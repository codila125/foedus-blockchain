package blockchain

import (
	"bytes"
	"testing"
)

// =============================================================================
// Block HashTransactions Tests
// =============================================================================

func TestBlock_HashTransactions_WithTransactions(t *testing.T) {
	t.Parallel()

	txs := []*Transaction{
		CreateTestCoinbaseTx("data1"),
		CreateTestCoinbaseTx("data2"),
		CreateTestCoinbaseTx("data3"),
	}

	block := &Block{
		Transactions: txs,
	}

	hash := block.HashTransactions()

	if len(hash) == 0 {
		t.Error("HashTransactions should return non-empty hash for non-empty transactions")
	}
}

func TestBlock_HashTransactions_Empty(t *testing.T) {
	t.Parallel()

	block := &Block{
		Transactions: []*Transaction{},
	}

	hash := block.HashTransactions()

	if len(hash) != 0 {
		t.Error("HashTransactions should return empty slice for empty transactions")
	}
}

func TestBlock_HashTransactions_Nil(t *testing.T) {
	t.Parallel()

	block := &Block{
		Transactions: nil,
	}

	hash := block.HashTransactions()

	if len(hash) != 0 {
		t.Error("HashTransactions should return empty slice for nil transactions")
	}
}

func TestBlock_HashTransactions_Deterministic(t *testing.T) {
	t.Parallel()

	txs := []*Transaction{
		CreateTestCoinbaseTx("data"),
	}

	block := &Block{Transactions: txs}

	hash1 := block.HashTransactions()
	hash2 := block.HashTransactions()

	if !bytes.Equal(hash1, hash2) {
		t.Error("HashTransactions should be deterministic")
	}
}

func TestBlock_HashTransactions_DifferentTxsDifferentHashes(t *testing.T) {
	t.Parallel()

	block1 := &Block{
		Transactions: []*Transaction{CreateTestCoinbaseTx("data1")},
	}

	block2 := &Block{
		Transactions: []*Transaction{CreateTestCoinbaseTx("data2")},
	}

	hash1 := block1.HashTransactions()
	hash2 := block2.HashTransactions()

	if bytes.Equal(hash1, hash2) {
		t.Error("Different transactions should produce different hashes")
	}
}

func TestBlock_HashTransactions_SingleTransaction(t *testing.T) {
	t.Parallel()

	block := &Block{
		Transactions: []*Transaction{CreateTestCoinbaseTx("data")},
	}

	hash := block.HashTransactions()

	// Merkle tree with single leaf should still produce valid hash
	if len(hash) == 0 {
		t.Error("Single transaction should produce valid hash")
	}
}

func TestBlock_HashTransactions_ManyTransactions(t *testing.T) {
	t.Parallel()

	txCount := 100
	txs := make([]*Transaction, txCount)
	for i := 0; i < txCount; i++ {
		txs[i] = CreateTestCoinbaseTx("")
	}

	block := &Block{Transactions: txs}

	hash := block.HashTransactions()

	if len(hash) == 0 {
		t.Error("Many transactions should produce valid hash")
	}
}

// =============================================================================
// Block HashContracts Tests
// =============================================================================

func TestBlock_HashContracts_WithContracts(t *testing.T) {
	t.Parallel()

	cts := []*Contract{
		CoinbaseOp("addr1", "data1"),
		CoinbaseOp("addr2", "data2"),
	}

	block := &Block{Contracts: cts}

	hash := block.HashContracts()

	if len(hash) == 0 {
		t.Error("HashContracts should return non-empty hash for non-empty contracts")
	}
}

func TestBlock_HashContracts_Empty(t *testing.T) {
	t.Parallel()

	block := &Block{Contracts: []*Contract{}}

	hash := block.HashContracts()

	if len(hash) != 0 {
		t.Error("HashContracts should return empty slice for empty contracts")
	}
}

func TestBlock_HashContracts_Nil(t *testing.T) {
	t.Parallel()

	block := &Block{Contracts: nil}

	hash := block.HashContracts()

	if len(hash) != 0 {
		t.Error("HashContracts should return empty slice for nil contracts")
	}
}

