package routes

import (
	"github.com/gorilla/mux"
)

// RegisterDMAssistantRoutes registers all DM Assistant-related routes
func RegisterDMAssistantRoutes(api *mux.Router, cfg *Config) {
	dmOnly := cfg.AuthMiddleware.RequireDM()
	
	// DM Assistant content generation
	api.HandleFunc("/dm/assistant/generate", 
		dmOnly(cfg.Handlers.GenerateDMContent)).Methods("POST")
	api.HandleFunc("/dm/assistant/npc/generate", 
		dmOnly(cfg.Handlers.GenerateNPC)).Methods("POST")
	api.HandleFunc("/dm/assistant/location/generate", 
		dmOnly(cfg.Handlers.GenerateLocation)).Methods("POST")
	api.HandleFunc("/dm/assistant/encounter/generate", 
		dmOnly(cfg.Handlers.GenerateEncounter)).Methods("POST")
	api.HandleFunc("/dm/assistant/quest/generate", 
		dmOnly(cfg.Handlers.GenerateQuest)).Methods("POST")
		
	// DM notes and session management
	api.HandleFunc("/dm/assistant/sessions/{sessionId}/notes", 
		dmOnly(cfg.Handlers.GetDMNotes)).Methods("GET")
	api.HandleFunc("/dm/assistant/sessions/{sessionId}/notes", 
		dmOnly(cfg.Handlers.SaveDMNote)).Methods("POST")
	api.HandleFunc("/dm/assistant/sessions/{sessionId}/notes/{noteId}", 
		dmOnly(cfg.Handlers.UpdateDMNote)).Methods("PUT")
	api.HandleFunc("/dm/assistant/sessions/{sessionId}/notes/{noteId}", 
		dmOnly(cfg.Handlers.DeleteDMNote)).Methods("DELETE")
		
	// Story elements
	api.HandleFunc("/dm/assistant/story/generate", 
		dmOnly(cfg.Handlers.GenerateStoryHook)).Methods("POST")
	api.HandleFunc("/dm/assistant/dialogue/generate", 
		dmOnly(cfg.Handlers.GenerateNPCDialogue)).Methods("POST")
}