package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

type CustomClassRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Role        string `json:"role"`               // e.g., "tank", "healer", "damage dealer", "support"
	Style       string `json:"style"`              // "balanced", "flavorful", "powerful"
	Features    string `json:"features,omitempty"` // Optional desired features
}

type AIClassGenerator struct {
	llmProvider LLMProvider
}

func NewAIClassGenerator(provider LLMProvider) *AIClassGenerator {
	return &AIClassGenerator{
		llmProvider: provider,
	}
}

func (g *AIClassGenerator) GenerateCustomClass(ctx context.Context, req CustomClassRequest) (*models.CustomClass, error) {
	prompt := g.buildClassPrompt(req)

	systemPrompt := `You are a D&D 5th Edition expert game designer creating balanced, interesting custom classes.
Your responses must be valid JSON matching the specified format exactly. Do not include any additional text or explanation outside the JSON.`

	response, err := g.llmProvider.GenerateCompletion(ctx, prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate class: %w", err)
	}

	class, err := g.parseClassResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse class response: %w", err)
	}

	// Validate and balance the class.
	if err := g.validateClass(class); err != nil {
		return nil, fmt.Errorf("class validation failed: %w", err)
	}

	// Calculate balance score.
	class.BalanceScore = g.calculateBalanceScore(class)
	class.PowerLevel = req.Style

	return class, nil
}

func (g *AIClassGenerator) buildClassPrompt(req CustomClassRequest) string {
	styleGuide := map[string]string{
		"balanced":  "Create a class that is well-balanced with existing D&D 5e classes, neither too powerful nor too weak.",
		"flavorful": "Focus on unique and interesting mechanics that enhance the roleplay experience, even if slightly unconventional.",
		"powerful":  "Create a slightly stronger class suitable for experienced players or high-difficulty campaigns.",
	}

	return fmt.Sprintf(`Create a custom class based on the following request:

Name: %s
Description: %s
Role: %s
Style: %s
Desired Features: %s

%s

Create a complete D&D 5e class following these guidelines:
1. Hit Die: Choose d6, d8, d10, or d12 based on the class's durability
2. Primary Ability: The main ability score this class relies on
3. Saving Throws: Two saving throw proficiencies (usually one strong, one weak)
4. Skills: List of available skills and how many the player can choose
5. Proficiencies: Armor, weapons, and tools
6. Starting Equipment: Basic gear for level 1
7. Class Features: Unique abilities gained at different levels (at least for levels 1-5)
8. Subclass: Name and level when subclasses are chosen (usually level 2 or 3)
9. Spellcasting (if applicable): Spell list type, spells known/prepared, spell slots

Ensure the class:
- Has a clear identity and role in a party
- Provides meaningful choices at each level
- Is balanced with existing D&D 5e classes
- Has interesting and fun mechanics
- Includes both combat and non-combat features

Respond with a JSON object in exactly this format:
{
  "name": "Class Name",
  "description": "Full description of the class",
  "hitDie": 8,
  "primaryAbility": "Intelligence",
  "savingThrowProficiencies": ["Intelligence", "Wisdom"],
  "skillProficiencies": ["Arcana", "History", "Insight", "Investigation", "Medicine", "Religion"],
  "skillChoices": 2,
  "armorProficiencies": ["Light armor"],
  "weaponProficiencies": ["Simple weapons"],
  "toolProficiencies": ["Herbalism kit"],
  "startingEquipment": "A spellbook, a component pouch, a scholar's pack, and leather armor",
  "classFeatures": [
    {
      "level": 1,
      "name": "Feature Name",
      "description": "Feature description",
      "usesPerRest": "1 + proficiency bonus",
      "restType": "long",
      "passive": false
    }
  ],
  "subclassName": "Archetype",
  "subclassLevel": 3,
  "subclasses": [
    {
      "name": "Subclass Name",
      "description": "Subclass description",
      "features": [
        {
          "level": 3,
          "name": "Subclass Feature",
          "description": "Feature description"
        }
      ]
    }
  ],
  "spellcastingAbility": "Intelligence",
  "spellList": ["wizard"],
  "cantripsKnownProgression": [2, 2, 2, 3, 3, 3, 3, 3, 3, 4],
  "spellsKnownProgression": [2, 3, 4, 5, 6, 7, 8, 9, 10, 11],
  "ritualCasting": true,
  "spellcastingFocus": "Arcane focus or component pouch",
  "dmNotes": "Any balance considerations or usage notes for the DM"
}

For non-spellcasters, omit the spellcasting fields. Ensure all features are balanced and appropriate for D&D 5e.`,
		req.Name,
		req.Description,
		req.Role,
		req.Style,
		req.Features,
		styleGuide[req.Style],
	)
}

