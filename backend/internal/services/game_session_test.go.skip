package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// Using MockGameSessionRepository from campaign_test.go

func TestGameSessionService_CreateSession(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name          string
		session       *models.GameSession
		setupMock     func(*MockGameSessionRepository)
		expectedError string
		validate      func(*testing.T, *models.GameSession)
	}{
		{
			name: "successful session creation",
			session: &models.GameSession{
				Name:        "Adventure Campaign",
				DMID:        "dm-123",
				Description: "A thrilling adventure",
			},
			setupMock: func(m *MockGameSessionRepository) {
				// Create session
				m.On("Create", ctx, mock.MatchedBy(func(s *models.GameSession) bool {
					return s.Name == "Adventure Campaign" &&
						s.DMID == "dm-123" &&
						s.Status == models.GameStatusActive
				})).Return(nil).Run(func(args mock.Arguments) {
					// Simulate DB setting ID and CreatedAt
					session := args.Get(1).(*models.GameSession)
					session.ID = "session-123"
					session.CreatedAt = now
				})
				// Add DM as participant
				m.On("AddParticipant", ctx, "session-123", "dm-123", (*string)(nil)).Return(nil)
			},
			validate: func(t *testing.T, s *models.GameSession) {
				assert.Equal(t, "session-123", s.ID)
				assert.Equal(t, models.GameStatusActive, s.Status)
				assert.Equal(t, now, s.CreatedAt)
			},
		},
		{
			name: "explicit status set",
			session: &models.GameSession{
				Name:   "Pending Campaign",
				DMID:   "dm-456",
				Status: models.GameStatusPending,
			},
			setupMock: func(m *MockGameSessionRepository) {
				m.On("Create", ctx, mock.MatchedBy(func(s *models.GameSession) bool {
					return s.Status == models.GameStatusPending
				})).Return(nil).Run(func(args mock.Arguments) {
					session := args.Get(1).(*models.GameSession)
					session.ID = "session-456"
				})
				m.On("AddParticipant", ctx, "session-456", "dm-456", (*string)(nil)).Return(nil)
			},
			validate: func(t *testing.T, s *models.GameSession) {
				assert.Equal(t, models.GameStatusPending, s.Status)
			},
		},
		{
			name: "missing session name",
			session: &models.GameSession{
				DMID: "dm-123",
			},
			expectedError: "session name is required",
		},
		{
			name: "missing DM ID",
			session: &models.GameSession{
				Name: "Adventure",
			},
			expectedError: "dungeon master user ID is required",
		},
		{
			name: "repository create error",
			session: &models.GameSession{
				Name: "Failed Campaign",
				DMID: "dm-789",
			},
			setupMock: func(m *MockGameSessionRepository) {
				m.On("Create", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "failed to create session",
		},
		{
			name: "add participant error",
			session: &models.GameSession{
				Name: "Participant Error Campaign",
				DMID: "dm-999",
			},
			setupMock: func(m *MockGameSessionRepository) {
				m.On("Create", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					session := args.Get(1).(*models.GameSession)
					session.ID = "session-999"
				})
				m.On("AddParticipant", ctx, "session-999", "dm-999", (*string)(nil)).Return(errors.New("participant error"))
			},
			expectedError: "failed to add DM as participant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewGameSessionService(mockRepo)
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

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_GetSessionByID(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name          string
		sessionID     string
		setupMock     func(*MockGameSessionRepository)
		expected      *models.GameSession
		expectedError string
	}{
		{
			name:      "successful retrieval",
			sessionID: "session-123",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByID", ctx, "session-123").Return(&models.GameSession{
					ID:          "session-123",
					Name:        "Adventure Campaign",
					DMID:        "dm-123",
					Status:      models.GameStatusActive,
					Description: "A thrilling adventure",
					CreatedAt:   now,
				}, nil)
			},
			expected: &models.GameSession{
				ID:          "session-123",
				Name:        "Adventure Campaign",
				DMID:        "dm-123",
				Status:      models.GameStatusActive,
				Description: "A thrilling adventure",
				CreatedAt:   now,
			},
		},
		{
			name:      "session not found",
			sessionID: "nonexistent",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:      "empty session ID",
			sessionID: "",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByID", ctx, "").Return(nil, errors.New("invalid ID"))
			},
			expectedError: "invalid ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewGameSessionService(mockRepo)
			session, err := service.GetSessionByID(ctx, tt.sessionID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, session)
			} else {
				require.NoError(t, err)
				require.NotNil(t, session)
				assert.Equal(t, tt.expected, session)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_UpdateSession(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name          string
		session       *models.GameSession
		setupMock     func(*MockGameSessionRepository)
		expectedError string
		validate      func(*testing.T, *models.GameSession)
	}{
		{
			name: "successful update",
			session: &models.GameSession{
				ID:          "session-123",
				Name:        "Updated Campaign",
				Status:      models.GameStatusPaused,
				Description: "Updated description",
			},
			setupMock: func(m *MockGameSessionRepository) {
				// GetByID returns existing session
				m.On("GetByID", ctx, "session-123").Return(&models.GameSession{
					ID:        "session-123",
					DMID:      "dm-123",
					CreatedAt: now,
				}, nil)
				// Update with preserved fields
				m.On("Update", ctx, mock.MatchedBy(func(s *models.GameSession) bool {
					return s.ID == "session-123" &&
						s.Name == "Updated Campaign" &&
						s.Status == models.GameStatusPaused &&
						s.DMID == "dm-123" && // Should preserve
						s.CreatedAt.Equal(now) // Should preserve
				})).Return(nil)
			},
			validate: func(t *testing.T, s *models.GameSession) {
				assert.Equal(t, "dm-123", s.DMID)
				assert.Equal(t, now, s.CreatedAt)
			},
		},
		{
			name:          "missing session ID",
			session:       &models.GameSession{Name: "No ID"},
			expectedError: "session ID is required",
		},
		{
			name: "session not found",
			session: &models.GameSession{
				ID:   "nonexistent",
				Name: "Updated",
			},
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "session not found",
		},
		{
			name: "repository update error",
			session: &models.GameSession{
				ID:   "session-123",
				Name: "Failed Update",
			},
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByID", ctx, "session-123").Return(&models.GameSession{
					ID:        "session-123",
					DMID:      "dm-123",
					CreatedAt: now,
				}, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewGameSessionService(mockRepo)
			err := service.UpdateSession(ctx, tt.session)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.session)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameSessionService_JoinSession(t *testing.T) {
	ctx := context.Background()
	characterID := "char-123"

	tests := []struct {
		name          string
		sessionID     string
		userID        string
		characterID   *string
		setupMock     func(*MockGameSessionRepository)
		expectedError string
	}{
		{
			name:        "successful join with character",
			sessionID:   "session-123",
			userID:      "user-456",
			characterID: &characterID,
			setupMock: func(m *MockGameSessionRepository) {
				// Check session exists
				m.On("GetByID", ctx, "session-123").Return(&models.GameSession{
					ID:   "session-123",
					DMID: "dm-123",
				}, nil)
				// Add participant
				m.On("AddParticipant", ctx, "session-123", "user-456", &characterID).Return(nil)
			},
		},
		{
			name:        "successful join without character",
			sessionID:   "session-123",
			userID:      "user-789",
			characterID: nil,
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByID", ctx, "session-123").Return(&models.GameSession{
					ID:   "session-123",
					DMID: "dm-123",
				}, nil)
				m.On("AddParticipant", ctx, "session-123", "user-789", (*string)(nil)).Return(nil)
			},
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			userID:        "user-123",
			expectedError: "session ID is required",
		},
		{
			name:          "empty user ID",
			sessionID:     "session-123",
			userID:        "",
			expectedError: "user ID is required",
		},
		{
			name:      "session not found",
			sessionID: "nonexistent",
			userID:    "user-123",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "session not found",
		},
		{
			name:      "add participant error",
			sessionID: "session-123",
			userID:    "user-999",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByID", ctx, "session-123").Return(&models.GameSession{
					ID: "session-123",
				}, nil)
				m.On("AddParticipant", ctx, "session-123", "user-999", (*string)(nil)).Return(errors.New("already joined"))
			},
			expectedError: "already joined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewGameSessionService(mockRepo)
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
		setupMock     func(*MockGameSessionRepository)
		expectedError string
	}{
		{
			name:      "successful leave",
			sessionID: "session-123",
			userID:    "user-456",
			setupMock: func(m *MockGameSessionRepository) {
				// Check session exists
				m.On("GetByID", ctx, "session-123").Return(&models.GameSession{
					ID:   "session-123",
					DMID: "dm-123",
				}, nil)
				// Remove participant
				m.On("RemoveParticipant", ctx, "session-123", "user-456").Return(nil)
			},
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			userID:        "user-123",
			expectedError: "session ID is required",
		},
		{
			name:          "empty user ID",
			sessionID:     "session-123",
			userID:        "",
			expectedError: "user ID is required",
		},
		{
			name:      "session not found",
			sessionID: "nonexistent",
			userID:    "user-123",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "session not found",
		},
		{
			name:      "DM cannot leave",
			sessionID: "session-123",
			userID:    "dm-123",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByID", ctx, "session-123").Return(&models.GameSession{
					ID:   "session-123",
					DMID: "dm-123",
				}, nil)
			},
			expectedError: "dungeon master cannot leave the session",
		},
		{
			name:      "remove participant error",
			sessionID: "session-123",
			userID:    "user-999",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByID", ctx, "session-123").Return(&models.GameSession{
					ID:   "session-123",
					DMID: "dm-123",
				}, nil)
				m.On("RemoveParticipant", ctx, "session-123", "user-999").Return(errors.New("not found"))
			},
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewGameSessionService(mockRepo)
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

