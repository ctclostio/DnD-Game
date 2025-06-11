package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// userRepository implements UserRepository interface
type userRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	// For SQLite compatibility, generate UUID and timestamps in application
	if r.db.DriverName() == "sqlite3" {
		user.ID = uuid.New().String()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		user.Role = "player" // Default role

		query := `
			INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`

		_, err := r.db.ExecContextRebind(ctx, query,
			user.ID, user.Username, user.Email, user.PasswordHash, user.Role,
			user.CreatedAt, user.UpdatedAt)
		if err != nil {
			// Check for constraint violations
			if strings.Contains(err.Error(), "UNIQUE") {
				if strings.Contains(err.Error(), "username") {
					return models.ErrDuplicateUsername
				}
				if strings.Contains(err.Error(), "email") {
					return models.ErrDuplicateEmail
				}
			}
			return fmt.Errorf("failed to create user: %w", err)
		}
		return nil
	}

	// PostgreSQL version with RETURNING clause
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES (?, ?, ?)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContextRebind(ctx, query, user.Username, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		// Check for constraint violations
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE") {
			if strings.Contains(err.Error(), "username") {
				return models.ErrDuplicateUsername
			}
			if strings.Contains(err.Error(), "email") {
				return models.ErrDuplicateEmail
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, COALESCE(role, 'player') as role, created_at, updated_at
		FROM users
		WHERE id = ?`

	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, COALESCE(role, 'player') as role, created_at, updated_at
		FROM users
		WHERE username = ?`

	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, COALESCE(role, 'player') as role, created_at, updated_at
		FROM users
		WHERE email = ?`

	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET username = ?, email = ?, password_hash = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at`

	err := r.db.QueryRowContextRebind(ctx, query, user.Username, user.Email, user.PasswordHash, user.ID).
		Scan(&user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.ErrUserNotFound
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContextRebind(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// List retrieves a paginated list of users
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	var users []*models.User
	query := `
		SELECT id, username, email, password_hash, COALESCE(role, 'player') as role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	query = r.db.Rebind(query)
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}
