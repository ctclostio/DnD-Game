package websocket

import (
	"context"
	"encoding/json"

	"github.com/gorilla/websocket"

	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	rooms      map[string]map[*Client]bool
	shutdown   chan struct{}
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	id       string
	username string
	roomID   string
	role     string // "player" or "dm"
}

type Message struct {
	Type     string          `json:"type"`
	RoomID   string          `json:"roomId"`
	PlayerID string          `json:"playerId"`
	Username string          `json:"username"`
	Role     string          `json:"role"`
	Data     json.RawMessage `json:"data"`
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		shutdown:   make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.shutdown:
			for client := range h.clients {
				close(client.send)
				_ = client.conn.Close()
			}
			return
		case client := <-h.register:
			h.clients[client] = true
			if client.roomID != "" {
				if h.rooms[client.roomID] == nil {
					h.rooms[client.roomID] = make(map[*Client]bool)
				}
				h.rooms[client.roomID][client] = true
			}
			logger.Info().
				Str("client_id", client.id).
				Str("username", client.username).
				Str("room_id", client.roomID).
				Str("role", client.role).
				Msg("Client connected to room")

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if client.roomID != "" && h.rooms[client.roomID] != nil {
					delete(h.rooms[client.roomID], client)
				}
				close(client.send)
				logger.Info().
					Str("client_id", client.id).
					Str("username", client.username).
					Str("room_id", client.roomID).
					Msg("Client disconnected from room")
			}

		case message := <-h.broadcast:
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				logger.Error().
					Err(err).
					Msg("Error unmarshaling message")
				continue
			}

			// Broadcast to room
			if msg.RoomID != "" && h.rooms[msg.RoomID] != nil {
				for client := range h.rooms[msg.RoomID] {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
						delete(h.rooms[msg.RoomID], client)
					}
				}
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error().
					Err(err).
					Str("client_id", c.id).
					Str("room_id", c.roomID).
					Msg("WebSocket read error")
			}
			break
		}
		c.hub.broadcast <- message
	}
}

func (c *Client) WritePump() {
	defer func() { _ = c.conn.Close() }()

	for message := range c.send {
		_ = c.conn.WriteMessage(websocket.TextMessage, message)
	}
	_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// Broadcast sends a message to the hub's broadcast channel
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// Shutdown gracefully stops the hub and closes all connections
func (h *Hub) Shutdown(_ context.Context) error {
	close(h.shutdown)
	return nil
}
