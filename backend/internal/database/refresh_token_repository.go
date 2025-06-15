package database

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID        string       `db:"id"`
	UserID    string       `db:"user_id"`
	TokenHash string       `db:"token_hash"`
	TokenID   string       `db:"token_id"`
	ExpiresAt time.Time    `db:"expires_at"`
	CreatedAt time.Time    `db:"created_at"`
	RevokedAt sql.NullTime `db:"revoked_at"`
}

// refreshTokenRepository handles refresh token database operations
type refreshTokenRepository struct {
	db *sqlx.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *sqlx.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create stores a new refresh token
func (r *refreshTokenRepository) Create(userID, tokenID, token string, expiresAt time.Time) error {
	// For SQLite compatibility, generate ID and timestamps
	id := uuid.New().String()
	createdAt := time.Now()

	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, token_id, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	query = r.db.Rebind(query)

	tokenHash := hashToken(token)
	_, err := r.db.Exec(query, id, userID, tokenHash, tokenID, expiresAt, createdAt)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// ValidateAndGet validates a refresh token and returns the associated user ID
func (r *refreshTokenRepository) ValidateAndGet(token string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_id, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = ? 
		  AND expires_at > CURRENT_TIMESTAMP
		  AND revoked_at IS NULL
	`
	query = r.db.Rebind(query)

	tokenHash := hashToken(token)
	var refreshToken RefreshToken
	err := r.db.Get(&refreshToken, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired refresh token")
		}
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	return &refreshToken, nil
}

// Revoke marks a refresh token as revoked
func (r *refreshTokenRepository) Revoke(tokenID string) error {
	query := `
		UPDATE refresh_tokens 
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE token_id = ? AND revoked_at IS NULL
	`
	query = r.db.Rebind(query)

	result, err := r.db.Exec(query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("refresh token not found or already revoked")
	}

	return nil
}

// RevokeAllForUser revokes all refresh tokens for a user
func (r *refreshTokenRepository) RevokeAllForUser(userID string) error {
	query := `
		UPDATE refresh_tokens 
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND revoked_at IS NULL
	`
	query = r.db.Rebind(query)

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user's refresh tokens: %w", err)
	}

	return nil
}

// CleanupExpired removes expired refresh tokens
func (r *refreshTokenRepository) CleanupExpired() error {
	// SQLite compatible version - use datetime function instead of INTERVAL
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < CURRENT_TIMESTAMP
		   OR revoked_at < datetime('now', '-30 days')
	`

	// For PostgreSQL, use the original query
	if r.db.DriverName() == "postgres" {
		query = `
			DELETE FROM refresh_tokens
			WHERE expires_at < CURRENT_TIMESTAMP
			   OR revoked_at < CURRENT_TIMESTAMP - INTERVAL '30 days'
		`
	}

	query = r.db.Rebind(query)

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return nil
}

// hashToken creates a SHA256 hash of the token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
