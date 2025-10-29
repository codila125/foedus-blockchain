// Package blockchain provides hashing functionalities for core data structures
// such as blocks, transactions, and contracts. These functions are essential
// for ensuring data integrity and creating unique identifiers for each component.
package blockchain

import (
	"crypto/sha256"

	"github.com/codila125/foedus-blockchain/merkle"
)

// HashTransactions generates a Merkle root hash for all transactions within a
// block. This allows for efficient verification of transaction inclusion without
// having to process the entire list of transactions. If there are no transactions,
// it returns an empty byte slice.
func (b *Block) HashTransactions() []byte {
	if len(b.Transactions) == 0 {
		return []byte{}
	}

	transactions := make([][]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		transactions[i] = tx.ID
	}
	tree := merkle.NewMerkleTree(transactions)

	return tree.RootNode.Data
}

// HashContracts computes the Merkle root hash for all smart contracts within a
// block. Similar to HashTransactions, this provides an efficient way to verify
// the integrity and inclusion of contracts. It returns an empty byte slice if
// the block contains no contracts.
func (b *Block) HashContracts() []byte {
	if len(b.Contracts) == 0 {
		return []byte{}
	}

	contracts := make([][]byte, len(b.Contracts))
	for i, contract := range b.Contracts {
		contracts[i] = contract.ID
	}

	tree := merkle.NewMerkleTree(contracts)

	return tree.RootNode.Data
}

// HashContract generates a unique, deterministic hash for a smart contract by
// serializing and hashing its immutable core components. This hash serves as the
// contract's permanent ID. Mutable fields like status and signatures are excluded
// to ensure the ID remains constant.
func (contract *Contract) HashContract() []byte {
	// Create a ContractCore from the current contract, excluding mutable fields.
	core := ContractCore{
		Title:          contract.Title,
		Description:    contract.Description,
		CreatorAddress: contract.CreatorAddress,
		Attachments:    contract.Attachments,
		Terms:          contract.Terms,
		CreatedAt:      contract.CreatedAt,
	}

	// Populate MilestonesCore from contract.Milestones, pre-allocating for efficiency.
	core.Milestones = make([]*MilestoneCore, len(contract.Milestones))
	for i, m := range contract.Milestones {
		core.Milestones[i] = &MilestoneCore{
			Title:       m.Title,
			Description: m.Description,
			Value:       m.Value,
			CreatedAt:   m.CreatedAt,
		}
	}

	// Populate PartiesCore from contract.Parties, pre-allocating for efficiency.
	core.Parties = make([]*PartyCore, len(contract.Parties))
	for i, p := range contract.Parties {
		core.Parties[i] = &PartyCore{
			Address:   p.Address,
			Role:      string(p.Role),
			PublicKey: p.PublicKey,
		}
	}

	var hash [32]byte
	serializedCore := core.SerializeContractCore()
	hash = sha256.Sum256(serializedCore)

	return hash[:]
}

// HashMilestones computes the SHA-256 hash of a milestone's immutable core
// attributes. This is used to generate a unique identifier for the milestone,
// ensuring its integrity within the contract.
func (milestone *Milestone) HashMilestones() []byte {
	core := MilestoneCore{
		Title:       milestone.Title,
		Description: milestone.Description,
		Value:       milestone.Value,
		CreatedAt:   milestone.CreatedAt,
	}

	hash := sha256.Sum256(core.SerializeMilestoneCore())

	return hash[:]
}

// HashTransaction generates a unique ID for a transaction by hashing its
// contents. A copy of the transaction is created with an empty ID field before
// serialization and hashing to ensure that the hash is deterministic and not
// influenced by any pre-existing ID.
func (tx *Transaction) HashTransaction() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.SerializeTransaction())

	return hash[:]
}
