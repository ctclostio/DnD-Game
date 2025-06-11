package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	// AccessToken is used for API authentication
	AccessToken TokenType = "access"
	// RefreshToken is used to refresh access tokens
	RefreshToken TokenType = "refresh"
)

// Claims represents the custom JWT claims for our application
type Claims struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"` // "player" or "dm"
	Type     TokenType `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// NewClaims creates a new Claims instance
func NewClaims(userID, username, email, role string, tokenType TokenType, duration time.Duration) *Claims {
	now := time.Now()
	return &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        GenerateTokenID(),
		},
	}
}

// Valid validates the claims
func (c *Claims) Validate() error {
	// Check custom claims
	if c.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if c.Type != AccessToken && c.Type != RefreshToken {
		return fmt.Errorf("invalid token type")
	}

	return nil
}

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expires
}
