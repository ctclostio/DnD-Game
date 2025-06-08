package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
)

// MockInventoryService is a mock implementation of the inventory service
type MockInventoryService struct {
	mock.Mock
}

func (m *MockInventoryService) AddItemToCharacter(characterID, itemID string, quantity int) error {
	args := m.Called(characterID, itemID, quantity)
	return args.Error(0)
}

func (m *MockInventoryService) RemoveItemFromCharacter(characterID, itemID string, quantity int) error {
	args := m.Called(characterID, itemID, quantity)
	return args.Error(0)
}

func (m *MockInventoryService) GetCharacterInventory(characterID string) ([]*models.InventoryItem, error) {
	args := m.Called(characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.InventoryItem), args.Error(1)
}

func (m *MockInventoryService) EquipItem(characterID, itemID string) error {
	args := m.Called(characterID, itemID)
	return args.Error(0)
}

func (m *MockInventoryService) UnequipItem(characterID, itemID string) error {
	args := m.Called(characterID, itemID)
	return args.Error(0)
}

func (m *MockInventoryService) AttuneToItem(characterID, itemID string) error {
	args := m.Called(characterID, itemID)
	return args.Error(0)
}

func (m *MockInventoryService) UnattuneFromItem(characterID, itemID string) error {
	args := m.Called(characterID, itemID)
	return args.Error(0)
}

func (m *MockInventoryService) GetCharacterCurrency(characterID string) (*models.Currency, error) {
	args := m.Called(characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Currency), args.Error(1)
}

func (m *MockInventoryService) UpdateCharacterCurrency(characterID string, copper, silver, electrum, gold, platinum int) error {
	args := m.Called(characterID, copper, silver, electrum, gold, platinum)
	return args.Error(0)
}

func (m *MockInventoryService) PurchaseItem(characterID, itemID string, quantity int) error {
	args := m.Called(characterID, itemID, quantity)
	return args.Error(0)
}

func (m *MockInventoryService) SellItem(characterID, itemID string, quantity int) error {
	args := m.Called(characterID, itemID, quantity)
	return args.Error(0)
}

func (m *MockInventoryService) GetCharacterWeight(characterID string) (*models.InventoryWeight, error) {
	args := m.Called(characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InventoryWeight), args.Error(1)
}

func (m *MockInventoryService) CreateItem(item *models.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockInventoryService) GetItemsByType(itemType models.ItemType) ([]*models.Item, error) {
	args := m.Called(itemType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Item), args.Error(1)
}

// Helper function to create test inventory handler
func createTestInventoryHandler() (*InventoryHandler, *MockInventoryService) {
	mockService := new(MockInventoryService)
	handler := NewInventoryHandler(&services.InventoryService{})
	// Inject the mock
	handler.inventoryService = mockService
	return handler, mockService
}

