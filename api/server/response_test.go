// Package server provides unit tests for response transformation functions
// that convert blockchain data structures to JSON-serializable formats.
package server

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/codila125/foedus-blockchain/blockchain"
)

// =============================================================================
// BlockResponse Tests
// =============================================================================

func TestBlockResponse_BasicConversion(t *testing.T) {
	block := &blockchain.Block{
		Hash:         []byte{0xab, 0xcd, 0xef},
		PrevHash:     []byte{0x12, 0x34, 0x56},
		Contracts:    []*blockchain.Contract{},
		Transactions: []*blockchain.Transaction{},
	}

	res := BlockResponse(block)

	expectedHash := hex.EncodeToString(block.Hash)
	if res.Hash != expectedHash {
		t.Errorf("Expected hash '%s', got '%s'", expectedHash, res.Hash)
	}

	expectedPrevHash := hex.EncodeToString(block.PrevHash)
	if res.PrevHash != expectedPrevHash {
		t.Errorf("Expected prevHash '%s', got '%s'", expectedPrevHash, res.PrevHash)
	}
}

func TestBlockResponse_EmptyHashes(t *testing.T) {
	block := &blockchain.Block{
		Hash:         []byte{},
		PrevHash:     []byte{},
		Contracts:    []*blockchain.Contract{},
		Transactions: []*blockchain.Transaction{},
	}

	res := BlockResponse(block)

	if res.Hash != "" {
		t.Errorf("Expected empty hash, got '%s'", res.Hash)
	}

	if res.PrevHash != "" {
		t.Errorf("Expected empty prevHash, got '%s'", res.PrevHash)
	}
}

func TestBlockResponse_GenesisBlock(t *testing.T) {
	genesisBlock := &blockchain.Block{
		Hash:         []byte{0x00, 0x01, 0x02},
		PrevHash:     []byte{},
		Contracts:    []*blockchain.Contract{},
		Transactions: []*blockchain.Transaction{},
	}

	res := BlockResponse(genesisBlock)

	if res.PrevHash != "" {
		t.Error("Genesis block should have empty PrevHash")
	}
}

// =============================================================================
// BlockContractResponse Tests
// =============================================================================

func TestBlockContractResponse_SingleContract(t *testing.T) {
	now := time.Now().Unix()
	contracts := []*blockchain.Contract{
		{
			ID:        []byte{0xab, 0xcd},
			Title:     "Test Contract",
			CreatedAt: now,
			UpdatedAt: now,
			Status:    blockchain.ContractActive,
		},
	}

	res := BlockContractResponse(contracts)

	if len(res) != 1 {
		t.Fatalf("Expected 1 contract, got %d", len(res))
	}

	if res[0].Title != "Test Contract" {
		t.Errorf("Expected title 'Test Contract', got '%s'", res[0].Title)
	}

	if res[0].Status != blockchain.ContractActive {
		t.Errorf("Expected status ACTIVE, got %s", res[0].Status)
	}
}

func TestBlockContractResponse_MultipleContracts(t *testing.T) {
	now := time.Now().Unix()
	contracts := []*blockchain.Contract{
		{ID: []byte{0x01}, Title: "Contract 1", CreatedAt: now, Status: blockchain.ContractDraft},
		{ID: []byte{0x02}, Title: "Contract 2", CreatedAt: now, Status: blockchain.ContractActive},
		{ID: []byte{0x03}, Title: "Contract 3", CreatedAt: now, Status: blockchain.ContractCompleted},
	}

	res := BlockContractResponse(contracts)

	if len(res) != 3 {
		t.Fatalf("Expected 3 contracts, got %d", len(res))
	}

	expectedTitles := []string{"Contract 1", "Contract 2", "Contract 3"}
	for i, contract := range res {
		if contract.Title != expectedTitles[i] {
			t.Errorf("Contract %d: expected title '%s', got '%s'", i, expectedTitles[i], contract.Title)
		}
	}
}

func TestBlockContractResponse_EmptyContracts(t *testing.T) {
	contracts := []*blockchain.Contract{}

	res := BlockContractResponse(contracts)

	if len(res) != 0 {
		t.Errorf("Expected empty/nil slice, got %d items", len(res))
	}
}

func TestBlockContractResponse_NilContracts(t *testing.T) {
	var contracts []*blockchain.Contract = nil

	res := BlockContractResponse(contracts)

	if len(res) != 0 {
		t.Errorf("Expected nil/empty slice for nil input")
	}
}

