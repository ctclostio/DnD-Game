package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// MockLLMProviderDMTest is a mock implementation for testing
type MockLLMProviderDMTest struct {
	mock.Mock
}

func (m *MockLLMProviderDMTest) GenerateCompletion(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	args := m.Called(ctx, prompt, systemPrompt)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProviderDMTest) GenerateContent(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	args := m.Called(ctx, prompt, systemPrompt)
	return args.String(0), args.Error(1)
}

func TestAIDMAssistantService_GenerateNPCDialogue(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       models.NPCDialogueRequest
		mockResponse  string
		mockError     error
		expected      string
		expectedError string
	}{
		{
			name: "successful tavern keeper dialogue",
			request: models.NPCDialogueRequest{
				NPCName:        "Gareth the Barkeep",
				NPCPersonality: []string{"gruff", "protective of locals", "secretly helpful"},
				DialogueStyle:  "Working class, uses simple words, occasional grunt",
				Situation:      "Late evening in a busy tavern",
				PlayerInput:    "Do you know anything about the strange disappearances?",
				PreviousContext: "First time meeting",
			},
			mockResponse: `"*grunts* Aye, folk been vanishin' alright. Started 'bout two weeks back, always near the old mill. But you didn't hear that from me, understand?"`,
			expected:     `"*grunts* Aye, folk been vanishin' alright. Started 'bout two weeks back, always near the old mill. But you didn't hear that from me, understand?"`,
		},
		{
			name: "noble with complex dialogue",
			request: models.NPCDialogueRequest{
				NPCName:        "Lady Elara Moonwhisper",
				NPCPersonality: []string{"cunning", "polite", "manipulative", "ambitious"},
				DialogueStyle:  "Formal, eloquent, uses metaphors, subtle threats",
				Situation:      "Private meeting in her manor",
				PlayerInput:    "I know you're involved with the thieves guild",
				PreviousContext: "Player previously helped her with a task",
			},
			mockResponse: `"My dear friend, one must be careful when casting shadows upon another's reputation. After all, even the moon has its dark side, yet it still illuminates the night. Perhaps we should discuss... mutual interests instead?"`,
			expected:     `"My dear friend, one must be careful when casting shadows upon another's reputation. After all, even the moon has its dark side, yet it still illuminates the night. Perhaps we should discuss... mutual interests instead?"`,
		},
		{
			name: "scared commoner",
			request: models.NPCDialogueRequest{
				NPCName:        "Timothy the Farmer",
				NPCPersonality: []string{"frightened", "honest", "simple"},
				DialogueStyle:  "Stuttering, nervous, rural dialect",
				Situation:      "After witnessing a monster attack",
				PlayerInput:    "What did you see?",
				PreviousContext: "Found hiding in barn",
			},
			mockResponse: `"I-it were terrible, m'lord! Big as a house it were, with eyes like fire and... and teeth like swords! Took me best cow, it did!"`,
			expected:     `"I-it were terrible, m'lord! Big as a house it were, with eyes like fire and... and teeth like swords! Took me best cow, it did!"`,
		},
		{
			name:          "empty NPC name",
			request:       models.NPCDialogueRequest{},
			expectedError: "NPC name is required",
		},
		{
			name: "LLM provider error",
			request: models.NPCDialogueRequest{
				NPCName:     "Test NPC",
				PlayerInput: "Hello",
			},
			mockError:     errors.New("LLM service unavailable"),
			expectedError: "failed to generate dialogue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderDMTest)

			if tt.expectedError == "" && tt.request.NPCName != "" {
				mockLLM.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(tt.mockResponse, tt.mockError)
			}

			service := NewAIDMAssistantService(mockLLM)
			result, err := service.GenerateNPCDialogue(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			if tt.request.NPCName != "" && tt.expectedError == "" {
				mockLLM.AssertExpectations(t)
			}
		})
	}
}

