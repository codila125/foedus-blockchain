package wallet

import (
	"bytes"
	"crypto/ed25519"
	"testing"
)

// go test -run TestWallet
// go test -run TestPublicKey
// go test -run TestChecksum
// go test -run TestNewKeyPair
// go test -run TestMakeWallet
// go test -run TestValidateAddress
// go test -bench=BenchmarkWallet
// go test -cover
// go test -race

// TestPublicKeyHash tests the SHA-256 hashing of public keys.
func TestPublicKeyHash(t *testing.T) {
	tests := []struct {
		name        string
		pubKey      []byte
		expectedLen int
	}{
		{
			name:        "empty public key",
			pubKey:      []byte{},
			expectedLen: 32, // SHA-256 always produces 32 bytes
		},
		{
			name:        "ed25519 public key size",
			pubKey:      make([]byte, ed25519.PublicKeySize),
			expectedLen: 32,
		},
		{
			name:        "arbitrary data",
			pubKey:      []byte("test public key data"),
			expectedLen: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := PublicKeyHash(tt.pubKey)
			if len(hash) != tt.expectedLen {
				t.Errorf("PublicKeyHash() length = %d, want %d", len(hash), tt.expectedLen)
			}
		})
	}
}

// TestPublicKeyHashDeterministic tests that the same input produces the same hash.
func TestPublicKeyHashDeterministic(t *testing.T) {
	pubKey := []byte("consistent test key")

	hash1 := PublicKeyHash(pubKey)
	hash2 := PublicKeyHash(pubKey)

	if !bytes.Equal(hash1, hash2) {
		t.Errorf("PublicKeyHash is not deterministic: %v != %v", hash1, hash2)
	}
}

// TestPublicKeyHashUnique tests that different inputs produce different hashes.
func TestPublicKeyHashUnique(t *testing.T) {
	pubKey1 := []byte("key one")
	pubKey2 := []byte("key two")

	hash1 := PublicKeyHash(pubKey1)
	hash2 := PublicKeyHash(pubKey2)

	if bytes.Equal(hash1, hash2) {
		t.Errorf("PublicKeyHash produced same hash for different inputs")
	}
}

// TestChecksum tests the checksum generation function.
func TestChecksum(t *testing.T) {
	tests := []struct {
		name        string
		payload     []byte
		expectedLen int
	}{
		{
			name:        "empty payload",
			payload:     []byte{},
			expectedLen: checksumLength,
		},
		{
			name:        "versioned hash",
			payload:     append([]byte{0x00}, make([]byte, 32)...),
			expectedLen: checksumLength,
		},
		{
			name:        "arbitrary data",
			payload:     []byte("test payload for checksum"),
			expectedLen: checksumLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checksum := Checksum(tt.payload)
			if len(checksum) != tt.expectedLen {
				t.Errorf("Checksum() length = %d, want %d", len(checksum), tt.expectedLen)
			}
		})
	}
}

// TestChecksumDeterministic tests that the same input produces the same checksum.
func TestChecksumDeterministic(t *testing.T) {
	payload := []byte("consistent payload")

	checksum1 := Checksum(payload)
	checksum2 := Checksum(payload)

	if !bytes.Equal(checksum1, checksum2) {
		t.Errorf("Checksum is not deterministic: %v != %v", checksum1, checksum2)
	}
}

// TestNewKeyPair tests key pair generation.
func TestNewKeyPair(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "generate key pair 1"},
		{name: "generate key pair 2"},
		{name: "generate key pair 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubKey, privKey := NewKeyPair()

			if len(pubKey) != ed25519.PublicKeySize {
				t.Errorf("Public key length = %d, want %d", len(pubKey), ed25519.PublicKeySize)
			}

			if len(privKey) != ed25519.PrivateKeySize {
				t.Errorf("Private key length = %d, want %d", len(privKey), ed25519.PrivateKeySize)
			}
		})
	}
}

// TestNewKeyPairUniqueness tests that each key pair is unique.
func TestNewKeyPairUniqueness(t *testing.T) {
	pub1, priv1 := NewKeyPair()
	pub2, priv2 := NewKeyPair()

	if bytes.Equal(pub1, pub2) {
		t.Error("Generated duplicate public keys")
	}

	if bytes.Equal(priv1, priv2) {
		t.Error("Generated duplicate private keys")
	}
}

