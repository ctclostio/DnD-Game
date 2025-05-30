package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
)

type CampaignHandler struct {
	campaignService  *services.CampaignService
	gameService      *services.GameSessionService
}

func NewCampaignHandler(campaignService *services.CampaignService, gameService *services.GameSessionService) *CampaignHandler {
	return &CampaignHandler{
		campaignService: campaignService,
		gameService:     gameService,
	}
}

// Story Arc Handlers

func (h *CampaignHandler) CreateStoryArc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
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

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can create story arcs")
		return
	}

	var req models.CreateStoryArcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	arc, err := h.campaignService.CreateStoryArc(ctx, sessionID, req)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusCreated, arc)
}

func (h *CampaignHandler) GenerateStoryArc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
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

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can generate story arcs")
		return
	}

	var req models.GenerateStoryArcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	arc, err := h.campaignService.GenerateStoryArc(ctx, sessionID, req)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusCreated, arc)
}

func (h *CampaignHandler) GetStoryArcs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	arcs, err := h.campaignService.GetStoryArcs(ctx, sessionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, arcs)
}

func (h *CampaignHandler) UpdateStoryArc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
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

	arcID, err := uuid.Parse(vars["arcId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid arc ID")
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can update story arcs")
		return
	}

	var req models.UpdateStoryArcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.campaignService.UpdateStoryArc(ctx, arcID, req); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{"message": "Story arc updated successfully"})
}

// Session Memory Handlers

func (h *CampaignHandler) CreateSessionMemory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
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

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can create session memories")
		return
	}

	var req models.CreateSessionMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	memory, err := h.campaignService.CreateSessionMemory(ctx, sessionID, req)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusCreated, memory)
}

func (h *CampaignHandler) GetSessionMemories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	memories, err := h.campaignService.GetSessionMemories(ctx, sessionID, limit)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, memories)
}

func (h *CampaignHandler) GenerateRecap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var req models.GenerateRecapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to 3 sessions if no body provided
		req.SessionCount = 3
	}

	recap, err := h.campaignService.GenerateRecap(ctx, sessionID, req.SessionCount)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, recap)
}

// Plot Thread Handlers

func (h *CampaignHandler) CreatePlotThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
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

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can create plot threads")
		return
	}

	var thread models.PlotThread
	if err := json.NewDecoder(r.Body).Decode(&thread); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.campaignService.CreatePlotThread(ctx, sessionID, &thread); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusCreated, thread)
}

func (h *CampaignHandler) GetPlotThreads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	activeOnly := r.URL.Query().Get("active") == "true"

	threads, err := h.campaignService.GetPlotThreads(ctx, sessionID, activeOnly)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, threads)
}

// Foreshadowing Handlers

func (h *CampaignHandler) GenerateForeshadowing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
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

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can generate foreshadowing")
		return
	}

	var req models.GenerateForeshadowingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	element, err := h.campaignService.GenerateForeshadowing(ctx, sessionID, req)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusCreated, element)
}

func (h *CampaignHandler) GetUnrevealedForeshadowing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
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

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can view foreshadowing")
		return
	}

	elements, err := h.campaignService.GetUnrevealedForeshadowing(ctx, sessionID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, elements)
}

func (h *CampaignHandler) RevealForeshadowing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	vars := mux.Vars(r)
	elementID, err := uuid.Parse(vars["elementId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid element ID")
		return
	}

	// Get session ID from query params
	sessionID, err := uuid.Parse(r.URL.Query().Get("sessionId"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can reveal foreshadowing")
		return
	}

	var req struct {
		SessionNumber int `json:"session_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.campaignService.RevealForeshadowing(ctx, elementID, req.SessionNumber); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{"message": "Foreshadowing revealed"})
}

// Timeline Handlers

func (h *CampaignHandler) AddTimelineEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
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

	// Verify user is in session
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Session not found")
		return
	}

	// Check if user is participant or DM
	isParticipant := false
	if session.DMID == claims.UserID {
		isParticipant = true
	} else {
		participants, err := h.gameService.GetSessionParticipants(ctx, sessionID.String())
		if err == nil {
			for _, p := range participants {
				if p.UserID == claims.UserID {
					isParticipant = true
					break
				}
			}
		}
	}

	if !isParticipant {
		sendErrorResponse(w, http.StatusForbidden, "User not in session")
		return
	}

	var event models.CampaignTimeline
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	event.GameSessionID = sessionID
	event.RealSessionDate = time.Now()

	if err := h.campaignService.AddTimelineEvent(ctx, &event); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusCreated, event)
}

func (h *CampaignHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Parse date range from query params
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var startDate, endDate time.Time
	if startStr != "" {
		startDate, _ = time.Parse(time.RFC3339, startStr)
	} else {
		startDate = time.Now().AddDate(-1, 0, 0) // Default to 1 year ago
	}

	if endStr != "" {
		endDate, _ = time.Parse(time.RFC3339, endStr)
	} else {
		endDate = time.Now().AddDate(0, 0, 1) // Default to tomorrow
	}

	events, err := h.campaignService.GetTimeline(ctx, sessionID, startDate, endDate)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, events)
}

// NPC Relationship Handlers

func (h *CampaignHandler) UpdateNPCRelationship(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
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

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.DMID != claims.UserID {
		sendErrorResponse(w, http.StatusForbidden, "Only the DM can update NPC relationships")
		return
	}

	var relationship models.NPCRelationship
	if err := json.NewDecoder(r.Body).Decode(&relationship); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	relationship.GameSessionID = sessionID

	if err := h.campaignService.UpdateNPCRelationship(ctx, &relationship); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, relationship)
}

func (h *CampaignHandler) GetNPCRelationships(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["sessionId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	npcID, err := uuid.Parse(vars["npcId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid NPC ID")
		return
	}

	relationships, err := h.campaignService.GetNPCRelationships(ctx, sessionID, npcID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusOK, relationships)
}