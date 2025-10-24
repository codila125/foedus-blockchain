// Package handler implements HTTP handlers for the Foedus Blockchain API.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/codila125/foedus-blockchain/api/server"
	"github.com/codila125/foedus-blockchain/wallet"
)

type Handler struct {
	server *server.Server
}

func NewHandler(server *server.Server) *Handler {
	return &Handler{
		server: server,
	}
}

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
