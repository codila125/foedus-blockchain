package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	badger "github.com/dgraph-io/badger/v4"
)

const (
	dbPath      = "./temp/blocks"
	dbFile      = "./temp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash []byte     // Hash of the last block in the chain
	Database *badger.DB // Reference to the BadgerDB database
}

type BlockChainIterator struct {
	CurrentHash []byte     // Hash of the current block in the iteration
	Database    *badger.DB // Reference to the BadgerDB database
}

func DBExists() bool {
	/*
		Checks if the blockchain database file exists.
		Returns true if the database file exists, false otherwise.
	*/
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func NewBlockChain(address string) *BlockChain {
	/*
		Creates a new blockchain with a genesis block and stores it in the database.
		'address' is the address to send the coinbase reward to.
		Returns a pointer to the newly created BlockChain instance.
	*/
	var lastHash []byte

	// Ensure a blockchain does not already exist
	if DBExists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	options := badger.DefaultOptions(dbPath) // Set default options for BadgerDB

	db, err := badger.Open(options)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData) // Create the coinbase transaction for the genesis block
		genesis := Genesis(cbtx)                 // Create the genesis block
		fmt.Println("Genesis created")
		err := txn.Set(genesis.Hash, genesis.Serialize()) // Store the genesis block in the database
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash) // Store the last hash pointer because it helps to find the last block

		lastHash = genesis.Hash // Set the last hash to the genesis block's hash

		return err
	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func ContinueBlockChain() *BlockChain {
	/*
		Continues an existing blockchain by loading it from the database.
		Returns a pointer to the BlockChain instance.
	*/
	if !DBExists() {
		fmt.Println("Blockchain does not exist")
		runtime.Goexit()
	}

	var lastHash []byte

	options := badger.DefaultOptions(dbPath)
	db, err := badger.Open(options)
	Handle(err)

	// Read the last hash from the database
	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		return err
	})

	Handle(err)

	chain := BlockChain{lastHash, db}
	return &chain
}

func (blockchain *BlockChain) AddBlock(transactions []*Transaction) *Block {
	/*
		Adds a new block with the given transactions to the blockchain.
		'transactions' is a slice of pointers to Transaction instances to be included in the new block.
		Returns a pointer to the newly added Block instance.
	*/
	var lastHash []byte

	// Get the last hash from the database
	err := blockchain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		return err
	})

	Handle(err)

	newBlock := CreateBlock(transactions, lastHash) // Create a new block with the transactions and previous hash

	// Store the new block in the database and update the last hash pointer
	err = blockchain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		blockchain.LastHash = newBlock.Hash
		return err
	})

	Handle(err)
	return newBlock
}

func (blockchain *BlockChain) Iterator() *BlockChainIterator {
	/*
		Creates and returns a new BlockChainIterator starting from the last block in the chain.
	*/
	iterator := &BlockChainIterator{blockchain.LastHash, blockchain.Database}
	return iterator
}

func (iter *BlockChainIterator) Next() *Block {
	/*
		Fetch the next block in the chain using the current hash stored in the iterator.
		Update the iterator's current hash to the previous block's hash for the next call.
		Return the deserialized block.
	*/
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash) // Get the block data for the current hash
		Handle(err)
		encodedBlock, err := item.ValueCopy(nil) // Copy the block data
		block = Deserialize(encodedBlock)        // Deserialize the block

		return err
	})

	Handle(err)

	iter.CurrentHash = block.PrevHash // Move to the previous block for the next iteration

	return block
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
	return Transaction{}, fmt.Errorf("Transaction is not found")
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

func (blockchain *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := blockchain.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
