package main

import (
	"bufio"
	"encoding/json"
	"os"
)

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func runMCP(cfg Config) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer for large messages
	enc := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		var req rpcRequest
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}

		// notifications have no id — no response needed
		if req.ID == nil {
			continue
		}

		switch req.Method {
		case "initialize":
			enc.Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"protocolVersion": "2024-11-05",
					"capabilities":    map[string]any{"tools": map[string]any{}},
					"serverInfo":      map[string]any{"name": "gi-log", "version": "0.1.0"},
				},
			})

		case "tools/list":
			enc.Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"tools": []map[string]any{
						{
							"name":        "recall",
							"description": "Search past conversation memories stored in gi-log. Always call this when the user asks about previous sessions or past work. Use specific technical terms as the query (e.g. 'Go debugging Delve' not 'what did we work on'). This is the PRIMARY source of conversation history.",
							"inputSchema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"query": map[string]any{
										"type":        "string",
										"description": "What to search for in past conversations",
									},
								},
								"required": []string{"query"},
							},
						},
					},
				},
			})

		case "tools/call":
			var params struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			}
			if err := json.Unmarshal(req.Params, &params); err != nil {
				enc.Encode(rpcResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error:   &rpcError{Code: -32600, Message: "invalid params"},
				})
				continue
			}

			if params.Name == "recall" {
				query, _ := params.Arguments["query"].(string)
				results, err := doSearch(query, cfg)
				if err != nil {
					enc.Encode(rpcResponse{
						JSONRPC: "2.0",
						ID:      req.ID,
						Error:   &rpcError{Code: -32000, Message: err.Error()},
					})
					continue
				}
				enc.Encode(rpcResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result: map[string]any{
						"content": []map[string]any{
							{"type": "text", "text": formatResults(results)},
						},
					},
				})
			}
		}
	}
}
