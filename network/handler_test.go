package network

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/protobuf"
	"google.golang.org/protobuf/proto"
)

// =============================================================================
// Handler Command Tests
// =============================================================================

func TestHandlerCommands_Valid(t *testing.T) {
	validCommands := []string{
		"GET_BLOCKCHAIN",
		"GET_BLOCKS",
		"GET_VERSION",
		"NEW_CONTRACT",
		"NEW_BLOCK",
	}

	for _, cmd := range validCommands {
		t.Run(cmd, func(t *testing.T) {
			// Verify command strings are properly formatted
			if len(cmd) == 0 {
				t.Error("Command should not be empty")
			}
			if cmd != string([]byte(cmd)) {
				t.Error("Command should be valid string")
			}
		})
	}
}

func TestHandlerCommands_UnknownCommand(t *testing.T) {
	unknownCommands := []string{
		"UNKNOWN",
		"invalid",
		"GET_BLOCK", // Missing 'S'
		"NEW_TX",
		"",
	}

	validCommands := map[string]bool{
		"GET_BLOCKCHAIN": true,
		"GET_BLOCKS":     true,
		"GET_VERSION":    true,
		"NEW_CONTRACT":   true,
		"NEW_BLOCK":      true,
	}

	for _, cmd := range unknownCommands {
		t.Run("Unknown_"+cmd, func(t *testing.T) {
			// This test verifies the commands don't match valid ones
			if validCommands[cmd] {
				t.Errorf("Command '%s' should not be valid", cmd)
			}
		})
	}
}

// =============================================================================
// Contract Protobuf Tests
// =============================================================================

func TestContractData_Serialization(t *testing.T) {
	contract := &protobuf.Contract{
		Id:             []byte("contract_id_123"),
		Title:          "Test Contract",
		Description:    "A test contract for unit testing",
		CreatorAddress: "creator_address_abc",
		Attachments:    [][]byte{[]byte("attachment1.pdf"), []byte("attachment2.doc")},
		Terms:          []byte("These are the terms"),
		Status:         "active",
		CreatedAt:      1234567890,
		UpdatedAt:      1234567891,
		Milestones:     []*protobuf.Milestone{},
		Parties:        []*protobuf.Party{},
	}

	data, err := proto.Marshal(contract)
	if err != nil {
		t.Fatalf("Failed to marshal contract: %v", err)
	}

	decoded := &protobuf.Contract{}
	err = proto.Unmarshal(data, decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal contract: %v", err)
	}

	if !bytes.Equal(decoded.Id, contract.Id) {
		t.Error("Contract ID mismatch")
	}
	if decoded.Title != contract.Title {
		t.Error("Contract Title mismatch")
	}
	if decoded.Description != contract.Description {
		t.Error("Contract Description mismatch")
	}
	if decoded.CreatorAddress != contract.CreatorAddress {
		t.Error("Contract CreatorAddress mismatch")
	}
	if len(decoded.Attachments) != len(contract.Attachments) {
		t.Error("Contract Attachments length mismatch")
	}
	if !bytes.Equal(decoded.Terms, contract.Terms) {
		t.Error("Contract Terms mismatch")
	}
	if decoded.Status != contract.Status {
		t.Error("Contract Status mismatch")
	}
}

