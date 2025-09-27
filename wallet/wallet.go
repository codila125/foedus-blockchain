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
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey []byte
	PublicKey  []byte
}

func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	return address
}

func ValidateAddress(address string) bool {
	fullHash := Base58Decode([]byte(address))
	actualChecksum := fullHash[len(fullHash)-checksumLength:]
	version := fullHash[0]
	pubKeyHash := fullHash[1 : len(fullHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return string(actualChecksum) == string(targetChecksum)
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pub := append(private.X.Bytes(), private.Y.Bytes()...)
	return *private, pub
}

func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private.D.Bytes(), public}

	return &wallet
}

func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)
	return pubHash[:]
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}

func (w *Wallet) ReconstructECDSAKey() (*ecdsa.PrivateKey, error) {
	curve := elliptic.P256() // Assuming P256 is always used for this blockchain

	d := new(big.Int).SetBytes(w.PrivateKey)

	// Reconstruct X and Y from PublicKey
	pubKeyLen := len(w.PublicKey)
	if pubKeyLen%2 != 0 {
		return nil, fmt.Errorf("invalid public key length for reconstruction")
	}
	x := new(big.Int).SetBytes(w.PublicKey[:pubKeyLen/2])
	y := new(big.Int).SetBytes(w.PublicKey[pubKeyLen/2:])

	publicKey := ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	privateKey := &ecdsa.PrivateKey{
		PublicKey: publicKey,
		D:         d,
	}

	return privateKey, nil
}

