package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/config"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/handlers"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

type authTestSuite struct {
	db          *database.DB
	router      *mux.Router
	authHandler *handlers.AuthHandler
	userService *services.UserService
	jwtService  *auth.JWTService
}

func setupAuthTestSuite(t *testing.T) *authTestSuite {
	t.Helper()

	// Set up test database
	db := testutil.SetupTestDB(t)

	// Create services
	userRepo := database.NewUserRepository(db)
	refreshTokenRepo := database.NewRefreshTokenRepository(db)
	userService := services.NewUserService(userRepo)
	refreshTokenService := services.NewRefreshTokenService(refreshTokenRepo)

	// Create JWT service
	cfg := &config.Config{
		JWTSecret:            "test-secret-key",
		JWTAccessExpiration:  15 * time.Minute,
		JWTRefreshExpiration: 7 * 24 * time.Hour,
	}
	jwtService := auth.NewJWTService(cfg)

	// Create auth handler
	authHandler := handlers.NewAuthHandler(userService, refreshTokenService, jwtService)

	// Set up router
	router := mux.NewRouter()
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/auth/refresh", authHandler.RefreshToken).Methods("POST")
	router.HandleFunc("/api/auth/logout", authHandler.Logout).Methods("POST")
	router.HandleFunc("/api/auth/me", auth.AuthMiddleware(jwtService)(http.HandlerFunc(authHandler.GetCurrentUser))).Methods("GET")

	return &authTestSuite{
		db:          db,
		router:      router,
		authHandler: authHandler,
		userService: userService,
		jwtService:  jwtService,
	}
}

func (s *authTestSuite) cleanup() {
	s.db.Close()
}

