package blockchain

type ContractStatus string

const (
	ContractDraft     ContractStatus = "DRAFT"
	ContractActive    ContractStatus = "ACTIVE"
	ContractCompleted ContractStatus = "COMPLETED"
	ContractDisputed  ContractStatus = "DISPUTED"
	ContractCancelled ContractStatus = "CANCELLED"
)

type MilestoneStatus string

const (
	MilestonePending   MilestoneStatus = "ACTIVE"
	MilestoneCompleted MilestoneStatus = "COMPLETED"
	MilestoneRejected  MilestoneStatus = "DISPUTED"
	MilestoneCancelled MilestoneStatus = "CANCELLED"
)

// Milestone represents a deliverable within a contract
type Milestone struct {
	ID          []byte          // Unique identifier for the milestone
	ContractID  []byte          // Reference to parent contract
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
	TermsHash      []byte         // Hash of contract terms
	DisputeHandler string         // Address of dispute arbitrator
	CreatorAddress string         // Address of contract creator
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
	RoleClient     ContractRole = "CLIENT"
	RoleContractor ContractRole = "CONTRACTOR"
	RoleArbitrator ContractRole = "ARBITRATOR"
	RoleCreator    ContractRole = "CREATOR"
)

type ContractOpType string

const (
	OpCreateContract    ContractOpType = "CREATE_CONTRACT"
	OpApproveContract   ContractOpType = "APPROVE_CONTRACT"
	OpCompleteMilestone ContractOpType = "COMPLETE_MILESTONE"
	OpApproveMilestone  ContractOpType = "APPROVE_MILESTONE"
	OpRejectMilestone   ContractOpType = "DISPUTE_MILESTONE"
	OpResolveDispute    ContractOpType = "RESOLVE_DISPUTE"
	OpCancelContract    ContractOpType = "CANCEL_CONTRACT"
)

type ContractOperation struct {
	ID          []byte         // Operation ID
	Type        ContractOpType // Type of operation
	ContractID  []byte         // Related contract ID
	MilestoneID []byte         // Related milestone ID (if applicable)
	Actor       string         // Address executing the operation
	Timestamp   int64          // When operation occurred
	Signature   []byte         // Digital signature
	Data        []byte         // Additional operation data
}
