package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

// Error messages
const (
	errInvalidSettlementID = "Invalid settlement ID"
	errInvalidReqBody = "Invalid request body"
	errInvalidSessionID = "Invalid session ID"
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
	_, sessionID, ok := ExtractUserAndSessionID(w, r)
	if !ok {
		return
	}

	// TODO: Verify user is DM of this session

	var req models.SettlementGenerationRequest
	if !DecodeAndValidateJSON(w, r, &req) {
		return
	}

	HandleServiceOperation(w, r, func() (*models.Settlement, error) {
		return h.settlementGen.GenerateSettlement(r.Context(), sessionID, req)
	}, http.StatusCreated)
}

// GetSettlements retrieves all settlements for a game session
func (h *WorldBuildingHandlers) GetSettlements(w http.ResponseWriter, r *http.Request) {
	_, sessionID, ok := ExtractUserAndSessionID(w, r)
	if !ok {
		return
	}

	HandleServiceListOperation(w, r, func() ([]*models.Settlement, error) {
		return h.worldRepo.GetSettlementsByGameSession(sessionID)
	})
}

// GetSettlement retrieves a specific settlement
func (h *WorldBuildingHandlers) GetSettlement(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	settlementID, err := uuid.Parse(vars["settlementId"])
	if err != nil {
		response.BadRequest(w, r, errInvalidSettlementID)
		return
	}

	settlement, err := h.worldRepo.GetSettlement(settlementID)
	if err != nil {
		response.NotFound(w, r, "settlement")
		return
	}

	response.JSON(w, r, http.StatusOK, settlement)
}

// CalculateItemPrice calculates the price of an item in a settlement's market
func (h *WorldBuildingHandlers) CalculateItemPrice(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	settlementID, err := uuid.Parse(vars["settlementId"])
	if err != nil {
		response.BadRequest(w, r, errInvalidSettlementID)
		return
	}

	var req struct {
		BasePrice float64 `json:"basePrice"`
		ItemType  string  `json:"itemType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, errInvalidReqBody)
		return
	}

	price, err := h.economicSim.CalculateItemPrice(settlementID, req.BasePrice, req.ItemType)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	resp := map[string]interface{}{
		"basePrice":     req.BasePrice,
		"adjustedPrice": price,
		"itemType":      req.ItemType,
		"settlementId":  settlementID,
	}

	response.JSON(w, r, http.StatusOK, resp)
}

// Faction handlers

// CreateFaction handles faction creation requests
func (h *WorldBuildingHandlers) CreateFaction(w http.ResponseWriter, r *http.Request) {
	_, sessionID, ok := ExtractUserAndSessionID(w, r)
	if !ok {
		return
	}

	// TODO: Verify user is DM of this session

	var req models.FactionCreationRequest
	if !DecodeAndValidateJSON(w, r, &req) {
		return
	}

	HandleServiceOperation(w, r, func() (*models.Faction, error) {
		return h.factionSystem.CreateFaction(r.Context(), sessionID, req)
	}, http.StatusCreated)
}

// GetFactions retrieves all factions for a game session
func (h *WorldBuildingHandlers) GetFactions(w http.ResponseWriter, r *http.Request) {
	_, sessionID, ok := ExtractUserAndSessionID(w, r)
	if !ok {
		return
	}

	HandleServiceListOperation(w, r, func() ([]*models.Faction, error) {
		return h.worldRepo.GetFactionsByGameSession(sessionID)
	})
}

// UpdateFactionRelationship updates the relationship between two factions
func (h *WorldBuildingHandlers) UpdateFactionRelationship(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	faction1ID, err := uuid.Parse(vars["faction1Id"])
	if err != nil {
		response.BadRequest(w, r, "Invalid faction 1 ID")
		return
	}

	faction2ID, err := uuid.Parse(vars["faction2Id"])
	if err != nil {
		response.BadRequest(w, r, "Invalid faction 2 ID")
		return
	}

	// TODO: Verify user is DM and factions belong to their session

	var req struct {
		Change int    `json:"change"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, errInvalidReqBody)
		return
	}

	err = h.factionSystem.UpdateFactionRelationship(r.Context(), faction1ID, faction2ID, req.Change, req.Reason)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

// SimulateFactionConflicts triggers faction conflict simulation
func (h *WorldBuildingHandlers) SimulateFactionConflicts(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		response.BadRequest(w, r, errInvalidSessionID)
		return
	}

	// TODO: Verify user is DM of this session

	events, err := h.factionSystem.SimulateFactionConflicts(r.Context(), sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, events)
}

// World Event handlers

// CreateWorldEvent generates a new world event
func (h *WorldBuildingHandlers) CreateWorldEvent(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		response.BadRequest(w, r, errInvalidSessionID)
		return
	}

	// TODO: Verify user is DM of this session

	var req struct {
		EventType models.WorldEventType `json:"eventType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, errInvalidReqBody)
		return
	}

	event, err := h.worldEventEngine.GenerateWorldEvent(r.Context(), sessionID, req.EventType)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, event)
}

// GetActiveWorldEvents retrieves all active world events
func (h *WorldBuildingHandlers) GetActiveWorldEvents(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		response.BadRequest(w, r, errInvalidSessionID)
		return
	}

	events, err := h.worldRepo.GetActiveWorldEvents(sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, events)
}

// ProgressWorldEvents advances the world event simulation
func (h *WorldBuildingHandlers) ProgressWorldEvents(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		response.BadRequest(w, r, errInvalidSessionID)
		return
	}

	// TODO: Verify user is DM of this session

	err = h.worldEventEngine.SimulateEventProgression(r.Context(), sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "progressed"})
}

// Trade Route handlers

// CreateTradeRoute creates a new trade route between settlements
func (h *WorldBuildingHandlers) CreateTradeRoute(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	// TODO: Verify user is DM

	var req struct {
		StartSettlementID string `json:"startSettlementId"`
		EndSettlementID   string `json:"endSettlementId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, errInvalidReqBody)
		return
	}

	startID, err := uuid.Parse(req.StartSettlementID)
	if err != nil {
		response.BadRequest(w, r, "Invalid start settlement ID")
		return
	}

	endID, err := uuid.Parse(req.EndSettlementID)
	if err != nil {
		response.BadRequest(w, r, "Invalid end settlement ID")
		return
	}

	route, err := h.economicSim.CreateTradeRoute(r.Context(), startID, endID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, route)
}

// SimulateEconomics runs the economic simulation
func (h *WorldBuildingHandlers) SimulateEconomics(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		response.BadRequest(w, r, errInvalidSessionID)
		return
	}

	// TODO: Verify user is DM of this session

	err = h.economicSim.SimulateEconomicCycle(r.Context(), sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "simulated"})
}

// GetSettlementMarket retrieves market conditions for a settlement
func (h *WorldBuildingHandlers) GetSettlementMarket(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	settlementID, err := uuid.Parse(vars["settlementId"])
	if err != nil {
		response.BadRequest(w, r, errInvalidSettlementID)
		return
	}

	market, err := h.worldRepo.GetMarketBySettlement(settlementID)
	if err != nil {
		response.NotFound(w, r, "market")
		return
	}

	response.JSON(w, r, http.StatusOK, market)
}
