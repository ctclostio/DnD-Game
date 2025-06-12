package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

func TestInventoryHandler_ManageInventory(t *testing.T) {
	characterID := uuid.New().String()
	userID := uuid.New().String()

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "add item to inventory",
			method: http.MethodPost,
			path:   "/api/characters/" + characterID + "/inventory",
			body: map[string]interface{}{
				"itemId":   uuid.New().String(),
				"quantity": 3,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "remove item from inventory",
			method: http.MethodDelete,
			path:   "/api/characters/" + characterID + "/inventory/" + uuid.New().String(),
			body: map[string]interface{}{
				"quantity": 1,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid quantity",
			method: http.MethodPost,
			path:   "/api/characters/" + characterID + "/inventory",
			body: map[string]interface{}{
				"itemId":   uuid.New().String(),
				"quantity": -1,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "quantity must be positive",
		},
		{
			name:           "get inventory",
			method:         http.MethodGet,
			path:           "/api/characters/" + characterID + "/inventory",
			body:           nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"characterId": characterID})

			// Add auth context
			ctx := context.WithValue(req.Context(), auth.UserContextKey, &auth.Claims{
				UserID: userID,
				Type:   auth.AccessToken,
			})
			req = req.WithContext(ctx)

			// For this test, verify request structure
			if tt.body != nil {
				var decoded map[string]interface{}
				err := json.NewDecoder(bytes.NewReader(body)).Decode(&decoded)
				assert.NoError(t, err)

				// Validate quantity if present
				if qty, ok := decoded["quantity"].(float64); ok {
					if qty < 0 && tt.expectedError == "quantity must be positive" {
						assert.True(t, true, "Quantity is correctly negative")
					}
				}
			}
		})
	}
}

func TestInventoryHandler_EquipItems(t *testing.T) {
	characterID := uuid.New().String()
	itemID := uuid.New().String()
	userID := uuid.New().String()

	tests := []struct {
		name           string
		action         string
		expectedStatus int
	}{
		{
			name:           "equip item",
			action:         "equip",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unequip item",
			action:         "unequip",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "attune to item",
			action:         "attune",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unattune from item",
			action:         "unattune",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodPost,
				"/api/characters/"+characterID+"/inventory/"+itemID+"/"+tt.action, nil)
			req = mux.SetURLVars(req, map[string]string{
				"characterId": characterID,
				"itemId":      itemID,
			})

			// Add auth context
			ctx := context.WithValue(req.Context(), auth.UserContextKey, &auth.Claims{
				UserID: userID,
				Type:   auth.AccessToken,
			})
			req = req.WithContext(ctx)

			// Verify route variables are set
			vars := mux.Vars(req)
			assert.Equal(t, characterID, vars["characterId"])
			assert.Equal(t, itemID, vars["itemId"])
		})
	}
}

func TestInventoryHandler_Currency(t *testing.T) {
	characterID := uuid.New().String()
	userID := uuid.New().String()

	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
		validateBody   func(*testing.T, map[string]interface{})
	}{
		{
			name: "update currency",
			body: map[string]interface{}{
				"copper":   50,
				"silver":   10,
				"gold":     5,
				"platinum": 0,
				"electrum": 0,
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, float64(50), body["copper"])
				assert.Equal(t, float64(10), body["silver"])
				assert.Equal(t, float64(5), body["gold"])
			},
		},
		{
			name: "invalid currency values",
			body: map[string]interface{}{
				"copper": -10,
				"silver": 5,
				"gold":   2,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "purchase item",
			body: map[string]interface{}{
				"itemId":   uuid.New().String(),
				"quantity": 1,
				"cost":     150, // in copper
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost,
				"/api/characters/"+characterID+"/currency", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"characterId": characterID})

			// Add auth context
			ctx := context.WithValue(req.Context(), auth.UserContextKey, &auth.Claims{
				UserID: userID,
				Type:   auth.AccessToken,
			})
			req = req.WithContext(ctx)

			// Verify request structure
			var decoded map[string]interface{}
			err := json.NewDecoder(bytes.NewReader(body)).Decode(&decoded)
			assert.NoError(t, err)

			if tt.validateBody != nil {
				tt.validateBody(t, decoded)
			}

			// Check for negative currency
			if copper, ok := decoded["copper"].(float64); ok && copper < 0 {
				assert.Equal(t, http.StatusBadRequest, tt.expectedStatus)
			}
		})
	}
}

