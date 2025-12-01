package blockchain

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"testing"
)

// =============================================================================
// CoinbaseTx Tests
// =============================================================================

func TestCoinbaseTx_Basic(t *testing.T) {
	t.Parallel()
	address := CreateTestWallet()
	data := "test_coinbase_data"

	tx := CoinbaseTx(address, data)

	if tx == nil {
		t.Fatal("CoinbaseTx should return non-nil transaction")
	}

	if len(tx.ID) == 0 {
		t.Error("Transaction ID should not be empty")
	}

	if len(tx.Inputs) != 1 {
		t.Errorf("Expected 1 input, got %d", len(tx.Inputs))
	}

	if len(tx.Outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(tx.Outputs))
	}
}

func TestCoinbaseTx_InputStructure(t *testing.T) {
	t.Parallel()
	tx := CreateTestCoinbaseTx("data")

	input := tx.Inputs[0]

	// Coinbase input has no reference transaction ID
	if len(input.ID) != 0 {
		t.Error("Coinbase input ID should be empty")
	}

	// Coinbase input has Out index of -1
	if input.Out != -1 {
		t.Errorf("Expected Out to be -1, got %d", input.Out)
	}

	// Coinbase input has no signature
	if input.Signature != nil {
		t.Error("Coinbase input should have no signature")
	}

	// Coinbase input has the data as PubKey
	if !bytes.Equal(input.PubKey, []byte("data")) {
		t.Errorf("Expected PubKey to be 'data', got '%s'", string(input.PubKey))
	}
}

func TestCoinbaseTx_OutputValue(t *testing.T) {
	t.Parallel()
	tx := CreateTestCoinbaseTx("data")

	if tx.Outputs[0].Value != 20 {
		t.Errorf("Expected coinbase reward of 20, got %d", tx.Outputs[0].Value)
	}
}

func TestCoinbaseTx_EmptyData(t *testing.T) {
	t.Parallel()
	tx := CoinbaseTx(CreateTestWallet(), "")

	// When data is empty, random data should be generated
	if len(tx.Inputs[0].PubKey) == 0 {
		t.Error("Expected random data when data is empty")
	}

	// Random data should be hex-encoded (40 chars for 20 bytes)
	if len(tx.Inputs[0].PubKey) != 40 {
		t.Logf("Expected 40-char random hex data, got %d chars", len(tx.Inputs[0].PubKey))
	}
}

func TestCoinbaseTx_DifferentDataProducesDifferentIDs(t *testing.T) {
	t.Parallel()
	address := CreateTestWallet()
	tx1 := CoinbaseTx(address, "data1")
	tx2 := CoinbaseTx(address, "data2")

	if bytes.Equal(tx1.ID, tx2.ID) {
		t.Error("Different coinbase data should produce different transaction IDs")
	}
}

func TestCoinbaseTx_DifferentAddressesProduceDifferentOutputs(t *testing.T) {
	t.Parallel()
	address1 := CreateTestWallet()
	address2 := CreateTestWallet()
	tx1 := CoinbaseTx(address1, "data")
	tx2 := CoinbaseTx(address2, "data")

	if bytes.Equal(tx1.Outputs[0].PubKeyHash, tx2.Outputs[0].PubKeyHash) {
		t.Error("Different addresses should produce different output PubKeyHash")
	}
}

// =============================================================================
// IsCoinbaseTx Tests
// =============================================================================

func TestIsCoinbaseTx_True(t *testing.T) {
	t.Parallel()
	tx := CoinbaseTx(CreateTestWallet(), "data")

	if !tx.IsCoinbaseTx() {
		t.Error("CoinbaseTx should be identified as coinbase transaction")
	}
}

func TestIsCoinbaseTx_False_MultipleInputs(t *testing.T) {
	t.Parallel()
	tx := &Transaction{
		ID: []byte("test"),
		Inputs: []TxInput{
			{ID: []byte{}, Out: -1, Signature: nil, PubKey: []byte("data")},
			{ID: []byte{}, Out: -1, Signature: nil, PubKey: []byte("data2")},
		},
		Outputs: []TxOutput{},
	}

	if tx.IsCoinbaseTx() {
		t.Error("Transaction with multiple inputs should not be coinbase")
	}
}

