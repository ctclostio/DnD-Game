package mocks

import (
	"context"
	"encoding/json"
)

// MockLLMProvider is a mock implementation of LLMProvider for testing
type MockLLMProvider struct {
	ResponseFunc func(context.Context, string, string) (string, error)
	Responses    map[string]string
}

func (m *MockLLMProvider) GenerateCompletion(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	if m.ResponseFunc != nil {
		return m.ResponseFunc(ctx, systemPrompt, prompt)
	}
	
	// Default responses for common AI requests
	if m.Responses != nil {
		for key, response := range m.Responses {
			if contains(prompt, key) || contains(systemPrompt, key) {
				return response, nil
			}
		}
	}
	
	// Default mock responses based on content
	switch {
	case contains(prompt, "race") || contains(systemPrompt, "race"):
		return mockRaceResponse(), nil
	case contains(prompt, "class") || contains(systemPrompt, "class"):
		return mockClassResponse(), nil
	case contains(prompt, "backstory"):
		return `{"backstory": "A mysterious wanderer with a hidden past..."}`, nil
	case contains(prompt, "name"):
		return `{"name": "Testarian McTestface"}`, nil
	case contains(prompt, "npc"):
		return mockNPCResponse(), nil
	case contains(prompt, "dialogue"):
		return `"Greetings, adventurer! How may I assist you today?"`, nil
	case contains(prompt, "location") || contains(prompt, "description"):
		return mockLocationResponse(), nil
	case contains(prompt, "combat") || contains(prompt, "narration"):
		return `"The battle rages on! Your sword strikes true!"`, nil
	case contains(prompt, "hazard"):
		return mockHazardResponse(), nil
	case contains(prompt, "story") || contains(prompt, "arc"):
		return mockStoryArcResponse(), nil
	case contains(prompt, "settlement"):
		return mockSettlementResponse(), nil
	case contains(prompt, "recap"):
		return mockRecapResponse(), nil
	default:
		return `{"response": "Mock AI response"}`, nil
	}
}

// GenerateContent is an alias for GenerateCompletion to satisfy the interface
func (m *MockLLMProvider) GenerateContent(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	return m.GenerateCompletion(ctx, prompt, systemPrompt)
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(str) > 0 && len(substr) > 0 && str[0:len(substr)] == substr || (len(str) > len(substr) && str[len(str)-len(substr):] == substr) || (len(substr) > 0 && len(str) > len(substr) && findSubstring(str, substr)))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func mockRaceResponse() string {
	race := map[string]interface{}{
		"name": "Shadowkin",
		"description": "A mysterious race from the shadow realm",
		"traits": []map[string]interface{}{
			{"name": "Darkvision", "description": "You can see in darkness"},
			{"name": "Shadow Step", "description": "Teleport through shadows"},
		},
		"abilityScoreIncrease": map[string]int{
			"dexterity": 2,
			"charisma": 1,
		},
		"size": "Medium",
		"speed": 30,
	}
	data, _ := json.Marshal(race)
	return string(data)
}

func mockClassResponse() string {
	class := map[string]interface{}{
		"name": "Shadowblade",
		"description": "A warrior who channels shadow magic",
		"hitDice": "1d10",
		"primaryAbility": "Dexterity",
		"savingThrows": []string{"Dexterity", "Intelligence"},
		"features": []map[string]interface{}{
			{
				"name": "Shadow Strike",
				"level": 1,
				"description": "Infuse your attacks with shadow energy",
			},
		},
	}
	data, _ := json.Marshal(class)
	return string(data)
}

func mockNPCResponse() string {
	npc := map[string]interface{}{
		"name": "Eldrin the Wise",
		"race": "Elf",
		"class": "Wizard",
		"level": 10,
		"personality": "Scholarly and mysterious",
		"motivation": "Seeks ancient knowledge",
		"description": "An elderly elf with piercing blue eyes",
	}
	data, _ := json.Marshal(npc)
	return string(data)
}

func mockLocationResponse() string {
	location := map[string]interface{}{
		"name": "The Whispering Woods",
		"type": "forest",
		"description": "Ancient trees whose leaves seem to whisper secrets",
		"atmosphere": "Mysterious and foreboding",
		"pointsOfInterest": []string{"Ancient shrine", "Hidden grove"},
		"potentialEncounters": []string{"Forest spirits", "Lost travelers"},
	}
	data, _ := json.Marshal(location)
	return string(data)
}

func mockHazardResponse() string {
	hazard := map[string]interface{}{
		"name": "Poisonous Mist",
		"type": "environmental",
		"description": "A thick, green mist that burns the lungs",
		"difficulty": 15,
		"savingThrow": "Constitution",
		"damage": "2d6 poison",
		"effect": "Poisoned condition for 1 hour on failed save",
	}
	data, _ := json.Marshal(hazard)
	return string(data)
}

func mockStoryArcResponse() string {
	arc := map[string]interface{}{
		"title": "The Shadow's Return",
		"description": "An ancient evil stirs in the depths",
		"acts": []map[string]interface{}{
			{
				"number": 1,
				"title": "Strange Omens",
				"description": "Mysterious events plague the land",
			},
		},
		"themes": []string{"corruption", "redemption"},
		"majorNPCs": []string{"The Shadow Lord", "The Oracle"},
	}
	data, _ := json.Marshal(arc)
	return string(data)
}

func mockSettlementResponse() string {
	settlement := map[string]interface{}{
		"name": "Riverside Haven",
		"type": "town",
		"population": 2500,
		"description": "A peaceful town by the river",
		"leadership": "Mayor Aldric",
		"notableLocations": []string{"The Rusty Anchor Inn", "Temple of Light"},
		"economy": "Trade and fishing",
		"currentEvents": []string{"Harvest festival approaching"},
	}
	data, _ := json.Marshal(settlement)
	return string(data)
}

func mockRecapResponse() string {
	recap := map[string]interface{}{
		"summary": "The party ventured into the ancient ruins...",
		"keyEvents": []string{
			"Defeated the stone guardian",
			"Discovered the hidden chamber",
			"Found the artifact",
		},
		"characterMoments": map[string]string{
			"Thorin": "Showed great bravery in battle",
			"Elara": "Solved the ancient puzzle",
		},
		"cliffhanger": "As you exit, you notice shadowy figures watching from afar...",
	}
	data, _ := json.Marshal(recap)
	return string(data)
}

// NewMockLLMProvider creates a new mock LLM provider
func NewMockLLMProvider() *MockLLMProvider {
	return &MockLLMProvider{}
}

// WithResponse adds a specific response for testing
func (m *MockLLMProvider) WithResponse(key, response string) *MockLLMProvider {
	if m.Responses == nil {
		m.Responses = make(map[string]string)
	}
	m.Responses[key] = response
	return m
}

// WithError returns an error for testing error cases
func (m *MockLLMProvider) WithError(err error) *MockLLMProvider {
	m.ResponseFunc = func(ctx context.Context, system, user string) (string, error) {
		return "", err
	}
	return m
}

// Reset clears all configured responses
func (m *MockLLMProvider) Reset() {
	m.ResponseFunc = nil
	m.Responses = nil
}