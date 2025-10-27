package wallet

import (
	"log"
	"crypto/rand"
	"crypto/sha256"
	"crypto/ed25519"
)

const (
	checksumLength = 4          // Length of the checksum in bytes
	version        = byte(0x00) // Version byte to prepend to the public key hash (0x00 for Bitcoin addresses)
)

type Wallet struct {
	PublicKey   ed25519.PublicKey   // Public key in bytes
	PrivateKey  ed25519.PrivateKey // Private key in bytes
}

func (w Wallet) Address() []byte {
	/*
		Generates a wallet address from the wallet's public key.
	*/
	pubHash := PublicKeyHash(w.PublicKey) // Hash the public key

	versionedHash := append([]byte{version}, pubHash...) // Prepend the version byte to the public key hash
	checksum := Checksum(versionedHash)                  // Compute the checksum of the versioned hash

	fullHash := append(versionedHash, checksum...) // Encode the full hash using Base58 to get the address
	address := Base58Encode(fullHash)              // Get the address in Base58 format

	return address
}

func ValidateAddress(address string) bool {
	/*
		Validates a given wallet address by checking its checksum.
		Returns true if the address is valid, false otherwise.
	*/
	fullHash := Base58Decode([]byte(address))
	actualChecksum := fullHash[len(fullHash)-checksumLength:] // Extract the checksum from the end of the address

	version := fullHash[0]                                             // Extract the version byte
	pubKeyHash := fullHash[1 : len(fullHash)-checksumLength]           // Extract the public key hash because checksum is at the end
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...)) // Recompute the checksum of the version and pubKeyHash

	return string(actualChecksum) == string(targetChecksum) // Compare the extracted checksum with the recomputed checksum
}

func NewKeyPair() (ed25519.PublicKey, ed25519.PrivateKey) {
	/*
		Generates a new ECDSA private and public key pair using the P256 curve.
		Returns the private key and the public key as a byte slice.
	*/

	public, private, err := ed25519.GenerateKey(rand.Reader) // Generate a new private key
	if err != nil {
		log.Panic(err)
	}

	return public, private
}

func MakeWallet() *Wallet {
	/*
		Creates a new wallet using a newly generated key pair.
	*/
	public, private := NewKeyPair()
	wallet := Wallet{PublicKey: public, PrivateKey: private}

	return &wallet
}

func PublicKeyHash(pubKey []byte) []byte {
	/*
		Generates a SHA-256 hash of the public key.
	*/
	pubHash := sha256.Sum256(pubKey)
	return pubHash[:]
}

func Checksum(payload []byte) []byte {
	/*
		Computes the checksum for a given payload.
		The checksum is the first four bytes of the double SHA-256 hash of the payload.
	*/
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}
