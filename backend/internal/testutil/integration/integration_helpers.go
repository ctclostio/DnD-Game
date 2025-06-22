package integration

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

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/config"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

// IntegrationTestContext contains all dependencies for integration testing
type IntegrationTestContext struct {
	T          *testing.T
	DB         *database.DB
	SQLXDB     *sqlx.DB
	Repos      *database.Repositories
	Services   interface{} // Avoid import cycle - will be *services.Services
	Router     *mux.Router
	JWTManager *auth.JWTManager
	WSHub      interface{} // Avoid import cycle - will be *websocket.Hub
	Logger     *logger.LoggerV2
	Config     *config.Config
}

// IntegrationTestOptions allows customization of test setup
type IntegrationTestOptions struct {
	SkipAuth      bool
	SkipWebSocket bool
	CustomRoutes  func(*mux.Router, *IntegrationTestContext)
	MockLLM       interface{} // Will be LLMProvider
}

// SetupIntegrationTest creates a complete test environment
func SetupIntegrationTest(t *testing.T, opts ...IntegrationTestOptions) (ctx *IntegrationTestContext, cleanup func()) {
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
	log, err := logger.NewV2(&logConfig)
	require.NoError(t, err)

	// Setup test database
	sqlxDB := testutil.SetupTestDB(t)
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
	repos.WorldBuilding = database.NewWorldBuildingRepository(db)
	repos.Narrative = database.NewNarrativeRepository(sqlxDB)
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

	// Services will be created by the test if needed
	// This avoids import cycles with the services package

	// WebSocket hub - create without importing websocket package to avoid cycle
	var hub interface{}
	if !options.SkipWebSocket {
		// Hub will be created in the test that needs it
		hub = nil
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
			Services:   nil, // Will be set by test if needed
			Router:     router,
			JWTManager: jwtManager,
			WSHub:      hub,
			Logger:     log,
			Config:     cfg,
		})
	}

	cleanup = func() {
		_ = sqlxDB.Close()
	}

	return &IntegrationTestContext{
		T:          t,
		DB:         db,
		SQLXDB:     sqlxDB,
		Repos:      repos,
		Services:   nil, // Will be set by test if needed
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

// DecodeResponseData decodes the data field from a wrapped response into the provided interface
func (ctx *IntegrationTestContext) DecodeResponseData(w *httptest.ResponseRecorder, v interface{}) {
	var resp response.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(ctx.T, err, "Failed to decode response wrapper: %s", w.Body.String())
	require.True(ctx.T, resp.Success, "Expected success response, got error: %v", resp.Error)

	// Marshal the data back to JSON then unmarshal into the target type
	// This handles the case where resp.Data is a map[string]interface{}
	dataBytes, err := json.Marshal(resp.Data)
	require.NoError(ctx.T, err, "Failed to marshal response data")

	err = json.Unmarshal(dataBytes, v)
	require.NoError(ctx.T, err, "Failed to unmarshal response data into target type")
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
func (ctx *IntegrationTestContext) CreateTestUser(username, email, _ string) string {
	userID := "user-" + username
	hashedPassword := "$2a$10$test-hash" // NOSONAR - Mock bcrypt hash for testing
	query := `INSERT INTO users (id, username, email, password_hash, role) VALUES (?, ?, ?, ?, ?)`
	_, err := ctx.SQLXDB.Exec(ctx.SQLXDB.Rebind(query), userID, username, email, hashedPassword, "player")
	require.NoError(ctx.T, err)
	return userID
}

// CreateTestCharacter creates a test character and returns the character ID
func (ctx *IntegrationTestContext) CreateTestCharacter(userID, name string) string {
	charID := "char-" + name
	testutil.SeedTestCharacter(ctx.T, ctx.SQLXDB, charID, userID, name)
	return charID
}

// CreateTestGameSession creates a test game session and returns the session ID
func (ctx *IntegrationTestContext) CreateTestGameSession(dmUserID, name, code string) string {
	sessionID := "session-" + code
	query := `INSERT INTO game_sessions (id, name, description, dm_user_id, code, is_active) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := ctx.SQLXDB.Exec(ctx.SQLXDB.Rebind(query), sessionID, name, "Test session", dmUserID, code, true)
	require.NoError(ctx.T, err)

	// Update max_players if the column exists
	updateQuery := `UPDATE game_sessions SET max_players = 6 WHERE id = ?`
	_, _ = ctx.SQLXDB.Exec(ctx.SQLXDB.Rebind(updateQuery), sessionID)

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
	testutil.TruncateTables(ctx.T, ctx.SQLXDB)
}
