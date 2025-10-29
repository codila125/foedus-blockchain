// Package server implements the core logic for the Foedus Blockchain's API server.
// It handles requests for creating wallets, checking balances, managing smart
// contracts, and interacting with the underlying blockchain and network layers.
package server

import (
	"context"
	"encoding/hex"
	"log"
	"time"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/network"
	"github.com/codila125/foedus-blockchain/wallet"
	"github.com/libp2p/go-libp2p/core/host"
)

// Server encapsulates the main components of the blockchain API server, including
// the port it runs on, the network host for P2P communication, and the blockchain
// instance it serves.
type Server struct {
	port       string
	sourceNode host.Host
	chain      *blockchain.BlockChain
}

// NewServer creates and initializes a new API server instance. It continues an
// existing blockchain from the specified port's data directory and sets up the
// P2P network node.
func NewServer(port string) *Server {
	chain := blockchain.ContinueBlockChain(port)
	sourceNode := network.RunSourceNode(chain, port)
	return &Server{
		port:       port,
		chain:      chain,
		sourceNode: sourceNode,
	}
}

// CreateWallet generates a new wallet, saves it to the node's wallet file, and
// returns the new wallet's address. It ensures that the wallet is persisted
// before confirming its creation.
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

// ListAddresses retrieves and returns all wallet addresses managed by the server's
// node. It loads the wallets from the file and returns a slice of address strings.
func (s *Server) ListAddresses(ctx context.Context) ([]string, error) {
	wallets, err := wallet.CreateWallets(s.port)
	if err != nil {
		return nil, err
	}
	addresses := wallets.GetAllAddresses()

	log.Printf("[SERVER] Listed all wallet addresses")

	return addresses, nil
}

// PrintChain returns a complete, chronologically ordered representation of the
// blockchain, from the most recent block to the genesis block.
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

// GetBalance calculates and returns the total balance for a given wallet address
// by summing up the values of all unspent transaction outputs (UTXOs) owned by that address.
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

// CreateContract facilitates the creation of a new smart contract. It constructs
// the contract with the provided parties, milestones, and terms, and then
// broadcasts it to the network for other nodes to process.
func (s *Server) CreateContract(ctx context.Context, req CreateContractReq) (string, error) {

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

	network.HandleSendContractRequest(s.sourceNode, contract)

	log.Printf("[SERVER] Contract created successfully with ID: %x", contract.ID)

	return hex.EncodeToString(contract.ID), nil
}

// ApproveContract handles the approval of a smart contract. It adds the approver's
// signature to the contract and then broadcasts the updated contract to the network.
func (s *Server) ApproveContract(ctx context.Context, contractID string, approverAddress string) error {
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

	network.HandleSendContractRequest(s.sourceNode, &contract)

	log.Printf("[SERVER] Contract ID %s approved by %s", contractID, approverAddress)
	return nil
}

// ApproveMilestone manages the approval of a contract's milestone. It records the
// approval with the provided evidence, updates the milestone's status, and
// broadcasts the modified contract to the network.
func (s *Server) ApproveMilestone(ctx context.Context, contractID string, milestoneID string, approverAddress string, evidence []byte) error {
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

	network.HandleSendContractRequest(s.sourceNode, updatedContract)

	log.Printf("[SERVER] Milestone ID %s in contract ID %s approved by %s", milestoneID, contractID, approverAddress)
	return nil
}

// ContractStatus retrieves and returns the current state of a smart contract,
// identified by its ID.
func (s *Server) ContractStatus(ctx context.Context, contractID string) (ContractRes, error) {
	log.Printf("[SERVER] Retrieving status for contract ID: %s", contractID)

	contract, err := s.chain.FindContract(contractID)
	if err != nil {
		return ContractRes{}, err
	}

	log.Printf("[SERVER] Retrieved status for contract ID %s: %s", contractID, contract.Status)

	return ContractResponse(contract), nil
}

// Close gracefully shuts down the server's resources, including the blockchain's
// database connection. It uses a context with a timeout to prevent the shutdown
// process from hanging.
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
