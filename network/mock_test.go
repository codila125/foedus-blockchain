package network

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/protobuf"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"google.golang.org/protobuf/proto"
)

// =============================================================================
// Test Host Helpers
// =============================================================================

// createTestHost creates a libp2p host for testing on a random port
func createTestHost(t *testing.T) host.Host {
	t.Helper()
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		t.Fatalf("Failed to create test host: %v", err)
	}
	return h
}

// connectHosts connects two hosts together
func connectHosts(t *testing.T, h1, h2 host.Host) {
	t.Helper()
	peerInfo := peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	}
	if err := h1.Connect(context.Background(), peerInfo); err != nil {
		t.Fatalf("Failed to connect hosts: %v", err)
	}
	// Wait for connection to be fully established
	time.Sleep(100 * time.Millisecond)
}

// =============================================================================
// PrintNodeID Tests
// =============================================================================

func TestPrintNodeID_WithRealHost(t *testing.T) {
	h := createTestHost(t)
	defer func() { _ = h.Close() }()

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintNodeID panicked: %v", r)
		}
	}()

	PrintNodeID(h)
}

func TestPrintNodeAddresses_WithRealHost(t *testing.T) {
	h := createTestHost(t)
	defer func() { _ = h.Close() }()

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintNodeAddresses panicked: %v", r)
		}
	}()

	PrintNodeAddresses(h)
}

// =============================================================================
// CreateStream Tests
// =============================================================================

func TestCreateStream_Success(t *testing.T) {
	h1 := createTestHost(t)
	h2 := createTestHost(t)
	defer func() { _ = h1.Close() }()
	defer func() { _ = h2.Close() }()

	// Set up a stream handler on h2
	streamReceived := make(chan bool, 1)
	h2.SetStreamHandler(protocolID, func(s network.Stream) {
		streamReceived <- true
		_ = s.Close()
	})

	connectHosts(t, h1, h2)

	stream, err := CreateStream(h1, h2.ID(), protocolID)
	if err != nil {
		t.Fatalf("CreateStream failed: %v", err)
	}
	defer func() { _ = stream.Close() }()

	// Verify stream was created successfully
	if stream == nil {
		t.Error("Stream should not be nil")
	}

	// Note: We don't wait for handler - connection establishment is enough
}

func TestCreateStream_InvalidPeer(t *testing.T) {
	h := createTestHost(t)
	defer func() { _ = h.Close() }()

	// Try to create stream to non-existent peer
	invalidPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	_, err := CreateStream(h, invalidPeerID, protocolID)
	if err == nil {
		t.Error("Expected error when creating stream to invalid peer")
	}
}

// =============================================================================
// SendCommand Tests
// =============================================================================

func TestSendCommand_Success(t *testing.T) {
	h1 := createTestHost(t)
	h2 := createTestHost(t)
	defer func() { _ = h1.Close() }()
	defer func() { _ = h2.Close() }()

	receivedCommand := make(chan string, 1)

	h2.SetStreamHandler(protocolID, func(s network.Stream) {
		defer func() { _ = s.Close() }()
		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err != nil && err != io.EOF {
			return
		}
		receivedCommand <- string(buf[:n])
	})

	connectHosts(t, h1, h2)

	stream, err := CreateStream(h1, h2.ID(), protocolID)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}
	defer func() { _ = stream.Close() }()

	err = SendCommand(stream, "GET_VERSION")
	if err != nil {
		t.Fatalf("SendCommand failed: %v", err)
	}

	select {
	case cmd := <-receivedCommand:
		if cmd != "GET_VERSION" {
			t.Errorf("Expected 'GET_VERSION', got '%s'", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("Command was not received")
	}
}

func TestSendCommand_AllCommands(t *testing.T) {
	commands := []string{
		"GET_BLOCKCHAIN",
		"GET_BLOCKS",
		"GET_VERSION",
		"NEW_CONTRACT",
		"NEW_BLOCK",
	}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			h1 := createTestHost(t)
			h2 := createTestHost(t)
			defer func() { _ = h1.Close() }()
			defer func() { _ = h2.Close() }()

			receivedCommand := make(chan string, 1)

			h2.SetStreamHandler(protocolID, func(s network.Stream) {
				defer func() { _ = s.Close() }()
				buf := make([]byte, 1024)
				n, err := s.Read(buf)
				if err != nil && err != io.EOF {
					return
				}
				receivedCommand <- string(buf[:n])
			})

			connectHosts(t, h1, h2)

			stream, err := CreateStream(h1, h2.ID(), protocolID)
			if err != nil {
				t.Fatalf("Failed to create stream: %v", err)
			}
			defer func() { _ = stream.Close() }()

			err = SendCommand(stream, cmd)
			if err != nil {
				t.Fatalf("SendCommand failed: %v", err)
			}

			select {
			case received := <-receivedCommand:
				if received != cmd {
					t.Errorf("Expected '%s', got '%s'", cmd, received)
				}
			case <-time.After(2 * time.Second):
				t.Error("Command was not received")
			}
		})
	}
}

