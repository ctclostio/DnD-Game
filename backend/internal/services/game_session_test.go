package services_test

import (
	"context"
	"errors"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/services/mocks"
)

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
			name:        "empty session ID",
			sessionID:   "",
			userID:      "user-123",
			characterID: nil,
			expectedError: "session ID is required",
		},
		{
			name:        "empty user ID",
			sessionID:   "session-123",
			userID:      "",
			characterID: nil,
			expectedError: "user ID is required",
		},
		{
			name:      "repository error",
			sessionID: "session-123",
			userID:    "user-123",
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
					DMID: "dm-456",  // Different from userID
				}
				m.On("GetByID", ctx, "session-123").Return(session, nil)
				m.On("RemoveParticipant", ctx, "session-123", "user-123").Return(nil)
			},
		},
		{
			name:      "repository error",
			sessionID: "session-123",
			userID:    "user-123",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{
					ID:   "session-123",
					DMID: "dm-456",  // Different from userID
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
						SessionID:   "session-123",
						UserID:      "dm-123",
						Role:        models.ParticipantRoleDM,
						IsOnline:    true,
						JoinedAt:    time.Now(),
					},
					{
						SessionID:   "session-123",
						UserID:      "player-123",
						CharacterID: "char-123",
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