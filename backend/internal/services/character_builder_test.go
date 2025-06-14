package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacterBuilder_BuildCharacter(t *testing.T) {
	// Create a temp directory for test data.
	tmpDir := t.TempDir()
	setupTestData(t, tmpDir)

	builder := NewCharacterBuilder(tmpDir)

	t.Run("successful character creation with standard race and class", func(t *testing.T) {
		// Test data.
		params := map[string]interface{}{
			"name":       "Aragorn",
			"race":       "human",
			"class":      "fighter",
			"background": "soldier",
			"alignment":  "Lawful Good",
			"abilityScores": map[string]int{
				"strength":     16,
				"dexterity":    14,
				"constitution": 15,
				"intelligence": 10,
				"wisdom":       13,
				"charisma":     12,
			},
		}

		// Execute.
		char, err := builder.BuildCharacter(params)

		// Assert.
		require.NoError(t, err)
		require.NotNil(t, char)
		assert.Equal(t, "Aragorn", char.Name)
		assert.Equal(t, "human", char.Race)
		assert.Equal(t, "fighter", char.Class)
		assert.Equal(t, 1, char.Level)

		// Verify racial bonuses applied (Human gets +1 to all).
		assert.Equal(t, 17, char.Attributes.Strength)
		assert.Equal(t, 15, char.Attributes.Dexterity)
		assert.Equal(t, 16, char.Attributes.Constitution)

		// Verify HP calculation (10 base + 3 CON modifier).
		assert.Equal(t, 13, char.MaxHitPoints)
		assert.Equal(t, 13, char.HitPoints)
	})

	t.Run("character creation with spell-casting class", func(t *testing.T) {
		params := map[string]interface{}{
			"name":       "Gandalf",
			"race":       "elf",
			"class":      "wizard",
			"background": "sage",
			"alignment":  "Neutral Good",
			"abilityScores": map[string]int{
				"strength":     8,
				"dexterity":    14,
				"constitution": 14,
				"intelligence": 16,
				"wisdom":       13,
				"charisma":     10,
			},
		}

		char, err := builder.BuildCharacter(params)

		require.NoError(t, err)
		require.NotNil(t, char)

		// Verify wizard-specific features.
		assert.NotEmpty(t, char.Spells.SpellSlots)
		assert.Equal(t, 2, char.Spells.SpellSlots[0].Total) // Level 1 spell slots
		assert.Equal(t, "Intelligence", char.Spells.SpellcastingAbility)

		// Verify spell save DC and attack bonus.
		intMod := builder.calculateModifier(char.Attributes.Intelligence)
		expectedDC := 8 + char.ProficiencyBonus + intMod
		assert.Equal(t, expectedDC, char.Spells.SpellSaveDC)
	})

	t.Run("character with custom race", func(t *testing.T) {
		customRaceID := "custom-race-123"
		params := map[string]interface{}{
			"name":         "Draconis",
			"race":         "custom",
			"customRaceId": customRaceID,
			"customRaceStats": map[string]interface{}{
				"name": "Dragonkin",
				"abilityScoreIncreases": map[string]interface{}{
					"strength": 2.0,
					"charisma": 1.0,
				},
				"size":       "Medium",
				"speed":      30.0,
				"languages":  []interface{}{"Common", "Draconic"},
				"darkvision": 60.0,
				"traits": []interface{}{
					map[string]interface{}{
						"name":        "Breath Weapon",
						"description": "You can use your action to exhale destructive energy.",
					},
				},
			},
			"class":      "paladin",
			"background": "noble",
			"alignment":  "Lawful Good",
			"abilityScores": map[string]int{
				"strength":     14,
				"dexterity":    10,
				"constitution": 13,
				"intelligence": 10,
				"wisdom":       12,
				"charisma":     14,
			},
		}

		char, err := builder.BuildCharacter(params)

		require.NoError(t, err)
		require.NotNil(t, char)
		assert.Equal(t, "Dragonkin", char.Race)
		assert.Equal(t, &customRaceID, char.CustomRaceID)

		// Verify custom racial bonuses applied.
		assert.Equal(t, 16, char.Attributes.Strength) // 14 + 2
		assert.Equal(t, 15, char.Attributes.Charisma) // 14 + 1
	})
}

func TestCharacterBuilder_GetAvailableOptions(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestData(t, tmpDir)

	builder := NewCharacterBuilder(tmpDir)

	options, err := builder.GetAvailableOptions()
	require.NoError(t, err)
	require.NotNil(t, options)

	// Check that all required fields are present.
	races, ok := options["races"].([]string)
	require.True(t, ok)
	assert.Contains(t, races, "human")
	assert.Contains(t, races, "elf")
	assert.Contains(t, races, "dwarf")

	classes, ok := options["classes"].([]string)
	require.True(t, ok)
	assert.Contains(t, classes, "fighter")
	assert.Contains(t, classes, "wizard")
	assert.Contains(t, classes, "cleric")

	backgrounds, ok := options["backgrounds"].([]string)
	require.True(t, ok)
	assert.Contains(t, backgrounds, "soldier")
	assert.Contains(t, backgrounds, "sage")
	assert.Contains(t, backgrounds, "noble")

	methods, ok := options["abilityScoreMethods"].([]string)
	require.True(t, ok)
	assert.Contains(t, methods, "standard_array")
	assert.Contains(t, methods, "point_buy")
	assert.Contains(t, methods, "roll_4d6")
	assert.Contains(t, methods, "custom")
}

