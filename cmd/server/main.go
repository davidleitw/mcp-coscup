package main

import (
	"log"
	"os"

	"coscup-mcp-server/mcp"
)

func main() {
	log.Println("🎯 Initializing COSCUP MCP Server...")

	// Create new COSCUP server instance
	server := mcp.NewCOSCUPServer()

	// Start the server (this will block)
	if err := server.Start(); err != nil {
		log.Printf("❌ Server error: %v", err)
		os.Exit(1)
	}
}