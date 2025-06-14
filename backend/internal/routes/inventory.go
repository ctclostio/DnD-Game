package routes

import (
	"github.com/ctclostio/DnD-Game/backend/internal/handlers"
	"github.com/gorilla/mux"
)

// RegisterInventoryRoutes registers all inventory-related routes.
func RegisterInventoryRoutes(api *mux.Router, cfg *Config) {
	// Check if inventory handler exists.
	if cfg.InventoryHandler == nil {
		return
	}

	inventoryHandler, ok := cfg.InventoryHandler.(*handlers.InventoryHandler)
	if !ok {
		return
	}

	auth := cfg.AuthMiddleware.Authenticate

	// Inventory management.
	api.HandleFunc("/characters/{characterId}/inventory",
		auth(inventoryHandler.GetCharacterInventory)).Methods("GET")
	api.HandleFunc("/characters/{characterId}/inventory",
		auth(inventoryHandler.AddItemToInventory)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/remove",
		auth(inventoryHandler.RemoveItemFromInventory)).Methods("POST")

	// Equipment management.
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/equip",
		auth(inventoryHandler.EquipItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/unequip",
		auth(inventoryHandler.UnequipItem)).Methods("POST")

	// Attunement.
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/attune",
		auth(inventoryHandler.AttuneItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/unattune",
		auth(inventoryHandler.UnattuneItem)).Methods("POST")

	// Currency management.
	api.HandleFunc("/characters/{characterId}/currency",
		auth(inventoryHandler.GetCharacterCurrency)).Methods("GET")
	api.HandleFunc("/characters/{characterId}/currency",
		auth(inventoryHandler.UpdateCharacterCurrency)).Methods("PUT")

	// Trading.
	api.HandleFunc("/characters/{characterId}/inventory/purchase",
		auth(inventoryHandler.PurchaseItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/sell",
		auth(inventoryHandler.SellItem)).Methods("POST")
}
