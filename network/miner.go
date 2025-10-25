package network

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"time"
	"github.com/codila125/foedus-blockchain/database"
	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
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

func RunMinerNode(port, walletAddress, Address string) {
    log.Printf("[MINER] Starting node with mining enabled for address: %s", Address)
    minerNode := createMinerNode()
    printNodeID(minerNode)
    printNodeAddresses(minerNode)
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
    
    log.Printf("[BLOCKCHAIN] Loading existing blockchain for node %s", port)
    chain := blockchain.ContinueBlockChain(port)
    defer chain.Database.Close()

	log.Printf("[MINER] Node is now ready to handle network requests and mine blocks")
    HandleNetworkRequests(minerNode, chain)
    
    // wait for interrupt signal to gracefully shut down the node
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
    <-ch
    log.Print("[MINER] Shutting down...")

    // shut the node down
    if err := minerNode.Close(); err != nil {
        panic(err)
    }
}
