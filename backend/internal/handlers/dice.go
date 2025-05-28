package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
)

type DiceRollRequest struct {
	GameSessionID string `json:"gameSessionId"`
	RollNotation  string `json:"rollNotation"` // e.g., "d20", "2d6", "1d8+3"
	Purpose       string `json:"purpose"`      // attack, damage, skill check, etc.
}

type DiceRollResponse struct {
	Roll    *models.DiceRoll `json:"roll"`
	Success bool             `json:"success"`
}

func (h *Handlers) RollDice(w http.ResponseWriter, r *http.Request) {
	var req DiceRollRequest
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

	// Validate game session ID
	if req.GameSessionID == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Game session ID is required")
		return
	}

	// Validate user is in the game session
	if err := h.gameService.ValidateUserInSession(r.Context(), req.GameSessionID, userID); err != nil {
		sendErrorResponse(w, http.StatusForbidden, "User is not a participant in this game session")
		return
	}

	// Create dice roll
	roll := &models.DiceRoll{
		GameSessionID: req.GameSessionID,
		UserID:        userID,
		RollNotation:  req.RollNotation,
		Purpose:       req.Purpose,
	}

	if err := h.diceService.RollDice(r.Context(), roll); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	response := DiceRollResponse{
		Roll:    roll,
		Success: true,
	}

	sendJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) GetDiceRolls(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	gameSessionID := r.URL.Query().Get("game_session_id")
	userID := r.URL.Query().Get("user_id")
	offset := 0
	limit := 50 // Default limit

	// Parse offset and limit
	// TODO: Add proper parsing for offset and limit

	var rolls []*models.DiceRoll
	var err error

	if gameSessionID != "" && userID != "" {
		rolls, err = h.diceService.GetRollsBySessionAndUser(r.Context(), gameSessionID, userID, offset, limit)
	} else if gameSessionID != "" {
		rolls, err = h.diceService.GetRollsBySession(r.Context(), gameSessionID, offset, limit)
	} else if userID != "" {
		rolls, err = h.diceService.GetRollsByUser(r.Context(), userID, offset, limit)
	} else {
		sendErrorResponse(w, http.StatusBadRequest, "Either game_session_id or user_id is required")
		return
	}

	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, rolls)
}