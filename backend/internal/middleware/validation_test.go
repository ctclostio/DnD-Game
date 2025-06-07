package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

func TestValidationMiddleware_CharacterCreation(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid character creation",
			body: map[string]interface{}{
				"name":  "Aragorn",
				"race":  "Human",
				"class": "Fighter",
				"abilities": map[string]int{
					"strength":     15,
					"dexterity":    13,
					"constitution": 14,
					"intelligence": 10,
					"wisdom":       12,
					"charisma":     8,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing required field",
			body: map[string]interface{}{
				"race":  "Human",
				"class": "Fighter",
				// Missing name
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
		{
			name: "invalid ability score too high",
			body: map[string]interface{}{
				"name":  "Invalid",
				"race":  "Human",
				"class": "Fighter",
				"abilities": map[string]int{
					"strength":     25, // Too high
					"dexterity":    13,
					"constitution": 14,
					"intelligence": 10,
					"wisdom":       12,
					"charisma":     8,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ability score must be between 3 and 18",
		},
		{
			name: "invalid ability score too low",
			body: map[string]interface{}{
				"name":  "Invalid",
				"race":  "Human",
				"class": "Fighter",
				"abilities": map[string]int{
					"strength":     15,
					"dexterity":    13,
					"constitution": 14,
					"intelligence": 10,
					"wisdom":       12,
					"charisma":     2, // Too low
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ability score must be between 3 and 18",
		},
		{
			name: "invalid race",
			body: map[string]interface{}{
				"name":  "Invalid",
				"race":  "Dragon", // Not a valid player race
				"class": "Fighter",
				"abilities": map[string]int{
					"strength":     15,
					"dexterity":    13,
					"constitution": 14,
					"intelligence": 10,
					"wisdom":       12,
					"charisma":     8,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid race",
		},
		{
			name: "invalid class",
			body: map[string]interface{}{
				"name":  "Invalid",
				"race":  "Human",
				"class": "Necromancer", // Not a standard class
				"abilities": map[string]int{
					"strength":     15,
					"dexterity":    13,
					"constitution": 14,
					"intelligence": 10,
					"wisdom":       12,
					"charisma":     8,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid class",
		},
		{
			name: "character name too long",
			body: map[string]interface{}{
				"name":  "ThisIsAnExtremelyLongCharacterNameThatExceedsTheMaximumAllowedLength",
				"race":  "Human",
				"class": "Fighter",
				"abilities": map[string]int{
					"strength":     15,
					"dexterity":    13,
					"constitution": 14,
					"intelligence": 10,
					"wisdom":       12,
					"charisma":     8,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name must be between 2 and 50 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			
			// Add character creation validation middleware
			router.POST("/characters", ValidateCharacterCreation(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})
			
			client := testutil.NewHTTPTestClient(t).SetRouter(router)
			resp := client.POST("/characters", tt.body)
			
			resp.AssertStatus(tt.expectedStatus)
			
			if tt.expectedError != "" {
				resp.AssertBodyContains(tt.expectedError)
			}
		})
	}
}

func TestValidationMiddleware_DiceRoll(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid dice notation",
			body: map[string]interface{}{
				"notation": "1d20+5",
				"purpose":  "Attack roll",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "complex valid notation",
			body: map[string]interface{}{
				"notation": "2d6+1d4+3",
				"purpose":  "Damage roll",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing notation",
			body: map[string]interface{}{
				"purpose": "Attack roll",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "notation is required",
		},
		{
			name: "invalid notation format",
			body: map[string]interface{}{
				"notation": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid dice notation",
		},
		{
			name: "too many dice",
			body: map[string]interface{}{
				"notation": "100d20", // Too many dice
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "maximum 20 dice allowed",
		},
		{
			name: "invalid die type",
			body: map[string]interface{}{
				"notation": "1d7", // d7 doesn't exist
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid die type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			
			router.POST("/dice/roll", ValidateDiceRoll(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})
			
			client := testutil.NewHTTPTestClient(t).SetRouter(router)
			resp := client.POST("/dice/roll", tt.body)
			
			resp.AssertStatus(tt.expectedStatus)
			
			if tt.expectedError != "" {
				resp.AssertBodyContains(tt.expectedError)
			}
		})
	}
}

func TestValidationMiddleware_CombatAction(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid attack action",
			body: map[string]interface{}{
				"action_type": "attack",
				"actor_id":    "char-1",
				"target_id":   "npc-1",
				"dice_roll":   18,
				"damage":      12,
				"damage_type": "slashing",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid healing action",
			body: map[string]interface{}{
				"action_type": "healing",
				"actor_id":    "char-2",
				"target_id":   "char-1",
				"healing":     8,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing action type",
			body: map[string]interface{}{
				"actor_id":  "char-1",
				"target_id": "npc-1",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "action_type is required",
		},
		{
			name: "invalid action type",
			body: map[string]interface{}{
				"action_type": "invalid",
				"actor_id":    "char-1",
				"target_id":   "npc-1",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid action type",
		},
		{
			name: "attack without damage",
			body: map[string]interface{}{
				"action_type": "attack",
				"actor_id":    "char-1",
				"target_id":   "npc-1",
				"dice_roll":   18,
				// Missing damage
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "damage is required for attack actions",
		},
		{
			name: "invalid damage type",
			body: map[string]interface{}{
				"action_type": "attack",
				"actor_id":    "char-1",
				"target_id":   "npc-1",
				"dice_roll":   18,
				"damage":      12,
				"damage_type": "nuclear", // Not a valid D&D damage type
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid damage type",
		},
		{
			name: "negative damage",
			body: map[string]interface{}{
				"action_type": "attack",
				"actor_id":    "char-1",
				"target_id":   "npc-1",
				"dice_roll":   18,
				"damage":      -5,
				"damage_type": "slashing",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "damage must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			
			router.POST("/combat/action", ValidateCombatAction(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})
			
			client := testutil.NewHTTPTestClient(t).SetRouter(router)
			resp := client.POST("/combat/action", tt.body)
			
			resp.AssertStatus(tt.expectedStatus)
			
			if tt.expectedError != "" {
				resp.AssertBodyContains(tt.expectedError)
			}
		})
	}
}

func TestValidationMiddleware_PaginationParams(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedError  string
		expectedPage   int
		expectedLimit  int
	}{
		{
			name:           "default pagination",
			query:          "",
			expectedStatus: http.StatusOK,
			expectedPage:   1,
			expectedLimit:  20,
		},
		{
			name:           "custom valid pagination",
			query:          "?page=2&limit=50",
			expectedStatus: http.StatusOK,
			expectedPage:   2,
			expectedLimit:  50,
		},
		{
			name:           "negative page",
			query:          "?page=-1",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "page must be positive",
		},
		{
			name:           "zero page",
			query:          "?page=0",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "page must be positive",
		},
		{
			name:           "limit too high",
			query:          "?limit=200",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "limit cannot exceed 100",
		},
		{
			name:           "negative limit",
			query:          "?limit=-10",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "limit must be positive",
		},
		{
			name:           "non-numeric page",
			query:          "?page=abc",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid page parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			
			router.GET("/items", ValidatePagination(), func(c *gin.Context) {
				page := c.GetInt("page")
				limit := c.GetInt("limit")
				c.JSON(http.StatusOK, gin.H{
					"page":  page,
					"limit": limit,
				})
			})
			
			req := httptest.NewRequest(http.MethodGet, "/items"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			require.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedError != "" {
				require.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, float64(tt.expectedPage), response["page"])
				require.Equal(t, float64(tt.expectedLimit), response["limit"])
			}
		})
	}
}

func TestValidationMiddleware_GameSessionParams(t *testing.T) {
	t.Run("valid session creation", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		
		router.POST("/sessions", ValidateGameSession(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "success"})
		})
		
		body := map[string]interface{}{
			"name":        "Epic Campaign",
			"max_players": 6,
			"description": "A grand adventure",
		}
		
		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.POST("/sessions", body)
		
		resp.AssertOK()
	})

	t.Run("session name too short", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		
		router.POST("/sessions", ValidateGameSession(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "success"})
		})
		
		body := map[string]interface{}{
			"name":        "A", // Too short
			"max_players": 6,
		}
		
		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.POST("/sessions", body)
		
		resp.AssertBadRequest()
		resp.AssertBodyContains("name must be at least 3 characters")
	})

	t.Run("invalid max players", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		
		router.POST("/sessions", ValidateGameSession(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "success"})
		})
		
		body := map[string]interface{}{
			"name":        "Epic Campaign",
			"max_players": 20, // Too many
		}
		
		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.POST("/sessions", body)
		
		resp.AssertBadRequest()
		resp.AssertBodyContains("max_players must be between 1 and 10")
	})
}

// Test custom validation rules
func TestValidationMiddleware_CustomRules(t *testing.T) {
	t.Run("spell slot validation", func(t *testing.T) {
		validator := NewSpellSlotValidator()
		
		tests := []struct {
			name    string
			slots   map[string]interface{}
			wantErr bool
		}{
			{
				name: "valid spell slots",
				slots: map[string]interface{}{
					"1": map[string]int{"total": 2, "used": 0},
					"2": map[string]int{"total": 1, "used": 1},
				},
				wantErr: false,
			},
			{
				name: "used exceeds total",
				slots: map[string]interface{}{
					"1": map[string]int{"total": 2, "used": 3},
				},
				wantErr: true,
			},
			{
				name: "negative values",
				slots: map[string]interface{}{
					"1": map[string]int{"total": -1, "used": 0},
				},
				wantErr: true,
			},
			{
				name: "invalid spell level",
				slots: map[string]interface{}{
					"10": map[string]int{"total": 1, "used": 0}, // Level 10 spells don't exist
				},
				wantErr: true,
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.Validate(tt.slots)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})
}