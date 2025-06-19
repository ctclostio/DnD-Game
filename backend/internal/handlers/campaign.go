package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/pkg/errors"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

type CampaignHandler struct {
	campaignService *services.CampaignService
	gameService     *services.GameSessionService
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
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		response.NotFound(w, r, constants.ErrSessionNotFound)
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can create story arcs")
		return
	}

	var req models.CreateStoryArcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	arc, err := h.campaignService.CreateStoryArc(ctx, sessionID, req)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, arc)
}

func (h *CampaignHandler) GenerateStoryArc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		response.NotFound(w, r, constants.ErrSessionNotFound)
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can generate story arcs")
		return
	}

	var req models.GenerateStoryArcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	arc, err := h.campaignService.GenerateStoryArc(ctx, sessionID, req)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, arc)
}

func (h *CampaignHandler) GetStoryArcs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	arcs, err := h.campaignService.GetStoryArcs(ctx, sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, arcs)
}

func (h *CampaignHandler) UpdateStoryArc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	arcID, err := uuid.Parse(vars[constants.ParamArcID])
	if err != nil {
		response.BadRequest(w, r, "Invalid arc ID")
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		response.NotFound(w, r, constants.ErrSessionNotFound)
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can update story arcs")
		return
	}

	var req models.UpdateStoryArcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	if err := h.campaignService.UpdateStoryArc(ctx, arcID, req); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"message": "Story arc updated successfully"})
}

// Session Memory Handlers

func (h *CampaignHandler) CreateSessionMemory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		response.NotFound(w, r, constants.ErrSessionNotFound)
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can create session memories")
		return
	}

	var req models.CreateSessionMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	memory, err := h.campaignService.CreateSessionMemory(ctx, sessionID, &req)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, memory)
}

func (h *CampaignHandler) GetSessionMemories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
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
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, memories)
}

func (h *CampaignHandler) GenerateRecap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	var req models.GenerateRecapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to 3 sessions if no body provided
		req.SessionCount = 3
	}

	recap, err := h.campaignService.GenerateRecap(ctx, sessionID, req.SessionCount)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, recap)
}

// Plot Thread Handlers

func (h *CampaignHandler) CreatePlotThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		response.NotFound(w, r, constants.ErrSessionNotFound)
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can create plot threads")
		return
	}

	var thread models.PlotThread
	if err := json.NewDecoder(r.Body).Decode(&thread); err != nil {
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	if err := h.campaignService.CreatePlotThread(ctx, sessionID, &thread); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, thread)
}

func (h *CampaignHandler) GetPlotThreads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	activeOnly := r.URL.Query().Get("active") == queryParamTrue

	threads, err := h.campaignService.GetPlotThreads(ctx, sessionID, activeOnly)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, threads)
}

// Foreshadowing Handlers

func (h *CampaignHandler) GenerateForeshadowing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		response.NotFound(w, r, constants.ErrSessionNotFound)
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can generate foreshadowing")
		return
	}

	var req models.GenerateForeshadowingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	element, err := h.campaignService.GenerateForeshadowing(ctx, sessionID, req)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, element)
}

func (h *CampaignHandler) GetUnrevealedForeshadowing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		response.NotFound(w, r, constants.ErrSessionNotFound)
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can view foreshadowing")
		return
	}

	elements, err := h.campaignService.GetUnrevealedForeshadowing(ctx, sessionID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, elements)
}

func (h *CampaignHandler) RevealForeshadowing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	elementID, err := uuid.Parse(vars[constants.ParamElementID])
	if err != nil {
		response.BadRequest(w, r, "Invalid element ID")
		return
	}

	// Get session ID from query params
	sessionID, err := uuid.Parse(r.URL.Query().Get(constants.ParamSessionID))
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		response.NotFound(w, r, constants.ErrSessionNotFound)
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can reveal foreshadowing")
		return
	}

	var req struct {
		SessionNumber int `json:"session_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	if err := h.campaignService.RevealForeshadowing(ctx, elementID, req.SessionNumber); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"message": "Foreshadowing revealed"})
}

// Timeline Handlers

func (h *CampaignHandler) AddTimelineEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	sessionID, err := h.parseSessionID(r)
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	// Verify user is in session
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		response.NotFound(w, r, constants.ErrSessionNotFound)
		return
	}

	// Check if user is participant or DM
	if !h.isUserInSession(ctx, session, claims.UserID) {
		response.Forbidden(w, r, "User not in session")
		return
	}

	var event models.CampaignTimeline
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	event.GameSessionID = sessionID
	event.RealSessionDate = time.Now()

	if err := h.campaignService.AddTimelineEvent(ctx, &event); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, event)
}

func (h *CampaignHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	// Parse date range from query params
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var startDate, endDate time.Time
	if startStr != "" {
		parsedStart, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			response.ErrorWithCode(w, r, errors.ErrCodeInvalidFormat, "invalid start date format")
			return
		}
		startDate = parsedStart
	} else {
		startDate = time.Now().AddDate(-1, 0, 0) // Default to 1 year ago
	}

	if endStr != "" {
		parsedEnd, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			response.ErrorWithCode(w, r, errors.ErrCodeInvalidFormat, "invalid end date format")
			return
		}
		endDate = parsedEnd
	} else {
		endDate = time.Now().AddDate(0, 0, 1) // Default to tomorrow
	}

	events, err := h.campaignService.GetTimeline(ctx, sessionID, startDate, endDate)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, events)
}

// NPC Relationship Handlers

func (h *CampaignHandler) UpdateNPCRelationship(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	// Verify user is DM
	session, err := h.gameService.GetSessionByID(ctx, sessionID.String())
	if err != nil {
		response.NotFound(w, r, constants.ErrSessionNotFound)
		return
	}

	if session.DMID != claims.UserID {
		response.Forbidden(w, r, "Only the DM can update NPC relationships")
		return
	}

	var relationship models.NPCRelationship
	if err := json.NewDecoder(r.Body).Decode(&relationship); err != nil {
		response.BadRequest(w, r, constants.ErrInvalidRequestBody)
		return
	}

	relationship.GameSessionID = sessionID

	if err := h.campaignService.UpdateNPCRelationship(ctx, &relationship); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, relationship)
}

func (h *CampaignHandler) GetNPCRelationships(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return
	}

	npcID, err := uuid.Parse(vars[constants.ParamNPCID])
	if err != nil {
		response.BadRequest(w, r, "Invalid NPC ID")
		return
	}

	relationships, err := h.campaignService.GetNPCRelationships(ctx, sessionID, npcID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, relationships)
}

// Helper functions to reduce cognitive complexity

// parseSessionID extracts and validates the session ID from request
func (h *CampaignHandler) parseSessionID(r *http.Request) (uuid.UUID, error) {
	vars := mux.Vars(r)
	return uuid.Parse(vars[constants.ParamSessionID])
}

// isUserInSession checks if the user is either the DM or a participant in the session
func (h *CampaignHandler) isUserInSession(ctx context.Context, session *models.GameSession, userID string) bool {
	// Check if user is DM
	if session.DMID == userID {
		return true
	}

	// Check if user is participant
	participants, err := h.gameService.GetSessionParticipants(ctx, session.ID)
	if err != nil {
		return false
	}

	for _, p := range participants {
		if p.UserID == userID {
			return true
		}
	}

	return false
}
