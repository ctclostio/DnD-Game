package handlers

import (
	"testing"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/config"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil"
	"github.com/ctclostio/DnD-Game/backend/internal/websocket"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// createTestServices creates a complete services instance for testing
func createTestServices(t *testing.T, repos *database.Repositories, jwtManager *auth.JWTManager, cfg *config.Config, log *logger.LoggerV2) *services.Services {
	// Create mock LLM provider
	llmProvider := &services.MockLLMProvider{}

	// Create services
	userService := services.NewUserService(repos.Users)
	refreshTokenService := services.NewRefreshTokenService(repos.RefreshTokens, jwtManager)

	// Create AI services with enhanced logger
	aiConfig := &services.AIConfig{Enabled: cfg.AI.Provider != "mock"}
	aiBattleMapGen := services.NewAIBattleMapGenerator(llmProvider, aiConfig, log)
	aiCampaignManager := services.NewAICampaignManager(llmProvider, aiConfig, log)

	// Create event bus with logger
	eventBus := services.NewEventBus(log) // eventBus - can be used if needed
	_ = eventBus // Mark as intentionally unused for now

	// Combat services
	combatService := services.NewCombatService()
	combatAutomationService := services.NewCombatAutomationService(repos.CombatAnalytics, repos.Characters, repos.NPCs)
	combatAnalyticsService := services.NewCombatAnalyticsService(repos.CombatAnalytics, combatService)

	// World building services
	settlementGenerator := services.NewSettlementGeneratorService(llmProvider, repos.WorldBuilding)
	factionSystem := services.NewFactionSystemService(llmProvider, repos.WorldBuilding)
	worldEventEngine := services.NewWorldEventEngineService(llmProvider, repos.WorldBuilding, factionSystem)
	economicSimulator := services.NewEconomicSimulatorService(repos.WorldBuilding)

	// Rule engine services
	diceRollService := services.NewDiceRollService(repos.DiceRolls)
	ruleEngine := services.NewRuleEngine(repos.RuleBuilder, diceRollService)

	// Create game session service with character repository
	gameSessionService := services.NewGameSessionService(repos.GameSessions)
	gameSessionService.SetCharacterRepository(repos.Characters)

	// Create service container
	return &services.Services{
		Users:              userService,
		Characters:         services.NewCharacterService(repos.Characters, repos.CustomClasses, llmProvider),
		GameSessions:       gameSessionService,
		DiceRolls:          diceRollService,
		Combat:             combatService,
		NPCs:               services.NewNPCService(repos.NPCs),
		Inventory:          services.NewInventoryService(repos.Inventory, repos.Characters),
		CustomRaces:        services.NewCustomRaceService(repos.CustomRaces, services.NewAIRaceGeneratorService(llmProvider)),
		DMAssistant:        services.NewDMAssistantService(repos.DMAssistant, services.NewAIDMAssistantService(llmProvider)),
		Encounters:         services.NewEncounterService(repos.Encounters, services.NewAIEncounterBuilder(llmProvider), combatService),
		Campaign:           services.NewCampaignService(repos.Campaign, repos.GameSessions, aiCampaignManager),
		CombatAutomation:   combatAutomationService,
		CombatAnalytics:    combatAnalyticsService,
		SettlementGen:      settlementGenerator,
		FactionSystem:      factionSystem,
		WorldEventEngine:   worldEventEngine,
		EconomicSim:        economicSimulator,
		RuleEngine:         ruleEngine,
		BalanceAnalyzer:    services.NewAIBalanceAnalyzer(cfg, llmProvider, ruleEngine, combatService),
		ConditionalReality: services.NewConditionalRealitySystem(ruleEngine),
		JWTManager:         jwtManager,
		RefreshTokens:      refreshTokenService,
		Config:             cfg,
		NarrativeEngine:    nil, // Can be set up if needed for specific tests
		WorldBuilding:      repos.WorldBuilding,
		RuleBuilder:        repos.RuleBuilder,
		AICampaignManager:  aiCampaignManager,
		BattleMapGen:       aiBattleMapGen,
	}
}

// SetupTestHandlers sets up handlers with all dependencies for integration testing
func SetupTestHandlers(t *testing.T, testCtx *testutil.IntegrationTestContext) (*Handlers, *websocket.Hub) {
	// Create WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Set JWT manager for WebSocket authentication
	websocket.SetJWTManager(testCtx.JWTManager)

	// Create services if not already provided
	var svc *services.Services
	if testCtx.Services != nil {
		// Type assertion if services were provided
		var ok bool
		svc, ok = testCtx.Services.(*services.Services)
		if !ok {
			t.Fatal("Services type assertion failed")
		}
	} else {
		// Create services
		svc = createTestServices(t, testCtx.Repos, testCtx.JWTManager, testCtx.Config, testCtx.Logger)
		testCtx.Services = svc
	}

	// Create handlers
	h := NewHandlers(svc, hub)

	return h, hub
}
