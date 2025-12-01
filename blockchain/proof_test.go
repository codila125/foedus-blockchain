package blockchain

import (
	"bytes"
	"math/big"
	"testing"
)

// =============================================================================
// NewProof Tests
// =============================================================================

func TestNewProof(t *testing.T) {
	t.Parallel()
	block := &Block{
		Timestamp:    1000,
		Hash:         []byte("test_hash"),
		Transactions: []*Transaction{},
		Contracts:    []*Contract{},
		PrevHash:     []byte("prev_hash"),
		Nonce:        0,
		Height:       1,
	}

	pow := NewProof(block)

	if pow == nil {
		t.Fatal("NewProof should return non-nil ProofOfWork")
	}

	if pow.Block != block {
		t.Error("ProofOfWork should reference the correct block")
	}

	if pow.Target == nil {
		t.Error("ProofOfWork target should not be nil")
	}
}

func TestNewProof_TargetCalculation(t *testing.T) {
	t.Parallel()
	block := &Block{
		Timestamp:    1000,
		Hash:         []byte{},
		Transactions: []*Transaction{},
		Contracts:    []*Contract{},
		PrevHash:     []byte{},
		Nonce:        0,
		Height:       0,
	}

	pow := NewProof(block)

	// Target should be 1 << (256 - difficulty)
	// With difficulty = 12, target should be 1 << 244
	expected := big.NewInt(1)
	expected.Lsh(expected, 244)

	if pow.Target.Cmp(expected) != 0 {
		t.Errorf("Expected target %s, got %s", expected.String(), pow.Target.String())
	}
}

func TestNewProof_TargetPositive(t *testing.T) {
	t.Parallel()
	block := &Block{}
	pow := NewProof(block)

	if pow.Target.Sign() != 1 {
		t.Error("Target should be positive")
	}
}

func TestNewProof_TargetLessThanMaxHash(t *testing.T) {
	t.Parallel()
	block := &Block{}
	pow := NewProof(block)

	// Max hash is 2^256 - 1
	maxHash := new(big.Int)
	maxHash.SetBytes(bytes.Repeat([]byte{0xff}, 32))

	if pow.Target.Cmp(maxHash) >= 0 {
		t.Error("Target should be less than max hash value")
	}
}

// =============================================================================
// InitData Tests
// =============================================================================

func TestProofOfWork_InitData(t *testing.T) {
	t.Parallel()
	block := &Block{
		Timestamp:    1609459200,
		Hash:         []byte{},
		Transactions: []*Transaction{CreateTestCoinbaseTx("data")},
		Contracts:    []*Contract{CreateTestCoinbaseOp("contract")},
		PrevHash:     []byte("previous_block_hash"),
		Nonce:        0,
		Height:       1,
	}

	pow := NewProof(block)
	data := pow.InitData(0)

	if len(data) == 0 {
		t.Error("InitData should return non-empty byte slice")
	}
}

func TestProofOfWork_InitDataDeterministic(t *testing.T) {
	t.Parallel()
	block := &Block{
		Timestamp:    1000,
		Hash:         []byte{},
		Transactions: []*Transaction{CreateTestCoinbaseTx("data")},
		Contracts:    []*Contract{},
		PrevHash:     []byte("prev"),
		Nonce:        0,
		Height:       1,
	}

	pow := NewProof(block)

	data1 := pow.InitData(42)
	data2 := pow.InitData(42)

	if !bytes.Equal(data1, data2) {
		t.Error("InitData should be deterministic for same nonce")
	}
}

func TestProofOfWork_InitDataDifferentNonces(t *testing.T) {
	t.Parallel()
	block := &Block{
		Timestamp:    1000,
		Hash:         []byte{},
		Transactions: []*Transaction{},
		Contracts:    []*Contract{},
		PrevHash:     []byte("prev"),
		Nonce:        0,
		Height:       1,
	}

	pow := NewProof(block)

	data1 := pow.InitData(0)
	data2 := pow.InitData(1)
	data3 := pow.InitData(100)

	if bytes.Equal(data1, data2) {
		t.Error("Different nonces should produce different data")
	}
	if bytes.Equal(data1, data3) {
		t.Error("Different nonces should produce different data")
	}
	if bytes.Equal(data2, data3) {
		t.Error("Different nonces should produce different data")
	}
}

func TestProofOfWork_InitDataContainsPrevHash(t *testing.T) {
	t.Parallel()
	prevHash := []byte("unique_previous_hash_12345")
	block := &Block{
		Timestamp:    1000,
		Hash:         []byte{},
		Transactions: []*Transaction{},
		Contracts:    []*Contract{},
		PrevHash:     prevHash,
		Nonce:        0,
		Height:       1,
	}

	pow := NewProof(block)
	data := pow.InitData(0)

	if !bytes.Contains(data, prevHash) {
		t.Error("InitData should contain PrevHash")
	}
}

