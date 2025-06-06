package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// MockCombatRepository is a mock implementation of CombatRepository
type MockCombatRepository struct {
	mock.Mock
}

// MockDiceRoller is a mock implementation of DiceRoller
type MockDiceRoller struct {
	mock.Mock
}

func (m *MockDiceRoller) Roll(dice string) (int, error) {
	args := m.Called(dice)
	return args.Int(0), args.Error(1)
}

func (m *MockDiceRoller) RollWithDetails(dice string) (*models.RollDetails, error) {
	args := m.Called(dice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RollDetails), args.Error(1)
}

func TestCombatService_InitiativeRoll(t *testing.T) {
	tests := []struct {
		name             string
		combatants       []models.Combatant
		setupMock        func(*MockDiceRoller)
		expectedOrder    []string // Expected order of combatant IDs
		expectedInitiatives map[string]int
	}{
		{
			name: "roll initiative for all combatants",
			combatants: []models.Combatant{
				{ID: "char-1", Name: "Fighter", Type: models.CombatantTypeCharacter, Initiative: 0},
				{ID: "char-2", Name: "Wizard", Type: models.CombatantTypeCharacter, Initiative: 0},
				{ID: "npc-1", Name: "Goblin", Type: models.CombatantTypeNPC, Initiative: 0},
			},
			setupMock: func(m *MockDiceRoller) {
				// Fighter rolls 15 + 2 (DEX) = 17
				m.On("Roll", "1d20+2").Return(17, nil).Once()
				// Wizard rolls 12 + 1 (DEX) = 13
				m.On("Roll", "1d20+1").Return(13, nil).Once()
				// Goblin rolls 8 + 2 (DEX) = 10
				m.On("Roll", "1d20+2").Return(10, nil).Once()
			},
			expectedOrder: []string{"char-1", "char-2", "npc-1"},
			expectedInitiatives: map[string]int{
				"char-1": 17,
				"char-2": 13,
				"npc-1":  10,
			},
		},
		{
			name: "handle tied initiatives",
			combatants: []models.Combatant{
				{ID: "char-1", Name: "Fighter", Type: models.CombatantTypeCharacter, Initiative: 0},
				{ID: "char-2", Name: "Wizard", Type: models.CombatantTypeCharacter, Initiative: 0},
			},
			setupMock: func(m *MockDiceRoller) {
				// Both roll the same
				m.On("Roll", "1d20+2").Return(15, nil).Once()
				m.On("Roll", "1d20+1").Return(15, nil).Once()
				// Tiebreaker rolls
				m.On("Roll", "1d20").Return(12, nil).Once()
				m.On("Roll", "1d20").Return(8, nil).Once()
			},
			expectedOrder: []string{"char-1", "char-2"},
			expectedInitiatives: map[string]int{
				"char-1": 15,
				"char-2": 15,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDice := new(MockDiceRoller)
			if tt.setupMock != nil {
				tt.setupMock(mockDice)
			}

			service := &CombatService{diceRoller: mockDice}
			result := service.RollInitiative(tt.combatants)

			// Check the order
			for i, expectedID := range tt.expectedOrder {
				assert.Equal(t, expectedID, result[i].ID)
			}

			// Check initiatives
			for _, combatant := range result {
				if expected, ok := tt.expectedInitiatives[combatant.ID]; ok {
					assert.Equal(t, expected, combatant.Initiative)
				}
			}

			mockDice.AssertExpectations(t)
		})
	}
}

