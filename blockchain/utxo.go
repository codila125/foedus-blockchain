package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/cockroachdb/pebble"
)

var UTXOPrefix = []byte("utxo-") // Prefix for UTXO entries in the database

type UTXOSet struct {
	Blockchain *BlockChain // Reference to the blockchain
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
			if !tx.IsCoinbaseTx() {
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

func (u *UTXOSet) Reindex() {
	/*
		Rebuilds the UTXO set into the database from scratch
		by deleting the previous UTXO set and scanning the entire
		blockchain to find all unspent transaction outputs.
	*/
	log.Printf("[UTXO] Starting UTXO set reindexing")

	db := u.Blockchain.Database.GetRawDB() // Get the database from the blockchain

	u.DeleteByPrefix(UTXOPrefix) // Delete existing UTXO entries with the specified prefix

	UTXOs := u.Blockchain.FindUTXO() // Find all unspent transaction outputs in the blockchain

	batch := db.NewBatch()

	for txID, outs := range UTXOs { // Iterate over each transaction ID and its outputs
		key, err := hex.DecodeString(txID)
		if err != nil {
			log.Panic(err)
		}
		key = append(UTXOPrefix, key...) // Prefix the key with "utxo-"

		err = batch.Set(key, outs.SerializeOutputs(), nil)
		Handle(err)
	}

	err := db.Apply(batch, &pebble.WriteOptions{Sync: true})
	batch.Close()
	Handle(err)

	count := u.CountTransactions()
	log.Printf("[UTXO] UTXO set reindexed successfully - %d transaction(s) in set", count)
}

func (u *UTXOSet) Update(block *Block) {
	/*
		Updates the UTXO set with the transactions from the given block.
		Removes spent outputs and adds new outputs from the block's transactions.
	*/
	log.Printf("[UTXO] Updating UTXO set with block %x", block.Hash)
	db := u.Blockchain.Database.GetRawDB()

	batch := db.NewBatch()

	for _, tx := range block.Transactions {
		if !tx.IsCoinbaseTx() {
			for _, in := range tx.Inputs {
				updatedOuts := TxOutputs{}
				inID := append(UTXOPrefix, in.ID...) // Prefix the input transaction ID with "utxo-"

				item, closer, err := db.Get(inID)
				if err != nil && err != pebble.ErrNotFound {
					closer.Close()
					Handle(err)
					continue
				}
				if err == pebble.ErrNotFound {
					continue
				}

				itemCopy := make([]byte, len(item))
				copy(itemCopy, item)
				closer.Close()

				outs := DeserializeOutputs(itemCopy) // Deserialize the outputs

				// Remove the spent output from the outputs list
				// by adding only those outputs which are not spent
				// to the updated outputs list
				// If all outputs are spent, the entry will be deleted
				// from the database
				// If some outputs remain unspent, the entry will be updated
				// with the remaining unspent outputs
				// This ensures that the UTXO set accurately reflects
				// the current state of unspent outputs after processing
				// the transactions in the block
				for outIdx, out := range outs.Outputs {
					if outIdx != in.Out {
						updatedOuts.Outputs = append(updatedOuts.Outputs, out)
					}
				}

				if len(updatedOuts.Outputs) == 0 {
					if err := batch.Delete(inID, &pebble.WriteOptions{}); err != nil {
						log.Panic(err)
					}
				} else {
					if err := batch.Set(inID, updatedOuts.SerializeOutputs(), &pebble.WriteOptions{}); err != nil {
						log.Panic(err)
					}
				}
			}

			// Add new outputs from the transaction to the UTXO set
			// by creating a new UTXO entry for the transaction ID
			// and storing the serialized outputs in the database
			// This ensures that the UTXO set includes all new outputs
			// created by the transactions in the block
			newOutputs := TxOutputs{}
			newOutputs.Outputs = append(newOutputs.Outputs, tx.Outputs...)

			txID := append(UTXOPrefix, tx.ID...)
			if err := batch.Set(txID, newOutputs.SerializeOutputs(), &pebble.WriteOptions{}); err != nil {
				log.Panic(err)
			}
		}
	}

	if err := db.Apply(batch, pebble.Sync); err != nil {
		log.Panic(err)
	}
	batch.Close()

	log.Printf("[UTXO] UTXO set updated successfully")
}

func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	/*
		Delete all UTXO entries with the given prefix in chunks
	*/
	db := u.Blockchain.Database.GetRawDB()

	iter, _ := db.NewIter(&pebble.IterOptions{})
	defer iter.Close()

	collectSize := 100000
	keysForDelete := make([][]byte, 0, collectSize)
	keysCollected := 0

	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		key := iter.Key()
		// Stop if we've gone past the prefix
		if !bytes.HasPrefix(key, prefix) {
			break
		}

		keyCopy := make([]byte, len(key))
		copy(keyCopy, key)
		keysForDelete = append(keysForDelete, keyCopy)
		keysCollected++

		// If we've collected enough keys, delete them in a batch
		if keysCollected == collectSize {
			// Delete collected keys in batch
			batch := db.NewBatch()

			for _, delKey := range keysForDelete {
				if err := batch.Delete(delKey, nil); err != nil {
					log.Panic(err)
				}
			}

			if err := db.Apply(batch, &pebble.WriteOptions{Sync: true}); err != nil {
				log.Panic(err)
			}
			batch.Close()

			keysForDelete = make([][]byte, 0, collectSize) // Reset the slice for the next batch
			keysCollected = 0
		}
	}

	// Delete any remaining keys that didn't make up a full batch
	if keysCollected > 0 {
		batch := db.NewBatch()

		for _, delKey := range keysForDelete {
			if err := batch.Delete(delKey, nil); err != nil {
				log.Panic(err)
			}
		}

		if err := db.Apply(batch, &pebble.WriteOptions{Sync: true}); err != nil {
			log.Panic(err)
		}
		batch.Close()
	}
}

