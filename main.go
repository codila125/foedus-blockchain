// Package main is the entry point for the Foedus blockchain application.
// It initializes and runs the application in one of two modes:
// as a command-line interface (CLI) for direct interaction or as an API server
// for network-based operations. The mode is determined by the presence of
// command-line arguments.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/codila125/foedus-blockchain/api"
	"github.com/codila125/foedus-blockchain/api/handler"
	"github.com/codila125/foedus-blockchain/api/server"
	"github.com/codila125/foedus-blockchain/cli"
)

const (
	// shutdownTimeout defines the maximum duration to wait for a graceful shutdown
	// before forcing the application to exit. This ensures that the server has
	// enough time to close connections and release resources properly.
	shutdownTimeout = 30 * time.Second
)

// main serves as the primary entry point for the application.
// If command-line arguments are provided, it launches the application in CLI mode
// to execute specific commands. Otherwise, it starts an API server, which requires
// the NODE_ID environment variable to be set for node identification within the network.
func main() {
	cliCmd := cli.CommandLine{}
	if len(os.Args) > 1 {
		cliCmd.Run()
		return
	}

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		log.Fatal("[SERVER] NODE_ID environment variable is not set")
	}

	serve := server.NewServer(nodeID)
	handle := handler.NewHandler(serve)
	api.RegisterRoutes(handle)
	srv := api.StartServer(":" + nodeID)

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("[SERVER] Server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	gracefulShutdown(srv, serve, serverErrors)
}

// gracefulShutdown manages the server's shutdown process in response to operating
// system signals. It listens for interrupt signals (SIGINT, SIGTERM) and initiates
// a clean shutdown, ensuring that all active processes are terminated gracefully.
// It also handles any fatal server errors that may occur during runtime.
func gracefulShutdown(srv *http.Server, appServer *server.Server, serverErrors chan error) {
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	select {
	case err := <-serverErrors:
		log.Fatalf("[SERVER] Fatal server error: %v", err)
	case sig := <-shutdownChan:
		log.Printf("[SERVER] Received signal: %v. Initiating graceful shutdown...", sig)
		shutdown(srv, appServer)
	}
}

// shutdown orchestrates the graceful termination of server resources within a
// specified timeout. It concurrently shuts down the HTTP server and closes
// underlying application resources, such as database connections. This function
// ensures that all components are properly cleaned up before the application exits.
func shutdown(srv *http.Server, appServer *server.Server) {
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Println("[SERVER] Starting shutdown sequence...")

	// Use a WaitGroup to manage concurrent shutdown operations.
	wg.Go(func() {
		log.Println("[SERVER] Shutting down HTTP server...")
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("[SERVER] HTTP server shutdown error: %v", err)
		} else {
			log.Println("[SERVER] HTTP server shut down successfully")
		}
	})

	// Concurrently close application-level resources.
	wg.Go(func() {
		if err := appServer.Close(ctx); err != nil {
			log.Printf("[SERVER] Error during resource cleanup: %v", err)
		}
	})

	// Wait for all shutdown tasks to complete or for the timeout to be reached.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("[SERVER] Shutdown completed successfully")
		os.Exit(0)
	case <-ctx.Done():
		log.Println("[SERVER] Shutdown timeout exceeded, forcing exit")
		os.Exit(1)
	}
}
