package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// MockLLMProviderDMTest is a mock implementation for testing
type MockLLMProviderDMTest struct {
	mock.Mock
}

func (m *MockLLMProviderDMTest) GenerateCompletion(ctx context.Context, prompt, systemPrompt string) (string, error) {
	args := m.Called(ctx, prompt, systemPrompt)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProviderDMTest) GenerateContent(ctx context.Context, prompt, systemPrompt string) (string, error) {
	args := m.Called(ctx, prompt, systemPrompt)
	return args.String(0), args.Error(1)
}

func TestAIDMAssistantService_GenerateNPCDialog(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       models.NPCDialogRequest
		mockResponse  string
		mockError     error
		expected      string
		expectedError bool
	}{
		{
			name: "successful tavern keeper dialog",
			request: models.NPCDialogRequest{
				NPCName:         "Gareth the Barkeep",
				NPCPersonality:  []string{"gruff", "protective of locals", "secretly helpful"},
				DialogStyle:     "Working class, uses simple words, occasional grunt",
				Situation:       "Late evening in a busy tavern",
				PlayerInput:     "Do you know anything about the strange disappearances?",
				PreviousContext: "First time meeting",
			},
			mockResponse: `"*grunts* Aye, folk been vanishin' alright. Started 'bout two weeks back, always near the old mill. But you didn't hear that from me, understand?"`,
			expected:     `"*grunts* Aye, folk been vanishin' alright. Started 'bout two weeks back, always near the old mill. But you didn't hear that from me, understand?"`,
		},
		{
			name: "LLM provider error",
			request: models.NPCDialogRequest{
				NPCName:     "Test NPC",
				PlayerInput: "Hello",
			},
			mockError:     errors.New("LLM service unavailable"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderDMTest)

			if tt.mockError != nil {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return("", tt.mockError)
			} else {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(tt.mockResponse, nil)
			}

			service := NewAIDMAssistantService(mockLLM)
			result, err := service.GenerateNPCDialog(ctx, &tt.request)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			mockLLM.AssertExpectations(t)
		})
	}
}

func TestAIDMAssistantService_GenerateLocationDescription(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       models.LocationDescriptionRequest
		mockResponse  string
		mockError     error
		expectedError bool
	}{
		{
			name: "successful dungeon description",
			request: models.LocationDescriptionRequest{
				LocationType:    "dungeon",
				LocationName:    "The Forgotten Crypt",
				Atmosphere:      "ominous",
				TimeOfDay:       "eternal darkness",
				Weather:         "stale air",
				SpecialFeatures: []string{"ancient altar", "glowing runes"},
			},
			mockResponse: `{
				"description": "The Forgotten Crypt stretches before you, its ancient stones slick with moisture.",
				"atmosphere": "An oppressive weight fills the air, as if the darkness itself is watching.",
				"notableFeatures": ["Ancient altar covered in dried blood", "Glowing runes pulsing with eldritch energy"],
				"availableActions": ["Investigate the altar", "Decipher the runes", "Search for hidden passages"],
				"secretsAndHidden": [
					{
						"description": "A hidden compartment behind the altar",
						"discoveryDC": 15,
						"discoveryHint": "The altar seems slightly askew"
					}
				],
				"environmentalEffects": "The runes pulse faster when living creatures approach"
			}`,
		},
		{
			name: "fallback to simple text response",
			request: models.LocationDescriptionRequest{
				LocationType: "forest",
				LocationName: "Whispering Woods",
				Atmosphere:   "mysterious",
			},
			mockResponse: "A dark forest with tall trees that seem to whisper secrets in the wind.",
		},
		{
			name: "LLM error",
			request: models.LocationDescriptionRequest{
				LocationType: "test",
				LocationName: "Test Location",
			},
			mockError:     errors.New("LLM service error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderDMTest)

			if tt.mockError != nil {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return("", tt.mockError)
			} else {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(tt.mockResponse, nil)
			}

			service := NewAIDMAssistantService(mockLLM)
			result, err := service.GenerateLocationDescription(ctx, &tt.request)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Name)
				assert.NotEmpty(t, result.Type)
			}

			mockLLM.AssertExpectations(t)
		})
	}
}

func TestAIDMAssistantService_GenerateCombatNarration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       models.CombatNarrationRequest
		mockResponse  string
		expected      string
		expectedError bool
	}{
		{
			name: "successful hit narration",
			request: models.CombatNarrationRequest{
				AttackerName:  "Thorin",
				TargetName:    "Goblin",
				WeaponOrSpell: "battleaxe",
				IsHit:         true,
				Damage:        12,
				TargetHP:      3,
				TargetMaxHP:   15,
				IsCritical:    false,
			},
			mockResponse: "Thorin's battleaxe bites deep into the goblin's shoulder, sending it reeling back with a shriek of pain!",
			expected:     "Thorin's battleaxe bites deep into the goblin's shoulder, sending it reeling back with a shriek of pain!",
		},
		{
			name: "killing blow narration",
			request: models.CombatNarrationRequest{
				AttackerName:  "Elara",
				TargetName:    "Orc",
				WeaponOrSpell: "fireball",
				IsHit:         true,
				Damage:        25,
				TargetHP:      0,
				TargetMaxHP:   30,
				IsCritical:    true,
			},
			mockResponse: "Elara's fireball engulfs the orc in a brilliant explosion! The creature's final scream is cut short as it collapses into smoldering ash.",
			expected:     "Elara's fireball engulfs the orc in a brilliant explosion! The creature's final scream is cut short as it collapses into smoldering ash.",
		},
		{
			name: "miss narration",
			request: models.CombatNarrationRequest{
				AttackerName:  "Gimli",
				TargetName:    "Dragon",
				WeaponOrSpell: "throwing axe",
				IsHit:         false,
			},
			mockResponse: "Gimli's throwing axe whistles past the dragon's head, sparking off the cavern wall behind it!",
			expected:     "Gimli's throwing axe whistles past the dragon's head, sparking off the cavern wall behind it!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderDMTest)
			mockLLM.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(tt.mockResponse, nil)

			service := NewAIDMAssistantService(mockLLM)
			result, err := service.GenerateCombatNarration(ctx, &tt.request)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)

			mockLLM.AssertExpectations(t)
		})
	}
}

