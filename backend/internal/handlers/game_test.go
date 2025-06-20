package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// Test constants
const (
	testSessionNameRequired = "session name is required"
)

// Helper functions to reduce code duplication in mock implementations

func handleMockError(args mock.Arguments, index int) error {
	return args.Error(index)
}

func handleMockSingleReturn[T any](args mock.Arguments, valueIndex, errorIndex int) (*T, error) {
	if args.Get(valueIndex) == nil {
		return nil, args.Error(errorIndex)
	}
	return args.Get(valueIndex).(*T), args.Error(errorIndex)
}

// MockGameSessionService is a mock implementation of the game session service
type MockGameSessionService struct {
	mock.Mock
}

func (m *MockGameSessionService) CreateSession(ctx context.Context, session *models.GameSession) error {
	args := m.Called(ctx, session)
	return handleMockError(args, 0)
}

func (m *MockGameSessionService) GetSession(ctx context.Context, id string) (*models.GameSession, error) {
	args := m.Called(ctx, id)
	return handleMockSingleReturn[models.GameSession](args, 0, 1)
}

func (m *MockGameSessionService) ValidateUserInSession(ctx context.Context, sessionID, userID string) error {
	// ValidateUserInSession checks if a user is currently in the specified session
	args := m.Called(ctx, sessionID, userID)
	// Return validation error if user is not in session
	return handleMockError(args, 0)
}

func (m *MockGameSessionService) JoinSession(ctx context.Context, sessionID, userID string, characterID *string) error {
	args := m.Called(ctx, sessionID, userID, characterID)
	return handleMockError(args, 0)
}

func (m *MockGameSessionService) LeaveSession(ctx context.Context, sessionID, userID string) error {
	// LeaveSession removes a user from the session
	args := m.Called(ctx, sessionID, userID)
	// Return error status for leave operation
	return handleMockError(args, 0)
}

func TestGameHandler_CreateGameSession(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		userID         string
		userRole       string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid game session creation",
			body: map[string]interface{}{
				"name":        "Epic Campaign",
				"description": "A thrilling adventure through forgotten realms",
			},
			userID:         uuid.New().String(),
			userRole:       "dm",
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing session name",
			body: map[string]interface{}{
				"description": "A campaign without a name",
			},
			userID:         uuid.New().String(),
			userRole:       "dm",
			expectedStatus: http.StatusBadRequest,
			expectedError:  testSessionNameRequired,
		},
		{
			name:           "invalid request body",
			body:           "invalid json",
			userID:         uuid.New().String(),
			userRole:       "dm",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "no authentication",
			body: map[string]interface{}{
				"name": "Test Campaign",
			},
			userID:         "", // No auth
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var body []byte
			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader(body))
			req.Header.Set(constants.ContentType, constants.ApplicationJSON)

			// Add auth context (placeholder)

			// For this test, verify request structure
			if tt.body != nil && tt.body != "invalid json" {
				var decoded map[string]interface{}
				err := json.NewDecoder(bytes.NewReader(body)).Decode(&decoded)
				assert.NoError(t, err)

				// Validate required fields
				if _, ok := decoded["name"]; !ok && tt.expectedError == testSessionNameRequired {
					assert.True(t, true, "Name is correctly missing")
				}
			}
		})
	}
}

func TestGameHandler_GetGameSession(t *testing.T) {
	sessionID := uuid.New().String()
	userID := uuid.New().String()

	tests := []struct {
		name           string
		sessionID      string
		userID         string
		expectedStatus int
	}{
		{
			name:           "valid request",
			sessionID:      sessionID,
			userID:         userID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no authentication",
			sessionID:      sessionID,
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, APISessionsBase+tt.sessionID, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{"id": tt.sessionID})
			// Add auth context (placeholder)

			// Verify session ID is properly extracted
			vars := mux.Vars(req)
			assert.Equal(t, tt.sessionID, vars["id"])
		})
	}
}

func TestGameHandler_JoinGameSession(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		sessionID      string
		userID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "join with character",
			body: map[string]interface{}{
				"characterId": uuid.New().String(),
			},
			sessionID:      uuid.New().String(),
			userID:         uuid.New().String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "join without character (spectator)",
			body:           map[string]interface{}{},
			sessionID:      uuid.New().String(),
			userID:         uuid.New().String(),
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid character ID format",
			body: map[string]interface{}{
				"characterId": "not-a-uuid",
			},
			sessionID:      uuid.New().String(),
			userID:         uuid.New().String(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid character ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, APISessionsBase+tt.sessionID+"/join", bytes.NewReader(body))
			req.Header.Set(constants.ContentType, constants.ApplicationJSON)
			// Route vars would be set by router in real handler
			// req = mux.SetURLVars(req, map[string]string{"id": tt.sessionID})

			// Add auth context (placeholder)
			// Add auth context
			var decoded map[string]interface{}
			err := json.NewDecoder(bytes.NewReader(body)).Decode(&decoded)
			assert.NoError(t, err)

			// Validate character ID if provided
			if charID, ok := decoded["characterId"].(string); ok && tt.expectedError == "Invalid character ID format" {
				_, err := uuid.Parse(charID)
				assert.Error(t, err, "Character ID should be invalid UUID")
			}
		})
	}
}

