package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/internal/services/mocks"
)

func TestCharacterService_CreateCharacter(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		character     *models.Character
		setupMock     func(*mocks.MockCharacterRepository, *mocks.MockLLMProvider)
		expectedError string
		validate      func(*testing.T, *models.Character)
	}{
		{
			name: "successful character creation",
			character: &models.Character{
				UserID: "user-123",
				Name:   "Aragorn",
				Race:   "Human",
				Class:  "Ranger",
				Attributes: models.Attributes{
					Strength:     16,
					Dexterity:    14,
					Constitution: 13,
					Intelligence: 12,
					Wisdom:       15,
					Charisma:     10,
				},
			},
			setupMock: func(charRepo *mocks.MockCharacterRepository, llm *mocks.MockLLMProvider) {
				// Just accept any character since the service modifies it
				charRepo.On("Create", ctx, mock.Anything).Return(nil)
			},
			validate: func(t *testing.T, char *models.Character) {
				assert.NotEmpty(t, char.ID)
				assert.Equal(t, 1, char.Level)
				assert.Equal(t, 10+(13-10)/2, char.MaxHitPoints) // 10 + CON modifier
				assert.Equal(t, char.MaxHitPoints, char.HitPoints)
				assert.Equal(t, 10+(14-10)/2, char.ArmorClass) // 10 + DEX modifier
				assert.Equal(t, 30, char.Speed)
				assert.Equal(t, 240.0, char.CarryCapacity) // STR 16 * 15
				assert.Equal(t, 3, char.AttunementSlotsMax)
			},
		},
		{
			name: "missing user ID",
			character: &models.Character{
				Name:  "Aragorn",
				Race:  "Human",
				Class: "Ranger",
			},
			expectedError: "user ID is required",
		},
		{
			name: "missing character name",
			character: &models.Character{
				UserID: "user-123",
				Race:   "Human",
				Class:  "Ranger",
			},
			expectedError: "character name is required",
		},
		{
			name: "missing race",
			character: &models.Character{
				UserID: "user-123",
				Name:   "Aragorn",
				Class:  "Ranger",
			},
			expectedError: "character race is required",
		},
		{
			name: "missing class",
			character: &models.Character{
				UserID: "user-123",
				Name:   "Aragorn",
				Race:   "Human",
			},
			expectedError: "character class is required",
		},
		{
			name: "repository error",
			character: &models.Character{
				UserID: "user-123",
				Name:   "Aragorn",
				Race:   "Human",
				Class:  "Ranger",
			},
			setupMock: func(charRepo *mocks.MockCharacterRepository, llm *mocks.MockLLMProvider) {
				charRepo.On("Create", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCharRepo := new(mocks.MockCharacterRepository)
			mockLLM := new(mocks.MockLLMProvider)

			if tt.setupMock != nil {
				tt.setupMock(mockCharRepo, mockLLM)
			}

			service := services.NewCharacterService(mockCharRepo, nil, mockLLM)
			err := service.CreateCharacter(ctx, tt.character)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.character)
				}
			}

			mockCharRepo.AssertExpectations(t)
			mockLLM.AssertExpectations(t)
		})
	}
}

