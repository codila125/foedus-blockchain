package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
)

const (
	DBPath      = "./temp/blocks_%s"
	genesisData = "Genesis Block - Go Blockchain Implementation"
)

type BlockChain struct {
	LastHash []byte     // Hash of the last block in the chain
	Database *badger.DB // Reference to the BadgerDB database
}

type BlockChainIterator struct {
	CurrentHash []byte     // Hash of the current block in the iteration
	Database    *badger.DB // Reference to the BadgerDB database
}

func DBExists(path string) bool {
	/*
		Checks if the blockchain database file exists.
		Returns true if the database file exists, false otherwise.
	*/
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}
	return true
}

func NewBlockChain(address string, nodeID string) *BlockChain {
	/*
		Creates a new blockchain with a genesis block and stores it in the database.
		'address' is the address to send the coinbase reward to.
		Returns a pointer to the newly created BlockChain instance.
	*/
	var lastHash []byte

	// Ensure a blockchain does not already exist
	path := fmt.Sprintf(DBPath, nodeID)
	if DBExists(path) {
		log.Printf("[BLOCKCHAIN] Blockchain already exists for node %s", nodeID)
		runtime.Goexit()
	}

	log.Printf("[BLOCKCHAIN] Initializing new blockchain for node %s", nodeID)

	if err := os.MkdirAll(path, 0o755); err != nil {
		log.Panic(err)
	}

	options := badger.DefaultOptions(path) // Set default options for BadgerDB

	db, err := openDB(path, options)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData) // Create the coinbase transaction for the genesis block
		genesis := Genesis(cbtx)                 // Create the genesis block
		log.Printf("[BLOCKCHAIN] Genesis block created - Hash: %x", genesis.Hash)
		err := txn.Set(genesis.Hash, genesis.Serialize()) // Store the genesis block in the database
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash) // Store the last hash pointer because it helps to find the last block

		lastHash = genesis.Hash // Set the last hash to the genesis block's hash

		return err
	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	log.Printf("[BLOCKCHAIN] Blockchain initialized successfully for node %s", nodeID)
	return &blockchain
}

func ContinueBlockChain(nodeID string) *BlockChain {
	/*
		Continues an existing blockchain by loading it from the database.
		Returns a pointer to the BlockChain instance.
	*/
	path := fmt.Sprintf(DBPath, nodeID)
	if !DBExists(path) {
		log.Printf("[BLOCKCHAIN] No existing blockchain found for node %s", nodeID)
		runtime.Goexit()
	}

	log.Printf("[BLOCKCHAIN] Loading existing blockchain for node %s", nodeID)

	var lastHash []byte

	options := badger.DefaultOptions(path) // Set default options for BadgerDB
	db, err := openDB(path, options)
	Handle(err)

	// Read the last hash from the database
	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		return err
	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	log.Printf("[BLOCKCHAIN] Blockchain loaded successfully with height %d", blockchain.GetBestHeight())
	return &blockchain
}

func (blockchain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	/*
		Adds a new block with the given transactions to the blockchain.
		'transactions' is a slice of pointers to Transaction instances to be included in the new block.
		Returns a pointer to the newly added Block instance.
	*/
	var lastHash []byte
	var lastHeight int

	log.Printf("[MINING] Starting block mining with %d transaction(s)", len(transactions))

	txMap := make(map[string]Transaction)
	for _, tx := range transactions {
		txMap[hex.EncodeToString(tx.ID)] = *tx
	}

	for _, tx := range transactions {
		if !blockchain.VerifyTransaction(tx, txMap) {
			log.Panicf("[MINING] Invalid transaction detected: %x", tx.ID)
		}
	}

	// Get the last hash from the database
	err := blockchain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)
		Handle(err)
		item, err = txn.Get(lastHash)
		Handle(err)
		blockData, err := item.ValueCopy(nil)
		Handle(err)
		block := Deserialize(blockData)
		lastHeight = block.Height

		return err
	})

	Handle(err)

	newBlock := CreateBlock(transactions, lastHash, lastHeight+1) // Create a new block with the transactions and previous hash

	// Store the new block in the database and update the last hash pointer
	err = blockchain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		blockchain.LastHash = newBlock.Hash
		return err
	})

	Handle(err)

	log.Printf("[MINING] Block mined successfully - Hash: %x, Height: %d", newBlock.Hash, newBlock.Height)

	utxoSet := UTXOSet{blockchain}
	utxoSet.Update(newBlock)

	return newBlock
}

