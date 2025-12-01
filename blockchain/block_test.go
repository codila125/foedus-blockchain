package blockchain

import (
	"bytes"
	"testing"
	"time"
)

// =============================================================================
// Block Structure Tests
// =============================================================================

func TestBlockFields(t *testing.T) {
	t.Parallel()
	txs := []*Transaction{CreateTestCoinbaseTx("test_data")}
	cts := []*Contract{CreateTestCoinbaseOp("test_contract")}
	prevHash := []byte("previous_hash")
	height := 1

	block := CreateBlock(txs, cts, prevHash, height)

	if block == nil {
		t.Fatal("CreateBlock should return non-nil block")
	}

	if block.Height != height {
		t.Errorf("Expected height %d, got %d", height, block.Height)
	}

	if !bytes.Equal(block.PrevHash, prevHash) {
		t.Errorf("Expected prevHash %x, got %x", prevHash, block.PrevHash)
	}

	if len(block.Transactions) != len(txs) {
		t.Errorf("Expected %d transactions, got %d", len(txs), len(block.Transactions))
	}

	if len(block.Contracts) != len(cts) {
		t.Errorf("Expected %d contracts, got %d", len(cts), len(block.Contracts))
	}

	if len(block.Hash) == 0 {
		t.Error("Block hash should not be empty after creation")
	}

	if block.Nonce < 0 {
		t.Error("Block nonce should be non-negative")
	}

	if block.Timestamp == 0 {
		t.Error("Block timestamp should not be zero")
	}
}

// =============================================================================
// CreateBlock Tests
// =============================================================================

func TestCreateBlock_WithTransactionsAndContracts(t *testing.T) {
	t.Parallel()
	txs := []*Transaction{
		CreateTestCoinbaseTx("data1"),
		CreateTestCoinbaseTx("data2"),
	}
	cts := []*Contract{
		CreateTestCoinbaseOp("contract1"),
		CreateTestCoinbaseOp("contract2"),
	}
	prevHash := []byte("test_prev_hash")
	height := 5

	block := CreateBlock(txs, cts, prevHash, height)

	if len(block.Transactions) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(block.Transactions))
	}

	if len(block.Contracts) != 2 {
		t.Errorf("Expected 2 contracts, got %d", len(block.Contracts))
	}

	if block.Height != 5 {
		t.Errorf("Expected height 5, got %d", block.Height)
	}
}

func TestCreateBlock_EmptyTransactionsAndContracts(t *testing.T) {
	t.Parallel()
	txs := []*Transaction{}
	cts := []*Contract{}
	prevHash := []byte("empty_prev_hash")
	height := 0

	block := CreateBlock(txs, cts, prevHash, height)

	if len(block.Transactions) != 0 {
		t.Errorf("Expected 0 transactions, got %d", len(block.Transactions))
	}

	if len(block.Contracts) != 0 {
		t.Errorf("Expected 0 contracts, got %d", len(block.Contracts))
	}

	if len(block.Hash) == 0 {
		t.Error("Block hash should not be empty even with empty txs/contracts")
	}
}

func TestCreateBlock_NilTransactionsAndContracts(t *testing.T) {
	t.Parallel()
	prevHash := []byte("nil_prev_hash")
	height := 1

	block := CreateBlock(nil, nil, prevHash, height)

	if len(block.Transactions) > 0 {
		t.Error("Expected empty transactions for nil input")
	}

	if len(block.Contracts) > 0 {
		t.Error("Expected empty contracts for nil input")
	}
}

func TestCreateBlock_EmptyPrevHash(t *testing.T) {
	t.Parallel()
	txs := []*Transaction{CreateTestCoinbaseTx("test")}
	cts := []*Contract{}
	prevHash := []byte{}
	height := 0

	block := CreateBlock(txs, cts, prevHash, height)

	if len(block.PrevHash) != 0 {
		t.Error("Expected empty PrevHash for genesis-like block")
	}

	if block.Height != 0 {
		t.Errorf("Expected height 0, got %d", block.Height)
	}
}

func TestCreateBlock_LargeHeight(t *testing.T) {
	t.Parallel()
	txs := []*Transaction{CreateTestCoinbaseTx("test")}
	cts := []*Contract{}
	prevHash := []byte("large_height_hash")
	height := 1000000

	block := CreateBlock(txs, cts, prevHash, height)

	if block.Height != height {
		t.Errorf("Expected height %d, got %d", height, block.Height)
	}
}

