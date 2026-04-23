package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

const n4lDriveUISchema = `{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "Query string to type into the SPA's search box (same syntax as N4Lquery)."
    },
    "session_id": {
      "type": "string",
      "description": "Optional SPA session id. If omitted, the first connected session is used."
    }
  },
  "required": ["name"],
  "additionalProperties": false
}`

func NewN4LdriveUIMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"N4LdriveUI",
		"Drive the SSTorytime SPA's search box: types the query into the browser's #name input and clicks Go!, rendering the full UI (orbits/panels/history) for the user to see. Returns a small ack, not the search payload — use N4Lquery if you need the raw data.",
		[]byte(n4lDriveUISchema),
	)
}

func N4LdriveUIHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var (
		queryStr  string
		sessionID string
	)
	if args, ok := request.Params.Arguments.(map[string]any); ok {
		if n, ok := args["name"].(string); ok {
			queryStr = n
		}
		if s, ok := args["session_id"].(string); ok {
			sessionID = s
		}
	}
	if queryStr == "" {
		return nil, fmt.Errorf("N4LdriveUI: missing required field 'name'")
	}

	if sessionID == "" {
		picked, err := PickSession()
		if err != nil {
			return nil, err
		}
		sessionID = picked
	}

	raw, err := CallSession(sessionID, "uiSearch", map[string]any{"name": queryStr}, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("N4LdriveUI: %w", err)
	}

	out := string(raw)
	if out == "" {
		out = "null"
	}
	var probe any
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("N4LdriveUI: session returned non-JSON payload: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(out)},
	}, nil
}
