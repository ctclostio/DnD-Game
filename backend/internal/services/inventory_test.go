package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/internal/services/mocks"
)

// runInventoryServiceTest is a helper to reduce duplication in table-driven tests
func runInventoryServiceTest(t *testing.T, tests []struct {
	name          string
	characterID   string
	itemID        string
	setupMock     func(*mocks.MockInventoryRepository)
	expectedError string
}, serviceAction func(*services.InventoryService, string, string) error) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewInventoryService(mockRepo, nil)
			err := serviceAction(service, tt.characterID, tt.itemID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// runInventoryServiceTestWithQuantity is a helper for tests that include quantity parameter
func runInventoryServiceTestWithQuantity(t *testing.T, tests []struct {
	name          string
	characterID   string
	itemID        string
	quantity      int
	setupMock     func(*mocks.MockInventoryRepository)
	expectedError string
}, serviceAction func(*services.InventoryService, string, string, int) error) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewInventoryService(mockRepo, nil)
			err := serviceAction(service, tt.characterID, tt.itemID, tt.quantity)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInventoryService_AddItemToCharacter(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		characterID   string
		itemID        string
		quantity      int
		setupMock     func(*mocks.MockInventoryRepository, *mocks.MockCharacterRepository)
		expectedError string
	}{
		{
			name:        "successful add item",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    2,
			setupMock: func(invRepo *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				// Character exists
				char := mocks.CreateTestCharacter("char-123", "user-123", "Test Character", "Human", "Fighter")
				charRepo.On("GetByID", ctx, "char-123").Return(char, nil)

				// Item exists
				item := mocks.CreateTestItem("item-456", "Healing Potion", models.ItemTypeConsumable, 50, 0.5)
				invRepo.On("GetItem", "item-456").Return(item, nil)

				// Add to inventory
				invRepo.On("AddItemToInventory", "char-123", "item-456", 2).Return(nil)
			},
		},
		{
			name:        "character not found",
			characterID: "nonexistent",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(_ *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				charRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:        "character nil response",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(_ *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				charRepo.On("GetByID", ctx, "char-123").Return(nil, nil)
			},
			expectedError: "character not found",
		},
		{
			name:        "item not found",
			characterID: "char-123",
			itemID:      "nonexistent",
			quantity:    1,
			setupMock: func(invRepo *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				char := mocks.CreateTestCharacter("char-123", "user-123", "Test Character", "Human", "Fighter")
				charRepo.On("GetByID", ctx, "char-123").Return(char, nil)
				invRepo.On("GetItem", "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:        "item nil response",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(invRepo *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				char := mocks.CreateTestCharacter("char-123", "user-123", "Test Character", "Human", "Fighter")
				charRepo.On("GetByID", ctx, "char-123").Return(char, nil)
				invRepo.On("GetItem", "item-456").Return(nil, nil)
			},
			expectedError: "item not found",
		},
		{
			name:        "add to inventory error",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(invRepo *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				char := mocks.CreateTestCharacter("char-123", "user-123", "Test Character", "Human", "Fighter")
				charRepo.On("GetByID", ctx, "char-123").Return(char, nil)

				item := mocks.CreateTestItem("item-456", "Healing Potion", models.ItemTypeConsumable, 50, 0.5)
				invRepo.On("GetItem", "item-456").Return(item, nil)

				invRepo.On("AddItemToInventory", "char-123", "item-456", 1).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name:        "zero quantity",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    0,
			setupMock: func(invRepo *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				char := mocks.CreateTestCharacter("char-123", "user-123", "Test Character", "Human", "Fighter")
				charRepo.On("GetByID", ctx, "char-123").Return(char, nil)

				item := mocks.CreateTestItem("item-456", "Healing Potion", models.ItemTypeConsumable, 50, 0.5)
				invRepo.On("GetItem", "item-456").Return(item, nil)

				invRepo.On("AddItemToInventory", "char-123", "item-456", 0).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvRepo := new(mocks.MockInventoryRepository)
			mockCharRepo := new(mocks.MockCharacterRepository)

			if tt.setupMock != nil {
				tt.setupMock(mockInvRepo, mockCharRepo)
			}

			service := services.NewInventoryService(mockInvRepo, mockCharRepo)
			err := service.AddItemToCharacter(tt.characterID, tt.itemID, tt.quantity)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockInvRepo.AssertExpectations(t)
			mockCharRepo.AssertExpectations(t)
		})
	}
}

func TestInventoryService_EquipItem(t *testing.T) {
	tests := []struct {
		name          string
		characterID   string
		itemID        string
		setupMock     func(*mocks.MockInventoryRepository)
		expectedError string
	}{
		{
			name:        "equip weapon successfully",
			characterID: "char-123",
			itemID:      "sword-456",
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with one-handed sword
				sword := mocks.CreateTestItem("sword-456", "Longsword", models.ItemTypeWeapon, 15, 3.0)
				sword.Properties["two_handed"] = false

				invItem := mocks.CreateTestInventoryItem("char-123", "sword-456", 1, false, false, sword)
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{invItem}, nil)

				// Equip the sword
				m.On("EquipItem", "char-123", "sword-456", true).Return(nil)
			},
		},
		{
			name:        "equip two-handed weapon",
			characterID: "char-123",
			itemID:      "greatsword-789",
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with two-handed sword
				greatsword := mocks.CreateTestItem("greatsword-789", "Greatsword", models.ItemTypeWeapon, 50, 6.0)
				greatsword.Properties["two_handed"] = true

				invItem := mocks.CreateTestInventoryItem("char-123", "greatsword-789", 1, false, false, greatsword)
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{invItem}, nil)

				// Equip the greatsword
				m.On("EquipItem", "char-123", "greatsword-789", true).Return(nil)
			},
		},
		{
			name:        "equip armor - unequip existing",
			characterID: "char-123",
			itemID:      "plate-mail",
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with existing armor equipped
				oldArmor := mocks.CreateTestItem("leather-armor", "Leather Armor", models.ItemTypeArmor, 10, 10.0)
				newArmor := mocks.CreateTestItem("plate-mail", "Plate Mail", models.ItemTypeArmor, 1500, 65.0)

				invItems := []*models.InventoryItem{
					mocks.CreateTestInventoryItem("char-123", "leather-armor", 1, true, false, oldArmor),
					mocks.CreateTestInventoryItem("char-123", "plate-mail", 1, false, false, newArmor),
				}
				m.On("GetCharacterInventory", "char-123").Return(invItems, nil)

				// Unequip old armor first
				m.On("EquipItem", "char-123", "leather-armor", false).Return(nil)
				// Equip new armor
				m.On("EquipItem", "char-123", "plate-mail", true).Return(nil)
			},
		},
		{
			name:        "not enough hands for two-handed weapon",
			characterID: "char-123",
			itemID:      "greatsword-789",
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Already have two one-handed weapons equipped
				sword1 := mocks.CreateTestItem("sword-1", "Short Sword", models.ItemTypeWeapon, 10, 2.0)
				sword1.Properties["two_handed"] = false
				sword2 := mocks.CreateTestItem("sword-2", "Dagger", models.ItemTypeWeapon, 2, 1.0)
				sword2.Properties["two_handed"] = false
				greatsword := mocks.CreateTestItem("greatsword-789", "Greatsword", models.ItemTypeWeapon, 50, 6.0)
				greatsword.Properties["two_handed"] = true

				invItems := []*models.InventoryItem{
					mocks.CreateTestInventoryItem("char-123", "sword-1", 1, true, false, sword1),
					mocks.CreateTestInventoryItem("char-123", "sword-2", 1, true, false, sword2),
					mocks.CreateTestInventoryItem("char-123", "greatsword-789", 1, false, false, greatsword),
				}
				m.On("GetCharacterInventory", "char-123").Return(invItems, nil)
			},
			expectedError: "not enough hands to equip this weapon",
		},
		{
			name:        "item not in inventory",
			characterID: "char-123",
			itemID:      "nonexistent",
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{}, nil)
			},
			expectedError: "item not found in inventory",
		},
		{
			name:        "inventory retrieval error",
			characterID: "char-123",
			itemID:      "item-456",
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetCharacterInventory", "char-123").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	runInventoryServiceTest(t, tests, func(service *services.InventoryService, characterID, itemID string) error {
		return service.EquipItem(characterID, itemID)
	})
}

