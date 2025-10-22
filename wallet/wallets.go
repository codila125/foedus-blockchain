package wallet

import (
	"fmt"
	"log"
)

const walletFile = "./temp/wallets_%s.data"

type Wallets struct {
	Wallets map[string]*Wallet // Map of wallet addresses to Wallet instances
}

func CreateWallets(nodeID string) (*Wallets, error) {
	/*
		Creates a new Wallets instance and loads existing wallets from the wallet file if it exists.
	*/
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFile(nodeID) // Load existing wallets from file

	return &wallets, err
}

func (ws *Wallets) AddWallet() string {
	/*
		Creates a new wallet, adds it to the Wallets instance, and returns the new address.
	*/
	wallet := MakeWallet()
	address := string(wallet.Address()) // Get the address of the new wallet

	ws.Wallets[address] = wallet

	return address
}

func (ws *Wallets) GetAllAddresses() []string {
	/*
		Returns a slice of all wallet addresses stored in the Wallets instance.
	*/
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (ws Wallets) GetWallet(address string) (Wallet, error) {
	/*
		Returns the Wallet instance associated with the given address.
	*/
	if _, exists := ws.Wallets[address]; !exists {
		log.Printf("[WALLET] Wallet not found for address: %s", address)
		return Wallet{}, fmt.Errorf("[WALLET] Wallet not found for address: %s", address)
	}
	return *ws.Wallets[address], nil
}
