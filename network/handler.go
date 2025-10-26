package network

import (
	"log"

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
