package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bioharz/mcp-terminal-tester/internal/mcp"
	"github.com/bioharz/mcp-terminal-tester/internal/utils"
)

func main() {
	// Initialize logger first
	utils.InitLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Shutting down server...")
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

	slog.Info("Starting MCP Terminal Tester", slog.String("mode", "stdio"))

	// Run the server
	if err := srv.Run(ctx); err != nil {
		slog.Error("Server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}