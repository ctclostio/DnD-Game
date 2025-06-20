package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/dice"
)

// Error messages
const (
	errFailedToRollDice = "Failed to roll dice"
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
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Calculate the modifier
	modifier := h.calculateTotalModifier(character, &req)
	req.Modifier = modifier

	// Perform the roll
	roll, allRolls, err := h.performDiceRoll(&req)
	if err != nil {
		http.Error(w, errFailedToRollDice, http.StatusInternalServerError)
		return
	}

	// Build response
	response := h.buildSkillCheckResponse(roll, modifier, allRolls, &req)

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// calculateTotalModifier calculates the total modifier for a skill check
func (h *Handlers) calculateTotalModifier(character *models.Character, req *SkillCheckRequest) int {
	if req.Modifier != 0 || req.Ability == "" {
		return req.Modifier
	}

	modifier := h.getAbilityModifier(character, req.Ability)

	// Add proficiency bonus if applicable
	if h.shouldAddProficiency(character, req) {
		modifier += character.ProficiencyBonus
	}

	return modifier
}

// shouldAddProficiency determines if proficiency bonus should be added
func (h *Handlers) shouldAddProficiency(character *models.Character, req *SkillCheckRequest) bool {
	if req.CheckType == "save" {
		return h.hasSavingThrowProficiency(character, req.Ability)
	}
	if req.CheckType == "skill" {
		return h.hasSkillProficiency(character, req.Skill)
	}
	return false
}

// performDiceRoll performs the dice roll based on advantage/disadvantage
func (h *Handlers) performDiceRoll(req *SkillCheckRequest) (int, []int, error) {
	roller := dice.NewRoller()

	if req.Advantage || req.Disadvantage {
		return h.rollWithAdvantageDisadvantage(roller, req.Advantage)
	}

	// Normal roll
	result, err := roller.Roll("1d20")
	if err != nil {
		return 0, nil, err
	}
	return result.Total, nil, nil
}

// rollWithAdvantageDisadvantage handles advantage/disadvantage rolls
func (h *Handlers) rollWithAdvantageDisadvantage(roller *dice.Roller, isAdvantage bool) (int, []int, error) {
	roll1, err := roller.Roll("1d20")
	if err != nil {
		return 0, nil, err
	}
	roll2, err := roller.Roll("1d20")
	if err != nil {
		return 0, nil, err
	}

	allRolls := []int{roll1.Total, roll2.Total}
	
	if isAdvantage {
		return max(roll1.Total, roll2.Total), allRolls, nil
	}
	return min(roll1.Total, roll2.Total), allRolls, nil
}

// buildSkillCheckResponse builds the response for a skill check
func (h *Handlers) buildSkillCheckResponse(roll, modifier int, allRolls []int, req *SkillCheckRequest) SkillCheckResponse {
	total := roll + modifier
	response := SkillCheckResponse{
		Roll:            roll,
		Modifier:        modifier,
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

	return response
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
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
	// Map of class to their saving throw proficiencies
	classProficiencies := map[string][]string{
		"fighter":   {constants.AbilityStrength, constants.AbilityConstitution},
		"barbarian": {constants.AbilityStrength, constants.AbilityConstitution},
		"rogue":     {constants.AbilityDexterity, constants.AbilityIntelligence},
		"ranger":    {constants.AbilityDexterity, constants.AbilityWisdom},
		"monk":      {constants.AbilityDexterity, constants.AbilityWisdom},
		"wizard":    {constants.AbilityIntelligence, constants.AbilityWisdom},
		"cleric":    {constants.AbilityWisdom, constants.AbilityCharisma},
		"druid":     {constants.AbilityIntelligence, constants.AbilityWisdom},
		"sorcerer":  {constants.AbilityConstitution, constants.AbilityCharisma},
		"warlock":   {constants.AbilityWisdom, constants.AbilityCharisma},
		"bard":      {constants.AbilityDexterity, constants.AbilityCharisma},
		"paladin":   {constants.AbilityWisdom, constants.AbilityCharisma},
	}

	proficiencies, exists := classProficiencies[character.Class]
	if !exists {
		return false
	}

	for _, prof := range proficiencies {
		if prof == ability {
			return true
		}
	}
	return false
}

func (h *Handlers) hasSkillProficiency(_ *models.Character, _ string) bool {
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
