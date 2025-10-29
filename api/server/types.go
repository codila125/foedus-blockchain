// Package server defines the data structures used for API requests and responses
// in the Foedus Blockchain. These types ensure that data is consistently
// structured and easily serializable to and from JSON.
package server

import (
	"time"

	"github.com/codila125/foedus-blockchain/blockchain"
)

// BlockRes is the JSON-serializable representation of a block for API responses.
// It provides a structured format for block data that is easy for clients to parse.
type BlockRes struct {
	PrevHash     string                `json:"prev_hash"`
	Hash         string                `json:"hash"`
	Contract     []BlockContractRes    `json:"contract"`
	Transactions []BlockTransactionRes `json:"transactions"`
}

// BlockContractRes represents a contract as it appears within a block in an API response.
// It includes essential details like the contract's ID, title, and status.
type BlockContractRes struct {
	ID        string                    `json:"id"`
	Title     string                    `json:"title"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
	Status    blockchain.ContractStatus `json:"status"`
}

// BlockTransactionRes is the JSON-serializable format for a transaction within a block.
// It includes the transaction's ID, inputs, and outputs.
type BlockTransactionRes struct {
	ID      string                      `json:"id"`
	Inputs  []BlockTransactionInputRes  `json:"inputs"`
	Outputs []BlockTransactionOutputRes `json:"outputs"`
}

// BlockTransactionInputRes represents a transaction input in a JSON-friendly format.
// It specifies the source transaction and output index.
type BlockTransactionInputRes struct {
	From string `json:"from"`
	Out  int    `json:"out"`
}

// BlockTransactionOutputRes represents a transaction output in a JSON-friendly format.
// It includes the recipient's public key hash and the value being transferred.
type BlockTransactionOutputRes struct {
	To    string `json:"to"`
	Value int    `json:"value"`
}

// CreateContractReq defines the structure of a request to create a new smart contract.
// It includes all the necessary information, such as title, description, parties, and milestones.
type CreateContractReq struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Creator     string            `json:"creator"`
	Milestones  []AddMilestoneReq `json:"milestones"`
	Parties     []AddPartyReq     `json:"parties"`
	Terms       string            `json:"terms"`
	Attachments []string          `json:"attachments"`
}

// AddMilestoneReq is the request structure for adding a new milestone to a contract.
// It contains the milestone's title, description, value, and due date.
type AddMilestoneReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Value       int    `json:"value"`
	DueDate     int64  `json:"due_date"`
}

// AddPartyReq defines the request structure for adding a new party to a contract.
// It includes the party's address and their role in the contract.
type AddPartyReq struct {
	Address string                  `json:"address"`
	Role    blockchain.ContractRole `json:"role"`
}

// ContractRes is the detailed JSON-serializable representation of a smart contract
// for API responses. It provides a comprehensive view of the contract's state.
type ContractRes struct {
	ID          string                    `json:"id"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Creator     string                    `json:"creator"`
	Milestones  []ContractMilestoneRes    `json:"milestones"`
	Parties     []ContractPartyRes        `json:"parties"`
	Terms       string                    `json:"terms"`
	Attachments []string                  `json:"attachments"`
	Status      blockchain.ContractStatus `json:"status"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

// ContractMilestoneRes is the JSON-serializable format for a contract milestone.
// It includes all details of the milestone, such as its status, due date, and evidence.
type ContractMilestoneRes struct {
	ID          string                     `json:"id"`
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Value       int                        `json:"value"`
	DueDate     time.Time                  `json:"due_date"`
	Status      blockchain.MilestoneStatus `json:"status"`
	CreatedAt   time.Time                  `json:"created_at"`
	Evidence    string                     `json:"evidence"`
	ApprovedBy  []string                   `json:"approved_by"`
	CompletedAt time.Time                  `json:"completed_at"`
}

// ContractPartyRes represents a party to a contract in a JSON-friendly format.
// It includes the party's address, role, and public key.
type ContractPartyRes struct {
	Address   string                  `json:"address"`
	Role      blockchain.ContractRole `json:"role"`
	PublicKey string                  `json:"public_key"`
}
