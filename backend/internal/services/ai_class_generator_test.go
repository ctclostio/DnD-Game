package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

func TestNewAIClassGenerator(t *testing.T) {
	mockLLM := &MockLLMProvider{}
	generator := NewAIClassGenerator(mockLLM)

	require.NotNil(t, generator)
	require.Equal(t, mockLLM, generator.llmProvider)
}

func TestAIClassGenerator_GenerateCustomClass(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		// Prepare valid AI response
		aiClass := map[string]interface{}{
			"name":            "Shadowdancer",
			"description":     "Masters of stealth and shadow magic",
			"hitDice":         "1d8",
			"primaryAbility":  "dexterity",
			"saves":           []string{"dexterity", "charisma"},
			"skillOptions":    []string{"Acrobatics", "Deception", "Investigation", "Performance", "Perception", "Sleight of Hand", "Stealth"},
			"skillCount":      3,
			"equipmentProficiencies": map[string][]string{
				"armor":   []string{"Light armor"},
				"weapons": []string{"Simple weapons", "shortswords", "rapiers", "hand crossbows"},
				"tools":   []string{"Thieves' tools"},
			},
			"features": []map[string]interface{}{
				{
					"level":       1,
					"name":        "Shadow Step",
					"description": "You can teleport between shadows",
				},
				{
					"level":       1,
					"name":        "Darkvision",
					"description": "You can see in darkness within 60 feet",
				},
				{
					"level":       3,
					"name":        "Shadow Clone",
					"description": "Create an illusory duplicate of yourself",
				},
			},
			"spellcasting": map[string]interface{}{
				"ability":      "charisma",
				"cantrips":     2,
				"spellsKnown":  4,
				"spellList":    []string{"minor illusion", "silent image", "darkness", "shadow blade"},
				"ritualCaster": false,
			},
			"subclasses": []map[string]interface{}{
				{
					"name":        "Path of Shadows",
					"description": "Focus on stealth and assassination",
					"features": []map[string]interface{}{
						{
							"level":       3,
							"name":        "Assassinate",
							"description": "Critical hits on surprised creatures",
						},
					},
				},
			},
			"multiclassingRequirements": map[string]interface{}{
				"minimumScores": map[string]int{
					"dexterity": 13,
				},
				"proficienciesGained": map[string][]string{
					"armor":   []string{"Light armor"},
					"weapons": []string{},
					"tools":   []string{},
				},
			},
			"balanceScore": 7,
			"powerLevel":   "moderate",
		}

		aiResponse, _ := json.Marshal(aiClass)
		mockLLM := &MockLLMProvider{
			Response: string(aiResponse),
		}

		generator := NewAIClassGenerator(mockLLM)

		req := CustomClassRequest{
			Name:        "Shadowdancer",
			Description: "A class that manipulates shadows",
			Role:        "stealth damage dealer",
			Style:       "balanced",
			Features:    "shadow magic and stealth",
		}

		ctx := testutil.TestContext()
		class, err := generator.GenerateCustomClass(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, class)
		require.Equal(t, "Shadowdancer", class.Name)
		require.Equal(t, "Masters of stealth and shadow magic", class.Description)
		require.Equal(t, 8, class.HitDie)
		require.Equal(t, "dexterity", class.PrimaryAbility)
		require.Len(t, class.ClassFeatures, 3)
		require.Equal(t, 7, class.BalanceScore)
	})

	t.Run("LLM provider error", func(t *testing.T) {
		mockLLM := &MockLLMProvider{
			Error: errors.New("API error"),
		}

		generator := NewAIClassGenerator(mockLLM)

		req := CustomClassRequest{
			Name:        "Test Class",
			Description: "Test description",
			Role:        "tank",
		}

		ctx := testutil.TestContext()
		class, err := generator.GenerateCustomClass(ctx, req)

		require.Error(t, err)
		require.Nil(t, class)
		require.Contains(t, err.Error(), "failed to generate class")
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		mockLLM := &MockLLMProvider{
			Response: "invalid json",
		}

		generator := NewAIClassGenerator(mockLLM)

		req := CustomClassRequest{
			Name:        "Test Class",
			Description: "Test description",
			Role:        "healer",
		}

		ctx := testutil.TestContext()
		class, err := generator.GenerateCustomClass(ctx, req)

		require.Error(t, err)
		require.Nil(t, class)
		require.Contains(t, err.Error(), "failed to parse class response")
	})

	t.Run("validation failure - overpowered", func(t *testing.T) {
		// Create an overpowered class
		aiClass := map[string]interface{}{
			"name":           "Godslayer",
			"description":    "Too powerful",
			"hitDice":        "1d12",
			"primaryAbility": "strength",
			"saves":          []string{"strength", "dexterity", "constitution"}, // Too many saves
			"features": []map[string]interface{}{
				{"level": 1, "name": "Divine Strike", "description": "Deal extra damage"},
				{"level": 1, "name": "Immortality", "description": "Cannot die"},
				{"level": 1, "name": "All-Seeing", "description": "See everything"},
				{"level": 1, "name": "Time Stop", "description": "Stop time"},
				{"level": 1, "name": "Reality Warp", "description": "Change reality"},
			},
			"balanceScore": 15, // Way too high
		}

		aiResponse, _ := json.Marshal(aiClass)
		mockLLM := &MockLLMProvider{
			Response: string(aiResponse),
		}

		generator := NewAIClassGenerator(mockLLM)

		req := CustomClassRequest{
			Name:        "Godslayer",
			Description: "An overpowered class",
			Role:        "damage dealer",
			Style:       "powerful",
		}

		ctx := testutil.TestContext()
		class, err := generator.GenerateCustomClass(ctx, req)

		require.Error(t, err)
		require.Nil(t, class)
		require.Contains(t, err.Error(), "class validation failed")
	})
}