func TestCombatService_Attack(t *testing.T) {
	tests := []struct {
		name          string
		attacker      *models.Combatant
		target        *models.Combatant
		weapon        *models.Weapon
		setupMock     func(*MockDiceRoller)
		expectedHit   bool
		expectedDamage int
		expectedError string
	}{
		{
			name: "successful hit",
			attacker: &models.Combatant{
				ID:   "char-1",
				Name: "Fighter",
				Stats: models.CombatantStats{
					Strength:  16,
					Dexterity: 14,
				},
			},
			target: &models.Combatant{
				ID: "npc-1",
				Name: "Goblin",
				AC: 15,
				HP: 7,
			},
			weapon: &models.Weapon{
				Name:       "Longsword",
				Damage:     "1d8",
				DamageType: "slashing",
				Properties: []string{"versatile"},
			},
			setupMock: func(m *MockDiceRoller) {
				// Attack roll: 1d20 + 3 (STR) + 2 (proficiency) = 18
				m.On("RollWithDetails", "1d20+5").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 20, Result: 13}},
					Modifier: 5,
					Total:    18,
				}, nil)
				// Damage roll: 1d8 + 3 (STR) = 8
				m.On("RollWithDetails", "1d8+3").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 8, Result: 5}},
					Modifier: 3,
					Total:    8,
				}, nil)
			},
			expectedHit:    true,
			expectedDamage: 8,
		},
		{
			name: "critical hit",
			attacker: &models.Combatant{
				ID:   "char-1",
				Name: "Rogue",
				Stats: models.CombatantStats{
					Strength:  10,
					Dexterity: 18,
				},
			},
			target: &models.Combatant{
				ID: "npc-1",
				Name: "Orc",
				AC: 13,
				HP: 15,
			},
			weapon: &models.Weapon{
				Name:       "Dagger",
				Damage:     "1d4",
				DamageType: "piercing",
				Properties: []string{"finesse", "light"},
			},
			setupMock: func(m *MockDiceRoller) {
				// Natural 20!
				m.On("RollWithDetails", "1d20+6").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 20, Result: 20}},
					Modifier: 6,
					Total:    26,
				}, nil)
				// Critical damage: 2d4 + 4 = 10
				m.On("RollWithDetails", "2d4+4").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 4, Result: 3}, {Sides: 4, Result: 3}},
					Modifier: 4,
					Total:    10,
				}, nil)
			},
			expectedHit:    true,
			expectedDamage: 10,
		},
		{
			name: "miss",
			attacker: &models.Combatant{
				ID:   "char-1",
				Name: "Wizard",
				Stats: models.CombatantStats{
					Strength:  8,
					Dexterity: 12,
				},
			},
			target: &models.Combatant{
				ID: "npc-1",
				Name: "Knight",
				AC: 18,
				HP: 52,
			},
			weapon: &models.Weapon{
				Name:       "Staff",
				Damage:     "1d6",
				DamageType: "bludgeoning",
			},
			setupMock: func(m *MockDiceRoller) {
				// Attack roll: too low
				m.On("RollWithDetails", "1d20+1").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 20, Result: 10}},
					Modifier: 1,
					Total:    11,
				}, nil)
			},
			expectedHit:    false,
			expectedDamage: 0,
		},
		{
			name: "natural 1 critical miss",
			attacker: &models.Combatant{
				ID:   "char-1",
				Name: "Fighter",
				Stats: models.CombatantStats{
					Strength: 16,
				},
			},
			target: &models.Combatant{
				ID: "npc-1",
				AC: 10,
				HP: 5,
			},
			weapon: &models.Weapon{
				Name:   "Greatsword",
				Damage: "2d6",
			},
			setupMock: func(m *MockDiceRoller) {
				// Natural 1!
				m.On("RollWithDetails", "1d20+5").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 20, Result: 1}},
					Modifier: 5,
					Total:    6,
				}, nil)
			},
			expectedHit:    false,
			expectedDamage: 0,
		},
		{
			name:          "nil attacker",
			attacker:      nil,
			target:        &models.Combatant{ID: "target"},
			weapon:        &models.Weapon{},
			expectedError: "attacker cannot be nil",
		},
		{
			name:          "nil target",
			attacker:      &models.Combatant{ID: "attacker"},
			target:        nil,
			weapon:        &models.Weapon{},
			expectedError: "target cannot be nil",
		},
		{
			name:          "nil weapon",
			attacker:      &models.Combatant{ID: "attacker"},
			target:        &models.Combatant{ID: "target"},
			weapon:        nil,
			expectedError: "weapon cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDice := new(MockDiceRoller)
			if tt.setupMock != nil {
				tt.setupMock(mockDice)
			}

			service := &CombatService{diceRoller: mockDice}
			hit, damage, err := service.Attack(tt.attacker, tt.target, tt.weapon)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedHit, hit)
				assert.Equal(t, tt.expectedDamage, damage)
			}

			mockDice.AssertExpectations(t)
		})
	}
}

