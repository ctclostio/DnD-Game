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
	"github.com/your-username/dnd-game/backend/internal/handlers"
	"github.com/your-username/dnd-game/backend/internal/testutil"
	ws "github.com/your-username/dnd-game/backend/internal/websocket"
)

func TestWebSocketHandlerIntegration(t *testing.T) {
	// Set development environment for origin validation
	origEnv := os.Getenv("GO_ENV")
	os.Setenv("GO_ENV", "development")
	defer os.Setenv("GO_ENV", origEnv)

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
		wsURL = fmt.Sprintf("%s?room=%s", wsURL, session.ID)

		// Connect to WebSocket with proper origin header
		header := http.Header{}
		header.Set("Origin", "http://localhost:3000")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer conn.Close()

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
		header.Set("Origin", "http://localhost:3000")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer conn.Close()

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
		header.Set("Origin", "http://localhost:3000")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer conn.Close()

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
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
		defer server.Close()

		// Create second user
		user2ID := testCtx.CreateTestUser("ws_user2", "ws_user2@example.com", "password123")
		_, err = testCtx.Repos.Users.GetByID(context.Background(), user2ID)
		require.NoError(t, err)
		
		tokenPair2, err := testCtx.JWTManager.GenerateTokenPair(user2ID, "ws_user2", "ws_user2@example.com", "player")
		require.NoError(t, err)
		token2 := tokenPair2.AccessToken

		// Setup headers with origin
		header := http.Header{}
		header.Set("Origin", "http://localhost:3000")

		// Connect first user
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		wsURL = fmt.Sprintf("%s?room=%s", wsURL, session.ID)
		conn1, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer conn1.Close()

		// Authenticate first user
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

		// Skip auth success message
		var authResp1 map[string]string
		err = conn1.ReadJSON(&authResp1)
		require.NoError(t, err)

		// Connect second user
		conn2, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer conn2.Close()

		// Authenticate second user
		var authReq2 map[string]string
		err = conn2.ReadJSON(&authReq2)
		require.NoError(t, err)

		authMsg2 := ws.AuthMessage{
			Type:  "auth",
			Token: token2,
			Room:  session.ID,
		}
		err = conn2.WriteJSON(authMsg2)
		require.NoError(t, err)

		// Skip auth success message
		var authResp2 map[string]string
		err = conn2.ReadJSON(&authResp2)
		require.NoError(t, err)

		// Send message from user1
		message := map[string]interface{}{
			"type":     "message",
			"roomId":   session.ID,
			"content":  "Hello from user1",
			"username": user.Username,
			"playerId": userID,
			"role":     "player",
			"data":     json.RawMessage(`{"content": "Hello from user1"}`),
		}
		err = conn1.WriteJSON(message)
		require.NoError(t, err)

		// User2 should receive the message
		var received map[string]interface{}
		conn2.SetReadDeadline(time.Now().Add(5 * time.Second))
		err = conn2.ReadJSON(&received)
		require.NoError(t, err)
		assert.Equal(t, "message", received["type"])
		assert.Equal(t, "Hello from user1", received["content"])
		assert.Equal(t, user.Username, received["username"])
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
		header.Set("Origin", "http://localhost:3000")

		// Connect to room 1
		wsURL1 := "ws" + strings.TrimPrefix(server.URL, "http")
		wsURL1 = fmt.Sprintf("%s?room=%s", wsURL1, session.ID)
		conn1, _, err := websocket.DefaultDialer.Dial(wsURL1, header)
		require.NoError(t, err)
		defer conn1.Close()

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
		wsURL2 = fmt.Sprintf("%s?room=%s", wsURL2, session2.ID)
		conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, header)
		require.NoError(t, err)
		defer conn2.Close()

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
		conn2.SetReadDeadline(time.Now().Add(1 * time.Second))
		var received map[string]interface{}
		err = conn2.ReadJSON(&received)
		assert.Error(t, err) // Should timeout
	})

	t.Run("reconnection handling", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		wsURL = fmt.Sprintf("%s?room=%s", wsURL, session.ID)

		// Setup headers with origin
		header := http.Header{}
		header.Set("Origin", "http://localhost:3000")

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
		conn1.Close()

		// Wait a bit
		time.Sleep(100 * time.Millisecond)

		// Reconnect
		conn2, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer conn2.Close()

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
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		wsURL = fmt.Sprintf("%s?room=%s", wsURL, session.ID)

		// Setup headers with origin
		header := http.Header{}
		header.Set("Origin", "http://localhost:3000")

		// Create multiple users
		numUsers := 5
		tokens := make([]string, numUsers)

		for i := 0; i < numUsers; i++ {
			username := fmt.Sprintf("concurrent_user_%d", i)
			email := fmt.Sprintf("concurrent_user_%d@example.com", i)
			uid := testCtx.CreateTestUser(username, email, "password123")
			
			tp, err := testCtx.JWTManager.GenerateTokenPair(uid, username, email, "player")
			require.NoError(t, err)
			tokens[i] = tp.AccessToken
		}

		// Connect all users concurrently
		var wg sync.WaitGroup
		connections := make([]*websocket.Conn, numUsers)

		for i := 0; i < numUsers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				// Connect
				conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
				if err != nil {
					t.Errorf("Failed to connect user %d: %v", idx, err)
					return
				}
				connections[idx] = conn

				// Authenticate
				var authReq map[string]string
				err = conn.ReadJSON(&authReq)
				if err != nil {
					t.Errorf("Failed to read auth request for user %d: %v", idx, err)
					return
				}

				authMsg := ws.AuthMessage{
					Type:  "auth",
					Token: tokens[idx],
					Room:  session.ID,
				}
				err = conn.WriteJSON(authMsg)
				if err != nil {
					t.Errorf("Failed to send auth for user %d: %v", idx, err)
					return
				}

				var authResp map[string]string
				err = conn.ReadJSON(&authResp)
				if err != nil {
					t.Errorf("Failed to read auth response for user %d: %v", idx, err)
					return
				}

				if authResp["type"] != "auth_success" {
					t.Errorf("Authentication failed for user %d", idx)
				}
			}(i)
		}

		wg.Wait()

		// Close all connections
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}
	})

	t.Run("origin validation", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Test with invalid origin (should fail in production mode)
		os.Setenv("GO_ENV", "production")
		defer os.Setenv("GO_ENV", "development")

		header := http.Header{}
		header.Set("Origin", "http://evil.com")

		_, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
		assert.Error(t, err)
		if resp != nil {
			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		}
	})
}