func TestAuthHandler_Register(t *testing.T) {
	suite := setupAuthTestSuite(t)
	defer suite.cleanup()

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
		validate       func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful registration",
			payload: map[string]interface{}{
				"username": "newuser",
				"email":    "newuser@example.com",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "newuser", response["username"])
				assert.Equal(t, "newuser@example.com", response["email"])
				assert.NotEmpty(t, response["id"])
				assert.NotEmpty(t, response["created_at"])
				_, hasPassword := response["password"]
				assert.False(t, hasPassword, "Password should not be in response")
			},
		},
		{
			name: "missing username",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username is required",
		},
		{
			name: "missing email",
			payload: map[string]interface{}{
				"username": "testuser",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email is required",
		},
		{
			name: "missing password",
			payload: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password is required",
		},
		{
			name: "invalid email format",
			payload: map[string]interface{}{
				"username": "testuser",
				"email":    "invalid-email",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid email format",
		},
		{
			name: "weak password",
			payload: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must be at least 8 characters",
		},
		{
			name: "duplicate username",
			payload: map[string]interface{}{
				"username": "existinguser",
				"email":    "new@example.com",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "username already exists",
		},
		{
			name: "duplicate email",
			payload: map[string]interface{}{
				"username": "newuser2",
				"email":    "existing@example.com",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "email already exists",
		},
		{
			name: "username too short",
			payload: map[string]interface{}{
				"username": "ab",
				"email":    "test@example.com",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username must be at least 3 characters",
		},
		{
			name: "username with invalid characters",
			payload: map[string]interface{}{
				"username": "user@name",
				"email":    "test@example.com",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username can only contain letters, numbers, and underscores",
		},
	}

	// Create existing user for duplicate tests
	existingUser := &models.User{
		Username: "existinguser",
		Email:    "existing@example.com",
		Password: "hashedpassword",
	}
	err := suite.userService.Create(context.Background(), existingUser)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.validate != nil {
				tt.validate(t, response)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	suite := setupAuthTestSuite(t)
	defer suite.cleanup()

	// Create test user
	testUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}
	err := suite.userService.Create(context.Background(), testUser)
	require.NoError(t, err)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
		validate       func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful login with username",
			payload: map[string]interface{}{
				"username": "testuser",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response map[string]interface{}) {
				assert.NotEmpty(t, response["access_token"])
				assert.NotEmpty(t, response["refresh_token"])
				assert.Equal(t, "Bearer", response["token_type"])
				assert.Equal(t, float64(900), response["expires_in"]) // 15 minutes

				user := response["user"].(map[string]interface{})
				assert.Equal(t, "testuser", user["username"])
				assert.Equal(t, "test@example.com", user["email"])
			},
		},
		{
			name: "successful login with email",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response map[string]interface{}) {
				assert.NotEmpty(t, response["access_token"])
				assert.NotEmpty(t, response["refresh_token"])
			},
		},
		{
			name: "missing credentials",
			payload: map[string]interface{}{
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username or email is required",
		},
		{
			name: "missing password",
			payload: map[string]interface{}{
				"username": "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password is required",
		},
		{
			name: "invalid username",
			payload: map[string]interface{}{
				"username": "nonexistent",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid credentials",
		},
		{
			name: "invalid password",
			payload: map[string]interface{}{
				"username": "testuser",
				"password": "WrongPassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid credentials",
		},
		{
			name: "both username and email provided",
			payload: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response map[string]interface{}) {
				assert.NotEmpty(t, response["access_token"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.validate != nil {
				tt.validate(t, response)
			}

			// Check for secure cookie on successful login
			if tt.expectedStatus == http.StatusOK {
				cookies := w.Result().Cookies()
				var refreshCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == "refresh_token" {
						refreshCookie = cookie
						break
					}
				}
				assert.NotNil(t, refreshCookie)
				assert.True(t, refreshCookie.HttpOnly)
				assert.True(t, refreshCookie.Secure)
				assert.Equal(t, http.SameSiteStrictMode, refreshCookie.SameSite)
			}
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	suite := setupAuthTestSuite(t)
	defer suite.cleanup()

	// Create test user and login to get tokens
	testUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}
	err := suite.userService.Create(context.Background(), testUser)
	require.NoError(t, err)

	// Login to get refresh token
	loginPayload := map[string]interface{}{
		"username": "testuser",
		"password": "SecurePass123!",
	}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	suite.router.ServeHTTP(loginW, loginReq)

	var loginResponse map[string]interface{}
	json.NewDecoder(loginW.Body).Decode(&loginResponse)
	validRefreshToken := loginResponse["refresh_token"].(string)
	refreshCookie := loginW.Result().Cookies()[0]

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
		expectedError  string
		validate       func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful refresh with cookie",
			setupRequest: func(req *http.Request) {
				req.AddCookie(refreshCookie)
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response map[string]interface{}) {
				assert.NotEmpty(t, response["access_token"])
				assert.NotEmpty(t, response["refresh_token"])
				assert.NotEqual(t, validRefreshToken, response["refresh_token"]) // Should be new token
			},
		},
		{
			name: "successful refresh with body",
			setupRequest: func(req *http.Request) {
				payload := map[string]interface{}{
					"refresh_token": validRefreshToken,
				}
				body, _ := json.Marshal(payload)
				req.Body = io.NopCloser(bytes.NewReader(body))
				req.ContentLength = int64(len(body))
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response map[string]interface{}) {
				assert.NotEmpty(t, response["access_token"])
			},
		},
		{
			name: "missing refresh token",
			setupRequest: func(req *http.Request) {
				// No cookie or body
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "refresh token is required",
		},
		{
			name: "invalid refresh token",
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: "invalid-token",
				})
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid refresh token",
		},
		{
			name: "expired refresh token",
			setupRequest: func(req *http.Request) {
				// Create an expired token
				expiredToken, _ := suite.jwtService.GenerateRefreshToken(testUser.ID.String(), -1*time.Hour)
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: expiredToken,
				})
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "refresh token expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/auth/refresh", nil)
			req.Header.Set("Content-Type", "application/json")
			
			if tt.setupRequest != nil {
				tt.setupRequest(req)
			}

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.validate != nil {
				tt.validate(t, response)
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	suite := setupAuthTestSuite(t)
	defer suite.cleanup()

	// Create test user and login
	testUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}
	err := suite.userService.Create(context.Background(), testUser)
	require.NoError(t, err)

	// Login to get tokens
	loginPayload := map[string]interface{}{
		"username": "testuser",
		"password": "SecurePass123!",
	}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	suite.router.ServeHTTP(loginW, loginReq)

	var loginResponse map[string]interface{}
	json.NewDecoder(loginW.Body).Decode(&loginResponse)
	accessToken := loginResponse["access_token"].(string)
	refreshToken := loginResponse["refresh_token"].(string)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
		validate       func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful logout",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer "+accessToken)
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Check response
				var response map[string]interface{}
				json.NewDecoder(w.Body).Decode(&response)
				assert.Equal(t, "logged out successfully", response["message"])

				// Check cookie is cleared
				cookies := w.Result().Cookies()
				var refreshCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == "refresh_token" {
						refreshCookie = cookie
						break
					}
				}
				assert.NotNil(t, refreshCookie)
				assert.Equal(t, "", refreshCookie.Value)
				assert.True(t, refreshCookie.MaxAge < 0)

				// Verify refresh token is invalidated
				refreshReq := httptest.NewRequest("POST", "/api/auth/refresh", nil)
				refreshReq.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: refreshToken,
				})
				refreshW := httptest.NewRecorder()
				suite.router.ServeHTTP(refreshW, refreshReq)
				assert.Equal(t, http.StatusUnauthorized, refreshW.Code)
			},
		},
		{
			name: "logout without auth token",
			setupRequest: func(req *http.Request) {
				// No auth header
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "logout with invalid token",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalid-token")
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/auth/logout", nil)
			
			if tt.setupRequest != nil {
				tt.setupRequest(req)
			}

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestAuthHandler_GetCurrentUser(t *testing.T) {
	suite := setupAuthTestSuite(t)
	defer suite.cleanup()

	// Create test user
	testUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}
	err := suite.userService.Create(context.Background(), testUser)
	require.NoError(t, err)

	// Generate access token
	accessToken, err := suite.jwtService.GenerateAccessToken(testUser.ID.String())
	require.NoError(t, err)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
		validate       func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful get current user",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer "+accessToken)
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "testuser", response["username"])
				assert.Equal(t, "test@example.com", response["email"])
				assert.NotEmpty(t, response["id"])
				_, hasPassword := response["password"]
				assert.False(t, hasPassword)
			},
		},
		{
			name: "missing auth header",
			setupRequest: func(req *http.Request) {
				// No auth header
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid token",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalid-token")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed auth header",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "InvalidFormat token")
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/auth/me", nil)
			
			if tt.setupRequest != nil {
				tt.setupRequest(req)
			}

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				
				if tt.validate != nil {
					tt.validate(t, response)
				}
			}
		})
	}
}

