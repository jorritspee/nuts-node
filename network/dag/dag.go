/*
 * Copyright (C) 2021 Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package dag

import (
	"context"
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/nuts-foundation/go-stoabs"
	"github.com/nuts-foundation/nuts-node/core"
	"github.com/nuts-foundation/nuts-node/crypto/hash"
	"github.com/nuts-foundation/nuts-node/network/log"
)

// metadataShelf is the name of the shelf that holds metadata.
const metadataShelf = "metadata"

// numberOfTransactionsKey is the key of the metadata property that holds the number of transactions on the DAG
const numberOfTransactionsKey = "tx_num"

// highestClockValue is the key of the metadata property that holds the highest lamport clock value of all transactions on the DAG
const highestClockValue = "lc_high"

// headRefKey is the of the metadata property that holds the latest HEAD of the DAG
const headRefKey = "head_ref"

// transactionsShelf is the name of the shelf that holds the actual transactions.
const transactionsShelf = "documents"

// headsShelf contains the name of the shelf the holds the heads.
const headsShelf = "heads"

// clockShelf is the name of the shelf that uses the Lamport clock as index to a set of TX refs.
const clockShelf = "clocks"

// TransactionCountDiagnostic is the name of the diagnostics result for the transaction count
const TransactionCountDiagnostic = "transaction_count"

type dag struct {
	db stoabs.KVStore
}

type numberOfTransactionsStatistic struct {
	numberOfTransactions uint
}

func (n numberOfTransactionsStatistic) Result() interface{} {
	return n.numberOfTransactions
}

func (n numberOfTransactionsStatistic) Name() string {
	return TransactionCountDiagnostic
}

func (n numberOfTransactionsStatistic) String() string {
	return fmt.Sprintf("%d", n.numberOfTransactions)
}

type dataSizeStatistic struct {
	sizeInBytes int64
}

func (s dataSizeStatistic) Result() interface{} {
	return s.sizeInBytes
}

func (s dataSizeStatistic) Name() string {
	return "stored_database_size_bytes"
}

func (s dataSizeStatistic) String() string {
	return fmt.Sprintf("%d", s.sizeInBytes)
}

// newDAG creates a DAG backed by the given database.
func newDAG(db stoabs.KVStore) *dag {
	return &dag{db: db}
}

func (d *dag) Migrate() error {
	return d.db.Write(context.Background(), func(tx stoabs.WriteTx) error {
		writer, err := tx.GetShelfWriter(metadataShelf)
		if err != nil {
			return err
		}
		// Migrate highest LC value
		// Todo: remove after V5 release
		highestLCBytes, err := writer.Get(stoabs.BytesKey(highestClockValue))
		if err != nil {
			return err
		}
		if highestLCBytes == nil {
			log.Logger().Info("Highest LC value not stored, migrating...")
			highestLC := d.getHighestClockLegacy(tx)
			err = d.setHighestClockValue(tx, highestLC)
			if err != nil {
				return err
			}
		}
		// Migrate number of TXs
		// Todo: remove after V5 release
		numberOfTXs, err := writer.Get(stoabs.BytesKey(numberOfTransactionsKey))
		if err != nil {
			return err
		}
		if numberOfTXs == nil {
			log.Logger().Info("Number of transactions not stored, migrating...")
			numberOfTXs := d.getNumberOfTransactionsLegacy(tx)
			err = d.setNumberOfTransactions(tx, numberOfTXs)
			if err != nil {
				return err
			}
		}

		// Migrate headsLegacy to single head
		// Todo: remove after V6 release => then remove headsShelf
		headRef, err := writer.Get(stoabs.BytesKey(headRefKey))
		if err != nil {
			return err
		}
		if headRef == nil {
			log.Logger().Info("Head not stored in metadata, migrating...")
			heads := d.headsLegacy(tx)
			if len(heads) != 0 { // ignore for empty node
				var latestHead hash.SHA256Hash
				var latestLC uint32

				for _, ref := range heads {
					transaction, err := getTransaction(ref, tx)
					if err != nil {
						return err
					}
					if transaction.Clock() >= latestLC {
						latestHead = ref
						latestLC = transaction.Clock()
					}
				}

				err = d.setHead(tx, latestHead)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (d *dag) diagnostics(ctx context.Context) []core.DiagnosticResult {
	var stats Statistics
	_ = d.db.Read(ctx, func(tx stoabs.ReadTx) error {
		stats = d.statistics(tx)
		return nil
	})

	result := make([]core.DiagnosticResult, 0)
	result = append(result, numberOfTransactionsStatistic{numberOfTransactions: stats.NumberOfTransactions})
	result = append(result, dataSizeStatistic{sizeInBytes: stats.DataSize})
	return result
}

func (d dag) headsLegacy(ctx stoabs.ReadTx) []hash.SHA256Hash {
	result := make([]hash.SHA256Hash, 0)
	reader := ctx.GetShelfReader(headsShelf)
	_ = reader.Iterate(func(key stoabs.Key, _ []byte) error {
		result = append(result, hash.FromSlice(key.Bytes())) // FromSlice() copies
		return nil
	}, stoabs.HashKey{})
	return result
}

func (d *dag) findBetweenLC(tx stoabs.ReadTx, startInclusive uint32, endExclusive uint32) ([]Transaction, error) {
	var result []Transaction

	err := d.visitBetweenLC(tx, startInclusive, endExclusive, func(transaction Transaction) bool {
		result = append(result, transaction)
		return true
	})
	if err != nil {
		// Make sure not to return results in case of error
		return nil, err
	}
	return result, nil
}

func (d *dag) visitBetweenLC(tx stoabs.ReadTx, startInclusive uint32, endExclusive uint32, visitor Visitor) error {
	reader := tx.GetShelfReader(clockShelf)

	// TODO: update to process in batches
	return reader.Range(stoabs.Uint32Key(startInclusive), stoabs.Uint32Key(endExclusive), func(_ stoabs.Key, value []byte) error {
		parsed := parseHashList(value)
		// according to RFC004, lower byte value refs go first
		sort.Slice(parsed, func(i, j int) bool {
			return parsed[i].Compare(parsed[j]) <= 0
		})
		for _, next := range parsed {
			transaction, err := getTransaction(next, tx)
			if err != nil {
				return err
			}
			visitor(transaction)
		}
		return nil
	}, true)
}

func (d *dag) isPresent(tx stoabs.ReadTx, ref hash.SHA256Hash) bool {
	return exists(tx.GetShelfReader(transactionsShelf), ref)
}

func (d *dag) add(tx stoabs.WriteTx, transactions ...Transaction) error {
	highestLC := d.getHighestClockValue(tx)
	headRef := hash.EmptyHash()

	for _, transaction := range transactions {
		if transaction != nil {
			if err := d.addSingle(tx, transaction); err != nil {
				return err
			}
			if transaction.Clock() > highestLC || transaction.Clock() == 0 {
				highestLC = transaction.Clock()
				headRef = transaction.Ref()
			}
		}
	}

	// update highest LC
	if err := d.setHighestClockValue(tx, highestLC); err != nil {
		return err
	}

	// update head
	if !headRef.Equals(hash.EmptyHash()) {
		if err := d.setHead(tx, headRef); err != nil {
			return err
		}
	}

	// update TX count
	txCount := d.getNumberOfTransactions(tx) + uint64(len(transactions))
	return d.setNumberOfTransactions(tx, txCount)
}

func (d dag) getNumberOfTransactions(tx stoabs.ReadTx) uint64 {
	value, err := tx.GetShelfReader(metadataShelf).Get(stoabs.BytesKey(numberOfTransactionsKey))
	if err != nil {
		log.Logger().
			WithError(err).
			Error("Unable to retrieve number of transactions")
		return 0
	}
	if value != nil {
		return bytesToCount(value)
	}
	return 0
}

func (d dag) setNumberOfTransactions(tx stoabs.WriteTx, count uint64) error {
	writer, err := tx.GetShelfWriter(metadataShelf)
	if err != nil {
		return err
	}
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes[:], count)

	return writer.Put(stoabs.BytesKey(numberOfTransactionsKey), bytes)
}

func (d dag) setHead(tx stoabs.WriteTx, ref hash.SHA256Hash) error {
	writer, err := tx.GetShelfWriter(metadataShelf)
	if err != nil {
		return err
	}

	return writer.Put(stoabs.BytesKey(headRefKey), ref.Slice())
}

func (d dag) getHighestClockValue(tx stoabs.ReadTx) uint32 {
	value, err := tx.GetShelfReader(metadataShelf).Get(stoabs.BytesKey(highestClockValue))
	if err != nil {
		log.Logger().
			WithError(err).
			Error("Unable to retrieve highest LC value")
		return 0
	}
	if value == nil {
		return 0
	}
	return bytesToClock(value)
}

// getHighestClockLegacy is used for migration.
// Remove after V5 or V6 release?
func (d dag) getHighestClockLegacy(tx stoabs.ReadTx) uint32 {
	reader := tx.GetShelfReader(clockShelf)
	var clock uint32
	err := reader.Iterate(func(key stoabs.Key, _ []byte) error {
		currentClock := uint32(key.(stoabs.Uint32Key))
		if currentClock > clock {
			clock = currentClock
		}
		return nil
	}, stoabs.Uint32Key(0))
	if err != nil {
		log.Logger().
			WithError(err).
			Error("Failed to read clock shelf")
		return 0
	}
	return clock
}

func (d dag) getHead(tx stoabs.ReadTx) (hash.SHA256Hash, error) {
	head, err := tx.GetShelfReader(metadataShelf).Get(stoabs.BytesKey(headRefKey))
	if err != nil {
		return hash.EmptyHash(), err
	}

	return hash.FromSlice(head), nil
}

// getNumberOfTransactionsLegacy is used for migration.
// Remove after V5 or V6 release?
func (d dag) getNumberOfTransactionsLegacy(tx stoabs.ReadTx) uint64 {
	reader := tx.GetShelfReader(transactionsShelf)
	return uint64(reader.Stats().NumEntries)
}

func (d dag) setHighestClockValue(tx stoabs.WriteTx, count uint32) error {
	writer, err := tx.GetShelfWriter(metadataShelf)
	if err != nil {
		return err
	}
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes[:], count)

	return writer.Put(stoabs.BytesKey(highestClockValue), bytes)
}

func (d dag) statistics(tx stoabs.ReadTx) Statistics {
	var result Statistics
	shelfStats := tx.GetShelfReader(transactionsShelf).Stats()
	result.DataSize = int64(shelfStats.ShelfSize)
	result.NumberOfTransactions = uint(d.getNumberOfTransactions(tx))

	return result
}

func (d *dag) addSingle(tx stoabs.WriteTx, transaction Transaction) error {
	ref := transaction.Ref()
	refKey := stoabs.NewHashKey(ref)
	transactions, lc, err := getBuckets(tx)
	if err != nil {
		return err
	}
	if exists(transactions, ref) {
		log.Logger().
			WithField(core.LogFieldTransactionRef, ref).
			Trace("Transaction already exists, not adding it again.")
		return nil
	}
	if len(transaction.Previous()) == 0 {
		if getRoots(lc) != nil {
			return errRootAlreadyExists
		}
	}
	if err := indexClockValue(tx, transaction); err != nil {
		return fmt.Errorf("unable to calculate LC value for %s: %w", ref, err)
	}
	return transactions.Put(refKey, transaction.Data())
}

func indexClockValue(tx stoabs.WriteTx, transaction Transaction) error {
	lc, err := tx.GetShelfWriter(clockShelf)
	if err != nil {
		return err
	}

	clockKey := stoabs.Uint32Key(transaction.Clock())
	ref := transaction.Ref()
	currentRefs, err := lc.Get(clockKey)
	if err != nil {
		return err
	}
	for _, cRef := range parseHashList(currentRefs) {
		if ref.Equals(cRef) {
			// should only be in the list once
			return nil
		}
	}
	if err := lc.Put(clockKey, appendHashList(currentRefs, ref)); err != nil {
		return err
	}

	log.Logger().
		WithField(core.LogFieldTransactionRef, ref).
		Tracef("Storing transaction logical clock (LC: %d)", clockKey)

	return nil
}

func bytesToClock(clockBytes []byte) uint32 {
	return binary.BigEndian.Uint32(clockBytes)
}

func bytesToCount(clockBytes []byte) uint64 {
	return binary.BigEndian.Uint64(clockBytes)
}

func getBuckets(tx stoabs.WriteTx) (transactions, lc stoabs.Writer, err error) {
	if transactions, err = tx.GetShelfWriter(transactionsShelf); err != nil {
		return
	}
	if lc, err = tx.GetShelfWriter(clockShelf); err != nil {
		return
	}
	return
}

func getRoots(lcBucket stoabs.Reader) []hash.SHA256Hash {
	roots, err := lcBucket.Get(stoabs.Uint32Key(0))
	if err != nil {
		return nil
	}
	return parseHashList(roots) // no need to copy, calls FromSlice() (which copies)
}

func getTransaction(hash hash.SHA256Hash, tx stoabs.ReadTx) (Transaction, error) {
	transactions := tx.GetShelfReader(transactionsShelf)

	transactionBytes, err := transactions.Get(stoabs.NewHashKey(hash))
	if err != nil || transactionBytes == nil {
		return nil, err
	}
	parsedTx, err := ParseTransaction(transactionBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse transaction %s: %w", hash, err)
	}

	return parsedTx, nil
}

// exists checks whether the transaction with the given ref exists.
func exists(transactions stoabs.Reader, ref hash.SHA256Hash) bool {
	val, _ := transactions.Get(stoabs.NewHashKey(ref))
	return val != nil
}

// parseHashList splits a list of concatenated hashes into separate hashes.
func parseHashList(input []byte) []hash.SHA256Hash {
	if len(input) == 0 {
		return nil
	}
	num := (len(input) - (len(input) % hash.SHA256HashSize)) / hash.SHA256HashSize
	result := make([]hash.SHA256Hash, num)
	for i := 0; i < num; i++ {
		result[i] = hash.FromSlice(input[i*hash.SHA256HashSize : i*hash.SHA256HashSize+hash.SHA256HashSize])
	}
	return result
}

func appendHashList(list []byte, h hash.SHA256Hash) []byte {
	newList := make([]byte, 0, len(list)+hash.SHA256HashSize)
	newList = append(newList, list...)
	newList = append(newList, h.Slice()...)
	return newList
}