func (blockchain *BlockChain) AddBlock(block *Block) error {
	/*
		Adds a block to the blockchain if it does not already exist.
		'block' is a pointer to the Block instance to be added.
		Returns an error if any operation fails, otherwise returns nil.
	*/

	err := blockchain.Database.Update(func(txn *badger.Txn) error {
		// Check if block already exists
		if _, err := txn.Get(block.Hash); err == nil {
			return nil // Block already exists
		}

		// Add the block to the database
		err := txn.Set(block.Hash, block.Serialize())
		if err != nil {
			return err
		}
		// Get the current last hash
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return fmt.Errorf("could not get last hash: %w", err)
		}
		lastHash, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("could not copy last hash: %w", err)
		}

		item, err = txn.Get(lastHash)
		if err != nil {
			return fmt.Errorf("could not get last block: %w", err)
		}
		lastBlock, err := item.ValueCopy(nil)
		Handle(err)
		lastBlockData := Deserialize(lastBlock)

		// Update the last hash only if the new block's height is greater
		if block.Height > lastBlockData.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			if err != nil {
				return fmt.Errorf("could not update last hash: %w", err)
			}
			blockchain.LastHash = block.Hash
			log.Printf("[BLOCKCHAIN] Block added and chain updated - Hash: %x, Height: %d", block.Hash, block.Height)
		} else {
			log.Printf("[BLOCKCHAIN] Block added to database - Hash: %x, Height: %d", block.Hash, block.Height)
		}

		return nil
	})

	return err
}

func (blockchain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	/*
		Retrieves a block from the blockchain by its hash.
		'blockHash' is the hash of the block to be retrieved.
		Returns the Block instance if found, otherwise returns an error.
	*/
	var block Block

	err := blockchain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(blockHash)
		Handle(err)
		blockData, err := item.ValueCopy(nil)
		Handle(err)
		block = *Deserialize(blockData)

		return err
	})

	Handle(err)
	return block, nil
}

func (blockchain *BlockChain) GetBlockHashes() [][]byte {
	/*
		Retrieves all block hashes in the blockchain.
		Returns a slice of byte slices, each representing a block hash.
	*/
	var blocks [][]byte
	iterator := blockchain.Iterator() // Create an iterator to traverse the blockchain

	for {
		block := iterator.Next()
		blocks = append(blocks, block.Hash)
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return blocks
}

func (blockchain *BlockChain) GetBestHeight() int {
	/*
		Returns the height of the latest block in the blockchain.
	*/
	var lastBlock Block

	err := blockchain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err := item.ValueCopy(nil)
		Handle(err)

		item, err = txn.Get(lastHash)
		Handle(err)
		blockData, err := item.ValueCopy(nil)
		Handle(err)

		lastBlock = *Deserialize(blockData)

		return err
	})

	Handle(err)
	return lastBlock.Height
}

