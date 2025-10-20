package server

import (
	"encoding/hex"
	"time"
	"github.com/codila125/foedus-blockchain/blockchain"
)
	
func BlockResponse(block *blockchain.Block) *BlockRes {
	return &BlockRes{
		Hash:         hex.EncodeToString(block.Hash),
		PrevHash:     hex.EncodeToString(block.PrevHash),
		Contract:     BlockContractResponse(block.Contracts),
		Transactions: BlockTransactionResponse(block.Transactions),
	}
}

func BlockContractResponse(contracts []*blockchain.Contract) []BlockContractRes {
	var contractRes []BlockContractRes
	for _, contract := range contracts {
		contractRes = append(contractRes, BlockContractRes{
			ID:        hex.EncodeToString(contract.ID),
			Title:     contract.Title,
			CreatedAt: time.Unix(contract.CreatedAt, 0),
			UpdatedAt: time.Unix(contract.UpdatedAt, 0),
			Status:    contract.Status,
		})
	}
	return contractRes
}

func BlockTransactionResponse(transactions []*blockchain.Transaction) []BlockTransactionRes {
	var transactionRes []BlockTransactionRes
	for _, tx := range transactions {
		transactionRes = append(transactionRes, BlockTransactionRes{
			ID:        hex.EncodeToString(tx.ID),
			Inputs:    BlockTransactionInputResponse(tx.Inputs),
			Outputs:   BlockTransactionOutputResponse(tx.Outputs),
		})
	}
	return transactionRes
}

func BlockTransactionInputResponse(inputs []blockchain.TxInput) []BlockTransactionInputRes {
	var inputRes []BlockTransactionInputRes
	for _, input := range inputs {
		inputRes = append(inputRes, BlockTransactionInputRes{
			From: hex.EncodeToString(input.ID),
			Out:  input.Out,
		})
	}
	return inputRes
}

func BlockTransactionOutputResponse(outputs []blockchain.TxOutput) []BlockTransactionOutputRes {
	var outputRes []BlockTransactionOutputRes
	for _, output := range outputs {
		outputRes = append(outputRes, BlockTransactionOutputRes{
			To:    hex.EncodeToString(output.PubKeyHash),
			Value: output.Value,
		})
	}
	return outputRes
}
