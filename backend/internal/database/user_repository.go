package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
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
	if r.db.DriverName() == "sqlite3" {
		return r.createSQLite(ctx, user)
	}
	return r.createPostgreSQL(ctx, user)
}

// createSQLite creates a user for SQLite database
func (r *userRepository) createSQLite(ctx context.Context, user *models.User) error {
	// Generate UUID and timestamps in application
	r.initializeUserDefaults(user)
	
	query := `
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	_, err := r.db.ExecContextRebind(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.Role,
		user.CreatedAt, user.UpdatedAt)
	
	return r.handleCreateError(err)
}

// createPostgreSQL creates a user for PostgreSQL database
func (r *userRepository) createPostgreSQL(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES (?, ?, ?)
		RETURNING id, created_at, updated_at`
	
	err := r.db.QueryRowContextRebind(ctx, query, user.Username, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	return r.handleCreateError(err)
}

// initializeUserDefaults sets default values for a new user
func (r *userRepository) initializeUserDefaults(user *models.User) {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Role = "player" // Default role
}

// handleCreateError handles errors from user creation
func (r *userRepository) handleCreateError(err error) error {
	if err == nil {
		return nil
	}
	
	// Check for constraint violations
	if r.isDuplicateError(err) {
		if strings.Contains(err.Error(), "username") {
			return models.ErrDuplicateUsername
		}
		if strings.Contains(err.Error(), "email") {
			return models.ErrDuplicateEmail
		}
	}
	
	return fmt.Errorf("failed to create user: %w", err)
}

// isDuplicateError checks if the error is a duplicate key error
func (r *userRepository) isDuplicateError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") || 
	       strings.Contains(errStr, "UNIQUE")
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