func TestAIDMAssistantService_GenerateEnvironmentDescription(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       models.EnvironmentRequest
		mockResponse  string
		expected      models.EnvironmentDescription
		expectedError string
	}{
		{
			name: "dungeon entrance description",
			request: models.EnvironmentRequest{
				Location:    "Ancient dungeon entrance",
				TimeOfDay:   "dusk",
				Weather:     "foggy",
				Atmosphere:  "ominous",
				KeyFeatures: []string{"carved stone door", "strange symbols", "recent tracks"},
			},
			mockResponse: `{
				"description": "As dusk settles, thick fog coils around the ancient stone entrance like ghostly fingers. The massive door, carved with symbols that seem to writhe in the dying light, stands partially ajar. Fresh tracks in the mud suggest you're not the first to find this place recently.",
				"sensory_details": {
					"sight": "Writhing shadows dance across worn stone, symbols glow faintly with an otherworldly light",
					"sound": "Distant dripping water echoes from within, the door groans on ancient hinges",
					"smell": "Damp earth mixed with something acrid and unnatural",
					"touch": "The stone feels unnaturally cold, almost pulling heat from your hands",
					"taste": "The fog leaves a metallic taste on your tongue"
				},
				"points_of_interest": [
					"Glowing symbols that react to proximity",
					"Fresh muddy footprints leading inside",
					"Scratches on the door frame - something tried to get out"
				],
				"potential_hazards": [
					"Unstable ceiling near entrance",
					"Magical ward on the door symbols",
					"Slippery moss-covered stones"
				],
				"hidden_elements": [
					"Secret compartment behind loose stone to the left of door",
					"Barely visible tripwire just inside entrance"
				]
			}`,
			expected: models.EnvironmentDescription{
				Description: "As dusk settles, thick fog coils around the ancient stone entrance like ghostly fingers. The massive door, carved with symbols that seem to writhe in the dying light, stands partially ajar. Fresh tracks in the mud suggest you're not the first to find this place recently.",
				SensoryDetails: models.SensoryDetails{
					Sight: "Writhing shadows dance across worn stone, symbols glow faintly with an otherworldly light",
					Sound: "Distant dripping water echoes from within, the door groans on ancient hinges",
					Smell: "Damp earth mixed with something acrid and unnatural",
					Touch: "The stone feels unnaturally cold, almost pulling heat from your hands",
					Taste: "The fog leaves a metallic taste on your tongue",
				},
				PointsOfInterest: []string{
					"Glowing symbols that react to proximity",
					"Fresh muddy footprints leading inside",
					"Scratches on the door frame - something tried to get out",
				},
				PotentialHazards: []string{
					"Unstable ceiling near entrance",
					"Magical ward on the door symbols",
					"Slippery moss-covered stones",
				},
				HiddenElements: []string{
					"Secret compartment behind loose stone to the left of door",
					"Barely visible tripwire just inside entrance",
				},
			},
		},
		{
			name: "tavern scene",
			request: models.EnvironmentRequest{
				Location:    "The Prancing Pony Tavern",
				TimeOfDay:   "evening",
				Weather:     "rainy",
				Atmosphere:  "lively",
				KeyFeatures: []string{"fireplace", "bar", "shadowy corner booth"},
			},
			mockResponse: `{
				"description": "The Prancing Pony thrums with life as rain patters against diamond-paned windows. Warm firelight dances across rough-hewn beams while patrons laugh over tankards of ale. The air is thick with pipe smoke and the promise of stories yet untold.",
				"sensory_details": {
					"sight": "Golden firelight, shadows dancing on weathered wood, rain-streaked windows",
					"sound": "Crackling fire, boisterous laughter, rain on the roof, a bard tuning a lute",
					"smell": "Wood smoke, spilled ale, roasting meat, wet wool from rain-soaked cloaks",
					"touch": "Warm air with occasional cool drafts, smooth worn wood of the bar",
					"taste": "Smoky air, the lingering taste of hearty stew"
				},
				"points_of_interest": [
					"Animated conversation at the bar about local rumors",
					"Hooded figure alone in the corner booth",
					"Bard preparing to perform"
				],
				"potential_hazards": [
					"Drunk patron looking for a fight",
					"Pickpocket working the crowd",
					"Loose floorboard near the stairs"
				],
				"hidden_elements": [
					"Message carved under corner table",
					"Barkeep's hidden weapon under the bar"
				]
			}`,
			expected: models.EnvironmentDescription{
				Description: "The Prancing Pony thrums with life as rain patters against diamond-paned windows. Warm firelight dances across rough-hewn beams while patrons laugh over tankards of ale. The air is thick with pipe smoke and the promise of stories yet untold.",
				SensoryDetails: models.SensoryDetails{
					Sight: "Golden firelight, shadows dancing on weathered wood, rain-streaked windows",
					Sound: "Crackling fire, boisterous laughter, rain on the roof, a bard tuning a lute",
					Smell: "Wood smoke, spilled ale, roasting meat, wet wool from rain-soaked cloaks",
					Touch: "Warm air with occasional cool drafts, smooth worn wood of the bar",
					Taste: "Smoky air, the lingering taste of hearty stew",
				},
				PointsOfInterest: []string{
					"Animated conversation at the bar about local rumors",
					"Hooded figure alone in the corner booth",
					"Bard preparing to perform",
				},
				PotentialHazards: []string{
					"Drunk patron looking for a fight",
					"Pickpocket working the crowd",
					"Loose floorboard near the stairs",
				},
				HiddenElements: []string{
					"Message carved under corner table",
					"Barkeep's hidden weapon under the bar",
				},
			},
		},
		{
			name:          "empty location",
			request:       models.EnvironmentRequest{},
			expectedError: "location is required",
		},
		{
			name: "invalid JSON response",
			request: models.EnvironmentRequest{
				Location: "Test location",
			},
			mockResponse:  "This is not valid JSON",
			expectedError: "failed to parse environment description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderDMTest)

			if tt.expectedError == "" && tt.request.Location != "" {
				mockLLM.On("Chat", mock.AnythingOfType("string")).Return(tt.mockResponse, nil)
			} else if tt.request.Location != "" && tt.expectedError == "failed to parse environment description" {
				mockLLM.On("Chat", mock.AnythingOfType("string")).Return(tt.mockResponse, nil)
			}

			service := NewAIDMAssistantService(mockLLM)
			result, err := service.GenerateEnvironmentDescription(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			if tt.request.Location != "" && tt.expectedError != "location is required" {
				mockLLM.AssertExpectations(t)
			}
		})
	}
}

