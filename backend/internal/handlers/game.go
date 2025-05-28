package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
)

func (h *Handlers) CreateGameSession(w http.ResponseWriter, r *http.Request) {
	var session models.GameSession
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user claims from auth context (DM role is enforced by middleware)
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Set the DM user ID
	session.DMUserID = claims.UserID

	if err := h.gameService.CreateSession(r.Context(), &session); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusCreated, session)
}

func (h *Handlers) GetGameSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session, err := h.gameService.GetSession(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	// Verify user is either DM or a participant
	isAuthorized := session.DMUserID == userID
	if !isAuthorized {
		// Check if user is a participant
		if err := h.gameService.ValidateUserInSession(r.Context(), id, userID); err == nil {
			isAuthorized = true
		}
	}

	if !isAuthorized {
		sendErrorResponse(w, http.StatusForbidden, "You don't have access to this game session")
		return
	}

	sendJSONResponse(w, http.StatusOK, session)
}

func (h *Handlers) UpdateGameSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get user claims from auth context (DM role is enforced by middleware)
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Verify the session exists and user is the DM
	existing, err := h.gameService.GetSession(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Game session not found")
		return
	}

	if existing.DMUserID != claims.UserID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can update the game session")
		return
	}

	var session models.GameSession
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	session.ID = id
	session.DMUserID = claims.UserID // Ensure DM can't be changed
	
	if err := h.gameService.UpdateSession(r.Context(), &session); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch updated session
	updated, err := h.gameService.GetSession(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, updated)
}

func (h *Handlers) JoinGameSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	var req struct {
		CharacterID *string `json:"characterId,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.gameService.JoinSession(r.Context(), sessionID, userID, req.CharacterID); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{"message": "Successfully joined game session"})
}

func (h *Handlers) LeaveGameSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.gameService.LeaveSession(r.Context(), sessionID, userID); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{"message": "Successfully left game session"})
}

func (h *Handlers) GetUserGameSessions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get sessions where user is participant
	sessions, err := h.gameService.GetSessionsByPlayer(r.Context(), userID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, sessions)
}