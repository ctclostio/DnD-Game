package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter tracks request rates per IP.
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

// visitor tracks the rate limit data for each visitor.
type visitor struct {
	lastSeen  time.Time
	count     int
	blocked   bool
	blockTime time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// cleanupVisitors removes old entries from the visitors map.
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getVisitor retrieves or creates a visitor entry.
func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			lastSeen: time.Now(),
			count:    0,
		}
		rl.visitors[ip] = v
	}

	return v
}

// isAllowed checks if a request from the given IP is allowed.
func (rl *RateLimiter) isAllowed(ip string) bool {
	v := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Check if visitor is blocked
	if v.blocked {
		// Check if block period has expired (exponential backoff)
		if now.Sub(v.blockTime) < rl.window*10 {
			return false
		}
		// Unblock
		v.blocked = false
		v.count = 0
	}

	// Reset count if window has passed
	if now.Sub(v.lastSeen) > rl.window {
		v.count = 0
		v.lastSeen = now
	}

	// Increment count
	v.count++

	// Check if limit exceeded
	if v.count > rl.rate {
		v.blocked = true
		v.blockTime = now
		return false
	}

	v.lastSeen = now
	return true
}

// Middleware returns a rate limiting middleware.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)

			if !rl.isAllowed(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.rate))
				w.Header().Set("X-RateLimit-Window", rl.window.String())
				w.Header().Set("Retry-After", rl.window.String())
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, `{"error":"Rate limit exceeded. Please try again later."}`)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getIP extracts the real IP address from the request.
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Get the first IP in the chain
		if ip := net.ParseIP(forwarded); ip != nil {
			return ip.String()
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		if ip := net.ParseIP(realIP); ip != nil {
			return ip.String()
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// AuthRateLimiter creates a rate limiter specifically for auth endpoints.
func AuthRateLimiter() *RateLimiter {
	// Allow 5 requests per minute for auth endpoints
	return NewRateLimiter(5, 1*time.Minute)
}

// APIRateLimiter creates a general rate limiter for API endpoints.
func APIRateLimiter() *RateLimiter {
	// Allow 100 requests per minute for general API endpoints
	return NewRateLimiter(100, 1*time.Minute)
}
