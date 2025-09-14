package blockchain

import (
	"bytes"
	"crypto/sha256"
)

type Block struct {
	Hash	 []byte
	Data     []byte
	PrevHash []byte
}

type BlockChain struct {
	Blocks []*Block
}

func (b *Block) CreateHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}

func CreateBlock(data, prevhash []byte) *Block {
	block := &Block{
		Data:    data,
		PrevHash: prevhash,
	}
	block.CreateHash()
	return block
}

func (blockchain *BlockChain) AddBlock(data string) {
	prevBlock := blockchain.Blocks[len(blockchain.Blocks)-1]
	newBlock := CreateBlock([]byte(data), prevBlock.Hash)
	blockchain.Blocks = append(blockchain.Blocks, newBlock)
}

func Genesis() *Block {
	return CreateBlock([]byte("Genesis Block"), []byte{})
}

func NewBlockChain() *BlockChain{
	return &BlockChain{
		Blocks: []*Block{Genesis()},
	}
}