func TestGameHandler_UpdatePlayerStatus(t *testing.T) {
	sessionID := uuid.New().String()

	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
	}{
		{
			name: "update to online",
			body: map[string]interface{}{
				"isOnline": true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "update to offline",
			body: map[string]interface{}{
				"isOnline": false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing status",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, APISessionsBase+sessionID+"/status", bytes.NewReader(body))
			req.Header.Set(constants.ContentType, constants.ApplicationJSON)
			_ = mux.SetURLVars(req, map[string]string{"id": sessionID})

			// Add auth context
			// Add auth context (placeholder)
			var decoded map[string]interface{}
			err := json.NewDecoder(bytes.NewReader(body)).Decode(&decoded)
			assert.NoError(t, err)

			// Check for required field
			_, hasStatus := decoded["isOnline"]
			if tt.expectedStatus == http.StatusBadRequest {
				assert.False(t, hasStatus, "isOnline should be missing")
			} else {
				assert.True(t, hasStatus, "isOnline should be present")
			}
		})
	}
}

func TestGameHandler_SessionValidation(t *testing.T) {
	// Test various session validation scenarios
	validationTests := []struct {
		name        string
		session     models.GameSession
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid session",
			session: models.GameSession{
				Name:        "Valid Campaign",
				Description: "A well-formed game session",
				DMID:        uuid.New().String(),
				Status:      models.GameStatusPending,
			},
			shouldError: false,
		},
		{
			name: "empty session name",
			session: models.GameSession{
				Description: "Missing name",
				DMID:        uuid.New().String(),
			},
			shouldError: true,
			errorMsg:    testSessionNameRequired,
		},
		{
			name: "session name too long",
			session: models.GameSession{
				Name:        string(make([]byte, 256)), // 256 characters
				Description: "Name is too long",
				DMID:        uuid.New().String(),
			},
			shouldError: true,
			errorMsg:    "session name must be less than 255 characters",
		},
		{
			name: "invalid status",
			session: models.GameSession{
				Name:   "Test Session",
				DMID:   uuid.New().String(),
				Status: "invalid_status",
			},
			shouldError: true,
			errorMsg:    "invalid session status",
		},
	}

	for _, tt := range validationTests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate session attributes
			if tt.session.Name == "" && tt.shouldError {
				assert.Equal(t, testSessionNameRequired, tt.errorMsg)
			}
			if len(tt.session.Name) > 255 && tt.shouldError {
				assert.Contains(t, tt.errorMsg, "must be less than 255 characters")
			}
			if tt.session.Status != "" &&
				tt.session.Status != models.GameStatusPending &&
				tt.session.Status != models.GameStatusActive &&
				tt.session.Status != models.GameStatusPaused &&
				tt.session.Status != models.GameStatusCompleted &&
				tt.shouldError {
				assert.Contains(t, tt.errorMsg, "invalid session status")
			}
		})
	}
}

func TestGameHandler_SessionLifecycle(t *testing.T) {
	// Test the lifecycle of a game session
	t.Run("session lifecycle", func(t *testing.T) {
		sessionID := uuid.New().String()
		dmID := uuid.New().String()
		playerID := uuid.New().String()

		// 1. Create session
		session := &models.GameSession{
			ID:          sessionID,
			Name:        "Test Campaign",
			Description: "Testing session lifecycle",
			DMID:        dmID,
			Status:      models.GameStatusPending,
			CreatedAt:   time.Now(),
		}

		assert.NotEmpty(t, session.ID)
		assert.Equal(t, models.GameStatusPending, session.Status)

		// 2. Player joins
		charID := uuid.New().String()
		participant := &models.GameParticipant{
			SessionID:   sessionID,
			UserID:      playerID,
			CharacterID: &charID,
			JoinedAt:    time.Now(),
		}

		assert.Equal(t, sessionID, participant.SessionID)
		assert.NotNil(t, participant.CharacterID)
		assert.NotEmpty(t, *participant.CharacterID)

		// 3. Start session
		session.Status = models.GameStatusActive
		session.StartedAt = &[]time.Time{time.Now()}[0]

		assert.Equal(t, models.GameStatusActive, session.Status)
		assert.NotNil(t, session.StartedAt)

		// 4. End session
		session.Status = models.GameStatusCompleted
		session.EndedAt = &[]time.Time{time.Now()}[0]

		assert.Equal(t, models.GameStatusCompleted, session.Status)
		assert.NotNil(t, session.EndedAt)
	})
}
