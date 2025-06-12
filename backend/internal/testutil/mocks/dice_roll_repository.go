package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// MockDiceRollRepository is a mock implementation of DiceRollRepository
type MockDiceRollRepository struct {
	mock.Mock
}

func (m *MockDiceRollRepository) Create(roll *models.DiceRoll) error {
	args := m.Called(roll)
	return args.Error(0)
}

func (m *MockDiceRollRepository) GetByID(id uuid.UUID) (*models.DiceRoll, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollRepository) GetBySession(sessionID uuid.UUID, limit, offset int) ([]*models.DiceRoll, error) {
	args := m.Called(sessionID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollRepository) GetByUser(userID uuid.UUID, limit, offset int) ([]*models.DiceRoll, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiceRoll), args.Error(1)
}
