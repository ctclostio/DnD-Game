package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
		// Skip caching for non-cacheable requests
		if !cm.isCacheableRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		cacheKey := cm.generateCacheKey(r)
		ctx := r.Context()

		// Try to serve from cache
		if served := cm.tryServeFromCache(w, r, ctx, cacheKey, next); served {
			return
		}

		// Handle cache miss
		cm.handleCacheMiss(w, r, ctx, cacheKey, next)
	})
}

// isCacheableRequest checks if request should be cached
func (cm *CacheMiddleware) isCacheableRequest(r *http.Request) bool {
	// Only cache GET and HEAD requests
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}

	// Check if route should be excluded
	return !cm.shouldExclude(r.URL.Path)
}

// tryServeFromCache attempts to serve response from cache
func (cm *CacheMiddleware) tryServeFromCache(w http.ResponseWriter, r *http.Request, ctx context.Context, cacheKey string, next http.Handler) bool {
	cached, err := cm.getFromCache(ctx, cacheKey)
	if err != nil || cached == nil {
		return false
	}

	age := time.Since(cached.CachedAt)
	
	// Serve fresh content
	if age <= cached.TTL {
		cm.serveCachedResponse(w, r, cached, false)
		return true
	}

	// Serve stale content while revalidating
	if age <= cached.TTL+cm.config.StaleWhileRevalidate {
		cm.serveCachedResponse(w, r, cached, true)
		go cm.revalidateInBackground(r, cacheKey, next)
		return true
	}

	return false
}

// handleCacheMiss handles requests when cache is missed
func (cm *CacheMiddleware) handleCacheMiss(w http.ResponseWriter, r *http.Request, ctx context.Context, cacheKey string, next http.Handler) {
	recorder := cm.createResponseRecorder(w)
	next.ServeHTTP(recorder, r)

	if !cm.shouldCache(recorder) {
		return
	}

	cm.cacheResponse(ctx, cacheKey, recorder)
}

// createResponseRecorder creates a new response recorder
func (cm *CacheMiddleware) createResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
		body:           &bytes.Buffer{},
	}
}

// cacheResponse saves the response to cache
func (cm *CacheMiddleware) cacheResponse(ctx context.Context, cacheKey string, recorder *responseRecorder) {
	ttl := cm.getTTL(recorder.headers)
	response := &CachedResponse{
		StatusCode: recorder.statusCode,
		Headers:    recorder.headers,
		Body:       recorder.body.Bytes(),
		CachedAt:   time.Now(),
		TTL:        ttl,
	}

	if err := cm.saveToCache(ctx, cacheKey, response, ttl); err != nil && cm.logger != nil {
		cm.logger.Error().
			Err(err).
			Str("cache_key", cacheKey).
			Msg("Failed to cache response")
	}
}

// generateCacheKey generates a cache key for the request
func (cm *CacheMiddleware) generateCacheKey(r *http.Request) string {
	parts := []string{
		r.Method,
		r.URL.Path,
	}

	// Add query parameters to key
	if queryPart := cm.buildQueryPart(r); queryPart != "" {
		parts = append(parts, queryPart)
	}

	// Add headers to key
	parts = append(parts, cm.buildHeaderParts(r)...)

	// Generate hash of all parts
	key := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(key))
	
	return "response:" + hex.EncodeToString(hash[:])
}

// buildQueryPart builds the query parameter part of the cache key
func (cm *CacheMiddleware) buildQueryPart(r *http.Request) string {
	if !cm.config.IncludeQuery || r.URL.RawQuery == "" {
		return ""
	}

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
	
	return strings.Join(queryParts, "&")
}

// buildHeaderParts builds the header parts of the cache key
func (cm *CacheMiddleware) buildHeaderParts(r *http.Request) []string {
	var headerParts []string
	
	for _, header := range cm.config.IncludeHeaders {
		if value := r.Header.Get(header); value != "" {
			headerParts = append(headerParts, fmt.Sprintf("%s:%s", header, value))
		}
	}
	
	return headerParts
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
	// Try to get TTL from Cache-Control header
	if ttl := cm.getTTLFromCacheControl(headers); ttl > 0 {
		return ttl
	}

	// Try to get TTL from Expires header
	if ttl := cm.getTTLFromExpires(headers); ttl > 0 {
		return ttl
	}

	return cm.config.DefaultTTL
}