func TestAuthHandler_RateLimiting(t *testing.T) {
	suite := setupAuthTestSuite(t)
	defer suite.cleanup()

	// Create test user
	testUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}
	err := suite.userService.Create(context.Background(), testUser)
	require.NoError(t, err)

	// Make multiple failed login attempts
	for i := 0; i < 6; i++ {
		payload := map[string]interface{}{
			"username": "testuser",
			"password": "WrongPassword",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "127.0.0.1:12345" // Same IP for all requests
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		if i < 5 {
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		} else {
			// After 5 failed attempts, should be rate limited
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
			
			var response map[string]interface{}
			json.NewDecoder(w.Body).Decode(&response)
			assert.Contains(t, response["error"], "too many failed login attempts")
		}
	}

	// Verify legitimate login still works from different IP
	payload := map[string]interface{}{
		"username": "testuser",
		"password": "SecurePass123!",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.1:12345" // Different IP
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_CSRF_Protection(t *testing.T) {
	suite := setupAuthTestSuite(t)
	defer suite.cleanup()

	// Create and login user
	testUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}
	err := suite.userService.Create(context.Background(), testUser)
	require.NoError(t, err)

	// Login to get tokens and CSRF token
	loginPayload := map[string]interface{}{
		"username": "testuser",
		"password": "SecurePass123!",
	}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	suite.router.ServeHTTP(loginW, loginReq)

	var loginResponse map[string]interface{}
	json.NewDecoder(loginW.Body).Decode(&loginResponse)
	accessToken := loginResponse["access_token"].(string)

	// Look for CSRF token in response headers or cookies
	csrfToken := loginW.Header().Get("X-CSRF-Token")
	if csrfToken == "" {
		// Check cookies
		for _, cookie := range loginW.Result().Cookies() {
			if cookie.Name == "csrf_token" {
				csrfToken = cookie.Value
				break
			}
		}
	}

	// Test state-changing request without CSRF token
	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Without CSRF protection, this would succeed
	// With CSRF protection, it should fail
	if csrfToken != "" {
		assert.Equal(t, http.StatusForbidden, w.Code)

		// Retry with CSRF token
		req = httptest.NewRequest("POST", "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("X-CSRF-Token", csrfToken)
		w = httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}