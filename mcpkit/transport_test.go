// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mcpkit

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHTTPServer_Health(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "test-mcp", Version: "1.0.0"})
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	httpSrv := NewHTTPServer(srv, HTTPServerConfig{
		Transport: "http",
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = httpSrv.Serve(ctx, l)
	}()

	client := &http.Client{Timeout: 3 * time.Second}
	url := fmt.Sprintf("http://%s/health", l.Addr().String())
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatal(err)
	}
	if data["status"] != "healthy" {
		t.Fatalf("unexpected health payload: %v", data)
	}
}

func TestHTTPServer_BearerAuth(t *testing.T) {
	const secret = "aegis-secret-token-12345"
	srv := NewServer(ServerInfo{Name: "test-mcp", Version: "1.0.0"})
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	httpSrv := NewHTTPServer(srv, HTTPServerConfig{
		Transport:   "http",
		BearerToken: secret,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = httpSrv.Serve(ctx, l)
	}()

	client := &http.Client{Timeout: 3 * time.Second}
	rpcURL := fmt.Sprintf("http://%s/rpc", l.Addr().String())

	pingReq := `{"jsonrpc":"2.0","id":1,"method":"ping"}`

	// 1. Request without token -> 401 Unauthorized
	resp, err := client.Post(rpcURL, "application/json", strings.NewReader(pingReq))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth header, got %d", resp.StatusCode)
	}

	// 2. Request with invalid token -> 401 Unauthorized
	req, _ := http.NewRequest(http.MethodPost, rpcURL, strings.NewReader(pingReq))
	req.Header.Set("Authorization", "Bearer wrong-secret")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong token, got %d", resp.StatusCode)
	}

	// 3. Request with valid Bearer header -> 200 OK
	req, _ = http.NewRequest(http.MethodPost, rpcURL, strings.NewReader(pingReq))
	req.Header.Set("Authorization", "Bearer "+secret)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with valid Bearer header, got %d", resp.StatusCode)
	}

	var rpcResp JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		t.Fatal(err)
	}
	if string(rpcResp.ID) != "1" {
		t.Fatalf("expected response ID 1, got %s", string(rpcResp.ID))
	}

	// 4. Request with valid query parameter token -> 200 OK
	queryURL := fmt.Sprintf("http://%s/rpc?token=%s", l.Addr().String(), secret)
	resp, err = client.Post(queryURL, "application/json", strings.NewReader(pingReq))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with valid query param token, got %d", resp.StatusCode)
	}
}

func TestHTTPServer_SSETransport(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "test-mcp", Version: "1.0.0"})
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	httpSrv := NewHTTPServer(srv, HTTPServerConfig{
		Transport: "sse",
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = httpSrv.Serve(ctx, l)
	}()

	client := &http.Client{Timeout: 5 * time.Second}
	sseURL := fmt.Sprintf("http://%s/sse", l.Addr().String())

	sseReq, err := http.NewRequest(http.MethodGet, sseURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	sseResp, err := client.Do(sseReq)
	if err != nil {
		t.Fatalf("GET /sse failed: %v", err)
	}
	defer sseResp.Body.Close()

	if sseResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for SSE, got %d", sseResp.StatusCode)
	}
	if ct := sseResp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("expected text/event-stream, got %q", ct)
	}

	reader := bufio.NewReader(sseResp.Body)
	var endpointLine string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("failed reading SSE stream: %v", err)
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: /messages?sessionId=") {
			endpointLine = line
			break
		}
	}

	endpointURI := strings.TrimPrefix(endpointLine, "data: ")
	postURL := fmt.Sprintf("http://%s%s", l.Addr().String(), endpointURI)

	initBody := `{"jsonrpc":"2.0","id":"init-1","method":"initialize"}`
	msgResp, err := client.Post(postURL, "application/json", strings.NewReader(initBody))
	if err != nil {
		t.Fatalf("POST to messages failed: %v", err)
	}
	msgResp.Body.Close()
	if msgResp.StatusCode != http.StatusAccepted && msgResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 202 Accepted or 200 OK, got %d", msgResp.StatusCode)
	}

	var messageReceived string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("failed reading message from SSE stream: %v", err)
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: {") {
			messageReceived = strings.TrimPrefix(line, "data: ")
			break
		}
	}

	var jsonResp JSONRPCResponse
	if err := json.Unmarshal([]byte(messageReceived), &jsonResp); err != nil {
		t.Fatalf("invalid JSON-RPC received over SSE: %v (raw: %s)", err, messageReceived)
	}
	if string(jsonResp.ID) != `"init-1"` {
		t.Fatalf("expected id init-1, got %s", string(jsonResp.ID))
	}
}

