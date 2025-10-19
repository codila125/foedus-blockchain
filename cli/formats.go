package cli

import (
	"fmt"

	"github.com/codila125/foedus-blockchain/blockchain"
)

func (cli *CommandLine) printBlockHeader(block *blockchain.Block, blockNumber int) {
	/*
		Prints the block header with hash, timestamp, nonce, height, and PoW validation.
	*/
	fmt.Println("\n" + "╔" + fmt.Sprintf("═══ BLOCK [#%d] ═══════════════════════════════════════════╗", blockNumber))
	fmt.Printf("║ Hash:     %x\n", block.Hash)
	fmt.Printf("║ Prev Hash: %x\n", block.PrevHash)
	fmt.Printf("║ Timestamp: %d\n", block.Timestamp)
	fmt.Printf("║ Nonce:     %d\n", block.Nonce)
	fmt.Printf("║ Height:    %d\n", block.Height)

	// Validate PoW
	pow := blockchain.NewProof(block)
	isValid := pow.Validate()
	powStatus := "✓ VALID"
	if !isValid {
		powStatus = "✗ INVALID"
	}
	fmt.Printf("║ PoW:       %s\n", powStatus)
	fmt.Println("╟" + "─────────────────────────────────────────────────────────────")
}

func (cli *CommandLine) printBlockContracts(block *blockchain.Block) {
	/*
		Prints all contracts in the block with proper formatting and indentation.
	*/
	fmt.Printf("║ CONTRACTS (%d):\n", len(block.Contracts))
	if len(block.Contracts) == 0 {
		fmt.Println("║   (No Contracts)")
	} else {
		for i, contract := range block.Contracts {
			cli.printContractWithBox(contract, i+1)
		}
	}
}

func (cli *CommandLine) printContractWithBox(contract *blockchain.Contract, contractNumber int) {
	/*
		Prints a single contract with a nicely formatted box.
	*/
	fmt.Printf("║   ┌─ Contract #%d ────────────────────────────────────────┐\n", contractNumber)
	fmt.Printf("║   │ ╭─── CONTRACT [%.32x] ───╮\n", contract.ID)
	fmt.Printf("║   │ │ Title: %s\n", contract.Title)
	fmt.Printf("║   │ │ Status: %s\n", contract.Status)
	fmt.Printf("║   │ │ Creator: %s\n", contract.CreatorAddress)

	fmt.Printf("║   │ ├─ PARTIES (%d)\n", len(contract.Parties))
	if len(contract.Parties) == 0 {
		fmt.Println("║   │ │   (No Parties)")
	} else {
		for _, party := range contract.Parties {
			signed := "✓"
			if len(party.Signature) == 0 {
				signed = "✗"
			}
			fmt.Printf("║   │ │   %s [%s] %s\n", signed, party.Role, party.Address)
		}
	}

	fmt.Printf("║   │ ├─ MILESTONES (%d)\n", len(contract.Milestones))
	if len(contract.Milestones) == 0 {
		fmt.Println("║   │ │   (No Milestones)")
	} else {
		for j, milestone := range contract.Milestones {
			fmt.Printf("║   │ │   [%d] %s (%s)\n", j+1, milestone.Title, milestone.Status)
			fmt.Printf("║   │ │       Value: %d\n", milestone.Value)
		}
	}

	fmt.Println("║   │ ╰──────────────────────────────────────────────────╯")
	fmt.Println("║   └───────────────────────────────────────────────────────────┘")
}

func (cli *CommandLine) printBlockTransactions(block *blockchain.Block) {
	/*
		Prints all transactions in the block with proper formatting and indentation.
	*/
	fmt.Println("╟" + "─────────────────────────────────────────────────────────────")
	fmt.Printf("║ TRANSACTIONS (%d):\n", len(block.Transactions))
	if len(block.Transactions) == 0 {
		fmt.Println("║   (No Transactions)")
	} else {
		for i, tx := range block.Transactions {
			cli.printTransactionWithBox(tx, i+1)
		}
	}
}

func (cli *CommandLine) printTransactionWithBox(tx *blockchain.Transaction, txNumber int) {
	/*
		Prints a single transaction with a nicely formatted box.
	*/
	fmt.Printf("║   ┌─ Transaction #%d ──────────────────────────────────────┐\n", txNumber)
	fmt.Printf("║   │ ╭─── TRANSACTION [%.32x] ───╮\n", tx.ID)

	// Inputs Section
	fmt.Printf("║   │ ├─ INPUTS (%d)\n", len(tx.Inputs))
	if len(tx.Inputs) == 0 {
		fmt.Println("║   │ │   (No Inputs)")
	} else {
		for i, input := range tx.Inputs {
			if tx.IsCoinbaseTx() {
				fmt.Printf("║   │ │   [COINBASE] → Reward Data: %s\n", input.PubKey)
			} else {
				fmt.Printf("║   │ │   [%d] From TX: %.16x...\n", i, input.ID)
				fmt.Printf("║   │ │       Output Index: %d\n", input.Out)
			}
		}
	}

	// Outputs Section
	fmt.Printf("║   │ ├─ OUTPUTS (%d)\n", len(tx.Outputs))
	if len(tx.Outputs) == 0 {
		fmt.Println("║   │ │   (No Outputs)")
	} else {
		for i, output := range tx.Outputs {
			fmt.Printf("║   │ │   [%d] To PubKeyHash: %.16x...\n", i, output.PubKeyHash)
			fmt.Printf("║   │ │       Value: %d\n", output.Value)
		}
	}

	fmt.Println("║   │ ╰──────────────────────────────────────────────────╯")
	fmt.Println("║   └───────────────────────────────────────────────────────────┘")
}

func (cli *CommandLine) printBlockFooter() {
	/*
		Prints the block footer to close the visual box.
	*/
	fmt.Println("╚" + "═══════════════════════════════════════════════════════════════")
}

