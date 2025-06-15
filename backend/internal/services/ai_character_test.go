package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

func TestNewAICharacterService(t *testing.T) {
	tests := []struct {
		name        string
		llmProvider LLMProvider
		wantEnabled bool
	}{
		{
			name:        "With LLM Provider",
			llmProvider: &MockLLMProvider{},
			wantEnabled: true,
		},
		{
			name:        "Without LLM Provider",
			llmProvider: nil,
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAICharacterService(tt.llmProvider)
			assert.NotNil(t, service)
			assert.Equal(t, tt.wantEnabled, service.IsEnabled())
		})
	}
}

func TestAICharacterService_GenerateCustomCharacter(t *testing.T) {
	validAIResponse := `{
		"race": "Elf",
		"subrace": "High Elf",
		"class": "Wizard",
		"subclass": "School of Evocation",
		"background": "Sage",
		"alignment": "Lawful Good",
		"attributes": {
			"strength": 8,
			"dexterity": 16,
			"constitution": 14,
			"intelligence": 18,
			"wisdom": 12,
			"charisma": 10
		},
		"hitDice": "1d6",
		"speed": 30,
		"features": [
			{
				"name": "Darkvision",
				"description": "You can see in dim light within 60 feet.",
				"level": 1,
				"source": "Race"
			}
		],
		"proficiencies": {
			"armor": ["None"],
			"weapons": ["Daggers", "Darts", "Slings", "Quarterstaffs", "Light crossbows"],
			"tools": ["None"],
			"languages": ["Common", "Elvish", "Draconic"]
		},
		"skills": ["Arcana", "History", "Insight", "Investigation"],
		"equipment": [
			{
				"name": "Spellbook",
				"type": "gear",
				"description": "A leather-bound book containing your spells"
			}
		],
		"personality": {
			"traits": ["I use polysyllabic words that convey the impression of great erudition."],
			"ideals": ["Knowledge. The path to power and self-improvement is through knowledge."],
			"bonds": ["I have an ancient text that holds terrible secrets that must not fall into the wrong hands."],
			"flaws": ["I speak without really thinking through my words, invariably insulting others."]
		},
		"backstory": "Born into a noble elven family, I spent my youth studying ancient texts and magical theory."
	}`

	tests := []struct {
		name              string
		request           CustomCharacterRequest
		mockResponse      string
		mockError         error
		expectError       bool
		validateCharacter func(t *testing.T, char *models.Character)
	}{
		{
			name: "Successful Character Generation",
			request: CustomCharacterRequest{
				Name:       "Aelindra Starweaver",
				Concept:    "A scholarly elven wizard obsessed with ancient magic",
				Race:       "Elf",
				Class:      "Wizard",
				Background: "Sage",
				Level:      5,
			},
			mockResponse: validAIResponse,
			expectError:  false,
			validateCharacter: func(t *testing.T, char *models.Character) {
				assert.Equal(t, "Aelindra Starweaver", char.Name)
				assert.Equal(t, "Elf", char.Race)
				assert.Equal(t, "High Elf", char.Subrace)
				assert.Equal(t, "Wizard", char.Class)
				assert.Equal(t, 5, char.Level)
				assert.Equal(t, 18, char.Attributes.Intelligence)
				assert.Equal(t, 3, char.ProficiencyBonus) // Level 5 = +3
				assert.Equal(t, 3, char.Initiative)       // DEX 16 = +3
			},
		},
		{
			name: "AI Service Error",
			request: CustomCharacterRequest{
				Name:    "Failed Character",
				Concept: "This will fail",
			},
			mockError:   errors.New("AI service unavailable"),
			expectError: true,
		},
		{
			name: "Invalid JSON Response",
			request: CustomCharacterRequest{
				Name:    "Invalid Response",
				Concept: "Bad JSON",
			},
			mockResponse: "This is not valid JSON",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := &MockLLMProvider{
				Response: tt.mockResponse,
				Error:    tt.mockError,
			}
			service := NewAICharacterService(mockLLM)

			character, err := service.GenerateCustomCharacter(&tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, character)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, character)
				if tt.validateCharacter != nil {
					tt.validateCharacter(t, character)
				}
			}
		})
	}
}

