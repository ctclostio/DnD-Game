package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// MockCharacterRepository is a mock implementation of CharacterRepository
type MockCharacterRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockCharacterRepository) Create(character *models.Character) error {
	args := m.Called(character)
	return args.Error(0)
}

// Get mocks the Get method
func (m *MockCharacterRepository) Get(id string) (*models.Character, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

// GetByUser mocks the GetByUser method
func (m *MockCharacterRepository) GetByUser(userID string) ([]*models.Character, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Character), args.Error(1)
}

// Update mocks the Update method
func (m *MockCharacterRepository) Update(character *models.Character) error {
	args := m.Called(character)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockCharacterRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// GetByName mocks the GetByName method
func (m *MockCharacterRepository) GetByName(name string) (*models.Character, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

// List mocks the List method
func (m *MockCharacterRepository) List(limit, offset int) ([]*models.Character, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Character), args.Error(1)
}

// Count mocks the Count method
func (m *MockCharacterRepository) Count() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// GetByUserAndName mocks the GetByUserAndName method
func (m *MockCharacterRepository) GetByUserAndName(userID, name string) (*models.Character, error) {
	args := m.Called(userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

// UpdateHitPoints mocks the UpdateHitPoints method
func (m *MockCharacterRepository) UpdateHitPoints(id string, currentHP, maxHP int) error {
	args := m.Called(id, currentHP, maxHP)
	return args.Error(0)
}

// UpdateLevel mocks the UpdateLevel method
func (m *MockCharacterRepository) UpdateLevel(id string, level, experience int) error {
	args := m.Called(id, level, experience)
	return args.Error(0)
}
