// Package network handles peer-to-peer networking for the Foedus blockchain,
// including block synchronization, broadcasting, and peer management.
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

// protocolID is the unique identifier for the Foedus blockchain protocol.
const protocolID = "/foedus/1.0.0"

// PrintNodeID logs the unique identifier (Peer ID) of the libp2p node.
// This is useful for debugging and identifying the node on the network.
func PrintNodeID(host host.Host) {
	log.Printf("[NETWORK] Node ID: %s", host.ID().String())
}

// PrintNodeAddresses logs all multiaddresses the node is listening on.
// These addresses can be used by other peers to connect to this node.
func PrintNodeAddresses(host host.Host) {
	addressesString := make([]string, 0)
	for _, address := range host.Addrs() {
		addressesString = append(addressesString, address.String())
	}
	log.Printf("[NETWORK] Node addresses: %s", strings.Join(addressesString, ", "))
}

// CreateStream establishes a new bidirectional stream to a peer using a specific protocol.
// It returns the stream handle or an error if the stream could not be established.
func CreateStream(node host.Host, peerID peerstore.ID, protocolID protocolpkg.ID) (network.Stream, error) {
	stream, err := node.NewStream(context.Background(), peerID, protocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	return stream, nil
}

// SendCommand sends a command as a simple string over a network stream.
// It returns an error if the write operation fails.
func SendCommand(stream network.Stream, command string) error {
	if _, err := stream.Write([]byte(command)); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}
	return nil
}

// ReceiveBlocks reads length-prefixed block messages from a stream, deserializes them,
// and processes them using a callback. It returns the number of blocks received,
// the hash of the last block, the maximum height, and any error encountered.
func ReceiveBlocks(stream network.Stream, processBlock func([]byte, *blockchain.Block) error) (int, []byte, int, error) {
	var blockCount int
	var lastHash []byte
	var maxHeight int

	reader := bufio.NewReader(stream)

	for {
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(reader, lenBuf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return blockCount, lastHash, maxHeight, fmt.Errorf("failed to read message length: %w", err)
		}

		messageLen := binary.BigEndian.Uint32(lenBuf)

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

		if cmd := blockDataProto.Command; cmd == "done" {
			break
		}

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

		if err := processBlock(serializedBlock, block); err != nil {
			return blockCount, lastHash, maxHeight, err
		}

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

// SendBlocks serializes and sends a list of blocks over a network stream.
// Each block is retrieved using a callback and sent as a length-prefixed protobuf message.
// It sends a "done" signal after all blocks have been sent.
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

		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
		if _, err := writer.Write(lenBuf); err != nil {
			return fmt.Errorf("failed to send block length: %w", err)
		}

		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("failed to send block data: %w", err)
		}
	}

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

// GetBlockchain requests and downloads the entire blockchain from a specified peer.
// This is typically used for initial synchronization when a node has no local blockchain.
// It stores the received blocks in the database.
func GetBlockchain(node host.Host, peerID peerstore.ID, nodeID string) error {
	log.Printf("[NETWORK] Requesting blockchain from peer: %s", peerID)

	stream, err := CreateStream(node, peerID, protocolID)
	if err != nil {
		return err
	}
	defer func() { _ = stream.Close() }()

	if err := SendCommand(stream, "GET_BLOCKCHAIN"); err != nil {
		return err
	}

	path := fmt.Sprintf(blockchain.DBPath, nodeID)
	if database.DBExists(path) {
		log.Printf("[DATABASE] Database already exists, skipping sync")
		return nil
	}

	db, err := database.OpenDB(path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	batchWriter := database.NewBatchWriter(db.GetRawDB(), 100)
	defer func() { _ = batchWriter.Close(false) }()

	processBlock := func(serializedBlock []byte, block *blockchain.Block) error {
		return batchWriter.Write(block.Hash, serializedBlock)
	}

	blockCount, lastHash, maxHeight, err := ReceiveBlocks(stream, processBlock)
	if err != nil {
		return err
	}

	if lastHash == nil {
		return fmt.Errorf("no blocks received from peer")
	}

	if err := batchWriter.Write([]byte(blockchain.LastHashKey), lastHash); err != nil {
		return err
	}

	if err := batchWriter.Close(true); err != nil {
		return err
	}

	log.Printf("[DATABASE] Synced %d blocks (max height: %d)", blockCount, maxHeight)
	return nil
}

// VersionInfo encapsulates a peer's blockchain height, last block hash, and node ID.
type VersionInfo struct {
	BestHeight int
	LastHash   []byte
	NodeID     string
}

// RequestVersionFromPeers queries all connected peers for their blockchain version information.
// It returns a map of peer IDs to their respective VersionInfo.
func RequestVersionFromPeers(node host.Host, chain *blockchain.BlockChain) map[peerstore.ID]VersionInfo {
	peers := node.Peerstore().Peers()
	versions := make(map[peerstore.ID]VersionInfo)

	localHeight := chain.GetBestHeight()

	connectedPeers := 0
	for _, peerID := range peers {
		if peerID != node.ID() && node.Network().Connectedness(peerID) == network.Connected {
			connectedPeers++
		}
	}

	log.Printf("[NETWORK] Requesting versions from %d connected peers (local height: %d)", connectedPeers, localHeight)

	for _, peerID := range peers {
		if peerID == node.ID() {
			continue
		}

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
			_ = stream.Close()
			continue
		}

		reader := bufio.NewReader(stream)
		lenBuf := make([]byte, 4)
		_, err = io.ReadFull(reader, lenBuf)
		if err != nil {
			log.Printf("[NETWORK] Error reading version length from %s: %v", peerID, err)
			_ = stream.Close()
			continue
		}

		messageLen := binary.BigEndian.Uint32(lenBuf)
		buf := make([]byte, messageLen)
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			log.Printf("[NETWORK] Error reading version data from %s: %v", peerID, err)
			_ = stream.Close()
			continue
		}

		versionDataProto := &protobuf.VersionData{}
		if err := proto.Unmarshal(buf, versionDataProto); err != nil {
			log.Printf("[NETWORK] Error unmarshaling version data from %s: %v", peerID, err)
			_ = stream.Close()
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

		_ = stream.Close()
	}

	return versions
}

