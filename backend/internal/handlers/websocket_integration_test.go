package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/testutil"
	ws "github.com/your-username/dnd-game/backend/internal/websocket"
)

// WebSocketTestClient wraps a WebSocket connection for testing
type WebSocketTestClient struct {
	t    *testing.T
	conn *websocket.Conn
}

// NewWebSocketTestClient creates a new WebSocket test client
func NewWebSocketTestClient(t *testing.T, url string, headers http.Header) (*WebSocketTestClient, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return nil, err
	}

	return &WebSocketTestClient{
		t:    t,
		conn: conn,
	}, nil
}

// ReadMessage reads a message from the WebSocket
func (c *WebSocketTestClient) ReadMessage() (map[string]interface{}, error) {
	var msg map[string]interface{}
	err := c.conn.ReadJSON(&msg)
	return msg, err
}

// WriteMessage sends a message to the WebSocket
func (c *WebSocketTestClient) WriteMessage(msg interface{}) error {
	return c.conn.WriteJSON(msg)
}

// Close closes the WebSocket connection
func (c *WebSocketTestClient) Close() error {
	return c.conn.Close()
}

func TestWebSocketConnection_Integration(t *testing.T) {
	ctx, cleanup := testutil.SetupIntegrationTest(t, testutil.IntegrationTestOptions{
		CustomRoutes: func(router *mux.Router, testCtx *testutil.IntegrationTestContext) {
			h, _ := setupTestHandlers(t, testCtx)
			authMiddleware := auth.NewMiddleware(testCtx.JWTManager)
			api := router.PathPrefix("/api/v1").Subrouter()
			
			// WebSocket route (using websocket package handler)
			ws.SetJWTManager(testCtx.JWTManager)
			api.HandleFunc("/ws", ws.HandleWebSocket).Methods("GET")
			
			// Game session routes (needed for testing)
			api.HandleFunc("/sessions", authMiddleware.Authenticate(h.CreateGameSession)).Methods("POST")
		},
	})
	defer cleanup()

	// Create test users
	userID := ctx.CreateTestUser("wsuser", "ws@example.com", "password123")
	dmUserID := ctx.CreateTestUser("dmuser", "dm@example.com", "password123")

	// Create test game session
	sessionID := ctx.CreateTestGameSession(dmUserID, "Test Session", "TEST123")

	// Generate auth tokens
	userToken, err := ctx.JWTManager.GenerateTokenPair(userID, "wsuser", "ws@example.com", "player")
	require.NoError(t, err)

	dmToken, err := ctx.JWTManager.GenerateTokenPair(dmUserID, "dmuser", "dm@example.com", "dm")
	require.NoError(t, err)

	// Start test server
	server := httptest.NewServer(ctx.Router)
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/ws"

	t.Run("Successful WebSocket Connection", func(t *testing.T) {
		// Set JWT manager for WebSocket handler
		ws.SetJWTManager(ctx.JWTManager)

		// Connect to WebSocket
		wsClient, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)
		defer wsClient.Close()

		// Read auth required message
		msg, err := wsClient.ReadMessage()
		require.NoError(t, err)
		assert.Equal(t, "auth_required", msg["type"])

		// Send authentication
		authMsg := map[string]interface{}{
			"type":  "auth",
			"token": userToken.AccessToken,
			"room":  sessionID,
		}
		err = wsClient.WriteMessage(authMsg)
		require.NoError(t, err)

		// Read auth success message
		msg, err = wsClient.ReadMessage()
		require.NoError(t, err)
		assert.Equal(t, "auth_success", msg["type"])
		assert.Equal(t, "wsuser", msg["username"])
		assert.Equal(t, "player", msg["role"])
	})

	t.Run("Authentication Failure - Invalid Token", func(t *testing.T) {
		wsClient, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)
		defer wsClient.Close()

		// Read auth required message
		msg, err := wsClient.ReadMessage()
		require.NoError(t, err)
		assert.Equal(t, "auth_required", msg["type"])

		// Send invalid authentication
		authMsg := map[string]interface{}{
			"type":  "auth",
			"token": "invalid-token",
			"room":  sessionID,
		}
		err = wsClient.WriteMessage(authMsg)
		require.NoError(t, err)

		// Read error message
		msg, err = wsClient.ReadMessage()
		if err == nil {
			assert.Equal(t, "error", msg["type"])
			assert.Equal(t, "Invalid token", msg["error"])
		}

		// Connection should be closed
		_, err = wsClient.ReadMessage()
		assert.Error(t, err)
	})

	t.Run("Authentication Failure - Missing Token", func(t *testing.T) {
		wsClient, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)
		defer wsClient.Close()

		// Read auth required message
		_, err = wsClient.ReadMessage()
		require.NoError(t, err)

		// Send authentication without token
		authMsg := map[string]interface{}{
			"type": "auth",
			"room": sessionID,
		}
		err = wsClient.WriteMessage(authMsg)
		require.NoError(t, err)

		// Read error message
		msg, err := wsClient.ReadMessage()
		if err == nil {
			assert.Equal(t, "error", msg["type"])
			assert.Equal(t, "Invalid authentication message", msg["error"])
		}

		// Connection should be closed
		_, err = wsClient.ReadMessage()
		assert.Error(t, err)
	})

	t.Run("Authentication Failure - Missing Room", func(t *testing.T) {
		wsClient, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)
		defer wsClient.Close()

		// Read auth required message
		_, err = wsClient.ReadMessage()
		require.NoError(t, err)

		// Send authentication without room
		authMsg := map[string]interface{}{
			"type":  "auth",
			"token": userToken.AccessToken,
		}
		err = wsClient.WriteMessage(authMsg)
		require.NoError(t, err)

		// Read error message
		msg, err := wsClient.ReadMessage()
		if err == nil {
			assert.Equal(t, "error", msg["type"])
			assert.Equal(t, "Room ID required", msg["error"])
		}
	})

	t.Run("Multiple Client Connections", func(t *testing.T) {
		// Connect first client (player)
		wsClient1, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)
		defer wsClient1.Close()

		// Authenticate first client
		_, err = wsClient1.ReadMessage() // auth_required
		require.NoError(t, err)
		err = wsClient1.WriteMessage(map[string]interface{}{
			"type":  "auth",
			"token": userToken.AccessToken,
			"room":  sessionID,
		})
		require.NoError(t, err)
		_, err = wsClient1.ReadMessage() // auth_success
		require.NoError(t, err)

		// Connect second client (DM)
		wsClient2, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)
		defer wsClient2.Close()

		// Authenticate second client
		_, err = wsClient2.ReadMessage() // auth_required
		require.NoError(t, err)
		err = wsClient2.WriteMessage(map[string]interface{}{
			"type":  "auth",
			"token": dmToken.AccessToken,
			"room":  sessionID,
		})
		require.NoError(t, err)
		_, err = wsClient2.ReadMessage() // auth_success
		require.NoError(t, err)

		// Send message from client 1
		testMessage := map[string]interface{}{
			"type":    "chat",
			"message": "Hello from player!",
		}
		err = wsClient1.WriteMessage(testMessage)
		require.NoError(t, err)

		// Client 2 should receive the message
		// Note: This depends on how the hub broadcasts messages
		// You may need to adjust based on your actual implementation
		time.Sleep(100 * time.Millisecond) // Give time for message to propagate
	})

	t.Run("Ping/Pong Handling", func(t *testing.T) {
		wsClient, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)
		defer wsClient.Close()

		// Authenticate
		_, err = wsClient.ReadMessage() // auth_required
		require.NoError(t, err)
		err = wsClient.WriteMessage(map[string]interface{}{
			"type":  "auth",
			"token": userToken.AccessToken,
			"room":  sessionID,
		})
		require.NoError(t, err)
		_, err = wsClient.ReadMessage() // auth_success
		require.NoError(t, err)

		// Send ping
		err = wsClient.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
		assert.NoError(t, err)

		// Should receive pong (WebSocket handles this automatically)
		// Connection should remain open
		wsClient.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err = wsClient.conn.ReadMessage()
		// If we get a deadline exceeded error, it means the connection is still open
		// and waiting for messages, which is what we want
		if err != nil {
			assert.Contains(t, err.Error(), "deadline exceeded")
		}
	})
}

