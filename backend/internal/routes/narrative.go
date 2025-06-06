package routes

import (
	"github.com/gorilla/mux"
)

// RegisterNarrativeRoutes registers all narrative engine-related routes
func RegisterNarrativeRoutes(api *mux.Router, cfg *Config) {
	auth := cfg.AuthMiddleware.Authenticate
	dmOnly := cfg.AuthMiddleware.RequireDM()
	
	// Story threads management
	api.HandleFunc("/narrative/threads", auth(cfg.Handlers.GetStoryThreads)).Methods("GET")
	api.HandleFunc("/narrative/threads", dmOnly(cfg.Handlers.CreateStoryThread)).Methods("POST")
	api.HandleFunc("/narrative/threads/{id}", auth(cfg.Handlers.GetStoryThread)).Methods("GET")
	api.HandleFunc("/narrative/threads/{id}", dmOnly(cfg.Handlers.UpdateStoryThread)).Methods("PUT")
	api.HandleFunc("/narrative/threads/{id}/advance", dmOnly(cfg.Handlers.AdvanceStoryThread)).Methods("POST")
	api.HandleFunc("/narrative/threads/{id}/resolve", dmOnly(cfg.Handlers.ResolveStoryThread)).Methods("POST")
	
	// Character memories and perspectives
	api.HandleFunc("/narrative/characters/{characterId}/memories", 
		auth(cfg.Handlers.GetCharacterMemories)).Methods("GET")
	api.HandleFunc("/narrative/characters/{characterId}/memories", 
		auth(cfg.Handlers.AddCharacterMemory)).Methods("POST")
	api.HandleFunc("/narrative/characters/{characterId}/perspective", 
		auth(cfg.Handlers.GetCharacterPerspective)).Methods("GET")
		
	// Consequence tracking
	api.HandleFunc("/narrative/consequences", auth(cfg.Handlers.GetConsequences)).Methods("GET")
	api.HandleFunc("/narrative/consequences/active", auth(cfg.Handlers.GetActiveConsequences)).Methods("GET")
	api.HandleFunc("/narrative/consequences/{id}/resolve", 
		dmOnly(cfg.Handlers.ResolveConsequence)).Methods("POST")
		
	// Narrative generation
	api.HandleFunc("/narrative/generate/recap", auth(cfg.Handlers.GenerateSessionRecap)).Methods("POST")
	api.HandleFunc("/narrative/generate/foreshadowing", 
		dmOnly(cfg.Handlers.GenerateForeshadowing)).Methods("POST")
	api.HandleFunc("/narrative/generate/plot-twist", 
		dmOnly(cfg.Handlers.GeneratePlotTwist)).Methods("POST")
		
	// Backstory integration
	api.HandleFunc("/narrative/backstory/hooks", auth(cfg.Handlers.GetBackstoryHooks)).Methods("GET")
	api.HandleFunc("/narrative/backstory/integrate", 
		dmOnly(cfg.Handlers.IntegrateBackstory)).Methods("POST")
}