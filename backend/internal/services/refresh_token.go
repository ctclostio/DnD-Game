package services

import (
	"fmt"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// RefreshTokenService handles refresh token operations
type RefreshTokenService struct {
	repo       database.RefreshTokenRepository
	jwtManager *auth.JWTManager
}

// NewRefreshTokenService creates a new refresh token service
func NewRefreshTokenService(repo database.RefreshTokenRepository, jwtManager *auth.JWTManager) *RefreshTokenService {
	return &RefreshTokenService{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

// Create stores a new refresh token
func (s *RefreshTokenService) Create(userID, refreshToken string) error {
	// Validate the token to get claims
	claims, err := s.jwtManager.ValidateToken(refreshToken, auth.RefreshToken)
	if err != nil {
		return fmt.Errorf("invalid refresh token: %w", err)
	}

	// Store the token
	return s.repo.Create(userID, claims.RegisteredClaims.ID, refreshToken, claims.RegisteredClaims.ExpiresAt.Time)
}

// RefreshAccessToken validates a refresh token and generates a new token pair
func (s *RefreshTokenService) RefreshAccessToken(refreshToken string) (*auth.TokenPair, string, error) {
	// Validate the refresh token in the database
	storedToken, err := s.repo.ValidateAndGet(refreshToken)
	if err != nil {
		return nil, "", err
	}

	// Validate the JWT itself
	claims, err := s.jwtManager.ValidateToken(refreshToken, auth.RefreshToken)
	if err != nil {
		return nil, "", err
	}

	// Revoke the old refresh token
	if err := s.repo.Revoke(storedToken.TokenID); err != nil {
		// Log error but continue - we don't want to fail the refresh
		// The old token will eventually expire anyway
	}

	// Generate new token pair
	tokenPair, err := s.jwtManager.GenerateTokenPair(claims.UserID, claims.Username, claims.Email, claims.Role)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate new tokens: %w", err)
	}

	return tokenPair, storedToken.UserID, nil
}

// Revoke marks a refresh token as revoked
func (s *RefreshTokenService) Revoke(tokenID string) error {
	return s.repo.Revoke(tokenID)
}

// RevokeAllForUser revokes all refresh tokens for a user
func (s *RefreshTokenService) RevokeAllForUser(userID string) error {
	return s.repo.RevokeAllForUser(userID)
}

// CleanupExpired removes expired refresh tokens
func (s *RefreshTokenService) CleanupExpired() error {
	return s.repo.CleanupExpired()
}

// StartCleanupTask starts a background task to clean up expired tokens
func (s *RefreshTokenService) StartCleanupTask(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := s.CleanupExpired(); err != nil {
				logger.Error().
					Err(err).
					Msg("Failed to cleanup expired tokens")
			}
		}
	}()
}