func TestAICharacterService_GenerateCustomCharacter_Disabled(t *testing.T) {
	service := NewAICharacterService(nil)

	req := CustomCharacterRequest{
		Name:    "Test",
		Concept: "Test concept",
	}
	_, err := service.GenerateCustomCharacter(&req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AI character generation is not enabled")
}

func TestAICharacterService_ValidateCustomContent(t *testing.T) {
	balancedResponse := `{
		"balanced": true,
		"issues": [],
		"suggestions": []
	}`

	unbalancedResponse := `{
		"balanced": false,
		"issues": ["Stats are too high", "Too many proficiencies"],
		"suggestions": ["Reduce ability scores", "Limit proficiencies to 2"]
	}`

	tests := []struct {
		name         string
		character    *models.Character
		mockResponse string
		mockError    error
		expectError  bool
		errorMessage string
	}{
		{
			name: "Balanced Character",
			character: &models.Character{
				Name:  "Balanced Hero",
				Race:  "Human",
				Class: "Fighter",
				Level: 1,
				Attributes: models.Attributes{
					Strength:     16,
					Dexterity:    14,
					Constitution: 15,
					Intelligence: 10,
					Wisdom:       12,
					Charisma:     8,
				},
			},
			mockResponse: balancedResponse,
			expectError:  false,
		},
		{
			name: "Unbalanced Character",
			character: &models.Character{
				Name:  "Overpowered Hero",
				Race:  "Custom Dragon",
				Class: "God Slayer",
				Level: 1,
				Attributes: models.Attributes{
					Strength:     20,
					Dexterity:    20,
					Constitution: 20,
					Intelligence: 20,
					Wisdom:       20,
					Charisma:     20,
				},
			},
			mockResponse: unbalancedResponse,
			expectError:  true,
			errorMessage: "character balance issues",
		},
		{
			name: "AI Service Error",
			character: &models.Character{
				Name: "Error Character",
			},
			mockError:   errors.New("AI service unavailable"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := &MockLLMProvider{
				Response: tt.mockResponse,
				Error:    tt.mockError,
			}
			service := NewAICharacterService(mockLLM)

			err := service.ValidateCustomContent(tt.character)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAICharacterService_ValidateCustomContent_Disabled(t *testing.T) {
	service := NewAICharacterService(nil)

	// Should skip validation when AI is disabled
	err := service.ValidateCustomContent(&models.Character{Name: "Test"})
	assert.NoError(t, err)
}

func TestAICharacterService_GenerateFallbackCharacter(t *testing.T) {
	tests := []struct {
		name              string
		request           CustomCharacterRequest
		validateCharacter func(t *testing.T, char *models.Character)
	}{
		{
			name: "Basic Fallback Character",
			request: CustomCharacterRequest{
				Name:    "Fallback Hero",
				Concept: "A simple adventurer",
			},
			validateCharacter: func(t *testing.T, char *models.Character) {
				assert.Equal(t, "Fallback Hero", char.Name)
				assert.Equal(t, "Custom", char.Race)
				assert.Equal(t, "Adventurer", char.Class)
				assert.Equal(t, 1, char.Level)
				assert.Equal(t, 30, char.Speed)
			},
		},
		{
			name: "Fallback with Custom Values",
			request: CustomCharacterRequest{
				Name:       "Custom Fallback",
				Concept:    "A unique hero",
				Race:       "Dwarf",
				Class:      "Cleric",
				Background: "Acolyte",
				Level:      3,
			},
			validateCharacter: func(t *testing.T, char *models.Character) {
				assert.Equal(t, "Custom Fallback", char.Name)
				assert.Equal(t, "Dwarf", char.Race)
				assert.Equal(t, "Cleric", char.Class)
				assert.Equal(t, "Acolyte", char.Background)
				assert.Equal(t, 3, char.Level)
				assert.Equal(t, 2, char.ProficiencyBonus) // Level 3 = +2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAICharacterService(nil) // No LLM provider

			character, err := service.GenerateFallbackCharacter(&tt.request)

			assert.NoError(t, err)
			assert.NotNil(t, character)
			tt.validateCharacter(t, character)

			// Verify resources contain concept
			assert.Equal(t, tt.request.Concept, character.Resources["concept"])
			assert.Equal(t, true, character.Resources["custom"])
		})
	}
}

func TestAICharacterService_CalculateModifier(t *testing.T) {
	service := NewAICharacterService(nil)

	tests := []struct {
		score    int
		expected int
	}{
		{1, -4}, // (1-10)/2 = -9/2 = -4 (integer division)
		{3, -3}, // (3-10)/2 = -7/2 = -3
		{8, -1},
		{10, 0},
		{11, 0},
		{12, 1},
		{13, 1},
		{14, 2},
		{15, 2},
		{16, 3},
		{18, 4},
		{20, 5},
		{30, 10},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Score_%d", tt.score), func(t *testing.T) {
			result := service.calculateModifier(tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAICharacterService_GetHitDiceValue(t *testing.T) {
	service := NewAICharacterService(nil)

	tests := []struct {
		hitDice  string
		expected int
	}{
		{"1d6", 6},
		{"1d8", 8},
		{"1d10", 10},
		{"1d12", 12},
		{"2d6", 8},     // Invalid format, defaults to 8
		{"", 8},        // Empty, defaults to 8
		{"invalid", 8}, // Invalid, defaults to 8
	}

	for _, tt := range tests {
		t.Run(tt.hitDice, func(t *testing.T) {
			result := service.getHitDiceValue(tt.hitDice)
			assert.Equal(t, tt.expected, result)
		})
	}
}
