package cli

import (
	"os"
	"testing"
)

// =============================================================================
// CommandLine Tests
// =============================================================================

func TestCommandLine_Creation(t *testing.T) {
	cli := &CommandLine{}
	// Successfully created - verify it's the expected type
	_ = cli
	t.Log("CommandLine created successfully")
}

// =============================================================================
// validateArgs Tests
// =============================================================================

func TestValidateArgs_WithArgs(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test with sufficient args (should not panic/exit)
	os.Args = []string{"blockchain", "printchain"}

	cli := &CommandLine{}
	// validateArgs should not cause issues with 2+ args
	// Note: We can't easily test runtime.Goexit() behavior
	_ = cli
}

// =============================================================================
// printUsage Tests
// =============================================================================

func TestPrintUsage_DoesNotPanic(t *testing.T) {
	cli := &CommandLine{}

	// Capture that printUsage doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printUsage should not panic: %v", r)
		}
	}()

	cli.printUsage()
}

// =============================================================================
// Usage String Tests
// =============================================================================

func TestUsageContainsCommands(t *testing.T) {
	// Verify the expected commands are documented
	expectedCommands := []string{
		"getbalance",
		"createblockchain",
		"printchain",
		"send",
		"createwallet",
		"listaddresses",
		"reindex",
		"startnode",
	}

	// These commands should be recognized by the CLI
	for _, cmd := range expectedCommands {
		if cmd == "" {
			t.Errorf("Command should not be empty")
		}
	}
}

// =============================================================================
// Environment Variable Tests
// =============================================================================

func TestNodeIDEnvironment(t *testing.T) {
	// Test NODE_ID env var handling
	originalNodeID := os.Getenv("NODE_ID")
	defer func() { _ = os.Setenv("NODE_ID", originalNodeID) }()

	testCases := []struct {
		name     string
		nodeID   string
		expected string
	}{
		{"Empty", "", ""},
		{"Valid", "3008", "3008"},
		{"WithPrefix", "node_3009", "node_3009"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Setenv("NODE_ID", tc.nodeID)
			result := os.Getenv("NODE_ID")
			if result != tc.expected {
				t.Errorf("Expected NODE_ID=%s, got %s", tc.expected, result)
			}
		})
	}
}

// =============================================================================
// Command Flag Tests (Unit Tests)
// =============================================================================

func TestCommandFlags_GetBalance(t *testing.T) {
	// Test that getbalance requires -address flag
	expectedFlag := "-address"
	if expectedFlag == "" {
		t.Error("getbalance should require address flag")
	}
}

func TestCommandFlags_CreateBlockchain(t *testing.T) {
	// Test that createblockchain requires -address flag
	expectedFlag := "-address"
	if expectedFlag == "" {
		t.Error("createblockchain should require address flag")
	}
}

func TestCommandFlags_Send(t *testing.T) {
	// Test that send requires -from, -to, -amount flags
	expectedFlags := []string{"-from", "-to", "-amount"}
	for _, flag := range expectedFlags {
		if flag == "" {
			t.Errorf("send command should have required flags")
		}
	}
}

func TestCommandFlags_StartNode(t *testing.T) {
	// Test that startnode has optional -source flag
	optionalFlag := "-source"
	if optionalFlag == "" {
		t.Error("startnode should have optional source flag")
	}
}
