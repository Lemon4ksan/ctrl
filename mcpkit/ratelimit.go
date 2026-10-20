// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mcpkit

import (
	"math"
	"sync"
	"time"

	"github.com/lemon4ksan/foundation/generic"
)

type clientBucket struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter implements an in-memory token-bucket rate limiter tracking client requests
// by IP address and evicting stale buckets to mitigate memory DoS.
type RateLimiter struct {
	mu           sync.Mutex
	rate         float64
	burst        int
	clients      map[string]*clientBucket
	staleTimeout time.Duration
	clock        func() time.Time
	stopPrune    chan struct{}
}

// NewRateLimiter initializes a token-bucket rate limiter with the specified refill rate (tokens/sec) and burst capacity.
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	r := generic.Ternary(rate > 0, rate, 50.0)
	b := generic.Ternary(burst > 0, burst, 100)
	rl := &RateLimiter{
		rate:         r,
		burst:        b,
		clients:      make(map[string]*clientBucket),
		staleTimeout: 3 * time.Minute,
		clock:        time.Now,
		stopPrune:    make(chan struct{}),
	}
	go rl.pruneLoop()
	return rl
}

// WithClock configures a custom time source on RateLimiter for deterministic testing.
func (rl *RateLimiter) WithClock(nowFunc func() time.Time) *RateLimiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if nowFunc != nil {
		rl.clock = nowFunc
	}
	return rl
}

// Allow reports whether a single request from the given client key (typically IP address) is permitted.
func (rl *RateLimiter) Allow(clientKey string) bool {
	return rl.AllowN(clientKey, 1.0)
}

// AllowN reports whether n tokens can be consumed for the given client key.
func (rl *RateLimiter) AllowN(clientKey string, n float64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.clock()
	bucket, ok := rl.clients[clientKey]
	if !ok {
		if float64(rl.burst) < n {
			return false
		}
		rl.clients[clientKey] = &clientBucket{
			tokens:     float64(rl.burst) - n,
			lastRefill: now,
		}
		return true
	}

	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens = math.Min(float64(rl.burst), bucket.tokens+elapsed*rl.rate)
	bucket.lastRefill = now

	if bucket.tokens >= n {
		bucket.tokens -= n
		return true
	}
	return false
}

func (rl *RateLimiter) pruneLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stopPrune:
			return
		case <-ticker.C:
			rl.PruneStale()
		}
	}
}

// PruneStale evicts client buckets that have been inactive longer than staleTimeout.
func (rl *RateLimiter) PruneStale() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.clock()
	for ip, bucket := range rl.clients {
		if now.Sub(bucket.lastRefill) > rl.staleTimeout {
			delete(rl.clients, ip)
		}
	}
}

// ClientCount returns the number of active client buckets currently tracked.
func (rl *RateLimiter) ClientCount() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.clients)
}

// HasClient reports whether a bucket currently exists for clientKey.
func (rl *RateLimiter) HasClient(clientKey string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	_, exists := rl.clients[clientKey]
	return exists
}

// SetLastRefill adjusts the last refill timestamp for clientKey (used by deterministic eviction tests).
func (rl *RateLimiter) SetLastRefill(clientKey string, t time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if b, ok := rl.clients[clientKey]; ok {
		b.lastRefill = t
	}
}

// Stop terminates the background bucket eviction goroutine.
func (rl *RateLimiter) Stop() {
	select {
	case <-rl.stopPrune:
	default:
		close(rl.stopPrune)
	}
}
