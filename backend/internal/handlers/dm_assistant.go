package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

// ProcessDMAssistantRequest handles real-time DM assistant requests
func (h *Handlers) ProcessDMAssistantRequest(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Parse request
	var req models.DMAssistantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Validate request type
	validTypes := map[string]bool{
		models.RequestTypeNPCDialog:           true,
		models.RequestTypeLocationDesc:        true,
		models.RequestTypeCombatNarration:     true,
		models.RequestTypeDeathDescription:    true,
		models.RequestTypePlotTwist:           true,
		models.RequestTypeEnvironmentalHazard: true,
		models.RequestTypeStoryHook:           true,
	}

	if !validTypes[req.Type] {
		response.BadRequest(w, r, "Invalid request type")
		return
	}

	// Convert userID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, r, "Invalid user ID")
		return
	}

	// Process the request
	result, err := h.dmAssistantService.ProcessRequest(r.Context(), userUUID, req)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// TODO: Support streaming via WebSocket when needed

	response.JSON(w, r, http.StatusOK, result)
}

// GetDMAssistantNPCs retrieves NPCs for a game session
func (h *Handlers) GetDMAssistantNPCs(w http.ResponseWriter, r *http.Request) {
	_, sessionID, ok := ExtractUserAndID(w, r, "sessionId")
	if !ok {
		return
	}

	HandleServiceListOperation(w, r, func() ([]*models.AINPC, error) {
		return h.dmAssistantService.GetNPCsBySession(r.Context(), sessionID)
	})
}

// GetDMAssistantNPC retrieves a specific NPC
func (h *Handlers) GetDMAssistantNPC(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	_, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Get NPC ID from URL
	vars := mux.Vars(r)
	npcID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.BadRequest(w, r, "Invalid NPC ID")
		return
	}

	// Get NPC
	npc, err := h.dmAssistantService.GetNPCByID(r.Context(), npcID)
	if err != nil {
		response.NotFound(w, r, "NPC")
		return
	}

	response.JSON(w, r, http.StatusOK, npc)
}

// CreateDMAssistantNPC creates a new AI-generated NPC
func (h *Handlers) CreateDMAssistantNPC(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Parse request
	var req struct {
		SessionID string                 `json:"sessionId"`
		Role      string                 `json:"role"`
		Context   map[string]interface{} `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		response.BadRequest(w, r, "Invalid session ID")
		return
	}

	// Convert userID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, r, "Invalid user ID")
		return
	}

	// Create NPC
	npc, err := h.dmAssistantService.CreateNPC(r.Context(), sessionID, userUUID, req.Role, req.Context)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, npc)
}

// GetDMAssistantLocations retrieves locations for a game session
func (h *Handlers) GetDMAssistantLocations(w http.ResponseWriter, r *http.Request) {
	_, sessionID, ok := ExtractUserAndID(w, r, "sessionId")
	if !ok {
		return
	}

	HandleServiceListOperation(w, r, func() ([]*models.AILocation, error) {
		return h.dmAssistantService.GetLocationsBySession(r.Context(), sessionID)
	})
}

// GetDMAssistantLocation retrieves a specific location
func (h *Handlers) GetDMAssistantLocation(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	_, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Get location ID from URL
	vars := mux.Vars(r)
	locationID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.BadRequest(w, r, "Invalid location ID")
		return
	}

	// Get location
	location, err := h.dmAssistantService.GetLocationByID(r.Context(), locationID)
	if err != nil {
		response.NotFound(w, r, "Location")
		return
	}

	response.JSON(w, r, http.StatusOK, location)
}

// GetDMAssistantStoryElements retrieves unused story elements for a session
func (h *Handlers) GetDMAssistantStoryElements(w http.ResponseWriter, r *http.Request) {
	// Authenticate user
	if !authenticateUser(w, r) {
		return
	}

	// Parse session ID
	sessionID, err := parseUUIDFromRequest(w, r, "sessionId")
	if err != nil {
		return
	}

	// Get story elements
	elements, err := h.dmAssistantService.GetUnusedStoryElements(r.Context(), sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, elements)
}

// MarkStoryElementUsed marks a story element as used
func (h *Handlers) MarkStoryElementUsed(w http.ResponseWriter, r *http.Request) {
	// Authenticate user
	if !authenticateUser(w, r) {
		return
	}

	// Parse element ID
	elementID, err := parseUUIDFromRequest(w, r, "id")
	if err != nil {
		return
	}

	// Mark as used
	if err := h.dmAssistantService.MarkStoryElementUsed(r.Context(), elementID); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "marked as used"})
}

// GetDMAssistantHazards retrieves environmental hazards for a location
func (h *Handlers) GetDMAssistantHazards(w http.ResponseWriter, r *http.Request) {
	// Authenticate user
	if !authenticateUser(w, r) {
		return
	}

	// Parse location ID
	locationID, err := parseUUIDFromRequest(w, r, "locationId")
	if err != nil {
		return
	}

	// Get hazards
	hazards, err := h.dmAssistantService.GetActiveHazards(r.Context(), locationID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, hazards)
}

// TriggerHazard marks a hazard as triggered
func (h *Handlers) TriggerHazard(w http.ResponseWriter, r *http.Request) {
	// Authenticate user
	if !authenticateUser(w, r) {
		return
	}

	// Parse hazard ID
	hazardID, err := parseUUIDFromRequest(w, r, "id")
	if err != nil {
		return
	}

	// Trigger hazard
	if err := h.dmAssistantService.TriggerHazard(r.Context(), hazardID); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "triggered"})
}

// handleStreamingDMAssistant handles WebSocket streaming for DM assistant
