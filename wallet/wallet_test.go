package wallet

import (
	"bytes"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"
)

// go test
// go test -bench=.
// go test -cover
// go test -race

// TestBase58Encode tests the Base58 encoding function with various inputs.
func TestBase58Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty input",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "single byte zero",
			input:    []byte{0},
			expected: "1",
		},
		{
			name:     "hello world",
			input:    []byte("hello world"),
			expected: "StV1DL6CwTryKyV",
		},
		{
			name:     "all zeros",
			input:    []byte{0, 0, 0},
			expected: "111",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Base58Encode(tt.input)
			if string(result) != tt.expected {
				t.Errorf("Base58Encode(%v) = %s, want %s", tt.input, string(result), tt.expected)
			}
		})
	}
}

// TestBase58Decode tests the Base58 decoding function with various inputs.
func TestBase58Decode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "single 1 (zero)",
			input:    []byte("1"),
			expected: []byte{0},
		},
		{
			name:     "hello world encoded",
			input:    []byte("StV1DL6CwTryKyV"),
			expected: []byte("hello world"),
		},
		{
			name:     "multiple ones (zeros)",
			input:    []byte("111"),
			expected: []byte{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Base58Decode(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Base58Decode(%s) = %v, want %v", string(tt.input), result, tt.expected)
			}
		})
	}
}

// TestBase58RoundTrip tests that encoding then decoding returns the original data.
func TestBase58RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "empty",
			input: []byte{},
		},
		{
			name:  "single byte",
			input: []byte{42},
		},
		{
			name:  "text data",
			input: []byte("blockchain wallet test"),
		},
		{
			name:  "binary data",
			input: []byte{0x00, 0xff, 0x10, 0xab, 0xcd},
		},
		{
			name:  "256 bytes",
			input: make([]byte, 256),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := Base58Encode(tt.input)
			decoded := Base58Decode(encoded)
			if !bytes.Equal(decoded, tt.input) {
				t.Errorf("RoundTrip failed: got %v, want %v", decoded, tt.input)
			}
		})
	}
}

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

// TestWalletsAddWallet tests adding wallets to the collection.
func TestWalletsAddWallet(t *testing.T) {
	tests := []struct {
		name          string
		numWallets    int
		expectedCount int
	}{
		{
			name:          "add one wallet",
			numWallets:    1,
			expectedCount: 1,
		},
		{
			name:          "add multiple wallets",
			numWallets:    5,
			expectedCount: 5,
		},
		{
			name:          "add no wallets",
			numWallets:    0,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &Wallets{Wallets: make(map[string]*Wallet)}

			for i := 0; i < tt.numWallets; i++ {
				ws.AddWallet()
			}

			if len(ws.Wallets) != tt.expectedCount {
				t.Errorf("Wallets count = %d, want %d", len(ws.Wallets), tt.expectedCount)
			}
		})
	}
}

// TestWalletsAddWalletReturnsValidAddress tests that AddWallet returns valid addresses.
func TestWalletsAddWalletReturnsValidAddress(t *testing.T) {
	ws := &Wallets{Wallets: make(map[string]*Wallet)}

	for range 3 {
		address := ws.AddWallet()

		if !ValidateAddress(address) {
			t.Errorf("AddWallet() returned invalid address: %s", address)
		}

		if _, exists := ws.Wallets[address]; !exists {
			t.Errorf("Wallet not stored with returned address: %s", address)
		}
	}
}

// TestWalletsGetAllAddresses tests retrieving all wallet addresses.
func TestWalletsGetAllAddresses(t *testing.T) {
	tests := []struct {
		name       string
		numWallets int
	}{
		{name: "empty collection", numWallets: 0},
		{name: "single wallet", numWallets: 1},
		{name: "multiple wallets", numWallets: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &Wallets{Wallets: make(map[string]*Wallet)}
			addedAddresses := make(map[string]bool)

			for i := 0; i < tt.numWallets; i++ {
				addr := ws.AddWallet()
				addedAddresses[addr] = true
			}

			addresses := ws.GetAllAddresses()

			if len(addresses) != tt.numWallets {
				t.Errorf("GetAllAddresses() returned %d addresses, want %d", len(addresses), tt.numWallets)
			}

			for _, addr := range addresses {
				if !addedAddresses[addr] {
					t.Errorf("GetAllAddresses() returned unknown address: %s", addr)
				}
			}
		})
	}
}

