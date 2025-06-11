package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/config"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/websocket"
	"github.com/your-username/dnd-game/backend/pkg/logger"
	"github.com/your-username/dnd-game/backend/pkg/response"
)

// IntegrationTestContext contains all dependencies for integration testing
type IntegrationTestContext struct {
	T          *testing.T
	DB         *database.DB
	SQLXDB     *sqlx.DB
	Repos      *database.Repositories
	Services   *services.Services
	Router     *mux.Router
	JWTManager *auth.JWTManager
	WSHub      *websocket.Hub
	Logger     *logger.LoggerV2
	Config     *config.Config
}

// IntegrationTestOptions allows customization of test setup
type IntegrationTestOptions struct {
	SkipAuth      bool
	SkipWebSocket bool
	CustomRoutes  func(*mux.Router, *IntegrationTestContext)
	MockLLM       services.LLMProvider
}

// SetupIntegrationTest creates a complete test environment
func SetupIntegrationTest(t *testing.T, opts ...IntegrationTestOptions) (*IntegrationTestContext, func()) {
	var options IntegrationTestOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	// Setup test logger
	logConfig := logger.ConfigV2{
		Level:       "debug",
		Pretty:      true,
		ServiceName: "test",
		Environment: "test",
	}
	log, err := logger.NewV2(logConfig)
	require.NoError(t, err)

	// Setup test database
	sqlxDB := SetupTestDB(t)
	db := &database.DB{DB: sqlxDB}

	// Create repositories
	repos := &database.Repositories{
		Users:           database.NewUserRepository(db),
		Characters:      database.NewCharacterRepository(db),
		GameSessions:    database.NewGameSessionRepository(db),
		DiceRolls:       database.NewDiceRollRepository(db),
		RefreshTokens:   database.NewRefreshTokenRepository(sqlxDB),
		NPCs:            database.NewNPCRepository(sqlxDB),
		Inventory:       database.NewInventoryRepository(db),
		CustomRaces:     database.NewCustomRaceRepository(sqlxDB),
		CustomClasses:   database.NewCustomClassRepository(db),
		DMAssistant:     database.NewDMAssistantRepository(sqlxDB),
		Encounters:      database.NewEncounterRepository(db),
		Campaign:        database.NewCampaignRepository(sqlxDB),
		CombatAnalytics: database.NewCombatAnalyticsRepository(sqlxDB),
		RuleBuilder:     database.NewRuleBuilderRepository(db),
	}

	// Create world building repositories
	worldBuildingRepo := database.NewWorldBuildingRepository(db)
	_ = database.NewNarrativeRepository(sqlxDB) // narrativeRepo - can be used if needed
	_ = database.NewEmergentWorldRepository(db) // emergentWorldRepo - can be used if needed

	// Create test config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:        "8080",
			Environment: "test",
		},
		Database: config.DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "test",
			Password:     "test",
			DatabaseName: "test",
			SSLMode:      "disable",
		},
		Auth: config.AuthConfig{
			JWTSecret:            "test-secret",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 24 * time.Hour,
		},
		AI: config.AIConfig{
			Provider: "mock",
			Enabled:  false,
		},
	}

	// JWT Manager
	jwtManager := auth.NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.AccessTokenDuration,
		cfg.Auth.RefreshTokenDuration,
	)

	// LLM Provider
	var llmProvider services.LLMProvider
	if options.MockLLM != nil {
		llmProvider = options.MockLLM
	} else {
		llmProvider = &services.MockLLMProvider{}
	}

	// Create services
	userService := services.NewUserService(repos.Users)
	refreshTokenService := services.NewRefreshTokenService(repos.RefreshTokens, jwtManager)
	
	// Create AI services with enhanced logger
	aiConfig := &services.AIConfig{Enabled: cfg.AI.Provider != "mock"}
	aiBattleMapGen := services.NewAIBattleMapGenerator(llmProvider, aiConfig, log)
	aiCampaignManager := services.NewAICampaignManager(llmProvider, aiConfig, log)
	
	// Create event bus with logger
	_ = services.NewEventBus(log) // eventBus - can be used if needed
	
	// Combat services
	combatService := services.NewCombatService()
	combatAutomationService := services.NewCombatAutomationService(repos.CombatAnalytics, repos.Characters, repos.NPCs)
	combatAnalyticsService := services.NewCombatAnalyticsService(repos.CombatAnalytics, combatService)
	
	// World building services
	settlementGenerator := services.NewSettlementGeneratorService(llmProvider, worldBuildingRepo)
	factionSystem := services.NewFactionSystemService(llmProvider, worldBuildingRepo)
	worldEventEngine := services.NewWorldEventEngineService(llmProvider, worldBuildingRepo, factionSystem)
	economicSimulator := services.NewEconomicSimulatorService(worldBuildingRepo)
	
	// Rule engine services
	diceRollService := services.NewDiceRollService(repos.DiceRolls)
	ruleEngine := services.NewRuleEngine(repos.RuleBuilder, diceRollService)
	
	// Create service container
	svc := &services.Services{
		Users:              userService,
		Characters:         services.NewCharacterService(repos.Characters, repos.CustomClasses, llmProvider),
		GameSessions:       services.NewGameSessionService(repos.GameSessions),
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
		WorldBuilding:      worldBuildingRepo,
		RuleBuilder:        repos.RuleBuilder,
		AICampaignManager:  aiCampaignManager,
		BattleMapGen:       aiBattleMapGen,
	}

	// WebSocket hub
	var hub *websocket.Hub
	if !options.SkipWebSocket {
		hub = websocket.NewHub()
		go hub.Run()
	}

	// Setup router
	router := mux.NewRouter()

	// Apply middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add request ID for testing
			ctx := context.WithValue(r.Context(), response.RequestIDKey, "test-request-id")
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	})

	// Apply custom routes
	if options.CustomRoutes != nil {
		options.CustomRoutes(router, &IntegrationTestContext{
			T:          t,
			DB:         db,
			SQLXDB:     sqlxDB,
			Repos:      repos,
			Services:   svc,
			Router:     router,
			JWTManager: jwtManager,
			WSHub:      hub,
			Logger:     log,
			Config:     cfg,
		})
	}

	cleanup := func() {
		if hub != nil {
			// Hub will be cleaned up when the test ends
			// The goroutine will exit when the test process ends
		}
		sqlxDB.Close()
	}

	return &IntegrationTestContext{
		T:          t,
		DB:         db,
		SQLXDB:     sqlxDB,
		Repos:      repos,
		Services:   svc,
		Router:     router,
		JWTManager: jwtManager,
		WSHub:      hub,
		Logger:     log,
		Config:     cfg,
	}, cleanup
}


