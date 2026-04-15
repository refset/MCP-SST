package main

import (
	"fmt"
	//"github.com/markburgess/MCP-SST/generated-server" 
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// 1. Create a new MCP server
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "my-server",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{},
	)


	fmt.Println(server)
	//generated.RegisterTools(s)

	/* 3. Serve via stdio
    if err := s.Serve(stdio.NewTransport()); err != nil {
        log.Fatalf("Fatal error: %v", err)
    }*/
}
