package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/websocket"
)

// ProcessDMAssistantRequest handles real-time DM assistant requests
func (h *Handlers) ProcessDMAssistantRequest(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request
	var req models.DMAssistantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request type
	validTypes := map[string]bool{
		models.RequestTypeNPCDialogue:         true,
		models.RequestTypeLocationDesc:        true,
		models.RequestTypeCombatNarration:     true,
		models.RequestTypeDeathDescription:    true,
		models.RequestTypePlotTwist:           true,
		models.RequestTypeEnvironmentalHazard: true,
		models.RequestTypeStoryHook:           true,
	}

	if !validTypes[req.Type] {
		respondWithError(w, http.StatusBadRequest, "Invalid request type")
		return
	}

	// Process the request
	result, err := h.services.DMAssistant.ProcessRequest(r.Context(), userID, req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If streaming is requested and this is a WebSocket upgrade request
	if req.StreamResponse && websocket.IsWebSocketUpgrade(r) {
		// Handle WebSocket streaming response
		h.handleStreamingDMAssistant(w, r, userID, req)
		return
	}

	respondWithJSON(w, http.StatusOK, result)
}

// GetDMAssistantNPCs retrieves NPCs for a game session
func (h *Handlers) GetDMAssistantNPCs(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	_, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get session ID from URL
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Get NPCs
	npcs, err := h.services.DMAssistant.GetNPCsBySession(r.Context(), sessionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, npcs)
}

// GetDMAssistantNPC retrieves a specific NPC
func (h *Handlers) GetDMAssistantNPC(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	_, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get NPC ID from URL
	vars := mux.Vars(r)
	npcID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid NPC ID")
		return
	}

	// Get NPC
	npc, err := h.services.DMAssistant.GetNPCByID(r.Context(), npcID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "NPC not found")
		return
	}

	respondWithJSON(w, http.StatusOK, npc)
}

// CreateDMAssistantNPC creates a new AI-generated NPC
func (h *Handlers) CreateDMAssistantNPC(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request
	var req struct {
		SessionID string                 `json:"sessionId"`
		Role      string                 `json:"role"`
		Context   map[string]interface{} `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Create NPC
	npc, err := h.services.DMAssistant.CreateNPC(r.Context(), sessionID, userID, req.Role, req.Context)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, npc)
}

// GetDMAssistantLocations retrieves locations for a game session
func (h *Handlers) GetDMAssistantLocations(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	_, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get session ID from URL
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Get locations
	locations, err := h.services.DMAssistant.GetLocationsBySession(r.Context(), sessionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, locations)
}

// GetDMAssistantLocation retrieves a specific location
func (h *Handlers) GetDMAssistantLocation(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	_, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get location ID from URL
	vars := mux.Vars(r)
	locationID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location ID")
		return
	}

	// Get location
	location, err := h.services.DMAssistant.GetLocationByID(r.Context(), locationID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Location not found")
		return
	}

	respondWithJSON(w, http.StatusOK, location)
}

// GetDMAssistantStoryElements retrieves unused story elements for a session
func (h *Handlers) GetDMAssistantStoryElements(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	_, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get session ID from URL
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Get story elements
	elements, err := h.services.DMAssistant.GetUnusedStoryElements(r.Context(), sessionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, elements)
}

// MarkStoryElementUsed marks a story element as used
func (h *Handlers) MarkStoryElementUsed(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	_, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get element ID from URL
	vars := mux.Vars(r)
	elementID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid element ID")
		return
	}

	// Mark as used
	if err := h.services.DMAssistant.MarkStoryElementUsed(r.Context(), elementID); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "marked as used"})
}

// GetDMAssistantHazards retrieves environmental hazards for a location
func (h *Handlers) GetDMAssistantHazards(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	_, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get location ID from URL
	vars := mux.Vars(r)
	locationID, err := uuid.Parse(vars["locationId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location ID")
		return
	}

	// Get hazards
	hazards, err := h.services.DMAssistant.GetActiveHazards(r.Context(), locationID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, hazards)
}

// TriggerHazard marks a hazard as triggered
func (h *Handlers) TriggerHazard(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	_, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get hazard ID from URL
	vars := mux.Vars(r)
	hazardID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid hazard ID")
		return
	}

	// Trigger hazard
	if err := h.services.DMAssistant.TriggerHazard(r.Context(), hazardID); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "triggered"})
}

// handleStreamingDMAssistant handles WebSocket streaming for DM assistant
func (h *Handlers) handleStreamingDMAssistant(w http.ResponseWriter, r *http.Request, userID uuid.UUID, req models.DMAssistantRequest) {
	// This would upgrade to WebSocket and stream responses
	// Implementation depends on your WebSocket setup
	// For now, we'll just return the non-streaming response
	result, err := h.services.DMAssistant.ProcessRequest(r.Context(), userID, req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, result)
}