func TestBlock_HashContracts_Deterministic(t *testing.T) {
	t.Parallel()

	cts := []*Contract{CoinbaseOp("addr", "data")}

	block := &Block{Contracts: cts}

	hash1 := block.HashContracts()
	hash2 := block.HashContracts()

	if !bytes.Equal(hash1, hash2) {
		t.Error("HashContracts should be deterministic")
	}
}

func TestBlock_HashContracts_DifferentContractsDifferentHashes(t *testing.T) {
	t.Parallel()

	block1 := &Block{Contracts: []*Contract{CoinbaseOp("addr1", "data1")}}
	block2 := &Block{Contracts: []*Contract{CoinbaseOp("addr2", "data2")}}

	hash1 := block1.HashContracts()
	hash2 := block2.HashContracts()

	if bytes.Equal(hash1, hash2) {
		t.Error("Different contracts should produce different hashes")
	}
}

func TestBlock_HashContracts_ManyContracts(t *testing.T) {
	t.Parallel()

	ctCount := 50
	cts := make([]*Contract, ctCount)
	for i := 0; i < ctCount; i++ {
		cts[i] = CoinbaseOp("addr", "")
	}

	block := &Block{Contracts: cts}

	hash := block.HashContracts()

	if len(hash) == 0 {
		t.Error("Many contracts should produce valid hash")
	}
}

// =============================================================================
// Contract HashContract Tests
// =============================================================================

func TestContract_HashContract(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		Title:          "Test Contract",
		Description:    "Test Description",
		CreatedAt:      1609459200,
		CreatorAddress: "creator_address",
		Milestones:     []*Milestone{},
		Parties:        []*Party{},
		Terms:          []byte("terms"),
		Attachments:    [][]byte{},
	}

	hash := contract.HashContract()

	if len(hash) != 32 { // SHA-256
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

func TestContract_HashContract_Deterministic(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		Title:          "Determinism Test",
		Description:    "Testing determinism",
		CreatedAt:      1000,
		CreatorAddress: "creator",
		Milestones:     []*Milestone{},
		Parties:        []*Party{},
		Terms:          []byte{},
		Attachments:    [][]byte{},
	}

	hash1 := contract.HashContract()
	hash2 := contract.HashContract()

	if !bytes.Equal(hash1, hash2) {
		t.Error("HashContract should be deterministic")
	}
}

func TestContract_HashContract_DifferentContractsDifferentHashes(t *testing.T) {
	t.Parallel()

	contract1 := &Contract{
		Title:          "Contract 1",
		Description:    "First contract",
		CreatedAt:      1000,
		CreatorAddress: "creator1",
		Milestones:     []*Milestone{},
		Parties:        []*Party{},
	}

	contract2 := &Contract{
		Title:          "Contract 2",
		Description:    "Second contract",
		CreatedAt:      2000,
		CreatorAddress: "creator2",
		Milestones:     []*Milestone{},
		Parties:        []*Party{},
	}

	hash1 := contract1.HashContract()
	hash2 := contract2.HashContract()

	if bytes.Equal(hash1, hash2) {
		t.Error("Different contracts should produce different hashes")
	}
}

func TestContract_HashContract_ExcludesMutableFields(t *testing.T) {
	t.Parallel()

	// Create two contracts with same immutable fields but different mutable fields
	contract1 := &Contract{
		Title:          "Same Title",
		Description:    "Same Description",
		CreatedAt:      1000,
		UpdatedAt:      2000, // Mutable - should be excluded
		Status:         ContractDraft,
		CreatorAddress: "creator",
		Milestones:     []*Milestone{},
		Parties:        []*Party{},
		Terms:          []byte{},
		Attachments:    [][]byte{},
	}

	contract2 := &Contract{
		Title:          "Same Title",
		Description:    "Same Description",
		CreatedAt:      1000,
		UpdatedAt:      3000, // Different UpdatedAt
		Status:         ContractActive,
		CreatorAddress: "creator",
		Milestones:     []*Milestone{},
		Parties:        []*Party{},
		Terms:          []byte{},
		Attachments:    [][]byte{},
	}

	hash1 := contract1.HashContract()
	hash2 := contract2.HashContract()

	// Hashes should be equal since only immutable fields are used
	if !bytes.Equal(hash1, hash2) {
		t.Error("Hash should only depend on immutable fields")
	}
}

