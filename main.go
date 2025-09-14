package main

import (
	"fmt"
	"go_blockchain/blockchain"
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
		fmt.Println()
	}
}