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
	shutdownTimeout = 30 * time.Second
)

func main() {
	cliCmd := cli.CommandLine{}
	if len(os.Args) > 1 {
		cliCmd.Run()
		return
	}

	// Start the API server
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		log.Fatal("[SERVER] NODE_ID environment variable is not set")
	}

	serve := server.NewServer(nodeID)
	handle := handler.NewHandler(serve)
	api.RegisterRoutes(handle)
	srv := api.StartServer(":" + nodeID)

	// Start the HTTP server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("[SERVER] Server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Set up graceful shutdown
	gracefulShutdown(srv, serve, serverErrors)
}

// gracefulShutdown handles signal interrupts and orchestrates resource cleanup
func gracefulShutdown(srv *http.Server, appServer *server.Server, serverErrors chan error) {
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	select {
	case err := <-serverErrors:
		log.Fatalf("[SERVER] Server error: %v", err)
	case sig := <-shutdownChan:
		log.Printf("[SERVER] Received shutdown signal: %v", sig)
		shutdown(srv, appServer)
	}
}

// shutdown orchestrates the graceful shutdown sequence
func shutdown(srv *http.Server, appServer *server.Server) {
	var wg sync.WaitGroup

	// Create a context with timeout for the entire shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Println("[SERVER] Starting shutdown sequence...")

	// Stop accepting new connections and wait for existing requests to complete
	wg.Go(func() {
		log.Println("[SERVER] Shutting down HTTP server...")
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("[SERVER] HTTP server shutdown error: %v", err)
		} else {
			log.Println("[SERVER] HTTP server shut down successfully")
		}
	})

	// Close database and other resources
	wg.Go(func() {
		if err := appServer.Close(ctx); err != nil {
			log.Printf("[SERVER] Error during resource cleanup: %v", err)
		}
	})

	// Wait for all shutdown tasks to complete or timeout
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

