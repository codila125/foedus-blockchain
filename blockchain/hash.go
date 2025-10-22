package blockchain

import (
	"crypto/sha256"
	
	"github.com/codila125/foedus-blockchain/merkle"
)

func (b *Block) HashTransactions() []byte {
	/*
		Computes the Merkle root of the block's transactions.
		Returns the Merkle root as a byte slice.
	*/
	var transactions [][]byte

	// Serialize each transaction and collect them
	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.ID)
	}
	if len(transactions) == 0 {
		return []byte{}
	}
	tree := merkle.NewMerkleTree(transactions) // Create a new Merkle tree from the transactions

	return tree.RootNode.Data
}

func (b *Block) HashContracts() []byte {
	/*
		Computes the Merkle root of the block's contracts.
		Returns the Merkle root as a byte slice.
	*/
	var contracts [][]byte

	// Serialize each contract and collect them
	for _, ct := range b.Contracts {
		contracts = append(contracts, ct.ID)
	}
	if len(contracts) == 0 {
		return []byte{}
	}

	tree := merkle.NewMerkleTree(contracts) // Create a new Merkle tree from the contracts

	return tree.RootNode.Data
}


// HashContract computes the hash of the immutable core of the contract.
// This hash serves as the permanent, unchanging ID of the contract.
// It explicitly excludes mutable fields like Status, UpdatedAt, and Party.Signature.
func (ct *Contract) HashContract() []byte {
	// Create a ContractCore from the current contract, excluding mutable fields
	core := ContractCore{
		Title:          ct.Title,
		Description:    ct.Description,
		CreatorAddress: ct.CreatorAddress,
		Attachments:    ct.Attachments,
		Terms:          ct.Terms,
		CreatedAt:      ct.CreatedAt,
	}

	// Populate MilestonesCore from ct.Milestones
	core.Milestones = make([]*MilestoneCore, len(ct.Milestones))
	for i, m := range ct.Milestones {
		core.Milestones[i] = &MilestoneCore{
			Title:       m.Title,
			Description: m.Description,
			Value:       m.Value,
		}
	}

	// Populate PartiesCore from ct.Parties
	core.Parties = make([]*PartyCore, len(ct.Parties))
	for i, p := range ct.Parties {
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


func (milestone *Milestone) HashMilestones() []byte {
	/*
		Hashes the milestones of the contract for ID generation.
	*/
	core := MilestoneCore{
		Title:       milestone.Title,
		Description: milestone.Description,
		Value:       milestone.Value,
		CreatedAt:   milestone.CreatedAt,
	}

	hash := sha256.Sum256(core.SerializeMilestoneCore())

	return hash[:]
}

func (tx *Transaction) HashTransaction() []byte {
	/*
		Computes the hash of the transaction.
		Returns the SHA-256 hash of the serialized transaction.
	*/
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.SerializeTransaction())

	return hash[:]
}
