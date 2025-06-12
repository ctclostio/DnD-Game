package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

func (h *Handlers) GetCharacters(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	characters, err := h.characterService.GetAllCharacters(r.Context(), userID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, characters)
}

func (h *Handlers) GetCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	character, err := h.characterService.GetCharacterByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, r, err.Error())
		return
	}

	// Verify the character belongs to the authenticated user
	if character.UserID != userID {
		response.Forbidden(w, r, "You don't have permission to access this character")
		return
	}

	response.JSON(w, r, http.StatusOK, character)
}

func (h *Handlers) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	var character models.Character
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}
	character.UserID = userID

	if err := h.characterService.CreateCharacter(r.Context(), &character); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, character)
}

func (h *Handlers) UpdateCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Verify ownership before update
	existing, err := h.characterService.GetCharacterByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, r, "Character not found")
		return
	}

	if existing.UserID != userID {
		response.Forbidden(w, r, "You don't have permission to update this character")
		return
	}

	var character models.Character
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	character.ID = id
	character.UserID = userID // Ensure user can't change ownership

	if err := h.characterService.UpdateCharacter(r.Context(), &character); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Fetch updated character
	updated, err := h.characterService.GetCharacterByID(r.Context(), id)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, updated)
}

func (h *Handlers) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Verify ownership before delete
	character, err := h.characterService.GetCharacterByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, r, "Character not found")
		return
	}

	if character.UserID != userID {
		response.Forbidden(w, r, "You don't have permission to delete this character")
		return
	}

	if err := h.characterService.DeleteCharacter(r.Context(), id); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusNoContent, nil)
}

// CastSpell handles spell casting requests
func (h *Handlers) CastSpell(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["id"]

	var req struct {
		SpellLevel int `json:"spellLevel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	err := h.characterService.UseSpellSlot(r.Context(), characterID, req.SpellLevel)
	if err != nil {
		response.BadRequest(w, r, err.Error())
		return
	}

	// Return updated character
	char, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, char)
}

// GenerateCustomClass handles AI generation of custom classes
func (h *Handlers) GenerateCustomClass(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Role        string `json:"role"`
		Style       string `json:"style"`
		Features    string `json:"features,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" || req.Description == "" {
		response.BadRequest(w, r, "Name and description are required")
		return
	}

	// Set defaults
	if req.Style == "" {
		req.Style = "balanced"
	}

	// Generate custom class
	customClass, err := h.characterService.GenerateCustomClass(r.Context(), userID, req.Name, req.Description, req.Role, req.Style, req.Features)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, customClass)
}

// GetCustomClasses returns the user's custom classes
func (h *Handlers) GetCustomClasses(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	includeUnapproved := r.URL.Query().Get("includeUnapproved") == "true"

	classes, err := h.characterService.GetUserCustomClasses(r.Context(), userID, includeUnapproved)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, classes)
}

// GetCustomClass returns a specific custom class
func (h *Handlers) GetCustomClass(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	classID := vars["id"]

	customClass, err := h.characterService.GetCustomClass(r.Context(), classID)
	if err != nil {
		response.NotFound(w, r, "Custom class not found")
		return
	}

	response.JSON(w, r, http.StatusOK, customClass)
}

// AddExperience handles adding XP to a character
func (h *Handlers) AddExperience(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["id"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Verify ownership
	character, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		response.NotFound(w, r, "Character not found")
		return
	}

	if character.UserID != userID {
		response.Forbidden(w, r, "You don't have permission to modify this character")
		return
	}

	var req struct {
		Experience int `json:"experience"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	if req.Experience <= 0 {
		response.BadRequest(w, r, "Experience must be positive")
		return
	}

	// Add experience
	err = h.characterService.AddExperience(r.Context(), characterID, req.Experience)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Return updated character
	char, err := h.characterService.GetCharacterByID(r.Context(), characterID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Calculate XP needed for next level
	xpForNext := h.characterService.GetXPForNextLevel(char.Level)

	resp := map[string]interface{}{
		"character":  char,
		"xpForNext":  xpForNext,
		"xpProgress": char.ExperiencePoints,
		"leveledUp":  false, // We could track this in the service
	}

	response.JSON(w, r, http.StatusOK, resp)
}

// Rest handles short and long rest requests
func (h *Handlers) Rest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["id"]

	var req struct {
		RestType string `json:"restType"` // "short" or "long"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	err := h.characterService.RestoreSpellSlots(r.Context(), characterID, req.RestType)
	if err != nil {
		response.BadRequest(w, r, err.Error())
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
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, char)
}