func (g *AIClassGenerator) parseClassResponse(response string) (*models.CustomClass, error) {
	// Clean up the response to extract JSON.
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var classData struct {
		Name                     string                `json:"name"`
		Description              string                `json:"description"`
		HitDie                   int                   `json:"hitDie"`
		PrimaryAbility           string                `json:"primaryAbility"`
		SavingThrowProficiencies []string              `json:"savingThrowProficiencies"`
		SkillProficiencies       []string              `json:"skillProficiencies"`
		SkillChoices             int                   `json:"skillChoices"`
		ArmorProficiencies       []string              `json:"armorProficiencies"`
		WeaponProficiencies      []string              `json:"weaponProficiencies"`
		ToolProficiencies        []string              `json:"toolProficiencies"`
		StartingEquipment        string                `json:"startingEquipment"`
		ClassFeatures            []models.ClassFeature `json:"classFeatures"`
		SubclassName             string                `json:"subclassName"`
		SubclassLevel            int                   `json:"subclassLevel"`
		Subclasses               []models.Subclass     `json:"subclasses"`
		SpellcastingAbility      string                `json:"spellcastingAbility,omitempty"`
		SpellList                []string              `json:"spellList,omitempty"`
		CantripsKnownProgression []int                 `json:"cantripsKnownProgression,omitempty"`
		SpellsKnownProgression   []int                 `json:"spellsKnownProgression,omitempty"`
		RitualCasting            bool                  `json:"ritualCasting"`
		SpellcastingFocus        string                `json:"spellcastingFocus,omitempty"`
		DMNotes                  string                `json:"dmNotes"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &classData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to models.CustomClass
	customClass := &models.CustomClass{
		Name:                     classData.Name,
		Description:              classData.Description,
		HitDie:                   classData.HitDie,
		PrimaryAbility:           classData.PrimaryAbility,
		SavingThrowProficiencies: classData.SavingThrowProficiencies,
		SkillProficiencies:       classData.SkillProficiencies,
		SkillChoices:             classData.SkillChoices,
		ArmorProficiencies:       classData.ArmorProficiencies,
		WeaponProficiencies:      classData.WeaponProficiencies,
		ToolProficiencies:        classData.ToolProficiencies,
		StartingEquipment:        classData.StartingEquipment,
		ClassFeatures:            classData.ClassFeatures,
		SubclassName:             classData.SubclassName,
		SubclassLevel:            classData.SubclassLevel,
		Subclasses:               classData.Subclasses,
		SpellcastingAbility:      classData.SpellcastingAbility,
		SpellList:                classData.SpellList,
		CantripsKnownProgression: classData.CantripsKnownProgression,
		SpellsKnownProgression:   classData.SpellsKnownProgression,
		RitualCasting:            classData.RitualCasting,
		SpellcastingFocus:        classData.SpellcastingFocus,
		DMNotes:                  classData.DMNotes,
	}

	// Generate spell slots progression for spellcasters.
	if customClass.SpellcastingAbility != "" {
		customClass.SpellSlotsProgression = g.generateSpellSlotProgression(customClass)
	}

	return customClass, nil
}

func (g *AIClassGenerator) generateSpellSlotProgression(class *models.CustomClass) map[string]interface{} {
	// Basic spell slot progression for full casters.
	// This can be customized based on the class type.
	fullCasterProgression := map[string]interface{}{
		"1":  []int{2, 0, 0, 0, 0, 0, 0, 0, 0},
		"2":  []int{3, 0, 0, 0, 0, 0, 0, 0, 0},
		"3":  []int{4, 2, 0, 0, 0, 0, 0, 0, 0},
		"4":  []int{4, 3, 0, 0, 0, 0, 0, 0, 0},
		"5":  []int{4, 3, 2, 0, 0, 0, 0, 0, 0},
		"6":  []int{4, 3, 3, 0, 0, 0, 0, 0, 0},
		"7":  []int{4, 3, 3, 1, 0, 0, 0, 0, 0},
		"8":  []int{4, 3, 3, 2, 0, 0, 0, 0, 0},
		"9":  []int{4, 3, 3, 3, 1, 0, 0, 0, 0},
		"10": []int{4, 3, 3, 3, 2, 0, 0, 0, 0},
		"11": []int{4, 3, 3, 3, 2, 1, 0, 0, 0},
		"12": []int{4, 3, 3, 3, 2, 1, 0, 0, 0},
		"13": []int{4, 3, 3, 3, 2, 1, 1, 0, 0},
		"14": []int{4, 3, 3, 3, 2, 1, 1, 0, 0},
		"15": []int{4, 3, 3, 3, 2, 1, 1, 1, 0},
		"16": []int{4, 3, 3, 3, 2, 1, 1, 1, 0},
		"17": []int{4, 3, 3, 3, 2, 1, 1, 1, 1},
		"18": []int{4, 3, 3, 3, 3, 1, 1, 1, 1},
		"19": []int{4, 3, 3, 3, 3, 2, 1, 1, 1},
		"20": []int{4, 3, 3, 3, 3, 2, 2, 1, 1},
	}

	return fullCasterProgression
}

func (g *AIClassGenerator) validateClass(class *models.CustomClass) error {
	// Validate hit die.
	validHitDice := map[int]bool{6: true, 8: true, 10: true, 12: true}
	if !validHitDice[class.HitDie] {
		return fmt.Errorf("invalid hit die: %d", class.HitDie)
	}

	// Validate primary ability.
	validAbilities := map[string]bool{
		"Strength": true, "Dexterity": true, "Constitution": true,
		"Intelligence": true, "Wisdom": true, "Charisma": true,
	}
	if !validAbilities[class.PrimaryAbility] {
		return fmt.Errorf("invalid primary ability: %s", class.PrimaryAbility)
	}

	// Validate saving throws (should have exactly 2).
	if len(class.SavingThrowProficiencies) != 2 {
		return fmt.Errorf("classes must have exactly 2 saving throw proficiencies")
	}

	// Validate skill choices.
	if class.SkillChoices < 2 || class.SkillChoices > 4 {
		class.SkillChoices = 2 // Default to 2 if out of range
	}

	// Ensure there are features for at least levels 1-3.
	hasLevel1Feature := false
	for _, feature := range class.ClassFeatures {
		if feature.Level == 1 {
			hasLevel1Feature = true
			break
		}
	}
	if !hasLevel1Feature {
		return fmt.Errorf("class must have at least one level 1 feature")
	}

	return nil
}

func (g *AIClassGenerator) calculateBalanceScore(class *models.CustomClass) int {
	score := 5 // Start with average score

	// Hit die scoring.
	hitDieScores := map[int]int{6: -2, 8: 0, 10: 1, 12: 2}
	score += hitDieScores[class.HitDie]

	// Armor proficiency scoring.
	if containsInClassGen(class.ArmorProficiencies, "Heavy armor") {
		score += 2
	} else if containsInClassGen(class.ArmorProficiencies, "Medium armor") {
		score += 1
	}

	// Spellcasting scoring.
	if class.SpellcastingAbility != "" {
		score += 2 // Spellcasters are generally more versatile
		if class.RitualCasting {
			score += 1
		}
	}

	// Feature count scoring.
	level5Features := 0
	for _, feature := range class.ClassFeatures {
		if feature.Level <= 5 {
			level5Features++
		}
	}
	if level5Features > 6 {
		score += 1 // Many early features
	} else if level5Features < 3 {
		score -= 1 // Few early features
	}

	// Skill choices scoring.
	if class.SkillChoices >= 4 {
		score += 1
	}

	// Cap the score between 1 and 10.
	if score < 1 {
		score = 1
	} else if score > 10 {
		score = 10
	}

	return score
}

func containsInClassGen(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