func TestHTTPServer_TLSTransport(t *testing.T) {
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "server.crt")
	keyFile := filepath.Join(tmpDir, "server.key")
	generateSelfSignedCert(t, certFile, keyFile)

	srv := NewServer(ServerInfo{Name: "test-mcp", Version: "1.0.0"})
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	httpSrv := NewHTTPServer(srv, HTTPServerConfig{
		Transport:   "http",
		TLSCertFile: certFile,
		TLSKeyFile:  keyFile,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = httpSrv.Serve(ctx, l)
	}()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // testing self-signed cert
	}
	client := &http.Client{Transport: tr, Timeout: 3 * time.Second}

	httpsURL := fmt.Sprintf("https://%s/rpc", l.Addr().String())
	reqBody := `{"jsonrpc":"2.0","id":"tls-test","method":"tools/list"}`
	resp, err := client.Post(httpsURL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("HTTPS request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK over HTTPS, got %d", resp.StatusCode)
	}

	var rpcResp JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		t.Fatal(err)
	}
	if string(rpcResp.ID) != `"tls-test"` {
		t.Fatalf("expected ID tls-test, got %s", string(rpcResp.ID))
	}
}

func generateSelfSignedCert(t *testing.T, certPath, keyPath string) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"MCPKit Security"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		t.Fatal(err)
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		t.Fatal(err)
	}

	keyOut, err := os.Create(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	defer keyOut.Close()
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}); err != nil {
		t.Fatal(err)
	}
}

func TestRateLimiter_TokenBucket(t *testing.T) {
	rl := NewRateLimiter(10.0, 3)
	defer rl.Stop()

	clientIP := "192.168.1.100"

	if !rl.Allow(clientIP) {
		t.Fatal("1st request should be allowed")
	}
	if !rl.Allow(clientIP) {
		t.Fatal("2nd request should be allowed")
	}
	if !rl.Allow(clientIP) {
		t.Fatal("3rd request should be allowed")
	}

	if rl.Allow(clientIP) {
		t.Fatal("4th request should be rejected (rate limited)")
	}

	otherIP := "10.0.0.1"
	if !rl.Allow(otherIP) {
		t.Fatal("request from other IP should be allowed")
	}

	time.Sleep(120 * time.Millisecond)
	if !rl.Allow(clientIP) {
		t.Fatal("request after refill duration should be allowed")
	}
}

func TestHTTPServer_RateLimiting(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "test-mcp", Version: "1.0.0"})
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	httpSrv := NewHTTPServer(srv, HTTPServerConfig{
		Transport:      "http",
		RateLimitRate:  1.0,
		RateLimitBurst: 2,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = httpSrv.Serve(ctx, l)
	}()

	client := &http.Client{Timeout: 3 * time.Second}
	rpcURL := fmt.Sprintf("http://%s/rpc", l.Addr().String())
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"ping"}`

	resp1, err := client.Post(rpcURL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp1.StatusCode)
	}

	resp2, err := client.Post(rpcURL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp2.StatusCode)
	}

	resp3, err := client.Post(rpcURL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests, got %d", resp3.StatusCode)
	}
	if retryAfter := resp3.Header.Get("Retry-After"); retryAfter != "1" {
		t.Fatalf("expected Retry-After: 1 header, got %q", retryAfter)
	}

	healthURL := fmt.Sprintf("http://%s/health", l.Addr().String())
	healthResp, err := client.Get(healthURL)
	if err != nil {
		t.Fatal(err)
	}
	healthResp.Body.Close()
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("health check should bypass rate limit, got %d", healthResp.StatusCode)
	}
}

func TestRateLimiter_StaleEviction(t *testing.T) {
	rl := NewRateLimiter(10.0, 5)
	defer rl.Stop()

	rl.Allow("1.1.1.1")
	rl.Allow("2.2.2.2")

	if rl.ClientCount() != 2 {
		t.Fatalf("expected 2 clients, got %d", rl.ClientCount())
	}
	rl.SetLastRefill("1.1.1.1", time.Now().Add(-5*time.Minute))

	rl.PruneStale()

	if rl.ClientCount() != 1 {
		t.Fatalf("expected 1 client after prune, got %d", rl.ClientCount())
	}
	if rl.HasClient("1.1.1.1") {
		t.Fatal("stale client 1.1.1.1 should have been evicted")
	}
	if !rl.HasClient("2.2.2.2") {
		t.Fatal("fresh client 2.2.2.2 should remain")
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter(1000000.0, 1000000)
	defer rl.Stop()

	clientIP := "10.10.10.10"
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rl.Allow(clientIP)
	}
}
