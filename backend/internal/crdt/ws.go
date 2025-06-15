package crdt

import (
	"net/http"

	"github.com/automerge/automerge-go"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}

func SyncHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	doc, _ := LoadDoc(id)
	state := automerge.NewSyncState(doc)
	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if mt != websocket.BinaryMessage {
			continue
		}
		if _, err := state.ReceiveMessage(msg); err != nil {
			return
		}
		for {
			m, ok := state.GenerateMessage()
			if !ok {
				break
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, m.Bytes()); err != nil {
				return
			}
		}
	}
}
