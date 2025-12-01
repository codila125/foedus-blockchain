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
// Integration Test Helpers
// =============================================================================

// createTestBlock creates a test block with configurable parameters
func createTestBlock(height int, hash, prevHash string) *blockchain.Block {
	return &blockchain.Block{
		Timestamp:    int64(1234567890 + height),
		Hash:         []byte(hash),
		PrevHash:     []byte(prevHash),
		Nonce:        height * 100,
		Height:       height,
		Transactions: []*blockchain.Transaction{},
		Contracts:    []*blockchain.Contract{},
	}
}

// createTestBlockchain creates a series of linked blocks
func createTestBlockchain(numBlocks int) []*blockchain.Block {
	blocks := make([]*blockchain.Block, numBlocks)

	for i := 0; i < numBlocks; i++ {
		var prevHash string
		if i == 0 {
			prevHash = ""
		} else {
			prevHash = string(blocks[i-1].Hash)
		}
		blocks[i] = createTestBlock(i, generateTestHash(i), prevHash)
	}

	return blocks
}

func generateTestHash(index int) string {
	hash := make([]byte, 32)
	binary.BigEndian.PutUint32(hash, uint32(index))
	return string(hash)
}

// =============================================================================
// End-to-End Block Transfer Tests
// =============================================================================

func TestBlockTransfer_SingleBlock(t *testing.T) {
	block := createTestBlock(1, "test_hash_1", "genesis_hash")

	serialized := block.SerializeBlock()
	if serialized == nil {
		t.Fatal("Failed to serialize block")
	}

	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Hash:    block.Hash,
		Height:  int32(block.Height),
		Data:    serialized,
	}

	wireData, err := proto.Marshal(blockData)
	if err != nil {
		t.Fatalf("Failed to marshal block data: %v", err)
	}

	decoded := &protobuf.BlockData{}
	err = proto.Unmarshal(wireData, decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal block data: %v", err)
	}

	recoveredBlock := blockchain.DeserializeBlock(decoded.Data)
	if recoveredBlock == nil {
		t.Fatal("Failed to deserialize block")
	}

	if recoveredBlock.Height != block.Height {
		t.Errorf("Height mismatch: expected %d, got %d", block.Height, recoveredBlock.Height)
	}
	if !bytes.Equal(recoveredBlock.Hash, block.Hash) {
		t.Error("Hash mismatch")
	}
	if !bytes.Equal(recoveredBlock.PrevHash, block.PrevHash) {
		t.Error("PrevHash mismatch")
	}
}

func TestBlockTransfer_MultipleBlocks(t *testing.T) {
	blocks := createTestBlockchain(5)

	for _, block := range blocks {
		serialized := block.SerializeBlock()
		if serialized == nil {
			t.Fatalf("Failed to serialize block at height %d", block.Height)
		}

		recovered := blockchain.DeserializeBlock(serialized)
		if recovered == nil {
			t.Fatalf("Failed to deserialize block at height %d", block.Height)
		}

		if recovered.Height != block.Height {
			t.Errorf("Height mismatch at %d", block.Height)
		}
	}
}

func TestBlockTransfer_ChainIntegrity(t *testing.T) {
	blocks := createTestBlockchain(10)

	for i := 1; i < len(blocks); i++ {
		if !bytes.Equal(blocks[i].PrevHash, blocks[i-1].Hash) {
			t.Errorf("Chain broken at block %d", i)
		}
	}
}

// =============================================================================
// Version Exchange Tests
// =============================================================================

func TestVersionExchange_Encoding(t *testing.T) {
	versionData := &protobuf.VersionData{
		Height:   100,
		LastHash: []byte("last_hash_value"),
		NodeId:   "server_node",
	}

	serverResponse, err := proto.Marshal(versionData)
	if err != nil {
		t.Fatalf("Failed to marshal version: %v", err)
	}

	decoded := &protobuf.VersionData{}
	err = proto.Unmarshal(serverResponse, decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal version: %v", err)
	}

	if decoded.Height != 100 {
		t.Errorf("Expected height 100, got %d", decoded.Height)
	}
	if decoded.NodeId != "server_node" {
		t.Errorf("Expected NodeId 'server_node', got %s", decoded.NodeId)
	}
}