func TestAIDMAssistantService_GeneratePlotHook(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       models.PlotHookRequest
		mockResponse  string
		expected      models.PlotHook
		expectedError string
	}{
		{
			name: "mystery plot hook",
			request: models.PlotHookRequest{
				Theme:            "mystery",
				PartyLevel:       5,
				Setting:          "coastal town",
				PartyComposition: []string{"fighter", "rogue", "wizard", "cleric"},
				CurrentSituation: "Party just arrived in town",
			},
			mockResponse: `{
				"title": "The Vanishing Tide",
				"hook": "As you enter the coastal town of Saltmere, you notice the usually bustling harbor is eerily quiet. A distraught fisherman approaches: 'Please, you look like capable folk! Every night when the tide goes out, someone disappears. Last night it was the lighthouse keeper's daughter. The town guard won't go near the old sea caves anymore.'",
				"background": "Two weeks ago, ancient ruins were exposed by an unusually low tide. Since then, townsfolk have been vanishing, leaving only wet footprints that lead to the sea.",
				"key_npcs": [
					{
						"name": "Captain Marea Stormwind",
						"role": "Retired Naval Officer",
						"motivation": "Protect the town, knows more than she lets on"
					},
					{
						"name": "Thessian the Sage",
						"role": "Local Scholar",
						"motivation": "Obsessed with the ruins, might be involved"
					}
				],
				"initial_clues": [
					"Victims all had nightmares about drowning before disappearing",
					"Strange algae found at disappearance sites doesn't match local varieties",
					"Old town records mention a 'Tide Cult' from 200 years ago"
				],
				"potential_rewards": [
					"500 gold from the town council",
					"Magical trident in the sea caves",
					"Favor of the local merchants guild"
				],
				"escalation": "If ignored, a party member will have the drowning nightmare and be compelled to walk to the sea at next low tide"
			}`,
			expected: models.PlotHook{
				Title: "The Vanishing Tide",
				Hook:  "As you enter the coastal town of Saltmere, you notice the usually bustling harbor is eerily quiet. A distraught fisherman approaches: 'Please, you look like capable folk! Every night when the tide goes out, someone disappears. Last night it was the lighthouse keeper's daughter. The town guard won't go near the old sea caves anymore.'",
				Background: "Two weeks ago, ancient ruins were exposed by an unusually low tide. Since then, townsfolk have been vanishing, leaving only wet footprints that lead to the sea.",
				KeyNPCs: []models.PlotNPC{
					{
						Name:       "Captain Marea Stormwind",
						Role:       "Retired Naval Officer",
						Motivation: "Protect the town, knows more than she lets on",
					},
					{
						Name:       "Thessian the Sage",
						Role:       "Local Scholar",
						Motivation: "Obsessed with the ruins, might be involved",
					},
				},
				InitialClues: []string{
					"Victims all had nightmares about drowning before disappearing",
					"Strange algae found at disappearance sites doesn't match local varieties",
					"Old town records mention a 'Tide Cult' from 200 years ago",
				},
				PotentialRewards: []string{
					"500 gold from the town council",
					"Magical trident in the sea caves",
					"Favor of the local merchants guild",
				},
				Escalation: "If ignored, a party member will have the drowning nightmare and be compelled to walk to the sea at next low tide",
			},
		},
		{
			name: "combat-focused plot hook",
			request: models.PlotHookRequest{
				Theme:            "action",
				PartyLevel:       8,
				Setting:          "mountain pass",
				PartyComposition: []string{"barbarian", "ranger", "sorcerer", "bard"},
				CurrentSituation: "Traveling to the capital",
			},
			mockResponse: `{
				"title": "Ambush at Thunder Peak",
				"hook": "The mountain path suddenly explodes with activity as boulders crash down from above! Through the dust, you see organized brigands taking positions. Their leader, wearing distinctive crimson armor, shouts: 'The Iron Lord sends his regards!'",
				"background": "The Iron Lord, a mysterious warlord, has been consolidating power in the mountains and now controls the main trade route.",
				"key_npcs": [
					{
						"name": "Korvan the Red",
						"role": "Brigand Captain",
						"motivation": "Loyal to Iron Lord, can be intimidated for information"
					}
				],
				"initial_clues": [
					"The brigands wear matching iron pendants",
					"They seem to have been waiting specifically for your party",
					"Map found on captain shows other ambush points"
				],
				"potential_rewards": [
					"Brigand's treasure cache nearby",
					"Crimson armor set",
					"Information about Iron Lord's fortress"
				],
				"escalation": "More brigand groups will pursue the party, growing stronger each time"
			}`,
			expected: models.PlotHook{
				Title:      "Ambush at Thunder Peak",
				Hook:       "The mountain path suddenly explodes with activity as boulders crash down from above! Through the dust, you see organized brigands taking positions. Their leader, wearing distinctive crimson armor, shouts: 'The Iron Lord sends his regards!'",
				Background: "The Iron Lord, a mysterious warlord, has been consolidating power in the mountains and now controls the main trade route.",
				KeyNPCs: []models.PlotNPC{
					{
						Name:       "Korvan the Red",
						Role:       "Brigand Captain",
						Motivation: "Loyal to Iron Lord, can be intimidated for information",
					},
				},
				InitialClues: []string{
					"The brigands wear matching iron pendants",
					"They seem to have been waiting specifically for your party",
					"Map found on captain shows other ambush points",
				},
				PotentialRewards: []string{
					"Brigand's treasure cache nearby",
					"Crimson armor set",
					"Information about Iron Lord's fortress",
				},
				Escalation: "More brigand groups will pursue the party, growing stronger each time",
			},
		},
		{
			name: "missing required fields",
			request: models.PlotHookRequest{
				Theme: "mystery",
			},
			expectedError: "party level must be greater than 0",
		},
		{
			name: "LLM error",
			request: models.PlotHookRequest{
				Theme:      "mystery",
				PartyLevel: 5,
			},
			mockResponse:  "",
			expectedError: "failed to generate plot hook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderDMTest)

			if tt.expectedError == "" && tt.request.PartyLevel > 0 {
				mockLLM.On("Chat", mock.AnythingOfType("string")).Return(tt.mockResponse, nil)
			} else if tt.request.PartyLevel > 0 && tt.expectedError == "failed to generate plot hook" {
				mockLLM.On("Chat", mock.AnythingOfType("string")).Return("", errors.New("LLM error"))
			}

			service := NewAIDMAssistantService(mockLLM)
			result, err := service.GeneratePlotHook(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			if tt.request.PartyLevel > 0 && tt.expectedError != "party level must be greater than 0" {
				mockLLM.AssertExpectations(t)
			}
		})
	}
}

