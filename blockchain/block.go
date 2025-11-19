// Package blockchain defines the fundamental data structures and core logic for
// the Foedus blockchain. It includes the implementation of blocks, which are
// the primary building units of the chain, and the mechanisms for their creation
// and validation.
package blockchain

import (
	"log"
	"time"
)

// Block represents a single unit in the blockchain. It contains a collection of
// transactions and smart contracts, a timestamp marking its creation time, and
// cryptographic hashes that link it to the preceding block. The block's integrity
// is secured by a proof-of-work nonce.
type Block struct {
	Timestamp    int64          // The time at which the block was created, in Unix format.
	Hash         []byte         // The cryptographic hash of the block's contents.
	Transactions []*Transaction // A list of transactions included in this block.
	Contracts    []*Contract    // A list of smart contracts included in this block.
	PrevHash     []byte         // The hash of the previous block in the chain, ensuring continuity.
	Nonce        int            // The nonce value found during the proof-of-work mining process.
	Height       int            // The block's position in the blockchain, also known as block number.
}

// CreateBlock constructs a new block, populates it with transactions and contracts,
// and links it to the previous block in the chain. It initiates the proof-of-work
// algorithm to find a valid hash and nonce for the block, thereby securing it.
// The timestamp is obtained from the current system time and is deterministic
// within the scope of a single block creation. For strict determinism across
// nodes with different clocks, use CreateBlockWithTimestamp instead.
func CreateBlock(txs []*Transaction, cts []*Contract, prevhash []byte, height int) *Block {
	return CreateBlockWithTimestamp(txs, cts, prevhash, height, time.Now().Unix())
}

// CreateBlockWithTimestamp constructs a new block with an explicitly provided timestamp.
// This enables deterministic block creation by allowing callers (such as consensus
// mechanisms) to control the timestamp value. This is essential for network consensus,
// as it prevents timestamp variations due to different system clocks across nodes.
// For mining/mempool scenarios, use CreateBlock instead, which uses the current time.
func CreateBlockWithTimestamp(txs []*Transaction, cts []*Contract, prevhash []byte, height int, timestamp int64) *Block {
	block := &Block{
		Timestamp:    timestamp,
		Transactions: txs,
		Contracts:    cts,
		PrevHash:     prevhash,
		Nonce:        0,
		Height:       height,
	}
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// Genesis generates the very first block in the blockchain, known as the genesis
// block. This block is unique as it has no preceding block. It is initialized
// with a coinbase transaction to reward the first miner and a base contract if applicable.
func Genesis(coinbase *Transaction, contractbase *Contract) *Block {
	return CreateBlock([]*Transaction{coinbase}, []*Contract{contractbase}, []byte{}, 0)
}

// Handle provides a standardized way of logging errors that occur within the
// blockchain package. It ensures that all errors are consistently reported,
// which simplifies debugging and monitoring.
func Handle(err error) {
	if err != nil {
		log.Println(err)
	}
}