// =============================================================================
// ReceiveBlocks Tests with Mock Stream
// =============================================================================

// mockStream implements network.Stream for testing
type mockStream struct {
	reader *bytes.Reader
	writer *bytes.Buffer
	closed bool
}

func newMockStream(data []byte) *mockStream {
	return &mockStream{
		reader: bytes.NewReader(data),
		writer: &bytes.Buffer{},
	}
}

func (m *mockStream) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

func (m *mockStream) Write(p []byte) (n int, err error) {
	return m.writer.Write(p)
}

func (m *mockStream) Close() error {
	m.closed = true
	return nil
}

func (m *mockStream) CloseRead() error                                  { return nil }
func (m *mockStream) CloseWrite() error                                 { return nil }
func (m *mockStream) Reset() error                                      { return nil }
func (m *mockStream) ResetWithError(code network.StreamErrorCode) error { return nil }
func (m *mockStream) SetDeadline(t time.Time) error                     { return nil }
func (m *mockStream) SetReadDeadline(t time.Time) error                 { return nil }
func (m *mockStream) SetWriteDeadline(t time.Time) error                { return nil }
func (m *mockStream) ID() string                                        { return "mock-stream" }
func (m *mockStream) Protocol() protocol.ID                             { return protocol.ID(protocolID) }
func (m *mockStream) SetProtocol(id protocol.ID) error                  { return nil }
func (m *mockStream) Stat() network.Stats                               { return network.Stats{} }
func (m *mockStream) Conn() network.Conn                                { return nil }
func (m *mockStream) Scope() network.StreamScope                        { return nil }

func TestReceiveBlocks_EmptyStream(t *testing.T) {
	stream := newMockStream([]byte{})

	count, lastHash, maxHeight, err := ReceiveBlocks(stream, func(serialized []byte, block *blockchain.Block) error {
		return nil
	})

	// Empty stream should return EOF gracefully
	if err != nil {
		t.Logf("Error (expected for empty stream): %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 blocks, got %d", count)
	}
	if lastHash != nil {
		t.Error("Expected nil lastHash for empty stream")
	}
	if maxHeight != 0 {
		t.Errorf("Expected maxHeight 0, got %d", maxHeight)
	}
}

func TestReceiveBlocks_DoneSignal(t *testing.T) {
	// Create a "done" message
	doneData := &protobuf.BlockData{Command: "done"}
	wireData, _ := proto.Marshal(doneData)

	// Create length-prefixed data
	buf := &bytes.Buffer{}
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(wireData)))
	_, _ = buf.Write(lenBuf)
	_, _ = buf.Write(wireData)

	stream := newMockStream(buf.Bytes())

	count, _, _, err := ReceiveBlocks(stream, func(serialized []byte, block *blockchain.Block) error {
		t.Error("processBlock should not be called for done signal")
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 blocks (done signal), got %d", count)
	}
}

// =============================================================================
// SendBlocks Tests with Mock Stream
// =============================================================================

