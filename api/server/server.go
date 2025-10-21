package server

import (
	"context"
	"encoding/hex"
	"log"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/wallet"
	"github.com/vrecan/death/v3"
)

type Server struct {
	port  string
	chain *blockchain.BlockChain
}

func NewServer(port string) *Server {
	chain := blockchain.ContinueBlockChain(port)
	go CloseDB(chain)
	return &Server{
		port:  port,
		chain: chain,
	}
}

func (s *Server) CreateWallet(ctx context.Context) (string, error) {
	wallets, err := wallet.CreateWallets(s.port)
	if err != nil {
		return "", err
	}
	address := wallets.AddWallet()
	wallets.SaveFile(s.port)
	
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

func (s *Server) PrintChain(ctx context.Context) ([]*BlockRes) {
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

	creatorWallet := wallets.GetWallet(req.Creator)
	partyWallet := wallets.GetWallet(req.Parties[0].Address)

	party := &blockchain.Party{
		Address:   req.Parties[0].Address,
		Role:      req.Parties[0].Role,
		PublicKey: partyWallet.PublicKey,
		Signature: []byte{},
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
			Evidence:	[]byte{},
			ApprovedBy: []string{},
			DisputedBy: []string{},
		}
		milestone.ID = milestone.HashMilestones()
		milestones = append(milestones, milestone)
	}

	contract := blockchain.CreateContract(req.Title, req.Description, &creatorWallet, milestones, []*blockchain.Party{party}, []byte(req.Terms), req.Attachments)
	cts := []*blockchain.Contract{contract}  // Include the contract transaction in the new block
	block := s.chain.MineBlock(nil, cts) // Mine a new block with the contract transaction

	log.Printf("[SERVER] Contract created and included in block %x", block.Hash)

	return hex.EncodeToString(contract.ID), nil
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

func CloseDB(chain *blockchain.BlockChain) {
	/*
		If a termination signal is received, this function ensures that the
		blockchain database is properly closed before exiting the program.
		Termination signals like SIGINT (Ctrl+C) and SIGTERM are handled.
	*/
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)       // Exit with a non-zero status to indicate termination
		defer runtime.Goexit() // Ensure all goroutines are terminated
		log.Printf("[SERVER] Shutting down node, closing database...")
		chain.Database.Close() // Close the blockchain database
		log.Printf("[SERVER] Database closed successfully")
	})
}