func TestBlockContractResponse_TimestampConversion(t *testing.T) {
	timestamp := int64(1700000000)
	contracts := []*blockchain.Contract{
		{
			ID:        []byte{0x01},
			Title:     "Timestamp Test",
			CreatedAt: timestamp,
			UpdatedAt: timestamp + 3600,
		},
	}

	res := BlockContractResponse(contracts)

	expectedCreatedAt := time.Unix(timestamp, 0)
	if !res[0].CreatedAt.Equal(expectedCreatedAt) {
		t.Errorf("Expected createdAt %v, got %v", expectedCreatedAt, res[0].CreatedAt)
	}

	expectedUpdatedAt := time.Unix(timestamp+3600, 0)
	if !res[0].UpdatedAt.Equal(expectedUpdatedAt) {
		t.Errorf("Expected updatedAt %v, got %v", expectedUpdatedAt, res[0].UpdatedAt)
	}
}

// =============================================================================
// BlockTransactionResponse Tests
// =============================================================================

func TestBlockTransactionResponse_SingleTransaction(t *testing.T) {
	transactions := []*blockchain.Transaction{
		{
			ID:      []byte{0xab, 0xcd, 0xef},
			Inputs:  []blockchain.TxInput{},
			Outputs: []blockchain.TxOutput{},
		},
	}

	res := BlockTransactionResponse(transactions)

	if len(res) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(res))
	}

	expectedID := hex.EncodeToString(transactions[0].ID)
	if res[0].ID != expectedID {
		t.Errorf("Expected ID '%s', got '%s'", expectedID, res[0].ID)
	}
}

func TestBlockTransactionResponse_MultipleTransactions(t *testing.T) {
	transactions := []*blockchain.Transaction{
		{ID: []byte{0x01}, Inputs: []blockchain.TxInput{}, Outputs: []blockchain.TxOutput{}},
		{ID: []byte{0x02}, Inputs: []blockchain.TxInput{}, Outputs: []blockchain.TxOutput{}},
		{ID: []byte{0x03}, Inputs: []blockchain.TxInput{}, Outputs: []blockchain.TxOutput{}},
	}

	res := BlockTransactionResponse(transactions)

	if len(res) != 3 {
		t.Fatalf("Expected 3 transactions, got %d", len(res))
	}
}

func TestBlockTransactionResponse_Empty(t *testing.T) {
	transactions := []*blockchain.Transaction{}

	res := BlockTransactionResponse(transactions)

	if len(res) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(res))
	}
}

// =============================================================================
// BlockTransactionInputResponse Tests
// =============================================================================

func TestBlockTransactionInputResponse_SingleInput(t *testing.T) {
	inputs := []blockchain.TxInput{
		{ID: []byte{0xab, 0xcd}, Out: 0},
	}

	res := BlockTransactionInputResponse(inputs)

	if len(res) != 1 {
		t.Fatalf("Expected 1 input, got %d", len(res))
	}

	expectedFrom := hex.EncodeToString(inputs[0].ID)
	if res[0].From != expectedFrom {
		t.Errorf("Expected From '%s', got '%s'", expectedFrom, res[0].From)
	}

	if res[0].Out != 0 {
		t.Errorf("Expected Out 0, got %d", res[0].Out)
	}
}

func TestBlockTransactionInputResponse_MultipleInputs(t *testing.T) {
	inputs := []blockchain.TxInput{
		{ID: []byte{0x01}, Out: 0},
		{ID: []byte{0x02}, Out: 1},
		{ID: []byte{0x03}, Out: 2},
	}

	res := BlockTransactionInputResponse(inputs)

	if len(res) != 3 {
		t.Fatalf("Expected 3 inputs, got %d", len(res))
	}

	for i, input := range res {
		if input.Out != i {
			t.Errorf("Input %d: expected Out %d, got %d", i, i, input.Out)
		}
	}
}

func TestBlockTransactionInputResponse_Empty(t *testing.T) {
	inputs := []blockchain.TxInput{}

	res := BlockTransactionInputResponse(inputs)

	if len(res) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(res))
	}
}

// =============================================================================
// BlockTransactionOutputResponse Tests
// =============================================================================

func TestBlockTransactionOutputResponse_SingleOutput(t *testing.T) {
	outputs := []blockchain.TxOutput{
		{PubKeyHash: []byte{0xab, 0xcd}, Value: 100},
	}

	res := BlockTransactionOutputResponse(outputs)

	if len(res) != 1 {
		t.Fatalf("Expected 1 output, got %d", len(res))
	}

	expectedTo := hex.EncodeToString(outputs[0].PubKeyHash)
	if res[0].To != expectedTo {
		t.Errorf("Expected To '%s', got '%s'", expectedTo, res[0].To)
	}

	if res[0].Value != 100 {
		t.Errorf("Expected Value 100, got %d", res[0].Value)
	}
}

