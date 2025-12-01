// Package server provides unit tests for the data structures and JSON
// serialization/deserialization of API request and response types.
package server

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/codila125/foedus-blockchain/blockchain"
)

// =============================================================================
// BlockRes Tests
// =============================================================================

func TestBlockRes_JSONMarshal(t *testing.T) {
	block := BlockRes{
		PrevHash:     "abc123",
		Hash:         "def456",
		Contract:     []BlockContractRes{},
		Transactions: []BlockTransactionRes{},
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("Failed to marshal BlockRes: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON output")
	}
}

func TestBlockRes_JSONUnmarshal(t *testing.T) {
	jsonStr := `{"prev_hash":"abc123","hash":"def456","contract":[],"transactions":[]}`

	var block BlockRes
	if err := json.Unmarshal([]byte(jsonStr), &block); err != nil {
		t.Fatalf("Failed to unmarshal BlockRes: %v", err)
	}

	if block.PrevHash != "abc123" {
		t.Errorf("Expected prev_hash 'abc123', got '%s'", block.PrevHash)
	}

	if block.Hash != "def456" {
		t.Errorf("Expected hash 'def456', got '%s'", block.Hash)
	}
}

func TestBlockRes_RoundTrip(t *testing.T) {
	original := BlockRes{
		PrevHash:     "test_prev_hash",
		Hash:         "test_hash",
		Contract:     []BlockContractRes{},
		Transactions: []BlockTransactionRes{},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded BlockRes
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Hash != original.Hash {
		t.Errorf("Hash mismatch: expected '%s', got '%s'", original.Hash, decoded.Hash)
	}
}

// =============================================================================
// BlockContractRes Tests
// =============================================================================

func TestBlockContractRes_JSONMarshal(t *testing.T) {
	now := time.Now()
	contract := BlockContractRes{
		ID:        "contract_id",
		Title:     "Test Contract",
		CreatedAt: now,
		UpdatedAt: now,
		Status:    blockchain.ContractActive,
	}

	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatalf("Failed to marshal BlockContractRes: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON output")
	}
}

func TestBlockContractRes_JSONUnmarshal(t *testing.T) {
	jsonStr := `{"id":"abc123","title":"Test","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z","status":"ACTIVE"}`

	var contract BlockContractRes
	if err := json.Unmarshal([]byte(jsonStr), &contract); err != nil {
		t.Fatalf("Failed to unmarshal BlockContractRes: %v", err)
	}

	if contract.ID != "abc123" {
		t.Errorf("Expected ID 'abc123', got '%s'", contract.ID)
	}

	if contract.Title != "Test" {
		t.Errorf("Expected title 'Test', got '%s'", contract.Title)
	}
}

func TestBlockContractRes_AllStatuses(t *testing.T) {
	statuses := []blockchain.ContractStatus{
		blockchain.ContractDraft,
		blockchain.ContractActive,
		blockchain.ContractCompleted,
		blockchain.ContractCancelled,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			contract := BlockContractRes{
				ID:     "id",
				Title:  "Title",
				Status: status,
			}

			data, err := json.Marshal(contract)
			if err != nil {
				t.Fatalf("Failed to marshal with status %s: %v", status, err)
			}

			var decoded BlockContractRes
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal with status %s: %v", status, err)
			}

			if decoded.Status != status {
				t.Errorf("Status mismatch: expected %s, got %s", status, decoded.Status)
			}
		})
	}
}

// =============================================================================
// BlockTransactionRes Tests
// =============================================================================

func TestBlockTransactionRes_JSONMarshal(t *testing.T) {
	tx := BlockTransactionRes{
		ID:      "tx_id",
		Inputs:  []BlockTransactionInputRes{},
		Outputs: []BlockTransactionOutputRes{},
	}

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("Failed to marshal BlockTransactionRes: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON output")
	}
}

