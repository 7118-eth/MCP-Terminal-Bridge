package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bioharz/mcp-terminal-tester/internal/mcp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		cancel()
	}()

	// Create and configure MCP server
	srv, err := mcp.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Get port from environment or use default
	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting MCP Terminal Tester on port %s", port)

	// Run the server
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}