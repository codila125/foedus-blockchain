package server

import (
	"log"
	"context"
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
	wallets, _ := wallet.CreateWallets(s.port)
	address := wallets.AddWallet()
	wallets.SaveFile(s.port)
	
	log.Printf("[SERVER] New wallet created successfully with address: %s\n", address)
	return address, nil
}

func (s *Server) ListAddresses(ctx context.Context) ([]string, error) {
	wallets, _ := wallet.CreateWallets(s.port)
	addresses := wallets.GetAllAddresses()

	log.Printf("[SERVER] Listing all wallet addresses:")

	return addresses, nil
}