// Package wallet provides functionalities for creating, managing, and using
// cryptographic wallets within the Foedus blockchain. It handles key generation,
// address creation, and validation, forming the core of user identity and
// transaction authorization.
package wallet

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"log"
)

const (
	// checksumLength specifies the length of the checksum in bytes, used for address validation.
	checksumLength = 4
	// version is a byte prefix used to identify the network (e.g., mainnet, testnet)
	// in the generated wallet addresses.
	version = byte(0x00)
)

// Wallet defines a cryptographic wallet, which includes a key pair for signing
// and verifying transactions. The public key is used to derive the wallet address,
// while the private key must be kept secret to maintain control over the funds.
type Wallet struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// Address generates a Base58-encoded public address for the wallet. The address
// is derived from the public key and includes a version prefix and a checksum
// to ensure its integrity and prevent typos.
func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	return address
}

// ValidateAddress checks if a given blockchain address is valid. It decodes the
// Base58-encoded address and verifies the checksum to ensure the address has not
// been tampered with or corrupted. It returns true if the address is valid.
func ValidateAddress(address string) bool {
	fullHash := Base58Decode([]byte(address))
	actualChecksum := fullHash[len(fullHash)-checksumLength:]

	version := fullHash[0]
	pubKeyHash := fullHash[1 : len(fullHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return string(actualChecksum) == string(targetChecksum)
}

// NewKeyPair creates a new Ed25519 public-private key pair. This function is
// fundamental for generating new wallets, as the key pair is essential for
// transaction signing and address generation. It panics if key generation fails.
func NewKeyPair() (ed25519.PublicKey, ed25519.PrivateKey) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Print(err)
	}

	return public, private
}

// MakeWallet creates and returns a new Wallet instance, complete with a newly
// generated public-private key pair. This is the standard method for creating
// a new wallet in the system.
func MakeWallet() *Wallet {
	public, private := NewKeyPair()
	wallet := Wallet{PublicKey: public, PrivateKey: private}

	return &wallet
}

// PublicKeyHash computes the SHA-256 hash of a public key. This is a crucial
// step in generating a wallet address, as it condenses the public key into a
// fixed-size hash.
func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)
	return pubHash[:]
}

// Checksum generates a checksum for a given payload by performing a double
// SHA-256 hash and taking the first few bytes. This checksum is appended to
// wallet addresses to ensure their integrity.
func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}
