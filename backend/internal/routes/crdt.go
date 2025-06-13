package routes

import (
	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/crdt"
)

func RegisterCRDTRoutes(router *mux.Router) {
	router.HandleFunc("/ws/v1/characters/{id}/sync", crdt.SyncHandler)
}
