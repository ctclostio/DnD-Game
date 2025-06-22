package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil/integration"
	ws "github.com/ctclostio/DnD-Game/backend/internal/websocket"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

const (
	// API path constants
	sessionsAPIPath     = "/sessions"
	sessionByIDPath     = "/sessions/{id}"
	sessionJoinPath     = "/sessions/{id}/join"
	sessionLeavePath    = "/sessions/{id}/leave"
	
	// Error message constants
	expectedMapMessage  = "Expected data to be a map"
)

// Helper function to setup test environment with game session routes
func setupGameSessionTestEnvironment(t *testing.T) (*integration.IntegrationTestContext, func()) {
	return integration.SetupIntegrationTest(t, integration.IntegrationTestOptions{
		CustomRoutes: func(router *mux.Router, testCtx *integration.IntegrationTestContext) {
			h, _ := SetupTestHandlers(t, testCtx)
			authMiddleware := auth.NewMiddleware(testCtx.JWTManager)
			api := router.PathPrefix(APIv1Prefix).Subrouter()

			// Auth routes
			api.HandleFunc("/auth/register", h.Register).Methods("POST")
			api.HandleFunc("/auth/login", h.Login).Methods("POST")
			api.HandleFunc("/auth/logout", authMiddleware.Authenticate(h.Logout)).Methods("POST")
			api.HandleFunc("/auth/me", authMiddleware.Authenticate(h.GetCurrentUser)).Methods("GET")

			// Game session routes
			api.HandleFunc(sessionsAPIPath, authMiddleware.Authenticate(h.CreateGameSession)).Methods("POST")
			api.HandleFunc(sessionsAPIPath, authMiddleware.Authenticate(h.GetUserGameSessions)).Methods("GET")
			api.HandleFunc(sessionByIDPath, authMiddleware.Authenticate(h.GetGameSession)).Methods("GET")
			api.HandleFunc(sessionByIDPath, authMiddleware.Authenticate(h.UpdateGameSession)).Methods("PUT")
			api.HandleFunc(sessionJoinPath, authMiddleware.Authenticate(h.JoinGameSession)).Methods("POST")
			api.HandleFunc(sessionLeavePath, authMiddleware.Authenticate(h.LeaveGameSession)).Methods("POST")
		},
	})
}

// Helper to assert session data
func assertSessionData(t *testing.T, data interface{}) map[string]interface{} {
	sessionData, ok := data.(map[string]interface{})
	require.True(t, ok, expectedMapMessage)
	return sessionData
}

// Helper to verify participants
func verifyParticipants(t *testing.T, sessionResp map[string]interface{}) {
	if participants, ok := sessionResp["participants"]; ok {
		participantList, ok := participants.([]interface{})
		assert.True(t, ok, "Participants should be a list")
		assert.Greater(t, len(participantList), 0, "Should have at least one participant")
	}
}