// AuthenticatedRequest creates an authenticated HTTP request
func (ctx *IntegrationTestContext) AuthenticatedRequest(method, url string, body interface{}, userID string) *http.Request {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(ctx.T, err)
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	// Generate token and add to request
	tokenPair, err := ctx.JWTManager.GenerateTokenPair(userID, "testuser", "test@example.com", "player")
	require.NoError(ctx.T, err)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)

	return req
}

// MakeRequest executes an HTTP request and returns the response
func (ctx *IntegrationTestContext) MakeRequest(req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)
	return w
}

// MakeAuthenticatedRequest is a convenience method combining AuthenticatedRequest and MakeRequest
func (ctx *IntegrationTestContext) MakeAuthenticatedRequest(method, url string, body interface{}, userID string) *httptest.ResponseRecorder {
	req := ctx.AuthenticatedRequest(method, url, body, userID)
	return ctx.MakeRequest(req)
}

// DecodeResponse decodes a JSON response into the provided interface
func (ctx *IntegrationTestContext) DecodeResponse(w *httptest.ResponseRecorder, v interface{}) {
	err := json.NewDecoder(w.Body).Decode(v)
	require.NoError(ctx.T, err, "Failed to decode response: %s", w.Body.String())
}

// AssertSuccessResponse asserts that the response is successful
func (ctx *IntegrationTestContext) AssertSuccessResponse(w *httptest.ResponseRecorder) response.Response {
	var resp response.Response
	ctx.DecodeResponse(w, &resp)
	require.True(ctx.T, resp.Success, "Expected success response, got error: %v", resp.Error)
	return resp
}

// AssertErrorResponse asserts that the response contains an error
func (ctx *IntegrationTestContext) AssertErrorResponse(w *httptest.ResponseRecorder, expectedMessage string) response.Response {
	var resp response.Response
	ctx.DecodeResponse(w, &resp)
	require.False(ctx.T, resp.Success, "Expected error response, got success")
	require.NotNil(ctx.T, resp.Error, "Expected error in response")
	if expectedMessage != "" {
		require.Contains(ctx.T, resp.Error.Message, expectedMessage, "Error message mismatch")
	}
	return resp
}

// CreateTestUser creates a test user and returns the user ID
func (ctx *IntegrationTestContext) CreateTestUser(username, email, password string) string {
	userID := "user-" + username
	hashedPassword := "$2a$10$test-hash" // Mock hash for testing
	query := `INSERT INTO users (id, username, email, password_hash, role) VALUES (?, ?, ?, ?, ?)`
	_, err := ctx.SQLXDB.Exec(ctx.SQLXDB.Rebind(query), userID, username, email, hashedPassword, "player")
	require.NoError(ctx.T, err)
	return userID
}

// CreateTestCharacter creates a test character and returns the character ID
func (ctx *IntegrationTestContext) CreateTestCharacter(userID, name string) string {
	charID := "char-" + name
	SeedTestCharacter(ctx.T, ctx.SQLXDB, charID, userID, name)
	return charID
}

// CreateTestGameSession creates a test game session and returns the session ID
func (ctx *IntegrationTestContext) CreateTestGameSession(dmUserID, name, code string) string {
	sessionID := "session-" + code
	query := `INSERT INTO game_sessions (id, name, description, dm_user_id, code, is_active) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := ctx.SQLXDB.Exec(ctx.SQLXDB.Rebind(query), sessionID, name, "Test session", dmUserID, code, true)
	require.NoError(ctx.T, err)
	return sessionID
}

// AddUserToSession adds a user to a game session
func (ctx *IntegrationTestContext) AddUserToSession(sessionID, userID string, characterID *string) {
	participantID := "participant-" + userID + "-" + sessionID
	query := `INSERT INTO game_participants (id, session_id, user_id, character_id, is_online) VALUES (?, ?, ?, ?, ?)`
	_, err := ctx.SQLXDB.Exec(ctx.SQLXDB.Rebind(query), participantID, sessionID, userID, characterID, false)
	require.NoError(ctx.T, err)
}

// CleanupDatabase truncates all tables for test isolation
func (ctx *IntegrationTestContext) CleanupDatabase() {
	TruncateTables(ctx.T, ctx.SQLXDB)
}