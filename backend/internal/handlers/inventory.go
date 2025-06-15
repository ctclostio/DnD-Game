package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
)

type InventoryHandler struct {
	inventoryService *services.InventoryService
}

func NewInventoryHandler(inventoryService *services.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

func (h *InventoryHandler) GetCharacterInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	inventory, err := h.inventoryService.GetCharacterInventory(characterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, inventory)
}

func (h *InventoryHandler) AddItemToInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	req, err := decodeItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.inventoryService.AddItemToCharacter(characterID, req.ItemID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccessResponse(w, "success")
}

func (h *InventoryHandler) RemoveItemFromInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	req, err := decodeItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.inventoryService.RemoveItemFromCharacter(characterID, req.ItemID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccessResponse(w, "success")
}

func (h *InventoryHandler) EquipItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]
	itemID := vars["itemId"]

	if err := h.inventoryService.EquipItem(characterID, itemID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendSuccessResponse(w, "equipped")
}

func (h *InventoryHandler) UnequipItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]
	itemID := vars["itemId"]

	if err := h.inventoryService.UnequipItem(characterID, itemID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendSuccessResponse(w, "unequipped")
}

func (h *InventoryHandler) AttuneItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]
	itemID := vars["itemId"]

	if err := h.inventoryService.AttuneToItem(characterID, itemID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendSuccessResponse(w, "attuned")
}

func (h *InventoryHandler) UnattuneItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]
	itemID := vars["itemId"]

	if err := h.inventoryService.UnattuneFromItem(characterID, itemID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendSuccessResponse(w, "unattuned")
}

func (h *InventoryHandler) GetCharacterCurrency(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	currency, err := h.inventoryService.GetCharacterCurrency(characterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, currency)
}

func (h *InventoryHandler) UpdateCharacterCurrency(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	var req struct {
		Copper   int `json:"copper"`
		Silver   int `json:"silver"`
		Electrum int `json:"electrum"`
		Gold     int `json:"gold"`
		Platinum int `json:"platinum"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.inventoryService.UpdateCharacterCurrency(characterID, req.Copper, req.Silver,
		req.Electrum, req.Gold, req.Platinum); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	currency, err := h.inventoryService.GetCharacterCurrency(characterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, currency)
}

func (h *InventoryHandler) PurchaseItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	req, err := decodeItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.inventoryService.PurchaseItem(characterID, req.ItemID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendSuccessResponse(w, "purchased")
}

func (h *InventoryHandler) SellItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	req, err := decodeItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.inventoryService.SellItem(characterID, req.ItemID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendSuccessResponse(w, "sold")
}

func (h *InventoryHandler) GetCharacterWeight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	weight, err := h.inventoryService.GetCharacterWeight(characterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, weight)
}

func (h *InventoryHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var item models.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.inventoryService.CreateItem(&item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, item)
}

func (h *InventoryHandler) GetItemsByType(w http.ResponseWriter, r *http.Request) {
	itemType := r.URL.Query().Get("type")
	if itemType == "" {
		http.Error(w, "type parameter required", http.StatusBadRequest)
		return
	}

	items, err := h.inventoryService.GetItemsByType(models.ItemType(itemType))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, items)
}
