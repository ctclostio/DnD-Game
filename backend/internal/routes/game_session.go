package routes

import (
	"github.com/gorilla/mux"
)

// RegisterGameSessionRoutes registers all game session-related routes.
func RegisterGameSessionRoutes(api *mux.Router, cfg *Config) {
	// Game session routes with role-based access.
	auth := cfg.AuthMiddleware.Authenticate
	dmOnly := cfg.AuthMiddleware.RequireDM()

	// Session management.
	api.HandleFunc("/game/sessions", dmOnly(cfg.Handlers.CreateGameSession)).Methods("POST")
	api.HandleFunc("/game/sessions/{id}", auth(cfg.Handlers.GetGameSession)).Methods("GET")
	api.HandleFunc("/game/sessions/{id}", dmOnly(cfg.Handlers.UpdateGameSession)).Methods("PUT")
	api.HandleFunc("/game/sessions/{id}/join", auth(cfg.Handlers.JoinGameSession)).Methods("POST")
	api.HandleFunc("/game/sessions/{id}/leave", auth(cfg.Handlers.LeaveGameSession)).Methods("POST")

	// Additional session routes.
	api.HandleFunc("/game/sessions", auth(cfg.Handlers.GetActiveSessions)).Methods("GET")
	api.HandleFunc("/game/sessions/{id}/players", auth(cfg.Handlers.GetSessionPlayers)).Methods("GET")
	api.HandleFunc("/game/sessions/{id}/kick/{playerId}",
		dmOnly(cfg.Handlers.KickPlayer)).Methods("POST")
}
