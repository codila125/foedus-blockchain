// Package server implements the gRPC server for the Foedus Blockchain.
package server

import (
	"context"
	"encoding/hex"
	"log"
	"time"

	"github.com/codila125/foedus-blockchain/network"
	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/wallet"
	"github.com/libp2p/go-libp2p/core/host"
)

type Server struct {
	port      string
	sourceNode host.Host
	chain     *blockchain.BlockChain
}

func NewServer(port string) *Server {
	chain := blockchain.ContinueBlockChain(port)
	sourceNode := network.RunSourceNode(chain)
	return &Server{
		port:      port,
		chain:     chain,
		sourceNode: sourceNode,
	}
}

func (s *Server) CreateWallet(ctx context.Context) (string, error) {
	wallets, _ := wallet.CreateWallets(s.port)
	address := wallets.AddWallet()
	err := wallets.SaveFile(s.port)
	if err != nil {
		log.Printf("[SERVER] Failed to save wallet: %v\n", err)
		return "", err
	}

	log.Printf("[SERVER] New wallet created successfully with address: %s\n", address)
	return address, nil
}

func (s *Server) ListAddresses(ctx context.Context) ([]string, error) {
	wallets, err := wallet.CreateWallets(s.port)
	if err != nil {
		return nil, err
	}
	addresses := wallets.GetAllAddresses()

	log.Printf("[SERVER] Listed all wallet addresses")

	return addresses, nil
}

func (s *Server) PrintChain(ctx context.Context) []*BlockRes {
	iterator := s.chain.Iterator()

	var blocks []*BlockRes
	for {
		block := iterator.Next()
		blocks = append(blocks, BlockResponse(block))
		if len(block.PrevHash) == 0 {
			break
		}
	}
	log.Printf("[SERVER] Served the blockchain successfully")

	return blocks
}

func (s *Server) GetBalance(ctx context.Context, address string) (int, error) {
	UTXOSet := blockchain.UTXOSet{Blockchain: s.chain}

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	log.Printf("[SERVER] Retrieved balance for address %s: %d", address, balance)

	return balance, nil
}

func (s *Server) CreateContract(ctx context.Context, req CreateContractReq) (string, error) {
	// Implementation for creating a contract goes here
	log.Printf("[SERVER] Creating contract: %s", req.Title)

	wallets, err := wallet.CreateWallets(s.port)
	if err != nil {
		return "", err
	}

	creatorWallet, err := wallets.GetWallet(req.Creator)
	if err != nil {
		return "", err
	}

	party := []*blockchain.Party{}
	for _, partyReq := range req.Parties {
		partyWallet, err := wallets.GetWallet(partyReq.Address)
		if err != nil {
			return "", err
		}
		party = append(party, &blockchain.Party{
			Address:   partyReq.Address,
			Role:      partyReq.Role,
			PublicKey: partyWallet.PublicKey,
			Signature: []byte{},
		})
	}

	milestones := []*blockchain.Milestone{}
	for _, milestoneReq := range req.Milestones {
		milestone := &blockchain.Milestone{
			Title:       milestoneReq.Title,
			Description: milestoneReq.Description,
			Value:       milestoneReq.Value,
			DueDate:     milestoneReq.DueDate,
			Status:      blockchain.MilestoneActive,
			CreatedAt:   time.Now().Unix(),
			CompletedAt: 0,
			Evidence:    []byte{},
			ApprovedBy:  []string{},
		}
		milestone.ID = milestone.HashMilestones()
		milestones = append(milestones, milestone)
	}

	attachments := [][]byte{}
	for _, att := range req.Attachments {
		attachments = append(attachments, []byte(att))
	}

	contract := blockchain.CreateContract(req.Title, req.Description, &creatorWallet, milestones, party, []byte(req.Terms), attachments)
	cts := []*blockchain.Contract{contract} // Include the contract transaction in the new block
	block := s.chain.MineBlock(nil, cts)    // Mine a new block with the contract transaction

	log.Printf("[SERVER] Contract created and included in block %x", block.Hash)

	return hex.EncodeToString(contract.ID), nil
}

