package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
)

func (h *Handlers) GetCharacters(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	characters, err := h.characterService.GetAllCharacters(r.Context(), userID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	sendJSONResponse(w, http.StatusOK, characters)
}

func (h *Handlers) GetCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	character, err := h.characterService.GetCharacterByID(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}
	
	// Verify the character belongs to the authenticated user
	if character.UserID != userID {
		sendErrorResponse(w, http.StatusForbidden, "You don't have permission to access this character")
		return
	}
	
	sendJSONResponse(w, http.StatusOK, character)
}

func (h *Handlers) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	var character models.Character
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	character.UserID = userID
	
	if err := h.characterService.CreateCharacter(r.Context(), &character); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	sendJSONResponse(w, http.StatusCreated, character)
}

func (h *Handlers) UpdateCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	// Verify ownership before update
	existing, err := h.characterService.GetCharacterByID(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Character not found")
		return
	}
	
	if existing.UserID != userID {
		sendErrorResponse(w, http.StatusForbidden, "You don't have permission to update this character")
		return
	}
	
	var character models.Character
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	character.ID = id
	character.UserID = userID // Ensure user can't change ownership
	
	if err := h.characterService.UpdateCharacter(r.Context(), &character); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	// Fetch updated character
	updated, err := h.characterService.GetCharacterByID(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	sendJSONResponse(w, http.StatusOK, updated)
}

func (h *Handlers) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	// Verify ownership before delete
	character, err := h.characterService.GetCharacterByID(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Character not found")
		return
	}
	
	if character.UserID != userID {
		sendErrorResponse(w, http.StatusForbidden, "You don't have permission to delete this character")
		return
	}
	
	if err := h.characterService.DeleteCharacter(r.Context(), id); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	sendJSONResponse(w, http.StatusNoContent, nil)
}

// CastSpell handles spell casting requests
func (h *Handlers) CastSpell(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["id"]

	var req struct {
		SpellLevel int `json:"spellLevel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.characterService.UseSpellSlot(r.Context(), characterID, req.SpellLevel)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return updated character
	char, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve updated character")
		return
	}

	sendJSONResponse(w, http.StatusOK, char)
}

// AddExperience handles adding XP to a character
func (h *Handlers) AddExperience(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["id"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Verify ownership
	character, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Character not found")
		return
	}

	if character.UserID != userID {
		sendErrorResponse(w, http.StatusForbidden, "You don't have permission to modify this character")
		return
	}

	var req struct {
		Experience int `json:"experience"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Experience <= 0 {
		sendErrorResponse(w, http.StatusBadRequest, "Experience must be positive")
		return
	}

	// Add experience
	err = h.characterService.AddExperience(r.Context(), characterID, req.Experience)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return updated character
	char, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve updated character")
		return
	}

	// Calculate XP needed for next level
	xpForNext := h.characterService.GetXPForNextLevel(char.Level)
	
	response := map[string]interface{}{
		"character":    char,
		"xpForNext":    xpForNext,
		"xpProgress":   char.ExperiencePoints,
		"leveledUp":    false, // We could track this in the service
	}

	sendJSONResponse(w, http.StatusOK, response)
}

// Rest handles short and long rest requests
func (h *Handlers) Rest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["id"]

	var req struct {
		RestType string `json:"restType"` // "short" or "long"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.characterService.RestoreSpellSlots(r.Context(), characterID, req.RestType)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// For long rest, also restore hit points
	if req.RestType == "long" {
		char, err := h.characterService.GetCharacterByID(r.Context(), characterID)
		if err == nil {
			char.HitPoints = char.MaxHitPoints
			h.characterService.UpdateCharacter(r.Context(), char)
		}
	}

	// Return updated character
	char, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve updated character")
		return
	}

	sendJSONResponse(w, http.StatusOK, char)
}