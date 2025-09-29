package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	badger "github.com/dgraph-io/badger/v4"
)

var UTXOPrefix = []byte("utxo-")

type UTXOSet struct {
	Blockchain *BlockChain
}

func (u *UTXOSet) Reindex() {
	db := u.Blockchain.Database

	u.DeleteByPrefix(UTXOPrefix)

	UTXOs := u.Blockchain.FindUTXO()

	err := db.Update(func(txn *badger.Txn) error {
		for txID, outs := range UTXOs {
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}
			key = append(UTXOPrefix, key...)

			err = txn.Set(key, outs.Serialize())
			Handle(err)
		}
		return nil
	})
	Handle(err)
}

func (u *UTXOSet) Update(block *Block) {
	db := u.Blockchain.Database

	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}
					inID := append(UTXOPrefix, in.ID...)
					item, err := txn.Get(inID)
					Handle(err)
					v, err := item.ValueCopy(nil)
					Handle(err)

					outs := DeserializeOutputs(v)
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
}

func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := u.Blockchain.Database.Update(func(txn *badger.Txn) error {
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

	collectSize := 100000
	u.Blockchain.Database.View(func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		options.PrefetchValues = false
		iterator := txn.NewIterator(options)
		defer iterator.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0

		for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
			key := iterator.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					log.Panic(err)
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}

func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.Database
	count := 0

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
	var UTXOs []TxOutput
	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		iterator := txn.NewIterator(options)
		defer iterator.Close()

		for iterator.Seek(UTXOPrefix); iterator.ValidForPrefix(UTXOPrefix); iterator.Next() {
			item := iterator.Item()
			v, err := item.ValueCopy(nil)
			Handle(err)
			outs := DeserializeOutputs(v)
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
	unspentOutputs := make(map[string][]int)
	acumulated := 0
	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		iterator := txn.NewIterator(options)
		defer iterator.Close()

		for iterator.Seek(UTXOPrefix); iterator.ValidForPrefix(UTXOPrefix); iterator.Next() {
			item := iterator.Item()
			key := item.Key()
			v, err := item.ValueCopy(nil)
			Handle(err)
			key = bytes.TrimPrefix(key, UTXOPrefix)
			txID := hex.EncodeToString(key)
			outs := DeserializeOutputs(v)

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
