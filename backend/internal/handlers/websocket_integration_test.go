package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/handlers"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil"
	ws "github.com/ctclostio/DnD-Game/backend/internal/websocket"
)

func TestWebSocketHandlerIntegration(t *testing.T) {
	// Set development environment for origin validation
	origEnv := os.Getenv("GO_ENV")
	require.NoError(t, os.Setenv("GO_ENV", "development"))
	defer func() { _ = os.Setenv("GO_ENV", origEnv) }()

	// Setup test context
	testCtx, cleanup := testutil.SetupIntegrationTest(t)
	defer cleanup()

	// Setup handlers with WebSocket hub
	_, _ = handlers.SetupTestHandlers(t, testCtx)

	// Create test user
	username := "ws_user"
	email := "ws_user@example.com"
	password := "password123"
	userID := testCtx.CreateTestUser(username, email, password)

	// Get the user from DB for the user object
	user, err := testCtx.Repos.Users.GetByID(context.Background(), userID)
	require.NoError(t, err)

	// Create game session
	sessionID := testCtx.CreateTestGameSession(userID, "ws_game", "WSGAME01")

	// Get the session from DB
	session, err := testCtx.Repos.GameSessions.GetByID(context.Background(), sessionID)
	require.NoError(t, err)

	// Login to get token
	tokenPair, err := testCtx.JWTManager.GenerateTokenPair(userID, username, email, "player")
	require.NoError(t, err)
	token := tokenPair.AccessToken

	t.Run("successful connection and authentication", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
		defer server.Close()

		// Convert http to ws URL
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		wsURL = fmt.Sprintf(constants.WebSocketURLFormat, wsURL, session.ID)

		// Connect to WebSocket with proper origin header
		header := http.Header{}
		header.Set("Origin", constants.LocalhostURL)
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer func() {
			if err := conn.Close(); err != nil {
				t.Logf(constants.WebSocketCloseError, err)
			}
		}()

		// Read auth required message
		var authReq map[string]string
		err = conn.ReadJSON(&authReq)
		require.NoError(t, err)
		assert.Equal(t, "auth_required", authReq["type"])

		// Send authentication
		authMsg := ws.AuthMessage{
			Type:  "auth",
			Token: token,
			Room:  session.ID,
		}
		err = conn.WriteJSON(authMsg)
		require.NoError(t, err)

		// Read auth success
		var authResp map[string]string
		err = conn.ReadJSON(&authResp)
		require.NoError(t, err)
		assert.Equal(t, "auth_success", authResp["type"])
		assert.Equal(t, user.Username, authResp["username"])
	})

	t.Run("authentication failure with invalid token", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
		defer server.Close()

		// Convert http to ws URL
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Connect to WebSocket with proper origin header
		header := http.Header{}
		header.Set("Origin", constants.LocalhostURL)
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer func() {
			if err := conn.Close(); err != nil {
				t.Logf(constants.WebSocketCloseError, err)
			}
		}()

		// Read auth required message
		var authReq map[string]string
		err = conn.ReadJSON(&authReq)
		require.NoError(t, err)

		// Send invalid authentication
		authMsg := ws.AuthMessage{
			Type:  "auth",
			Token: "invalid_token",
			Room:  session.ID,
		}
		err = conn.WriteJSON(authMsg)
		require.NoError(t, err)

		// Read error response
		var errorResp map[string]string
		err = conn.ReadJSON(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "error", errorResp["type"])
		assert.Equal(t, "Invalid token", errorResp["error"])
	})

	t.Run("authentication failure with missing room", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
		defer server.Close()

		// Convert http to ws URL (no room parameter)
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Connect to WebSocket with proper origin header
		header := http.Header{}
		header.Set("Origin", constants.LocalhostURL)
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer func() {
			if err := conn.Close(); err != nil {
				t.Logf(constants.WebSocketCloseError, err)
			}
		}()

		// Read auth required message
		var authReq map[string]string
		err = conn.ReadJSON(&authReq)
		require.NoError(t, err)

		// Send authentication without room
		authMsg := ws.AuthMessage{
			Type:  "auth",
			Token: token,
			// Room field is empty
		}
		err = conn.WriteJSON(authMsg)
		require.NoError(t, err)

		// Read error response
		var errorResp map[string]string
		err = conn.ReadJSON(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "error", errorResp["type"])
		assert.Equal(t, "Room ID required", errorResp["error"])
	})

	t.Run("message broadcasting", func(t *testing.T) {
		testMessageBroadcasting(t, testCtx, sessionID, session, user, userID, token)
	})

	t.Run("room isolation", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
		defer server.Close()

		// Create another game session
		sessionID2 := testCtx.CreateTestGameSession(userID, "ws_game2", "WSGAME02")
		session2, err := testCtx.Repos.GameSessions.GetByID(context.Background(), sessionID2)
		require.NoError(t, err)

		// Setup headers with origin
		header := http.Header{}
		header.Set("Origin", constants.LocalhostURL)

		// Connect to room 1
		wsURL1 := "ws" + strings.TrimPrefix(server.URL, "http")
		wsURL1 = fmt.Sprintf(constants.WebSocketURLFormat, wsURL1, session.ID)
		conn1, _, err := websocket.DefaultDialer.Dial(wsURL1, header)
		require.NoError(t, err)
		defer func() { _ = conn1.Close() }()

		// Authenticate in room 1
		var authReq1 map[string]string
		err = conn1.ReadJSON(&authReq1)
		require.NoError(t, err)

		authMsg1 := ws.AuthMessage{
			Type:  "auth",
			Token: token,
			Room:  session.ID,
		}
		err = conn1.WriteJSON(authMsg1)
		require.NoError(t, err)

		var authResp1 map[string]string
		err = conn1.ReadJSON(&authResp1)
		require.NoError(t, err)

		// Connect to room 2
		wsURL2 := "ws" + strings.TrimPrefix(server.URL, "http")
		wsURL2 = fmt.Sprintf(constants.WebSocketURLFormat, wsURL2, session2.ID)
		conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, header)
		require.NoError(t, err)
		defer func() { _ = conn2.Close() }()

		// Authenticate in room 2
		var authReq2 map[string]string
		err = conn2.ReadJSON(&authReq2)
		require.NoError(t, err)

		authMsg2 := ws.AuthMessage{
			Type:  "auth",
			Token: token,
			Room:  session2.ID,
		}
		err = conn2.WriteJSON(authMsg2)
		require.NoError(t, err)

		var authResp2 map[string]string
		err = conn2.ReadJSON(&authResp2)
		require.NoError(t, err)

		// Send message to room 1
		message := map[string]interface{}{
			"type":     "message",
			"roomId":   session.ID,
			"content":  "Room 1 message",
			"username": user.Username,
			"playerId": userID,
			"role":     "player",
			"data":     json.RawMessage(`{"content": "Room 1 message"}`),
		}
		err = conn1.WriteJSON(message)
		require.NoError(t, err)

		// Room 2 should not receive the message
		_ = conn2.SetReadDeadline(time.Now().Add(1 * time.Second))
		var received map[string]interface{}
		err = conn2.ReadJSON(&received)
		assert.Error(t, err) // Should timeout
	})

	t.Run("reconnection handling", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		wsURL = fmt.Sprintf(constants.WebSocketURLFormat, wsURL, session.ID)

		// Setup headers with origin
		header := http.Header{}
		header.Set("Origin", constants.LocalhostURL)

		// First connection
		conn1, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		require.NoError(t, err)

		// Authenticate
		var authReq map[string]string
		err = conn1.ReadJSON(&authReq)
		require.NoError(t, err)

		authMsg := ws.AuthMessage{
			Type:  "auth",
			Token: token,
			Room:  session.ID,
		}
		err = conn1.WriteJSON(authMsg)
		require.NoError(t, err)

		var authResp map[string]string
		err = conn1.ReadJSON(&authResp)
		require.NoError(t, err)

		// Close first connection
		_ = conn1.Close()

		// Wait a bit
		time.Sleep(100 * time.Millisecond)

		// Reconnect
		conn2, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer func() { _ = conn2.Close() }()

		// Authenticate again
		err = conn2.ReadJSON(&authReq)
		require.NoError(t, err)

		err = conn2.WriteJSON(authMsg)
		require.NoError(t, err)

		err = conn2.ReadJSON(&authResp)
		require.NoError(t, err)
		assert.Equal(t, "auth_success", authResp["type"])
	})

	t.Run("concurrent connections", func(t *testing.T) {
		testConcurrentConnections(t, testCtx, sessionID, session)
	})

	t.Run("origin validation", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Test with invalid origin (should fail in production mode)
		require.NoError(t, os.Setenv("GO_ENV", "production"))
		defer func() { _ = os.Setenv("GO_ENV", "development") }()

		header := http.Header{}
		header.Set("Origin", "http://evil.com")

		_, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
		assert.Error(t, err)
		if resp != nil {
			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		}
	})
}