func TestAIDMAssistantService_SuggestRuling(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       models.RulingRequest
		mockResponse  string
		expected      models.RulingSuggestion
		expectedError string
	}{
		{
			name: "creative spell use ruling",
			request: models.RulingRequest{
				Situation:   "Player wants to use Shape Water cantrip to create ice daggers and throw them",
				RulesContext: "Shape Water can freeze water and move it, but doesn't mention creating weapons",
				PlayerIntent: "Deal damage using a creative application of a utility cantrip",
			},
			mockResponse: `{
				"ruling": "Allow it with limitations: The player can create and throw ice daggers using Shape Water, but as an improvised weapon attack (1d4 damage) rather than a spell attack. Creating the dagger requires the cantrip action, throwing is a separate action.",
				"reasoning": "This rewards creative thinking while maintaining game balance. Shape Water can freeze water into simple shapes, and ice daggers are reasonable. However, since the cantrip isn't designed for damage, treating them as improvised weapons prevents it from overshadowing actual damage cantrips.",
				"precedent": "Similar to allowing Prestidigitation to create small distractions or Mold Earth to create difficult terrain - utility cantrips can have combat applications with appropriate limitations.",
				"alternatives": [
					"Require a Dexterity check to properly form the daggers",
					"Allow it once per combat as a surprise tactic",
					"Require existing water source within 5 feet"
				],
				"balance_considerations": "This ruling keeps the cantrip useful without making it better than Fire Bolt or other damage cantrips. The two-action requirement (create then throw) balances the versatility."
			}`,
			expected: models.RulingSuggestion{
				Ruling:    "Allow it with limitations: The player can create and throw ice daggers using Shape Water, but as an improvised weapon attack (1d4 damage) rather than a spell attack. Creating the dagger requires the cantrip action, throwing is a separate action.",
				Reasoning: "This rewards creative thinking while maintaining game balance. Shape Water can freeze water into simple shapes, and ice daggers are reasonable. However, since the cantrip isn't designed for damage, treating them as improvised weapons prevents it from overshadowing actual damage cantrips.",
				Precedent: "Similar to allowing Prestidigitation to create small distractions or Mold Earth to create difficult terrain - utility cantrips can have combat applications with appropriate limitations.",
				Alternatives: []string{
					"Require a Dexterity check to properly form the daggers",
					"Allow it once per combat as a surprise tactic",
					"Require existing water source within 5 feet",
				},
				BalanceConsiderations: "This ruling keeps the cantrip useful without making it better than Fire Bolt or other damage cantrips. The two-action requirement (create then throw) balances the versatility.",
			},
		},
		{
			name: "skill check combination",
			request: models.RulingRequest{
				Situation:   "Rogue wants to use Acrobatics instead of Athletics to climb because they're parkour-trained",
				RulesContext: "Rules specify Athletics for climbing",
				PlayerIntent: "Use character backstory to justify alternate skill use",
			},
			mockResponse: `{
				"ruling": "Allow Acrobatics for this specific climb, but at a higher DC (+2 to +5 depending on the surface). The rogue must describe how they're using momentum and agility rather than strength.",
				"reasoning": "Character backstory and creative problem-solving should be rewarded. The increased DC reflects that while possible, it's more challenging than the standard approach.",
				"precedent": "Similar to allowing Intimidation with Strength instead of Charisma, or Medicine with Intelligence for academic knowledge.",
				"alternatives": [
					"Allow it at normal DC but only for certain surfaces (buildings, not natural cliffs)",
					"Require proficiency in both skills",
					"Give advantage on Athletics checks for climbing instead"
				],
				"balance_considerations": "This doesn't break balance as it's still consuming the same resources (action, skill check) with appropriate difficulty adjustments."
			}`,
			expected: models.RulingSuggestion{
				Ruling:    "Allow Acrobatics for this specific climb, but at a higher DC (+2 to +5 depending on the surface). The rogue must describe how they're using momentum and agility rather than strength.",
				Reasoning: "Character backstory and creative problem-solving should be rewarded. The increased DC reflects that while possible, it's more challenging than the standard approach.",
				Precedent: "Similar to allowing Intimidation with Strength instead of Charisma, or Medicine with Intelligence for academic knowledge.",
				Alternatives: []string{
					"Allow it at normal DC but only for certain surfaces (buildings, not natural cliffs)",
					"Require proficiency in both skills",
					"Give advantage on Athletics checks for climbing instead",
				},
				BalanceConsiderations: "This doesn't break balance as it's still consuming the same resources (action, skill check) with appropriate difficulty adjustments.",
			},
		},
		{
			name:          "empty situation",
			request:       models.RulingRequest{},
			expectedError: "situation description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderDMTest)

			if tt.expectedError == "" && tt.request.Situation != "" {
				mockLLM.On("Chat", mock.AnythingOfType("string")).Return(tt.mockResponse, nil)
			}

			service := NewAIDMAssistantService(mockLLM)
			result, err := service.SuggestRuling(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			if tt.request.Situation != "" && tt.expectedError == "" {
				mockLLM.AssertExpectations(t)
			}
		})
	}
}

