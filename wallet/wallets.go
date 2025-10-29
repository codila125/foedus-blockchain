// Package wallet provides the functionality to manage a collection of wallets,
// including their creation, storage, and retrieval. It acts as a container for
// individual wallet instances, allowing them to be accessed by their unique addresses.
package wallet

import (
	"fmt"
	"log"
)

// walletFile defines the file path pattern for storing and loading wallet data.
// The '%s' placeholder is replaced with a node-specific identifier to ensure
// that each node maintains its own set of wallets.
const walletFile = "./temp/wallets_%s.data"

// Wallets represents a collection of Wallet instances, indexed by their
// Base58-encoded addresses. This structure provides an in-memory store for
// managing multiple wallets within a single node.
type Wallets struct {
	Wallets map[string]*Wallet
}

// CreateWallets initializes and returns a new Wallets collection. It attempts
// to load any existing wallets from the file system for the given nodeID.
// If no wallet file is found, an empty collection is created.
func CreateWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFile(nodeID)

	return &wallets, err
}

// AddWallet creates a new wallet, adds it to the collection, and returns its
// unique address. The new wallet is immediately available for use.
func (ws *Wallets) AddWallet() string {
	wallet := MakeWallet()
	address := string(wallet.Address())

	ws.Wallets[address] = wallet

	return address
}

// GetAllAddresses returns a slice containing all the wallet addresses currently
// managed in the collection. This is useful for listing all available wallets.
func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet retrieves a wallet from the collection by its address. It returns
// the Wallet instance if found; otherwise, it returns an error indicating that
// the wallet does not exist.
func (ws Wallets) GetWallet(address string) (Wallet, error) {
	if _, exists := ws.Wallets[address]; !exists {
		log.Printf("[WALLET] Wallet not found for address: %s", address)
		return Wallet{}, fmt.Errorf("[WALLET] wallet not found for address: %s", address)
	}
	return *ws.Wallets[address], nil
}
