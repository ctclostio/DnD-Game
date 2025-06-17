package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

// ExtractUserAndSessionID extracts user ID and session ID from request
func ExtractUserAndSessionID(w http.ResponseWriter, r *http.Request) (userID string, sessionID uuid.UUID, ok bool) {
	user, userOk := auth.GetUserFromContext(r.Context())
	if !userOk {
		response.Unauthorized(w, r, constants.ErrUnauthorized)
		return "", uuid.Nil, false
	}

	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars[constants.ParamSessionID])
	if err != nil {
		response.BadRequest(w, r, constants.ErrInvalidSessionID)
		return "", uuid.Nil, false
	}

	return user.ID, sessionID, true
}

// ExtractUserAndID extracts user ID and a generic UUID from request
func ExtractUserAndID(w http.ResponseWriter, r *http.Request, idParam string) (userID string, id uuid.UUID, ok bool) {
	userIDStr, userOk := auth.GetUserIDFromContext(r.Context())
	if !userOk {
		response.Unauthorized(w, r, "")
		return "", uuid.Nil, false
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars[idParam])
	if err != nil {
		response.BadRequest(w, r, "Invalid "+idParam)
		return "", uuid.Nil, false
	}

	return userIDStr, id, true
}

// DecodeAndValidateJSON decodes JSON request body into target struct
func DecodeAndValidateJSON(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return false
	}
	return true
}

// HandleServiceOperation executes a service operation and handles response
func HandleServiceOperation[T any](w http.ResponseWriter, r *http.Request, 
	operation func() (*T, error), 
	successStatus int) {
	
	result, err := operation()
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, successStatus, result)
}

// HandleServiceListOperation executes a service list operation and handles response
func HandleServiceListOperation[T any](w http.ResponseWriter, r *http.Request, 
	operation func() ([]*T, error)) {
	
	results, err := operation()
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, results)
}

// VerifyCharacterOwnership verifies that the user owns the character
func VerifyCharacterOwnership(w http.ResponseWriter, r *http.Request, 
	characterRepo database.CharacterRepository, 
	characterID, userID string) (*models.Character, bool) {
	
	character, err := characterRepo.GetByID(r.Context(), characterID)
	if err != nil {
		http.Error(w, "Character not found", http.StatusNotFound)
		return nil, false
	}

	if character.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return nil, false
	}

	return character, true
}

// HandleCharacterOwnedCreation handles creation of resources that require character ownership verification
func HandleCharacterOwnedCreation[T any](w http.ResponseWriter, r *http.Request, 
	characterRepo database.CharacterRepository,
	getCharacterID func(*T) string,
	createFunc func(*T) error) {
	
	claims := auth.GetClaimsFromContext(r.Context())
	
	var entity T
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Verify character ownership
	characterID := getCharacterID(&entity)
	if _, ok := VerifyCharacterOwnership(w, r, characterRepo, characterID, claims.UserID); !ok {
		return
	}
	
	if err := createFunc(&entity); err != nil {
		http.Error(w, "Failed to create resource", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(entity); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}