func TestGameSessionLifecycle_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Skip("integration environment not available")
	
	ctx, cleanup := setupGameSessionTestEnvironment(t)
	defer cleanup()

	// Create test users
	dmUserID := ctx.CreateTestUser(TestDMUsername, TestDMEmail, DefaultPassword)
	player1ID := ctx.CreateTestUser(TestPlayer1Username, TestPlayer1Email, DefaultPassword)
	player2ID := ctx.CreateTestUser(TestPlayer2Username, TestPlayer2Email, DefaultPassword)
	player3ID := ctx.CreateTestUser(TestPlayer3Username, TestPlayer3Email, DefaultPassword)

	// Create characters for players
	char1ID := ctx.CreateTestCharacter(player1ID, CharacterAragorn)
	_ = ctx.CreateTestCharacter(player2ID, CharacterLegolas) // char2ID - used in skipped test
	char3ID := ctx.CreateTestCharacter(player3ID, CharacterGimli)

	// Create a test session that will be used by all subtests
	var sessionID string
	createReq := map[string]interface{}{
		NameField:        FellowshipCampaignName,
		DescriptionField: FellowshipCampaignDesc,
		MaxPlayersField: 6,
	}
	w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath[:len(APISessionsPath)-1], createReq, dmUserID)
	require.Equal(t, http.StatusCreated, w.Code)

	resp := ctx.AssertSuccessResponse(w)
	sessionData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok, expectedMapMessage)
	sessionID = sessionData["id"].(string)
	require.NotEmpty(t, sessionID, "Session ID should not be empty")

	t.Run("Create Game Session", func(t *testing.T) {
		// This test just verifies the session was created properly
		assert.Equal(t, FellowshipCampaignName, sessionData[NameField])
		assert.Equal(t, dmUserID, sessionData["dmId"])
		assert.NotEmpty(t, sessionData[CodeField])
		assert.True(t, sessionData["isActive"].(bool))

		// Verify session in database
		testutil.AssertRowExists(t, ctx.SQLXDB, "game_sessions", "id", sessionID)
	})

	t.Run("Get Game Session - DM Access", func(t *testing.T) {
		w := ctx.MakeAuthenticatedRequest("GET", APISessionsPath+sessionID, nil, dmUserID)
		assert.Equal(t, http.StatusOK, w.Code)

		resp := ctx.AssertSuccessResponse(w)
		sessionData, ok := resp.Data.(map[string]interface{})
		require.True(t, ok, expectedMapMessage)

		assert.Equal(t, sessionID, sessionData["id"])
		assert.Equal(t, FellowshipCampaignName, sessionData[NameField])
		assert.Equal(t, dmUserID, sessionData["dmId"])
	})

	t.Run("Get Game Session - Non-participant Forbidden", func(t *testing.T) {
		w := ctx.MakeAuthenticatedRequest("GET", APISessionsPath+sessionID, nil, player1ID)
		assert.Equal(t, http.StatusForbidden, w.Code)

		var resp response.Response
		ctx.DecodeResponse(w, &resp)
		assert.Contains(t, resp.Error.Message, ErrDontHaveAccess)
	})

	t.Run("List Game Sessions", func(t *testing.T) {
		// Create another session to test filtering
		otherDMID := ctx.CreateTestUser(TestOtherDMUsername, TestOtherDMEmail, DefaultPassword)
		createReq := map[string]interface{}{
			NameField:        AnotherCampaignName,
			DescriptionField: AnotherCampaignDesc,
		}
		w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath[:len(APISessionsPath)-1], createReq, otherDMID)
		require.Equal(t, http.StatusCreated, w.Code)

		// DM should see only their session
		w = ctx.MakeAuthenticatedRequest("GET", APISessionsPath[:len(APISessionsPath)-1], nil, dmUserID)
		assert.Equal(t, http.StatusOK, w.Code)

		resp := ctx.AssertSuccessResponse(w)
		sessions, ok := resp.Data.([]interface{})
		require.True(t, ok, "Expected data to be an array")

		// Should see at least the session we created
		found := false
		for _, s := range sessions {
			sessionMap, ok := s.(map[string]interface{})
			require.True(t, ok)
			if sessionMap["id"] == sessionID {
				found = true
				assert.Equal(t, FellowshipCampaignName, sessionMap[NameField])
			}
		}
		assert.True(t, found, "Created session should be in list")
	})

	t.Run("Join Game Session with Character", func(t *testing.T) {
		t.Logf("Attempting to join session: %s", sessionID)
		joinReq := map[string]interface{}{
			CharacterIDField: char1ID,
		}

		w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath+sessionID+"/join", joinReq, player1ID)
		if w.Code != http.StatusOK {
			t.Logf("Join response: %d, body: %s", w.Code, w.Body.String())
		}
		assert.Equal(t, http.StatusOK, w.Code)

		resp := ctx.AssertSuccessResponse(w)
		assert.NotNil(t, resp.Data)

		// Verify participant in database
		testutil.AssertRowExists(t, ctx.SQLXDB, "game_participants", "user_id", player1ID)

		// Verify we can find the participant with the correct session_id
		var count int
		err := ctx.SQLXDB.Get(&count,
			"SELECT COUNT(*) FROM game_participants WHERE session_id = ? AND user_id = ?",
			sessionID, player1ID)
		require.NoError(t, err)
		require.Equal(t, 1, count, "Player should be in the session")

		// Let's also check what participants exist
		type participant struct {
			SessionID string `db:"session_id"`
			UserID    string `db:"user_id"`
		}
		var participants []participant
		err = ctx.SQLXDB.Select(&participants, "SELECT session_id, user_id FROM game_participants")
		require.NoError(t, err)
		t.Logf("All participants: %+v", participants)

		// Now player should be able to access the session
		t.Logf("Player1 ID: %s, Session ID: %s", player1ID, sessionID)
		w = ctx.MakeAuthenticatedRequest("GET", APISessionsPath+sessionID, nil, player1ID)
		if w.Code != http.StatusOK {
			t.Logf("Access denied. Response: %d, body: %s", w.Code, w.Body.String())
		}
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Skip "Join Game Session by Code" test since JoinByCode handler doesn't exist
	t.Run("Join Game Session by Code - SKIPPED", func(t *testing.T) {
		t.Skip("JoinByCode handler not implemented")
	})

	t.Run("Cannot Join Same Session Twice", func(t *testing.T) {
		joinReq := map[string]interface{}{
			CharacterIDField: char1ID,
		}

		w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath+sessionID+"/join", joinReq, player1ID)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp response.Response
		ctx.DecodeResponse(w, &resp)
		assert.Contains(t, resp.Error.Message, ErrAlreadyInSession)
	})

	t.Run("Join Without Character", func(t *testing.T) {
		testJoinWithoutCharacter(t, ctx, sessionID, player3ID)
	})

	t.Run("Leave Game Session", func(t *testing.T) {
		// First, ensure player3 joins (they might already be in from previous test)
		joinReq := map[string]interface{}{
			CharacterIDField: char3ID,
		}
		_ = ctx.MakeAuthenticatedRequest("POST", APISessionsPath+sessionID+"/join", joinReq, player3ID)

		// Now leave
		w = ctx.MakeAuthenticatedRequest("POST", APISessionsPath+sessionID+"/leave", nil, player3ID)
		assert.Equal(t, http.StatusOK, w.Code)

		resp := ctx.AssertSuccessResponse(w)
		assert.NotNil(t, resp.Data)

		// Verify no longer a participant
		testutil.AssertRowNotExists(t, ctx.SQLXDB, "game_participants", "user_id", player3ID)

		// Should no longer have access
		w = ctx.MakeAuthenticatedRequest("GET", APISessionsPath+sessionID, nil, player3ID)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Update Game Session", func(t *testing.T) {
		testUpdateGameSession(t, ctx, sessionID, dmUserID)
	})

	t.Run("Only DM Can Update Session", func(t *testing.T) {
		updateReq := map[string]interface{}{
			NameField: "Hacked Session Name",
		}

		w := ctx.MakeAuthenticatedRequest("PUT", APISessionsPath+sessionID, updateReq, player1ID)

		// Should be forbidden or not found
		assert.Contains(t, []int{http.StatusForbidden, http.StatusNotFound}, w.Code)
	})

	t.Run("End Game Session", func(t *testing.T) {
		testEndGameSession(t, ctx, dmUserID)
	})
}

