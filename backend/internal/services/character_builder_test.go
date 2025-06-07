package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-org/dnd-game/internal/models"
	"github.com/your-org/dnd-game/internal/testutil"
)

func TestCharacterBuilder_CreateCharacter(t *testing.T) {
	t.Run("successful character creation with standard race and class", func(t *testing.T) {
		// Setup mocks
		mockCharRepo := new(testutil.MockCharacterRepository)
		mockUserRepo := new(testutil.MockUserRepository)
		
		builder := NewCharacterBuilder(mockCharRepo, mockUserRepo)
		
		// Test data
		user := testutil.NewUserBuilder().WithID(1).Build()
		charInput := &models.CharacterCreationInput{
			Name:  "Aragorn",
			Race:  "Human",
			Class: "Fighter",
			Background: "Noble",
			AbilityScores: models.AbilityScores{
				Strength:     16,
				Dexterity:    14,
				Constitution: 15,
				Intelligence: 10,
				Wisdom:       13,
				Charisma:     12,
			},
		}
		
		// Setup expectations
		mockUserRepo.On("GetByID", user.ID).Return(user, nil)
		mockCharRepo.On("Create", mock.MatchedBy(func(char *models.Character) bool {
			return char.Name == "Aragorn" &&
				char.Race == "Human" &&
				char.Class == "Fighter" &&
				char.Level == 1 &&
				char.Abilities.Strength == 17 && // Human +1 to all
				char.HitPoints == 12 // 10 base + 2 CON modifier
		})).Return(nil)
		
		// Execute
		ctx := testutil.TestContext()
		ctx = context.WithValue(ctx, "user_id", user.ID)
		
		char, err := builder.CreateCharacter(ctx, charInput)
		
		// Assert
		require.NoError(t, err)
		require.NotNil(t, char)
		require.Equal(t, "Aragorn", char.Name)
		require.Equal(t, "Human", char.Race)
		require.Equal(t, "Fighter", char.Class)
		require.Equal(t, 1, char.Level)
		
		// Verify racial bonuses applied
		require.Equal(t, 17, char.Abilities.Strength)
		require.Equal(t, 15, char.Abilities.Dexterity)
		
		// Verify HP calculation
		require.Equal(t, 12, char.MaxHitPoints)
		require.Equal(t, 12, char.HitPoints)
		
		// Verify skills and proficiencies
		require.Contains(t, char.Proficiencies, "All Armor")
		require.Contains(t, char.Proficiencies, "All Weapons")
		
		mockCharRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("character creation with spell-casting class", func(t *testing.T) {
		mockCharRepo := new(testutil.MockCharacterRepository)
		mockUserRepo := new(testutil.MockUserRepository)
		
		builder := NewCharacterBuilder(mockCharRepo, mockUserRepo)
		
		user := testutil.NewUserBuilder().WithID(1).Build()
		charInput := &models.CharacterCreationInput{
			Name:  "Gandalf",
			Race:  "Elf",
			Class: "Wizard",
			Background: "Sage",
			AbilityScores: models.AbilityScores{
				Strength:     8,
				Dexterity:    14,
				Constitution: 14,
				Intelligence: 16,
				Wisdom:       13,
				Charisma:     10,
			},
		}
		
		mockUserRepo.On("GetByID", user.ID).Return(user, nil)
		mockCharRepo.On("Create", mock.MatchedBy(func(char *models.Character) bool {
			// Verify spell slots are initialized
			slots, hasSlots := char.SpellSlots["1"]
			return char.Class == "Wizard" &&
				hasSlots &&
				slots.Total == 2 &&
				slots.Used == 0 &&
				len(char.KnownSpells) > 0
		})).Return(nil)
		
		ctx := testutil.TestContext()
		ctx = context.WithValue(ctx, "user_id", user.ID)
		
		char, err := builder.CreateCharacter(ctx, charInput)
		
		require.NoError(t, err)
		require.NotNil(t, char)
		
		// Verify wizard-specific features
		require.NotEmpty(t, char.SpellSlots)
		require.Equal(t, 2, char.SpellSlots["1"].Total)
		require.Contains(t, char.Equipment, "Spellbook")
		require.Greater(t, len(char.KnownSpells), 0)
		
		mockCharRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("invalid ability scores", func(t *testing.T) {
		mockCharRepo := new(testutil.MockCharacterRepository)
		mockUserRepo := new(testutil.MockUserRepository)
		
		builder := NewCharacterBuilder(mockCharRepo, mockUserRepo)
		
		charInput := &models.CharacterCreationInput{
			Name:  "Invalid",
			Race:  "Human",
			Class: "Fighter",
			AbilityScores: models.AbilityScores{
				Strength:     25, // Too high
				Dexterity:    14,
				Constitution: 15,
				Intelligence: 10,
				Wisdom:       13,
				Charisma:     2, // Too low
			},
		}
		
		ctx := testutil.TestContext()
		ctx = context.WithValue(ctx, "user_id", int64(1))
		
		_, err := builder.CreateCharacter(ctx, charInput)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "ability score")
	})

	t.Run("duplicate character name", func(t *testing.T) {
		mockCharRepo := new(testutil.MockCharacterRepository)
		mockUserRepo := new(testutil.MockUserRepository)
		
		builder := NewCharacterBuilder(mockCharRepo, mockUserRepo)
		
		user := testutil.NewUserBuilder().WithID(1).Build()
		existingChars := []*models.Character{
			testutil.NewCharacterBuilder().WithName("Aragorn").Build(),
		}
		
		charInput := &models.CharacterCreationInput{
			Name:  "Aragorn", // Duplicate name
			Race:  "Human",
			Class: "Fighter",
			AbilityScores: models.AbilityScores{
				Strength:     16,
				Dexterity:    14,
				Constitution: 15,
				Intelligence: 10,
				Wisdom:       13,
				Charisma:     12,
			},
		}
		
		mockUserRepo.On("GetByID", user.ID).Return(user, nil)
		mockCharRepo.On("GetByUserID", user.ID).Return(existingChars, nil)
		
		ctx := testutil.TestContext()
		ctx = context.WithValue(ctx, "user_id", user.ID)
		
		_, err := builder.CreateCharacter(ctx, charInput)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})
}

func TestCharacterBuilder_ApplyRacialBonuses(t *testing.T) {
	builder := &CharacterBuilder{}
	
	tests := []struct {
		name          string
		race          string
		baseAbilities models.AbilityScores
		expected      models.AbilityScores
	}{
		{
			name: "human bonus (+1 to all)",
			race: "Human",
			baseAbilities: models.AbilityScores{
				Strength:     10,
				Dexterity:    10,
				Constitution: 10,
				Intelligence: 10,
				Wisdom:       10,
				Charisma:     10,
			},
			expected: models.AbilityScores{
				Strength:     11,
				Dexterity:    11,
				Constitution: 11,
				Intelligence: 11,
				Wisdom:       11,
				Charisma:     11,
			},
		},
		{
			name: "elf bonus (+2 DEX)",
			race: "Elf",
			baseAbilities: models.AbilityScores{
				Strength:     10,
				Dexterity:    14,
				Constitution: 10,
				Intelligence: 10,
				Wisdom:       10,
				Charisma:     10,
			},
			expected: models.AbilityScores{
				Strength:     10,
				Dexterity:    16,
				Constitution: 10,
				Intelligence: 10,
				Wisdom:       10,
				Charisma:     10,
			},
		},
		{
			name: "dwarf bonus (+2 CON)",
			race: "Dwarf",
			baseAbilities: models.AbilityScores{
				Strength:     10,
				Dexterity:    10,
				Constitution: 14,
				Intelligence: 10,
				Wisdom:       10,
				Charisma:     10,
			},
			expected: models.AbilityScores{
				Strength:     10,
				Dexterity:    10,
				Constitution: 16,
				Intelligence: 10,
				Wisdom:       10,
				Charisma:     10,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.applyRacialBonuses(tt.race, tt.baseAbilities)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCharacterBuilder_CalculateHP(t *testing.T) {
	builder := &CharacterBuilder{}
	
	tests := []struct {
		name     string
		class    string
		conMod   int
		expected int
	}{
		{
			name:     "fighter with +2 CON",
			class:    "Fighter",
			conMod:   2,
			expected: 12, // 10 base + 2 CON
		},
		{
			name:     "wizard with +1 CON",
			class:    "Wizard",
			conMod:   1,
			expected: 7, // 6 base + 1 CON
		},
		{
			name:     "barbarian with +3 CON",
			class:    "Barbarian",
			conMod:   3,
			expected: 15, // 12 base + 3 CON
		},
		{
			name:     "rogue with -1 CON",
			class:    "Rogue",
			conMod:   -1,
			expected: 7, // 8 base - 1 CON (minimum 1, but 7 is still valid)
		},
		{
			name:     "wizard with -2 CON",
			class:    "Wizard",
			conMod:   -2,
			expected: 4, // 6 base - 2 CON
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.calculateHP(tt.class, tt.conMod)
			require.Equal(t, tt.expected, result)
			require.Greater(t, result, 0, "HP must always be positive")
		})
	}
}

func TestCharacterBuilder_InitializeSpellSlots(t *testing.T) {
	builder := &CharacterBuilder{}
	
	tests := []struct {
		name     string
		class    string
		level    int
		expected map[string]models.SpellSlotInfo
	}{
		{
			name:  "level 1 wizard",
			class: "Wizard",
			level: 1,
			expected: map[string]models.SpellSlotInfo{
				"1": {Total: 2, Used: 0},
			},
		},
		{
			name:  "level 3 cleric",
			class: "Cleric",
			level: 3,
			expected: map[string]models.SpellSlotInfo{
				"1": {Total: 4, Used: 0},
				"2": {Total: 2, Used: 0},
			},
		},
		{
			name:  "level 5 sorcerer",
			class: "Sorcerer",
			level: 5,
			expected: map[string]models.SpellSlotInfo{
				"1": {Total: 4, Used: 0},
				"2": {Total: 3, Used: 0},
				"3": {Total: 2, Used: 0},
			},
		},
		{
			name:     "level 1 fighter (no spells)",
			class:    "Fighter",
			level:    1,
			expected: map[string]models.SpellSlotInfo{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.initializeSpellSlots(tt.class, tt.level)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCharacterBuilder_AssignStartingEquipment(t *testing.T) {
	builder := &CharacterBuilder{}
	
	tests := []struct {
		name          string
		class         string
		background    string
		expectedItems []string
	}{
		{
			name:       "fighter with soldier background",
			class:      "Fighter",
			background: "Soldier",
			expectedItems: []string{
				"Chain Mail",
				"Longsword",
				"Shield",
				"Insignia of Rank",
			},
		},
		{
			name:       "wizard with sage background",
			class:      "Wizard",
			background: "Sage",
			expectedItems: []string{
				"Spellbook",
				"Component Pouch",
				"Scholar's Pack",
				"Letter from Colleague",
			},
		},
		{
			name:       "rogue with criminal background",
			class:      "Rogue",
			background: "Criminal",
			expectedItems: []string{
				"Leather Armor",
				"Shortsword",
				"Thieves' Tools",
				"Crowbar",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.assignStartingEquipment(tt.class, tt.background)
			
			for _, item := range tt.expectedItems {
				require.Contains(t, result, item, "Should have %s", item)
			}
		})
	}
}

func TestCharacterBuilder_WithCustomRaceAndClass(t *testing.T) {
	// This test would check integration with custom races/classes
	t.Run("custom race application", func(t *testing.T) {
		mockCharRepo := new(testutil.MockCharacterRepository)
		mockUserRepo := new(testutil.MockUserRepository)
		mockCustomRaceRepo := new(MockCustomRaceRepository)
		
		builder := NewCharacterBuilder(mockCharRepo, mockUserRepo)
		builder.customRaceRepo = mockCustomRaceRepo
		
		user := testutil.NewUserBuilder().WithID(1).Build()
		customRace := &models.CustomRace{
			ID:     1,
			UserID: 1,
			Name:   "Dragonkin",
			Traits: models.RaceTraits{
				AbilityScoreIncreases: map[string]int{
					"strength":    2,
					"charisma":    1,
				},
				Speed:       30,
				Size:        "Medium",
				Languages:   []string{"Common", "Draconic"},
				Features:    []string{"Breath Weapon", "Damage Resistance"},
			},
		}
		
		charInput := &models.CharacterCreationInput{
			Name:         "Draconis",
			CustomRaceID: &customRace.ID,
			Class:        "Paladin",
			AbilityScores: models.AbilityScores{
				Strength:     14,
				Dexterity:    10,
				Constitution: 13,
				Intelligence: 10,
				Wisdom:       12,
				Charisma:     14,
			},
		}
		
		mockUserRepo.On("GetByID", user.ID).Return(user, nil)
		mockCustomRaceRepo.On("GetByID", customRace.ID).Return(customRace, nil)
		mockCharRepo.On("Create", mock.MatchedBy(func(char *models.Character) bool {
			return char.Race == "Dragonkin" &&
				char.Abilities.Strength == 16 && // 14 + 2
				char.Abilities.Charisma == 15 && // 14 + 1
				char.Speed == 30
		})).Return(nil)
		
		ctx := testutil.TestContext()
		ctx = context.WithValue(ctx, "user_id", user.ID)
		
		char, err := builder.CreateCharacter(ctx, charInput)
		
		require.NoError(t, err)
		require.NotNil(t, char)
		require.Equal(t, "Dragonkin", char.Race)
		require.Contains(t, char.Traits, "Breath Weapon")
		
		mockCharRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockCustomRaceRepo.AssertExpectations(t)
	})
}

// Mock for custom race repository
type MockCustomRaceRepository struct {
	mock.Mock
}

func (m *MockCustomRaceRepository) GetByID(id int64) (*models.CustomRace, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CustomRace), args.Error(1)
}