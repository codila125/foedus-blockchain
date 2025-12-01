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
// Benchmarks for Serialization Operations
// =============================================================================

func BenchmarkBlockSerialization(b *testing.B) {
	block := &blockchain.Block{
		Timestamp:    1234567890,
		Hash:         []byte("test_hash_32_bytes_exactly!!!!!"),
		PrevHash:     []byte("prev_hash_32_bytes_exactly!!!!!"),
		Nonce:        12345,
		Height:       100,
		Transactions: []*blockchain.Transaction{},
		Contracts:    []*blockchain.Contract{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = block.SerializeBlock()
	}
}

func BenchmarkBlockDeserialization(b *testing.B) {
	block := &blockchain.Block{
		Timestamp:    1234567890,
		Hash:         []byte("test_hash_32_bytes_exactly!!!!!"),
		PrevHash:     []byte("prev_hash_32_bytes_exactly!!!!!"),
		Nonce:        12345,
		Height:       100,
		Transactions: []*blockchain.Transaction{},
		Contracts:    []*blockchain.Contract{},
	}
	serialized := block.SerializeBlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = blockchain.DeserializeBlock(serialized)
	}
}

func BenchmarkBlockDataProtobuf_Marshal(b *testing.B) {
	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Hash:    []byte("test_hash_32_bytes_exactly!!!!!"),
		Height:  100,
		Data:    make([]byte, 1000),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(blockData)
	}
}

func BenchmarkBlockDataProtobuf_Unmarshal(b *testing.B) {
	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Hash:    []byte("test_hash_32_bytes_exactly!!!!!"),
		Height:  100,
		Data:    make([]byte, 1000),
	}
	data, _ := proto.Marshal(blockData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoded := &protobuf.BlockData{}
		_ = proto.Unmarshal(data, decoded)
	}
}

func BenchmarkVersionDataProtobuf_Marshal(b *testing.B) {
	versionData := &protobuf.VersionData{
		Height:   100,
		LastHash: []byte("test_hash_32_bytes_exactly!!!!!"),
		NodeId:   "node_3000",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(versionData)
	}
}

func BenchmarkVersionDataProtobuf_Unmarshal(b *testing.B) {
	versionData := &protobuf.VersionData{
		Height:   100,
		LastHash: []byte("test_hash_32_bytes_exactly!!!!!"),
		NodeId:   "node_3000",
	}
	data, _ := proto.Marshal(versionData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoded := &protobuf.VersionData{}
		_ = proto.Unmarshal(data, decoded)
	}
}

func BenchmarkHeightDataProtobuf(b *testing.B) {
	heightData := &protobuf.HeightData{Height: 1000000}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := proto.Marshal(heightData)
		decoded := &protobuf.HeightData{}
		_ = proto.Unmarshal(data, decoded)
	}
}

// =============================================================================
// Benchmarks for Contract Serialization
// =============================================================================

func BenchmarkContractProtobuf_Simple(b *testing.B) {
	contract := &protobuf.Contract{
		Id:             []byte("contract_123"),
		Title:          "Test Contract",
		Description:    "Test Description",
		CreatorAddress: "creator_address",
		Status:         "active",
		CreatedAt:      1234567890,
		UpdatedAt:      1234567890,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := proto.Marshal(contract)
		decoded := &protobuf.Contract{}
		_ = proto.Unmarshal(data, decoded)
	}
}

func BenchmarkContractProtobuf_WithMilestones(b *testing.B) {
	milestones := make([]*protobuf.Milestone, 10)
	for i := 0; i < 10; i++ {
		milestones[i] = &protobuf.Milestone{
			Id:          []byte{byte(i)},
			Title:       "Milestone",
			Description: "Description",
			Value:       int32(i * 100),
			Status:      "pending",
		}
	}

	contract := &protobuf.Contract{
		Id:         []byte("contract_123"),
		Title:      "Test Contract",
		Milestones: milestones,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := proto.Marshal(contract)
		decoded := &protobuf.Contract{}
		_ = proto.Unmarshal(data, decoded)
	}
}

