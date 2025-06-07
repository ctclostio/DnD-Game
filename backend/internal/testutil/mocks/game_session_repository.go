package mocks

import (
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockGameSessionRepository is a mock implementation of GameSessionRepository
type MockGameSessionRepository struct {
	mock.Mock
}

func (m *MockGameSessionRepository) Create(session *models.GameSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetByID(id uuid.UUID) (*models.GameSession, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) GetByDM(dmID uuid.UUID) ([]*models.GameSession, error) {
	args := m.Called(dmID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) GetByPlayer(playerID uuid.UUID) ([]*models.GameSession, error) {
	args := m.Called(playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) Update(session *models.GameSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockGameSessionRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockGameSessionRepository) List(limit, offset int) ([]*models.GameSession, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) AddParticipant(sessionID, userID uuid.UUID, characterID *uuid.UUID) error {
	args := m.Called(sessionID, userID, characterID)
	return args.Error(0)
}

func (m *MockGameSessionRepository) RemoveParticipant(sessionID, userID uuid.UUID) error {
	args := m.Called(sessionID, userID)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetParticipants(sessionID uuid.UUID) ([]*models.GameParticipant, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameParticipant), args.Error(1)
}

func (m *MockGameSessionRepository) UpdateParticipantCharacter(sessionID, userID uuid.UUID, characterID uuid.UUID) error {
	args := m.Called(sessionID, userID, characterID)
	return args.Error(0)
}