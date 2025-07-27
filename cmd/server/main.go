package main

import (
	"flag"
	"log"
	"os"

	"mcp-coscup/mcp"
)

func main() {
	log.Println("Initializing COSCUP MCP Server...")

	// Parse command line flags
	mode := flag.String("mode", "stdio", "Server mode: stdio or http")
	flag.Parse()

	// Check environment variable if flag not explicitly set
	if *mode == "stdio" {
		if envMode := os.Getenv("MCP_MODE"); envMode != "" {
			*mode = envMode
		}
	}

	log.Printf("Starting server in %s mode", *mode)

	// Create new COSCUP server instance
	server := mcp.NewCOSCUPServer()

	var err error
	switch *mode {
	case "http":
		err = server.StartHTTP()
	case "stdio":
		err = server.Start()
	default:
		log.Printf("Unknown mode: %s. Supported modes: stdio, http", *mode)
		os.Exit(1)
	}

	if err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