func TestAIClassGenerator_buildClassPrompt(t *testing.T) {
	generator := NewAIClassGenerator(&MockLLMProvider{})

	tests := []struct {
		name     string
		req      CustomClassRequest
		contains []string
	}{
		{
			name: "complete request",
			req: CustomClassRequest{
				Name:        "Spellblade",
				Description: "A warrior who weaves magic into combat",
				Role:        "hybrid damage dealer",
				Style:       "balanced",
				Features:    "elemental weapon enchantments and defensive magic",
			},
			contains: []string{
				"Spellblade",
				"warrior who weaves magic",
				"hybrid damage dealer",
				"balanced",
				"elemental weapon enchantments",
				"JSON format",
			},
		},
		{
			name: "minimal request",
			req: CustomClassRequest{
				Name:        "Basic Class",
				Description: "A simple test class",
				Role:        "support",
			},
			contains: []string{
				"Basic Class",
				"simple test class",
				"support",
				"JSON format",
			},
		},
		{
			name: "tank role",
			req: CustomClassRequest{
				Name:        "Defender",
				Description: "Protects allies",
				Role:        "tank",
				Style:       "flavorful",
			},
			contains: []string{
				"tank",
				"high hit points",
				"defensive abilities",
				"flavorful",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := generator.buildClassPrompt(tt.req)

			for _, expected := range tt.contains {
				require.Contains(t, prompt, expected)
			}
		})
	}
}

