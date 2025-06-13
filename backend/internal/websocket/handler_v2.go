package websocket

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/middleware"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// AuthMessageV2 represents the authentication message
type AuthMessageV2 struct {
	Type  string `json:"type"`
	Token string `json:"token"`
	Room  string `json:"room"`
}

// MessageV2 represents a WebSocket message
type MessageV2 struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	Room    string
	Message MessageV2
	Sender  string
}

// HandlerV2 is the enhanced WebSocket handler with structured logging
type HandlerV2 struct {
	hub            *Hub
	jwtManager     *auth.JWTManager
	log            *logger.LoggerV2
	upgrader       websocket.Upgrader
	allowedOrigins []string
}

// NewHandlerV2 creates a new WebSocket handler with logging
func NewHandlerV2(hub *Hub, jwtManager *auth.JWTManager, log *logger.LoggerV2) *HandlerV2 {
	// Configure allowed origins
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:8080",
	}

	// Add production origin from environment
	if prodOrigin := os.Getenv("PRODUCTION_ORIGIN"); prodOrigin != "" {
		allowedOrigins = append(allowedOrigins, prodOrigin)
	}

	h := &HandlerV2{
		hub:            hub,
		jwtManager:     jwtManager,
		log:            log,
		allowedOrigins: allowedOrigins,
	}

	// Configure upgrader
	h.upgrader = websocket.Upgrader{
		CheckOrigin:       h.checkOrigin,
		EnableCompression: true,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		HandshakeTimeout:  10 * time.Second,
	}

	return h
}

// checkOrigin validates the origin of WebSocket connections
func (h *HandlerV2) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	// Log origin check
	h.log.WithContext(r.Context()).Debug().
		Str("origin", origin).
		Strs("allowed_origins", h.allowedOrigins).
		Msg("Checking WebSocket origin")

	// In development, allow empty origin
	if os.Getenv("GO_ENV") == "development" && origin == "" {
		h.log.Debug().Msg("Allowing empty origin in development")
		return true
	}

	// Check against allowed origins
	allowed := middleware.ValidateOrigin(h.allowedOrigins, origin)

	if !allowed {
		h.log.Warn().
			Str("origin", origin).
			Msg("WebSocket connection rejected - invalid origin")
	}

	return allowed
}

// HandleWebSocket handles WebSocket connections with structured logging
func (h *HandlerV2) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get request context with IDs
	ctx := r.Context()
	log := h.log.WithContext(ctx)

	// Extract connection metadata
	clientIP := r.RemoteAddr
	userAgent := r.UserAgent()

	// Log connection attempt
	log.Info().
		Str("client_ip", clientIP).
		Str("user_agent", userAgent).
		Str("origin", r.Header.Get("Origin")).
		Msg("WebSocket connection attempt")

	// Upgrade connection
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().
			Err(err).
			Str("client_ip", clientIP).
			Msg("WebSocket upgrade failed")
		return
	}

	// Generate client ID
	clientID := "ws-" + time.Now().Format("20060102150405")

	// Log successful upgrade
	log.Info().
		Str("client_id", clientID).
		Str("client_ip", clientIP).
		Msg("WebSocket connection established")

	// Configure connection
	_ = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		log.Debug().
			Str("client_id", clientID).
			Msg("Received pong, extending read deadline")
		return nil
	})

	// Create temporary client for authentication
	// Using a minimal client for authentication phase
	tempConn := conn

	// Send authentication request
	authRequest := map[string]string{
		"type":    "auth_required",
		"message": "Please authenticate",
	}

	authData, _ := json.Marshal(authRequest)
	if err := tempConn.WriteMessage(websocket.TextMessage, authData); err != nil {
		log.Error().
			Err(err).
			Str("client_id", clientID).
			Msg("Failed to send auth request")
		conn.Close()
		return
	}

	// Wait for authentication
	_ = conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	var authMsg AuthMessageV2
	if err := conn.ReadJSON(&authMsg); err != nil {
		log.Warn().
			Err(err).
			Str("client_id", clientID).
			Msg("Failed to read auth message")
		errorMsg := map[string]string{"type": "error", "message": "Authentication failed"}
		errorData, _ := json.Marshal(errorMsg)
		_ = tempConn.WriteMessage(websocket.TextMessage, errorData)
		conn.Close()
		return
	}

	// Validate authentication
	if authMsg.Type != "auth" || authMsg.Token == "" {
		log.Warn().
			Str("client_id", clientID).
			Str("auth_type", authMsg.Type).
			Bool("has_token", authMsg.Token != "").
			Msg("Invalid auth message")
		errorMsg := map[string]string{"type": "error", "message": "Invalid authentication message"}
		errorData, _ := json.Marshal(errorMsg)
		_ = tempConn.WriteMessage(websocket.TextMessage, errorData)
		conn.Close()
		return
	}

	// Verify JWT token
	claims, err := h.jwtManager.ValidateToken(authMsg.Token, auth.AccessToken)
	if err != nil {
		log.Warn().
			Err(err).
			Str("client_id", clientID).
			Msg("Token validation failed")
		errorMsg := map[string]string{"type": "error", "message": "Invalid token"}
		errorData, _ := json.Marshal(errorMsg)
		_ = tempConn.WriteMessage(websocket.TextMessage, errorData)
		conn.Close()
		return
	}
	userID := claims.UserID

	// Create authenticated client
	client := &Client{
		id:       clientID,
		username: claims.Username,
		conn:     conn,
		send:     make(chan []byte, 256),
		hub:      h.hub,
		role:     claims.Role,
	}

	// Join room if specified
	if authMsg.Room != "" {
		client.roomID = authMsg.Room
		// Logger already has room context from creation

		log.Info().
			Str("room", authMsg.Room).
			Str("client_id", clientID).
			Msg("Client joining room")
	}

	// Register client
	client.hub.register <- client

	// Send success message
	successMsg := map[string]interface{}{
		"type":      "auth_success",
		"message":   "Authentication successful",
		"user_id":   userID,
		"client_id": clientID,
	}

	successData, _ := json.Marshal(successMsg)
	if err := conn.WriteMessage(websocket.TextMessage, successData); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to send auth success")
	}

	// Log successful authentication
	log.Info().
		Str("client_id", clientID).
		Str("user_id", userID).
		Str("room", client.roomID).
		Msg("WebSocket client authenticated and registered")

	// Reset read deadline for normal operation
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}

// Handler V2 uses the standard Client type from hub.go
// All WebSocket communication is handled by the standard Client ReadPump and WritePump methods
