package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTManager(t *testing.T) {
	secret := "test-secret-key"
	accessDuration := 15 * time.Minute
	refreshDuration := 7 * 24 * time.Hour

	manager := NewJWTManager(secret, accessDuration, refreshDuration)

	assert.NotNil(t, manager)
	assert.Equal(t, secret, manager.secretKey)
	assert.Equal(t, accessDuration, manager.accessTokenDuration)
	assert.Equal(t, refreshDuration, manager.refreshTokenDuration)
}

func TestGenerateTokenPair(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour)

	userID := "user-123"
	username := "testuser"
	email := "test@example.com"
	role := "player"

	tokenPair, err := manager.GenerateTokenPair(userID, username, email, role)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.NotEqual(t, tokenPair.AccessToken, tokenPair.RefreshToken)
	assert.Equal(t, int64(900), tokenPair.ExpiresIn) // 15 minutes in seconds
}

func TestValidateToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour)

	t.Run("valid access token", func(t *testing.T) {
		userID := "user-123"
		username := "testuser"
		email := "test@example.com"
		role := "player"

		tokenPair, err := manager.GenerateTokenPair(userID, username, email, role)
		require.NoError(t, err)

		claims, err := manager.ValidateToken(tokenPair.AccessToken, AccessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, AccessToken, claims.Type)
	})

	t.Run("valid refresh token", func(t *testing.T) {
		userID := "user-456"
		username := "anotheruser"
		email := "another@example.com"
		role := "dm"

		tokenPair, err := manager.GenerateTokenPair(userID, username, email, role)
		require.NoError(t, err)

		claims, err := manager.ValidateToken(tokenPair.RefreshToken, RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, RefreshToken, claims.Type)
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := manager.ValidateToken("invalid-token", AccessToken)
		assert.Error(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create manager with very short duration
		shortManager := NewJWTManager("test-secret", 1*time.Millisecond, 1*time.Millisecond)

		tokenPair, err := shortManager.GenerateTokenPair("user-123", "testuser", "test@example.com", "player")
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		_, err = shortManager.ValidateToken(tokenPair.AccessToken, AccessToken)
		assert.Error(t, err)
		assert.Equal(t, ErrExpiredToken, err)
	})

	t.Run("wrong secret", func(t *testing.T) {
		// Generate token with one secret
		manager1 := NewJWTManager("secret1", 15*time.Minute, 24*time.Hour)
		tokenPair, err := manager1.GenerateTokenPair("user-123", "testuser", "test@example.com", "player")
		require.NoError(t, err)

		// Try to validate with different secret
		manager2 := NewJWTManager("secret2", 15*time.Minute, 24*time.Hour)
		_, err = manager2.ValidateToken(tokenPair.AccessToken, AccessToken)
		assert.Error(t, err)
	})

	t.Run("wrong token type", func(t *testing.T) {
		tokenPair, err := manager.GenerateTokenPair("user-123", "testuser", "test@example.com", "player")
		require.NoError(t, err)

		// Try to validate access token as refresh token
		_, err = manager.ValidateToken(tokenPair.AccessToken, RefreshToken)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTokenType, err)
	})
}

func TestRefreshAccessToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour)

	t.Run("valid refresh token", func(t *testing.T) {
		userID := "user-123"
		username := "testuser"
		email := "test@example.com"
		role := "player"

		tokenPair, err := manager.GenerateTokenPair(userID, username, email, role)
		require.NoError(t, err)

		newTokenPair, err := manager.RefreshAccessToken(tokenPair.RefreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, newTokenPair.AccessToken)
		assert.NotEmpty(t, newTokenPair.RefreshToken)

		// Validate new access token
		claims, err := manager.ValidateToken(newTokenPair.AccessToken, AccessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, AccessToken, claims.Type)
	})

	t.Run("with access token instead of refresh token", func(t *testing.T) {
		tokenPair, err := manager.GenerateTokenPair("user-123", "testuser", "test@example.com", "player")
		require.NoError(t, err)

		_, err = manager.RefreshAccessToken(tokenPair.AccessToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid refresh token")
	})

	t.Run("with invalid token", func(t *testing.T) {
		_, err := manager.RefreshAccessToken("invalid-token")
		assert.Error(t, err)
	})
}

func TestGenerateTokenID(t *testing.T) {
	// Test multiple times to ensure uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateTokenID()
		assert.NotEmpty(t, id)
		assert.False(t, ids[id], "Token ID should be unique")
		ids[id] = true
	}
}

func TestExtractTokenFromHeader(t *testing.T) {
	t.Run("valid bearer token", func(t *testing.T) {
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test"
		header := "Bearer " + token

		extracted, err := ExtractTokenFromHeader(header)
		require.NoError(t, err)
		assert.Equal(t, token, extracted)
	})

	t.Run("empty header", func(t *testing.T) {
		_, err := ExtractTokenFromHeader("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authorization header is required")
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := ExtractTokenFromHeader("InvalidFormat token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid authorization header format")
	})

	t.Run("missing bearer prefix", func(t *testing.T) {
		_, err := ExtractTokenFromHeader("token-without-bearer")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid authorization header format")
	})
}

func TestClaims_Validate(t *testing.T) {
	t.Run("valid claims", func(t *testing.T) {
		claims := &Claims{
			UserID: "user-123",
			Type:   AccessToken,
		}
		err := claims.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing user ID", func(t *testing.T) {
		claims := &Claims{
			Type: AccessToken,
		}
		err := claims.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user_id is required")
	})

	t.Run("invalid token type", func(t *testing.T) {
		claims := &Claims{
			UserID: "user-123",
			Type:   TokenType("invalid"),
		}
		err := claims.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token type")
	})
}

func TestNewClaims(t *testing.T) {
	userID := "user-123"
	username := "testuser"
	email := "test@example.com"
	role := "player"
	tokenType := AccessToken
	duration := 15 * time.Minute

	claims := NewClaims(userID, username, email, role, tokenType, duration)

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, tokenType, claims.Type)
	assert.NotEmpty(t, claims.RegisteredClaims.ID)
	assert.NotNil(t, claims.RegisteredClaims.ExpiresAt)
	assert.NotNil(t, claims.RegisteredClaims.IssuedAt)
	assert.NotNil(t, claims.RegisteredClaims.NotBefore)

	// Check expiration is set correctly
	expectedExpiry := time.Now().Add(duration)
	actualExpiry := claims.RegisteredClaims.ExpiresAt.Time
	assert.WithinDuration(t, expectedExpiry, actualExpiry, 1*time.Second)
}