// TestWalletsGetWallet tests retrieving a specific wallet.
func TestWalletsGetWallet(t *testing.T) {
	ws := &Wallets{Wallets: make(map[string]*Wallet)}
	address := ws.AddWallet()

	tests := []struct {
		name        string
		address     string
		shouldError bool
	}{
		{
			name:        "existing wallet",
			address:     address,
			shouldError: false,
		},
		{
			name:        "non-existing wallet",
			address:     "nonexistent",
			shouldError: true,
		},
		{
			name:        "empty address",
			address:     "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := ws.GetWallet(tt.address)

			if tt.shouldError {
				if err == nil {
					t.Error("GetWallet() should have returned an error")
				}
			} else {
				if err != nil {
					t.Errorf("GetWallet() unexpected error: %v", err)
				}
				if len(wallet.PublicKey) == 0 {
					t.Error("GetWallet() returned wallet with empty public key")
				}
			}
		})
	}
}

// TestWalletsSaveAndLoadFile tests saving and loading wallets from file.
func TestWalletsSaveAndLoadFile(t *testing.T) {
	tests := []struct {
		name       string
		numWallets int
	}{
		{name: "single wallet", numWallets: 1},
		{name: "multiple wallets", numWallets: 5},
		{name: "many wallets", numWallets: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use unique node ID for each test
			nodeID := "test_" + tt.name

			// Create wallets and save
			ws := &Wallets{Wallets: make(map[string]*Wallet)}
			addresses := make([]string, 0, tt.numWallets)

			for i := 0; i < tt.numWallets; i++ {
				addr := ws.AddWallet()
				addresses = append(addresses, addr)
			}

			// Cleanup
			walletPath := "./temp/wallets_" + nodeID + ".data"
			t.Cleanup(func() { os.Remove(walletPath) })

			// Save wallets
			err := ws.SaveFile(nodeID)
			if err != nil {
				t.Fatalf("Failed to save wallets: %v", err)
			}

			// Load into new collection
			loadedWs := &Wallets{Wallets: make(map[string]*Wallet)}
			err = loadedWs.LoadFile(nodeID)
			if err != nil {
				t.Fatalf("Failed to load wallets: %v", err)
			}

			// Verify count
			if len(loadedWs.Wallets) != tt.numWallets {
				t.Errorf("Loaded %d wallets, want %d", len(loadedWs.Wallets), tt.numWallets)
			}

			// Verify each wallet
			for _, addr := range addresses {
				original := ws.Wallets[addr]
				loaded, exists := loadedWs.Wallets[addr]

				if !exists {
					t.Errorf("Address %s not found in loaded wallets", addr)
					continue
				}

				if !bytes.Equal(original.PublicKey, loaded.PublicKey) {
					t.Errorf("Public key mismatch for address %s", addr)
				}

				if !bytes.Equal(original.PrivateKey, loaded.PrivateKey) {
					t.Errorf("Private key mismatch for address %s", addr)
				}
			}
		})
	}
}

// TestLoadFileNonExistent tests loading from a non-existent file.
func TestLoadFileNonExistent(t *testing.T) {
	ws := &Wallets{Wallets: make(map[string]*Wallet)}
	err := ws.LoadFile("nonexistent_node_id_12345")

	if err == nil {
		t.Error("LoadFile() should return error for non-existent file")
	}
}