func TestBlockTransactionRes_WithInputsOutputs(t *testing.T) {
	tx := BlockTransactionRes{
		ID: "tx_id",
		Inputs: []BlockTransactionInputRes{
			{From: "input1", Out: 0},
			{From: "input2", Out: 1},
		},
		Outputs: []BlockTransactionOutputRes{
			{To: "output1", Value: 100},
			{To: "output2", Value: 50},
		},
	}

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded BlockTransactionRes
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(decoded.Inputs) != 2 {
		t.Errorf("Expected 2 inputs, got %d", len(decoded.Inputs))
	}

	if len(decoded.Outputs) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(decoded.Outputs))
	}
}

// =============================================================================
// BlockTransactionInputRes Tests
// =============================================================================

func TestBlockTransactionInputRes_JSONFields(t *testing.T) {
	input := BlockTransactionInputRes{
		From: "source_tx",
		Out:  0,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, "\"from\"") {
		t.Error("JSON should contain 'from' field")
	}
	if !strings.Contains(jsonStr, "\"out\"") {
		t.Error("JSON should contain 'out' field")
	}
}

func TestBlockTransactionInputRes_RoundTrip(t *testing.T) {
	original := BlockTransactionInputRes{
		From: "test_from",
		Out:  42,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded BlockTransactionInputRes
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.From != original.From {
		t.Errorf("From mismatch: expected '%s', got '%s'", original.From, decoded.From)
	}

	if decoded.Out != original.Out {
		t.Errorf("Out mismatch: expected %d, got %d", original.Out, decoded.Out)
	}
}

// =============================================================================
// BlockTransactionOutputRes Tests
// =============================================================================

func TestBlockTransactionOutputRes_JSONFields(t *testing.T) {
	output := BlockTransactionOutputRes{
		To:    "recipient",
		Value: 1000,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, "\"to\"") {
		t.Error("JSON should contain 'to' field")
	}
	if !strings.Contains(jsonStr, "\"value\"") {
		t.Error("JSON should contain 'value' field")
	}
}

func TestBlockTransactionOutputRes_ZeroValue(t *testing.T) {
	output := BlockTransactionOutputRes{
		To:    "recipient",
		Value: 0,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded BlockTransactionOutputRes
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Value != 0 {
		t.Errorf("Expected value 0, got %d", decoded.Value)
	}
}

// =============================================================================
// CreateContractReq Tests
// =============================================================================

func TestCreateContractReq_FullRequest(t *testing.T) {
	req := CreateContractReq{
		Title:       "Test Contract",
		Description: "A test contract description",
		Creator:     "creator_address",
		Milestones: []AddMilestoneReq{
			{Title: "Milestone 1", Description: "First milestone", Value: 1000, DueDate: 1735689600},
		},
		Parties: []AddPartyReq{
			{Address: "party1", Role: blockchain.RoleContractor},
		},
		Terms:       "Contract terms and conditions",
		Attachments: []string{"doc1.pdf", "doc2.pdf"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal CreateContractReq: %v", err)
	}

	var decoded CreateContractReq
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CreateContractReq: %v", err)
	}

	if decoded.Title != req.Title {
		t.Errorf("Title mismatch: expected '%s', got '%s'", req.Title, decoded.Title)
	}

	if len(decoded.Milestones) != 1 {
		t.Errorf("Expected 1 milestone, got %d", len(decoded.Milestones))
	}

	if len(decoded.Parties) != 1 {
		t.Errorf("Expected 1 party, got %d", len(decoded.Parties))
	}

	if len(decoded.Attachments) != 2 {
		t.Errorf("Expected 2 attachments, got %d", len(decoded.Attachments))
	}
}

func TestCreateContractReq_EmptyFields(t *testing.T) {
	req := CreateContractReq{}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal empty request: %v", err)
	}

	var decoded CreateContractReq
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal empty request: %v", err)
	}

	if decoded.Title != "" {
		t.Error("Expected empty title")
	}
}

