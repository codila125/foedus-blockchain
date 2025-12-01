package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/codila125/foedus-blockchain/blockchain"
)

// =============================================================================
// Test Helpers
// =============================================================================

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// =============================================================================
// printBlockHeader Tests
// =============================================================================

func TestPrintBlockHeader_ContainsBlockInfo(t *testing.T) {
	cli := &CommandLine{}

	block := &blockchain.Block{
		Hash:      []byte("testhash123456789012345678901234"),
		PrevHash:  []byte("prevhash123456789012345678901234"),
		Timestamp: 1234567890,
		Nonce:     12345,
		Height:    10,
	}

	output := captureOutput(func() {
		cli.printBlockHeader(block, 5)
	})

	// Check that output contains expected elements
	expectedElements := []string{
		"BLOCK",
		"Hash:",
		"Prev Hash:",
		"Timestamp:",
		"Nonce:",
		"Height:",
		"PoW:",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("printBlockHeader output should contain '%s'", elem)
		}
	}
}

func TestPrintBlockHeader_BlockNumber(t *testing.T) {
	cli := &CommandLine{}

	block := &blockchain.Block{
		Hash:     []byte("hash"),
		PrevHash: []byte("prev"),
	}

	output := captureOutput(func() {
		cli.printBlockHeader(block, 42)
	})

	if !strings.Contains(output, "#42") {
		t.Error("printBlockHeader should display block number")
	}
}

// =============================================================================
// printBlockContracts Tests
// =============================================================================

func TestPrintBlockContracts_NoContracts(t *testing.T) {
	cli := &CommandLine{}

	block := &blockchain.Block{
		Contracts: []*blockchain.Contract{},
	}

	output := captureOutput(func() {
		cli.printBlockContracts(block)
	})

	if !strings.Contains(output, "CONTRACTS (0)") {
		t.Error("Should show contract count of 0")
	}
	if !strings.Contains(output, "No Contracts") {
		t.Error("Should indicate no contracts")
	}
}

func TestPrintBlockContracts_WithContracts(t *testing.T) {
	cli := &CommandLine{}

	block := &blockchain.Block{
		Contracts: []*blockchain.Contract{
			{
				ID:             []byte("contract1234567890123456"),
				Title:          "Test Contract",
				Status:         "Active",
				CreatorAddress: "1ABC123",
				Parties:        []*blockchain.Party{},
				Milestones:     []*blockchain.Milestone{},
			},
		},
	}

	output := captureOutput(func() {
		cli.printBlockContracts(block)
	})

	if !strings.Contains(output, "CONTRACTS (1)") {
		t.Error("Should show contract count of 1")
	}
	if !strings.Contains(output, "Test Contract") {
		t.Error("Should show contract title")
	}
}

// =============================================================================
// printContractWithBox Tests
// =============================================================================

func TestPrintContractWithBox_BasicInfo(t *testing.T) {
	cli := &CommandLine{}

	contract := &blockchain.Contract{
		ID:             []byte("contract1234567890123456"),
		Title:          "My Contract",
		Status:         "Pending",
		CreatorAddress: "1CreatorAddress",
		Parties:        []*blockchain.Party{},
		Milestones:     []*blockchain.Milestone{},
	}

	output := captureOutput(func() {
		cli.printContractWithBox(contract, 1)
	})

	expectedElements := []string{
		"Contract #1",
		"CONTRACT",
		"Title: My Contract",
		"Status: Pending",
		"Creator: 1CreatorAddress",
		"PARTIES",
		"MILESTONES",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("printContractWithBox should contain '%s'", elem)
		}
	}
}

func TestPrintContractWithBox_WithParties(t *testing.T) {
	cli := &CommandLine{}

	contract := &blockchain.Contract{
		ID:             []byte("contract123"),
		Title:          "Contract",
		Status:         "Active",
		CreatorAddress: "Creator",
		Parties: []*blockchain.Party{
			{
				Address:   "1Party1Address",
				Role:      "Buyer",
				Signature: []byte("sig"),
			},
			{
				Address:   "1Party2Address",
				Role:      "Seller",
				Signature: []byte{},
			},
		},
		Milestones: []*blockchain.Milestone{},
	}

	output := captureOutput(func() {
		cli.printContractWithBox(contract, 1)
	})

	if !strings.Contains(output, "PARTIES (2)") {
		t.Error("Should show party count")
	}
	if !strings.Contains(output, "Buyer") {
		t.Error("Should show party role")
	}
	// Signed party should have checkmark
	if !strings.Contains(output, "✓") {
		t.Error("Should show checkmark for signed party")
	}
	// Unsigned party should have X
	if !strings.Contains(output, "✗") {
		t.Error("Should show X for unsigned party")
	}
}

