package network

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/database"
	"github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
)

const protocolID = "/foedus/1.0.0"

func printNodeID(host host.Host) {
	log.Printf("[NETWORK] Node ID: %s", host.ID().String())
}

func printNodeAddresses(host host.Host) {
	addressesString := make([]string, 0)
	for _, address := range host.Addrs() {
		addressesString = append(addressesString, address.String())
	}
	log.Printf("[NETWORK] Node addresses: %s", strings.Join(addressesString, ", "))
}

func GetBlockchain(node host.Host, peerID peerstore.ID, nodeID string) error {
    log.Printf("[NETWORK] Requesting blockchain from peer: %s", peerID)

    stream, err := node.NewStream(context.Background(), peerID, protocolID)
    if err != nil {
        return fmt.Errorf("failed to create stream: %w", err)
    }
    defer stream.Close()

    // Send request
    request := "GET_BLOCKCHAIN"
    _, err = stream.Write([]byte(request))
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }

    log.Printf("[NETWORK] Blockchain request sent, waiting for blocks...")

    // Initialize database for this node
    path := fmt.Sprintf(blockchain.DBPath, nodeID)
    if database.DBExists(path) {
        log.Printf("[DATABASE] Database already exists at %s, skipping sync", path)
        return nil
    }

    // Create new database to store received blocks
    db, err := database.OpenDB(path)
    if err != nil {
        return fmt.Errorf("failed to open database: %w", err)
    }

    rawDB := db.GetRawDB()
    batch := rawDB.NewBatch()

    // Receive blocks from peer
    decoder := gob.NewDecoder(stream)
    blockCount := 0
    var lastHash []byte
    maxHeight := -1

    for {
        var blockData map[string]interface{}
        if err := decoder.Decode(&blockData); err != nil {
            if strings.Contains(err.Error(), "EOF") {
                log.Printf("[NETWORK] EOF reached, blockchain sync complete")
                break
            }
            batch.Close()
            db.Close()
            return fmt.Errorf("failed to decode block: %w", err)
        }

        // Check if this is the end signal
        if command, ok := blockData["command"]; ok {
            if cmdStr, isString := command.(string); isString && cmdStr == "done" {
                log.Printf("[NETWORK] Received completion signal from peer")
                break
            }
        }

        // Extract block data
        if serializedBlock, ok := blockData["data"].([]byte); ok {
            block := blockchain.DeserializeBlock(serializedBlock)
            if block == nil {
                log.Printf("[NETWORK] Failed to deserialize block")
                continue
            }

            // Store block in batch
            err := batch.Set(block.Hash, block.SerializeBlock(), nil)
            if err != nil {
                batch.Close()
                db.Close()
                return fmt.Errorf("failed to store block in batch: %w", err)
            }

            // Track the block with the highest height
            if block.Height > maxHeight {
                maxHeight = block.Height
                lastHash = block.Hash
            }

            blockCount++

            log.Printf("[NETWORK] ← Received and stored block %x (height: %d)", block.Hash, block.Height)

            // Flush batch every 100 blocks to avoid memory overflow
            if blockCount%100 == 0 {
                err = rawDB.Apply(batch, &pebble.WriteOptions{Sync: true})
                if err != nil {
                    batch.Close()
                    db.Close()
                    return fmt.Errorf("failed to apply batch: %w", err)
                }
                log.Printf("[DATABASE] Flushed %d blocks to database", blockCount)
                batch.Close()
                batch = rawDB.NewBatch()
            }
        }
    }

    // Ensure we received at least one block
    if lastHash == nil {
        batch.Close()
        db.Close()
        return fmt.Errorf("no blocks received from peer")
    }

    // Store the last hash pointer (the block with highest height)
    err = batch.Set([]byte(blockchain.LastHashKey), lastHash, nil)
    if err != nil {
        batch.Close()
        db.Close()
        return fmt.Errorf("failed to store last hash: %w", err)
    }
    log.Printf("[DATABASE] Storing last hash: %x (height: %d)", lastHash, maxHeight)

    // Apply final batch with sync
    if err := rawDB.Apply(batch, &pebble.WriteOptions{Sync: true}); err != nil {
        batch.Close()
        db.Close()
        return fmt.Errorf("failed to apply final batch: %w", err)
    }

    batch.Close()

    // Force a manual flush to ensure all data is written
    if err := rawDB.Flush(); err != nil {
        db.Close()
        return fmt.Errorf("failed to flush database: %w", err)
    }

    log.Printf("[DATABASE] Forcing compaction...")
    if err := rawDB.Compact([]byte(blockchain.LastHashKey), []byte{0xff, 0xff, 0xff, 0xff}, true); err != nil {
        log.Printf("[DATABASE] Warning: compaction failed: %v", err)
    }

    // Verify the last hash was stored before closing
    verifyHash, closer, err := rawDB.Get([]byte(blockchain.LastHashKey))
    if err != nil {
        db.Close()
        return fmt.Errorf("verification failed - last hash not found after sync: %w", err)
    }
    log.Printf("[DATABASE] Verified last hash stored: %x", verifyHash)
    closer.Close()

    // Close database properly
    log.Printf("[DATABASE] Closing database...")
    if err := db.Close(); err != nil {
        return fmt.Errorf("failed to close database: %w", err)
    }

    // Give filesystem time to sync
    log.Printf("[DATABASE] Waiting for filesystem sync...")
    time.Sleep(500 * time.Millisecond)

    log.Printf("[DATABASE] ✓ Blockchain sync complete - %d blocks stored with max height %d", blockCount, maxHeight)
    return nil
}

func HandleGetBlockchainRequest(s network.Stream, chain *blockchain.BlockChain) {
    defer s.Close()

    encoder := gob.NewEncoder(s)

    // Get all block hashes from the blockchain
    blockHashes := chain.GetBlockHashes()
    log.Printf("[NETWORK] Sending %d blocks to peer %s", len(blockHashes), s.Conn().RemotePeer())

    // Send each block
    for _, hash := range blockHashes {
        block, err := chain.GetBlock(hash)
        if err != nil {
            log.Printf("[NETWORK] Error retrieving block %x: %v", hash, err)
            continue
        }

        blockData := map[string]interface{}{
            "command": "block",
            "hash":    block.Hash,
            "height":  block.Height,
            "data":    block.SerializeBlock(),
        }

        if err := encoder.Encode(blockData); err != nil {
            log.Printf("[NETWORK] Error sending block: %v", err)
            return
        }

        log.Printf("[NETWORK] → Sent block %x (height: %d)", hash, block.Height)

        // Add small delay to avoid overwhelming the network
        time.Sleep(10 * time.Millisecond)
    }

    // Send completion signal - use map[string]interface{} for consistency
    done := map[string]interface{}{"command": "done"}
    if err := encoder.Encode(done); err != nil {
        log.Printf("[NETWORK] Error sending completion signal: %v", err)
    }

    log.Printf("[NETWORK] Blockchain transmission complete to peer %s", s.Conn().RemotePeer())
}

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
