// Package cli implements the command line interface for the blockchain application.
package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/network"
	"github.com/codila125/foedus-blockchain/wallet"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	/*
		Prints the usage instructions for the command line interface.
	*/
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address <ADDRESS> : get the balance for an address")
	fmt.Println(" createblockchain -address <ADDRESS> : creates a blockchain and sends genesis reward to address")
	fmt.Println(" printchain : prints the blocks in the chain")
	fmt.Println(" send -from <FROM> -to <TO> -amount <AMOUNT> -mine : send amount of coins from address to address")
	fmt.Println(" createwallet : creates a new Wallet")
	fmt.Println(" listaddresses : lists the addresses in our wallet file")
	fmt.Println(" reindex : rebuilds the UTXO and ICCT sets")
	fmt.Println(" startnode -miner <ADDRESS> : starts a node with ID specified in NODE_ID env. var. -miner enables mining")
}

func (cli *CommandLine) validateArgs() {
	/*
			Validates that at least one command line argument is provided.
		   If no arguments are provided, it prints the usage instructions and exits the program.
	*/
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) listAddresses(nodeID string) {
	/*
		Lists all wallet addresses stored in the wallet file.
	*/
	wallets, _ := wallet.CreateWallets(nodeID)
	log.Printf("[CLI] Listing all wallet addresses:")

	addresses := wallets.GetAllAddresses()
	for _, address := range addresses {
		log.Printf("[CLI] - %s\n", address)
	}
}

func (cli *CommandLine) createWallet(nodeID string) {
	/*
		Creates a new wallet, saves it to the wallet file, and prints the new address.
	*/
	wallets, _ := wallet.CreateWallets(nodeID)
	address := wallets.AddWallet()
	err := wallets.SaveFile(nodeID)

	if err != nil {
		log.Printf("[WALLET] Failed to save wallet: %v\n", err)
		return
	}

	log.Printf("[WALLET] New wallet created successfully with address: %s\n", address)
}

func (cli *CommandLine) printChain(nodeID string) {
	/*
		Prints all the blocks in the blockchain along with their details in a formatted block structure.
	*/
	chain := blockchain.ContinueBlockChain(nodeID) // Load the existing blockchain
	defer chain.Database.Close()
	iter := chain.Iterator()

	blockNumber := 0
	// Iterate through the blocks in the blockchain and print their details
	for {
		block := iter.Next() // Get the next block

		cli.printBlockHeader(block, blockNumber)
		cli.printBlockContracts(block)
		cli.printBlockTransactions(block)
		cli.printBlockFooter()

		blockNumber++
		// Break the loop if we reach the genesis block (no previous hash)
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockChain(address string, nodeID string) {
	/*
		Creates a new blockchain and sends the genesis block reward to the specified address.
		Also initializes the UTXO set for the new blockchain.
	*/
	if !wallet.ValidateAddress(address) { // Validate the provided address
		log.Panic("[CLI] Invalid address provided")
	}

	log.Printf("[CLI] Creating new blockchain for address: %s", address)
	chain := blockchain.NewBlockChain(address, nodeID) // Create a new blockchain with the genesis block

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex() // Rebuild the UTXO set from the blockchain

	ICCTSet := blockchain.ICCTSet{Blockchain: chain}
	ICCTSet.Reindex() // Rebuild the ICCT set from the blockchain

	chain.Database.Close() // Close the database connection

	log.Printf("[CLI] ✓ Blockchain created successfully")
}

func (cli *CommandLine) getBalance(address string, nodeID string) {
	/*
		Calculates and prints the balance of the specified address by summing its unspent transaction outputs (UTXOs).
	*/
	if !wallet.ValidateAddress(address) {
		log.Panic("[CLI] Invalid address provided")
	}

	log.Printf("[CLI] Fetching balance for address: %s", address)

	// Load the existing blockchain and rebuild the UTXO set
	chain := blockchain.ContinueBlockChain(nodeID) // Load the existing blockchain
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	defer chain.Database.Close()

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))   // Decode the address to get the public key hash
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]       // Remove the version byte and checksum
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash) // Find all unspent transactions for the public key hash

	// Sum the values of all unspent transaction outputs to get the balance
	for _, out := range UTXOs {
		balance += out.Value
	}

	log.Printf("[CLI] Balance of %s retrieved: %d", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int, nodeID string, mineNow bool) {
	/*
		Creates and sends a new transaction from one address to another, including a coinbase transaction for the sender.
		Updates the UTXO set after adding the new block to the blockchain.
	*/
	// Validate the provided addresses
	if !wallet.ValidateAddress(from) {
		log.Panic("[CLI] Invalid sender address")
	}
	if !wallet.ValidateAddress(to) {
		log.Panic("[CLI] Invalid recipient address")
	}

	log.Printf("[CLI] Initiating transaction: %d from %s to %s", amount, from, to)

	// Load the existing blockchain and UTXO set
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	wallets, err := wallet.CreateWallets(nodeID) // Load existing wallets
	if err != nil {
		log.Panic(err)
	}
	wallet, err := wallets.GetWallet(from) // Get the wallet for the sender's address
	if err != nil {
		log.Panic(err)
	}

	// Create a new transaction from the sender to the recipient
	tx := blockchain.NewTransaction(&wallet, to, amount, &UTXOSet)
	if mineNow {
		log.Printf("[CLI] Mining transaction locally")
		cbTx := blockchain.CoinbaseTx(from, "")    // Create a coinbase transaction for the sender
		txs := []*blockchain.Transaction{cbTx, tx} // Include the coinbase transaction in the new block
		newBlock := chain.MineBlock(txs, nil)      // Mine a new block with the transactions                   // Update the UTXO set with the new block
		log.Printf("[CLI] ✓ Transaction mined in block %x", newBlock.Hash)
	} else {
		log.Printf("[CLI] Sending transaction to network")
		network.SendTx(network.KnownNodes[0], tx) // Send the transaction to a known node in the network
		log.Printf("[CLI] ✓ Transaction sent to network")
	}
}

