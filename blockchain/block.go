// Package blockchain implements a simple blockchain with basic functionalities.
package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash         []byte         // Hash of the block
	Transactions []*Transaction // List of transactions included in the block
	PrevHash     []byte         // Hash of the previous block
	Nonce        int            // Nonce used for mining
}

func CreateBlock(txs []*Transaction, prevhash []byte) *Block {
	/*
		Creates a new block with the given transactions and previous block hash.
	*/
	block := &Block{
		Transactions: txs,
		PrevHash:     prevhash,
		Nonce:        0, // Nonce is the number which will be found by Proof of Work
	}
	pow := NewProof(block)   // Create a new Proof of Work for the block
	nonce, hash := pow.Run() // Run the Proof of Work to find a valid nonce and hash

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func (b *Block) HashTransactions() []byte {
	/*
		Computes the Merkle root of the block's transactions.
		Returns the Merkle root as a byte slice.
	*/
	var transactions [][]byte

	// Serialize each transaction and collect them
	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	tree := NewMerkleTree(transactions) // Create a new Merkle tree from the transactions

	return tree.RootNode.Data
}

func Genesis(coinbase *Transaction) *Block {
	/*
		Creates the genesis block with a coinbase transaction.
		'coinbase' is the coinbase transaction that rewards the miner.
	*/
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

func (b *Block) Serialize() []byte {
	/*
		Serializes the block into a byte slice.
	*/

	var result bytes.Buffer
	enc := gob.NewEncoder(&result)

	err := enc.Encode(b)
	Handle(err)

	return result.Bytes()
}

func Deserialize(data []byte) *Block {
	/*
		Deserializes a byte slice into a Block.
	*/
	var block Block

	dec := gob.NewDecoder(bytes.NewReader(data))

	err := dec.Decode(&block)
	Handle(err)

	return &block
}

func Handle(err error) {
	/*
		Handles errors by logging them.
	*/
	if err != nil {
		log.Panic(err)
	}
}
