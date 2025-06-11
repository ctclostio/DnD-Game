package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/middleware"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/pkg/errors"
	"github.com/your-username/dnd-game/backend/pkg/response"
)

// CharacterHandlerV2 demonstrates the new standardized error handling
type CharacterHandlerV2 struct {
	characterService CharacterService
}

// NewCharacterHandlerV2 creates a new character handler with standardized error handling
func NewCharacterHandlerV2(characterService CharacterService) *CharacterHandlerV2 {
	return &CharacterHandlerV2{
		characterService: characterService,
	}
}

// GetCharacters retrieves all characters for the authenticated user
func (h *CharacterHandlerV2) GetCharacters(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	characters, err := h.characterService.GetAllCharacters(r.Context(), userID.String())
	if err != nil {
		return errors.NewInternalError("Failed to retrieve characters", err).
			WithCode(string(errors.ErrCodeDatabaseError))
	}

	response.JSON(w, r, http.StatusOK, map[string]interface{}{
		"characters": characters,
		"count":      len(characters),
	})
	return nil
}

// GetCharacter retrieves a specific character
func (h *CharacterHandlerV2) GetCharacter(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	vars := mux.Vars(r)
	characterID := vars["id"]

	// Validate character ID format
	if _, err := uuid.Parse(characterID); err != nil {
		return errors.NewValidationError("Invalid character ID format").
			WithCode(string(errors.ErrCodeInvalidInput)).
			WithDetails(map[string]interface{}{
				"character_id": characterID,
			})
	}

	character, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		// Check if it's a not found error
		if errors.IsNotFound(err) {
			return errors.NewNotFoundError("character").
				WithCode(string(errors.ErrCodeCharacterNotFound))
		}
		return errors.NewInternalError("Failed to retrieve character", err).
			WithCode(string(errors.ErrCodeDatabaseError))
	}

	// Verify ownership
	if character.UserID != userID.String() {
		return errors.NewAuthorizationError("You don't have permission to access this character").
			WithCode(string(errors.ErrCodeCharacterNotOwned)).
			WithDetails(map[string]interface{}{
				"character_id": characterID,
			})
	}

	response.JSON(w, r, http.StatusOK, character)
	return nil
}

// CreateCharacter creates a new character
func (h *CharacterHandlerV2) CreateCharacter(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	var req CreateCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.NewBadRequestError("Invalid request body").
			WithCode(string(errors.ErrCodeInvalidInput))
	}

	// Validate request
	if validationErr := h.validateCreateCharacterRequest(&req); validationErr != nil {
		return validationErr
	}

	// Check character limit
	characters, err := h.characterService.GetAllCharacters(r.Context(), userID.String())
	if err != nil {
		return errors.NewInternalError("Failed to check character limit", err).
			WithCode(string(errors.ErrCodeDatabaseError))
	}

	const maxCharacters = 10
	if len(characters) >= maxCharacters {
		return errors.NewBadRequestError("Character limit reached").
			WithCode(string(errors.ErrCodeCharacterLimitReached)).
			WithDetails(map[string]interface{}{
				"limit":   maxCharacters,
				"current": len(characters),
			})
	}

	// Create character
	character := &models.Character{
		UserID:     userID.String(),
		Name:       req.Name,
		Race:       req.Race,
		Class:      req.Class,
		Level:      req.Level,
		Background: req.Background,
		Alignment:  req.Alignment,
		Attributes: req.Attributes,
	}

	if err := h.characterService.CreateCharacter(r.Context(), character); err != nil {
		// Check for duplicate name
		if errors.IsDuplicate(err) {
			return errors.NewConflictError("A character with this name already exists").
				WithCode(string(errors.ErrCodeDuplicateEntry)).
				WithDetails(map[string]interface{}{
					"name": req.Name,
				})
		}
		return errors.NewInternalError("Failed to create character", err).
			WithCode(string(errors.ErrCodeDatabaseError))
	}

	response.JSON(w, r, http.StatusCreated, character)
	return nil
}

