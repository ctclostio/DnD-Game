package testhelpers

import (
	"context"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/google/uuid"
)

// CreateTestClaims creates test JWT claims with sensible defaults
func CreateTestClaims(userID, username, email, role string) *auth.Claims {
	return auth.NewClaims(userID, username, email, role, auth.AccessToken, time.Hour)
}

// CreateAuthContext creates a context with authentication claims
func CreateAuthContext(userID, role string) context.Context {
	claims := CreateTestClaims(userID, "testuser", "test@example.com", role)
	return context.WithValue(context.Background(), auth.UserContextKey, claims)
}

// CreateAuthContextWithClaims creates a context with custom claims
func CreateAuthContextWithClaims(claims *auth.Claims) context.Context {
	return context.WithValue(context.Background(), auth.UserContextKey, claims)
}

// CreateDMContext creates a context for a DM user
func CreateDMContext(userID string) context.Context {
	return CreateAuthContext(userID, "dm")
}

// CreatePlayerContext creates a context for a player user
func CreatePlayerContext(userID string) context.Context {
	return CreateAuthContext(userID, "player")
}

// CreateAdminContext creates a context for an admin user
func CreateAdminContext(userID string) context.Context {
	return CreateAuthContext(userID, "admin")
}

// TestUser represents a test user with commonly needed fields
type TestUser struct {
	ID       string
	Username string
	Email    string
	Role     string
	Claims   *auth.Claims
}

// NewTestUser creates a new test user with generated ID
func NewTestUser(role string) *TestUser {
	userID := uuid.New().String()
	username := "testuser_" + userID[:8]
	email := username + "@example.com"

	user := &TestUser{
		ID:       userID,
		Username: username,
		Email:    email,
		Role:     role,
	}

	user.Claims = CreateTestClaims(userID, username, email, role)
	return user
}

// NewTestDM creates a test DM user
func NewTestDM() *TestUser {
	return NewTestUser("dm")
}

// NewTestPlayer creates a test player user
func NewTestPlayer() *TestUser {
	return NewTestUser("player")
}

// GetContext returns a context with this user's claims
func (u *TestUser) GetContext() context.Context {
	return CreateAuthContextWithClaims(u.Claims)
}

// ExtractUserID extracts user ID from context (helper for tests)
func ExtractUserID(ctx context.Context) (string, bool) {
	claims, ok := ctx.Value(auth.UserContextKey).(*auth.Claims)
	if !ok || claims == nil {
		return "", false
	}
	return claims.UserID, true
}

// ExtractClaims extracts claims from context (helper for tests)
func ExtractClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(auth.UserContextKey).(*auth.Claims)
	return claims, ok
}
