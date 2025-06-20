package routes

import (
	"github.com/gorilla/mux"
)

// Route paths
const (
	settlementByIDPath = "/world/settlements/{id}"
	factionByIDPath = "/world/factions/{id}"
)

// RegisterWorldBuildingRoutes registers all world building-related routes
func RegisterWorldBuildingRoutes(api *mux.Router, cfg *Config) {
	auth := cfg.AuthMiddleware.Authenticate
	dmOnly := cfg.AuthMiddleware.RequireDM()

	// Settlement management
	api.HandleFunc("/world/settlements", auth(cfg.Handlers.GetSettlements)).Methods("GET")
	api.HandleFunc("/world/settlements", dmOnly(cfg.Handlers.CreateSettlement)).Methods("POST")
	api.HandleFunc(settlementByIDPath, auth(cfg.Handlers.GetSettlement)).Methods("GET")
	api.HandleFunc(settlementByIDPath, dmOnly(cfg.Handlers.UpdateSettlement)).Methods("PUT")
	api.HandleFunc(settlementByIDPath, dmOnly(cfg.Handlers.DeleteSettlement)).Methods("DELETE")
	api.HandleFunc("/world/settlements/generate", dmOnly(cfg.Handlers.GenerateSettlement)).Methods("POST")

	// Faction management
	api.HandleFunc("/world/factions", auth(cfg.Handlers.GetFactions)).Methods("GET")
	api.HandleFunc("/world/factions", dmOnly(cfg.Handlers.CreateFaction)).Methods("POST")
	api.HandleFunc(factionByIDPath, auth(cfg.Handlers.GetFaction)).Methods("GET")
	api.HandleFunc(factionByIDPath, dmOnly(cfg.Handlers.UpdateFaction)).Methods("PUT")
	api.HandleFunc(factionByIDPath, dmOnly(cfg.Handlers.DeleteFaction)).Methods("DELETE")
	api.HandleFunc("/world/factions/relationships", auth(cfg.Handlers.GetFactionRelationships)).Methods("GET")
	api.HandleFunc("/world/factions/relationships", dmOnly(cfg.Handlers.UpdateFactionRelationship)).Methods("PUT")

	// World events
	api.HandleFunc("/world/events", auth(cfg.Handlers.GetWorldEvents)).Methods("GET")
	api.HandleFunc("/world/events/active", auth(cfg.Handlers.GetActiveWorldEvents)).Methods("GET")
	api.HandleFunc("/world/events/trigger", dmOnly(cfg.Handlers.TriggerWorldEvent)).Methods("POST")
	api.HandleFunc("/world/events/{id}/resolve", dmOnly(cfg.Handlers.ResolveWorldEvent)).Methods("POST")

	// Culture and lore
	api.HandleFunc("/world/cultures", auth(cfg.Handlers.GetCultures)).Methods("GET")
	api.HandleFunc("/world/cultures", dmOnly(cfg.Handlers.CreateCulture)).Methods("POST")
	api.HandleFunc("/world/cultures/{id}", auth(cfg.Handlers.GetCulture)).Methods("GET")
	api.HandleFunc("/world/cultures/{id}", dmOnly(cfg.Handlers.UpdateCulture)).Methods("PUT")

	// Economic simulation
	api.HandleFunc("/world/economy/simulate", dmOnly(cfg.Handlers.SimulateEconomy)).Methods("POST")
	api.HandleFunc("/world/economy/status", auth(cfg.Handlers.GetEconomicStatus)).Methods("GET")
}
