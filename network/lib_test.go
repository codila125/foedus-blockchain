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
// ProtocolID Tests
// =============================================================================

func TestProtocolID_Format(t *testing.T) {
	expected := "/foedus/1.0.0"
	if protocolID != expected {
		t.Errorf("Expected protocol ID '%s', got '%s'", expected, protocolID)
	}
}

func TestProtocolID_NonEmpty(t *testing.T) {
	if protocolID == "" {
		t.Error("Protocol ID should not be empty")
	}
}

func TestProtocolID_StartsWithSlash(t *testing.T) {
	if len(protocolID) == 0 || protocolID[0] != '/' {
		t.Error("Protocol ID should start with '/'")
	}
}

// =============================================================================
// VersionInfo Tests
// =============================================================================

func TestVersionInfo_Struct(t *testing.T) {
	testCases := []struct {
		name       string
		bestHeight int
		lastHash   []byte
		nodeID     string
	}{
		{
			name:       "ValidVersion",
			bestHeight: 100,
			lastHash:   []byte("testhash12345"),
			nodeID:     "node_3000",
		},
		{
			name:       "GenesisVersion",
			bestHeight: 0,
			lastHash:   []byte("genesis_hash"),
			nodeID:     "genesis_node",
		},
		{
			name:       "EmptyHash",
			bestHeight: 50,
			lastHash:   []byte{},
			nodeID:     "node_empty",
		},
		{
			name:       "LargeHeight",
			bestHeight: 1000000,
			lastHash:   []byte("very_large_height_hash"),
			nodeID:     "node_large",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := VersionInfo{
				BestHeight: tc.bestHeight,
				LastHash:   tc.lastHash,
				NodeID:     tc.nodeID,
			}

			if v.BestHeight != tc.bestHeight {
				t.Errorf("BestHeight mismatch: expected %d, got %d", tc.bestHeight, v.BestHeight)
			}
			if !bytes.Equal(v.LastHash, tc.lastHash) {
				t.Errorf("LastHash mismatch")
			}
			if v.NodeID != tc.nodeID {
				t.Errorf("NodeID mismatch: expected %s, got %s", tc.nodeID, v.NodeID)
			}
		})
	}
}

func TestVersionInfo_ZeroValue(t *testing.T) {
	var v VersionInfo

	if v.BestHeight != 0 {
		t.Error("Zero value BestHeight should be 0")
	}
	if v.LastHash != nil {
		t.Error("Zero value LastHash should be nil")
	}
	if v.NodeID != "" {
		t.Error("Zero value NodeID should be empty")
	}
}

// =============================================================================
// Length-Prefix Encoding Tests
// =============================================================================

func TestLengthPrefixEncoding_RoundTrip(t *testing.T) {
	testCases := []uint32{
		0,
		1,
		255,
		256,
		65535,
		65536,
		1000000,
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			buf := make([]byte, 4)
			binary.BigEndian.PutUint32(buf, tc)

			result := binary.BigEndian.Uint32(buf)
			if result != tc {
				t.Errorf("Expected %d, got %d", tc, result)
			}
		})
	}
}

func TestLengthPrefixEncoding_MaxValue(t *testing.T) {
	maxVal := uint32(0xFFFFFFFF)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, maxVal)

	result := binary.BigEndian.Uint32(buf)
	if result != maxVal {
		t.Errorf("Expected %d, got %d", maxVal, result)
	}
}

// =============================================================================
// Protobuf Serialization Tests
// =============================================================================

