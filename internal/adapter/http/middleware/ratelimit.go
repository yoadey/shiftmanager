package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// tokenBucket implements a simple token-bucket rate limiter per IP.
type tokenBucket struct {
	tokens    float64
	maxTokens float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

func newTokenBucket(rpm int) *tokenBucket {
	max := float64(rpm)
	return &tokenBucket{
		tokens:     max,
		maxTokens:  max,
		refillRate: max / 60.0,
		lastRefill: time.Now(),
	}
}

func (b *tokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// ipRateLimiter maps IPs to their token buckets.
type ipRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	rpm     int
}

func newIPRateLimiter(rpm int) *ipRateLimiter {
	rl := &ipRateLimiter{
		buckets: make(map[string]*tokenBucket),
		rpm:     rpm,
	}
	// Periodically evict stale buckets to prevent memory leak.
	go rl.evictLoop()
	return rl
}

func (rl *ipRateLimiter) get(ip string) *tokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b, ok := rl.buckets[ip]
	if !ok {
		b = newTokenBucket(rl.rpm)
		rl.buckets[ip] = b
	}
	return b
}

func (rl *ipRateLimiter) evictLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for ip, b := range rl.buckets {
			b.mu.Lock()
			stale := b.lastRefill.Before(cutoff)
			b.mu.Unlock()
			if stale {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns a middleware that limits requests per minute per IP.
// rpm = 0 disables rate limiting.
func RateLimit(rpm int) func(http.Handler) http.Handler {
	if rpm <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	limiter := newIPRateLimiter(rpm)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			// Respect X-Forwarded-For for reverse-proxy deployments.
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				// Use the first (client) IP.
				parts := splitComma(forwarded)
				if len(parts) > 0 {
					ip = trimSpace(parts[0])
				}
			}

			if !limiter.get(ip).Allow() {
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func splitComma(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func trimSpace(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	j := len(s)
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t') {
		j--
	}
	return s[i:j]
}