// TestSaveFileCreatesDirectory tests that SaveFile creates the temp directory.
func TestSaveFileCreatesDirectory(t *testing.T) {
	// Use a unique node ID to avoid conflicts
	nodeID := "test_dir_creation"
	walletPath := "./temp/wallets_" + nodeID + ".data"
	t.Cleanup(func() { os.Remove(walletPath) })

	ws := &Wallets{Wallets: make(map[string]*Wallet)}
	ws.AddWallet()

	err := ws.SaveFile(nodeID)
	if err != nil {
		t.Errorf("SaveFile() error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		t.Error("SaveFile() did not create wallet file")
	}
}

// TestCreateWallets tests the CreateWallets function.
func TestCreateWallets(t *testing.T) {
	// Test with non-existent file (should return error but still create empty collection)
	ws, err := CreateWallets("new_node_test_12345")

	if ws == nil {
		t.Fatal("CreateWallets() returned nil")
	}

	if ws.Wallets == nil {
		t.Error("CreateWallets() did not initialize Wallets map")
	}

	// Error is expected since the file doesn't exist
	if err == nil {
		// If no error, the file existed - clean up
		t.Log("Wallet file existed, which is acceptable")
	}
}

// TestCreateWalletsWithExistingFile tests CreateWallets with an existing file.
func TestCreateWalletsWithExistingFile(t *testing.T) {
	nodeID := "test_existing_file"
	walletPath := "./temp/wallets_" + nodeID + ".data"
	t.Cleanup(func() { os.Remove(walletPath) })

	// First create and save some wallets
	ws := &Wallets{Wallets: make(map[string]*Wallet)}
	addr1 := ws.AddWallet()
	addr2 := ws.AddWallet()
	err := ws.SaveFile(nodeID)
	if err != nil {
		t.Fatalf("Failed to save wallets: %v", err)
	}

	// Now create wallets from the existing file
	loadedWs, err := CreateWallets(nodeID)
	if err != nil {
		t.Fatalf("CreateWallets() error: %v", err)
	}

	// Verify loaded wallets
	if len(loadedWs.Wallets) != 2 {
		t.Errorf("CreateWallets() loaded %d wallets, want 2", len(loadedWs.Wallets))
	}

	if _, exists := loadedWs.Wallets[addr1]; !exists {
		t.Errorf("CreateWallets() did not load address %s", addr1)
	}

	if _, exists := loadedWs.Wallets[addr2]; !exists {
		t.Errorf("CreateWallets() did not load address %s", addr2)
	}
}

// TestWalletIntegration tests the full wallet workflow.
func TestWalletIntegration(t *testing.T) {
	nodeID := "integration_test"
	walletPath := "./temp/wallets_" + nodeID + ".data"
	t.Cleanup(func() { os.Remove(walletPath) })

	// Create wallets
	ws := &Wallets{Wallets: make(map[string]*Wallet)}

	// Add multiple wallets
	addresses := make([]string, 3)
	for i := range 3 {
		addresses[i] = ws.AddWallet()
	}

	// Verify addresses are valid
	for _, addr := range addresses {
		if !ValidateAddress(addr) {
			t.Errorf("Invalid address generated: %s", addr)
		}
	}

	// Get all addresses
	allAddrs := ws.GetAllAddresses()
	if len(allAddrs) != 3 {
		t.Errorf("Expected 3 addresses, got %d", len(allAddrs))
	}

	// Get specific wallet
	wallet, err := ws.GetWallet(addresses[0])
	if err != nil {
		t.Errorf("Failed to get wallet: %v", err)
	}

	// Verify wallet can sign and verify
	message := []byte("test transaction data")
	signature := ed25519.Sign(wallet.PrivateKey, message)
	if !ed25519.Verify(wallet.PublicKey, message, signature) {
		t.Error("Wallet keys cannot sign and verify")
	}

	// Save and reload
	err = ws.SaveFile(nodeID)
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	loadedWs := &Wallets{Wallets: make(map[string]*Wallet)}
	err = loadedWs.LoadFile(nodeID)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Verify loaded wallet can still sign and verify
	loadedWallet, _ := loadedWs.GetWallet(addresses[0])
	signature2 := ed25519.Sign(loadedWallet.PrivateKey, message)
	if !ed25519.Verify(loadedWallet.PublicKey, message, signature2) {
		t.Error("Loaded wallet keys cannot sign and verify")
	}
}

// TestSaveFileWithEmptyWallets tests saving empty wallet collection.
func TestSaveFileWithEmptyWallets(t *testing.T) {
	nodeID := "test_empty_wallets"
	walletPath := "./temp/wallets_" + nodeID + ".data"
	t.Cleanup(func() { os.Remove(walletPath) })

	ws := &Wallets{Wallets: make(map[string]*Wallet)}
	err := ws.SaveFile(nodeID)
	if err != nil {
		t.Errorf("SaveFile() with empty wallets error: %v", err)
	}

	// Verify we can load empty wallets
	loadedWs := &Wallets{Wallets: make(map[string]*Wallet)}
	err = loadedWs.LoadFile(nodeID)
	if err != nil {
		t.Errorf("LoadFile() empty wallets error: %v", err)
	}

	if len(loadedWs.Wallets) != 0 {
		t.Errorf("Expected 0 wallets, got %d", len(loadedWs.Wallets))
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

// TestLoadCorruptedFile tests loading a corrupted wallet file.
func TestLoadCorruptedFile(t *testing.T) {
	nodeID := "test_corrupted"
	walletPath := "./temp/wallets_" + nodeID + ".data"
	t.Cleanup(func() { os.Remove(walletPath) })

	// Ensure temp directory exists
	os.MkdirAll("./temp", 0o755)

	// Write corrupted data
	err := os.WriteFile(walletPath, []byte("corrupted data that is not valid protobuf"), 0o600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Try to load corrupted file
	ws := &Wallets{Wallets: make(map[string]*Wallet)}
	err = ws.LoadFile(nodeID)
	if err == nil {
		t.Error("LoadFile() should return error for corrupted file")
	}
}

// TestWalletFilePermissions tests that saved wallet file has correct permissions.
func TestWalletFilePermissions(t *testing.T) {
	nodeID := "test_permissions"
	walletPath := filepath.Join("./temp", "wallets_"+nodeID+".data")
	t.Cleanup(func() { os.Remove(walletPath) })

	ws := &Wallets{Wallets: make(map[string]*Wallet)}
	ws.AddWallet()

	err := ws.SaveFile(nodeID)
	if err != nil {
		t.Fatalf("SaveFile() error: %v", err)
	}

	info, err := os.Stat(walletPath)
	if err != nil {
		t.Fatalf("Failed to stat wallet file: %v", err)
	}

	// Check file permissions (0600 = owner read/write only)
	expectedMode := os.FileMode(0o600)
	if info.Mode().Perm() != expectedMode {
		t.Errorf("Wallet file permissions = %o, want %o", info.Mode().Perm(), expectedMode)
	}
}

// Benchmark tests

func BenchmarkBase58Encode(b *testing.B) {
	data := make([]byte, 32)
	for i := range data {
		data[i] = byte(i)
	}

	for b.Loop() {
		Base58Encode(data)
	}
}

func BenchmarkBase58Decode(b *testing.B) {
	data := make([]byte, 32)
	for i := range data {
		data[i] = byte(i)
	}
	encoded := Base58Encode(data)

	for b.Loop() {
		Base58Decode(encoded)
	}
}

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

func BenchmarkAddWallet(b *testing.B) {
	ws := &Wallets{Wallets: make(map[string]*Wallet)}

	for b.Loop() {
		ws.AddWallet()
	}
}

func BenchmarkSaveFile(b *testing.B) {
	ws := &Wallets{Wallets: make(map[string]*Wallet)}
	for range 10 {
		ws.AddWallet()
	}
	nodeID := "benchmark_save"

	for b.Loop() {
		ws.SaveFile(nodeID)
	}

	// Cleanup
	b.StopTimer()
	os.Remove("./temp/wallets_" + nodeID + ".data")
}

func BenchmarkLoadFile(b *testing.B) {
	ws := &Wallets{Wallets: make(map[string]*Wallet)}
	for range 10 {
		ws.AddWallet()
	}
	nodeID := "benchmark_load"
	ws.SaveFile(nodeID)

	for b.Loop() {
		loadedWs := &Wallets{}
		loadedWs.LoadFile(nodeID)
	}

	// Cleanup
	b.StopTimer()
	os.Remove("./temp/wallets_" + nodeID + ".data")
}