func TestInventoryService_AttuneToItem(t *testing.T) {
	tests := []struct {
		name          string
		characterID   string
		itemID        string
		setupMock     func(*mocks.MockInventoryRepository)
		expectedError string
	}{
		{
			name:        "successful attunement",
			characterID: "char-123",
			itemID:      "ring-456",
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with attuneable item
				ring := mocks.CreateTestItem("ring-456", "Ring of Protection", models.ItemTypeMagic, 500, 0.1)
				ring.RequiresAttunement = true

				invItem := mocks.CreateTestInventoryItem("char-123", "ring-456", 1, true, false, ring)
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{invItem}, nil)

				// Attune to item
				m.On("AttuneItem", "char-123", "ring-456").Return(nil)
			},
		},
		{
			name:        "item does not require attunement",
			characterID: "char-123",
			itemID:      "sword-456",
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with non-attuneable item
				sword := mocks.CreateTestItem("sword-456", "Longsword", models.ItemTypeWeapon, 15, 3.0)
				sword.RequiresAttunement = false

				invItem := mocks.CreateTestInventoryItem("char-123", "sword-456", 1, true, false, sword)
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{invItem}, nil)
			},
			expectedError: "item does not require attunement",
		},
		{
			name:        "already attuned to item",
			characterID: "char-123",
			itemID:      "ring-456",
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with already attuned item
				ring := mocks.CreateTestItem("ring-456", "Ring of Protection", models.ItemTypeMagic, 500, 0.1)
				ring.RequiresAttunement = true

				invItem := mocks.CreateTestInventoryItem("char-123", "ring-456", 1, true, true, ring)
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{invItem}, nil)
			},
			expectedError: "already attuned to this item",
		},
		{
			name:        "item not in inventory",
			characterID: "char-123",
			itemID:      "nonexistent",
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{}, nil)
			},
			expectedError: "item not found in inventory",
		},
		{
			name:        "inventory retrieval error",
			characterID: "char-123",
			itemID:      "ring-456",
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetCharacterInventory", "char-123").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name:        "attune error",
			characterID: "char-123",
			itemID:      "ring-456",
			setupMock: func(m *mocks.MockInventoryRepository) {
				ring := mocks.CreateTestItem("ring-456", "Ring of Protection", models.ItemTypeMagic, 500, 0.1)
				ring.RequiresAttunement = true

				invItem := mocks.CreateTestInventoryItem("char-123", "ring-456", 1, true, false, ring)
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{invItem}, nil)
				m.On("AttuneItem", "char-123", "ring-456").Return(errors.New("attunement failed"))
			},
			expectedError: "attunement failed",
		},
	}

	runInventoryServiceTest(t, tests, func(service *services.InventoryService, characterID, itemID string) error {
		return service.AttuneToItem(characterID, itemID)
	})
}