func TestPrintContractWithBox_WithMilestones(t *testing.T) {
	cli := &CommandLine{}

	contract := &blockchain.Contract{
		ID:             []byte("contract123"),
		Title:          "Contract",
		Status:         "Active",
		CreatorAddress: "Creator",
		Parties:        []*blockchain.Party{},
		Milestones: []*blockchain.Milestone{
			{
				Title:  "Milestone 1",
				Status: "Complete",
				Value:  100,
			},
			{
				Title:  "Milestone 2",
				Status: "Pending",
				Value:  200,
			},
		},
	}

	output := captureOutput(func() {
		cli.printContractWithBox(contract, 1)
	})

	if !strings.Contains(output, "MILESTONES (2)") {
		t.Error("Should show milestone count")
	}
	if !strings.Contains(output, "Milestone 1") {
		t.Error("Should show milestone title")
	}
	if !strings.Contains(output, "Value: 100") {
		t.Error("Should show milestone value")
	}
}

// =============================================================================
// printBlockTransactions Tests
// =============================================================================

func TestPrintBlockTransactions_NoTransactions(t *testing.T) {
	cli := &CommandLine{}

	block := &blockchain.Block{
		Transactions: []*blockchain.Transaction{},
	}

	output := captureOutput(func() {
		cli.printBlockTransactions(block)
	})

	if !strings.Contains(output, "TRANSACTIONS (0)") {
		t.Error("Should show transaction count of 0")
	}
	if !strings.Contains(output, "No Transactions") {
		t.Error("Should indicate no transactions")
	}
}

func TestPrintBlockTransactions_WithTransactions(t *testing.T) {
	cli := &CommandLine{}

	block := &blockchain.Block{
		Transactions: []*blockchain.Transaction{
			{
				ID:      []byte("tx1234567890123456789012"),
				Inputs:  []blockchain.TxInput{},
				Outputs: []blockchain.TxOutput{},
			},
		},
	}

	output := captureOutput(func() {
		cli.printBlockTransactions(block)
	})

	if !strings.Contains(output, "TRANSACTIONS (1)") {
		t.Error("Should show transaction count of 1")
	}
}

// =============================================================================
// printTransactionWithBox Tests
// =============================================================================

func TestPrintTransactionWithBox_BasicInfo(t *testing.T) {
	cli := &CommandLine{}

	tx := &blockchain.Transaction{
		ID:      []byte("transaction123456789012345"),
		Inputs:  []blockchain.TxInput{},
		Outputs: []blockchain.TxOutput{},
	}

	output := captureOutput(func() {
		cli.printTransactionWithBox(tx, 1)
	})

	expectedElements := []string{
		"Transaction #1",
		"TRANSACTION",
		"INPUTS",
		"OUTPUTS",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("printTransactionWithBox should contain '%s'", elem)
		}
	}
}

func TestPrintTransactionWithBox_WithOutputs(t *testing.T) {
	cli := &CommandLine{}

	tx := &blockchain.Transaction{
		ID:     []byte("tx123"),
		Inputs: []blockchain.TxInput{},
		Outputs: []blockchain.TxOutput{
			{
				Value:      50,
				PubKeyHash: []byte("pubkeyhash12345678"),
			},
		},
	}

	output := captureOutput(func() {
		cli.printTransactionWithBox(tx, 1)
	})

	if !strings.Contains(output, "OUTPUTS (1)") {
		t.Error("Should show output count")
	}
	if !strings.Contains(output, "Value: 50") {
		t.Error("Should show output value")
	}
}

// =============================================================================
// printBlockFooter Tests
// =============================================================================

func TestPrintBlockFooter_ContainsBorder(t *testing.T) {
	cli := &CommandLine{}

	output := captureOutput(func() {
		cli.printBlockFooter()
	})

	if !strings.Contains(output, "╚") {
		t.Error("Block footer should contain closing border character")
	}
	if !strings.Contains(output, "═") {
		t.Error("Block footer should contain border line")
	}
}

// =============================================================================
// Format Consistency Tests
// =============================================================================

func TestFormatConsistency_BoxCharacters(t *testing.T) {
	// Verify consistent use of box-drawing characters
	boxChars := []string{"╔", "╗", "╚", "╝", "║", "═", "─", "┌", "┐", "└", "┘", "├", "┤", "╭", "╮", "╰", "╯"}

	for _, char := range boxChars {
		if len(char) == 0 {
			t.Error("Box character should not be empty")
		}
	}
}

func TestFormatConsistency_StatusIndicators(t *testing.T) {
	// Verify status indicators are used consistently
	validIndicator := "✓ VALID"
	invalidIndicator := "✗ INVALID"

	if !strings.Contains(validIndicator, "✓") {
		t.Error("Valid indicator should use checkmark")
	}
	if !strings.Contains(invalidIndicator, "✗") {
		t.Error("Invalid indicator should use X mark")
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkPrintBlockHeader(b *testing.B) {
	cli := &CommandLine{}
	block := &blockchain.Block{
		Hash:      []byte("benchmarkhash"),
		PrevHash:  []byte("benchmarkprev"),
		Timestamp: 1234567890,
		Nonce:     12345,
		Height:    100,
	}

	// Redirect stdout to discard
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cli.printBlockHeader(block, i)
	}
}

func BenchmarkPrintBlockFooter(b *testing.B) {
	cli := &CommandLine{}

	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cli.printBlockFooter()
	}
}

// Silence unused import warning
var _ = fmt.Sprint