func TestAIDMAssistantService_GenerateTreasure(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       models.TreasureRequest
		mockResponse  string
		expected      models.TreasureHoard
		expectedError string
	}{
		{
			name: "level 5 dungeon treasure",
			request: models.TreasureRequest{
				ChallengeRating: 5,
				TreasureType:    "dungeon hoard",
				Context:         "Ancient wizard's laboratory",
				PartySize:       4,
			},
			mockResponse: `{
				"total_value": 750,
				"coins": {
					"copper": 0,
					"silver": 200,
					"gold": 150,
					"platinum": 10
				},
				"gems": [
					{
						"name": "Moss Agate",
						"value": 10,
						"quantity": 3,
						"description": "Translucent green stone with moss-like inclusions"
					},
					{
						"name": "Amethyst",
						"value": 100,
						"quantity": 1,
						"description": "Deep purple crystal, faintly glowing"
					}
				],
				"art_objects": [
					{
						"name": "Ancient Spellbook",
						"value": 250,
						"description": "Leather-bound tome with silver clasps, contains theoretical magic essays"
					}
				],
				"magic_items": [
					{
						"name": "Wand of Magic Detection",
						"rarity": "uncommon",
						"description": "This wand has 3 charges. While holding it, you can expend 1 charge as an action to cast Detect Magic. Regains 1d3 charges at dawn.",
						"properties": ["requires attunement by a spellcaster"]
					},
					{
						"name": "Dust of Disappearance",
						"rarity": "uncommon",
						"description": "Found in a small packet, this powder resembles very fine sand. When thrown into the air, it covers a 10-foot square, rendering creatures invisible for 2d4 minutes.",
						"properties": ["consumable", "one use"]
					}
				],
				"special_items": [
					"Wizard's research notes on planar travel (plot hook)",
					"Map showing location of sister laboratory"
				]
			}`,
			expected: models.TreasureHoard{
				TotalValue: 750,
				Coins: models.CoinageBreakdown{
					Copper:   0,
					Silver:   200,
					Gold:     150,
					Platinum: 10,
				},
				Gems: []models.Gem{
					{
						Name:        "Moss Agate",
						Value:       10,
						Quantity:    3,
						Description: "Translucent green stone with moss-like inclusions",
					},
					{
						Name:        "Amethyst",
						Value:       100,
						Quantity:    1,
						Description: "Deep purple crystal, faintly glowing",
					},
				},
				ArtObjects: []models.ArtObject{
					{
						Name:        "Ancient Spellbook",
						Value:       250,
						Description: "Leather-bound tome with silver clasps, contains theoretical magic essays",
					},
				},
				MagicItemDetails: []models.MagicItem{
					{
						Name:        "Wand of Magic Detection",
						Rarity:      "uncommon",
						Description: "This wand has 3 charges. While holding it, you can expend 1 charge as an action to cast Detect Magic. Regains 1d3 charges at dawn.",
						Properties:  []string{"requires attunement by a spellcaster"},
					},
					{
						Name:        "Dust of Disappearance",
						Rarity:      "uncommon",
						Description: "Found in a small packet, this powder resembles very fine sand. When thrown into the air, it covers a 10-foot square, rendering creatures invisible for 2d4 minutes.",
						Properties:  []string{"consumable", "one use"},
					},
				},
				SpecialItems: []string{
					"Wizard's research notes on planar travel (plot hook)",
					"Map showing location of sister laboratory",
				},
			},
		},
		{
			name:          "invalid challenge rating",
			request:       models.TreasureRequest{},
			expectedError: "challenge rating must be between 0 and 30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMProviderDMTest)

			if tt.expectedError == "" && tt.request.ChallengeRating > 0 {
				mockLLM.On("Chat", mock.AnythingOfType("string")).Return(tt.mockResponse, nil)
			}

			service := NewAIDMAssistantService(mockLLM)
			result, err := service.GenerateTreasure(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			if tt.request.ChallengeRating > 0 && tt.expectedError == "" {
				mockLLM.AssertExpectations(t)
			}
		})
	}
}

