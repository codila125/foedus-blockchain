package blockchain

import (
	"bytes"
	"testing"
)

// =============================================================================
// Contract Status Tests
// =============================================================================

func TestContractStatus_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status ContractStatus
		value  string
	}{
		{ContractDraft, "DRAFT"},
		{ContractActive, "ACTIVE"},
		{ContractCompleted, "COMPLETED"},
		{ContractCancelled, "CANCELLED"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.value {
			t.Errorf("Expected %s, got %s", tt.value, string(tt.status))
		}
	}
}

func TestContractStatus_Uniqueness(t *testing.T) {
	t.Parallel()

	statuses := []ContractStatus{
		ContractDraft,
		ContractActive,
		ContractCompleted,
		ContractCancelled,
	}

	seen := make(map[ContractStatus]bool)
	for _, status := range statuses {
		if seen[status] {
			t.Errorf("Duplicate status found: %s", status)
		}
		seen[status] = true
	}
}

// =============================================================================
// Milestone Status Tests
// =============================================================================

func TestMilestoneStatus_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status MilestoneStatus
		value  string
	}{
		{MilestoneActive, "ACTIVE"},
		{MilestoneCompleted, "COMPLETED"},
		{MilestoneCancelled, "CANCELLED"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.value {
			t.Errorf("Expected %s, got %s", tt.value, string(tt.status))
		}
	}
}

// =============================================================================
// Contract Role Tests
// =============================================================================

func TestContractRole_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		role  ContractRole
		value string
	}{
		{RoleContractor, "CONTRACTOR"},
		{RoleArbitrator, "ARBITRATOR"},
		{RoleCreator, "CREATOR"},
	}

	for _, tt := range tests {
		if string(tt.role) != tt.value {
			t.Errorf("Expected %s, got %s", tt.value, string(tt.role))
		}
	}
}

// =============================================================================
// CoinbaseOp Tests
// =============================================================================

func TestCoinbaseOp_Basic(t *testing.T) {
	t.Parallel()

	creator := "test_creator"
	data := "test_data"

	contract := CoinbaseOp(creator, data)

	if contract == nil {
		t.Fatal("CoinbaseOp should return non-nil contract")
	}

	if len(contract.ID) == 0 {
		t.Error("Contract ID should not be empty")
	}

	if contract.Title != "Coinbase Foedus" {
		t.Errorf("Expected title 'Coinbase Foedus', got '%s'", contract.Title)
	}

	if contract.Description != data {
		t.Errorf("Expected description '%s', got '%s'", data, contract.Description)
	}

	if contract.CreatorAddress != creator {
		t.Errorf("Expected creator '%s', got '%s'", creator, contract.CreatorAddress)
	}

	if contract.Status != ContractActive {
		t.Errorf("Expected status ACTIVE, got %s", contract.Status)
	}
}

func TestCoinbaseOp_EmptyMilestonesAndParties(t *testing.T) {
	t.Parallel()

	contract := CoinbaseOp("creator", "data")

	if len(contract.Milestones) != 0 {
		t.Errorf("Expected 0 milestones, got %d", len(contract.Milestones))
	}

	if len(contract.Parties) != 0 {
		t.Errorf("Expected 0 parties, got %d", len(contract.Parties))
	}
}

func TestCoinbaseOp_EmptyData(t *testing.T) {
	t.Parallel()

	contract := CoinbaseOp("creator", "")

	// When data is empty, random data should be generated
	if contract.Description == "" {
		t.Error("Expected random data when description is empty")
	}

	// Random data should be hex-encoded
	if len(contract.Description) != 40 {
		t.Logf("Expected 40-char random hex data, got %d chars", len(contract.Description))
	}
}

func TestCoinbaseOp_HashCalculation(t *testing.T) {
	t.Parallel()

	contract := CoinbaseOp("creator", "data")

	// ID should be a hash of the contract
	if len(contract.ID) != 32 { // SHA-256 produces 32 bytes
		t.Errorf("Expected ID length 32, got %d", len(contract.ID))
	}

	// Recalculate hash and verify
	expectedHash := contract.HashContract()
	if !bytes.Equal(contract.ID, expectedHash) {
		t.Error("Contract ID should match its hash")
	}
}

