// Package wallet provides the functionality for persisting and loading wallet
// data to and from the file system. It uses Protocol Buffers for serialization
// to ensure data is stored in a compact and efficient format.
package wallet

import (
	"fmt"
	"os"

	"github.com/codila125/foedus-blockchain/protobuf"
	"google.golang.org/protobuf/proto"
)

// LoadFile populates the Wallets collection from a data file on disk. It reads
// the serialized wallet data, unmarshals it using Protocol Buffers, and restores
// the wallets into the in-memory map. If the file does not exist, it returns an error.
func (ws *Wallets) LoadFile(nodeID string) error {
	walletPath := fmt.Sprintf(walletFile, nodeID)

	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		return fmt.Errorf("[WALLET] wallet file does not exist: %w", err)
	}

	fileContent, err := os.ReadFile(walletPath)
	if err != nil {
		return fmt.Errorf("[WALLET] failed to read wallet file: %w", err)
	}

	protoWallets := &protobuf.Wallets{}
	if err := proto.Unmarshal(fileContent, protoWallets); err != nil {
		return fmt.Errorf("[WALLET] failed to unmarshal wallets: %w", err)
	}

	ws.Wallets = make(map[string]*Wallet)
	for address, w := range protoWallets.Wallets {
		ws.Wallets[address] = &Wallet{
			PrivateKey: w.PrivateKey,
			PublicKey:  w.PublicKey,
		}
	}

	return nil
}

// SaveFile serializes the current collection of wallets and writes them to a
// data file. It ensures the target directory exists, marshals the wallets into
// the Protocol Buffers format, and saves the data to disk with restricted permissions.
func (ws *Wallets) SaveFile(nodeID string) error {
	walletPath := fmt.Sprintf(walletFile, nodeID)

	if err := os.MkdirAll("./temp", 0o755); err != nil {
		return fmt.Errorf("[WALLET] failed to create temp directory: %w", err)
	}

	protoWallets := &protobuf.Wallets{
		Wallets: make(map[string]*protobuf.Wallet),
	}

	for address, w := range ws.Wallets {
		protoWallets.Wallets[address] = &protobuf.Wallet{
			PrivateKey: w.PrivateKey,
			PublicKey:  w.PublicKey,
		}
	}

	data, err := proto.Marshal(protoWallets)
	if err != nil {
		return fmt.Errorf("[WALLET] failed to marshal wallets: %w", err)
	}

	err = os.WriteFile(walletPath, data, 0o600)
	if err != nil {
		return fmt.Errorf("[WALLET] failed to save wallets to file: %w", err)
	}
	return nil
}