func TestIsCoinbaseTx_False_NonEmptyID(t *testing.T) {
	t.Parallel()
	tx := &Transaction{
		ID: []byte("test"),
		Inputs: []TxInput{
			{ID: []byte("some_tx_id"), Out: -1, Signature: nil, PubKey: []byte("data")},
		},
		Outputs: []TxOutput{},
	}

	if tx.IsCoinbaseTx() {
		t.Error("Transaction with non-empty input ID should not be coinbase")
	}
}

func TestIsCoinbaseTx_False_NonNegativeOut(t *testing.T) {
	t.Parallel()
	tx := &Transaction{
		ID: []byte("test"),
		Inputs: []TxInput{
			{ID: []byte{}, Out: 0, Signature: nil, PubKey: []byte("data")},
		},
		Outputs: []TxOutput{},
	}

	if tx.IsCoinbaseTx() {
		t.Error("Transaction with Out >= 0 should not be coinbase")
	}
}

// =============================================================================
// Transaction Hash Tests
// =============================================================================

func TestTransaction_HashTransaction(t *testing.T) {
	t.Parallel()
	tx := CoinbaseTx(CreateTestWallet(), "data")

	hash := tx.HashTransaction()

	if len(hash) != 32 { // SHA-256 produces 32-byte hash
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

func TestTransaction_HashDeterministic(t *testing.T) {
	t.Parallel()
	tx := &Transaction{
		ID: []byte("test_id"),
		Inputs: []TxInput{
			{ID: []byte("input_id"), Out: 0, Signature: nil, PubKey: []byte("pubkey")},
		},
		Outputs: []TxOutput{
			{Value: 100, PubKeyHash: []byte("pubkeyhash")},
		},
	}

	hash1 := tx.HashTransaction()
	hash2 := tx.HashTransaction()

	if !bytes.Equal(hash1, hash2) {
		t.Error("HashTransaction should be deterministic")
	}
}

func TestTransaction_DifferentTxsDifferentHashes(t *testing.T) {
	t.Parallel()
	tx1 := &Transaction{
		ID: []byte("id1"),
		Inputs: []TxInput{
			{ID: []byte("input1"), Out: 0, Signature: nil, PubKey: []byte("pk1")},
		},
		Outputs: []TxOutput{
			{Value: 100, PubKeyHash: []byte("hash1")},
		},
	}

	tx2 := &Transaction{
		ID: []byte("id2"),
		Inputs: []TxInput{
			{ID: []byte("input2"), Out: 1, Signature: nil, PubKey: []byte("pk2")},
		},
		Outputs: []TxOutput{
			{Value: 200, PubKeyHash: []byte("hash2")},
		},
	}

	if bytes.Equal(tx1.HashTransaction(), tx2.HashTransaction()) {
		t.Error("Different transactions should have different hashes")
	}
}

// =============================================================================
// TrimmedCopy Tests
// =============================================================================

func TestTransaction_TrimmedCopy(t *testing.T) {
	t.Parallel()
	original := &Transaction{
		ID: []byte("original_id"),
		Inputs: []TxInput{
			{ID: []byte("in1"), Out: 0, Signature: []byte("sig1"), PubKey: []byte("pk1")},
			{ID: []byte("in2"), Out: 1, Signature: []byte("sig2"), PubKey: []byte("pk2")},
		},
		Outputs: []TxOutput{
			{Value: 50, PubKeyHash: []byte("hash1")},
			{Value: 30, PubKeyHash: []byte("hash2")},
		},
	}

	trimmed := original.TrimmedCopy()

	// ID should be copied
	if !bytes.Equal(trimmed.ID, original.ID) {
		t.Error("Trimmed copy should have same ID")
	}

	// Inputs should have ID and Out copied, but not Signature and PubKey
	for i, input := range trimmed.Inputs {
		if !bytes.Equal(input.ID, original.Inputs[i].ID) {
			t.Errorf("Input %d ID should be copied", i)
		}
		if input.Out != original.Inputs[i].Out {
			t.Errorf("Input %d Out should be copied", i)
		}
		if input.Signature != nil {
			t.Errorf("Input %d Signature should be nil in trimmed copy", i)
		}
		if input.PubKey != nil {
			t.Errorf("Input %d PubKey should be nil in trimmed copy", i)
		}
	}

	// Outputs should be fully copied
	for i, output := range trimmed.Outputs {
		if output.Value != original.Outputs[i].Value {
			t.Errorf("Output %d Value should be copied", i)
		}
		if !bytes.Equal(output.PubKeyHash, original.Outputs[i].PubKeyHash) {
			t.Errorf("Output %d PubKeyHash should be copied", i)
		}
	}
}

func TestTransaction_TrimmedCopyIndependence(t *testing.T) {
	t.Parallel()
	original := &Transaction{
		ID:      []byte("id"),
		Inputs:  []TxInput{{ID: []byte("in"), Out: 0, Signature: []byte("sig"), PubKey: []byte("pk")}},
		Outputs: []TxOutput{{Value: 100, PubKeyHash: []byte("hash")}},
	}

	trimmed := original.TrimmedCopy()

	// Modify trimmed copy
	trimmed.ID = []byte("modified_id")
	trimmed.Inputs[0].ID = []byte("modified_in")
	trimmed.Outputs[0].Value = 999

	// Original should not be affected
	if bytes.Equal(original.ID, []byte("modified_id")) {
		t.Error("Modifying trimmed copy should not affect original ID")
	}
	if bytes.Equal(original.Inputs[0].ID, []byte("modified_in")) {
		t.Error("Modifying trimmed copy should not affect original input ID")
	}
	if original.Outputs[0].Value == 999 {
		t.Error("Modifying trimmed copy should not affect original output value")
	}

	// Verify trimmed copy was actually modified
	if !bytes.Equal(trimmed.ID, []byte("modified_id")) {
		t.Error("Trimmed copy ID should be modified")
	}
	if !bytes.Equal(trimmed.Inputs[0].ID, []byte("modified_in")) {
		t.Error("Trimmed copy input ID should be modified")
	}
	if trimmed.Outputs[0].Value != 999 {
		t.Error("Trimmed copy output value should be modified")
	}
}

// =============================================================================
// Sign and Verify Tests
// =============================================================================

func TestTransaction_SignAndVerify(t *testing.T) {
	t.Parallel()

	// Create key pair
	pubKey, privKey, _ := ed25519.GenerateKey(nil)

	// Create a previous transaction that we'll reference
	prevTx := CreateTestCoinbaseTx("prev_data")
	prevTx.ID = []byte("prev_tx_id")
	prevTx.Outputs = []TxOutput{
		{Value: 100, PubKeyHash: []byte("test_pubkey_hash")},
	}

	// Create transaction to sign
	tx := &Transaction{
		ID: []byte("test_tx_id"),
		Inputs: []TxInput{
			{ID: prevTx.ID, Out: 0, Signature: nil, PubKey: pubKey},
		},
		Outputs: []TxOutput{
			{Value: 50, PubKeyHash: []byte("recipient_hash")},
			{Value: 50, PubKeyHash: []byte("change_hash")},
		},
	}

	prevTXs := map[string]Transaction{
		hex.EncodeToString(prevTx.ID): *prevTx,
	}

	// Sign the transaction
	tx.Sign(privKey, prevTXs)

	// Verify each input has a signature
	for i, input := range tx.Inputs {
		if len(input.Signature) == 0 {
			t.Errorf("Input %d should have signature after signing", i)
		}
	}

	// Verify the transaction
	if !tx.Verify(prevTXs) {
		t.Error("Transaction verification should pass after signing")
	}
}

func TestTransaction_VerifyFailsWithWrongKey(t *testing.T) {
	t.Parallel()

	pubKey1, privKey1, _ := ed25519.GenerateKey(nil)
	pubKey2, _, _ := ed25519.GenerateKey(nil)

	prevTx := CreateTestCoinbaseTx("data")
	prevTx.ID = []byte("prev_id")
	prevTx.Outputs = []TxOutput{{Value: 100, PubKeyHash: []byte("hash")}}

	tx := &Transaction{
		ID: []byte("tx_id"),
		Inputs: []TxInput{
			{ID: prevTx.ID, Out: 0, Signature: nil, PubKey: pubKey1},
		},
		Outputs: []TxOutput{{Value: 100, PubKeyHash: []byte("out_hash")}},
	}

	prevTXs := map[string]Transaction{hex.EncodeToString(prevTx.ID): *prevTx}

	// Sign with key1
	tx.Sign(privKey1, prevTXs)

	// Replace public key with different key
	tx.Inputs[0].PubKey = pubKey2

	// Verification should fail
	if tx.Verify(prevTXs) {
		t.Error("Verification should fail with wrong public key")
	}
}

func TestTransaction_SignCoinbase(t *testing.T) {
	t.Parallel()

	_, privKey, _ := ed25519.GenerateKey(nil)
	tx := CoinbaseTx(CreateTestWallet(), "data")

	// Signing a coinbase tx should not panic and should be a no-op
	tx.Sign(privKey, nil)

	// Input should still have nil signature
	if tx.Inputs[0].Signature != nil {
		t.Error("Coinbase transaction should not be signed")
	}
}

func TestTransaction_VerifyCoinbase(t *testing.T) {
	t.Parallel()

	tx := CoinbaseTx(CreateTestWallet(), "data")

	// Coinbase verification should always return true
	if !tx.Verify(nil) {
		t.Error("Coinbase transaction verification should always return true")
	}
}

// =============================================================================
// sortedInputIndices Tests
// =============================================================================

func TestTransaction_SortedInputIndices(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		ID: []byte("tx"),
		Inputs: []TxInput{
			{ID: []byte{0x03}, Out: 0}, // index 0, ID: "03"
			{ID: []byte{0x01}, Out: 1}, // index 1, ID: "01"
			{ID: []byte{0x02}, Out: 2}, // index 2, ID: "02"
		},
		Outputs: []TxOutput{},
	}

	indices := tx.sortedInputIndices()

	// Expected order: "01" (idx 1), "02" (idx 2), "03" (idx 0)
	if indices[0] != 1 {
		t.Errorf("Expected first index to be 1, got %d", indices[0])
	}
	if indices[1] != 2 {
		t.Errorf("Expected second index to be 2, got %d", indices[1])
	}
	if indices[2] != 0 {
		t.Errorf("Expected third index to be 0, got %d", indices[2])
	}
}

func TestTransaction_SortedInputIndices_SingleInput(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		ID:      []byte("tx"),
		Inputs:  []TxInput{{ID: []byte{0x01}, Out: 0}},
		Outputs: []TxOutput{},
	}

	indices := tx.sortedInputIndices()

	if len(indices) != 1 {
		t.Errorf("Expected 1 index, got %d", len(indices))
	}
	if indices[0] != 0 {
		t.Errorf("Expected index 0, got %d", indices[0])
	}
}

