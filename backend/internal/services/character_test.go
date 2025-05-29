package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// Mock repository
type MockCharacterRepository struct {
	mock.Mock
}

func (m *MockCharacterRepository) Create(ctx context.Context, character *models.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockCharacterRepository) GetByID(ctx context.Context, id string) (*models.Character, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Character, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Character), args.Error(1)
}

func (m *MockCharacterRepository) Update(ctx context.Context, character *models.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockCharacterRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCharacterRepository) List(ctx context.Context, offset, limit int) ([]*models.Character, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Character), args.Error(1)
}

func TestCharacterService_CreateCharacter(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(MockCharacterRepository)
		service := NewCharacterService(mockRepo)

		character := &models.Character{
			UserID: "user-123",
			Name:   "Thorin",
			Race:   "Dwarf",
			Class:  "Fighter",
			Level:  1,
			Attributes: models.Attributes{
				Strength:     16,
				Dexterity:    12,
				Constitution: 14,
				Intelligence: 10,
				Wisdom:       13,
				Charisma:     8,
			},
		}

		mockRepo.On("Create", ctx, mock.MatchedBy(func(c *models.Character) bool {
			return c.UserID == character.UserID &&
				c.Name == character.Name &&
				c.MaxHitPoints == 12 && // 10 + CON modifier (2)
				c.HitPoints == 12 &&
				c.ArmorClass == 11 && // 10 + DEX modifier (1)
				c.Speed == 30 &&
				c.CarryCapacity == 240 && // STR (16) * 15
				c.AttunementSlotsMax == 3
		})).Return(nil)

		err := service.CreateCharacter(ctx, character)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation errors", func(t *testing.T) {
		mockRepo := new(MockCharacterRepository)
		service := NewCharacterService(mockRepo)

		tests := []struct {
			name      string
			character *models.Character
			errMsg    string
		}{
			{
				name: "missing user ID",
				character: &models.Character{
					Name:  "Test",
					Race:  "Human",
					Class: "Fighter",
				},
				errMsg: "user ID is required",
			},
			{
				name: "missing name",
				character: &models.Character{
					UserID: "user-123",
					Race:   "Human",
					Class:  "Fighter",
				},
				errMsg: "character name is required",
			},
			{
				name: "missing race",
				character: &models.Character{
					UserID: "user-123",
					Name:   "Test",
					Class:  "Fighter",
				},
				errMsg: "character race is required",
			},
			{
				name: "missing class",
				character: &models.Character{
					UserID: "user-123",
					Name:   "Test",
					Race:   "Human",
				},
				errMsg: "character class is required",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := service.CreateCharacter(ctx, tt.character)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			})
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockCharacterRepository)
		service := NewCharacterService(mockRepo)

		character := &models.Character{
			UserID: "user-123",
			Name:   "Test",
			Race:   "Human",
			Class:  "Fighter",
		}

		mockRepo.On("Create", ctx, mock.Anything).Return(errors.New("database error"))

		err := service.CreateCharacter(ctx, character)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

func TestCharacterService_GetAllCharacters(t *testing.T) {
	ctx := context.Background()

	t.Run("get characters for user", func(t *testing.T) {
		mockRepo := new(MockCharacterRepository)
		service := NewCharacterService(mockRepo)

		userID := "user-123"
		expectedChars := []*models.Character{
			{ID: "char-1", UserID: userID, Name: "Char1"},
			{ID: "char-2", UserID: userID, Name: "Char2"},
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(expectedChars, nil)

		chars, err := service.GetAllCharacters(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedChars, chars)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty user ID returns empty list", func(t *testing.T) {
		mockRepo := new(MockCharacterRepository)
		service := NewCharacterService(mockRepo)

		chars, err := service.GetAllCharacters(ctx, "")
		assert.NoError(t, err)
		assert.Empty(t, chars)
		mockRepo.AssertNotCalled(t, "GetByUserID")
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockCharacterRepository)
		service := NewCharacterService(mockRepo)

		userID := "user-123"
		mockRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("database error"))

		chars, err := service.GetAllCharacters(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, chars)
		mockRepo.AssertExpectations(t)
	})
}

func TestCharacterService_UpdateCharacter(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(MockCharacterRepository)
		service := NewCharacterService(mockRepo)

		existingChar := &models.Character{
			ID:        "char-123",
			UserID:    "user-123",
			Name:      "Old Name",
			CreatedAt: timeNow(),
		}

		updateChar := &models.Character{
			ID:   "char-123",
			Name: "New Name",
		}

		mockRepo.On("GetByID", ctx, "char-123").Return(existingChar, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(c *models.Character) bool {
			return c.ID == "char-123" &&
				c.UserID == "user-123" && // Preserved from existing
				c.Name == "New Name" &&
				c.CreatedAt.Equal(existingChar.CreatedAt) // Preserved
		})).Return(nil)

		err := service.UpdateCharacter(ctx, updateChar)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("missing character ID", func(t *testing.T) {
		mockRepo := new(MockCharacterRepository)
		service := NewCharacterService(mockRepo)

		err := service.UpdateCharacter(ctx, &models.Character{Name: "Test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "character ID is required")
		mockRepo.AssertNotCalled(t, "GetByID")
	})

	t.Run("character not found", func(t *testing.T) {
		mockRepo := new(MockCharacterRepository)
		service := NewCharacterService(mockRepo)

		mockRepo.On("GetByID", ctx, "char-123").Return(nil, errors.New("not found"))

		err := service.UpdateCharacter(ctx, &models.Character{ID: "char-123"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "character not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestCharacterService_CalculateHitPoints(t *testing.T) {
	service := &CharacterService{}

	tests := []struct {
		class        string
		level        int
		constitution int
		expected     int
	}{
		{"fighter", 1, 14, 12},    // 10 + 2 (CON mod)
		{"wizard", 1, 10, 6},       // 6 + 0 (CON mod)
		{"barbarian", 1, 16, 15},  // 12 + 3 (CON mod)
		{"fighter", 5, 14, 40},    // 10 + 4*8 + 5*2 (CON mod)
		{"wizard", 3, 12, 15},     // 6 + 2*4 + 3*1 (CON mod)
		{"unknown", 1, 14, 10},    // Default: 8 + 2 (CON mod)
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			hp := service.CalculateHitPoints(tt.class, tt.level, tt.constitution)
			assert.Equal(t, tt.expected, hp)
		})
	}
}

func TestGetModifier(t *testing.T) {
	tests := []struct {
		ability  int
		expected int
	}{
		{1, -5},
		{2, -4},
		{3, -4},
		{4, -3},
		{5, -3},
		{6, -2},
		{7, -2},
		{8, -1},
		{9, -1},
		{10, 0},
		{11, 0},
		{12, 1},
		{13, 1},
		{14, 2},
		{15, 2},
		{16, 3},
		{17, 3},
		{18, 4},
		{19, 4},
		{20, 5},
		{30, 10},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.ability)), func(t *testing.T) {
			mod := getModifier(tt.ability)
			assert.Equal(t, tt.expected, mod)
		})
	}
}

func TestCalculateCarryCapacity(t *testing.T) {
	tests := []struct {
		strength int
		expected float64
	}{
		{10, 150},
		{15, 225},
		{20, 300},
		{8, 120},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.strength)), func(t *testing.T) {
			capacity := CalculateCarryCapacity(tt.strength)
			assert.Equal(t, tt.expected, capacity)
		})
	}
}

// Helper function
func timeNow() time.Time {
	return time.Now().Truncate(time.Second)
}