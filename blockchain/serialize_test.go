package blockchain

import (
	"bytes"
	"testing"
)

// =============================================================================
// SerializeOutputs / DeserializeOutputs Tests
// =============================================================================

func TestSerializeDeserializeOutputs(t *testing.T) {
	t.Parallel()

	outputs := TxOutputs{
		Outputs: []TxOutput{
			{Value: 100, PubKeyHash: []byte("hash1")},
			{Value: 200, PubKeyHash: []byte("hash2")},
			{Value: 300, PubKeyHash: []byte("hash3")},
		},
	}

	serialized := outputs.SerializeOutputs()
	deserialized := DeserializeOutputs(serialized)

	if len(deserialized.Outputs) != len(outputs.Outputs) {
		t.Errorf("Expected %d outputs, got %d", len(outputs.Outputs), len(deserialized.Outputs))
	}

	for i, out := range outputs.Outputs {
		if deserialized.Outputs[i].Value != out.Value {
			t.Errorf("Output %d value mismatch: expected %d, got %d", i, out.Value, deserialized.Outputs[i].Value)
		}
		if !bytes.Equal(deserialized.Outputs[i].PubKeyHash, out.PubKeyHash) {
			t.Errorf("Output %d PubKeyHash mismatch", i)
		}
	}
}

func TestSerializeOutputs_Empty(t *testing.T) {
	t.Parallel()

	outputs := TxOutputs{Outputs: []TxOutput{}}

	serialized := outputs.SerializeOutputs()
	deserialized := DeserializeOutputs(serialized)

	if len(deserialized.Outputs) != 0 {
		t.Errorf("Expected 0 outputs, got %d", len(deserialized.Outputs))
	}
}

func TestSerializeOutputs_SingleOutput(t *testing.T) {
	t.Parallel()

	outputs := TxOutputs{
		Outputs: []TxOutput{
			{Value: 50, PubKeyHash: []byte("single_hash")},
		},
	}

	serialized := outputs.SerializeOutputs()
	deserialized := DeserializeOutputs(serialized)

	if len(deserialized.Outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(deserialized.Outputs))
	}

	if deserialized.Outputs[0].Value != 50 {
		t.Errorf("Expected value 50, got %d", deserialized.Outputs[0].Value)
	}
}

func TestSerializeOutputs_LargeValues(t *testing.T) {
	t.Parallel()

	outputs := TxOutputs{
		Outputs: []TxOutput{
			{Value: 2147483647, PubKeyHash: []byte("max_int32_hash")}, // Max int32
		},
	}

	serialized := outputs.SerializeOutputs()
	deserialized := DeserializeOutputs(serialized)

	if deserialized.Outputs[0].Value != 2147483647 {
		t.Errorf("Expected value 2147483647, got %d", deserialized.Outputs[0].Value)
	}
}

func TestSerializeOutputs_LargePubKeyHash(t *testing.T) {
	t.Parallel()

	largeHash := make([]byte, 256)
	for i := range largeHash {
		largeHash[i] = byte(i)
	}

	outputs := TxOutputs{
		Outputs: []TxOutput{
			{Value: 100, PubKeyHash: largeHash},
		},
	}

	serialized := outputs.SerializeOutputs()
	deserialized := DeserializeOutputs(serialized)

	if !bytes.Equal(deserialized.Outputs[0].PubKeyHash, largeHash) {
		t.Error("Large PubKeyHash not preserved after serialization")
	}
}

// =============================================================================
// SerializeTransaction / DeserializeTransaction Tests
// =============================================================================

func TestSerializeDeserializeTransaction(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		ID: []byte("tx_id_123"),
		Inputs: []TxInput{
			{ID: []byte("input1"), Out: 0, Signature: []byte("sig1"), PubKey: []byte("pk1")},
			{ID: []byte("input2"), Out: 1, Signature: []byte("sig2"), PubKey: []byte("pk2")},
		},
		Outputs: []TxOutput{
			{Value: 50, PubKeyHash: []byte("hash1")},
			{Value: 30, PubKeyHash: []byte("hash2")},
		},
	}

	serialized := tx.SerializeTransaction()
	deserialized := DeserializeTransaction(serialized)

	if !bytes.Equal(deserialized.ID, tx.ID) {
		t.Error("Transaction ID mismatch")
	}

	if len(deserialized.Inputs) != len(tx.Inputs) {
		t.Errorf("Expected %d inputs, got %d", len(tx.Inputs), len(deserialized.Inputs))
	}

	if len(deserialized.Outputs) != len(tx.Outputs) {
		t.Errorf("Expected %d outputs, got %d", len(tx.Outputs), len(deserialized.Outputs))
	}

	for i, in := range tx.Inputs {
		if !bytes.Equal(deserialized.Inputs[i].ID, in.ID) {
			t.Errorf("Input %d ID mismatch", i)
		}
		if deserialized.Inputs[i].Out != in.Out {
			t.Errorf("Input %d Out mismatch", i)
		}
		if !bytes.Equal(deserialized.Inputs[i].Signature, in.Signature) {
			t.Errorf("Input %d Signature mismatch", i)
		}
		if !bytes.Equal(deserialized.Inputs[i].PubKey, in.PubKey) {
			t.Errorf("Input %d PubKey mismatch", i)
		}
	}

	for i, out := range tx.Outputs {
		if deserialized.Outputs[i].Value != out.Value {
			t.Errorf("Output %d Value mismatch", i)
		}
		if !bytes.Equal(deserialized.Outputs[i].PubKeyHash, out.PubKeyHash) {
			t.Errorf("Output %d PubKeyHash mismatch", i)
		}
	}
}

