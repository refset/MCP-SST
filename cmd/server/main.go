package main

import (
    "log"
    "os"
    // Import your generated code
    "github.com/markburgess/MCP-SST" 
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    // 1. Create a new MCP server
    s := mcp.NewServer()

    // 2. Register tools from generated code
    generated.RegisterTools(s)

    // 3. Serve via stdio
    if err := s.Serve(stdio.NewTransport()); err != nil {
        log.Fatalf("Fatal error: %v", err)
    }
}
