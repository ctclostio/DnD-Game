package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/pkg/response"
)

// EmergentWorldHandlers manages handlers for the emergent world system
type EmergentWorldHandlers struct {
	livingEcosystem    *services.LivingEcosystemService
	factionPersonality *services.FactionPersonalityService
	proceduralCulture  *services.ProceduralCultureService
	worldRepo          *database.EmergentWorldRepository
	gameService        *services.GameSessionService
	factionRepo        *database.WorldBuildingRepository
}

// NewEmergentWorldHandlers creates new emergent world handlers
func NewEmergentWorldHandlers(
	livingEcosystem *services.LivingEcosystemService,
	factionPersonality *services.FactionPersonalityService,
	proceduralCulture *services.ProceduralCultureService,
	worldRepo *database.EmergentWorldRepository,
	gameService *services.GameSessionService,
	factionRepo *database.WorldBuildingRepository,
) *EmergentWorldHandlers {
	return &EmergentWorldHandlers{
		livingEcosystem:    livingEcosystem,
		factionPersonality: factionPersonality,
		proceduralCulture:  proceduralCulture,
		worldRepo:          worldRepo,
		gameService:        gameService,
		factionRepo:        factionRepo,
	}
}

// SimulateWorld handles POST /api/sessions/{sessionId}/world/simulate
func (h *EmergentWorldHandlers) SimulateWorld(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Verify user is DM
	userID, _ := auth.GetUserIDFromContext(r.Context())
	session, err := h.gameService.GetSessionByID(r.Context(), sessionID)
	if err != nil || session.DMID != userID {
		response.Forbidden(w, r, "Only the DM can simulate world progress")
		return
	}

	// Run simulation
	err = h.livingEcosystem.SimulateWorldProgress(r.Context(), sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Get recent events
	events, err := h.worldRepo.GetWorldEvents(sessionID, 10, false)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]interface{}{
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
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, events)
}

// GetWorldState handles GET /api/sessions/{sessionId}/world/state
func (h *EmergentWorldHandlers) GetWorldState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	state, err := h.worldRepo.GetWorldState(sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, state)
}

// NPC Autonomy Handlers

// GetNPCGoals handles GET /api/npcs/{npcId}/goals
func (h *EmergentWorldHandlers) GetNPCGoals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	npcID := vars["npcId"]

	goals, err := h.worldRepo.GetNPCGoals(npcID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, goals)
}

// CreateNPCGoal handles POST /api/npcs/{npcId}/goals
func (h *EmergentWorldHandlers) CreateNPCGoal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	npcID := vars["npcId"]

	var goal models.NPCGoal
	if err := json.NewDecoder(r.Body).Decode(&goal); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	goal.NPCID = npcID
	goal.ID = uuid.New().String()

	if err := h.worldRepo.CreateNPCGoal(&goal); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, goal)
}

// GetNPCSchedule handles GET /api/npcs/{npcId}/schedule
func (h *EmergentWorldHandlers) GetNPCSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	npcID := vars["npcId"]

	schedule, err := h.worldRepo.GetNPCSchedule(npcID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, schedule)
}

// Faction Personality Handlers

// InitializeFactionPersonality handles POST /api/factions/{factionId}/personality
func (h *EmergentWorldHandlers) InitializeFactionPersonality(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	factionID := vars["factionId"]

	// Get faction
	factionUUID, _ := uuid.Parse(factionID)
	faction, err := h.factionRepo.GetFaction(factionUUID)
	if err != nil {
		response.NotFound(w, r, "Faction not found")
		return
	}

	// Initialize personality
	personality, err := h.factionPersonality.InitializeFactionPersonality(r.Context(), faction)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, personality)
}

// GetFactionPersonality handles GET /api/factions/{factionId}/personality
func (h *EmergentWorldHandlers) GetFactionPersonality(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	factionID := vars["factionId"]

	personality, err := h.worldRepo.GetFactionPersonality(factionID)
	if err != nil {
		response.NotFound(w, r, "Faction personality not found")
		return
	}

	response.JSON(w, r, http.StatusOK, personality)
}

// MakeFactionDecision handles POST /api/factions/{factionId}/decide
func (h *EmergentWorldHandlers) MakeFactionDecision(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	factionID := vars["factionId"]

	var decision models.FactionDecision
	if err := json.NewDecoder(r.Body).Decode(&decision); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	result, err := h.factionPersonality.MakeFactionDecision(r.Context(), factionID, decision)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, result)
}

