// Package network implements networking functionalities for a blockchain application.
package network

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"syscall"

	"go_blockchain/blockchain"

	"github.com/vrecan/death/v3"
)

const (
	protocol      = "tcp" // Network protocol used for communication
	version       = 1
	commandLength = 12
)

var (
	nodeAddress     string                                    // Address of main node
	mineAddress    string                                    // Address of node used for mining
	KnownNodes      = []string{"localhost:3000"}              // List of known nodes in the network, already contains the main node
	blocksInTransit = [][]byte{}                              // List of block hashes that are in transit (being sent or received)
	memoryPool      = make(map[string]blockchain.Transaction) // Pool of transactions in memory
)

type Addr struct {
	Addrlist []string // List of node addresses
}

type Block struct {
	AddrFrom string // Address of the node sending the block
	Block    []byte // Serialized block data
}

type GetBlocks struct {
	AddrFrom string // Address of the node from which blocks are requested
}

type GetData struct {
	AddrFrom string // Address of the node from which data is requested
	Type     string // Type of data being requested (e.g., "block" or "tx")
	ID       []byte // ID of the requested data (block hash or transaction ID)
}

type Inv struct {
	AddrFrom string   // Address of the node sending the inventory
	Type     string   // Type of inventory being sent (e.g., "block" or "tx")
	Items    [][]byte // List of item IDs (block hashes or transaction IDs)
}

type Tx struct {
	AddrFrom    string // Address of the node sending the transaction
	Transaction []byte // Transaction data
}

type Version struct {
	Version    int    // Blockchain version to identify when blockchains are not the same
	BestHeight int    // Length of the blockchain of that version
	AddrFrom   string // Address of the node sending the version message
}

func CommandToBytes(command string) []byte {
	/*
		Converts a command string to a byte array of fixed length (commandLength).
	*/
	var bytes [commandLength]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func BytesToCommand(bytes []byte) string {
	/*
		Converts a byte array back to a command string by removing spaces.
	*/
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return string(command)
}

func GobEncode(data any) []byte {
	/*
		Returns the gob encoded byte array of the input data structure.
	*/
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func SendData(addr string, data []byte) {
	/*
		Sends data to a specified address over a TCP connection.
	*/
	conn, err := net.Dial(protocol, addr) // Establish a connection to the specified address
	// If connection fails, remove the node from KnownNodes
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range KnownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}
		KnownNodes = updatedNodes
		return
	}

	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data)) // Send the data over the connection by copying it to the connection
	if err != nil {
		log.Panic(err)
	}
}

func SendAddr(address string) {
	/*
		To send the list of known nodes to a specific node in the network.
	*/
	nodes := Addr{KnownNodes}
	nodes.Addrlist = append(nodes.Addrlist, nodeAddress)
	payload := GobEncode(nodes)
	req := append(CommandToBytes("addr"), payload...)

	log.Printf("--> Sending addresses to %s", address)
	SendData(address, req)
}

func SendBlock(addr string, b *blockchain.Block) {
	/*
		To send a block to a specific node in the network.
	*/
	data := Block{nodeAddress, b.Serialize()}
	payload := GobEncode(data)
	req := append(CommandToBytes("block"), payload...)

	log.Printf("--> Sending block %x to %s", b.Hash, addr)
	SendData(addr, req)
}

func SendInv(address, kind string, items [][]byte) {
	/*
		To send an inventory of blocks or transactions to a specific node in the network.
	*/
	inventory := Inv{nodeAddress, kind, items}
	payload := GobEncode(inventory)
	req := append(CommandToBytes("inv"), payload...)

	log.Printf("--> Sending inventory of %d %s to %s", len(items), kind, address)
	SendData(address, req)
}

func SendTx(addr string, tnx *blockchain.Transaction) {
	/*
		To send a transaction to a specific node in the network.
	*/
	data := Tx{nodeAddress, tnx.Serialize()}
	payload := GobEncode(data)
	req := append(CommandToBytes("tx"), payload...)

	log.Printf("--> Sending transaction %x to %s", tnx.ID, addr)
	SendData(addr, req)
}

func SendVersion(addr string, chain *blockchain.BlockChain) {
	/*
		To send the version of the blockchain to a specific node in the network.
	*/
	bestHeight := chain.GetBestHeight()
	payload := GobEncode(Version{version, bestHeight, nodeAddress})

	req := append(CommandToBytes("version"), payload...)

	log.Printf("--> Sending version to %s", addr)
	SendData(addr, req)
}

func SendGetBlocks(addr string) {
	/*
		To request the list of blocks from a specific node in the network.
	*/
	payload := GobEncode(GetBlocks{nodeAddress})
	req := append(CommandToBytes("getblocks"), payload...)

	log.Printf("--> Sending getblocks to %s", addr)
	SendData(addr, req)
}

