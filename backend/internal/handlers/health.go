package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
)

func HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "dnd-game-api",
	}); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}
