package main

import (
	"fmt"
	"go_blockchain/blockchain"
	"strconv"
)

func main() {
	fmt.Println("Hello, Go Blockchain!")
	newNetwork := blockchain.NewBlockChain()
	newNetwork.AddBlock("First Block after Genesis")
	newNetwork.AddBlock("Second Block after Genesis")

	for _, block := range newNetwork.Blocks {

		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}

