// Package network implements the peer-to-peer networking layer for the Foedus blockchain.
// It uses libp2p to handle node communication, data synchronization, and the propagation
// of new blocks and contracts across the network.
package network

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/protobuf"
	"github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
)

// HandleNetworkRequests sets up stream handlers for all supported network protocols.
// It listens for incoming messages from peers and dispatches them to the appropriate
// handler based on the command received.
func HandleNetworkRequests(h host.Host, chain *blockchain.BlockChain, nodeID string) {
	h.SetStreamHandler(protocolID, func(stream network.Stream) {
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
		// Responds with the entire local blockchain.
		case "GET_BLOCKCHAIN":
			HandleGetBlockchainRequest(h, stream, chain)
		// Responds with blocks newer than a specified height.
		case "GET_BLOCKS":
			HandleGetBlocksRequest(h, stream, chain)
		// Responds with the node's current blockchain version information.
		case "GET_VERSION":
			HandleGetVersionRequest(h, stream, chain, nodeID)
		// Handles a new contract, mines it into a block, and broadcasts the block.
		case "NEW_CONTRACT":
			HandleReceiveContractRequest(h, stream, chain)
		// Handles a new block and adds it to the local blockchain.
		case "NEW_BLOCK":
			HandleReceiveNewBlockRequest(h, stream, chain)
		default:
			stream.Close()
		}
	})
}

// HandleGetBlockchainRequest responds to a peer's request for the entire blockchain.
// It sends all block hashes in chronological order, allowing the peer to reconstruct the chain.
func HandleGetBlockchainRequest(h host.Host, s network.Stream, chain *blockchain.BlockChain) {
	defer s.Close()

	blockHashes := chain.GetBlockHashes()
	log.Printf("[NETWORK] Sending %d blocks to peer %s", len(blockHashes), s.Conn().RemotePeer())

	// The GetBlockHashes method returns hashes from newest to oldest.
	// Reverse the slice to send them in chronological order (genesis first).
	for i, j := 0, len(blockHashes)-1; i < j; i, j = i+1, j-1 {
		blockHashes[i], blockHashes[j] = blockHashes[j], blockHashes[i]
	}

	if err := SendBlocks(s, blockHashes, chain.GetBlock); err != nil {
		log.Printf("[NETWORK] Error sending blocks: %v", err)
		return
	}

	log.Printf("[NETWORK] Blockchain transmission complete")
}

// HandleGetBlocksRequest responds to a peer's request for blocks after a specific height.
// This is used for syncing a peer that is partially behind the current chain height.
func HandleGetBlocksRequest(h host.Host, s network.Stream, chain *blockchain.BlockChain) {
	defer s.Close()

	reader := bufio.NewReader(s)
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(reader, lenBuf)
	if err != nil {
		log.Printf("[NETWORK] Error reading height length from stream: %v", err)
		return
	}

	messageLen := binary.BigEndian.Uint32(lenBuf)
	buf := make([]byte, messageLen)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		log.Printf("[NETWORK] Error reading height data from stream: %v", err)
		return
	}

	heightData := &protobuf.HeightData{}
	err = proto.Unmarshal(buf, heightData)
	if err != nil {
		log.Printf("[NETWORK] Error unmarshaling height data: %v", err)
		return
	}
	requestedHeight := int(heightData.Height)

	log.Printf("[NETWORK] Peer %s requesting blocks after height %d", s.Conn().RemotePeer(), requestedHeight)

	allBlockHashes := chain.GetBlockHashes()

	// Filter the block hashes to include only those with a height greater than the peer's height.
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

	if err := SendBlocks(s, blocksToSend, chain.GetBlock); err != nil {
		log.Printf("[NETWORK] Error sending blocks: %v", err)
		return
	}

	log.Printf("[NETWORK] Block transmission complete")
}

