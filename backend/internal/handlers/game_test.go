package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
)

// MockGameSessionService is a mock implementation of the game session service
type MockGameSessionService struct {
	mock.Mock
}

func (m *MockGameSessionService) CreateSession(ctx context.Context, session *models.GameSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockGameSessionService) GetSession(ctx context.Context, id string) (*models.GameSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameSessionService) GetSessionByID(ctx context.Context, id string) (*models.GameSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameSessionService) GetGameSession(ctx context.Context, id string) (*models.GameSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameSessionService) GetSessionsByDM(ctx context.Context, dmUserID string) ([]*models.GameSession, error) {
	args := m.Called(ctx, dmUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionService) GetSessionsByPlayer(ctx context.Context, userID string) ([]*models.GameSession, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionService) UpdateSession(ctx context.Context, session *models.GameSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockGameSessionService) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGameSessionService) JoinSession(ctx context.Context, sessionID, userID string, characterID *string) error {
	args := m.Called(ctx, sessionID, userID, characterID)
	return args.Error(0)
}

func (m *MockGameSessionService) LeaveSession(ctx context.Context, sessionID, userID string) error {
	args := m.Called(ctx, sessionID, userID)
	return args.Error(0)
}

func (m *MockGameSessionService) UpdatePlayerOnlineStatus(ctx context.Context, sessionID, userID string, isOnline bool) error {
	args := m.Called(ctx, sessionID, userID, isOnline)
	return args.Error(0)
}

func (m *MockGameSessionService) GetSessionParticipants(ctx context.Context, sessionID string) ([]*models.GameParticipant, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameParticipant), args.Error(1)
}

func (m *MockGameSessionService) ValidateUserInSession(ctx context.Context, sessionID, userID string) error {
	args := m.Called(ctx, sessionID, userID)
	return args.Error(0)
}

// Helper to create test game handlers
func createTestGameHandlers() (*Handlers, *MockGameSessionService) {
	mockGameService := new(MockGameSessionService)
	mockServices := &services.Services{
		GameSessions: mockGameService,
	}
	handlers := NewHandlers(mockServices, nil)
	return handlers, mockGameService
}

func TestHandlers_CreateGameSession(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		userID         string
		userRole       string
		requestBody    interface{}
		setupMock      func(*MockGameSessionService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:     "successful creation by DM",
			userID:   "dm-123",
			userRole: "dm",
			requestBody: map[string]interface{}{
				"name":        "Epic Campaign",
				"description": "A thrilling adventure",
			},
			setupMock: func(m *MockGameSessionService) {
				m.On("CreateSession", mock.Anything, mock.MatchedBy(func(s *models.GameSession) bool {
					return s.Name == "Epic Campaign" &&
						s.Description == "A thrilling adventure" &&
						s.DMID == "dm-123"
				})).Return(nil).Run(func(args mock.Arguments) {
					// Simulate setting ID and timestamp
					session := args.Get(1).(*models.GameSession)
					session.ID = "session-new"
					session.CreatedAt = now
				})
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body []byte) {
				var session models.GameSession
				err := json.Unmarshal(body, &session)
				require.NoError(t, err)
				assert.Equal(t, "session-new", session.ID)
				assert.Equal(t, "Epic Campaign", session.Name)
				assert.Equal(t, "dm-123", session.DMID)
			},
		},
		{
			name:           "invalid request body",
			userID:         "dm-123",
			userRole:       "dm",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "Invalid request body", response["error"])
			},
		},
		{
			name:     "service error",
			userID:   "dm-123",
			userRole: "dm",
			requestBody: map[string]interface{}{
				"name": "Campaign",
			},
			setupMock: func(m *MockGameSessionService) {
				m.On("CreateSession", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "database error", response["error"])
			},
		},
		{
			name:           "unauthorized - no auth context",
			requestBody:    map[string]interface{}{"name": "Campaign"},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockGameService := createTestGameHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockGameService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/game-sessions", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", tt.userRole)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handlers.CreateGameSession(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockGameService.AssertExpectations(t)
		})
	}
}

