package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/your-username/dnd-game/backend/internal/models"
)

// MockGameSessionRepository is a mock implementation of database.GameSessionRepository
type MockGameSessionRepository struct {
	mock.Mock
}

func (m *MockGameSessionRepository) Create(ctx context.Context, session *models.GameSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetByID(ctx context.Context, id string) (*models.GameSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) GetByDMUserID(ctx context.Context, dmUserID string) ([]*models.GameSession, error) {
	args := m.Called(ctx, dmUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) GetByParticipantUserID(ctx context.Context, userID string) ([]*models.GameSession, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) Update(ctx context.Context, session *models.GameSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockGameSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGameSessionRepository) List(ctx context.Context, offset, limit int) ([]*models.GameSession, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) AddParticipant(ctx context.Context, sessionID, userID string, characterID *string) error {
	args := m.Called(ctx, sessionID, userID, characterID)
	return args.Error(0)
}

func (m *MockGameSessionRepository) RemoveParticipant(ctx context.Context, sessionID, userID string) error {
	args := m.Called(ctx, sessionID, userID)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetParticipants(ctx context.Context, sessionID string) ([]*models.GameParticipant, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameParticipant), args.Error(1)
}

func (m *MockGameSessionRepository) UpdateParticipantOnlineStatus(ctx context.Context, sessionID, userID string, isOnline bool) error {
	args := m.Called(ctx, sessionID, userID, isOnline)
	return args.Error(0)
}

func TestNewGameSessionService(t *testing.T) {
	mockRepo := new(MockGameSessionRepository)
	service := NewGameSessionService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
}

func TestGameSessionService_CreateSession(t *testing.T) {
	mockRepo := new(MockGameSessionRepository)
	service := NewGameSessionService(mockRepo)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		session := &models.GameSession{
			Name: "Epic Campaign",
			DMID: "dm-123",
		}

		mockRepo.On("Create", ctx, mock.MatchedBy(func(s *models.GameSession) bool {
			return s.Name == session.Name && s.DMID == session.DMID
		})).Return(nil).Once()

		err := service.CreateSession(ctx, session)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("missing session name", func(t *testing.T) {
		session := &models.GameSession{
			DMID: "dm-123",
		}

		err := service.CreateSession(ctx, session)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session name is required")
	})

	t.Run("missing DM ID", func(t *testing.T) {
		session := &models.GameSession{
			Name: "Epic Campaign",
		}

		err := service.CreateSession(ctx, session)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dungeon master user ID is required")
	})
}

func TestGameSessionService_GetSessionByID(t *testing.T) {
	mockRepo := new(MockGameSessionRepository)
	service := NewGameSessionService(mockRepo)
	ctx := context.Background()

	t.Run("session found", func(t *testing.T) {
		sessionID := "session-123"
		expectedSession := &models.GameSession{
			ID:   sessionID,
			DMID: "dm-123",
			Name: "Epic Campaign",
		}

		mockRepo.On("GetByID", ctx, sessionID).Return(expectedSession, nil).Once()

		session, err := service.GetSessionByID(ctx, sessionID)

		require.NoError(t, err)
		assert.Equal(t, expectedSession, session)
		mockRepo.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		sessionID := "nonexistent"

		mockRepo.On("GetByID", ctx, sessionID).Return(nil, sql.ErrNoRows).Once()

		session, err := service.GetSessionByID(ctx, sessionID)

		assert.Error(t, err)
		assert.Nil(t, session)
		mockRepo.AssertExpectations(t)
	})
}

func TestGameSessionService_JoinSession(t *testing.T) {
	mockRepo := new(MockGameSessionRepository)
	service := NewGameSessionService(mockRepo)
	ctx := context.Background()

	t.Run("successful join", func(t *testing.T) {
		sessionID := "session-123"
		userID := "user-456"
		characterID := "char-789"

		session := &models.GameSession{
			ID:   sessionID,
			DMID: "dm-123",
		}

		existingParticipants := []*models.GameParticipant{
			{
				SessionID: sessionID,
				UserID:    "dm-123",
			},
		}

		mockRepo.On("GetByID", ctx, sessionID).Return(session, nil).Once()
		mockRepo.On("GetParticipants", ctx, sessionID).Return(existingParticipants, nil).Once()
		mockRepo.On("AddParticipant", ctx, sessionID, userID, &characterID).Return(nil).Once()

		err := service.JoinSession(ctx, sessionID, userID, &characterID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		sessionID := "nonexistent"
		userID := "user-456"
		characterID := "char-789"

		mockRepo.On("GetByID", ctx, sessionID).Return(nil, sql.ErrNoRows).Once()

		err := service.JoinSession(ctx, sessionID, userID, &characterID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user already in session", func(t *testing.T) {
		sessionID := "session-123"
		userID := "user-456"
		characterID := "char-789"

		session := &models.GameSession{
			ID:   sessionID,
			DMID: "dm-123",
		}

		existingParticipants := []*models.GameParticipant{
			{
				SessionID: sessionID,
				UserID:    userID, // User already in session
			},
		}

		mockRepo.On("GetByID", ctx, sessionID).Return(session, nil).Once()
		mockRepo.On("GetParticipants", ctx, sessionID).Return(existingParticipants, nil).Once()

		err := service.JoinSession(ctx, sessionID, userID, &characterID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user already in session")
		mockRepo.AssertExpectations(t)
	})
}

func TestGameSessionService_LeaveSession(t *testing.T) {
	mockRepo := new(MockGameSessionRepository)
	service := NewGameSessionService(mockRepo)
	ctx := context.Background()

	t.Run("successful leave", func(t *testing.T) {
		sessionID := "session-123"
		userID := "user-456"

		session := &models.GameSession{
			ID:   sessionID,
			DMID: "dm-123",
		}

		mockRepo.On("GetByID", ctx, sessionID).Return(session, nil).Once()
		mockRepo.On("RemoveParticipant", ctx, sessionID, userID).Return(nil).Once()

		err := service.LeaveSession(ctx, sessionID, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DM cannot leave", func(t *testing.T) {
		sessionID := "session-123"
		dmID := "dm-123"

		session := &models.GameSession{
			ID:   sessionID,
			DMID: dmID,
		}

		mockRepo.On("GetByID", ctx, sessionID).Return(session, nil).Once()

		err := service.LeaveSession(ctx, sessionID, dmID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dungeon master cannot leave their own session")
		mockRepo.AssertExpectations(t)
	})
}

func TestGameSessionService_GetSessionsByDM(t *testing.T) {
	mockRepo := new(MockGameSessionRepository)
	service := NewGameSessionService(mockRepo)
	ctx := context.Background()

	t.Run("sessions found", func(t *testing.T) {
		dmID := "dm-123"
		expectedSessions := []*models.GameSession{
			{
				ID:   "session-1",
				DMID: dmID,
				Name: "Campaign 1",
			},
			{
				ID:   "session-2",
				DMID: dmID,
				Name: "Campaign 2",
			},
		}

		mockRepo.On("GetByDMUserID", ctx, dmID).Return(expectedSessions, nil).Once()

		sessions, err := service.GetSessionsByDM(ctx, dmID)

		require.NoError(t, err)
		assert.Equal(t, expectedSessions, sessions)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no sessions found", func(t *testing.T) {
		dmID := "dm-123"

		mockRepo.On("GetByDMUserID", ctx, dmID).Return([]*models.GameSession{}, nil).Once()

		sessions, err := service.GetSessionsByDM(ctx, dmID)

		require.NoError(t, err)
		assert.Empty(t, sessions)
		mockRepo.AssertExpectations(t)
	})
}

func TestGameSessionService_GetSessionParticipants(t *testing.T) {
	mockRepo := new(MockGameSessionRepository)
	service := NewGameSessionService(mockRepo)
	ctx := context.Background()

	t.Run("participants found", func(t *testing.T) {
		sessionID := "session-123"
		expectedParticipants := []*models.GameParticipant{
			{
				SessionID: sessionID,
				UserID:    "dm-123",
				Role:      models.ParticipantRoleDM,
			},
			{
				SessionID:   sessionID,
				UserID:      "player-1",
				CharacterID: "char-1",
				Role:        models.ParticipantRolePlayer,
			},
		}

		mockRepo.On("GetParticipants", ctx, sessionID).Return(expectedParticipants, nil).Once()

		participants, err := service.GetSessionParticipants(ctx, sessionID)

		require.NoError(t, err)
		assert.Equal(t, expectedParticipants, participants)
		mockRepo.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		sessionID := "session-123"
		dbError := errors.New("database connection failed")

		mockRepo.On("GetParticipants", ctx, sessionID).Return(nil, dbError).Once()

		participants, err := service.GetSessionParticipants(ctx, sessionID)

		assert.Error(t, err)
		assert.Nil(t, participants)
		mockRepo.AssertExpectations(t)
	})
}