func TestGameSessionService_ValidateUserInSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		sessionID     string
		userID        string
		setupMock     func(*MockGameSessionRepository)
		expectedError string
	}{
		{
			name:      "user is participant",
			sessionID: "session-123",
			userID:    "user-456",
			setupMock: func(m *MockGameSessionRepository) {
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
			setupMock: func(m *MockGameSessionRepository) {
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
			setupMock: func(m *MockGameSessionRepository) {
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
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetParticipants", ctx, "session-123").Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get participants",
		},
		{
			name:      "session not found",
			sessionID: "nonexistent",
			userID:    "user-123",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetParticipants", ctx, "nonexistent").Return([]*models.GameParticipant{}, nil)
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "session not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewGameSessionService(mockRepo)
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
		setupMock     func(*MockGameSessionRepository)
		expected      []*models.GameSession
		expectedError string
	}{
		{
			name:     "successful retrieval",
			dmUserID: "dm-123",
			setupMock: func(m *MockGameSessionRepository) {
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
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByDMUserID", ctx, "dm-456").Return([]*models.GameSession{}, nil)
			},
			expected: []*models.GameSession{},
		},
		{
			name:     "repository error",
			dmUserID: "dm-789",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetByDMUserID", ctx, "dm-789").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewGameSessionService(mockRepo)
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
		setupMock     func(*MockGameSessionRepository)
		expectedError string
	}{
		{
			name:      "set player online",
			sessionID: "session-123",
			userID:    "user-456",
			isOnline:  true,
			setupMock: func(m *MockGameSessionRepository) {
				m.On("UpdateParticipantOnlineStatus", ctx, "session-123", "user-456", true).Return(nil)
			},
		},
		{
			name:      "set player offline",
			sessionID: "session-123",
			userID:    "user-456",
			isOnline:  false,
			setupMock: func(m *MockGameSessionRepository) {
				m.On("UpdateParticipantOnlineStatus", ctx, "session-123", "user-456", false).Return(nil)
			},
		},
		{
			name:      "participant not found",
			sessionID: "session-123",
			userID:    "nonexistent",
			isOnline:  true,
			setupMock: func(m *MockGameSessionRepository) {
				m.On("UpdateParticipantOnlineStatus", ctx, "session-123", "nonexistent", true).Return(errors.New("participant not found"))
			},
			expectedError: "participant not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewGameSessionService(mockRepo)
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

func TestGameSessionService_GetSessionParticipants(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name          string
		sessionID     string
		setupMock     func(*MockGameSessionRepository)
		expected      []*models.GameParticipant
		expectedError string
	}{
		{
			name:      "successful retrieval",
			sessionID: "session-123",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetParticipants", ctx, "session-123").Return([]*models.GameParticipant{
					{
						SessionID:   "session-123",
						UserID:      "user-456",
						CharacterID: "char-456",
						Role:        models.ParticipantRolePlayer,
						IsOnline:    true,
						JoinedAt:    now,
						User: &models.User{
							ID:       "user-456",
							Username: "player1",
						},
					},
					{
						SessionID:   "session-123",
						UserID:      "user-789",
						CharacterID: "char-789",
						Role:        models.ParticipantRolePlayer,
						IsOnline:    false,
						JoinedAt:    now.Add(-1 * time.Hour),
						User: &models.User{
							ID:       "user-789",
							Username: "player2",
						},
					},
				}, nil)
			},
			expected: []*models.GameParticipant{
				{
					SessionID:   "session-123",
					UserID:      "user-456",
					CharacterID: "char-456",
					Role:        models.ParticipantRolePlayer,
					IsOnline:    true,
					JoinedAt:    now,
					User: &models.User{
						ID:       "user-456",
						Username: "player1",
					},
				},
				{
					SessionID:   "session-123",
					UserID:      "user-789",
					CharacterID: "char-789",
					Role:        models.ParticipantRolePlayer,
					IsOnline:    false,
					JoinedAt:    now.Add(-1 * time.Hour),
					User: &models.User{
						ID:       "user-789",
						Username: "player2",
					},
				},
			},
		},
		{
			name:      "empty participants",
			sessionID: "session-456",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetParticipants", ctx, "session-456").Return([]*models.GameParticipant{}, nil)
			},
			expected: []*models.GameParticipant{},
		},
		{
			name:      "repository error",
			sessionID: "session-789",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("GetParticipants", ctx, "session-789").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewGameSessionService(mockRepo)
			participants, err := service.GetSessionParticipants(ctx, tt.sessionID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, participants)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, participants)
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
		setupMock     func(*MockGameSessionRepository)
		expectedError string
	}{
		{
			name:      "successful deletion",
			sessionID: "session-123",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("Delete", ctx, "session-123").Return(nil)
			},
		},
		{
			name:      "session not found",
			sessionID: "nonexistent",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("Delete", ctx, "nonexistent").Return(errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:      "repository error",
			sessionID: "session-456",
			setupMock: func(m *MockGameSessionRepository) {
				m.On("Delete", ctx, "session-456").Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockGameSessionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewGameSessionService(mockRepo)
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

// Test aliases
func TestGameSessionService_Aliases(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockGameSessionRepository)
	service := NewGameSessionService(mockRepo)

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

// Test concurrent operations
func TestGameSessionService_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockGameSessionRepository)
	service := NewGameSessionService(mockRepo)

	// Set up expectations for concurrent calls
	for i := 0; i < 10; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		userID := fmt.Sprintf("user-%d", i)
		
		// Mock session retrieval
		mockRepo.On("GetByID", ctx, sessionID).Return(&models.GameSession{
			ID:   sessionID,
			DMID: "dm-123",
		}, nil).Maybe()
		
		// Mock join session
		mockRepo.On("AddParticipant", ctx, sessionID, userID, (*string)(nil)).Return(nil).Maybe()
		
		// Mock online status update
		mockRepo.On("UpdateParticipantOnlineStatus", ctx, sessionID, userID, true).Return(nil).Maybe()
	}

	// Run concurrent operations
	done := make(chan bool, 30)
	
	// Join sessions
	for i := 0; i < 10; i++ {
		go func(id int) {
			sessionID := fmt.Sprintf("session-%d", id)
			userID := fmt.Sprintf("user-%d", id)
			err := service.JoinSession(ctx, sessionID, userID, nil)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Update online status
	for i := 0; i < 10; i++ {
		go func(id int) {
			sessionID := fmt.Sprintf("session-%d", id)
			userID := fmt.Sprintf("user-%d", id)
			err := service.UpdatePlayerOnlineStatus(ctx, sessionID, userID, true)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Get sessions
	for i := 0; i < 10; i++ {
		go func(id int) {
			sessionID := fmt.Sprintf("session-%d", id)
			session, err := service.GetSessionByID(ctx, sessionID)
			assert.NoError(t, err)
			assert.Equal(t, sessionID, session.ID)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 30; i++ {
		<-done
	}

	mockRepo.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkGameSessionService_CreateSession(b *testing.B) {
	ctx := context.Background()
	mockRepo := new(MockGameSessionRepository)
	service := NewGameSessionService(mockRepo)

	// Set up mock to always succeed
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)
	mockRepo.On("AddParticipant", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session := &models.GameSession{
			Name: fmt.Sprintf("Session %d", i),
			DMID: fmt.Sprintf("dm-%d", i),
		}
		_ = service.CreateSession(ctx, session)
	}
}

func BenchmarkGameSessionService_ValidateUserInSession(b *testing.B) {
	ctx := context.Background()
	mockRepo := new(MockGameSessionRepository)
	service := NewGameSessionService(mockRepo)

	// Set up mock
	participants := []*models.GameParticipant{
		{UserID: "user-1"},
		{UserID: "user-2"},
		{UserID: "user-3"},
		{UserID: "user-4"},
		{UserID: "user-5"},
	}
	mockRepo.On("GetParticipants", ctx, "session-123").Return(participants, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.ValidateUserInSession(ctx, "session-123", "user-3")
	}
}