// UpdateCharacter updates an existing character
func (h *CharacterHandlerV2) UpdateCharacter(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	vars := mux.Vars(r)
	characterID := vars["id"]

	// Validate character ID
	if _, err := uuid.Parse(characterID); err != nil {
		return errors.NewValidationError("Invalid character ID format").
			WithCode(string(errors.ErrCodeInvalidInput))
	}

	// Verify ownership first
	existing, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		if errors.IsNotFound(err) {
			return errors.NewNotFoundError("character").
				WithCode(string(errors.ErrCodeCharacterNotFound))
		}
		return errors.NewInternalError("Failed to verify character ownership", err).
			WithCode(string(errors.ErrCodeDatabaseError))
	}

	if existing.UserID != userID.String() {
		return errors.NewAuthorizationError("You don't have permission to update this character").
			WithCode(string(errors.ErrCodeCharacterNotOwned))
	}

	// Parse update request
	var req UpdateCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.NewBadRequestError("Invalid request body").
			WithCode(string(errors.ErrCodeInvalidInput))
	}

	// Validate update request
	if validationErr := h.validateUpdateCharacterRequest(&req); validationErr != nil {
		return validationErr
	}

	// Apply updates
	character := &models.Character{
		ID:     characterID,
		UserID: userID.String(), // Ensure ownership can't be changed
	}

	// Only update provided fields
	if req.Name != nil {
		character.Name = *req.Name
	}
	if req.Level != nil {
		character.Level = *req.Level
	}
	if req.HitPoints != nil {
		character.HitPoints = *req.HitPoints
	}
	if req.ExperiencePoints != nil {
		character.ExperiencePoints = *req.ExperiencePoints
	}

	if err := h.characterService.UpdateCharacter(r.Context(), character); err != nil {
		return errors.NewInternalError("Failed to update character", err).
			WithCode(string(errors.ErrCodeDatabaseError))
	}

	// Fetch and return updated character
	updated, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		return errors.NewInternalError("Failed to retrieve updated character", err).
			WithCode(string(errors.ErrCodeDatabaseError))
	}

	response.JSON(w, r, http.StatusOK, updated)
	return nil
}

// DeleteCharacter deletes a character
func (h *CharacterHandlerV2) DeleteCharacter(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	vars := mux.Vars(r)
	characterID := vars["id"]

	// Validate character ID
	if _, err := uuid.Parse(characterID); err != nil {
		return errors.NewValidationError("Invalid character ID format").
			WithCode(string(errors.ErrCodeInvalidInput))
	}

	// Verify ownership
	character, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		if errors.IsNotFound(err) {
			return errors.NewNotFoundError("character").
				WithCode(string(errors.ErrCodeCharacterNotFound))
		}
		return errors.NewInternalError("Failed to verify character ownership", err).
			WithCode(string(errors.ErrCodeDatabaseError))
	}

	if character.UserID != userID.String() {
		return errors.NewAuthorizationError("You don't have permission to delete this character").
			WithCode(string(errors.ErrCodeCharacterNotOwned))
	}

	// TODO: Check if character is in active game session
	// This would require checking the game_participants table or adding a CurrentGameSessionID field to Character

	if err := h.characterService.DeleteCharacter(r.Context(), characterID); err != nil {
		return errors.NewInternalError("Failed to delete character", err).
			WithCode(string(errors.ErrCodeDatabaseError))
	}

	response.JSON(w, r, http.StatusOK, map[string]string{
		"message": "Character deleted successfully",
		"id":      characterID,
	})
	return nil
}

// LevelUp levels up a character
func (h *CharacterHandlerV2) LevelUp(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	vars := mux.Vars(r)
	characterID := vars["id"]

	// Validate character ID
	if _, err := uuid.Parse(characterID); err != nil {
		return errors.NewValidationError("Invalid character ID format").
			WithCode(string(errors.ErrCodeInvalidInput))
	}

	// Verify ownership
	character, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		if errors.IsNotFound(err) {
			return errors.NewNotFoundError("character").
				WithCode(string(errors.ErrCodeCharacterNotFound))
		}
		return errors.NewInternalError("Failed to retrieve character", err).
			WithCode(string(errors.ErrCodeDatabaseError))
	}

	if character.UserID != userID.String() {
		return errors.NewAuthorizationError("You don't have permission to level up this character").
			WithCode(string(errors.ErrCodeCharacterNotOwned))
	}

	// Check if character can level up
	if character.Level >= 20 {
		return errors.NewBadRequestError("Character is already at maximum level").
			WithCode(string(errors.ErrCodeOutOfRange)).
			WithDetails(map[string]interface{}{
				"current_level": character.Level,
				"max_level":     20,
			})
	}

	// Parse level up choices
	var req LevelUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.NewBadRequestError("Invalid request body").
			WithCode(string(errors.ErrCodeInvalidInput))
	}

	// Perform level up
	updatedCharacter, err := h.characterService.LevelUp(r.Context(), characterID, req.HitPointIncrease, req.AttributeIncrease)
	if err != nil {
		return errors.NewInternalError("Failed to level up character", err).
			WithCode(string(errors.ErrCodeDatabaseError))
	}

	response.JSON(w, r, http.StatusOK, updatedCharacter)
	return nil
}

