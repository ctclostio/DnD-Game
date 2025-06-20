package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/internal/services/mocks"
)

// Test constants
const (
	// Repository method names
	testMethodGetItem               = "GetItem"
	testMethodGetCharacterInventory = "GetCharacterInventory"
	testMethodGetCharacterCurrency  = "GetCharacterCurrency"
	testMethodUpdateCharCurrency    = "UpdateCharacterCurrency"
	testMethodRemoveItem            = "RemoveItemFromInventory"
	testMethodEquipItem             = "EquipItem"
	testMethodAddItem               = "AddItemToInventory"
	testMethodGetByID               = "GetByID"
	
	// Test values
	testIDNonexistent   = "nonexistent"
	testRaceHuman       = "Human"
	testClassFighter    = "Fighter"
	testItemGreatsword  = "greatsword-789"
	testItemSword1      = "sword-1"
	testItemPotion      = "potion-456"
	testItemPlateMail   = "plate-mail"
	testItemLeatherArmor = "leather-armor"
	testItemExpensive   = "expensive-item"
	testItemArrow       = "arrow-789"
	testPropTwoHanded   = "two_handed"
	
	// Error messages
	testErrNotFound         = "not found"
	testErrItemNotFound     = "item not found"
	testErrUpdateFailed     = "update failed"
	testErrItemNotInInv     = "item not in inventory"
	testErrInsufficientFunds = "insufficient funds"
	testErrInventoryRepository = "repository error"
	
	// Type strings
	testTypeModelsItem = "*models.Item"
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
			characterID: constants.TestCharacterID,
			itemID:      constants.TestItemID,
			quantity:    2,
			setupMock: func(invRepo *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				// Character exists
				char := mocks.CreateTestCharacter(constants.TestCharacterID, constants.TestUserID, constants.TestCharacterName, testRaceHuman, testClassFighter)
				charRepo.On(testMethodGetByID, ctx, constants.TestCharacterID).Return(char, nil)

				// Item exists
				item := mocks.CreateTestItem(constants.TestItemID, constants.TestHealingPotion, models.ItemTypeConsumable, 50, 0.5)
				invRepo.On(testMethodGetItem, constants.TestItemID).Return(item, nil)

				// Add to inventory
				invRepo.On("AddItemToInventory", constants.TestCharacterID, constants.TestItemID, 2).Return(nil)
			},
		},
		{
			name:        constants.TestCharacterNotFound,
			characterID: testIDNonexistent,
			itemID:      constants.TestItemID,
			quantity:    1,
			setupMock: func(_ *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				charRepo.On(testMethodGetByID, ctx, testIDNonexistent).Return(nil, errors.New(testErrNotFound))
			},
			expectedError: testErrNotFound,
		},
		{
			name:        "character nil response",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestItemID,
			quantity:    1,
			setupMock: func(_ *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				charRepo.On(testMethodGetByID, ctx, constants.TestCharacterID).Return(nil, nil)
			},
			expectedError: constants.TestCharacterNotFound,
		},
		{
			name:        testErrItemNotFound,
			characterID: constants.TestCharacterID,
			itemID:      testIDNonexistent,
			quantity:    1,
			setupMock: func(invRepo *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				char := mocks.CreateTestCharacter(constants.TestCharacterID, constants.TestUserID, constants.TestCharacterName, testRaceHuman, testClassFighter)
				charRepo.On(testMethodGetByID, ctx, constants.TestCharacterID).Return(char, nil)
				invRepo.On(testMethodGetItem, testIDNonexistent).Return(nil, errors.New(testErrNotFound))
			},
			expectedError: testErrNotFound,
		},
		{
			name:        "item nil response",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestItemID,
			quantity:    1,
			setupMock: func(invRepo *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				char := mocks.CreateTestCharacter(constants.TestCharacterID, constants.TestUserID, constants.TestCharacterName, testRaceHuman, testClassFighter)
				charRepo.On(testMethodGetByID, ctx, constants.TestCharacterID).Return(char, nil)
				invRepo.On(testMethodGetItem, constants.TestItemID).Return(nil, nil)
			},
			expectedError: testErrItemNotFound,
		},
		{
			name:        "add to inventory error",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestItemID,
			quantity:    1,
			setupMock: func(invRepo *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				char := mocks.CreateTestCharacter(constants.TestCharacterID, constants.TestUserID, constants.TestCharacterName, testRaceHuman, testClassFighter)
				charRepo.On(testMethodGetByID, ctx, constants.TestCharacterID).Return(char, nil)

				item := mocks.CreateTestItem(constants.TestItemID, constants.TestHealingPotion, models.ItemTypeConsumable, 50, 0.5)
				invRepo.On(testMethodGetItem, constants.TestItemID).Return(item, nil)

				invRepo.On("AddItemToInventory", constants.TestCharacterID, constants.TestItemID, 1).Return(errors.New(constants.TestDatabaseError))
			},
			expectedError: constants.TestDatabaseError,
		},
		{
			name:        "zero quantity",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestItemID,
			quantity:    0,
			setupMock: func(invRepo *mocks.MockInventoryRepository, charRepo *mocks.MockCharacterRepository) {
				char := mocks.CreateTestCharacter(constants.TestCharacterID, constants.TestUserID, constants.TestCharacterName, testRaceHuman, testClassFighter)
				charRepo.On(testMethodGetByID, ctx, constants.TestCharacterID).Return(char, nil)

				item := mocks.CreateTestItem(constants.TestItemID, constants.TestHealingPotion, models.ItemTypeConsumable, 50, 0.5)
				invRepo.On(testMethodGetItem, constants.TestItemID).Return(item, nil)

				invRepo.On("AddItemToInventory", constants.TestCharacterID, constants.TestItemID, 0).Return(nil)
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
			characterID: constants.TestCharacterID,
			itemID:      constants.TestSwordID,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with one-handed sword
				sword := mocks.CreateTestItem(constants.TestSwordID, constants.TestLongsword, models.ItemTypeWeapon, 15, 3.0)
				sword.Properties["two_handed"] = false

				invItem := mocks.CreateTestInventoryItem(constants.TestCharacterID, constants.TestSwordID, 1, false, false, sword)
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return([]*models.InventoryItem{invItem}, nil)

				// Equip the sword
				m.On("EquipItem", constants.TestCharacterID, constants.TestSwordID, true).Return(nil)
			},
		},
		{
			name:        "equip two-handed weapon",
			characterID: constants.TestCharacterID,
			itemID:      testItemGreatsword,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with two-handed sword
				greatsword := mocks.CreateTestItem(testItemGreatsword, "Greatsword", models.ItemTypeWeapon, 50, 6.0)
				greatsword.Properties["two_handed"] = true

				invItem := mocks.CreateTestInventoryItem(constants.TestCharacterID, testItemGreatsword, 1, false, false, greatsword)
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return([]*models.InventoryItem{invItem}, nil)

				// Equip the greatsword
				m.On("EquipItem", constants.TestCharacterID, testItemGreatsword, true).Return(nil)
			},
		},
		{
			name:        "equip armor - unequip existing",
			characterID: constants.TestCharacterID,
			itemID:      testItemPlateMail,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with existing armor equipped
				oldArmor := mocks.CreateTestItem(testItemLeatherArmor, "Leather Armor", models.ItemTypeArmor, 10, 10.0)
				newArmor := mocks.CreateTestItem(testItemPlateMail, "Plate Mail", models.ItemTypeArmor, 1500, 65.0)

				invItems := []*models.InventoryItem{
					mocks.CreateTestInventoryItem(constants.TestCharacterID, testItemLeatherArmor, 1, true, false, oldArmor),
					mocks.CreateTestInventoryItem(constants.TestCharacterID, testItemPlateMail, 1, false, false, newArmor),
				}
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return(invItems, nil)

				// Unequip old armor first
				m.On("EquipItem", constants.TestCharacterID, testItemLeatherArmor, false).Return(nil)
				// Equip new armor
				m.On("EquipItem", constants.TestCharacterID, testItemPlateMail, true).Return(nil)
			},
		},
		{
			name:        "not enough hands for two-handed weapon",
			characterID: constants.TestCharacterID,
			itemID:      testItemGreatsword,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Already have two one-handed weapons equipped
				sword1 := mocks.CreateTestItem(testItemSword1, "Short Sword", models.ItemTypeWeapon, 10, 2.0)
				sword1.Properties["two_handed"] = false
				sword2 := mocks.CreateTestItem("sword-2", "Dagger", models.ItemTypeWeapon, 2, 1.0)
				sword2.Properties["two_handed"] = false
				greatsword := mocks.CreateTestItem(testItemGreatsword, "Greatsword", models.ItemTypeWeapon, 50, 6.0)
				greatsword.Properties["two_handed"] = true

				invItems := []*models.InventoryItem{
					mocks.CreateTestInventoryItem(constants.TestCharacterID, testItemSword1, 1, true, false, sword1),
					mocks.CreateTestInventoryItem(constants.TestCharacterID, "sword-2", 1, true, false, sword2),
					mocks.CreateTestInventoryItem(constants.TestCharacterID, testItemGreatsword, 1, false, false, greatsword),
				}
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return(invItems, nil)
			},
			expectedError: "not enough hands to equip this weapon",
		},
		{
			name:        testErrItemNotInInv,
			characterID: constants.TestCharacterID,
			itemID:      testIDNonexistent,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return([]*models.InventoryItem{}, nil)
			},
			expectedError: "item not found in inventory",
		},
		{
			name:        "inventory retrieval error",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestItemID,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return(nil, errors.New(constants.TestDatabaseError))
			},
			expectedError: constants.TestDatabaseError,
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
			characterID: constants.TestCharacterID,
			itemID:      constants.TestRingID,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with attuneable item
				ring := mocks.CreateTestItem(constants.TestRingID, constants.TestRingOfProtection, models.ItemTypeMagic, 500, 0.1)
				ring.RequiresAttunement = true

				invItem := mocks.CreateTestInventoryItem(constants.TestCharacterID, constants.TestRingID, 1, true, false, ring)
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return([]*models.InventoryItem{invItem}, nil)

				// Attune to item
				m.On("AttuneItem", constants.TestCharacterID, constants.TestRingID).Return(nil)
			},
		},
		{
			name:        "item does not require attunement",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestSwordID,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with non-attuneable item
				sword := mocks.CreateTestItem(constants.TestSwordID, constants.TestLongsword, models.ItemTypeWeapon, 15, 3.0)
				sword.RequiresAttunement = false

				invItem := mocks.CreateTestInventoryItem(constants.TestCharacterID, constants.TestSwordID, 1, true, false, sword)
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return([]*models.InventoryItem{invItem}, nil)
			},
			expectedError: "item does not require attunement",
		},
		{
			name:        "already attuned to item",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestRingID,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get inventory with already attuned item
				ring := mocks.CreateTestItem(constants.TestRingID, constants.TestRingOfProtection, models.ItemTypeMagic, 500, 0.1)
				ring.RequiresAttunement = true

				invItem := mocks.CreateTestInventoryItem(constants.TestCharacterID, constants.TestRingID, 1, true, true, ring)
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return([]*models.InventoryItem{invItem}, nil)
			},
			expectedError: "already attuned to this item",
		},
		{
			name:        testErrItemNotInInv,
			characterID: constants.TestCharacterID,
			itemID:      testIDNonexistent,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return([]*models.InventoryItem{}, nil)
			},
			expectedError: "item not found in inventory",
		},
		{
			name:        "inventory retrieval error",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestRingID,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return(nil, errors.New(constants.TestDatabaseError))
			},
			expectedError: constants.TestDatabaseError,
		},
		{
			name:        "attune error",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestRingID,
			setupMock: func(m *mocks.MockInventoryRepository) {
				ring := mocks.CreateTestItem(constants.TestRingID, constants.TestRingOfProtection, models.ItemTypeMagic, 500, 0.1)
				ring.RequiresAttunement = true

				invItem := mocks.CreateTestInventoryItem(constants.TestCharacterID, constants.TestRingID, 1, true, false, ring)
				m.On(testMethodGetCharacterInventory, constants.TestCharacterID).Return([]*models.InventoryItem{invItem}, nil)
				m.On("AttuneItem", constants.TestCharacterID, constants.TestRingID).Return(errors.New("attunement failed"))
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
			characterID: constants.TestCharacterID,
			copper:      5,
			silver:      3,
			gold:        10,
			setupMock: func(m *mocks.MockInventoryRepository) {
				existingCurrency := &models.Currency{
					CharacterID: constants.TestCharacterID,
					Copper:      10,
					Silver:      5,
					Gold:        2,
				}
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(existingCurrency, nil)
				m.On("UpdateCharacterCurrency", mock.MatchedBy(func(c *models.Currency) bool {
					return c.Copper == 15 && c.Silver == 8 && c.Gold == 12
				})).Return(nil)
			},
		},
		{
			name:        "subtract currency successfully",
			characterID: constants.TestCharacterID,
			copper:      -5,
			silver:      -2,
			gold:        -1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				existingCurrency := &models.Currency{
					CharacterID: constants.TestCharacterID,
					Copper:      10,
					Silver:      5,
					Gold:        2,
				}
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(existingCurrency, nil)
				m.On("UpdateCharacterCurrency", mock.MatchedBy(func(c *models.Currency) bool {
					return c.Copper == 5 && c.Silver == 3 && c.Gold == 1
				})).Return(nil)
			},
		},
		{
			name:        testErrInsufficientFunds,
			characterID: constants.TestCharacterID,
			gold:        -10,
			setupMock: func(m *mocks.MockInventoryRepository) {
				existingCurrency := &models.Currency{
					CharacterID: constants.TestCharacterID,
					Gold:        2,
				}
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(existingCurrency, nil)
			},
			expectedError: testErrInsufficientFunds,
		},
		{
			name:        "get currency error",
			characterID: constants.TestCharacterID,
			gold:        10,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(nil, errors.New(constants.TestDatabaseError))
			},
			expectedError: constants.TestDatabaseError,
		},
		{
			name:        "update currency error",
			characterID: constants.TestCharacterID,
			gold:        10,
			setupMock: func(m *mocks.MockInventoryRepository) {
				existingCurrency := &models.Currency{
					CharacterID: constants.TestCharacterID,
					Gold:        2,
				}
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(existingCurrency, nil)
				m.On("UpdateCharacterCurrency", mock.Anything).Return(errors.New(testErrUpdateFailed))
			},
			expectedError: testErrUpdateFailed,
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
			characterID: constants.TestCharacterID,
			itemID:      testItemPotion,
			quantity:    3,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get item
				potion := mocks.CreateTestItem(testItemPotion, "Healing Potion", models.ItemTypeConsumable, 50, 0.5)
				m.On("GetItem", testItemPotion).Return(potion, nil)

				// Get currency (has 200 gold = 20000 copper)
				currency := &models.Currency{
					CharacterID: constants.TestCharacterID,
					Gold:        200,
				}
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(currency, nil)

				// Update currency (150 copper cost for 3 potions)
				m.On("UpdateCharacterCurrency", mock.MatchedBy(func(c *models.Currency) bool {
					// 20000 - 150 = 19850 copper = 19 platinum, 8 gold, 1 electrum
					return c.CharacterID == constants.TestCharacterID &&
						c.Platinum == 19 &&
						c.Gold == 8 &&
						c.Electrum == 1 &&
						c.Silver == 0 &&
						c.Copper == 0
				})).Return(nil)

				// Add items to inventory
				m.On("AddItemToInventory", constants.TestCharacterID, testItemPotion, 3).Return(nil)
			},
		},
		{
			name:        testErrItemNotFound,
			characterID: constants.TestCharacterID,
			itemID:      testIDNonexistent,
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetItem", "nonexistent").Return(nil, errors.New(testErrNotFound))
			},
			expectedError: testErrNotFound,
		},
		{
			name:        "item nil response",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestItemID,
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetItem", constants.TestItemID).Return(nil, nil)
			},
			expectedError: testErrItemNotFound,
		},
		{
			name:        testErrInsufficientFunds,
			characterID: constants.TestCharacterID,
			itemID:      testItemExpensive,
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Expensive item
				item := mocks.CreateTestItem("expensive-item", "Plate Armor", models.ItemTypeArmor, 150000, 65.0)
				m.On("GetItem", testItemExpensive).Return(item, nil)

				// Poor character
				currency := &models.Currency{
					CharacterID: constants.TestCharacterID,
					Gold:        10,
				}
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(currency, nil)
			},
			expectedError: testErrInsufficientFunds,
		},
		{
			name:        "currency update error",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestItemID,
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				item := mocks.CreateTestItem(constants.TestItemID, "Item", models.ItemTypeOther, 10, 1.0)
				m.On("GetItem", constants.TestItemID).Return(item, nil)

				currency := &models.Currency{
					CharacterID: constants.TestCharacterID,
					Gold:        100,
				}
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(currency, nil)
				m.On("UpdateCharacterCurrency", mock.Anything).Return(errors.New(testErrUpdateFailed))
			},
			expectedError: testErrUpdateFailed,
		},
		{
			name:        "add to inventory error",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestItemID,
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				item := mocks.CreateTestItem(constants.TestItemID, "Item", models.ItemTypeOther, 10, 1.0)
				m.On("GetItem", constants.TestItemID).Return(item, nil)

				currency := &models.Currency{
					CharacterID: constants.TestCharacterID,
					Gold:        100,
				}
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(currency, nil)
				m.On("UpdateCharacterCurrency", mock.Anything).Return(nil)
				m.On("AddItemToInventory", constants.TestCharacterID, constants.TestItemID, 1).Return(errors.New("inventory full"))
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
			characterID: constants.TestCharacterID,
			itemID:      constants.TestSwordID,
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Get item (worth 100 gold)
				sword := mocks.CreateTestItem(constants.TestSwordID, "Longsword", models.ItemTypeWeapon, 10000, 3.0)
				m.On("GetItem", constants.TestSwordID).Return(sword, nil)

				// Remove from inventory
				m.On("RemoveItemFromInventory", constants.TestCharacterID, constants.TestSwordID, 1).Return(nil)

				// Get current currency
				currency := &models.Currency{
					CharacterID: constants.TestCharacterID,
					Gold:        10,
				}
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(currency, nil)

				// Update currency (sale price is 50% = 5000 copper)
				m.On("UpdateCharacterCurrency", mock.MatchedBy(func(c *models.Currency) bool {
					// 1000 + 5000 = 6000 copper = 6 platinum
					return c.CharacterID == constants.TestCharacterID &&
						c.Platinum == 6 &&
						c.Gold == 0 &&
						c.Electrum == 0 &&
						c.Silver == 0 &&
						c.Copper == 0
				})).Return(nil)
			},
		},
		{
			name:        testErrItemNotFound,
			characterID: constants.TestCharacterID,
			itemID:      testIDNonexistent,
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetItem", "nonexistent").Return(nil, errors.New(testErrNotFound))
			},
			expectedError: testErrNotFound,
		},
		{
			name:        "remove from inventory error",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestItemID,
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				item := mocks.CreateTestItem(constants.TestItemID, "Item", models.ItemTypeOther, 100, 1.0)
				m.On("GetItem", constants.TestItemID).Return(item, nil)
				m.On("RemoveItemFromInventory", constants.TestCharacterID, constants.TestItemID, 1).Return(errors.New(testErrItemNotInInv))
			},
			expectedError: testErrItemNotInInv,
		},
		{
			name:        "currency retrieval error",
			characterID: constants.TestCharacterID,
			itemID:      constants.TestItemID,
			quantity:    1,
			setupMock: func(m *mocks.MockInventoryRepository) {
				item := mocks.CreateTestItem(constants.TestItemID, "Item", models.ItemTypeOther, 100, 1.0)
				m.On("GetItem", constants.TestItemID).Return(item, nil)
				m.On("RemoveItemFromInventory", constants.TestCharacterID, constants.TestItemID, 1).Return(nil)
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(nil, errors.New(constants.TestDatabaseError))
			},
			expectedError: constants.TestDatabaseError,
		},
		{
			name:        "sell multiple items",
			characterID: constants.TestCharacterID,
			itemID:      testItemArrow,
			quantity:    20,
			setupMock: func(m *mocks.MockInventoryRepository) {
				// Arrows worth 1 copper each
				arrow := mocks.CreateTestItem(testItemArrow, "Arrow", models.ItemTypeOther, 1, 0.05)
				m.On("GetItem", testItemArrow).Return(arrow, nil)

				m.On("RemoveItemFromInventory", constants.TestCharacterID, testItemArrow, 20).Return(nil)

				currency := &models.Currency{
					CharacterID: constants.TestCharacterID,
					Copper:      5,
				}
				m.On(testMethodGetCharacterCurrency, constants.TestCharacterID).Return(currency, nil)

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
			characterID: constants.TestCharacterID,
			setupMock: func(m *mocks.MockInventoryRepository) {
				weight := &models.InventoryWeight{
					CurrentWeight:     75.5,
					CarryCapacity:     150.0,
					Encumbered:        false,
					HeavilyEncumbered: false,
				}
				m.On("GetCharacterWeight", constants.TestCharacterID).Return(weight, nil)
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
			name:        testErrInventoryRepository,
			characterID: "char-789",
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetCharacterWeight", "char-789").Return(nil, errors.New(constants.TestDatabaseError))
			},
			expectedError: constants.TestDatabaseError,
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
				m.On("CreateItem", mock.AnythingOfType(testTypeModelsItem)).Return(nil)
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
				m.On("CreateItem", mock.AnythingOfType(testTypeModelsItem)).Return(nil)
			},
		},
		{
			name: testErrInventoryRepository,
			item: &models.Item{
				Name: "Failed Item",
			},
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("CreateItem", mock.AnythingOfType(testTypeModelsItem)).Return(errors.New(constants.TestDatabaseError))
			},
			expectedError: constants.TestDatabaseError,
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
					mocks.CreateTestItem(testItemSword1, "Longsword", models.ItemTypeWeapon, 15, 3.0),
					mocks.CreateTestItem("axe-1", "Battleaxe", models.ItemTypeWeapon, 10, 4.0),
					mocks.CreateTestItem("bow-1", "Longbow", models.ItemTypeWeapon, 50, 2.0),
				}
				m.On("GetItemsByType", models.ItemTypeWeapon).Return(weapons, nil)
			},
			expected: []*models.Item{
				mocks.CreateTestItem(testItemSword1, "Longsword", models.ItemTypeWeapon, 15, 3.0),
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
			name:     testErrInventoryRepository,
			itemType: models.ItemTypeMagic,
			setupMock: func(m *mocks.MockInventoryRepository) {
				m.On("GetItemsByType", models.ItemTypeMagic).Return(nil, errors.New(constants.TestDatabaseError))
			},
			expectedError: constants.TestDatabaseError,
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
		mockInvRepo.On("RemoveItemFromInventory", constants.TestCharacterID, constants.TestItemID, 2).Return(nil).Once()
		err := service.RemoveItemFromCharacter(constants.TestCharacterID, constants.TestItemID, 2)
		assert.NoError(t, err)
	})

	t.Run("GetCharacterInventory", func(t *testing.T) {
		expected := []*models.InventoryItem{
			mocks.CreateTestInventoryItem(constants.TestCharacterID, constants.TestItemID, 1, false, false, nil),
		}
		mockInvRepo.On("GetCharacterInventory", constants.TestCharacterID).Return(expected, nil).Once()

		inventory, err := service.GetCharacterInventory(constants.TestCharacterID)
		assert.NoError(t, err)
		assert.Equal(t, expected, inventory)
	})

	t.Run("UnequipItem", func(t *testing.T) {
		mockInvRepo.On("EquipItem", constants.TestCharacterID, constants.TestItemID, false).Return(nil).Once()
		err := service.UnequipItem(constants.TestCharacterID, constants.TestItemID)
		assert.NoError(t, err)
	})

	t.Run("UnattuneFromItem", func(t *testing.T) {
		mockInvRepo.On("UnattuneItem", constants.TestCharacterID, constants.TestItemID).Return(nil).Once()
		err := service.UnattuneFromItem(constants.TestCharacterID, constants.TestItemID)
		assert.NoError(t, err)
	})

	t.Run("GetCharacterCurrency", func(t *testing.T) {
		expected := &models.Currency{
			CharacterID: constants.TestCharacterID,
			Gold:        100,
		}
		mockInvRepo.On("GetCharacterCurrency", constants.TestCharacterID).Return(expected, nil).Once()

		currency, err := service.GetCharacterCurrency(constants.TestCharacterID)
		assert.NoError(t, err)
		assert.Equal(t, expected, currency)
	})

	mockInvRepo.AssertExpectations(t)
}