func (s *Server) ApproveContract(ctx context.Context, contractID string, approverAddress string) error {
	// Implementation for approving a contract goes here
	log.Printf("[SERVER] Approving contract ID: %s by approver: %s", contractID, approverAddress)

	wallets, err := wallet.CreateWallets(s.port)
	if err != nil {
		return err
	}

	approverWallet, err := wallets.GetWallet(approverAddress)
	if err != nil {
		return err
	}
	contract, err := s.chain.FindContract(contractID)
	if err != nil {
		return err
	}

	err = contract.ApproveContract(&approverWallet)
	if err != nil {
		return err
	}

	cts := []*blockchain.Contract{&contract} // Include the updated contract transaction in the new block
	block := s.chain.MineBlock(nil, cts)     // Mine a new block with the updated contract transaction

	log.Printf("[SERVER] Contract ID %s approved by %s and included in block %x", contractID, approverAddress, block.Hash)
	return nil
}

func (s *Server) ApproveMilestone(ctx context.Context, contractID string, milestoneID string, approverAddress string, evidence []byte) error {
	// Implementation for approving a milestone goes here
	log.Printf("[SERVER] Approving milestone ID: %s in contract ID: %s by approver: %s", milestoneID, contractID, approverAddress)

	wallets, err := wallet.CreateWallets(s.port)
	if err != nil {
		return err
	}

	approverWallet, err := wallets.GetWallet(approverAddress)
	if err != nil {
		return err
	}
	contract, err := s.chain.FindContract(contractID)
	if err != nil {
		return err
	}

	milestoneIDBytes, err := hex.DecodeString(milestoneID)
	if err != nil {
		return err
	}

	updatedContract, err := contract.ApproveMilestone(&approverWallet, milestoneIDBytes, evidence)
	if err != nil {
		return err
	}

	cts := []*blockchain.Contract{updatedContract} // Include the updated contract transaction in the new block
	block := s.chain.MineBlock(nil, cts)           // Mine a new block with the updated contract transaction

	log.Printf("[SERVER] Milestone ID %s in contract ID %s approved by %s and included in block %x", milestoneID, contractID, approverAddress, block.Hash)
	return nil
}

func (s *Server) ContractStatus(ctx context.Context, contractID string) (ContractRes, error) {
	// Implementation for retrieving contract status goes here
	log.Printf("[SERVER] Retrieving status for contract ID: %s", contractID)

	contract, err := s.chain.FindContract(contractID)
	if err != nil {
		return ContractRes{}, err
	}

	log.Printf("[SERVER] Retrieved status for contract ID %s: %s", contractID, contract.Status)

	return ContractResponse(contract), nil
}

// Close gracefully closes the database and associated resources
func (s *Server) Close(ctx context.Context) error {
	log.Println("[SERVER] Closing server resources...")

	if s.chain == nil {
		log.Println("[SERVER] Blockchain is nil, skipping close")
		return nil
	}

	if s.chain.Database == nil {
		log.Println("[SERVER] Database is nil, skipping close")
		return nil
	}

	// Close database with timeout
	dbCloseDone := make(chan error, 1)
	go func() {
		dbCloseDone <- s.chain.Database.Close()
	}()

	select {
	case err := <-dbCloseDone:
		if err != nil {
			log.Printf("[SERVER] Error closing database: %v", err)
			return err
		}
		log.Println("[SERVER] Database closed successfully")
	case <-ctx.Done():
		log.Println("[SERVER] Database closure timeout exceeded")
		return ctx.Err()
	}

	return nil
}

// GetChain returns the blockchain instance
func (s *Server) GetChain() *blockchain.BlockChain {
	return s.chain
}

