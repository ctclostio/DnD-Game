package websocket

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/middleware"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// getAllowedOrigins returns the list of allowed origins for CORS
func getAllowedOrigins() []string {
	// Default development origins
	origins := []string{
		"http://localhost:3000",
		"http://localhost:8080",
	}

	// Add production origin from environment
	if prodOrigin := os.Getenv("PRODUCTION_ORIGIN"); prodOrigin != "" {
		origins = append(origins, prodOrigin)
	}
	
	return origins
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		// In development, also allow empty origin
		if os.Getenv("GO_ENV") == constants.EnvDevelopment && origin == "" {
			return true
		}

		return middleware.ValidateOrigin(getAllowedOrigins(), origin)
	},
	// Enable compression
	EnableCompression: true,
}

var (
	hub        *Hub
	jwtManager *auth.JWTManager
)

// InitHub initializes and starts the websocket hub
func InitHub() *Hub {
	if hub == nil {
		hub = NewHub()
		go hub.Run()
	}
	return hub
}

// GetHub returns the websocket hub instance
func GetHub() *Hub {
	if hub == nil {
		return InitHub()
	}
	return hub
}

// SetJWTManager sets the JWT manager for WebSocket authentication
func SetJWTManager(manager *auth.JWTManager) {
	jwtManager = manager
}

// AuthMessage represents the authentication message
type AuthMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
	Room  string `json:"room"`
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Log the connection attempt
	logger.Info().
		Str("origin", r.Header.Get("Origin")).
		Str("remote_addr", r.RemoteAddr).
		Str("user_agent", r.Header.Get("User-Agent")).
		Msg("WebSocket connection attempt")

	// Upgrade connection first
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error().
			Err(err).
			Str("origin", r.Header.Get("Origin")).
			Str("remote_addr", r.RemoteAddr).
			Msg("WebSocket upgrade error")
		return
	}

	// Set initial timeouts
	_ = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Create temporary client for authentication
	tempClient := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	// Send authentication request
	authRequest := map[string]string{
		"type":    "auth_required",
		"message": "Please authenticate",
	}

	if err := conn.WriteJSON(authRequest); err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to send auth request")
		conn.Close()
		return
	}

	// Wait for authentication message
	var authMsg AuthMessage
	if err := conn.ReadJSON(&authMsg); err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to read auth message")
		_ = conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Authentication failed",
		})
		conn.Close()
		return
	}

	// Validate authentication message
	if authMsg.Type != constants.AuthType || authMsg.Token == "" {
		_ = conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Invalid authentication message",
		})
		conn.Close()
		return
	}

	// Validate token
	if jwtManager == nil {
		logger.Error().
			Msg("JWT manager not initialized")
		_ = conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Internal server error",
		})
		conn.Close()
		return
	}

	claims, err := jwtManager.ValidateToken(authMsg.Token, auth.AccessToken)
	if err != nil {
		logger.Warn().
			Err(err).
			Str("remote_addr", r.RemoteAddr).
			Msg("Token validation failed")
		_ = conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Invalid token",
		})
		conn.Close()
		return
	}

	// Validate room ID
	roomID := authMsg.Room
	if roomID == "" {
		// Try to get from query params as fallback
		roomID = r.URL.Query().Get("room")
		if roomID == "" {
			_ = conn.WriteJSON(map[string]string{
				"type":  "error",
				"error": "Room ID required",
			})
			conn.Close()
			return
		}
	}

	// Create authenticated client
	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     tempClient.send,
		id:       claims.UserID,
		username: claims.Username,
		roomID:   roomID,
		role:     claims.Role,
	}

	// Send authentication success
	_ = conn.WriteJSON(map[string]string{
		"type":     "auth_success",
		"message":  "Authentication successful",
		"username": client.username,
		"role":     client.role,
	})

	// Remove read deadline after successful auth
	_ = conn.SetReadDeadline(time.Time{})

	// Register client with hub
	client.hub.register <- client

	// Start client pumps
	go client.WritePump()
	go client.ReadPump()
}