func TestSerializeTransaction_Empty(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		ID:      []byte{},
		Inputs:  []TxInput{},
		Outputs: []TxOutput{},
	}

	serialized := tx.SerializeTransaction()
	deserialized := DeserializeTransaction(serialized)

	if len(deserialized.Inputs) != 0 {
		t.Errorf("Expected 0 inputs, got %d", len(deserialized.Inputs))
	}

	if len(deserialized.Outputs) != 0 {
		t.Errorf("Expected 0 outputs, got %d", len(deserialized.Outputs))
	}
}

func TestSerializeTransaction_CoinbaseTx(t *testing.T) {
	t.Parallel()

	tx := CreateTestCoinbaseTx("test_data")

	serialized := tx.SerializeTransaction()
	deserialized := DeserializeTransaction(serialized)

	if !bytes.Equal(deserialized.ID, tx.ID) {
		t.Error("Coinbase transaction ID mismatch after serialization")
	}

	if len(deserialized.Inputs) != 1 {
		t.Errorf("Expected 1 input for coinbase, got %d", len(deserialized.Inputs))
	}

	if len(deserialized.Outputs) != 1 {
		t.Errorf("Expected 1 output for coinbase, got %d", len(deserialized.Outputs))
	}
}

// =============================================================================
// SerializeForSigning Tests
// =============================================================================

func TestSerializeForSigning(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		ID: []byte("tx_id"),
		Inputs: []TxInput{
			{ID: []byte("in1"), Out: 0, Signature: []byte("sig"), PubKey: []byte("pk")},
		},
		Outputs: []TxOutput{
			{Value: 100, PubKeyHash: []byte("hash")},
		},
	}

	serialized := tx.SerializeForSigning()

	if len(serialized) == 0 {
		t.Error("SerializeForSigning should return non-empty data")
	}
}

func TestSerializeForSigning_ExcludesSignatureAndPubKey(t *testing.T) {
	t.Parallel()

	tx1 := &Transaction{
		ID: []byte("tx_id"),
		Inputs: []TxInput{
			{ID: []byte("in1"), Out: 0, Signature: []byte("signature1"), PubKey: []byte("pubkey1")},
		},
		Outputs: []TxOutput{
			{Value: 100, PubKeyHash: []byte("hash")},
		},
	}

	tx2 := &Transaction{
		ID: []byte("tx_id"),
		Inputs: []TxInput{
			{ID: []byte("in1"), Out: 0, Signature: []byte("different_sig"), PubKey: []byte("different_pk")},
		},
		Outputs: []TxOutput{
			{Value: 100, PubKeyHash: []byte("hash")},
		},
	}

	serialized1 := tx1.SerializeForSigning()
	serialized2 := tx2.SerializeForSigning()

	// Should be equal because signatures and pubkeys are excluded
	if !bytes.Equal(serialized1, serialized2) {
		t.Error("SerializeForSigning should exclude Signature and PubKey")
	}
}

func TestSerializeForSigning_Deterministic(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		ID: []byte("deterministic_tx"),
		Inputs: []TxInput{
			{ID: []byte("in"), Out: 5, Signature: nil, PubKey: nil},
		},
		Outputs: []TxOutput{
			{Value: 200, PubKeyHash: []byte("h")},
		},
	}

	serialized1 := tx.SerializeForSigning()
	serialized2 := tx.SerializeForSigning()

	if !bytes.Equal(serialized1, serialized2) {
		t.Error("SerializeForSigning should be deterministic")
	}
}

// =============================================================================
// SerializeForVerification Tests
// =============================================================================

