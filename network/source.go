package network

import (
	"fmt"
	"log"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
)

// createSourceNode creates and initializes a libp2p host node, which serves as the primary
// entry point for other peers to connect to the network. It listens on a well-known port.
// Returns the host node and any error encountered during creation.
func createSourceNode() (host.Host, error) {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/3009",
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create source node: %w", err)
	}

	return node, nil
}

// RunSourceNode initializes and runs the source node, which acts as a stable anchor
// in the network. It sets up peer event listeners and handles incoming network requests,
// facilitating blockchain synchronization and communication for all other nodes.
// Returns nil and logs an error if the source node cannot be created.
func RunSourceNode(chain *blockchain.BlockChain, nodeID string) host.Host {
	sourceNode, err := createSourceNode()
	if err != nil {
		log.Printf("[SOURCE NODE] Failed to create source node: %v", err)
		return nil
	}
	PrintNodeID(sourceNode)
	PrintNodeAddresses(sourceNode)
	peerInfo := peerstore.AddrInfo{
		ID:    sourceNode.ID(),
		Addrs: sourceNode.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		log.Println("Error converting AddrInfo to P2P addresses:", err)
		return nil
	}
	log.Printf("[SOURCE NODE] Source Node Multiaddresses:")
	for _, addr := range addrs {
		log.Printf("[SOURCE NODE] Address: %s", addr.String())
	}

	// Set up peer connection event listeners
	SetupPeerEventListeners(sourceNode, chain, nodeID)

	HandleNetworkRequests(sourceNode, chain, nodeID)

	return sourceNode
}
