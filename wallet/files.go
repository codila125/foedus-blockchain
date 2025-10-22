package wallet

import (
	"fmt"
	"os"

	"github.com/codila125/foedus-blockchain/protobuf"
	"google.golang.org/protobuf/proto"
)

func (ws *Wallets) LoadFile(nodeID string) error {
	/*
		Loads wallets from the wallet file into the Wallets instance.
		Returns an error if the file does not exist or if there is an issue reading it.
	*/
	walletPath := fmt.Sprintf(walletFile, nodeID)

	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		return fmt.Errorf("[WALLET] Wallet file does not exist: %w", err)
	}

	// Read the wallet file
	fileContent, err := os.ReadFile(walletPath)
	if err != nil {
		return fmt.Errorf("[WALLET] Failed to read wallet file: %w", err)
	}

	protoWallets := &protobuf.Wallets{}
	if err := proto.Unmarshal(fileContent, protoWallets); err != nil {
		return fmt.Errorf("[WALLET] Failed to unmarshal wallets: %w", err)
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

func (ws *Wallets) SaveFile(nodeID string) error {
	/*
		Saves the Wallets instance to the wallet file.
	*/
	walletPath := fmt.Sprintf(walletFile, nodeID)

	if err := os.MkdirAll("./temp", 0o755); err != nil {
		return fmt.Errorf("[WALLET] Failed to create temp directory: %w", err)
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
		return fmt.Errorf("[WALLET] Failed to marshal wallets: %w", err)
	}

	err = os.WriteFile(walletPath, data, 0o600) // Write the buffer content to the wallet file
	if err != nil {
		return fmt.Errorf("[WALLET] Failed to save wallets to file: %w", err)
	}
	return nil
}

