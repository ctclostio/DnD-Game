package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/pkg/response"
)

// CreateNPC handles NPC creation (DM only)
func (h *Handlers) CreateNPC(w http.ResponseWriter, r *http.Request) {
	// Get user claims from auth context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	var npc models.NPC
	if err := json.NewDecoder(r.Body).Decode(&npc); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Verify user is DM of the game session
	session, err := h.gameService.GetSession(r.Context(), npc.GameSessionID)
	if err != nil {
		response.NotFound(w, r, "game session")
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can create NPCs")
		return
	}

	// Set creator
	npc.CreatedBy = claims.UserID

	if err := h.npcService.CreateNPC(r.Context(), &npc); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, npc)
}

// GetNPC retrieves an NPC by ID
func (h *Handlers) GetNPC(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	npcID := vars["id"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	npc, err := h.npcService.GetNPC(r.Context(), npcID)
	if err != nil {
		response.NotFound(w, r, "NPC")
		return
	}

	// Verify user has access to this NPC's game session
	if err := h.gameService.ValidateUserInSession(r.Context(), npc.GameSessionID, userID); err != nil {
		response.Forbidden(w, r, "You don't have access to this NPC")
		return
	}

	response.JSON(w, r, http.StatusOK, npc)
}

// GetNPCsBySession retrieves all NPCs for a game session
func (h *Handlers) GetNPCsBySession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Verify user has access to this game session
	if err := h.gameService.ValidateUserInSession(r.Context(), sessionID, userID); err != nil {
		response.Forbidden(w, r, "You don't have access to this game session")
		return
	}

	npcs, err := h.npcService.GetNPCsByGameSession(r.Context(), sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, npcs)
}

// UpdateNPC updates an NPC (DM only)
func (h *Handlers) UpdateNPC(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	npcID := vars["id"]

	// Get user claims from auth context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Get existing NPC
	existingNPC, err := h.npcService.GetNPC(r.Context(), npcID)
	if err != nil {
		response.NotFound(w, r, "NPC")
		return
	}

	// Verify user is DM of the game session
	session, err := h.gameService.GetSession(r.Context(), existingNPC.GameSessionID)
	if err != nil {
		response.NotFound(w, r, "game session")
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can update NPCs")
		return
	}

	var npc models.NPC
	if err := json.NewDecoder(r.Body).Decode(&npc); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Ensure ID and GameSessionID can't be changed
	npc.ID = npcID
	npc.GameSessionID = existingNPC.GameSessionID
	npc.CreatedBy = existingNPC.CreatedBy

	if err := h.npcService.UpdateNPC(r.Context(), &npc); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, npc)
}

// DeleteNPC deletes an NPC (DM only)
func (h *Handlers) DeleteNPC(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	npcID := vars["id"]

	// Get user claims from auth context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Get NPC to verify permissions
	npc, err := h.npcService.GetNPC(r.Context(), npcID)
	if err != nil {
		response.NotFound(w, r, "NPC")
		return
	}

	// Verify user is DM of the game session
	session, err := h.gameService.GetSession(r.Context(), npc.GameSessionID)
	if err != nil {
		response.NotFound(w, r, "game session")
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can delete NPCs")
		return
	}

	if err := h.npcService.DeleteNPC(r.Context(), npcID); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusNoContent, nil)
}

// SearchNPCs searches for NPCs
func (h *Handlers) SearchNPCs(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Parse query parameters
	filter := models.NPCSearchFilter{
		GameSessionID:    r.URL.Query().Get("gameSessionId"),
		Name:             r.URL.Query().Get("name"),
		Type:             r.URL.Query().Get("type"),
		Size:             r.URL.Query().Get("size"),
		IncludeTemplates: r.URL.Query().Get("includeTemplates") == "true",
	}

	// Parse CR range
	if minCR := r.URL.Query().Get("minCR"); minCR != "" {
		if cr, err := strconv.ParseFloat(minCR, 64); err == nil {
			filter.MinCR = cr
		}
	}

	if maxCR := r.URL.Query().Get("maxCR"); maxCR != "" {
		if cr, err := strconv.ParseFloat(maxCR, 64); err == nil {
			filter.MaxCR = cr
		}
	}

	// If searching within a game session, verify access
	if filter.GameSessionID != "" {
		if err := h.gameService.ValidateUserInSession(r.Context(), filter.GameSessionID, userID); err != nil {
			response.Forbidden(w, r, "You don't have access to this game session")
			return
		}
	}

	npcs, err := h.npcService.SearchNPCs(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, npcs)
}

// GetNPCTemplates retrieves all available NPC templates
func (h *Handlers) GetNPCTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.npcService.GetTemplates(r.Context())
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, templates)
}

// CreateNPCFromTemplate creates an NPC from a template (DM only)
func (h *Handlers) CreateNPCFromTemplate(w http.ResponseWriter, r *http.Request) {
	// Get user claims from auth context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	var req struct {
		TemplateID    string `json:"templateId"`
		GameSessionID string `json:"gameSessionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Verify user is DM of the game session
	session, err := h.gameService.GetSession(r.Context(), req.GameSessionID)
	if err != nil {
		response.NotFound(w, r, "game session")
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can create NPCs")
		return
	}

	npc, err := h.npcService.CreateFromTemplate(r.Context(), req.TemplateID, req.GameSessionID, claims.UserID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, npc)
}

// NPCQuickActions handles quick actions on NPCs (damage, heal, etc.)
func (h *Handlers) NPCQuickActions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	npcID := vars["id"]
	action := vars["action"]

	// Get user claims from auth context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Get NPC to verify permissions
	npc, err := h.npcService.GetNPC(r.Context(), npcID)
	if err != nil {
		response.NotFound(w, r, "NPC")
		return
	}

	// Verify user is DM of the game session
	session, err := h.gameService.GetSession(r.Context(), npc.GameSessionID)
	if err != nil {
		response.NotFound(w, r, "game session")
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can perform quick actions on NPCs")
		return
	}

	switch action {
	case "damage":
		var req struct {
			Amount     int    `json:"amount"`
			DamageType string `json:"damageType"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, r, "Invalid request body")
			return
		}

		if err := h.npcService.ApplyDamage(r.Context(), npcID, req.Amount, req.DamageType); err != nil {
			response.InternalServerError(w, r, err)
			return
		}

	case "heal":
		var req struct {
			Amount int `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, r, "Invalid request body")
			return
		}

		if err := h.npcService.HealNPC(r.Context(), npcID, req.Amount); err != nil {
			response.InternalServerError(w, r, err)
			return
		}

	case "initiative":
		initiative, err := h.npcService.RollInitiative(r.Context(), npcID)
		if err != nil {
			response.InternalServerError(w, r, err)
			return
		}

		response.JSON(w, r, http.StatusOK, map[string]int{"initiative": initiative})
		return

	default:
		response.BadRequest(w, r, "Invalid action")
		return
	}

	// Return updated NPC
	updatedNPC, err := h.npcService.GetNPC(r.Context(), npcID)
	if err != nil {
		response.InternalServerError(w, r, errors.New("Failed to retrieve updated NPC"))
		return
	}

	response.JSON(w, r, http.StatusOK, updatedNPC)
}
