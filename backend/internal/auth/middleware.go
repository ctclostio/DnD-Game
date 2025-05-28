package auth

import (
	"context"
	"net/http"
	"strings"
)

// ContextKey represents the type for context keys
type ContextKey string

const (
	// UserContextKey is the key for user claims in request context
	UserContextKey ContextKey = "user_claims"
)

// Middleware provides authentication middleware functions
type Middleware struct {
	jwtManager *JWTManager
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware(jwtManager *JWTManager) *Middleware {
	return &Middleware{
		jwtManager: jwtManager,
	}
}

// Authenticate is a middleware that validates JWT tokens
func (m *Middleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from header
		authHeader := r.Header.Get("Authorization")
		token, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := m.jwtManager.ValidateToken(token, AccessToken)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OptionalAuthenticate is a middleware that validates JWT tokens if present but doesn't require them
func (m *Middleware) OptionalAuthenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from header if present
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			token, err := ExtractTokenFromHeader(authHeader)
			if err == nil {
				// Validate token but don't fail if invalid
				claims, err := m.jwtManager.ValidateToken(token, AccessToken)
				if err == nil {
					// Add claims to context if valid
					ctx := context.WithValue(r.Context(), UserContextKey, claims)
					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
	}
}

// RequireRole is a middleware that checks if the user has a specific role
func (m *Middleware) RequireRole(role string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.Authenticate(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !strings.EqualFold(claims.Role, role) {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireDM is a middleware that requires the user to be a Dungeon Master
func (m *Middleware) RequireDM() func(http.HandlerFunc) http.HandlerFunc {
	return m.RequireRole("dm")
}

// RequirePlayer is a middleware that requires the user to be a Player
func (m *Middleware) RequirePlayer() func(http.HandlerFunc) http.HandlerFunc {
	return m.RequireRole("player")
}

// GetUserFromContext retrieves user claims from the request context
func GetUserFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*Claims)
	return claims, ok
}

// GetUserIDFromContext is a helper to get just the user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := GetUserFromContext(ctx)
	if !ok {
		return "", false
	}
	return claims.UserID, true
}