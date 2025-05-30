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

	svc := &services.Services{
		Users:            services.NewUserService(repos.Users),
		Characters:       services.NewCharacterService(repos.Characters, repos.CustomClasses, llmProvider),
		GameSessions:     services.NewGameSessionService(repos.GameSessions),
		DiceRolls:        services.NewDiceRollService(repos.DiceRolls),
		Combat:           combatService,
		NPCs:             services.NewNPCService(repos.NPCs),
		Inventory:        services.NewInventoryService(repos.Inventory, repos.Characters),
		CustomRaces:      customRaceService,
		DMAssistant:      dmAssistantService,
		Encounters:       encounterService,
		Campaign:         campaignService,
		CombatAutomation: combatAutomationService,
		CombatAnalytics:  combatAnalyticsService,
		SettlementGen:    settlementGenerator,
		FactionSystem:    factionSystem,
		WorldEventEngine: worldEventEngine,
		EconomicSim:      economicSimulator,
		JWTManager:       jwtManager,
		RefreshTokens:    refreshTokenService,
		Config:           cfg,
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

	// Create authentication middleware
	authMiddleware := auth.NewMiddleware(jwtManager)

	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", h.HealthCheck).Methods("GET")
	
	// Auth routes (public)
	api.HandleFunc("/auth/register", h.Register).Methods("POST")
	api.HandleFunc("/auth/login", h.Login).Methods("POST")
	api.HandleFunc("/auth/refresh", h.RefreshToken).Methods("POST")
	api.HandleFunc("/auth/logout", authMiddleware.Authenticate(h.Logout)).Methods("POST")
	api.HandleFunc("/auth/me", authMiddleware.Authenticate(h.GetCurrentUser)).Methods("GET")
	
	// Character creation routes (protected)
	api.HandleFunc("/characters/options", authMiddleware.Authenticate(charCreationHandler.GetCharacterOptions)).Methods("GET")
	api.HandleFunc("/characters/create", authMiddleware.Authenticate(charCreationHandler.CreateCharacter)).Methods("POST")
	api.HandleFunc("/characters/create-custom", authMiddleware.Authenticate(charCreationHandler.CreateCustomCharacter)).Methods("POST")
	api.HandleFunc("/characters/validate", authMiddleware.Authenticate(charCreationHandler.ValidateCharacter)).Methods("POST")
	api.HandleFunc("/characters/roll-abilities", authMiddleware.Authenticate(charCreationHandler.RollAbilityScores)).Methods("POST")
	
	// Character routes (protected)
	api.HandleFunc("/characters", authMiddleware.Authenticate(h.GetCharacters)).Methods("GET")
	api.HandleFunc("/characters", authMiddleware.Authenticate(h.CreateCharacter)).Methods("POST")
	api.HandleFunc("/characters/{id}", authMiddleware.Authenticate(h.GetCharacter)).Methods("GET")
	api.HandleFunc("/characters/{id}", authMiddleware.Authenticate(h.UpdateCharacter)).Methods("PUT")
	api.HandleFunc("/characters/{id}", authMiddleware.Authenticate(h.DeleteCharacter)).Methods("DELETE")
	api.HandleFunc("/characters/{id}/cast-spell", authMiddleware.Authenticate(h.CastSpell)).Methods("POST")
	api.HandleFunc("/characters/{id}/rest", authMiddleware.Authenticate(h.Rest)).Methods("POST")
	api.HandleFunc("/characters/{id}/add-experience", authMiddleware.Authenticate(h.AddExperience)).Methods("POST")
	
	// Dice roll routes (protected)
	api.HandleFunc("/dice/roll", authMiddleware.Authenticate(h.RollDice)).Methods("POST")
	
	// Game session routes (protected)
	api.HandleFunc("/game/sessions", authMiddleware.RequireDM()(h.CreateGameSession)).Methods("POST")
	api.HandleFunc("/game/sessions/{id}", authMiddleware.Authenticate(h.GetGameSession)).Methods("GET")
	api.HandleFunc("/game/sessions/{id}", authMiddleware.RequireDM()(h.UpdateGameSession)).Methods("PUT")
	api.HandleFunc("/game/sessions/{id}/join", authMiddleware.Authenticate(h.JoinGameSession)).Methods("POST")
	api.HandleFunc("/game/sessions/{id}/leave", authMiddleware.Authenticate(h.LeaveGameSession)).Methods("POST")

	// Combat routes (protected)
	api.HandleFunc("/combat/start", authMiddleware.Authenticate(h.StartCombat)).Methods("POST")
	api.HandleFunc("/combat/{id}", authMiddleware.Authenticate(h.GetCombat)).Methods("GET")
	api.HandleFunc("/combat/session/{sessionId}", authMiddleware.Authenticate(h.GetCombatBySession)).Methods("GET")
	api.HandleFunc("/combat/{id}/next-turn", authMiddleware.Authenticate(h.NextTurn)).Methods("POST")
	api.HandleFunc("/combat/{id}/action", authMiddleware.Authenticate(h.ProcessCombatAction)).Methods("POST")
	api.HandleFunc("/combat/{id}/end", authMiddleware.Authenticate(h.EndCombat)).Methods("POST")
	api.HandleFunc("/combat/{id}/combatants/{combatantId}/save", authMiddleware.Authenticate(h.MakeSavingThrow)).Methods("POST")
	api.HandleFunc("/combat/{id}/combatants/{combatantId}/damage", authMiddleware.Authenticate(h.ApplyDamage)).Methods("POST")
	api.HandleFunc("/combat/{id}/combatants/{combatantId}/heal", authMiddleware.Authenticate(h.HealCombatant)).Methods("POST")

	// NPC routes (protected)
	api.HandleFunc("/npcs", authMiddleware.RequireDM()(h.CreateNPC)).Methods("POST")
	api.HandleFunc("/npcs/{id}", authMiddleware.Authenticate(h.GetNPC)).Methods("GET")
	api.HandleFunc("/npcs/{id}", authMiddleware.RequireDM()(h.UpdateNPC)).Methods("PUT")
	api.HandleFunc("/npcs/{id}", authMiddleware.RequireDM()(h.DeleteNPC)).Methods("DELETE")
	api.HandleFunc("/npcs/session/{sessionId}", authMiddleware.Authenticate(h.GetNPCsBySession)).Methods("GET")
	api.HandleFunc("/npcs/search", authMiddleware.Authenticate(h.SearchNPCs)).Methods("GET")
	api.HandleFunc("/npcs/templates", authMiddleware.Authenticate(h.GetNPCTemplates)).Methods("GET")
	api.HandleFunc("/npcs/create-from-template", authMiddleware.RequireDM()(h.CreateNPCFromTemplate)).Methods("POST")
	api.HandleFunc("/npcs/{id}/action/{action}", authMiddleware.RequireDM()(h.NPCQuickActions)).Methods("POST")
	
	// Skill check routes (protected)
	api.HandleFunc("/skill-check", authMiddleware.Authenticate(h.PerformSkillCheck)).Methods("POST")
	api.HandleFunc("/characters/{id}/checks", authMiddleware.Authenticate(h.GetCharacterChecks)).Methods("GET")
	
	// Inventory routes (protected)
	api.HandleFunc("/characters/{characterId}/inventory", authMiddleware.Authenticate(inventoryHandler.GetCharacterInventory)).Methods("GET")
	api.HandleFunc("/characters/{characterId}/inventory", authMiddleware.Authenticate(inventoryHandler.AddItemToInventory)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/remove", authMiddleware.Authenticate(inventoryHandler.RemoveItemFromInventory)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/equip", authMiddleware.Authenticate(inventoryHandler.EquipItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/unequip", authMiddleware.Authenticate(inventoryHandler.UnequipItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/attune", authMiddleware.Authenticate(inventoryHandler.AttuneItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/unattune", authMiddleware.Authenticate(inventoryHandler.UnattuneItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/currency", authMiddleware.Authenticate(inventoryHandler.GetCharacterCurrency)).Methods("GET")
	api.HandleFunc("/characters/{characterId}/currency", authMiddleware.Authenticate(inventoryHandler.UpdateCharacterCurrency)).Methods("PUT")
	api.HandleFunc("/characters/{characterId}/inventory/purchase", authMiddleware.Authenticate(inventoryHandler.PurchaseItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/sell", authMiddleware.Authenticate(inventoryHandler.SellItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/weight", authMiddleware.Authenticate(inventoryHandler.GetCharacterWeight)).Methods("GET")
	api.HandleFunc("/items", authMiddleware.RequireDM()(inventoryHandler.CreateItem)).Methods("POST")
	api.HandleFunc("/items", authMiddleware.Authenticate(inventoryHandler.GetItemsByType)).Methods("GET")

	// Custom race routes (protected)
	api.HandleFunc("/custom-races", authMiddleware.Authenticate(h.CreateCustomRace)).Methods("POST")
	api.HandleFunc("/custom-races", authMiddleware.Authenticate(h.GetUserCustomRaces)).Methods("GET")
	api.HandleFunc("/custom-races/public", authMiddleware.Authenticate(h.GetPublicCustomRaces)).Methods("GET")
	api.HandleFunc("/custom-races/pending", authMiddleware.RequireDM()(h.GetPendingCustomRaces)).Methods("GET")
	api.HandleFunc("/custom-races/{id}", authMiddleware.Authenticate(h.GetCustomRace)).Methods("GET")
	api.HandleFunc("/custom-races/{id}/stats", authMiddleware.Authenticate(h.GetCustomRaceStats)).Methods("GET")
	api.HandleFunc("/custom-races/{id}/approve", authMiddleware.RequireDM()(h.ApproveCustomRace)).Methods("POST")
	api.HandleFunc("/custom-races/{id}/reject", authMiddleware.RequireDM()(h.RejectCustomRace)).Methods("POST")
	
	// Custom class routes (protected)
	api.HandleFunc("/characters/custom-classes/generate", authMiddleware.Authenticate(h.GenerateCustomClass)).Methods("POST")
	api.HandleFunc("/characters/custom-classes", authMiddleware.Authenticate(h.GetCustomClasses)).Methods("GET")
	api.HandleFunc("/characters/custom-classes/{id}", authMiddleware.Authenticate(h.GetCustomClass)).Methods("GET")
	api.HandleFunc("/custom-races/{id}/revision", authMiddleware.RequireDM()(h.RequestRevisionCustomRace)).Methods("POST")
	api.HandleFunc("/custom-races/{id}/public", authMiddleware.Authenticate(h.MakeCustomRacePublic)).Methods("POST")

	// Encounter Builder routes (protected, DM only)
	api.HandleFunc("/encounters/generate", authMiddleware.RequireDM()(h.GenerateEncounter)).Methods("POST")
	api.HandleFunc("/encounters/{id}", authMiddleware.Authenticate(h.GetEncounter)).Methods("GET")
	api.HandleFunc("/encounters/session/{sessionId}", authMiddleware.Authenticate(h.GetSessionEncounters)).Methods("GET")
	api.HandleFunc("/encounters/{id}/start", authMiddleware.RequireDM()(h.StartEncounter)).Methods("POST")
	api.HandleFunc("/encounters/{id}/complete", authMiddleware.RequireDM()(h.CompleteEncounter)).Methods("POST")
	api.HandleFunc("/encounters/{id}/scale", authMiddleware.RequireDM()(h.ScaleEncounter)).Methods("POST")
	api.HandleFunc("/encounters/{id}/tactical-suggestion", authMiddleware.RequireDM()(h.GetTacticalSuggestion)).Methods("POST")
	api.HandleFunc("/encounters/{id}/events", authMiddleware.Authenticate(h.LogEncounterEvent)).Methods("POST")
	api.HandleFunc("/encounters/{id}/events", authMiddleware.Authenticate(h.GetEncounterEvents)).Methods("GET")
	api.HandleFunc("/encounters/{id}/enemies/{enemyId}", authMiddleware.RequireDM()(h.UpdateEnemyStatus)).Methods("PATCH")
	api.HandleFunc("/encounters/{id}/reinforcements", authMiddleware.RequireDM()(h.TriggerReinforcements)).Methods("POST")
	api.HandleFunc("/encounters/{id}/objectives/check", authMiddleware.RequireDM()(h.CheckObjectives)).Methods("POST")

	// DM Assistant routes (protected, DM only)
	api.HandleFunc("/dm-assistant", authMiddleware.RequireDM()(h.ProcessDMAssistantRequest)).Methods("POST")
	api.HandleFunc("/dm-assistant/sessions/{sessionId}/npcs", authMiddleware.RequireDM()(h.GetDMAssistantNPCs)).Methods("GET")
	api.HandleFunc("/dm-assistant/npcs", authMiddleware.RequireDM()(h.CreateDMAssistantNPC)).Methods("POST")
	api.HandleFunc("/dm-assistant/npcs/{id}", authMiddleware.RequireDM()(h.GetDMAssistantNPC)).Methods("GET")
	api.HandleFunc("/dm-assistant/sessions/{sessionId}/locations", authMiddleware.RequireDM()(h.GetDMAssistantLocations)).Methods("GET")
	api.HandleFunc("/dm-assistant/locations/{id}", authMiddleware.RequireDM()(h.GetDMAssistantLocation)).Methods("GET")
	api.HandleFunc("/dm-assistant/sessions/{sessionId}/story-elements", authMiddleware.RequireDM()(h.GetDMAssistantStoryElements)).Methods("GET")
	api.HandleFunc("/dm-assistant/story-elements/{id}/use", authMiddleware.RequireDM()(h.MarkStoryElementUsed)).Methods("POST")
	api.HandleFunc("/dm-assistant/locations/{locationId}/hazards", authMiddleware.RequireDM()(h.GetDMAssistantHazards)).Methods("GET")
	api.HandleFunc("/dm-assistant/hazards/{id}/trigger", authMiddleware.RequireDM()(h.TriggerHazard)).Methods("POST")
	
	// Campaign Management routes (protected)
	// Story Arc routes
	api.HandleFunc("/sessions/{sessionId}/story-arcs", authMiddleware.RequireDM()(campaignHandler.CreateStoryArc)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/story-arcs/generate", authMiddleware.RequireDM()(campaignHandler.GenerateStoryArc)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/story-arcs", authMiddleware.Authenticate(campaignHandler.GetStoryArcs)).Methods("GET")
	api.HandleFunc("/sessions/{sessionId}/story-arcs/{arcId}", authMiddleware.RequireDM()(campaignHandler.UpdateStoryArc)).Methods("PUT")
	
	// Session Memory routes
	api.HandleFunc("/sessions/{sessionId}/memories", authMiddleware.RequireDM()(campaignHandler.CreateSessionMemory)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/memories", authMiddleware.Authenticate(campaignHandler.GetSessionMemories)).Methods("GET")
	api.HandleFunc("/sessions/{sessionId}/recap", authMiddleware.Authenticate(campaignHandler.GenerateRecap)).Methods("POST")
	
	// Plot Thread routes
	api.HandleFunc("/sessions/{sessionId}/plot-threads", authMiddleware.RequireDM()(campaignHandler.CreatePlotThread)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/plot-threads", authMiddleware.Authenticate(campaignHandler.GetPlotThreads)).Methods("GET")
	
	// Foreshadowing routes
	api.HandleFunc("/sessions/{sessionId}/foreshadowing", authMiddleware.RequireDM()(campaignHandler.GenerateForeshadowing)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/foreshadowing/unrevealed", authMiddleware.RequireDM()(campaignHandler.GetUnrevealedForeshadowing)).Methods("GET")
	api.HandleFunc("/foreshadowing/{elementId}/reveal", authMiddleware.RequireDM()(campaignHandler.RevealForeshadowing)).Methods("POST")
	
	// Timeline routes
	api.HandleFunc("/sessions/{sessionId}/timeline", authMiddleware.Authenticate(campaignHandler.AddTimelineEvent)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/timeline", authMiddleware.Authenticate(campaignHandler.GetTimeline)).Methods("GET")
	
	// NPC Relationship routes
	api.HandleFunc("/sessions/{sessionId}/npc-relationships", authMiddleware.RequireDM()(campaignHandler.UpdateNPCRelationship)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/npcs/{npcId}/relationships", authMiddleware.Authenticate(campaignHandler.GetNPCRelationships)).Methods("GET")
	
	// Combat Automation routes
	// Auto-resolution
	api.HandleFunc("/sessions/{sessionId}/combat/auto-resolve", authMiddleware.RequireDM()(combatAutomationHandler.AutoResolveCombat)).Methods("POST")
	
	// Smart Initiative
	api.HandleFunc("/sessions/{sessionId}/combat/smart-initiative", authMiddleware.Authenticate(combatAutomationHandler.SmartInitiative)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/initiative-rules", authMiddleware.RequireDM()(combatAutomationHandler.SetInitiativeRules)).Methods("POST")
	
	// Battle Maps
	api.HandleFunc("/sessions/{sessionId}/battle-maps", authMiddleware.RequireDM()(combatAutomationHandler.GenerateBattleMap)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/battle-maps", authMiddleware.Authenticate(combatAutomationHandler.GetBattleMaps)).Methods("GET")
	api.HandleFunc("/battle-maps/{mapId}", authMiddleware.Authenticate(combatAutomationHandler.GetBattleMap)).Methods("GET")
	
	// Combat Analytics
	api.HandleFunc("/combat/{combatId}/analytics", authMiddleware.Authenticate(combatAutomationHandler.GetCombatAnalytics)).Methods("GET")
	api.HandleFunc("/sessions/{sessionId}/combat-history", authMiddleware.Authenticate(combatAutomationHandler.GetSessionCombatHistory)).Methods("GET")
	
	// World Building routes
	// Settlement routes
	api.HandleFunc("/sessions/{sessionId}/settlements/generate", authMiddleware.RequireDM()(worldBuildingHandler.GenerateSettlement)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/settlements", authMiddleware.Authenticate(worldBuildingHandler.GetSettlements)).Methods("GET")
	api.HandleFunc("/settlements/{settlementId}", authMiddleware.Authenticate(worldBuildingHandler.GetSettlement)).Methods("GET")
	api.HandleFunc("/settlements/{settlementId}/market", authMiddleware.Authenticate(worldBuildingHandler.GetSettlementMarket)).Methods("GET")
	api.HandleFunc("/settlements/{settlementId}/calculate-price", authMiddleware.Authenticate(worldBuildingHandler.CalculateItemPrice)).Methods("POST")
	
	// Faction routes
	api.HandleFunc("/sessions/{sessionId}/factions", authMiddleware.RequireDM()(worldBuildingHandler.CreateFaction)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/factions", authMiddleware.Authenticate(worldBuildingHandler.GetFactions)).Methods("GET")
	api.HandleFunc("/factions/{faction1Id}/relationships/{faction2Id}", authMiddleware.RequireDM()(worldBuildingHandler.UpdateFactionRelationship)).Methods("PUT")
	api.HandleFunc("/sessions/{sessionId}/factions/simulate-conflicts", authMiddleware.RequireDM()(worldBuildingHandler.SimulateFactionConflicts)).Methods("POST")
	
	// World Event routes
	api.HandleFunc("/sessions/{sessionId}/world-events", authMiddleware.RequireDM()(worldBuildingHandler.CreateWorldEvent)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/world-events/active", authMiddleware.Authenticate(worldBuildingHandler.GetActiveWorldEvents)).Methods("GET")
	api.HandleFunc("/sessions/{sessionId}/world-events/progress", authMiddleware.RequireDM()(worldBuildingHandler.ProgressWorldEvents)).Methods("POST")
	
	// Trade Route and Economic routes
	api.HandleFunc("/trade-routes", authMiddleware.RequireDM()(worldBuildingHandler.CreateTradeRoute)).Methods("POST")
	api.HandleFunc("/sessions/{sessionId}/economy/simulate", authMiddleware.RequireDM()(worldBuildingHandler.SimulateEconomics)).Methods("POST")

	// Initialize WebSocket with JWT manager
	websocket.SetJWTManager(jwtManager)
	
	// WebSocket endpoint
	router.HandleFunc("/ws", websocket.HandleWebSocket)

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/build/")))

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		ExposedHeaders:   []string{"Authorization"},
	})

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