func TestCoinbaseOp_DifferentCreators(t *testing.T) {
	t.Parallel()

	contract1 := CoinbaseOp("creator1", "data")
	contract2 := CoinbaseOp("creator2", "data")

	if bytes.Equal(contract1.ID, contract2.ID) {
		t.Error("Different creators should produce different contract IDs")
	}
}

func TestCoinbaseOp_DifferentData(t *testing.T) {
	t.Parallel()

	contract1 := CoinbaseOp("creator", "data1")
	contract2 := CoinbaseOp("creator", "data2")

	if bytes.Equal(contract1.ID, contract2.ID) {
		t.Error("Different data should produce different contract IDs")
	}
}

// =============================================================================
// IsCoinbaseOp Tests
// =============================================================================

func TestIsCoinbaseOp_True(t *testing.T) {
	t.Parallel()

	contract := CoinbaseOp("creator", "data")

	if !contract.IsCoinbaseOp() {
		t.Error("CoinbaseOp contract should be identified as coinbase")
	}
}

func TestIsCoinbaseOp_False_WrongTitle(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		ID:         []byte("id"),
		Title:      "Regular Contract",
		Milestones: []*Milestone{},
		Parties:    []*Party{},
	}

	if contract.IsCoinbaseOp() {
		t.Error("Contract with wrong title should not be coinbase")
	}
}

func TestIsCoinbaseOp_False_HasMilestones(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		ID:    []byte("id"),
		Title: "Coinbase Foedus",
		Milestones: []*Milestone{
			{ID: []byte("m1"), Title: "Milestone 1"},
		},
		Parties: []*Party{},
	}

	if contract.IsCoinbaseOp() {
		t.Error("Contract with milestones should not be coinbase")
	}
}

func TestIsCoinbaseOp_False_HasParties(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		ID:         []byte("id"),
		Title:      "Coinbase Foedus",
		Milestones: []*Milestone{},
		Parties: []*Party{
			{Address: "addr1", Role: RoleCreator},
		},
	}

	if contract.IsCoinbaseOp() {
		t.Error("Contract with parties should not be coinbase")
	}
}

// =============================================================================
// Contract Structure Tests
// =============================================================================

func TestContract_Fields(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		ID:             []byte("contract_id"),
		Title:          "Test Contract",
		Description:    "Test Description",
		CreatedAt:      1609459200,
		UpdatedAt:      1609459300,
		Status:         ContractDraft,
		Milestones:     []*Milestone{{ID: []byte("m1"), Title: "M1", Value: 100}},
		Parties:        []*Party{{Address: "addr1", Role: RoleCreator}},
		Terms:          []byte("contract_terms"),
		CreatorAddress: "creator_address",
		Attachments:    [][]byte{[]byte("attachment1")},
	}

	if string(contract.ID) != "contract_id" {
		t.Error("ID not set correctly")
	}

	if contract.Title != "Test Contract" {
		t.Error("Title not set correctly")
	}

	if contract.Description != "Test Description" {
		t.Error("Description not set correctly")
	}

	if contract.CreatedAt != 1609459200 {
		t.Error("CreatedAt not set correctly")
	}

	if contract.UpdatedAt != 1609459300 {
		t.Error("UpdatedAt not set correctly")
	}

	if contract.Status != ContractDraft {
		t.Error("Status not set correctly")
	}

	if !bytes.Equal(contract.Terms, []byte("contract_terms")) {
		t.Error("Terms not set correctly")
	}

	if contract.CreatorAddress != "creator_address" {
		t.Error("CreatorAddress not set correctly")
	}

	if len(contract.Attachments) != 1 {
		t.Error("Attachments not set correctly")
	}

	if len(contract.Milestones) != 1 {
		t.Error("Milestones not set correctly")
	}

	if len(contract.Parties) != 1 {
		t.Error("Parties not set correctly")
	}
}

// =============================================================================
// Milestone Structure Tests
// =============================================================================