func TestCreateContractReq_JSONFieldNames(t *testing.T) {
	req := CreateContractReq{
		Title:       "Test",
		Description: "Desc",
		Creator:     "Creator",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)
	expectedFields := []string{"\"title\"", "\"description\"", "\"creator\""}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON should contain field %s", field)
		}
	}
}

// =============================================================================
// AddMilestoneReq Tests
// =============================================================================

func TestAddMilestoneReq_JSONMarshal(t *testing.T) {
	milestone := AddMilestoneReq{
		Title:       "Milestone Title",
		Description: "Milestone Description",
		Value:       5000,
		DueDate:     1735689600,
	}

	data, err := json.Marshal(milestone)
	if err != nil {
		t.Fatalf("Failed to marshal AddMilestoneReq: %v", err)
	}

	var decoded AddMilestoneReq
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AddMilestoneReq: %v", err)
	}

	if decoded.Title != milestone.Title {
		t.Errorf("Title mismatch: expected '%s', got '%s'", milestone.Title, decoded.Title)
	}

	if decoded.Value != milestone.Value {
		t.Errorf("Value mismatch: expected %d, got %d", milestone.Value, decoded.Value)
	}

	if decoded.DueDate != milestone.DueDate {
		t.Errorf("DueDate mismatch: expected %d, got %d", milestone.DueDate, decoded.DueDate)
	}
}

func TestAddMilestoneReq_JSONFieldNames(t *testing.T) {
	milestone := AddMilestoneReq{
		Title:       "Test",
		Description: "Desc",
		Value:       100,
		DueDate:     1735689600,
	}

	data, err := json.Marshal(milestone)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)
	expectedFields := []string{"\"title\"", "\"description\"", "\"value\"", "\"due_date\""}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON should contain field %s", field)
		}
	}
}

// =============================================================================
// AddPartyReq Tests
// =============================================================================

func TestAddPartyReq_JSONMarshal(t *testing.T) {
	party := AddPartyReq{
		Address: "party_address",
		Role:    blockchain.RoleContractor,
	}

	data, err := json.Marshal(party)
	if err != nil {
		t.Fatalf("Failed to marshal AddPartyReq: %v", err)
	}

	var decoded AddPartyReq
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AddPartyReq: %v", err)
	}

	if decoded.Address != party.Address {
		t.Errorf("Address mismatch: expected '%s', got '%s'", party.Address, decoded.Address)
	}

	if decoded.Role != party.Role {
		t.Errorf("Role mismatch: expected %s, got %s", party.Role, decoded.Role)
	}
}

func TestAddPartyReq_AllRoles(t *testing.T) {
	roles := []blockchain.ContractRole{
		blockchain.RoleContractor,
		blockchain.RoleArbitrator,
		blockchain.RoleCreator,
	}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			party := AddPartyReq{
				Address: "address",
				Role:    role,
			}

			data, err := json.Marshal(party)
			if err != nil {
				t.Fatalf("Failed to marshal with role %s: %v", role, err)
			}

			var decoded AddPartyReq
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal with role %s: %v", role, err)
			}

			if decoded.Role != role {
				t.Errorf("Role mismatch: expected %s, got %s", role, decoded.Role)
			}
		})
	}
}

// =============================================================================
// ContractRes Tests
// =============================================================================

func TestContractRes_FullResponse(t *testing.T) {
	now := time.Now()
	res := ContractRes{
		ID:          "contract_id",
		Title:       "Contract Title",
		Description: "Contract Description",
		Creator:     "creator_addr",
		Milestones:  []ContractMilestoneRes{},
		Parties:     []ContractPartyRes{},
		Terms:       "Terms and conditions",
		Attachments: []string{"doc.pdf"},
		Status:      blockchain.ContractActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("Failed to marshal ContractRes: %v", err)
	}

	var decoded ContractRes
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ContractRes: %v", err)
	}

	if decoded.ID != res.ID {
		t.Errorf("ID mismatch: expected '%s', got '%s'", res.ID, decoded.ID)
	}

	if decoded.Status != res.Status {
		t.Errorf("Status mismatch: expected %s, got %s", res.Status, decoded.Status)
	}
}