func TestSerializeForVerification(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		ID: []byte("tx_id"),
		Inputs: []TxInput{
			{ID: []byte("in1"), Out: 0, Signature: []byte("sig"), PubKey: []byte("pk")},
			{ID: []byte("in2"), Out: 1, Signature: []byte("sig2"), PubKey: []byte("pk2")},
		},
		Outputs: []TxOutput{
			{Value: 100, PubKeyHash: []byte("hash")},
		},
	}

	serialized0 := tx.SerializeForVerification(0)
	serialized1 := tx.SerializeForVerification(1)

	if len(serialized0) == 0 {
		t.Error("SerializeForVerification should return non-empty data")
	}

	// Different input indices should produce different results
	if bytes.Equal(serialized0, serialized1) {
		t.Error("Different input indices should produce different serialization")
	}
}

// =============================================================================
// SerializeBlock / DeserializeBlock Tests
// =============================================================================

func TestSerializeDeserializeBlock(t *testing.T) {
	t.Parallel()

	block := CreateBlock(
		[]*Transaction{CreateTestCoinbaseTx("data")},
		[]*Contract{CreateTestCoinbaseOp("contract")},
		[]byte("prev_hash"),
		5,
	)

	serialized := block.SerializeBlock()
	deserialized := DeserializeBlock(serialized)

	if deserialized == nil {
		t.Fatal("DeserializeBlock should not return nil")
	}

	if !bytes.Equal(deserialized.Hash, block.Hash) {
		t.Error("Block Hash mismatch")
	}

	if !bytes.Equal(deserialized.PrevHash, block.PrevHash) {
		t.Error("Block PrevHash mismatch")
	}

	if deserialized.Height != block.Height {
		t.Errorf("Expected height %d, got %d", block.Height, deserialized.Height)
	}

	if deserialized.Nonce != block.Nonce {
		t.Errorf("Expected nonce %d, got %d", block.Nonce, deserialized.Nonce)
	}

	if deserialized.Timestamp != block.Timestamp {
		t.Errorf("Expected timestamp %d, got %d", block.Timestamp, deserialized.Timestamp)
	}

	if len(deserialized.Transactions) != len(block.Transactions) {
		t.Errorf("Expected %d transactions, got %d", len(block.Transactions), len(deserialized.Transactions))
	}

	if len(deserialized.Contracts) != len(block.Contracts) {
		t.Errorf("Expected %d contracts, got %d", len(block.Contracts), len(deserialized.Contracts))
	}
}

func TestSerializeBlock_GenesisBlock(t *testing.T) {
	t.Parallel()

	genesis := Genesis(CreateTestCoinbaseTx("data"), CreateTestCoinbaseOp("contract"))

	serialized := genesis.SerializeBlock()
	deserialized := DeserializeBlock(serialized)

	if deserialized.Height != 0 {
		t.Errorf("Genesis block should have height 0, got %d", deserialized.Height)
	}

	if len(deserialized.PrevHash) != 0 {
		t.Error("Genesis block should have empty PrevHash")
	}
}

func TestSerializeBlock_EmptyBlock(t *testing.T) {
	t.Parallel()

	block := CreateBlock([]*Transaction{}, []*Contract{}, []byte{}, 0)

	serialized := block.SerializeBlock()
	deserialized := DeserializeBlock(serialized)

	if len(deserialized.Transactions) != 0 {
		t.Errorf("Expected 0 transactions, got %d", len(deserialized.Transactions))
	}

	if len(deserialized.Contracts) != 0 {
		t.Errorf("Expected 0 contracts, got %d", len(deserialized.Contracts))
	}
}

func TestSerializeBlock_ManyTransactions(t *testing.T) {
	t.Parallel()

	txCount := 50
	txs := make([]*Transaction, txCount)
	for i := 0; i < txCount; i++ {
		txs[i] = CreateTestCoinbaseTx("")
	}

	block := CreateBlock(txs, []*Contract{}, []byte("prev"), 1)

	serialized := block.SerializeBlock()
	deserialized := DeserializeBlock(serialized)

	if len(deserialized.Transactions) != txCount {
		t.Errorf("Expected %d transactions, got %d", txCount, len(deserialized.Transactions))
	}
}

// =============================================================================
// SerializeMilestoneCore Tests
// =============================================================================

func TestSerializeMilestoneCore(t *testing.T) {
	t.Parallel()

	core := &MilestoneCore{
		Title:       "Test Milestone",
		Description: "Test Description",
		Value:       1000,
		CreatedAt:   1609459200,
	}

	serialized := core.SerializeMilestoneCore()

	if len(serialized) == 0 {
		t.Error("SerializeMilestoneCore should return non-empty data")
	}
}

func TestSerializeMilestoneCore_Deterministic(t *testing.T) {
	t.Parallel()

	core := &MilestoneCore{
		Title:       "Deterministic",
		Description: "Testing",
		Value:       500,
		CreatedAt:   1000,
	}

	serialized1 := core.SerializeMilestoneCore()
	serialized2 := core.SerializeMilestoneCore()

	if !bytes.Equal(serialized1, serialized2) {
		t.Error("SerializeMilestoneCore should be deterministic")
	}
}