func TestCombatService_SavingThrow(t *testing.T) {
	tests := []struct {
		name         string
		character    *models.Character
		saveType     string
		dc           int
		setupMock    func(*MockDiceRoller)
		expected     bool
		expectedRoll int
		expectedError string
	}{
		{
			name: "successful strength save",
			character: &models.Character{
				SavingThrows: models.SavingThrows{
					Strength: 5,
				},
			},
			saveType: "strength",
			dc:       15,
			setupMock: func(m *MockDiceRoller) {
				m.On("RollWithDetails", "1d20+5").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 20, Result: 12}},
					Modifier: 5,
					Total:    17,
				}, nil)
			},
			expected:     true,
			expectedRoll: 17,
		},
		{
			name: "failed dexterity save",
			character: &models.Character{
				SavingThrows: models.SavingThrows{
					Dexterity: 2,
				},
			},
			saveType: "dexterity",
			dc:       18,
			setupMock: func(m *MockDiceRoller) {
				m.On("RollWithDetails", "1d20+2").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 20, Result: 10}},
					Modifier: 2,
					Total:    12,
				}, nil)
			},
			expected:     false,
			expectedRoll: 12,
		},
		{
			name: "natural 20 auto success",
			character: &models.Character{
				SavingThrows: models.SavingThrows{
					Wisdom: -2,
				},
			},
			saveType: "wisdom",
			dc:       25,
			setupMock: func(m *MockDiceRoller) {
				m.On("RollWithDetails", "1d20-2").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 20, Result: 20}},
					Modifier: -2,
					Total:    18,
				}, nil)
			},
			expected:     true,
			expectedRoll: 18,
		},
		{
			name: "natural 1 auto fail",
			character: &models.Character{
				SavingThrows: models.SavingThrows{
					Constitution: 8,
				},
			},
			saveType: "constitution",
			dc:       5,
			setupMock: func(m *MockDiceRoller) {
				m.On("RollWithDetails", "1d20+8").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 20, Result: 1}},
					Modifier: 8,
					Total:    9,
				}, nil)
			},
			expected:     false,
			expectedRoll: 9,
		},
		{
			name:         "invalid save type",
			character:    &models.Character{},
			saveType:     "invalid",
			dc:           15,
			expectedError: "invalid save type",
		},
		{
			name:         "nil character",
			character:    nil,
			saveType:     "strength",
			dc:           15,
			expectedError: "character cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDice := new(MockDiceRoller)
			if tt.setupMock != nil {
				tt.setupMock(mockDice)
			}

			service := &CombatService{diceRoller: mockDice}
			success, roll, err := service.SavingThrow(tt.character, tt.saveType, tt.dc)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, success)
				assert.Equal(t, tt.expectedRoll, roll)
			}

			mockDice.AssertExpectations(t)
		})
	}
}

func TestCombatService_SkillCheck(t *testing.T) {
	tests := []struct {
		name         string
		character    *models.Character
		skill        string
		dc           int
		setupMock    func(*MockDiceRoller)
		expected     bool
		expectedRoll int
		expectedError string
	}{
		{
			name: "successful stealth check",
			character: &models.Character{
				Skills: models.Skills{
					Stealth: 5,
				},
			},
			skill: "stealth",
			dc:    15,
			setupMock: func(m *MockDiceRoller) {
				m.On("RollWithDetails", "1d20+5").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 20, Result: 11}},
					Modifier: 5,
					Total:    16,
				}, nil)
			},
			expected:     true,
			expectedRoll: 16,
		},
		{
			name: "failed perception check",
			character: &models.Character{
				Skills: models.Skills{
					Perception: 3,
				},
			},
			skill: "perception",
			dc:    20,
			setupMock: func(m *MockDiceRoller) {
				m.On("RollWithDetails", "1d20+3").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 20, Result: 8}},
					Modifier: 3,
					Total:    11,
				}, nil)
			},
			expected:     false,
			expectedRoll: 11,
		},
		{
			name: "athletics check with negative modifier",
			character: &models.Character{
				Skills: models.Skills{
					Athletics: -1,
				},
			},
			skill: "athletics",
			dc:    10,
			setupMock: func(m *MockDiceRoller) {
				m.On("RollWithDetails", "1d20-1").Return(&models.RollDetails{
					Dice:     []models.DieResult{{Sides: 20, Result: 12}},
					Modifier: -1,
					Total:    11,
				}, nil)
			},
			expected:     true,
			expectedRoll: 11,
		},
		{
			name:         "invalid skill",
			character:    &models.Character{},
			skill:        "invalid_skill",
			dc:           15,
			expectedError: "invalid skill",
		},
		{
			name:         "nil character",
			character:    nil,
			skill:        "stealth",
			dc:           15,
			expectedError: "character cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDice := new(MockDiceRoller)
			if tt.setupMock != nil {
				tt.setupMock(mockDice)
			}

			service := &CombatService{diceRoller: mockDice}
			success, roll, err := service.SkillCheck(tt.character, tt.skill, tt.dc)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, success)
				assert.Equal(t, tt.expectedRoll, roll)
			}

			mockDice.AssertExpectations(t)
		})
	}
}

