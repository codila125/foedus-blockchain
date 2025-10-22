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
	if len(b.Transactions) == 0 {
		return []byte{}
	}
	
	transactions := make([][]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		transactions[i] = tx.ID
	}
	tree := merkle.NewMerkleTree(transactions) // Create a new Merkle tree from the transactions

	return tree.RootNode.Data
}

func (b *Block) HashContracts() []byte {
	/*
		Computes the Merkle root of the block's contracts.
		Returns the Merkle root as a byte slice.
	*/
	if len(b.Contracts) == 0 {
		return []byte{}
	}
	
	contracts := make([][]byte, len(b.Contracts))
	for i, ct := range b.Contracts {
		contracts[i] = ct.ID
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

	// Populate MilestonesCore from ct.Milestones (pre-allocate)
	core.Milestones = make([]*MilestoneCore, len(ct.Milestones))
	for i, m := range ct.Milestones {
		core.Milestones[i] = &MilestoneCore{
			Title:       m.Title,
			Description: m.Description,
			Value:       m.Value,
			CreatedAt:   m.CreatedAt,
		}
	}

	// Populate PartiesCore from ct.Parties (pre-allocate)
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
