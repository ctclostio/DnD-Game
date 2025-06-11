package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/pkg/response"
)

func (h *Handlers) CreateGameSession(w http.ResponseWriter, r *http.Request) {
	var session models.GameSession
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Get user claims from auth context (DM role is enforced by middleware)
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Set the DM user ID
	session.DMID = claims.UserID

	if err := h.gameService.CreateSession(r.Context(), &session); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, session)
}

func (h *Handlers) GetGameSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	session, err := h.gameService.GetSession(r.Context(), id)
	if err != nil {
		response.NotFound(w, r, err.Error())
		return
	}

	// Verify user is either DM or a participant
	isAuthorized := session.DMID == userID
	if !isAuthorized {
		// Check if user is a participant
		if err := h.gameService.ValidateUserInSession(r.Context(), id, userID); err == nil {
			isAuthorized = true
		}
	}

	if !isAuthorized {
		response.Forbidden(w, r, "You don't have access to this game session")
		return
	}

	response.JSON(w, r, http.StatusOK, session)
}

func (h *Handlers) UpdateGameSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get user claims from auth context (DM role is enforced by middleware)
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Verify the session exists and user is the DM
	existing, err := h.gameService.GetSession(r.Context(), id)
	if err != nil {
		response.NotFound(w, r, "Game session not found")
		return
	}

	if existing.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can update the game session")
		return
	}

	// Decode into a map to handle partial updates
	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Apply only the fields that were provided
	if name, ok := updateData["name"].(string); ok {
		existing.Name = name
	}
	if desc, ok := updateData["description"].(string); ok {
		existing.Description = desc
	}
	// Check both snake_case and camelCase for compatibility
	if isActive, ok := updateData["is_active"].(bool); ok {
		existing.IsActive = isActive
	} else if isActive, ok := updateData["isActive"].(bool); ok {
		existing.IsActive = isActive
	}
	if maxPlayers, ok := updateData["max_players"].(float64); ok {
		existing.MaxPlayers = int(maxPlayers)
	} else if maxPlayers, ok := updateData["maxPlayers"].(float64); ok {
		existing.MaxPlayers = int(maxPlayers)
	}
	if isPublic, ok := updateData["is_public"].(bool); ok {
		existing.IsPublic = isPublic
	} else if isPublic, ok := updateData["isPublic"].(bool); ok {
		existing.IsPublic = isPublic
	}
	if requiresInvite, ok := updateData["requires_invite"].(bool); ok {
		existing.RequiresInvite = requiresInvite
	} else if requiresInvite, ok := updateData["requiresInvite"].(bool); ok {
		existing.RequiresInvite = requiresInvite
	}

	if err := h.gameService.UpdateSession(r.Context(), existing); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Fetch updated session
	updated, err := h.gameService.GetSession(r.Context(), id)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, updated)
}

func (h *Handlers) JoinGameSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	var req struct {
		CharacterID *string `json:"character_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	if err := h.gameService.JoinSession(r.Context(), sessionID, userID, req.CharacterID); err != nil {
		response.BadRequest(w, r, err.Error())
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"message": "Successfully joined game session"})
}

func (h *Handlers) LeaveGameSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	if err := h.gameService.LeaveSession(r.Context(), sessionID, userID); err != nil {
		response.BadRequest(w, r, err.Error())
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"message": "Successfully left game session"})
}

func (h *Handlers) GetUserGameSessions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Get sessions where user is participant
	sessions, err := h.gameService.GetSessionsByPlayer(r.Context(), userID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, sessions)
}

// GetActiveSessions returns all active game sessions (public or user is participant)
func (h *Handlers) GetActiveSessions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Get all sessions where user is participant
	sessions, err := h.gameService.GetSessionsByPlayer(r.Context(), userID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Filter for active sessions only
	activeSessions := make([]*models.GameSession, 0)
	for _, session := range sessions {
		if session.IsActive && session.Status != models.GameStatusCompleted {
			activeSessions = append(activeSessions, session)
		}
	}

	response.JSON(w, r, http.StatusOK, activeSessions)
}

// GetSessionPlayers returns all players in a game session
func (h *Handlers) GetSessionPlayers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Verify user is in the session
	if err := h.gameService.ValidateUserInSession(r.Context(), sessionID, userID); err != nil {
		response.Forbidden(w, r, "You are not a participant in this session")
		return
	}

	// Get all participants
	participants, err := h.gameService.GetSessionParticipants(r.Context(), sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, participants)
}

// KickPlayer removes a player from the game session (DM only)
func (h *Handlers) KickPlayer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]
	playerID := vars["playerId"]

	// Get user claims from auth context (DM role is enforced by middleware)
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// Verify the session exists and user is the DM
	session, err := h.gameService.GetSession(r.Context(), sessionID)
	if err != nil {
		response.NotFound(w, r, "Game session not found")
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can kick players")
		return
	}

	// Prevent DM from kicking themselves
	if playerID == claims.UserID {
		response.BadRequest(w, r, "DM cannot kick themselves")
		return
	}

	// Remove the player
	if err := h.gameService.KickPlayer(r.Context(), sessionID, playerID); err != nil {
		response.BadRequest(w, r, err.Error())
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{
		"message": "Player kicked successfully",
	})
}