func TestVersionExchange_WithWireFormat(t *testing.T) {
	versionData := &protobuf.VersionData{
		Height:   500,
		LastHash: []byte("test_hash_123456"),
		NodeId:   "node_3000",
	}

	protoData, _ := proto.Marshal(versionData)

	buf := &bytes.Buffer{}
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(protoData)))
	buf.Write(lenBuf)
	buf.Write(protoData)

	readLen := make([]byte, 4)
	_, _ = buf.Read(readLen)
	msgLen := binary.BigEndian.Uint32(readLen)
	msgBuf := make([]byte, msgLen)
	_, _ = buf.Read(msgBuf)

	decoded := &protobuf.VersionData{}
	_ = proto.Unmarshal(msgBuf, decoded)

	if decoded.Height != 500 {
		t.Errorf("Height mismatch")
	}
}

// =============================================================================
// Height Request/Response Tests
// =============================================================================

func TestHeightRequest_Encoding(t *testing.T) {
	heights := []int32{0, 10, 100, 1000, 100000}

	for _, height := range heights {
		request := &protobuf.HeightData{Height: height}

		data, err := proto.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal height %d: %v", height, err)
		}

		decoded := &protobuf.HeightData{}
		err = proto.Unmarshal(data, decoded)
		if err != nil {
			t.Fatalf("Failed to unmarshal height %d: %v", height, err)
		}

		if decoded.Height != height {
			t.Errorf("Height mismatch: expected %d, got %d", height, decoded.Height)
		}
	}
}

// =============================================================================
// Version Comparison Tests
// =============================================================================

func TestVersionComparison_SyncDecision(t *testing.T) {
	localHeight := 100

	testCases := []struct {
		name       string
		peerHeight int
		shouldSync bool
	}{
		{"PeerBehind", 50, false},
		{"PeerEqual", 100, false},
		{"PeerAhead", 150, true},
		{"PeerFarAhead", 1000, true},
		{"PeerAtGenesis", 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			needsSync := tc.peerHeight > localHeight
			if needsSync != tc.shouldSync {
				t.Errorf("Expected shouldSync=%v, got %v", tc.shouldSync, needsSync)
			}
		})
	}
}

// =============================================================================
// Contract Broadcasting Tests
// =============================================================================