// Stub implementations for testing - these would normally be in the actual service file
func (s *AIDMAssistantService) GenerateEnvironmentDescription(ctx context.Context, req models.EnvironmentRequest) (models.EnvironmentDescription, error) {
	if req.Location == "" {
		return models.EnvironmentDescription{}, errors.New("location is required")
	}

	prompt := "Generate environment description based on request"
	response, err := s.llmProvider.GenerateCompletion(ctx, prompt, "")
	if err != nil {
		return models.EnvironmentDescription{}, err
	}

	var result models.EnvironmentDescription
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return models.EnvironmentDescription{}, errors.New("failed to parse environment description")
	}

	return result, nil
}

func (s *AIDMAssistantService) GeneratePlotHook(ctx context.Context, req models.PlotHookRequest) (models.PlotHook, error) {
	if req.PartyLevel <= 0 {
		return models.PlotHook{}, errors.New("party level must be greater than 0")
	}

	prompt := "Generate plot hook based on request"
	response, err := s.llmProvider.GenerateCompletion(ctx, prompt, "")
	if err != nil {
		return models.PlotHook{}, errors.New("failed to generate plot hook")
	}

	var result models.PlotHook
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return models.PlotHook{}, errors.New("failed to parse plot hook")
	}

	return result, nil
}

func (s *AIDMAssistantService) SuggestRuling(ctx context.Context, req models.RulingRequest) (models.RulingSuggestion, error) {
	if req.Situation == "" {
		return models.RulingSuggestion{}, errors.New("situation description is required")
	}

	prompt := "Suggest ruling based on request"
	response, err := s.llmProvider.GenerateCompletion(ctx, prompt, "")
	if err != nil {
		return models.RulingSuggestion{}, err
	}

	var result models.RulingSuggestion
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return models.RulingSuggestion{}, err
	}

	return result, nil
}

func (s *AIDMAssistantService) GenerateTreasure(ctx context.Context, req models.TreasureRequest) (models.TreasureHoard, error) {
	if req.ChallengeRating < 0 || req.ChallengeRating > 30 {
		return models.TreasureHoard{}, errors.New("challenge rating must be between 0 and 30")
	}

	prompt := "Generate treasure based on request"
	response, err := s.llmProvider.GenerateCompletion(ctx, prompt, "")
	if err != nil {
		return models.TreasureHoard{}, err
	}

	var result models.TreasureHoard
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return models.TreasureHoard{}, err
	}

	return result, nil
}