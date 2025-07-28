package mcp

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// COSCUPServer represents the COSCUP MCP server
type COSCUPServer struct {
	mcpServer *server.MCPServer
}

// getAvailableToolsList dynamically generates a list of available tools
func getAvailableToolsList() string {
	tools := CreateMCPTools()
	var toolNames []string
	for name := range tools {
		toolNames = append(toolNames, name)
	}
	sort.Strings(toolNames)
	return strings.Join(toolNames, ", ")
}

// NewCOSCUPServer creates a new COSCUP MCP server instance
func NewCOSCUPServer() *COSCUPServer {
	return &COSCUPServer{}
}

// Start initializes and starts the MCP server
func (s *COSCUPServer) Start() error {
	log.Println("Starting COSCUP MCP Server...")

	// COSCUP data is automatically loaded via init() when the package loads
	log.Println("COSCUP session data ready")

	// Create MCP server
	s.mcpServer = server.NewMCPServer(
		"COSCUP Schedule Planner",
		"1.0.0",
		server.WithLogging(),
		server.WithToolCapabilities(false),
	)

	// Register all tools
	if err := s.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Start cleanup routine for old sessions
	go s.startCleanupRoutine()

	log.Println("COSCUP MCP Server is ready!")
	log.Printf("Available tools: %s", getAvailableToolsList())

	// Start serving (this will block)
	return server.ServeStdio(s.mcpServer)
}

// registerTools registers all MCP tools with their handlers
func (s *COSCUPServer) registerTools() error {
	tools := CreateMCPTools()
	handlers := GetToolHandlers()

	for toolName, tool := range tools {
		handler, exists := handlers[toolName]
		if !exists {
			return fmt.Errorf("no handler found for tool: %s", toolName)
		}

		s.mcpServer.AddTool(tool, handler)
		log.Printf("Registered tool: %s", toolName)
	}

	return nil
}

// StartHTTP initializes and starts the MCP server in HTTP mode
func (s *COSCUPServer) StartHTTP() error {
	log.Println("Starting COSCUP MCP Server in HTTP mode...")

	// COSCUP data is automatically loaded via init() when the package loads
	log.Println("COSCUP session data ready")

	// Create MCP server
	s.mcpServer = server.NewMCPServer(
		"COSCUP Schedule Planner",
		"1.0.0",
		server.WithLogging(),
		server.WithToolCapabilities(false),
	)

	// Register all tools
	if err := s.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Start cleanup routine for old sessions
	go s.startCleanupRoutine()

	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("COSCUP MCP Server is ready!")
	log.Printf("Available tools: %s", getAvailableToolsList())
	log.Printf("Starting HTTP server on port %s", port)

	// Create a custom HTTP server with both MCP and health endpoints
	mux := http.NewServeMux()

	// Add health check endpoints
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/", s.healthHandler) // Also respond to root path

	// Create StreamableHTTP server with custom endpoint path
	httpServer := server.NewStreamableHTTPServer(s.mcpServer,
		server.WithEndpointPath("/mcp"),
	)

	// Handle MCP endpoints with connection logging
	mux.Handle("/mcp", s.loggingMiddleware(httpServer))
	mux.Handle("/mcp/", s.loggingMiddleware(httpServer))

	// Start HTTP server
	log.Printf("HTTP Server listening on :%s", port)
	return http.ListenAndServe(":"+port, mux)
}

// healthHandler provides a simple health check endpoint
func (s *COSCUPServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"coscup-mcp-server","version":"1.0.0"}`))
}


// loggingMiddleware logs HTTP requests for debugging
func (s *COSCUPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Call the next handler
		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf("[HTTP] %s %s completed in %v", r.Method, r.URL.Path, duration)
	})
}

// startCleanupRoutine starts a background routine to cleanup old sessions
func (s *COSCUPServer) startCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour) // cleanup every hour
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Running session cleanup...")
		CleanupOldSessions()
		stats := GetSessionStats()
		log.Printf("Active sessions: %v", stats["active_sessions"])
	}
}
