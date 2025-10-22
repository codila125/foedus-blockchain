package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/cockroachdb/pebble"
)

var ICCTPrefix = []byte("icct-") // Prefix for Incomplete Contract Tracking entries in the database

type ICCTSet struct {
	Blockchain *BlockChain // Reference to the blockchain
}

// ContractState wraps a contract with its milestones for storage
type ContractState struct {
	Contract   *Contract
	Milestones []*Milestone
}

func (blockchain *BlockChain) FindICCT() map[string]ContractState {
	/*
		Scans the entire blockchain to find all incomplete contracts.
		Returns a map where the key is the contract ID and the value is the latest contract state.
	*/
	ICCT := make(map[string]ContractState) // ICCT map to hold incomplete contracts

	iterator := blockchain.Iterator() // Create an iterator to traverse the blockchain

	for {
		block := iterator.Next() // Get the next block

		for _, ct := range block.Contracts { // Iterate over each contract in the block
			ctID := hex.EncodeToString(ct.ID) // Encode contract ID to string

			// Always update with the latest version of the contract
			state := ContractState{
				Contract:   ct,
				Milestones: ct.Milestones,
			}

			// Only track incomplete contracts (DRAFT or ACTIVE)
			if ct.Status == ContractDraft || ct.Status == ContractActive {
				ICCT[ctID] = state
			} else {
				// Remove completed or cancelled contracts
				delete(ICCT, ctID)
			}
		}

		// If we've reached the genesis block, stop iterating
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return ICCT
}

func (i *ICCTSet) Reindex() {
	/*
		Rebuilds the ICCT set into the database from scratch
		by deleting the previous ICCT set and scanning the entire
		blockchain to find all incomplete contracts.
	*/
	log.Printf("[ICCT] Starting ICCT set reindexing")

	db := i.Blockchain.Database.GetRawDB() // Get the database from the blockchain

	i.DeleteByPrefix(ICCTPrefix) // Delete existing ICCT entries with the specified prefix

	ICCTs := i.Blockchain.FindICCT() // Find all incomplete contracts in the blockchain

	batch := db.NewBatch()

	for ctID, state := range ICCTs { // Iterate over each contract ID and its state
		key, err := hex.DecodeString(ctID)
		if err != nil {
			log.Panic(err)
		}
		key = append(ICCTPrefix, key...) // Prefix the key with "icct-"

		err = batch.Set(key, state.SerializeContractState(), nil)
		Handle(err)
	}

	err := db.Apply(batch, &pebble.WriteOptions{Sync: true})
	batch.Close()
	Handle(err)

	count := i.CountContracts()
	log.Printf("[ICCT] ICCT set reindexed successfully - %d incomplete contract(s) in set", count)
}

func (i *ICCTSet) Update(block *Block) {
	/*
		Updates the ICCT set with the contracts from the given block.
		Adds new incomplete contracts and updates existing ones with their latest state.
		Removes completed or cancelled contracts from the set.
	*/
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
				log.Panic(err)
			}
			log.Printf("[ICCT] Contract %x updated - Status: %s", ct.ID, ct.Status)
		} else {
			// Contract is COMPLETED or CANCELLED, remove from ICCT set
			if err := batch.Delete(ctID, &pebble.WriteOptions{}); err != nil {
				// If key doesn't exist, that's okay
				if err != pebble.ErrNotFound {
					log.Panic(err)
				}
			}
			log.Printf("[ICCT] Contract %x removed - Status: %s", ct.ID, ct.Status)
		}
	}

	if err := db.Apply(batch, pebble.Sync); err != nil {
		log.Panic(err)
	}
	batch.Close()

	log.Printf("[ICCT] ICCT set updated successfully")
}

func (i *ICCTSet) DeleteByPrefix(prefix []byte) {
	/*
		Delete all ICCT entries with the given prefix in chunks
	*/
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

func (i ICCTSet) CountContracts() int {
	/*
		Counts the number of incomplete contracts in the ICCT set.
		Returns the count as an integer.
	*/
	db := i.Blockchain.Database.GetRawDB()
	count := 0

	// Create iterator for prefix scan
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

func (i ICCTSet) GetContract(contractID []byte) (*Contract, error) {
	/*
		Retrieves the latest state of a contract from the ICCT set.
		Returns the contract and nil error if found, otherwise returns an error.
	*/
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
