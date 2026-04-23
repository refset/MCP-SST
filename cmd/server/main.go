package main

// MCP bridge between Claude-Code (stdio) and an SSTorytime SPA tab
// (WebSocket). See generated-server/mcptools/hub.go for the hub
// protocol.
//
// Run:
//   go run ./cmd/server                  # stdio MCP + ws hub on :8889
//   go run ./cmd/server -ws :9100        # alternate ws port
//
// Install in Claude Code:
//   claude mcp add sst -- go run ./cmd/server

import (
	"flag"
	"fmt"
	"os"

	"github.com/markburgess/MCP-SST/generated-server/mcptools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	wsAddr := flag.String("ws", ":8889", "address for the SPA websocket hub")
	flag.Usage = Usage
	flag.Parse()

	// 1. Start the WS hub (non-blocking).
	mcptools.StartWebSocketHub(*wsAddr)

	// 2. Build the MCP server + register tools.
	mcpserver := server.NewMCPServer("MCP-SSTorytime", "1.1.0")
	mcpserver.AddTool(mcptools.NewN4LqueryMCPTool(), mcptools.N4LqueryHandler)
	mcpserver.AddTool(mcptools.NewListSessionsMCPTool(), mcptools.ListSessionsHandler)

	// 3. Serve MCP over stdio — the transport Claude Code expects
	// when registered with `claude mcp add <name> -- <cmd>`.
	if err := server.ServeStdio(mcpserver); err != nil {
		fmt.Fprintf(os.Stderr, "MCP stdio server: %v\n", err)
		os.Exit(1)
	}
}

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-ws :port]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}
