package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

type AICharacterService struct {
	llmProvider LLMProvider
	aiEnabled   bool
}

type CustomCharacterRequest struct {
	Name       string `json:"name"`
	Concept    string `json:"concept"`
	Race       string `json:"race,omitempty"`
	Class      string `json:"class,omitempty"`
	Background string `json:"background,omitempty"`
	Ruleset    string `json:"ruleset,omitempty"`
	Level      int    `json:"level,omitempty"`
}

func NewAICharacterService(llmProvider LLMProvider) *AICharacterService {
	return &AICharacterService{
		llmProvider: llmProvider,
		aiEnabled:   llmProvider != nil,
	}
}

func (s *AICharacterService) IsEnabled() bool {
	return s.aiEnabled && s.llmProvider != nil
}

func (s *AICharacterService) GenerateCustomCharacter(req CustomCharacterRequest) (*models.Character, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("AI character generation is not enabled")
	}

	// Build the prompt
	prompt := s.buildCharacterPrompt(req)
	systemPrompt := "You are a D&D character creation assistant. Create balanced, interesting characters that follow game rules. Your response must be valid JSON matching the specified format."

	// Call AI API
	response, err := s.llmProvider.GenerateCompletion(context.Background(), prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate character: %w", err)
	}

	// Parse AI response into character
	character, err := s.parseAIResponse(response, req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return character, nil
}

func (s *AICharacterService) ValidateCustomContent(character *models.Character) error {
	if !s.IsEnabled() {
		// Skip validation if AI is not enabled
		return nil
	}

	prompt := fmt.Sprintf(`Validate this D&D character for game balance:
Name: %s
Race: %s
Class: %s
Level: %d
Stats: STR %d, DEX %d, CON %d, INT %d, WIS %d, CHA %d

Is this character balanced for play? If not, suggest adjustments. Respond with JSON:
{
  "balanced": true/false,
  "issues": ["list of issues"],
  "suggestions": ["list of suggestions"]
}`, character.Name, character.Race, character.Class, character.Level,
		character.Attributes.Strength, character.Attributes.Dexterity,
		character.Attributes.Constitution, character.Attributes.Intelligence,
		character.Attributes.Wisdom, character.Attributes.Charisma)

	systemPrompt := "You are a D&D character validator. Analyze the character for game balance and provide feedback. Your response must be valid JSON."

	response, err := s.llmProvider.GenerateCompletion(context.Background(), prompt, systemPrompt)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Parse validation response
	var validation struct {
		Balanced    bool     `json:"balanced"`
		Issues      []string `json:"issues"`
		Suggestions []string `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(response), &validation); err != nil {
		// If parsing fails, assume it's balanced
		return nil
	}

	if !validation.Balanced && len(validation.Issues) > 0 {
		return fmt.Errorf("character balance issues: %s", strings.Join(validation.Issues, "; "))
	}

	return nil
}

func (s *AICharacterService) buildCharacterPrompt(req CustomCharacterRequest) string {
	ruleset := req.Ruleset
	if ruleset == "" {
		ruleset = "D&D 5e"
	}

	level := req.Level
	if level == 0 {
		level = 1
	}

	prompt := fmt.Sprintf(`Create a %s character with the following details:
Name: %s
Concept: %s`, ruleset, req.Name, req.Concept)

	if req.Race != "" {
		prompt += fmt.Sprintf("\nRace: %s (or similar if custom)", req.Race)
	}
	if req.Class != "" {
		prompt += fmt.Sprintf("\nClass: %s (or similar if custom)", req.Class)
	}
	if req.Background != "" {
		prompt += fmt.Sprintf("\nBackground: %s", req.Background)
	}

	prompt += fmt.Sprintf("\nLevel: %d", level)

	prompt += `

Generate a complete character with:
1. Ability scores (using standard array or point buy)
2. Racial traits and abilities
3. Class features appropriate for the level
4. Background skills and proficiencies
5. Equipment and starting gear
6. Personality traits, ideals, bonds, and flaws
7. Any unique abilities that fit the concept

Respond with a JSON object in this format:
{
  "race": "race name",
  "subrace": "subrace if applicable",
  "class": "class name",
  "subclass": "subclass if applicable",
  "background": "background name",
  "alignment": "alignment",
  "attributes": {
    "strength": 10,
    "dexterity": 10,
    "constitution": 10,
    "intelligence": 10,
    "wisdom": 10,
    "charisma": 10
  },
  "hitDice": "" + constants.DiceD8 + "",
  "speed": 30,
  "features": [
    {
      "name": "Feature Name",
      "description": "Feature description",
      "level": 1,
      "source": "Race/Class/Background"
    }
  ],
  "proficiencies": {
    "armor": ["list of armor"],
    "weapons": ["list of weapons"],
    "tools": ["list of tools"],
    "languages": ["list of languages"]
  },
  "skills": ["list of skill proficiencies"],
  "equipment": [
    {
      "name": "Item name",
      "type": "weapon/armor/gear",
      "description": "Item description"
    }
  ],
  "personality": {
    "traits": ["trait 1", "trait 2"],
    "ideals": ["ideal"],
    "bonds": ["bond"],
    "flaws": ["flaw"]
  },
  "backstory": "Brief character backstory"
}`

	return prompt
}

func (s *AICharacterService) parseAIResponse(aiResponse string, req CustomCharacterRequest) (*models.Character, error) {
	// Try to extract JSON from the response
	startIdx := strings.Index(aiResponse, "{")
	endIdx := strings.LastIndex(aiResponse, "}")

	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("no valid JSON found in AI response")
	}

	jsonStr := aiResponse[startIdx : endIdx+1]

	var aiChar struct {
		Race       string `json:"race"`
		Subrace    string `json:"subrace"`
		Class      string `json:"class"`
		Subclass   string `json:"subclass"`
		Background string `json:"background"`
		Alignment  string `json:"alignment"`
		Attributes struct {
			Strength     int `json:"strength"`
			Dexterity    int `json:"dexterity"`
			Constitution int `json:"constitution"`
			Intelligence int `json:"intelligence"`
			Wisdom       int `json:"wisdom"`
			Charisma     int `json:"charisma"`
		} `json:"attributes"`
		HitDice       string               `json:"hitDice"`
		Speed         int                  `json:"speed"`
		Features      []models.Feature     `json:"features"`
		Proficiencies models.Proficiencies `json:"proficiencies"`
		Skills        []string             `json:"skills"`
		Equipment     []models.Item        `json:"equipment"`
		Personality   map[string][]string  `json:"personality"`
		Backstory     string               `json:"backstory"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &aiChar); err != nil {
		return nil, err
	}

	// Build the character model
	character := &models.Character{
		Name:          req.Name,
		Race:          aiChar.Race,
		Subrace:       aiChar.Subrace,
		Class:         aiChar.Class,
		Subclass:      aiChar.Subclass,
		Background:    aiChar.Background,
		Alignment:     aiChar.Alignment,
		Level:         req.Level,
		HitDice:       aiChar.HitDice,
		Speed:         aiChar.Speed,
		Attributes:    models.Attributes(aiChar.Attributes),
		Features:      aiChar.Features,
		Proficiencies: aiChar.Proficiencies,
		Equipment:     aiChar.Equipment,
	}

	if character.Level == 0 {
		character.Level = 1
	}

	// Calculate derived stats
	character.ProficiencyBonus = ((character.Level - 1) / 4) + 2
	character.Initiative = s.calculateModifier(character.Attributes.Dexterity)

	// Calculate HP
	hitDiceValue := s.getHitDiceValue(aiChar.HitDice)
	character.MaxHitPoints = hitDiceValue + s.calculateModifier(character.Attributes.Constitution)
	character.HitPoints = character.MaxHitPoints

	// Calculate saving throws
	character.SavingThrows = s.calculateSavingThrows(character)

	// Store personality and backstory in resources
	character.Resources = map[string]interface{}{
		"personality": aiChar.Personality,
		"backstory":   aiChar.Backstory,
	}

	return character, nil
}

