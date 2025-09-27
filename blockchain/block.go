// Package blockchain implements a simple blockchain with basic functionalities.
package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

// type BlockChain struct {
// 	Blocks []*Block
// }

//func (b *Block) CreateHash() {
//	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
//	hash := sha256.Sum256(info)
//	b.Hash = hash[:]
//}

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

// func (blockchain *BlockChain) AddBlock(data string) {
// 	prevBlock := blockchain.Blocks[len(blockchain.Blocks)-1]
// 	newBlock := CreateBlock([]byte(data), prevBlock.Hash)
// 	blockchain.Blocks = append(blockchain.Blocks, newBlock)
// }

func (b *Block) HashTransactions() []byte {
	var transactions [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(transactions, []byte{}))
	return txHash[:]
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

// func NewBlockChain() *BlockChain {
// 	return &BlockChain{
// 		Blocks: []*Block{Genesis()},
// 	}
// }

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