// =============================================================================
// SerializeContractCore Tests
// =============================================================================

func TestSerializeContractCore(t *testing.T) {
	t.Parallel()

	core := &ContractCore{
		Title:          "Test Contract",
		Description:    "Test Description",
		CreatedAt:      1000,
		Milestones:     []*MilestoneCore{},
		Parties:        []*PartyCore{},
		Terms:          []byte("terms"),
		CreatorAddress: "creator",
		Attachments:    [][]byte{},
	}

	serialized := core.SerializeContractCore()

	if len(serialized) == 0 {
		t.Error("SerializeContractCore should return non-empty data")
	}
}

func TestSerializeContractCore_WithMilestones(t *testing.T) {
	t.Parallel()

	core := &ContractCore{
		Title:       "Contract",
		Description: "Desc",
		CreatedAt:   1000,
		Milestones: []*MilestoneCore{
			{Title: "M1", Description: "D1", Value: 100, CreatedAt: 1000},
			{Title: "M2", Description: "D2", Value: 200, CreatedAt: 2000},
		},
		Parties:        []*PartyCore{},
		CreatorAddress: "creator",
	}

	serialized := core.SerializeContractCore()

	if len(serialized) == 0 {
		t.Error("SerializeContractCore with milestones should return non-empty data")
	}
}

func TestSerializeContractCore_WithParties(t *testing.T) {
	t.Parallel()

	core := &ContractCore{
		Title:       "Contract",
		Description: "Desc",
		CreatedAt:   1000,
		Milestones:  []*MilestoneCore{},
		Parties: []*PartyCore{
			{Address: "addr1", Role: "CREATOR", PublicKey: []byte("pk1")},
			{Address: "addr2", Role: "CONTRACTOR", PublicKey: []byte("pk2")},
		},
		CreatorAddress: "creator",
	}

	serialized := core.SerializeContractCore()

	if len(serialized) == 0 {
		t.Error("SerializeContractCore with parties should return non-empty data")
	}
}

func TestSerializeContractCore_Deterministic(t *testing.T) {
	t.Parallel()

	core := &ContractCore{
		Title:          "Determinism Test",
		Description:    "Testing",
		CreatedAt:      1000,
		Milestones:     []*MilestoneCore{{Title: "M", Description: "D", Value: 100, CreatedAt: 1000}},
		Parties:        []*PartyCore{{Address: "a", Role: "R", PublicKey: []byte("p")}},
		Terms:          []byte("terms"),
		CreatorAddress: "c",
		Attachments:    [][]byte{[]byte("att")},
	}

	serialized1 := core.SerializeContractCore()
	serialized2 := core.SerializeContractCore()

	if !bytes.Equal(serialized1, serialized2) {
		t.Error("SerializeContractCore should be deterministic")
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestSerialization_NilValues(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		ID:      nil,
		Inputs:  nil,
		Outputs: nil,
	}

	// Should handle nil gracefully
	serialized := tx.SerializeTransaction()
	if len(serialized) == 0 {
		t.Log("Nil transaction produces empty serialization")
	}
}

func TestSerialization_LargeData(t *testing.T) {
	t.Parallel()

	// Create large transaction
	inputCount := 100
	inputs := make([]TxInput, inputCount)
	for i := 0; i < inputCount; i++ {
		inputs[i] = TxInput{
			ID:        make([]byte, 64),
			Out:       i,
			Signature: make([]byte, 64),
			PubKey:    make([]byte, 32),
		}
	}

	outputCount := 100
	outputs := make([]TxOutput, outputCount)
	for i := 0; i < outputCount; i++ {
		outputs[i] = TxOutput{
			Value:      i * 100,
			PubKeyHash: make([]byte, 32),
		}
	}

	tx := &Transaction{
		ID:      make([]byte, 32),
		Inputs:  inputs,
		Outputs: outputs,
	}

	serialized := tx.SerializeTransaction()
	deserialized := DeserializeTransaction(serialized)

	if len(deserialized.Inputs) != inputCount {
		t.Errorf("Expected %d inputs after deserialization, got %d", inputCount, len(deserialized.Inputs))
	}

	if len(deserialized.Outputs) != outputCount {
		t.Errorf("Expected %d outputs after deserialization, got %d", outputCount, len(deserialized.Outputs))
	}
}

func TestSerialization_ZeroValues(t *testing.T) {
	t.Parallel()

	outputs := TxOutputs{
		Outputs: []TxOutput{
			{Value: 0, PubKeyHash: []byte{}},
		},
	}

	serialized := outputs.SerializeOutputs()
	deserialized := DeserializeOutputs(serialized)

	if deserialized.Outputs[0].Value != 0 {
		t.Error("Zero value should be preserved")
	}
}
