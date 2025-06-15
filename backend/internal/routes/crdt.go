package routes

import (
	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/crdt"
)

func RegisterCRDTRoutes(router *mux.Router) {
	router.HandleFunc("/ws/v1/characters/{id}/sync", crdt.SyncHandler)
}
