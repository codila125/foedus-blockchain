package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/codila125/foedus-blockchain/api/server"
)

type Handler struct {
	ctx   context.Context
	server *server.Server
}

func NewHandler(server *server.Server) *Handler {
	return &Handler{
		ctx:    context.Background(),
		server: server,
	}
}

func (h *Handler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	address, err := h.server.CreateWallet(h.ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"address": address})
}

func (h *Handler) ListAddresses(w http.ResponseWriter, r *http.Request) {
	addresses, err := h.server.ListAddresses(h.ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string][]string{"addresses": addresses})
}

func (h *Handler) PrintChain(w http.ResponseWriter, r *http.Request) {
	blocks := h.server.PrintChain(h.ctx)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(blocks)
}