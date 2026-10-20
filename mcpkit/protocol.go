// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mcpkit provides a universal, domain-agnostic Model Context Protocol
// (MCP v1.0, specification 2024-11-05) JSON-RPC 2.0 router, generic typed tool
// handler adapters, STDIO framing loop, and HTTP/SSE transport with Bearer
// authentication, TLS termination, and token-bucket rate limiting.
//
// It contains zero business-logic dependencies so that any CLI or agent daemon
// in the ecosystem (stasis, vortex, etc.) can expose MCP tools by registering
// schemas and strongly-typed handlers on a [Server].
package mcpkit

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ProtocolVersion is the MCP protocol revision negotiated during `initialize`.
const ProtocolVersion = "2024-11-05"

// Standard JSON-RPC 2.0 error codes (https://www.jsonrpc.org/specification#error_object).
const (
	ErrCodeParseError     = -32700 // Invalid JSON was received by the server.
	ErrCodeInvalidRequest = -32600 // The JSON sent is not a valid Request object.
	ErrCodeMethodNotFound = -32601 // The method does not exist or is not available.
	ErrCodeInvalidParams  = -32602 // Invalid method parameter(s).
	ErrCodeInternalError  = -32603 // Internal JSON-RPC error.
)

// JSONRPCRequest represents an incoming JSON-RPC 2.0 request or notification message.
// Per JSON-RPC 2.0 §4.1, a request without an ID (or with ID == "null") is treated
// as a one-way notification and must not produce a response payload.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// IsNotification reports whether the request is a one-way JSON-RPC 2.0 notification.
func (r JSONRPCRequest) IsNotification() bool {
	return len(r.ID) == 0 || string(r.ID) == "null"
}

// JSONRPCResponse represents an outgoing JSON-RPC 2.0 reply to the MCP host.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a structured JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ServerInfo identifies an MCP server during the `initialize` handshake.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolDefinition describes an MCP v1.0 tool and its JSON Schema input specification.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// ContentBlock represents an element in an MCP CallToolResult content array
// (e.g., "text" or base64-encoded "image").
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// CallToolResult is the standard MCP v1.0 response envelope for `tools/call`.
type CallToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// TextResult constructs a successful CallToolResult containing a single text block.
func TextResult(msg string) CallToolResult {
	return CallToolResult{
		Content: []ContentBlock{{Type: "text", Text: msg}},
	}
}

// ErrorResult constructs a CallToolResult with IsError=true from an error.
func ErrorResult(err error) CallToolResult {
	if err == nil {
		return TextResult("")
	}
	return CallToolResult{
		Content: []ContentBlock{{Type: "text", Text: err.Error()}},
		IsError: true,
	}
}

// ImageResult constructs a CallToolResult containing a base64-encoded image block
// followed by any optional telemetry or annotation text blocks.
func ImageResult(dataB64, mimeType string, extraBlocks ...ContentBlock) CallToolResult {
	blocks := make([]ContentBlock, 0, 1+len(extraBlocks))
	blocks = append(blocks, ContentBlock{
		Type:     "image",
		Data:     dataB64,
		MimeType: mimeType,
	})
	blocks = append(blocks, extraBlocks...)
	return CallToolResult{Content: blocks}
}

// RequestHandler abstracts any MCP JSON-RPC 2.0 dispatcher capable of handling
// a single request message over STDIO, HTTP POST, or SSE transports.
type RequestHandler interface {
	HandleRequest(ctx context.Context, req JSONRPCRequest) *JSONRPCResponse
}

// MethodHandlerFunc processes a custom JSON-RPC 2.0 method registered on [Server].
type MethodHandlerFunc func(ctx context.Context, req JSONRPCRequest) *JSONRPCResponse

// ToolHandlerFunc processes a single `tools/call` invocation with raw JSON arguments.
type ToolHandlerFunc func(ctx context.Context, rawArgs json.RawMessage) CallToolResult

// TypedHandler adapts a strongly-typed Go function into a ToolHandlerFunc.
// If rawArgs fails JSON unmarshaling, an ErrorResult is returned automatically.
func TypedHandler[T any](fn func(ctx context.Context, args T) CallToolResult) ToolHandlerFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) CallToolResult {
		var args T
		if len(rawArgs) > 0 {
			if err := json.Unmarshal(rawArgs, &args); err != nil {
				return ErrorResult(err)
			}
		}
		return fn(ctx, args)
	}
}

