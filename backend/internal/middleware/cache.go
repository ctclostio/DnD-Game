package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/cache"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// CacheConfig contains cache middleware configuration
type CacheConfig struct {
	// DefaultTTL is the default cache duration
	DefaultTTL time.Duration
	
	// MaxBodySize is the maximum response body size to cache (in bytes)
	MaxBodySize int64
	
	// IncludeQuery determines if query parameters are included in cache key
	IncludeQuery bool
	
	// IncludeHeaders lists headers to include in cache key
	IncludeHeaders []string
	
	// ExcludeRoutes lists route patterns to exclude from caching
	ExcludeRoutes []string
	
	// CacheableStatusCodes lists which status codes to cache
	CacheableStatusCodes []int
	
	// StaleWhileRevalidate allows serving stale content while revalidating
	StaleWhileRevalidate time.Duration
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		DefaultTTL:           5 * time.Minute,
		MaxBodySize:          1 * 1024 * 1024, // 1MB
		IncludeQuery:         true,
		IncludeHeaders:       []string{"Accept", "Accept-Language"},
		ExcludeRoutes:        []string{"/api/auth/", "/api/ws/", "/health"},
		CacheableStatusCodes: []int{http.StatusOK, http.StatusNoContent},
		StaleWhileRevalidate: 1 * time.Minute,
	}
}

// CacheMiddleware provides HTTP response caching
type CacheMiddleware struct {
	cache  *cache.Cache
	config *CacheConfig
	logger *logger.LoggerV2
}

// NewCacheMiddleware creates a new cache middleware
func NewCacheMiddleware(c *cache.Cache, cfg *CacheConfig, log *logger.LoggerV2) *CacheMiddleware {
	if cfg == nil {
		cfg = DefaultCacheConfig()
	}
	
	return &CacheMiddleware{
		cache:  c,
		config: cfg,
		logger: log,
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
	CachedAt   time.Time           `json:"cached_at"`
	TTL        time.Duration       `json:"ttl"`
}

// Handler returns the cache middleware handler
func (cm *CacheMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only cache GET and HEAD requests
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		// Check if route should be excluded
		if cm.shouldExclude(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Generate cache key
		cacheKey := cm.generateCacheKey(r)

		// Try to get from cache
		ctx := r.Context()
		cached, err := cm.getFromCache(ctx, cacheKey)
		if err == nil && cached != nil {
			// Check if still fresh
			if time.Since(cached.CachedAt) <= cached.TTL {
				cm.serveCachedResponse(w, r, cached, false)
				return
			}

			// Check if we can serve stale while revalidating
			if time.Since(cached.CachedAt) <= cached.TTL+cm.config.StaleWhileRevalidate {
				cm.serveCachedResponse(w, r, cached, true)
				
				// Revalidate in background
				go cm.revalidateInBackground(r, cacheKey)
				return
			}
		}

		// Cache miss or expired - capture response
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			headers:        make(http.Header),
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(recorder, r)

		// Cache the response if appropriate
		if cm.shouldCache(recorder) {
			ttl := cm.getTTL(recorder.headers)
			response := &CachedResponse{
				StatusCode: recorder.statusCode,
				Headers:    recorder.headers,
				Body:       recorder.body.Bytes(),
				CachedAt:   time.Now(),
				TTL:        ttl,
			}

			if err := cm.saveToCache(ctx, cacheKey, response, ttl); err != nil {
				if cm.logger != nil {
					cm.logger.Error().
						Err(err).
						Str("cache_key", cacheKey).
						Msg("Failed to cache response")
				}
			}
		}
	})
}

// generateCacheKey generates a cache key for the request
func (cm *CacheMiddleware) generateCacheKey(r *http.Request) string {
	parts := []string{
		r.Method,
		r.URL.Path,
	}

	// Include query parameters if configured
	if cm.config.IncludeQuery && r.URL.RawQuery != "" {
		// Sort query parameters for consistent keys
		query := r.URL.Query()
		keys := make([]string, 0, len(query))
		for k := range query {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var queryParts []string
		for _, k := range keys {
			for _, v := range query[k] {
				queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, v))
			}
		}
		parts = append(parts, strings.Join(queryParts, "&"))
	}

	// Include specified headers
	for _, header := range cm.config.IncludeHeaders {
		if value := r.Header.Get(header); value != "" {
			parts = append(parts, fmt.Sprintf("%s:%s", header, value))
		}
	}

	// Generate hash of all parts
	key := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(key))
	
	return "response:" + hex.EncodeToString(hash[:])
}

