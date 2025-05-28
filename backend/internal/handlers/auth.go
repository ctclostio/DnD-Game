package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Register handles user registration
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Validate email format
	if !strings.Contains(req.Email, "@") {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	// Validate password strength
	if len(req.Password) < 8 {
		sendErrorResponse(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	// Register user
	user, err := h.userService.Register(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "username already taken") {
			sendErrorResponse(w, http.StatusConflict, "Username already exists")
		} else if strings.Contains(err.Error(), "email already registered") {
			sendErrorResponse(w, http.StatusConflict, "Email already exists")
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		}
		return
	}

	// Set default role
	user.Role = "player"

	// Generate tokens
	tokenPair, err := h.jwtManager.GenerateTokenPair(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store refresh token
	if err := h.refreshTokenService.Create(user.ID, tokenPair.RefreshToken); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	// Create response
	response := models.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         *user,
	}

	sendJSONResponse(w, http.StatusCreated, response)
}

// Login handles user login
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Get user by username
	user, err := h.userService.GetByUsername(r.Context(), req.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate tokens
	tokenPair, err := h.jwtManager.GenerateTokenPair(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store refresh token
	if err := h.refreshTokenService.Create(user.ID, tokenPair.RefreshToken); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	// Create response
	response := models.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         *user,
	}

	sendJSONResponse(w, http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *Handlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.RefreshToken == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// Refresh access token
	tokenPair, userID, err := h.refreshTokenService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrExpiredToken) {
			sendErrorResponse(w, http.StatusUnauthorized, "Refresh token has expired")
		} else {
			sendErrorResponse(w, http.StatusUnauthorized, "Invalid refresh token")
		}
		return
	}

	// Get user details
	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to get user details")
		return
	}

	// Store new refresh token
	if err := h.refreshTokenService.Create(user.ID, tokenPair.RefreshToken); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	// Create response
	response := models.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         *user,
	}

	sendJSONResponse(w, http.StatusOK, response)
}

// Logout handles user logout
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Revoke all refresh tokens for the user
	if err := h.refreshTokenService.RevokeAllForUser(claims.UserID); err != nil {
		// Log error but don't fail the logout
		// The access token will still expire
	}

	// Response
	response := map[string]string{
		"message": "Successfully logged out",
	}
	sendJSONResponse(w, http.StatusOK, response)
}

// GetCurrentUser returns the current authenticated user
func (h *Handlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user details
	user, err := h.userService.GetByID(r.Context(), claims.UserID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	sendJSONResponse(w, http.StatusOK, user)
}