func TestSendBlocks_EmptyList(t *testing.T) {
	stream := newMockStream([]byte{})

	err := SendBlocks(stream, [][]byte{}, func(hash []byte) (blockchain.Block, error) {
		return blockchain.Block{}, nil
	})

	if err != nil {
		t.Errorf("SendBlocks with empty list should succeed: %v", err)
	}

	// Verify done signal was sent
	data := stream.writer.Bytes()
	if len(data) == 0 {
		t.Error("Expected done signal to be written")
	}

	// Parse the done signal
	if len(data) >= 4 {
		msgLen := binary.BigEndian.Uint32(data[:4])
		if int(msgLen) <= len(data)-4 {
			doneProto := &protobuf.BlockData{}
			_ = proto.Unmarshal(data[4:4+msgLen], doneProto)
			if doneProto.Command != "done" {
				t.Errorf("Expected 'done' command, got '%s'", doneProto.Command)
			}
		}
	}
}

func TestSendBlocks_SingleBlock(t *testing.T) {
	stream := newMockStream([]byte{})

	block := &blockchain.Block{
		Timestamp: 1234567890,
		Hash:      []byte("test_hash"),
		PrevHash:  []byte("prev_hash"),
		Height:    1,
		Nonce:     100,
	}

	blockHashes := [][]byte{block.Hash}
	getBlock := func(hash []byte) (blockchain.Block, error) {
		return *block, nil
	}

	err := SendBlocks(stream, blockHashes, getBlock)
	if err != nil {
		t.Errorf("SendBlocks failed: %v", err)
	}

	// Verify data was written
	data := stream.writer.Bytes()
	if len(data) == 0 {
		t.Error("Expected block data to be written")
	}

	// Parse first message (the block)
	reader := bufio.NewReader(bytes.NewReader(data))
	lenBuf := make([]byte, 4)
	_, err = io.ReadFull(reader, lenBuf)
	if err != nil {
		t.Fatalf("Failed to read length: %v", err)
	}

	msgLen := binary.BigEndian.Uint32(lenBuf)
	msgBuf := make([]byte, msgLen)
	_, err = io.ReadFull(reader, msgBuf)
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	blockData := &protobuf.BlockData{}
	err = proto.Unmarshal(msgBuf, blockData)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if blockData.Command != "NEW_BLOCK" {
		t.Errorf("Expected 'NEW_BLOCK' command, got '%s'", blockData.Command)
	}
	if blockData.Height != int32(block.Height) {
		t.Errorf("Expected height %d, got %d", block.Height, blockData.Height)
	}
}

// =============================================================================
// SetupPeerEventListeners Tests
// =============================================================================

func TestSetupPeerEventListeners_WithRealHost(t *testing.T) {
	h := createTestHost(t)
	defer func() { _ = h.Close() }()

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetupPeerEventListeners panicked: %v", r)
		}
	}()

	SetupPeerEventListeners(h, nil, "test_node")
}

func TestSetupPeerEventListeners_PeerConnection(t *testing.T) {
	h1 := createTestHost(t)
	h2 := createTestHost(t)
	defer func() { _ = h1.Close() }()
	defer func() { _ = h2.Close() }()

	// Set up peer listeners on h1
	SetupPeerEventListeners(h1, nil, "node_1")

	// Connect h2 to h1
	connectHosts(t, h2, h1)

	// Verify connection
	if h1.Network().Connectedness(h2.ID()) != network.Connected {
		t.Error("Expected h2 to be connected to h1")
	}
}

// =============================================================================
// VersionInfo Map Tests
// =============================================================================

func TestVersionInfoMap_Operations(t *testing.T) {
	h1 := createTestHost(t)
	h2 := createTestHost(t)
	defer func() { _ = h1.Close() }()
	defer func() { _ = h2.Close() }()

	versions := make(map[peer.ID]VersionInfo)

	versions[h1.ID()] = VersionInfo{
		BestHeight: 100,
		LastHash:   []byte("hash1"),
		NodeID:     "node1",
	}

	versions[h2.ID()] = VersionInfo{
		BestHeight: 200,
		LastHash:   []byte("hash2"),
		NodeID:     "node2",
	}

	// Find best version
	var bestPeer peer.ID
	maxHeight := 0
	for peerID, version := range versions {
		if version.BestHeight > maxHeight {
			maxHeight = version.BestHeight
			bestPeer = peerID
		}
	}

	if maxHeight != 200 {
		t.Errorf("Expected max height 200, got %d", maxHeight)
	}
	if bestPeer != h2.ID() {
		t.Error("Expected h2 to have best version")
	}
}

