package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventory)
}

func (h *InventoryHandler) AddItemToInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	var req struct {
		ItemID   string `json:"item_id"`
		Quantity int    `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	if err := h.inventoryService.AddItemToCharacter(characterID, req.ItemID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *InventoryHandler) RemoveItemFromInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	var req struct {
		ItemID   string `json:"item_id"`
		Quantity int    `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	if err := h.inventoryService.RemoveItemFromCharacter(characterID, req.ItemID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *InventoryHandler) EquipItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]
	itemID := vars["itemId"]

	if err := h.inventoryService.EquipItem(characterID, itemID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "equipped"})
}

func (h *InventoryHandler) UnequipItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]
	itemID := vars["itemId"]

	if err := h.inventoryService.UnequipItem(characterID, itemID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "unequipped"})
}

func (h *InventoryHandler) AttuneItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]
	itemID := vars["itemId"]

	if err := h.inventoryService.AttuneToItem(characterID, itemID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "attuned"})
}

func (h *InventoryHandler) UnattuneItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]
	itemID := vars["itemId"]

	if err := h.inventoryService.UnattuneFromItem(characterID, itemID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "unattuned"})
}

func (h *InventoryHandler) GetCharacterCurrency(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	currency, err := h.inventoryService.GetCharacterCurrency(characterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currency)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currency)
}

func (h *InventoryHandler) PurchaseItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	var req struct {
		ItemID   string `json:"item_id"`
		Quantity int    `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	if err := h.inventoryService.PurchaseItem(characterID, req.ItemID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "purchased"})
}

func (h *InventoryHandler) SellItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	var req struct {
		ItemID   string `json:"item_id"`
		Quantity int    `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	if err := h.inventoryService.SellItem(characterID, req.ItemID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "sold"})
}

func (h *InventoryHandler) GetCharacterWeight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	weight, err := h.inventoryService.GetCharacterWeight(characterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weight)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