// Helper functions to reduce cognitive complexity

// testMessageBroadcasting tests message broadcasting between users
func testMessageBroadcasting(t *testing.T, testCtx *testutil.IntegrationTestContext, sessionID string, session interface{}, user interface{}, userID string, token string) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
	defer server.Close()

	// Create second user
	user2ID := testCtx.CreateTestUser("ws_user2", "ws_user2@example.com", "password123")
	_, err := testCtx.Repos.Users.GetByID(context.Background(), user2ID)
	require.NoError(t, err)

	tokenPair2, err := testCtx.JWTManager.GenerateTokenPair(user2ID, "ws_user2", "ws_user2@example.com", "player")
	require.NoError(t, err)
	token2 := tokenPair2.AccessToken

	// Setup connections
	conn1, conn2 := setupTwoUserConnections(t, server, sessionID, token, token2)
	defer func() { _ = conn1.Close() }()
	defer func() { _ = conn2.Close() }()

	// Send message from user1
	message := buildTestMessage(sessionID, "Hello from user1", getUsernameFromUser(user), userID)
	err = conn1.WriteJSON(message)
	require.NoError(t, err)

	// User2 should receive the message
	verifyMessageReceived(t, conn2, "Hello from user1", getUsernameFromUser(user))
}

// setupTwoUserConnections establishes authenticated connections for two users
func setupTwoUserConnections(t *testing.T, server *httptest.Server, sessionID, token1, token2 string) (*websocket.Conn, *websocket.Conn) {
	header := http.Header{}
	header.Set("Origin", constants.LocalhostURL)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	wsURL = fmt.Sprintf(constants.WebSocketURLFormat, wsURL, sessionID)

	// Connect and authenticate first user
	conn1 := connectAndAuthenticate(t, wsURL, header, token1, sessionID)

	// Connect and authenticate second user
	conn2 := connectAndAuthenticate(t, wsURL, header, token2, sessionID)

	return conn1, conn2
}

