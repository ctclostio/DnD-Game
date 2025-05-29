package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/websocket"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	userService         *services.UserService
	characterService    *services.CharacterService
	gameService         *services.GameSessionService
	diceService         *services.DiceRollService
	combatService       *services.CombatService
	npcService          *services.NPCService
	inventoryService    *services.InventoryService
	encounterService    *services.EncounterService
	jwtManager          *auth.JWTManager
	refreshTokenService *services.RefreshTokenService
	websocketHub        *websocket.Hub
}

// NewHandlers creates a new handlers instance
func NewHandlers(svc *services.Services, hub *websocket.Hub) *Handlers {
	return &Handlers{
		userService:         svc.Users,
		characterService:    svc.Characters,
		gameService:         svc.GameSessions,
		diceService:         svc.DiceRolls,
		combatService:       svc.Combat,
		npcService:          svc.NPCs,
		inventoryService:    svc.Inventory,
		encounterService:    svc.Encounters,
		jwtManager:          svc.JWTManager,
		refreshTokenService: svc.RefreshTokens,
		websocketHub:        hub,
	}
}

// HealthCheck handles health check requests
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "healthy",
		"service": "dnd-game-backend",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to send JSON response
func sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper function to send error response
func sendErrorResponse(w http.ResponseWriter, status int, message string) {
	response := map[string]string{
		"error": message,
	}
	sendJSONResponse(w, status, response)
}