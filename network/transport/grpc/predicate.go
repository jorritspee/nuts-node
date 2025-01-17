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

package grpc

import (
	"github.com/nuts-foundation/go-did/did"
	"github.com/nuts-foundation/nuts-node/network/transport"
)

// Predicate matches the connection based on a condition
type Predicate interface {
	Match(conn Connection) bool
}

type peerIDPredicate struct {
	peerID transport.PeerID
}

// ByPeerID filters the connection on the Peer ID
func ByPeerID(peerID transport.PeerID) Predicate {
	return peerIDPredicate{peerID: peerID}
}

func (predicate peerIDPredicate) Match(conn Connection) bool {
	return conn.Peer().ID == predicate.peerID
}

type connectedPredicate struct {
	connected bool
}

// ByConnected filters the connection by the connection state (if it's connected)
func ByConnected() Predicate {
	return connectedPredicate{connected: true}
}

// ByNotConnected filters the connection by the connection state (if it's not connected)
func ByNotConnected() Predicate {
	return connectedPredicate{connected: false}
}

func (predicate connectedPredicate) Match(conn Connection) bool {
	return conn.IsConnected() == predicate.connected
}

type nodeDIDPredicate struct {
	nodeDID did.DID
}

func (predicate nodeDIDPredicate) Match(conn Connection) bool {
	return conn.Peer().NodeDID.Equals(predicate.nodeDID)
}

// ByNodeDID filters the connection by the node DID
func ByNodeDID(nodeDID did.DID) Predicate {
	return nodeDIDPredicate{nodeDID: nodeDID}
}
