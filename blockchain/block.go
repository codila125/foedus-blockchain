// Package blockchain implements a simple blockchain with basic functionalities.
package blockchain

import (
	"log"
	"time"

	"github.com/codila125/foedus-blockchain/merkle"
)

type Block struct {
	Timestamp    int64          // Timestamp of the block creation
	Hash         []byte         // Hash of the block
	Transactions []*Transaction // List of transactions included in the block
	Contracts    []*Contract    // List of contracts included in the block
	PrevHash     []byte         // Hash of the previous block
	Nonce        int            // Nonce used for mining
	Height       int            // Height of the block in the blockchain
}

func CreateBlock(txs []*Transaction, cts []*Contract, prevhash []byte, height int) *Block {
	/*
		Creates a new block with the given contracts, transactions, and previous block hash.
	*/
	block := &Block{
		Timestamp:    time.Now().Unix(),
		Transactions: txs,
		Contracts:    cts,
		PrevHash:     prevhash,
		Nonce:        0,      // Nonce is the number which will be found by Proof of Work
		Height:       height, // Height will be set when adding the block to the blockchain
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
		transactions = append(transactions, tx.ID)
	}
	if len(transactions) == 0 {
		return []byte{}
	}
	tree := merkle.NewMerkleTree(transactions) // Create a new Merkle tree from the transactions

	return tree.RootNode.Data
}

func (b *Block) HashContracts() []byte {
	/*
		Computes the Merkle root of the block's contracts.
		Returns the Merkle root as a byte slice.
	*/
	var contracts [][]byte

	// Serialize each contract and collect them
	for _, ct := range b.Contracts {
		contracts = append(contracts, ct.ID)
	}
	if len(contracts) == 0 {
		return []byte{}
	}

	tree := merkle.NewMerkleTree(contracts) // Create a new Merkle tree from the contracts

	return tree.RootNode.Data
}

func Genesis(coinbase *Transaction, contractbase *Contract) *Block {
	/*
		Creates the genesis block with a coinbase contract.
		'coinbase' is the coinbase contract that rewards the miner.
	*/
	return CreateBlock([]*Transaction{coinbase}, []*Contract{contractbase}, []byte{}, 0)
}

func Handle(err error) {
	/*
		Handles errors by logging them.
	*/
	if err != nil {
		log.Println(err)
	}
}
