package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/config"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/handlers"
	"github.com/ctclostio/DnD-Game/backend/internal/middleware"
	"github.com/ctclostio/DnD-Game/backend/internal/routes"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/internal/websocket"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

func main() {
	// Initialize logger
	log := initializeLogger()

	// Load and validate configuration
	cfg := loadConfiguration(log)

	// Log configuration details
	logConfiguration(log, cfg)
	
	// Warn if running in development mode
	warnDevelopmentMode(log, cfg)

	// Initialize database
	db, repos := initializeDatabase(cfg, log)
	defer closeDatabase(db, log)

	// Initialize auth manager
	jwtManager := initializeAuthManager(cfg, log)

	// Create AI provider
	llmProvider := createAIProvider(cfg, log)

	// Initialize all services
	svc := initializeServices(cfg, db, repos, llmProvider, jwtManager, log)

	// Start refresh token cleanup
	startRefreshTokenCleanup(svc.RefreshTokens, log)

	// Initialize WebSocket hub
	hub := initializeWebSocket(jwtManager, log)

	// Create handlers
	h := handlers.NewHandlers(svc, db, hub)
	log.Info().Msg("Handlers initialized")

	// Setup HTTP server
	handler := setupHTTPServer(cfg, h, jwtManager, log)

	// Run server and handle shutdown
	runServer(cfg, handler, svc.RefreshTokens, hub, log)

	log.Info().Msg("Server shutdown complete")
}

// initializeLogger creates and configures the logger
func initializeLogger() *logger.LoggerV2 {
	logConfig := logger.ConfigV2{
		Level:        getEnvOrDefault("LOG_LEVEL", "info"),
		Pretty:       getEnvOrDefault("LOG_PRETTY", "false") == "true",
		CallerInfo:   true,
		StackTrace:   true,
		ServiceName:  "dnd-game-backend",
		Environment:  getEnvOrDefault("ENVIRONMENT", getEnvOrDefault("ENV", "production")),
		TimeFormat:   time.RFC3339Nano,
		SamplingRate: 1.0,
		Fields: logger.Fields{
			"version": "1.0.0",
			"pid":     os.Getpid(),
		},
	}

	log, err := logger.NewV2(&logConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log.Info().
		Str("service", logConfig.ServiceName).
		Str("environment", logConfig.Environment).
		Msg("Starting D&D Game Backend")

	return log
}

// loadConfiguration loads and validates the configuration
func loadConfiguration(log *logger.LoggerV2) *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid configuration")
	}

	return cfg
}

// logConfiguration logs the configuration details
func logConfiguration(log *logger.LoggerV2, cfg *config.Config) {
	log.Info().
		Str("database_host", cfg.Database.Host).
		Str("server_port", cfg.Server.Port).
		Bool("ai_enabled", cfg.AI.Provider != constants.MockProvider).
		Msg("Configuration loaded successfully")
}

// warnDevelopmentMode warns if running in development mode
func warnDevelopmentMode(log *logger.LoggerV2, cfg *config.Config) {
	if cfg.Server.Environment == "development" {
		log.Warn().
			Msg("⚠️  SERVER IS RUNNING IN DEVELOPMENT MODE - NOT SUITABLE FOR PRODUCTION")
		log.Warn().
			Msg("⚠️  Security features are relaxed. Set ENV=production for production use")
	}
}

// initializeDatabase initializes the database connection
func initializeDatabase(cfg *config.Config, log *logger.LoggerV2) (*database.DB, *database.Repositories) {
	log.Info().Msg("Initializing database connection")
	db, repos, err := database.InitializeWithLogging(cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	log.Info().Msg("Database initialized successfully")
	return db, repos
}

// closeDatabase closes the database connection
func closeDatabase(db *database.DB, log *logger.LoggerV2) {
	if err := db.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close database connection")
	}
}

// initializeAuthManager creates the JWT manager
func initializeAuthManager(cfg *config.Config, log *logger.LoggerV2) *auth.JWTManager {
	jwtManager := auth.NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.AccessTokenDuration,
		cfg.Auth.RefreshTokenDuration,
	)
	log.Info().Msg("JWT manager initialized")
	return jwtManager
}

