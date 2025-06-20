package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

func (h *Handlers) CreateGameSession(w http.ResponseWriter, r *http.Request) {
	var session models.GameSession
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	// Get user claims from auth context (DM role is enforced by middleware)
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, constants.ErrUnauthorized)
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
		response.Unauthorized(w, r, constants.ErrUnauthorized)
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
		response.Unauthorized(w, r, constants.ErrUnauthorized)
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
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	// Apply updates from request data
	applyGameSessionUpdates(existing, updateData)

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
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, constants.ErrUnauthorized)
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
		response.Unauthorized(w, r, constants.ErrUnauthorized)
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
		response.Unauthorized(w, r, constants.ErrUnauthorized)
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
		response.Unauthorized(w, r, constants.ErrUnauthorized)
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

	// Validate user session access
	err := validateUserSession(w, r, h.gameService, sessionID)
	if err != nil {
		return // Response already sent by helper
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
		response.Unauthorized(w, r, constants.ErrUnauthorized)
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

// Helper function to apply game session updates from request data
func applyGameSessionUpdates(session *models.GameSession, updateData map[string]interface{}) {
	// Apply only the fields that were provided
	if name, ok := updateData["name"].(string); ok {
		session.Name = name
	}
	if desc, ok := updateData["description"].(string); ok {
		session.Description = desc
	}
	
	// Check both snake_case and camelCase for compatibility
	applyBoolUpdate(&session.IsActive, updateData, "is_active", "isActive")
	applyIntUpdate(&session.MaxPlayers, updateData, "max_players", "maxPlayers")
	applyBoolUpdate(&session.IsPublic, updateData, "is_public", "isPublic")
	applyBoolUpdate(&session.RequiresInvite, updateData, "requires_invite", "requiresInvite")
}

// applyBoolUpdate updates a bool field from either snake_case or camelCase key
func applyBoolUpdate(field *bool, data map[string]interface{}, snakeKey, camelKey string) {
	if val, ok := data[snakeKey].(bool); ok {
		*field = val
	} else if val, ok := data[camelKey].(bool); ok {
		*field = val
	}
}

// applyIntUpdate updates an int field from either snake_case or camelCase key
func applyIntUpdate(field *int, data map[string]interface{}, snakeKey, camelKey string) {
	if val, ok := data[snakeKey].(float64); ok {
		*field = int(val)
	} else if val, ok := data[camelKey].(float64); ok {
		*field = int(val)
	}
}
