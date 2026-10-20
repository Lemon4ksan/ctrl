// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mcpkit

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lemon4ksan/foundation/generic"
)

const maxRPCBodyBytes = 16 * 1024 * 1024

// HTTPServerConfig specifies configuration options for the MCP HTTP/SSE transport.
type HTTPServerConfig struct {
	Addr                string        // Listen address (default: "127.0.0.1")
	Port                int           // Listen port (default: 8080; 0 for dynamic)
	Transport           string        // "sse" or "http" (default: "sse")
	BearerToken         string        // Optional shared secret for Bearer token authentication
	TLSCertFile         string        // Path to TLS certificate PEM file
	TLSKeyFile          string        // Path to TLS private key PEM file
	ReadTimeout         time.Duration // HTTP read timeout (default: 30s)
	RateLimitRate       float64       // Refill rate in tokens/sec (default: 50.0)
	RateLimitBurst      int           // Maximum burst token capacity (default: 100)
	DisableRateLimiting bool          // Set true to disable token-bucket rate limiting
	ServiceName         string        // Service identifier in /health and WWW-Authenticate realm (default: "stasis-mcp")
	BannerText          string        // Optional plain-text banner returned on GET /
}

type sseClient struct {
	id      string
	msgChan chan []byte
	done    chan struct{}
}

// HTTPServer wraps an MCP JSON-RPC RequestHandler with HTTP and Server-Sent Events (SSE) transports.
type HTTPServer struct {
	server    RequestHandler
	cfg       HTTPServerConfig
	listener  net.Listener
	httpSrv   *http.Server
	clientsMu sync.RWMutex
	clients   map[string]*sseClient
	limiter   *RateLimiter
}

// NewHTTPServer constructs a new HTTP/SSE transport server backed by server.
func NewHTTPServer(server RequestHandler, cfg HTTPServerConfig) *HTTPServer {
	cfg.Addr = generic.Coalesce(cfg.Addr, "127.0.0.1")
	if cfg.Port <= 0 && cfg.Port != -1 {
		cfg.Port = 8080
	}
	cfg.Transport = generic.Coalesce(strings.ToLower(cfg.Transport), "sse")
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 30 * time.Second
	}
	if cfg.RateLimitRate <= 0 {
		cfg.RateLimitRate = 50.0
	}
	if cfg.RateLimitBurst <= 0 {
		cfg.RateLimitBurst = 100
	}
	cfg.ServiceName = generic.Coalesce(cfg.ServiceName, "stasis-mcp")
	cfg.BannerText = generic.Coalesce(cfg.BannerText, "MCP v1.0 Server (endpoints: /sse, /messages, /rpc)\n")

	var limiter *RateLimiter
	if !cfg.DisableRateLimiting {
		limiter = NewRateLimiter(cfg.RateLimitRate, cfg.RateLimitBurst)
	}

	return &HTTPServer{
		server:  server,
		cfg:     cfg,
		clients: make(map[string]*sseClient),
		limiter: limiter,
	}
}

// Config returns the configuration currently applied to the HTTP server.
func (h *HTTPServer) Config() HTTPServerConfig {
	return h.cfg
}

// Addr returns the net.Addr on which the server is listening, or nil if not listening.
func (h *HTTPServer) Addr() net.Addr {
	if h.listener != nil {
		return h.listener.Addr()
	}
	return nil
}

// Start binds to the configured network address and serves HTTP/SSE requests until ctx is cancelled.
func (h *HTTPServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", h.cfg.Addr, h.cfg.Port)
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	return h.Serve(ctx, l)
}

