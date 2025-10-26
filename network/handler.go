package network

import (
	"log"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
)
func HandleNetworkRequests(s host.Host, chain *blockchain.BlockChain) {
	s.SetStreamHandler(protocolID, func(s network.Stream) {
		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err != nil {
			log.Printf("[NETWORK] Error reading from stream: %v", err)
			s.Close()
			return
		}

		command := string(buf[:n])
		log.Printf("[NETWORK] ← Received command from %s: %s", s.Conn().RemotePeer(), command)

		if command == "GET_BLOCKCHAIN" {
			HandleGetBlockchainRequest(s, chain)
		} else {
			s.Close()
		}
	})
}