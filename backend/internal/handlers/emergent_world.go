package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
)

// EmergentWorldHandlers manages handlers for the emergent world system
type EmergentWorldHandlers struct {
	livingEcosystem      *services.LivingEcosystemService
	factionPersonality   *services.FactionPersonalityService
	proceduralCulture    *services.ProceduralCultureService
	worldRepo            *database.EmergentWorldRepository
}

// NewEmergentWorldHandlers creates new emergent world handlers
func NewEmergentWorldHandlers(
	livingEcosystem *services.LivingEcosystemService,
	factionPersonality *services.FactionPersonalityService,
	proceduralCulture *services.ProceduralCultureService,
	worldRepo *database.EmergentWorldRepository,
) *EmergentWorldHandlers {
	return &EmergentWorldHandlers{
		livingEcosystem:    livingEcosystem,
		factionPersonality: factionPersonality,
		proceduralCulture:  proceduralCulture,
		worldRepo:          worldRepo,
	}
}

// SimulateWorld handles POST /api/sessions/{sessionId}/world/simulate
func (h *EmergentWorldHandlers) SimulateWorld(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Verify user is DM
	userID := auth.GetUserIDFromContext(r.Context())
	session, err := h.gameService.GetByID(r.Context(), sessionID)
	if err != nil || session.DMUserID != userID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can simulate world progress")
		return
	}

	// Run simulation
	err = h.livingEcosystem.SimulateWorldProgress(r.Context(), sessionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to simulate world progress")
		return
	}

	// Get recent events
	events, err := h.worldRepo.GetWorldEvents(sessionID, 10, false)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve world events")
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "World simulation completed",
		"events":  events,
	})
}

// GetWorldEvents handles GET /api/sessions/{sessionId}/world/events
func (h *EmergentWorldHandlers) GetWorldEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Parse query parameters
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	onlyVisible := r.URL.Query().Get("visible") == "true"

	events, err := h.worldRepo.GetWorldEvents(sessionID, limit, onlyVisible)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve world events")
		return
	}

	sendJSONResponse(w, http.StatusOK, events)
}

// GetWorldState handles GET /api/sessions/{sessionId}/world/state
func (h *EmergentWorldHandlers) GetWorldState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	state, err := h.worldRepo.GetWorldState(sessionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve world state")
		return
	}

	sendJSONResponse(w, http.StatusOK, state)
}

// NPC Autonomy Handlers

// GetNPCGoals handles GET /api/npcs/{npcId}/goals
func (h *EmergentWorldHandlers) GetNPCGoals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	npcID := vars["npcId"]

	goals, err := h.worldRepo.GetNPCGoals(npcID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve NPC goals")
		return
	}

	sendJSONResponse(w, http.StatusOK, goals)
}

// CreateNPCGoal handles POST /api/npcs/{npcId}/goals
func (h *EmergentWorldHandlers) CreateNPCGoal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	npcID := vars["npcId"]

	var goal models.NPCGoal
	if err := json.NewDecoder(r.Body).Decode(&goal); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	goal.NPCID = npcID
	goal.ID = generateUUID()

	if err := h.worldRepo.CreateNPCGoal(&goal); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to create NPC goal")
		return
	}

	sendJSONResponse(w, http.StatusCreated, goal)
}

// GetNPCSchedule handles GET /api/npcs/{npcId}/schedule
func (h *EmergentWorldHandlers) GetNPCSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	npcID := vars["npcId"]

	schedule, err := h.worldRepo.GetNPCSchedule(npcID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve NPC schedule")
		return
	}

	sendJSONResponse(w, http.StatusOK, schedule)
}

// Faction Personality Handlers

// InitializeFactionPersonality handles POST /api/factions/{factionId}/personality
func (h *EmergentWorldHandlers) InitializeFactionPersonality(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	factionID := vars["factionId"]

	// Get faction
	faction, err := h.factionRepo.GetFaction(factionID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Faction not found")
		return
	}

	// Initialize personality
	personality, err := h.factionPersonality.InitializeFactionPersonality(r.Context(), faction)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to initialize faction personality")
		return
	}

	sendJSONResponse(w, http.StatusCreated, personality)
}

// GetFactionPersonality handles GET /api/factions/{factionId}/personality
func (h *EmergentWorldHandlers) GetFactionPersonality(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	factionID := vars["factionId"]

	personality, err := h.worldRepo.GetFactionPersonality(factionID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Faction personality not found")
		return
	}

	sendJSONResponse(w, http.StatusOK, personality)
}

// MakeFactionDecision handles POST /api/factions/{factionId}/decide
func (h *EmergentWorldHandlers) MakeFactionDecision(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	factionID := vars["factionId"]

	var decision services.FactionDecision
	if err := json.NewDecoder(r.Body).Decode(&decision); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.factionPersonality.MakeFactionDecision(r.Context(), factionID, decision)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to make faction decision")
		return
	}

	sendJSONResponse(w, http.StatusOK, result)
}