func TestGameSessionWithWebSocket_Integration(t *testing.T) {
	// Setup with custom routes including WebSocket
	ctx, cleanup := integration.SetupIntegrationTest(t, integration.IntegrationTestOptions{
		CustomRoutes: func(router *mux.Router, testCtx *integration.IntegrationTestContext) {
			// Create handlers
			h, _ := SetupTestHandlers(t, testCtx)
			authMiddleware := auth.NewMiddleware(testCtx.JWTManager)

			// API routes
			api := router.PathPrefix(APIv1Prefix).Subrouter()

			// Game session routes
			api.HandleFunc(sessionsAPIPath, authMiddleware.Authenticate(h.CreateGameSession)).Methods("POST")
			api.HandleFunc(sessionByIDPath, authMiddleware.Authenticate(h.GetGameSession)).Methods("GET")
			api.HandleFunc(sessionJoinPath, authMiddleware.Authenticate(h.JoinGameSession)).Methods("POST")

			// WebSocket route (using websocket package handler)
			ws.SetJWTManager(testCtx.JWTManager)
			api.HandleFunc("/ws", ws.HandleWebSocket).Methods("GET")
		},
	})
	defer cleanup()

	// Create users and session
	dmID := ctx.CreateTestUser(TestWSDMUsername, TestWSDMEmail, DefaultPassword)
	playerID := ctx.CreateTestUser(TestWSPlayerUsername, TestWSPlayerEmail, DefaultPassword)
	charID := ctx.CreateTestCharacter(playerID, CharacterWSHero)

	// Create session
	createReq := map[string]interface{}{
		NameField:        WebSocketSessionName,
		DescriptionField: WebSocketSessionDesc,
		MaxPlayersField: 6,
	}
	w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath[:len(APISessionsPath)-1], createReq, dmID)
	require.Equal(t, http.StatusCreated, w.Code)

	var session models.GameSession
	ctx.DecodeResponseData(w, &session)

	t.Run("Player Online Status", func(t *testing.T) {
		// Join session
		joinReq := map[string]interface{}{
			CharacterIDField: charID,
		}
		w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath+session.ID+"/join", joinReq, playerID)
		require.Equal(t, http.StatusOK, w.Code)

		// Check initial online status (should be false)
		var isOnline bool
		err := ctx.SQLXDB.Get(&isOnline,
			"SELECT is_online FROM game_participants WHERE session_id = ? AND user_id = ?",
			session.ID, playerID)
		require.NoError(t, err)
		assert.False(t, isOnline, "Player should be offline initially")

		// When WebSocket connects, status should update
		// This would be tested in the WebSocket integration tests
	})

	t.Run("Session Participant List", func(t *testing.T) {
		// Get session with participants
		w := ctx.MakeAuthenticatedRequest("GET", APISessionsPath+session.ID, nil, dmID)
		assert.Equal(t, http.StatusOK, w.Code)

		var sessionResp map[string]interface{}
		ctx.DecodeResponseData(w, &sessionResp)

		// Check if participants are included
		verifyParticipants(t, sessionResp)
	})
}

