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
	CreatedAt time.Time                  `json:"created_at"`
	UpdatedAt time.Time                  `json:"updated_at"`
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