// =============================================================================
// CreateBlockWithTimestamp Tests
// =============================================================================

func TestCreateBlockWithTimestamp(t *testing.T) {
	t.Parallel()
	txs := []*Transaction{CreateTestCoinbaseTx("test_data")}
	cts := []*Contract{}
	prevHash := []byte("timestamp_test_hash")
	height := 1
	fixedTimestamp := int64(1609459200)

	block := CreateBlockWithTimestamp(txs, cts, prevHash, height, fixedTimestamp)

	if block.Timestamp != fixedTimestamp {
		t.Errorf("Expected timestamp %d, got %d", fixedTimestamp, block.Timestamp)
	}
}

func TestCreateBlockWithTimestamp_Determinism(t *testing.T) {
	t.Parallel()
	txs := []*Transaction{CreateTestCoinbaseTx("same_data")}
	cts := []*Contract{CreateTestCoinbaseOp("same_contract")}
	prevHash := []byte("determinism_prev_hash")
	height := 1
	timestamp := int64(1609459200)

	block1 := CreateBlockWithTimestamp(txs, cts, prevHash, height, timestamp)
	block2 := CreateBlockWithTimestamp(txs, cts, prevHash, height, timestamp)

	if block1.Timestamp != block2.Timestamp {
		t.Error("Timestamps should be identical for deterministic block creation")
	}

	if block1.Height != block2.Height {
		t.Error("Heights should be identical")
	}
}

func TestCreateBlockWithTimestamp_ZeroTimestamp(t *testing.T) {
	t.Parallel()
	txs := []*Transaction{}
	cts := []*Contract{}
	prevHash := []byte("zero_ts")
	height := 0
	timestamp := int64(0)

	block := CreateBlockWithTimestamp(txs, cts, prevHash, height, timestamp)

	if block.Timestamp != 0 {
		t.Errorf("Expected timestamp 0, got %d", block.Timestamp)
	}
}

func TestCreateBlockWithTimestamp_NegativeTimestamp(t *testing.T) {
	t.Parallel()
	txs := []*Transaction{}
	cts := []*Contract{}
	prevHash := []byte("negative_ts")
	height := 0
	timestamp := int64(-1)

	block := CreateBlockWithTimestamp(txs, cts, prevHash, height, timestamp)

	if block.Timestamp != -1 {
		t.Errorf("Expected timestamp -1, got %d", block.Timestamp)
	}
}

func TestCreateBlockWithTimestamp_FutureTimestamp(t *testing.T) {
	t.Parallel()
	txs := []*Transaction{}
	cts := []*Contract{}
	prevHash := []byte("future_ts")
	height := 0
	futureTime := time.Now().Add(365 * 24 * time.Hour).Unix()

	block := CreateBlockWithTimestamp(txs, cts, prevHash, height, futureTime)

	if block.Timestamp != futureTime {
		t.Errorf("Expected future timestamp %d, got %d", futureTime, block.Timestamp)
	}
}

// =============================================================================
// Genesis Block Tests
// =============================================================================

func TestGenesis(t *testing.T) {
	t.Parallel()
	coinbase := CreateTestCoinbaseTx("genesis_data")
	contractbase := CreateTestCoinbaseOp("genesis_contract")

	genesis := Genesis(coinbase, contractbase)

	if genesis == nil {
		t.Fatal("Genesis should return non-nil block")
	}

	if genesis.Height != 0 {
		t.Errorf("Expected genesis height 0, got %d", genesis.Height)
	}

	if len(genesis.PrevHash) != 0 {
		t.Error("Genesis block should have empty PrevHash")
	}

	if len(genesis.Transactions) != 1 {
		t.Errorf("Expected 1 transaction in genesis, got %d", len(genesis.Transactions))
	}

	if len(genesis.Contracts) != 1 {
		t.Errorf("Expected 1 contract in genesis, got %d", len(genesis.Contracts))
	}

	if len(genesis.Hash) == 0 {
		t.Error("Genesis block should have non-empty hash")
	}

	if genesis.Nonce < 0 {
		t.Error("Genesis block nonce should be non-negative")
	}
}

