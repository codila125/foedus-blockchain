package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"log"
	"os"
)

const walletFile = "./temp/wallets.data"

type Wallets struct {
	Wallets map[string]*Wallet // Map of wallet addresses to Wallet instances
}

func CreateWallets() (*Wallets, error) {
	/*
		Creates a new Wallets instance and loads existing wallets from the wallet file if it exists.
	*/
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFile() // Load existing wallets from file

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

func (ws Wallets) GetWallet(address string) Wallet {
	/*
		Returns the Wallet instance associated with the given address.
	*/
	return *ws.Wallets[address]
}

func (ws *Wallets) LoadFile() error {
	/*
		Loads wallets from the wallet file into the Wallets instance.
		Returns an error if the file does not exist or if there is an issue reading it.
	*/
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	var wallets Wallets

	// Read the wallet file
	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	gob.Register(elliptic.P256())                           // Register the elliptic curve for gob encoding/decoding
	decoder := gob.NewDecoder(bytes.NewReader(fileContent)) // Create a decoder for the file content
	err = decoder.Decode(&wallets)                          // Decode the file content into the wallets struct
	if err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}

func (ws *Wallets) SaveFile() {
	/*
		Saves the Wallets instance to the wallet file.
	*/
	var content bytes.Buffer

	gob.Register(elliptic.P256()) // Register the elliptic curve for gob encoding/decoding

	encoder := gob.NewEncoder(&content) // Create an encoder for the content buffer
	err := encoder.Encode(ws)           // Encode the Wallets instance into the buffer
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, content.Bytes(), 0644) // Write the buffer content to the wallet file
	if err != nil {
		log.Panic(err)
	}
}