func TestMilestone_Fields(t *testing.T) {
	t.Parallel()

	milestone := &Milestone{
		ID:          []byte("milestone_id"),
		Title:       "Test Milestone",
		Description: "Milestone Description",
		Value:       1000,
		DueDate:     1609459200,
		Status:      MilestoneActive,
		CreatedAt:   1609459100,
		CompletedAt: 0,
		Evidence:    []byte("evidence_hash"),
		ApprovedBy:  []string{"addr1", "addr2"},
	}

	if string(milestone.ID) != "milestone_id" {
		t.Error("ID not set correctly")
	}

	if milestone.Title != "Test Milestone" {
		t.Error("Title not set correctly")
	}

	if milestone.Description != "Milestone Description" {
		t.Error("Description not set correctly")
	}

	if milestone.Value != 1000 {
		t.Error("Value not set correctly")
	}

	if milestone.DueDate != 1609459200 {
		t.Error("DueDate not set correctly")
	}

	if milestone.Status != MilestoneActive {
		t.Error("Status not set correctly")
	}

	if milestone.CreatedAt != 1609459100 {
		t.Error("CreatedAt not set correctly")
	}

	if milestone.CompletedAt != 0 {
		t.Error("CompletedAt not set correctly")
	}

	if !bytes.Equal(milestone.Evidence, []byte("evidence_hash")) {
		t.Error("Evidence not set correctly")
	}

	if len(milestone.ApprovedBy) != 2 {
		t.Error("ApprovedBy not set correctly")
	}
}

func TestMilestone_ZeroValue(t *testing.T) {
	t.Parallel()

	milestone := &Milestone{
		ID:    []byte("m"),
		Title: "Test",
		Value: 0,
	}

	if milestone.Value != 0 {
		t.Error("Milestone should accept zero value")
	}
	if err := milestone.Validate(); err != nil {
		t.Errorf("Zero value should be valid: %v", err)
	}
}

func TestMilestone_NegativeValue(t *testing.T) {
	t.Parallel()

	// Struct accepts negative, but validation should fail
	milestone := &Milestone{
		ID:    []byte("m"),
		Title: "Test",
		Value: -100,
	}

	if err := milestone.Validate(); err != ErrMilestoneNegativeValue {
		t.Errorf("Expected ErrMilestoneNegativeValue, got: %v", err)
	}
}

func TestMilestone_EmptyTitle(t *testing.T) {
	t.Parallel()

	milestone := &Milestone{
		ID:    []byte("m"),
		Title: "",
		Value: 100,
	}

	if err := milestone.Validate(); err != ErrMilestoneEmptyTitle {
		t.Errorf("Expected ErrMilestoneEmptyTitle, got: %v", err)
	}
}

// =============================================================================
// Party Structure Tests
// =============================================================================

func TestParty_Fields(t *testing.T) {
	t.Parallel()

	party := &Party{
		Address:   "party_address",
		Role:      RoleContractor,
		PublicKey: []byte("public_key"),
		Signature: []byte("signature"),
	}

	if party.Address != "party_address" {
		t.Error("Address not set correctly")
	}

	if party.Role != RoleContractor {
		t.Error("Role not set correctly")
	}

	if !bytes.Equal(party.PublicKey, []byte("public_key")) {
		t.Error("PublicKey not set correctly")
	}

	if !bytes.Equal(party.Signature, []byte("signature")) {
		t.Error("Signature not set correctly")
	}
}

func TestParty_AllRoles(t *testing.T) {
	t.Parallel()

	roles := []ContractRole{RoleContractor, RoleArbitrator, RoleCreator}

	for _, role := range roles {
		party := &Party{
			Address: "addr",
			Role:    role,
		}

		if party.Role != role {
			t.Errorf("Expected role %s, got %s", role, party.Role)
		}

		if err := party.Validate(); err != nil {
			t.Errorf("Party with role %s should be valid: %v", role, err)
		}
	}
}

func TestParty_EmptyAddress(t *testing.T) {
	t.Parallel()

	party := &Party{
		Address: "",
		Role:    RoleContractor,
	}

	if err := party.Validate(); err != ErrPartyEmptyAddress {
		t.Errorf("Expected ErrPartyEmptyAddress, got: %v", err)
	}
}

func TestParty_EmptyRole(t *testing.T) {
	t.Parallel()

	party := &Party{
		Address: "addr",
		Role:    "",
	}

	if err := party.Validate(); err != ErrPartyEmptyRole {
		t.Errorf("Expected ErrPartyEmptyRole, got: %v", err)
	}
}

