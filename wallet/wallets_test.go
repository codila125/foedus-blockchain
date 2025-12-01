package wallet

import (
	"crypto/ed25519"
	"os"
	"testing"
)

// go test -run TestWallets
// go test -run TestCreate
// go test -bench=BenchmarkAddWallet

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
	t.Cleanup(func() { _ = os.Remove(walletPath) })

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
	t.Cleanup(func() { _ = os.Remove(walletPath) })

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

// Benchmark tests for wallets collection

func BenchmarkAddWallet(b *testing.B) {
	ws := &Wallets{Wallets: make(map[string]*Wallet)}

	for b.Loop() {
		ws.AddWallet()
	}
}
