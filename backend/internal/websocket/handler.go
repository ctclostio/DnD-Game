package websocket

import (
	"log"
	"net/http"
	"os"
	"time"
	
	"github.com/gorilla/websocket"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/middleware"
)

var allowedOrigins []string

func init() {
	// Default development origins
	allowedOrigins = []string{
		"http://localhost:3000",
		"http://localhost:8080",
	}
	
	// Add production origin from environment
	if prodOrigin := os.Getenv("PRODUCTION_ORIGIN"); prodOrigin != "" {
		allowedOrigins = append(allowedOrigins, prodOrigin)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		
		// In development, also allow empty origin
		if os.Getenv("GO_ENV") == "development" && origin == "" {
			return true
		}
		
		return middleware.ValidateOrigin(allowedOrigins, origin)
	},
	// Enable compression
	EnableCompression: true,
}

var hub = NewHub()

var jwtManager *auth.JWTManager

func init() {
	go hub.Run()
}

// GetHub returns the websocket hub instance
func GetHub() *Hub {
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
	log.Printf("WebSocket connection attempt from origin: %s", r.Header.Get("Origin"))

	// Upgrade connection first
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Set initial timeouts
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
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
		log.Printf("Failed to send auth request: %v", err)
		conn.Close()
		return
	}

	// Wait for authentication message
	var authMsg AuthMessage
	if err := conn.ReadJSON(&authMsg); err != nil {
		log.Printf("Failed to read auth message: %v", err)
		conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Authentication failed",
		})
		conn.Close()
		return
	}

	// Validate authentication message
	if authMsg.Type != "auth" || authMsg.Token == "" {
		conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Invalid authentication message",
		})
		conn.Close()
		return
	}

	// Validate token
	if jwtManager == nil {
		log.Println("JWT manager not initialized")
		conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Internal server error",
		})
		conn.Close()
		return
	}

	claims, err := jwtManager.ValidateToken(authMsg.Token, auth.AccessToken)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		conn.WriteJSON(map[string]string{
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
			conn.WriteJSON(map[string]string{
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
	conn.WriteJSON(map[string]string{
		"type":     "auth_success",
		"message":  "Authentication successful",
		"username": client.username,
		"role":     client.role,
	})

	// Remove read deadline after successful auth
	conn.SetReadDeadline(time.Time{})

	// Register client with hub
	client.hub.register <- client

	// Start client pumps
	go client.WritePump()
	go client.ReadPump()
}