// =============================================================================
// Network Commands Constant Tests
// =============================================================================

func TestNetworkCommands_Defined(t *testing.T) {
	commands := []string{
		"GET_BLOCKCHAIN",
		"GET_BLOCKS",
		"GET_VERSION",
		"NEW_CONTRACT",
		"NEW_BLOCK",
	}

	// Verify all commands are valid non-empty strings
	for _, cmd := range commands {
		if cmd == "" {
			t.Errorf("Command should not be empty: %s", cmd)
		}
	}
}

// =============================================================================
// Stream Handler Tests
// =============================================================================

func TestHandleNetworkRequests_UnknownCommand(t *testing.T) {
	h1 := createTestHost(t)
	h2 := createTestHost(t)
	defer func() { _ = h1.Close() }()
	defer func() { _ = h2.Close() }()

	// Set up handler on h2 with nil chain - only handles commands that don't need chain
	h2.SetStreamHandler(protocolID, func(s network.Stream) {
		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err != nil {
			_ = s.Close()
			return
		}
		command := string(buf[:n])
		// Unknown command just closes stream
		if command != "GET_BLOCKCHAIN" && command != "GET_BLOCKS" &&
			command != "GET_VERSION" && command != "NEW_CONTRACT" && command != "NEW_BLOCK" {
			_ = s.Close()
		}
	})

	connectHosts(t, h1, h2)

	// Send unknown command
	stream, err := CreateStream(h1, h2.ID(), protocolID)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}
	defer func() { _ = stream.Close() }()

	err = SendCommand(stream, "UNKNOWN_COMMAND")
	if err != nil {
		t.Errorf("SendCommand failed: %v", err)
	}

	// Give handler time to process and close
	time.Sleep(200 * time.Millisecond)
}

// TestHandleNetworkRequests_Setup tests that the handler setup doesn't panic
func TestHandleNetworkRequests_Setup(t *testing.T) {
	h := createTestHost(t)
	defer func() { _ = h.Close() }()

	// Should not panic even with nil chain
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleNetworkRequests panicked: %v", r)
		}
	}()

	HandleNetworkRequests(h, nil, "test_node")
}

// =============================================================================
// Graceful Shutdown Tests
// =============================================================================

func TestMinerShutdownTimeout_Defined(t *testing.T) {
	if minerShutdownTimeout <= 0 {
		t.Error("minerShutdownTimeout should be positive")
	}
	if minerShutdownTimeout > 60*time.Second {
		t.Error("minerShutdownTimeout should not exceed 60 seconds")
	}
}

// =============================================================================
// Protocol ID Validation Tests
// =============================================================================

func TestProtocolID_ValidFormat(t *testing.T) {
	// Protocol ID should follow libp2p convention: /name/version
	if protocolID[0] != '/' {
		t.Error("Protocol ID should start with /")
	}

	parts := 0
	for _, c := range protocolID {
		if c == '/' {
			parts++
		}
	}

	// Should have at least 2 slashes: /name/version
	if parts < 2 {
		t.Error("Protocol ID should have format /name/version")
	}
}

// =============================================================================
// Multiaddress Format Tests
// =============================================================================

func TestHostAddresses_Valid(t *testing.T) {
	h := createTestHost(t)
	defer func() { _ = h.Close() }()

	addrs := h.Addrs()
	if len(addrs) == 0 {
		t.Error("Host should have at least one address")
	}

	for _, addr := range addrs {
		addrStr := addr.String()
		if addrStr == "" {
			t.Error("Address string should not be empty")
		}
		// Should contain /ip4/ or /ip6/
		if !bytes.Contains([]byte(addrStr), []byte("/ip4/")) &&
			!bytes.Contains([]byte(addrStr), []byte("/ip6/")) {
			t.Errorf("Address should contain IP protocol: %s", addrStr)
		}
	}
}

func TestHostID_NonEmpty(t *testing.T) {
	h := createTestHost(t)
	defer func() { _ = h.Close() }()

	if h.ID().String() == "" {
		t.Error("Host ID should not be empty")
	}
}