// =============================================================================
// ContractCore and MilestoneCore Tests
// =============================================================================

func TestContractCore_Fields(t *testing.T) {
	t.Parallel()

	core := ContractCore{
		Title:          "Core Title",
		Description:    "Core Description",
		CreatedAt:      1000,
		Milestones:     []*MilestoneCore{{Title: "M1", Value: 100}},
		Parties:        []*PartyCore{{Address: "addr1", Role: "CREATOR"}},
		Terms:          []byte("terms"),
		CreatorAddress: "creator",
		Attachments:    [][]byte{{0x01}},
	}

	if core.Title != "Core Title" {
		t.Error("Title not set correctly")
	}

	if core.Description != "Core Description" {
		t.Error("Description not set correctly")
	}

	if core.CreatedAt != 1000 {
		t.Error("CreatedAt not set correctly")
	}

	if len(core.Milestones) != 1 {
		t.Error("Milestones not set correctly")
	}

	if len(core.Parties) != 1 {
		t.Error("Parties not set correctly")
	}

	if !bytes.Equal(core.Terms, []byte("terms")) {
		t.Error("Terms not set correctly")
	}

	if core.CreatorAddress != "creator" {
		t.Error("CreatorAddress not set correctly")
	}

	if len(core.Attachments) != 1 {
		t.Error("Attachments not set correctly")
	}
}

func TestMilestoneCore_Fields(t *testing.T) {
	t.Parallel()

	core := MilestoneCore{
		Title:       "Core Milestone",
		Description: "Core Milestone Description",
		Value:       500,
		CreatedAt:   2000,
	}

	if core.Title != "Core Milestone" {
		t.Error("Title not set correctly")
	}

	if core.Description != "Core Milestone Description" {
		t.Error("Description not set correctly")
	}

	if core.Value != 500 {
		t.Error("Value not set correctly")
	}

	if core.CreatedAt != 2000 {
		t.Error("CreatedAt not set correctly")
	}
}

func TestPartyCore_Fields(t *testing.T) {
	t.Parallel()

	core := PartyCore{
		Address:   "core_address",
		Role:      "CONTRACTOR",
		PublicKey: []byte("pk"),
	}

	if core.Address != "core_address" {
		t.Error("Address not set correctly")
	}

	if core.Role != "CONTRACTOR" {
		t.Error("Role not set correctly")
	}

	if !bytes.Equal(core.PublicKey, []byte("pk")) {
		t.Error("PublicKey not set correctly")
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestContract_EmptyContract(t *testing.T) {
	t.Parallel()

	contract := &Contract{}

	if len(contract.ID) > 0 {
		t.Error("Empty contract should have nil or empty ID")
	}

	if contract.Status != "" {
		t.Error("Empty contract should have empty status")
	}

	// Empty contract should fail validation
	if err := contract.Validate(); err != ErrContractEmptyTitle {
		t.Errorf("Expected ErrContractEmptyTitle for empty contract, got: %v", err)
	}
}

func TestContract_EmptyTitle(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		Title:          "",
		CreatorAddress: "creator",
	}

	if err := contract.Validate(); err != ErrContractEmptyTitle {
		t.Errorf("Expected ErrContractEmptyTitle, got: %v", err)
	}
}

func TestContract_EmptyCreator(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		Title:          "Test Contract",
		CreatorAddress: "",
	}

	if err := contract.Validate(); err != ErrContractEmptyCreator {
		t.Errorf("Expected ErrContractEmptyCreator, got: %v", err)
	}
}

func TestContract_ValidContract(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		Title:          "Test Contract",
		CreatorAddress: "creator",
		Milestones: []*Milestone{
			{Title: "Milestone 1", Value: 100},
		},
		Parties: []*Party{
			{Address: "addr", Role: RoleContractor},
		},
	}

	if err := contract.Validate(); err != nil {
		t.Errorf("Valid contract should pass validation: %v", err)
	}
}

func TestContract_InvalidMilestone(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		Title:          "Test Contract",
		CreatorAddress: "creator",
		Milestones: []*Milestone{
			{Title: "Valid", Value: 100},
			{Title: "", Value: 50}, // Invalid: empty title
		},
	}

	if err := contract.Validate(); err != ErrMilestoneEmptyTitle {
		t.Errorf("Expected ErrMilestoneEmptyTitle for contract with invalid milestone, got: %v", err)
	}
}

