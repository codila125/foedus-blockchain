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
			"command": "block",
			"hash":    block.Hash,
			"height":  block.Height,
			"data":    block.SerializeBlock(),
		}

		if err := encoder.Encode(blockData); err != nil {
			return fmt.Errorf("failed to encode block: %w", err)
		}
	}

	// Send completion signal
	if err := encoder.Encode(map[string]interface{}{"command": "done"}); err != nil {
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

func HandleGetBlockchainRequest(s network.Stream, chain *blockchain.BlockChain) {
	defer s.Close()

	blockHashes := chain.GetBlockHashes()
	log.Printf("[NETWORK] Sending %d blocks to peer %s", len(blockHashes), s.Conn().RemotePeer())

	encoder := gob.NewEncoder(s)

	// Use modular SendBlocks function
	if err := SendBlocks(encoder, blockHashes, chain.GetBlock); err != nil {
		log.Printf("[NETWORK] Error sending blocks: %v", err)
		return
	}

	log.Printf("[NETWORK] Blockchain transmission complete")
}