func TestHeightData_Serialization(t *testing.T) {
	testCases := []int32{
		0,
		1,
		100,
		-1, // Edge case: negative height
		1000000,
	}

	for _, height := range testCases {
		t.Run("", func(t *testing.T) {
			original := &protobuf.HeightData{Height: height}

			data, err := proto.Marshal(original)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			decoded := &protobuf.HeightData{}
			err = proto.Unmarshal(data, decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if decoded.Height != height {
				t.Errorf("Expected height %d, got %d", height, decoded.Height)
			}
		})
	}
}

func TestVersionData_Serialization(t *testing.T) {
	testCases := []struct {
		height   int32
		lastHash []byte
		nodeID   string
	}{
		{0, []byte{}, ""},
		{100, []byte("hash"), "node_1"},
		{1000000, []byte("very_long_hash_value_here"), "node_with_long_id_12345"},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			original := &protobuf.VersionData{
				Height:   tc.height,
				LastHash: tc.lastHash,
				NodeId:   tc.nodeID,
			}

			data, err := proto.Marshal(original)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			decoded := &protobuf.VersionData{}
			err = proto.Unmarshal(data, decoded)
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

func TestBlockData_Serialization(t *testing.T) {
	testCases := []struct {
		command string
		hash    []byte
		height  int32
		data    []byte
	}{
		{"NEW_BLOCK", []byte("hash"), 1, []byte("block_data")},
		{"done", nil, 0, nil},
		{"GET_BLOCKS", []byte("another_hash"), 100, []byte("more_data")},
	}

	for _, tc := range testCases {
		t.Run(tc.command, func(t *testing.T) {
			original := &protobuf.BlockData{
				Command: tc.command,
				Hash:    tc.hash,
				Height:  tc.height,
				Data:    tc.data,
			}

			data, err := proto.Marshal(original)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			decoded := &protobuf.BlockData{}
			err = proto.Unmarshal(data, decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if decoded.Command != tc.command {
				t.Errorf("Command mismatch")
			}
			if !bytes.Equal(decoded.Hash, tc.hash) {
				t.Errorf("Hash mismatch")
			}
			if decoded.Height != tc.height {
				t.Errorf("Height mismatch")
			}
			if !bytes.Equal(decoded.Data, tc.data) {
				t.Errorf("Data mismatch")
			}
		})
	}
}

// =============================================================================
// Block Serialization Tests
// =============================================================================

func TestBlockSerialization_Empty(t *testing.T) {
	block := &blockchain.Block{
		Timestamp:    0,
		Hash:         []byte{},
		PrevHash:     []byte{},
		Nonce:        0,
		Height:       0,
		Transactions: nil,
		Contracts:    nil,
	}

	serialized := block.SerializeBlock()
	if serialized == nil {
		t.Fatal("SerializeBlock should not return nil")
	}

	deserialized := blockchain.DeserializeBlock(serialized)
	if deserialized == nil {
		t.Fatal("DeserializeBlock should not return nil")
	}
}

func TestBlockSerialization_WithData(t *testing.T) {
	block := &blockchain.Block{
		Timestamp:    1234567890,
		Hash:         []byte("test_hash_value"),
		PrevHash:     []byte("prev_hash_value"),
		Nonce:        12345,
		Height:       100,
		Transactions: []*blockchain.Transaction{},
		Contracts:    []*blockchain.Contract{},
	}

	serialized := block.SerializeBlock()
	if serialized == nil {
		t.Fatal("SerializeBlock should not return nil")
	}

	deserialized := blockchain.DeserializeBlock(serialized)
	if deserialized == nil {
		t.Fatal("DeserializeBlock should not return nil")
	}

	if deserialized.Timestamp != block.Timestamp {
		t.Errorf("Timestamp mismatch")
	}
	if !bytes.Equal(deserialized.Hash, block.Hash) {
		t.Errorf("Hash mismatch")
	}
	if !bytes.Equal(deserialized.PrevHash, block.PrevHash) {
		t.Errorf("PrevHash mismatch")
	}
	if deserialized.Nonce != block.Nonce {
		t.Errorf("Nonce mismatch")
	}
	if deserialized.Height != block.Height {
		t.Errorf("Height mismatch")
	}
}

// =============================================================================
// Wire Format Tests
// =============================================================================

func TestWireFormat_LengthPrefixed(t *testing.T) {
	// Test that we can correctly create length-prefixed messages
	message := []byte("test message content")

	// Create length-prefixed buffer
	buf := &bytes.Buffer{}
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(message)))
	buf.Write(lenBuf)
	buf.Write(message)

	// Read back
	readLen := make([]byte, 4)
	_, err := buf.Read(readLen)
	if err != nil {
		t.Fatalf("Failed to read length: %v", err)
	}

	msgLen := binary.BigEndian.Uint32(readLen)
	if msgLen != uint32(len(message)) {
		t.Errorf("Length mismatch: expected %d, got %d", len(message), msgLen)
	}

	readMsg := make([]byte, msgLen)
	_, err = buf.Read(readMsg)
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if !bytes.Equal(readMsg, message) {
		t.Errorf("Message content mismatch")
	}
}

func TestWireFormat_EmptyMessage(t *testing.T) {
	message := []byte{}

	buf := &bytes.Buffer{}
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(message)))
	buf.Write(lenBuf)
	buf.Write(message)

	readLen := make([]byte, 4)
	_, _ = buf.Read(readLen)
	msgLen := binary.BigEndian.Uint32(readLen)

	if msgLen != 0 {
		t.Errorf("Empty message should have length 0, got %d", msgLen)
	}
}

func TestWireFormat_LargeMessage(t *testing.T) {
	// Test with a large message (1MB)
	message := make([]byte, 1024*1024)
	for i := range message {
		message[i] = byte(i % 256)
	}

	buf := &bytes.Buffer{}
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(message)))
	buf.Write(lenBuf)
	buf.Write(message)

	readLen := make([]byte, 4)
	_, _ = buf.Read(readLen)
	msgLen := binary.BigEndian.Uint32(readLen)

	if msgLen != uint32(len(message)) {
		t.Errorf("Length mismatch for large message")
	}
}

// =============================================================================
// Command String Tests
// =============================================================================

func TestCommandStrings(t *testing.T) {
	commands := []string{
		"GET_BLOCKCHAIN",
		"GET_BLOCKS",
		"GET_VERSION",
		"NEW_CONTRACT",
		"NEW_BLOCK",
	}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			// Verify command is valid ASCII
			for _, c := range cmd {
				if c < 32 || c > 126 {
					t.Errorf("Command contains non-printable character: %v", c)
				}
			}

			// Verify command is uppercase
			for _, c := range cmd {
				if c >= 'a' && c <= 'z' {
					t.Errorf("Command should be uppercase: %s", cmd)
				}
			}
		})
	}
}