func (blockchain *BlockChain) FindUTXO() map[string]TxOutputs {
	/*
		Scans the entire blockchain to find all unspent transaction outputs (UTXOs).
		Returns a map where the key is the transaction ID and the value is the corresponding unspent outputs.
	*/
	UTXO := make(map[string]TxOutputs) // UTXO map to hold unspent transaction outputs
	spentTXs := make(map[string][]int) // Map to track spent transaction outputs

	iterator := blockchain.Iterator() // Create an iterator to traverse the blockchain

	for {
		block := iterator.Next() // Get the next block

		for _, tx := range block.Transactions { // Iterate over each transaction in the block
			txID := hex.EncodeToString(tx.ID) // Encode transaction ID to string

		Outputs:
			for outIdx, out := range tx.Outputs { // Iterate over each output in the transaction
				// If the output is already spent, skip it
				if spentTXs[txID] != nil {
					for _, spentOut := range spentTXs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]                       // Initialize outputs for this transaction ID
				outs.Outputs = append(outs.Outputs, out) // Add the unspent output
				UTXO[txID] = outs                        // Update the UTXO map with the new output
			}
			// If the transaction is not a coinbase, mark its inputs as spent
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXs[inTxID] = append(spentTXs[inTxID], in.Out) // Mark the output as spent
				}
			}
		}
		// If we've reached the genesis block, stop iterating
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO
}

func (blockchain *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	/*
		Finds and returns a transaction by its ID by scanning through all blocks in the blockchain.
		Returns the transaction if found, otherwise returns an error.
	*/
	iterator := blockchain.Iterator()
	// Iterate through the blocks in the blockchain
	for {
		block := iterator.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (blockchain *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	/*
		Signs a transaction using the provided private key.
		'tx' is the transaction to be signed.
		'privKey' is the ECDSA private key used for signing.
	*/
	if tx.IsCoinbase() {
		return
	}
	prevTXs := make(map[string]Transaction)
	for _, in := range tx.Inputs {
		prevTX, err := blockchain.FindTransaction(in.ID) // Find the previous transaction referenced by the input
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs) // Sign the transaction with the private key and previous transactions
}

func (blockchain *BlockChain) VerifyTransaction(tx *Transaction, txMap map[string]Transaction) bool {
	/*
	   Verifies the signatures of a transaction.
	   'tx' is the transaction to be verified.
	   'txMap' is a map of other transactions in the same block/pool.
	   Returns true if the transaction is valid, false otherwise.
	*/
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		// First, check if the previous transaction is in the current pool of transactions.
		if prevTx, ok := txMap[hex.EncodeToString(in.ID)]; ok {
			prevTXs[hex.EncodeToString(prevTx.ID)] = prevTx
		} else {
			// If not in the pool, search the blockchain.
			prevTX, err := blockchain.FindTransaction(in.ID)
			if err != nil {
				log.Printf("[VERIFY] Transaction verification failed - Parent transaction %x not found", in.ID)
				return false
			}
			prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
		}
	}

	if !tx.Verify(prevTXs) {
		log.Printf("[VERIFY] Transaction %x signature verification failed", tx.ID)
		return false
	}

	return true
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	/*
		Attempts to open the Badger database with the original options.
		Lock files exist if the previous instance did not close properly.
		If it fails, it removes the lock file and retries opening the database.
	*/
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil { // Remove the lock file
		return nil, err
	}
	log.Printf("[DATABASE] Retrying database connection after removing lock file")
	retryOpts := originalOpts  // Retry opening the database with the original options
	retryOpts.ReadOnly = false // Ensure ReadOnly is false for retry
	return badger.Open(retryOpts)
}

func openDB(dir string, options badger.Options) (*badger.DB, error) {
	/*
		Attempts to open the Badger database with the specified options.
		If it fails, it retries opening the database with modified options.
	*/
	if db, err := badger.Open(options); err != nil {
		if strings.Contains(err.Error(), "LOCK") { // Check if the error is related to a lock file
			log.Printf("[DATABASE] Database locked, attempting recovery")
			if db, err := retry(dir, options); err == nil {
				log.Printf("[DATABASE] Database opened successfully after retry")
				return db, nil
			}
			log.Printf("[DATABASE] Failed to open database after retry")
		}
		return nil, err
	} else {
		return db, nil
	}
}
