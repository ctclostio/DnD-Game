package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// MockInventoryRepository is a mock implementation of database.InventoryRepository
type MockInventoryRepository struct {
	mock.Mock
}

// CreateItem mocks the CreateItem method
func (m *MockInventoryRepository) CreateItem(item *models.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

// GetItem mocks the GetItem method
func (m *MockInventoryRepository) GetItem(itemID string) (*models.Item, error) {
	args := m.Called(itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

// GetItemsByType mocks the GetItemsByType method
func (m *MockInventoryRepository) GetItemsByType(itemType models.ItemType) ([]*models.Item, error) {
	args := m.Called(itemType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Item), args.Error(1)
}

// AddItemToInventory mocks the AddItemToInventory method
func (m *MockInventoryRepository) AddItemToInventory(characterID, itemID string, quantity int) error {
	args := m.Called(characterID, itemID, quantity)
	return args.Error(0)
}

// RemoveItemFromInventory mocks the RemoveItemFromInventory method
func (m *MockInventoryRepository) RemoveItemFromInventory(characterID, itemID string, quantity int) error {
	args := m.Called(characterID, itemID, quantity)
	return args.Error(0)
}

// GetCharacterInventory mocks the GetCharacterInventory method
func (m *MockInventoryRepository) GetCharacterInventory(characterID string) ([]*models.InventoryItem, error) {
	args := m.Called(characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.InventoryItem), args.Error(1)
}

// EquipItem mocks the EquipItem method
func (m *MockInventoryRepository) EquipItem(characterID, itemID string, equip bool) error {
	args := m.Called(characterID, itemID, equip)
	return args.Error(0)
}

// AttuneItem mocks the AttuneItem method
func (m *MockInventoryRepository) AttuneItem(characterID, itemID string) error {
	args := m.Called(characterID, itemID)
	return args.Error(0)
}

// UnattuneItem mocks the UnattuneItem method
func (m *MockInventoryRepository) UnattuneItem(characterID, itemID string) error {
	args := m.Called(characterID, itemID)
	return args.Error(0)
}

// GetCharacterCurrency mocks the GetCharacterCurrency method
func (m *MockInventoryRepository) GetCharacterCurrency(characterID string) (*models.Currency, error) {
	args := m.Called(characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Currency), args.Error(1)
}

// CreateCharacterCurrency mocks the CreateCharacterCurrency method
func (m *MockInventoryRepository) CreateCharacterCurrency(currency *models.Currency) error {
	args := m.Called(currency)
	return args.Error(0)
}

// UpdateCharacterCurrency mocks the UpdateCharacterCurrency method
func (m *MockInventoryRepository) UpdateCharacterCurrency(currency *models.Currency) error {
	args := m.Called(currency)
	return args.Error(0)
}

// GetCharacterWeight mocks the GetCharacterWeight method
func (m *MockInventoryRepository) GetCharacterWeight(characterID string) (*models.InventoryWeight, error) {
	args := m.Called(characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InventoryWeight), args.Error(1)
}

// Helper function to create test items
func createTestItem(id, name string, itemType models.ItemType, value int, weight float64) *models.Item {
	return &models.Item{
		ID:         id,
		Name:       name,
		Type:       itemType,
		Rarity:     models.ItemRarityCommon,
		Weight:     weight,
		Value:      value,
		Properties: make(models.ItemProperties),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Helper function to create test inventory item
func createTestInventoryItem(characterID, itemID string, quantity int, equipped, attuned bool, item *models.Item) *models.InventoryItem {
	return &models.InventoryItem{
		ID:          fmt.Sprintf("inv-%s-%s", characterID, itemID),
		CharacterID: characterID,
		ItemID:      itemID,
		Quantity:    quantity,
		Equipped:    equipped,
		Attuned:     attuned,
		Item:        item,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestInventoryService_AddItemToCharacter(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		characterID   string
		itemID        string
		quantity      int
		setupMock     func(*MockInventoryRepository, *MockCharacterRepository)
		expectedError string
	}{
		{
			name:        "successful add item",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    2,
			setupMock: func(mi *MockInventoryRepository, mc *MockCharacterRepository) {
				// Character exists
				mc.On("GetByID", ctx, "char-123").Return(&models.Character{
					ID:   "char-123",
					Name: "Test Character",
				}, nil)
				// Item exists
				mi.On("GetItem", "item-456").Return(&models.Item{
					ID:    "item-456",
					Name:  "Healing Potion",
					Type:  models.ItemTypeConsumable,
					Value: 50,
				}, nil)
				// Add to inventory
				mi.On("AddItemToInventory", "char-123", "item-456", 2).Return(nil)
			},
		},
		{
			name:        "character not found",
			characterID: "nonexistent",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(mi *MockInventoryRepository, mc *MockCharacterRepository) {
				mc.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:        "character nil response",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(mi *MockInventoryRepository, mc *MockCharacterRepository) {
				mc.On("GetByID", ctx, "char-123").Return(nil, nil)
			},
			expectedError: "character not found",
		},
		{
			name:        "item not found",
			characterID: "char-123",
			itemID:      "nonexistent",
			quantity:    1,
			setupMock: func(mi *MockInventoryRepository, mc *MockCharacterRepository) {
				mc.On("GetByID", ctx, "char-123").Return(&models.Character{
					ID: "char-123",
				}, nil)
				mi.On("GetItem", "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:        "item nil response",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(mi *MockInventoryRepository, mc *MockCharacterRepository) {
				mc.On("GetByID", ctx, "char-123").Return(&models.Character{
					ID: "char-123",
				}, nil)
				mi.On("GetItem", "item-456").Return(nil, nil)
			},
			expectedError: "item not found",
		},
		{
			name:        "add to inventory error",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(mi *MockInventoryRepository, mc *MockCharacterRepository) {
				mc.On("GetByID", ctx, "char-123").Return(&models.Character{
					ID: "char-123",
				}, nil)
				mi.On("GetItem", "item-456").Return(&models.Item{
					ID: "item-456",
				}, nil)
				mi.On("AddItemToInventory", "char-123", "item-456", 1).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name:        "zero quantity",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    0,
			setupMock: func(mi *MockInventoryRepository, mc *MockCharacterRepository) {
				mc.On("GetByID", ctx, "char-123").Return(&models.Character{
					ID: "char-123",
				}, nil)
				mi.On("GetItem", "item-456").Return(&models.Item{
					ID: "item-456",
				}, nil)
				mi.On("AddItemToInventory", "char-123", "item-456", 0).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvRepo := new(MockInventoryRepository)
			mockCharRepo := new(MockCharacterRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockInvRepo, mockCharRepo)
			}

			service := NewInventoryService(mockInvRepo, mockCharRepo)
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
		setupMock     func(*MockInventoryRepository)
		expectedError string
	}{
		{
			name:        "equip weapon successfully",
			characterID: "char-123",
			itemID:      "sword-456",
			setupMock: func(m *MockInventoryRepository) {
				// Get inventory with one-handed sword
				sword := createTestItem("sword-456", "Longsword", models.ItemTypeWeapon, 15, 3.0)
				sword.Properties["two_handed"] = false
				
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{
					createTestInventoryItem("char-123", "sword-456", 1, false, false, sword),
				}, nil)
				// Equip the sword
				m.On("EquipItem", "char-123", "sword-456", true).Return(nil)
			},
		},
		{
			name:        "equip two-handed weapon",
			characterID: "char-123",
			itemID:      "greatsword-789",
			setupMock: func(m *MockInventoryRepository) {
				// Get inventory with two-handed sword
				greatsword := createTestItem("greatsword-789", "Greatsword", models.ItemTypeWeapon, 50, 6.0)
				greatsword.Properties["two_handed"] = true
				
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{
					createTestInventoryItem("char-123", "greatsword-789", 1, false, false, greatsword),
				}, nil)
				// Equip the greatsword
				m.On("EquipItem", "char-123", "greatsword-789", true).Return(nil)
			},
		},
		{
			name:        "equip armor - unequip existing",
			characterID: "char-123",
			itemID:      "plate-mail",
			setupMock: func(m *MockInventoryRepository) {
				// Get inventory with existing armor equipped
				oldArmor := createTestItem("leather-armor", "Leather Armor", models.ItemTypeArmor, 10, 10.0)
				newArmor := createTestItem("plate-mail", "Plate Mail", models.ItemTypeArmor, 1500, 65.0)
				
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{
					createTestInventoryItem("char-123", "leather-armor", 1, true, false, oldArmor),
					createTestInventoryItem("char-123", "plate-mail", 1, false, false, newArmor),
				}, nil)
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
			setupMock: func(m *MockInventoryRepository) {
				// Already have two one-handed weapons equipped
				sword1 := createTestItem("sword-1", "Short Sword", models.ItemTypeWeapon, 10, 2.0)
				sword1.Properties["two_handed"] = false
				sword2 := createTestItem("sword-2", "Dagger", models.ItemTypeWeapon, 2, 1.0)
				sword2.Properties["two_handed"] = false
				greatsword := createTestItem("greatsword-789", "Greatsword", models.ItemTypeWeapon, 50, 6.0)
				greatsword.Properties["two_handed"] = true
				
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{
					createTestInventoryItem("char-123", "sword-1", 1, true, false, sword1),
					createTestInventoryItem("char-123", "sword-2", 1, true, false, sword2),
					createTestInventoryItem("char-123", "greatsword-789", 1, false, false, greatsword),
				}, nil)
			},
			expectedError: "not enough hands to equip this weapon",
		},
		{
			name:        "item not in inventory",
			characterID: "char-123",
			itemID:      "nonexistent",
			setupMock: func(m *MockInventoryRepository) {
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{}, nil)
			},
			expectedError: "item not found in inventory",
		},
		{
			name:        "inventory retrieval error",
			characterID: "char-123",
			itemID:      "item-456",
			setupMock: func(m *MockInventoryRepository) {
				m.On("GetCharacterInventory", "char-123").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewInventoryService(mockRepo, nil)
			err := service.EquipItem(tt.characterID, tt.itemID)

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

func TestInventoryService_AttuneToItem(t *testing.T) {
	tests := []struct {
		name          string
		characterID   string
		itemID        string
		setupMock     func(*MockInventoryRepository)
		expectedError string
	}{
		{
			name:        "successful attunement",
			characterID: "char-123",
			itemID:      "ring-456",
			setupMock: func(m *MockInventoryRepository) {
				// Get inventory with attuneable item
				ring := createTestItem("ring-456", "Ring of Protection", models.ItemTypeMagic, 500, 0.1)
				ring.RequiresAttunement = true
				
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{
					createTestInventoryItem("char-123", "ring-456", 1, true, false, ring),
				}, nil)
				// Attune to item
				m.On("AttuneItem", "char-123", "ring-456").Return(nil)
			},
		},
		{
			name:        "item does not require attunement",
			characterID: "char-123",
			itemID:      "sword-456",
			setupMock: func(m *MockInventoryRepository) {
				// Get inventory with non-attuneable item
				sword := createTestItem("sword-456", "Longsword", models.ItemTypeWeapon, 15, 3.0)
				sword.RequiresAttunement = false
				
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{
					createTestInventoryItem("char-123", "sword-456", 1, true, false, sword),
				}, nil)
			},
			expectedError: "item does not require attunement",
		},
		{
			name:        "already attuned to item",
			characterID: "char-123",
			itemID:      "ring-456",
			setupMock: func(m *MockInventoryRepository) {
				// Get inventory with already attuned item
				ring := createTestItem("ring-456", "Ring of Protection", models.ItemTypeMagic, 500, 0.1)
				ring.RequiresAttunement = true
				
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{
					createTestInventoryItem("char-123", "ring-456", 1, true, true, ring),
				}, nil)
			},
			expectedError: "already attuned to this item",
		},
		{
			name:        "item not in inventory",
			characterID: "char-123",
			itemID:      "nonexistent",
			setupMock: func(m *MockInventoryRepository) {
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{}, nil)
			},
			expectedError: "item not found in inventory",
		},
		{
			name:        "inventory retrieval error",
			characterID: "char-123",
			itemID:      "ring-456",
			setupMock: func(m *MockInventoryRepository) {
				m.On("GetCharacterInventory", "char-123").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name:        "attune error",
			characterID: "char-123",
			itemID:      "ring-456",
			setupMock: func(m *MockInventoryRepository) {
				ring := createTestItem("ring-456", "Ring of Protection", models.ItemTypeMagic, 500, 0.1)
				ring.RequiresAttunement = true
				
				m.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{
					createTestInventoryItem("char-123", "ring-456", 1, true, false, ring),
				}, nil)
				m.On("AttuneItem", "char-123", "ring-456").Return(errors.New("attunement failed"))
			},
			expectedError: "attunement failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewInventoryService(mockRepo, nil)
			err := service.AttuneToItem(tt.characterID, tt.itemID)

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

func TestInventoryService_UpdateCharacterCurrency(t *testing.T) {
	tests := []struct {
		name          string
		characterID   string
		copper        int
		silver        int
		electrum      int
		gold          int
		platinum      int
		setupMock     func(*MockInventoryRepository)
		expectedError string
		validate      func(*testing.T, *models.Currency)
	}{
		{
			name:        "add currency successfully",
			characterID: "char-123",
			copper:      5,
			silver:      3,
			gold:        10,
			setupMock: func(m *MockInventoryRepository) {
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
			setupMock: func(m *MockInventoryRepository) {
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
			setupMock: func(m *MockInventoryRepository) {
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
			setupMock: func(m *MockInventoryRepository) {
				m.On("GetCharacterCurrency", "char-123").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name:        "update currency error",
			characterID: "char-123",
			gold:        10,
			setupMock: func(m *MockInventoryRepository) {
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
			mockRepo := new(MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewInventoryService(mockRepo, nil)
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
		setupMock     func(*MockInventoryRepository)
		expectedError string
	}{
		{
			name:        "successful purchase",
			characterID: "char-123",
			itemID:      "potion-456",
			quantity:    3,
			setupMock: func(m *MockInventoryRepository) {
				// Get item
				potion := createTestItem("potion-456", "Healing Potion", models.ItemTypeConsumable, 50, 0.5)
				m.On("GetItem", "potion-456").Return(potion, nil)
				
				// Get currency (has 200 gold = 20000 copper)
				currency := &models.Currency{
					CharacterID: "char-123",
					Gold:        200,
				}
				m.On("GetCharacterCurrency", "char-123").Return(currency, nil)
				
				// Update currency (150 copper cost for 3 potions)
				m.On("UpdateCharacterCurrency", mock.MatchedBy(func(c *models.Currency) bool {
					// 20000 - 150 = 19850 copper = 198 gold, 5 silver
					return c.Gold == 198 && c.Silver == 5 && c.Copper == 0
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
			setupMock: func(m *MockInventoryRepository) {
				m.On("GetItem", "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:        "item nil response",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(m *MockInventoryRepository) {
				m.On("GetItem", "item-456").Return(nil, nil)
			},
			expectedError: "item not found",
		},
		{
			name:        "insufficient funds",
			characterID: "char-123",
			itemID:      "expensive-item",
			quantity:    1,
			setupMock: func(m *MockInventoryRepository) {
				// Expensive item
				item := createTestItem("expensive-item", "Plate Armor", models.ItemTypeArmor, 150000, 65.0)
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
			setupMock: func(m *MockInventoryRepository) {
				item := createTestItem("item-456", "Item", models.ItemTypeOther, 10, 1.0)
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
			setupMock: func(m *MockInventoryRepository) {
				item := createTestItem("item-456", "Item", models.ItemTypeOther, 10, 1.0)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewInventoryService(mockRepo, nil)
			err := service.PurchaseItem(tt.characterID, tt.itemID, tt.quantity)

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

func TestInventoryService_SellItem(t *testing.T) {
	tests := []struct {
		name          string
		characterID   string
		itemID        string
		quantity      int
		setupMock     func(*MockInventoryRepository)
		expectedError string
	}{
		{
			name:        "successful sale",
			characterID: "char-123",
			itemID:      "sword-456",
			quantity:    1,
			setupMock: func(m *MockInventoryRepository) {
				// Get item (worth 100 gold)
				sword := createTestItem("sword-456", "Longsword", models.ItemTypeWeapon, 10000, 3.0)
				m.On("GetItem", "sword-456").Return(sword, nil)
				
				// Remove from inventory
				m.On("RemoveItemFromInventory", "char-123", "sword-456", 1).Return(nil)
				
				// Get current currency
				currency := &models.Currency{
					CharacterID: "char-123",
					Gold:        10,
				}
				m.On("GetCharacterCurrency", "char-123").Return(currency, nil)
				
				// Update currency (sale price is 50% = 5000 copper = 50 gold)
				m.On("UpdateCharacterCurrency", mock.MatchedBy(func(c *models.Currency) bool {
					// 10 gold + 50 gold = 60 gold
					return c.Gold == 60 && c.Silver == 0 && c.Copper == 0
				})).Return(nil)
			},
		},
		{
			name:        "item not found",
			characterID: "char-123",
			itemID:      "nonexistent",
			quantity:    1,
			setupMock: func(m *MockInventoryRepository) {
				m.On("GetItem", "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:        "remove from inventory error",
			characterID: "char-123",
			itemID:      "item-456",
			quantity:    1,
			setupMock: func(m *MockInventoryRepository) {
				item := createTestItem("item-456", "Item", models.ItemTypeOther, 100, 1.0)
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
			setupMock: func(m *MockInventoryRepository) {
				item := createTestItem("item-456", "Item", models.ItemTypeOther, 100, 1.0)
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
			setupMock: func(m *MockInventoryRepository) {
				// Arrows worth 1 copper each
				arrow := createTestItem("arrow-789", "Arrow", models.ItemTypeOther, 1, 0.05)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewInventoryService(mockRepo, nil)
			err := service.SellItem(tt.characterID, tt.itemID, tt.quantity)

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

func TestInventoryService_GetCharacterWeight(t *testing.T) {
	tests := []struct {
		name          string
		characterID   string
		setupMock     func(*MockInventoryRepository)
		expected      *models.InventoryWeight
		expectedError string
	}{
		{
			name:        "get weight successfully",
			characterID: "char-123",
			setupMock: func(m *MockInventoryRepository) {
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
			setupMock: func(m *MockInventoryRepository) {
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
			setupMock: func(m *MockInventoryRepository) {
				m.On("GetCharacterWeight", "char-789").Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewInventoryService(mockRepo, nil)
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
		setupMock     func(*MockInventoryRepository)
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
			setupMock: func(m *MockInventoryRepository) {
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
			setupMock: func(m *MockInventoryRepository) {
				m.On("CreateItem", mock.AnythingOfType("*models.Item")).Return(nil)
			},
		},
		{
			name: "repository error",
			item: &models.Item{
				Name: "Failed Item",
			},
			setupMock: func(m *MockInventoryRepository) {
				m.On("CreateItem", mock.AnythingOfType("*models.Item")).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewInventoryService(mockRepo, nil)
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
		setupMock     func(*MockInventoryRepository)
		expected      []*models.Item
		expectedError string
	}{
		{
			name:     "get weapons successfully",
			itemType: models.ItemTypeWeapon,
			setupMock: func(m *MockInventoryRepository) {
				weapons := []*models.Item{
					createTestItem("sword-1", "Longsword", models.ItemTypeWeapon, 15, 3.0),
					createTestItem("axe-1", "Battleaxe", models.ItemTypeWeapon, 10, 4.0),
					createTestItem("bow-1", "Longbow", models.ItemTypeWeapon, 50, 2.0),
				}
				m.On("GetItemsByType", models.ItemTypeWeapon).Return(weapons, nil)
			},
			expected: []*models.Item{
				createTestItem("sword-1", "Longsword", models.ItemTypeWeapon, 15, 3.0),
				createTestItem("axe-1", "Battleaxe", models.ItemTypeWeapon, 10, 4.0),
				createTestItem("bow-1", "Longbow", models.ItemTypeWeapon, 50, 2.0),
			},
		},
		{
			name:     "get consumables - empty result",
			itemType: models.ItemTypeConsumable,
			setupMock: func(m *MockInventoryRepository) {
				m.On("GetItemsByType", models.ItemTypeConsumable).Return([]*models.Item{}, nil)
			},
			expected: []*models.Item{},
		},
		{
			name:     "repository error",
			itemType: models.ItemTypeMagic,
			setupMock: func(m *MockInventoryRepository) {
				m.On("GetItemsByType", models.ItemTypeMagic).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInventoryRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewInventoryService(mockRepo, nil)
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
	mockInvRepo := new(MockInventoryRepository)
	service := NewInventoryService(mockInvRepo, nil)

	t.Run("RemoveItemFromCharacter", func(t *testing.T) {
		mockInvRepo.On("RemoveItemFromInventory", "char-123", "item-456", 2).Return(nil).Once()
		err := service.RemoveItemFromCharacter("char-123", "item-456", 2)
		assert.NoError(t, err)
	})

	t.Run("GetCharacterInventory", func(t *testing.T) {
		expected := []*models.InventoryItem{
			createTestInventoryItem("char-123", "item-456", 1, false, false, nil),
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

// Test concurrent operations
func TestInventoryService_ConcurrentOperations(t *testing.T) {
	mockInvRepo := new(MockInventoryRepository)
	mockCharRepo := new(MockCharacterRepository)
	service := NewInventoryService(mockInvRepo, mockCharRepo)
	ctx := context.Background()

	// Set up expectations for concurrent calls
	for i := 0; i < 10; i++ {
		characterID := fmt.Sprintf("char-%d", i)
		itemID := fmt.Sprintf("item-%d", i)
		
		// Mock character exists
		mockCharRepo.On("GetByID", ctx, characterID).Return(&models.Character{
			ID: characterID,
		}, nil).Maybe()
		
		// Mock item exists
		mockInvRepo.On("GetItem", itemID).Return(&models.Item{
			ID: itemID,
		}, nil).Maybe()
		
		// Mock add to inventory
		mockInvRepo.On("AddItemToInventory", characterID, itemID, 1).Return(nil).Maybe()
		
		// Mock get inventory
		mockInvRepo.On("GetCharacterInventory", characterID).Return([]*models.InventoryItem{}, nil).Maybe()
	}

	// Run concurrent operations
	done := make(chan bool, 20)
	
	// Add items
	for i := 0; i < 10; i++ {
		go func(id int) {
			characterID := fmt.Sprintf("char-%d", id)
			itemID := fmt.Sprintf("item-%d", id)
			err := service.AddItemToCharacter(characterID, itemID, 1)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Get inventories
	for i := 0; i < 10; i++ {
		go func(id int) {
			characterID := fmt.Sprintf("char-%d", id)
			_, err := service.GetCharacterInventory(characterID)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	mockInvRepo.AssertExpectations(t)
	mockCharRepo.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkInventoryService_AddItemToCharacter(b *testing.B) {
	mockInvRepo := new(MockInventoryRepository)
	mockCharRepo := new(MockCharacterRepository)
	service := NewInventoryService(mockInvRepo, mockCharRepo)
	ctx := context.Background()

	// Set up mocks to always succeed
	mockCharRepo.On("GetByID", ctx, mock.Anything).Return(&models.Character{
		ID: "char-123",
	}, nil)
	mockInvRepo.On("GetItem", mock.Anything).Return(&models.Item{
		ID: "item-456",
	}, nil)
	mockInvRepo.On("AddItemToInventory", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.AddItemToCharacter("char-123", "item-456", 1)
	}
}

func BenchmarkInventoryService_EquipItem(b *testing.B) {
	mockInvRepo := new(MockInventoryRepository)
	service := NewInventoryService(mockInvRepo, nil)

	// Set up mock
	sword := createTestItem("sword-456", "Longsword", models.ItemTypeWeapon, 15, 3.0)
	sword.Properties["two_handed"] = false
	
	mockInvRepo.On("GetCharacterInventory", "char-123").Return([]*models.InventoryItem{
		createTestInventoryItem("char-123", "sword-456", 1, false, false, sword),
	}, nil)
	mockInvRepo.On("EquipItem", "char-123", "sword-456", true).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.EquipItem("char-123", "sword-456")
	}
}