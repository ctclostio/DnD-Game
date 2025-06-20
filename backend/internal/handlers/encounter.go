package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

// GenerateEncounter creates a new AI-generated encounter
func (h *Handlers) GenerateEncounter(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Get user role to check if they're a DM
	isDM := r.Context().Value("isDM").(bool)
	if !isDM {
		response.Forbidden(w, r, "Only DMs can create encounters")
		return
	}

	var req struct {
		GameSessionID    string   `json:"gameSessionId"`
		PartyLevel       int      `json:"partyLevel"`
		PartySize        int      `json:"partySize"`
		PartyComposition []string `json:"partyComposition"`
		Difficulty       string   `json:"difficulty"`
		EncounterType    string   `json:"encounterType"`
		Location         string   `json:"location"`
		NarrativeContext string   `json:"narrativeContext"`
		SpecialRequests  string   `json:"specialRequests,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, ErrInvalidRequestBody)
		return
	}

	// Validate required fields
	if req.GameSessionID == "" || req.PartyLevel < 1 || req.PartySize < 1 {
		response.BadRequest(w, r, "Missing required fields")
		return
	}

	// Set defaults
	if req.Difficulty == "" {
		req.Difficulty = "medium"
	}
	if req.EncounterType == "" {
		req.EncounterType = "combat"
	}

	// Create encounter request
	encounterReq := services.EncounterRequest{
		PartyLevel:       req.PartyLevel,
		PartySize:        req.PartySize,
		PartyComposition: req.PartyComposition,
		Difficulty:       req.Difficulty,
		EncounterType:    req.EncounterType,
		Location:         req.Location,
		NarrativeContext: req.NarrativeContext,
		SpecialRequests:  req.SpecialRequests,
	}

	// Generate encounter
	encounter, err := h.encounterService.GenerateEncounter(r.Context(), &encounterReq, req.GameSessionID, userID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, encounter)
}

// GetEncounter retrieves an encounter by ID
func (h *Handlers) GetEncounter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encounterID := vars["id"]

	encounter, err := h.encounterService.GetEncounter(r.Context(), encounterID)
	if err != nil {
		response.NotFound(w, r, "Encounter not found")
		return
	}

	response.JSON(w, r, http.StatusOK, encounter)
}

// GetSessionEncounters retrieves all encounters for a game session
func (h *Handlers) GetSessionEncounters(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	encounters, err := h.encounterService.GetEncountersBySession(r.Context(), sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, encounters)
}

// StartEncounter begins an encounter
func (h *Handlers) StartEncounter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encounterID := vars["id"]

	// Check if user is DM
	isDM := r.Context().Value("isDM").(bool)
	if !isDM {
		response.Forbidden(w, r, "Only DMs can start encounters")
		return
	}

	if err := h.encounterService.StartEncounter(r.Context(), encounterID); err != nil {
		response.BadRequest(w, r, err.Error())
		return
	}

	// Return updated encounter
	encounter, _ := h.encounterService.GetEncounter(r.Context(), encounterID)
	response.JSON(w, r, http.StatusOK, encounter)
}

// CompleteEncounter marks an encounter as completed
func (h *Handlers) CompleteEncounter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encounterID := vars["id"]

	// Check if user is DM
	isDM := r.Context().Value("isDM").(bool)
	if !isDM {
		response.Forbidden(w, r, "Only DMs can complete encounters")
		return
	}

	var req struct {
		Outcome string `json:"outcome"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, ErrInvalidRequestBody)
		return
	}

	if err := h.encounterService.CompleteEncounter(r.Context(), encounterID, req.Outcome); err != nil {
		response.BadRequest(w, r, err.Error())
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "completed"})
}

