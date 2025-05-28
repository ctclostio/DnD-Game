package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"context"
	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/websocket"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *Handlers) StartCombat(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)

	var req struct {
		GameSessionID string              `json:"gameSessionId"`
		Combatants    []models.Combatant  `json:"combatants"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Verify user is in the game session
	session, err := h.gameService.GetGameSession(r.Context(), req.GameSessionID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Game session not found")
		return
	}

	// Check if user is DM
	if session.DMUserID != claims.UserID {
		respondWithError(w, http.StatusForbidden, "Only the DM can start combat")
		return
	}

	combat, err := h.combatService.StartCombat(r.Context(), req.GameSessionID, req.Combatants)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Broadcast combat start
	h.broadcastCombatUpdate(req.GameSessionID, models.CombatUpdate{
		Type:    models.UpdateTypeCombatStart,
		Combat:  combat,
		Message: "Combat has begun!",
	})

	respondWithJSON(w, http.StatusCreated, combat)
}

func (h *Handlers) GetCombat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	combatID := vars["id"]

	combat, err := h.combatService.GetCombat(r.Context(), combatID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Combat not found")
		return
	}

	respondWithJSON(w, http.StatusOK, combat)
}

func (h *Handlers) GetCombatBySession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	combat, err := h.combatService.GetCombatBySession(r.Context(), sessionID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "No active combat for session")
		return
	}

	respondWithJSON(w, http.StatusOK, combat)
}

func (h *Handlers) NextTurn(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)
	vars := mux.Vars(r)
	combatID := vars["id"]

	combat, err := h.combatService.GetCombat(r.Context(), combatID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Combat not found")
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetGameSession(r.Context(), combat.GameSessionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get game session")
		return
	}

	if session.DMUserID != claims.UserID {
		respondWithError(w, http.StatusForbidden, "Only the DM can advance turns")
		return
	}

	combatant, err := h.combatService.NextTurn(r.Context(), combatID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Broadcast turn change
	h.broadcastCombatUpdate(combat.GameSessionID, models.CombatUpdate{
		Type:    models.UpdateTypeTurnStart,
		Combat:  combat,
		Message: combatant.Name + "'s turn",
	})

	respondWithJSON(w, http.StatusOK, combatant)
}

func (h *Handlers) ProcessCombatAction(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)
	vars := mux.Vars(r)
	combatID := vars["id"]

	var request models.CombatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	combat, err := h.combatService.GetCombat(r.Context(), combatID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Combat not found")
		return
	}

	// Verify user can control the actor
	if !h.canControlCombatant(r.Context(), claims.UserID, combat, request.ActorID) {
		respondWithError(w, http.StatusForbidden, "You cannot control this combatant")
		return
	}

	action, err := h.combatService.ProcessAction(r.Context(), combatID, request)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated combat state
	updatedCombat, _ := h.combatService.GetCombat(r.Context(), combatID)

	// Broadcast action
	h.broadcastCombatUpdate(combat.GameSessionID, models.CombatUpdate{
		Type:    models.UpdateTypeAction,
		Combat:  updatedCombat,
		Action:  action,
		Message: action.Description,
	})

	respondWithJSON(w, http.StatusOK, action)
}

func (h *Handlers) EndCombat(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)
	vars := mux.Vars(r)
	combatID := vars["id"]

	combat, err := h.combatService.GetCombat(r.Context(), combatID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Combat not found")
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetGameSession(r.Context(), combat.GameSessionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get game session")
		return
	}

	if session.DMUserID != claims.UserID {
		respondWithError(w, http.StatusForbidden, "Only the DM can end combat")
		return
	}

	if err := h.combatService.EndCombat(r.Context(), combatID); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Broadcast combat end
	h.broadcastCombatUpdate(combat.GameSessionID, models.CombatUpdate{
		Type:    models.UpdateTypeCombatEnd,
		Message: "Combat has ended",
	})

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Combat ended"})
}

func (h *Handlers) MakeSavingThrow(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)
	vars := mux.Vars(r)
	combatID := vars["id"]
	combatantID := vars["combatantId"]

	var req struct {
		Ability      string `json:"ability"`
		DC           int    `json:"dc"`
		Advantage    bool   `json:"advantage"`
		Disadvantage bool   `json:"disadvantage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	combat, err := h.combatService.GetCombat(r.Context(), combatID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Combat not found")
		return
	}

	// Verify user can control the combatant
	if !h.canControlCombatant(r.Context(), claims.UserID, combat, combatantID) {
		respondWithError(w, http.StatusForbidden, "You cannot control this combatant")
		return
	}

	roll, success, err := h.combatService.MakeSavingThrow(r.Context(), combatID, combatantID, req.Ability, req.DC, req.Advantage, req.Disadvantage)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]interface{}{
		"roll":    roll,
		"success": success,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) ApplyDamage(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)
	vars := mux.Vars(r)
	combatID := vars["id"]
	combatantID := vars["combatantId"]

	var req struct {
		Damage []models.Damage `json:"damage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	combat, err := h.combatService.GetCombat(r.Context(), combatID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Combat not found")
		return
	}

	// Verify user is DM or damage is self-inflicted
	session, _ := h.gameService.GetGameSession(r.Context(), combat.GameSessionID)
	if session.DMUserID != claims.UserID && !h.canControlCombatant(r.Context(), claims.UserID, combat, combatantID) {
		respondWithError(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	totalDamage, err := h.combatService.ApplyDamage(r.Context(), combatID, combatantID, req.Damage)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated combat state
	updatedCombat, _ := h.combatService.GetCombat(r.Context(), combatID)

	// Find combatant name
	var combatantName string
	for _, c := range updatedCombat.Combatants {
		if c.ID == combatantID {
			combatantName = c.Name
			break
		}
	}

	// Broadcast HP change
	h.broadcastCombatUpdate(combat.GameSessionID, models.CombatUpdate{
		Type:    models.UpdateTypeHPChange,
		Combat:  updatedCombat,
		Message: fmt.Sprintf("%s takes %d damage", combatantName, totalDamage),
	})

	respondWithJSON(w, http.StatusOK, map[string]int{"totalDamage": totalDamage})
}

func (h *Handlers) HealCombatant(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)
	vars := mux.Vars(r)
	combatID := vars["id"]
	combatantID := vars["combatantId"]

	var req struct {
		Healing int `json:"healing"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	combat, err := h.combatService.GetCombat(r.Context(), combatID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Combat not found")
		return
	}

	// Verify user is DM or healing is self-inflicted
	session, _ := h.gameService.GetGameSession(r.Context(), combat.GameSessionID)
	if session.DMUserID != claims.UserID && !h.canControlCombatant(r.Context(), claims.UserID, combat, combatantID) {
		respondWithError(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	if err := h.combatService.HealCombatant(r.Context(), combatID, combatantID, req.Healing); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated combat state
	updatedCombat, _ := h.combatService.GetCombat(r.Context(), combatID)

	// Find combatant name
	var combatantName string
	for _, c := range updatedCombat.Combatants {
		if c.ID == combatantID {
			combatantName = c.Name
			break
		}
	}

	// Broadcast HP change
	h.broadcastCombatUpdate(combat.GameSessionID, models.CombatUpdate{
		Type:    models.UpdateTypeHPChange,
		Combat:  updatedCombat,
		Message: fmt.Sprintf("%s heals for %d HP", combatantName, req.Healing),
	})

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Healing applied"})
}

// Helper function to check if a user can control a combatant
func (h *Handlers) canControlCombatant(ctx context.Context, userID string, combat *models.Combat, combatantID string) bool {
	// Find the combatant
	var combatant *models.Combatant
	for i := range combat.Combatants {
		if combat.Combatants[i].ID == combatantID {
			combatant = &combat.Combatants[i]
			break
		}
	}

	if combatant == nil {
		return false
	}

	// DM can control all combatants
	session, err := h.gameService.GetGameSession(ctx, combat.GameSessionID)
	if err == nil && session.DMUserID == userID {
		return true
	}

	// Players can only control their own characters
	if combatant.IsPlayerCharacter && combatant.CharacterID != "" {
		character, err := h.characterService.GetCharacter(ctx, combatant.CharacterID)
		if err == nil && character.UserID == userID {
			return true
		}
	}

	return false
}

// Helper function to broadcast combat updates
func (h *Handlers) broadcastCombatUpdate(gameSessionID string, update models.CombatUpdate) {
	message := websocket.Message{
		Type:   "combat",
		RoomID: gameSessionID,
		Data:   nil,
	}
	
	data, err := json.Marshal(update)
	if err != nil {
		return
	}
	
	message.Data = data
	
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return
	}
	
	h.websocketHub.Broadcast(msgBytes)
}