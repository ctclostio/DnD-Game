package handlers

// Authentication API documentation

// GetCSRFToken godoc
// @Summary Get CSRF token
// @Description Get a CSRF token for subsequent requests
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "CSRF token set"
// @Router /auth/csrf [get]

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration details"
// @Success 201 {object} models.AuthResponse "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 409 {object} map[string]string "User already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]

// Login godoc
// @Summary User login
// @Description Authenticate user and receive access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.AuthResponse "Login successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]

// RefreshToken godoc
// @Summary Refresh access token
// @Description Exchange a refresh token for a new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.TokenResponse "Token refreshed successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid refresh token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/refresh [post]

// Logout godoc
// @Summary User logout
// @Description Invalidate the user's refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]string "Logout successful"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /auth/logout [post]