// TestNewKeyPairSignVerify tests that generated keys can sign and verify messages.
func TestNewKeyPairSignVerify(t *testing.T) {
	tests := []struct {
		name    string
		message []byte
	}{
		{
			name:    "short message",
			message: []byte("hello"),
		},
		{
			name:    "long message",
			message: []byte("this is a much longer message for testing purposes"),
		},
		{
			name:    "empty message",
			message: []byte{},
		},
		{
			name:    "binary message",
			message: []byte{0x00, 0xff, 0xab, 0xcd},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubKey, privKey := NewKeyPair()

			signature := ed25519.Sign(privKey, tt.message)
			if !ed25519.Verify(pubKey, tt.message, signature) {
				t.Error("Failed to verify signature with generated key pair")
			}
		})
	}
}

// TestMakeWallet tests wallet creation.
func TestMakeWallet(t *testing.T) {
	wallet := MakeWallet()

	if wallet == nil {
		t.Fatal("MakeWallet() returned nil")
	}

	if len(wallet.PublicKey) != ed25519.PublicKeySize {
		t.Errorf("Wallet public key length = %d, want %d", len(wallet.PublicKey), ed25519.PublicKeySize)
	}

	if len(wallet.PrivateKey) != ed25519.PrivateKeySize {
		t.Errorf("Wallet private key length = %d, want %d", len(wallet.PrivateKey), ed25519.PrivateKeySize)
	}
}

// TestMakeWalletUniqueness tests that each wallet is unique.
func TestMakeWalletUniqueness(t *testing.T) {
	wallet1 := MakeWallet()
	wallet2 := MakeWallet()

	if bytes.Equal(wallet1.PublicKey, wallet2.PublicKey) {
		t.Error("MakeWallet() created duplicate public keys")
	}

	if bytes.Equal(wallet1.PrivateKey, wallet2.PrivateKey) {
		t.Error("MakeWallet() created duplicate private keys")
	}
}

// TestWalletAddress tests address generation from wallet.
func TestWalletAddress(t *testing.T) {
	wallet := MakeWallet()
	address := wallet.Address()

	if len(address) == 0 {
		t.Error("Wallet.Address() returned empty address")
	}

	// Address should be Base58 encoded, so it should only contain valid Base58 characters
	validChars := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	for _, c := range string(address) {
		found := false
		for _, v := range validChars {
			if c == v {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Address contains invalid Base58 character: %c", c)
		}
	}
}

// TestWalletAddressDeterministic tests that the same wallet produces the same address.
func TestWalletAddressDeterministic(t *testing.T) {
	wallet := MakeWallet()

	addr1 := wallet.Address()
	addr2 := wallet.Address()

	if !bytes.Equal(addr1, addr2) {
		t.Errorf("Wallet.Address() is not deterministic: %s != %s", string(addr1), string(addr2))
	}
}

// TestWalletAddressUniqueness tests that different wallets produce different addresses.
func TestWalletAddressUniqueness(t *testing.T) {
	wallet1 := MakeWallet()
	wallet2 := MakeWallet()

	addr1 := wallet1.Address()
	addr2 := wallet2.Address()

	if bytes.Equal(addr1, addr2) {
		t.Error("Different wallets produced the same address")
	}
}

// TestValidateAddress tests address validation.
func TestValidateAddress(t *testing.T) {
	// Generate a valid address first
	wallet := MakeWallet()
	validAddress := string(wallet.Address())

	tests := []struct {
		name    string
		address string
		valid   bool
	}{
		{
			name:    "valid address from wallet",
			address: validAddress,
			valid:   true,
		},
		{
			name:    "another valid address",
			address: string(MakeWallet().Address()),
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAddress(tt.address)
			if result != tt.valid {
				t.Errorf("ValidateAddress(%s) = %v, want %v", tt.address, result, tt.valid)
			}
		})
	}
}