func TestBlockTransactionOutputResponse_MultipleOutputs(t *testing.T) {
	outputs := []blockchain.TxOutput{
		{PubKeyHash: []byte{0x01}, Value: 50},
		{PubKeyHash: []byte{0x02}, Value: 75},
		{PubKeyHash: []byte{0x03}, Value: 125},
	}

	res := BlockTransactionOutputResponse(outputs)

	if len(res) != 3 {
		t.Fatalf("Expected 3 outputs, got %d", len(res))
	}

	expectedValues := []int{50, 75, 125}
	for i, output := range res {
		if output.Value != expectedValues[i] {
			t.Errorf("Output %d: expected Value %d, got %d", i, expectedValues[i], output.Value)
		}
	}
}

func TestBlockTransactionOutputResponse_ZeroValue(t *testing.T) {
	outputs := []blockchain.TxOutput{
		{PubKeyHash: []byte{0x01}, Value: 0},
	}

	res := BlockTransactionOutputResponse(outputs)

	if res[0].Value != 0 {
		t.Errorf("Expected Value 0, got %d", res[0].Value)
	}
}

// =============================================================================
// ContractResponse Tests
// =============================================================================

func TestContractResponse_FullContract(t *testing.T) {
	now := time.Now().Unix()
	contract := blockchain.Contract{
		ID:             []byte{0xab, 0xcd, 0xef},
		Title:          "Full Contract",
		Description:    "A complete contract",
		CreatorAddress: "creator_addr",
		CreatedAt:      now,
		UpdatedAt:      now,
		Status:         blockchain.ContractActive,
		Milestones:     []*blockchain.Milestone{},
		Parties:        []*blockchain.Party{},
		Terms:          []byte("Contract terms"),
		Attachments:    [][]byte{},
	}

	res := ContractResponse(contract)

	expectedID := hex.EncodeToString(contract.ID)
	if res.ID != expectedID {
		t.Errorf("Expected ID '%s', got '%s'", expectedID, res.ID)
	}

	if res.Title != "Full Contract" {
		t.Errorf("Expected title 'Full Contract', got '%s'", res.Title)
	}

	if res.Creator != "creator_addr" {
		t.Errorf("Expected creator 'creator_addr', got '%s'", res.Creator)
	}

	if res.Terms != "Contract terms" {
		t.Errorf("Expected terms 'Contract terms', got '%s'", res.Terms)
	}
}

func TestContractResponse_WithMilestones(t *testing.T) {
	now := time.Now().Unix()
	contract := blockchain.Contract{
		ID:    []byte{0x01},
		Title: "Contract with Milestones",
		Milestones: []*blockchain.Milestone{
			{ID: []byte{0x01}, Title: "Milestone 1", Value: 100, CreatedAt: now},
			{ID: []byte{0x02}, Title: "Milestone 2", Value: 200, CreatedAt: now},
		},
		Parties:     []*blockchain.Party{},
		CreatedAt:   now,
		UpdatedAt:   now,
		Attachments: [][]byte{},
	}

	res := ContractResponse(contract)

	if len(res.Milestones) != 2 {
		t.Fatalf("Expected 2 milestones, got %d", len(res.Milestones))
	}
}

func TestContractResponse_WithParties(t *testing.T) {
	now := time.Now().Unix()
	contract := blockchain.Contract{
		ID:    []byte{0x01},
		Title: "Contract with Parties",
		Parties: []*blockchain.Party{
			{Address: "addr1", Role: blockchain.RoleContractor, PublicKey: []byte{0x01}},
			{Address: "addr2", Role: blockchain.RoleArbitrator, PublicKey: []byte{0x02}},
		},
		Milestones:  []*blockchain.Milestone{},
		CreatedAt:   now,
		UpdatedAt:   now,
		Attachments: [][]byte{},
	}

	res := ContractResponse(contract)

	if len(res.Parties) != 2 {
		t.Fatalf("Expected 2 parties, got %d", len(res.Parties))
	}
}

func TestContractResponse_WithAttachments(t *testing.T) {
	now := time.Now().Unix()
	contract := blockchain.Contract{
		ID:         []byte{0x01},
		Title:      "Contract with Attachments",
		Milestones: []*blockchain.Milestone{},
		Parties:    []*blockchain.Party{},
		CreatedAt:  now,
		UpdatedAt:  now,
		Attachments: [][]byte{
			[]byte("doc1.pdf"),
			[]byte("doc2.pdf"),
			[]byte("https://example.com/file"),
		},
	}

	res := ContractResponse(contract)

	if len(res.Attachments) != 3 {
		t.Fatalf("Expected 3 attachments, got %d", len(res.Attachments))
	}
}

