package websocket

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost during development
		return true
	},
}

var hub = NewHub()

func init() {
	go hub.Run()
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Get room ID and player ID from query params
	roomID := r.URL.Query().Get("room")
	playerID := r.URL.Query().Get("player")

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		id:     playerID,
		roomID: roomID,
	}

	client.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}