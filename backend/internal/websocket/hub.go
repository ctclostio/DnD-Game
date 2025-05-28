package websocket

import (
	"encoding/json"
	"log"
	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	rooms      map[string]map[*Client]bool
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
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if client.roomID != "" {
				if h.rooms[client.roomID] == nil {
					h.rooms[client.roomID] = make(map[*Client]bool)
				}
				h.rooms[client.roomID][client] = true
			}
			log.Printf("Client %s connected to room %s", client.id, client.roomID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if client.roomID != "" && h.rooms[client.roomID] != nil {
					delete(h.rooms[client.roomID], client)
				}
				close(client.send)
				log.Printf("Client %s disconnected from room %s", client.id, client.roomID)
			}

		case message := <-h.broadcast:
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
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
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.hub.broadcast <- message
	}
}

func (c *Client) WritePump() {
	defer c.conn.Close()
	
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// Broadcast sends a message to the hub's broadcast channel
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}