// Helper function to create test items
func createTestItem(id, name string, itemType models.ItemType) *models.Item {
	return &models.Item{
		ID:         id,
		Name:       name,
		Type:       itemType,
		Weight:     1.0,
		Value:      100,
		Properties: make(models.ItemProperties),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func TestInventoryHandler_GetCharacterInventory(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		setupMock      func(*MockInventoryService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "successful retrieval",
			characterID: "char-123",
			setupMock: func(m *MockInventoryService) {
				items := []*models.InventoryItem{
					{
						ID:          "inv-1",
						CharacterID: "char-123",
						ItemID:      "sword-1",
						Quantity:    1,
						Equipped:    true,
						Item:        createTestItem("sword-1", "Longsword", models.ItemTypeWeapon),
					},
					{
						ID:          "inv-2",
						CharacterID: "char-123",
						ItemID:      "potion-1",
						Quantity:    5,
						Equipped:    false,
						Item:        createTestItem("potion-1", "Healing Potion", models.ItemTypeConsumable),
					},
				}
				m.On("GetCharacterInventory", "char-123").Return(items, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var items []*models.InventoryItem
				err := json.Unmarshal(body, &items)
				require.NoError(t, err)
				assert.Len(t, items, 2)
				assert.Equal(t, "Longsword", items[0].Item.Name)
				assert.True(t, items[0].Equipped)
				assert.Equal(t, 5, items[1].Quantity)
			},
		},
		{
			name:        "empty inventory",
			characterID: "char-456",
			setupMock: func(m *MockInventoryService) {
				m.On("GetCharacterInventory", "char-456").Return([]*models.InventoryItem{}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var items []*models.InventoryItem
				err := json.Unmarshal(body, &items)
				require.NoError(t, err)
				assert.Empty(t, items)
			},
		},
		{
			name:        "service error",
			characterID: "char-789",
			setupMock: func(m *MockInventoryService) {
				m.On("GetCharacterInventory", "char-789").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestInventoryHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/characters/"+tt.characterID+"/inventory", nil)
			req = mux.SetURLVars(req, map[string]string{
				"characterId": tt.characterID,
			})

			rr := httptest.NewRecorder()
			handler.GetCharacterInventory(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInventoryHandler_AddItemToInventory(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		requestBody    interface{}
		setupMock      func(*MockInventoryService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "successful add",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"item_id":  "sword-456",
				"quantity": 2,
			},
			setupMock: func(m *MockInventoryService) {
				m.On("AddItemToCharacter", "char-123", "sword-456", 2).Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "success", response["status"])
			},
		},
		{
			name:        "default quantity to 1",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"item_id":  "potion-789",
				"quantity": 0,
			},
			setupMock: func(m *MockInventoryService) {
				m.On("AddItemToCharacter", "char-123", "potion-789", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "negative quantity defaults to 1",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"item_id":  "potion-789",
				"quantity": -5,
			},
			setupMock: func(m *MockInventoryService) {
				m.On("AddItemToCharacter", "char-123", "potion-789", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			characterID:    "char-123",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"item_id":  "item-999",
				"quantity": 1,
			},
			setupMock: func(m *MockInventoryService) {
				m.On("AddItemToCharacter", "char-123", "item-999", 1).Return(errors.New("item not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "item not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestInventoryHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/characters/"+tt.characterID+"/inventory", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{
				"characterId": tt.characterID,
			})

			rr := httptest.NewRecorder()
			handler.AddItemToInventory(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInventoryHandler_EquipItem(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		itemID         string
		setupMock      func(*MockInventoryService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "successful equip",
			characterID: "char-123",
			itemID:      "sword-456",
			setupMock: func(m *MockInventoryService) {
				m.On("EquipItem", "char-123", "sword-456").Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "equipped", response["status"])
			},
		},
		{
			name:        "equip error - not enough hands",
			characterID: "char-123",
			itemID:      "greatsword-789",
			setupMock: func(m *MockInventoryService) {
				m.On("EquipItem", "char-123", "greatsword-789").Return(errors.New("not enough hands to equip this weapon"))
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "not enough hands")
			},
		},
		{
			name:        "item not in inventory",
			characterID: "char-123",
			itemID:      "nonexistent",
			setupMock: func(m *MockInventoryService) {
				m.On("EquipItem", "char-123", "nonexistent").Return(errors.New("item not found in inventory"))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestInventoryHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/characters/"+tt.characterID+"/inventory/"+tt.itemID+"/equip", nil)
			req = mux.SetURLVars(req, map[string]string{
				"characterId": tt.characterID,
				"itemId":      tt.itemID,
			})

			rr := httptest.NewRecorder()
			handler.EquipItem(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInventoryHandler_AttuneItem(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		itemID         string
		setupMock      func(*MockInventoryService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "successful attunement",
			characterID: "char-123",
			itemID:      "ring-456",
			setupMock: func(m *MockInventoryService) {
				m.On("AttuneToItem", "char-123", "ring-456").Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "attuned", response["status"])
			},
		},
		{
			name:        "attunement error - doesn't require attunement",
			characterID: "char-123",
			itemID:      "sword-789",
			setupMock: func(m *MockInventoryService) {
				m.On("AttuneToItem", "char-123", "sword-789").Return(errors.New("item does not require attunement"))
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "does not require attunement")
			},
		},
		{
			name:        "already attuned",
			characterID: "char-123",
			itemID:      "ring-456",
			setupMock: func(m *MockInventoryService) {
				m.On("AttuneToItem", "char-123", "ring-456").Return(errors.New("already attuned to this item"))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestInventoryHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/characters/"+tt.characterID+"/inventory/"+tt.itemID+"/attune", nil)
			req = mux.SetURLVars(req, map[string]string{
				"characterId": tt.characterID,
				"itemId":      tt.itemID,
			})

			rr := httptest.NewRecorder()
			handler.AttuneItem(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInventoryHandler_UpdateCharacterCurrency(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		requestBody    interface{}
		setupMock      func(*MockInventoryService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "successful currency update",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"copper":   5,
				"silver":   3,
				"electrum": 0,
				"gold":     10,
				"platinum": 1,
			},
			setupMock: func(m *MockInventoryService) {
				m.On("UpdateCharacterCurrency", "char-123", 5, 3, 0, 10, 1).Return(nil)
				
				updatedCurrency := &models.Currency{
					CharacterID: "char-123",
					Copper:      15,
					Silver:      8,
					Gold:        12,
					Platinum:    1,
				}
				m.On("GetCharacterCurrency", "char-123").Return(updatedCurrency, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var currency models.Currency
				err := json.Unmarshal(body, &currency)
				require.NoError(t, err)
				assert.Equal(t, 15, currency.Copper)
				assert.Equal(t, 8, currency.Silver)
				assert.Equal(t, 12, currency.Gold)
			},
		},
		{
			name:           "invalid request body",
			characterID:    "char-123",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "insufficient funds error",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"gold": -100,
			},
			setupMock: func(m *MockInventoryService) {
				m.On("UpdateCharacterCurrency", "char-123", 0, 0, 0, -100, 0).Return(errors.New("insufficient funds"))
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "insufficient funds")
			},
		},
		{
			name:        "get currency error after update",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"gold": 10,
			},
			setupMock: func(m *MockInventoryService) {
				m.On("UpdateCharacterCurrency", "char-123", 0, 0, 0, 10, 0).Return(nil)
				m.On("GetCharacterCurrency", "char-123").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestInventoryHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/characters/"+tt.characterID+"/currency", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{
				"characterId": tt.characterID,
			})

			rr := httptest.NewRecorder()
			handler.UpdateCharacterCurrency(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInventoryHandler_PurchaseItem(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		requestBody    interface{}
		setupMock      func(*MockInventoryService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "successful purchase",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"item_id":  "potion-456",
				"quantity": 3,
			},
			setupMock: func(m *MockInventoryService) {
				m.On("PurchaseItem", "char-123", "potion-456", 3).Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "purchased", response["status"])
			},
		},
		{
			name:        "insufficient funds",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"item_id":  "expensive-item",
				"quantity": 1,
			},
			setupMock: func(m *MockInventoryService) {
				m.On("PurchaseItem", "char-123", "expensive-item", 1).Return(errors.New("insufficient funds"))
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "insufficient funds")
			},
		},
		{
			name:        "item not found",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"item_id": "nonexistent",
			},
			setupMock: func(m *MockInventoryService) {
				m.On("PurchaseItem", "char-123", "nonexistent", 1).Return(errors.New("item not found"))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestInventoryHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/characters/"+tt.characterID+"/purchase", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{
				"characterId": tt.characterID,
			})

			rr := httptest.NewRecorder()
			handler.PurchaseItem(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInventoryHandler_SellItem(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		requestBody    interface{}
		setupMock      func(*MockInventoryService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "successful sale",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"item_id":  "sword-456",
				"quantity": 1,
			},
			setupMock: func(m *MockInventoryService) {
				m.On("SellItem", "char-123", "sword-456", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "sold", response["status"])
			},
		},
		{
			name:        "item not in inventory",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"item_id":  "nonexistent",
				"quantity": 1,
			},
			setupMock: func(m *MockInventoryService) {
				m.On("SellItem", "char-123", "nonexistent", 1).Return(errors.New("item not in inventory"))
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "item not in inventory")
			},
		},
		{
			name:        "invalid quantity",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"item_id":  "item-789",
				"quantity": -5,
			},
			setupMock: func(m *MockInventoryService) {
				// Negative quantity defaults to 1
				m.On("SellItem", "char-123", "item-789", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestInventoryHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/characters/"+tt.characterID+"/sell", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{
				"characterId": tt.characterID,
			})

			rr := httptest.NewRecorder()
			handler.SellItem(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInventoryHandler_GetCharacterWeight(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		setupMock      func(*MockInventoryService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "not encumbered",
			characterID: "char-123",
			setupMock: func(m *MockInventoryService) {
				weight := &models.InventoryWeight{
					CurrentWeight:     75.5,
					CarryCapacity:     150.0,
					Encumbered:        false,
					HeavilyEncumbered: false,
				}
				m.On("GetCharacterWeight", "char-123").Return(weight, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var weight models.InventoryWeight
				err := json.Unmarshal(body, &weight)
				require.NoError(t, err)
				assert.Equal(t, 75.5, weight.CurrentWeight)
				assert.Equal(t, 150.0, weight.CarryCapacity)
				assert.False(t, weight.Encumbered)
			},
		},
		{
			name:        "heavily encumbered",
			characterID: "char-456",
			setupMock: func(m *MockInventoryService) {
				weight := &models.InventoryWeight{
					CurrentWeight:     350.0,
					CarryCapacity:     150.0,
					Encumbered:        true,
					HeavilyEncumbered: true,
				}
				m.On("GetCharacterWeight", "char-456").Return(weight, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var weight models.InventoryWeight
				err := json.Unmarshal(body, &weight)
				require.NoError(t, err)
				assert.True(t, weight.Encumbered)
				assert.True(t, weight.HeavilyEncumbered)
			},
		},
		{
			name:        "service error",
			characterID: "char-789",
			setupMock: func(m *MockInventoryService) {
				m.On("GetCharacterWeight", "char-789").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestInventoryHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/characters/"+tt.characterID+"/weight", nil)
			req = mux.SetURLVars(req, map[string]string{
				"characterId": tt.characterID,
			})

			rr := httptest.NewRecorder()
			handler.GetCharacterWeight(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInventoryHandler_CreateItem(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockInventoryService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name: "successful item creation",
			requestBody: map[string]interface{}{
				"name":                "Flaming Sword",
				"type":                "weapon",
				"rarity":              "rare",
				"weight":              3.5,
				"value":               5000,
				"requires_attunement": true,
				"description":         "A sword wreathed in magical flames",
			},
			setupMock: func(m *MockInventoryService) {
				m.On("CreateItem", mock.MatchedBy(func(item *models.Item) bool {
					return item.Name == "Flaming Sword" &&
						item.Type == models.ItemTypeWeapon &&
						item.Rarity == models.ItemRarityRare &&
						item.RequiresAttunement == true
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var item models.Item
				err := json.Unmarshal(body, &item)
				require.NoError(t, err)
				assert.Equal(t, "Flaming Sword", item.Name)
				assert.Equal(t, models.ItemRarityRare, item.Rarity)
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: map[string]interface{}{
				"name": "Test Item",
				"type": "weapon",
			},
			setupMock: func(m *MockInventoryService) {
				m.On("CreateItem", mock.Anything).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestInventoryHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/items", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.CreateItem(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInventoryHandler_GetItemsByType(t *testing.T) {
	tests := []struct {
		name           string
		itemType       string
		setupMock      func(*MockInventoryService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:     "get weapons",
			itemType: "weapon",
			setupMock: func(m *MockInventoryService) {
				weapons := []*models.Item{
					createTestItem("sword-1", "Longsword", models.ItemTypeWeapon),
					createTestItem("axe-1", "Battleaxe", models.ItemTypeWeapon),
					createTestItem("bow-1", "Longbow", models.ItemTypeWeapon),
				}
				m.On("GetItemsByType", models.ItemTypeWeapon).Return(weapons, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var items []*models.Item
				err := json.Unmarshal(body, &items)
				require.NoError(t, err)
				assert.Len(t, items, 3)
				assert.Equal(t, "Longsword", items[0].Name)
				assert.Equal(t, models.ItemTypeWeapon, items[0].Type)
			},
		},
		{
			name:     "get consumables - empty result",
			itemType: "consumable",
			setupMock: func(m *MockInventoryService) {
				m.On("GetItemsByType", models.ItemTypeConsumable).Return([]*models.Item{}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var items []*models.Item
				err := json.Unmarshal(body, &items)
				require.NoError(t, err)
				assert.Empty(t, items)
			},
		},
		{
			name:           "missing type parameter",
			itemType:       "",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "type parameter required")
			},
		},
		{
			name:     "service error",
			itemType: "magic",
			setupMock: func(m *MockInventoryService) {
				m.On("GetItemsByType", models.ItemTypeMagic).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestInventoryHandler()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			url := "/api/items"
			if tt.itemType != "" {
				url += "?type=" + tt.itemType
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			rr := httptest.NewRecorder()
			handler.GetItemsByType(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}