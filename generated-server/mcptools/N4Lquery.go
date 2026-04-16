package mcptools

import (
	"net/http"
	"net/url"
	"io"
	"context"
	"fmt"
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

// Input Schema for the N4Lquery tool

const N4LqueryInputSchema = `{
  "properties": {
    "body": {
      "properties": {
        "chapcontext": {
          "description": "string used for web group updates",
          "type": "string"
        },
        "name": {
          "description": "A search string in the N4L query language",
          "type": "string"
        },
        "nclass": {
          "description": "NPtr string classification for direct lookup",
          "type": "string"
        },
        "ncptr": {
          "description": "NPtr hash number for direct lookup",
          "type": "string"
        }
      },
      "required": [
        "name"
      ],
      "type": "object"
    }
  },
  "required": [
    "body"
  ],
  "type": "object"
}`

// Response Template for the N4Lquery tool (Status: 200, Content-Type: application/json)

const N4LqueryResponseTemplate_A = `# API Response Information

Below is the response template for this API endpoint.

The template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.

**Status Code:** 200

**Content-Type:** application/json

> Default for all non error reponses

## Response Structure

- Structure (Type: object):
  - **Response** (Type: string):
  - **Content** (Type: Combinator):
    - **One Of the following structures**:
      - **Option 1**: An array of nearest neighbours of the selected knowable text (Type: array):
        - **Items** (Type: array):
          - **Items**: An element in orbit around the knowable text (Type: object):
            - **Ctx**: The context key words of the orbiting item (Type: string):
            - **Dst**: An internal direct reference to a knowable item (Type: object):
              - **CPtr** (Type: integer):
              - **Class** (Type: integer):
            - **Text**: The dereferenced text of the orbiting item (Type: string):
            - **Wgt**: The link weight of the arrow (Type: number):
            - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
              - **Lat** (Type: number):
              - **Lon** (Type: number):
              - **R** (Type: number):
              - **X** (Type: number):
              - **Y** (Type: number):
              - **Z** (Type: number):
            - **STindex**: The Semantic Spacetime type (0-6) of the arrow (Type: integer):
            - **Arrow**: The arrow type relating the orbiting item to the base knowable text (Type: string):
            - **OOO**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
              - **R** (Type: number):
              - **X** (Type: number):
              - **Y** (Type: number):
              - **Z** (Type: number):
              - **Lat** (Type: number):
              - **Lon** (Type: number):
            - **Radius**: The distance of the related item from the base knowable text (Type: integer):
      - **Option 2**: A structure listing paths from a root node (Type: object):
        - **SuperNodes**: A list of nodes identified by their symmetrical roles in the paths (Type: array):
          - **Items** (Type: string):
        - **Title**: The text of the root node (Type: string):
        - **BTWC**: An array of Betweenness Centrality scores for the path matrix (Type: array):
          - **Items** (Type: string):
        - **Paths**: An array of paths starting from the root node (Type: array):
          - **Items** (Type: array):
            - **Items**: A curated journey through the knowable graph (Type: object):
              - **NPtr**: An internal direct reference to a knowable item (Type: object):
                - **CPtr** (Type: integer):
                - **Class** (Type: integer):
              - **STindex**: The Semantic Spacetime (0-6) type of the arrow (Type: integer):
              - **Chp**: The chapter the text belongs to (Type: string):
              - **Wgt**: The weight assigned to the arrow link (Type: number):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **R** (Type: number):
                - **X** (Type: number):
                - **Y** (Type: number):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
              - **Arr**: The arrow type of the next leg of the path (Type: integer):
              - **Ctx**: The context keywords the text belongs to (Type: string):
              - **Line**: The line number reference of the text in the original N4L notes (Type: integer):
              - **Name**: The text for this location in the path (Type: string):
        - **RootNode**: An internal direct reference to a knowable item (Type: object):
          - **CPtr** (Type: integer):
          - **Class** (Type: integer):
      - **Option 3**: A structure listing paths from a root node (Type: object):
        - **SuperNodes**: A list of nodes identified by their symmetrical roles in the paths (Type: array):
          - **Items** (Type: string):
        - **Title**: The text of the root node (Type: string):
        - **BTWC**: An array of Betweenness Centrality scores for the path matrix (Type: array):
          - **Items** (Type: string):
        - **Paths**: An array of paths starting from the root node (Type: array):
          - **Items** (Type: array):
            - **Items**: A curated journey through the knowable graph (Type: object):
              - **Wgt**: The weight assigned to the arrow link (Type: number):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **Y** (Type: number):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
                - **R** (Type: number):
                - **X** (Type: number):
              - **Arr**: The arrow type of the next leg of the path (Type: integer):
              - **Ctx**: The context keywords the text belongs to (Type: string):
              - **Line**: The line number reference of the text in the original N4L notes (Type: integer):
              - **Name**: The text for this location in the path (Type: string):
              - **NPtr**: An internal direct reference to a knowable item (Type: object):
                - **Class** (Type: integer):
                - **CPtr** (Type: integer):
              - **STindex**: The Semantic Spacetime (0-6) type of the arrow (Type: integer):
              - **Chp**: The chapter the text belongs to (Type: string):
        - **RootNode**: An internal direct reference to a knowable item (Type: object):
          - **CPtr** (Type: integer):
          - **Class** (Type: integer):
      - **Option 4**: An array of NodeEvents forming a story sequence about the search topic (Type: array):
        - **Items**: The content pointed to by a NodePtr (Type: object):
          - **L**: The string length of the text (Type: integer):
          - **NPtr**: An internal direct reference to a knowable item (Type: object):
            - **CPtr** (Type: integer):
            - **Class** (Type: integer):
          - **Text**: The knowable data (Type: string):
          - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
            - **X** (Type: number):
            - **Y** (Type: number):
            - **Z** (Type: number):
            - **Lat** (Type: number):
            - **Lon** (Type: number):
            - **R** (Type: number):
          - **Chap**: The chapter containing the text (Type: string):
          - **Context**: The context keywords explaining the relevance of the text (Type: string):
      - **Option 5**: A literal rendering of the original N4L page notes (Type: object):
        - **Context**: The context keywords labelling the knowable item (Type: string):
        - **Notes**: An array of the related notes matching the knowable item (Type: array):
          - **Items** (Type: array):
            - **Items**: A curated journey through the knowable graph (Type: object):
              - **Line**: The line number reference of the text in the original N4L notes (Type: integer):
              - **Name**: The text for this location in the path (Type: string):
              - **NPtr**: An internal direct reference to a knowable item (Type: object):
                - **CPtr** (Type: integer):
                - **Class** (Type: integer):
              - **STindex**: The Semantic Spacetime (0-6) type of the arrow (Type: integer):
              - **Chp**: The chapter the text belongs to (Type: string):
              - **Wgt**: The weight assigned to the arrow link (Type: number):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **R** (Type: number):
                - **X** (Type: number):
                - **Y** (Type: number):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
              - **Arr**: The arrow type of the next leg of the path (Type: integer):
              - **Ctx**: The context keywords the text belongs to (Type: string):
        - **Title**: The knowable item (Type: string):
      - **Option 6**: A table of contents matching chapters and contexts (Type: array):
        - **Items** (Type: object):
          - **Context** (Type: array):
            - **Items** (Type: object):
              - **Reln** (Type: array):
                - **Items** (Type: integer):
              - **Text** (Type: string):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **X** (Type: number):
                - **Y** (Type: number):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
                - **R** (Type: number):
          - **Single** (Type: array):
            - **Items** (Type: object):
              - **Reln** (Type: array):
                - **Items** (Type: integer):
              - **Text** (Type: string):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
                - **R** (Type: number):
                - **X** (Type: number):
                - **Y** (Type: number):
          - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
            - **Y** (Type: number):
            - **Z** (Type: number):
            - **Lat** (Type: number):
            - **Lon** (Type: number):
            - **R** (Type: number):
            - **X** (Type: number):
          - **Chapter** (Type: string):
          - **Common** (Type: array):
            - **Items** (Type: object):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **Y** (Type: number):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
                - **R** (Type: number):
                - **X** (Type: number):
              - **Reln** (Type: array):
                - **Items** (Type: integer):
              - **Text** (Type: string):
      - **Option 7** (Type: array):
        - **Items**: Spceific information explaining a single type of arrow (Type: object):
          - **Long** (Type: string):
          - **Short** (Type: string):
          - **ASTtype** (Type: integer):
          - **ArrPtr** (Type: integer):
          - **ISTtype** (Type: integer):
          - **InvL** (Type: string):
          - **InvPtr** (Type: integer):
          - **InvS** (Type: string):
`

// NewN4LqueryMCPTool creates the MCP Tool instance for N4Lquery

func NewN4LqueryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"N4Lquery",
		"Interpret search command",
		[]byte(N4LqueryInputSchema),
	)
}