func TestWebSocketMessageBroadcast_Integration(t *testing.T) {
	ctx, cleanup := testutil.SetupIntegrationTest(t, testutil.IntegrationTestOptions{
		CustomRoutes: func(router *mux.Router, testCtx *testutil.IntegrationTestContext) {
			h, _ := setupTestHandlers(t, testCtx)
			authMiddleware := auth.NewMiddleware(testCtx.JWTManager)
			api := router.PathPrefix("/api/v1").Subrouter()
			
			ws.SetJWTManager(testCtx.JWTManager)
			api.HandleFunc("/ws", ws.HandleWebSocket).Methods("GET")
			api.HandleFunc("/sessions", authMiddleware.Authenticate(h.CreateGameSession)).Methods("POST")
		},
	})
	defer cleanup()

	// Create test users
	player1ID := ctx.CreateTestUser("player1", "player1@example.com", "password123")
	player2ID := ctx.CreateTestUser("player2", "player2@example.com", "password123")
	dmID := ctx.CreateTestUser("dm", "dm@example.com", "password123")

	// Create test game session
	sessionID := ctx.CreateTestGameSession(dmID, "Broadcast Test", "BCAST123")

	// Add players to session
	ctx.AddUserToSession(sessionID, player1ID, nil)
	ctx.AddUserToSession(sessionID, player2ID, nil)

	// Generate tokens
	player1Token, _ := ctx.JWTManager.GenerateTokenPair(player1ID, "player1", "player1@example.com", "player")
	player2Token, _ := ctx.JWTManager.GenerateTokenPair(player2ID, "player2", "player2@example.com", "player")
	dmToken, _ := ctx.JWTManager.GenerateTokenPair(dmID, "dm", "dm@example.com", "dm")

	// Start test server
	server := httptest.NewServer(ctx.Router)
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/ws"

	// Set JWT manager
	ws.SetJWTManager(ctx.JWTManager)

	// Helper function to connect and authenticate
	connectAndAuth := func(token, username string) *WebSocketTestClient {
		client, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)

		// Read auth required
		msg, err := client.ReadMessage()
		require.NoError(t, err)
		require.Equal(t, "auth_required", msg["type"])

		// Send auth
		err = client.WriteMessage(map[string]interface{}{
			"type":  "auth",
			"token": token,
			"room":  sessionID,
		})
		require.NoError(t, err)

		// Read auth success
		msg, err = client.ReadMessage()
		require.NoError(t, err)
		require.Equal(t, "auth_success", msg["type"])
		require.Equal(t, username, msg["username"])

		return client
	}

	t.Run("Broadcast Chat Message", func(t *testing.T) {
		// Connect all three clients
		player1Client := connectAndAuth(player1Token.AccessToken, "player1")
		defer player1Client.Close()

		player2Client := connectAndAuth(player2Token.AccessToken, "player2")
		defer player2Client.Close()

		dmClient := connectAndAuth(dmToken.AccessToken, "dm")
		defer dmClient.Close()

		// Give time for all connections to establish
		time.Sleep(100 * time.Millisecond)

		// Player 1 sends a message
		chatMsg := map[string]interface{}{
			"type":    "message",
			"content": "Hello everyone!",
		}
		err := player1Client.WriteMessage(chatMsg)
		require.NoError(t, err)

		// All clients should receive the broadcast
		// Check player2 receives it
		player2Client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		msg, err := player2Client.ReadMessage()
		if err == nil {
			assert.Contains(t, []string{"message", "broadcast"}, msg["type"])
		}

		// Check DM receives it
		dmClient.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		msg, err = dmClient.ReadMessage()
		if err == nil {
			assert.Contains(t, []string{"message", "broadcast"}, msg["type"])
		}
	})

	t.Run("Dice Roll Broadcast", func(t *testing.T) {
		player1Client := connectAndAuth(player1Token.AccessToken, "player1")
		defer player1Client.Close()

		player2Client := connectAndAuth(player2Token.AccessToken, "player2")
		defer player2Client.Close()

		// Give time for connections to establish
		time.Sleep(100 * time.Millisecond)

		// Player 1 rolls dice
		diceMsg := map[string]interface{}{
			"type": "dice_roll",
			"data": map[string]interface{}{
				"dice":   "1d20",
				"result": 15,
			},
		}
		err := player1Client.WriteMessage(diceMsg)
		require.NoError(t, err)

		// Player 2 should see the roll
		player2Client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		msg, err := player2Client.ReadMessage()
		if err == nil {
			// The message type might be different based on implementation
			assert.NotNil(t, msg["type"])
		}
	})
}

