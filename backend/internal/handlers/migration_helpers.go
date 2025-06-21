package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

// DEPRECATED: These functions are kept for backward compatibility during migration
// They now redirect to the new standardized response package

// sendJSONResponse sends a JSON response (DEPRECATED - use response.JSON)
// Deprecated: Use response.JSON instead
//
//lint:ignore U1000 retained for backward compatibility
func deprecatedSendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	// Create a minimal request with empty context for backward compatibility
	r, _ := http.NewRequest("", "", http.NoBody)
	response.JSON(w, r, status, data)
}

// sendErrorResponse sends an error response (DEPRECATED - use response.Error)
// Deprecated: Use response.Error or response.ErrorWithCode instead
//
//lint:ignore U1000 retained for backward compatibility
func deprecatedSendErrorResponse(w http.ResponseWriter, _ int, message string) {
	// Create a minimal request with empty context for backward compatibility
	r, _ := http.NewRequest("", "", http.NoBody)
	response.BadRequest(w, r, message)
}

// respondWithJSON sends a JSON response (DEPRECATED - use response.JSON)
// Deprecated: Use response.JSON instead
// This variant was used by older handler implementations
//
//lint:ignore U1000 retained for backward compatibility
func deprecatedRespondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	// Create a minimal request with empty context for backward compatibility
	// This function wraps response.JSON for legacy code that used respondWithJSON naming
	r, _ := http.NewRequest("GET", "/", http.NoBody)
	response.JSON(w, r, status, data)
}

// respondWithError sends an error response (DEPRECATED - use response.Error)
// Deprecated: Use response.Error or response.ErrorWithCode instead
//
//lint:ignore U1000 retained for backward compatibility
func deprecatedRespondWithError(w http.ResponseWriter, status int, message string) {
	// Create a minimal request with empty context for backward compatibility
	r, _ := http.NewRequest("", "", http.NoBody)

	switch status {
	case http.StatusBadRequest:
		response.BadRequest(w, r, message)
	case http.StatusUnauthorized:
		response.Unauthorized(w, r, message)
	case http.StatusForbidden:
		response.Forbidden(w, r, message)
	case http.StatusNotFound:
		response.NotFound(w, r, message)
	default:
		response.BadRequest(w, r, message)
	}
}

// parseJSON parses JSON request body (kept for compatibility)
//
//lint:ignore U1000 retained for backward compatibility
func parseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