func TestContract_HashContract_WithMilestones(t *testing.T) {
	t.Parallel()

	milestones := []*Milestone{
		{
			Title:       "Milestone 1",
			Description: "First milestone",
			Value:       100,
			CreatedAt:   1000,
		},
		{
			Title:       "Milestone 2",
			Description: "Second milestone",
			Value:       200,
			CreatedAt:   2000,
		},
	}

	contract := &Contract{
		Title:          "Contract with Milestones",
		Description:    "Has milestones",
		CreatedAt:      1000,
		CreatorAddress: "creator",
		Milestones:     milestones,
		Parties:        []*Party{},
	}

	hash := contract.HashContract()

	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

func TestContract_HashContract_WithParties(t *testing.T) {
	t.Parallel()

	parties := []*Party{
		{
			Address:   "addr1",
			Role:      RoleCreator,
			PublicKey: []byte("pk1"),
		},
		{
			Address:   "addr2",
			Role:      RoleContractor,
			PublicKey: []byte("pk2"),
		},
	}

	contract := &Contract{
		Title:          "Contract with Parties",
		Description:    "Has parties",
		CreatedAt:      1000,
		CreatorAddress: "creator",
		Milestones:     []*Milestone{},
		Parties:        parties,
	}

	hash := contract.HashContract()

	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

func TestContract_HashContract_WithAttachments(t *testing.T) {
	t.Parallel()

	attachments := [][]byte{
		[]byte("attachment1_hash"),
		[]byte("attachment2_hash"),
	}

	contract := &Contract{
		Title:          "Contract with Attachments",
		Description:    "Has attachments",
		CreatedAt:      1000,
		CreatorAddress: "creator",
		Milestones:     []*Milestone{},
		Parties:        []*Party{},
		Terms:          []byte{},
		Attachments:    attachments,
	}

	hash := contract.HashContract()

	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

func TestContract_HashContract_EmptyContract(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		Title:          "",
		Description:    "",
		CreatedAt:      0,
		CreatorAddress: "",
		Milestones:     []*Milestone{},
		Parties:        []*Party{},
	}

	hash := contract.HashContract()

	// Empty contract should still produce valid hash
	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

// =============================================================================
// Milestone HashMilestones Tests
// =============================================================================

func TestMilestone_HashMilestones(t *testing.T) {
	t.Parallel()

	milestone := &Milestone{
		Title:       "Test Milestone",
		Description: "Test Description",
		Value:       1000,
		CreatedAt:   1609459200,
	}

	hash := milestone.HashMilestones()

	if len(hash) != 32 { // SHA-256
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

func TestMilestone_HashMilestones_Deterministic(t *testing.T) {
	t.Parallel()

	milestone := &Milestone{
		Title:       "Determinism Test",
		Description: "Testing determinism",
		Value:       500,
		CreatedAt:   1000,
	}

	hash1 := milestone.HashMilestones()
	hash2 := milestone.HashMilestones()

	if !bytes.Equal(hash1, hash2) {
		t.Error("HashMilestones should be deterministic")
	}
}

func TestMilestone_HashMilestones_DifferentMilestonesDifferentHashes(t *testing.T) {
	t.Parallel()

	milestone1 := &Milestone{
		Title:       "Milestone 1",
		Description: "First",
		Value:       100,
		CreatedAt:   1000,
	}

	milestone2 := &Milestone{
		Title:       "Milestone 2",
		Description: "Second",
		Value:       200,
		CreatedAt:   2000,
	}

	hash1 := milestone1.HashMilestones()
	hash2 := milestone2.HashMilestones()

	if bytes.Equal(hash1, hash2) {
		t.Error("Different milestones should produce different hashes")
	}
}

func TestMilestone_HashMilestones_ExcludesMutableFields(t *testing.T) {
	t.Parallel()

	milestone1 := &Milestone{
		Title:       "Same Title",
		Description: "Same Description",
		Value:       100,
		CreatedAt:   1000,
		Status:      MilestoneActive, // Mutable
		CompletedAt: 0,               // Mutable
	}

	milestone2 := &Milestone{
		Title:       "Same Title",
		Description: "Same Description",
		Value:       100,
		CreatedAt:   1000,
		Status:      MilestoneCompleted, // Different status
		CompletedAt: 2000,               // Different completed time
	}

	hash1 := milestone1.HashMilestones()
	hash2 := milestone2.HashMilestones()

	// Hashes should be equal since only immutable fields are used
	if !bytes.Equal(hash1, hash2) {
		t.Error("Milestone hash should only depend on immutable fields")
	}
}

func TestMilestone_HashMilestones_ZeroValue(t *testing.T) {
	t.Parallel()

	milestone := &Milestone{
		Title:       "Zero Value Milestone",
		Description: "Has zero value",
		Value:       0,
		CreatedAt:   1000,
	}

	hash := milestone.HashMilestones()

	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

func TestMilestone_HashMilestones_EmptyMilestone(t *testing.T) {
	t.Parallel()

	milestone := &Milestone{
		Title:       "",
		Description: "",
		Value:       0,
		CreatedAt:   0,
	}

	hash := milestone.HashMilestones()

	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

// =============================================================================
// Edge Cases for Hashing
// =============================================================================

func TestHash_LargeData(t *testing.T) {
	t.Parallel()

	// Create contract with large description
	largeDescription := make([]byte, 10000)
	for i := range largeDescription {
		largeDescription[i] = byte(i % 256)
	}

	contract := &Contract{
		Title:          "Large Contract",
		Description:    string(largeDescription),
		CreatedAt:      1000,
		CreatorAddress: "creator",
		Milestones:     []*Milestone{},
		Parties:        []*Party{},
	}

	hash := contract.HashContract()

	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

func TestHash_UnicodeData(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		Title:          "合同标题 🎉",
		Description:    "Description with émojis 🚀 and spëcial châràctérs",
		CreatedAt:      1000,
		CreatorAddress: "creator_with_unicode_💰",
		Milestones:     []*Milestone{},
		Parties:        []*Party{},
	}

	hash := contract.HashContract()

	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

func TestHash_OrderOfMilestones(t *testing.T) {
	t.Parallel()

	milestones1 := []*Milestone{
		{Title: "A", Description: "First", Value: 100, CreatedAt: 1000},
		{Title: "B", Description: "Second", Value: 200, CreatedAt: 2000},
	}

	milestones2 := []*Milestone{
		{Title: "B", Description: "Second", Value: 200, CreatedAt: 2000},
		{Title: "A", Description: "First", Value: 100, CreatedAt: 1000},
	}

	contract1 := &Contract{
		Title:          "Contract",
		Description:    "Desc",
		CreatedAt:      1000,
		CreatorAddress: "creator",
		Milestones:     milestones1,
		Parties:        []*Party{},
	}

	contract2 := &Contract{
		Title:          "Contract",
		Description:    "Desc",
		CreatedAt:      1000,
		CreatorAddress: "creator",
		Milestones:     milestones2,
		Parties:        []*Party{},
	}

	hash1 := contract1.HashContract()
	hash2 := contract2.HashContract()

	// Order of milestones should affect the hash
	if bytes.Equal(hash1, hash2) {
		t.Error("Different milestone order should produce different hashes")
	}
}

func TestHash_OrderOfParties(t *testing.T) {
	t.Parallel()

	parties1 := []*Party{
		{Address: "addr1", Role: RoleCreator, PublicKey: []byte("pk1")},
		{Address: "addr2", Role: RoleContractor, PublicKey: []byte("pk2")},
	}

	parties2 := []*Party{
		{Address: "addr2", Role: RoleContractor, PublicKey: []byte("pk2")},
		{Address: "addr1", Role: RoleCreator, PublicKey: []byte("pk1")},
	}

	contract1 := &Contract{
		Title:          "Contract",
		Description:    "Desc",
		CreatedAt:      1000,
		CreatorAddress: "creator",
		Milestones:     []*Milestone{},
		Parties:        parties1,
	}

	contract2 := &Contract{
		Title:          "Contract",
		Description:    "Desc",
		CreatedAt:      1000,
		CreatorAddress: "creator",
		Milestones:     []*Milestone{},
		Parties:        parties2,
	}

	hash1 := contract1.HashContract()
	hash2 := contract2.HashContract()

	// Order of parties should affect the hash
	if bytes.Equal(hash1, hash2) {
		t.Error("Different party order should produce different hashes")
	}
}
