package wallet

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// go test -run TestFile
// go test -run TestSave
// go test -run TestLoad
// go test -bench=BenchmarkSave
// go test -bench=BenchmarkLoad

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
			t.Cleanup(func() { _ = os.Remove(walletPath) })

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
	t.Cleanup(func() { _ = os.Remove(walletPath) })

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

// TestSaveFileWithEmptyWallets tests saving empty wallet collection.
func TestSaveFileWithEmptyWallets(t *testing.T) {
	nodeID := "test_empty_wallets"
	walletPath := "./temp/wallets_" + nodeID + ".data"
	t.Cleanup(func() { _ = os.Remove(walletPath) })

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

// TestLoadCorruptedFile tests loading a corrupted wallet file.
func TestLoadCorruptedFile(t *testing.T) {
	nodeID := "test_corrupted"
	walletPath := "./temp/wallets_" + nodeID + ".data"
	t.Cleanup(func() { _ = os.Remove(walletPath) })

	// Ensure temp directory exists
	_ = os.MkdirAll("./temp", 0o755)

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
	t.Cleanup(func() { _ = os.Remove(walletPath) })

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

// Benchmark tests for file operations

func BenchmarkSaveFile(b *testing.B) {
	ws := &Wallets{Wallets: make(map[string]*Wallet)}
	for range 10 {
		ws.AddWallet()
	}
	nodeID := "benchmark_save"

	for b.Loop() {
		_ = ws.SaveFile(nodeID)
	}

	// Cleanup
	b.StopTimer()
	_ = os.Remove("./temp/wallets_" + nodeID + ".data")
}

func BenchmarkLoadFile(b *testing.B) {
	ws := &Wallets{Wallets: make(map[string]*Wallet)}
	for range 10 {
		ws.AddWallet()
	}
	nodeID := "benchmark_load"
	_ = ws.SaveFile(nodeID)

	for b.Loop() {
		loadedWs := &Wallets{}
		_ = loadedWs.LoadFile(nodeID)
	}

	// Cleanup
	b.StopTimer()
	_ = os.Remove("./temp/wallets_" + nodeID + ".data")
}
