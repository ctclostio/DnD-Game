package routes

import (
	"github.com/gorilla/mux"
	"github.com/ctclostio/DnD-Game/backend/internal/handlers"
)

// RegisterCharacterRoutes registers all character-related routes
func RegisterCharacterRoutes(api *mux.Router, cfg *Config) {
	// All character routes require authentication
	auth := cfg.AuthMiddleware.Authenticate

	// Character creation routes
	if cfg.CharCreationHandler != nil {
		if h, ok := cfg.CharCreationHandler.(*handlers.CharacterCreationHandler); ok {
			api.HandleFunc("/characters/options", auth(h.GetCharacterOptions)).Methods("GET")
			api.HandleFunc("/characters/create", auth(h.CreateCharacter)).Methods("POST")
			api.HandleFunc("/characters/create-custom", auth(h.CreateCustomCharacter)).Methods("POST")
			api.HandleFunc("/characters/validate", auth(h.ValidateCharacter)).Methods("POST")
			api.HandleFunc("/characters/roll-abilities", auth(h.RollAbilityScores)).Methods("POST")
		}
	}

	// Character CRUD routes
	api.HandleFunc("/characters", auth(cfg.Handlers.GetCharacters)).Methods("GET")
	api.HandleFunc("/characters", auth(cfg.Handlers.CreateCharacter)).Methods("POST")
	api.HandleFunc("/characters/{id}", auth(cfg.Handlers.GetCharacter)).Methods("GET")
	api.HandleFunc("/characters/{id}", auth(cfg.Handlers.UpdateCharacter)).Methods("PUT")
	api.HandleFunc("/characters/{id}", auth(cfg.Handlers.DeleteCharacter)).Methods("DELETE")

	// Character action routes
	api.HandleFunc("/characters/{id}/cast-spell", auth(cfg.Handlers.CastSpell)).Methods("POST")
	api.HandleFunc("/characters/{id}/rest", auth(cfg.Handlers.Rest)).Methods("POST")
	api.HandleFunc("/characters/{id}/add-experience", auth(cfg.Handlers.AddExperience)).Methods("POST")

	// Skill check routes
	api.HandleFunc("/skill-check", auth(cfg.Handlers.PerformSkillCheck)).Methods("POST")
	api.HandleFunc("/characters/{id}/checks", auth(cfg.Handlers.GetCharacterChecks)).Methods("GET")

	// Dice roll routes
	api.HandleFunc("/dice/roll", auth(cfg.Handlers.RollDice)).Methods("POST")
}