func TestAIDMAssistantService_GeneratePlotTwist(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		context       map[string]interface{}
		mockResponse  string
		expectedError bool
	}{
		{
			name: "successful plot twist generation",
			context: map[string]interface{}{
				"current_quest": "Find the missing prince",
				"location":      "Capital city",
				"party_level":   5,
			},
			mockResponse: `{
				"title": "The Prince's Dark Secret",
				"description": "The missing prince wasn't kidnapped - he fled after discovering he's actually the bastard son of the villain the party has been hunting.",
				"suggestedTiming": "When the party finds evidence of the prince's whereabouts",
				"prerequisites": ["Party must trust the royal family", "Villain's identity should be known"],
				"consequences": ["Royal family reputation damaged", "Prince becomes potential ally or enemy"],
				"foreshadowingHints": ["Prince seemed troubled before disappearing", "Villain has been avoiding the capital"],
				"impactLevel": "major"
			}`,
		},
		{
			name:          "LLM error",
			context:       map[string]interface{}{"test": "data"},
			mockResponse:  "",
			expectedError: true,
		},
		{
			name:          "invalid JSON response",
			context:       map[string]interface{}{"test": "data"},
			mockResponse:  "This is not valid JSON",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderDMTest)

			if tt.expectedError && tt.mockResponse == "" {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return("", errors.New("LLM error"))
			} else {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(tt.mockResponse, nil)
			}

			service := NewAIDMAssistantService(mockLLM)
			result, err := service.GeneratePlotTwist(ctx, tt.context)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Title)
				assert.NotEmpty(t, result.Description)
			}

			mockLLM.AssertExpectations(t)
		})
	}
}

func TestAIDMAssistantService_GenerateEnvironmentalHazard(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		locationType  string
		difficulty    int
		mockResponse  string
		expectedError bool
	}{
		{
			name:         "successful hazard generation",
			locationType: "dungeon",
			difficulty:   5,
			mockResponse: `{
				"name": "Collapsing Ceiling",
				"description": "Cracks spider-web across the ancient ceiling, dust raining down with each footstep",
				"triggerCondition": "Loud noise or vibration (combat, explosion, etc.)",
				"effectDescription": "Chunks of stone crash down in a 10-foot radius",
				"mechanicalEffects": {
					"save": "DEX",
					"difficultyClass": 14,
					"damage": "2d10",
					"damageType": "bludgeoning",
					"additionalEffects": "Knocked prone on failed save"
				},
				"avoidanceHints": "DC 12 Perception to notice the unstable ceiling",
				"isTrap": false,
				"isNatural": true
			}`,
		},
		{
			name:          "invalid JSON response",
			locationType:  "forest",
			difficulty:    3,
			mockResponse:  "Not valid JSON",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderDMTest)
			mockLLM.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(tt.mockResponse, nil)

			service := NewAIDMAssistantService(mockLLM)
			result, err := service.GenerateEnvironmentalHazard(ctx, tt.locationType, tt.difficulty)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Name)
				assert.NotEmpty(t, result.Description)
			}

			mockLLM.AssertExpectations(t)
		})
	}
}

func TestAIDMAssistantService_GenerateNPC(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		role          string
		context       map[string]interface{}
		mockResponse  string
		expectedError bool
	}{
		{
			name: "successful NPC generation",
			role: "Mysterious Merchant",
			context: map[string]interface{}{
				"location": "Desert trading post",
				"theme":    "Arabian Nights inspired",
			},
			mockResponse: `{
				"name": "Rashid al-Maliki",
				"race": "Human",
				"occupation": "Traveling merchant and information broker",
				"personalityTraits": ["shrewd", "well-connected", "superstitious"],
				"appearance": "Weathered face with deep laugh lines, colorful silk robes, numerous rings",
				"voiceDescription": "Smooth baritone with musical accent, often speaks in riddles",
				"motivations": "Seeks rare artifacts and forbidden knowledge to sell to the highest bidder",
				"secrets": "Is actually an agent of a djinn, bound to gather information",
				"dialogStyle": "Flowery speech, uses many metaphors, addresses everyone as 'friend'",
				"relationshipToParty": "Cautiously interested, sees potential profit"
			}`,
		},
		{
			name:          "LLM error",
			role:          "Guard Captain",
			context:       map[string]interface{}{},
			mockResponse:  "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderDMTest)

			if tt.expectedError {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return("", errors.New("LLM error"))
			} else {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(tt.mockResponse, nil)
			}

			service := NewAIDMAssistantService(mockLLM)
			result, err := service.GenerateNPC(ctx, tt.role, tt.context)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Name)
				assert.NotEmpty(t, result.Race)
				assert.NotEmpty(t, result.Occupation)
			}

			mockLLM.AssertExpectations(t)
		})
	}
}
