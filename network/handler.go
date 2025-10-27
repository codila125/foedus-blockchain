package network

import (
	"log"
	"encoding/gob"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
)

func HandleNetworkRequests(s host.Host, chain *blockchain.BlockChain, nodeID string) {
	s.SetStreamHandler(protocolID, func(stream network.Stream) {
		buf := make([]byte, 1024)
		n, err := stream.Read(buf)
		if err != nil {
			log.Printf("[NETWORK] Error reading from stream: %v", err)
			stream.Close()
			return
		}

		command := string(buf[:n])
		log.Printf("[NETWORK] ← Received command from %s: %s", stream.Conn().RemotePeer(), command)

		switch command {
		case "GET_BLOCKCHAIN":
			HandleGetBlockchainRequest(stream, chain)
		case "GET_BLOCKS":
			HandleGetBlocksRequest(stream, chain)
		case "GET_VERSION":
			HandleGetVersionRequest(stream, chain, nodeID)
		case "NEW_CONTRACT":
			HandleReceiveContractRequest(stream, s, chain)
		case "NEW_BLOCK":
			HandleReceiveNewBlockRequest(stream, chain)
		default:
			stream.Close()
		}
	})
}

func HandleGetBlockchainRequest(s network.Stream, chain *blockchain.BlockChain) {
	defer s.Close()

	blockHashes := chain.GetBlockHashes()
	log.Printf("[NETWORK] Sending %d blocks to peer %s", len(blockHashes), s.Conn().RemotePeer())

	// Reverse block hashes to send in chronological order (genesis first)
	for i, j := 0, len(blockHashes)-1; i < j; i, j = i+1, j-1 {
		blockHashes[i], blockHashes[j] = blockHashes[j], blockHashes[i]
	}

	encoder := gob.NewEncoder(s)

	// Use modular SendBlocks function
	if err := SendBlocks(encoder, blockHashes, chain.GetBlock); err != nil {
		log.Printf("[NETWORK] Error sending blocks: %v", err)
		return
	}

	log.Printf("[NETWORK] Blockchain transmission complete")
}

// HandleGetBlocksRequest responds with only blocks after the requested height
func HandleGetBlocksRequest(s network.Stream, chain *blockchain.BlockChain) {
	defer s.Close()

	// Receive the height from requester
	decoder := gob.NewDecoder(s)
	var heightData map[string]any
	if err := decoder.Decode(&heightData); err != nil {
		log.Printf("[NETWORK] Error decoding height request: %v", err)
		return
	}

	requestedHeight, ok := heightData["height"].(int)
	if !ok {
		log.Printf("[NETWORK] Invalid height data received")
		return
	}

	log.Printf("[NETWORK] Peer %s requesting blocks after height %d", s.Conn().RemotePeer(), requestedHeight)

	// Get all block hashes
	allBlockHashes := chain.GetBlockHashes()

	// Filter to only send blocks with height > requestedHeight
	var blocksToSend [][]byte
	for _, hash := range allBlockHashes {
		block, err := chain.GetBlock(hash)
		if err != nil {
			log.Printf("[NETWORK] Error retrieving block %x: %v", hash, err)
			continue
		}

		if block.Height > requestedHeight {
			blocksToSend = append(blocksToSend, hash)
		}
	}

	log.Printf("[NETWORK] Sending %d blocks (after height %d) to peer %s", len(blocksToSend), requestedHeight, s.Conn().RemotePeer())

	encoder := gob.NewEncoder(s)

	// Send filtered blocks
	if err := SendBlocks(encoder, blocksToSend, chain.GetBlock); err != nil {
		log.Printf("[NETWORK] Error sending blocks: %v", err)
		return
	}

	log.Printf("[NETWORK] Block transmission complete")
}

// HandleGetVersionRequest responds to version requests with local blockchain info
func HandleGetVersionRequest(s network.Stream, chain *blockchain.BlockChain, nodeID string) {
	defer s.Close()

	height := chain.GetBestHeight()
	lastHash := chain.LastHash

	log.Printf("[NETWORK] Sending version info to %s (height: %d)", s.Conn().RemotePeer(), height)

	encoder := gob.NewEncoder(s)
	versionData := map[string]any{
		"height":   height,
		"lastHash": lastHash,
		"nodeID":   nodeID,
	}

	if err := encoder.Encode(versionData); err != nil {
		log.Printf("[NETWORK] Error sending version info: %v", err)
	}
}

func HandleSendContractRequest(node host.Host, contract *blockchain.Contract) {
	peers := node.Peerstore().Peers()
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

		encoder := gob.NewEncoder(stream)
		if err := SendCommand(stream, "NEW_CONTRACT"); err != nil {
			log.Printf("[NETWORK] Error sending command to %s: %v", peerID, err)
			continue
		}

		contractData := map[string]any{
			"contract": contract,
		}

		if err := encoder.Encode(contractData); err != nil {
			log.Printf("[NETWORK] Error sending contract to %s: %v", peerID, err)
		} else {
			log.Printf("[NETWORK] Sent new contract to peer %s", peerID)
		}
	}
}

func HandleReceiveContractRequest(s network.Stream, node host.Host, chain *blockchain.BlockChain) {
	defer s.Close()

	decoder := gob.NewDecoder(s)
	var contractData map[string]any
	if err := decoder.Decode(&contractData); err != nil {
		log.Printf("[NETWORK] Error decoding contract data: %v", err)
		return
	}

	contract, ok := contractData["contract"].(*blockchain.Contract)
	if !ok {
		log.Printf("[NETWORK] Invalid contract data received")
		return
	}
	log.Printf("[NETWORK] Received new contract ID %x from peer %s", contract.ID, s.Conn().RemotePeer())

	// Add contract to blockchain
	cts := []*blockchain.Contract{contract}
	block := chain.MineBlock(nil, cts)
	log.Printf("[NETWORK] New contract ID %x included in block %x", contract.ID, block.Hash)

	// Broadcast new block to all peers using dedicated function
	BroadcastBlock(node, block)
}

func HandleReceiveNewBlockRequest(s network.Stream, chain *blockchain.BlockChain) {

	decoder := gob.NewDecoder(s)
	var blockData map[string]any
	if err := decoder.Decode(&blockData); err != nil {
		log.Printf("[NETWORK] Error decoding new block data: %v", err)
		return
	}

	serializedBlock, ok := blockData["data"].([]byte)
	if !ok {
		log.Printf("[NETWORK] Invalid block data received")
		return
	}

	block := blockchain.DeserializeBlock(serializedBlock)
	if block == nil {
		log.Printf("[NETWORK] Failed to deserialize received block")
		return
	}

	// Add block to blockchain
	if err := chain.AddBlock(block); err != nil {
		log.Printf("[NETWORK] Error adding new block %x: %v", block.Hash, err)
		return
	}

	log.Printf("[NETWORK] New block %x added to blockchain from peer %s", block.Hash, s.Conn().RemotePeer())

	s.Close()
}
