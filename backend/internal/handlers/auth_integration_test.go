package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/handlers"
	"github.com/ctclostio/DnD-Game/backend/internal/middleware"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

func TestAuthFlow_Integration(t *testing.T) {
	// Setup test context
	ctx, cleanup := testutil.SetupIntegrationTest(t)
	defer cleanup()

	// Create logger
	cfg := logger.DefaultConfig()
	log, err := logger.NewV2(&cfg)
	require.NoError(t, err)

	// Create services
	userService := services.NewUserService(ctx.Repos.Users)
	refreshTokenService := services.NewRefreshTokenService(ctx.Repos.RefreshTokens, ctx.JWTManager)

	svc := &services.Services{
		DB:            ctx.DB,
		Users:         userService,
		RefreshTokens: refreshTokenService,
		JWTManager:    ctx.JWTManager,
	}

	// Create handlers and setup routes
	h := handlers.NewHandlers(svc, ctx.DB, nil)

	router := mux.NewRouter()
	api := router.PathPrefix(constants.APIv1Prefix).Subrouter()

	// Apply middleware
	api.Use(middleware.RequestIDMiddleware)
	api.Use(middleware.LoggingMiddleware(log))

	// Auth routes
	api.HandleFunc("/auth/register", h.Register).Methods("POST")
	api.HandleFunc("/auth/login", h.Login).Methods("POST")
	api.HandleFunc("/auth/refresh", h.RefreshToken).Methods("POST")
	api.HandleFunc("/auth/logout", auth.NewMiddleware(ctx.JWTManager).Authenticate(h.Logout)).Methods("POST")
	api.HandleFunc("/auth/me", auth.NewMiddleware(ctx.JWTManager).Authenticate(h.GetCurrentUser)).Methods("GET")

	// Test data
	testUser := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}

	t.Run("Register New User", func(t *testing.T) {
		body, _ := json.Marshal(testUser)
		req := httptest.NewRequest("POST", constants.AuthRegisterPath, bytes.NewBuffer(body))
		req.Header.Set(constants.ContentType, constants.ApplicationJSON)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		// Verify response has auth tokens
		authData, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, authData["access_token"])
		assert.NotEmpty(t, authData["refresh_token"])
		assert.NotNil(t, authData["user"])

		// Verify user in database
		var count int
		err = ctx.SQLXDB.Get(&count, "SELECT COUNT(*) FROM users WHERE username = ?", testUser.Username)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Cannot Register Duplicate Username", func(t *testing.T) {
		// First register a user
		duplicateUser := models.RegisterRequest{
			Username: "duplicateuser",
			Email:    "duplicate@example.com",
			Password: "SecurePass123!",
		}

		body, _ := json.Marshal(duplicateUser)
		req := httptest.NewRequest("POST", constants.AuthRegisterPath, bytes.NewBuffer(body))
		req.Header.Set(constants.ContentType, constants.ApplicationJSON)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		// Try to register the same username again
		body, _ = json.Marshal(duplicateUser)
		req = httptest.NewRequest("POST", constants.AuthRegisterPath, bytes.NewBuffer(body))
		req.Header.Set(constants.ContentType, constants.ApplicationJSON)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code) // 409 is the correct status for duplicate

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Error.Message, "Username") // Capital U
	})

	t.Run("Login with Valid Credentials", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Username: testUser.Username,
			Password: testUser.Password,
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", constants.AuthLoginPath, bytes.NewBuffer(body))
		req.Header.Set(constants.ContentType, constants.ApplicationJSON)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		authData, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, authData["access_token"])
		assert.NotEmpty(t, authData["refresh_token"])
	})

	t.Run("Login with Invalid Password", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Username: testUser.Username,
			Password: "WrongPassword123!",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", constants.AuthLoginPath, bytes.NewBuffer(body))
		req.Header.Set(constants.ContentType, constants.ApplicationJSON)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Error.Message, "Invalid")
	})

	t.Run("Get Current User", func(t *testing.T) {
		// First login to get token
		loginReq := models.LoginRequest{
			Username: testUser.Username,
			Password: testUser.Password,
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", constants.AuthLoginPath, bytes.NewBuffer(body))
		req.Header.Set(constants.ContentType, constants.ApplicationJSON)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var loginResp response.Response
		err := json.NewDecoder(w.Body).Decode(&loginResp)
		require.NoError(t, err)

		authData := loginResp.Data.(map[string]interface{})
		accessToken := authData["access_token"].(string)

		// Now test /me endpoint
		req = httptest.NewRequest("GET", constants.AuthMePath, http.NoBody)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var meResp response.Response
		err = json.NewDecoder(w.Body).Decode(&meResp)
		require.NoError(t, err)
		assert.True(t, meResp.Success)

		userData, ok := meResp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, testUser.Username, userData["username"])
		assert.Equal(t, testUser.Email, userData["email"])
	})

	t.Run("Refresh Token", func(t *testing.T) {
		// First login to get tokens
		loginReq := models.LoginRequest{
			Username: testUser.Username,
			Password: testUser.Password,
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", constants.AuthLoginPath, bytes.NewBuffer(body))
		req.Header.Set(constants.ContentType, constants.ApplicationJSON)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var loginResp response.Response
		err := json.NewDecoder(w.Body).Decode(&loginResp)
		require.NoError(t, err)

		authData := loginResp.Data.(map[string]interface{})
		refreshToken := authData["refresh_token"].(string)

		// Now test refresh
		refreshReq := models.RefreshTokenRequest{
			RefreshToken: refreshToken,
		}

		body, _ = json.Marshal(refreshReq)
		req = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set(constants.ContentType, constants.ApplicationJSON)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var refreshResp response.Response
		err = json.NewDecoder(w.Body).Decode(&refreshResp)
		require.NoError(t, err)
		assert.True(t, refreshResp.Success)

		newAuthData, ok := refreshResp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, newAuthData["access_token"])
		assert.NotEmpty(t, newAuthData["refresh_token"])
		assert.NotEqual(t, refreshToken, newAuthData["refresh_token"], "Should get new refresh token")
	})

	t.Run("Logout", func(t *testing.T) {
		// First login to get tokens
		loginReq := models.LoginRequest{
			Username: testUser.Username,
			Password: testUser.Password,
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", constants.AuthLoginPath, bytes.NewBuffer(body))
		req.Header.Set(constants.ContentType, constants.ApplicationJSON)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var loginResp response.Response
		err := json.NewDecoder(w.Body).Decode(&loginResp)
		require.NoError(t, err)

		authData := loginResp.Data.(map[string]interface{})
		accessToken := authData["access_token"].(string)
		refreshToken := authData["refresh_token"].(string)

		// Now logout
		req = httptest.NewRequest("POST", "/api/v1/auth/logout", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify refresh token is invalidated
		refreshReq := models.RefreshTokenRequest{
			RefreshToken: refreshToken,
		}

		body, _ = json.Marshal(refreshReq)
		req = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set(constants.ContentType, constants.ApplicationJSON)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Unauthorized Access Without Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", constants.AuthMePath, http.NoBody)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Token Format", func(t *testing.T) {
		req := httptest.NewRequest("GET", constants.AuthMePath, http.NoBody)
		req.Header.Set("Authorization", "InvalidTokenFormat")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Expired Token", func(t *testing.T) {
		// Create an expired token
		_, err := ctx.JWTManager.GenerateTokenPair("test-id", "test", "test@example.com", "player")
		require.NoError(t, err)

		// Wait for token to expire (if access token duration is short enough for testing)
		// Or mock the time if your JWT manager supports it
		// For now, we'll skip the actual expiration test
		t.Skip("Skipping expired token test - requires time manipulation")
	})
}

func TestPasswordValidation_Integration(t *testing.T) {
	ctx, cleanup := testutil.SetupIntegrationTest(t)
	defer cleanup()

	router := setupTestRouter(t, ctx)

	tests := []struct {
		name           string
		password       string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Too Short",
			password:       "Aa1!",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password",
		},
		{
			name:           "No Uppercase",
			password:       "password123!",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password",
		},
		{
			name:           "No Lowercase",
			password:       "PASSWORD123!",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password",
		},
		{
			name:           "No Number",
			password:       "Password!",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password",
		},
		{
			name:           "Valid Password",
			password:       "ValidPass123!",
			expectedStatus: http.StatusCreated,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createTestRegisterRequest(i, tt.password)
			w := executeRegisterRequest(t, router, user)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				validateErrorResponse(t, w, tt.expectedError)
			}
		})
	}
}

