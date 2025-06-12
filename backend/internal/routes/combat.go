package routes

import (
	"github.com/gorilla/mux"
)

// RegisterCombatRoutes registers all combat-related routes
func RegisterCombatRoutes(api *mux.Router, cfg *Config) {
	// All combat routes require authentication
	auth := cfg.AuthMiddleware.Authenticate

	// Combat management
	api.HandleFunc("/combat/start", auth(cfg.Handlers.StartCombat)).Methods("POST")
	api.HandleFunc("/combat/{id}", auth(cfg.Handlers.GetCombat)).Methods("GET")
	api.HandleFunc("/combat/session/{sessionId}", auth(cfg.Handlers.GetCombatBySession)).Methods("GET")
	api.HandleFunc("/combat/{id}/next-turn", auth(cfg.Handlers.NextTurn)).Methods("POST")
	api.HandleFunc("/combat/{id}/action", auth(cfg.Handlers.ProcessCombatAction)).Methods("POST")
	api.HandleFunc("/combat/{id}/end", auth(cfg.Handlers.EndCombat)).Methods("POST")

	// Combatant actions
	api.HandleFunc("/combat/{id}/combatants/{combatantId}/save",
		auth(cfg.Handlers.MakeSavingThrow)).Methods("POST")
	api.HandleFunc("/combat/{id}/combatants/{combatantId}/damage",
		auth(cfg.Handlers.ApplyDamage)).Methods("POST")
	api.HandleFunc("/combat/{id}/combatants/{combatantId}/heal",
		auth(cfg.Handlers.HealCombatant)).Methods("POST")

	// Combat automation routes (commented out until handlers are implemented)
	// api.HandleFunc("/combat/{id}/automate", auth(cfg.Handlers.AutomateCombat)).Methods("POST")
	// api.HandleFunc("/combat/{id}/suggestion", auth(cfg.Handlers.GetCombatSuggestion)).Methods("GET")

	// Combat analytics routes (commented out until handlers are implemented)
	// api.HandleFunc("/combat/{id}/analytics", auth(cfg.Handlers.GetCombatAnalytics)).Methods("GET")
	// api.HandleFunc("/sessions/{sessionId}/combat-analytics",
	//	auth(cfg.Handlers.GetSessionCombatAnalytics)).Methods("GET")
	// api.HandleFunc("/characters/{characterId}/combat-stats",
	//	auth(cfg.Handlers.GetCharacterCombatStats)).Methods("GET")
}
