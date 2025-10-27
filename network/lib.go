// Package network handles peer-to-peer networking for the Foedus blockchain
package network

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"strings"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/database"
	"github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	protocolpkg "github.com/libp2p/go-libp2p/core/protocol"
)

const protocolID = "/foedus/1.0.0"

func init() {
	// Register types for gob encoding/decoding
	gob.Register(&blockchain.Contract{})
	gob.Register(&blockchain.Block{})
}

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
func ReceiveBlocks(decoder *gob.Decoder, processBlock func([]byte, *blockchain.Block) error) (int, []byte, int, error) {
	var blockCount int
	var lastHash []byte
	var maxHeight int

	for {
		var blockData map[string]any
		if err := decoder.Decode(&blockData); err != nil {
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			return blockCount, lastHash, maxHeight, fmt.Errorf("failed to decode block: %w", err)
		}

		// Check for completion signal
		if cmd, ok := blockData["command"].(string); ok && cmd == "done" {
			break
		}

		// Extract block
		serializedBlock, ok := blockData["data"].([]byte)
		if !ok {
			continue
		}

		block := blockchain.DeserializeBlock(serializedBlock)
		if block == nil {
			continue
		}

		// Process block using callback
		if err := processBlock(serializedBlock, block); err != nil {
			return blockCount, lastHash, maxHeight, err
		}

		// Track highest block
		if block.Height > maxHeight {
			maxHeight = block.Height
			lastHash = block.Hash
		}

		blockCount++
	}

	return blockCount, lastHash, maxHeight, nil
}

// SendBlocks encodes and sends blocks over a stream
func SendBlocks(encoder *gob.Encoder, blockHashes [][]byte, getBlock func([]byte) (blockchain.Block, error)) error {
	for _, hash := range blockHashes {
		block, err := getBlock(hash)
		if err != nil {
			log.Printf("[NETWORK] Error retrieving block %x: %v", hash, err)
			continue
		}

		blockData := map[string]any{
			"command": "NEW_BLOCK",
			"hash":    block.Hash,
			"height":  block.Height,
			"data":    block.SerializeBlock(),
		}

		if err := encoder.Encode(blockData); err != nil {
			return fmt.Errorf("failed to encode block: %w", err)
		}
	}

	// Send completion signal
	if err := encoder.Encode(map[string]any{"command": "done"}); err != nil {
		return fmt.Errorf("failed to send completion signal: %w", err)
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
	decoder := gob.NewDecoder(stream)
	blockCount, lastHash, maxHeight, err := ReceiveBlocks(decoder, processBlock)
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

		// Receive version response
		decoder := gob.NewDecoder(stream)
		var versionData map[string]any
		if err := decoder.Decode(&versionData); err != nil {
			log.Printf("[NETWORK] Error decoding version from %s: %v", peerID, err)
			stream.Close()
			continue
		}

		height, ok1 := versionData["height"].(int)
		lastHash, ok2 := versionData["lastHash"].([]byte)
		nodeID, ok3 := versionData["nodeID"].(string)

		if ok1 && ok2 && ok3 {
			versions[peerID] = VersionInfo{
				BestHeight: height,
				LastHash:   lastHash,
				NodeID:     nodeID,
			}
			log.Printf("[NETWORK] ✓ Peer %s version: height=%d, node=%s", peerID, height, nodeID)
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
	encoder := gob.NewEncoder(stream)
	heightData := map[string]any{
		"height": localHeight,
	}
	if err := encoder.Encode(heightData); err != nil {
		return fmt.Errorf("failed to send height: %w", err)
	}

	// Receive only missing blocks
	decoder := gob.NewDecoder(stream)
	var newBlocksAdded int

	for {
		var blockData map[string]any
		if err := decoder.Decode(&blockData); err != nil {
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			return fmt.Errorf("failed to decode block: %w", err)
		}

		// Check for completion signal
		if cmd, ok := blockData["command"].(string); ok && cmd == "done" {
			break
		}

		// Extract block
		serializedBlock, ok := blockData["data"].([]byte)
		if !ok {
			continue
		}

		block := blockchain.DeserializeBlock(serializedBlock)
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

		encoder := gob.NewEncoder(stream)
		blockData := map[string]any{
			"data": block.SerializeBlock(),
		}

		if err := encoder.Encode(blockData); err != nil {
			log.Printf("[NETWORK] Error broadcasting block to %s: %v", peerID, err)
		} else {
			log.Printf("[NETWORK] ✓ Broadcasted new block %x to peer %s", block.Hash, peerID)
		}
	}
}