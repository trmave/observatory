package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"observatory/internal/domain"
	"os"
)

// Simple MCP Server over Stdio
type MCPServer struct {
	repo domain.ProviderRepository
}

func NewMCPServer(repo domain.ProviderRepository) *MCPServer {
	return &MCPServer{repo: repo}
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func (s *MCPServer) Serve() {
	decoder := json.NewDecoder(os.Stdin)
	for {
		var req JSONRPCRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		go s.handleRequest(req)
	}
}

func (s *MCPServer) handleRequest(req JSONRPCRequest) {
	var res JSONRPCResponse
	res.JSONRPC = "2.0"
	res.ID = req.ID

	switch req.Method {
	case "initialize":
		res.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"serverInfo": map[string]string{
				"name":    "observatory-mcp",
				"version": "1.0.0",
			},
		}
	case "tools/list":
		res.Result = map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "get_infrastructure_status",
					"description": "Obtiene el estado actual de todos los proveedores de nube monitoreados.",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		}
	case "tools/call":
		providers, _ := s.repo.ListAll(context.Background())
		statusText := "Estado actual de la infraestructura:\n"
		for _, p := range providers {
			statusText += fmt.Sprintf("- %s (%s): %s\n", p.Name, p.Slug, p.CurrentStatus)
		}
		res.Result = map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": statusText,
				},
			},
		}
	}

	json.NewEncoder(os.Stdout).Encode(res)
}
