package handlers

import (
	"context"
	"net/http"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// validateUserSession validates that a user is authenticated and has access to a game session
// Returns an error if validation fails
func validateUserSession(w http.ResponseWriter, r *http.Request, gameService interface {
	ValidateUserInSession(ctx context.Context, sessionID, userID string) error
}, sessionID string) error {
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "Unauthorized")
		return http.ErrNotSupported // Using a sentinel error to indicate response was already sent
	}

	// Verify user has access to this game session
	if err := gameService.ValidateUserInSession(r.Context(), sessionID, userID); err != nil {
		response.Forbidden(w, r, "You don't have access to this game session")
		return err
	}

	return nil
}

// parseUUIDFromRequest parses a UUID from the request vars with the given key
// Sends a BadRequest response if parsing fails
func parseUUIDFromRequest(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, error) {
	vars := mux.Vars(r)
	idStr := vars[key]
	
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, r, "Invalid "+key)
		return uuid.Nil, err
	}
	
	return id, nil
}

// authenticateUser ensures the user is authenticated
// Sends an Unauthorized response if authentication fails
func authenticateUser(w http.ResponseWriter, r *http.Request) bool {
	_, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, r, "")
		return false
	}
	return true
}