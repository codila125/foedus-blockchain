// Package blockchain implements a simple blockchain with basic functionalities.
package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

func CreateBlock(txs []*Transaction, prevhash []byte) *Block {
	block := &Block{
		Transactions: txs,
		PrevHash:     prevhash,
		Nonce:        0,
	}
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	tree := NewMerkleTree(transactions)

	return tree.RootNode.Data
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	enc := gob.NewEncoder(&result)

	err := enc.Encode(b)
	Handle(err)

	return result.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	dec := gob.NewDecoder(bytes.NewReader(data))

	err := dec.Decode(&block)
	Handle(err)

	return &block
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