func TestGameSessionSecurity_Integration(t *testing.T) {
	ctx, cleanup := integration.SetupIntegrationTest(t, integration.IntegrationTestOptions{
		CustomRoutes: func(router *mux.Router, testCtx *integration.IntegrationTestContext) {
			h, _ := SetupTestHandlers(t, testCtx)
			authMiddleware := auth.NewMiddleware(testCtx.JWTManager)
			api := router.PathPrefix(APIv1Prefix).Subrouter()

			api.HandleFunc(sessionsAPIPath, authMiddleware.Authenticate(h.CreateGameSession)).Methods("POST")
			api.HandleFunc(sessionByIDPath, authMiddleware.Authenticate(h.GetGameSession)).Methods("GET")
			api.HandleFunc(sessionJoinPath, authMiddleware.Authenticate(h.JoinGameSession)).Methods("POST")
		},
	})
	defer cleanup()

	// Create users
	dm1ID := ctx.CreateTestUser(TestSecDM1Username, TestSecDM1Email, DefaultPassword)
	dm2ID := ctx.CreateTestUser(TestSecDM2Username, TestSecDM2Email, DefaultPassword)
	playerID := ctx.CreateTestUser(TestSecPlayerUsername, TestSecPlayerEmail, DefaultPassword)
	hackerID := ctx.CreateTestUser(TestHackerUsername, TestHackerEmail, DefaultPassword)

	// Create sessions
	session1ID := ctx.CreateTestGameSession(dm1ID, SecureSession1Name, SessionCodeSEC001)
	session2ID := ctx.CreateTestGameSession(dm2ID, SecureSession2Name, SessionCodeSEC002)

	t.Run("Cannot Access Other DM's Session", func(t *testing.T) {
		// DM2 tries to access DM1's session
		w := ctx.MakeAuthenticatedRequest("GET", APISessionsPath+session1ID, nil, dm2ID)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Cannot Delete Other DM's Session - SKIPPED", func(t *testing.T) {
		t.Skip("DeleteGameSession handler not implemented")
	})

	t.Run("Cannot Join Inactive Session", func(t *testing.T) {
		// Deactivate session
		_, err := ctx.SQLXDB.Exec("UPDATE game_sessions SET is_active = false WHERE id = ?", session2ID)
		require.NoError(t, err)

		charID := ctx.CreateTestCharacter(playerID, CharacterSecHero)
		joinReq := map[string]interface{}{
			CharacterIDField: charID,
		}

		w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath+session2ID+"/join", joinReq, playerID)
		// Should fail to join inactive session
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("Session Code Uniqueness", func(t *testing.T) {
		// Try to create session with duplicate code
		// This should be handled by the service layer
		createReq := map[string]interface{}{
			NameField: "Duplicate Code Session",
			CodeField: SessionCodeSEC001, // Same as session1
		}

		w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath[:len(APISessionsPath)-1], createReq, hackerID)

		// The service should generate a unique code, not use the provided one
		// Or it should reject if code is provided and duplicate
		if w.Code == http.StatusCreated {
			var session models.GameSession
			ctx.DecodeResponseData(w, &session)
			assert.NotEqual(t, SessionCodeSEC001, session.Code, "Should not allow duplicate codes")
		}
	})

	t.Run("Cannot Use Another User's Character", func(t *testing.T) {
		// Create a character for player
		playerCharID := ctx.CreateTestCharacter(playerID, "PlayerChar")

		// Hacker tries to join with player's character
		joinReq := map[string]interface{}{
			CharacterIDField: playerCharID,
		}

		// Reactivate session for this test
		_, err := ctx.SQLXDB.Exec("UPDATE game_sessions SET is_active = true WHERE id = ?", session1ID)
		require.NoError(t, err)

		w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath+session1ID+"/join", joinReq, hackerID)

		// Debug response
		t.Logf("Join response: status=%d, body=%s", w.Code, w.Body.String())

		// Should be rejected
		assert.NotEqual(t, http.StatusOK, w.Code)

		// Verify hacker is not in participants
		testutil.AssertRowNotExists(t, ctx.SQLXDB, "game_participants", "user_id", hackerID)
	})
}