// =============================================================================
// ContractMilestoneResponse Tests
// =============================================================================

func TestContractMilestoneResponse_SingleMilestone(t *testing.T) {
	now := time.Now().Unix()
	milestones := []*blockchain.Milestone{
		{
			ID:          []byte{0xab, 0xcd},
			Title:       "Test Milestone",
			Description: "Description",
			Value:       1000,
			DueDate:     now + 86400,
			Status:      blockchain.MilestoneActive,
			CreatedAt:   now,
			CompletedAt: 0,
			Evidence:    []byte("evidence"),
			ApprovedBy:  []string{"addr1"},
		},
	}

	res := ContractMilestoneResponse(milestones)

	if len(res) != 1 {
		t.Fatalf("Expected 1 milestone, got %d", len(res))
	}

	if res[0].Title != "Test Milestone" {
		t.Errorf("Expected title 'Test Milestone', got '%s'", res[0].Title)
	}

	if res[0].Value != 1000 {
		t.Errorf("Expected value 1000, got %d", res[0].Value)
	}

	if res[0].Status != blockchain.MilestoneActive {
		t.Errorf("Expected status ACTIVE, got %s", res[0].Status)
	}
}

func TestContractMilestoneResponse_Empty(t *testing.T) {
	milestones := []*blockchain.Milestone{}

	res := ContractMilestoneResponse(milestones)

	if len(res) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(res))
	}
}

func TestContractMilestoneResponse_AllStatuses(t *testing.T) {
	now := time.Now().Unix()
	milestones := []*blockchain.Milestone{
		{ID: []byte{0x01}, Title: "Active", Status: blockchain.MilestoneActive, CreatedAt: now},
		{ID: []byte{0x02}, Title: "Completed", Status: blockchain.MilestoneCompleted, CreatedAt: now},
		{ID: []byte{0x03}, Title: "Cancelled", Status: blockchain.MilestoneCancelled, CreatedAt: now},
	}

	res := ContractMilestoneResponse(milestones)

	expectedStatuses := []blockchain.MilestoneStatus{
		blockchain.MilestoneActive,
		blockchain.MilestoneCompleted,
		blockchain.MilestoneCancelled,
	}

	for i, milestone := range res {
		if milestone.Status != expectedStatuses[i] {
			t.Errorf("Milestone %d: expected status %s, got %s", i, expectedStatuses[i], milestone.Status)
		}
	}
}

// =============================================================================
// ContractPartyResponse Tests
// =============================================================================

func TestContractPartyResponse_SingleParty(t *testing.T) {
	parties := []*blockchain.Party{
		{
			Address:   "party_address",
			Role:      blockchain.RoleContractor,
			PublicKey: []byte{0xab, 0xcd, 0xef},
		},
	}

	res := ContractPartyResponse(parties)

	if len(res) != 1 {
		t.Fatalf("Expected 1 party, got %d", len(res))
	}

	if res[0].Address != "party_address" {
		t.Errorf("Expected address 'party_address', got '%s'", res[0].Address)
	}

	if res[0].Role != blockchain.RoleContractor {
		t.Errorf("Expected role CONTRACTOR, got %s", res[0].Role)
	}

	expectedPubKey := hex.EncodeToString(parties[0].PublicKey)
	if res[0].PublicKey != expectedPubKey {
		t.Errorf("Expected publicKey '%s', got '%s'", expectedPubKey, res[0].PublicKey)
	}
}

func TestContractPartyResponse_AllRoles(t *testing.T) {
	parties := []*blockchain.Party{
		{Address: "addr1", Role: blockchain.RoleContractor, PublicKey: []byte{0x01}},
		{Address: "addr2", Role: blockchain.RoleArbitrator, PublicKey: []byte{0x02}},
		{Address: "addr3", Role: blockchain.RoleCreator, PublicKey: []byte{0x03}},
	}

	res := ContractPartyResponse(parties)

	if len(res) != 3 {
		t.Fatalf("Expected 3 parties, got %d", len(res))
	}

	expectedRoles := []blockchain.ContractRole{
		blockchain.RoleContractor,
		blockchain.RoleArbitrator,
		blockchain.RoleCreator,
	}

	for i, party := range res {
		if party.Role != expectedRoles[i] {
			t.Errorf("Party %d: expected role %s, got %s", i, expectedRoles[i], party.Role)
		}
	}
}

