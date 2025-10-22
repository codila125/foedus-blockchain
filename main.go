package main

import (
	"log"
	"os"

	"github.com/codila125/foedus-blockchain/api"
	"github.com/codila125/foedus-blockchain/api/handler"
	"github.com/codila125/foedus-blockchain/api/server"
	"github.com/codila125/foedus-blockchain/cli"
)

func main() {
	defer os.Exit(0)
	cli := cli.CommandLine{}
	if len(os.Args) > 1 {
		cli.Run()
	} else {
		// Start the API server
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			log.Fatal("[SERVER] NODE_ID environment variable is not set")
		}
		serve := server.NewServer(nodeID)
		handle := handler.NewHandler(serve)
		api.RegisterRoutes(handle)
		err := api.Start(":" + nodeID)
		if err != nil {
			log.Fatalf("[SERVER] Failed to start API server: %v", err)
		}
	}
}
