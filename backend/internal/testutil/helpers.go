package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// TestLogger returns a logger for testing
func TestLogger() *logger.Logger {
	cfg := logger.Config{
		Level:  "debug",
		Pretty: false,
	}
	return logger.New(cfg)
}

// TestContextWithT returns a context with test values
func TestContextWithT(t *testing.T) context.Context {
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
		ID:           "test-char-id",
		UserID:       "test-user-id",
		Name:         "Test Character",
		Race:         "human",
		Class:        "fighter",
		Level:        1,
		HitPoints:    10,
		MaxHitPoints: 10,
		ArmorClass:   10,
		Attributes: models.Attributes{
			Strength:     15,
			Dexterity:    12,
			Constitution: 14,
			Intelligence: 10,
			Wisdom:       13,
			Charisma:     8,
		},
	}
}

// TestJWTManager returns a JWT manager for testing
func TestJWTManager() *auth.JWTManager {
	return auth.NewJWTManager("test-secret-key-32-characters-long", 15*time.Minute, 7*24*time.Hour)
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
		tokenPair, err := jwtManager.GenerateTokenPair(user.ID, user.Username, user.Email, user.Role)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)

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

// RandomString generates a random string of specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