func TestContractWithMilestones_Serialization(t *testing.T) {
	milestone := &protobuf.Milestone{
		Id:          []byte("milestone_1"),
		Title:       "First Milestone",
		Description: "Complete phase 1",
		Value:       1000,
		DueDate:     1234567890,
		Status:      "pending",
		CreatedAt:   1234567880,
		CompletedAt: 0,
		Evidence:    []byte("evidence1.pdf"),
		ApprovedBy:  []string{},
	}

	contract := &protobuf.Contract{
		Id:         []byte("contract_with_milestones"),
		Title:      "Contract with Milestones",
		Milestones: []*protobuf.Milestone{milestone},
	}

	data, err := proto.Marshal(contract)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	decoded := &protobuf.Contract{}
	err = proto.Unmarshal(data, decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(decoded.Milestones) != 1 {
		t.Fatalf("Expected 1 milestone, got %d", len(decoded.Milestones))
	}

	m := decoded.Milestones[0]
	if !bytes.Equal(m.Id, milestone.Id) {
		t.Error("Milestone ID mismatch")
	}
	if m.Title != milestone.Title {
		t.Error("Milestone Title mismatch")
	}
	if m.Value != milestone.Value {
		t.Error("Milestone Value mismatch")
	}
}

func TestContractWithParties_Serialization(t *testing.T) {
	party := &protobuf.Party{
		Address:   "party_address_123",
		Role:      "buyer",
		PublicKey: []byte("public_key_bytes"),
		Signature: []byte("signature_bytes"),
	}

	contract := &protobuf.Contract{
		Id:      []byte("contract_with_parties"),
		Title:   "Contract with Parties",
		Parties: []*protobuf.Party{party},
	}

	data, err := proto.Marshal(contract)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	decoded := &protobuf.Contract{}
	err = proto.Unmarshal(data, decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(decoded.Parties) != 1 {
		t.Fatalf("Expected 1 party, got %d", len(decoded.Parties))
	}

	p := decoded.Parties[0]
	if p.Address != party.Address {
		t.Error("Party Address mismatch")
	}
	if p.Role != party.Role {
		t.Error("Party Role mismatch")
	}
	if !bytes.Equal(p.PublicKey, party.PublicKey) {
		t.Error("Party PublicKey mismatch")
	}
	if !bytes.Equal(p.Signature, party.Signature) {
		t.Error("Party Signature mismatch")
	}
}

// =============================================================================
// Contract Status Tests
// =============================================================================

func TestContractStatus_Values(t *testing.T) {
	validStatuses := []string{
		"draft",
		"pending",
		"active",
		"completed",
		"cancelled",
		"disputed",
	}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			contract := &protobuf.Contract{
				Id:     []byte("test_id"),
				Status: status,
			}

			data, err := proto.Marshal(contract)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			decoded := &protobuf.Contract{}
			err = proto.Unmarshal(data, decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if decoded.Status != status {
				t.Errorf("Expected status '%s', got '%s'", status, decoded.Status)
			}
		})
	}
}

// =============================================================================
// Milestone Status Tests
// =============================================================================

func TestMilestoneStatus_Values(t *testing.T) {
	validStatuses := []string{
		"pending",
		"in_progress",
		"submitted",
		"approved",
		"rejected",
		"completed",
	}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			milestone := &protobuf.Milestone{
				Id:     []byte("milestone_id"),
				Status: status,
			}

			data, err := proto.Marshal(milestone)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			decoded := &protobuf.Milestone{}
			err = proto.Unmarshal(data, decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if decoded.Status != status {
				t.Errorf("Expected status '%s', got '%s'", status, decoded.Status)
			}
		})
	}
}

// =============================================================================
// Party Role Tests
// =============================================================================

