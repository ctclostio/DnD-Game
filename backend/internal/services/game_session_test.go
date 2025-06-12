package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/internal/services/mocks"
)

// Helper function to get string pointer
func stringPtr(s string) *string {
	return &s
}

func TestGameSessionService_CreateSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		session       *models.GameSession
		setupMock     func(*mocks.MockGameSessionRepository)
		expectedError string
		validate      func(*testing.T, *models.GameSession)
	}{
		{
			name: "successful session creation",
			session: &models.GameSession{
				DMID:        "dm-123",
				Name:        "Lost Mines of Phandelver",
				Description: "A classic D&D 5e adventure",
			},
			setupMock: func(sessionRepo *mocks.MockGameSessionRepository) {
				sessionRepo.On("Create", ctx, mock.MatchedBy(func(s *models.GameSession) bool {
					return s.DMID == "dm-123" &&
						s.Name == "Lost Mines of Phandelver" &&
						s.Description == "A classic D&D 5e adventure" &&
						s.Status == models.GameStatusActive
				})).Return(nil).Run(func(args mock.Arguments) {
					// Simulate the repository setting the ID
					session := args.Get(1).(*models.GameSession)
					session.ID = "session-123"
				})
				sessionRepo.On("AddParticipant", ctx, "session-123", "dm-123", (*string)(nil)).Return(nil)
			},
			validate: func(t *testing.T, session *models.GameSession) {
				assert.Equal(t, "dm-123", session.DMID)
				assert.Equal(t, "Lost Mines of Phandelver", session.Name)
				assert.Equal(t, models.GameStatusActive, session.Status)
			},
		},
		{
			name: "missing session name",
			session: &models.GameSession{
				DMID:        "dm-123",
				Name:        "",
				Description: "A test session",
			},
			expectedError: "session name is required",
		},
		{
			name: "missing DM ID",
			session: &models.GameSession{
				DMID:        "",
				Name:        "Test Session",
				Description: "A test session",
			},
			expectedError: "dungeon master user ID is required",
		},
		{
			name: "database error",
			session: &models.GameSession{
				DMID:        "dm-123",
				Name:        "Test Session",
				Description: "A test session",
			},
			setupMock: func(sessionRepo *mocks.MockGameSessionRepository) {
				sessionRepo.On("Create", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "failed to create session",
		},
		{
			name: "add participant error",
			session: &models.GameSession{
				DMID:        "dm-123",
				Name:        "Test Session",
				Description: "A test session",
			},
			setupMock: func(sessionRepo *mocks.MockGameSessionRepository) {
				sessionRepo.On("Create", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					session := args.Get(1).(*models.GameSession)
					session.ID = "session-123"
				})
				sessionRepo.On("AddParticipant", ctx, "session-123", "dm-123", (*string)(nil)).Return(errors.New("participant error"))
			},
			expectedError: "failed to add DM as participant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSessionRepo := new(mocks.MockGameSessionRepository)

			if tt.setupMock != nil {
				tt.setupMock(mockSessionRepo)
			}

			service := services.NewGameSessionService(mockSessionRepo)
			err := service.CreateSession(ctx, tt.session)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.session)
				}
			}

			mockSessionRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_GetSessionByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		sessionID     string
		setupMock     func(*mocks.MockGameSessionRepository)
		expectedError string
		validate      func(*testing.T, *models.GameSession)
	}{
		{
			name:      "successful get session",
			sessionID: "session-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{
					ID:          "session-123",
					DMID:        "dm-123",
					Name:        "Lost Mines",
					Description: "Adventure",
					Status:      models.GameStatusActive,
					CreatedAt:   time.Now(),
				}
				m.On("GetByID", ctx, "session-123").Return(session, nil)
			},
			validate: func(t *testing.T, session *models.GameSession) {
				assert.Equal(t, "session-123", session.ID)
				assert.Equal(t, "dm-123", session.DMID)
				assert.Equal(t, "Lost Mines", session.Name)
			},
		},
		{
			name:      "session not found",
			sessionID: "nonexistent",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewGameSessionService(mockRepo)
			session, err := service.GetSessionByID(ctx, tt.sessionID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, session)
			} else {
				require.NoError(t, err)
				require.NotNil(t, session)
				if tt.validate != nil {
					tt.validate(t, session)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_JoinSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		sessionID     string
		userID        string
		characterID   *string
		setupMock     func(*mocks.MockGameSessionRepository)
		expectedError string
	}{
		{
			name:      "successful add participant with character",
			sessionID: "session-123",
			userID:    "user-123",
			characterID: func() *string {
				s := "char-456"
				return &s
			}(),
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{ID: "session-123"}
				m.On("GetByID", ctx, "session-123").Return(session, nil)
				m.On("AddParticipant", ctx, "session-123", "user-123", mock.Anything).Return(nil)
			},
		},
		{
			name:        "successful add participant without character",
			sessionID:   "session-123",
			userID:      "user-123",
			characterID: nil,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{ID: "session-123"}
				m.On("GetByID", ctx, "session-123").Return(session, nil)
				m.On("AddParticipant", ctx, "session-123", "user-123", (*string)(nil)).Return(nil)
			},
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			userID:        "user-123",
			characterID:   nil,
			expectedError: "session ID is required",
		},
		{
			name:          "empty user ID",
			sessionID:     "session-123",
			userID:        "",
			characterID:   nil,
			expectedError: "user ID is required",
		},
		{
			name:        "repository error",
			sessionID:   "session-123",
			userID:      "user-123",
			characterID: nil,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{ID: "session-123"}
				m.On("GetByID", ctx, "session-123").Return(session, nil)
				m.On("AddParticipant", ctx, "session-123", "user-123", (*string)(nil)).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewGameSessionService(mockRepo)
			err := service.JoinSession(ctx, tt.sessionID, tt.userID, tt.characterID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_LeaveSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		sessionID     string
		userID        string
		setupMock     func(*mocks.MockGameSessionRepository)
		expectedError string
	}{
		{
			name:      "successful removal",
			sessionID: "session-123",
			userID:    "user-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{
					ID:   "session-123",
					DMID: "dm-456", // Different from userID
				}
				m.On("GetByID", ctx, "session-123").Return(session, nil)
				m.On("RemoveParticipant", ctx, "session-123", "user-123").Return(nil)
			},
		},
		{
			name:      "DM cannot leave",
			sessionID: "session-123",
			userID:    "dm-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{
					ID:   "session-123",
					DMID: "dm-123", // Same as userID
				}
				m.On("GetByID", ctx, "session-123").Return(session, nil)
			},
			expectedError: "dungeon master cannot leave the session",
		},
		{
			name:      "repository error",
			sessionID: "session-123",
			userID:    "user-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{
					ID:   "session-123",
					DMID: "dm-456", // Different from userID
				}
				m.On("GetByID", ctx, "session-123").Return(session, nil)
				m.On("RemoveParticipant", ctx, "session-123", "user-123").Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewGameSessionService(mockRepo)
			err := service.LeaveSession(ctx, tt.sessionID, tt.userID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_UpdateSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		session       *models.GameSession
		setupMock     func(*mocks.MockGameSessionRepository)
		expectedError string
	}{
		{
			name: "successful update",
			session: &models.GameSession{
				ID:          "session-123",
				Name:        "Updated Session",
				Description: "Updated description",
			},
			setupMock: func(m *mocks.MockGameSessionRepository) {
				existing := &models.GameSession{
					ID:        "session-123",
					DMID:      "dm-123",
					CreatedAt: time.Now(),
				}
				m.On("GetByID", ctx, "session-123").Return(existing, nil)
				m.On("Update", ctx, mock.MatchedBy(func(s *models.GameSession) bool {
					return s.ID == "session-123" && s.Name == "Updated Session"
				})).Return(nil)
			},
		},
		{
			name: "missing session ID",
			session: &models.GameSession{
				Name: "Updated Session",
			},
			expectedError: "session ID is required",
		},
		{
			name: "repository error",
			session: &models.GameSession{
				ID:   "session-123",
				Name: "Updated Session",
			},
			setupMock: func(m *mocks.MockGameSessionRepository) {
				existing := &models.GameSession{
					ID:        "session-123",
					DMID:      "dm-123",
					CreatedAt: time.Now(),
				}
				m.On("GetByID", ctx, "session-123").Return(existing, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewGameSessionService(mockRepo)
			err := service.UpdateSession(ctx, tt.session)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_DeleteSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		sessionID     string
		setupMock     func(*mocks.MockGameSessionRepository)
		expectedError string
	}{
		{
			name:      "successful deletion",
			sessionID: "session-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("Delete", ctx, "session-123").Return(nil)
			},
		},
		{
			name:      "repository error",
			sessionID: "session-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("Delete", ctx, "session-123").Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewGameSessionService(mockRepo)
			err := service.DeleteSession(ctx, tt.sessionID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_GetSessionParticipants(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		sessionID     string
		setupMock     func(*mocks.MockGameSessionRepository)
		expectedError string
		validate      func(*testing.T, []*models.GameParticipant)
	}{
		{
			name:      "successful get participants",
			sessionID: "session-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				participants := []*models.GameParticipant{
					{
						SessionID: "session-123",
						UserID:    "dm-123",
						Role:      models.ParticipantRoleDM,
						IsOnline:  true,
						JoinedAt:  time.Now(),
					},
					{
						SessionID:   "session-123",
						UserID:      "player-123",
						CharacterID: stringPtr("char-123"),
						Role:        models.ParticipantRolePlayer,
						IsOnline:    false,
						JoinedAt:    time.Now(),
					},
				}
				m.On("GetParticipants", ctx, "session-123").Return(participants, nil)
			},
			validate: func(t *testing.T, participants []*models.GameParticipant) {
				assert.Len(t, participants, 2)
				assert.Equal(t, models.ParticipantRoleDM, participants[0].Role)
				assert.Equal(t, models.ParticipantRolePlayer, participants[1].Role)
			},
		},
		{
			name:      "repository error",
			sessionID: "session-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetParticipants", ctx, "session-123").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewGameSessionService(mockRepo)
			participants, err := service.GetSessionParticipants(ctx, tt.sessionID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, participants)
			} else {
				require.NoError(t, err)
				require.NotNil(t, participants)
				if tt.validate != nil {
					tt.validate(t, participants)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_ValidateUserInSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		sessionID     string
		userID        string
		setupMock     func(*mocks.MockGameSessionRepository)
		expectedError string
	}{
		{
			name:      "user is participant",
			sessionID: "session-123",
			userID:    "user-456",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetParticipants", ctx, "session-123").Return([]*models.GameParticipant{
					{
						SessionID: "session-123",
						UserID:    "user-456",
						Role:      models.ParticipantRolePlayer,
					},
					{
						SessionID: "session-123",
						UserID:    "user-789",
						Role:      models.ParticipantRolePlayer,
					},
				}, nil)
			},
		},
		{
			name:      "user is DM",
			sessionID: "session-123",
			userID:    "dm-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				// Not in participants list
				m.On("GetParticipants", ctx, "session-123").Return([]*models.GameParticipant{}, nil)
				// But is the DM
				m.On("GetByID", ctx, "session-123").Return(&models.GameSession{
					ID:   "session-123",
					DMID: "dm-123",
				}, nil)
			},
		},
		{
			name:      "user not in session",
			sessionID: "session-123",
			userID:    "user-999",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetParticipants", ctx, "session-123").Return([]*models.GameParticipant{
					{
						SessionID: "session-123",
						UserID:    "user-456",
					},
				}, nil)
				m.On("GetByID", ctx, "session-123").Return(&models.GameSession{
					ID:   "session-123",
					DMID: "dm-123",
				}, nil)
			},
			expectedError: "user is not a participant in this session",
		},
		{
			name:      "get participants error",
			sessionID: "session-123",
			userID:    "user-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetParticipants", ctx, "session-123").Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get participants",
		},
		{
			name:      "session not found",
			sessionID: "nonexistent",
			userID:    "user-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetParticipants", ctx, "nonexistent").Return([]*models.GameParticipant{}, nil)
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "session not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewGameSessionService(mockRepo)
			err := service.ValidateUserInSession(ctx, tt.sessionID, tt.userID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_GetSessionsByDM(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name          string
		dmUserID      string
		setupMock     func(*mocks.MockGameSessionRepository)
		expected      []*models.GameSession
		expectedError string
	}{
		{
			name:     "successful retrieval",
			dmUserID: "dm-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetByDMUserID", ctx, "dm-123").Return([]*models.GameSession{
					{
						ID:        "session-1",
						Name:      "Campaign 1",
						DMID:      "dm-123",
						Status:    models.GameStatusActive,
						CreatedAt: now,
					},
					{
						ID:        "session-2",
						Name:      "Campaign 2",
						DMID:      "dm-123",
						Status:    models.GameStatusPaused,
						CreatedAt: now.Add(-24 * time.Hour),
					},
				}, nil)
			},
			expected: []*models.GameSession{
				{
					ID:        "session-1",
					Name:      "Campaign 1",
					DMID:      "dm-123",
					Status:    models.GameStatusActive,
					CreatedAt: now,
				},
				{
					ID:        "session-2",
					Name:      "Campaign 2",
					DMID:      "dm-123",
					Status:    models.GameStatusPaused,
					CreatedAt: now.Add(-24 * time.Hour),
				},
			},
		},
		{
			name:     "no sessions found",
			dmUserID: "dm-456",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetByDMUserID", ctx, "dm-456").Return([]*models.GameSession{}, nil)
			},
			expected: []*models.GameSession{},
		},
		{
			name:     "repository error",
			dmUserID: "dm-789",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetByDMUserID", ctx, "dm-789").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewGameSessionService(mockRepo)
			sessions, err := service.GetSessionsByDM(ctx, tt.dmUserID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, sessions)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, sessions)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_UpdatePlayerOnlineStatus(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		sessionID     string
		userID        string
		isOnline      bool
		setupMock     func(*mocks.MockGameSessionRepository)
		expectedError string
	}{
		{
			name:      "set player online",
			sessionID: "session-123",
			userID:    "user-456",
			isOnline:  true,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("UpdateParticipantOnlineStatus", ctx, "session-123", "user-456", true).Return(nil)
			},
		},
		{
			name:      "set player offline",
			sessionID: "session-123",
			userID:    "user-456",
			isOnline:  false,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("UpdateParticipantOnlineStatus", ctx, "session-123", "user-456", false).Return(nil)
			},
		},
		{
			name:      "participant not found",
			sessionID: "session-123",
			userID:    "nonexistent",
			isOnline:  true,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("UpdateParticipantOnlineStatus", ctx, "session-123", "nonexistent", true).Return(errors.New("participant not found"))
			},
			expectedError: "participant not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewGameSessionService(mockRepo)
			err := service.UpdatePlayerOnlineStatus(ctx, tt.sessionID, tt.userID, tt.isOnline)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Test aliases
func TestGameSessionService_Aliases(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockGameSessionRepository)
	service := services.NewGameSessionService(mockRepo)

	// Setup mock for all alias calls
	expectedSession := &models.GameSession{
		ID:   "session-123",
		Name: "Test Session",
	}
	mockRepo.On("GetByID", ctx, "session-123").Return(expectedSession, nil)

	// Test GetGameSession alias
	session1, err := service.GetGameSession(ctx, "session-123")
	require.NoError(t, err)
	assert.Equal(t, expectedSession, session1)

	// Test GetSession alias
	session2, err := service.GetSession(ctx, "session-123")
	require.NoError(t, err)
	assert.Equal(t, expectedSession, session2)

	mockRepo.AssertExpectations(t)
}
