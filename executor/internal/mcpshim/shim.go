package mcpshim

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Allowed tools: approvals / Telegram / memory only (§5.3).
// This shim is stdio on the worker — not a TCP tunnel to :18789 (§4.4, §12.9).
const (
	ToolApprovals = "approvals"
	ToolTelegram  = "telegram"
	ToolMemory    = "memory"
)

var allowed = []string{ToolApprovals, ToolTelegram, ToolMemory}

var denied = map[string]struct{}{
	"exec": {}, "spawn": {}, "shell": {}, "fs_write": {}, "science-run": {},
}

// Tool describes a stub MCP tool.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListTools returns the only advertised tools.
func ListTools() []Tool {
	return []Tool{
		{Name: ToolApprovals, Description: "OpenClaw approvals surface (human gate, not a shell)"},
		{Name: ToolTelegram, Description: "Telegram notifications via OpenClaw"},
		{Name: ToolMemory, Description: "Long-term memory via OpenClaw"},
	}
}

// CallResult is a stub tools/call response.
type CallResult struct {
	Denied  bool   `json:"denied"`
	Message string `json:"message"`
}

// CallTool advertises only approvals/telegram/memory and denies exec/spawn/shell/fs_write/science-run.
func CallTool(name string) CallResult {
	n := strings.TrimSpace(name)
	if _, ok := denied[n]; ok {
		return CallResult{Denied: true, Message: "deny"}
	}
	for _, t := range allowed {
		if n == t {
			return CallResult{Denied: false, Message: "stub: not implemented"}
		}
	}
	return CallResult{Denied: true, Message: "deny"}
}

// RPCRequest is a minimal JSON-RPC 2.0 request.
type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// HandleLine processes one NDJSON JSON-RPC request for the stdio shim.
func HandleLine(line []byte) ([]byte, error) {
	var req RPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		return marshalRPCError(nil, -32700, "parse error"), nil
	}
	switch req.Method {
	case "initialize":
		return marshalRPC(req.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo":      map[string]string{"name": "aeon-mcp-shim", "version": "0.0.0"},
			"capabilities":    map[string]any{"tools": map[string]any{}},
		}), nil
	case "tools/list", "list-tools":
		return marshalRPC(req.ID, map[string]any{"tools": ListTools()}), nil
	case "tools/call":
		var p struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(req.Params, &p)
		res := CallTool(p.Name)
		if res.Denied {
			return marshalRPC(req.ID, map[string]any{
				"isError": true,
				"content": []map[string]string{{"type": "text", "text": "deny"}},
			}), nil
		}
		return marshalRPC(req.ID, map[string]any{
			"content": []map[string]string{{"type": "text", "text": res.Message}},
		}), nil
	default:
		return marshalRPCError(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method)), nil
	}
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func marshalRPC(id json.RawMessage, result any) []byte {
	if len(id) == 0 {
		id = []byte("null")
	}
	b, _ := json.Marshal(rpcResponse{JSONRPC: "2.0", ID: id, Result: result})
	return b
}

func marshalRPCError(id json.RawMessage, code int, msg string) []byte {
	if len(id) == 0 {
		id = []byte("null")
	}
	b, _ := json.Marshal(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	})
	return b
}