func (s *AICharacterService) calculateModifier(score int) int {
	return (score - 10) / 2
}

func (s *AICharacterService) getHitDiceValue(hitDice string) int {
	switch hitDice {
	case "1d6":
		return 6
	case constants.DiceD8:
		return 8
	case "1d10":
		return 10
	case constants.DiceD12:
		return 12
	default:
		return 8
	}
}

func (s *AICharacterService) calculateSavingThrows(character *models.Character) models.SavingThrows {
	// Basic saving throws without class proficiencies
	// In a full implementation, this would check class data
	return models.SavingThrows{
		Strength: models.SavingThrow{
			Modifier:    s.calculateModifier(character.Attributes.Strength),
			Proficiency: false,
		},
		Dexterity: models.SavingThrow{
			Modifier:    s.calculateModifier(character.Attributes.Dexterity),
			Proficiency: false,
		},
		Constitution: models.SavingThrow{
			Modifier:    s.calculateModifier(character.Attributes.Constitution),
			Proficiency: false,
		},
		Intelligence: models.SavingThrow{
			Modifier:    s.calculateModifier(character.Attributes.Intelligence),
			Proficiency: false,
		},
		Wisdom: models.SavingThrow{
			Modifier:    s.calculateModifier(character.Attributes.Wisdom),
			Proficiency: false,
		},
		Charisma: models.SavingThrow{
			Modifier:    s.calculateModifier(character.Attributes.Charisma),
			Proficiency: false,
		},
	}
}

// GenerateFallbackCharacter creates a character without AI when the service is disabled
func (s *AICharacterService) GenerateFallbackCharacter(req CustomCharacterRequest) (*models.Character, error) {
	// Create a basic character based on the concept
	character := &models.Character{
		Name:       req.Name,
		Race:       "Custom",
		Class:      "Adventurer",
		Background: "Wanderer",
		Alignment:  "True Neutral",
		Level:      1,
		HitDice:    constants.DiceD8,
		Speed:      30,
		Attributes: models.Attributes{
			Strength:     12,
			Dexterity:    14,
			Constitution: 13,
			Intelligence: 10,
			Wisdom:       12,
			Charisma:     10,
		},
		Proficiencies: models.Proficiencies{
			Armor:     []string{"Light armor"},
			Weapons:   []string{"Simple weapons"},
			Tools:     []string{},
			Languages: []string{"Common"},
		},
		Features: []models.Feature{
			{
				Name:        "Adaptable",
				Description: "Your custom heritage grants you versatility in your abilities.",
				Level:       1,
				Source:      "Custom Race",
			},
		},
	}

	// Override with provided values
	if req.Race != "" {
		character.Race = req.Race
	}
	if req.Class != "" {
		character.Class = req.Class
	}
	if req.Background != "" {
		character.Background = req.Background
	}
	if req.Level > 0 {
		character.Level = req.Level
	}

	// Calculate derived stats
	character.ProficiencyBonus = ((character.Level - 1) / 4) + 2
	character.Initiative = s.calculateModifier(character.Attributes.Dexterity)
	character.MaxHitPoints = 8 + s.calculateModifier(character.Attributes.Constitution)
	character.HitPoints = character.MaxHitPoints
	character.SavingThrows = s.calculateSavingThrows(character)

	character.Resources = map[string]interface{}{
		"concept": req.Concept,
		"custom":  true,
	}

	return character, nil
}
