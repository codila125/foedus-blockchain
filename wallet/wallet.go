package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"
)

const (
	checksumLength = 4          // Length of the checksum in bytes
	version        = byte(0x00) // Version byte to prepend to the public key hash (0x00 for Bitcoin addresses)
)

type Wallet struct {
	PrivateKey []byte // Private key in bytes
	PublicKey  []byte // Public key in bytes
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

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	/*
		Generates a new ECDSA private and public key pair using the P256 curve.
		Returns the private key and the public key as a byte slice.
	*/
	curve := elliptic.P256() // Using P256 curve for key generation

	private, err := ecdsa.GenerateKey(curve, rand.Reader) // Generate a new private key
	if err != nil {
		log.Panic(err)
	}

	pub := append(private.X.Bytes(), private.Y.Bytes()...) // Concatenate X and Y coordinates of the private key to form the public key
	return *private, pub
}

func MakeWallet() *Wallet {
	/*
		Creates a new wallet using a newly generated key pair.
	*/
	private, public := NewKeyPair()
	wallet := Wallet{private.D.Bytes(), public}

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

func (w *Wallet) ReconstructECDSAKey() (*ecdsa.PrivateKey, error) {
	/*
		Reconstructs the ECDSA private key from the wallet's private key bytes.
		Returns the reconstructed private key.
	*/
	curve := elliptic.P256() // Assuming P256 is always used for this blockchain

	d := new(big.Int).SetBytes(w.PrivateKey) // Set D from the private key bytes

	// Reconstruct X and Y from PublicKey
	pubKeyLen := len(w.PublicKey)
	if pubKeyLen%2 != 0 {
		return nil, fmt.Errorf("invalid public key length for reconstruction")
	}

	x := new(big.Int).SetBytes(w.PublicKey[:pubKeyLen/2]) // X is the first half of the public key bytes
	y := new(big.Int).SetBytes(w.PublicKey[pubKeyLen/2:]) // Y is the second half of the public key bytes

	// Reconstruct the public key
	publicKey := ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	// Reconstruct the private key
	privateKey := &ecdsa.PrivateKey{
		PublicKey: publicKey,
		D:         d,
	}

	return privateKey, nil
}