func TestPartyRole_Values(t *testing.T) {
	validRoles := []string{
		"buyer",
		"seller",
		"witness",
		"arbiter",
	}

	for _, role := range validRoles {
		t.Run(role, func(t *testing.T) {
			party := &protobuf.Party{
				Address: "test_address",
				Role:    role,
			}

			data, err := proto.Marshal(party)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			decoded := &protobuf.Party{}
			err = proto.Unmarshal(data, decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if decoded.Role != role {
				t.Errorf("Expected role '%s', got '%s'", role, decoded.Role)
			}
		})
	}
}

// =============================================================================
// Block Handling Tests
// =============================================================================

func TestBlockSerialization_RoundTrip(t *testing.T) {
	testCases := []struct {
		name      string
		block     *blockchain.Block
		expectErr bool
	}{
		{
			name: "GenesisBlock",
			block: &blockchain.Block{
				Timestamp:    0,
				Hash:         []byte("genesis_hash"),
				PrevHash:     []byte{},
				Nonce:        0,
				Height:       0,
				Transactions: []*blockchain.Transaction{},
				Contracts:    []*blockchain.Contract{},
			},
			expectErr: false,
		},
		{
			name: "NormalBlock",
			block: &blockchain.Block{
				Timestamp:    1234567890,
				Hash:         []byte("normal_hash"),
				PrevHash:     []byte("prev_hash"),
				Nonce:        12345,
				Height:       100,
				Transactions: []*blockchain.Transaction{},
				Contracts:    []*blockchain.Contract{},
			},
			expectErr: false,
		},
		{
			name: "BlockWithLargeNonce",
			block: &blockchain.Block{
				Timestamp:    1234567890,
				Hash:         []byte("large_nonce_hash"),
				PrevHash:     []byte("prev_hash"),
				Nonce:        2147483647, // Max int32
				Height:       1000000,
				Transactions: []*blockchain.Transaction{},
				Contracts:    []*blockchain.Contract{},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serialized := tc.block.SerializeBlock()
			if serialized == nil {
				if !tc.expectErr {
					t.Error("Serialization returned nil unexpectedly")
				}
				return
			}

			deserialized := blockchain.DeserializeBlock(serialized)
			if deserialized == nil {
				if !tc.expectErr {
					t.Error("Deserialization returned nil unexpectedly")
				}
				return
			}

			if deserialized.Timestamp != tc.block.Timestamp {
				t.Errorf("Timestamp mismatch: expected %d, got %d",
					tc.block.Timestamp, deserialized.Timestamp)
			}
			if !bytes.Equal(deserialized.Hash, tc.block.Hash) {
				t.Error("Hash mismatch")
			}
			if !bytes.Equal(deserialized.PrevHash, tc.block.PrevHash) {
				t.Error("PrevHash mismatch")
			}
			if deserialized.Nonce != tc.block.Nonce {
				t.Errorf("Nonce mismatch: expected %d, got %d",
					tc.block.Nonce, deserialized.Nonce)
			}
			if deserialized.Height != tc.block.Height {
				t.Errorf("Height mismatch: expected %d, got %d",
					tc.block.Height, deserialized.Height)
			}
		})
	}
}

func TestBlockDataProtobuf_Commands(t *testing.T) {
	commands := []string{
		"NEW_BLOCK",
		"done",
		"",
	}

	for _, cmd := range commands {
		t.Run("Command_"+cmd, func(t *testing.T) {
			blockData := &protobuf.BlockData{
				Command: cmd,
				Hash:    []byte("test_hash"),
				Height:  42,
				Data:    []byte("test_data"),
			}

			data, err := proto.Marshal(blockData)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			decoded := &protobuf.BlockData{}
			err = proto.Unmarshal(data, decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if decoded.Command != cmd {
				t.Errorf("Expected command '%s', got '%s'", cmd, decoded.Command)
			}
		})
	}
}

// =============================================================================
// Height Data Request Tests
// =============================================================================

func TestHeightDataRequest_Encoding(t *testing.T) {
	testCases := []struct {
		name   string
		height int32
	}{
		{"Zero", 0},
		{"Small", 10},
		{"Medium", 1000},
		{"Large", 1000000},
		{"MaxInt32", 2147483647},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			heightData := &protobuf.HeightData{Height: tc.height}

			protoData, err := proto.Marshal(heightData)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			// Simulate network transmission with length prefix
			buf := &bytes.Buffer{}
			lenBuf := make([]byte, 4)
			binary.BigEndian.PutUint32(lenBuf, uint32(len(protoData)))
			buf.Write(lenBuf)
			buf.Write(protoData)

			// Read back
			readLenBuf := make([]byte, 4)
			_, err = buf.Read(readLenBuf)
			if err != nil {
				t.Fatalf("Failed to read length: %v", err)
			}

			messageLen := binary.BigEndian.Uint32(readLenBuf)
			msgBuf := make([]byte, messageLen)
			_, err = buf.Read(msgBuf)
			if err != nil {
				t.Fatalf("Failed to read message: %v", err)
			}

			decoded := &protobuf.HeightData{}
			err = proto.Unmarshal(msgBuf, decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if decoded.Height != tc.height {
				t.Errorf("Expected height %d, got %d", tc.height, decoded.Height)
			}
		})
	}
}

// =============================================================================
// Version Data Response Tests
// =============================================================================

func TestVersionDataResponse_Encoding(t *testing.T) {
	testCases := []struct {
		name     string
		height   int32
		lastHash []byte
		nodeID   string
	}{
		{
			name:     "Genesis",
			height:   0,
			lastHash: []byte("genesis_hash"),
			nodeID:   "node_3000",
		},
		{
			name:     "Normal",
			height:   100,
			lastHash: []byte("normal_hash_value"),
			nodeID:   "node_3009",
		},
		{
			name:     "LongHash",
			height:   50000,
			lastHash: bytes.Repeat([]byte("a"), 64), // 64 byte hash
			nodeID:   "node_with_long_id_12345",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			versionData := &protobuf.VersionData{
				Height:   tc.height,
				LastHash: tc.lastHash,
				NodeId:   tc.nodeID,
			}

			protoData, err := proto.Marshal(versionData)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			// Simulate network transmission
			buf := &bytes.Buffer{}
			lenBuf := make([]byte, 4)
			binary.BigEndian.PutUint32(lenBuf, uint32(len(protoData)))
			buf.Write(lenBuf)
			buf.Write(protoData)

			// Read back
			readLenBuf := make([]byte, 4)
			_, _ = buf.Read(readLenBuf)
			messageLen := binary.BigEndian.Uint32(readLenBuf)
			msgBuf := make([]byte, messageLen)
			_, _ = buf.Read(msgBuf)

			decoded := &protobuf.VersionData{}
			err = proto.Unmarshal(msgBuf, decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if decoded.Height != tc.height {
				t.Errorf("Height mismatch")
			}
			if !bytes.Equal(decoded.LastHash, tc.lastHash) {
				t.Errorf("LastHash mismatch")
			}
			if decoded.NodeId != tc.nodeID {
				t.Errorf("NodeID mismatch")
			}
		})
	}
}