// =============================================================================
// ContractMilestoneRes Tests
// =============================================================================

func TestContractMilestoneRes_JSONMarshal(t *testing.T) {
	now := time.Now()
	milestone := ContractMilestoneRes{
		ID:          "milestone_id",
		Title:       "Milestone Title",
		Description: "Milestone Description",
		Value:       2000,
		DueDate:     now.Add(30 * 24 * time.Hour),
		Status:      blockchain.MilestoneActive,
		CreatedAt:   now,
		Evidence:    "evidence_hash",
		ApprovedBy:  []string{"addr1", "addr2"},
		CompletedAt: time.Time{},
	}

	data, err := json.Marshal(milestone)
	if err != nil {
		t.Fatalf("Failed to marshal ContractMilestoneRes: %v", err)
	}

	var decoded ContractMilestoneRes
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ContractMilestoneRes: %v", err)
	}

	if decoded.ID != milestone.ID {
		t.Errorf("ID mismatch: expected '%s', got '%s'", milestone.ID, decoded.ID)
	}

	if len(decoded.ApprovedBy) != 2 {
		t.Errorf("Expected 2 approvers, got %d", len(decoded.ApprovedBy))
	}
}

func TestContractMilestoneRes_AllStatuses(t *testing.T) {
	statuses := []blockchain.MilestoneStatus{
		blockchain.MilestoneActive,
		blockchain.MilestoneCompleted,
		blockchain.MilestoneCancelled,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			milestone := ContractMilestoneRes{
				ID:     "id",
				Title:  "Title",
				Status: status,
			}

			data, err := json.Marshal(milestone)
			if err != nil {
				t.Fatalf("Failed to marshal with status %s: %v", status, err)
			}

			var decoded ContractMilestoneRes
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal with status %s: %v", status, err)
			}

			if decoded.Status != status {
				t.Errorf("Status mismatch: expected %s, got %s", status, decoded.Status)
			}
		})
	}
}

// =============================================================================
// ContractPartyRes Tests
// =============================================================================

func TestContractPartyRes_JSONMarshal(t *testing.T) {
	party := ContractPartyRes{
		Address:   "party_address",
		Role:      blockchain.RoleContractor,
		PublicKey: "public_key_hex",
	}

	data, err := json.Marshal(party)
	if err != nil {
		t.Fatalf("Failed to marshal ContractPartyRes: %v", err)
	}

	var decoded ContractPartyRes
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ContractPartyRes: %v", err)
	}

	if decoded.Address != party.Address {
		t.Errorf("Address mismatch: expected '%s', got '%s'", party.Address, decoded.Address)
	}

	if decoded.PublicKey != party.PublicKey {
		t.Errorf("PublicKey mismatch: expected '%s', got '%s'", party.PublicKey, decoded.PublicKey)
	}
}

func TestContractPartyRes_JSONFieldNames(t *testing.T) {
	party := ContractPartyRes{
		Address:   "addr",
		Role:      blockchain.RoleContractor,
		PublicKey: "key",
	}

	data, err := json.Marshal(party)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)
	expectedFields := []string{"\"address\"", "\"role\"", "\"public_key\""}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON should contain field %s", field)
		}
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestEdgeCase_SpecialCharacters(t *testing.T) {
	req := CreateContractReq{
		Title:       "Contract with \"quotes\" and 'apostrophes'",
		Description: "Line1\nLine2\tTabbed",
		Terms:       "Terms with <html> & special chars",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal special characters: %v", err)
	}

	var decoded CreateContractReq
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal special characters: %v", err)
	}

	if decoded.Title != req.Title {
		t.Errorf("Title with special chars mismatch")
	}
}

