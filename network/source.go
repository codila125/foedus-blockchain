package network

import (
	"log"
	"time"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

func createSourceNode() host.Host {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/8006",
		),
	)
	if err != nil {
		panic(err)
	}

	return node
}

func RunSourceNode(chain *blockchain.BlockChain, nodeID string) host.Host {
	sourceNode := createSourceNode()
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

// SetupPeerEventListeners sets up listeners for peer connect/disconnect events
func SetupPeerEventListeners(h host.Host, chain *blockchain.BlockChain, nodeID string) {
	notifee := &PeerNotifee{
		host:   h,
		chain:  chain,
		nodeID: nodeID,
	}
	h.Network().Notify(notifee)
}

// PeerNotifee implements network.Notifiee to listen for peer events
type PeerNotifee struct {
	host   host.Host
	chain  *blockchain.BlockChain
	nodeID string
}

// Listen is called when the network starts listening on a new multiaddr
func (p *PeerNotifee) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when the network stops listening on a multiaddr
func (p *PeerNotifee) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection is established
func (p *PeerNotifee) Connected(net network.Network, conn network.Conn) {
	remotePeer := conn.RemotePeer()
	remoteAddr := conn.RemoteMultiaddr()

	log.Printf("[SOURCE NODE] ✓ Peer connected: %s", remotePeer.String())
	log.Printf("[SOURCE NODE]   Remote Address: %s", remoteAddr.String())
	log.Printf("[SOURCE NODE]   Total peers: %d", len(p.host.Network().Peers()))

	// Check version and sync if peer has longer chain
	// Add delay to allow peer's stream handler to initialize
	if p.chain != nil && p.nodeID != "" {
		go func() {
			// Wait for peer to be fully ready (stream handler initialized)
			time.Sleep(2 * time.Second)

			// Check if peer is still connected
			if p.host.Network().Connectedness(remotePeer) != network.Connected {
				log.Printf("[SOURCE NODE] Peer %s disconnected before sync check", remotePeer)
				return
			}

			if err := SyncToLatestBlockchain(p.host, p.chain, p.nodeID); err != nil {
				log.Printf("[SOURCE NODE] Sync check failed: %v", err)
			}
		}()
	}
}

// Disconnected is called when a connection is closed
func (p *PeerNotifee) Disconnected(net network.Network, conn network.Conn) {
	remotePeer := conn.RemotePeer()

	log.Printf("[SOURCE NODE] ✗ Peer disconnected: %s", remotePeer.String())
	log.Printf("[SOURCE NODE]   Total peers: %d", len(p.host.Network().Peers()))
}

// OpenedStream is called when a stream is opened
func (p *PeerNotifee) OpenedStream(net network.Network, stream network.Stream) {}

// ClosedStream is called when a stream is closed
func (p *PeerNotifee) ClosedStream(net network.Network, stream network.Stream) {}
