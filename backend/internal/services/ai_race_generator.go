package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// AIRaceGeneratorService handles AI-powered custom race generation.
type AIRaceGeneratorService struct {
	llmProvider LLMProvider
}

// NewAIRaceGeneratorService creates a new AI race generator service.
func NewAIRaceGeneratorService(llmProvider LLMProvider) *AIRaceGeneratorService {
	return &AIRaceGeneratorService{
		llmProvider: llmProvider,
	}
}

// GenerateCustomRace uses AI to generate a balanced custom race.
func (s *AIRaceGeneratorService) GenerateCustomRace(ctx context.Context, request models.CustomRaceRequest) (*models.CustomRaceGenerationResult, error) {
	systemPrompt := s.buildSystemPrompt()
	userPrompt := s.buildUserPrompt(request)

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate race: %w", err)
	}

	// Parse the JSON response.
	var result models.CustomRaceGenerationResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Validate and sanitize the result.
	if err := s.validateGeneratedRace(&result); err != nil {
		return nil, fmt.Errorf("generated race failed validation: %w", err)
	}

	return &result, nil
}

func (s *AIRaceGeneratorService) buildSystemPrompt() string {
	return `You are a Dungeons & Dragons 5th Edition expert game master helping create balanced custom player races.

Your task is to generate a complete, balanced custom race based on the player's description. Follow these guidelines:

1. BALANCE REQUIREMENTS:
   - Total ability score increases should sum to +3 or +4 (e.g., +2 to one ability, +1 to another)
   - Powerful traits should be balanced with limitations
   - Use existing D&D 5e races as reference for power level
   - Assign a balance score from 1-10 (5-6 is perfectly balanced, 7-8 is strong but acceptable, 9-10 needs revision)

2. REQUIRED ATTRIBUTES:
   - Name: A unique, thematic race name
   - Description: 2-3 sentences describing appearance and culture
   - Size: Usually Medium or Small
   - Speed: Usually 25-35 feet (25 for Small, 30 for Medium)
   - At least 2-3 unique racial traits
   - Languages: Common + 1-2 additional languages

3. OPTIONAL ATTRIBUTES (include if thematically appropriate):
   - Darkvision (usually 60 feet)
   - Damage resistances (limit to 1-2 types)
   - Skill proficiencies (limit to 1-2)
   - Tool/weapon proficiencies

4. TRAIT DESIGN:
   - Each trait should have a clear name and mechanical description
   - Avoid traits that are too complex or game-breaking
   - Reference existing D&D mechanics when possible

5. OUTPUT FORMAT:
   You MUST return ONLY valid JSON with no additional text or explanation. Return ONLY the JSON object matching this structure:
   {
     "name": "Race Name",
     "description": "Physical description and cultural background",
     "abilityScoreIncreases": {"strength": 2, "constitution": 1},
     "size": "Medium",
     "speed": 30,
     "traits": [
       {"name": "Trait Name", "description": "Mechanical description of what this trait does"}
     ],
     "languages": ["Common", "Draconic"],
     "darkvision": 60,
     "resistances": ["fire"],
     "immunities": [],
     "skillProficiencies": ["Intimidation"],
     "toolProficiencies": [],
     "weaponProficiencies": [],
     "armorProficiencies": [],
     "balanceScore": 6,
     "balanceExplanation": "This race is well-balanced because..."
   }

IMPORTANT: Ensure all arrays use valid D&D 5e terms. For ability scores, use lowercase keys: strength, dexterity, constitution, intelligence, wisdom, charisma.`
}

func (s *AIRaceGeneratorService) buildUserPrompt(request models.CustomRaceRequest) string {
	return fmt.Sprintf(`Create a D&D 5e custom race with the following specifications:

Name: %s
Description: %s

Generate a complete, balanced race following the system prompt guidelines. Ensure it's thematically consistent with the description while maintaining game balance.`,
		request.Name,
		request.Description)
}

func (s *AIRaceGeneratorService) validateGeneratedRace(race *models.CustomRaceGenerationResult) error {
	// Validate ability score increases.
	totalASI := 0
	validAbilities := map[string]bool{
		"strength": true, "dexterity": true, "constitution": true,
		"intelligence": true, "wisdom": true, "charisma": true,
	}

	for ability, increase := range race.AbilityScoreIncreases {
		if !validAbilities[strings.ToLower(ability)] {
			return fmt.Errorf("invalid ability score: %s", ability)
		}
		if increase < -2 || increase > 3 {
			return fmt.Errorf("ability score increase out of range: %s +%d", ability, increase)
		}
		totalASI += increase
	}

	if totalASI < 1 || totalASI > 6 {
		return fmt.Errorf("total ability score increases (%d) outside acceptable range (1-6)", totalASI)
	}

	// Validate size.
	validSize := false
	for _, size := range models.ValidSizes {
		if race.Size == size {
			validSize = true
			break
		}
	}
	if !validSize {
		return fmt.Errorf("invalid size: %s", race.Size)
	}

	// Validate speed.
	if race.Speed < 20 || race.Speed > 40 {
		return fmt.Errorf("speed %d outside acceptable range (20-40)", race.Speed)
	}

	// Validate traits.
	if len(race.Traits) < 1 {
		return fmt.Errorf("race must have at least one trait")
	}
	if len(race.Traits) > 6 {
		return fmt.Errorf("too many traits (%d), maximum is 6", len(race.Traits))
	}

	// Validate languages.
	if len(race.Languages) < 1 {
		return fmt.Errorf("race must know at least one language")
	}
	if len(race.Languages) > 4 {
		return fmt.Errorf("too many languages (%d), maximum is 4", len(race.Languages))
	}

	// Validate darkvision.
	if race.Darkvision != 0 && race.Darkvision != 30 && race.Darkvision != 60 && race.Darkvision != 120 {
		return fmt.Errorf("invalid darkvision range: %d", race.Darkvision)
	}

	// Validate damage types for resistances/immunities.
	for _, resistance := range race.Resistances {
		if !isValidDamageType(resistance) {
			return fmt.Errorf("invalid damage resistance type: %s", resistance)
		}
	}

	for _, immunity := range race.Immunities {
		if !isValidDamageType(immunity) {
			return fmt.Errorf("invalid damage immunity type: %s", immunity)
		}
	}

	// Validate balance score.
	if race.BalanceScore < 1 || race.BalanceScore > 10 {
		return fmt.Errorf("balance score %d outside range (1-10)", race.BalanceScore)
	}

	return nil
}

func isValidDamageType(damageType string) bool {
	normalized := strings.ToLower(damageType)
	for _, valid := range models.ValidDamageTypes {
		if normalized == valid {
			return true
		}
	}
	return false
}
