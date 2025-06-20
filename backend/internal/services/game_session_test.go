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
	"github.com/ctclostio/DnD-Game/backend/internal/services/testutil"
)

// Test constants
const (
	// Session names and descriptions
	testSessionName          = "Test Session"
	testSessionLostMines     = "Lost Mines of Phandelver"
	testSessionDescription   = "A test session"
	testSessionUpdatedName   = "Updated Session"
	
	// User and DM IDs
	testDMID456             = "dm-456"
	testUserID456           = "user-456"
	
	// Error messages
	testErrSessionNotFound   = "session not found"
	testErrGameNotFound     = "not found"
	testErrGameRepository   = "repository error"
	testErrParticipantNotFound = "participant not found"
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
				DMID:        testutil.TestDMID,
				Name:        testSessionLostMines,
				Description: "A classic D&D 5e adventure",
			},
			setupMock: func(sessionRepo *mocks.MockGameSessionRepository) {
				sessionRepo.On("Create", ctx, mock.MatchedBy(func(s *models.GameSession) bool {
					return s.DMID == testutil.TestDMID &&
						s.Name == testSessionLostMines &&
						s.Description == "A classic D&D 5e adventure" &&
						s.Status == models.GameStatusPending
				})).Return(nil).Run(func(args mock.Arguments) {
					// Simulate the repository setting the ID
					session := args.Get(1).(*models.GameSession)
					session.ID = testutil.TestSessionID
				})
				sessionRepo.On("AddParticipant", ctx, testutil.TestSessionID, testutil.TestDMID, (*string)(nil)).Return(nil)
			},
			validate: func(t *testing.T, session *models.GameSession) {
				assert.Equal(t, testutil.TestDMID, session.DMID)
				assert.Equal(t, testSessionLostMines, session.Name)
				assert.Equal(t, models.GameStatusPending, session.Status)
			},
		},
		{
			name: "missing session name",
			session: &models.GameSession{
				DMID:        testutil.TestDMID,
				Name:        "",
				Description: testSessionDescription,
			},
			expectedError: "session name is required",
		},
		{
			name: "missing DM ID",
			session: &models.GameSession{
				DMID:        "",
				Name:        testSessionName,
				Description: testSessionDescription,
			},
			expectedError: "dungeon master user ID is required",
		},
		{
			name: testutil.TestDatabaseError,
			session: &models.GameSession{
				DMID:        testutil.TestDMID,
				Name:        testSessionName,
				Description: testSessionDescription,
			},
			setupMock: func(sessionRepo *mocks.MockGameSessionRepository) {
				sessionRepo.On("Create", ctx, mock.Anything).Return(errors.New(testutil.TestDatabaseError))
			},
			expectedError: "failed to create session",
		},
		{
			name: "add participant error",
			session: &models.GameSession{
				DMID:        testutil.TestDMID,
				Name:        testSessionName,
				Description: testSessionDescription,
			},
			setupMock: func(sessionRepo *mocks.MockGameSessionRepository) {
				sessionRepo.On("Create", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					session := args.Get(1).(*models.GameSession)
					session.ID = testutil.TestSessionID
				})
				sessionRepo.On("AddParticipant", ctx, testutil.TestSessionID, testutil.TestDMID, (*string)(nil)).Return(errors.New("participant error"))
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
			sessionID: testutil.TestSessionID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{
					ID:          testutil.TestSessionID,
					DMID:        testutil.TestDMID,
					Name:        "Lost Mines",
					Description: "Adventure",
					Status:      models.GameStatusActive,
					CreatedAt:   time.Now(),
				}
				m.On("GetByID", ctx, testutil.TestSessionID).Return(session, nil)
			},
			validate: func(t *testing.T, session *models.GameSession) {
				assert.Equal(t, testutil.TestSessionID, session.ID)
				assert.Equal(t, testutil.TestDMID, session.DMID)
				assert.Equal(t, "Lost Mines", session.Name)
			},
		},
		{
			name:      testErrSessionNotFound,
			sessionID: "nonexistent",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New(testErrGameNotFound))
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
			sessionID: testutil.TestSessionID,
			userID:    testutil.TestUserID,
			characterID: func() *string {
				s := "char-456"
				return &s
			}(),
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{ID: testutil.TestSessionID, IsActive: true, Status: models.GameStatusActive}
				m.On("GetByID", ctx, testutil.TestSessionID).Return(session, nil)
				m.On("GetParticipants", ctx, testutil.TestSessionID).Return([]*models.GameParticipant{}, nil)
				m.On("AddParticipant", ctx, testutil.TestSessionID, testutil.TestUserID, mock.Anything).Return(nil)
			},
		},
		{
			name:        "successful add participant without character",
			sessionID:   testutil.TestSessionID,
			userID:      testutil.TestUserID,
			characterID: nil,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{ID: testutil.TestSessionID, IsActive: true, Status: models.GameStatusActive}
				m.On("GetByID", ctx, testutil.TestSessionID).Return(session, nil)
				m.On("GetParticipants", ctx, testutil.TestSessionID).Return([]*models.GameParticipant{}, nil)
				m.On("AddParticipant", ctx, testutil.TestSessionID, testutil.TestUserID, (*string)(nil)).Return(nil)
			},
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			userID:        testutil.TestUserID,
			characterID:   nil,
			expectedError: "session ID is required",
		},
		{
			name:          "empty user ID",
			sessionID:     testutil.TestSessionID,
			userID:        "",
			characterID:   nil,
			expectedError: "user ID is required",
		},
		{
			name:        testErrGameRepository,
			sessionID:   testutil.TestSessionID,
			userID:      testutil.TestUserID,
			characterID: nil,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{ID: testutil.TestSessionID, IsActive: true, Status: models.GameStatusActive}
				m.On("GetByID", ctx, testutil.TestSessionID).Return(session, nil)
				m.On("GetParticipants", ctx, testutil.TestSessionID).Return([]*models.GameParticipant{}, nil)
				m.On("AddParticipant", ctx, testutil.TestSessionID, testutil.TestUserID, (*string)(nil)).Return(errors.New(testutil.TestDatabaseError))
			},
			expectedError: testutil.TestDatabaseError,
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
			sessionID: testutil.TestSessionID,
			userID:    testutil.TestUserID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{
					ID:   testutil.TestSessionID,
					DMID: testDMID456, // Different from userID
				}
				m.On("GetByID", ctx, testutil.TestSessionID).Return(session, nil)
				m.On("RemoveParticipant", ctx, testutil.TestSessionID, testutil.TestUserID).Return(nil)
			},
		},
		{
			name:      "DM cannot leave",
			sessionID: testutil.TestSessionID,
			userID:    testutil.TestDMID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{
					ID:   testutil.TestSessionID,
					DMID: testutil.TestDMID, // Same as userID
				}
				m.On("GetByID", ctx, testutil.TestSessionID).Return(session, nil)
			},
			expectedError: "dungeon master cannot leave the session",
		},
		{
			name:      testErrGameRepository,
			sessionID: testutil.TestSessionID,
			userID:    testutil.TestUserID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				session := &models.GameSession{
					ID:   testutil.TestSessionID,
					DMID: testDMID456, // Different from userID
				}
				m.On("GetByID", ctx, testutil.TestSessionID).Return(session, nil)
				m.On("RemoveParticipant", ctx, testutil.TestSessionID, testutil.TestUserID).Return(errors.New(testutil.TestDatabaseError))
			},
			expectedError: testutil.TestDatabaseError,
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
				ID:          testutil.TestSessionID,
				Name:        testSessionUpdatedName,
				Description: "Updated description",
			},
			setupMock: func(m *mocks.MockGameSessionRepository) {
				existing := &models.GameSession{
					ID:        testutil.TestSessionID,
					DMID:      testutil.TestDMID,
					CreatedAt: time.Now(),
				}
				m.On("GetByID", ctx, testutil.TestSessionID).Return(existing, nil)
				m.On("Update", ctx, mock.MatchedBy(func(s *models.GameSession) bool {
					return s.ID == testutil.TestSessionID && s.Name == testSessionUpdatedName
				})).Return(nil)
			},
		},
		{
			name: "missing session ID",
			session: &models.GameSession{
				Name: testSessionUpdatedName,
			},
			expectedError: "session ID is required",
		},
		{
			name: testErrGameRepository,
			session: &models.GameSession{
				ID:   testutil.TestSessionID,
				Name: testSessionUpdatedName,
			},
			setupMock: func(m *mocks.MockGameSessionRepository) {
				existing := &models.GameSession{
					ID:        testutil.TestSessionID,
					DMID:      testutil.TestDMID,
					CreatedAt: time.Now(),
				}
				m.On("GetByID", ctx, testutil.TestSessionID).Return(existing, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New(testutil.TestDatabaseError))
			},
			expectedError: testutil.TestDatabaseError,
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
			sessionID: testutil.TestSessionID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("Delete", ctx, testutil.TestSessionID).Return(nil)
			},
		},
		{
			name:      testErrGameRepository,
			sessionID: testutil.TestSessionID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("Delete", ctx, testutil.TestSessionID).Return(errors.New(testutil.TestDatabaseError))
			},
			expectedError: testutil.TestDatabaseError,
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
			sessionID: testutil.TestSessionID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				participants := []*models.GameParticipant{
					{
						SessionID: testutil.TestSessionID,
						UserID:    testutil.TestDMID,
						Role:      models.ParticipantRoleDM,
						IsOnline:  true,
						JoinedAt:  time.Now(),
					},
					{
						SessionID:   testutil.TestSessionID,
						UserID:      "player-123",
						CharacterID: stringPtr(testutil.TestCharacterID),
						Role:        models.ParticipantRolePlayer,
						IsOnline:    false,
						JoinedAt:    time.Now(),
					},
				}
				m.On("GetParticipants", ctx, testutil.TestSessionID).Return(participants, nil)
			},
			validate: func(t *testing.T, participants []*models.GameParticipant) {
				assert.Len(t, participants, 2)
				assert.Equal(t, models.ParticipantRoleDM, participants[0].Role)
				assert.Equal(t, models.ParticipantRolePlayer, participants[1].Role)
			},
		},
		{
			name:      testErrGameRepository,
			sessionID: testutil.TestSessionID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetParticipants", ctx, testutil.TestSessionID).Return(nil, errors.New(testutil.TestDatabaseError))
			},
			expectedError: testutil.TestDatabaseError,
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
			sessionID: testutil.TestSessionID,
			userID:    testUserID456,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetParticipants", ctx, testutil.TestSessionID).Return([]*models.GameParticipant{
					{
						SessionID: testutil.TestSessionID,
						UserID:    testUserID456,
						Role:      models.ParticipantRolePlayer,
					},
					{
						SessionID: testutil.TestSessionID,
						UserID:    "user-789",
						Role:      models.ParticipantRolePlayer,
					},
				}, nil)
			},
		},
		{
			name:      "user is DM",
			sessionID: testutil.TestSessionID,
			userID:    testutil.TestDMID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				// Not in participants list
				m.On("GetParticipants", ctx, testutil.TestSessionID).Return([]*models.GameParticipant{}, nil)
				// But is the DM
				m.On("GetByID", ctx, testutil.TestSessionID).Return(&models.GameSession{
					ID:   testutil.TestSessionID,
					DMID: testutil.TestDMID,
				}, nil)
			},
		},
		{
			name:      "user not in session",
			sessionID: testutil.TestSessionID,
			userID:    "user-999",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetParticipants", ctx, testutil.TestSessionID).Return([]*models.GameParticipant{
					{
						SessionID: testutil.TestSessionID,
						UserID:    testUserID456,
					},
				}, nil)
				m.On("GetByID", ctx, testutil.TestSessionID).Return(&models.GameSession{
					ID:   testutil.TestSessionID,
					DMID: testutil.TestDMID,
				}, nil)
			},
			expectedError: "user is not a participant in this session",
		},
		{
			name:      "get participants error",
			sessionID: testutil.TestSessionID,
			userID:    testutil.TestUserID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetParticipants", ctx, testutil.TestSessionID).Return(nil, errors.New(testutil.TestDatabaseError))
			},
			expectedError: "failed to get participants",
		},
		{
			name:      testErrSessionNotFound,
			sessionID: "nonexistent",
			userID:    testutil.TestUserID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetParticipants", ctx, "nonexistent").Return([]*models.GameParticipant{}, nil)
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New(testErrGameNotFound))
			},
			expectedError: testErrSessionNotFound,
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
			dmUserID: testutil.TestDMID,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetByDMUserID", ctx, testutil.TestDMID).Return([]*models.GameSession{
					{
						ID:        "session-1",
						Name:      "Campaign 1",
						DMID:      testutil.TestDMID,
						Status:    models.GameStatusActive,
						CreatedAt: now,
					},
					{
						ID:        "session-2",
						Name:      "Campaign 2",
						DMID:      testutil.TestDMID,
						Status:    models.GameStatusPaused,
						CreatedAt: now.Add(-24 * time.Hour),
					},
				}, nil)
			},
			expected: []*models.GameSession{
				{
					ID:        "session-1",
					Name:      "Campaign 1",
					DMID:      testutil.TestDMID,
					Status:    models.GameStatusActive,
					CreatedAt: now,
				},
				{
					ID:        "session-2",
					Name:      "Campaign 2",
					DMID:      testutil.TestDMID,
					Status:    models.GameStatusPaused,
					CreatedAt: now.Add(-24 * time.Hour),
				},
			},
		},
		{
			name:     "no sessions found",
			dmUserID: testDMID456,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetByDMUserID", ctx, testDMID456).Return([]*models.GameSession{}, nil)
			},
			expected: []*models.GameSession{},
		},
		{
			name:     testErrGameRepository,
			dmUserID: "dm-789",
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("GetByDMUserID", ctx, "dm-789").Return(nil, errors.New(testutil.TestDatabaseError))
			},
			expectedError: testutil.TestDatabaseError,
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
			sessionID: testutil.TestSessionID,
			userID:    testUserID456,
			isOnline:  true,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("UpdateParticipantOnlineStatus", ctx, testutil.TestSessionID, testUserID456, true).Return(nil)
			},
		},
		{
			name:      "set player offline",
			sessionID: testutil.TestSessionID,
			userID:    testUserID456,
			isOnline:  false,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("UpdateParticipantOnlineStatus", ctx, testutil.TestSessionID, testUserID456, false).Return(nil)
			},
		},
		{
			name:      testErrParticipantNotFound,
			sessionID: testutil.TestSessionID,
			userID:    "nonexistent",
			isOnline:  true,
			setupMock: func(m *mocks.MockGameSessionRepository) {
				m.On("UpdateParticipantOnlineStatus", ctx, testutil.TestSessionID, "nonexistent", true).Return(errors.New(testErrParticipantNotFound))
			},
			expectedError: testErrParticipantNotFound,
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
		ID:   testutil.TestSessionID,
		Name: testSessionName,
	}
	mockRepo.On("GetByID", ctx, testutil.TestSessionID).Return(expectedSession, nil)

	// Test GetGameSession alias
	session1, err := service.GetGameSession(ctx, testutil.TestSessionID)
	require.NoError(t, err)
	assert.Equal(t, expectedSession, session1)

	// Test GetSession alias
	session2, err := service.GetSession(ctx, testutil.TestSessionID)
	require.NoError(t, err)
	assert.Equal(t, expectedSession, session2)

	mockRepo.AssertExpectations(t)
}