func TestInventoryService_UpdateCharacterCurrency(t *testing.T) {
	tests := []struct {
		name          string
		characterID   string
		copper        int
		silver        int
		electrum      int
		gold          int
		platinum      int
		setupMock     func(*mocks.MockInventoryRepository)
		expectedError string
		validate      func(*testing.T, *models.Currency)
	}{
		{
			name:        "add currency successfully",
			characterID: "char-123",
			copper:      5,
			silver:      3,
			gold:        10,
			setupMock: func(m *mocks.MockInventoryRepository) {
				existingCurrency := &models.Currency{
					CharacterID: "char-123",
					Copper:      10,
					Silver:      5,
					Gold:        2,
				}
				m.On("GetCharacterCurrency", "char-123").Return(existingCurrency, nil)
				m.On("UpdateCharacterCurrency", mock.MatchedBy(func(c *models.Currency) bool {
					return c.Copper == 15 && c.Silver == 8 && c.Gold == 12
				})).Return(nil)
			},
		},
		{
			name:        "subtract currency successfully",
			characterID: "char-123",
			copper:      -5,
			silver:      -2,
			gold:        -1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				existingCurrency := &models.Currency{
					CharacterID: "char-123",
					Copper:      10,
					Silver:      5,
					Gold:        2,
				}
				m.On("GetCharacterCurrency", "char-123").Return(existingCurrency, nil)
				m.On("UpdateCharacterCurrency", mock.MatchedBy(func(c *models.Currency) bool {
					return c.Copper == 5 && c.Silver == 3 && c.Gold == 1
				})).Return(nil)
			},
		},
		{
			name:        "insufficient funds",
			characterID: "char-123",
			gold:        -10,
			setupMock: func(m *mocks.MockInventoryRepository) {
				existingCurrency := &models.Currency{
					CharacterID: "char-123",
					Gold:        2,
				}
				m.On("GetCharacterCurrency", "char-123").Return(existingCurrency, nil)
			},
			expectedError: "insufficient funds",
		},
		{
			name:        "get currency error",
			characterID: "char-123",
			gold:        10,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetCharacterCurrency", "char-123").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name:        "update currency error",
			characterID: "char-123",
			gold:        10,
			setupMock: func(m *mocks.MockInventoryRepository) {
				existingCurrency := &models.Currency{
					CharacterID: "char-123",
					Gold:        2,
				}
				m.On("GetCharacterCurrency", "char-123").Return(existingCurrency, nil)
				m.On("UpdateCharacterCurrency", mock.Anything).Return(errors.New("update failed"))
			},
			expectedError: "update failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewInventoryService(mockRepo, nil)
			err := service.UpdateCharacterCurrency(tt.characterID, tt.copper, tt.silver, tt.electrum, tt.gold, tt.platinum)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInventoryService_PurchaseItem(t *testing.T) {
	tests := []struct {
		name          string
		characterID   string
		itemID        string
		quantity      int
		setupMock     func(*mocks.MockInventoryRepository)
		expectedError string
	}{
		{
			name:        "successful purchase",
			characterID: "char-123",
			itemID:      "potion-456",
			quantity:    3,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get item
				potion := mocks.CreateTestItem("potion-456", "Healing Potion", models.ItemTypeConsumable, 50, 0.5)
				m.On("GetItem", "potion-456").Return(potion, nil)

				// Get currency (has 200 gold = 20000 copper)
				currency := &models.Currency{
					CharacterID: "char-123",
					Gold:        200,
				}
				m.On("GetCharacterCurrency", "char-123").Return(currency, nil)

				// Update currency (150 copper cost for 3 potions)
				m.On("UpdateCharacterCurrency", mock.MatchedBy(func(c *models.Currency) bool {
					// 20000 - 150 = 19850 copper = 19 platinum, 8 gold, 1 electrum
					return c.CharacterID == "char-123" &&
						c.Platinum == 19 &&
						c.Gold == 8 &&
						c.Electrum == 1 &&
						c.Silver == 0 &&
						c.Copper == 0
				})).Return(nil)

				// Add items to inventory
				m.On("AddItemToInventory", "char-123", "potion-456", 3).Return(nil)
			},
		},
		{
			name:        "item not found",
			characterID: "char-123",
			itemID:      "nonexistent",
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetItem", "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:        "item nil response",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetItem", "item-456").Return(nil, nil)
			},
			expectedError: "item not found",
		},
		{
			name:        "insufficient funds",
			characterID: "char-123",
			itemID:      "expensive-item",
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Expensive item
				item := mocks.CreateTestItem("expensive-item", "Plate Armor", models.ItemTypeArmor, 150000, 65.0)
				m.On("GetItem", "expensive-item").Return(item, nil)

				// Poor character
				currency := &models.Currency{
					CharacterID: "char-123",
					Gold:        10,
				}
				m.On("GetCharacterCurrency", "char-123").Return(currency, nil)
			},
			expectedError: "insufficient funds",
		},
		{
			name:        "currency update error",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				item := mocks.CreateTestItem("item-456", "Item", models.ItemTypeOther, 10, 1.0)
				m.On("GetItem", "item-456").Return(item, nil)

				currency := &models.Currency{
					CharacterID: "char-123",
					Gold:        100,
				}
				m.On("GetCharacterCurrency", "char-123").Return(currency, nil)
				m.On("UpdateCharacterCurrency", mock.Anything).Return(errors.New("update failed"))
			},
			expectedError: "update failed",
		},
		{
			name:        "add to inventory error",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				item := mocks.CreateTestItem("item-456", "Item", models.ItemTypeOther, 10, 1.0)
				m.On("GetItem", "item-456").Return(item, nil)

				currency := &models.Currency{
					CharacterID: "char-123",
					Gold:        100,
				}
				m.On("GetCharacterCurrency", "char-123").Return(currency, nil)
				m.On("UpdateCharacterCurrency", mock.Anything).Return(nil)
				m.On("AddItemToInventory", "char-123", "item-456", 1).Return(errors.New("inventory full"))
			},
			expectedError: "inventory full",
		},
	}

	runInventoryServiceTestWithQuantity(t, tests, func(service *services.InventoryService, characterID, itemID string, quantity int) error {
		return service.PurchaseItem(characterID, itemID, quantity)
	})
}

