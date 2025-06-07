package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

func TestRefreshTokenService_Create(t *testing.T) {
	t.Run("successful token storage", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		userID := "user-123"
		refreshToken := "valid.refresh.token"
		expiresAt := time.Now().Add(24 * time.Hour)
		
		claims := &auth.Claims{
			UserID:   userID,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "player",
			Type:     auth.RefreshToken,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        "token-123",
				ExpiresAt: jwt.NewNumericDate(expiresAt),
			},
		}
		
		mockJWT.On("ValidateToken", refreshToken, auth.RefreshToken).Return(claims, nil)
		mockRepo.On("Create", userID, "token-123", refreshToken, mock.MatchedBy(func(t time.Time) bool {
			return t.Equal(expiresAt.Round(time.Second))
		})).Return(nil)
		
		err := service.Create(userID, refreshToken)
		
		require.NoError(t, err)
		mockJWT.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		userID := "user-123"
		refreshToken := "invalid.refresh.token"
		
		mockJWT.On("ValidateToken", refreshToken, auth.RefreshToken).
			Return(nil, auth.ErrInvalidToken)
		
		err := service.Create(userID, refreshToken)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid refresh token")
		mockJWT.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("expired token", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		userID := "user-123"
		refreshToken := "expired.refresh.token"
		
		mockJWT.On("ValidateToken", refreshToken, auth.RefreshToken).
			Return(nil, auth.ErrExpiredToken)
		
		err := service.Create(userID, refreshToken)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid refresh token")
		mockJWT.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		userID := "user-123"
		refreshToken := "valid.refresh.token"
		expiresAt := time.Now().Add(24 * time.Hour)
		
		claims := &auth.Claims{
			UserID:   userID,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "player",
			Type:     auth.RefreshToken,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        "token-123",
				ExpiresAt: jwt.NewNumericDate(expiresAt),
			},
		}
		
		mockJWT.On("ValidateToken", refreshToken, auth.RefreshToken).Return(claims, nil)
		mockRepo.On("Create", userID, "token-123", refreshToken, mock.Anything).
			Return(errors.New("database error"))
		
		err := service.Create(userID, refreshToken)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "database error")
		mockJWT.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestRefreshTokenService_RefreshAccessToken(t *testing.T) {
	t.Run("successful token refresh", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		refreshToken := "valid.refresh.token"
		userID := "user-123"
		
		storedToken := &database.RefreshToken{
			ID:        "1",
			UserID:    userID,
			TokenID:   "token-123",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		
		claims := &auth.Claims{
			UserID:   userID,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "player",
			Type:     auth.RefreshToken,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        "token-123",
				ExpiresAt: jwt.NewNumericDate(storedToken.ExpiresAt),
			},
		}
		
		newTokenPair := &auth.TokenPair{
			AccessToken:  "new.access.token",
			RefreshToken: "new.refresh.token",
			ExpiresIn:    3600,
		}
		
		mockRepo.On("ValidateAndGet", refreshToken).Return(storedToken, nil)
		mockJWT.On("ValidateToken", refreshToken, auth.RefreshToken).Return(claims, nil)
		mockRepo.On("Revoke", "token-123").Return(nil)
		mockJWT.On("GenerateTokenPair", userID, "testuser", "test@example.com", "player").
			Return(newTokenPair, nil)
		
		result, returnedUserID, err := service.RefreshAccessToken(refreshToken)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, newTokenPair, result)
		require.Equal(t, userID, returnedUserID)
		mockRepo.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
	})

	t.Run("token not found in database", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		refreshToken := "invalid.refresh.token"
		
		mockRepo.On("ValidateAndGet", refreshToken).
			Return(nil, errors.New("invalid or expired refresh token"))
		
		result, userID, err := service.RefreshAccessToken(refreshToken)
		
		require.Error(t, err)
		require.Nil(t, result)
		require.Empty(t, userID)
		require.Contains(t, err.Error(), "invalid or expired refresh token")
		mockRepo.AssertExpectations(t)
		mockJWT.AssertNotCalled(t, "ValidateToken")
	})

	t.Run("invalid JWT", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		refreshToken := "invalid.jwt.token"
		userID := "user-123"
		
		storedToken := &database.RefreshToken{
			ID:        "1",
			UserID:    userID,
			TokenID:   "token-123",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		
		mockRepo.On("ValidateAndGet", refreshToken).Return(storedToken, nil)
		mockJWT.On("ValidateToken", refreshToken, auth.RefreshToken).
			Return(nil, auth.ErrInvalidToken)
		
		result, returnedUserID, err := service.RefreshAccessToken(refreshToken)
		
		require.Error(t, err)
		require.Nil(t, result)
		require.Empty(t, returnedUserID)
		mockRepo.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
	})

	t.Run("revoke old token failure (non-fatal)", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		refreshToken := "valid.refresh.token"
		userID := "user-123"
		
		storedToken := &database.RefreshToken{
			ID:        "1",
			UserID:    userID,
			TokenID:   "token-123",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		
		claims := &auth.Claims{
			UserID:   userID,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "player",
			Type:     auth.RefreshToken,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        "token-123",
				ExpiresAt: jwt.NewNumericDate(storedToken.ExpiresAt),
			},
		}
		
		newTokenPair := &auth.TokenPair{
			AccessToken:  "new.access.token",
			RefreshToken: "new.refresh.token",
			ExpiresIn:    3600,
		}
		
		mockRepo.On("ValidateAndGet", refreshToken).Return(storedToken, nil)
		mockJWT.On("ValidateToken", refreshToken, auth.RefreshToken).Return(claims, nil)
		mockRepo.On("Revoke", "token-123").Return(errors.New("revoke failed"))
		mockJWT.On("GenerateTokenPair", userID, "testuser", "test@example.com", "player").
			Return(newTokenPair, nil)
		
		// Should still succeed despite revoke failure
		result, returnedUserID, err := service.RefreshAccessToken(refreshToken)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, newTokenPair, result)
		require.Equal(t, userID, returnedUserID)
		mockRepo.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
	})

	t.Run("generate new tokens failure", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		refreshToken := "valid.refresh.token"
		userID := "user-123"
		
		storedToken := &database.RefreshToken{
			ID:        "1",
			UserID:    userID,
			TokenID:   "token-123",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		
		claims := &auth.Claims{
			UserID:   userID,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "player",
			Type:     auth.RefreshToken,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        "token-123",
				ExpiresAt: jwt.NewNumericDate(storedToken.ExpiresAt),
			},
		}
		
		mockRepo.On("ValidateAndGet", refreshToken).Return(storedToken, nil)
		mockJWT.On("ValidateToken", refreshToken, auth.RefreshToken).Return(claims, nil)
		mockRepo.On("Revoke", "token-123").Return(nil)
		mockJWT.On("GenerateTokenPair", userID, "testuser", "test@example.com", "player").
			Return(nil, errors.New("failed to generate tokens"))
		
		result, returnedUserID, err := service.RefreshAccessToken(refreshToken)
		
		require.Error(t, err)
		require.Nil(t, result)
		require.Empty(t, returnedUserID)
		require.Contains(t, err.Error(), "failed to generate new tokens")
		mockRepo.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
	})
}