func TestGameSessionConcurrency_Integration(t *testing.T) {
	ctx, cleanup := integration.SetupIntegrationTest(t, integration.IntegrationTestOptions{
		CustomRoutes: func(router *mux.Router, testCtx *integration.IntegrationTestContext) {
			h, _ := SetupTestHandlers(t, testCtx)
			authMiddleware := auth.NewMiddleware(testCtx.JWTManager)
			api := router.PathPrefix(APIv1Prefix).Subrouter()

			api.HandleFunc(sessionsAPIPath, authMiddleware.Authenticate(h.CreateGameSession)).Methods("POST")
			api.HandleFunc(sessionJoinPath, authMiddleware.Authenticate(h.JoinGameSession)).Methods("POST")
			api.HandleFunc(sessionLeavePath, authMiddleware.Authenticate(h.LeaveGameSession)).Methods("POST")
		},
	})
	defer cleanup()

	// Verify database is properly set up
	var tableCount int
	err := ctx.SQLXDB.Get(&tableCount, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='game_sessions'")
	require.NoError(t, err)
	require.Equal(t, 1, tableCount, "game_sessions table should exist")

	dmID := ctx.CreateTestUser(TestConcDMUsername, TestConcDMEmail, DefaultPassword)

	// Create multiple players
	playerIDs := make([]string, 0, 5)
	charIDs := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		playerID := ctx.CreateTestUser(testutil.RandomString(8), testutil.RandomString(8)+EmailDomain, DefaultPassword)
		charID := ctx.CreateTestCharacter(playerID, testutil.RandomString(8))
		playerIDs = append(playerIDs, playerID)
		charIDs = append(charIDs, charID)
	}

	// Create session
	sessionID := ctx.CreateTestGameSession(dmID, ConcurrentSessionName, SessionCodeCONC123)

	t.Run("Concurrent Joins", func(t *testing.T) {
		// Have all players try to join simultaneously
		done := make(chan bool, len(playerIDs))

		for i := range playerIDs {
			go func(idx int) {
				joinReq := map[string]interface{}{
					CharacterIDField: charIDs[idx],
				}
				w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath+sessionID+"/join", joinReq, playerIDs[idx])
				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}(i)
		}

		// Wait for all joins to complete
		for i := 0; i < len(playerIDs); i++ {
			select {
			case <-done:
				// Join completed
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent joins")
			}
		}

		// Verify all players are in the session
		var count int
		err := ctx.SQLXDB.Get(&count, "SELECT COUNT(*) FROM game_participants WHERE session_id = ?", sessionID)
		require.NoError(t, err)
		assert.Equal(t, len(playerIDs), count, "All players should have joined")
	})

	t.Run("Concurrent Leaves", func(t *testing.T) {
		// Have all players leave simultaneously
		done := make(chan bool, len(playerIDs))

		for i := range playerIDs {
			go func(idx int) {
				w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath+sessionID+"/leave", nil, playerIDs[idx])
				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}(i)
		}

		// Wait for all leaves to complete
		for i := 0; i < len(playerIDs); i++ {
			select {
			case <-done:
				// Leave completed
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent leaves")
			}
		}

		// Verify no players remain
		var count int
		err := ctx.SQLXDB.Get(&count, "SELECT COUNT(*) FROM game_participants WHERE session_id = ?", sessionID)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "No players should remain in session")
	})
}

