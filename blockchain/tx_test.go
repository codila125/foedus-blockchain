package blockchain

import (
	"bytes"
	"testing"

	"github.com/codila125/foedus-blockchain/wallet"
)

// =============================================================================
// TxOutput Tests
// =============================================================================

func TestTxOutput_Fields(t *testing.T) {
	t.Parallel()

	output := TxOutput{
		Value:      100,
		PubKeyHash: []byte("test_hash"),
	}

	if output.Value != 100 {
		t.Errorf("Expected value 100, got %d", output.Value)
	}

	if !bytes.Equal(output.PubKeyHash, []byte("test_hash")) {
		t.Error("PubKeyHash mismatch")
	}
}

func TestTxOutput_ZeroValue(t *testing.T) {
	t.Parallel()

	output := TxOutput{Value: 0, PubKeyHash: []byte("hash")}

	if output.Value != 0 {
		t.Error("TxOutput should accept zero value")
	}
	// Zero value is valid
	if err := output.Validate(); err != nil {
		t.Errorf("Zero value should be valid: %v", err)
	}
}

func TestTxOutput_NegativeValue(t *testing.T) {
	t.Parallel()

	// Struct accepts negative, but validation should fail
	output := TxOutput{Value: -100, PubKeyHash: []byte("hash")}

	if err := output.Validate(); err != ErrNegativeValue {
		t.Errorf("Expected ErrNegativeValue, got: %v", err)
	}
}

func TestTxOutput_EmptyPubKeyHash(t *testing.T) {
	t.Parallel()

	output := TxOutput{Value: 100, PubKeyHash: []byte{}}

	if err := output.Validate(); err != ErrEmptyAddress {
		t.Errorf("Expected ErrEmptyAddress for empty PubKeyHash, got: %v", err)
	}
}

func TestTxOutput_NilPubKeyHash(t *testing.T) {
	t.Parallel()

	output := TxOutput{Value: 100, PubKeyHash: nil}

	if err := output.Validate(); err != ErrEmptyAddress {
		t.Errorf("Expected ErrEmptyAddress for nil PubKeyHash, got: %v", err)
	}
}

// =============================================================================
// NewTxOutput Tests
// =============================================================================

func TestNewTxOutput(t *testing.T) {
	t.Parallel()

	// Create a valid wallet to get a valid address
	w := wallet.MakeWallet()
	address := string(w.Address())

	output := NewTxOutput(100, address)

	if output == nil {
		t.Fatal("NewTxOutput should return non-nil output")
	}

	if output.Value != 100 {
		t.Errorf("Expected value 100, got %d", output.Value)
	}

	if len(output.PubKeyHash) == 0 {
		t.Error("PubKeyHash should not be empty after Lock")
	}
}

func TestNewTxOutput_ZeroValue(t *testing.T) {
	t.Parallel()

	w := wallet.MakeWallet()
	address := string(w.Address())

	output := NewTxOutput(0, address)

	if output.Value != 0 {
		t.Errorf("Expected value 0, got %d", output.Value)
	}
}

func TestNewTxOutput_LargeValue(t *testing.T) {
	t.Parallel()

	w := wallet.MakeWallet()
	address := string(w.Address())
	largeValue := 2147483647 // Max int32

	output := NewTxOutput(largeValue, address)

	if output.Value != largeValue {
		t.Errorf("Expected value %d, got %d", largeValue, output.Value)
	}
}

// =============================================================================
// Lock Tests
// =============================================================================

func TestTxOutput_Lock(t *testing.T) {
	t.Parallel()

	w := wallet.MakeWallet()
	address := w.Address()

	output := &TxOutput{Value: 100, PubKeyHash: nil}
	output.Lock(address)

	if len(output.PubKeyHash) == 0 {
		t.Error("PubKeyHash should be set after Lock")
	}

	// PubKeyHash should match the wallet's public key hash
	expectedHash := wallet.PublicKeyHash(w.PublicKey)
	if !bytes.Equal(output.PubKeyHash, expectedHash) {
		t.Error("PubKeyHash should match wallet's public key hash")
	}
}

func TestTxOutput_Lock_ExtractsPubKeyHashFromAddress(t *testing.T) {
	t.Parallel()

	w := wallet.MakeWallet()
	address := w.Address()

	// Address format: version(1) + pubKeyHash + checksum(4)
	// Lock should extract just the pubKeyHash

	output := &TxOutput{Value: 50}
	output.Lock(address)

	// The extracted pubKeyHash should be 32 bytes (SHA-256 of public key)
	if len(output.PubKeyHash) != 32 {
		t.Errorf("Expected PubKeyHash length 32, got %d", len(output.PubKeyHash))
	}
}

