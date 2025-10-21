package server

import (
	"context"
	"log"

	"github.com/codila125/foedus-blockchain/blockchain"
	"github.com/codila125/foedus-blockchain/wallet"
)

type Server struct {
	port string
}

func NewServer(port string) *Server {
	return &Server{
		port: port,
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

	chain := blockchain.ContinueBlockChain(s.port)
	defer chain.Database.Close()
	iterator := chain.Iterator()

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
	
	chain := blockchain.ContinueBlockChain(s.port)
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	defer chain.Database.Close()

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