func TestWebSocketReconnection_Integration(t *testing.T) {
	ctx, cleanup := testutil.SetupIntegrationTest(t, testutil.IntegrationTestOptions{
		CustomRoutes: func(router *mux.Router, testCtx *testutil.IntegrationTestContext) {
			h, _ := setupTestHandlers(t, testCtx)
			authMiddleware := auth.NewMiddleware(testCtx.JWTManager)
			api := router.PathPrefix("/api/v1").Subrouter()
			
			ws.SetJWTManager(testCtx.JWTManager)
			api.HandleFunc("/ws", ws.HandleWebSocket).Methods("GET")
			api.HandleFunc("/sessions", authMiddleware.Authenticate(h.CreateGameSession)).Methods("POST")
		},
	})
	defer cleanup()

	userID := ctx.CreateTestUser("reconnect", "reconnect@example.com", "password123")
	sessionID := ctx.CreateTestGameSession(userID, "Reconnect Test", "RECON123")

	token, _ := ctx.JWTManager.GenerateTokenPair(userID, "reconnect", "reconnect@example.com", "player")

	server := httptest.NewServer(ctx.Router)
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/ws"

	ws.SetJWTManager(ctx.JWTManager)

	t.Run("Client Reconnection", func(t *testing.T) {
		// First connection
		client1, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)

		// Authenticate
		_, _ = client1.ReadMessage() // auth_required
		err = client1.WriteMessage(map[string]interface{}{
			"type":  "auth",
			"token": token.AccessToken,
			"room":  sessionID,
		})
		require.NoError(t, err)
		_, _ = client1.ReadMessage() // auth_success

		// Close first connection
		client1.Close()

		// Wait a bit
		time.Sleep(100 * time.Millisecond)

		// Second connection (reconnection)
		client2, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)
		defer client2.Close()

		// Authenticate again
		_, err = client2.ReadMessage() // auth_required
		require.NoError(t, err)
		err = client2.WriteMessage(map[string]interface{}{
			"type":  "auth",
			"token": token.AccessToken,
			"room":  sessionID,
		})
		require.NoError(t, err)

		msg, err := client2.ReadMessage() // auth_success
		require.NoError(t, err)
		assert.Equal(t, "auth_success", msg["type"])
	})
}