// getTTLFromCacheControl extracts TTL from Cache-Control header
func (cm *CacheMiddleware) getTTLFromCacheControl(headers http.Header) time.Duration {
	cacheControl := headers.Get("Cache-Control")
	if cacheControl == "" {
		return 0
	}

	parts := strings.Split(cacheControl, ",")
	for _, part := range parts {
		if ttl := cm.parseMaxAge(strings.TrimSpace(part)); ttl > 0 {
			return ttl
		}
	}

	return 0
}

// parseMaxAge parses max-age directive from Cache-Control
func (cm *CacheMiddleware) parseMaxAge(directive string) time.Duration {
	if !strings.HasPrefix(directive, "max-age=") {
		return 0
	}

	seconds := strings.TrimPrefix(directive, "max-age=")
	if seconds == "" {
		return 0
	}

	var s int
	if _, err := fmt.Sscanf(seconds, "%d", &s); err != nil || s <= 0 {
		return 0
	}

	return time.Duration(s) * time.Second
}

// getTTLFromExpires extracts TTL from Expires header
func (cm *CacheMiddleware) getTTLFromExpires(headers http.Header) time.Duration {
	expires := headers.Get("Expires")
	if expires == "" {
		return 0
	}

	t, err := http.ParseTime(expires)
	if err != nil {
		return 0
	}

	ttl := time.Until(t)
	if ttl <= 0 {
		return 0
	}

	return ttl
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
		if _, err := w.Write(cached.Body); err != nil {
			// Log write error but don't fail - response is already committed
			if cm.logger != nil {
				cm.logger.Error().Err(err).Msg("Failed to write cached response body")
			}
		}
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
func (cm *CacheMiddleware) revalidateInBackground(r *http.Request, cacheKey string, next http.Handler) {
	// Clone the request for background processing
	req := r.Clone(context.Background())
	
	// Create a mock response writer to capture the response
	mock := &mockResponseWriter{
		headers: make(http.Header),
		body:    bytes.Buffer{},
		status:  http.StatusOK,
	}
	recorder := &responseRecorder{
		ResponseWriter: mock,
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
		body:           &bytes.Buffer{},
		written:        false,
	}

	// Execute the request against the actual handler
	next.ServeHTTP(recorder, req)
	
	// Check if we should cache the response
	if cm.shouldCache(recorder) {
		// Cache the response
		cm.cacheResponse(req.Context(), cacheKey, recorder)
		
		if cm.logger != nil {
			cm.logger.Debug().
				Str("path", r.URL.Path).
				Str("cache_key", cacheKey).
				Int("status", recorder.statusCode).
				Int("body_size", recorder.body.Len()).
				Msg("Background revalidation completed")
		}
	} else if cm.logger != nil {
		cm.logger.Debug().
			Str("path", r.URL.Path).
			Str("cache_key", cacheKey).
			Int("status", recorder.statusCode).
			Msg("Background revalidation completed - response not cached")
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
	patterns := []string{
		// Always invalidate the specific resource
		fmt.Sprintf("response:*%s*", path),
	}

	// Add endpoint-specific patterns
	patterns = append(patterns, im.getEndpointPatterns(path)...)

	return patterns
}

// getEndpointPatterns returns additional patterns based on the endpoint type
func (im *InvalidationMiddleware) getEndpointPatterns(path string) []string {
	var patterns []string

	// Define pattern rules for different endpoints
	patternRules := map[string][]string{
		"/characters":     {"response:*characters*"},
		"/game-sessions":  {"response:*game-sessions*", "response:*sessions*"},
		"/campaigns":      {"response:*campaigns*"},
		"/inventory":      {"response:*inventory*", "response:*items*"},
		"/users":          {"response:*users*"},
	}

	// Check each rule and add matching patterns
	for endpoint, endpointPatterns := range patternRules {
		if strings.Contains(path, endpoint) {
			patterns = append(patterns, endpointPatterns...)
			
			// For specific resource updates, also invalidate the list
			if strings.Contains(path, endpoint+"/") {
				patterns = append(patterns, fmt.Sprintf("response:*%s*", path))
			}
		}
	}

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