// connectAndAuthenticate connects to WebSocket and authenticates
func connectAndAuthenticate(t *testing.T, wsURL string, header http.Header, token, roomID string) *websocket.Conn {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	require.NoError(t, err)

	// Read auth required
	var authReq map[string]string
	err = conn.ReadJSON(&authReq)
	require.NoError(t, err)

	// Send auth
	authMsg := ws.AuthMessage{
		Type:  "auth",
		Token: token,
		Room:  roomID,
	}
	err = conn.WriteJSON(authMsg)
	require.NoError(t, err)

	// Skip auth success
	var authResp map[string]string
	err = conn.ReadJSON(&authResp)
	require.NoError(t, err)

	return conn
}

// buildTestMessage creates a test message
func buildTestMessage(sessionID, content, username, userID string) map[string]interface{} {
	return map[string]interface{}{
		"type":     "message",
		"roomId":   sessionID,
		"content":  content,
		"username": username,
		"playerId": userID,
		"role":     "player",
		"data":     json.RawMessage(fmt.Sprintf(`{"content": "%s"}`, content)),
	}
}

// verifyMessageReceived verifies that a message was received correctly
func verifyMessageReceived(t *testing.T, conn *websocket.Conn, expectedContent, expectedUsername string) {
	var received map[string]interface{}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	err := conn.ReadJSON(&received)
	require.NoError(t, err)
	assert.Equal(t, "message", received["type"])
	assert.Equal(t, expectedContent, received["content"])
	assert.Equal(t, expectedUsername, received["username"])
}