func TestCharacterBuilder_RollAbilityScores(t *testing.T) {
	builder := &CharacterBuilder{}

	t.Run("standard array", func(t *testing.T) {
		scores, err := builder.RollAbilityScores("standard_array")
		require.NoError(t, err)

		// Check that standard array values are present.
		expectedValues := []int{15, 14, 13, 12, 10, 8}
		var actualValues []int
		for _, ability := range []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"} {
			actualValues = append(actualValues, scores[ability])
		}

		// Sort to compare (order doesn't matter for standard array).
		assert.ElementsMatch(t, expectedValues, actualValues)
	})

	t.Run("roll 4d6", func(t *testing.T) {
		scores, err := builder.RollAbilityScores("roll_4d6")
		require.NoError(t, err)

		// Check that all abilities have valid scores (3-18).
		for ability, score := range scores {
			assert.GreaterOrEqual(t, score, 3, "Ability %s should be at least 3", ability)
			assert.LessOrEqual(t, score, 18, "Ability %s should be at most 18", ability)
		}
	})

	t.Run("point buy", func(t *testing.T) {
		scores, err := builder.RollAbilityScores("point_buy")
		require.NoError(t, err)

		// Check that all abilities start at 8.
		for _, score := range scores {
			assert.Equal(t, 8, score)
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		_, err := builder.RollAbilityScores("invalid_method")
		assert.Error(t, err)
	})
}

