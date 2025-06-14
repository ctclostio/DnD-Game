package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

// DEPRECATED: These functions are kept for backward compatibility during migration.
// They now redirect to the new standardized response package.
// sendJSONResponse sends a JSON response (DEPRECATED - use response.JSON)
// Deprecated: Use response.JSON instead
//
//lint:ignore U1000 retained for backward compatibility.
func deprecatedSendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	// Create a minimal request with empty context for backward compatibility.
	r, _ := http.NewRequest("", "", nil)
	response.JSON(w, r, status, data)
}

// sendErrorResponse sends an error response (DEPRECATED - use response.Error)
// Deprecated: Use response.Error or response.ErrorWithCode instead
//
//lint:ignore U1000 retained for backward compatibility.
func deprecatedSendErrorResponse(w http.ResponseWriter, status int, message string) {
	// Create a minimal request with empty context for backward compatibility.
	r, _ := http.NewRequest("", "", nil)
	response.BadRequest(w, r, message)
}

// respondWithJSON sends a JSON response (DEPRECATED - use response.JSON)
// Deprecated: Use response.JSON instead
//
//lint:ignore U1000 retained for backward compatibility.
func deprecatedRespondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	// Create a minimal request with empty context for backward compatibility.
	r, _ := http.NewRequest("", "", nil)
	response.JSON(w, r, status, data)
}

// respondWithError sends an error response (DEPRECATED - use response.Error)
// Deprecated: Use response.Error or response.ErrorWithCode instead
//
//lint:ignore U1000 retained for backward compatibility.
func deprecatedRespondWithError(w http.ResponseWriter, status int, message string) {
	// Create a minimal request with empty context for backward compatibility.
	r, _ := http.NewRequest("", "", nil)

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

// parseJSON parses JSON request body (kept for compatibility).
//
//lint:ignore U1000 retained for backward compatibility.
func parseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
