// Package network handles peer-to-peer networking for the Foedus blockchain
package network

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/database"
	"github.com/codila125/foedus-blockchain/protobuf"
	"github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	protocolpkg "github.com/libp2p/go-libp2p/core/protocol"
	"google.golang.org/protobuf/proto"
)

const protocolID = "/foedus/1.0.0"

// PrintNodeID logs the node's unique identifier
func PrintNodeID(host host.Host) {
	log.Printf("[NETWORK] Node ID: %s", host.ID().String())
}

// PrintNodeAddresses logs all listening addresses for the node
func PrintNodeAddresses(host host.Host) {
	addressesString := make([]string, 0)
	for _, address := range host.Addrs() {
		addressesString = append(addressesString, address.String())
	}
	log.Printf("[NETWORK] Node addresses: %s", strings.Join(addressesString, ", "))
}

// CreateStream establishes a new stream to a peer with the given protocol
func CreateStream(node host.Host, peerID peerstore.ID, protocolID protocolpkg.ID) (network.Stream, error) {
	stream, err := node.NewStream(context.Background(), peerID, protocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	return stream, nil
}

// SendCommand sends a command string over a stream
func SendCommand(stream network.Stream, command string) error {
	if _, err := stream.Write([]byte(command)); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}
	return nil
}

// ReceiveBlocks decodes and processes blocks from a stream
func ReceiveBlocks(stream network.Stream, processBlock func([]byte, *blockchain.Block) error) (int, []byte, int, error) {
	var blockCount int
	var lastHash []byte
	var maxHeight int

	reader := bufio.NewReader(stream)

	for {
		// Read length prefix (4 bytes, big-endian)
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(reader, lenBuf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return blockCount, lastHash, maxHeight, fmt.Errorf("failed to read message length: %w", err)
		}

		messageLen := binary.BigEndian.Uint32(lenBuf)

		// Read the actual message
		buf := make([]byte, messageLen)
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return blockCount, lastHash, maxHeight, fmt.Errorf("failed to read message data: %w", err)
		}

		blockDataProto := &protobuf.BlockData{}
		if err := proto.Unmarshal(buf, blockDataProto); err != nil {
			return blockCount, lastHash, maxHeight, fmt.Errorf("failed to decode block: %w", err)
		}

		// Check for completion signal
		if cmd := blockDataProto.Command; cmd == "done" {
			break
		}

		// Extract block
		serializedBlock := blockDataProto.Data
		if serializedBlock == nil {
			log.Printf("[NETWORK] Invalid block data format, skipping")
			continue
		}

		block := blockchain.DeserializeBlock(serializedBlock)
		if block == nil {
			log.Printf("[NETWORK] Failed to deserialize block, skipping")
			continue
		}

		// Process block using callback
		if err := processBlock(serializedBlock, block); err != nil {
			return blockCount, lastHash, maxHeight, err
		}

		// Track the latest block (since blocks are sent in chronological order)
		if block.Height >= maxHeight {
			maxHeight = block.Height
			lastHash = block.Hash
		}

		blockCount++

		if blockCount == 1 {
			log.Printf("[NETWORK] Received genesis block (height: %d)", block.Height)
		}
	}

	return blockCount, lastHash, maxHeight, nil
}

