package network

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"time"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/database"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

const (
	// minerShutdownTimeout defines the maximum duration to wait for a graceful shutdown
	// before forcing the miner node to exit.
	minerShutdownTimeout = 30 * time.Second
)

// createMinerNode initializes a new libp2p host configured for mining operations.
// It listens on a predefined TCP port and returns the configured host.
// Returns nil and error if the host cannot be created.
func createMinerNode() (host.Host, error) {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/3010",
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create miner node: %w", err)
	}

	return node, nil
}

// RunMinerNode starts and manages a miner node. It handles connecting to the
// network, synchronizing the blockchain, and processing network requests. It also
// ensures a graceful shutdown on receiving termination signals.
func RunMinerNode(port, Address string) {
	log.Printf("[MINER] Starting node with mining enabled for address: %s", Address)
	minerNode, err := createMinerNode()
	if err != nil {
		log.Printf("[MINER] CRITICAL: Failed to create miner node: %v", err)
		return
	}
	PrintNodeID(minerNode)
	PrintNodeAddresses(minerNode)
	addr, err := multiaddr.NewMultiaddr(Address)
	if err != nil {
		log.Printf("[MINER] CRITICAL: Invalid multiaddress %s: %v", Address, err)
		_ = minerNode.Close()
		return
	}
	peer, err := peerstore.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Printf("[MINER] CRITICAL: Failed to parse peer address: %v", err)
		_ = minerNode.Close()
		return
	}
	if err := minerNode.Connect(context.Background(), *peer); err != nil {
		log.Printf("[MINER] CRITICAL: Failed to connect to peer %s: %v", peer.ID, err)
		_ = minerNode.Close()
		return
	}

	path := fmt.Sprintf(blockchain.DBPath, port)
	if !database.DBExists(path) {
		log.Printf("[BLOCKCHAIN] No existing blockchain found for node %s", port)
		if err := GetBlockchain(minerNode, peer.ID, port); err != nil {
			log.Printf("[BLOCKCHAIN] Failed to sync blockchain: %v", err)
		}
		log.Printf("[BLOCKCHAIN] Successfully synchronized blockchain for node %s", port)

		// Allow time for the database to settle after initial synchronization.
		log.Printf("[BLOCKCHAIN] Waiting for database to settle...")
		time.Sleep(1 * time.Second)
	} else {
		log.Printf("[BLOCKCHAIN] Existing blockchain found for node %s", port)
	}

	chain, err := blockchain.ContinueBlockChain(port)
	if err != nil {
		log.Printf("[MINER] CRITICAL: Failed to load blockchain: %v", err)
		_ = minerNode.Close()
		return
	}
	defer func() { _ = chain.Database.Close() }()

	_ = SyncToLatestBlockchain(minerNode, chain, port)

	log.Printf("[MINER] Node is now ready to handle network requests and mine blocks")
	HandleNetworkRequests(minerNode, chain, port)

	// Set up a handler for graceful shutdown on termination signals.
	gracefulMinerShutdown(minerNode, chain)
}

// gracefulMinerShutdown listens for system signals (SIGINT, SIGTERM) and initiates
// a graceful shutdown of the miner node, ensuring all resources are released properly.
func gracefulMinerShutdown(minerNode host.Host, chain *blockchain.BlockChain) {
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	sig := <-shutdownChan
	log.Printf("[MINER] Shutting down: %v", sig)
	shutdownMinerNode(minerNode, chain)
}

// shutdownMinerNode manages the graceful shutdown of the miner node. It closes the
// libp2p host and blockchain database concurrently, with a timeout to prevent indefinite hanging.
func shutdownMinerNode(minerNode host.Host, chain *blockchain.BlockChain) {
	var wg sync.WaitGroup

	// Establish a context with a timeout for the shutdown process.
	ctx, cancel := context.WithTimeout(context.Background(), minerShutdownTimeout)
	defer cancel()

	log.Println("[MINER] Starting shutdown sequence...")

	// Concurrently close the network node to stop accepting new connections.
	wg.Go(func() {
		log.Println("[MINER] Closing network node...")
		if err := minerNode.Close(); err != nil {
			log.Printf("[MINER] Error closing network node: %v", err)
		} else {
			log.Println("[MINER] Network node closed successfully")
		}
	})

	// Concurrently close the blockchain database.
	wg.Go(func() {
		log.Println("[MINER] Closing blockchain database...")
		if err := chain.Database.Close(); err != nil {
			log.Printf("[MINER] Error closing database: %v", err)
		} else {
			log.Println("[MINER] Blockchain database closed successfully")
		}
	})

	// Wait for all shutdown operations to complete or for the timeout to be reached.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("[MINER] Shutdown completed successfully")
		os.Exit(0)
	case <-ctx.Done():
		log.Println("[MINER] Shutdown timeout exceeded, forcing exit")
		os.Exit(1)
	}
}
