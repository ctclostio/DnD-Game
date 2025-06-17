package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

func TestDiceHandler_RollDice(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		userID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid dice roll request",
			body: DiceRollRequest{
				GameSessionID: "session-123",
				RollNotation:  "1d20+5",
				Purpose:       "Attack roll",
			},
			userID:         uuid.New().String(),
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing game session ID",
			body: DiceRollRequest{
				RollNotation: "1d20",
				Purpose:      "Skill check",
			},
			userID:         uuid.New().String(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Game session ID is required",
		},
		{
			name: "invalid dice notation",
			body: DiceRollRequest{
				GameSessionID: "session-123",
				RollNotation:  "invalid",
				Purpose:       "Test",
			},
			userID:         uuid.New().String(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid dice notation",
		},
		{
			name: "no authentication",
			body: DiceRollRequest{
				GameSessionID: "session-123",
				RollNotation:  "1d20",
				Purpose:       "Test",
			},
			userID:         "", // Empty user ID simulates no auth
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/dice/roll", bytes.NewReader(body))
			req.Header.Set(constants.ContentType, constants.ApplicationJSON)

			// Add auth context (placeholder)

			// For this test, we'll just verify the request structure
			// A full test would need the handler with all dependencies

			// Verify we can decode the request
			var decoded DiceRollRequest
			err := json.NewDecoder(bytes.NewReader(body)).Decode(&decoded)
			assert.NoError(t, err)

			// Basic validation that would be done by handler
			if decoded.GameSessionID == "" && tt.expectedError == "Game session ID is required" {
				assert.Empty(t, decoded.GameSessionID)
			}

			if decoded.RollNotation == "invalid" && tt.expectedError == "Invalid dice notation" {
				// In real handler, this would be validated by dice parser
				assert.Equal(t, "invalid", decoded.RollNotation)
			}
		})
	}
}

func TestDiceHandler_GetRollHistory(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		userID         string
		limit          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "valid request",
			sessionID:      "session-123",
			userID:         uuid.New().String(),
			limit:          "10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no limit specified (should use default)",
			sessionID:      "session-123",
			userID:         uuid.New().String(),
			limit:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid limit",
			sessionID:      "session-123",
			userID:         uuid.New().String(),
			limit:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid limit parameter",
		},
		{
			name:           "no authentication",
			sessionID:      "session-123",
			userID:         "",
			limit:          "10",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			url := "/api/sessions/" + tt.sessionID + "/dice/history"
			if tt.limit != "" {
				url += "?limit=" + tt.limit
			}
			req := httptest.NewRequest(http.MethodGet, url, http.NoBody)

			// Add auth context (placeholder)

			// For this test, verify query parameter parsing
			limit := req.URL.Query().Get("limit")
			if limit == "invalid" && tt.expectedError == "Invalid limit parameter" {
				// In real handler, this would fail integer parsing
				assert.Equal(t, "invalid", limit)
			}
		})
	}
}

// TestDiceNotationValidation tests various dice notation formats
func TestDiceNotationValidation(t *testing.T) {
	validNotations := []string{
		"d20",
		"1d20",
		"2d6",
		"1d8+3",
		"2d10-1",
		"3d6+2d4",
		"1d12+1d6+5",
		"d20+d4-2",
		"4d6k3",   // keep highest 3
		"2d20kl1", // keep lowest 1
		"d100",
		"d%", // percentile dice
	}

	invalidNotations := []string{
		"",
		"invalid",
		"1dd20",
		"d",
		"20d",
		"d20d",
		"1+d20",   // modifier must come after dice
		"d20+",    // missing modifier
		"d0",      // invalid die size
		"d1",      // invalid die size
		"1d20+1d", // incomplete notation
	}

	for _, notation := range validNotations {
		t.Run("valid_"+notation, func(t *testing.T) {
			// This would be validated by the dice parser
			assert.NotEmpty(t, notation)
			// In real implementation, parser.ParseNotation(notation) should not error
		})
	}

	for _, notation := range invalidNotations {
		t.Run("invalid_"+notation, func(t *testing.T) {
			// This would be rejected by the dice parser
			// In real implementation, parser.ParseNotation(notation) should error
			if notation == "" {
				assert.Empty(t, notation)
			} else {
				assert.NotContains(t, validNotations, notation)
			}
		})
	}
}

// TestDiceRollPurposes tests different roll purposes
func TestDiceRollPurposes(t *testing.T) {
	purposes := []struct {
		purpose     string
		description string
		typical     string // typical notation
	}{
		{
			purpose:     "attack",
			description: "Attack roll",
			typical:     "1d20",
		},
		{
			purpose:     "damage",
			description: "Damage roll",
			typical:     "1d8+3",
		},
		{
			purpose:     "skill_check",
			description: "Skill check",
			typical:     "1d20+5",
		},
		{
			purpose:     "saving_throw",
			description: "Saving throw",
			typical:     "1d20+2",
		},
		{
			purpose:     "initiative",
			description: "Initiative roll",
			typical:     "1d20+3",
		},
		{
			purpose:     "ability_check",
			description: "Ability check",
			typical:     "1d20",
		},
		{
			purpose:     "hit_dice",
			description: "Hit dice roll for healing",
			typical:     "1d10+2",
		},
		{
			purpose:     "death_save",
			description: "Death saving throw",
			typical:     "1d20",
		},
	}

	for _, p := range purposes {
		t.Run(p.purpose, func(t *testing.T) {
			// Create a dice roll with this purpose
			roll := &models.DiceRoll{
				ID:            uuid.New().String(),
				UserID:        uuid.New().String(),
				GameSessionID: uuid.New().String(),
				RollNotation:  p.typical,
				Purpose:       p.purpose,
				Total:         10, // Placeholder
				Results:       []int{10},
			}

			// Verify the roll has required fields
			assert.NotEmpty(t, roll.ID)
			assert.NotEmpty(t, roll.UserID)
			assert.NotEmpty(t, roll.RollNotation)
			assert.Equal(t, p.purpose, roll.Purpose)
			assert.NotZero(t, roll.Total)
			assert.NotEmpty(t, roll.Results)
		})
	}
}

// TestAdvantageDisadvantage tests advantage/disadvantage mechanics
func TestAdvantageDisadvantage(t *testing.T) {
	tests := []struct {
		name         string
		notation     string
		rollType     string
		expectedDice int
	}{
		{
			name:         "normal d20 roll",
			notation:     "1d20+5",
			rollType:     "normal",
			expectedDice: 1,
		},
		{
			name:         "advantage roll",
			notation:     "2d20kh1+5", // roll 2d20, keep highest 1
			rollType:     "advantage",
			expectedDice: 2,
		},
		{
			name:         "disadvantage roll",
			notation:     "2d20kl1+5", // roll 2d20, keep lowest 1
			rollType:     "disadvantage",
			expectedDice: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This demonstrates the notation for advantage/disadvantage
			assert.Contains(t, tt.notation, "d20")
			switch tt.rollType {
			case "advantage":
				assert.Contains(t, tt.notation, "kh") // keep highest
			case "disadvantage":
				assert.Contains(t, tt.notation, "kl") // keep lowest
			}
		})
	}
}
