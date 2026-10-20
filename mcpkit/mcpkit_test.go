// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mcpkit

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestServer_STDIOAndTypedHandlers(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "test-mcp", Version: "1.0.0"})

	type echoArgs struct {
		Message string `json:"message"`
	}
	srv.RegisterTool(
		ToolDefinition{
			Name:        "echo",
			Description: "Echoes a message",
			InputSchema: map[string]any{"type": "object"},
		},
		TypedHandler(func(_ context.Context, args echoArgs) CallToolResult {
			if args.Message == "fail" {
				return ErrorResult(errors.New("forced failure"))
			}
			return TextResult("echo: " + args.Message)
		}),
	)

	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello"}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"unknown_tool","arguments":{}}}`,
	}

	in := strings.NewReader(strings.Join(requests, "\n") + "\n")
	var out bytes.Buffer

	if err := srv.Serve(context.Background(), in, &out); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 responses, got %d:\n%s", len(lines), out.String())
	}

	var callResp JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[3]), &callResp); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(callResp.Result)
	if !strings.Contains(string(raw), "echo: hello") {
		t.Fatalf("expected 'echo: hello' in response, got %s", string(raw))
	}
}

func TestHTTPServer_SSEAndRateLimit(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "test-mcp", Version: "1.0.0"})
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	httpSrv := NewHTTPServer(srv, HTTPServerConfig{
		Transport:      "sse",
		BearerToken:    "secret-token",
		RateLimitRate:  10.0,
		RateLimitBurst: 5,
		ServiceName:    "test-mcp",
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = httpSrv.Serve(ctx, l)
	}()

	client := &http.Client{Timeout: 3 * time.Second}

	// 1. Health check bypasses auth
	healthResp, err := client.Get(fmt.Sprintf("http://%s/health", l.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}
	healthResp.Body.Close()
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on /health, got %d", healthResp.StatusCode)
	}

	// 2. Connect SSE with query token
	sseResp, err := client.Get(fmt.Sprintf("http://%s/sse?token=secret-token", l.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}
	defer sseResp.Body.Close()

	reader := bufio.NewReader(sseResp.Body)
	var endpointURI string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: /messages?sessionId=") {
			endpointURI = strings.TrimPrefix(line, "data: ")
			break
		}
	}

	postURL := fmt.Sprintf("http://%s%s&token=secret-token", l.Addr().String(), endpointURI)
	msgResp, err := client.Post(
		postURL,
		"application/json",
		strings.NewReader(`{"jsonrpc":"2.0","id":"1","method":"ping"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	msgResp.Body.Close()
	if msgResp.StatusCode != http.StatusAccepted && msgResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 202/200, got %d", msgResp.StatusCode)
	}
}

func TestToolBuilderAndTextHandlers(t *testing.T) {
	def := NewTool("demo_tool", "Demo tool").
		String("name", "User name", Required()).
		Enum("mode", "Run mode", []string{"fast", "safe"}, Default("fast")).
		Int("count", "Iteration count", Min(1)).
		Number("scale", "Scale factor", Min(0.1), Max(1.0)).
		Bool("verbose", "Verbose output").
		StringArray("tags", "Tag list").
		IntArray("region", "ROI region", MinItems(4), MaxItems(4)).
		ObjectArray("steps", "Batch steps", func(item *ToolBuilder) {
			item.String("op", "Operation", Required())
		}, Required(), MinItems(1)).
		Build()

	if def.Name != "demo_tool" {
		t.Fatalf("unexpected tool name: %s", def.Name)
	}
	props, ok := def.InputSchema["properties"].(map[string]any)
	if !ok || len(props) != 8 {
		t.Fatalf("expected 8 properties, got %+v", def.InputSchema)
	}

	srv := NewServer(ServerInfo{Name: "builder-test", Version: "1.0.0"}, def)
	srv.RegisterHandler("demo_tool", TextHandler(func(_ context.Context, args struct {
		Name string `json:"name"`
	},
	) (string, error) {
		return "hello " + args.Name, nil
	}))
	srv.RegisterTool(
		NewTool("ping_no_args", "No args").Build(),
		NoArgsHandler(func(_ context.Context) (string, error) {
			return "pong", nil
		}),
	)

	res := srv.CallTool(context.Background(), "demo_tool", json.RawMessage(`{"name":"world"}`))
	if res.IsError || len(res.Content) == 0 || res.Content[0].Text != "hello world" {
		t.Fatalf("unexpected TextHandler result: %+v", res)
	}
	resNoArgs := srv.CallTool(context.Background(), "ping_no_args", nil)
	if resNoArgs.IsError || len(resNoArgs.Content) == 0 || resNoArgs.Content[0].Text != "pong" {
		t.Fatalf("unexpected NoArgsHandler result: %+v", resNoArgs)
	}
}