func TestRateLimiting_Integration(t *testing.T) {
	t.Skip("Rate limiting not yet implemented")

	// This test would verify that:
	// 1. Multiple rapid login attempts are rate limited
	// 2. Rate limit resets after time window
	// 3. Different endpoints have different rate limits
}

// Helper functions to reduce cognitive complexity

// setupTestRouter creates a test router with necessary middleware and handlers
func setupTestRouter(t *testing.T, ctx *testutil.IntegrationTestContext) *mux.Router {
	cfg := logger.DefaultConfig()
	log, err := logger.NewV2(&cfg)
	require.NoError(t, err)

	svc := &services.Services{
		DB:            ctx.DB,
		Users:         services.NewUserService(ctx.Repos.Users),
		RefreshTokens: services.NewRefreshTokenService(ctx.Repos.RefreshTokens, ctx.JWTManager),
		JWTManager:    ctx.JWTManager,
	}

	h := handlers.NewHandlers(svc, ctx.DB, nil)

	router := mux.NewRouter()
	api := router.PathPrefix(constants.APIv1Prefix).Subrouter()
	api.Use(middleware.LoggingMiddleware(log))
	api.HandleFunc("/auth/register", h.Register).Methods("POST")
	
	return router
}

// createTestRegisterRequest creates a unique register request for testing
func createTestRegisterRequest(index int, password string) models.RegisterRequest {
	return models.RegisterRequest{
		Username: "testuser" + string(rune('0'+index)),
		Email:    "test" + string(rune('0'+index)) + "@example.com",
		Password: password,
	}
}