// SendBlocks encodes and sends blocks over an existing stream
func SendBlocks(stream network.Stream, blockHashes [][]byte, getBlock func([]byte) (blockchain.Block, error)) error {
	writer := bufio.NewWriter(stream)

	for _, hash := range blockHashes {
		block, err := getBlock(hash)
		if err != nil {
			log.Printf("[NETWORK] Error retrieving block %x: %v", hash, err)
			continue
		}

		blockData := &protobuf.BlockData{
			Command: "NEW_BLOCK",
			Hash:    block.Hash,
			Height:  int32(block.Height),
			Data:    block.SerializeBlock(),
		}

		data, err := proto.Marshal(blockData)
		if err != nil {
			return fmt.Errorf("failed to marshal block: %w", err)
		}

		// Send length-prefixed block data
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
		if _, err := writer.Write(lenBuf); err != nil {
			return fmt.Errorf("failed to send block length: %w", err)
		}

		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("failed to send block data: %w", err)
		}
	}

	// Send completion signal
	doneData := &protobuf.BlockData{
		Command: "done",
	}

	data, err := proto.Marshal(doneData)
	if err != nil {
		return fmt.Errorf("failed to marshal done signal: %w", err)
	}

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	if _, err := writer.Write(lenBuf); err != nil {
		return fmt.Errorf("failed to send done signal length: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("failed to send done signal: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

func GetBlockchain(node host.Host, peerID peerstore.ID, nodeID string) error {
	log.Printf("[NETWORK] Requesting blockchain from peer: %s", peerID)

	// Create stream and send request
	stream, err := CreateStream(node, peerID, protocolID)
	if err != nil {
		return err
	}
	defer stream.Close()

	if err := SendCommand(stream, "GET_BLOCKCHAIN"); err != nil {
		return err
	}

	// Check if database already exists
	path := fmt.Sprintf(blockchain.DBPath, nodeID)
	if database.DBExists(path) {
		log.Printf("[DATABASE] Database already exists, skipping sync")
		return nil
	}

	// Open database
	db, err := database.OpenDB(path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Initialize batch writer
	batchWriter := database.NewBatchWriter(db.GetRawDB(), 100)
	defer batchWriter.Close(false)

	// Process blocks callback
	processBlock := func(serializedBlock []byte, block *blockchain.Block) error {
		return batchWriter.Write(block.Hash, serializedBlock)
	}

	// Receive and store blocks
	blockCount, lastHash, maxHeight, err := ReceiveBlocks(stream, processBlock)
	if err != nil {
		return err
	}

	if lastHash == nil {
		return fmt.Errorf("no blocks received from peer")
	}

	// Store last hash
	if err := batchWriter.Write([]byte(blockchain.LastHashKey), lastHash); err != nil {
		return err
	}

	// Final sync flush
	if err := batchWriter.Close(true); err != nil {
		return err
	}

	log.Printf("[DATABASE] Synced %d blocks (max height: %d)", blockCount, maxHeight)
	return nil
}

// VersionInfo represents the blockchain version information
type VersionInfo struct {
	BestHeight int    // Height of the latest block
	LastHash   []byte // Hash of the latest block
	NodeID     string // Identifier of the node
}

// RequestVersionFromPeers requests blockchain version from all connected peers
func RequestVersionFromPeers(node host.Host, chain *blockchain.BlockChain) map[peerstore.ID]VersionInfo {
	peers := node.Peerstore().Peers()
	versions := make(map[peerstore.ID]VersionInfo)

	localHeight := chain.GetBestHeight()

	// Count only connected peers
	connectedPeers := 0
	for _, peerID := range peers {
		if peerID != node.ID() && node.Network().Connectedness(peerID) == network.Connected {
			connectedPeers++
		}
	}

	log.Printf("[NETWORK] Requesting versions from %d connected peers (local height: %d)", connectedPeers, localHeight)

	for _, peerID := range peers {
		if peerID == node.ID() {
			continue // Skip self
		}

		// Check if peer is actually connected
		if node.Network().Connectedness(peerID) != network.Connected {
			log.Printf("[NETWORK] Skipping disconnected peer %s", peerID)
			continue
		}

		stream, err := CreateStream(node, peerID, protocolID)
		if err != nil {
			log.Printf("[NETWORK] Error creating stream to %s: %v", peerID, err)
			continue
		}

		if err := SendCommand(stream, "GET_VERSION"); err != nil {
			log.Printf("[NETWORK] Error sending GET_VERSION to %s: %v", peerID, err)
			stream.Close()
			continue
		}

		// Receive version response with length prefix
		reader := bufio.NewReader(stream)
		lenBuf := make([]byte, 4)
		_, err = io.ReadFull(reader, lenBuf)
		if err != nil {
			log.Printf("[NETWORK] Error reading version length from %s: %v", peerID, err)
			stream.Close()
			continue
		}

		messageLen := binary.BigEndian.Uint32(lenBuf)
		buf := make([]byte, messageLen)
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			log.Printf("[NETWORK] Error reading version data from %s: %v", peerID, err)
			stream.Close()
			continue
		}

		versionDataProto := &protobuf.VersionData{}
		if err := proto.Unmarshal(buf, versionDataProto); err != nil {
			log.Printf("[NETWORK] Error unmarshaling version data from %s: %v", peerID, err)
			stream.Close()
			continue
		}

		if versionDataProto.LastHash != nil && versionDataProto.NodeId != "" {
			versions[peerID] = VersionInfo{
				BestHeight: int(versionDataProto.Height),
				LastHash:   versionDataProto.LastHash,
				NodeID:     versionDataProto.NodeId,
			}
			log.Printf("[NETWORK] ✓ Peer %s version: height=%d, node=%s", peerID, versionDataProto.Height, versionDataProto.NodeId)
		}

		stream.Close()
	}

	return versions
}

// SyncToLatestBlockchain syncs the local blockchain with the peer that has the longest chain
func SyncToLatestBlockchain(node host.Host, chain *blockchain.BlockChain, nodeID string) error {
	localHeight := chain.GetBestHeight()
	versions := RequestVersionFromPeers(node, chain)

	if len(versions) == 0 {
		log.Printf("[NETWORK] No peers available for synchronization")
		return fmt.Errorf("no peers available")
	}

	// Find peer with the longest chain
	var bestPeerID peerstore.ID
	var bestVersion VersionInfo
	maxHeight := localHeight

	for peerID, version := range versions {
		if version.BestHeight > maxHeight {
			maxHeight = version.BestHeight
			bestPeerID = peerID
			bestVersion = version
		}
	}

	if maxHeight <= localHeight {
		log.Printf("[NETWORK] ✓ Local blockchain is up to date (height: %d)", localHeight)
		return nil
	}

	log.Printf("[NETWORK] Found peer %s with longer chain (height: %d vs local: %d)",
		bestPeerID, bestVersion.BestHeight, localHeight)
	log.Printf("[NETWORK] Starting synchronization from peer %s...", bestPeerID)

	// Sync missing blocks from the best peer
	if err := SyncMissingBlocks(node, bestPeerID, chain); err != nil {
		return fmt.Errorf("failed to sync blockchain from peer %s: %w", bestPeerID, err)
	}

	log.Printf("[NETWORK] ✓ Successfully synchronized to height %d", bestVersion.BestHeight)
	return nil
}

// SyncMissingBlocks requests and adds missing blocks from a peer to the existing blockchain
func SyncMissingBlocks(node host.Host, peerID peerstore.ID, chain *blockchain.BlockChain) error {
	localHeight := chain.GetBestHeight()
	log.Printf("[NETWORK] Requesting blocks after height %d from peer: %s", localHeight, peerID)

	// Create stream and send request
	stream, err := CreateStream(node, peerID, protocolID)
	if err != nil {
		return err
	}
	defer stream.Close()

	// Send command with local height
	if err := SendCommand(stream, "GET_BLOCKS"); err != nil {
		return err
	}

	// Send our current height so peer knows what to send
	heightData := &protobuf.HeightData{
		Height: int32(localHeight),
	}

	protoData, err := proto.Marshal(heightData)
	if err != nil {
		return fmt.Errorf("failed to marshal height data: %w", err)
	}

	// Send length-prefixed height data
	writer := bufio.NewWriter(stream)
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(protoData)))
	if _, err := writer.Write(lenBuf); err != nil {
		return fmt.Errorf("failed to send height length: %w", err)
	}

	if _, err := writer.Write(protoData); err != nil {
		return fmt.Errorf("failed to send height data: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush height data: %w", err)
	}

	// Receive only missing blocks
	var newBlocksAdded int
	blockReader := bufio.NewReader(stream)

	for {
		// Read length prefix (4 bytes, big-endian)
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(blockReader, lenBuf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return fmt.Errorf("failed to read block message length: %w", err)
		}

		messageLen := binary.BigEndian.Uint32(lenBuf)

		// Read the actual block message
		buf := make([]byte, messageLen)
		_, err = io.ReadFull(blockReader, buf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return fmt.Errorf("failed to read block message data: %w", err)
		}

		blockDataProto := &protobuf.BlockData{}
		if err := proto.Unmarshal(buf, blockDataProto); err != nil {
			return fmt.Errorf("failed to decode block: %w", err)
		}

		// Check for completion signal
		if cmd := blockDataProto.Command; cmd == "done" {
			break
		}

		block := blockchain.DeserializeBlock(blockDataProto.Data)
		if block == nil {
			continue
		}

		// Add block to chain
		if err := chain.AddBlock(block); err != nil {
			log.Printf("[NETWORK] Error adding block %x (height: %d): %v", block.Hash, block.Height, err)
			continue
		}

		newBlocksAdded++
		log.Printf("[NETWORK] ✓ Added block %x (height: %d)", block.Hash, block.Height)
	}

	if newBlocksAdded > 0 {
		log.Printf("[NETWORK] Successfully added %d new blocks", newBlocksAdded)
	} else {
		log.Printf("[NETWORK] No new blocks received (already synchronized)")
	}

	return nil
}