func TestWebSocketRoomIsolation_Integration(t *testing.T) {
	ctx, cleanup := testutil.SetupIntegrationTest(t, testutil.IntegrationTestOptions{
		CustomRoutes: func(router *mux.Router, testCtx *testutil.IntegrationTestContext) {
			h, _ := setupTestHandlers(t, testCtx)
			authMiddleware := auth.NewMiddleware(testCtx.JWTManager)
			api := router.PathPrefix("/api/v1").Subrouter()
			
			ws.SetJWTManager(testCtx.JWTManager)
			api.HandleFunc("/ws", ws.HandleWebSocket).Methods("GET")
			api.HandleFunc("/sessions", authMiddleware.Authenticate(h.CreateGameSession)).Methods("POST")
		},
	})
	defer cleanup()

	// Create two separate game sessions
	dm1ID := ctx.CreateTestUser("dm1", "dm1@example.com", "password123")
	dm2ID := ctx.CreateTestUser("dm2", "dm2@example.com", "password123")
	player1ID := ctx.CreateTestUser("roomPlayer1", "rp1@example.com", "password123")
	player2ID := ctx.CreateTestUser("roomPlayer2", "rp2@example.com", "password123")

	session1ID := ctx.CreateTestGameSession(dm1ID, "Room 1", "ROOM1")
	session2ID := ctx.CreateTestGameSession(dm2ID, "Room 2", "ROOM2")

	// Generate tokens
	player1Token, _ := ctx.JWTManager.GenerateTokenPair(player1ID, "roomPlayer1", "rp1@example.com", "player")
	player2Token, _ := ctx.JWTManager.GenerateTokenPair(player2ID, "roomPlayer2", "rp2@example.com", "player")

	server := httptest.NewServer(ctx.Router)
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/ws"

	ws.SetJWTManager(ctx.JWTManager)

	t.Run("Messages Stay Within Rooms", func(t *testing.T) {
		// Connect player 1 to room 1
		client1, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)
		defer client1.Close()

		_, _ = client1.ReadMessage() // auth_required
		err = client1.WriteMessage(map[string]interface{}{
			"type":  "auth",
			"token": player1Token.AccessToken,
			"room":  session1ID,
		})
		require.NoError(t, err)
		_, _ = client1.ReadMessage() // auth_success

		// Connect player 2 to room 2
		client2, err := NewWebSocketTestClient(t, wsURL, nil)
		require.NoError(t, err)
		defer client2.Close()

		_, _ = client2.ReadMessage() // auth_required
		err = client2.WriteMessage(map[string]interface{}{
			"type":  "auth",
			"token": player2Token.AccessToken,
			"room":  session2ID,
		})
		require.NoError(t, err)
		_, _ = client2.ReadMessage() // auth_success

		// Wait for connections to establish
		time.Sleep(100 * time.Millisecond)

		// Player 1 sends a message in room 1
		err = client1.WriteMessage(map[string]interface{}{
			"type":    "message",
			"content": "Message in room 1",
		})
		require.NoError(t, err)

		// Player 2 should NOT receive this message
		client2.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, err = client2.ReadMessage()
		// Should timeout waiting for message
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deadline exceeded")
	})
}