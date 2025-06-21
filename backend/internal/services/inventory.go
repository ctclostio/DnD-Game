package services

import (
	"context"
	"fmt"

	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// Error message constants
const (
	errMsgItemNotFound = "item not found"
)

type InventoryService struct {
	inventoryRepo database.InventoryRepository
	characterRepo database.CharacterRepository
}

func NewInventoryService(inventoryRepo database.InventoryRepository, characterRepo database.CharacterRepository) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		characterRepo: characterRepo,
	}
}

func (s *InventoryService) AddItemToCharacter(characterID, itemID string, quantity int) error {
	character, err := s.characterRepo.GetByID(context.Background(), characterID)
	if err != nil {
		return err
	}
	if character == nil {
		return fmt.Errorf("character not found")
	}

	item, err := s.inventoryRepo.GetItem(itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return fmt.Errorf(errMsgItemNotFound)
	}

	return s.inventoryRepo.AddItemToInventory(characterID, itemID, quantity)
}

func (s *InventoryService) RemoveItemFromCharacter(characterID, itemID string, quantity int) error {
	return s.inventoryRepo.RemoveItemFromInventory(characterID, itemID, quantity)
}

func (s *InventoryService) GetCharacterInventory(characterID string) ([]*models.InventoryItem, error) {
	return s.inventoryRepo.GetCharacterInventory(characterID)
}

func (s *InventoryService) EquipItem(characterID, itemID string) error {
	inventory, err := s.inventoryRepo.GetCharacterInventory(characterID)
	if err != nil {
		return err
	}

	targetItem := s.findItemInInventory(inventory, itemID)
	if targetItem == nil {
		return fmt.Errorf("item not found in inventory")
	}

	// Handle armor equipping
	if targetItem.Item.Type == models.ItemTypeArmor {
		if err := s.unequipExistingArmor(characterID, itemID, inventory); err != nil {
			return err
		}
	}

	// Handle weapon equipping
	if targetItem.Item.Type == models.ItemTypeWeapon {
		if err := s.validateWeaponSlots(targetItem, itemID, inventory); err != nil {
			return err
		}
	}

	return s.inventoryRepo.EquipItem(characterID, itemID, true)
}

// findItemInInventory searches for an item in the inventory
func (s *InventoryService) findItemInInventory(inventory []*models.InventoryItem, itemID string) *models.InventoryItem {
	for _, inv := range inventory {
		if inv.ItemID == itemID {
			return inv
		}
	}
	return nil
}