func TestAIClassGenerator_parseClassResponse(t *testing.T) {
	generator := NewAIClassGenerator(&MockLLMProvider{})

	t.Run("valid response", func(t *testing.T) {
		response := `{
			"name": "Test Class",
			"description": "A test class",
			"hitDice": "1d10",
			"primaryAbility": "strength",
			"saves": ["strength", "constitution"],
			"features": [
				{
					"level": 1,
					"name": "Test Feature",
					"description": "Does something"
				}
			],
			"balanceScore": 5
		}`

		class, err := generator.parseClassResponse(response)

		require.NoError(t, err)
		require.NotNil(t, class)
		require.Equal(t, "Test Class", class.Name)
		require.Equal(t, 10, class.HitDie)
		require.Len(t, class.ClassFeatures, 1)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		response := "not json"

		class, err := generator.parseClassResponse(response)

		require.Error(t, err)
		require.Nil(t, class)
	})

	t.Run("missing required fields", func(t *testing.T) {
		response := `{
			"name": "Incomplete Class"
		}`

		class, err := generator.parseClassResponse(response)

		require.Error(t, err)
		require.Nil(t, class)
		require.Contains(t, err.Error(), "missing required fields")
	})
}

func TestAIClassGenerator_generateSpellSlotProgression(t *testing.T) {
	generator := NewAIClassGenerator(&MockLLMProvider{})

	tests := []struct {
		name              string
		class             *models.CustomClass
		expectedCantrips  int
		expectedLevel1    int
		expectedMaxLevel  int
	}{
		{
			name: "full caster",
			class: &models.CustomClass{
				SpellcastingAbility: "intelligence",
			},
			expectedCantrips: 3,
			expectedLevel1:   2,
			expectedMaxLevel: 9,
		},
		{
			name: "half caster",
			class: &models.CustomClass{
				SpellcastingAbility: "wisdom",
				SpellsKnownProgression: []int{0, 0, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11}, // Indicates half-caster
			},
			expectedCantrips: 0,
			expectedLevel1:   2,
			expectedMaxLevel: 5,
		},
		{
			name: "non-caster",
			class: &models.CustomClass{
				Name: "Fighter",
			},
			expectedCantrips: 0,
			expectedLevel1:   0,
			expectedMaxLevel: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progression := generator.generateSpellSlotProgression(tt.class)

			if tt.expectedCantrips > 0 {
				level1 := progression["1"].(map[string]int)
				require.Equal(t, tt.expectedCantrips, level1["cantrips"])
			}

			if tt.expectedLevel1 > 0 {
				level1 := progression["1"].(map[string]int)
				require.Equal(t, tt.expectedLevel1, level1["1st"])
			}

			// Check max spell level
			if tt.expectedMaxLevel > 0 {
				level20 := progression["20"].(map[string]int)
				maxLevelKey := string(rune(tt.expectedMaxLevel)) + "th"
				if tt.expectedMaxLevel == 1 {
					maxLevelKey = "1st"
				} else if tt.expectedMaxLevel == 2 {
					maxLevelKey = "2nd"
				} else if tt.expectedMaxLevel == 3 {
					maxLevelKey = "3rd"
				}
				require.Contains(t, level20, maxLevelKey)
			}
		})
	}
}

