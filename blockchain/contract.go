// Package blockchain defines the data structures for smart contracts, including
// their status, milestones, and participating parties. These structures are fundamental
// for creating and managing legally binding agreements on the Foedus blockchain.
package blockchain

import "errors"

// ContractStatus represents the lifecycle state of a smart contract. Each status
// indicates a distinct phase, from creation to completion or cancellation.
type ContractStatus string

// Defines the set of possible statuses for a contract.
const (
	// ContractDraft indicates that the contract has been created but is not yet active.
	// In this state, parties can review the terms before committing.
	ContractDraft ContractStatus = "DRAFT"
	// ContractActive signifies that all parties have agreed to the terms, and the
	// contract is now legally binding and in effect.
	ContractActive ContractStatus = "ACTIVE"
	// ContractCompleted means that all milestones have been fulfilled and the
	// contract has been successfully concluded.
	ContractCompleted ContractStatus = "COMPLETED"
	// ContractCancelled indicates that the contract has been terminated prematurely,
	// either by mutual agreement or due to a dispute.
	ContractCancelled ContractStatus = "CANCELLED"
)

// MilestoneStatus represents the state of an individual milestone within a contract.
// It tracks the progress of specific deliverables.
type MilestoneStatus string

// Defines the set of possible statuses for a contract milestone.
const (
	// MilestoneActive indicates that the milestone is currently in progress and
	// has not yet been completed.
	MilestoneActive MilestoneStatus = "ACTIVE"
	// MilestoneCompleted signifies that the deliverable for the milestone has been
	// approved and payment can be released.
	MilestoneCompleted MilestoneStatus = "COMPLETED"
	// MilestoneCancelled indicates that the milestone has been voided and will not
	// be pursued.
	MilestoneCancelled MilestoneStatus = "CANCELLED"
)

// Milestone represents a specific, measurable deliverable within a smart contract.
// Each milestone has its own value, deadline, and approval mechanism, allowing for
// incremental progress and payment.
type Milestone struct {
	ID          []byte          // A unique identifier for the milestone, typically a hash of its core attributes.
	Title       string          // A concise title for the milestone.
	Description string          // A detailed description of the milestone's requirements and deliverables.
	Value       int             // The payment amount associated with the successful completion of this milestone.
	DueDate     int64           // The deadline for completing the milestone, represented as a Unix timestamp.
	Status      MilestoneStatus // The current status of the milestone (e.g., Active, Completed).
	CreatedAt   int64           // The timestamp of when the milestone was created.
	CompletedAt int64           // The timestamp of when the milestone was marked as completed. Zero if not completed.
	Evidence    []byte          // A cryptographic hash (e.g., IPFS hash) of the evidence or deliverables submitted for this milestone.
	ApprovedBy  []string        // A list of wallet addresses of the parties who have approved the milestone's completion.
}

// Contract represents a legally binding smart contract on the blockchain. It
// encapsulates all the terms, conditions, parties, and milestones that define
// an agreement.
type Contract struct {
	ID             []byte         // A unique identifier for the contract, derived from its immutable core components.
	Title          string         // The official title of the contract.
	Description    string         // A comprehensive description of the contract's purpose and scope.
	CreatedAt      int64          // The timestamp of when the contract was created.
	UpdatedAt      int64          // The timestamp of the last update made to the contract.
	Status         ContractStatus // The current status of the contract in its lifecycle (e.g., Draft, Active).
	Milestones     []*Milestone   // A list of all milestones that constitute the contract's deliverables.
	Parties        []*Party       // The participants involved in the contract, including their roles and signatures.
	Terms          []byte         // A cryptographic hash of the detailed terms and conditions of the contract.
	CreatorAddress string         // The wallet address of the party who created the contract.
	Attachments    [][]byte       // A list of cryptographic hashes for any attached documents (e.g., specifications, legal notices).
}

// Party represents a participant in a smart contract. Each party has a defined
// role and provides a digital signature to signify their agreement to the contract terms.
type Party struct {
	Address   string       // The unique wallet address of the party.
	Role      ContractRole // The role of the party within the contract (e.g., Contractor, Arbitrator).
	PublicKey []byte       // The public key of the party, used for signature verification.
	Signature []byte       // The digital signature of the contract's core data, proving the party's consent.
}

// ContractRole defines the specific role that a party plays within a contract.
type ContractRole string

// Defines the set of possible roles for a party in a contract.
const (
	// RoleContractor is a party responsible for delivering the work or services
	// defined in the contract's milestones.
	RoleContractor ContractRole = "CONTRACTOR"
	// RoleArbitrator is a neutral third party designated to resolve disputes
	// between the other parties.
	RoleArbitrator ContractRole = "ARBITRATOR"
	// RoleCreator is the party who initially drafts and proposes the contract.
	RoleCreator ContractRole = "CREATOR"
)

// ContractCore represents the immutable components of a contract. This structure
// is used to generate the unique contract ID by hashing its contents, ensuring
// that the ID is deterministic and tamper-proof.
type ContractCore struct {
	Title          string
	Description    string
	CreatedAt      int64
	Milestones     []*MilestoneCore
	Parties        []*PartyCore
	Terms          []byte
	CreatorAddress string
	Attachments    [][]byte
}

// PartyCore represents the immutable attributes of a contract party. It is used
// as part of the ContractCore structure for generating the contract ID. The
// signature is excluded because it is generated after the contract ID is created.
type PartyCore struct {
	Address   string
	Role      string
	PublicKey []byte
}

// MilestoneCore represents the immutable attributes of a milestone. This is used
// within the ContractCore to ensure that the fundamental aspects of a milestone
// cannot be changed after the contract is created. Mutable fields like status
// are excluded.
type MilestoneCore struct {
	Title       string
	Description string
	Value       int
	CreatedAt   int64
}

// ErrMilestoneNegativeValue is returned when a milestone has a negative value.
var ErrMilestoneNegativeValue = errors.New("milestone value cannot be negative")

// ErrMilestoneEmptyTitle is returned when a milestone has an empty title.
var ErrMilestoneEmptyTitle = errors.New("milestone title cannot be empty")

// ErrContractEmptyTitle is returned when a contract has an empty title.
var ErrContractEmptyTitle = errors.New("contract title cannot be empty")

// ErrContractEmptyCreator is returned when a contract has an empty creator address.
var ErrContractEmptyCreator = errors.New("contract creator address cannot be empty")

// ErrPartyEmptyAddress is returned when a party has an empty address.
var ErrPartyEmptyAddress = errors.New("party address cannot be empty")

// ErrPartyEmptyRole is returned when a party has an empty role.
var ErrPartyEmptyRole = errors.New("party role cannot be empty")

// Validate checks if a Milestone has valid values.
func (m *Milestone) Validate() error {
	if m.Value < 0 {
		return ErrMilestoneNegativeValue
	}
	if m.Title == "" {
		return ErrMilestoneEmptyTitle
	}
	return nil
}

// Validate checks if a Contract has valid values.
func (c *Contract) Validate() error {
	if c.Title == "" {
		return ErrContractEmptyTitle
	}
	if c.CreatorAddress == "" {
		return ErrContractEmptyCreator
	}
	for _, m := range c.Milestones {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	for _, p := range c.Parties {
		if err := p.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate checks if a Party has valid values.
func (p *Party) Validate() error {
	if p.Address == "" {
		return ErrPartyEmptyAddress
	}
	if p.Role == "" {
		return ErrPartyEmptyRole
	}
	return nil
}
