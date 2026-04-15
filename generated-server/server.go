package MCP_SST

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/markburgess/MCP-SST/generated-server/mcptools"
)

// NewMCPServer creates and returns an MCP server with all tools registered
func NewMCPServer() *server.MCPServer {
	// Create a new MCP server
	s := server.NewMCPServer(
		"MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// Register all tools
	s.AddTool(mcptools.NewN4LqueryMCPTool(), mcptools.N4LqueryHandler)

	return s
}
