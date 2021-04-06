/*
 * Nuts node
 * Copyright (C) 2021. Nuts community
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
 */

package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nuts-foundation/go-did/did"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/nuts-foundation/nuts-node/core"
	api "github.com/nuts-foundation/nuts-node/vdr/api/v1"
)

// FlagSet contains flags relevant for the VDR instance
func FlagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("vdr", pflag.ContinueOnError)
	return flagSet
}

// Cmd contains sub-commands for the remote client
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vdr",
		Short: "Verifiable Data VDR commands",
	}

	cmd.AddCommand(createCmd())

	cmd.AddCommand(resolveCmd())

	cmd.AddCommand(updateCmd())

	cmd.AddCommand(deactivateCmd())

	cmd.AddCommand(addVerificationMethodCmd())

	cmd.AddCommand(deleteVerificationMethodCmd())

	return cmd
}

func createCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-did",
		Short: "Registers a new DID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := httpClient(core.NewClientConfig(cmd.Flags())).Create()
			if err != nil {
				return fmt.Errorf("unable to create new DID: %v", err)
			}
			bytes, _ := json.MarshalIndent(doc, "", "  ")
			cmd.Println(string(bytes))
			return nil
		},
	}
}

func updateCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update [DID] [hash] [file]",
		Short: "Update a DID with the given DID document, this replaces the DID document. " +
			"If no file is given, a pipe is assumed. The hash is needed to prevent concurrent updates.",
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			hash := args[1]

			var bytes []byte
			var err error
			if len(args) == 3 {
				// read from file
				bytes, err = os.ReadFile(args[2])
				if err != nil {
					return fmt.Errorf("failed to read file %s: %w", args[2], err)
				}
			} else {
				// read from stdin
				bytes, err = readFromStdin()
				if err != nil {
					return fmt.Errorf("failed to read from pipe: %w", err)
				}
			}

			// parse
			var didDoc did.Document
			if err = json.Unmarshal(bytes, &didDoc); err != nil {
				return fmt.Errorf("failed to parse DID document: %w", err)
			}

			if _, err = httpClient(core.NewClientConfig(cmd.Flags())).Update(id, hash, didDoc); err != nil {
				return fmt.Errorf("failed to update DID document: %w", err)
			}

			cmd.Println("DID document updated")
			return nil
		},
	}
}

func resolveCmd() *cobra.Command {
	var printMetadata bool
	var printDocument bool
	result := &cobra.Command{
		Use:   "resolve [DID]",
		Short: "Resolve a DID document based on its DID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, meta, err := httpClient(core.NewClientConfig(cmd.Flags())).Get(args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve DID document: %v", err)
			}

			var toPrint []interface{}
			if !printMetadata && !printDocument {
				toPrint = append(toPrint, doc, meta)
			} else {
				if printDocument {
					toPrint = append(toPrint, doc)
				}
				if printMetadata {
					toPrint = append(toPrint, meta)
				}
			}
			for _, o := range toPrint {
				bytes, _ := json.MarshalIndent(o, "", "  ")
				cmd.Printf("%s\n", string(bytes))
			}

			return nil
		},
	}
	result.Flags().BoolVar(&printMetadata, "metadata", false, "Pass 'true' to only print the metadata (unless other flags are provided as well).")
	result.Flags().BoolVar(&printDocument, "document", false, "Pass 'true' to only print the document (unless other flags are provided as well).")
	return result
}

func deactivateCmd() *cobra.Command {
	result := &cobra.Command{
		Use:   "deactivate [DID]",
		Short: "Deactivate a DID document based on its DID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !askYesNo("This will delete the DID document, are you sure?", cmd) {
				cmd.Println("Deactivation cancelled")
				return nil
			}
			err := httpClient(core.NewClientConfig(cmd.Flags())).Deactivate(args[0])
			if err != nil {
				return fmt.Errorf("failed to deactivate DID document: %v", err)
			}
			cmd.Println("DID document deactivated")

			return nil
		},
	}
	return result
}

func addVerificationMethodCmd() *cobra.Command {
	result := &cobra.Command{
		Use:   "addvm [DID]",
		Short: "Add a verification method key to the DID document.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			verificationMethod, err := httpClient(core.NewClientConfig(cmd.Flags())).AddNewVerificationMethod(args[0])
			if err != nil {
				return fmt.Errorf("failed to add a new verification method to DID document: %s", err.Error())
			}
			bytes, _ := json.MarshalIndent(verificationMethod, "", "  ")
			cmd.Printf("%s\n", string(bytes))
			return nil
		},
	}

	return result
}

func deleteVerificationMethodCmd() *cobra.Command {
	result := &cobra.Command{
		Use:   "delvm [DID] [kid]",
		Short: "Deletes a verification method from the DID document.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := httpClient(core.NewClientConfig(cmd.Flags())).DeleteVerificationMethod(args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to delete the verification method from DID document: %s", err.Error())
			}
			cmd.Println("Verification method deleted from the DID document.")
			return nil
		},
	}

	return result
}

func askYesNo(question string, cmd *cobra.Command) (answer bool) {
	reader := bufio.NewReader(cmd.InOrStdin())
	question += "[yes/no]: "

	for true {
		cmd.Print(question)
		s, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		s = strings.TrimSuffix(s, "\n")
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			answer = true
			break
		} else if s == "n" || s == "no" {
			break
		}
		cmd.Println("invalid answer")
	}
	return
}

func readFromStdin() ([]byte, error) {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		return nil, errors.New("expected piped input")
	}
	return io.ReadAll(bufio.NewReader(os.Stdin))
}

// httpClient creates a remote client
func httpClient(config core.ClientConfig) api.HTTPClient {
	return api.HTTPClient{
		ServerAddress: config.GetAddress(),
		Timeout:       config.Timeout,
	}
}