// N4LqueryHandler is the handler function for the N4Lquery tool.
// This function is automatically generated. Users should implement the actual
// logic within this function body to integrate with backend APIs.
// You can generate types, http client and helpers for parsing request params to facilitate the implementation.

func N4LqueryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {


	/* type CallToolRequest struct {
                  JSONRPC string `json:"jsonrpc"`
                  ID      string `json:"id"`
                  Method  string `json:"method"` // Always "tools/call"
                  Params  struct {
                  Name      string                 `json:"name"`
                  Arguments map[string]interface{} `json:"arguments"`
                  } `json:"params"`
          }*/

	var mcp_search_command string
	
	if args, ok := request.Params.Arguments.(map[string]any); ok {
		mcp_search_command = args["command_request"].(string)
		fmt.Println("DEBUG ARG string:",request.Params.Name,"ARGS",mcp_search_command)
	}
	
	// Make HTTP calls or interact with services as needed.
	// Return an *mcp.CallToolResult with the response payload, or an error.

	// We need to submit a simple form to http_server

	formdata := url.Values{
		"name": { mcp_search_command },
	}

	// Reroute query to the SST server
	uri := "https://127.0.0.1:8443/searchN4L"

	var body []byte
	
	fmt.Println("Test the certificate",uri,formdata)
	
	resp, err := http.PostForm(uri, formdata)
	
	if err != nil {
		fmt.Printf("POST: Unable to forward request: %s\n", "N4Lquery")
		body = SelfSignedForm(uri,mcp_search_command,formdata)
		return nil, fmt.Errorf("%s not implemented", "N4Lquery")
	}
	
	defer resp.Body.Close()
	
	body, _ = io.ReadAll(resp.Body)
	
	/* type CallToolResult struct {
                 Content []Content `json:"content"`
                 IsError bool      `json:"isError,omitempty"`
        } */
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Text: fmt.Sprintf("%v", string(body))}},
	}, nil

	return nil, fmt.Errorf("%s not implemented", "N4Lquery")
}


// *********************************************************************

func SelfSignedForm(uri,query string,formdata url.Values) []byte {
	
	// curl -Iv https://127.0.0.1:8443 --cacert cert.pem

	caCert, err := os.ReadFile("../server/cert.pem")
	
	if err != nil {
		fmt.Println("Couldn't load server's self-signed certificate file",err)
		return nil
	}
	
	// 2. Create a CertPool and add the CA certificate
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	
	// 3. Configure TLS with the custom CertPool
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	
	// 4. Create an HTTP client with the custom TLS configuration
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	
	fmt.Println("Try to connect FORM",uri,formdata)
	
	resp, err2 := client.PostForm(uri, formdata)
	
	if err2 != nil {
		fmt.Printf("POST: Unable to forward request: %s\n", "N4Lquery")
		return nil
	}
	
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	fmt.Println("SELF_SIGNED",string(body))
	return body
}
