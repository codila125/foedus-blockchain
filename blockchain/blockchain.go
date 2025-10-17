package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/cockroachdb/pebble"
	"github.com/codila125/foedus-blockchain/database"
)

const (
	DBPath      = "./temp/blocks_%s"
	genesisData = "Genesis Block - Foedus"
)

type BlockChain struct {
	LastHash []byte             // Hash of the last block in the chain
	Database *database.PebbleDB // Reference to the PebbleDB database
}

type BlockChainIterator struct {
	CurrentHash []byte             // Hash of the current block in the iteration
	Database    *database.PebbleDB // Reference to the PebbleDB database
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
	if database.DBExists(path) {
		log.Printf("[BLOCKCHAIN] Blockchain already exists for node %s", nodeID)
		runtime.Goexit()
	}

	log.Printf("[BLOCKCHAIN] Initializing new blockchain for node %s", nodeID)

	if err := os.MkdirAll(path, 0o755); err != nil {
		log.Panic(err)
	}

	db, err := database.OpenDB(path)
	Handle(err)

	rawDB := db.GetRawDB()
	batch := rawDB.NewBatch()
	defer batch.Close()

	// Create the genesis block and store it in the database
	cbtx := CoinbaseTx(address, genesisData) // Create the coinbase transaction for the genesis block
	genesis := Genesis(cbtx)                 // Create the genesis block
	log.Printf("[BLOCKCHAIN] Genesis block created - Hash: %x", genesis.Hash)

	err = batch.Set(genesis.Hash, genesis.Serialize(), nil) // Store the genesis block in the database
	Handle(err)
	err = batch.Set([]byte("lh"), genesis.Hash, nil) // Store the last hash pointer because it helps to find the last block
	Handle(err)
	lastHash = genesis.Hash // Set the last hash to the genesis block's hash

	err = rawDB.Apply(batch, &pebble.WriteOptions{Sync: true})
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
	if !database.DBExists(path) {
		log.Printf("[BLOCKCHAIN] No existing blockchain found for node %s", nodeID)
		runtime.Goexit()
	}

	log.Printf("[BLOCKCHAIN] Loading existing blockchain for node %s", nodeID)

	var lastHash []byte

	db, err := database.OpenDB(path)
	Handle(err)

	// Read the last hash from the database
	lastHashBytes, err := db.Get([]byte("lh"))
	if err == nil {
		lastHash = make([]byte, len(lastHashBytes))
		copy(lastHash, lastHashBytes)
	} else {
		log.Panicf("[BLOCKCHAIN] Failed to retrieve last hash for node %s: %v", nodeID, err)
	}

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

	db := blockchain.Database.GetRawDB()

	// Get the last hash from the database
	lastHashBytes, closer, err := db.Get([]byte("lh"))
	if err == nil {
		lastHash = make([]byte, len(lastHashBytes))
		copy(lastHash, lastHashBytes)
		closer.Close()
	} else {
		log.Panicf("[BLOCKCHAIN] Failed to retrieve last hash: %v", err)
	}

	lastBlockData, closer, err := db.Get(lastHash)
	if err != nil {
		log.Panicf("[BLOCKCHAIN] Failed to retrieve last block: %v", err)
	}
	blockDataCopy := make([]byte, len(lastBlockData))
	copy(blockDataCopy, lastBlockData)
	closer.Close()

	lastBlock := Deserialize(blockDataCopy)
	lastHeight = lastBlock.Height
	newBlock := CreateBlock(transactions, lastHash, lastHeight+1) // Create a new block with the transactions and previous hash

	// Store the new block in the database and update the last hash pointer
	batch := db.NewBatch()
	defer batch.Close()

	batch.Set(newBlock.Hash, newBlock.Serialize(), nil)
	batch.Set([]byte("lh"), newBlock.Hash, nil)

	blockchain.LastHash = newBlock.Hash
	log.Printf("[MINING] Block mined successfully - Hash: %x, Height: %d", newBlock.Hash, newBlock.Height)

	err = db.Apply(batch, &pebble.WriteOptions{Sync: true})
	Handle(err)

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

	db := blockchain.Database.GetRawDB()
	batch := db.NewBatch()
	defer batch.Close()

	_, closer, err := db.Get(block.Hash)
	if err == nil {
		closer.Close()
		return nil // Block already exists
	}

	//Store block in database
	err = batch.Set(block.Hash, block.Serialize(), nil)
	if err != nil {
		return err
	}

	lastHash, closer, err := db.Get([]byte("lh"))
	if err != nil {
		return fmt.Errorf("could not get last hash: %w", err)
	}
	lastHashCopy := make([]byte, len(lastHash))
	copy(lastHashCopy, lastHash)
	closer.Close()

	lastBlockData, closer, err := db.Get(lastHashCopy)
	if err != nil {
		return fmt.Errorf("could not get last block: %w", err)
	}
	blockDataCopy := make([]byte, len(lastBlockData))
	copy(blockDataCopy, lastBlockData)
	closer.Close()

	lastBlock := Deserialize(blockDataCopy)

	if lastBlock == nil {
		return fmt.Errorf("could not deserialize last block")
	}

	// Update the last hash only if the new block's height is greater
	if block.Height > lastBlock.Height {
		err = batch.Set([]byte("lh"), block.Hash, nil)
		if err != nil {
			return fmt.Errorf("could not update last hash: %w", err)
		}
		blockchain.LastHash = block.Hash
		log.Printf("[BLOCKCHAIN] Block added and chain updated - Hash: %x, Height: %d", block.Hash, block.Height)
	} else {
		log.Printf("[BLOCKCHAIN] Block added to database - Hash: %x, Height: %d", block.Hash, block.Height)
	}

	err = db.Apply(batch, &pebble.WriteOptions{Sync: true})
	Handle(err)

	return nil
}

func (blockchain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	/*
		Retrieves a block from the blockchain by its hash.
		'blockHash' is the hash of the block to be retrieved.
		Returns the Block instance if found, otherwise returns an error.
	*/
	var block Block

	db := blockchain.Database.GetRawDB()
	blockData, closer, err := db.Get(blockHash)
	if err != nil {
		return block, fmt.Errorf("could not get block: %w", err)
	}
	blockDataCopy := make([]byte, len(blockData))
	copy(blockDataCopy, blockData)
	closer.Close()

	block = *Deserialize(blockDataCopy)

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
	db := blockchain.Database.GetRawDB()

	lastHash, closer, err := db.Get([]byte("lh"))
	if err != nil {
		Handle(err)
	}
	lastHashCopy := make([]byte, len(lastHash))
	copy(lastHashCopy, lastHash)
	closer.Close()

	lastBlockData, closer, err := db.Get(lastHashCopy)
	if err != nil {
		Handle(err)
	}
	blockDataCopy := make([]byte, len(lastBlockData))
	copy(blockDataCopy, lastBlockData)
	closer.Close()

	lastBlock := *Deserialize(blockDataCopy)
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