// initializeServices creates all application services
func initializeServices(
	cfg *config.Config,
	db *database.DB,
	repos *database.Repositories,
	llmProvider services.LLMProvider,
	jwtManager *auth.JWTManager,
	log *logger.LoggerV2,
) *services.Services {
	log.Info().Msg("Initializing AI services")
	
	// AI services
	aiRaceGenerator := services.NewAIRaceGeneratorService(llmProvider)
	aiDMAssistant := services.NewAIDMAssistantService(llmProvider)
	aiEncounterBuilder := services.NewAIEncounterBuilder(llmProvider)
	aiCampaignManager := services.NewAICampaignManager(llmProvider, &services.AIConfig{Enabled: cfg.AI.Provider != constants.MockProvider}, log)
	aiBattleMapGenerator := services.NewAIBattleMapGenerator(llmProvider, &services.AIConfig{Enabled: cfg.AI.Provider != constants.MockProvider}, log)

	log.Info().Msg("Initializing core services")

	// Token service
	refreshTokenService := services.NewRefreshTokenService(repos.RefreshTokens, jwtManager)

	// Combat services
	combatService := services.NewCombatService()
	combatAutomationService := services.NewCombatAutomationService(repos.CombatAnalytics, repos.Characters, repos.NPCs)
	combatAnalyticsService := services.NewCombatAnalyticsService(repos.CombatAnalytics, combatService)

	// World building services
	worldBuildingRepo := database.NewWorldBuildingRepository(db)
	settlementGenerator := services.NewSettlementGeneratorService(llmProvider, worldBuildingRepo)
	factionSystem := services.NewFactionSystemService(llmProvider, worldBuildingRepo)
	worldEventEngine := services.NewWorldEventEngineService(llmProvider, worldBuildingRepo, factionSystem)
	economicSimulator := services.NewEconomicSimulatorService(worldBuildingRepo)

	// Narrative engine
	narrativeEngine, err := services.NewNarrativeEngine(cfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create narrative engine - continuing without it")
		narrativeEngine = nil
	} else {
		log.Info().Msg("Narrative engine initialized")
	}

	// Rule builder services
	diceRollService := services.NewDiceRollService(repos.DiceRolls)
	ruleEngine := services.NewRuleEngine(repos.RuleBuilder, diceRollService)
	balanceAnalyzer := services.NewAIBalanceAnalyzer(cfg, llmProvider, ruleEngine, combatService)
	conditionalReality := services.NewConditionalRealitySystem(ruleEngine)

	// Game session service with security dependencies
	gameSessionService := services.NewGameSessionService(repos.GameSessions)
	gameSessionService.SetCharacterRepository(repos.Characters)
	gameSessionService.SetUserRepository(repos.Users)

	// Aggregate all services
	return &services.Services{
		DB:                 db,
		Users:              services.NewUserService(repos.Users),
		Characters:         services.NewCharacterService(repos.Characters, repos.CustomClasses, llmProvider),
		GameSessions:       gameSessionService,
		DiceRolls:          diceRollService,
		Combat:             combatService,
		NPCs:               services.NewNPCService(repos.NPCs),
		Inventory:          services.NewInventoryService(repos.Inventory, repos.Characters),
		CustomRaces:        services.NewCustomRaceService(repos.CustomRaces, aiRaceGenerator),
		DMAssistant:        services.NewDMAssistantService(repos.DMAssistant, aiDMAssistant),
		Encounters:         services.NewEncounterService(repos.Encounters, aiEncounterBuilder, combatService),
		Campaign:           services.NewCampaignService(repos.Campaign, repos.GameSessions, aiCampaignManager),
		CombatAutomation:   combatAutomationService,
		CombatAnalytics:    combatAnalyticsService,
		SettlementGen:      settlementGenerator,
		FactionSystem:      factionSystem,
		WorldEventEngine:   worldEventEngine,
		EconomicSim:        economicSimulator,
		RuleEngine:         ruleEngine,
		BalanceAnalyzer:    balanceAnalyzer,
		ConditionalReality: conditionalReality,
		JWTManager:         jwtManager,
		RefreshTokens:      refreshTokenService,
		Config:             cfg,
		NarrativeEngine:    narrativeEngine,
		WorldBuilding:      worldBuildingRepo,
		RuleBuilder:        repos.RuleBuilder,
		AICampaignManager:  aiCampaignManager,
		BattleMapGen:       aiBattleMapGenerator,
	}
}

