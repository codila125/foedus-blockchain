package network

import (
	"log"
	"time"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
	multiaddr "github.com/multiformats/go-multiaddr"
)

// PeerNotifee implements the libp2p network.Notifiee interface to handle peer-related events,
// such as connections and disconnections. This allows the node to react dynamically to changes
// in network topology.
type PeerNotifee struct {
	host   host.Host
	chain  *blockchain.BlockChain
	nodeID string
}

// Listen is called when the network begins listening on a new multiaddress.
// This implementation is a no-op as no action is needed on this event.
func (p *PeerNotifee) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when the network stops listening on a multiaddress.
// This implementation is a no-op as no action is needed on this event.
func (p *PeerNotifee) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a new connection to a peer is established. It logs the event
// and triggers a blockchain synchronization check to ensure the local node is up-to-date.
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

// Disconnected is called when a connection to a peer is closed. It logs the event
// to provide visibility into network churn.
func (p *PeerNotifee) Disconnected(net network.Network, conn network.Conn) {
	remotePeer := conn.RemotePeer()

	log.Printf("[SOURCE NODE] ✗ Peer disconnected: %s", remotePeer.String())
	log.Printf("[SOURCE NODE]   Total peers: %d", len(p.host.Network().Peers()))
}

// OpenedStream is called when a new stream is opened to a peer.
// This implementation is a no-op as no action is needed on this event.
func (p *PeerNotifee) OpenedStream(net network.Network, stream network.Stream) {}

// ClosedStream is called when a stream to a peer is closed.
// This implementation is a no-op as no action is needed on this event.
func (p *PeerNotifee) ClosedStream(net network.Network, stream network.Stream) {}

// SetupPeerEventListeners registers a notifiee to monitor peer connection and disconnection events in the network.
func SetupPeerEventListeners(h host.Host, chain *blockchain.BlockChain, nodeID string) {
	notifee := &PeerNotifee{
		host:   h,
		chain:  chain,
		nodeID: nodeID,
	}
	h.Network().Notify(notifee)
}