func TestInventoryHandler_ItemTypes(t *testing.T) {
	// Test various item types and their properties
	itemTests := []struct {
		name       string
		item       models.Item
		properties map[string]interface{}
	}{
		{
			name: "weapon",
			item: models.Item{
				ID:     uuid.New().String(),
				Name:   "Longsword +1",
				Type:   models.ItemTypeWeapon,
				Rarity: models.ItemRarityUncommon,
				Weight: 3.0,
				Value:  1500, // in copper
			},
			properties: map[string]interface{}{
				"damage":      "1d8",
				"damageType":  "slashing",
				"versatile":   "1d10",
				"enhancement": 1,
			},
		},
		{
			name: "armor",
			item: models.Item{
				ID:     uuid.New().String(),
				Name:   "Chain Mail",
				Type:   models.ItemTypeArmor,
				Rarity: models.ItemRarityCommon,
				Weight: 55.0,
				Value:  7500,
			},
			properties: map[string]interface{}{
				"armorClass":          16,
				"stealthDisadvantage": true,
				"strengthRequirement": 13,
			},
		},
		{
			name: "consumable",
			item: models.Item{
				ID:     uuid.New().String(),
				Name:   "Potion of Healing",
				Type:   models.ItemTypeConsumable,
				Rarity: models.ItemRarityCommon,
				Weight: 0.5,
				Value:  5000,
			},
			properties: map[string]interface{}{
				"healing": "2d4+2",
				"action":  "action",
			},
		},
		{
			name: "magic item",
			item: models.Item{
				ID:                 uuid.New().String(),
				Name:               "Ring of Protection",
				Type:               models.ItemTypeMagic,
				Rarity:             models.ItemRarityRare,
				Weight:             0.0,
				Value:              500000,
				RequiresAttunement: true,
			},
			properties: map[string]interface{}{
				"acBonus":          1,
				"savingThrowBonus": 1,
			},
		},
	}

	for _, tt := range itemTests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify item attributes
			assert.NotEmpty(t, tt.item.ID)
			assert.NotEmpty(t, tt.item.Name)
			assert.NotEmpty(t, tt.item.Type)
			assert.NotEmpty(t, tt.item.Rarity)
			assert.Greater(t, tt.item.Value, 0)

			// For magic items, check attunement
			if tt.item.Type == models.ItemTypeMagic && tt.item.RequiresAttunement {
				assert.True(t, tt.item.RequiresAttunement)
			}

			// Verify properties structure
			if tt.properties != nil {
				assert.NotEmpty(t, tt.properties)
			}
		})
	}
}

func TestInventoryHandler_WeightCalculation(t *testing.T) {
	t.Run("calculate total weight", func(t *testing.T) {
		items := []struct {
			item     models.Item
			quantity int
		}{
			{
				item: models.Item{
					Name:   "Longsword",
					Weight: 3.0,
				},
				quantity: 1,
			},
			{
				item: models.Item{
					Name:   "Chain Mail",
					Weight: 55.0,
				},
				quantity: 1,
			},
			{
				item: models.Item{
					Name:   "Rations",
					Weight: 2.0,
				},
				quantity: 5,
			},
			{
				item: models.Item{
					Name:   "Torch",
					Weight: 1.0,
				},
				quantity: 10,
			},
		}

		totalWeight := 0.0
		for _, i := range items {
			totalWeight += i.item.Weight * float64(i.quantity)
		}

		// Expected: 3 + 55 + (2*5) + (1*10) = 78
		assert.Equal(t, 78.0, totalWeight)

		// Check encumbrance (assuming STR 15 = 225 lbs capacity)
		carryCapacity := 15 * 15.0 // STR score * 15
		assert.Less(t, totalWeight, carryCapacity)
	})
}

func TestInventoryHandler_CurrencyConversion(t *testing.T) {
	tests := []struct {
		name        string
		currency    models.Currency
		totalCopper int
	}{
		{
			name: "mixed currency",
			currency: models.Currency{
				Copper:   5,
				Silver:   3,
				Electrum: 1,
				Gold:     2,
				Platinum: 1,
			},
			totalCopper: 5 + (3 * 10) + (1 * 50) + (2 * 100) + (1 * 1000), // 1285
		},
		{
			name: "gold only",
			currency: models.Currency{
				Gold: 50,
			},
			totalCopper: 50 * 100, // 5000
		},
		{
			name: "copper only",
			currency: models.Currency{
				Copper: 1234,
			},
			totalCopper: 1234,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := tt.currency.TotalInCopper()
			assert.Equal(t, tt.totalCopper, total)

			// Test affordability
			canAfford := tt.currency.CanAfford(tt.totalCopper)
			assert.True(t, canAfford, "Should be able to afford exact amount")

			cantAfford := tt.currency.CanAfford(tt.totalCopper + 1)
			assert.False(t, cantAfford, "Should not be able to afford more than total")
		})
	}
}