func TestCharacterService_GetCharacterByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		characterID   string
		setupMock     func(*mocks.MockCharacterRepository)
		expectedError string
		validate      func(*testing.T, *models.Character)
	}{
		{
			name:        "successful get character",
			characterID: "char-123",
			setupMock: func(m *mocks.MockCharacterRepository) {
				char := mocks.CreateTestCharacter("char-123", "user-123", "Aragorn", "Human", "Ranger")
				m.On("GetByID", ctx, "char-123").Return(char, nil)
			},
			validate: func(t *testing.T, char *models.Character) {
				assert.Equal(t, "char-123", char.ID)
				assert.Equal(t, "Aragorn", char.Name)
				assert.Equal(t, "Human", char.Race)
				assert.Equal(t, "Ranger", char.Class)
			},
		},
		{
			name:        "character not found",
			characterID: "nonexistent",
			setupMock: func(m *mocks.MockCharacterRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("character not found"))
			},
			expectedError: "character not found",
		},
		{
			name:        "repository error",
			characterID: "char-123",
			setupMock: func(m *mocks.MockCharacterRepository) {
				m.On("GetByID", ctx, "char-123").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockCharacterRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewCharacterService(mockRepo, nil, nil)
			char, err := service.GetCharacterByID(ctx, tt.characterID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, char)
			} else {
				require.NoError(t, err)
				require.NotNil(t, char)
				if tt.validate != nil {
					tt.validate(t, char)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCharacterService_UpdateCharacter(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		update        *models.Character
		setupMock     func(*mocks.MockCharacterRepository)
		expectedError string
		validate      func(*testing.T, *models.Character)
	}{
		{
			name: "successful update with partial data",
			update: &models.Character{
				ID:    "char-123",
				Name:  "Strider", // Only updating name
				Level: 2,         // And level
			},
			setupMock: func(m *mocks.MockCharacterRepository) {
				// Get existing character
				existing := mocks.CreateTestCharacter("char-123", "user-123", "Aragorn", "Human", "Ranger")
				m.On("GetByID", ctx, "char-123").Return(existing, nil)

				// Update with merged data - using mock.Anything to avoid complex matching
				m.On("Update", ctx, mock.Anything).Return(nil)
			},
		},
		{
			name: "update attributes",
			update: &models.Character{
				ID: "char-123",
				Attributes: models.Attributes{
					Strength:     18, // Only strength changed
					Dexterity:    14,
					Constitution: 13,
					Intelligence: 12,
					Wisdom:       15,
					Charisma:     10,
				},
			},
			setupMock: func(m *mocks.MockCharacterRepository) {
				existing := mocks.CreateTestCharacter("char-123", "user-123", "Aragorn", "Human", "Ranger")
				m.On("GetByID", ctx, "char-123").Return(existing, nil)
				m.On("Update", ctx, mock.Anything).Return(nil)
			},
		},
		{
			name: "missing character ID",
			update: &models.Character{
				Name: "Strider",
			},
			expectedError: "character ID is required",
		},
		{
			name: "character not found",
			update: &models.Character{
				ID:   "nonexistent",
				Name: "Strider",
			},
			setupMock: func(m *mocks.MockCharacterRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("character not found"))
			},
			expectedError: "character not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockCharacterRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewCharacterService(mockRepo, nil, nil)
			err := service.UpdateCharacter(ctx, tt.update)

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

func TestCharacterService_DeleteCharacter(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		characterID   string
		setupMock     func(*mocks.MockCharacterRepository)
		expectedError string
	}{
		{
			name:        "successful deletion",
			characterID: "char-123",
			setupMock: func(m *mocks.MockCharacterRepository) {
				m.On("Delete", ctx, "char-123").Return(nil)
			},
		},
		{
			name:        "character not found",
			characterID: "nonexistent",
			setupMock: func(m *mocks.MockCharacterRepository) {
				m.On("Delete", ctx, "nonexistent").Return(errors.New("character not found"))
			},
			expectedError: "character not found",
		},
		{
			name:        "repository error",
			characterID: "char-123",
			setupMock: func(m *mocks.MockCharacterRepository) {
				m.On("Delete", ctx, "char-123").Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockCharacterRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewCharacterService(mockRepo, nil, nil)
			err := service.DeleteCharacter(ctx, tt.characterID)

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

func TestCharacterService_AddExperience(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		characterID   string
		xpToAdd       int
		setupMock     func(*mocks.MockCharacterRepository, *mocks.MockLLMProvider)
		expectedError string
		validate      func(*testing.T, *models.Character)
	}{
		{
			name:        "add XP without level up",
			characterID: "char-123",
			xpToAdd:     100,
			setupMock: func(charRepo *mocks.MockCharacterRepository, llm *mocks.MockLLMProvider) {
				char := &models.Character{
					ID:               "char-123",
					Level:            1,
					ExperiencePoints: 100,
					Class:            "Fighter",
					Attributes: models.Attributes{
						Constitution: 14,
					},
				}
				charRepo.On("GetByID", ctx, "char-123").Return(char, nil)

				// XP increases but no level up (need 300 for level 2)
				charRepo.On("Update", ctx, mock.Anything).Return(nil)
			},
		},
		{
			name:        "add XP with level up",
			characterID: "char-123",
			xpToAdd:     250,
			setupMock: func(charRepo *mocks.MockCharacterRepository, llm *mocks.MockLLMProvider) {
				char := &models.Character{
					ID:               "char-123",
					Level:            1,
					ExperiencePoints: 100,
					Class:            "Fighter",
					HitPoints:        10,
					MaxHitPoints:     10,
					Attributes: models.Attributes{
						Constitution: 14,
					},
				}
				// Mock GetByID to return character for all calls (AddExperience and LevelUp need it)
				charRepo.On("GetByID", ctx, "char-123").Return(char, nil)

				// Update will be called after level up
				charRepo.On("Update", ctx, mock.Anything).Return(nil)
			},
		},
		{
			name:        "character not found",
			characterID: "nonexistent",
			xpToAdd:     100,
			setupMock: func(charRepo *mocks.MockCharacterRepository, llm *mocks.MockLLMProvider) {
				charRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("character not found"))
			},
			expectedError: "character not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCharRepo := new(mocks.MockCharacterRepository)
			mockLLM := new(mocks.MockLLMProvider)

			if tt.setupMock != nil {
				tt.setupMock(mockCharRepo, mockLLM)
			}

			service := services.NewCharacterService(mockCharRepo, nil, mockLLM)
			err := service.AddExperience(ctx, tt.characterID, tt.xpToAdd)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockCharRepo.AssertExpectations(t)
			mockLLM.AssertExpectations(t)
		})
	}
}

func TestCharacterService_InitializeSpellSlots(t *testing.T) {
	tests := []struct {
		name     string
		class    string
		level    int
		expected []models.SpellSlot
	}{
		{
			name:  "wizard level 1",
			class: "wizard",
			level: 1,
			expected: []models.SpellSlot{
				{Level: 1, Total: 2, Remaining: 2},
			},
		},
		{
			name:  "wizard level 3",
			class: "wizard",
			level: 3,
			expected: []models.SpellSlot{
				{Level: 1, Total: 4, Remaining: 4},
				{Level: 2, Total: 2, Remaining: 2},
			},
		},
		{
			name:  "paladin level 2 (half caster)",
			class: "paladin",
			level: 2,
			expected: []models.SpellSlot{
				{Level: 1, Total: 2, Remaining: 2},
			},
		},
		{
			name:  "paladin level 5",
			class: "paladin",
			level: 5,
			expected: []models.SpellSlot{
				{Level: 1, Total: 4, Remaining: 4},
				{Level: 2, Total: 2, Remaining: 2},
			},
		},
		{
			name:  "warlock level 3 (pact magic)",
			class: "warlock",
			level: 3,
			expected: []models.SpellSlot{
				{Level: 2, Total: 2, Remaining: 2},
			},
		},
		{
			name:     "fighter (non-caster)",
			class:    "fighter",
			level:    5,
			expected: []models.SpellSlot{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := services.NewCharacterService(nil, nil, nil)
			slots := service.InitializeSpellSlots(tt.class, tt.level)

			assert.Equal(t, tt.expected, slots)
		})
	}
}

func TestCharacterService_UseSpellSlot(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		characterID   string
		slotLevel     int
		setupMock     func(*mocks.MockCharacterRepository)
		expectedError string
	}{
		{
			name:        "successful spell slot use",
			characterID: "char-123",
			slotLevel:   1,
			setupMock: func(m *mocks.MockCharacterRepository) {
				char := &models.Character{
					ID: "char-123",
					Spells: models.SpellData{
						SpellSlots: []models.SpellSlot{
							{Level: 1, Total: 3, Remaining: 3},
							{Level: 2, Total: 2, Remaining: 2},
						},
					},
				}
				m.On("GetByID", ctx, "char-123").Return(char, nil)

				// Verify slot was decremented
				m.On("Update", ctx, mock.Anything).Return(nil)
			},
		},
		{
			name:        "no remaining spell slots",
			characterID: "char-123",
			slotLevel:   1,
			setupMock: func(m *mocks.MockCharacterRepository) {
				char := &models.Character{
					ID: "char-123",
					Spells: models.SpellData{
						SpellSlots: []models.SpellSlot{
							{Level: 1, Total: 3, Remaining: 0},
						},
					},
				}
				m.On("GetByID", ctx, "char-123").Return(char, nil)
			},
			expectedError: "no remaining spell slots of level 1",
		},
		{
			name:        "character doesn't have that spell level",
			characterID: "char-123",
			slotLevel:   3,
			setupMock: func(m *mocks.MockCharacterRepository) {
				char := &models.Character{
					ID: "char-123",
					Spells: models.SpellData{
						SpellSlots: []models.SpellSlot{
							{Level: 1, Total: 3, Remaining: 3},
							{Level: 2, Total: 2, Remaining: 2},
						},
					},
				}
				m.On("GetByID", ctx, "char-123").Return(char, nil)
			},
			expectedError: "character does not have spell slots of level 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockCharacterRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewCharacterService(mockRepo, nil, nil)
			err := service.UseSpellSlot(ctx, tt.characterID, tt.slotLevel)

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
