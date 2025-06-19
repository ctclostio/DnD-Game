package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCSRFCookieSecurityFlags validates that CSRF cookies have proper security flags
func TestCSRFCookieSecurityFlags(t *testing.T) {
	store := NewCSRFStore()
	
	tests := []struct {
		name           string
		isProduction   bool
		expectedSecure bool
		description    string
	}{
		{
			name:           "Production environment sets Secure flag",
			isProduction:   true,
			expectedSecure: true,
			description:    "In production, cookies must have Secure=true to ensure HTTPS-only transmission",
		},
		{
			name:           "Development environment allows non-secure",
			isProduction:   false,
			expectedSecure: false,
			description:    "In development, Secure=false allows local HTTP testing",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			
			// Wrap with CSRF middleware
			csrfHandler := CSRFMiddleware(store, tt.isProduction)(handler)
			
			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			rec := httptest.NewRecorder()
			
			// Execute request
			csrfHandler.ServeHTTP(rec, req)
			
			// Check cookie was set
			cookies := rec.Result().Cookies()
			if len(cookies) == 0 {
				t.Fatal("Expected CSRF cookie to be set")
			}
			
			// Find CSRF cookie
			var csrfCookie *http.Cookie
			for _, cookie := range cookies {
				if cookie.Name == "csrf_token" {
					csrfCookie = cookie
					break
				}
			}
			
			if csrfCookie == nil {
				t.Fatal("CSRF cookie not found")
			}
			
			// Validate security flags
			if csrfCookie.Secure != tt.expectedSecure {
				t.Errorf("Secure flag mismatch: got %v, want %v", csrfCookie.Secure, tt.expectedSecure)
			}
			
			// Validate other security requirements
			if csrfCookie.HttpOnly != false {
				t.Error("HttpOnly must be false for CSRF tokens to be accessible by JavaScript")
			}
			
			if csrfCookie.SameSite != http.SameSiteStrictMode {
				t.Errorf("SameSite should be Strict, got %v", csrfCookie.SameSite)
			}
			
			if csrfCookie.Path != "/" {
				t.Errorf("Path should be /, got %s", csrfCookie.Path)
			}
			
			// Log security configuration for clarity
			t.Logf("Cookie security configuration - Secure: %v, HttpOnly: %v, SameSite: %v",
				csrfCookie.Secure, csrfCookie.HttpOnly, csrfCookie.SameSite)
		})
	}
}

// TestCSRFCookieSecurityHeaders validates security headers in responses
func TestCSRFCookieSecurityHeaders(t *testing.T) {
	store := NewCSRFStore()
	
	// Test production configuration
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	csrfHandler := CSRFMiddleware(store, true)(handler)
	
	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	
	csrfHandler.ServeHTTP(rec, req)
	
	// Verify Set-Cookie header contains security attributes
	setCookieHeader := rec.Header().Get("Set-Cookie")
	if setCookieHeader == "" {
		t.Fatal("Expected Set-Cookie header")
	}
	
	// In production, must have Secure attribute
	if !strings.Contains(setCookieHeader, "Secure") {
		t.Error("Production cookie must have Secure attribute in Set-Cookie header")
	}
	
	// Must have SameSite=Strict
	if !strings.Contains(setCookieHeader, "SameSite=Strict") {
		t.Error("Cookie must have SameSite=Strict attribute")
	}
	
	// Should NOT have HttpOnly (CSRF tokens need JS access)
	if strings.Contains(setCookieHeader, "HttpOnly") {
		t.Error("CSRF cookie should not have HttpOnly attribute")
	}
}

// TestCSRFTokenValidation ensures CSRF validation works correctly
func TestCSRFTokenValidation(t *testing.T) {
	store := NewCSRFStore()
	
	// Generate a token
	token, err := store.GenerateToken()
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	// Validate it exists
	if !store.ValidateToken(token) {
		t.Error("Generated token should be valid")
	}
	
	// Validate non-existent token fails
	if store.ValidateToken("invalid-token") {
		t.Error("Invalid token should not validate")
	}
}