// LenientHandler adapts a strongly-typed Go function into a ToolHandlerFunc that
// tolerates empty or partially omitted JSON arguments without failing prior to invocation.
func LenientHandler[T any](fn func(ctx context.Context, args T) CallToolResult) ToolHandlerFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) CallToolResult {
		var args T
		if len(rawArgs) > 0 {
			_ = json.Unmarshal(rawArgs, &args)
		}
		return fn(ctx, args)
	}
}

// Server is a universal, domain-agnostic MCP v1.0 JSON-RPC 2.0 router.
// It implements the `initialize`, `ping`, `tools/list`, and `tools/call` methods,
// supports custom method registration via [Server.RegisterMethod], and provides
// a high-throughput newline-delimited STDIO loop.
type Server struct {
	info     ServerInfo
	tools    []ToolDefinition
	handlers map[string]ToolHandlerFunc
	methods  map[string]MethodHandlerFunc
}

// NewServer constructs a universal MCP v1.0 server with the given ServerInfo
// and optional initial tool definitions.
func NewServer(info ServerInfo, defs ...ToolDefinition) *Server {
	return &Server{
		info:     info,
		tools:    append([]ToolDefinition(nil), defs...),
		handlers: make(map[string]ToolHandlerFunc, len(defs)),
		methods:  make(map[string]MethodHandlerFunc),
	}
}

// RegisterHandler binds a ToolHandlerFunc to an existing tool name.
func (s *Server) RegisterHandler(name string, handler ToolHandlerFunc) {
	s.handlers[name] = handler
}

// RegisterTool registers both a tool schema definition and its handler function.
func (s *Server) RegisterTool(def ToolDefinition, handler ToolHandlerFunc) {
	s.tools = append(s.tools, def)
	s.handlers[def.Name] = handler
}

// RegisterMethod registers a custom JSON-RPC 2.0 method handler on the server,
// enabling extension with custom protocol methods (e.g. resources/list, prompts/list)
// without modifying the core dispatcher.
func (s *Server) RegisterMethod(method string, handler MethodHandlerFunc) {
	if s.methods == nil {
		s.methods = make(map[string]MethodHandlerFunc)
	}
	s.methods[method] = handler
}

// Serve processes newline-delimited JSON-RPC 2.0 messages from r and writes
// responses to w until EOF or context cancellation.
// The scanner buffer is sized up to 16 MiB to accommodate high-resolution base64
// screenshots or large DOM/diff payloads.
func (s *Server) Serve(ctx context.Context, r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	enc := json.NewEncoder(w)

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			_ = enc.Encode(JSONRPCResponse{
				JSONRPC: "2.0",
				Error:   &RPCError{Code: ErrCodeParseError, Message: "parse error: " + err.Error()},
			})
			continue
		}

		resp := s.HandleRequest(ctx, req)
		if resp != nil {
			if err := enc.Encode(resp); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

// HandleRequest dispatches a single JSON-RPC 2.0 request and returns a response
// (or nil for one-way notifications).
func (s *Server) HandleRequest(ctx context.Context, req JSONRPCRequest) *JSONRPCResponse {
	if req.IsNotification() {
		return nil
	}

	if customHandler, ok := s.methods[req.Method]; ok && customHandler != nil {
		return customHandler(ctx, req)
	}

	switch req.Method {
	case "initialize":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": ProtocolVersion,
				"serverInfo": map[string]any{
					"name":    s.info.Name,
					"version": s.info.Version,
				},
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
			},
		}

	case "ping":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{},
		}

	case "tools/list":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": s.tools,
			},
		}

	case "tools/call":
		var callParams struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			return &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: ErrCodeInvalidParams, Message: "invalid tools/call params: " + err.Error()},
			}
		}
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  s.CallTool(ctx, callParams.Name, callParams.Arguments),
		}

	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: ErrCodeMethodNotFound, Message: "method not found: " + req.Method},
		}
	}
}

// CallTool invokes a registered tool handler by name.
func (s *Server) CallTool(ctx context.Context, name string, rawArgs json.RawMessage) CallToolResult {
	handler, ok := s.handlers[name]
	if !ok {
		return ErrorResult(fmt.Errorf("unknown tool %q", name))
	}
	return handler(ctx, rawArgs)
}
