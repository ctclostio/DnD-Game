package routes

import (
	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/handlers"
	"github.com/ctclostio/DnD-Game/backend/internal/middleware"
)

// Config holds all dependencies needed for route registration
type Config struct {
	Handlers                *handlers.Handlers
	CharCreationHandler     interface{} // Add specific handler interfaces as needed
	InventoryHandler        interface{}
	CampaignHandler         interface{}
	CombatAutomationHandler interface{}
	WorldBuildingHandler    interface{}
	NarrativeHandler        interface{}
	AuthMiddleware          *auth.Middleware
	CSRFStore               *auth.CSRFStore
	AuthRateLimiter         *middleware.RateLimiter
	APIRateLimiter          *middleware.RateLimiter
	IsProduction            bool
}

// RegisterRoutes sets up all application routes
func RegisterRoutes(router *mux.Router, cfg *Config) {
	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Apply CSRF middleware to all routes
	api.Use(auth.CSRFMiddleware(cfg.CSRFStore, cfg.IsProduction))

	// Apply general API rate limiting
	api.Use(cfg.APIRateLimiter.Middleware())

	// Health check endpoints (no auth required, outside rate limiting)
	router.HandleFunc("/health", cfg.Handlers.Health).Methods("GET")
	router.HandleFunc("/health/live", cfg.Handlers.LivenessProbe).Methods("GET")
	router.HandleFunc("/health/ready", cfg.Handlers.ReadinessProbe).Methods("GET")

	// Detailed health requires authentication
	api.HandleFunc("/health/detailed",
		cfg.AuthMiddleware.Authenticate(cfg.Handlers.DetailedHealth)).Methods("GET")

	// CSRF token endpoint
	api.HandleFunc("/csrf-token", cfg.Handlers.GetCSRFToken).Methods("GET")

	// Swagger documentation - only in development mode
	if !cfg.IsProduction {
		router.HandleFunc("/swagger", handlers.SwaggerUI).Methods("GET")
		api.HandleFunc("/swagger.json", cfg.Handlers.SwaggerJSON).Methods("GET")
	}

	// Register route groups
	RegisterAuthRoutes(api, cfg)
	RegisterCharacterRoutes(api, cfg)
	RegisterCombatRoutes(api, cfg)
	RegisterGameSessionRoutes(api, cfg)
	RegisterNPCRoutes(api, cfg)
	RegisterInventoryRoutes(api, cfg)
	RegisterDMAssistantRoutes(api, cfg)
	RegisterWorldBuildingRoutes(api, cfg)
	RegisterRuleBuilderRoutes(api, cfg)
	RegisterNarrativeRoutes(api, cfg)
	RegisterCRDTRoutes(router)
}
