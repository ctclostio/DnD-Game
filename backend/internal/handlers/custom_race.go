package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// CreateCustomRace handles creating a new custom race
func (h *Handlers) CreateCustomRace(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request
	var req models.CustomRaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create custom race
	customRace, err := h.services.CustomRaces.CreateCustomRace(r.Context(), userID, req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, customRace)
}

// GetCustomRace retrieves a custom race by ID
func (h *Handlers) GetCustomRace(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get custom race
	customRace, err := h.services.CustomRaces.GetCustomRace(r.Context(), raceID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Custom race not found")
		return
	}

	// Check if user has permission to view this race
	if customRace.CreatedBy != userID && !customRace.IsPublic && customRace.ApprovalStatus != models.ApprovalStatusApproved {
		// Check if user is DM
		user, err := h.services.Users.GetByID(r.Context(), userID.String())
		if err != nil || user.Role != "dm" {
			respondWithError(w, http.StatusForbidden, "You don't have permission to view this race")
			return
		}
	}

	respondWithJSON(w, http.StatusOK, customRace)
}

// GetUserCustomRaces retrieves all custom races created by a user
func (h *Handlers) GetUserCustomRaces(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get custom races
	races, err := h.services.CustomRaces.GetUserCustomRaces(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, races)
}

// GetPublicCustomRaces retrieves all approved public custom races
func (h *Handlers) GetPublicCustomRaces(w http.ResponseWriter, r *http.Request) {
	// Get public races
	races, err := h.services.CustomRaces.GetPublicCustomRaces(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, races)
}

// ApproveCustomRace approves a custom race (DM only)
func (h *Handlers) ApproveCustomRace(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request
	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Approve race
	if err := h.services.CustomRaces.ApproveCustomRace(r.Context(), raceID, userID, req.Notes); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

// RejectCustomRace rejects a custom race (DM only)
func (h *Handlers) RejectCustomRace(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request
	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Reject race
	if err := h.services.CustomRaces.RejectCustomRace(r.Context(), raceID, userID, req.Notes); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

// RequestRevisionCustomRace requests changes to a custom race (DM only)
func (h *Handlers) RequestRevisionCustomRace(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request
	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Request revision
	if err := h.services.CustomRaces.RequestRevision(r.Context(), raceID, userID, req.Notes); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "revision_needed"})
}

// MakeCustomRacePublic makes a custom race available to all players
func (h *Handlers) MakeCustomRacePublic(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Make race public
	if err := h.services.CustomRaces.MakePublic(r.Context(), raceID, userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "public"})
}

// GetPendingCustomRaces retrieves all custom races pending approval (DM only)
func (h *Handlers) GetPendingCustomRaces(w http.ResponseWriter, r *http.Request) {
	// Get pending races
	races, err := h.services.CustomRaces.GetPendingApproval(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, races)
}

// GetCustomRaceStats retrieves race stats for character creation
func (h *Handlers) GetCustomRaceStats(w http.ResponseWriter, r *http.Request) {
	// Get race ID from URL
	vars := mux.Vars(r)
	raceID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid race ID")
		return
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Validate user can use this race
	if err := h.services.CustomRaces.ValidateCustomRaceForCharacter(r.Context(), raceID, userID); err != nil {
		respondWithError(w, http.StatusForbidden, err.Error())
		return
	}

	// Get race stats
	stats, err := h.services.CustomRaces.GetCustomRaceStats(r.Context(), raceID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}