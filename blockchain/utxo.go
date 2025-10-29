// Package blockchain manages the Unspent Transaction Output (UTXO) set, which is
// a critical component for tracking the ownership of cryptocurrency. The UTXO set
// is an index of all unspent outputs, enabling efficient validation of new
// transactions.
package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/cockroachdb/pebble"
)

// UTXOPrefix is a database key prefix used to distinguish UTXO entries from other
// data stored in the database. This helps in organizing and querying the UTXO set.
var UTXOPrefix = []byte("utxo-")

// UTXOSet provides a high-level interface for managing the collection of unspent
// transaction outputs. It is tightly coupled with the blockchain to ensure that
// the UTXO set is always in sync with the state of the chain.
type UTXOSet struct {
	Blockchain *BlockChain
}

// FindUTXO scans the entire blockchain to identify all unspent transaction outputs.
// It iterates through each block, tracking which outputs have been spent and which
// remain available. This function is foundational for building the UTXO set from scratch.
// It returns a map where keys are transaction IDs and values are the unspent outputs.
func (blockchain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXs := make(map[string]map[int]bool)

	iterator := blockchain.Iterator()

	for {
		block := iterator.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			for outIdx, out := range tx.Outputs {
				if spentTXs[txID] != nil && spentTXs[txID][outIdx] {
					continue
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsCoinbaseTx() {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					if spentTXs[inTxID] == nil {
						spentTXs[inTxID] = make(map[int]bool)
					}
					spentTXs[inTxID][in.Out] = true
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO
}

// Reindex rebuilds the UTXO set from the ground up by scanning the entire
// blockchain. This operation is typically performed when the UTXO set is created
// for the first time or if it becomes corrupted. It ensures the UTXO index is
// consistent with the blockchain's history.
func (u *UTXOSet) Reindex() {
	log.Printf("[UTXO] Starting UTXO set reindexing")

	db := u.Blockchain.Database.GetRawDB()

	u.DeleteByPrefix(UTXOPrefix)

	UTXOs := u.Blockchain.FindUTXO()

	batch := db.NewBatch()

	for txID, outs := range UTXOs {
		key, err := hex.DecodeString(txID)
		if err != nil {
			log.Panic(err)
		}
		key = append(UTXOPrefix, key...)

		err = batch.Set(key, outs.SerializeOutputs(), nil)
		Handle(err)
	}

	err := db.Apply(batch, &pebble.WriteOptions{Sync: true})
	batch.Close()
	Handle(err)

	count := u.CountTransactions()
	log.Printf("[UTXO] UTXO set reindexed successfully - %d transaction(s) in set", count)
}

// Update modifies the UTXO set to reflect the transactions in a newly added block.
// It removes outputs that have been spent and adds the new outputs created in the
// block's transactions. This keeps the UTXO set current with the latest state of
// the blockchain.
func (u *UTXOSet) Update(block *Block) {
	log.Printf("[UTXO] Updating UTXO set with block %x", block.Hash)
	db := u.Blockchain.Database.GetRawDB()

	batch := db.NewBatch()

	for _, tx := range block.Transactions {
		if !tx.IsCoinbaseTx() {
			for _, in := range tx.Inputs {
				updatedOuts := TxOutputs{}
				inID := append(UTXOPrefix, in.ID...)

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

				outs := DeserializeOutputs(itemCopy)

				// Remove spent outputs by creating a new list of unspent ones.
				for outIdx, out := range outs.Outputs {
					if outIdx != in.Out {
						updatedOuts.Outputs = append(updatedOuts.Outputs, out)
					}
				}

				if len(updatedOuts.Outputs) == 0 {
					if err := batch.Delete(inID, nil); err != nil {
						log.Panic(err)
					}
				} else {
					if err := batch.Set(inID, updatedOuts.SerializeOutputs(), nil); err != nil {
						log.Panic(err)
					}
				}
			}
		}

		// Add new outputs from the transaction to the UTXO set.
		newOutputs := TxOutputs{}
		newOutputs.Outputs = append(newOutputs.Outputs, tx.Outputs...)

		txID := append(UTXOPrefix, tx.ID...)
		if err := batch.Set(txID, newOutputs.SerializeOutputs(), nil); err != nil {
			log.Panic(err)
		}
	}

	if err := db.Apply(batch, pebble.Sync); err != nil {
		log.Panic(err)
	}
	batch.Close()

	log.Printf("[UTXO] UTXO set updated successfully")
}

// DeleteByPrefix removes all UTXO entries from the database that match a given
// prefix. This is used during reindexing to clear the old UTXO set. The deletion
// is performed in batches to handle large datasets efficiently without consuming
// excessive memory.
func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	db := u.Blockchain.Database.GetRawDB()

	iter, _ := db.NewIter(&pebble.IterOptions{})
	defer func() {
		_ = iter.Close()
	}()

	collectSize := 100000
	keysForDelete := make([][]byte, 0, collectSize)
	keysCollected := 0

	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		key := iter.Key()
		if !bytes.HasPrefix(key, prefix) {
			break
		}

		keyCopy := make([]byte, len(key))
		copy(keyCopy, key)
		keysForDelete = append(keysForDelete, keyCopy)
		keysCollected++

		// Delete keys in batches to manage memory usage.
		if keysCollected == collectSize {
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
			keysForDelete = make([][]byte, 0, collectSize)
			keysCollected = 0
		}
	}

	// Delete any remaining keys.
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

// CountTransactions returns the total number of transactions currently held in
// the UTXO set. This provides a quick way to gauge the size of the UTXO set.
func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.Database.GetRawDB()
	count := 0

	iter, _ := db.NewIter(&pebble.IterOptions{})
	defer func() {
		_ = iter.Close()
	}()

	for iter.SeekGE(UTXOPrefix); iter.Valid(); iter.Next() {
		if !bytes.HasPrefix(iter.Key(), UTXOPrefix) {
			break
		}
		count++
	}

	return count
}

// FindUnspentTransactions retrieves all unspent transaction outputs (UTXOs)
// that are locked with the given public key hash. This function is essential
// for calculating a user's balance and gathering inputs for a new transaction.
func (u UTXOSet) FindUnspentTransactions(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	db := u.Blockchain.Database.GetRawDB()

	iter, _ := db.NewIter(&pebble.IterOptions{})
	defer func() {
		_ = iter.Close()
	}()

	for iter.SeekGE(UTXOPrefix); iter.Valid(); iter.Next() {
		if !bytes.HasPrefix(iter.Key(), UTXOPrefix) {
			break
		}

		v := iter.Value()
		valueCopy := make([]byte, len(v))
		copy(valueCopy, v)

		outs := DeserializeOutputs(valueCopy)

		for _, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// FindSpendableOutputs discovers unspent transaction outputs that can be used to
// fund a new transaction. It accumulates outputs locked with the given public key
// hash until the total value meets or exceeds the required amount. It returns the
// accumulated value and a map of transaction IDs to their spendable output indices.
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Database.GetRawDB()

	iter, _ := db.NewIter(&pebble.IterOptions{})
	defer func() {
		_ = iter.Close()
	}()

	for iter.SeekGE(UTXOPrefix); iter.Valid(); iter.Next() {
		if !bytes.HasPrefix(iter.Key(), UTXOPrefix) {
			break
		}

		key := iter.Key()
		v := iter.Value()

		keyTrimmed := bytes.TrimPrefix(key, UTXOPrefix)
		keyTrimmedCopy := make([]byte, len(keyTrimmed))
		copy(keyTrimmedCopy, keyTrimmed)
		txID := hex.EncodeToString(keyTrimmedCopy)

		valueCopy := make([]byte, len(v))
		copy(valueCopy, v)
		outs := DeserializeOutputs(valueCopy)

		for outIdx, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
			}
		}
	}

	return accumulated, unspentOutputs
}
