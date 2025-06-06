package websocket

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/middleware"
	"github.com/your-username/dnd-game/backend/pkg/logger"
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

// AuthMessage represents the authentication message
type AuthMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
	Room  string `json:"room"`
}

// Message represents a WebSocket message
type Message struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	Room    string
	Message Message
	Sender  string
}

// HandlerV2 is the enhanced WebSocket handler with structured logging
type HandlerV2 struct {
	hub        *Hub
	jwtManager *auth.JWTManager
	log        *logger.LoggerV2
	upgrader   websocket.Upgrader
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
		hub:        hub,
		jwtManager: jwtManager,
		log:        log,
		allowedOrigins: allowedOrigins,
	}
	
	// Configure upgrader
	h.upgrader = websocket.Upgrader{
		CheckOrigin: h.checkOrigin,
		EnableCompression: true,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		HandshakeTimeout: 10 * time.Second,
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
	clientIP := getClientIP(r)
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
	clientID := generateClientID()
	
	// Log successful upgrade
	log.Info().
		Str("client_id", clientID).
		Str("client_ip", clientIP).
		Msg("WebSocket connection established")
	
	// Configure connection
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		log.Debug().
			Str("client_id", clientID).
			Msg("Received pong, extending read deadline")
		return nil
	})
	
	// Create temporary client for authentication
	tempClient := &Client{
		ID:   clientID,
		conn: conn,
		send: make(chan []byte, 256),
		log:  log.WithField("client_id", clientID),
	}
	
	// Send authentication request
	authRequest := map[string]string{
		"type":    "auth_required",
		"message": "Please authenticate",
	}
	
	if err := tempClient.sendJSON(authRequest); err != nil {
		log.Error().
			Err(err).
			Str("client_id", clientID).
			Msg("Failed to send auth request")
		conn.Close()
		return
	}
	
	// Wait for authentication
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	
	var authMsg AuthMessage
	if err := conn.ReadJSON(&authMsg); err != nil {
		log.Warn().
			Err(err).
			Str("client_id", clientID).
			Msg("Failed to read auth message")
		tempClient.sendError("Authentication failed")
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
		tempClient.sendError("Invalid authentication message")
		conn.Close()
		return
	}
	
	// Verify JWT token
	userID, err := h.jwtManager.VerifyToken(authMsg.Token)
	if err != nil {
		log.Warn().
			Err(err).
			Str("client_id", clientID).
			Msg("Token validation failed")
		tempClient.sendError("Invalid token")
		conn.Close()
		return
	}
	
	// Create authenticated client
	client := &Client{
		ID:     clientID,
		UserID: userID,
		conn:   conn,
		send:   make(chan []byte, 256),
		hub:    h.hub,
		log:    log.WithFields(map[string]interface{}{
			"client_id": clientID,
			"user_id":   userID,
		}),
	}
	
	// Join room if specified
	if authMsg.Room != "" {
		client.room = authMsg.Room
		client.log = client.log.WithField("room", authMsg.Room)
		
		client.log.Info().
			Str("room", authMsg.Room).
			Msg("Client joining room")
	}
	
	// Register client
	client.hub.register <- client
	
	// Send success message
	successMsg := map[string]interface{}{
		"type":    "auth_success",
		"message": "Authentication successful",
		"user_id": userID,
		"client_id": clientID,
	}
	
	if err := client.sendJSON(successMsg); err != nil {
		client.log.Error().
			Err(err).
			Msg("Failed to send auth success")
	}
	
	// Log successful authentication
	client.log.Info().
		Str("room", client.room).
		Msg("WebSocket client authenticated and registered")
	
	// Reset read deadline for normal operation
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	
	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// Client represents a WebSocket client with logging
type Client struct {
	ID     string
	UserID string
	conn   *websocket.Conn
	send   chan []byte
	hub    *Hub
	room   string
	log    *logger.LoggerV2
}

// sendJSON sends a JSON message to the client
func (c *Client) sendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		c.log.Error().
			Err(err).
			Msg("Failed to marshal JSON")
		return err
	}
	
	select {
	case c.send <- data:
		return nil
	default:
		c.log.Warn().Msg("Client send buffer full")
		return fmt.Errorf("send buffer full")
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(message string) {
	errorMsg := map[string]string{
		"type":    "error",
		"message": message,
	}
	c.sendJSON(errorMsg)
}

// readPump handles incoming messages from the client
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		c.log.Info().Msg("Client disconnected")
	}()
	
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	
	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Error().
					Err(err).
					Msg("Unexpected WebSocket close")
			} else {
				c.log.Debug().
					Err(err).
					Msg("WebSocket read error")
			}
			break
		}
		
		// Log received message
		c.log.Debug().
			Str("message_type", message.Type).
			Interface("data", message.Data).
			Msg("Received WebSocket message")
		
		// Handle message based on type
		switch message.Type {
		case "ping":
			c.sendJSON(map[string]string{"type": "pong"})
			
		case "join_room":
			if roomID, ok := message.Data["room_id"].(string); ok {
				c.room = roomID
				c.log = c.log.WithField("room", roomID)
				c.log.Info().Msg("Client joined room")
			}
			
		case "leave_room":
			oldRoom := c.room
			c.room = ""
			c.log.Info().
				Str("old_room", oldRoom).
				Msg("Client left room")
			
		case "broadcast":
			// Broadcast to room
			if c.room != "" {
				c.hub.broadcast <- &BroadcastMessage{
					Room:    c.room,
					Message: message,
					Sender:  c.ID,
				}
				c.log.Debug().
					Str("room", c.room).
					Msg("Broadcasting message to room")
			}
			
		default:
			c.log.Warn().
				Str("message_type", message.Type).
				Msg("Unknown message type")
		}
	}
}

// writePump handles outgoing messages to the client
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.log.Debug().Msg("Send channel closed")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.log.Error().
					Err(err).
					Msg("Failed to get writer")
				return
			}
			w.Write(message)
			
			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}
			
			if err := w.Close(); err != nil {
				c.log.Error().
					Err(err).
					Msg("Failed to close writer")
				return
			}
			
			c.log.Debug().
				Int("message_count", n+1).
				Msg("Sent messages to client")
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.log.Debug().
					Err(err).
					Msg("Failed to send ping")
				return
			}
		}
	}
}

// Helper functions

func generateClientID() string {
	return fmt.Sprintf("ws_%s", uuid.New().String())
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}