func SendGetData(addr string, kind string, id []byte) {
	/*
		To request the list of blocks from a specific node in the network.
	*/
	payload := GobEncode(GetData{nodeAddress, kind, id})
	req := append(CommandToBytes("getdata"), payload...)

	log.Printf("--> Sending getdata for %s %x to %s", kind, id, addr)
	SendData(addr, req)
}

func HandleAddr(request []byte) {
	/*
		Handles the "addr" command by decoding the received data and updating the list of known nodes.
	*/
	var buff bytes.Buffer
	var payload Addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	KnownNodes = append(KnownNodes, payload.Addrlist...)
	log.Printf("Received addresses, now have %d known nodes", len(KnownNodes))
	RequestBlocks()
}

func RequestBlocks() {
	/*
		Requests the list of blocks from all known nodes in the network.
	*/
	for _, node := range KnownNodes {
		SendGetBlocks(node) // Send a "getblocks" request to each known node
	}
}

func HandleBlock(request []byte, chain *blockchain.BlockChain) {
	/*
	   Handles the "block" command by decoding the received block data and adding it to the blockchain.
	*/
	var buff bytes.Buffer
	var payload Block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := blockchain.Deserialize(blockData)

	log.Printf("Received a new block from %s!", payload.AddrFrom)

	err = chain.AddBlock(block)
	if err != nil {
		log.Printf("Could not add block: %v", err)
		return
	}

	log.Printf("Added block %x", block.Hash)

	// If there are more blocks in transit, request the next one
	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		SendGetData(payload.AddrFrom, "block", blockHash)
		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXO := blockchain.UTXOSet{Blockchain: chain}
		UTXO.Reindex()
	}
}

func HandleGetBlocks(request []byte, chain *blockchain.BlockChain) {
	/*
		Handles the "getblocks" command by sending the list of block hashes to the requesting node.
	*/
	var buff bytes.Buffer
	var payload GetBlocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := chain.GetBlockHashes()
	SendInv(payload.AddrFrom, "block", blocks)
}

func HandleGetData(request []byte, chain *blockchain.BlockChain) {
	/*
		Handles the "getdata" command by sending the requested block or transaction to the requesting node.
	*/
	var buff bytes.Buffer
	var payload GetData

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := chain.GetBlock([]byte(payload.ID)) // Get the requested block from the blockchain
		if err != nil {
			return
		}
		SendBlock(payload.AddrFrom, &block) // Send the block to the requesting node
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID) // Convert the transaction ID to a string
		tx := memoryPool[txID]                 // Get the requested transaction from the memory pool

		if tx.ID == nil {
			return
		}
		SendTx(payload.AddrFrom, &tx) // Send the transaction to the requesting node
	}
}

func HandleVersion(request []byte, chain *blockchain.BlockChain) {
	/*
		Handles the "version" command by comparing the blockchain versions and heights,
		and requesting blocks if the local blockchain is behind.
	*/
	var buff bytes.Buffer
	var payload Version

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	bestHeight := chain.GetBestHeight() // Get the best height of the local blockchain
	log.Printf(">>> VERSION received from %s (their height: %d, our height: %d)\n", payload.AddrFrom, payload.BestHeight, bestHeight)

	if bestHeight < payload.BestHeight { // If the local blockchain is behind
		log.Printf(">>> We are behind! Requesting blocks from %s\n", payload.AddrFrom)
		SendGetBlocks(payload.AddrFrom) // Request blocks from the sender
	} else if bestHeight > payload.BestHeight { // If the local blockchain is ahead
		log.Printf(">>> We are ahead! Sending blocks to %s\n", payload.AddrFrom)
		SendVersion(payload.AddrFrom, chain) // Send the local version to the sender
	} else {
		log.Printf(">>> We are up-to-date with %s\n", payload.AddrFrom)
	}

	if !NodeIsKnown(payload.AddrFrom) { // If the sender is not in the list of known nodes
		KnownNodes = append(KnownNodes, payload.AddrFrom) // Add the sender to the list of known nodes
		log.Printf(">>> Added %s to known nodes\n", payload.AddrFrom)
	}
}

func NodeIsKnown(addr string) bool {
	/*
		Checks if a node address is already in the list of known nodes.
	*/
	for _, node := range KnownNodes {
		if node == addr {
			return true
		}
	}
	return false
}

func HandleTx(request []byte, chain *blockchain.BlockChain) {
	/*
		Handles the "tx" command by processing the received transaction.
	*/
	var buff bytes.Buffer
	var payload Tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := blockchain.DeserializeTransaction(txData)
	memoryPool[hex.EncodeToString(tx.ID)] = tx

	log.Printf("Received transaction %x from %s", tx.ID, payload.AddrFrom)

	if nodeAddress == KnownNodes[0] {
		for _, node := range KnownNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				SendInv(node, "tx", [][]byte{tx.ID}) // Notify other nodes about the new transaction
			}
		}
	}

	if len(memoryPool) >= 1 && len(mineAddress) > 0 {
		MineTx(chain) // If the node is a miner and has enough transactions, mine a new block
	}
}

