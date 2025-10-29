// Package handler provides the HTTP request handlers for the Foedus Blockchain's
// RESTful API. These handlers are responsible for processing incoming requests,
// interacting with the blockchain server, and returning appropriate responses.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/codila125/foedus-blockchain/api/server"
	"github.com/codila125/foedus-blockchain/wallet"
)

// Handler encapsulates the server logic and provides methods to handle API requests.
// It acts as a bridge between the HTTP routing layer and the core blockchain server.
type Handler struct {
	server *server.Server
}

// NewHandler creates and returns a new Handler instance. It requires a server
// instance to be provided, which it uses to process blockchain-related operations.
func NewHandler(server *server.Server) *Handler {
	return &Handler{
		server: server,
	}
}

// CreateWallet handles the request to generate a new cryptographic wallet.
// On success, it returns the new wallet's address with an HTTP 201 Created status.
// On failure, it returns an HTTP 500 Internal Server Error.
func (h *Handler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	address, err := h.server.CreateWallet(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"address": address})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ListAddresses handles the request to retrieve all wallet addresses stored on the node.
// It returns a JSON array of addresses with an HTTP 200 OK status.
// On failure, it returns an HTTP 500 Internal Server Error.
func (h *Handler) ListAddresses(w http.ResponseWriter, r *http.Request) {
	addresses, err := h.server.ListAddresses(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string][]string{"addresses": addresses})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// PrintChain handles the request to retrieve the entire blockchain.
// It returns a JSON representation of all blocks in the chain with an HTTP 200 OK status.
// On failure, it returns an HTTP 500 Internal Server Error.
func (h *Handler) PrintChain(w http.ResponseWriter, r *http.Request) {
	blocks := h.server.PrintChain(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(blocks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetBalance handles the request to retrieve the balance of a specific wallet address.
// It validates the address format and returns the balance as a JSON object.
// It returns an HTTP 400 Bad Request for invalid addresses and HTTP 500 for other errors.
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	address := chi.URLParam(r, "address")

	if !wallet.ValidateAddress(address) {
		http.Error(w, "Invalid wallet address", http.StatusBadRequest)
		return
	}

	balance, err := h.server.GetBalance(r.Context(), address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]int{"balance": balance})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreateContract handles the request to create a new smart contract.
// It validates the addresses of the creator and all parties involved.
// On success, it returns the new contract's ID with an HTTP 201 Created status.
func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	creator := chi.URLParam(r, "address")
	if !wallet.ValidateAddress(creator) {
		http.Error(w, "Invalid creator address: "+creator, http.StatusBadRequest)
		return
	}
	var req server.CreateContractReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Creator = creator

	for _, party := range req.Parties {
		if !wallet.ValidateAddress(party.Address) {
			http.Error(w, "Invalid party address: "+party.Address, http.StatusBadRequest)
			return
		}
	}

	contractID, err := h.server.CreateContract(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"contract_id": contractID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetContract handles the request to retrieve the status of a specific smart contract.
// It returns a JSON representation of the contract with an HTTP 200 OK status.
// On failure, it returns an HTTP 500 Internal Server Error.
func (h *Handler) GetContract(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractID")

	contract, err := h.server.ContractStatus(r.Context(), contractID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contract)
}

// ApproveContract handles the request to approve a smart contract.
// It validates the approver's address, updates the contract's state, and
// returns the updated contract as a JSON object with an HTTP 200 OK status.
func (h *Handler) ApproveContract(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractID")
	approver := chi.URLParam(r, "address")
	if !wallet.ValidateAddress(approver) {
		http.Error(w, "Invalid approver address: "+approver, http.StatusBadRequest)
		return
	}

	err := h.server.ApproveContract(r.Context(), contractID, approver)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contract, err := h.server.ContractStatus(r.Context(), contractID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contract)
}

// ApproveMilestone handles the request to approve a milestone within a smart contract.
// It validates the approver's address, processes the provided evidence, and updates
// the milestone's status. It returns the updated contract with an HTTP 200 OK status.
func (h *Handler) ApproveMilestone(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractID")
	milestoneID := chi.URLParam(r, "milestoneID")
	approver := chi.URLParam(r, "address")
	if !wallet.ValidateAddress(approver) {
		http.Error(w, "Invalid approver address: "+approver, http.StatusBadRequest)
		return
	}

	var req struct {
		Evidence string `json:"evidence"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.server.ApproveMilestone(r.Context(), contractID, milestoneID, approver, []byte(req.Evidence))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contract, err := h.server.ContractStatus(r.Context(), contractID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contract)
}
