package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
)

type CharacterCreationHandler struct {
	characterService  *services.CharacterService
	characterBuilder  *services.CharacterBuilder
	aiCharService     *services.AICharacterService
	customRaceService *services.CustomRaceService
}

func NewCharacterCreationHandler(cs *services.CharacterService, crs *services.CustomRaceService, llmProvider services.LLMProvider) *CharacterCreationHandler {
	dataPath := filepath.Join(".", "data")
	return &CharacterCreationHandler{
		characterService:  cs,
		characterBuilder:  services.NewCharacterBuilder(dataPath),
		aiCharService:     services.NewAICharacterService(llmProvider),
		customRaceService: crs,
	}
}

// GetCharacterOptions returns available races, classes, backgrounds for character creation
func (h *CharacterCreationHandler) GetCharacterOptions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	options, err := h.characterBuilder.GetAvailableOptions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add user's custom races
	if h.customRaceService != nil {
		userUUID, err := uuid.Parse(userID)
		if err == nil {
			// Get user's own custom races
			userRaces, err := h.customRaceService.GetUserCustomRaces(r.Context(), userUUID)
			if err == nil {
				customRaceOptions := make([]map[string]interface{}, 0)
				for _, race := range userRaces {
					if race.ApprovalStatus == models.ApprovalStatusApproved || race.CreatedBy == userUUID {
						customRaceOptions = append(customRaceOptions, map[string]interface{}{
							"id":          race.ID,
							"name":        race.Name,
							"description": race.Description,
							"status":      race.ApprovalStatus,
							"isCustom":    true,
						})
					}
				}

				// Get public custom races
				publicRaces, err := h.customRaceService.GetPublicCustomRaces(r.Context())
				if err == nil {
					for _, race := range publicRaces {
						customRaceOptions = append(customRaceOptions, map[string]interface{}{
							"id":          race.ID,
							"name":        race.Name,
							"description": race.Description,
							"status":      race.ApprovalStatus,
							"isCustom":    true,
							"isPublic":    true,
						})
					}
				}

				options["customRaces"] = customRaceOptions
			}
		}
	}

	// Add AI availability status
	options["aiEnabled"] = h.aiCharService.IsEnabled()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(options); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateCharacter handles standard D&D character creation
func (h *CharacterCreationHandler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name               string         `json:"name"`
		Race               string         `json:"race"`
		CustomRaceID       string         `json:"customRaceId,omitempty"`
		Subrace            string         `json:"subrace,omitempty"`
		Class              string         `json:"class"`
		Background         string         `json:"background"`
		Alignment          string         `json:"alignment"`
		AbilityScoreMethod string         `json:"abilityScoreMethod"`
		AbilityScores      map[string]int `json:"abilityScores"`
		SelectedSkills     []string       `json:"selectedSkills,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Build character using character builder
	params := map[string]interface{}{
		"name":          req.Name,
		"race":          req.Race,
		"customRaceId":  req.CustomRaceID,
		"subrace":       req.Subrace,
		"class":         req.Class,
		"background":    req.Background,
		"alignment":     req.Alignment,
		"abilityScores": req.AbilityScores,
	}

	// If using a custom race, validate and get race data
	if req.CustomRaceID != "" {
		customRaceUUID, err := uuid.Parse(req.CustomRaceID)
		if err != nil {
			http.Error(w, "Invalid custom race ID", http.StatusBadRequest)
			return
		}

		userUUID, _ := uuid.Parse(userID)
		if err := h.customRaceService.ValidateCustomRaceForCharacter(r.Context(), customRaceUUID, userUUID); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// Get custom race stats and add to params
		raceStats, err := h.customRaceService.GetCustomRaceStats(r.Context(), customRaceUUID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		params["customRaceStats"] = raceStats

		// Increment usage counter
		go func() { _ = h.customRaceService.IncrementUsage(r.Context(), customRaceUUID) }()
	}

	character, err := h.characterBuilder.BuildCharacter(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set user ID and generate character ID
	character.ID = uuid.New().String()
	character.UserID = userID

	// Save character to database
	if err := h.characterService.CreateCharacter(r.Context(), character); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(character); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateCustomCharacter handles AI-assisted custom character creation
func (h *CharacterCreationHandler) CreateCustomCharacter(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req services.CustomCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.Concept == "" {
		http.Error(w, "Name and concept are required", http.StatusBadRequest)
		return
	}

	var character *models.Character
	var err error

	// Try AI generation first if enabled
	if h.aiCharService.IsEnabled() {
		character, err = h.aiCharService.GenerateCustomCharacter(&req)
		if err != nil {
			// Fall back to basic generation
			character, err = h.aiCharService.GenerateFallbackCharacter(&req)
		}
	} else {
		// Use fallback if AI is not enabled
		character, err = h.aiCharService.GenerateFallbackCharacter(&req)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate the custom character
	if err := h.aiCharService.ValidateCustomContent(character); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set user ID and generate character ID
	character.ID = uuid.New().String()
	character.UserID = userID

	// Save character to database
	if err := h.characterService.CreateCharacter(r.Context(), character); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(character); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// RollAbilityScores generates ability scores using specified method
func (h *CharacterCreationHandler) RollAbilityScores(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Method string `json:"method"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	scores, err := h.characterBuilder.RollAbilityScores(req.Method)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"scores": scores,
		"method": req.Method,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ValidateCharacter validates a character build
func (h *CharacterCreationHandler) ValidateCharacter(w http.ResponseWriter, r *http.Request) {
	var character models.Character
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Perform validation
	errors := h.validateCharacterBuild(&character)

	response := map[string]interface{}{
		"valid":  len(errors) == 0,
		"errors": errors,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *CharacterCreationHandler) validateCharacterBuild(character *models.Character) []string {
	var errors []string

	// Validate ability scores
	attrs := character.Attributes
	if attrs.Strength < 3 || attrs.Strength > 20 ||
		attrs.Dexterity < 3 || attrs.Dexterity > 20 ||
		attrs.Constitution < 3 || attrs.Constitution > 20 ||
		attrs.Intelligence < 3 || attrs.Intelligence > 20 ||
		attrs.Wisdom < 3 || attrs.Wisdom > 20 ||
		attrs.Charisma < 3 || attrs.Charisma > 20 {
		errors = append(errors, "Ability scores must be between 3 and 20")
	}

	// Validate level
	if character.Level < 1 || character.Level > 20 {
		errors = append(errors, "Level must be between 1 and 20")
	}

	// Validate required fields
	if character.Name == "" {
		errors = append(errors, "Character name is required")
	}
	if character.Race == "" {
		errors = append(errors, "Character race is required")
	}
	if character.Class == "" {
		errors = append(errors, "Character class is required")
	}

	return errors
}
