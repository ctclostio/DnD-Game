package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders adds security headers to HTTP responses.
func SecurityHeaders(isDevelopment bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content Security Policy
			cspDirectives := []string{
				"default-src 'self'",
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'", // Allow inline scripts for development
				"style-src 'self' 'unsafe-inline'",                // Allow inline styles
				"img-src 'self' data: https:",
				"font-src 'self' data:",
				"connect-src 'self' ws: wss:",
				"frame-ancestors 'none'",
				"base-uri 'self'",
				"form-action 'self'",
			}

			// In production, tighten CSP
			if !isDevelopment {
				cspDirectives = []string{
					"default-src 'self'",
					"script-src 'self'",
					"style-src 'self'",
					"img-src 'self' data: https:",
					"font-src 'self'",
					"connect-src 'self' wss:",
					"frame-ancestors 'none'",
					"base-uri 'self'",
					"form-action 'self'",
					"upgrade-insecure-requests",
				}
			}

			w.Header().Set("Content-Security-Policy", strings.Join(cspDirectives, "; "))

			// Other security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			// HSTS (HTTP Strict Transport Security) - only in production
			if !isDevelopment {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			// Remove server header
			w.Header().Del("Server")

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateOrigin checks if the origin is allowed.
func ValidateOrigin(allowedOrigins []string, origin string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
		// Handle wildcard subdomains
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:]
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}
