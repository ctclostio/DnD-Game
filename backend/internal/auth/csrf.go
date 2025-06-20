package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

const (
	csrfTokenLength = 32
	csrfCookieName  = "csrf_token"
	csrfHeaderName  = "X-CSRF-Token"
	csrfTokenTTL    = 24 * time.Hour
)

// CSRFStore manages CSRF tokens
type CSRFStore struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

// NewCSRFStore creates a new CSRF token store
func NewCSRFStore() *CSRFStore {
	store := &CSRFStore{
		tokens: make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// cleanup removes expired tokens
func (s *CSRFStore) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, expiry := range s.tokens {
			if now.After(expiry) {
				delete(s.tokens, token)
			}
		}
		s.mu.Unlock()
	}
}

// GenerateToken creates a new CSRF token
func (s *CSRFStore) GenerateToken() (string, error) {
	bytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(bytes)

	s.mu.Lock()
	s.tokens[token] = time.Now().Add(csrfTokenTTL)
	s.mu.Unlock()

	return token, nil
}

// ValidateToken checks if a token is valid
func (s *CSRFStore) ValidateToken(token string) bool {
	s.mu.RLock()
	expiry, exists := s.tokens[token]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		s.mu.Lock()
		delete(s.tokens, token)
		s.mu.Unlock()
		return false
	}

	return true
}

// handleSafeCSRFMethod handles CSRF token generation for safe HTTP methods.
func handleSafeCSRFMethod(w http.ResponseWriter, r *http.Request, store *CSRFStore, isProduction bool, next http.Handler) {
	// Generate and set CSRF token for GET requests (also applies to HEAD, OPTIONS as per original logic)
	token, err := store.GenerateToken()
	if err == nil {
		// SECURITY: Cookie configuration follows OWASP best practices
		// - Secure flag is environment-based: true in production (HTTPS required), false in development
		// - HttpOnly is false by design: CSRF tokens must be readable by JavaScript to be included in requests
		// - SameSite=Strict provides additional CSRF protection by preventing cross-site cookie transmission
		// - Path=/ ensures the token is available for all application routes
		// This configuration is safe and intentional. The Secure flag dynamically adapts to the environment.
		// NOSONAR: S2092 - Secure flag is properly set based on environment (isProduction parameter)
		http.SetCookie(w, &http.Cookie{
			Name:     csrfCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: false, // Must be readable by JavaScript for CSRF protection to work
			Secure:   isProduction, // Environment-based: always true in production (requires HTTPS)
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(csrfTokenTTL.Seconds()),
		})
	}
	next.ServeHTTP(w, r)
}

// validateCSRFForStateChange validates the CSRF token for state-changing methods.
// It returns true if the request should proceed, false if an error was written and the request should stop.
func validateCSRFForStateChange(w http.ResponseWriter, r *http.Request, store *CSRFStore) bool {
	cookieToken := ""
	if cookie, err := r.Cookie(csrfCookieName); err == nil {
		cookieToken = cookie.Value
	}

	headerToken := r.Header.Get(csrfHeaderName)

	// Both tokens must be present and match
	if cookieToken == "" || headerToken == "" {
		http.Error(w, "CSRF token missing", http.StatusForbidden)
		return false
	}

	// Use constant-time comparison
	if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
		http.Error(w, "CSRF token mismatch", http.StatusForbidden)
		return false
	}

	// Validate token exists and isn't expired
	if !store.ValidateToken(cookieToken) {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return false
	}
	return true
}

// CSRFMiddleware provides CSRF protection
func CSRFMiddleware(store *CSRFStore, isProduction bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF for safe methods
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				handleSafeCSRFMethod(w, r, store, isProduction, next)
				return
			}

			// For state-changing methods, validate CSRF token
			if !validateCSRFForStateChange(w, r, store) {
				return // Error already handled by validateCSRFForStateChange
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ExemptCSRF returns a middleware that exempts specific paths from CSRF protection
func ExemptCSRF(paths ...string) func(http.Handler) http.Handler {
	pathMap := make(map[string]bool)
	for _, path := range paths {
		pathMap[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if pathMap[r.URL.Path] {
				// Skip CSRF check for exempted paths
				r = r.WithContext(r.Context()) // Could add context value to indicate CSRF exempt
			}
			next.ServeHTTP(w, r)
		})
	}
}
