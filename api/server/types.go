package server

import (
	"time"

	"github.com/codila125/foedus-blockchain/blockchain"
)

type BlockRes struct {
	PrevHash     string                `json:"prev_hash"`
	Hash         string                `json:"hash"`
	Contract     []BlockContractRes    `json:"contract"`
	Transactions []BlockTransactionRes `json:"transactions"`
}

type BlockContractRes struct {
	ID        string                    `json:"id"`
	Title     string                    `json:"title"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
	Status    blockchain.ContractStatus `json:"status"`
}

type BlockTransactionRes struct {
	ID      string                      `json:"id"`
	Inputs  []BlockTransactionInputRes  `json:"inputs"`
	Outputs []BlockTransactionOutputRes `json:"outputs"`
}

type BlockTransactionInputRes struct {
	From string `json:"from"`
	Out  int    `json:"out"`
}

type BlockTransactionOutputRes struct {
	To    string `json:"to"`
	Value int    `json:"value"`
}

type CreateContractReq struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Creator     string            `json:"creator"`
	Milestones  []AddMilestoneReq `json:"milestones"`
	Parties     []AddPartyReq     `json:"parties"`
	Terms       string            `json:"terms"`
	Attachments []string          `json:"attachments"`
}

type AddMilestoneReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Value       int    `json:"value"`
	DueDate     int64  `json:"due_date"`
}

type AddPartyReq struct {
	Address string                  `json:"address"`
	Role    blockchain.ContractRole `json:"role"`
}

type ContractRes struct {
	ID          string                    `json:"id"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Creator     string                    `json:"creator"`
	Milestones  []ContractMilestoneRes    `json:"milestones"`
	Parties     []ContractPartyRes        `json:"parties"`
	Terms       string                    `json:"terms"`
	Attachments []string                 `json:"attachments"`
	Status      blockchain.ContractStatus `json:"status"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

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

type ContractPartyRes struct {
	Address   string                  `json:"address"`
	Role      blockchain.ContractRole `json:"role"`
	PublicKey string                  `json:"public_key"`
}
