package websocket

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/your-username/dnd-game/backend/internal/auth"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost during development
		return true
	},
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

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter or Authorization header
	token := r.URL.Query().Get("token")
	if token == "" {
		// Try to get from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			var err error
			token, err = auth.ExtractTokenFromHeader(authHeader)
			if err != nil {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}
		}
	}

	if token == "" {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Validate token
	if jwtManager == nil {
		log.Println("JWT manager not initialized")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	claims, err := jwtManager.ValidateToken(token, auth.AccessToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Get room ID from query params
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		conn.Close()
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		id:       claims.UserID,
		username: claims.Username,
		roomID:   roomID,
		role:     claims.Role,
	}

	client.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}