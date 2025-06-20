package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/handlers"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil"
)

// Test constants
const (
	testGameSessionsPath = "/api/v1/game/sessions/"
	skipShortModeMsg = "skipping integration test in short mode"
	skipIntegrationMsg = "integration environment not available"
)

// Test data structure to hold test setup
type testData struct {
	ctx       context.Context
	testCtx   *testutil.IntegrationTestContext
	h         *handlers.Handlers
	svc       *services.Services
	dm        *models.User
	player1   *models.User
	player2   *models.User
	player3   *models.User
	char1     *models.Character
	char2     *models.Character
	charHigh  *models.Character
	session   *models.GameSession
}

// setupTestData creates all test data needed for security tests
func setupTestData(t *testing.T) (*testData, func()) {
	ctx, testCtx, h, svc, cleanup := setupTestEnvironment(t)

	// Create test users
	dm := createTestUser(t, testCtx, "dm_user", "dm@example.com", "dm")
	player1 := createTestUser(t, testCtx, "player1", "player1@example.com", "player")
	player2 := createTestUser(t, testCtx, "player2", "player2@example.com", "player")
	player3 := createTestUser(t, testCtx, "player3", "player3@example.com", "player")

	// Create test characters
	char1 := createTestCharacter(t, testCtx, player1.ID, "Fighter", 5)
	char2 := createTestCharacter(t, testCtx, player2.ID, "Wizard", 3)
	charHighLevel := createTestCharacter(t, testCtx, player3.ID, "Paladin", 10)

	// Create test session
	session := &models.GameSession{
		DMID:                  dm.ID,
		Name:                  "Security Test Session",
		Description:           "Testing security features",
		MaxPlayers:            4,
		IsPublic:              false,
		RequiresInvite:        true,
		AllowedCharacterLevel: 5,
	}
	err := svc.GameSessions.CreateSession(ctx, session)
	require.NoError(t, err)

	return &testData{
		ctx:      ctx,
		testCtx:  testCtx,
		h:        h,
		svc:      svc,
		dm:       dm,
		player1:  player1,
		player2:  player2,
		player3:  player3,
		char1:    char1,
		char2:    char2,
		charHigh: charHighLevel,
		session:  session,
	}, cleanup
}

// Helper to verify error response
func verifyErrorResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedError string) {
	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	
	errMap, ok := response["error"].(map[string]interface{})
	require.True(t, ok, "Expected error object in response")
	assert.Contains(t, errMap["message"], expectedError)
}

func TestGameSessionSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip(skipShortModeMsg)
	}
	t.Skip(skipIntegrationMsg)

}

func TestJoinSessionSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip(skipShortModeMsg)
	}
	t.Skip(skipIntegrationMsg)

	td, cleanup := setupTestData(t)
	defer cleanup()

	// First, have player1 join the session
	body := map[string]interface{}{"character_id": td.char1.ID}
	req := createAuthenticatedRequest("POST", testGameSessionsPath+td.session.ID+"/join", body, td.player1.ID, td.session.ID)
	rr := executeRequest(td.h, req, td.h.JoinGameSession)
	require.Equal(t, http.StatusOK, rr.Code, "Player1 should be able to join initially")

	t.Run("Cannot join twice", func(t *testing.T) {
		body := map[string]interface{}{"character_id": td.char1.ID}
		req := createAuthenticatedRequest("POST", testGameSessionsPath+td.session.ID+"/join", body, td.player1.ID, td.session.ID)
		rr := executeRequest(td.h, req, td.h.JoinGameSession)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		verifyErrorResponse(t, rr, "already in this session")
	})

	t.Run("Cannot join with another user's character", func(t *testing.T) {
		body := map[string]interface{}{"character_id": td.char1.ID} // Belongs to player1
		req := createAuthenticatedRequest("POST", testGameSessionsPath+td.session.ID+"/join", body, td.player2.ID, td.session.ID)
		rr := executeRequest(td.h, req, td.h.JoinGameSession)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		verifyErrorResponse(t, rr, "don't own this character")
	})

	t.Run("Cannot join with high level character", func(t *testing.T) {
		// First leave the session
		_ = td.svc.GameSessions.LeaveSession(td.ctx, td.session.ID, td.player1.ID)

		body := map[string]interface{}{"character_id": td.charHigh.ID}
		req := createAuthenticatedRequest("POST", testGameSessionsPath+td.session.ID+"/join", body, td.player1.ID, td.session.ID)
		rr := executeRequest(td.h, req, td.h.JoinGameSession)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		verifyErrorResponse(t, rr, "exceeds session limit")
	})

	t.Run("Can join without character", func(t *testing.T) {
		body := map[string]interface{}{}
		req := createAuthenticatedRequest("POST", testGameSessionsPath+td.session.ID+"/join", body, td.player2.ID, td.session.ID)
		rr := executeRequest(td.h, req, td.h.JoinGameSession)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Cannot join full session", func(t *testing.T) {
		// Fill the session to capacity
		td.session.MaxPlayers = 2 // DM + 1 player
		_ = td.testCtx.Repos.GameSessions.Update(td.ctx, td.session)

		body := map[string]interface{}{}
		req := createAuthenticatedRequest("POST", testGameSessionsPath+td.session.ID+"/join", body, "another_user", td.session.ID)
		rr := executeRequest(td.h, req, td.h.JoinGameSession)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		verifyErrorResponse(t, rr, "session is full")
	})

}

