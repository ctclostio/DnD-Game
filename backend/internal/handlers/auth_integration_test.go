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
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/handlers"
	"github.com/your-username/dnd-game/backend/internal/middleware"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/testutil"
	"github.com/your-username/dnd-game/backend/pkg/logger"
	"github.com/your-username/dnd-game/backend/pkg/response"
)

func TestAuthFlow_Integration(t *testing.T) {
	// Setup test context
	ctx, cleanup := testutil.SetupIntegrationTest(t)
	defer cleanup()

	// Create logger
	log, err := logger.NewV2(logger.DefaultConfig())
	require.NoError(t, err)

	// Create services
	userService := services.NewUserService(ctx.Repos.Users)
	refreshTokenService := services.NewRefreshTokenService(ctx.Repos.RefreshTokens, ctx.JWTManager)

	svc := &services.Services{
		Users:         userService,
		RefreshTokens: refreshTokenService,
		JWTManager:    ctx.JWTManager,
	}

	// Create handlers and setup routes
	h := handlers.NewHandlers(svc, nil)

	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()

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
		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

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
		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		// Try to register the same username again
		body, _ = json.Marshal(duplicateUser)
		req = httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

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
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

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
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

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
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var loginResp response.Response
		err := json.NewDecoder(w.Body).Decode(&loginResp)
		require.NoError(t, err)

		authData := loginResp.Data.(map[string]interface{})
		accessToken := authData["access_token"].(string)

		// Now test /me endpoint
		req = httptest.NewRequest("GET", "/api/v1/auth/me", nil)
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
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

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
		req.Header.Set("Content-Type", "application/json")

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
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

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
		req = httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
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
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Unauthorized Access Without Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Token Format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
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

	log, err := logger.NewV2(logger.DefaultConfig())
	require.NoError(t, err)

	svc := &services.Services{
		Users:         services.NewUserService(ctx.Repos.Users),
		RefreshTokens: services.NewRefreshTokenService(ctx.Repos.RefreshTokens, ctx.JWTManager),
		JWTManager:    ctx.JWTManager,
	}

	h := handlers.NewHandlers(svc, nil)

	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.LoggingMiddleware(log))
	api.HandleFunc("/auth/register", h.Register).Methods("POST")

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
			// Use unique username for each test
			user := models.RegisterRequest{
				Username: "testuser" + string(rune('0'+i)),
				Email:    "test" + string(rune('0'+i)) + "@example.com",
				Password: tt.password,
			}

			body, _ := json.Marshal(user)
			req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Print the actual response for debugging
				body := w.Body.String()
				t.Logf("Response body: %s", body)

				// For 500 errors, the middleware returns a different format
				if w.Code == http.StatusInternalServerError {
					var errResp map[string]interface{}
					err := json.Unmarshal([]byte(body), &errResp)
					require.NoError(t, err)
					if msg, ok := errResp["message"].(string); ok {
						assert.Contains(t, msg, tt.expectedError)
					}
				} else {
					var resp response.Response
					err := json.Unmarshal([]byte(body), &resp)
					require.NoError(t, err)
					assert.False(t, resp.Success)
					if resp.Error != nil {
						assert.Contains(t, resp.Error.Message, tt.expectedError)
					} else {
						t.Errorf("Expected error but got nil")
					}
				}
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

// Helper to create authenticated request
func createAuthRequest(t *testing.T, method, url string, body interface{}, token string) *http.Request {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}
