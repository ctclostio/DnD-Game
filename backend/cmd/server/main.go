package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/config"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/handlers"
	"github.com/your-username/dnd-game/backend/internal/middleware"
	"github.com/your-username/dnd-game/backend/internal/routes"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/websocket"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize database
	db, repos, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.AccessTokenDuration,
		cfg.Auth.RefreshTokenDuration,
	)

	// Create AI provider
	var llmProvider services.LLMProvider
	switch cfg.AI.Provider {
	case "openai":
		llmProvider = services.NewOpenAIProvider(cfg.AI.APIKey, cfg.AI.Model)
	case "anthropic":
		llmProvider = services.NewAnthropicProvider(cfg.AI.APIKey, cfg.AI.Model)
	case "openrouter":
		llmProvider = services.NewOpenRouterProvider(cfg.AI.APIKey, cfg.AI.Model)
	default:
		llmProvider = &services.MockLLMProvider{}
	}

	// Create AI race generator
	aiRaceGenerator := services.NewAIRaceGeneratorService(llmProvider)
	
	// Create AI DM assistant
	aiDMAssistant := services.NewAIDMAssistantService(llmProvider)

	// Create services
	refreshTokenService := services.NewRefreshTokenService(repos.RefreshTokens, jwtManager)
	
	// Start refresh token cleanup task
	refreshTokenService.StartCleanupTask(1 * time.Hour)

	// Create custom race service
	customRaceService := services.NewCustomRaceService(repos.CustomRaces, aiRaceGenerator)
	
	// Create DM assistant service
	dmAssistantService := services.NewDMAssistantService(repos.DMAssistant, aiDMAssistant)
	
	// Create AI encounter builder
	aiEncounterBuilder := services.NewAIEncounterBuilder(llmProvider)
	
	// Create combat service first (needed by encounter service)
	combatService := services.NewCombatService()
	
	// Create encounter service
	encounterService := services.NewEncounterService(repos.Encounters, aiEncounterBuilder, combatService)
	
	// Create AI campaign manager
	aiCampaignManager := services.NewAICampaignManager(llmProvider, &services.AIConfig{Enabled: cfg.AI.Provider != "mock"})
	
	// Create campaign service
	campaignService := services.NewCampaignService(repos.Campaign, repos.GameSessions, aiCampaignManager)
	
	// Create AI battle map generator
	aiBattleMapGenerator := services.NewAIBattleMapGenerator(llmProvider, &services.AIConfig{Enabled: cfg.AI.Provider != "mock"})
	
	// Create combat automation service
	combatAutomationService := services.NewCombatAutomationService(repos.CombatAnalytics, repos.Characters, repos.NPCs)
	
	// Create combat analytics service
	combatAnalyticsService := services.NewCombatAnalyticsService(repos.CombatAnalytics, combatService)
	
	// Create world building repository
	worldBuildingRepo := database.NewWorldBuildingRepository(db.StdDB())
	
	// Create world building services
	settlementGenerator := services.NewSettlementGeneratorService(llmProvider, worldBuildingRepo)
	factionSystem := services.NewFactionSystemService(llmProvider, worldBuildingRepo)
	worldEventEngine := services.NewWorldEventEngineService(llmProvider, worldBuildingRepo, factionSystem)
	economicSimulator := services.NewEconomicSimulatorService(worldBuildingRepo)
	
	// Create narrative engine
	narrativeEngine, err := services.NewNarrativeEngine(cfg)
	if err != nil {
		log.Printf("Failed to create narrative engine: %v", err)
		// Continue without narrative engine rather than failing completely
	}
	
	// Create rule builder services
	ruleEngine := services.NewRuleEngine(repos.RuleBuilder, services.NewDiceRollService(repos.DiceRolls))
	balanceAnalyzer := services.NewAIBalanceAnalyzer(cfg, llmProvider, ruleEngine, combatService)
	conditionalReality := services.NewConditionalRealitySystem(ruleEngine)

	svc := &services.Services{
		Users:              services.NewUserService(repos.Users),
		Characters:         services.NewCharacterService(repos.Characters, repos.CustomClasses, llmProvider),
		GameSessions:       services.NewGameSessionService(repos.GameSessions),
		DiceRolls:          services.NewDiceRollService(repos.DiceRolls),
		Combat:             combatService,
		NPCs:               services.NewNPCService(repos.NPCs),
		Inventory:          services.NewInventoryService(repos.Inventory, repos.Characters),
		CustomRaces:        customRaceService,
		DMAssistant:        dmAssistantService,
		Encounters:         encounterService,
		Campaign:           campaignService,
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
	}

	// Get websocket hub
	wsHub := websocket.GetHub()
	
	// Create handlers with services
	h := handlers.NewHandlers(svc, wsHub)
	
	// Create character creation handler
	charCreationHandler := handlers.NewCharacterCreationHandler(svc.Characters, svc.CustomRaces, llmProvider)
	
	// Create inventory handler
	inventoryHandler := handlers.NewInventoryHandler(svc.Inventory)
	
	// Create campaign handler
	campaignHandler := handlers.NewCampaignHandler(svc.Campaign, svc.GameSessions)
	
	// Create combat automation handler
	combatAutomationHandler := handlers.NewCombatAutomationHandler(
		svc.CombatAutomation,
		svc.CombatAnalytics,
		svc.Characters,
		svc.GameSessions,
		aiBattleMapGenerator,
	)
	
	// Create world building handler
	worldBuildingHandler := handlers.NewWorldBuildingHandlers(
		svc.SettlementGen,
		svc.FactionSystem,
		svc.WorldEventEngine,
		svc.EconomicSim,
		worldBuildingRepo,
	)
	
	// Create narrative handler
	var narrativeHandler *handlers.NarrativeHandlers
	if narrativeEngine != nil {
		narrativeHandler = handlers.NewNarrativeHandlers(
			narrativeEngine,
			repos.Narrative,
			repos.Characters,
			repos.GameSessions,
		)
	}

	// Create authentication middleware
	authMiddleware := auth.NewMiddleware(jwtManager)
	
	// Create CSRF store
	csrfStore := auth.NewCSRFStore()

	// Create logger for middleware
	logger := log.New(os.Stdout, "[DND-GAME] ", log.LstdFlags|log.Lshortfile)

	// Create rate limiters
	authRateLimiter := middleware.AuthRateLimiter()
	apiRateLimiter := middleware.APIRateLimiter()

	router := mux.NewRouter()
	
	// Apply global middleware
	router.Use(middleware.Recovery(logger))
	
	// Apply security headers
	isDevelopment := os.Getenv("GO_ENV") == "development"
	router.Use(middleware.SecurityHeaders(isDevelopment))

	// Configure route dependencies
	routeConfig := &routes.Config{
		Handlers:                h,
		CharCreationHandler:     charCreationHandler,
		InventoryHandler:        inventoryHandler,
		CampaignHandler:         campaignHandler,
		CombatAutomationHandler: combatAutomationHandler,
		WorldBuildingHandler:    worldBuildingHandler,
		NarrativeHandler:        narrativeHandler,
		AuthMiddleware:          authMiddleware,
		CSRFStore:               csrfStore,
		AuthRateLimiter:         authRateLimiter,
		APIRateLimiter:          apiRateLimiter,
	}
	
	// Register all routes
	routes.RegisterRoutes(router, routeConfig)
	
	// Initialize WebSocket with JWT manager
	websocket.SetJWTManager(jwtManager)
	
	// WebSocket endpoint
	router.HandleFunc("/ws", websocket.HandleWebSocket)
	
	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/build/")))

	// Setup CORS
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:8080"}
	
	// Add production origins from environment
	if prodOrigin := os.Getenv("PRODUCTION_ORIGIN"); prodOrigin != "" {
		allowedOrigins = append(allowedOrigins, prodOrigin)
	}
	
	// In production, use strict CORS
	corsOptions := cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		ExposedHeaders:   []string{"Authorization"},
		MaxAge:           86400, // 24 hours
	}
	
	// In development, allow all headers
	if isDevelopment {
		corsOptions.AllowedHeaders = []string{"*"}
		corsOptions.Debug = true
	}
	
	c := cors.New(corsOptions)

	handler := c.Handler(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}