// GetFactionAgendas handles GET /api/factions/{factionId}/agendas
func (h *EmergentWorldHandlers) GetFactionAgendas(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	factionID := vars["factionId"]

	agendas, err := h.worldRepo.GetFactionAgendas(factionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve faction agendas")
		return
	}

	sendJSONResponse(w, http.StatusOK, agendas)
}

// RecordFactionInteraction handles POST /api/factions/{factionId}/interactions
func (h *EmergentWorldHandlers) RecordFactionInteraction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	factionID := vars["factionId"]

	var interaction services.PlayerInteraction
	if err := json.NewDecoder(r.Body).Decode(&interaction); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.factionPersonality.LearnFromInteraction(r.Context(), factionID, interaction)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to record faction interaction")
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Faction interaction recorded successfully",
	})
}

// Procedural Culture Handlers

// GenerateCulture handles POST /api/sessions/{sessionId}/cultures/generate
func (h *EmergentWorldHandlers) GenerateCulture(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var params services.CultureGenParameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	culture, err := h.proceduralCulture.GenerateCulture(r.Context(), sessionID, params)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to generate culture")
		return
	}

	sendJSONResponse(w, http.StatusCreated, culture)
}

// GetCultures handles GET /api/sessions/{sessionId}/cultures
func (h *EmergentWorldHandlers) GetCultures(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	cultures, err := h.worldRepo.GetCulturesBySession(sessionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve cultures")
		return
	}

	sendJSONResponse(w, http.StatusOK, cultures)
}

// GetCulture handles GET /api/cultures/{cultureId}
func (h *EmergentWorldHandlers) GetCulture(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cultureID := vars["cultureId"]

	culture, err := h.worldRepo.GetCulture(cultureID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Culture not found")
		return
	}

	sendJSONResponse(w, http.StatusOK, culture)
}

// InteractWithCulture handles POST /api/cultures/{cultureId}/interact
func (h *EmergentWorldHandlers) InteractWithCulture(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cultureID := vars["cultureId"]

	var action services.PlayerCulturalAction
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.proceduralCulture.RespondToPlayerAction(r.Context(), cultureID, action)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to process cultural interaction")
		return
	}

	// Get updated culture
	culture, _ := h.worldRepo.GetCulture(cultureID)

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Cultural interaction processed",
		"culture": culture,
	})
}

// GetCultureLanguage handles GET /api/cultures/{cultureId}/language
func (h *EmergentWorldHandlers) GetCultureLanguage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cultureID := vars["cultureId"]

	culture, err := h.worldRepo.GetCulture(cultureID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Culture not found")
		return
	}

	sendJSONResponse(w, http.StatusOK, culture.Language)
}

// GetCultureBeliefs handles GET /api/cultures/{cultureId}/beliefs
func (h *EmergentWorldHandlers) GetCultureBeliefs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cultureID := vars["cultureId"]

	culture, err := h.worldRepo.GetCulture(cultureID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Culture not found")
		return
	}

	sendJSONResponse(w, http.StatusOK, culture.BeliefSystem)
}

// GetCultureCustoms handles GET /api/cultures/{cultureId}/customs
func (h *EmergentWorldHandlers) GetCultureCustoms(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cultureID := vars["cultureId"]

	culture, err := h.worldRepo.GetCulture(cultureID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Culture not found")
		return
	}

	sendJSONResponse(w, http.StatusOK, culture.Customs)
}

// GetSimulationLogs handles GET /api/sessions/{sessionId}/world/logs
func (h *EmergentWorldHandlers) GetSimulationLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Only DM can view simulation logs
	userID := auth.GetUserIDFromContext(r.Context())
	session, err := h.gameService.GetByID(r.Context(), sessionID)
	if err != nil || session.DMUserID != userID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can view simulation logs")
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	logs, err := h.worldRepo.GetSimulationLogs(sessionID, limit)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve simulation logs")
		return
	}

	sendJSONResponse(w, http.StatusOK, logs)
}

// TriggerWorldEvent handles POST /api/sessions/{sessionId}/world/events
func (h *EmergentWorldHandlers) TriggerWorldEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Only DM can trigger world events
	userID := auth.GetUserIDFromContext(r.Context())
	session, err := h.gameService.GetByID(r.Context(), sessionID)
	if err != nil || session.DMUserID != userID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can trigger world events")
		return
	}

	var event models.WorldEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	event.SessionID = sessionID
	event.ID = generateUUID()

	if err := h.worldRepo.CreateWorldEvent(&event); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to create world event")
		return
	}

	sendJSONResponse(w, http.StatusCreated, event)
}