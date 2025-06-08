package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
)

// WorldBuildingHandlers handles world building related requests
type WorldBuildingHandlers struct {
	settlementGen    *services.SettlementGeneratorService
	factionSystem    *services.FactionSystemService
	worldEventEngine *services.WorldEventEngineService
	economicSim      *services.EconomicSimulatorService
	worldRepo        services.WorldBuildingRepository
}

// NewWorldBuildingHandlers creates a new world building handlers instance
func NewWorldBuildingHandlers(
	settlementGen *services.SettlementGeneratorService,
	factionSystem *services.FactionSystemService,
	worldEventEngine *services.WorldEventEngineService,
	economicSim *services.EconomicSimulatorService,
	worldRepo services.WorldBuildingRepository,
) *WorldBuildingHandlers {
	return &WorldBuildingHandlers{
		settlementGen:    settlementGen,
		factionSystem:    factionSystem,
		worldEventEngine: worldEventEngine,
		economicSim:      economicSim,
		worldRepo:        worldRepo,
	}
}

// Settlement handlers

// GenerateSettlement handles settlement generation requests
func (h *WorldBuildingHandlers) GenerateSettlement(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// TODO: Verify user is DM of this session

	var req models.SettlementGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	settlement, err := h.settlementGen.GenerateSettlement(r.Context(), sessionID, req)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to generate settlement")
		return
	}

	sendJSONResponse(w, http.StatusCreated, settlement)
}

// GetSettlements retrieves all settlements for a game session
func (h *WorldBuildingHandlers) GetSettlements(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	settlements, err := h.worldRepo.GetSettlementsByGameSession(sessionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to get settlements")
		return
	}

	sendJSONResponse(w, http.StatusOK, settlements)
}

// GetSettlement retrieves a specific settlement
func (h *WorldBuildingHandlers) GetSettlement(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	settlementID, err := uuid.Parse(vars["settlementId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid settlement ID")
		return
	}

	settlement, err := h.worldRepo.GetSettlement(settlementID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Settlement not found")
		return
	}

	sendJSONResponse(w, http.StatusOK, settlement)
}

// CalculateItemPrice calculates the price of an item in a settlement's market
func (h *WorldBuildingHandlers) CalculateItemPrice(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	settlementID, err := uuid.Parse(vars["settlementId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid settlement ID")
		return
	}

	var req struct {
		BasePrice float64 `json:"basePrice"`
		ItemType  string  `json:"itemType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	price, err := h.economicSim.CalculateItemPrice(settlementID, req.BasePrice, req.ItemType)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to calculate price")
		return
	}

	response := map[string]interface{}{
		"basePrice":     req.BasePrice,
		"adjustedPrice": price,
		"itemType":      req.ItemType,
		"settlementId":  settlementID,
	}

	sendJSONResponse(w, http.StatusOK, response)
}

// Faction handlers

// CreateFaction handles faction creation requests
func (h *WorldBuildingHandlers) CreateFaction(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// TODO: Verify user is DM of this session

	var req models.FactionCreationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	faction, err := h.factionSystem.CreateFaction(r.Context(), sessionID, req)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to create faction")
		return
	}

	sendJSONResponse(w, http.StatusCreated, faction)
}

// GetFactions retrieves all factions for a game session
func (h *WorldBuildingHandlers) GetFactions(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	factions, err := h.worldRepo.GetFactionsByGameSession(sessionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to get factions")
		return
	}

	sendJSONResponse(w, http.StatusOK, factions)
}

// UpdateFactionRelationship updates the relationship between two factions
func (h *WorldBuildingHandlers) UpdateFactionRelationship(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	faction1ID, err := uuid.Parse(vars["faction1Id"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid faction 1 ID")
		return
	}

	faction2ID, err := uuid.Parse(vars["faction2Id"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid faction 2 ID")
		return
	}

	// TODO: Verify user is DM and factions belong to their session

	var req struct {
		Change int    `json:"change"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.factionSystem.UpdateFactionRelationship(r.Context(), faction1ID, faction2ID, req.Change, req.Reason)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to update faction relationship")
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

// SimulateFactionConflicts triggers faction conflict simulation
func (h *WorldBuildingHandlers) SimulateFactionConflicts(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// TODO: Verify user is DM of this session

	events, err := h.factionSystem.SimulateFactionConflicts(r.Context(), sessionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to simulate faction conflicts")
		return
	}

	sendJSONResponse(w, http.StatusOK, events)
}

// World Event handlers

// CreateWorldEvent generates a new world event
func (h *WorldBuildingHandlers) CreateWorldEvent(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// TODO: Verify user is DM of this session

	var req struct {
		EventType models.WorldEventType `json:"eventType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	event, err := h.worldEventEngine.GenerateWorldEvent(r.Context(), sessionID, req.EventType)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to generate world event")
		return
	}

	sendJSONResponse(w, http.StatusCreated, event)
}

// GetActiveWorldEvents retrieves all active world events
func (h *WorldBuildingHandlers) GetActiveWorldEvents(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	events, err := h.worldRepo.GetActiveWorldEvents(sessionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to get world events")
		return
	}

	sendJSONResponse(w, http.StatusOK, events)
}

// ProgressWorldEvents advances the world event simulation
func (h *WorldBuildingHandlers) ProgressWorldEvents(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// TODO: Verify user is DM of this session

	err = h.worldEventEngine.SimulateEventProgression(r.Context(), sessionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to progress world events")
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{"status": "progressed"})
}

// Trade Route handlers

// CreateTradeRoute creates a new trade route between settlements
func (h *WorldBuildingHandlers) CreateTradeRoute(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// TODO: Verify user is DM

	var req struct {
		StartSettlementID string `json:"startSettlementId"`
		EndSettlementID   string `json:"endSettlementId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	startID, err := uuid.Parse(req.StartSettlementID)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid start settlement ID")
		return
	}

	endID, err := uuid.Parse(req.EndSettlementID)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid end settlement ID")
		return
	}

	route, err := h.economicSim.CreateTradeRoute(r.Context(), startID, endID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to create trade route")
		return
	}

	sendJSONResponse(w, http.StatusCreated, route)
}

// SimulateEconomics runs the economic simulation
func (h *WorldBuildingHandlers) SimulateEconomics(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// TODO: Verify user is DM of this session

	err = h.economicSim.SimulateEconomicCycle(r.Context(), sessionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to simulate economics")
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{"status": "simulated"})
}

// GetSettlementMarket retrieves market conditions for a settlement
func (h *WorldBuildingHandlers) GetSettlementMarket(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	settlementID, err := uuid.Parse(vars["settlementId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid settlement ID")
		return
	}

	market, err := h.worldRepo.GetMarketBySettlement(settlementID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Market not found")
		return
	}

	sendJSONResponse(w, http.StatusOK, market)
}