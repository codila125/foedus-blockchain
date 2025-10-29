// Package server provides utility functions for converting core blockchain data
// structures into JSON-serializable response formats. These functions ensure that
// the API responses are consistent, human-readable, and suitable for client-side
// consumption.
package server

import (
	"encoding/hex"
	"time"

	"github.com/codila125/foedus-blockchain/blockchain"
)

// BlockResponse transforms a blockchain.Block into a BlockRes, which is a
// simplified, JSON-friendly representation of a block. It encodes binary data,
// such as hashes, into hexadecimal strings.
func BlockResponse(block *blockchain.Block) *BlockRes {
	return &BlockRes{
		Hash:         hex.EncodeToString(block.Hash),
		PrevHash:     hex.EncodeToString(block.PrevHash),
		Contract:     BlockContractResponse(block.Contracts),
		Transactions: BlockTransactionResponse(block.Transactions),
	}
}

// BlockContractResponse converts a slice of blockchain.Contract pointers into a
// slice of BlockContractRes. This function is used to prepare contract data for
// inclusion in a block response.
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

// BlockTransactionResponse converts a slice of blockchain.Transaction pointers
// into a slice of BlockTransactionRes, making them suitable for API responses.
func BlockTransactionResponse(transactions []*blockchain.Transaction) []BlockTransactionRes {
	var transactionRes []BlockTransactionRes
	for _, tx := range transactions {
		transactionRes = append(transactionRes, BlockTransactionRes{
			ID:      hex.EncodeToString(tx.ID),
			Inputs:  BlockTransactionInputResponse(tx.Inputs),
			Outputs: BlockTransactionOutputResponse(tx.Outputs),
		})
	}
	return transactionRes
}

// BlockTransactionInputResponse transforms a slice of blockchain.TxInput into a
// slice of BlockTransactionInputRes for JSON serialization.
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

// BlockTransactionOutputResponse converts a slice of blockchain.TxOutput into a
// slice of BlockTransactionOutputRes, preparing them for API output.
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

// ContractResponse converts a blockchain.Contract into a detailed ContractRes
// object for client-side consumption. It formats timestamps, encodes binary data,
// and prepares nested structures like milestones and parties.
func ContractResponse(contract blockchain.Contract) ContractRes {
	return ContractRes{
		ID:          hex.EncodeToString(contract.ID),
		Title:       contract.Title,
		Description: contract.Description,
		Creator:     contract.CreatorAddress,
		Milestones:  ContractMilestoneResponse(contract.Milestones),
		Parties:     ContractPartyResponse(contract.Parties),
		Terms:       string(contract.Terms),
		Attachments: ContractAttachmentResponse(contract.Attachments),
		Status:      contract.Status,
		CreatedAt:   time.Unix(contract.CreatedAt, 0),
		UpdatedAt:   time.Unix(contract.UpdatedAt, 0),
	}
}

// ContractMilestoneResponse transforms a slice of blockchain.Milestone pointers
// into a slice of ContractMilestoneRes, making them ready for JSON output.
func ContractMilestoneResponse(milestones []*blockchain.Milestone) []ContractMilestoneRes {
	var milestoneRes []ContractMilestoneRes
	for _, milestone := range milestones {
		milestoneRes = append(milestoneRes, ContractMilestoneRes{
			ID:          hex.EncodeToString(milestone.ID),
			Title:       milestone.Title,
			Description: milestone.Description,
			Value:       milestone.Value,
			DueDate:     time.Unix(milestone.DueDate, 0),
			Status:      milestone.Status,
			CreatedAt:   time.Unix(milestone.CreatedAt, 0),
			CompletedAt: time.Unix(milestone.CompletedAt, 0),
			Evidence:    string(milestone.Evidence),
			ApprovedBy:  milestone.ApprovedBy,
		})
	}
	return milestoneRes
}

// ContractPartyResponse converts a slice of blockchain.Party pointers into a
// slice of ContractPartyRes for API responses.
func ContractPartyResponse(parties []*blockchain.Party) []ContractPartyRes {
	var partyRes []ContractPartyRes
	for _, party := range parties {
		partyRes = append(partyRes, ContractPartyRes{
			Address:   party.Address,
			Role:      party.Role,
			PublicKey: hex.EncodeToString(party.PublicKey),
		})
	}
	return partyRes
}

// ContractAttachmentResponse converts a slice of byte slices, representing
// attachments, into a slice of strings for easier handling in JSON.
func ContractAttachmentResponse(attachments [][]byte) []string {
	var attachmentRes []string
	for _, attachment := range attachments {
		attachmentRes = append(attachmentRes, string(attachment))
	}
	return attachmentRes
}