// getUsernameFromUser extracts username from user interface
func getUsernameFromUser(user interface{}) string {
	// Try to access Username field using reflection or type assertion
	if u, ok := user.(interface{ Username string }); ok {
		return u.Username
	}
	// If it has a Username field
	type userStruct struct{ Username string }
	if u, ok := user.(*userStruct); ok {
		return u.Username
	}
	// Try models.User type
	if u, ok := user.(*struct{ Username string }); ok {
		return u.Username
	}
	return "unknown"
}

// testConcurrentConnections tests handling of multiple concurrent connections
func testConcurrentConnections(t *testing.T, testCtx *testutil.IntegrationTestContext, sessionID string, session interface{}) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	wsURL = fmt.Sprintf(constants.WebSocketURLFormat, wsURL, sessionID)

	// Create multiple users
	numUsers := 5
	tokens := createMultipleTestUsers(t, testCtx, numUsers)

	// Connect all users concurrently
	connections := connectUsersConcurrently(t, wsURL, sessionID, tokens)

	// Close all connections
	for _, conn := range connections {
		if conn != nil {
			if err := conn.Close(); err != nil {
				t.Logf(constants.WebSocketCloseError, err)
			}
		}
	}
}

// createMultipleTestUsers creates multiple test users and returns their tokens
func createMultipleTestUsers(t *testing.T, testCtx *testutil.IntegrationTestContext, numUsers int) []string {
	tokens := make([]string, numUsers)

	for i := 0; i < numUsers; i++ {
		username := fmt.Sprintf("concurrent_user_%d", i)
		email := fmt.Sprintf("concurrent_user_%d@example.com", i)
		uid := testCtx.CreateTestUser(username, email, "password123")

		tp, err := testCtx.JWTManager.GenerateTokenPair(uid, username, email, "player")
		require.NoError(t, err)
		tokens[i] = tp.AccessToken
	}

	return tokens
}

// connectUsersConcurrently connects multiple users concurrently
func connectUsersConcurrently(t *testing.T, wsURL, sessionID string, tokens []string) []*websocket.Conn {
	var wg sync.WaitGroup
	connections := make([]*websocket.Conn, len(tokens))
	header := http.Header{}
	header.Set("Origin", constants.LocalhostURL)

	for i := 0; i < len(tokens); i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			conn := authenticateUser(t, wsURL, header, tokens[idx], sessionID, idx)
			connections[idx] = conn
		}(i)
	}

	wg.Wait()
	return connections
}

// authenticateUser connects and authenticates a single user
func authenticateUser(t *testing.T, wsURL string, header http.Header, token, sessionID string, userIdx int) *websocket.Conn {
	// Connect
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Errorf("Failed to connect user %d: %v", userIdx, err)
		return nil
	}

	// Authenticate
	var authReq map[string]string
	err = conn.ReadJSON(&authReq)
	if err != nil {
		t.Errorf("Failed to read auth request for user %d: %v", userIdx, err)
		return conn
	}

	authMsg := ws.AuthMessage{
		Type:  "auth",
		Token: token,
		Room:  sessionID,
	}
	err = conn.WriteJSON(authMsg)
	if err != nil {
		t.Errorf("Failed to send auth for user %d: %v", userIdx, err)
		return conn
	}

	var authResp map[string]string
	err = conn.ReadJSON(&authResp)
	if err != nil {
		t.Errorf("Failed to read auth response for user %d: %v", userIdx, err)
		return conn
	}

	if authResp["type"] != "auth_success" {
		t.Errorf("Authentication failed for user %d", userIdx)
	}

	return conn
}
