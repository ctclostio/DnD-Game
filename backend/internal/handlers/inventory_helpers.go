package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
)

// sendJSONResponse sends a JSON response with proper error handling
func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// sendSuccessResponse sends a generic success response
func sendSuccessResponse(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": message}); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// decodeItemRequest decodes and validates an item request from the request body
type ItemRequest struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

func decodeItemRequest(r *http.Request) (*ItemRequest, error) {
	var req ItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	// Ensure quantity is at least 1
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	return &req, nil
}