func TestGetGameSessionAuthorization(t *testing.T) {
	if testing.Short() {
		t.Skip(skipShortModeMsg)
	}
	t.Skip(skipIntegrationMsg)

	td, cleanup := setupTestData(t)
	defer cleanup()

	// Ensure player1 can join the session
	_ = td.svc.GameSessions.LeaveSession(td.ctx, td.session.ID, td.player1.ID)
	td.session.MaxPlayers = 4
	err := td.testCtx.Repos.GameSessions.Update(td.ctx, td.session)
	require.NoError(t, err)
	_ = td.svc.GameSessions.JoinSession(td.ctx, td.session.ID, td.player1.ID, &td.char1.ID)

	// Create private session
	privateSession := &models.GameSession{
		DMID:        td.dm.ID,
		Name:        "Private Session",
		Description: "Should not be visible to non-participants",
		IsPublic:    false,
	}
	err = td.svc.GameSessions.CreateSession(td.ctx, privateSession)
	require.NoError(t, err)

	t.Run("DM can view their session", func(t *testing.T) {
		req := createAuthenticatedRequest("GET", testGameSessionsPath+privateSession.ID, nil, td.dm.ID, privateSession.ID)
		rr := executeRequest(td.h, req, td.h.GetGameSession)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Non-participant cannot view private session", func(t *testing.T) {
		req := createAuthenticatedRequest("GET", testGameSessionsPath+privateSession.ID, nil, td.player1.ID, privateSession.ID)
		rr := executeRequest(td.h, req, td.h.GetGameSession)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("Participant can view session", func(t *testing.T) {
		req := createAuthenticatedRequest("GET", testGameSessionsPath+td.session.ID, nil, td.player1.ID, td.session.ID)
		rr := executeRequest(td.h, req, td.h.GetGameSession)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

}

// Helper to create kick player request
func createKickPlayerRequest(sessionID, playerToKick, userID string) *http.Request {
	body := map[string]interface{}{}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", testGameSessionsPath+sessionID+"/kick/"+playerToKick, bytes.NewBuffer(jsonBody))
	req.Header.Set(constants.ContentType, constants.ApplicationJSON)
	claims := &auth.Claims{
		UserID:   userID,
		Username: "test",
		Email:    handlers.TestEmail,
		Role:     "dm",
	}
	req = req.WithContext(context.WithValue(req.Context(), auth.UserContextKey, claims))
	req = mux.SetURLVars(req, map[string]string{
		"id":       sessionID,
		"playerId": playerToKick,
	})
	return req
}

func TestKickPlayerSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip(skipShortModeMsg)
	}
	t.Skip(skipIntegrationMsg)

	td, cleanup := setupTestData(t)
	defer cleanup()

	// Ensure player2 is in the session for kick testing
	_ = td.svc.GameSessions.LeaveSession(td.ctx, td.session.ID, td.player2.ID)
	err := td.svc.GameSessions.JoinSession(td.ctx, td.session.ID, td.player2.ID, &td.char2.ID)
	require.NoError(t, err)

	t.Run("DM can kick player", func(t *testing.T) {
		req := createKickPlayerRequest(td.session.ID, td.player2.ID, td.dm.ID)
		rr := executeRequest(td.h, req, td.h.KickPlayer)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Non-DM cannot kick player", func(t *testing.T) {
		req := createKickPlayerRequest(td.session.ID, td.player2.ID, td.player1.ID)
		rr := executeRequest(td.h, req, td.h.KickPlayer)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		verifyErrorResponse(t, rr, "Only the DM can kick")
	})

	t.Run("DM cannot kick themselves", func(t *testing.T) {
		req := createKickPlayerRequest(td.session.ID, td.dm.ID, td.dm.ID)
		rr := executeRequest(td.h, req, td.h.KickPlayer)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		verifyErrorResponse(t, rr, "cannot kick themselves")
	})

	t.Run("Cannot kick non-existent player", func(t *testing.T) {
		req := createKickPlayerRequest(td.session.ID, "non_existent_id", td.dm.ID)
		rr := executeRequest(td.h, req, td.h.KickPlayer)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		verifyErrorResponse(t, rr, "not in this session")
	})

}

func TestSessionStateSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip(skipShortModeMsg)
	}
	t.Skip(skipIntegrationMsg)

	td, cleanup := setupTestData(t)
	defer cleanup()

	t.Run("Cannot join inactive session", func(t *testing.T) {
		// Create inactive session
		inactiveSession := &models.GameSession{
			DMID: td.dm.ID,
			Name: "Inactive Session",
		}
		err := td.svc.GameSessions.CreateSession(td.ctx, inactiveSession)
		require.NoError(t, err)

		// Update session to be inactive
		inactiveSession.IsActive = false
		err = td.testCtx.Repos.GameSessions.Update(td.ctx, inactiveSession)
		require.NoError(t, err)

		// Try to join inactive session
		req := createAuthenticatedRequest("POST", testGameSessionsPath+inactiveSession.ID+"/join", map[string]interface{}{}, td.player1.ID, inactiveSession.ID)
		rr := executeRequest(td.h, req, td.h.JoinGameSession)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		verifyErrorResponse(t, rr, "not active")
	})

	t.Run("Cannot join completed session", func(t *testing.T) {
		// Create completed session
		completedSession := &models.GameSession{
			DMID:   td.dm.ID,
			Name:   "Completed Session",
			Status: models.GameStatusCompleted,
		}
		err := td.svc.GameSessions.CreateSession(td.ctx, completedSession)
		require.NoError(t, err)

		// Try to join completed session
		req := createAuthenticatedRequest("POST", testGameSessionsPath+completedSession.ID+"/join", map[string]interface{}{}, td.player1.ID, completedSession.ID)
		rr := executeRequest(td.h, req, td.h.JoinGameSession)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		verifyErrorResponse(t, rr, "completed session")
	})
}

