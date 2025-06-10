package services

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/services/mocks"
)

func TestRefreshTokenService_Create(t *testing.T) {
	t.Run("successful token storage", func(t *testing.T) {
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		tokenID := "token-123"
		
		mockRepo.On("Revoke", tokenID).Return(nil)
		
		err := service.Revoke(tokenID)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("token not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		userID := "user-123"
		
		mockRepo.On("RevokeAllForUser", userID).Return(nil)
		
		err := service.RevokeAllForUser(userID)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		mockRepo.On("CleanupExpired").Return(nil)
		
		err := service.CleanupExpired()
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("cleanup error", func(t *testing.T) {
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
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
	t.Run("cleanup task starts successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRefreshTokenRepository)
		mockJWT := new(mocks.MockJWTManager)
		
		service := NewRefreshTokenService(mockRepo, mockJWT)
		
		// Use a channel to signal when cleanup is called
		cleanupCalled := make(chan bool, 1)
		
		// Set up expectations for cleanup to be called
		mockRepo.On("CleanupExpired").Return(nil).Run(func(args mock.Arguments) {
			select {
			case cleanupCalled <- true:
			default:
				// Channel is full, ignore extra calls
			}
		}).Maybe() // Allow any number of calls
		
		// Start cleanup task with short interval for testing
		interval := 50 * time.Millisecond
		service.StartCleanupTask(interval)
		
		// Wait for at least one cleanup call
		select {
		case <-cleanupCalled:
			// Success - cleanup was called
		case <-time.After(200 * time.Millisecond):
			t.Fatal("cleanup was not called within expected time")
		}
	})
}

