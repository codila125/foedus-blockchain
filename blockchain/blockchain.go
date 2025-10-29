// Package blockchain provides the core implementation of the Foedus blockchain,
// including its data structures, creation, and maintenance. It manages the
// chain of blocks, their persistence in a database, and the consensus mechanisms
// that ensure the integrity and security of the ledger.
package blockchain

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/cockroachdb/pebble"
	"github.com/codila125/foedus-blockchain/database"
)

const (
	// DBPath defines the directory pattern for storing blockchain data. Each node
	// has its own subdirectory, identified by a unique node ID.
	DBPath = "./temp/blocks_%s"
	// genesisData is a constant string embedded in the coinbase transaction of the
	// genesis block, marking the beginning of the blockchain.
	genesisData = "Genesis Foedus"
	// LastHashKey is the database key used to store the hash of the most recent
	// block in the chain. This allows for quick access to the tip of the blockchain.
	LastHashKey = "lh"
)

// BlockChain represents the full blockchain. It holds a reference to the hash of
// the latest block and the database connection where the blocks are stored.
type BlockChain struct {
	LastHash []byte             // The hash of the most recent block in the chain.
	Database *database.PebbleDB // The underlying database for persistent storage.
}

// BlockChainIterator provides a mechanism to traverse the blockchain from the
// newest block to the oldest. It maintains the hash of the current block being
// examined.
type BlockChainIterator struct {
	CurrentHash []byte             // The hash of the current block in the iteration.
	Database    *database.PebbleDB // A reference to the blockchain's database.
}

// NewBlockChain creates and initializes a new blockchain for a specific node.
// If a blockchain already exists at the specified path, the function will exit.
// Otherwise, it creates a genesis block, stores it in the database, and sets it
// as the first and last block of the chain.
func NewBlockChain(address string, nodeID string) *BlockChain {
	var lastHash []byte

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
	defer func() {
		_ = batch.Close()
	}()

	cbtx := CoinbaseTx(address, genesisData)
	cbct := CoinbaseOp(address, genesisData)
	genesis := Genesis(cbtx, cbct)
	log.Printf("[BLOCKCHAIN] Genesis block created - Hash: %x", genesis.Hash)

	err = batch.Set(genesis.Hash, genesis.SerializeBlock(), nil)
	Handle(err)
	err = batch.Set([]byte(LastHashKey), genesis.Hash, nil)
	Handle(err)
	lastHash = genesis.Hash

	err = rawDB.Apply(batch, &pebble.WriteOptions{Sync: true})
	Handle(err)

	blockchain := BlockChain{lastHash, db}
	log.Printf("[BLOCKCHAIN] Blockchain initialized successfully for node %s", nodeID)
	return &blockchain
}