func TestTxOutput_Lock_DifferentAddresses(t *testing.T) {
	t.Parallel()

	w1 := wallet.MakeWallet()
	w2 := wallet.MakeWallet()

	output1 := &TxOutput{Value: 100}
	output2 := &TxOutput{Value: 100}

	output1.Lock(w1.Address())
	output2.Lock(w2.Address())

	if bytes.Equal(output1.PubKeyHash, output2.PubKeyHash) {
		t.Error("Different addresses should produce different PubKeyHash")
	}
}

// =============================================================================
// IsLockedWithKey Tests
// =============================================================================

func TestTxOutput_IsLockedWithKey_True(t *testing.T) {
	t.Parallel()

	w := wallet.MakeWallet()
	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)

	output := NewTxOutput(100, string(w.Address()))

	if !output.IsLockedWithKey(pubKeyHash) {
		t.Error("Output should be locked with the correct public key hash")
	}
}

func TestTxOutput_IsLockedWithKey_False(t *testing.T) {
	t.Parallel()

	w1 := wallet.MakeWallet()
	w2 := wallet.MakeWallet()

	output := NewTxOutput(100, string(w1.Address()))

	w2PubKeyHash := wallet.PublicKeyHash(w2.PublicKey)

	if output.IsLockedWithKey(w2PubKeyHash) {
		t.Error("Output should not be locked with different public key hash")
	}
}

func TestTxOutput_IsLockedWithKey_EmptyHash(t *testing.T) {
	t.Parallel()

	w := wallet.MakeWallet()
	output := NewTxOutput(100, string(w.Address()))

	if output.IsLockedWithKey([]byte{}) {
		t.Error("Output should not match empty key hash")
	}
}

func TestTxOutput_IsLockedWithKey_NilHash(t *testing.T) {
	t.Parallel()

	w := wallet.MakeWallet()
	output := NewTxOutput(100, string(w.Address()))

	if output.IsLockedWithKey(nil) {
		t.Error("Output should not match nil key hash")
	}
}

// =============================================================================
// TxInput Tests
// =============================================================================

func TestTxInput_Fields(t *testing.T) {
	t.Parallel()

	input := TxInput{
		ID:        []byte("tx_id"),
		Out:       0,
		Signature: []byte("signature"),
		PubKey:    []byte("public_key"),
	}

	if !bytes.Equal(input.ID, []byte("tx_id")) {
		t.Error("ID mismatch")
	}

	if input.Out != 0 {
		t.Errorf("Expected Out 0, got %d", input.Out)
	}

	if !bytes.Equal(input.Signature, []byte("signature")) {
		t.Error("Signature mismatch")
	}

	if !bytes.Equal(input.PubKey, []byte("public_key")) {
		t.Error("PubKey mismatch")
	}
}

func TestTxInput_NegativeOut(t *testing.T) {
	t.Parallel()

	// Coinbase transaction has Out = -1 and empty ID
	input := TxInput{
		ID:        []byte{},
		Out:       -1,
		Signature: nil,
		PubKey:    []byte("data"),
	}

	if input.Out != -1 {
		t.Errorf("Expected Out -1, got %d", input.Out)
	}

	// Coinbase input should be valid
	if err := input.Validate(); err != nil {
		t.Errorf("Coinbase input should be valid: %v", err)
	}
}

func TestTxInput_InvalidNegativeOut(t *testing.T) {
	t.Parallel()

	// Non-coinbase with negative Out should be invalid
	input := TxInput{
		ID:        []byte("some_tx_id"),
		Out:       -1,
		Signature: nil,
		PubKey:    nil,
	}

	if err := input.Validate(); err == nil {
		t.Error("Non-coinbase input with negative Out should be invalid")
	}
}

func TestTxInput_EmptyIDWithPositiveOut(t *testing.T) {
	t.Parallel()

	// Empty ID with positive Out is invalid (not a valid coinbase)
	input := TxInput{
		ID:        []byte{},
		Out:       0,
		Signature: []byte{},
		PubKey:    []byte{},
	}

	if err := input.Validate(); err == nil {
		t.Error("Empty ID with non-negative Out should be invalid")
	}
}

func TestTxInput_ValidInput(t *testing.T) {
	t.Parallel()

	input := TxInput{
		ID:        []byte("valid_tx_id"),
		Out:       0,
		Signature: []byte("sig"),
		PubKey:    []byte("pk"),
	}

	if err := input.Validate(); err != nil {
		t.Errorf("Valid input should pass validation: %v", err)
	}
}

func TestTxInput_NilFields(t *testing.T) {
	t.Parallel()

	input := TxInput{
		ID:        nil,
		Out:       0,
		Signature: nil,
		PubKey:    nil,
	}

	// Nil ID with non-negative Out is invalid
	if err := input.Validate(); err == nil {
		t.Error("Nil ID with non-negative Out should be invalid")
	}
}

