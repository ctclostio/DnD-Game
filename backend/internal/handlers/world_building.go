package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	userID, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// TODO: Verify user is DM of this session

	var req models.SettlementGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	settlement, err := h.settlementGen.GenerateSettlement(r.Context(), sessionID, req)
	if err != nil {
		sendErrorResponse(w, "Failed to generate settlement", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, settlement, http.StatusCreated)
}

// GetSettlements retrieves all settlements for a game session
func (h *WorldBuildingHandlers) GetSettlements(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	settlements, err := h.worldRepo.GetSettlementsByGameSession(sessionID)
	if err != nil {
		sendErrorResponse(w, "Failed to get settlements", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, settlements, http.StatusOK)
}

// GetSettlement retrieves a specific settlement
func (h *WorldBuildingHandlers) GetSettlement(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	settlementID, err := uuid.Parse(vars["settlementId"])
	if err != nil {
		sendErrorResponse(w, "Invalid settlement ID", http.StatusBadRequest)
		return
	}

	settlement, err := h.worldRepo.GetSettlement(settlementID)
	if err != nil {
		sendErrorResponse(w, "Settlement not found", http.StatusNotFound)
		return
	}

	sendJSONResponse(w, settlement, http.StatusOK)
}

// CalculateItemPrice calculates the price of an item in a settlement's market
func (h *WorldBuildingHandlers) CalculateItemPrice(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	settlementID, err := uuid.Parse(vars["settlementId"])
	if err != nil {
		sendErrorResponse(w, "Invalid settlement ID", http.StatusBadRequest)
		return
	}

	var req struct {
		BasePrice float64 `json:"basePrice"`
		ItemType  string  `json:"itemType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	price, err := h.economicSim.CalculateItemPrice(settlementID, req.BasePrice, req.ItemType)
	if err != nil {
		sendErrorResponse(w, "Failed to calculate price", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"basePrice":     req.BasePrice,
		"adjustedPrice": price,
		"itemType":      req.ItemType,
		"settlementId":  settlementID,
	}

	sendJSONResponse(w, response, http.StatusOK)
}

// Faction handlers

// CreateFaction handles faction creation requests
func (h *WorldBuildingHandlers) CreateFaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// TODO: Verify user is DM of this session

	var req models.FactionCreationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	faction, err := h.factionSystem.CreateFaction(r.Context(), sessionID, req)
	if err != nil {
		sendErrorResponse(w, "Failed to create faction", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, faction, http.StatusCreated)
}

// GetFactions retrieves all factions for a game session
func (h *WorldBuildingHandlers) GetFactions(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	factions, err := h.worldRepo.GetFactionsByGameSession(sessionID)
	if err != nil {
		sendErrorResponse(w, "Failed to get factions", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, factions, http.StatusOK)
}

// UpdateFactionRelationship updates the relationship between two factions
func (h *WorldBuildingHandlers) UpdateFactionRelationship(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	faction1ID, err := uuid.Parse(vars["faction1Id"])
	if err != nil {
		sendErrorResponse(w, "Invalid faction 1 ID", http.StatusBadRequest)
		return
	}

	faction2ID, err := uuid.Parse(vars["faction2Id"])
	if err != nil {
		sendErrorResponse(w, "Invalid faction 2 ID", http.StatusBadRequest)
		return
	}

	// TODO: Verify user is DM and factions belong to their session

	var req struct {
		Change int    `json:"change"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.factionSystem.UpdateFactionRelationship(r.Context(), faction1ID, faction2ID, req.Change, req.Reason)
	if err != nil {
		sendErrorResponse(w, "Failed to update faction relationship", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]string{"status": "updated"}, http.StatusOK)
}

// SimulateFactionConflicts triggers faction conflict simulation
func (h *WorldBuildingHandlers) SimulateFactionConflicts(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// TODO: Verify user is DM of this session

	events, err := h.factionSystem.SimulateFactionConflicts(r.Context(), sessionID)
	if err != nil {
		sendErrorResponse(w, "Failed to simulate faction conflicts", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, events, http.StatusOK)
}

// World Event handlers

// CreateWorldEvent generates a new world event
func (h *WorldBuildingHandlers) CreateWorldEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// TODO: Verify user is DM of this session

	var req struct {
		EventType models.WorldEventType `json:"eventType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	event, err := h.worldEventEngine.GenerateWorldEvent(r.Context(), sessionID, req.EventType)
	if err != nil {
		sendErrorResponse(w, "Failed to generate world event", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, event, http.StatusCreated)
}

// GetActiveWorldEvents retrieves all active world events
func (h *WorldBuildingHandlers) GetActiveWorldEvents(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	events, err := h.worldRepo.GetActiveWorldEvents(sessionID)
	if err != nil {
		sendErrorResponse(w, "Failed to get world events", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, events, http.StatusOK)
}

// ProgressWorldEvents advances the world event simulation
func (h *WorldBuildingHandlers) ProgressWorldEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// TODO: Verify user is DM of this session

	err = h.worldEventEngine.SimulateEventProgression(r.Context(), sessionID)
	if err != nil {
		sendErrorResponse(w, "Failed to progress world events", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]string{"status": "progressed"}, http.StatusOK)
}

// Trade Route handlers

// CreateTradeRoute creates a new trade route between settlements
func (h *WorldBuildingHandlers) CreateTradeRoute(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Verify user is DM

	var req struct {
		StartSettlementID string `json:"startSettlementId"`
		EndSettlementID   string `json:"endSettlementId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	startID, err := uuid.Parse(req.StartSettlementID)
	if err != nil {
		sendErrorResponse(w, "Invalid start settlement ID", http.StatusBadRequest)
		return
	}

	endID, err := uuid.Parse(req.EndSettlementID)
	if err != nil {
		sendErrorResponse(w, "Invalid end settlement ID", http.StatusBadRequest)
		return
	}

	route, err := h.economicSim.CreateTradeRoute(r.Context(), startID, endID)
	if err != nil {
		sendErrorResponse(w, "Failed to create trade route", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, route, http.StatusCreated)
}

// SimulateEconomics runs the economic simulation
func (h *WorldBuildingHandlers) SimulateEconomics(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// TODO: Verify user is DM of this session

	err = h.economicSim.SimulateEconomicCycle(r.Context(), sessionID)
	if err != nil {
		sendErrorResponse(w, "Failed to simulate economics", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]string{"status": "simulated"}, http.StatusOK)
}

// GetSettlementMarket retrieves market conditions for a settlement
func (h *WorldBuildingHandlers) GetSettlementMarket(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	settlementID, err := uuid.Parse(vars["settlementId"])
	if err != nil {
		sendErrorResponse(w, "Invalid settlement ID", http.StatusBadRequest)
		return
	}

	market, err := h.worldRepo.GetMarketBySettlement(settlementID)
	if err != nil {
		sendErrorResponse(w, "Market not found", http.StatusNotFound)
		return
	}

	sendJSONResponse(w, market, http.StatusOK)
}