func TestContract_InvalidParty(t *testing.T) {
	t.Parallel()

	contract := &Contract{
		Title:          "Test Contract",
		CreatorAddress: "creator",
		Parties: []*Party{
			{Address: "", Role: RoleContractor}, // Invalid: empty address
		},
	}

	if err := contract.Validate(); err != ErrPartyEmptyAddress {
		t.Errorf("Expected ErrPartyEmptyAddress for contract with invalid party, got: %v", err)
	}
}

func TestContract_WithMultipleMilestones(t *testing.T) {
	t.Parallel()

	milestones := make([]*Milestone, 10)
	for i := 0; i < 10; i++ {
		milestones[i] = &Milestone{
			ID:     []byte{byte(i)},
			Title:  "Milestone",
			Value:  i * 100,
			Status: MilestoneActive,
		}
	}

	contract := &Contract{
		ID:         []byte("contract_with_milestones"),
		Milestones: milestones,
	}

	if !bytes.Equal(contract.ID, []byte("contract_with_milestones")) {
		t.Error("Contract ID not set correctly")
	}

	if len(contract.Milestones) != 10 {
		t.Errorf("Expected 10 milestones, got %d", len(contract.Milestones))
	}

	// Verify milestone fields are correctly set
	for i, m := range contract.Milestones {
		if !bytes.Equal(m.ID, []byte{byte(i)}) {
			t.Errorf("Milestone %d ID mismatch", i)
		}
		if m.Title != "Milestone" {
			t.Errorf("Milestone %d Title mismatch", i)
		}
		if m.Value != i*100 {
			t.Errorf("Milestone %d Value mismatch", i)
		}
		if m.Status != MilestoneActive {
			t.Errorf("Milestone %d Status mismatch", i)
		}
	}
}

func TestContract_WithMultipleParties(t *testing.T) {
	t.Parallel()

	parties := []*Party{
		{Address: "addr1", Role: RoleCreator},
		{Address: "addr2", Role: RoleContractor},
		{Address: "addr3", Role: RoleArbitrator},
	}

	contract := &Contract{
		ID:      []byte("contract_with_parties"),
		Parties: parties,
	}

	if !bytes.Equal(contract.ID, []byte("contract_with_parties")) {
		t.Error("ID not set correctly")
	}

	if len(contract.Parties) != 3 {
		t.Errorf("Expected 3 parties, got %d", len(contract.Parties))
	}

	// Verify party fields
	if contract.Parties[0].Address != "addr1" || contract.Parties[0].Role != RoleCreator {
		t.Error("Party 0 fields mismatch")
	}
	if contract.Parties[1].Address != "addr2" || contract.Parties[1].Role != RoleContractor {
		t.Error("Party 1 fields mismatch")
	}
	if contract.Parties[2].Address != "addr3" || contract.Parties[2].Role != RoleArbitrator {
		t.Error("Party 2 fields mismatch")
	}
}

func TestMilestone_WithApprovals(t *testing.T) {
	t.Parallel()

	approvers := []string{"addr1", "addr2", "addr3", "addr4", "addr5"}

	milestone := &Milestone{
		ID:         []byte("m"),
		Title:      "Test",
		ApprovedBy: approvers,
	}

	if len(milestone.ApprovedBy) != 5 {
		t.Errorf("Expected 5 approvers, got %d", len(milestone.ApprovedBy))
	}

	if !bytes.Equal(milestone.ID, []byte("m")) {
		t.Error("ID not set correctly")
	}

	if milestone.Title != "Test" {
		t.Error("Title not set correctly")
	}
}

func TestContract_LargeAttachments(t *testing.T) {
	t.Parallel()

	attachments := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		attachments[i] = make([]byte, 32) // IPFS hash size
	}

	contract := &Contract{
		ID:          []byte("contract_with_attachments"),
		Attachments: attachments,
	}

	if len(contract.Attachments) != 100 {
		t.Errorf("Expected 100 attachments, got %d", len(contract.Attachments))
	}

	if !bytes.Equal(contract.ID, []byte("contract_with_attachments")) {
		t.Error("ID not set correctly")
	}
}