// ScaleEncounter adjusts encounter difficulty
func (h *Handlers) ScaleEncounter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encounterID := vars["id"]

	// Check if user is DM
	isDM := r.Context().Value("isDM").(bool)
	if !isDM {
		response.Forbidden(w, r, "Only DMs can scale encounters")
		return
	}

	var req struct {
		Difficulty string `json:"difficulty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, ErrInvalidRequestBody)
		return
	}

	encounter, err := h.encounterService.ScaleEncounter(r.Context(), encounterID, req.Difficulty)
	if err != nil {
		response.BadRequest(w, r, err.Error())
		return
	}

	response.JSON(w, r, http.StatusOK, encounter)
}

// GetTacticalSuggestion provides AI tactical advice
func (h *Handlers) GetTacticalSuggestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encounterID := vars["id"]

	// Check if user is DM
	isDM := r.Context().Value("isDM").(bool)
	if !isDM {
		response.Forbidden(w, r, "Only DMs can request tactical suggestions")
		return
	}

	var req struct {
		Situation string `json:"situation"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, ErrInvalidRequestBody)
		return
	}

	suggestion, err := h.encounterService.GetTacticalSuggestion(r.Context(), encounterID, req.Situation)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{
		"suggestion": suggestion,
	})
}

// LogEncounterEvent records an event during the encounter
func (h *Handlers) LogEncounterEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encounterID := vars["id"]

	var req struct {
		Round            int                    `json:"round"`
		EventType        string                 `json:"eventType"`
		ActorType        string                 `json:"actorType"`
		ActorID          string                 `json:"actorId,omitempty"`
		ActorName        string                 `json:"actorName"`
		Description      string                 `json:"description"`
		MechanicalEffect map[string]interface{} `json:"mechanicalEffect,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, ErrInvalidRequestBody)
		return
	}

	err := h.encounterService.LogCombatEvent(
		r.Context(),
		services.CombatEventParams{
			EncounterID:      encounterID,
			Round:            req.Round,
			EventType:        req.EventType,
			ActorType:        req.ActorType,
			ActorID:          req.ActorID,
			ActorName:        req.ActorName,
			Description:      req.Description,
			MechanicalEffect: req.MechanicalEffect,
		},
	)

	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "logged"})
}

// GetEncounterEvents retrieves encounter events
func (h *Handlers) GetEncounterEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encounterID := vars["id"]

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	events, err := h.encounterService.GetEncounterEvents(r.Context(), encounterID, limit)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, events)
}

// UpdateEnemyStatus updates an enemy during combat
func (h *Handlers) UpdateEnemyStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enemyID := vars["enemyId"]

	// Check if user is DM
	isDM := r.Context().Value("isDM").(bool)
	if !isDM {
		response.Forbidden(w, r, "Only DMs can update enemy status")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		response.BadRequest(w, r, ErrInvalidRequestBody)
		return
	}

	if err := h.encounterService.UpdateEnemyStatus(r.Context(), enemyID, updates); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

// TriggerReinforcements activates a reinforcement wave
func (h *Handlers) TriggerReinforcements(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encounterID := vars["id"]

	// Check if user is DM
	isDM := r.Context().Value("isDM").(bool)
	if !isDM {
		response.Forbidden(w, r, "Only DMs can trigger reinforcements")
		return
	}

	var req struct {
		WaveIndex int `json:"waveIndex"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, ErrInvalidRequestBody)
		return
	}

	if err := h.encounterService.TriggerReinforcements(r.Context(), encounterID, req.WaveIndex); err != nil {
		response.BadRequest(w, r, err.Error())
		return
	}

	// Return updated encounter
	encounter, _ := h.encounterService.GetEncounter(r.Context(), encounterID)
	response.JSON(w, r, http.StatusOK, encounter)
}

// CheckObjectives evaluates encounter objectives
func (h *Handlers) CheckObjectives(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encounterID := vars["id"]

	// Check if user is DM
	isDM := r.Context().Value("isDM").(bool)
	if !isDM {
		response.Forbidden(w, r, "Only DMs can check objectives")
		return
	}

	if err := h.encounterService.CheckObjectives(r.Context(), encounterID); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "checked"})
}
