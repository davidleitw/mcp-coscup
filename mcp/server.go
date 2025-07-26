package mcp

import (
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// COSCUPServer represents the COSCUP MCP server
type COSCUPServer struct {
	mcpServer *server.MCPServer
}

// NewCOSCUPServer creates a new COSCUP MCP server instance
func NewCOSCUPServer() *COSCUPServer {
	return &COSCUPServer{}
}

// Start initializes and starts the MCP server
func (s *COSCUPServer) Start() error {
	log.Println("Starting COSCUP MCP Server...")

	// Load COSCUP data first
	log.Println("Loading COSCUP session data...")
	if err := LoadCOSCUPData(); err != nil {
		return fmt.Errorf("failed to load COSCUP data: %w", err)
	}
	log.Println("COSCUP data loaded successfully")

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
	log.Println("Available tools: start_planning, choose_session, get_options, get_schedule, get_next_session, get_session_detail, finish_planning, help")

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

// GetStats returns server statistics
func (s *COSCUPServer) GetStats() map[string]any {
	sessionStats := GetSessionStats()

	return map[string]any{
		"server_name":   "COSCUP Schedule Planner",
		"version":       "1.0.0",
		"uptime":        time.Now().Format(time.RFC3339),
		"session_stats": sessionStats,
		"available_tools": []string{
			"start_planning",
			"choose_session",
			"get_options",
			"get_schedule",
			"get_next_session",
			"get_session_detail",
			"finish_planning",
			"help",
		},
	}
}
