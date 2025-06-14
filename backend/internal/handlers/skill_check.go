package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/dice"
)

type SkillCheckRequest struct {
	CharacterID  string `json:"characterId"`
	CheckType    string `json:"checkType"` // "skill", "save", "ability"
	Skill        string `json:"skill"`     // e.g., "athletics", "perception"
	Ability      string `json:"ability"`   // e.g., "strength", "dexterity"
	Modifier     int    `json:"modifier"`
	Advantage    bool   `json:"advantage"`
	Disadvantage bool   `json:"disadvantage"`
	DC           int    `json:"dc,omitempty"`
}

type SkillCheckResponse struct {
	Roll            int   `json:"roll"`
	Modifier        int   `json:"modifier"`
	Total           int   `json:"total"`
	Success         bool  `json:"success,omitempty"`
	CriticalSuccess bool  `json:"criticalSuccess"`
	CriticalFailure bool  `json:"criticalFailure"`
	Advantage       bool  `json:"advantage"`
	Disadvantage    bool  `json:"disadvantage"`
	AllRolls        []int `json:"allRolls,omitempty"`
}

// PerformSkillCheck handles skill checks and saving throws
func (h *Handlers) PerformSkillCheck(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var req SkillCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify character ownership
	character, err := h.characterService.GetCharacterByID(r.Context(), req.CharacterID)
	if err != nil {
		http.Error(w, "Character not found", http.StatusNotFound)
		return
	}

	if character.UserID != userID {
		// Only the character owner can perform skill checks on their character
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Calculate the ability modifier if not provided
	if req.Modifier == 0 && req.Ability != "" {
		req.Modifier = h.getAbilityModifier(character, req.Ability)

		// Add proficiency bonus for saving throws if applicable
		if req.CheckType == "save" && h.hasSavingThrowProficiency(character, req.Ability) {
			req.Modifier += character.ProficiencyBonus
		}

		// Add proficiency bonus for skill checks if applicable
		if req.CheckType == "skill" && h.hasSkillProficiency(character, req.Skill) {
			req.Modifier += character.ProficiencyBonus
		}
	}

	// Perform the roll
	roller := dice.NewRoller()
	var roll, total int
	var allRolls []int

	if req.Advantage || req.Disadvantage {
		// Roll twice
		roll1, err := roller.Roll("1d20")
		if err != nil {
			http.Error(w, "Failed to roll dice", http.StatusInternalServerError)
			return
		}
		roll2, err := roller.Roll("1d20")
		if err != nil {
			http.Error(w, "Failed to roll dice", http.StatusInternalServerError)
			return
		}
		allRolls = []int{roll1.Total, roll2.Total}

		if req.Advantage {
			if roll1.Total > roll2.Total {
				roll = roll1.Total
			} else {
				roll = roll2.Total
			}
		} else { // Disadvantage
			if roll1.Total < roll2.Total {
				roll = roll1.Total
			} else {
				roll = roll2.Total
			}
		}
	} else {
		// Normal roll
		result, err := roller.Roll("1d20")
		if err != nil {
			http.Error(w, "Failed to roll dice", http.StatusInternalServerError)
			return
		}
		roll = result.Total
	}

	total = roll + req.Modifier

	response := SkillCheckResponse{
		Roll:            roll,
		Modifier:        req.Modifier,
		Total:           total,
		CriticalSuccess: roll == 20,
		CriticalFailure: roll == 1,
		Advantage:       req.Advantage,
		Disadvantage:    req.Disadvantage,
		AllRolls:        allRolls,
	}

	// Check against DC if provided
	if req.DC > 0 {
		response.Success = total >= req.DC
	}

	// Log the roll in the game session if applicable
	// Note: Since characters don't have a direct GameSessionID field,
	// we would need to query the game_participants table to find active sessions
	// For now, we'll skip the websocket broadcast
	// TODO: Implement session lookup if needed

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetCharacterChecks returns available checks for a character
func (h *Handlers) GetCharacterChecks(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	characterID := mux.Vars(r)["id"]

	character, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		http.Error(w, "Character not found", http.StatusNotFound)
		return
	}

	if character.UserID != userID {
		// For now, only the owner can view character checks
		// TODO: Implement DM permission check through game_participants table
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Build response with all available checks
	response := map[string]interface{}{
		"savingThrows": h.getSavingThrows(character),
		"skills":       h.getSkills(character),
		"abilities":    h.getAbilityChecks(character),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) getAbilityModifier(character *models.Character, ability string) int {
	switch ability {
	case constants.AbilityStrength:
		return (character.Attributes.Strength - 10) / 2
	case constants.AbilityDexterity:
		return (character.Attributes.Dexterity - 10) / 2
	case constants.AbilityConstitution:
		return (character.Attributes.Constitution - 10) / 2
	case constants.AbilityIntelligence:
		return (character.Attributes.Intelligence - 10) / 2
	case constants.AbilityWisdom:
		return (character.Attributes.Wisdom - 10) / 2
	case constants.AbilityCharisma:
		return (character.Attributes.Charisma - 10) / 2
	default:
		return 0
	}
}

func (h *Handlers) hasSavingThrowProficiency(character *models.Character, ability string) bool {
	// This would check character's class saving throw proficiencies
	// For now, returning based on common class proficiencies
	switch character.Class {
	case "fighter", "barbarian":
		return ability == constants.AbilityStrength || ability == constants.AbilityConstitution
	case "rogue", "ranger", "monk":
		return ability == constants.AbilityDexterity || ability == constants.AbilityIntelligence
	case "wizard":
		return ability == constants.AbilityIntelligence || ability == constants.AbilityWisdom
	case "cleric", "druid":
		return ability == constants.AbilityWisdom || ability == constants.AbilityCharisma
	case "sorcerer", "warlock", "bard":
		return ability == constants.AbilityCharisma || ability == constants.AbilityWisdom
	case "paladin":
		return ability == constants.AbilityWisdom || ability == constants.AbilityCharisma
	default:
		return false
	}
}

func (h *Handlers) hasSkillProficiency(character *models.Character, skill string) bool {
	// This would check character's skill proficiencies from background/class
	// For now, returning true for some common proficiencies
	// In a real implementation, this would check character.Skills array
	return false // Would need to implement skill proficiency tracking
}

func (h *Handlers) getSavingThrows(character *models.Character) []map[string]interface{} {
	abilities := []string{constants.AbilityStrength, constants.AbilityDexterity, constants.AbilityConstitution, constants.AbilityIntelligence, constants.AbilityWisdom, constants.AbilityCharisma}
	saves := make([]map[string]interface{}, 0)

	for _, ability := range abilities {
		modifier := h.getAbilityModifier(character, ability)
		isProficient := h.hasSavingThrowProficiency(character, ability)
		if isProficient {
			modifier += character.ProficiencyBonus
		}

		saves = append(saves, map[string]interface{}{
			"name":       ability,
			"modifier":   modifier,
			"proficient": isProficient,
		})
	}

	return saves
}

func (h *Handlers) getSkills(character *models.Character) []map[string]interface{} {
	// D&D 5e skills mapped to their abilities
	skills := []struct {
		name    string
		ability string
	}{
		{"acrobatics", constants.AbilityDexterity},
		{"animal handling", constants.AbilityWisdom},
		{"arcana", constants.AbilityIntelligence},
		{"athletics", constants.AbilityStrength},
		{"deception", constants.AbilityCharisma},
		{"history", constants.AbilityIntelligence},
		{"insight", constants.AbilityWisdom},
		{"intimidation", constants.AbilityCharisma},
		{"investigation", constants.AbilityIntelligence},
		{"medicine", constants.AbilityWisdom},
		{"nature", constants.AbilityIntelligence},
		{"perception", constants.AbilityWisdom},
		{"performance", constants.AbilityCharisma},
		{"persuasion", constants.AbilityCharisma},
		{"religion", constants.AbilityIntelligence},
		{"sleight of hand", constants.AbilityDexterity},
		{"stealth", constants.AbilityDexterity},
		{"survival", constants.AbilityWisdom},
	}

	skillList := make([]map[string]interface{}, 0)
	for _, skill := range skills {
		modifier := h.getAbilityModifier(character, skill.ability)
		// Would check proficiency here

		skillList = append(skillList, map[string]interface{}{
			"name":       skill.name,
			"ability":    skill.ability,
			"modifier":   modifier,
			"proficient": false, // Would check actual proficiencies
		})
	}

	return skillList
}

func (h *Handlers) getAbilityChecks(character *models.Character) []map[string]interface{} {
	abilities := []string{constants.AbilityStrength, constants.AbilityDexterity, constants.AbilityConstitution, constants.AbilityIntelligence, constants.AbilityWisdom, constants.AbilityCharisma}
	checks := make([]map[string]interface{}, 0)

	for _, ability := range abilities {
		modifier := h.getAbilityModifier(character, ability)
		checks = append(checks, map[string]interface{}{
			"name":     ability,
			"modifier": modifier,
		})
	}

	return checks
}
