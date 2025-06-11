package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/pkg/errors"
	"github.com/your-username/dnd-game/backend/pkg/response"
	"golang.org/x/crypto/bcrypt"
)

// GetCSRFToken handles CSRF token generation
func (h *Handlers) GetCSRFToken(w http.ResponseWriter, r *http.Request) {
	// The CSRF middleware will automatically set the cookie
	// This endpoint just needs to return success
	response.JSON(w, r, http.StatusOK, map[string]string{
		"message": "CSRF token set",
	})
}

// Register handles user registration
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		response.ErrorWithCode(w, r, errors.ErrCodeMissingRequired, "Username, email, and password are required")
		return
	}

	// Validate email format
	if !strings.Contains(req.Email, "@") {
		response.ErrorWithCode(w, r, errors.ErrCodeInvalidFormat, "Invalid email format")
		return
	}

	// Validate password strength
	if err := validatePassword(req.Password); err != nil {
		response.ErrorWithCode(w, r, errors.ErrCodeInvalidPassword)
		return
	}

	// Register user
	user, err := h.userService.Register(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "username already taken") {
			response.ErrorWithCode(w, r, errors.ErrCodeUserExists, "Username already exists")
		} else if strings.Contains(err.Error(), "email already registered") {
			response.ErrorWithCode(w, r, errors.ErrCodeUserExists, "Email already exists")
		} else {
			response.InternalServerError(w, r, err)
		}
		return
	}

	// Set default role
	user.Role = "player"

	// Generate tokens
	tokenPair, err := h.jwtManager.GenerateTokenPair(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Store refresh token
	if err := h.refreshTokenService.Create(user.ID, tokenPair.RefreshToken); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Create response
	authResponse := models.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         *user,
	}

	response.JSON(w, r, http.StatusCreated, authResponse)
}

// Login handles user login
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		response.ErrorWithCode(w, r, errors.ErrCodeMissingRequired, "Username and password are required")
		return
	}

	// Get user by username
	user, err := h.userService.GetByUsername(r.Context(), req.Username)
	if err != nil {
		response.ErrorWithCode(w, r, errors.ErrCodeInvalidCredentials)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.ErrorWithCode(w, r, errors.ErrCodeInvalidCredentials)
		return
	}

	// Generate tokens
	tokenPair, err := h.jwtManager.GenerateTokenPair(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Store refresh token
	if err := h.refreshTokenService.Create(user.ID, tokenPair.RefreshToken); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Create response
	authResponse := models.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         *user,
	}

	response.JSON(w, r, http.StatusOK, authResponse)
}

// RefreshToken handles token refresh
func (h *Handlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Validate input
	if req.RefreshToken == "" {
		response.ErrorWithCode(w, r, errors.ErrCodeMissingRequired, "Refresh token is required")
		return
	}

	// Refresh access token
	tokenPair, userID, err := h.refreshTokenService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		if err == auth.ErrExpiredToken {
			response.ErrorWithCode(w, r, errors.ErrCodeTokenExpired)
		} else {
			response.ErrorWithCode(w, r, errors.ErrCodeTokenInvalid)
		}
		return
	}

	// Get user details
	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Store new refresh token
	if err := h.refreshTokenService.Create(user.ID, tokenPair.RefreshToken); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Create response
	authResponse := models.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         *user,
	}

	response.JSON(w, r, http.StatusOK, authResponse)
}

// Logout handles user logout
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Revoke all refresh tokens for the user
	if err := h.refreshTokenService.RevokeAllForUser(claims.UserID); err != nil {
		// Log error but don't fail the logout
		// The access token will still expire
	}

	// Response
	logoutResponse := map[string]string{
		"message": "Successfully logged out",
	}
	response.JSON(w, r, http.StatusOK, logoutResponse)
}

// GetCurrentUser returns the current authenticated user
func (h *Handlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Get user details
	user, err := h.userService.GetByID(r.Context(), claims.UserID)
	if err != nil {
		response.ErrorWithCode(w, r, errors.ErrCodeUserNotFound)
		return
	}

	response.JSON(w, r, http.StatusOK, user)
}

// validatePassword checks if the password meets security requirements
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	
	hasUpper := false
	hasLower := false
	hasNumber := false
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}
	}
	
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	
	return nil
}