package api

import (
	"log"
	"net/http"

	"github.com/codila125/foedus-blockchain/api/handler"
	"github.com/go-chi/chi/v5"
)

var r *chi.Mux

func RegisterRoutes(handler *handler.Handler) *chi.Mux {
	r = chi.NewRouter()

	r.Route("/blockchain", func(r chi.Router) {
		r.Post("/createwallet", handler.CreateWallet)
		r.Post("/listaddresses", handler.ListAddresses)
	})
	return r
}

func Start(port string) error {
	log.Printf("[SERVER] Starting server node in PORT%s\n", port)
	return http.ListenAndServe(port, r)
}