func TestGenesis_NilCoinbase(t *testing.T) {
	t.Parallel()
	contractbase := CreateTestCoinbaseOp("data")

	defer func() {
		if r := recover(); r != nil {
			t.Log("Genesis with nil coinbase panics as expected")
		}
	}()

	genesis := Genesis(nil, contractbase)
	if genesis != nil {
		if len(genesis.Transactions) > 0 && genesis.Transactions[0] != nil {
			t.Error("Expected nil transaction in genesis with nil coinbase")
		}
	}
}

func TestGenesis_NilContractbase(t *testing.T) {
	t.Parallel()
	coinbase := CreateTestCoinbaseTx("data")

	defer func() {
		if r := recover(); r != nil {
			t.Log("Genesis with nil contractbase panics as expected")
		}
	}()

	genesis := Genesis(coinbase, nil)
	if genesis != nil {
		if len(genesis.Contracts) > 0 && genesis.Contracts[0] != nil {
			t.Error("Expected nil contract in genesis with nil contractbase")
		}
	}
}

// =============================================================================
// Block Hash Tests
// =============================================================================

func TestBlock_HashNotEmpty(t *testing.T) {
	t.Parallel()
	block := CreateBlock([]*Transaction{}, []*Contract{}, []byte{}, 0)

	if len(block.Hash) == 0 {
		t.Error("Block hash should not be empty")
	}
}

func TestBlock_HashLength(t *testing.T) {
	t.Parallel()
	block := CreateBlock([]*Transaction{CreateTestCoinbaseTx("data")}, []*Contract{}, []byte("prev"), 1)

	if len(block.Hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(block.Hash))
	}
}

func TestBlock_DifferentInputsDifferentHashes(t *testing.T) {
	t.Parallel()

	block1 := CreateBlockWithTimestamp(
		[]*Transaction{CreateTestCoinbaseTx("data1")},
		[]*Contract{},
		[]byte("prev1"),
		1,
		1000000,
	)

	block2 := CreateBlockWithTimestamp(
		[]*Transaction{CreateTestCoinbaseTx("data2")},
		[]*Contract{},
		[]byte("prev2"),
		2,
		2000000,
	)

	if bytes.Equal(block1.Hash, block2.Hash) {
		t.Error("Different blocks should have different hashes")
	}
}

// =============================================================================
// Edge Cases and Boundary Tests
// =============================================================================

func TestBlock_VeryLargePrevHash(t *testing.T) {
	t.Parallel()
	largePrevHash := make([]byte, 1024)
	for i := range largePrevHash {
		largePrevHash[i] = byte(i % 256)
	}

	block := CreateBlock([]*Transaction{}, []*Contract{}, largePrevHash, 0)

	if !bytes.Equal(block.PrevHash, largePrevHash) {
		t.Error("Block should preserve large PrevHash")
	}
}

func TestBlock_ManyTransactions(t *testing.T) {
	t.Parallel()
	txCount := 100
	txs := make([]*Transaction, txCount)
	for i := 0; i < txCount; i++ {
		txs[i] = CreateTestCoinbaseTx("")
	}

	block := CreateBlock(txs, []*Contract{}, []byte("prev"), 1)

	if len(block.Transactions) != txCount {
		t.Errorf("Expected %d transactions, got %d", txCount, len(block.Transactions))
	}
}

func TestBlock_ManyContracts(t *testing.T) {
	t.Parallel()
	ctCount := 50
	cts := make([]*Contract, ctCount)
	for i := 0; i < ctCount; i++ {
		cts[i] = CreateTestCoinbaseOp("")
	}

	block := CreateBlock([]*Transaction{}, cts, []byte("prev"), 1)

	if len(block.Contracts) != ctCount {
		t.Errorf("Expected %d contracts, got %d", ctCount, len(block.Contracts))
	}
}

func TestBlock_MaxIntHeight(t *testing.T) {
	t.Parallel()
	maxHeight := int(^uint(0) >> 1)

	block := CreateBlock([]*Transaction{}, []*Contract{}, []byte{}, maxHeight)

	if block.Height != maxHeight {
		t.Errorf("Expected max height %d, got %d", maxHeight, block.Height)
	}
}

// =============================================================================
// HandleCritical Function Tests
// =============================================================================

func TestHandleCritical_NilError(t *testing.T) {
	t.Parallel()
	HandleCritical(nil, "test context")
}

func TestHandleCritical_NonNilError(t *testing.T) {
	t.Parallel()
	HandleCritical(bytes.ErrTooLarge, "test context")
}