func TestCombatService_ApplyDamage(t *testing.T) {
	tests := []struct {
		name             string
		target           *models.Combatant
		damage           int
		damageType       string
		expectedHP       int
		expectedStatus   string
		expectedError    string
	}{
		{
			name: "normal damage",
			target: &models.Combatant{
				ID:    "char-1",
				Name:  "Fighter",
				HP:    45,
				MaxHP: 45,
			},
			damage:         10,
			damageType:     "slashing",
			expectedHP:     35,
			expectedStatus: "alive",
		},
		{
			name: "damage reduces to 0",
			target: &models.Combatant{
				ID:    "char-1",
				Name:  "Wizard",
				HP:    8,
				MaxHP: 25,
			},
			damage:         15,
			damageType:     "fire",
			expectedHP:     0,
			expectedStatus: "unconscious",
		},
		{
			name: "massive damage instant death",
			target: &models.Combatant{
				ID:    "char-1",
				Name:  "Commoner",
				HP:    4,
				MaxHP: 4,
			},
			damage:         12, // More than max HP
			damageType:     "bludgeoning",
			expectedHP:     0,
			expectedStatus: "dead",
		},
		{
			name: "resistance halves damage",
			target: &models.Combatant{
				ID:         "char-1",
				Name:       "Barbarian",
				HP:         50,
				MaxHP:      50,
				Resistances: []string{"slashing", "piercing", "bludgeoning"},
			},
			damage:         20,
			damageType:     "slashing",
			expectedHP:     40, // 20 / 2 = 10 damage taken
			expectedStatus: "alive",
		},
		{
			name: "vulnerability doubles damage",
			target: &models.Combatant{
				ID:              "npc-1",
				Name:            "Fire Elemental",
				HP:              30,
				MaxHP:           30,
				Vulnerabilities: []string{"cold"},
			},
			damage:         10,
			damageType:     "cold",
			expectedHP:     10, // 10 * 2 = 20 damage taken
			expectedStatus: "alive",
		},
		{
			name: "immunity negates damage",
			target: &models.Combatant{
				ID:         "npc-1",
				Name:       "Ghost",
				HP:         20,
				MaxHP:      20,
				Immunities: []string{"necrotic", "poison"},
			},
			damage:         50,
			damageType:     "necrotic",
			expectedHP:     20, // No damage taken
			expectedStatus: "alive",
		},
		{
			name:          "nil target",
			target:        nil,
			damage:        10,
			damageType:    "fire",
			expectedError: "target cannot be nil",
		},
		{
			name: "negative damage (healing)",
			target: &models.Combatant{
				ID:    "char-1",
				HP:    10,
				MaxHP: 30,
			},
			damage:         -15,
			damageType:     "healing",
			expectedHP:     25,
			expectedStatus: "alive",
		},
		{
			name: "healing cannot exceed max HP",
			target: &models.Combatant{
				ID:    "char-1",
				HP:    25,
				MaxHP: 30,
			},
			damage:         -10,
			damageType:     "healing",
			expectedHP:     30,
			expectedStatus: "alive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &CombatService{}
			err := service.ApplyDamage(tt.target, tt.damage, tt.damageType)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedHP, tt.target.HP)
				assert.Equal(t, tt.expectedStatus, tt.target.Status)
			}
		})
	}
}