// Validation helpers

func (h *CharacterHandlerV2) validateCreateCharacterRequest(req *CreateCharacterRequest) error {
	validationErrors := &errors.ValidationErrors{}

	if req.Name == "" {
		validationErrors.Add("name", "Name is required")
	} else if len(req.Name) > 50 {
		validationErrors.Add("name", "Name must be 50 characters or less")
	}

	if req.Race == "" {
		validationErrors.Add("race", "Race is required")
	}

	if req.Class == "" {
		validationErrors.Add("class", "Class is required")
	}

	if req.Level < 1 || req.Level > 20 {
		validationErrors.Add("level", "Level must be between 1 and 20")
	}

	// Validate attributes
	attrs := []struct {
		name  string
		value int
	}{
		{"strength", req.Attributes.Strength},
		{"dexterity", req.Attributes.Dexterity},
		{"constitution", req.Attributes.Constitution},
		{"intelligence", req.Attributes.Intelligence},
		{"wisdom", req.Attributes.Wisdom},
		{"charisma", req.Attributes.Charisma},
	}

	for _, attr := range attrs {
		if attr.value < 3 || attr.value > 20 {
			validationErrors.Add(attr.name, "Attribute must be between 3 and 20")
		}
	}

	if validationErrors.HasErrors() {
		return validationErrors.ToAppError().WithCode(string(errors.ErrCodeValidationFailed))
	}

	return nil
}

func (h *CharacterHandlerV2) validateUpdateCharacterRequest(req *UpdateCharacterRequest) error {
	validationErrors := &errors.ValidationErrors{}

	if req.Name != nil && *req.Name == "" {
		validationErrors.Add("name", "Name cannot be empty")
	}

	if req.Level != nil && (*req.Level < 1 || *req.Level > 20) {
		validationErrors.Add("level", "Level must be between 1 and 20")
	}

	if req.HitPoints != nil && *req.HitPoints < 0 {
		validationErrors.Add("hit_points", "Hit points cannot be negative")
	}

	if req.ExperiencePoints != nil && *req.ExperiencePoints < 0 {
		validationErrors.Add("experience_points", "Experience points cannot be negative")
	}

	if validationErrors.HasErrors() {
		return validationErrors.ToAppError().WithCode(string(errors.ErrCodeValidationFailed))
	}

	return nil
}

// RegisterRoutesV2 registers the character routes with new error handling
func (h *CharacterHandlerV2) RegisterRoutesV2(r *mux.Router) {
	// All routes use authenticated handlers
	r.HandleFunc("/characters", middleware.AuthenticatedHandler(h.GetCharacters)).Methods("GET")
	r.HandleFunc("/characters", middleware.AuthenticatedHandler(h.CreateCharacter)).Methods("POST")
	r.HandleFunc("/characters/{id}", middleware.AuthenticatedHandler(h.GetCharacter)).Methods("GET")
	r.HandleFunc("/characters/{id}", middleware.AuthenticatedHandler(h.UpdateCharacter)).Methods("PUT")
	r.HandleFunc("/characters/{id}", middleware.AuthenticatedHandler(h.DeleteCharacter)).Methods("DELETE")
	r.HandleFunc("/characters/{id}/level-up", middleware.AuthenticatedHandler(h.LevelUp)).Methods("POST")
}

// Request/Response types

type CreateCharacterRequest struct {
	Name       string            `json:"name"`
	Race       string            `json:"race"`
	Class      string            `json:"class"`
	Level      int               `json:"level"`
	Background string            `json:"background"`
	Alignment  string            `json:"alignment"`
	Attributes models.Attributes `json:"attributes"`
}

type UpdateCharacterRequest struct {
	Name             *string `json:"name,omitempty"`
	Level            *int    `json:"level,omitempty"`
	HitPoints        *int    `json:"hit_points,omitempty"`
	ExperiencePoints *int    `json:"experience_points,omitempty"`
}

type LevelUpRequest struct {
	HitPointIncrease  int    `json:"hit_point_increase"`
	AttributeIncrease string `json:"attribute_increase"` // e.g., "strength", "dexterity"
}
