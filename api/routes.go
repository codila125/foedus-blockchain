// Package api configures and manages the HTTP routes for the Foedus blockchain's
// API server. It defines the endpoints for interacting with the blockchain,
// such as creating wallets, checking balances, and managing smart contracts.
package api

import (
	"log"
	"net/http"

	"github.com/codila125/foedus-blockchain/api/handler"
	"github.com/go-chi/chi/v5"
)

var r *chi.Mux

// RegisterRoutes initializes the API router and maps the HTTP endpoints to their
// corresponding handler functions. This function organizes all the blockchain-related
// routes under the "/blockchain" prefix and returns the configured router.
func RegisterRoutes(handler *handler.Handler) *chi.Mux {
	r = chi.NewRouter()

	r.Route("/blockchain", func(r chi.Router) {
		// GET endpoints for read-only operations
		r.Get("/createwallet", handler.CreateWallet)
		r.Get("/listaddresses", handler.ListAddresses)
		r.Get("/printchain", handler.PrintChain)
		r.Get("/getbalance/{address}", handler.GetBalance)
		r.Get("/getcontract/{contractID}", handler.GetContract)

		// POST endpoints for state-changing operations
		r.Post("/createcontract/{address}", handler.CreateContract)
		r.Post("/approvecontract", handler.ApproveContract)
		r.Post("/approvemilestone", handler.ApproveMilestone)
		r.Post("/cancelcontract", handler.CancelContract)
	})
	return r
}

// StartServer creates and configures an HTTP server to listen on the specified
// port. It assigns the registered routes to the server's handler and returns
// the server instance, ready to be started.
func StartServer(port string) *http.Server {
	log.Printf("[SERVER] Starting server node on PORT%s\n", port)
	return &http.Server{
		Addr:    port,
		Handler: r,
	}
}
