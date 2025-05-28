package models

import (
	"fmt"
	"time"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never expose password hash in JSON
	Role         string    `json:"role" db:"role"`        // "player" or "dm"
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// UserWithCharacters includes a user's characters
type UserWithCharacters struct {
	User
	Characters []Character `json:"characters"`
}

// GameParticipant represents a user participating in a game session
type GameParticipant struct {
	GameSessionID string     `json:"gameSessionId" db:"game_session_id"`
	UserID        string     `json:"userId" db:"user_id"`
	CharacterID   *string    `json:"characterId,omitempty" db:"character_id"`
	IsOnline      bool       `json:"isOnline" db:"is_online"`
	JoinedAt      time.Time  `json:"joinedAt" db:"joined_at"`
	User          *User      `json:"user,omitempty"`
	Character     *Character `json:"character,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expires
	TokenType    string `json:"token_type"`
	User         User   `json:"user"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetConfirm represents password reset confirmation
type PasswordResetConfirm struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Validate performs validation on User
func (u *User) Validate() error {
	if u.Username == "" {
		return ErrInvalidUsername
	}
	if u.Email == "" {
		return ErrInvalidEmail
	}
	return nil
}

// Custom errors for user operations
var (
	ErrUserNotFound      = fmt.Errorf("user not found")
	ErrDuplicateUsername = fmt.Errorf("username already exists")
	ErrDuplicateEmail    = fmt.Errorf("email already exists")
	ErrInvalidUsername   = fmt.Errorf("invalid username")
	ErrInvalidEmail      = fmt.Errorf("invalid email")
	ErrInvalidPassword   = fmt.Errorf("invalid password")
)