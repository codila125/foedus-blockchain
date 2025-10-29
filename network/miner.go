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
	minerShutdownTimeout = 30 * time.Second
)

func createMinerNode() host.Host {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/8007",
		),
	)
	if err != nil {
		panic(err)
	}

	return node
}

func RunMinerNode(port, Address string) {
	log.Printf("[MINER] Starting node with mining enabled for address: %s", Address)
	minerNode := createMinerNode()
	PrintNodeID(minerNode)
	PrintNodeAddresses(minerNode)
	addr, err := multiaddr.NewMultiaddr(Address)
	if err != nil {
		panic(err)
	}
	peer, err := peerstore.AddrInfoFromP2pAddr(addr)
	if err != nil {
		panic(err)
	}
	if err := minerNode.Connect(context.Background(), *peer); err != nil {
		panic(err)
	}

	path := fmt.Sprintf(blockchain.DBPath, port)
	if !database.DBExists(path) {
		log.Printf("[BLOCKCHAIN] No existing blockchain found for node %s", port)
		if err := GetBlockchain(minerNode, peer.ID, port); err != nil {
			log.Panicf("[BLOCKCHAIN] Failed to sync blockchain: %v", err)
		}
		log.Printf("[BLOCKCHAIN] Successfully synchronized blockchain for node %s", port)

		// Give extra time for database to settle after sync
		log.Printf("[BLOCKCHAIN] Waiting for database to settle...")
		time.Sleep(1 * time.Second)
	} else {
		log.Printf("[BLOCKCHAIN] Existing blockchain found for node %s", port)
	}

	chain := blockchain.ContinueBlockChain(port)
	defer chain.Database.Close()

	SyncToLatestBlockchain(minerNode, chain, port)

	log.Printf("[MINER] Node is now ready to handle network requests and mine blocks")
	HandleNetworkRequests(minerNode, chain, port)

	// Set up graceful shutdown
	gracefulMinerShutdown(minerNode, chain)
}

// gracefulMinerShutdown handles signal interrupts and orchestrates resource cleanup for miner nodes
func gracefulMinerShutdown(minerNode host.Host, chain *blockchain.BlockChain) {
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	sig := <-shutdownChan
	log.Printf("[MINER] Shutting down: %v", sig)
	shutdownMinerNode(minerNode, chain)
}

// shutdownMinerNode orchestrates the graceful shutdown sequence for the miner node
func shutdownMinerNode(minerNode host.Host, chain *blockchain.BlockChain) {
	var wg sync.WaitGroup

	// Create a context with timeout for the entire shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), minerShutdownTimeout)
	defer cancel()

	log.Println("[MINER] Starting shutdown sequence...")

	// Stop accepting new network connections and clean up streams
	wg.Go(func() {
		log.Println("[MINER] Closing network node...")
		if err := minerNode.Close(); err != nil {
			log.Printf("[MINER] Error closing network node: %v", err)
		} else {
			log.Println("[MINER] Network node closed successfully")
		}
	})

	// Close database and other blockchain resources
	wg.Go(func() {
		log.Println("[MINER] Closing blockchain database...")
		if err := chain.Database.Close(); err != nil {
			log.Printf("[MINER] Error closing database: %v", err)
		} else {
			log.Println("[MINER] Blockchain database closed successfully")
		}
	})

	// Wait for all shutdown tasks to complete or timeout
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