func MineTx(chain *blockchain.BlockChain) {
	/*
		Handles the mining of transactions into a new block.
	*/
    var txs []*blockchain.Transaction

    // Make a slice of transaction IDs to ensure consistent processing
    var txIDs []string
    for id := range memoryPool {
        txIDs = append(txIDs, id)
    }

    // Process transactions in a deterministic order
    for _, id := range txIDs {
        // Get the transaction from memory pool
        tx := memoryPool[id]
        fmt.Printf("Verifying tx: %x\n", tx.ID)
        
        if chain.VerifyTransaction(&tx, memoryPool) {
            // Store a persistent copy of the verified transaction
            verifiedTx := tx // Make a copy
            txs = append(txs, &verifiedTx)
            log.Printf("Transaction %x passed verification", tx.ID)
        } else {
            log.Printf("Transaction %x failed verification - INVALID", tx.ID)
        }
    }

    if len(txs) == 0 {
        log.Println("All transactions are invalid")
        return
    }

	// Mine a new block with the transactions.
	cbTx := blockchain.CoinbaseTx(mineAddress, "")
	txs = append(txs, cbTx)

	newBlock := chain.MineBlock(txs)
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()

	log.Println("New block mined")

	// Clear the memory pool.
	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		delete(memoryPool, txID) // Remove the transaction from the memory pool
	}

	// Notify other nodes about the new block
	for _, node := range KnownNodes {
		if node != nodeAddress {
			SendInv(node, "block", [][]byte{newBlock.Hash}) // Send an inventory of the new block hash to other nodes
		}
	}

	if len(memoryPool) > 0 {
		MineTx(chain) // If there are still transactions in the memory pool, mine another block
	}
}

func HandleInv(request []byte, chain *blockchain.BlockChain) {
	/*
		Handles the "inv" command by processing the received inventory of blocks or transactions.
	*/
	var buff bytes.Buffer
	var payload Inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Received inventory with %d %s from %s", len(payload.Items), payload.Type, payload.AddrFrom)

	if payload.Type == "block" {
		blocksInTransit = payload.Items
		blockHash := payload.Items[0] // Get the first block hash in the inventory
		log.Printf("Requesting block %x\n", blockHash)
		SendGetData(payload.AddrFrom, "block", blockHash) // Request the block from the sender

		blocksInTransit = blocksInTransit[1:] // Remove the requested block hash from the list
		log.Printf(">>> Blocks remaining in transit: %d\n", len(blocksInTransit))
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]                            // Get the first transaction ID in the inventory
		if memoryPool[hex.EncodeToString(txID)].ID == nil { // If the transaction is not already in the memory pool
			log.Printf("Requesting transaction %x\n", txID)
			SendGetData(payload.AddrFrom, "tx", txID) // Request the transaction from the sender
		}
	}
}

func HandleConnection(conn net.Conn, chain *blockchain.BlockChain) {
	/*
		Handles incoming connections by reading commands and dispatching them to appropriate handlers.
	*/
	req, err := io.ReadAll(conn)
	defer conn.Close()

	if err != nil {
		if err != io.EOF {
			log.Panic(err)
		}
		return
	}
	command := BytesToCommand(req[:commandLength])
	log.Printf("<-- Received %s command", command)

	switch command {
	case "addr":
		HandleAddr(req)
	case "block":
		HandleBlock(req, chain)
	case "inv":
		HandleInv(req, chain)
	case "getblocks":
		HandleGetBlocks(req, chain)
	case "getdata":
		HandleGetData(req, chain)
	case "tx":
		HandleTx(req, chain)
	case "version":
		HandleVersion(req, chain)
	default:
		log.Println("Unknown command")
	}
}

func CloseDB(chain *blockchain.BlockChain) {
	/*
		If a termination signal is received, this function ensures that the
		blockchain database is properly closed before exiting the program.
		Termination signals like SIGINT (Ctrl+C) and SIGTERM are handled.
	*/
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)       // Exit with a non-zero status to indicate termination
		defer runtime.Goexit() // Ensure all goroutines are terminated
		chain.Database.Close() // Close the blockchain database
	})
}

func StartServer(nodeID, minerAddress string) {
	/*
		Starts the blockchain server and listens for incoming connections.
	*/
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	mineAddress = minerAddress

	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	go CloseDB(chain)

	if nodeAddress != KnownNodes[0] {
		SendVersion(KnownNodes[0], chain)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go HandleConnection(conn, chain)

	}
}