// =============================================================================
// Benchmarks for Length-Prefix Encoding
// =============================================================================

func BenchmarkLengthPrefixWrite(b *testing.B) {
	data := make([]byte, 1000)
	buf := &bytes.Buffer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
		buf.Write(lenBuf)
		buf.Write(data)
	}
}

func BenchmarkLengthPrefixRead(b *testing.B) {
	data := make([]byte, 1000)
	original := &bytes.Buffer{}
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	original.Write(lenBuf)
	original.Write(data)
	originalBytes := original.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(originalBytes)
		readLen := make([]byte, 4)
		_, _ = buf.Read(readLen)
		msgLen := binary.BigEndian.Uint32(readLen)
		msgBuf := make([]byte, msgLen)
		_, _ = buf.Read(msgBuf)
	}
}

// =============================================================================
// Benchmarks for Block Chain Operations
// =============================================================================

func BenchmarkBlockchainCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blocks := make([]*blockchain.Block, 100)
		var prevHash []byte
		for j := 0; j < 100; j++ {
			blocks[j] = &blockchain.Block{
				Timestamp:    int64(1000000 + j),
				Hash:         []byte{byte(j)},
				PrevHash:     prevHash,
				Height:       j,
				Transactions: []*blockchain.Transaction{},
				Contracts:    []*blockchain.Contract{},
			}
			prevHash = blocks[j].Hash
		}
	}
}

func BenchmarkVersionInfoCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VersionInfo{
			BestHeight: 100,
			LastHash:   []byte("test_hash"),
			NodeID:     "node_3000",
		}
	}
}

// =============================================================================
// Benchmarks for Large Data
// =============================================================================

func BenchmarkLargeBlockData_1KB(b *testing.B) {
	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Hash:    []byte("hash"),
		Height:  100,
		Data:    make([]byte, 1024),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := proto.Marshal(blockData)
		decoded := &protobuf.BlockData{}
		_ = proto.Unmarshal(data, decoded)
	}
}

func BenchmarkLargeBlockData_100KB(b *testing.B) {
	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Hash:    []byte("hash"),
		Height:  100,
		Data:    make([]byte, 100*1024),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := proto.Marshal(blockData)
		decoded := &protobuf.BlockData{}
		_ = proto.Unmarshal(data, decoded)
	}
}

func BenchmarkLargeBlockData_1MB(b *testing.B) {
	blockData := &protobuf.BlockData{
		Command: "NEW_BLOCK",
		Hash:    []byte("hash"),
		Height:  100,
		Data:    make([]byte, 1024*1024),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := proto.Marshal(blockData)
		decoded := &protobuf.BlockData{}
		_ = proto.Unmarshal(data, decoded)
	}
}

// =============================================================================
// Memory Allocation Benchmarks
// =============================================================================

func BenchmarkBlockDataAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blockData := &protobuf.BlockData{
			Command: "NEW_BLOCK",
			Hash:    []byte("test_hash"),
			Height:  100,
			Data:    make([]byte, 1000),
		}
		_, _ = proto.Marshal(blockData)
	}
}

func BenchmarkVersionInfoAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VersionInfo{
			BestHeight: 100,
			LastHash:   make([]byte, 32),
			NodeID:     "node_3000",
		}
	}
}

// =============================================================================
// Comparison Benchmarks
// =============================================================================

func BenchmarkBytesEqual_Small(b *testing.B) {
	a := []byte("small_hash")
	c := []byte("small_hash")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bytes.Equal(a, c)
	}
}

func BenchmarkBytesEqual_32Bytes(b *testing.B) {
	a := make([]byte, 32)
	c := make([]byte, 32)
	copy(c, a)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bytes.Equal(a, c)
	}
}

func BenchmarkBytesEqual_64Bytes(b *testing.B) {
	a := make([]byte, 64)
	c := make([]byte, 64)
	copy(c, a)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bytes.Equal(a, c)
	}
}
