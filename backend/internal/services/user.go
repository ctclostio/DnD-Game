package services

import (
	"context"
	"fmt"

	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo database.UserRepository
}

func NewUserService(repo database.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// Register creates a new user account
func (s *UserService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	// Validate input
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters long")
	}

	// Check if username already exists
	existingUser, _ := s.repo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		logger.Warn().
			Str("username", req.Username).
			Msg("Registration attempt with existing username")
		return nil, fmt.Errorf("username already taken")
	}

	// Check if email already exists
	existingUser, _ = s.repo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		logger.Warn().
			Str("email", req.Email).
			Msg("Registration attempt with existing email")
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		logger.Error().
			Err(err).
			Str("username", req.Username).
			Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info().
		Str("user_id", user.ID).
		Str("username", user.Username).
		Str("email", user.Email).
		Msg("User registered successfully")

	return user, nil
}

// Login authenticates a user and returns a token
func (s *UserService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	// Get user by username
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		logger.Warn().
			Str("username", req.Username).
			Msg("Login attempt with non-existent username")
		return nil, fmt.Errorf("invalid username or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		logger.Warn().
			Str("username", req.Username).
			Str("user_id", user.ID).
			Msg("Login attempt with incorrect password")
		return nil, fmt.Errorf("invalid username or password")
	}

	logger.Info().
		Str("user_id", user.ID).
		Str("username", user.Username).
		Msg("User logged in successfully")

	// This should be handled by the auth handler, not here
	// Just return the user for now
	return &models.AuthResponse{
		AccessToken:  "", // Will be filled by auth handler
		RefreshToken: "", // Will be filled by auth handler
		ExpiresIn:    0,  // Will be filled by auth handler
		TokenType:    "Bearer",
		User:         *user,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	return s.repo.GetByID(ctx, id)
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	// Validate user ID
	if user.ID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Check if user exists
	existing, err := s.repo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Preserve password hash and created at
	user.PasswordHash = existing.PasswordHash
	user.CreatedAt = existing.CreatedAt

	return s.repo.Update(ctx, user)
}

// ChangePassword updates user password
func (s *UserService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// Get user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("invalid password")
	}

	// Validate new password
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	return s.repo.Update(ctx, user)
}

// DeleteUser deletes a user account
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("user ID is required")
	}
	return s.repo.Delete(ctx, id)
}

// GetByUsername retrieves a user by username
func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.repo.GetByUsername(ctx, username)
}

// GetByID retrieves a user by ID (alias for GetUserByID)
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
	return s.GetUserByID(ctx, id)
}