// =============================================================================
// TxOutputs Tests
// =============================================================================

func TestTxOutputs_Fields(t *testing.T) {
	t.Parallel()

	outputs := TxOutputs{
		Outputs: []TxOutput{
			{Value: 100, PubKeyHash: []byte("hash1")},
			{Value: 200, PubKeyHash: []byte("hash2")},
		},
	}

	if len(outputs.Outputs) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(outputs.Outputs))
	}

	if outputs.Outputs[0].Value != 100 {
		t.Error("First output value mismatch")
	}

	if outputs.Outputs[1].Value != 200 {
		t.Error("Second output value mismatch")
	}
}

func TestTxOutputs_Empty(t *testing.T) {
	t.Parallel()

	outputs := TxOutputs{Outputs: []TxOutput{}}

	if len(outputs.Outputs) != 0 {
		t.Error("Empty outputs should have length 0")
	}
}

func TestTxOutputs_Nil(t *testing.T) {
	t.Parallel()

	outputs := TxOutputs{Outputs: nil}

	if outputs.Outputs != nil {
		t.Error("Nil outputs should remain nil")
	}
}

func TestTxOutputs_Append(t *testing.T) {
	t.Parallel()

	outputs := TxOutputs{Outputs: []TxOutput{}}

	outputs.Outputs = append(outputs.Outputs, TxOutput{Value: 50, PubKeyHash: []byte("h1")})
	outputs.Outputs = append(outputs.Outputs, TxOutput{Value: 75, PubKeyHash: []byte("h2")})

	if len(outputs.Outputs) != 2 {
		t.Errorf("Expected 2 outputs after append, got %d", len(outputs.Outputs))
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestTxOutput_VeryLargePubKeyHash(t *testing.T) {
	t.Parallel()

	largeHash := make([]byte, 1024)
	for i := range largeHash {
		largeHash[i] = byte(i % 256)
	}

	output := TxOutput{Value: 100, PubKeyHash: largeHash}

	if !bytes.Equal(output.PubKeyHash, largeHash) {
		t.Error("Large PubKeyHash should be preserved")
	}

	if output.Value != 100 {
		t.Error("Value should be preserved")
	}
}

func TestTxInput_VeryLargeSignature(t *testing.T) {
	t.Parallel()

	largeSig := make([]byte, 512)
	for i := range largeSig {
		largeSig[i] = byte(i % 256)
	}

	input := TxInput{
		ID:        []byte("id"),
		Out:       0,
		Signature: largeSig,
		PubKey:    []byte("pk"),
	}

	if !bytes.Equal(input.Signature, largeSig) {
		t.Error("Large Signature should be preserved")
	}

	if !bytes.Equal(input.ID, []byte("id")) {
		t.Error("ID should be preserved")
	}

	if input.Out != 0 {
		t.Error("Out should be preserved")
	}

	if !bytes.Equal(input.PubKey, []byte("pk")) {
		t.Error("PubKey should be preserved")
	}
}

func TestTxOutputs_ManyOutputs(t *testing.T) {
	t.Parallel()

	count := 1000
	outputs := TxOutputs{Outputs: make([]TxOutput, count)}

	for i := 0; i < count; i++ {
		outputs.Outputs[i] = TxOutput{
			Value:      i,
			PubKeyHash: []byte{byte(i % 256)},
		}
	}

	if len(outputs.Outputs) != count {
		t.Errorf("Expected %d outputs, got %d", count, len(outputs.Outputs))
	}

	// Verify first and last values
	if outputs.Outputs[0].Value != 0 {
		t.Error("First output value mismatch")
	}

	if outputs.Outputs[count-1].Value != count-1 {
		t.Error("Last output value mismatch")
	}

	// Verify first and last PubKeyHash
	if !bytes.Equal(outputs.Outputs[0].PubKeyHash, []byte{byte(0)}) {
		t.Error("First output PubKeyHash mismatch")
	}

	if !bytes.Equal(outputs.Outputs[count-1].PubKeyHash, []byte{byte((count - 1) % 256)}) {
		t.Error("Last output PubKeyHash mismatch")
	}
}

func TestTxOutput_MaxIntValue(t *testing.T) {
	t.Parallel()

	maxInt := int(^uint(0) >> 1) // Max int value

	output := TxOutput{Value: maxInt, PubKeyHash: []byte("hash")}

	if output.Value != maxInt {
		t.Errorf("Expected max int value %d, got %d", maxInt, output.Value)
	}

	if !bytes.Equal(output.PubKeyHash, []byte("hash")) {
		t.Error("PubKeyHash should be preserved")
	}
}
