package main

import (
	"net/http"
	"sync"
	"time"
)

// rateLimiter tracks request counts per IP using a sliding window
type rateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientWindow
	limit    int
	window   time.Duration
	cleanup  time.Duration
}

type clientWindow struct {
	count    int
	resetAt  time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		clients: make(map[string]*clientWindow),
		limit:   limit,
		window:  window,
		cleanup: window * 2,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cw, exists := rl.clients[ip]
	if !exists || now.After(cw.resetAt) {
		rl.clients[ip] = &clientWindow{count: 1, resetAt: now.Add(rl.window)}
		return true
	}
	cw.count++
	return cw.count <= rl.limit
}

func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, cw := range rl.clients {
			if now.After(cw.resetAt) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// rateLimitMiddleware limits requests per IP (100 requests per minute)
func rateLimitMiddleware(next http.Handler) http.Handler {
	limiter := newRateLimiter(100, time.Minute)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = fwd
		}
		if !limiter.allow(ip) {
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "rate limit exceeded",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