func TestInventoryService_SellItem(t *testing.T) {
	tests := []struct {
		name          string
		characterID   string
		itemID        string
		quantity      int
		setupMock     func(*mocks.MockInventoryRepository)
		expectedError string
	}{
		{
			name:        "successful sale",
			characterID: "char-123",
			itemID:      "sword-456",
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get item (worth 100 gold)
				sword := mocks.CreateTestItem("sword-456", "Longsword", models.ItemTypeWeapon, 10000, 3.0)
				m.On("GetItem", "sword-456").Return(sword, nil)

				// Remove from inventory
				m.On("RemoveItemFromInventory", "char-123", "sword-456", 1).Return(nil)

				// Get current currency
				currency := &models.Currency{
					CharacterID: "char-123",
					Gold:        10,
				}
				m.On("GetCharacterCurrency", "char-123").Return(currency, nil)

				// Update currency (sale price is 50% = 5000 copper)
				m.On("UpdateCharacterCurrency", mock.MatchedBy(func(c *models.Currency) bool {
					// 1000 + 5000 = 6000 copper = 6 platinum
					return c.CharacterID == "char-123" &&
						c.Platinum == 6 &&
						c.Gold == 0 &&
						c.Electrum == 0 &&
						c.Silver == 0 &&
						c.Copper == 0
				})).Return(nil)
			},
		},
		{
			name:        "item not found",
			characterID: "char-123",
			itemID:      "nonexistent",
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetItem", "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:        "remove from inventory error",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				item := mocks.CreateTestItem("item-456", "Item", models.ItemTypeOther, 100, 1.0)
				m.On("GetItem", "item-456").Return(item, nil)
				m.On("RemoveItemFromInventory", "char-123", "item-456", 1).Return(errors.New("item not in inventory"))
			},
			expectedError: "item not in inventory",
		},
		{
			name:        "currency retrieval error",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				item := mocks.CreateTestItem("item-456", "Item", models.ItemTypeOther, 100, 1.0)
				m.On("GetItem", "item-456").Return(item, nil)
				m.On("RemoveItemFromInventory", "char-123", "item-456", 1).Return(nil)
				m.On("GetCharacterCurrency", "char-123").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name:        "sell multiple items",
			characterID: "char-123",
			itemID:      "arrow-789",
			quantity:    20,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Arrows worth 1 copper each
				arrow := mocks.CreateTestItem("arrow-789", "Arrow", models.ItemTypeOther, 1, 0.05)
				m.On("GetItem", "arrow-789").Return(arrow, nil)

				m.On("RemoveItemFromInventory", "char-123", "arrow-789", 20).Return(nil)

				currency := &models.Currency{
					CharacterID: "char-123",
					Copper:      5,
				}
				m.On("GetCharacterCurrency", "char-123").Return(currency, nil)

				// Sale price is 10 copper (20 * 1 / 2)
				m.On("UpdateCharacterCurrency", mock.MatchedBy(func(c *models.Currency) bool {
					// 5 + 10 = 15 copper = 1 silver, 5 copper
					return c.Silver == 1 && c.Copper == 5
				})).Return(nil)
			},
		},
	}

	runInventoryServiceTestWithQuantity(t, tests, func(service *services.InventoryService, characterID, itemID string, quantity int) error {
		return service.SellItem(characterID, itemID, quantity)
	})
}

