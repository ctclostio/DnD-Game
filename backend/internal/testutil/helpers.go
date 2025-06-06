package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/pkg/logger"
)

// TestLogger returns a logger for testing
func TestLogger() *logger.Logger {
	cfg := logger.Config{
		Level:  "debug",
		Pretty: false,
	}
	return logger.New(cfg)
}

// TestContext returns a context with test values
func TestContext(t *testing.T) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.RequestIDKey, "test-request-id")
	return ctx
}

// TestUser returns a test user
func TestUser() *models.User {
	return &models.User{
		ID:       "test-user-id",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "player",
	}
}

// TestCharacter returns a test character
func TestCharacter() *models.Character {
	return &models.Character{
		ID:        "test-char-id",
		UserID:    "test-user-id",
		Name:      "Test Character",
		Race:      "human",
		Class:     "fighter",
		Level:     1,
		HitPoints: 10,
		MaxHP:     10,
		ArmorClass: 10,
		Attributes: map[string]int{
			"strength":     15,
			"dexterity":    12,
			"constitution": 14,
			"intelligence": 10,
			"wisdom":       13,
			"charisma":     8,
		},
	}
}

// TestJWTManager returns a JWT manager for testing
func TestJWTManager() *auth.JWTManager {
	return auth.NewJWTManager("test-secret-key-32-characters-long", "15m", "7d")
}

// AuthenticatedRequest creates an authenticated HTTP request
func AuthenticatedRequest(t *testing.T, method, path string, body interface{}, user *models.User) *http.Request {
	t.Helper()
	
	var req *http.Request
	var err error
	
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		req, err = http.NewRequest(method, path, bytes.NewReader(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, path, nil)
		require.NoError(t, err)
	}
	
	// Add authentication
	if user != nil {
		jwtManager := TestJWTManager()
		token, err := jwtManager.GenerateToken(user.ID, user.Username, user.Role, auth.AccessToken)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		
		// Add user to context
		ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
		req = req.WithContext(ctx)
	}
	
	return req
}

// ExecuteRequest executes an HTTP request and returns the response
func ExecuteRequest(req *http.Request, router *mux.Router) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// ParseResponse parses JSON response body
func ParseResponse(t *testing.T, rr *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	err := json.NewDecoder(rr.Body).Decode(v)
	require.NoError(t, err)
}

// AssertErrorResponse asserts that the response is an error with expected status
func AssertErrorResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	require.Equal(t, expectedStatus, rr.Code)
	
	var response map[string]interface{}
	ParseResponse(t, rr, &response)
	
	require.Contains(t, response, "type")
	require.Contains(t, response, "message")
}

// MockDB provides a mock database for testing
type MockDB struct {
	PingFunc func() error
}

func (m *MockDB) Ping() error {
	if m.PingFunc != nil {
		return m.PingFunc()
	}
	return nil
}