// unequipExistingArmor unequips any currently equipped armor
func (s *InventoryService) unequipExistingArmor(characterID, itemID string, inventory []*models.InventoryItem) error {
	for _, inv := range inventory {
		if inv.Equipped && inv.Item.Type == models.ItemTypeArmor && inv.ItemID != itemID {
			if err := s.inventoryRepo.EquipItem(characterID, inv.ItemID, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateWeaponSlots checks if there are enough hand slots for the weapon
func (s *InventoryService) validateWeaponSlots(targetItem *models.InventoryItem, itemID string, inventory []*models.InventoryItem) error {
	weaponSlots := s.getWeaponSlots(targetItem)
	currentSlots := s.countEquippedWeaponSlots(inventory, itemID)

	if currentSlots+weaponSlots > 2 {
		return fmt.Errorf("not enough hands to equip this weapon")
	}
	return nil
}

// getWeaponSlots returns the number of hand slots required for a weapon
func (s *InventoryService) getWeaponSlots(item *models.InventoryItem) int {
	if item.Item.Properties["two_handed"] == true {
		return 2
	}
	return 1
}

// countEquippedWeaponSlots counts the number of hand slots used by equipped weapons
func (s *InventoryService) countEquippedWeaponSlots(inventory []*models.InventoryItem, excludeItemID string) int {
	currentSlots := 0
	for _, inv := range inventory {
		if inv.Equipped && inv.Item.Type == models.ItemTypeWeapon && inv.ItemID != excludeItemID {
			if inv.Item.Properties["two_handed"] == true {
				currentSlots += 2
			} else {
				currentSlots++
			}
		}
	}
	return currentSlots
}

func (s *InventoryService) UnequipItem(characterID, itemID string) error {
	return s.inventoryRepo.EquipItem(characterID, itemID, false)
}

func (s *InventoryService) AttuneToItem(characterID, itemID string) error {
	inventory, err := s.inventoryRepo.GetCharacterInventory(characterID)
	if err != nil {
		return err
	}

	found := false
	for _, inv := range inventory {
		if inv.ItemID == itemID {
			found = true
			if !inv.Item.RequiresAttunement {
				return fmt.Errorf("item does not require attunement")
			}
			if inv.Attuned {
				return fmt.Errorf("already attuned to this item")
			}
			break
		}
	}

	if !found {
		return fmt.Errorf("item not found in inventory")
	}

	return s.inventoryRepo.AttuneItem(characterID, itemID)
}

func (s *InventoryService) UnattuneFromItem(characterID, itemID string) error {
	return s.inventoryRepo.UnattuneItem(characterID, itemID)
}

func (s *InventoryService) GetCharacterCurrency(characterID string) (*models.Currency, error) {
	return s.inventoryRepo.GetCharacterCurrency(characterID)
}

func (s *InventoryService) UpdateCharacterCurrency(characterID string, copper, silver, electrum, gold, platinum int) error {
	currency, err := s.inventoryRepo.GetCharacterCurrency(characterID)
	if err != nil {
		return err
	}

	currency.Copper += copper
	currency.Silver += silver
	currency.Electrum += electrum
	currency.Gold += gold
	currency.Platinum += platinum

	if currency.Copper < 0 || currency.Silver < 0 || currency.Electrum < 0 ||
		currency.Gold < 0 || currency.Platinum < 0 {
		return fmt.Errorf("insufficient funds")
	}

	return s.inventoryRepo.UpdateCharacterCurrency(currency)
}

func (s *InventoryService) PurchaseItem(characterID, itemID string, quantity int) error {
	item, err := s.inventoryRepo.GetItem(itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return fmt.Errorf(errMsgItemNotFound)
	}

	totalCost := item.Value * quantity
	currency, err := s.inventoryRepo.GetCharacterCurrency(characterID)
	if err != nil {
		return err
	}

	if !currency.CanAfford(totalCost) {
		return fmt.Errorf("insufficient funds")
	}

	if !currency.Subtract(totalCost) {
		return fmt.Errorf("failed to subtract currency")
	}

	if err := s.inventoryRepo.UpdateCharacterCurrency(currency); err != nil {
		return err
	}

	return s.inventoryRepo.AddItemToInventory(characterID, itemID, quantity)
}

func (s *InventoryService) SellItem(characterID, itemID string, quantity int) error {
	item, err := s.inventoryRepo.GetItem(itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return fmt.Errorf(errMsgItemNotFound)
	}

	salePrice := (item.Value * quantity) / 2

	if err := s.inventoryRepo.RemoveItemFromInventory(characterID, itemID, quantity); err != nil {
		return err
	}

	currency, err := s.inventoryRepo.GetCharacterCurrency(characterID)
	if err != nil {
		return err
	}

	total := currency.TotalInCopper() + salePrice

	currency.Platinum = total / 1000
	total %= 1000

	currency.Gold = total / 100
	total %= 100

	currency.Electrum = total / 50
	total %= 50

	currency.Silver = total / 10
	currency.Copper = total % 10

	return s.inventoryRepo.UpdateCharacterCurrency(currency)
}

func (s *InventoryService) GetCharacterWeight(characterID string) (*models.InventoryWeight, error) {
	return s.inventoryRepo.GetCharacterWeight(characterID)
}

func (s *InventoryService) CreateItem(item *models.Item) error {
	return s.inventoryRepo.CreateItem(item)
}

func (s *InventoryService) GetItemsByType(itemType models.ItemType) ([]*models.Item, error) {
	return s.inventoryRepo.GetItemsByType(itemType)
}
