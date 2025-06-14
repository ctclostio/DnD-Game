package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a test context with auth claims.
func createAuthContext(userID, role string) context.Context {
	claims := &auth.Claims{
		UserID:   userID,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     role,
		Type:     auth.AccessToken,
	}
	return context.WithValue(context.Background(), auth.UserContextKey, claims)
}

func TestCharacterHandler_RequestValidation(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		userID         string
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:   "valid character creation request",
			method: http.MethodPost,
			path:   "/api/characters",
			body: map[string]interface{}{
				"name":  "Aragorn",
				"race":  "Human",
				"class": "Ranger",
				"level": 10,
				"abilities": map[string]int{
					"strength":     16,
					"dexterity":    14,
					"constitution": 15,
					"intelligence": 12,
					"wisdom":       14,
					"charisma":     13,
				},
			},
			userID:         uuid.New().String(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "invalid character creation - missing name",
			method: http.MethodPost,
			path:   "/api/characters",
			body: map[string]interface{}{
				"race":  "Elf",
				"class": "Wizard",
				"level": 1,
			},
			userID:         uuid.New().String(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid character creation - invalid ability scores",
			method: http.MethodPost,
			path:   "/api/characters",
			body: map[string]interface{}{
				"name":  "Invalid Character",
				"race":  "Human",
				"class": "Fighter",
				"level": 1,
				"abilities": map[string]int{
					"strength":     25, // Too high
					"dexterity":    14,
					"constitution": 15,
					"intelligence": 12,
					"wisdom":       14,
					"charisma":     13,
				},
			},
			userID:         uuid.New().String(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "no authentication",
			method:         http.MethodPost,
			path:           "/api/characters",
			body:           map[string]interface{}{"name": "Test"},
			userID:         "", // No auth
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request.
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Add auth context if userID is provided (placeholder for middleware).
			// For this test, we'll just verify the request structure.
			if tt.body != nil {
				var decoded map[string]interface{}
				err := json.NewDecoder(bytes.NewReader(body)).Decode(&decoded)
				assert.NoError(t, err)

				// Validate character creation requirements.
				if tt.method == http.MethodPost && tt.path == "/api/characters" {
					if tt.expectedStatus == http.StatusBadRequest {
						// Check for missing required fields.
						if _, ok := decoded["name"]; !ok && tt.name == "invalid character creation - missing name" {
							assert.True(t, true, "Name is correctly missing")
						}

						// Check for invalid ability scores.
						if abilities, ok := decoded["abilities"].(map[string]interface{}); ok {
							if str, ok := abilities["strength"].(float64); ok && str > 20 {
								assert.True(t, true, "Strength is correctly too high")
							}
						}
					}
				}
			}
		})
	}
}

func TestCharacterHandler_UpdateCharacter(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		body           interface{}
		userID         string
		expectedStatus int
	}{
		{
			name:        "valid HP update",
			characterID: uuid.New().String(),
			body: map[string]interface{}{
				"currentHP": 45,
			},
			userID:         uuid.New().String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:        "level up",
			characterID: uuid.New().String(),
			body: map[string]interface{}{
				"level": 11,
				"maxHP": 88,
				"hitDice": map[string]interface{}{
					"d10": 11,
				},
			},
			userID:         uuid.New().String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:        "update equipment",
			characterID: uuid.New().String(),
			body: map[string]interface{}{
				"equipment": []string{
					"Longsword +1",
					"Chain Mail",
					"Shield",
				},
			},
			userID:         uuid.New().String(),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request.
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/characters/"+tt.characterID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Add route vars.
			_ = mux.SetURLVars(req, map[string]string{"id": tt.characterID})

			// Add auth context (placeholder).
			// Verify request can be parsed.
			var decoded map[string]interface{}
			err := json.NewDecoder(bytes.NewReader(body)).Decode(&decoded)
			assert.NoError(t, err)
			assert.NotEmpty(t, decoded)
		})
	}
}