func TestRefreshTokenService_Revoke(t *testing.T) {
	t.Run("successful revocation", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		tokenID := "token-123"
		
		mockRepo.On("Revoke", tokenID).Return(nil)
		
		err := service.Revoke(tokenID)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("token not found", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		tokenID := "non-existent-token"
		
		mockRepo.On("Revoke", tokenID).
			Return(errors.New("refresh token not found or already revoked"))
		
		err := service.Revoke(tokenID)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "refresh token not found or already revoked")
		mockRepo.AssertExpectations(t)
	})
}

func TestRefreshTokenService_RevokeAllForUser(t *testing.T) {
	t.Run("successful revocation of all user tokens", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		userID := "user-123"
		
		mockRepo.On("RevokeAllForUser", userID).Return(nil)
		
		err := service.RevokeAllForUser(userID)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		userID := "user-123"
		
		mockRepo.On("RevokeAllForUser", userID).
			Return(errors.New("database error"))
		
		err := service.RevokeAllForUser(userID)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

func TestRefreshTokenService_CleanupExpired(t *testing.T) {
	t.Run("successful cleanup", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		mockRepo.On("CleanupExpired").Return(nil)
		
		err := service.CleanupExpired()
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("cleanup error", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		mockRepo.On("CleanupExpired").
			Return(errors.New("cleanup failed"))
		
		err := service.CleanupExpired()
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "cleanup failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestRefreshTokenService_StartCleanupTask(t *testing.T) {
	t.Run("cleanup task runs periodically", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		// Set up expectations for cleanup to be called at least once
		mockRepo.On("CleanupExpired").Return(nil).Times(2)
		
		// Start cleanup task with short interval for testing
		interval := 100 * time.Millisecond
		service.StartCleanupTask(interval)
		
		// Wait for cleanup to run at least twice
		time.Sleep(250 * time.Millisecond)
		
		// Verify cleanup was called
		mockRepo.AssertExpectations(t)
	})

	t.Run("cleanup task continues on error", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		mockJWT := new(MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		// First call fails, second succeeds
		mockRepo.On("CleanupExpired").Return(errors.New("cleanup error")).Once()
		mockRepo.On("CleanupExpired").Return(nil).Once()
		
		// Start cleanup task with short interval for testing
		interval := 100 * time.Millisecond
		service.StartCleanupTask(interval)
		
		// Wait for cleanup to run at least twice
		time.Sleep(250 * time.Millisecond)
		
		// Verify cleanup was called despite error
		mockRepo.AssertExpectations(t)
	})
}

// Mock implementations
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(userID, tokenID string, token string, expiresAt time.Time) error {
	args := m.Called(userID, tokenID, token, expiresAt)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) ValidateAndGet(token string) (*database.RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) Revoke(tokenID string) error {
	args := m.Called(tokenID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) CleanupExpired() error {
	args := m.Called()
	return args.Error(0)
}

type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateTokenPair(userID, username, email, role string) (*auth.TokenPair, error) {
	args := m.Called(userID, username, email, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

func (m *MockJWTManager) ValidateToken(tokenString string, expectedType auth.TokenType) (*auth.Claims, error) {
	args := m.Called(tokenString, expectedType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}