// Helper functions
func createTestUser(t *testing.T, ctx *testutil.IntegrationTestContext, username, email, role string) *models.User {
	user := &models.User{
		Username: username,
		Email:    email,
		Role:     role,
	}
	err := ctx.Repos.Users.Create(context.Background(), user)
	require.NoError(t, err)
	return user
}

func createTestCharacter(t *testing.T, ctx *testutil.IntegrationTestContext, userID, name string, level int) *models.Character {
	char := &models.Character{
		UserID: userID,
		Name:   name,
		Level:  level,
		Class:  "Fighter",
		Race:   "Human",
	}
	err := ctx.Repos.Characters.Create(context.Background(), char)
	require.NoError(t, err)
	return char
}

// Helper functions to reduce cognitive complexity

// setupTestEnvironment initializes the test environment
func setupTestEnvironment(t *testing.T) (context.Context, *testutil.IntegrationTestContext, *handlers.Handlers, *services.Services, func()) {
	ctx := context.Background()
	testCtx, cleanup := testutil.SetupIntegrationTest(t)

	// Create services with repositories
	gameService := services.NewGameSessionService(testCtx.Repos.GameSessions)
	gameService.SetCharacterRepository(testCtx.Repos.Characters)
	gameService.SetUserRepository(testCtx.Repos.Users)

	svc := &services.Services{
		DB:           testCtx.DB,
		Users:        services.NewUserService(testCtx.Repos.Users),
		Characters:   services.NewCharacterService(testCtx.Repos.Characters, nil, nil),
		GameSessions: gameService,
		JWTManager:   testCtx.JWTManager,
	}

	// Create handlers
	h := handlers.NewHandlers(svc, testCtx.DB, nil)

	return ctx, testCtx, h, svc, cleanup
}

// createAuthenticatedRequest creates a request with user authentication
func createAuthenticatedRequest(method, url string, body interface{}, userID, sessionID string) *http.Request {
	var bodyReader *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set(constants.ContentType, constants.ApplicationJSON)
	
	claims := &auth.Claims{
		UserID:   userID,
		Username: "testuser",
		Email:    handlers.TestEmail,
		Role:     "player",
	}
	req = req.WithContext(context.WithValue(req.Context(), auth.UserContextKey, claims))
	
	if sessionID != "" {
		req = mux.SetURLVars(req, map[string]string{"id": sessionID})
	}
	
	return req
}

// executeRequest executes a handler request and returns the response
func executeRequest(h *handlers.Handlers, req *http.Request, handler func(http.ResponseWriter, *http.Request)) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler(rr, req)
	return rr
}
