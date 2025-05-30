package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
)

type CombatAutomationHandler struct {
	combatAutomation *services.CombatAutomationService
	combatAnalytics  *services.CombatAnalyticsService
	characterService *services.CharacterService
	gameService      *services.GameSessionService
	mapGenerator     *services.AIBattleMapGenerator
}

func NewCombatAutomationHandler(
	combatAutomation *services.CombatAutomationService,
	combatAnalytics *services.CombatAnalyticsService,
	characterService *services.CharacterService,
	gameService *services.GameSessionService,
	mapGenerator *services.AIBattleMapGenerator,
) *CombatAutomationHandler {
	return &CombatAutomationHandler{
		combatAutomation: combatAutomation,
		combatAnalytics:  combatAnalytics,
		characterService: characterService,
		gameService:      gameService,
		mapGenerator:     mapGenerator,
	}
}

// AutoResolveCombat handles quick combat resolution
func (h *CombatAutomationHandler) AutoResolveCombat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		respondWithError(w, http.StatusForbidden, "Only the DM can auto-resolve combat")
		return
	}

	var req models.AutoResolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get party characters
	participants, err := h.gameService.GetSessionParticipants(ctx, sessionID.String())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get participants")
		return
	}

	var characters []*models.Character
	for _, p := range participants {
		if p.CharacterID != "" {
			char, err := h.characterService.GetCharacterByID(ctx, p.CharacterID)
			if err == nil {
				characters = append(characters, char)
			}
		}
	}

	if len(characters) == 0 {
		respondWithError(w, http.StatusBadRequest, "No characters in party")
		return
	}

	// Auto-resolve the combat
	resolution, err := h.combatAutomation.AutoResolveCombat(ctx, sessionID, characters, req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, resolution)
}

// SmartInitiative handles automatic initiative rolling
func (h *CombatAutomationHandler) SmartInitiative(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Verify user is in session
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Check if user is participant or DM
	isParticipant := session.DMID == claims.UserID
	if !isParticipant {
		participants, _ := h.gameService.GetSessionParticipants(ctx, sessionID.String())
		for _, p := range participants {
			if p.UserID == claims.UserID {
				isParticipant = true
				break
			}
		}
	}

	if !isParticipant {
		respondWithError(w, http.StatusForbidden, "User not in session")
		return
	}

	var req models.SmartInitiativeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Calculate initiative for all combatants
	initiatives, err := h.combatAutomation.SmartInitiative(ctx, sessionID, req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, initiatives)
}

// GenerateBattleMap creates a tactical map
func (h *CombatAutomationHandler) GenerateBattleMap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		respondWithError(w, http.StatusForbidden, "Only the DM can generate battle maps")
		return
	}

	var req models.GenerateBattleMapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Generate the battle map
	battleMap, err := h.mapGenerator.GenerateBattleMap(ctx, req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Save to database
	battleMap.GameSessionID = sessionID
	if err := h.combatAutomation.SaveBattleMap(ctx, battleMap); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save battle map")
		return
	}

	respondWithJSON(w, http.StatusCreated, battleMap)
}

// GetCombatAnalytics retrieves combat analytics report
func (h *CombatAutomationHandler) GetCombatAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	combatID, err := uuid.Parse(vars["combatId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid combat ID")
		return
	}

	// Get combat analytics
	analytics, err := h.combatAnalytics.GetCombatAnalytics(ctx, combatID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Combat analytics not found")
		return
	}

	// Get combatant reports
	combatantAnalytics, err := h.combatAnalytics.GetCombatantAnalytics(ctx, analytics.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get combatant analytics")
		return
	}

	// Build full report
	report := models.CombatAnalyticsReport{
		Analytics: analytics,
		CombatantReports: h.buildCombatantReports(combatantAnalytics),
	}

	respondWithJSON(w, http.StatusOK, report)
}

// GetSessionCombatHistory retrieves combat history for a session
func (h *CombatAutomationHandler) GetSessionCombatHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Get all combat analytics for session
	analytics, err := h.combatAnalytics.GetCombatAnalyticsBySession(ctx, sessionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get combat history")
		return
	}

	// Get auto-resolutions
	resolutions, err := h.combatAutomation.GetAutoResolutionsBySession(ctx, sessionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get auto-resolutions")
		return
	}

	response := map[string]interface{}{
		"combat_analytics":   analytics,
		"auto_resolutions":   resolutions,
		"total_combats":      len(analytics) + len(resolutions),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetBattleMaps retrieves battle maps for a session
func (h *CombatAutomationHandler) GetBattleMaps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	maps, err := h.combatAutomation.GetBattleMapsBySession(ctx, sessionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get battle maps")
		return
	}

	respondWithJSON(w, http.StatusOK, maps)
}

// GetBattleMap retrieves a specific battle map
func (h *CombatAutomationHandler) GetBattleMap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	mapID, err := uuid.Parse(vars["mapId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid map ID")
		return
	}

	battleMap, err := h.combatAutomation.GetBattleMap(ctx, mapID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Battle map not found")
		return
	}

	respondWithJSON(w, http.StatusOK, battleMap)
}

// SetInitiativeRules sets special initiative rules for entities
func (h *CombatAutomationHandler) SetInitiativeRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		respondWithError(w, http.StatusForbidden, "Only the DM can set initiative rules")
		return
	}

	var rule models.SmartInitiativeRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	rule.GameSessionID = sessionID
	if err := h.combatAutomation.SetInitiativeRule(ctx, &rule); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, rule)
}

// Helper methods

func (h *CombatAutomationHandler) buildCombatantReports(analytics []*models.CombatantAnalytics) []*models.CombatantReport {
	reports := make([]*models.CombatantReport, len(analytics))
	
	for i, a := range analytics {
		reports[i] = &models.CombatantReport{
			Analytics:         a,
			PerformanceRating: h.ratePerformance(a),
			Highlights:        h.generateHighlights(a),
		}
	}
	
	return reports
}

func (h *CombatAutomationHandler) ratePerformance(stats *models.CombatantAnalytics) string {
	score := 0
	
	if stats.AttacksMade > 0 {
		hitRate := float64(stats.AttacksHit) / float64(stats.AttacksMade)
		if hitRate > 0.75 {
			score += 3
		} else if hitRate > 0.5 {
			score += 2
		}
	}
	
	if stats.FinalHP > 0 {
		score += 2
	}
	
	if stats.DamageDealt > stats.DamageTaken*2 {
		score += 2
	}
	
	if score >= 6 {
		return "excellent"
	} else if score >= 4 {
		return "good"
	} else if score >= 2 {
		return "fair"
	}
	return "poor"
}

func (h *CombatAutomationHandler) generateHighlights(stats *models.CombatantAnalytics) []string {
	highlights := []string{}
	
	if stats.AttacksMade > 0 {
		hitRate := float64(stats.AttacksHit) / float64(stats.AttacksMade)
		if hitRate > 0.75 {
			highlights = append(highlights, fmt.Sprintf("%.0f%% hit rate", hitRate*100))
		}
	}
	
	if stats.CriticalHits > 1 {
		highlights = append(highlights, fmt.Sprintf("%d critical hits", stats.CriticalHits))
	}
	
	if stats.DamageDealt > 50 {
		highlights = append(highlights, fmt.Sprintf("%d damage dealt", stats.DamageDealt))
	}
	
	return highlights
}