func TestContractPartyResponse_Empty(t *testing.T) {
	parties := []*blockchain.Party{}

	res := ContractPartyResponse(parties)

	if len(res) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(res))
	}
}

// =============================================================================
// ContractAttachmentResponse Tests
// =============================================================================

func TestContractAttachmentResponse_SingleAttachment(t *testing.T) {
	attachments := [][]byte{
		[]byte("document.pdf"),
	}

	res := ContractAttachmentResponse(attachments)

	if len(res) != 1 {
		t.Fatalf("Expected 1 attachment, got %d", len(res))
	}

	if res[0] != "document.pdf" {
		t.Errorf("Expected attachment 'document.pdf', got '%s'", res[0])
	}
}

func TestContractAttachmentResponse_MultipleAttachments(t *testing.T) {
	attachments := [][]byte{
		[]byte("doc1.pdf"),
		[]byte("doc2.pdf"),
		[]byte("https://example.com/file"),
	}

	res := ContractAttachmentResponse(attachments)

	if len(res) != 3 {
		t.Fatalf("Expected 3 attachments, got %d", len(res))
	}

	expectedAttachments := []string{"doc1.pdf", "doc2.pdf", "https://example.com/file"}
	for i, att := range res {
		if att != expectedAttachments[i] {
			t.Errorf("Attachment %d: expected '%s', got '%s'", i, expectedAttachments[i], att)
		}
	}
}

func TestContractAttachmentResponse_Empty(t *testing.T) {
	attachments := [][]byte{}

	res := ContractAttachmentResponse(attachments)

	if len(res) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(res))
	}
}

func TestContractAttachmentResponse_UnicodeContent(t *testing.T) {
	attachments := [][]byte{
		[]byte("文档.pdf"),
		[]byte("dokument_αβγ.pdf"),
	}

	res := ContractAttachmentResponse(attachments)

	if res[0] != "文档.pdf" {
		t.Errorf("Unicode attachment mismatch: expected '文档.pdf', got '%s'", res[0])
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestEdgeCase_LargeIDBytes(t *testing.T) {
	largeID := make([]byte, 1000)
	for i := range largeID {
		largeID[i] = byte(i % 256)
	}

	hexStr := hex.EncodeToString(largeID)

	if len(hexStr) != 2000 {
		t.Errorf("Expected hex string length 2000, got %d", len(hexStr))
	}
}

func TestEdgeCase_EmptyPubKeyHash(t *testing.T) {
	outputs := []blockchain.TxOutput{
		{PubKeyHash: []byte{}, Value: 100},
	}

	res := BlockTransactionOutputResponse(outputs)

	if res[0].To != "" {
		t.Errorf("Expected empty 'To' for empty PubKeyHash, got '%s'", res[0].To)
	}
}

func TestEdgeCase_NegativeTimestamp(t *testing.T) {
	negativeTimestamp := int64(-86400)
	tm := time.Unix(negativeTimestamp, 0)

	if tm.Unix() != negativeTimestamp {
		t.Errorf("Expected timestamp %d, got %d", negativeTimestamp, tm.Unix())
	}
}

func TestEdgeCase_ZeroTimestamp(t *testing.T) {
	zeroTimestamp := int64(0)
	tm := time.Unix(zeroTimestamp, 0)

	expectedTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if !tm.UTC().Equal(expectedTime) {
		t.Errorf("Zero timestamp should equal Unix epoch")
	}
}

func TestEdgeCase_LargeValue(t *testing.T) {
	outputs := []blockchain.TxOutput{
		{PubKeyHash: []byte{0x01}, Value: 2147483647},
	}

	res := BlockTransactionOutputResponse(outputs)

	if res[0].Value != 2147483647 {
		t.Errorf("Expected max int32 value, got %d", res[0].Value)
	}
}

// =============================================================================
// Hex Encoding Consistency Tests
// =============================================================================

func TestHexEncoding_ConsistentOutput(t *testing.T) {
	testBytes := []byte{0xab, 0xcd, 0xef, 0x12, 0x34}

	encoded1 := hex.EncodeToString(testBytes)
	encoded2 := hex.EncodeToString(testBytes)

	if encoded1 != encoded2 {
		t.Error("Hex encoding should be consistent")
	}
}

func TestHexEncoding_LowerCase(t *testing.T) {
	testBytes := []byte{0xAB, 0xCD, 0xEF}
	encoded := hex.EncodeToString(testBytes)

	if encoded != "abcdef" {
		t.Errorf("Expected lowercase hex 'abcdef', got '%s'", encoded)
	}
}
