package routes

import (
	"github.com/gorilla/mux"
)

// Route paths
const (
	npcByIDPath = "/npcs/{id}"
)

// RegisterNPCRoutes registers all NPC-related routes
func RegisterNPCRoutes(api *mux.Router, cfg *Config) {
	auth := cfg.AuthMiddleware.Authenticate
	dmOnly := cfg.AuthMiddleware.RequireDM()

	// NPC CRUD operations (DM only for create/update/delete)
	api.HandleFunc("/npcs", dmOnly(cfg.Handlers.CreateNPC)).Methods("POST")
	api.HandleFunc(npcByIDPath, auth(cfg.Handlers.GetNPC)).Methods("GET")
	api.HandleFunc(npcByIDPath, dmOnly(cfg.Handlers.UpdateNPC)).Methods("PUT")
	api.HandleFunc(npcByIDPath, dmOnly(cfg.Handlers.DeleteNPC)).Methods("DELETE")

	// NPC queries
	api.HandleFunc("/npcs/session/{sessionId}", auth(cfg.Handlers.GetNPCsBySession)).Methods("GET")
	api.HandleFunc("/npcs/search", auth(cfg.Handlers.SearchNPCs)).Methods("GET")
	api.HandleFunc("/npcs/templates", auth(cfg.Handlers.GetNPCTemplates)).Methods("GET")

	// NPC creation helpers
	api.HandleFunc("/npcs/create-from-template",
		dmOnly(cfg.Handlers.CreateNPCFromTemplate)).Methods("POST")
	api.HandleFunc("/npcs/{id}/action/{action}",
		dmOnly(cfg.Handlers.NPCQuickActions)).Methods("POST")
}