// GetFactionAgendas handles GET /api/factions/{factionId}/agendas
func (h *EmergentWorldHandlers) GetFactionAgendas(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	factionID := vars["factionId"]

	agendas, err := h.worldRepo.GetFactionAgendas(factionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, agendas)
}

// RecordFactionInteraction handles POST /api/factions/{factionId}/interactions
func (h *EmergentWorldHandlers) RecordFactionInteraction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	factionID := vars["factionId"]

	var interaction models.PlayerInteraction
	if err := json.NewDecoder(r.Body).Decode(&interaction); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	err := h.factionPersonality.LearnFromInteraction(r.Context(), factionID, interaction)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{
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
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	culture, err := h.proceduralCulture.GenerateCulture(r.Context(), sessionID, params)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, culture)
}

// GetCultures handles GET /api/sessions/{sessionId}/cultures
func (h *EmergentWorldHandlers) GetCultures(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	cultures, err := h.worldRepo.GetCulturesBySession(sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, cultures)
}

// GetCulture handles GET /api/cultures/{cultureId}
func (h *EmergentWorldHandlers) GetCulture(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cultureID := vars["cultureId"]

	culture, err := h.worldRepo.GetCulture(cultureID)
	if err != nil {
		response.NotFound(w, r, "Culture not found")
		return
	}

	response.JSON(w, r, http.StatusOK, culture)
}

// InteractWithCulture handles POST /api/cultures/{cultureId}/interact
func (h *EmergentWorldHandlers) InteractWithCulture(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cultureID := vars["cultureId"]

	var action services.PlayerCulturalAction
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	err := h.proceduralCulture.RespondToPlayerAction(r.Context(), cultureID, action)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Get updated culture
	culture, _ := h.worldRepo.GetCulture(cultureID)

	response.JSON(w, r, http.StatusOK, map[string]interface{}{
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
		response.NotFound(w, r, "Culture not found")
		return
	}

	response.JSON(w, r, http.StatusOK, culture.Language)
}

// GetCultureBeliefs handles GET /api/cultures/{cultureId}/beliefs
func (h *EmergentWorldHandlers) GetCultureBeliefs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cultureID := vars["cultureId"]

	culture, err := h.worldRepo.GetCulture(cultureID)
	if err != nil {
		response.NotFound(w, r, "Culture not found")
		return
	}

	response.JSON(w, r, http.StatusOK, culture.BeliefSystem)
}

// GetCultureCustoms handles GET /api/cultures/{cultureId}/customs
func (h *EmergentWorldHandlers) GetCultureCustoms(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cultureID := vars["cultureId"]

	culture, err := h.worldRepo.GetCulture(cultureID)
	if err != nil {
		response.NotFound(w, r, "Culture not found")
		return
	}

	response.JSON(w, r, http.StatusOK, culture.Customs)
}

// GetSimulationLogs handles GET /api/sessions/{sessionId}/world/logs
func (h *EmergentWorldHandlers) GetSimulationLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Only DM can view simulation logs
	userID, _ := auth.GetUserIDFromContext(r.Context())
	session, err := h.gameService.GetSessionByID(r.Context(), sessionID)
	if err != nil || session.DMID != userID {
		response.Forbidden(w, r, "Only the DM can view simulation logs")
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
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, logs)
}

// TriggerWorldEvent handles POST /api/sessions/{sessionId}/world/events
func (h *EmergentWorldHandlers) TriggerWorldEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Only DM can trigger world events
	userID, _ := auth.GetUserIDFromContext(r.Context())
	session, err := h.gameService.GetSessionByID(r.Context(), sessionID)
	if err != nil || session.DMID != userID {
		response.Forbidden(w, r, "Only the DM can trigger world events")
		return
	}

	var event models.WorldEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	parsedSessionID, err := uuid.Parse(sessionID)
	if err != nil {
		response.BadRequest(w, r, "Invalid session ID format")
		return
	}
	event.GameSessionID = parsedSessionID
	event.ID = uuid.New()

	// Convert WorldEvent to EmergentWorldEvent for creation
	emergentEvent := &models.EmergentWorldEvent{
		ID:               event.ID.String(),
		SessionID:        sessionID,
		EventType:        string(event.Type),
		Title:            event.Name,
		Description:      event.Description,
		Impact:           make(map[string]interface{}),
		AffectedEntities: []string{},
		IsPlayerVisible:  true,
		OccurredAt:       event.CreatedAt,
	}

	if err := h.worldRepo.CreateWorldEvent(emergentEvent); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, event)
}
