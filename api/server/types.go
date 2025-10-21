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
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	Creator        string            `json:"creator"`
	Milestones     []AddMilestoneReq `json:"milestones"`
	Parties        []AddPartyReq     `json:"parties"`
	Terms          string            `json:"terms"`
	DisputeHandler string            `json:"dispute_handler"`
	Attachments    [][]byte          `json:"attachments"`
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