// Serve accepts incoming connections on l and serves HTTP/SSE requests until ctx is cancelled.
func (h *HTTPServer) Serve(ctx context.Context, l net.Listener) error {
	h.listener = l
	mux := http.NewServeMux()

	mux.HandleFunc("/sse", h.handleSSE)
	mux.HandleFunc("/messages", h.handleMessages)
	mux.HandleFunc("/rpc", h.handleRPC)
	mux.HandleFunc("/", h.handleRoot)
	mux.HandleFunc("/health", h.handleHealth)

	handler := h.rateLimitMiddleware(h.authMiddleware(mux))

	h.httpSrv = &http.Server{
		Handler:     handler,
		ReadTimeout: h.cfg.ReadTimeout,
	}

	errChan := make(chan error, 1)
	go func() {
		if h.cfg.TLSCertFile != "" && h.cfg.TLSKeyFile != "" {
			errChan <- h.httpSrv.ServeTLS(l, h.cfg.TLSCertFile, h.cfg.TLSKeyFile)
		} else {
			errChan <- h.httpSrv.Serve(l)
		}
	}()

	select {
	case <-ctx.Done():
		if h.limiter != nil {
			h.limiter.Stop()
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = h.httpSrv.Shutdown(shutdownCtx)
		return nil
	case err := <-errChan:
		if h.limiter != nil {
			h.limiter.Stop()
		}
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (h *HTTPServer) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.cfg.DisableRateLimiting && h.limiter != nil {
			if r.URL.Path != "/health" {
				clientIP := extractClientIP(r)
				if !h.limiter.Allow(clientIP) {
					w.Header().Set("Retry-After", "1")
					w.Header().Set("X-RateLimit-Limit", strconv.Itoa(h.cfg.RateLimitBurst))
					http.Error(w, "Too Many Requests: rate limit exceeded", http.StatusTooManyRequests)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Limiter returns the underlying token-bucket rate limiter.
func (h *HTTPServer) Limiter() *RateLimiter {
	return h.limiter
}

func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func (h *HTTPServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		if h.cfg.BearerToken != "" {
			token := ""
			authHeader := r.Header.Get("Authorization")
			if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
				token = after
			} else if qTok := r.URL.Query().Get("token"); qTok != "" {
				token = qTok
			}

			if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(h.cfg.BearerToken)) != 1 {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm=%q`, h.cfg.ServiceName))
				http.Error(w, "Unauthorized: invalid or missing Bearer token", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (h *HTTPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	sessionID := generateSessionID()
	client := &sseClient{
		id:      sessionID,
		msgChan: make(chan []byte, 64),
		done:    make(chan struct{}),
	}

	h.clientsMu.Lock()
	h.clients[sessionID] = client
	h.clientsMu.Unlock()

	defer func() {
		h.clientsMu.Lock()
		delete(h.clients, sessionID)
		h.clientsMu.Unlock()
		close(client.done)
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	endpointURI := "/messages?sessionId=" + sessionID
	_, _ = fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpointURI)
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-client.msgChan:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func decodeRPCRequest(w http.ResponseWriter, r *http.Request) (JSONRPCRequest, bool) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed (expected POST)", http.StatusMethodNotAllowed)
		return JSONRPCRequest{}, false
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRPCBodyBytes))
	if err != nil {
		http.Error(w, "Failed to read request body: "+err.Error(), http.StatusBadRequest)
		return JSONRPCRequest{}, false
	}

	var rpcReq JSONRPCRequest
	if err := json.Unmarshal(body, &rpcReq); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: ErrCodeParseError, Message: "parse error: " + err.Error()},
		})
		return JSONRPCRequest{}, false
	}

	return rpcReq, true
}

func (h *HTTPServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	rpcReq, ok := decodeRPCRequest(w, r)
	if !ok {
		return
	}

	resp := h.server.HandleRequest(r.Context(), rpcReq)
	if resp == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal JSON-RPC response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	deliveredSSE := false
	if sessionID != "" {
		h.clientsMu.RLock()
		client, exists := h.clients[sessionID]
		h.clientsMu.RUnlock()
		if exists {
			select {
			case client.msgChan <- respBytes:
				deliveredSSE = true
			default:
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if deliveredSSE {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_, _ = w.Write(respBytes)
}

func (h *HTTPServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	rpcReq, ok := decodeRPCRequest(w, r)
	if !ok {
		return
	}

	resp := h.server.HandleRequest(r.Context(), rpcReq)
	if resp == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

func (h *HTTPServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodPost {
		h.handleRPC(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprint(w, h.cfg.BannerText)
}

func (h *HTTPServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSONResponse(w, http.StatusOK, map[string]any{
		"status":  "healthy",
		"service": h.cfg.ServiceName,
	})
}

func writeJSONResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func generateSessionID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
