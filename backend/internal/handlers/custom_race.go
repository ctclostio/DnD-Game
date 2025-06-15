package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/errors"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

// CreateCustomRace handles creating a new custom race
func (h *Handlers) CreateCustomRace(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Parse request
	var req models.CustomRaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Convert userID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, r, "Invalid user ID")
		return
	}

	// Create custom race
	customRace, err := h.customRaceService.CreateCustomRace(r.Context(), userUUID, req)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, customRace)
}

// GetCustomRace retrieves a custom race by ID
func (h *Handlers) GetCustomRace(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.BadRequest(w, r, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Get custom race
	customRace, err := h.customRaceService.GetCustomRace(r.Context(), raceID)
	if err != nil {
		response.NotFound(w, r, "Custom race")
		return
	}

	// Convert userID to UUID for comparison
	userUUID, _ := uuid.Parse(userID)

	// Check if user has permission to view this race
	if customRace.CreatedBy != userUUID && !customRace.IsPublic && customRace.ApprovalStatus != models.ApprovalStatusApproved {
		// Check if user is DM
		user, err := h.userService.GetByID(r.Context(), userID)
		if err != nil || user.Role != "dm" {
			response.ErrorWithCode(w, r, errors.ErrCodeInsufficientPrivilege, "You don't have permission to view this race")
			return
		}
	}

	response.JSON(w, r, http.StatusOK, customRace)
}

// GetUserCustomRaces retrieves all custom races created by a user
func (h *Handlers) GetUserCustomRaces(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Convert userID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, r, "Invalid user ID")
		return
	}

	// Get custom races
	races, err := h.customRaceService.GetUserCustomRaces(r.Context(), userUUID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, races)
}

// GetPublicCustomRaces retrieves all approved public custom races
func (h *Handlers) GetPublicCustomRaces(w http.ResponseWriter, r *http.Request) {
	// Get public races
	races, err := h.customRaceService.GetPublicCustomRaces(r.Context())
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, races)
}

// ApproveCustomRace approves a custom race (DM only)
func (h *Handlers) ApproveCustomRace(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.BadRequest(w, r, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Parse request
	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Convert userID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, r, "Invalid user ID")
		return
	}

	// Approve race
	if err := h.customRaceService.ApproveCustomRace(r.Context(), raceID, userUUID, req.Notes); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "approved"})
}

// RejectCustomRace rejects a custom race (DM only)
func (h *Handlers) RejectCustomRace(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.BadRequest(w, r, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Parse request
	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Convert userID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, r, "Invalid user ID")
		return
	}

	// Reject race
	if err := h.customRaceService.RejectCustomRace(r.Context(), raceID, userUUID, req.Notes); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "rejected"})
}

// RequestRevisionCustomRace requests changes to a custom race (DM only)
func (h *Handlers) RequestRevisionCustomRace(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.BadRequest(w, r, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Parse request
	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Convert userID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, r, "Invalid user ID")
		return
	}

	// Request revision
	if err := h.customRaceService.RequestRevision(r.Context(), raceID, userUUID, req.Notes); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "revision_needed"})
}

// MakeCustomRacePublic makes a custom race available to all players
func (h *Handlers) MakeCustomRacePublic(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.BadRequest(w, r, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Convert userID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, r, "Invalid user ID")
		return
	}

	// Make race public
	if err := h.customRaceService.MakePublic(r.Context(), raceID, userUUID); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, map[string]string{"status": "public"})
}

// GetPendingCustomRaces retrieves all custom races pending approval (DM only)
func (h *Handlers) GetPendingCustomRaces(w http.ResponseWriter, r *http.Request) {
	// Get pending races
	races, err := h.customRaceService.GetPendingApproval(r.Context())
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, races)
}

// GetCustomRaceStats retrieves race stats for character creation
func (h *Handlers) GetCustomRaceStats(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.BadRequest(w, r, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return
	}

	// Convert userID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, r, "Invalid user ID")
		return
	}

	// Validate user can use this race
	if err := h.customRaceService.ValidateCustomRaceForCharacter(r.Context(), raceID, userUUID); err != nil {
		response.Forbidden(w, r, err.Error())
		return
	}

	// Get race stats
	stats, err := h.customRaceService.GetCustomRaceStats(r.Context(), raceID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, stats)
}