func (u UTXOSet) CountTransactions() int {
	/*
		Counts the number of transactions in the UTXO set.
		Returns the count as an integer.
	*/
	db := u.Blockchain.Database.GetRawDB()
	count := 0

	// Create iterator for prefix scan
	iter, _ := db.NewIter(&pebble.IterOptions{})
	defer iter.Close()

	// Count UTXO entries with the specified prefix
	for iter.SeekGE(UTXOPrefix); iter.Valid(); iter.Next() {
		if !bytes.HasPrefix(iter.Key(), UTXOPrefix) {
			break
		}
		count++
	}

	return count
}

func (u UTXOSet) FindUnspentTransactions(pubKeyHash []byte) []TxOutput {
	/*
		Finds and returns all unspent transaction outputs (UTXOs) for the given public key hash.
	*/
	var UTXOs []TxOutput
	db := u.Blockchain.Database.GetRawDB()

	// Create iterator for prefix scan
	iter, _ := db.NewIter(&pebble.IterOptions{})
	defer iter.Close()

	// Iterate over all UTXO entries with the specified prefix
	for iter.SeekGE(UTXOPrefix); iter.Valid(); iter.Next() {
		if !bytes.HasPrefix(iter.Key(), UTXOPrefix) {
			break
		}

		// Copy value before using it
		v := iter.Value()
		valueCopy := make([]byte, len(v))
		copy(valueCopy, v)

		outs := DeserializeOutputs(valueCopy) // Deserialize the outputs

		// Check each output to see if it is locked with the given public key hash i.e., belongs to the address
		for _, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	/*
		Finds and returns spendable outputs for a given public key hash that sum up to at least the specified amount.
		Returns the total accumulated amount and a map of transaction IDs to output indices.
	*/
	unspentOutputs := make(map[string][]int) // Map to hold unspent outputs
	accumulated := 0
	db := u.Blockchain.Database.GetRawDB()

	// Create iterator for prefix scan
	iter, _ := db.NewIter(&pebble.IterOptions{})
	defer iter.Close()

	// Iterate over all UTXO entries
	for iter.SeekGE(UTXOPrefix); iter.Valid(); iter.Next() {
		if !bytes.HasPrefix(iter.Key(), UTXOPrefix) {
			break
		}

		key := iter.Key()
		v := iter.Value()

		// Copy key and value before using them
		keyTrimmed := bytes.TrimPrefix(key, UTXOPrefix)
		keyTrimmedCopy := make([]byte, len(keyTrimmed))
		copy(keyTrimmedCopy, keyTrimmed)
		txID := hex.EncodeToString(keyTrimmedCopy)

		valueCopy := make([]byte, len(v))
		copy(valueCopy, v)
		outs := DeserializeOutputs(valueCopy)

		// Check each output to see if it is locked with the given public key hash
		// If it is, add its value to the accumulated amount and record its index
		// Stop if we've accumulated enough to cover the requested amount
		for outIdx, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
			}
		}
	}

	return accumulated, unspentOutputs
}