func TestEdgeCase_UnicodeStrings(t *testing.T) {
	req := CreateContractReq{
		Title:       "合同 - Contract 契約",
		Description: "描述 αβγδ émojis 🎉",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal unicode: %v", err)
	}

	var decoded CreateContractReq
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal unicode: %v", err)
	}

	if decoded.Title != req.Title {
		t.Errorf("Unicode title mismatch: expected '%s', got '%s'", req.Title, decoded.Title)
	}
}

func TestEdgeCase_EmptyArrays(t *testing.T) {
	req := CreateContractReq{
		Title:       "Contract",
		Milestones:  []AddMilestoneReq{},
		Parties:     []AddPartyReq{},
		Attachments: []string{},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal empty arrays: %v", err)
	}

	var decoded CreateContractReq
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal empty arrays: %v", err)
	}

	if len(decoded.Milestones) != 0 {
		t.Error("Expected empty milestones array")
	}
}

func TestEdgeCase_NullArraysFromJSON(t *testing.T) {
	jsonStr := `{"title":"Contract","milestones":null,"parties":null,"attachments":null}`

	var decoded CreateContractReq
	if err := json.Unmarshal([]byte(jsonStr), &decoded); err != nil {
		t.Fatalf("Failed to unmarshal null arrays: %v", err)
	}

	if decoded.Milestones != nil {
		t.Error("Expected nil milestones for null JSON value")
	}
}

func TestEdgeCase_LargeNumbers(t *testing.T) {
	milestone := AddMilestoneReq{
		Title:   "Large value",
		Value:   2147483647,
		DueDate: 9999999999,
	}

	data, err := json.Marshal(milestone)
	if err != nil {
		t.Fatalf("Failed to marshal large numbers: %v", err)
	}

	var decoded AddMilestoneReq
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal large numbers: %v", err)
	}

	if decoded.Value != milestone.Value {
		t.Errorf("Large value mismatch: expected %d, got %d", milestone.Value, decoded.Value)
	}
}

func TestEdgeCase_ZeroValues(t *testing.T) {
	milestone := AddMilestoneReq{
		Title:   "Zero milestone",
		Value:   0,
		DueDate: 0,
	}

	data, err := json.Marshal(milestone)
	if err != nil {
		t.Fatalf("Failed to marshal zero values: %v", err)
	}

	var decoded AddMilestoneReq
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal zero values: %v", err)
	}

	if decoded.Value != 0 {
		t.Errorf("Expected value 0, got %d", decoded.Value)
	}

	if decoded.DueDate != 0 {
		t.Errorf("Expected due_date 0, got %d", decoded.DueDate)
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkBlockRes_Marshal(b *testing.B) {
	block := BlockRes{
		Hash:         "abc123def456",
		PrevHash:     "789ghi012jkl",
		Contract:     []BlockContractRes{},
		Transactions: []BlockTransactionRes{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(block)
	}
}

func BenchmarkBlockRes_Unmarshal(b *testing.B) {
	jsonStr := `{"prev_hash":"abc123","hash":"def456","contract":[],"transactions":[]}`
	data := []byte(jsonStr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var block BlockRes
		_ = json.Unmarshal(data, &block)
	}
}

func BenchmarkCreateContractReq_Marshal(b *testing.B) {
	req := CreateContractReq{
		Title:       "Benchmark Contract",
		Description: "A contract for benchmarking",
		Creator:     "creator_address",
		Milestones: []AddMilestoneReq{
			{Title: "M1", Value: 100, DueDate: 1735689600},
			{Title: "M2", Value: 200, DueDate: 1735689600},
		},
		Parties: []AddPartyReq{
			{Address: "addr1", Role: blockchain.RoleContractor},
		},
		Terms:       "Terms",
		Attachments: []string{"doc1", "doc2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(req)
	}
}