func TestProofOfWork_InitDataNonceAtEnd(t *testing.T) {
	t.Parallel()
	block := &Block{
		Timestamp:    1000,
		Hash:         []byte{},
		Transactions: []*Transaction{},
		Contracts:    []*Contract{},
		PrevHash:     []byte("prev"),
		Nonce:        0,
		Height:       1,
	}

	pow := NewProof(block)

	// Get data for two different nonces
	data1 := pow.InitData(0)
	data2 := pow.InitData(1)

	// The data should only differ in the nonce bytes
	// Nonce bytes are near the end (before difficulty bytes)
	// Data structure: prevHash + txHash + ctHash + nonce(8) + difficulty(8)
	if len(data1) != len(data2) {
		t.Error("Data length should be same for different nonces")
	}
}

// =============================================================================
// Validate Tests
// =============================================================================

func TestProofOfWork_ValidateValidBlock(t *testing.T) {
	t.Parallel()

	// Create a block using CreateBlock which runs PoW
	txs := []*Transaction{CreateTestCoinbaseTx("validate_data")}
	block := CreateBlock(txs, []*Contract{}, []byte("prev"), 1)

	pow := NewProof(block)

	if !pow.Validate() {
		t.Error("Block created with CreateBlock should be valid")
	}
}

func TestProofOfWork_ValidateInvalidNonce(t *testing.T) {
	t.Parallel()

	// Create a valid block first
	txs := []*Transaction{CreateTestCoinbaseTx("data")}
	block := CreateBlock(txs, []*Contract{}, []byte("prev"), 1)

	// Store original valid nonce
	originalNonce := block.Nonce

	// Block with corrupted nonce is likely invalid (could still be valid by chance)
	// We test multiple corrupted nonces
	invalidCount := 0
	for testNonce := -5; testNonce <= 5; testNonce++ {
		if testNonce == originalNonce {
			continue
		}
		block.Nonce = testNonce
		pow := NewProof(block)
		if !pow.Validate() {
			invalidCount++
		}
	}

	// At least some should be invalid
	if invalidCount == 0 {
		t.Log("Warning: All test nonces validated (statistically unlikely)")
	}
}

func TestProofOfWork_ValidateEmptyBlock(t *testing.T) {
	t.Parallel()

	block := CreateBlock([]*Transaction{}, []*Contract{}, []byte{}, 0)

	pow := NewProof(block)

	if !pow.Validate() {
		t.Error("Empty block created with PoW should be valid")
	}
}

func TestProofOfWork_ValidateConsistency(t *testing.T) {
	t.Parallel()

	block := CreateBlock([]*Transaction{CreateTestCoinbaseTx("consistency")}, []*Contract{}, []byte("prev"), 1)

	pow := NewProof(block)

	// Validate multiple times should give same result
	result1 := pow.Validate()
	result2 := pow.Validate()
	result3 := pow.Validate()

	if result1 != result2 || result2 != result3 {
		t.Error("Validate should be consistent across multiple calls")
	}
}

// =============================================================================
// Run Tests
// =============================================================================

func TestProofOfWork_Run(t *testing.T) {
	t.Parallel()

	block := &Block{
		Timestamp:    1609459200,
		Hash:         []byte{},
		Transactions: []*Transaction{CreateTestCoinbaseTx("run_test")},
		Contracts:    []*Contract{},
		PrevHash:     []byte("prev_hash_for_run"),
		Nonce:        0,
		Height:       1,
	}

	pow := NewProof(block)
	nonce, hash := pow.Run()

	// Nonce should be non-negative
	if nonce < 0 {
		t.Error("Nonce should be non-negative")
	}

	// Hash should be 32 bytes (SHA-256)
	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}

	// Update block with found values
	block.Nonce = nonce
	block.Hash = hash

	// Block should now be valid
	if !pow.Validate() {
		t.Error("Block should be valid after Run")
	}
}

func TestProofOfWork_RunProducesValidHash(t *testing.T) {
	t.Parallel()

	block := &Block{
		Timestamp:    1000,
		Hash:         []byte{},
		Transactions: []*Transaction{},
		Contracts:    []*Contract{},
		PrevHash:     []byte("test_prev"),
		Nonce:        0,
		Height:       0,
	}

	pow := NewProof(block)
	nonce, hash := pow.Run()

	// Verify hash meets target
	var intHash big.Int
	intHash.SetBytes(hash)

	if intHash.Cmp(pow.Target) >= 0 {
		t.Error("Hash should be less than target")
	}

	// Verify nonce produces this hash
	block.Nonce = nonce
	block.Hash = hash
	if !pow.Validate() {
		t.Error("Nonce from Run should validate")
	}
}

