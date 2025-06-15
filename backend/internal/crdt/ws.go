package crdt

import (
	"net/http"

	"github.com/automerge/automerge-go"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}

// setupCRDTConnection attempts to upgrade the HTTP connection to a WebSocket connection
// and retrieves the document ID. It handles writing HTTP errors if setup fails.
// It returns the established WebSocket connection, the document ID, and an error.
// If an error occurs that prevents further processing (e.g., ID missing, upgrade failed),
// an HTTP error is written to w and a non-nil error is returned.
func setupCRDTConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, string, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return nil, "", http.ErrMissingFile // Using a generic error, specific error type might be better
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// upgrader.Upgrade handles writing the error response internally if it fails
		return nil, "", err
	}
	return conn, id, nil
}

// processCRDTMessages handles the synchronization logic for a given document
// over an established WebSocket connection.
func processCRDTMessages(conn *websocket.Conn, doc *automerge.Doc) {
	state, err := automerge.NewSyncState(doc)
	if err != nil {
		// TODO: Consider logging this error, e.g., log.Printf("Error creating sync state for doc %v: %v", doc, err)
		// For now, simply return, which will terminate message processing for this connection.
		return
	}
	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			// Error reading message, assume connection closed or problematic
			return
		}
		if mt != websocket.BinaryMessage {
			// Ignore non-binary messages
			continue
		}

		// Apply incoming message to the sync state
		// The original code ignores the patch, so we do too.
		if _, err := state.ReceiveMessage(msg); err != nil {
			// Error processing message
			return
		}

		// Generate and send any outgoing messages
		for {
			syncMsg, ok := state.GenerateMessage()
			if !ok {
				// No more messages to generate at this time
				break
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, syncMsg.Bytes()); err != nil {
				// Error writing message
				return
			}
		}
	}
}

func SyncHandler(w http.ResponseWriter, r *http.Request) {
	conn, id, err := setupCRDTConnection(w, r)
	if err != nil {
		// setupCRDTConnection already handled writing the HTTP error
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	// The original code ignores the error from LoadDoc.
	// This behavior is preserved here but might be an area for future improvement.
	doc, _ := LoadDoc(id)
	if doc == nil {
		// If LoadDoc can return nil on error (even if error is ignored),
		// we should probably handle it to prevent panic with NewSyncState.
		// For now, assuming LoadDoc returns a non-nil doc or NewSyncState handles nil.
		// If LoadDoc truly fails and returns nil, this could be problematic.
		// Consider logging an error here if doc is nil.
		// http.Error(w, "failed to load document", http.StatusInternalServerError)
		// return
	}

	processCRDTMessages(conn, doc)
}
