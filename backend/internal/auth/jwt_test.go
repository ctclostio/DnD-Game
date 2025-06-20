package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testJWTSecret = "test-secret"
	testUserID    = "user-123"
	testEmail     = "test@example.com"
	testUsername  = "testuser"
	testRole      = "player"
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
	manager := NewJWTManager(testJWTSecret, 15*time.Minute, 24*time.Hour)

	userID := testUserID
	username := testUsername
	email := testEmail
	role := testRole

	tokenPair, err := manager.GenerateTokenPair(userID, username, email, role)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.NotEqual(t, tokenPair.AccessToken, tokenPair.RefreshToken)
	assert.Equal(t, int64(900), tokenPair.ExpiresIn) // 15 minutes in seconds
}

func TestValidateToken(t *testing.T) {
	manager := NewJWTManager(testJWTSecret, 15*time.Minute, 24*time.Hour)

	// Helper function to test token validation
	validateTokenHelper := func(t *testing.T, userID, username, email, role string, tokenType TokenType) {
		tokenPair, err := manager.GenerateTokenPair(userID, username, email, role)
		require.NoError(t, err)

		var token string
		if tokenType == AccessToken {
			token = tokenPair.AccessToken
		} else {
			token = tokenPair.RefreshToken
		}

		claims, err := manager.ValidateToken(token, tokenType)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, tokenType, claims.Type)
	}

	t.Run("valid access token", func(t *testing.T) {
		validateTokenHelper(t, testUserID, testUsername, testEmail, testRole, AccessToken)
	})

	t.Run("valid refresh token", func(t *testing.T) {
		validateTokenHelper(t, "user-456", "anotheruser", "another@example.com", "dm", RefreshToken)
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := manager.ValidateToken("invalid-token", AccessToken)
		assert.Error(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create manager with very short duration
		shortManager := NewJWTManager(testJWTSecret, 1*time.Millisecond, 1*time.Millisecond)

		tokenPair, err := shortManager.GenerateTokenPair(testUserID, testUsername, testEmail, testRole)
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
		tokenPair, err := manager1.GenerateTokenPair(testUserID, testUsername, testEmail, testRole)
		require.NoError(t, err)

		// Try to validate with different secret
		manager2 := NewJWTManager("secret2", 15*time.Minute, 24*time.Hour)
		_, err = manager2.ValidateToken(tokenPair.AccessToken, AccessToken)
		assert.Error(t, err)
	})

	t.Run("wrong token type", func(t *testing.T) {
		tokenPair, err := manager.GenerateTokenPair(testUserID, testUsername, testEmail, testRole)
		require.NoError(t, err)

		// Try to validate access token as refresh token
		_, err = manager.ValidateToken(tokenPair.AccessToken, RefreshToken)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTokenType, err)
	})
}

func TestRefreshAccessToken(t *testing.T) {
	manager := NewJWTManager(testJWTSecret, 15*time.Minute, 24*time.Hour)

	t.Run("valid refresh token", func(t *testing.T) {
		userID := testUserID
		username := testUsername
		email := testEmail
		role := testRole

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
		tokenPair, err := manager.GenerateTokenPair(testUserID, testUsername, testEmail, testRole)
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
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test" // NOSONAR - Mock JWT token for testing
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

func TestClaimsValidate(t *testing.T) {
	t.Run("valid claims", func(t *testing.T) {
		claims := &Claims{
			UserID: testUserID,
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
			UserID: testUserID,
			Type:   TokenType("invalid"),
		}
		err := claims.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token type")
	})
}

func TestNewClaims(t *testing.T) {
	userID := testUserID
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
	assert.NotEmpty(t, claims.ID)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)

	// Check expiration is set correctly
	expectedExpiry := time.Now().Add(duration)
	actualExpiry := claims.ExpiresAt.Time
	assert.WithinDuration(t, expectedExpiry, actualExpiry, 1*time.Second)
}