func TestAIClassGenerator_validateClass(t *testing.T) {
	generator := NewAIClassGenerator(&MockLLMProvider{})

	tests := []struct {
		name    string
		class   *models.CustomClass
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid balanced class",
			class: &models.CustomClass{
				Name:           "Valid Class",
				Description:    "A balanced class",
				HitDie:         8,
				PrimaryAbility: "wisdom",
				BalanceScore:   6,
				ClassFeatures: []models.ClassFeature{
					{Level: 1, Name: "Feature 1"},
					{Level: 3, Name: "Feature 2"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			class: &models.CustomClass{
				Description: "No name",
				HitDie:      8,
			},
			wantErr: true,
			errMsg:  "class name cannot be empty",
		},
		{
			name: "invalid hit dice",
			class: &models.CustomClass{
				Name:    "Bad Dice",
				HitDie: 13,
			},
			wantErr: true,
			errMsg:  "invalid hit dice",
		},
		{
			name: "too many features at level 1",
			class: &models.CustomClass{
				Name:    "Feature Heavy",
				HitDie: 8,
				ClassFeatures: []models.ClassFeature{
					{Level: 1, Name: "Feature 1"},
					{Level: 1, Name: "Feature 2"},
					{Level: 1, Name: "Feature 3"},
					{Level: 1, Name: "Feature 4"},
					{Level: 1, Name: "Feature 5"},
				},
			},
			wantErr: true,
			errMsg:  "too many features at level 1",
		},
		{
			name: "overpowered class",
			class: &models.CustomClass{
				Name:         "Overpowered",
				HitDie:       12,
				BalanceScore: 12,
			},
			wantErr: true,
			errMsg:  "class is too powerful",
		},
		{
			name: "invalid primary ability",
			class: &models.CustomClass{
				Name:           "Bad Ability",
				HitDie:         8,
				PrimaryAbility: "luck",
			},
			wantErr: true,
			errMsg:  "invalid primary ability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generator.validateClass(tt.class)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAIClassGenerator_calculateBalanceScore(t *testing.T) {
	generator := NewAIClassGenerator(&MockLLMProvider{})

	tests := []struct {
		name          string
		class         *models.CustomClass
		expectedScore int
	}{
		{
			name: "basic fighter-like class",
			class: &models.CustomClass{
				HitDie:         10,
				PrimaryAbility: "strength",
				ClassFeatures: []models.ClassFeature{
					{Level: 1, Name: "Fighting Style"},
					{Level: 2, Name: "Second Wind"},
				},
			},
			expectedScore: 5,
		},
		{
			name: "caster class",
			class: &models.CustomClass{
				HitDie:         6,
				PrimaryAbility: "intelligence",
				SpellcastingAbility: "intelligence",
				ClassFeatures: []models.ClassFeature{
					{Level: 1, Name: "Spellcasting"},
				},
			},
			expectedScore: 6,
		},
		{
			name: "hybrid class with many features",
			class: &models.CustomClass{
				HitDie:         8,
				PrimaryAbility: "charisma",
				SpellcastingAbility: "charisma",
				ClassFeatures: []models.ClassFeature{
					{Level: 1, Name: "Feature 1"},
					{Level: 1, Name: "Feature 2"},
					{Level: 3, Name: "Feature 3"},
					{Level: 5, Name: "Feature 4"},
					{Level: 7, Name: "Feature 5"},
				},
			},
			expectedScore: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := generator.calculateBalanceScore(tt.class)
			require.Equal(t, tt.expectedScore, score)
		})
	}
}

func TestAIClassGenerator_ConcurrentGeneration(t *testing.T) {
	// Prepare a valid response
	aiClass := map[string]interface{}{
		"name":           "Concurrent Class",
		"description":    "Test class",
		"hitDice":        "1d8",
		"primaryAbility": "wisdom",
		"saves":          []string{"wisdom", "charisma"},
		"features": []map[string]interface{}{
			{
				"level":       1,
				"name":        "Test Feature",
				"description": "A test feature",
			},
		},
		"balanceScore": 5,
	}

	aiResponse, _ := json.Marshal(aiClass)
	mockLLM := &MockLLMProvider{
		Response: string(aiResponse),
	}

	generator := NewAIClassGenerator(mockLLM)

	// Run multiple generations concurrently
	const numGoroutines = 10
	errors := make(chan error, numGoroutines)
	classes := make(chan *models.CustomClass, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			req := CustomClassRequest{
				Name:        "Concurrent" + string(rune(id)),
				Description: "Test class",
				Role:        "support",
			}

			ctx := context.Background()
			class, err := generator.GenerateCustomClass(ctx, req)
			if err != nil {
				errors <- err
			} else {
				classes <- class
			}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-errors:
			require.NoError(t, err)
		case class := <-classes:
			require.NotNil(t, class)
			require.Equal(t, "Concurrent Class", class.Name)
			successCount++
		}
	}

	require.Equal(t, numGoroutines, successCount)
}

// Integration test
func TestAIClassGenerator_Integration(t *testing.T) {
	t.Run("complete class generation flow", func(t *testing.T) {
		// Create a complex but balanced class
		aiClass := map[string]interface{}{
			"name":            "Mystic Knight",
			"description":     "Warriors who blend martial prowess with arcane power",
			"hitDice":         "1d10",
			"primaryAbility":  "strength",
			"saves":           []string{"constitution", "intelligence"},
			"skillOptions":    []string{"Arcana", "Athletics", "History", "Insight", "Investigation", "Persuasion"},
			"skillCount":      2,
			"equipmentProficiencies": map[string][]string{
				"armor":   []string{"All armor", "shields"},
				"weapons": []string{"Simple weapons", "martial weapons"},
				"tools":   []string{},
			},
			"features": []map[string]interface{}{
				{
					"level":       1,
					"name":        "Arcane Weapon",
					"description": "Infuse weapon with magical energy",
				},
				{
					"level":       2,
					"name":        "Fighting Style",
					"description": "Choose a combat specialization",
				},
				{
					"level":       3,
					"name":        "Mystic Strike",
					"description": "Channel spells through weapon attacks",
				},
			},
			"spellcasting": map[string]interface{}{
				"ability":      "intelligence",
				"cantrips":     0,
				"spellsKnown":  3,
				"spellList":    []string{"shield", "magic missile", "misty step"},
				"ritualCaster": true,
			},
			"subclasses": []map[string]interface{}{
				{
					"name":        "Order of the Spell Blade",
					"description": "Focus on weaving spells into combat",
					"features": []map[string]interface{}{
						{
							"level":       3,
							"name":        "Spell Parry",
							"description": "Use reactions to deflect spells",
						},
					},
				},
			},
			"balanceScore": 7,
			"powerLevel":   "balanced",
		}

		aiResponse, _ := json.Marshal(aiClass)
		mockLLM := &MockLLMProvider{
			Response: string(aiResponse),
		}

		generator := NewAIClassGenerator(mockLLM)

		req := CustomClassRequest{
			Name:        "Mystic Knight",
			Description: "A warrior-mage hybrid class",
			Role:        "hybrid damage dealer",
			Style:       "balanced",
			Features:    "spell channeling through weapons",
		}

		ctx := testutil.TestContext()
		class, err := generator.GenerateCustomClass(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, class)

		// Validate the complete class
		err = generator.validateClass(class)
		require.NoError(t, err)

		// Check balance
		require.Equal(t, 7, class.BalanceScore)
		require.Equal(t, "balanced", class.PowerLevel)

		// Verify spell progression was generated
		require.NotNil(t, class.SpellSlotsProgression)

		// Verify it has appropriate features for a hybrid class
		require.NotEmpty(t, class.SpellcastingAbility)
		require.Equal(t, 10, class.HitDie) // Good HP for a hybrid
	})
}

// Benchmark tests
func BenchmarkAIClassGenerator_GenerateCustomClass(b *testing.B) {
	aiResponse := `{
		"name": "Benchmark Class",
		"description": "Test",
		"hitDice": "1d8",
		"primaryAbility": "wisdom",
		"saves": ["wisdom", "dexterity"],
		"features": [
			{"level": 1, "name": "Feature", "description": "Test"}
		],
		"balanceScore": 5
	}`

	mockLLM := &MockLLMProvider{Response: aiResponse}
	generator := NewAIClassGenerator(mockLLM)

	req := CustomClassRequest{
		Name:        "Benchmark",
		Description: "Test class",
		Role:        "support",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.GenerateCustomClass(ctx, req)
	}
}

func BenchmarkAIClassGenerator_validateClass(b *testing.B) {
	generator := NewAIClassGenerator(&MockLLMProvider{})
	class := &models.CustomClass{
		Name:           "Test Class",
		Description:    "A test class",
		HitDie:         8,
		PrimaryAbility: "wisdom",
		BalanceScore:   6,
		ClassFeatures: []models.ClassFeature{
			{Level: 1, Name: "Feature 1"},
			{Level: 3, Name: "Feature 2"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.validateClass(class)
	}
}