// startRefreshTokenCleanup starts the refresh token cleanup task
func startRefreshTokenCleanup(service *services.RefreshTokenService, log *logger.LoggerV2) {
	service.StartCleanupTask(1 * time.Hour)
	log.Info().Msg("Refresh token cleanup task started")
}

// initializeWebSocket initializes the WebSocket hub
func initializeWebSocket(jwtManager *auth.JWTManager, log *logger.LoggerV2) *websocket.Hub {
	hub := websocket.InitHub()
	websocket.SetJWTManager(jwtManager)
	log.Info().Msg("WebSocket hub started")
	return hub
}

// setupHTTPServer configures the HTTP server with all middleware and routes
func setupHTTPServer(
	cfg *config.Config,
	h *handlers.Handlers,
	jwtManager *auth.JWTManager,
	log *logger.LoggerV2,
) http.Handler {
	r := mux.NewRouter()

	// Add middleware
	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.RequestContextMiddleware)
	r.Use(middleware.LoggingMiddleware(log))
	r.Use(middleware.ErrorHandlerV2(log))
	
	isDevelopment := cfg.Server.Environment == "development"
	r.Use(middleware.SecurityHeaders(isDevelopment))

	// Create CSRF store
	csrfStore := auth.NewCSRFStore()

	// Create auth middleware
	authMiddleware := auth.NewMiddleware(jwtManager)

	// Create rate limiters
	authRateLimiter := middleware.NewRateLimiter(15, time.Minute)
	apiRateLimiter := middleware.NewRateLimiter(200, time.Minute)

	// Setup route config
	routeConfig := &routes.Config{
		Handlers:        h,
		AuthMiddleware:  authMiddleware,
		CSRFStore:       csrfStore,
		AuthRateLimiter: authRateLimiter,
		APIRateLimiter:  apiRateLimiter,
		IsProduction:    !isDevelopment,
	}

	// Setup all routes
	routes.RegisterRoutes(r, routeConfig)
	log.Info().Msg("Routes configured")

	// Setup CORS
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:8080", "http://192.168.1.161:3000"}
	if cfg.Server.Environment == "production" {
		allowedOrigins = []string{"https://yourdomain.com"}
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID", "X-Correlation-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
	})

	return c.Handler(r)
}

// runServer starts the HTTP server and handles graceful shutdown
func runServer(
	cfg *config.Config,
	handler http.Handler,
	refreshTokenService *services.RefreshTokenService,
	hub *websocket.Hub,
	log *logger.LoggerV2,
) {
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Info().
			Str("port", cfg.Server.Port).
			Str("address", srv.Addr).
			Msg("HTTP server starting")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	log.Info().
		Str("port", cfg.Server.Port).
		Msg("D&D Game Backend is running")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info().
		Str("signal", sig.String()).
		Msg("Shutdown signal received")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	refreshTokenService.StopCleanupTask()

	if err := hub.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown websocket hub")
	}
}

// Helper function to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// createAIProvider creates the appropriate LLM provider based on configuration
func createAIProvider(cfg *config.Config, log *logger.LoggerV2) services.LLMProvider {
	switch cfg.AI.Provider {
	case "openai":
		log.Info().Str("provider", "openai").Str("model", cfg.AI.Model).Msg("OpenAI provider initialized")
		return services.NewOpenAIProvider(cfg.AI.APIKey, cfg.AI.Model)
	case "anthropic":
		log.Info().Str("provider", "anthropic").Str("model", cfg.AI.Model).Msg("Anthropic provider initialized")
		return services.NewAnthropicProvider(cfg.AI.APIKey, cfg.AI.Model)
	case "openrouter":
		log.Info().Str("provider", "openrouter").Str("model", cfg.AI.Model).Msg("OpenRouter provider initialized")
		return services.NewOpenRouterProvider(cfg.AI.APIKey, cfg.AI.Model)
	default:
		log.Warn().Msg("Using mock LLM provider")
		return &services.MockLLMProvider{}
	}
}