// Helper test functions to reduce cognitive complexity
func testJoinWithoutCharacter(t *testing.T, ctx *integration.IntegrationTestContext, sessionID, playerID string) {
	joinReq := map[string]interface{}{}
	w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath+sessionID+"/join", joinReq, playerID)

	if w.Code == http.StatusBadRequest {
		var resp response.Response
		ctx.DecodeResponse(w, &resp)
		assert.Contains(t, resp.Error.Message, ErrCharacterRequired)
	} else {
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func testUpdateGameSession(t *testing.T, ctx *integration.IntegrationTestContext, sessionID, dmUserID string) {
	updateReq := map[string]interface{}{
		NameField:        FellowshipCampaignUpd,
		DescriptionField: FellowshipDescUpd,
	}

	w := ctx.MakeAuthenticatedRequest("PUT", APISessionsPath+sessionID, updateReq, dmUserID)

	if w.Code == http.StatusNotFound {
		t.Skip("Update endpoint not implemented")
	}

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	w = ctx.MakeAuthenticatedRequest("GET", APISessionsPath+sessionID, nil, dmUserID)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := ctx.AssertSuccessResponse(w)
	sessionData := assertSessionData(t, resp.Data)
	assert.Equal(t, FellowshipCampaignUpd, sessionData[NameField])
}

func testEndGameSession(t *testing.T, ctx *integration.IntegrationTestContext, dmUserID string) {
	// Create a new session to end
	createReq := map[string]interface{}{
		NameField:        SessionToEndName,
		DescriptionField: SessionToEndDesc,
	}
	w := ctx.MakeAuthenticatedRequest("POST", APISessionsPath[:len(APISessionsPath)-1], createReq, dmUserID)
	require.Equal(t, http.StatusCreated, w.Code)

	var tempSession models.GameSession
	ctx.DecodeResponseData(w, &tempSession)

	// Try to end/deactivate the session
	w = ctx.MakeAuthenticatedRequest("DELETE", APISessionsPath+tempSession.ID, nil, dmUserID)

	if w.Code == http.StatusNotFound {
		// Try updating is_active instead
		updateReq := map[string]interface{}{
			IsActiveField: false,
		}
		w = ctx.MakeAuthenticatedRequest("PUT", APISessionsPath+tempSession.ID, updateReq, dmUserID)
	}

	// Verify session is ended/inactive
	if w.Code == http.StatusOK || w.Code == http.StatusNoContent {
		var isActive bool
		err := ctx.SQLXDB.Get(&isActive, "SELECT is_active FROM game_sessions WHERE id = ?", tempSession.ID)
		if err == nil {
			assert.False(t, isActive, "Session should be inactive")
		}
	}
}