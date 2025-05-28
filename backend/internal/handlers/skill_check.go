package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/pkg/dice"
)

type SkillCheckRequest struct {
	CharacterID string `json:"characterId"`
	CheckType   string `json:"checkType"` // "skill", "save", "ability"
	Skill       string `json:"skill"`     // e.g., "athletics", "perception"
	Ability     string `json:"ability"`   // e.g., "strength", "dexterity"
	Modifier    int    `json:"modifier"`
	Advantage   bool   `json:"advantage"`
	Disadvantage bool  `json:"disadvantage"`
	DC          int    `json:"dc,omitempty"`
}

type SkillCheckResponse struct {
	Roll      int    `json:"roll"`
	Modifier  int    `json:"modifier"`
	Total     int    `json:"total"`
	Success   bool   `json:"success,omitempty"`
	CriticalSuccess bool `json:"criticalSuccess"`
	CriticalFailure bool `json:"criticalFailure"`
	Advantage bool   `json:"advantage"`
	Disadvantage bool `json:"disadvantage"`
	AllRolls  []int  `json:"allRolls,omitempty"`
}

// PerformSkillCheck handles skill checks and saving throws
func (h *Handler) PerformSkillCheck(w http.ResponseWriter, r *http.Request) {
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
		// Check if user is DM of the game session
		if character.GameSessionID != "" {
			session, err := h.gameService.GetGameSession(r.Context(), character.GameSessionID)
			if err != nil || session.DmID != userID {
				http.Error(w, "Unauthorized", http.StatusForbidden)
				return
			}
		} else {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
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
		roll1 := roller.Roll("1d20")
		roll2 := roller.Roll("1d20")
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
		result := roller.Roll("1d20")
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
	if character.GameSessionID != "" {
		checkName := req.Skill
		if checkName == "" {
			checkName = req.Ability
		}
		if req.CheckType == "save" {
			checkName += " save"
		}
		
		h.websocketHub.Broadcast(character.GameSessionID, map[string]interface{}{
			"type": "skillCheck",
			"characterName": character.Name,
			"checkType": req.CheckType,
			"checkName": checkName,
			"roll": roll,
			"modifier": req.Modifier,
			"total": total,
			"success": response.Success,
			"dc": req.DC,
			"criticalSuccess": response.CriticalSuccess,
			"criticalFailure": response.CriticalFailure,
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCharacterChecks returns available checks for a character
func (h *Handler) GetCharacterChecks(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	characterID := mux.Vars(r)["id"]
	
	character, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		http.Error(w, "Character not found", http.StatusNotFound)
		return
	}
	
	if character.UserID != userID {
		// Check if user is DM
		if character.GameSessionID != "" {
			session, err := h.gameService.GetGameSession(r.Context(), character.GameSessionID)
			if err != nil || session.DmID != userID {
				http.Error(w, "Unauthorized", http.StatusForbidden)
				return
			}
		} else {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}
	
	// Build response with all available checks
	response := map[string]interface{}{
		"savingThrows": h.getSavingThrows(character),
		"skills": h.getSkills(character),
		"abilities": h.getAbilityChecks(character),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) getAbilityModifier(character *models.Character, ability string) int {
	switch ability {
	case "strength":
		return (character.Strength - 10) / 2
	case "dexterity":
		return (character.Dexterity - 10) / 2
	case "constitution":
		return (character.Constitution - 10) / 2
	case "intelligence":
		return (character.Intelligence - 10) / 2
	case "wisdom":
		return (character.Wisdom - 10) / 2
	case "charisma":
		return (character.Charisma - 10) / 2
	default:
		return 0
	}
}

func (h *Handler) hasSavingThrowProficiency(character *models.Character, ability string) bool {
	// This would check character's class saving throw proficiencies
	// For now, returning based on common class proficiencies
	switch character.Class {
	case "fighter", "barbarian":
		return ability == "strength" || ability == "constitution"
	case "rogue", "ranger", "monk":
		return ability == "dexterity" || ability == "intelligence"
	case "wizard":
		return ability == "intelligence" || ability == "wisdom"
	case "cleric", "druid":
		return ability == "wisdom" || ability == "charisma"
	case "sorcerer", "warlock", "bard":
		return ability == "charisma" || ability == "wisdom"
	case "paladin":
		return ability == "wisdom" || ability == "charisma"
	default:
		return false
	}
}

func (h *Handler) hasSkillProficiency(character *models.Character, skill string) bool {
	// This would check character's skill proficiencies from background/class
	// For now, returning true for some common proficiencies
	// In a real implementation, this would check character.Skills array
	return false // Would need to implement skill proficiency tracking
}

func (h *Handler) getSavingThrows(character *models.Character) []map[string]interface{} {
	abilities := []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"}
	saves := make([]map[string]interface{}, 0)
	
	for _, ability := range abilities {
		modifier := h.getAbilityModifier(character, ability)
		isProficient := h.hasSavingThrowProficiency(character, ability)
		if isProficient {
			modifier += character.ProficiencyBonus
		}
		
		saves = append(saves, map[string]interface{}{
			"name": ability,
			"modifier": modifier,
			"proficient": isProficient,
		})
	}
	
	return saves
}

func (h *Handler) getSkills(character *models.Character) []map[string]interface{} {
	// D&D 5e skills mapped to their abilities
	skills := []struct {
		name    string
		ability string
	}{
		{"acrobatics", "dexterity"},
		{"animal handling", "wisdom"},
		{"arcana", "intelligence"},
		{"athletics", "strength"},
		{"deception", "charisma"},
		{"history", "intelligence"},
		{"insight", "wisdom"},
		{"intimidation", "charisma"},
		{"investigation", "intelligence"},
		{"medicine", "wisdom"},
		{"nature", "intelligence"},
		{"perception", "wisdom"},
		{"performance", "charisma"},
		{"persuasion", "charisma"},
		{"religion", "intelligence"},
		{"sleight of hand", "dexterity"},
		{"stealth", "dexterity"},
		{"survival", "wisdom"},
	}
	
	skillList := make([]map[string]interface{}, 0)
	for _, skill := range skills {
		modifier := h.getAbilityModifier(character, skill.ability)
		// Would check proficiency here
		
		skillList = append(skillList, map[string]interface{}{
			"name": skill.name,
			"ability": skill.ability,
			"modifier": modifier,
			"proficient": false, // Would check actual proficiencies
		})
	}
	
	return skillList
}

func (h *Handler) getAbilityChecks(character *models.Character) []map[string]interface{} {
	abilities := []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"}
	checks := make([]map[string]interface{}, 0)
	
	for _, ability := range abilities {
		modifier := h.getAbilityModifier(character, ability)
		checks = append(checks, map[string]interface{}{
			"name": ability,
			"modifier": modifier,
		})
	}
	
	return checks
}