func TestInitializeSpellSlots(t *testing.T) {
	tests := []struct {
		name     string
		class    string
		level    int
		expected []models.SpellSlot
	}{
		{
			name:  "level 1 wizard",
			class: "Wizard",
			level: 1,
			expected: []models.SpellSlot{
				{Level: 1, Total: 2, Remaining: 2},
			},
		},
		{
			name:  "level 3 cleric",
			class: "Cleric",
			level: 3,
			expected: []models.SpellSlot{
				{Level: 1, Total: 4, Remaining: 4},
				{Level: 2, Total: 2, Remaining: 2},
			},
		},
		{
			name:  "level 5 sorcerer",
			class: "Sorcerer",
			level: 5,
			expected: []models.SpellSlot{
				{Level: 1, Total: 4, Remaining: 4},
				{Level: 2, Total: 3, Remaining: 3},
				{Level: 3, Total: 2, Remaining: 2},
			},
		},
		{
			name:     "level 1 fighter (no spells)",
			class:    "Fighter",
			level:    1,
			expected: []models.SpellSlot{},
		},
		{
			name:  "level 5 ranger (half caster)",
			class: "Ranger",
			level: 5,
			expected: []models.SpellSlot{
				{Level: 1, Total: 4, Remaining: 4},
				{Level: 2, Total: 2, Remaining: 2},
			},
		},
		{
			name:  "level 1 warlock (pact magic)",
			class: "Warlock",
			level: 1,
			expected: []models.SpellSlot{
				{Level: 1, Total: 1, Remaining: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InitializeSpellSlots(tt.class, tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCharacterBuilder_CalculateModifier(t *testing.T) {
	builder := &CharacterBuilder{}

	tests := []struct {
		score    int
		expected int
	}{
		{1, -4}, // (1-10)/2 = -9/2 = -4 (truncates towards zero)
		{2, -4},
		{3, -3}, // (3-10)/2 = -7/2 = -3
		{4, -3},
		{5, -2}, // (5-10)/2 = -5/2 = -2
		{6, -2},
		{7, -1}, // (7-10)/2 = -3/2 = -1
		{8, -1},
		{9, 0}, // (9-10)/2 = -1/2 = 0
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
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("score %d", tt.score), func(t *testing.T) {
			result := builder.calculateModifier(tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCharacterBuilder_CalculateProficiencyBonus(t *testing.T) {
	builder := &CharacterBuilder{}

	tests := []struct {
		level    int
		expected int
	}{
		{1, 2},
		{2, 2},
		{3, 2},
		{4, 2},
		{5, 3},
		{6, 3},
		{7, 3},
		{8, 3},
		{9, 4},
		{10, 4},
		{11, 4},
		{12, 4},
		{13, 5},
		{14, 5},
		{15, 5},
		{16, 5},
		{17, 6},
		{18, 6},
		{19, 6},
		{20, 6},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("level %d", tt.level), func(t *testing.T) {
			result := builder.calculateProficiencyBonus(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to setup test data.
func setupTestData(t *testing.T, dir string) {
	// Create directory structure.
	for _, subdir := range []string{"races", "classes", "backgrounds"} {
		err := os.MkdirAll(filepath.Join(dir, subdir), 0755)
		require.NoError(t, err)
	}

	// Create minimal test data files.
	raceData := map[string]interface{}{
		"human": map[string]interface{}{
			"name": "Human",
			"abilityScoreIncrease": map[string]int{
				"strength":     1,
				"dexterity":    1,
				"constitution": 1,
				"intelligence": 1,
				"wisdom":       1,
				"charisma":     1,
			},
			"size":      "Medium",
			"speed":     30,
			"languages": []string{"Common"},
			"traits":    []map[string]interface{}{},
		},
		"elf": map[string]interface{}{
			"name":                 "Elf",
			"abilityScoreIncrease": map[string]int{"dexterity": 2},
			"size":                 "Medium",
			"speed":                30,
			"languages":            []string{"Common", "Elvish"},
			"traits":               []map[string]interface{}{},
		},
		"dwarf": map[string]interface{}{
			"name":                 "Dwarf",
			"abilityScoreIncrease": map[string]int{"constitution": 2},
			"size":                 "Medium",
			"speed":                25,
			"languages":            []string{"Common", "Dwarvish"},
			"traits":               []map[string]interface{}{},
		},
	}

	for name, data := range raceData {
		file := filepath.Join(dir, "races", name+".json")
		content, _ := json.MarshalIndent(data, "", "  ")
		err := os.WriteFile(file, content, 0644)
		require.NoError(t, err)
	}

	classData := map[string]interface{}{
		"fighter": map[string]interface{}{
			"name":                     "Fighter",
			"hitDice":                  "1d10",
			"primaryAbility":           "Strength",
			"savingThrowProficiencies": []string{"Strength", "Constitution"},
			"skillChoices":             map[string]interface{}{},
			"features":                 map[string]interface{}{},
		},
		"wizard": map[string]interface{}{
			"name":                     "Wizard",
			"hitDice":                  "1d6",
			"primaryAbility":           "Intelligence",
			"savingThrowProficiencies": []string{"Intelligence", "Wisdom"},
			"skillChoices":             map[string]interface{}{},
			"features":                 map[string]interface{}{},
			"spellcasting": map[string]interface{}{
				"ability": "Intelligence",
				"cantripsKnown": []interface{}{
					map[string]interface{}{"level": 1.0, "known": 3.0},
				},
			},
		},
		"cleric": map[string]interface{}{
			"name":                     "Cleric",
			"hitDice":                  "1d8",
			"primaryAbility":           "Wisdom",
			"savingThrowProficiencies": []string{"Wisdom", "Charisma"},
			"skillChoices":             map[string]interface{}{},
			"features":                 map[string]interface{}{},
			"spellcasting": map[string]interface{}{
				"ability": "Wisdom",
			},
		},
		"paladin": map[string]interface{}{
			"name":                     "Paladin",
			"hitDice":                  "1d10",
			"primaryAbility":           "Strength",
			"savingThrowProficiencies": []string{"Wisdom", "Charisma"},
			"skillChoices":             map[string]interface{}{},
			"features":                 map[string]interface{}{},
		},
	}

	for name, data := range classData {
		file := filepath.Join(dir, "classes", name+".json")
		content, _ := json.MarshalIndent(data, "", "  ")
		err := os.WriteFile(file, content, 0644)
		require.NoError(t, err)
	}

	backgroundData := map[string]interface{}{
		"soldier": map[string]interface{}{
			"name":               "Soldier",
			"skillProficiencies": []string{"Athletics", "Intimidation"},
			"languages":          1,
			"toolProficiencies":  []string{"Gaming Set"},
			"equipment":          []string{"Insignia of Rank", "Trophy"},
			"feature":            map[string]interface{}{},
		},
		"sage": map[string]interface{}{
			"name":               "Sage",
			"skillProficiencies": []string{"Arcana", "History"},
			"languages":          2,
			"toolProficiencies":  []string{},
			"equipment":          []string{"Letter from Colleague", "Quill"},
			"feature":            map[string]interface{}{},
		},
		"noble": map[string]interface{}{
			"name":               "Noble",
			"skillProficiencies": []string{"History", "Persuasion"},
			"languages":          1,
			"toolProficiencies":  []string{"Gaming Set"},
			"equipment":          []string{"Signet Ring", "Scroll of Pedigree"},
			"feature":            map[string]interface{}{},
		},
	}

	for name, data := range backgroundData {
		file := filepath.Join(dir, "backgrounds", name+".json")
		content, _ := json.MarshalIndent(data, "", "  ")
		err := os.WriteFile(file, content, 0644)
		require.NoError(t, err)
	}
}