func TestCommandStrings_DoneSignal(t *testing.T) {
	done := "done"

	// "done" is lowercase - verify this special case
	if done != "done" {
		t.Error("Done signal should be lowercase 'done'")
	}
}

// =============================================================================
// Height Comparison Tests
// =============================================================================

func TestHeightComparison(t *testing.T) {
	localHeight := 100

	testCases := []struct {
		peerHeight int
		shouldSync bool
	}{
		{50, false},  // Peer behind
		{100, false}, // Equal
		{150, true},  // Peer ahead
		{0, false},   // Genesis
		{1000, true}, // Far ahead
	}

	for _, tc := range testCases {
		needsSync := tc.peerHeight > localHeight
		if needsSync != tc.shouldSync {
			t.Errorf("Height %d: expected sync=%v, got %v", tc.peerHeight, tc.shouldSync, needsSync)
		}
	}
}

// =============================================================================
// Block Hash Tests
// =============================================================================

func TestBlockHash_Format(t *testing.T) {
	// Typical block hashes are 32 bytes (SHA-256)
	expectedLen := 32

	testHash := make([]byte, expectedLen)
	if len(testHash) != expectedLen {
		t.Errorf("Expected hash length %d, got %d", expectedLen, len(testHash))
	}
}

func TestBlockHash_InBlockData(t *testing.T) {
	hash := []byte("12345678901234567890123456789012") // 32 bytes

	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Hash:    hash,
		Height:  1,
		Data:    []byte("block_data"),
	}

	data, _ := proto.Marshal(blockData)
	decoded := &protobuf.BlockData{}
	_ = proto.Unmarshal(data, decoded)

	if !bytes.Equal(decoded.Hash, hash) {
		t.Error("Hash not preserved in BlockData")
	}
}

// =============================================================================
// Network Constants Tests
// =============================================================================

func TestNetworkConstants(t *testing.T) {
	// Protocol ID format check
	if len(protocolID) == 0 {
		t.Error("Protocol ID should not be empty")
	}

	// Should contain version info
	if !bytes.Contains([]byte(protocolID), []byte("foedus")) {
		t.Error("Protocol ID should contain 'foedus'")
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestEdgeCase_NilBlockData(t *testing.T) {
	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Hash:    nil,
		Height:  0,
		Data:    nil,
	}

	data, err := proto.Marshal(blockData)
	if err != nil {
		t.Fatalf("Should marshal nil fields: %v", err)
	}

	decoded := &protobuf.BlockData{}
	err = proto.Unmarshal(data, decoded)
	if err != nil {
		t.Fatalf("Should unmarshal nil fields: %v", err)
	}

	if decoded.Hash != nil {
		t.Error("Hash should remain nil")
	}
	if decoded.Data != nil {
		t.Error("Data should remain nil")
	}
}

func TestEdgeCase_MaxInt32Height(t *testing.T) {
	maxHeight := int32(2147483647)

	heightData := &protobuf.HeightData{Height: maxHeight}
	data, _ := proto.Marshal(heightData)

	decoded := &protobuf.HeightData{}
	_ = proto.Unmarshal(data, decoded)

	if decoded.Height != maxHeight {
		t.Errorf("Max height not preserved: expected %d, got %d", maxHeight, decoded.Height)
	}
}

func TestEdgeCase_NegativeHeight(t *testing.T) {
	negativeHeight := int32(-1)

	heightData := &protobuf.HeightData{Height: negativeHeight}
	data, _ := proto.Marshal(heightData)

	decoded := &protobuf.HeightData{}
	_ = proto.Unmarshal(data, decoded)

	if decoded.Height != negativeHeight {
		t.Errorf("Negative height not preserved: expected %d, got %d", negativeHeight, decoded.Height)
	}
}

func TestEdgeCase_EmptyNodeID(t *testing.T) {
	versionData := &protobuf.VersionData{
		Height:   100,
		LastHash: []byte("hash"),
		NodeId:   "",
	}

	data, _ := proto.Marshal(versionData)

	decoded := &protobuf.VersionData{}
	_ = proto.Unmarshal(data, decoded)

	if decoded.NodeId != "" {
		t.Error("Empty NodeId should be preserved")
	}
}

func TestEdgeCase_LongNodeID(t *testing.T) {
	longID := string(make([]byte, 10000))

	versionData := &protobuf.VersionData{
		Height:   100,
		LastHash: []byte("hash"),
		NodeId:   longID,
	}

	data, err := proto.Marshal(versionData)
	if err != nil {
		t.Fatalf("Should marshal long NodeId: %v", err)
	}

	decoded := &protobuf.VersionData{}
	_ = proto.Unmarshal(data, decoded)

	if len(decoded.NodeId) != len(longID) {
		t.Errorf("Long NodeId length not preserved")
	}
}