// HandleGetVersionRequest responds to a version request from a peer. It sends the local
// blockchain's height and the hash of the latest block, allowing peers to compare chain states.
func HandleGetVersionRequest(h host.Host, s network.Stream, chain *blockchain.BlockChain, nodeID string) {
	defer s.Close()

	height := chain.GetBestHeight()
	lastHash := chain.LastHash

	log.Printf("[NETWORK] Sending version info to %s (height: %d)", s.Conn().RemotePeer(), height)

	versionData := &protobuf.VersionData{
		Height:   int32(height),
		NodeId:   nodeID,
		LastHash: lastHash,
	}

	data, err := proto.Marshal(versionData)
	if err != nil {
		log.Printf("[NETWORK] Error marshaling version info: %v", err)
		return
	}

	writer := bufio.NewWriter(s)

	// Send length-prefixed version data
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	if _, err := writer.Write(lenBuf); err != nil {
		log.Printf("[NETWORK] Error sending version info length: %v", err)
		return
	}

	if _, err := writer.Write(data); err != nil {
		log.Printf("[NETWORK] Error sending version info: %v", err)
		return
	}

	if err := writer.Flush(); err != nil {
		log.Printf("[NETWORK] Error flushing version info: %v", err)
		return
	}
}

// HandleSendContractRequest broadcasts a newly created contract to all connected peers.
// This ensures that new contracts are propagated throughout the network to be included in a future block.
func HandleSendContractRequest(node host.Host, contract *blockchain.Contract) {
	peers := node.Peerstore().Peers()
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
		defer stream.Close()

		if err := SendCommand(stream, "NEW_CONTRACT"); err != nil {
			log.Printf("[NETWORK] Error sending command to %s: %v", peerID, err)
			continue
		}

		contractData := &protobuf.Contract{
			Id:             contract.ID,
			Title:          contract.Title,
			Description:    contract.Description,
			CreatorAddress: contract.CreatorAddress,
			Attachments:    contract.Attachments,
			Terms:          contract.Terms,
			Status:         string(contract.Status),
			CreatedAt:      contract.CreatedAt,
			UpdatedAt:      contract.UpdatedAt,
			Milestones:     []*protobuf.Milestone{},
			Parties:        []*protobuf.Party{},
		}

		for _, m := range contract.Milestones {
			protoMilestone := &protobuf.Milestone{
				Id:          m.ID,
				Title:       m.Title,
				Description: m.Description,
				Value:       int32(m.Value),
				DueDate:     m.DueDate,
				Status:      string(m.Status),
				CreatedAt:   m.CreatedAt,
				CompletedAt: m.CompletedAt,
				Evidence:    m.Evidence,
				ApprovedBy:  m.ApprovedBy,
			}
			contractData.Milestones = append(contractData.Milestones, protoMilestone)
		}

		for _, p := range contract.Parties {
			protoParty := &protobuf.Party{
				Address:   p.Address,
				Role:      string(p.Role),
				PublicKey: p.PublicKey,
				Signature: p.Signature,
			}
			contractData.Parties = append(contractData.Parties, protoParty)
		}

		if data, err := proto.Marshal(contractData); err != nil {
			log.Printf("[NETWORK] Error marshaling contract data to %s: %v", peerID, err)
			continue
		} else {
			writer := bufio.NewWriter(stream)
			lenBuf := make([]byte, 4)
			binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
			if _, err := writer.Write(lenBuf); err != nil {
				log.Printf("[NETWORK] Error sending contract length to %s: %v", peerID, err)
				continue
			}

			if _, err := writer.Write(data); err != nil {
				log.Printf("[NETWORK] Error sending contract data to %s: %v", peerID, err)
				continue
			}

			if err := writer.Flush(); err != nil {
				log.Printf("[NETWORK] Error flushing contract data to %s: %v", peerID, err)
				continue
			}

			log.Printf("[NETWORK] Sent new contract to peer %s", peerID)
		}
	}
}

