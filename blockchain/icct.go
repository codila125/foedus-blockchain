package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/cockroachdb/pebble"
)

// ICCTPrefix is the key prefix used for storing incomplete contract entries in the database.
var ICCTPrefix = []byte("icct-")

// ICCTSet manages the set of incomplete contracts for efficient tracking and lookup.
type ICCTSet struct {
	Blockchain *BlockChain
}

// ContractState wraps a contract with its milestones for efficient storage and retrieval.
type ContractState struct {
	Contract   *Contract
	Milestones []*Milestone
}

// FindICCT scans the entire blockchain to build a map of all incomplete contracts.
// It iterates through each block and, for each contract, records its latest state.
// A contract is considered incomplete if its status is 'ContractDraft' or 'ContractActive'.
// If a contract transitions to a completed or cancelled state, it is removed from the map.
// The resulting map provides a point-in-time snapshot of all active and pending contracts.
func (blockchain *BlockChain) FindICCT() map[string]ContractState {
	ICCT := make(map[string]ContractState)

	iterator := blockchain.Iterator()

	for {
		block := iterator.Next()

		for _, ct := range block.Contracts {
			ctID := hex.EncodeToString(ct.ID)

			// Always update with the latest version of the contract
			state := ContractState{
				Contract:   ct,
				Milestones: ct.Milestones,
			}

			if ct.Status == ContractDraft || ct.Status == ContractActive {
				ICCT[ctID] = state
			} else {
				delete(ICCT, ctID)
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return ICCT
}

// Reindex clears and rebuilds the Incomplete Contract (ICCT) set from the blockchain.
// This function first deletes all existing ICCT entries and then repopulates the set
// by scanning the entire blockchain, ensuring the index is perfectly synchronized with
// the chain's history.
func (i *ICCTSet) Reindex() {
	log.Printf("[ICCT] Starting ICCT set reindexing")

	db := i.Blockchain.Database.GetRawDB()

	i.DeleteByPrefix(ICCTPrefix)

	ICCTs := i.Blockchain.FindICCT()

	batch := db.NewBatch()

	for ctID, state := range ICCTs {
		key, err := hex.DecodeString(ctID)
		if err != nil {
			log.Print(err)
		}
		key = append(ICCTPrefix, key...)

		err = batch.Set(key, state.SerializeContractState(), nil)
		Handle(err)
	}

	err := db.Apply(batch, &pebble.WriteOptions{Sync: true})
	batch.Close()
	Handle(err)

	count := i.CountContracts()
	log.Printf("[ICCT] ICCT set reindexed successfully - %d incomplete contract(s) in set", count)
}

// Update processes a new block and applies its contract-related changes to the ICCT set.
// It iterates through the contracts in the block:
// - If a contract's status is 'ContractDraft' or 'ContractActive', it is added or updated in the set.
// - If a contract's status is terminal (e.g., 'Completed', 'Cancelled'), it is removed from the set.
// This keeps the ICCT set consistent with the latest state of the blockchain.
func (i *ICCTSet) Update(block *Block) {
	log.Printf("[ICCT] Updating ICCT set with block %x", block.Hash)
	db := i.Blockchain.Database.GetRawDB()

	batch := db.NewBatch()

	for _, ct := range block.Contracts {

		ctID := append(ICCTPrefix, ct.ID...)

		// If contract is incomplete (DRAFT or ACTIVE), store/update it
		if ct.Status == ContractDraft || ct.Status == ContractActive {
			state := ContractState{
				Contract:   ct,
				Milestones: ct.Milestones,
			}

			if err := batch.Set(ctID, state.SerializeContractState(), &pebble.WriteOptions{}); err != nil {
				log.Print(err)
			}
			log.Printf("[ICCT] Contract %x updated - Status: %s", ct.ID, ct.Status)
		} else {
			// Contract is COMPLETED or CANCELLED, remove from ICCT set
			if err := batch.Delete(ctID, &pebble.WriteOptions{}); err != nil {
				// If key doesn't exist, that's okay
				if err != pebble.ErrNotFound {
					log.Print(err)
				}
			}
			log.Printf("[ICCT] Contract %x removed - Status: %s", ct.ID, ct.Status)
		}
	}

	if err := db.Apply(batch, pebble.Sync); err != nil {
		log.Print(err)
	}
	batch.Close()

	log.Printf("[ICCT] ICCT set updated successfully")
}

// DeleteByPrefix removes all database entries that match the given key prefix.
// The deletion is performed in batches to manage memory usage and avoid creating a single,
// excessively large database transaction, which could impact performance.
func (i *ICCTSet) DeleteByPrefix(prefix []byte) {
	db := i.Blockchain.Database.GetRawDB()

	iter, _ := db.NewIter(&pebble.IterOptions{})
	defer func() {
		_ = iter.Close()
	}()

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
					log.Print(err)
				}
			}

			if err := db.Apply(batch, &pebble.WriteOptions{Sync: true}); err != nil {
				log.Print(err)
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
				log.Print(err)
			}
		}

		if err := db.Apply(batch, &pebble.WriteOptions{Sync: true}); err != nil {
			log.Print(err)
		}
		batch.Close()
	}
}

// CountContracts iterates through the database and returns the total number of entries
// in the ICCT set, effectively counting all incomplete contracts.
func (i ICCTSet) CountContracts() int {
	db := i.Blockchain.Database.GetRawDB()
	count := 0

	iter, _ := db.NewIter(&pebble.IterOptions{})
	defer func() {
		_ = iter.Close()
	}()

	// Count ICCT entries with the specified prefix
	for iter.SeekGE(ICCTPrefix); iter.Valid(); iter.Next() {
		if !bytes.HasPrefix(iter.Key(), ICCTPrefix) {
			break
		}
		count++
	}

	return count
}

// GetContract retrieves the latest state of a specific contract from the ICCT set using its ID.
// It returns the deserialized contract if found, or an error if the contract does not exist
// in the set or if a database error occurs.
func (i ICCTSet) GetContract(contractID []byte) (*Contract, error) {
	db := i.Blockchain.Database.GetRawDB()
	key := append(ICCTPrefix, contractID...)

	data, closer, err := db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, err
		}
		return nil, err
	}
	defer closer.Close()

	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	state := DeserializeContractState(dataCopy)
	return state.Contract, nil
}
