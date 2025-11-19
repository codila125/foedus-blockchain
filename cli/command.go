package cli

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/network"
	"github.com/codila125/foedus-blockchain/wallet"
)

// CommandLine serves as the entry point for all command-line operations.
// It provides a structured way to handle user input and execute corresponding
// blockchain functionalities.
type CommandLine struct{}

// printUsage displays a comprehensive list of available commands and their
// usage instructions. This function is automatically called when the user
// provides invalid or insufficient arguments.
func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address <ADDRESS> : get the balance for an address")
	fmt.Println(" createblockchain -address <ADDRESS> : creates a blockchain and sends genesis reward to address")
	fmt.Println(" printchain : prints the blocks in the chain")
	fmt.Println(" send -from <FROM> -to <TO> -amount <AMOUNT> : send amount of coins from address to address")
	fmt.Println(" createwallet : creates a new Wallet")
	fmt.Println(" listaddresses : lists the addresses in our wallet file")
	fmt.Println(" reindex : rebuilds the UTXO and ICCT sets")
	fmt.Println(" startnode -source <ADDRESS> : starts a node with ID specified in NODE_ID env. var. -miner enables mining")
}

// validateArgs checks if the user has provided the minimum required arguments
// to run a command. If not, it prints the usage information and exits.
func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

// listAddresses retrieves and displays all wallet addresses stored on the
// current node. It provides a simple way for users to view their available
// addresses.
func (cli *CommandLine) listAddresses(nodeID string) {
	wallets, _ := wallet.CreateWallets(nodeID)
	log.Printf("[CLI] Listing all wallet addresses:")

	addresses := wallets.GetAllAddresses()
	for _, address := range addresses {
		log.Printf("[CLI] - %s\n", address)
	}
}

// createWallet generates a new cryptographic key pair (wallet), saves it to
// the node's storage, and prints the new wallet address. This allows users
// to create new identities for transacting on the blockchain.
func (cli *CommandLine) createWallet(nodeID string) {
	wallets, _ := wallet.CreateWallets(nodeID)
	address := wallets.AddWallet()
	err := wallets.SaveFile(nodeID)
	if err != nil {
		log.Printf("[WALLET] Failed to save wallet: %v\n", err)
		return
	}

	log.Printf("[WALLET] New wallet created successfully with address: %s\n", address)
}

// printChain iterates through the entire blockchain and prints a detailed
// view of each block, including its header, transactions, and contracts.
// This is useful for debugging and verifying the chain's integrity.
func (cli *CommandLine) printChain(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID)
	defer func() {
		_ = chain.Database.Close()
	}()
	iter := chain.Iterator()

	blockNumber := 0
	for {
		block := iter.Next()

		cli.printBlockHeader(block, blockNumber)
		cli.printBlockContracts(block)
		cli.printBlockTransactions(block)
		cli.printBlockFooter()

		blockNumber++
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

// createBlockChain initializes a new blockchain with a genesis block, which
// is the first block in the chain. It also creates the initial UTXO and ICCT
// sets, which are essential for processing future transactions and contracts.
func (cli *CommandLine) createBlockChain(address string, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Print("[CLI] Invalid address provided")
	}

	log.Printf("[CLI] Creating new blockchain for address: %s", address)
	chain := blockchain.NewBlockChain(address, nodeID)

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()

	ICCTSet := blockchain.ICCTSet{Blockchain: chain}
	ICCTSet.Reindex()

	chain.Database.Close()

	log.Printf("[CLI] ✓ Blockchain created successfully")
}

// getBalance calculates and displays the total balance of a given wallet
// address. It does this by summing the values of all unspent transaction
// outputs (UTXOs) associated with the address's public key hash.
func (cli *CommandLine) getBalance(address string, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Print("[CLI] Invalid address provided")
	}

	log.Printf("[CLI] Fetching balance for address: %s", address)

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	defer chain.Database.Close()

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	log.Printf("[CLI] Balance of %s retrieved: %d", address, balance)
}

// send facilitates the transfer of a specified amount of currency from a
// sender's address to a recipient's address. It creates a new transaction,
// mines a new block to include it, and updates the UTXO set accordingly.
func (cli *CommandLine) send(from, to string, amount int, nodeID string) {
	if !wallet.ValidateAddress(from) {
		log.Print("[CLI] Invalid sender address")
	}
	if !wallet.ValidateAddress(to) {
		log.Print("[CLI] Invalid recipient address")
	}

	log.Printf("[CLI] Initiating transaction: %d from %s to %s", amount, from, to)

	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		log.Print(err)
	}
	wallet, err := wallets.GetWallet(from)
	if err != nil {
		log.Print(err)
	}

	tx := blockchain.NewTransaction(&wallet, to, amount, &UTXOSet)
	log.Printf("[CLI] Mining transaction locally")
	cbTx := blockchain.CoinbaseTx(from, "")
	txs := []*blockchain.Transaction{cbTx, tx}
	newBlock := chain.MineBlock(txs, nil)
	log.Printf("[CLI] ✓ Transaction mined in block %x", newBlock.Hash)
}

// reindex rebuilds the Unspent Transaction Output (UTXO) and Incomplete
// Contract (ICCT) sets from the blockchain data. This operation is crucial
// for ensuring data consistency and can resolve discrepancies that may arise
// during network operations.
func (cli *CommandLine) reindex(nodeID string) {
	log.Printf("[CLI] Starting UTXO and ICCT reindex operation")
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()
	count := UTXOSet.CountTransactions()
	log.Printf("[CLI] ✓ UTXO reindex complete - %d transaction(s) in set", count)

	ICCTSet := blockchain.ICCTSet{Blockchain: chain}
	ICCTSet.Reindex()
	contractCount := ICCTSet.CountContracts()
	log.Printf("[CLI] ✓ ICCT reindex complete - %d incomplete contract(s) in set", contractCount)
}

// startNode launches a new node and connects it to the blockchain network.
// It can optionally start in mining mode, which allows the node to create
// new blocks and earn rewards, sent to the specified miner address.
func (cli *CommandLine) startNode(nodeID string, sourceAddress string) {
	network.RunMinerNode(nodeID, sourceAddress)
}