func TestInventoryService_GetCharacterWeight(t *testing.T) {
	tests := []struct {
		name          string
		characterID   string
		setupMock     func(*mocks.MockInventoryRepository)
		expected      *models.InventoryWeight
		expectedError string
	}{
		{
			name:        "get weight successfully",
			characterID: "char-123",
			setupMock: func(m *mocks.MockInventoryRepository) {
				weight := &models.InventoryWeight{
					CurrentWeight:     75.5,
					CarryCapacity:     150.0,
					Encumbered:        false,
					HeavilyEncumbered: false,
				}
				m.On("GetCharacterWeight", "char-123").Return(weight, nil)
			},
			expected: &models.InventoryWeight{
				CurrentWeight:     75.5,
				CarryCapacity:     150.0,
				Encumbered:        false,
				HeavilyEncumbered: false,
			},
		},
		{
			name:        "encumbered character",
			characterID: "char-456",
			setupMock: func(m *mocks.MockInventoryRepository) {
				weight := &models.InventoryWeight{
					CurrentWeight:     180.0,
					CarryCapacity:     150.0,
					Encumbered:        true,
					HeavilyEncumbered: false,
				}
				m.On("GetCharacterWeight", "char-456").Return(weight, nil)
			},
			expected: &models.InventoryWeight{
				CurrentWeight:     180.0,
				CarryCapacity:     150.0,
				Encumbered:        true,
				HeavilyEncumbered: false,
			},
		},
		{
			name:        "repository error",
			characterID: "char-789",
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetCharacterWeight", "char-789").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewInventoryService(mockRepo, nil)
			weight, err := service.GetCharacterWeight(tt.characterID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, weight)
			} else {
				require.NoError(t, err)
				require.NotNil(t, weight)
				assert.Equal(t, tt.expected, weight)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInventoryService_CreateItem(t *testing.T) {
	tests := []struct {
		name          string
		item          *models.Item
		setupMock     func(*mocks.MockInventoryRepository)
		expectedError string
	}{
		{
			name: "create item successfully",
			item: &models.Item{
				Name:               "Custom Sword",
				Type:               models.ItemTypeWeapon,
				Rarity:             models.ItemRarityUncommon,
				Weight:             3.5,
				Value:              500,
				RequiresAttunement: false,
			},
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("CreateItem", mock.AnythingOfType("*models.Item")).Return(nil)
			},
		},
		{
			name: "create magic item with attunement",
			item: &models.Item{
				Name:                   "Ring of Spell Storing",
				Type:                   models.ItemTypeMagic,
				Rarity:                 models.ItemRarityRare,
				Weight:                 0.1,
				Value:                  5000,
				RequiresAttunement:     true,
				AttunementRequirements: "by a spellcaster",
			},
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("CreateItem", mock.AnythingOfType("*models.Item")).Return(nil)
			},
		},
		{
			name: "repository error",
			item: &models.Item{
				Name: "Failed Item",
			},
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("CreateItem", mock.AnythingOfType("*models.Item")).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewInventoryService(mockRepo, nil)
			err := service.CreateItem(tt.item)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInventoryService_GetItemsByType(t *testing.T) {
	tests := []struct {
		name          string
		itemType      models.ItemType
		setupMock     func(*mocks.MockInventoryRepository)
		expected      []*models.Item
		expectedError string
	}{
		{
			name:     "get weapons successfully",
			itemType: models.ItemTypeWeapon,
			setupMock: func(m *mocks.MockInventoryRepository) {
				weapons := []*models.Item{
					mocks.CreateTestItem("sword-1", "Longsword", models.ItemTypeWeapon, 15, 3.0),
					mocks.CreateTestItem("axe-1", "Battleaxe", models.ItemTypeWeapon, 10, 4.0),
					mocks.CreateTestItem("bow-1", "Longbow", models.ItemTypeWeapon, 50, 2.0),
				}
				m.On("GetItemsByType", models.ItemTypeWeapon).Return(weapons, nil)
			},
			expected: []*models.Item{
				mocks.CreateTestItem("sword-1", "Longsword", models.ItemTypeWeapon, 15, 3.0),
				mocks.CreateTestItem("axe-1", "Battleaxe", models.ItemTypeWeapon, 10, 4.0),
				mocks.CreateTestItem("bow-1", "Longbow", models.ItemTypeWeapon, 50, 2.0),
			},
		},
		{
			name:     "get consumables - empty result",
			itemType: models.ItemTypeConsumable,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetItemsByType", models.ItemTypeConsumable).Return([]*models.Item{}, nil)
			},
			expected: []*models.Item{},
		},
		{
			name:     "repository error",
			itemType: models.ItemTypeMagic,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetItemsByType", models.ItemTypeMagic).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewInventoryService(mockRepo, nil)
			items, err := service.GetItemsByType(tt.itemType)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, items)
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(items))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Test simple pass-through methods
func TestInventoryService_PassThroughMethods(t *testing.T) {
	mockInvRepo := new(mocks.MockInventoryRepository)
	service := services.NewInventoryService(mockInvRepo, nil)

	t.Run("RemoveItemFromCharacter", func(t *testing.T) {
		mockInvRepo.On("RemoveItemFromInventory", "char-123", "item-456", 2).Return(nil).Once()
		err := service.RemoveItemFromCharacter("char-123", "item-456", 2)
		assert.NoError(t, err)
	})

	t.Run("GetCharacterInventory", func(t *testing.T) {
		expected := []*models.InventoryItem{
			mocks.CreateTestInventoryItem("char-123", "item-456", 1, false, false, nil),
		}
		mockInvRepo.On("GetCharacterInventory", "char-123").Return(expected, nil).Once()

		inventory, err := service.GetCharacterInventory("char-123")
		assert.NoError(t, err)
		assert.Equal(t, expected, inventory)
	})

	t.Run("UnequipItem", func(t *testing.T) {
		mockInvRepo.On("EquipItem", "char-123", "item-456", false).Return(nil).Once()
		err := service.UnequipItem("char-123", "item-456")
		assert.NoError(t, err)
	})

	t.Run("UnattuneFromItem", func(t *testing.T) {
		mockInvRepo.On("UnattuneItem", "char-123", "item-456").Return(nil).Once()
		err := service.UnattuneFromItem("char-123", "item-456")
		assert.NoError(t, err)
	})

	t.Run("GetCharacterCurrency", func(t *testing.T) {
		expected := &models.Currency{
			CharacterID: "char-123",
			Gold:        100,
		}
		mockInvRepo.On("GetCharacterCurrency", "char-123").Return(expected, nil).Once()

		currency, err := service.GetCharacterCurrency("char-123")
		assert.NoError(t, err)
		assert.Equal(t, expected, currency)
	})

	mockInvRepo.AssertExpectations(t)
}