// ContinueBlockChain loads an existing blockchain from the database for a given node.
// It retrieves the hash of the last block to set the current tip of the chain.
// If no blockchain is found, the application exits.
func ContinueBlockChain(nodeID string) *BlockChain {
	path := fmt.Sprintf(DBPath, nodeID)
	if !database.DBExists(path) {
		log.Printf("[BLOCKCHAIN] No existing blockchain found for node %s", nodeID)
		runtime.Goexit()
	}

	log.Printf("[BLOCKCHAIN] Loading existing blockchain for node %s", nodeID)

	var lastHash []byte

	db, err := database.OpenDB(path)
	Handle(err)

	lastHashBytes, err := db.Get([]byte(LastHashKey))
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

// MineBlock adds a new block to the blockchain. It validates all transactions
// and contracts, performs the Proof-of-Work to find a valid hash, and then adds
// the new block to the database. It also updates the UTXO and ICCT sets to reflect
// the new state.
func (blockchain *BlockChain) MineBlock(transactions []*Transaction, contracts []*Contract) *Block {
	var lastHash []byte
	var lastHeight int

	log.Printf("[MINING] Starting block mining with %d transaction(s) and %d contract(s)", len(transactions), len(contracts))

	txMap := make(map[string]Transaction)

	for _, tx := range transactions {
		txMap[hex.EncodeToString(tx.ID)] = *tx
	}

	for _, tx := range transactions {
		if !blockchain.VerifyTransaction(tx, txMap) {
			log.Panicf("[MINING] Invalid transaction detected: %x", tx.ID)
		}
	}

	for _, ct := range contracts {
		if !blockchain.VerifyContract(ct) {
			log.Panicf("[MINING] Invalid contract detected: %x", ct.ID)
		}
	}

	db := blockchain.Database.GetRawDB()

	lastHashBytes, closer, err := db.Get([]byte(LastHashKey))
	if err != nil {
		log.Panicf("[BLOCKCHAIN] Failed to retrieve last hash: %v", err)
	}
	lastHash = make([]byte, len(lastHashBytes))
	copy(lastHash, lastHashBytes)
	_ = closer.Close()

	lastBlockData, closer, err := db.Get(lastHash)
	if err != nil {
		log.Panicf("[BLOCKCHAIN] Failed to retrieve last block: %v", err)
	}
	blockDataCopy := make([]byte, len(lastBlockData))
	copy(blockDataCopy, lastBlockData)
	_ = closer.Close()

	lastBlock := DeserializeBlock(blockDataCopy)
	lastHeight = lastBlock.Height
	newBlock := CreateBlock(transactions, contracts, lastHash, lastHeight+1)

	batch := db.NewBatch()
	defer func() {
		_ = batch.Close()
	}()

	err = batch.Set(newBlock.Hash, newBlock.SerializeBlock(), nil)
	if err != nil {
		log.Panicf("[BLOCKCHAIN] Failed to store new block: %v", err)
	}
	err = batch.Set([]byte(LastHashKey), newBlock.Hash, nil)
	if err != nil {
		log.Panicf("[BLOCKCHAIN] Failed to update last hash: %v", err)
	}

	blockchain.LastHash = newBlock.Hash
	log.Printf("[MINING] Block mined successfully - Hash: %x, Height: %d", newBlock.Hash, newBlock.Height)

	err = db.Apply(batch, &pebble.WriteOptions{Sync: true})
	Handle(err)

	utxoSet := UTXOSet{blockchain}
	utxoSet.Update(newBlock)

	icctSet := ICCTSet{blockchain}
	icctSet.Update(newBlock)

	return newBlock
}

// AddBlock incorporates a new block into the blockchain, typically received from
// another node in the network. It validates the block and, if it is valid and
// extends the longest chain, updates the blockchain's tip. The UTXO and ICCT sets
// are also updated accordingly.
func (blockchain *BlockChain) AddBlock(block *Block) error {
	db := blockchain.Database.GetRawDB()
	batch := db.NewBatch()
	defer func() {
		_ = batch.Close()
	}()

	_, closer, err := db.Get(block.Hash)
	if err == nil {
		_ = closer.Close()
		return nil
	}

	err = batch.Set(block.Hash, block.SerializeBlock(), nil)
	if err != nil {
		return err
	}

	lastHash, closer, err := db.Get([]byte(LastHashKey))
	if err != nil {
		return fmt.Errorf("could not get last hash: %w", err)
	}
	lastHashCopy := make([]byte, len(lastHash))
	copy(lastHashCopy, lastHash)
	_ = closer.Close()

	lastBlockData, closer, err := db.Get(lastHashCopy)
	if err != nil {
		return fmt.Errorf("could not get last block: %w", err)
	}
	blockDataCopy := make([]byte, len(lastBlockData))
	copy(blockDataCopy, lastBlockData)
	_ = closer.Close()

	lastBlock := DeserializeBlock(blockDataCopy)

	if lastBlock == nil {
		return fmt.Errorf("could not deserialize last block")
	}

	if block.Height > lastBlock.Height {
		err = batch.Set([]byte(LastHashKey), block.Hash, nil)
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

	utxoSet := UTXOSet{blockchain}
	utxoSet.Update(block)

	icctSet := ICCTSet{blockchain}
	icctSet.Update(block)

	return nil
}

// GetBlock retrieves a specific block from the blockchain using its hash.
// It reads the block's data from the database and deserializes it into a
// Block struct.
func (blockchain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	db := blockchain.Database.GetRawDB()
	blockData, closer, err := db.Get(blockHash)
	if err != nil {
		return block, fmt.Errorf("could not get block: %w", err)
	}
	blockDataCopy := make([]byte, len(blockData))
	copy(blockDataCopy, blockData)
	closer.Close()

	block = *DeserializeBlock(blockDataCopy)

	return block, nil
}

// GetBlockHashes returns a list of all block hashes in the blockchain, starting
// from the most recent block and traversing back to the genesis block.
func (blockchain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	iterator := blockchain.Iterator()

	for {
		block := iterator.Next()
		blocks = append(blocks, block.Hash)
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return blocks
}

// GetBestHeight returns the height of the latest block in the blockchain. This
// is a key metric for understanding the current length and state of the chain.
func (blockchain *BlockChain) GetBestHeight() int {
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

	lastBlock := *DeserializeBlock(blockDataCopy)
	return lastBlock.Height
}