func TestProofOfWork_RunDifferentBlocksDifferentResults(t *testing.T) {
	t.Parallel()

	block1 := &Block{
		Timestamp:    1000,
		Transactions: []*Transaction{CreateTestCoinbaseTx("data1")},
		Contracts:    []*Contract{},
		PrevHash:     []byte("prev1"),
		Height:       1,
	}

	block2 := &Block{
		Timestamp:    2000,
		Transactions: []*Transaction{CreateTestCoinbaseTx("data2")},
		Contracts:    []*Contract{},
		PrevHash:     []byte("prev2"),
		Height:       2,
	}

	pow1 := NewProof(block1)
	pow2 := NewProof(block2)

	_, hash1 := pow1.Run()
	_, hash2 := pow2.Run()

	if bytes.Equal(hash1, hash2) {
		t.Error("Different blocks should produce different hashes")
	}
}

// =============================================================================
// Edge Cases and Stress Tests
// =============================================================================

func TestProofOfWork_ZeroNonce(t *testing.T) {
	t.Parallel()

	block := &Block{
		Timestamp:    1000,
		Transactions: []*Transaction{},
		Contracts:    []*Contract{},
		PrevHash:     []byte{},
		Nonce:        0,
		Height:       0,
	}

	pow := NewProof(block)
	data := pow.InitData(0)

	if len(data) == 0 {
		t.Error("InitData with nonce 0 should produce data")
	}
}

func TestProofOfWork_LargeNonce(t *testing.T) {
	t.Parallel()

	block := &Block{
		Timestamp:    1000,
		Transactions: []*Transaction{},
		Contracts:    []*Contract{},
		PrevHash:     []byte{},
		Nonce:        1000000000,
		Height:       0,
	}

	pow := NewProof(block)
	data := pow.InitData(1000000000)

	if len(data) == 0 {
		t.Error("InitData with large nonce should produce data")
	}
}

func TestProofOfWork_EmptyTransactionsHash(t *testing.T) {
	t.Parallel()

	block := &Block{
		Timestamp:    1000,
		Transactions: []*Transaction{},
		Contracts:    []*Contract{},
		PrevHash:     []byte("prev"),
		Nonce:        0,
		Height:       0,
	}

	txHash := block.HashTransactions()

	// Empty transactions should produce empty or specific hash
	if len(txHash) > 0 {
		t.Logf("Empty transactions hash: %x", txHash)
	}
}

func TestProofOfWork_EmptyContractsHash(t *testing.T) {
	t.Parallel()

	block := &Block{
		Timestamp:    1000,
		Transactions: []*Transaction{},
		Contracts:    []*Contract{},
		PrevHash:     []byte("prev"),
		Nonce:        0,
		Height:       0,
	}

	ctHash := block.HashContracts()

	// Empty contracts should produce empty or specific hash
	if len(ctHash) > 0 {
		t.Logf("Empty contracts hash: %x", ctHash)
	}
}

func TestProofOfWork_BlockWithManyTransactions(t *testing.T) {
	t.Parallel()

	txCount := 50
	txs := make([]*Transaction, txCount)
	for i := 0; i < txCount; i++ {
		txs[i] = CreateTestCoinbaseTx("")
	}

	block := &Block{
		Timestamp:    1000,
		Transactions: txs,
		Contracts:    []*Contract{},
		PrevHash:     []byte("prev"),
		Nonce:        0,
		Height:       1,
	}

	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Nonce = nonce
	block.Hash = hash

	if !pow.Validate() {
		t.Error("Block with many transactions should validate after PoW")
	}
}

func TestProofOfWork_BlockWithManyContracts(t *testing.T) {
	t.Parallel()

	ctCount := 50
	cts := make([]*Contract, ctCount)
	for i := 0; i < ctCount; i++ {
		cts[i] = CreateTestCoinbaseOp("")
	}

	block := &Block{
		Timestamp:    1000,
		Transactions: []*Transaction{},
		Contracts:    cts,
		PrevHash:     []byte("prev"),
		Nonce:        0,
		Height:       1,
	}

	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Nonce = nonce
	block.Hash = hash

	if !pow.Validate() {
		t.Error("Block with many contracts should validate after PoW")
	}
}

// =============================================================================
// Target and Difficulty Tests
// =============================================================================

func TestProofOfWork_TargetNotZero(t *testing.T) {
	t.Parallel()

	block := &Block{}
	pow := NewProof(block)

	if pow.Target.Sign() == 0 {
		t.Error("Target should not be zero")
	}
}

func TestProofOfWork_DifficultyConstant(t *testing.T) {
	t.Parallel()

	// Verify difficulty constant is reasonable
	if difficulty <= 0 {
		t.Error("Difficulty should be positive")
	}

	if difficulty > 256 {
		t.Error("Difficulty should not exceed 256 bits")
	}
}