func TestCombatService_CalculateAC(t *testing.T) {
	tests := []struct {
		name      string
		character *models.Character
		armor     *models.Armor
		shield    bool
		expected  int
	}{
		{
			name: "no armor",
			character: &models.Character{
				AbilityScores: models.AbilityScores{
					Dexterity: 14,
				},
			},
			armor:    nil,
			shield:   false,
			expected: 12, // 10 + 2 (DEX)
		},
		{
			name: "light armor",
			character: &models.Character{
				AbilityScores: models.AbilityScores{
					Dexterity: 16,
				},
			},
			armor: &models.Armor{
				Name:     "Leather Armor",
				Type:     "light",
				BaseAC:   11,
				MaxDexBonus: 10, // No limit for light armor
			},
			shield:   false,
			expected: 14, // 11 + 3 (DEX)
		},
		{
			name: "medium armor with dex cap",
			character: &models.Character{
				AbilityScores: models.AbilityScores{
					Dexterity: 18,
				},
			},
			armor: &models.Armor{
				Name:     "Breastplate",
				Type:     "medium",
				BaseAC:   14,
				MaxDexBonus: 2,
			},
			shield:   false,
			expected: 16, // 14 + 2 (capped DEX bonus)
		},
		{
			name: "heavy armor ignores dex",
			character: &models.Character{
				AbilityScores: models.AbilityScores{
					Dexterity: 20,
				},
			},
			armor: &models.Armor{
				Name:     "Plate Armor",
				Type:     "heavy",
				BaseAC:   18,
				MaxDexBonus: 0,
			},
			shield:   false,
			expected: 18, // Just base AC
		},
		{
			name: "armor with shield",
			character: &models.Character{
				AbilityScores: models.AbilityScores{
					Dexterity: 14,
				},
			},
			armor: &models.Armor{
				Name:     "Chain Mail",
				Type:     "heavy",
				BaseAC:   16,
				MaxDexBonus: 0,
			},
			shield:   true,
			expected: 18, // 16 + 2 (shield)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &CombatService{}
			result := service.CalculateAC(tt.character, tt.armor, tt.shield)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// CombatService implementation stub for testing
type CombatService struct {
	diceRoller DiceRoller
}

func (s *CombatService) RollInitiative(combatants []models.Combatant) []models.Combatant {
	// Implementation would go here
	return combatants
}

func (s *CombatService) Attack(attacker, target *models.Combatant, weapon *models.Weapon) (bool, int, error) {
	if attacker == nil {
		return false, 0, errors.New("attacker cannot be nil")
	}
	if target == nil {
		return false, 0, errors.New("target cannot be nil")
	}
	if weapon == nil {
		return false, 0, errors.New("weapon cannot be nil")
	}
	// Implementation would go here
	return false, 0, nil
}

func (s *CombatService) SavingThrow(character *models.Character, saveType string, dc int) (bool, int, error) {
	if character == nil {
		return false, 0, errors.New("character cannot be nil")
	}
	validSaves := map[string]bool{
		"strength": true, "dexterity": true, "constitution": true,
		"intelligence": true, "wisdom": true, "charisma": true,
	}
	if !validSaves[saveType] {
		return false, 0, errors.New("invalid save type")
	}
	// Implementation would go here
	return false, 0, nil
}

func (s *CombatService) SkillCheck(character *models.Character, skill string, dc int) (bool, int, error) {
	if character == nil {
		return false, 0, errors.New("character cannot be nil")
	}
	validSkills := map[string]bool{
		"acrobatics": true, "animal_handling": true, "arcana": true, "athletics": true,
		"deception": true, "history": true, "insight": true, "intimidation": true,
		"investigation": true, "medicine": true, "nature": true, "perception": true,
		"performance": true, "persuasion": true, "religion": true, "sleight_of_hand": true,
		"stealth": true, "survival": true,
	}
	if !validSkills[skill] {
		return false, 0, errors.New("invalid skill")
	}
	// Implementation would go here
	return false, 0, nil
}

func (s *CombatService) ApplyDamage(target *models.Combatant, damage int, damageType string) error {
	if target == nil {
		return errors.New("target cannot be nil")
	}
	// Implementation would go here
	return nil
}

func (s *CombatService) CalculateAC(character *models.Character, armor *models.Armor, shield bool) int {
	// Implementation would go here
	return 10
}

// DiceRoller interface for dice rolling
type DiceRoller interface {
	Roll(dice string) (int, error)
	RollWithDetails(dice string) (*models.RollDetails, error)
}