func TestTransaction_SortedInputIndices_Empty(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		ID:      []byte("tx"),
		Inputs:  []TxInput{},
		Outputs: []TxOutput{},
	}

	indices := tx.sortedInputIndices()

	if len(indices) != 0 {
		t.Errorf("Expected 0 indices for empty inputs, got %d", len(indices))
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestTransaction_EmptyInputsAndOutputs(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		ID:      []byte("empty_tx"),
		Inputs:  []TxInput{},
		Outputs: []TxOutput{},
	}

	// Should not panic
	hash := tx.HashTransaction()
	if len(hash) != 32 {
		t.Error("Hash should still be 32 bytes for empty transaction")
	}

	trimmed := tx.TrimmedCopy()
	if len(trimmed.Inputs) != 0 || len(trimmed.Outputs) != 0 {
		t.Error("Trimmed copy of empty tx should also be empty")
	}
}

func TestTransaction_VeryLargeTransaction(t *testing.T) {
	t.Parallel()

	// Create transaction with many inputs and outputs
	inputCount := 100
	outputCount := 100

	inputs := make([]TxInput, inputCount)
	for i := 0; i < inputCount; i++ {
		inputs[i] = TxInput{
			ID:        []byte("input_id_" + string(rune(i))),
			Out:       i,
			Signature: nil,
			PubKey:    []byte("pubkey"),
		}
	}

	outputs := make([]TxOutput, outputCount)
	for i := 0; i < outputCount; i++ {
		outputs[i] = TxOutput{
			Value:      i + 1,
			PubKeyHash: []byte("hash_" + string(rune(i))),
		}
	}

	tx := &Transaction{
		ID:      []byte("large_tx"),
		Inputs:  inputs,
		Outputs: outputs,
	}

	// Should handle large transactions without issue
	hash := tx.HashTransaction()
	if len(hash) != 32 {
		t.Error("Large transaction should still produce valid hash")
	}

	trimmed := tx.TrimmedCopy()
	if len(trimmed.Inputs) != inputCount {
		t.Errorf("Expected %d inputs in trimmed copy, got %d", inputCount, len(trimmed.Inputs))
	}
	if len(trimmed.Outputs) != outputCount {
		t.Errorf("Expected %d outputs in trimmed copy, got %d", outputCount, len(trimmed.Outputs))
	}
}