func BroadcastBlock(node host.Host, block *blockchain.Block) {
	peers := node.Peerstore().Peers()

	// Count only connected peers
	connectedPeers := 0
	for _, peerID := range peers {
		if peerID != node.ID() && node.Network().Connectedness(peerID) == network.Connected {
			connectedPeers++
		}
	}

	log.Printf("[NETWORK] Broadcasting new block %x to %d connected peers", block.Hash, connectedPeers)

	for _, peerID := range peers {
		if peerID == node.ID() {
			continue // Skip self
		}

		// Check if peer is actually connected
		if node.Network().Connectedness(peerID) != network.Connected {
			log.Printf("[NETWORK] Skipping disconnected peer %s", peerID)
			continue
		}

		stream, err := CreateStream(node, peerID, protocolID)
		if err != nil {
			log.Printf("[NETWORK] Error creating stream to %s: %v", peerID, err)
			continue
		}
		defer stream.Close()

		if err := SendCommand(stream, "NEW_BLOCK"); err != nil {
			log.Printf("[NETWORK] Error sending NEW_BLOCK command to %s: %v", peerID, err)
			continue
		}

		// Wrap block in BlockData protobuf message
		blockDataProto := &protobuf.BlockData{
			Command: "NEW_BLOCK",
			Hash:    block.Hash,
			Height:  int32(block.Height),
			Data:    block.SerializeBlock(),
		}

		data, err := proto.Marshal(blockDataProto)
		if err != nil {
			log.Printf("[NETWORK] Error marshaling block data: %v", err)
			continue
		}

		// Send length-prefixed block data
		writer := bufio.NewWriter(stream)
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
		if _, err := writer.Write(lenBuf); err != nil {
			log.Printf("[NETWORK] Error sending block length to %s: %v", peerID, err)
			continue
		}

		if _, err := writer.Write(data); err != nil {
			log.Printf("[NETWORK] Error broadcasting block to %s: %v", peerID, err)
			continue
		}

		if err := writer.Flush(); err != nil {
			log.Printf("[NETWORK] Error flushing block data to %s: %v", peerID, err)
			continue
		}

		log.Printf("[NETWORK] ✓ Broadcasted new block %x to peer %s", block.Hash, peerID)
	}
}