func (cli *CommandLine) reindex(nodeID string) {
	/*
		Rebuilds the UTXO set and ICCT set from the current state of the blockchain.
	*/
	log.Printf("[CLI] Starting UTXO and ICCT reindex operation")
	chain := blockchain.ContinueBlockChain(nodeID) // Load the existing blockchain
	defer chain.Database.Close()

	UTXOSet := blockchain.UTXOSet{Blockchain: chain} // Create a UTXO set instance
	UTXOSet.Reindex()                                // Rebuild the UTXO set
	count := UTXOSet.CountTransactions()             // Count the number of transactions in the UTXO set
	log.Printf("[CLI] ✓ UTXO reindex complete - %d transaction(s) in set", count)

	ICCTSet := blockchain.ICCTSet{Blockchain: chain} // Create an ICCT set instance
	ICCTSet.Reindex()                                // Rebuild the ICCT set
	contractCount := ICCTSet.CountContracts()        // Count the number of contracts in the ICCT set
	log.Printf("[CLI] ✓ ICCT reindex complete - %d incomplete contract(s) in set", contractCount)
}

func (cli *CommandLine) startNode(nodeID string, minerAddress string) {
	/*
		Starts a new node in the blockchain network.
		If a miner address is provided, the node will also mine new blocks and send rewards to that address.
	*/
	if len(minerAddress) > 0 {
		if !wallet.ValidateAddress(minerAddress) {
			log.Panic("[CLI] Invalid miner address provided")
		}
	}
	network.StartServer(nodeID, minerAddress)
}

func (cli *CommandLine) Run() {
	/*
		Parses and executes the command line arguments.
	*/
	cli.validateArgs()
	nodeID := os.Getenv("NODE_ID") // Get the node ID from the environment variable
	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!")
		runtime.Goexit()
	}

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexCmd := flag.NewFlagSet("reindex", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")

	switch os.Args[1] {
	case "reindex":
		err := reindexCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress, nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}
	if reindexCmd.Parsed() {
		cli.reindex(nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}

	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			runtime.Goexit()
		}
		cli.startNode(nodeID, *startNodeMiner)
	}
}
