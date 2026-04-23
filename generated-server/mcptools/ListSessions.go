package mcptools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

const listSessionsSchema = `{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`

func NewListSessionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"sst_list_sessions",
		"List connected SPA sessions (session id + current repo target). Use this first if you don't know which session_id to target.",
		[]byte(listSessionsSchema),
	)
}

func ListSessionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	out, err := json.MarshalIndent(ListSessions(), "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(string(out))},
	}, nil
}
