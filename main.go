package main

import (
	"github.com/codila125/foedus-blockchain/api"
	"github.com/codila125/foedus-blockchain/api/handler"
	"github.com/codila125/foedus-blockchain/api/server"
	"github.com/codila125/foedus-blockchain/cli"
	"os"
)

func main() {
	defer os.Exit(0)
	cli := cli.CommandLine{}
	if len(os.Args) > 1 {
		cli.Run()
	} else {
		// Start the API server
		nodeID := os.Getenv("NODE_ID")
		serve := server.NewServer(nodeID)
		handle := handler.NewHandler(serve)
		api.RegisterRoutes(handle)
		api.Start(":" + nodeID)
	}
}
