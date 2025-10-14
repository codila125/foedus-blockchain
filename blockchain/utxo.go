package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	badger "github.com/dgraph-io/badger/v4"
)

var UTXOPrefix = []byte("utxo-") // Prefix for UTXO entries in the database

type UTXOSet struct {
	Blockchain *BlockChain // Reference to the blockchain
}

func (u *UTXOSet) Reindex() {
	/*
		Rebuilds the UTXO set into the database from scratch
		by deleting the previous UTXO set and scanning the entire
		blockchain to find all unspent transaction outputs.
	*/
	log.Printf("[UTXO] Starting UTXO set reindexing")
	db := u.Blockchain.Database // Get the database from the blockchain

	u.DeleteByPrefix(UTXOPrefix) // Delete existing UTXO entries with the specified prefix

	UTXOs := u.Blockchain.FindUTXO() // Find all unspent transaction outputs in the blockchain

	err := db.Update(func(txn *badger.Txn) error {
		for txID, outs := range UTXOs { // Iterate over each transaction ID and its outputs
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}
			key = append(UTXOPrefix, key...) // Prefix the key with "utxo-"

			err = txn.Set(key, outs.Serialize()) // Store the serialized outputs in the database
			Handle(err)
		}
		return nil
	})
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
	db := u.Blockchain.Database

	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}
					inID := append(UTXOPrefix, in.ID...) // Prefix the input transaction ID with "utxo-"
					item, err := txn.Get(inID)           // Get the UTXO entry for the input transaction ID
					Handle(err)
					v, err := item.ValueCopy(nil) // Copy the value of the item
					Handle(err)

					outs := DeserializeOutputs(v) // Deserialize the outputs

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
						if err := txn.Delete(inID); err != nil {
							log.Panic(err)
						} else {
							if err := txn.Set(inID, updatedOuts.Serialize()); err != nil {
								log.Panic(err)
							}
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
				if err := txn.Set(txID, newOutputs.Serialize()); err != nil {
					log.Panic(err)
				}
			}
		}
		return nil
	})
	Handle(err)
	log.Printf("[UTXO] UTXO set updated successfully")
}

func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	/*
		Delete all UTXO entries with the given prefix in chunks
	*/
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := u.Blockchain.Database.Update(func(txn *badger.Txn) error { // Perform deletion atomically
			// Iterate over collected keys and delete them
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	collectSize := 100000                                    // Number of keys to collect before deleting in batch
	u.Blockchain.Database.View(func(txn *badger.Txn) error { // Read-only transaction
		options := badger.DefaultIteratorOptions
		options.PrefetchValues = false // We only need keys, not values, so disable prefetching values, saving memory
		iterator := txn.NewIterator(options)
		defer iterator.Close()

		keysForDelete := make([][]byte, 0, collectSize) // Slice to hold keys to be deleted
		keysCollected := 0

		for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() { // Iterate over keys with the given prefix
			key := iterator.Item().KeyCopy(nil)        // Copy the key to avoid referencing iterator memory
			keysForDelete = append(keysForDelete, key) // Add key to the slice for deletion
			keysCollected++
			// If we've collected enough keys, delete them in a batch
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					log.Panic(err)
				}
				keysForDelete = make([][]byte, 0, collectSize) // Reset the slice for the next batch
				keysCollected = 0
			}
		}
		// Delete any remaining keys that didn't make up a full batch
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}

func (u UTXOSet) CountTransactions() int {
	/*
		Counts the number of transactions in the UTXO set.
		Returns the count as an integer.
	*/
	db := u.Blockchain.Database
	count := 0

	// Count the number of UTXO entries with the specified prefix
	err := db.View(func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		iterator := txn.NewIterator(options)

		defer iterator.Close()
		for iterator.Seek(UTXOPrefix); iterator.ValidForPrefix(UTXOPrefix); iterator.Next() {
			count++
		}
		return nil
	})
	Handle(err)
	return count
}

func (u UTXOSet) FindUnspentTransactions(pubKeyHash []byte) []TxOutput {
	/*
		Finds and returns all unspent transaction outputs (UTXOs) for the given public key hash.
	*/
	var UTXOs []TxOutput
	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		iterator := txn.NewIterator(options)
		defer iterator.Close()

		// Iterate over all UTXO entries with the specified prefix
		for iterator.Seek(UTXOPrefix); iterator.ValidForPrefix(UTXOPrefix); iterator.Next() {
			item := iterator.Item()       // Get the current item
			v, err := item.ValueCopy(nil) // Copy the value of the item
			Handle(err)
			outs := DeserializeOutputs(v) // Deserialize the outputs

			// Check each output to see if it is locked with the given public key hash i.e., belongs to the address
			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	Handle(err)

	return UTXOs
}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	/*
		Finds and returns spendable outputs for a given public key hash that sum up to at least the specified amount.
		Returns the total accumulated amount and a map of transaction IDs to output indices.
	*/
	unspentOutputs := make(map[string][]int) // Map to hold unspent outputs
	acumulated := 0
	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		iterator := txn.NewIterator(options)
		defer iterator.Close()

		for iterator.Seek(UTXOPrefix); iterator.ValidForPrefix(UTXOPrefix); iterator.Next() {
			item := iterator.Item()
			key := item.KeyCopy(nil)      // Copy the key to avoid referencing iterator memory
			v, err := item.ValueCopy(nil) // Copy the value of the item
			Handle(err)
			key = bytes.TrimPrefix(key, UTXOPrefix) // Remove the prefix to get the original transaction ID
			txID := hex.EncodeToString(key)         // Encode the transaction ID to a string
			outs := DeserializeOutputs(v)           // Deserialize the outputs

			// Check each output to see if it is locked with the given public key hash
			// If it is, add its value to the accumulated amount and record its index
			// Stop if we've accumulated enough to cover the requested amount
			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && acumulated < amount {
					acumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})
	Handle(err)

	return acumulated, unspentOutputs
}
