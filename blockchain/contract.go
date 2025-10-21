package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
)

type ContractStatus string

const (
	ContractDraft     ContractStatus = "DRAFT"
	ContractActive    ContractStatus = "ACTIVE"
	ContractCompleted ContractStatus = "COMPLETED"
	ContractCancelled ContractStatus = "CANCELLED"
)

type MilestoneStatus string

const (
	MilestoneActive   MilestoneStatus = "ACTIVE"
	MilestoneCompleted MilestoneStatus = "COMPLETED"
	MilestoneRejected  MilestoneStatus = "DISPUTED"
	MilestoneCancelled MilestoneStatus = "CANCELLED"
)

// Milestone represents a deliverable within a contract
type Milestone struct {
	ID          []byte          // Unique identifier for the milestone
	Title       string          // Milestone title/description
	Description string          // Detailed description
	Value       int             // Payment amount for this milestone
	DueDate     int64           // Unix timestamp for due date
	Status      MilestoneStatus // Current status
	CreatedAt   int64           // Creation timestamp
	CompletedAt int64           // Completion timestamp (0 if not completed)
	Evidence    []byte          // Hash of evidence/deliverables (IPFS hash, etc.)
	ApprovedBy  []string        // Addresses of parties who approved the milestone
	DisputedBy  []string        // Addresses of parties who disputed the milestone
}

// Contract represents an agreement between parties with milestone-based payments
type Contract struct {
	ID             []byte         // Unique contract identifier
	Title          string         // Contract title
	Description    string         // Detailed description
	CreatedAt      int64          // Creation timestamp
	UpdatedAt      int64          // Last update timestamp
	Status         ContractStatus // Current contract status
	Milestones     []*Milestone   // Array of milestones
	Parties        []*Party       // Involved parties (use pointers so signatures persist)
	Terms      []byte         // Hash of contract terms
	CreatorAddress string         // Address of contract creator
	Attachments    [][]byte       // Array of attachment hashes (IPFS hashes, etc.)
}

// Party represents a participant in a contract
type Party struct {
	Address   string       // Wallet address
	Role      ContractRole // Role in the contract (e.g., "CLIENT", "CONTRACTOR")
	PublicKey []byte       // Public key of the party
	Signature []byte       // Digital signature of the party
}

type ContractRole string

const (
	RoleContractor ContractRole = "CONTRACTOR"
	RoleArbitrator ContractRole = "ARBITRATOR"
	RoleCreator    ContractRole = "CREATOR"
)

type ContractCore struct {
    Title          string
    Description    string
    CreatedAt      int64
    Milestones     []*MilestoneCore // Use a core version of Milestone
    Parties        []*PartyCore     // Use a core version of Party
	Terms      []byte
    DisputeHandler string
    CreatorAddress string
	Attachments    [][]byte       // Array of attachment hashes (IPFS hashes, etc.)
}

// PartyCore represents the immutable parts of a Party for contract ID generation.
// Signature is excluded as it's mutable.
type PartyCore struct {
    Address   string
    Role      string
    PublicKey []byte
}

// MilestoneCore represents the immutable parts of a Milestone for contract ID generation.
// Status, CompletedAt, Evidence are mutable and excluded.
type MilestoneCore struct {
    Title       string
    Description string
    Value       int
	CreatedAt   int64
}

// SerializeContractCore serializes the immutable core of the contract into a byte array.
func (cc *ContractCore) SerializeContractCore() []byte {
    var buff bytes.Buffer
    enc := gob.NewEncoder(&buff)
    err := enc.Encode(cc)
    if err != nil {
        log.Panic(err)
    }
    return buff.Bytes()
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
        Terms:         ct.Terms,
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