// executeRegisterRequest executes a register request and returns the response recorder
func executeRegisterRequest(t *testing.T, router *mux.Router, user models.RegisterRequest) *httptest.ResponseRecorder {
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", constants.AuthRegisterPath, bytes.NewBuffer(body))
	req.Header.Set(constants.ContentType, constants.ApplicationJSON)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	return w
}

// validateErrorResponse validates the error response based on status code
func validateErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedError string) {
	body := w.Body.String()
	t.Logf("Response body: %s", body)

	// For 500 errors, the middleware returns a different format
	if w.Code == http.StatusInternalServerError {
		validateInternalServerError(t, body, expectedError)
	} else {
		validateStandardError(t, body, expectedError)
	}
}

// validateInternalServerError validates 500 error responses
func validateInternalServerError(t *testing.T, body string, expectedError string) {
	var errResp map[string]interface{}
	err := json.Unmarshal([]byte(body), &errResp)
	require.NoError(t, err)
	
	if msg, ok := errResp["message"].(string); ok {
		assert.Contains(t, msg, expectedError)
	}
}

// validateStandardError validates standard error responses
func validateStandardError(t *testing.T, body string, expectedError string) {
	var resp response.Response
	err := json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	
	if resp.Error != nil {
		assert.Contains(t, resp.Error.Message, expectedError)
	} else {
		t.Errorf("Expected error but got nil")
	}
}

// Helper to create authenticated request