func TestHandlers_GetGameSession(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		sessionID      string
		userID         string
		setupMock      func(*MockGameSessionService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:      "successful retrieval - DM",
			sessionID: "session-123",
			userID:    "dm-123",
			setupMock: func(m *MockGameSessionService) {
				session := &models.GameSession{
					ID:          "session-123",
					DMID:        "dm-123",
					Name:        "Epic Campaign",
					Status:      models.GameStatusActive,
					CreatedAt:   now,
				}
				m.On("GetSession", mock.Anything, "session-123").Return(session, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var session models.GameSession
				err := json.Unmarshal(body, &session)
				require.NoError(t, err)
				assert.Equal(t, "session-123", session.ID)
				assert.Equal(t, "Epic Campaign", session.Name)
			},
		},
		{
			name:      "successful retrieval - participant",
			sessionID: "session-123",
			userID:    "player-456",
			setupMock: func(m *MockGameSessionService) {
				session := &models.GameSession{
					ID:   "session-123",
					DMID: "dm-123",
					Name: "Epic Campaign",
				}
				m.On("GetSession", mock.Anything, "session-123").Return(session, nil)
				// User is a participant
				m.On("ValidateUserInSession", mock.Anything, "session-123", "player-456").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "forbidden - not participant or DM",
			sessionID: "session-123",
			userID:    "user-789",
			setupMock: func(m *MockGameSessionService) {
				session := &models.GameSession{
					ID:   "session-123",
					DMID: "dm-123",
				}
				m.On("GetSession", mock.Anything, "session-123").Return(session, nil)
				// User is not a participant
				m.On("ValidateUserInSession", mock.Anything, "session-123", "user-789").Return(errors.New("not participant"))
			},
			expectedStatus: http.StatusForbidden,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "don't have access")
			},
		},
		{
			name:      "session not found",
			sessionID: "nonexistent",
			userID:    "user-123",
			setupMock: func(m *MockGameSessionService) {
				m.On("GetSession", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unauthorized",
			sessionID:      "session-123",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockGameService := createTestGameHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockGameService)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/game-sessions/"+tt.sessionID, nil)
			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", "player")
				req = req.WithContext(ctx)
			}

			req = mux.SetURLVars(req, map[string]string{
				"id": tt.sessionID,
			})

			rr := httptest.NewRecorder()
			handlers.GetGameSession(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockGameService.AssertExpectations(t)
		})
	}
}

func TestHandlers_UpdateGameSession(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		userID         string
		userRole       string
		requestBody    interface{}
		setupMock      func(*MockGameSessionService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:      "successful update by DM",
			sessionID: "session-123",
			userID:    "dm-123",
			userRole:  "dm",
			requestBody: map[string]interface{}{
				"name":        "Updated Campaign",
				"description": "New description",
				"status":      "paused",
			},
			setupMock: func(m *MockGameSessionService) {
				// Check ownership
				existingSession := &models.GameSession{
					ID:   "session-123",
					DMID: "dm-123",
					Name: "Original Campaign",
				}
				m.On("GetSession", mock.Anything, "session-123").Return(existingSession, nil).Once()

				// Update session
				m.On("UpdateSession", mock.Anything, mock.MatchedBy(func(s *models.GameSession) bool {
					return s.ID == "session-123" &&
						s.Name == "Updated Campaign" &&
						s.DMID == "dm-123" // Should preserve DM
				})).Return(nil)

				// Get updated session
				updatedSession := &models.GameSession{
					ID:          "session-123",
					DMID:        "dm-123",
					Name:        "Updated Campaign",
					Description: "New description",
					Status:      models.GameStatusPaused,
				}
				m.On("GetSession", mock.Anything, "session-123").Return(updatedSession, nil).Once()
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var session models.GameSession
				err := json.Unmarshal(body, &session)
				require.NoError(t, err)
				assert.Equal(t, "Updated Campaign", session.Name)
				assert.Equal(t, models.GameStatusPaused, session.Status)
			},
		},
		{
			name:      "forbidden - not the DM",
			sessionID: "session-123",
			userID:    "dm-456",
			userRole:  "dm",
			requestBody: map[string]interface{}{
				"name": "Updated",
			},
			setupMock: func(m *MockGameSessionService) {
				existingSession := &models.GameSession{
					ID:   "session-123",
					DMID: "dm-123", // Different DM
				}
				m.On("GetSession", mock.Anything, "session-123").Return(existingSession, nil)
			},
			expectedStatus: http.StatusForbidden,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Only the DM")
			},
		},
		{
			name:      "session not found",
			sessionID: "nonexistent",
			userID:    "dm-123",
			userRole:  "dm",
			requestBody: map[string]interface{}{
				"name": "Updated",
			},
			setupMock: func(m *MockGameSessionService) {
				m.On("GetSession", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid request body",
			sessionID:      "session-123",
			userID:         "dm-123",
			userRole:       "dm",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
			setupMock: func(m *MockGameSessionService) {
				existingSession := &models.GameSession{
					ID:   "session-123",
					DMID: "dm-123",
				}
				m.On("GetSession", mock.Anything, "session-123").Return(existingSession, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockGameService := createTestGameHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockGameService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/game-sessions/"+tt.sessionID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", tt.userRole)
				req = req.WithContext(ctx)
			}

			req = mux.SetURLVars(req, map[string]string{
				"id": tt.sessionID,
			})

			rr := httptest.NewRecorder()
			handlers.UpdateGameSession(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockGameService.AssertExpectations(t)
		})
	}
}