// HandleReceiveContractRequest processes an incoming contract from a peer. The contract is
// mined into a new block, and the new block is then broadcast to the network.
func HandleReceiveContractRequest(h host.Host, s network.Stream, chain *blockchain.BlockChain) {
	defer s.Close()

	reader := bufio.NewReader(s)
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(reader, lenBuf)
	if err != nil {
		log.Printf("[NETWORK] Error reading contract length from stream: %v", err)
		return
	}

	messageLen := binary.BigEndian.Uint32(lenBuf)
	buf := make([]byte, messageLen)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		log.Printf("[NETWORK] Error reading contract data from stream: %v", err)
		return
	}

	contract := &protobuf.Contract{}
	if err := proto.Unmarshal(buf, contract); err != nil {
		log.Printf("[NETWORK] Error unmarshaling contract data: %v", err)
		return
	}

	log.Printf("[NETWORK] Received new contract ID %x from peer %s", contract.Id, s.Conn().RemotePeer())

	ct := &blockchain.Contract{
		ID:             contract.Id,
		Title:          contract.Title,
		Description:    contract.Description,
		CreatorAddress: contract.CreatorAddress,
		Attachments:    contract.Attachments,
		Terms:          contract.Terms,
		Status:         blockchain.ContractStatus(contract.Status),
		CreatedAt:      contract.CreatedAt,
		UpdatedAt:      contract.UpdatedAt,
		Milestones:     []*blockchain.Milestone{},
		Parties:        []*blockchain.Party{},
	}

	for _, protoM := range contract.Milestones {
		m := &blockchain.Milestone{
			ID:          protoM.Id,
			Title:       protoM.Title,
			Description: protoM.Description,
			Value:       int(protoM.Value),
			DueDate:     protoM.DueDate,
			Status:      blockchain.MilestoneStatus(protoM.Status),
			CreatedAt:   protoM.CreatedAt,
			CompletedAt: protoM.CompletedAt,
			Evidence:    protoM.Evidence,
			ApprovedBy:  protoM.ApprovedBy,
		}
		ct.Milestones = append(ct.Milestones, m)
	}

	for _, protoP := range contract.Parties {
		p := &blockchain.Party{
			Address:   protoP.Address,
			Role:      blockchain.ContractRole(protoP.Role),
			PublicKey: protoP.PublicKey,
			Signature: protoP.Signature,
		}
		ct.Parties = append(ct.Parties, p)
	}

	block := chain.MineBlock(nil, []*blockchain.Contract{ct})
	log.Printf("[NETWORK] New contract ID %x included in block %x", contract.Id, block.Hash)

	BroadcastBlock(h, block)
}

// HandleReceiveNewBlockRequest processes an incoming block from a peer. It validates the
// block and, if valid, adds it to the local blockchain.
func HandleReceiveNewBlockRequest(h host.Host, s network.Stream, chain *blockchain.BlockChain) {
	defer s.Close()

	reader := bufio.NewReader(s)
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(reader, lenBuf)
	if err != nil {
		log.Printf("[NETWORK] Error reading new block length from stream: %v", err)
		return
	}

	messageLen := binary.BigEndian.Uint32(lenBuf)
	buf := make([]byte, messageLen)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		log.Printf("[NETWORK] Error reading new block data from stream: %v", err)
		return
	}

	blockDataProto := &protobuf.BlockData{}
	if err := proto.Unmarshal(buf, blockDataProto); err != nil {
		log.Printf("[NETWORK] Error unmarshaling new block data: %v", err)
		return
	}

	block := blockchain.DeserializeBlock(blockDataProto.Data)
	if block == nil {
		log.Printf("[NETWORK] Failed to deserialize received block")
		return
	}

	if err := chain.AddBlock(block); err != nil {
		log.Printf("[NETWORK] Error adding new block %x: %v", block.Hash, err)
		return
	}

	log.Printf("[NETWORK] New block %x added to blockchain from peer %s", block.Hash, s.Conn().RemotePeer())
}