// shouldExclude checks if the route should be excluded from caching
func (cm *CacheMiddleware) shouldExclude(path string) bool {
	for _, pattern := range cm.config.ExcludeRoutes {
		if strings.HasPrefix(path, pattern) {
			return true
		}
	}
	return false
}

// shouldCache determines if the response should be cached
func (cm *CacheMiddleware) shouldCache(rec *responseRecorder) bool {
	// Check status code
	validStatus := false
	for _, code := range cm.config.CacheableStatusCodes {
		if rec.statusCode == code {
			validStatus = true
			break
		}
	}
	if !validStatus {
		return false
	}

	// Check response size
	if int64(rec.body.Len()) > cm.config.MaxBodySize {
		return false
	}

	// Check cache control headers
	cacheControl := rec.headers.Get("Cache-Control")
	if strings.Contains(cacheControl, "no-cache") || strings.Contains(cacheControl, "no-store") {
		return false
	}

	return true
}

// getTTL extracts TTL from response headers or uses default
func (cm *CacheMiddleware) getTTL(headers http.Header) time.Duration {
	// Check Cache-Control max-age
	cacheControl := headers.Get("Cache-Control")
	if cacheControl != "" {
		parts := strings.Split(cacheControl, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "max-age=") {
				if seconds := strings.TrimPrefix(part, "max-age="); seconds != "" {
					var s int
					if _, err := fmt.Sscanf(seconds, "%d", &s); err == nil && s > 0 {
						return time.Duration(s) * time.Second
					}
				}
			}
		}
	}

	// Check Expires header
	if expires := headers.Get("Expires"); expires != "" {
		if t, err := http.ParseTime(expires); err == nil {
			ttl := time.Until(t)
			if ttl > 0 {
				return ttl
			}
		}
	}

	return cm.config.DefaultTTL
}