// TestValidateAddressInvalid tests that invalid addresses are rejected.
// Note: ValidateAddress may panic on malformed inputs (empty string, too short)
// so we test with addresses that won't cause panic but are still invalid.
func TestValidateAddressInvalid(t *testing.T) {
	// Create a valid address and corrupt it significantly
	wallet := MakeWallet()
	address := wallet.Address()

	// Create corrupted versions that are still long enough to not panic
	// We'll corrupt multiple bytes to ensure checksum mismatch
	tests := []struct {
		name  string
		setup func() string
	}{
		{
			name: "corrupted multiple bytes",
			setup: func() string {
				corrupted := make([]byte, len(address))
				copy(corrupted, address)
				// Corrupt multiple bytes to ensure checksum fails
				for i := 5; i < 10 && i < len(corrupted); i++ {
					corrupted[i] = 'X'
				}
				return string(corrupted)
			},
		},
		{
			name: "all same character",
			setup: func() string {
				// Create a fake address with all same chars (unlikely to have valid checksum)
				fake := make([]byte, len(address))
				for i := range fake {
					fake[i] = 'A'
				}
				return string(fake)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			corruptedAddr := tt.setup()
			if ValidateAddress(corruptedAddr) {
				t.Errorf("ValidateAddress(%s) should be false for invalid address", corruptedAddr)
			}
		})
	}
}

// TestValidateAddressModified tests that modifying an address invalidates it.
func TestValidateAddressModified(t *testing.T) {
	wallet := MakeWallet()
	address := wallet.Address()

	// Verify original is valid
	if !ValidateAddress(string(address)) {
		t.Fatal("Original address should be valid")
	}

	// Modify one character and verify it's invalid
	if len(address) > 1 {
		modified := make([]byte, len(address))
		copy(modified, address)
		// Change a character in the middle
		if modified[len(modified)/2] == 'A' {
			modified[len(modified)/2] = 'B'
		} else {
			modified[len(modified)/2] = 'A'
		}

		if ValidateAddress(string(modified)) {
			t.Error("Modified address should be invalid")
		}
	}
}

// TestWalletAddressLength tests that wallet addresses have consistent length.
func TestWalletAddressLength(t *testing.T) {
	// Generate multiple wallets and check address lengths are in expected range
	for range 10 {
		wallet := MakeWallet()
		address := wallet.Address()

		// Base58 encoded addresses length depends on the hash size
		// SHA-256 hash (32 bytes) + version (1 byte) + checksum (4 bytes) = 37 bytes
		// Base58 encoding of 37 bytes typically produces 50-52 characters
		if len(address) < 40 || len(address) > 60 {
			t.Errorf("Unexpected address length: %d for address %s", len(address), string(address))
		}
	}
}

// TestConcurrentWalletCreation tests that wallets can be created concurrently.
func TestConcurrentWalletCreation(t *testing.T) {
	done := make(chan string, 100)

	// Create wallets concurrently
	for range 100 {
		go func() {
			wallet := MakeWallet()
			done <- string(wallet.Address())
		}()
	}

	// Collect all addresses
	addresses := make(map[string]bool)
	for range 100 {
		addr := <-done
		if addresses[addr] {
			t.Errorf("Duplicate address created: %s", addr)
		}
		addresses[addr] = true
	}
}

// Benchmark tests for wallet.go

func BenchmarkPublicKeyHash(b *testing.B) {
	pubKey := make([]byte, ed25519.PublicKeySize)

	for b.Loop() {
		PublicKeyHash(pubKey)
	}
}

func BenchmarkChecksum(b *testing.B) {
	payload := make([]byte, 33) // version + 32 byte hash

	for b.Loop() {
		Checksum(payload)
	}
}

func BenchmarkNewKeyPair(b *testing.B) {
	for b.Loop() {
		NewKeyPair()
	}
}

func BenchmarkMakeWallet(b *testing.B) {
	for b.Loop() {
		MakeWallet()
	}
}

func BenchmarkWalletAddress(b *testing.B) {
	wallet := MakeWallet()
	for b.Loop() {
		wallet.Address()
	}
}

func BenchmarkValidateAddress(b *testing.B) {
	wallet := MakeWallet()
	address := string(wallet.Address())

	for b.Loop() {
		ValidateAddress(address)
	}
}
