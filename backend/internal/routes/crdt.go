package routes

import (
	"github.com/ctclostio/DnD-Game/backend/internal/crdt"
	"github.com/gorilla/mux"
)

func RegisterCRDTRoutes(router *mux.Router) {
	router.HandleFunc("/ws/v1/characters/{id}/sync", crdt.SyncHandler)
}