func TestHandlers_JoinGameSession(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		userID         string
		requestBody    interface{}
		setupMock      func(*MockGameSessionService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:      "successful join with character",
			sessionID: "session-123",
			userID:    "player-456",
			requestBody: map[string]interface{}{
				"characterId": "char-789",
			},
			setupMock: func(m *MockGameSessionService) {
				charID := "char-789"
				m.On("JoinSession", mock.Anything, "session-123", "player-456", &charID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "Successfully joined game session", response["message"])
			},
		},
		{
			name:        "successful join without character",
			sessionID:   "session-123",
			userID:      "player-456",
			requestBody: map[string]interface{}{},
			setupMock: func(m *MockGameSessionService) {
				m.On("JoinSession", mock.Anything, "session-123", "player-456", (*string)(nil)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			sessionID:      "session-123",
			userID:         "player-456",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "join error - session full",
			sessionID: "session-123",
			userID:    "player-456",
			requestBody: map[string]interface{}{
				"characterId": "char-789",
			},
			setupMock: func(m *MockGameSessionService) {
				charID := "char-789"
				m.On("JoinSession", mock.Anything, "session-123", "player-456", &charID).Return(errors.New("session is full"))
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "session is full", response["error"])
			},
		},
		{
			name:           "unauthorized",
			sessionID:      "session-123",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockGameService := createTestGameHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockGameService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/game-sessions/"+tt.sessionID+"/join", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", "player")
				req = req.WithContext(ctx)
			}

			req = mux.SetURLVars(req, map[string]string{
				"id": tt.sessionID,
			})

			rr := httptest.NewRecorder()
			handlers.JoinGameSession(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockGameService.AssertExpectations(t)
		})
	}
}

func TestHandlers_LeaveGameSession(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		userID         string
		setupMock      func(*MockGameSessionService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:      "successful leave",
			sessionID: "session-123",
			userID:    "player-456",
			setupMock: func(m *MockGameSessionService) {
				m.On("LeaveSession", mock.Anything, "session-123", "player-456").Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "Successfully left game session", response["message"])
			},
		},
		{
			name:      "leave error - DM cannot leave",
			sessionID: "session-123",
			userID:    "dm-123",
			setupMock: func(m *MockGameSessionService) {
				m.On("LeaveSession", mock.Anything, "session-123", "dm-123").Return(errors.New("dungeon master cannot leave the session"))
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "dungeon master cannot leave")
			},
		},
		{
			name:           "unauthorized",
			sessionID:      "session-123",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockGameService := createTestGameHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockGameService)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/game-sessions/"+tt.sessionID+"/leave", nil)

			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", "player")
				req = req.WithContext(ctx)
			}

			req = mux.SetURLVars(req, map[string]string{
				"id": tt.sessionID,
			})

			rr := httptest.NewRecorder()
			handlers.LeaveGameSession(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockGameService.AssertExpectations(t)
		})
	}
}

func TestHandlers_GetUserGameSessions(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockGameSessionService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:   "successful retrieval",
			userID: "player-123",
			setupMock: func(m *MockGameSessionService) {
				sessions := []*models.GameSession{
					{
						ID:        "session-1",
						Name:      "Campaign 1",
						DMID:      "dm-456",
						Status:    models.GameStatusActive,
						CreatedAt: now,
					},
					{
						ID:        "session-2",
						Name:      "Campaign 2",
						DMID:      "dm-789",
						Status:    models.GameStatusPaused,
						CreatedAt: now.Add(-24 * time.Hour),
					},
				}
				m.On("GetSessionsByPlayer", mock.Anything, "player-123").Return(sessions, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var sessions []*models.GameSession
				err := json.Unmarshal(body, &sessions)
				require.NoError(t, err)
				assert.Len(t, sessions, 2)
				assert.Equal(t, "Campaign 1", sessions[0].Name)
				assert.Equal(t, "Campaign 2", sessions[1].Name)
			},
		},
		{
			name:   "empty sessions list",
			userID: "player-456",
			setupMock: func(m *MockGameSessionService) {
				m.On("GetSessionsByPlayer", mock.Anything, "player-456").Return([]*models.GameSession{}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var sessions []*models.GameSession
				err := json.Unmarshal(body, &sessions)
				require.NoError(t, err)
				assert.Empty(t, sessions)
			},
		},
		{
			name:   "service error",
			userID: "player-789",
			setupMock: func(m *MockGameSessionService) {
				m.On("GetSessionsByPlayer", mock.Anything, "player-789").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "database error", response["error"])
			},
		},
		{
			name:           "unauthorized",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockGameService := createTestGameHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockGameService)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/game-sessions", nil)

			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", "player")
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handlers.GetUserGameSessions(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockGameService.AssertExpectations(t)
		})
	}
}