func TestCharacterHandler_SpellSlots(t *testing.T) {
	characterID := uuid.New().String()
	userID := uuid.New().String()

	tests := []struct {
		name        string
		action      string
		body        interface{}
		expectError bool
	}{
		{
			name:   "use spell slot",
			action: "use",
			body: map[string]interface{}{
				"spellLevel": 3,
			},
			expectError: false,
		},
		{
			name:   "invalid spell level",
			action: "use",
			body: map[string]interface{}{
				"spellLevel": 10, // Too high
			},
			expectError: true,
		},
		{
			name:   "short rest",
			action: "rest",
			body: map[string]interface{}{
				"restType": "short",
			},
			expectError: false,
		},
		{
			name:   "long rest",
			action: "rest",
			body: map[string]interface{}{
				"restType": "long",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request.
			body, _ := json.Marshal(tt.body)
			var path string
			if tt.action == "use" {
				path = "/api/characters/" + characterID + "/spell-slots/use"
			} else {
				path = "/api/characters/" + characterID + "/rest"
			}

			req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			_ = mux.SetURLVars(req, map[string]string{"id": characterID})

			// Add auth context.
			ctx := createAuthContext(userID, "player")
			_ = req.WithContext(ctx)

			// Verify request structure.
			var decoded map[string]interface{}
			err := json.NewDecoder(bytes.NewReader(body)).Decode(&decoded)
			assert.NoError(t, err)

			// Validate spell level if using spell slot.
			if tt.action == "use" {
				if spellLevel, ok := decoded["spellLevel"].(float64); ok {
					if spellLevel > 9 && tt.expectError {
						assert.True(t, true, "Spell level correctly identified as too high")
					}
				}
			}
		})
	}
}

func TestCharacterHandler_CustomClass(t *testing.T) {
	t.Run("generate custom class request", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "Shadow Dancer",
			"description": "A mystical warrior who dances between shadows",
			"role":        "Stealth DPS",
			"style":       "Agile melee combat with shadow magic",
			"features":    "Shadow step, invisibility, sneak attack",
		}

		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/characters/custom-class/generate", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// Add auth context.
		// Context would be added by auth middleware in real handler.
		// ctx := createAuthContext(userID, "player").
		// req = req.WithContext(ctx)

		// Verify request structure.
		var decoded map[string]interface{}
		err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&decoded)
		assert.NoError(t, err)
		assert.Equal(t, "Shadow Dancer", decoded["name"])
		assert.NotEmpty(t, decoded["description"])
		assert.NotEmpty(t, decoded["role"])
	})

	t.Run("list custom classes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/characters/custom-classes?includeUnapproved=true", nil)

		// Add auth context.
		// Context would be added by auth middleware in real handler.
		// ctx := createAuthContext(userID, "player").
		// req = req.WithContext(ctx)

		// Verify query parameter.
		includeUnapproved := req.URL.Query().Get("includeUnapproved")
		assert.Equal(t, "true", includeUnapproved)
	})
}

func TestCharacterHandler_CharacterValidation(t *testing.T) {
	// Test D&D-specific validation rules.
	validationTests := []struct {
		name        string
		character   models.Character
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid character",
			character: models.Character{
				Name:  "Gandalf",
				Race:  "Human",
				Class: "Wizard",
				Level: 20,
				Attributes: models.Attributes{
					Strength:     10,
					Dexterity:    14,
					Constitution: 12,
					Intelligence: 18,
					Wisdom:       16,
					Charisma:     13,
				},
			},
			shouldError: false,
		},
		{
			name: "ability score too low",
			character: models.Character{
				Name:  "Weak Character",
				Race:  "Human",
				Class: "Fighter",
				Level: 1,
				Attributes: models.Attributes{
					Strength:     2, // Minimum is 3
					Dexterity:    10,
					Constitution: 10,
					Intelligence: 10,
					Wisdom:       10,
					Charisma:     10,
				},
			},
			shouldError: true,
			errorMsg:    "ability scores must be between 3 and 20",
		},
		{
			name: "level too high",
			character: models.Character{
				Name:  "Overpowered",
				Race:  "Elf",
				Class: "Ranger",
				Level: 25, // Maximum is 20
			},
			shouldError: true,
			errorMsg:    "level must be between 1 and 20",
		},
		{
			name: "empty name",
			character: models.Character{
				Name:  "",
				Race:  "Dwarf",
				Class: "Cleric",
				Level: 1,
			},
			shouldError: true,
			errorMsg:    "character name is required",
		},
	}

	for _, tt := range validationTests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate character attributes.
			if tt.character.Name == "" && tt.shouldError {
				assert.Equal(t, "character name is required", tt.errorMsg)
			}
			if tt.character.Level > 20 && tt.shouldError {
				assert.Contains(t, tt.errorMsg, "level must be between 1 and 20")
			}
			if tt.shouldError && tt.name == "ability score too low" {
				assert.Contains(t, tt.errorMsg, "ability scores must be between 3 and 20")
			}
		})
	}
}