func TestContractBroadcast_Encoding(t *testing.T) {
	contract := &protobuf.Contract{
		Id:             []byte("contract_123"),
		Title:          "Test Contract",
		Description:    "Test Description",
		CreatorAddress: "creator_addr",
		Status:         "active",
		CreatedAt:      1234567890,
		UpdatedAt:      1234567890,
		Milestones: []*protobuf.Milestone{
			{
				Id:          []byte("milestone_1"),
				Title:       "Milestone 1",
				Description: "First milestone",
				Value:       1000,
				Status:      "pending",
			},
		},
		Parties: []*protobuf.Party{
			{
				Address:   "party_addr_1",
				Role:      "buyer",
				PublicKey: []byte("pubkey"),
			},
		},
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

	if decoded.Title != contract.Title {
		t.Error("Title mismatch")
	}
	if len(decoded.Milestones) != 1 {
		t.Error("Milestones count mismatch")
	}
	if len(decoded.Parties) != 1 {
		t.Error("Parties count mismatch")
	}
}

func TestContractBroadcast_ComplexContract(t *testing.T) {
	// Multiple milestones and parties
	milestones := make([]*protobuf.Milestone, 5)
	for i := 0; i < 5; i++ {
		milestones[i] = &protobuf.Milestone{
			Id:     []byte{byte(i)},
			Title:  "Milestone",
			Value:  int32(i * 100),
			Status: "pending",
		}
	}

	parties := make([]*protobuf.Party, 3)
	roles := []string{"buyer", "seller", "witness"}
	for i := 0; i < 3; i++ {
		parties[i] = &protobuf.Party{
			Address: "addr",
			Role:    roles[i],
		}
	}

	contract := &protobuf.Contract{
		Id:         []byte("complex_contract"),
		Title:      "Complex Contract",
		Milestones: milestones,
		Parties:    parties,
	}

	data, _ := proto.Marshal(contract)
	decoded := &protobuf.Contract{}
	_ = proto.Unmarshal(data, decoded)

	if len(decoded.Milestones) != 5 {
		t.Errorf("Expected 5 milestones, got %d", len(decoded.Milestones))
	}
	if len(decoded.Parties) != 3 {
		t.Errorf("Expected 3 parties, got %d", len(decoded.Parties))
	}
}

// =============================================================================
// Block Broadcast Tests
// =============================================================================

func TestBlockBroadcast_Format(t *testing.T) {
	block := createTestBlock(10, "broadcast_hash", "prev_hash")

	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Hash:    block.Hash,
		Height:  int32(block.Height),
		Data:    block.SerializeBlock(),
	}

	data, err := proto.Marshal(blockData)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	buf := &bytes.Buffer{}
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	buf.Write(lenBuf)
	buf.Write(data)

	if buf.Len() != 4+len(data) {
		t.Errorf("Wire format length mismatch: expected %d, got %d", 4+len(data), buf.Len())
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestErrorHandling_CorruptedProtobuf(t *testing.T) {
	corruptedData := []byte{0xFF, 0xFE, 0xFD, 0xFC}

	blockData := &protobuf.BlockData{}
	err := proto.Unmarshal(corruptedData, blockData)
	// Protobuf may or may not error depending on the corruption
	_ = err // Just ensure no panic
}

func TestErrorHandling_EmptyBlockData(t *testing.T) {
	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Hash:    nil,
		Height:  0,
		Data:    nil,
	}

	data, err := proto.Marshal(blockData)
	if err != nil {
		t.Fatalf("Failed to marshal empty block data: %v", err)
	}

	decoded := &protobuf.BlockData{}
	err = proto.Unmarshal(data, decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Hash != nil {
		t.Error("Hash should be nil")
	}
	if decoded.Data != nil {
		t.Error("Data should be nil")
	}
}

// =============================================================================
// Large Data Tests
// =============================================================================

func TestLargeBlockData(t *testing.T) {
	largeData := make([]byte, 100000) // 100KB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Hash:    largeData[:32],
		Height:  1000,
		Data:    largeData,
	}

	data, err := proto.Marshal(blockData)
	if err != nil {
		t.Fatalf("Failed to marshal large block: %v", err)
	}

	if len(data) < len(largeData) {
		t.Error("Encoded data seems too small")
	}

	decoded := &protobuf.BlockData{}
	err = proto.Unmarshal(data, decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal large block: %v", err)
	}

	if len(decoded.Data) != len(largeData) {
		t.Errorf("Data length mismatch: expected %d, got %d", len(largeData), len(decoded.Data))
	}
}

func TestLargeContractWithManyMilestones(t *testing.T) {
	milestones := make([]*protobuf.Milestone, 100)
	for i := range milestones {
		milestones[i] = &protobuf.Milestone{
			Id:          []byte{byte(i)},
			Title:       "Milestone",
			Description: "Description",
			Value:       int32(i * 100),
			Status:      "pending",
		}
	}

	contract := &protobuf.Contract{
		Id:         []byte("large_contract"),
		Title:      "Large Contract",
		Milestones: milestones,
	}

	data, err := proto.Marshal(contract)
	if err != nil {
		t.Fatalf("Failed to marshal large contract: %v", err)
	}

	decoded := &protobuf.Contract{}
	err = proto.Unmarshal(data, decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal large contract: %v", err)
	}

	if len(decoded.Milestones) != 100 {
		t.Errorf("Expected 100 milestones, got %d", len(decoded.Milestones))
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestEdgeCase_GenesisBlock(t *testing.T) {
	block := createTestBlock(0, "genesis", "")

	if block.Height != 0 {
		t.Error("Genesis block should have height 0")
	}
	if len(block.PrevHash) != 0 {
		t.Error("Genesis block should have empty prev hash")
	}
}

func TestEdgeCase_MaxInt32HeightBlock(t *testing.T) {
	block := &blockchain.Block{
		Height: 2147483647,
		Hash:   []byte("max_height_hash"),
	}

	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Height:  int32(block.Height),
		Hash:    block.Hash,
	}

	data, _ := proto.Marshal(blockData)
	decoded := &protobuf.BlockData{}
	_ = proto.Unmarshal(data, decoded)

	if decoded.Height != 2147483647 {
		t.Errorf("Max height not preserved")
	}
}

func TestEdgeCase_ZeroTimestamp(t *testing.T) {
	block := &blockchain.Block{
		Timestamp:    0,
		Hash:         []byte("zero_ts"),
		PrevHash:     []byte{},
		Height:       0,
		Transactions: []*blockchain.Transaction{},
		Contracts:    []*blockchain.Contract{},
	}

	serialized := block.SerializeBlock()
	deserialized := blockchain.DeserializeBlock(serialized)

	if deserialized.Timestamp != 0 {
		t.Error("Zero timestamp not preserved")
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestConcurrentVersionInfoAccess(t *testing.T) {
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			v := VersionInfo{
				BestHeight: idx * 10,
				LastHash:   []byte{byte(idx)},
				NodeID:     string(rune('A' + idx)),
			}
			_ = v.BestHeight
			_ = v.LastHash
			_ = v.NodeID
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestConcurrentBlockSerialization(t *testing.T) {
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			block := createTestBlock(idx, generateTestHash(idx), "")
			serialized := block.SerializeBlock()
			_ = blockchain.DeserializeBlock(serialized)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// =============================================================================
// Block Data Command Tests
// =============================================================================

func TestBlockDataCommands(t *testing.T) {
	commands := []string{
		"NEW_BLOCK",
		"done",
		"GET_BLOCKS",
	}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			blockData := &protobuf.BlockData{
				Command: cmd,
				Height:  1,
			}

			data, _ := proto.Marshal(blockData)
			decoded := &protobuf.BlockData{}
			_ = proto.Unmarshal(data, decoded)

			if decoded.Command != cmd {
				t.Errorf("Command mismatch: expected %s, got %s", cmd, decoded.Command)
			}
		})
	}
}

// =============================================================================
// Block Chain Simulation Tests
// =============================================================================

func TestBlockChainSimulation_AddBlocks(t *testing.T) {
	// Simulate adding multiple blocks
	var lastHash []byte
	blockCount := 20

	for i := 0; i < blockCount; i++ {
		block := &blockchain.Block{
			Timestamp:    int64(1000000 + i),
			Hash:         []byte{byte(i), byte(i + 1), byte(i + 2)},
			PrevHash:     lastHash,
			Height:       i,
			Nonce:        i * 1000,
			Transactions: []*blockchain.Transaction{},
			Contracts:    []*blockchain.Contract{},
		}

		// Serialize and deserialize
		serialized := block.SerializeBlock()
		recovered := blockchain.DeserializeBlock(serialized)

		if recovered.Height != i {
			t.Errorf("Block %d height mismatch", i)
		}

		lastHash = block.Hash
	}
}

func TestBlockChainSimulation_VerifyChain(t *testing.T) {
	blocks := createTestBlockchain(50)

	// Verify chain integrity
	for i := 1; i < len(blocks); i++ {
		currentBlock := blocks[i]
		prevBlock := blocks[i-1]

		if !bytes.Equal(currentBlock.PrevHash, prevBlock.Hash) {
			t.Errorf("Chain integrity broken at block %d", i)
		}

		if currentBlock.Height != i {
			t.Errorf("Block %d has wrong height: %d", i, currentBlock.Height)
		}
	}
}
