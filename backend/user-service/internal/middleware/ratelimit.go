package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter per IP
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens   int
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per window
// window: time window duration
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// Clean up old visitors every 5 minutes
	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, v := range rl.visitors {
		if now.Sub(v.lastSeen) > rl.window*2 {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			tokens:   rl.rate,
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
		return v
	}

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(v.lastSeen)
	if elapsed > rl.window {
		v.tokens = rl.rate
		v.lastSeen = now
	}

	return v
}

func (rl *RateLimiter) allow(ip string) bool {
	v := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if v.tokens > 0 {
		v.tokens--
		v.lastSeen = time.Now()
		return true
	}

	return false
}

// Middleware returns a rate limiting middleware
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)

		if !rl.allow(ip) {
			w.Header().Set("Retry-After", rl.window.String())
			http.Error(w, `{"error":"rate_limit_exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (if behind proxy/load balancer)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Get first IP in chain
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
