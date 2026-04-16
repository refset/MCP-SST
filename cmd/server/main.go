package main

import (
	"fmt"
	"os"
	"github.com/markburgess/MCP-SST/generated-server/mcptools" 
	//"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)


func main() {
	// 1. Create a new MCP server

	mcpserver := server.NewMCPServer("MCP-SSTorytime", "1.0.0")
	

	
	// 2. Register a tool (AddTool) with a handler function

	mcpserver.AddTool(mcptools.NewN4LqueryMCPTool(),mcptools.N4LqueryHandler)

	httpServer := server.NewStreamableHTTPServer(mcpserver)


	// Start the server
	if err := httpServer.Start(":8888"); err != nil {
		fmt.Println("Server failed to start: %v", err)
		os.Exit(-1)
	}
}