// SyncToLatestBlockchain identifies the peer with the highest blockchain and synchronizes with it.
// If the local blockchain is already up-to-date, it does nothing.
func SyncToLatestBlockchain(node host.Host, chain *blockchain.BlockChain, nodeID string) error {
	localHeight := chain.GetBestHeight()
	versions := RequestVersionFromPeers(node, chain)

	if len(versions) == 0 {
		log.Printf("[NETWORK] No peers available for synchronization")
		return fmt.Errorf("no peers available")
	}

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

	if err := SyncMissingBlocks(node, bestPeerID, chain); err != nil {
		return fmt.Errorf("failed to sync blockchain from peer %s: %w", bestPeerID, err)
	}

	log.Printf("[NETWORK] ✓ Successfully synchronized to height %d", bestVersion.BestHeight)
	return nil
}

// SyncMissingBlocks requests and adds blocks from a peer that are missing from the local blockchain.
// It starts requesting blocks after the current best height of the local chain.
func SyncMissingBlocks(node host.Host, peerID peerstore.ID, chain *blockchain.BlockChain) error {
	localHeight := chain.GetBestHeight()
	log.Printf("[NETWORK] Requesting blocks after height %d from peer: %s", localHeight, peerID)

	stream, err := CreateStream(node, peerID, protocolID)
	if err != nil {
		return err
	}
	defer func() { _ = stream.Close() }()

	if err := SendCommand(stream, "GET_BLOCKS"); err != nil {
		return err
	}

	heightData := &protobuf.HeightData{
		Height: int32(localHeight),
	}

	protoData, err := proto.Marshal(heightData)
	if err != nil {
		return fmt.Errorf("failed to marshal height data: %w", err)
	}

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

	var newBlocksAdded int
	blockReader := bufio.NewReader(stream)

	for {
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(blockReader, lenBuf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return fmt.Errorf("failed to read block message length: %w", err)
		}

		messageLen := binary.BigEndian.Uint32(lenBuf)

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

		if cmd := blockDataProto.Command; cmd == "done" {
			break
		}

		block := blockchain.DeserializeBlock(blockDataProto.Data)
		if block == nil {
			continue
		}

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

// BroadcastBlock sends a newly mined block to all connected peers in the network.
// This ensures that all nodes are aware of the latest block.
func BroadcastBlock(node host.Host, block *blockchain.Block) {
	peers := node.Peerstore().Peers()

	connectedPeers := 0
	for _, peerID := range peers {
		if peerID != node.ID() && node.Network().Connectedness(peerID) == network.Connected {
			connectedPeers++
		}
	}

	log.Printf("[NETWORK] Broadcasting new block %x to %d connected peers", block.Hash, connectedPeers)

	for _, peerID := range peers {
		if peerID == node.ID() {
			continue
		}

		if node.Network().Connectedness(peerID) != network.Connected {
			log.Printf("[NETWORK] Skipping disconnected peer %s", peerID)
			continue
		}

		stream, err := CreateStream(node, peerID, protocolID)
		if err != nil {
			log.Printf("[NETWORK] Error creating stream to %s: %v", peerID, err)
			continue
		}
		defer func() { _ = stream.Close() }()

		if err := SendCommand(stream, "NEW_BLOCK"); err != nil {
			log.Printf("[NETWORK] Error sending NEW_BLOCK command to %s: %v", peerID, err)
			continue
		}

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