// getFromCache retrieves a response from cache
func (cm *CacheMiddleware) getFromCache(ctx context.Context, key string) (*CachedResponse, error) {
	var response CachedResponse
	err := cm.cache.GetJSON(ctx, key, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// saveToCache saves a response to cache
func (cm *CacheMiddleware) saveToCache(ctx context.Context, key string, response *CachedResponse, ttl time.Duration) error {
	// Add extra time for stale-while-revalidate
	cacheTTL := ttl + cm.config.StaleWhileRevalidate
	return cm.cache.SetJSON(ctx, key, response, cacheTTL)
}

// serveCachedResponse writes the cached response to the client
func (cm *CacheMiddleware) serveCachedResponse(w http.ResponseWriter, r *http.Request, cached *CachedResponse, stale bool) {
	// Copy headers
	for k, v := range cached.Headers {
		w.Header()[k] = v
	}

	// Add cache headers
	w.Header().Set("X-Cache", "HIT")
	w.Header().Set("X-Cache-Key", cm.generateCacheKey(r))
	w.Header().Set("Age", fmt.Sprintf("%.0f", time.Since(cached.CachedAt).Seconds()))

	if stale {
		w.Header().Set("X-Cache-Status", "STALE")
		w.Header().Set("Warning", `110 - "Response is stale"`)
	}

	// Write status code
	w.WriteHeader(cached.StatusCode)

	// Write body (unless HEAD request)
	if r.Method != http.MethodHead {
		w.Write(cached.Body)
	}

	// Log cache hit
	if cm.logger != nil {
		cm.logger.Debug().
			Str("path", r.URL.Path).
			Str("cache_status", "HIT").
			Bool("stale", stale).
			Dur("age", time.Since(cached.CachedAt)).
			Msg("Served from cache")
	}
}

// revalidateInBackground refreshes cache in the background
func (cm *CacheMiddleware) revalidateInBackground(r *http.Request, cacheKey string) {
	// Clone the request for background processing
	req := r.Clone(context.Background())
	
	// Create a mock response writer to capture the response
	recorder := &responseRecorder{
		ResponseWriter: &mockResponseWriter{},
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
		body:           &bytes.Buffer{},
	}

	// TODO: Execute the request against the actual handler
	// This would require access to the handler chain
	
	if cm.logger != nil {
		cm.logger.Debug().
			Str("path", r.URL.Path).
			Str("cache_key", cacheKey).
			Msg("Background revalidation started")
	}
}

// responseRecorder captures the response for caching
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       *bytes.Buffer
	written    bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.written {
		r.statusCode = code
		// Copy headers
		for k, v := range r.ResponseWriter.Header() {
			r.headers[k] = v
		}
		r.ResponseWriter.WriteHeader(code)
		r.written = true
	}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// mockResponseWriter is used for background revalidation
type mockResponseWriter struct {
	headers http.Header
	body    bytes.Buffer
	status  int
}

func (m *mockResponseWriter) Header() http.Header {
	if m.headers == nil {
		m.headers = make(http.Header)
	}
	return m.headers
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	return m.body.Write(b)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

// InvalidationMiddleware handles cache invalidation
type InvalidationMiddleware struct {
	cache  *cache.Cache
	logger *logger.LoggerV2
}

// NewInvalidationMiddleware creates a new invalidation middleware
func NewInvalidationMiddleware(c *cache.Cache, log *logger.LoggerV2) *InvalidationMiddleware {
	return &InvalidationMiddleware{
		cache:  c,
		logger: log,
	}
}

// Handler returns the invalidation middleware handler
func (im *InvalidationMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only invalidate on mutating methods
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Capture response to check status
		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		// Invalidate cache on successful mutations
		if recorder.statusCode >= 200 && recorder.statusCode < 300 {
			im.invalidateRelatedCache(r)
		}
	})
}

// invalidateRelatedCache invalidates cache entries related to the request
func (im *InvalidationMiddleware) invalidateRelatedCache(r *http.Request) {
	ctx := r.Context()
	path := r.URL.Path

	// Define invalidation patterns based on routes
	patterns := im.getInvalidationPatterns(r.Method, path)

	for _, pattern := range patterns {
		if err := im.cache.Invalidate(ctx, pattern); err != nil {
			if im.logger != nil {
				im.logger.Error().
					Err(err).
					Str("pattern", pattern).
					Msg("Failed to invalidate cache")
			}
		} else if im.logger != nil {
			im.logger.Debug().
				Str("method", r.Method).
				Str("path", path).
				Str("pattern", pattern).
				Msg("Cache invalidated")
		}
	}
}

// getInvalidationPatterns returns cache patterns to invalidate based on the request
func (im *InvalidationMiddleware) getInvalidationPatterns(method, path string) []string {
	var patterns []string

	// Character-related endpoints
	if strings.Contains(path, "/characters") {
		if strings.HasSuffix(path, "/characters") {
			// Creating a new character - invalidate list
			patterns = append(patterns, "response:*characters*")
		} else if matched := strings.Contains(path, "/characters/"); matched {
			// Updating specific character - invalidate that character and lists
			patterns = append(patterns, 
				fmt.Sprintf("response:*%s*", path),
				"response:*characters*",
			)
		}
	}

	// Game session endpoints
	if strings.Contains(path, "/game-sessions") {
		patterns = append(patterns,
			"response:*game-sessions*",
			"response:*sessions*",
		)
	}

	// Generic pattern for the specific resource
	patterns = append(patterns, fmt.Sprintf("response:*%s*", path))

	return patterns
}

// statusRecorder captures just the status code
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if !s.written {
		s.statusCode = code
		s.ResponseWriter.WriteHeader(code)
		s.written = true
	}
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if !s.written {
		s.WriteHeader(http.StatusOK)
	}
	return s.ResponseWriter.Write(b)
}