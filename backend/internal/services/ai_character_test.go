package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// MockLLMProvider is a mock implementation of LLMProvider interface
type MockLLMProviderForTest struct {
	mock.Mock
}

func (m *MockLLMProviderForTest) GenerateCompletion(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	args := m.Called(ctx, prompt, systemPrompt)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProviderForTest) GenerateContent(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	args := m.Called(ctx, prompt, systemPrompt)
	return args.String(0), args.Error(1)
}

func TestNewAICharacterService(t *testing.T) {
	tests := []struct {
		name        string
		llmProvider LLMProvider
		wantEnabled bool
	}{
		{
			name:        "With LLM Provider",
			llmProvider: &MockLLMProviderForTest{},
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
			},
			{
				"name": "Keen Senses",
				"description": "You have proficiency in the Perception skill.",
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
			},
			{
				"name": "Component Pouch",
				"type": "gear",
				"description": "A small pouch containing spell components"
			}
		],
		"personality": {
			"traits": ["I use polysyllabic words that convey the impression of great erudition.", "I've read every book in the world's greatest libraries—or I like to boast that I have."],
			"ideals": ["Knowledge. The path to power and self-improvement is through knowledge."],
			"bonds": ["I have an ancient text that holds terrible secrets that must not fall into the wrong hands."],
			"flaws": ["I speak without really thinking through my words, invariably insulting others."]
		},
		"backstory": "Born into a noble elven family, I spent my youth studying ancient texts and magical theory."
	}`

	tests := []struct {
		name             string
		request          CustomCharacterRequest
		mockResponse     string
		mockError        error
		expectError      bool
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
				assert.Equal(t, "School of Evocation", char.Subclass)
				assert.Equal(t, 5, char.Level)
				assert.Equal(t, 18, char.Attributes.Intelligence)
				assert.Equal(t, 3, char.ProficiencyBonus) // Level 5 = +3
				assert.Equal(t, 3, char.Initiative) // DEX 16 = +3
				assert.Equal(t, 8, char.MaxHitPoints) // 6 + CON mod (2)
			},
		},
		{
			name: "Generation with Default Values",
			request: CustomCharacterRequest{
				Name:    "Simple Character",
				Concept: "A basic adventurer",
			},
			mockResponse: validAIResponse,
			expectError:  false,
			validateCharacter: func(t *testing.T, char *models.Character) {
				assert.Equal(t, 1, char.Level) // Default level
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
		{
			name: "Partial JSON Response",
			request: CustomCharacterRequest{
				Name:    "Partial Response",
				Concept: "Incomplete JSON",
			},
			mockResponse: `Some text before JSON {"race": "Human"} and some after`,
			expectError:  true, // Missing required fields
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderForTest)
			service := NewAICharacterService(mockLLM)

			if tt.mockError == nil && tt.mockResponse != "" {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, nil)
			} else if tt.mockError != nil {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.Anything, mock.Anything).
					Return("", tt.mockError)
			}

			character, err := service.GenerateCustomCharacter(tt.request)

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

			mockLLM.AssertExpectations(t)
		})
	}
}

func TestAICharacterService_GenerateCustomCharacter_Disabled(t *testing.T) {
	service := NewAICharacterService(nil)
	
	_, err := service.GenerateCustomCharacter(CustomCharacterRequest{
		Name:    "Test",
		Concept: "Test concept",
	})
	
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
		{
			name: "Invalid Response Format",
			character: &models.Character{
				Name: "Invalid Response",
			},
			mockResponse: "Not a JSON response",
			expectError:  false, // Falls back to assuming balanced
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderForTest)
			service := NewAICharacterService(mockLLM)

			if tt.mockError == nil && tt.mockResponse != "" {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, nil)
			} else if tt.mockError != nil {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.Anything, mock.Anything).
					Return("", tt.mockError)
			}

			err := service.ValidateCustomContent(tt.character)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			mockLLM.AssertExpectations(t)
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
			
			character, err := service.GenerateFallbackCharacter(tt.request)
			
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
		{1, -5},
		{3, -4},
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
		{"2d6", 8},  // Invalid format, defaults to 8
		{"", 8},      // Empty, defaults to 8
		{"invalid", 8}, // Invalid, defaults to 8
	}

	for _, tt := range tests {
		t.Run(tt.hitDice, func(t *testing.T) {
			result := service.getHitDiceValue(tt.hitDice)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAICharacterService_BuildCharacterPrompt(t *testing.T) {
	service := NewAICharacterService(nil)
	
	tests := []struct {
		name     string
		request  CustomCharacterRequest
		validate func(t *testing.T, prompt string)
	}{
		{
			name: "Complete Request",
			request: CustomCharacterRequest{
				Name:       "Thorin",
				Concept:    "A dwarven warrior seeking redemption",
				Race:       "Dwarf",
				Class:      "Paladin",
				Background: "Soldier",
				Ruleset:    "Pathfinder",
				Level:      5,
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Pathfinder")
				assert.Contains(t, prompt, "Name: Thorin")
				assert.Contains(t, prompt, "Concept: A dwarven warrior seeking redemption")
				assert.Contains(t, prompt, "Race: Dwarf")
				assert.Contains(t, prompt, "Class: Paladin")
				assert.Contains(t, prompt, "Background: Soldier")
				assert.Contains(t, prompt, "Level: 5")
			},
		},
		{
			name: "Minimal Request",
			request: CustomCharacterRequest{
				Name:    "Simple",
				Concept: "Basic hero",
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "D&D 5e") // Default ruleset
				assert.Contains(t, prompt, "Level: 1") // Default level
				assert.NotContains(t, prompt, "Race:")
				assert.NotContains(t, prompt, "Class:")
				assert.NotContains(t, prompt, "Background:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := service.buildCharacterPrompt(tt.request)
			tt.validate(t, prompt)
			
			// Always check for required JSON format instructions
			assert.Contains(t, prompt, "Generate a complete character with:")
			assert.Contains(t, prompt, "Respond with a JSON object in this format:")
		})
	}
}

func TestAICharacterService_CalculateSavingThrows(t *testing.T) {
	service := NewAICharacterService(nil)
	
	character := &models.Character{
		Attributes: models.Attributes{
			Strength:     14, // +2
			Dexterity:    16, // +3
			Constitution: 12, // +1
			Intelligence: 10, // +0
			Wisdom:       13, // +1
			Charisma:     8,  // -1
		},
	}
	
	saves := service.calculateSavingThrows(character)
	
	assert.Equal(t, 2, saves.Strength.Modifier)
	assert.Equal(t, 3, saves.Dexterity.Modifier)
	assert.Equal(t, 1, saves.Constitution.Modifier)
	assert.Equal(t, 0, saves.Intelligence.Modifier)
	assert.Equal(t, 1, saves.Wisdom.Modifier)
	assert.Equal(t, -1, saves.Charisma.Modifier)
	
	// All should be non-proficient in this basic implementation
	assert.False(t, saves.Strength.Proficiency)
	assert.False(t, saves.Dexterity.Proficiency)
	assert.False(t, saves.Constitution.Proficiency)
	assert.False(t, saves.Intelligence.Proficiency)
	assert.False(t, saves.Wisdom.Proficiency)
	assert.False(t, saves.Charisma.Proficiency)
}

func TestAICharacterService_ParseAIResponse_EdgeCases(t *testing.T) {
	service := NewAICharacterService(nil)
	
	tests := []struct {
		name        string
		aiResponse  string
		request     CustomCharacterRequest
		expectError bool
	}{
		{
			name:        "No JSON in response",
			aiResponse:  "This is just plain text with no JSON",
			expectError: true,
		},
		{
			name:        "Malformed JSON",
			aiResponse:  `{"race": "Elf", "class": `, // Incomplete JSON
			expectError: true,
		},
		{
			name: "JSON with extra text",
			aiResponse: `Here's your character: {"race": "Elf", "class": "Wizard", "attributes": {"strength": 10, "dexterity": 14, "constitution": 12, "intelligence": 16, "wisdom": 12, "charisma": 10}} That's all!`,
			request: CustomCharacterRequest{
				Name:  "Test",
				Level: 1,
			},
			expectError: false,
		},
		{
			name: "Multiple JSON objects",
			aiResponse: `{"wrong": "object"} {"race": "Human", "class": "Fighter", "attributes": {"strength": 16, "dexterity": 12, "constitution": 14, "intelligence": 10, "wisdom": 11, "charisma": 10}}`,
			request: CustomCharacterRequest{
				Name:  "Multi",
				Level: 1,
			},
			expectError: false, // Should parse the last valid JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.parseAIResponse(tt.aiResponse, tt.request)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkAICharacterService_GenerateCustomCharacter(b *testing.B) {
	mockLLM := new(MockLLMProviderForTest)
	service := NewAICharacterService(mockLLM)
	
	validResponse := `{
		"race": "Human",
		"class": "Fighter",
		"attributes": {
			"strength": 16,
			"dexterity": 14,
			"constitution": 15,
			"intelligence": 10,
			"wisdom": 12,
			"charisma": 8
		},
		"hitDice": "1d10",
		"speed": 30,
		"features": [],
		"proficiencies": {
			"armor": ["All armor", "Shields"],
			"weapons": ["Simple weapons", "Martial weapons"],
			"tools": [],
			"languages": ["Common"]
		},
		"skills": ["Athletics", "Intimidation"],
		"equipment": [],
		"personality": {
			"traits": ["Brave"],
			"ideals": ["Honor"],
			"bonds": ["Family"],
			"flaws": ["Stubborn"]
		},
		"backstory": "A simple soldier."
	}`
	
	mockLLM.On("GenerateCompletion", mock.Anything, mock.Anything, mock.Anything).
		Return(validResponse, nil)
	
	request := CustomCharacterRequest{
		Name:    "Benchmark Hero",
		Concept: "Test character",
		Level:   1,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateCustomCharacter(request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAICharacterService_ParseAIResponse(b *testing.B) {
	service := NewAICharacterService(nil)
	
	response := `{
		"race": "Elf",
		"class": "Ranger",
		"attributes": {
			"strength": 12,
			"dexterity": 17,
			"constitution": 14,
			"intelligence": 12,
			"wisdom": 15,
			"charisma": 10
		},
		"hitDice": "1d10",
		"speed": 35,
		"features": [
			{"name": "Darkvision", "description": "60 feet", "level": 1, "source": "Race"}
		],
		"proficiencies": {
			"armor": ["Light armor", "Medium armor", "Shields"],
			"weapons": ["Simple weapons", "Martial weapons"],
			"tools": [],
			"languages": ["Common", "Elvish"]
		},
		"skills": ["Animal Handling", "Perception", "Survival"],
		"equipment": [],
		"personality": {
			"traits": ["Silent", "Observant"],
			"ideals": ["Nature"],
			"bonds": ["Forest"],
			"flaws": ["Loner"]
		},
		"backstory": "Guardian of the forest."
	}`
	
	request := CustomCharacterRequest{
		Name:  "Benchmark",
		Level: 1,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.parseAIResponse(response, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}