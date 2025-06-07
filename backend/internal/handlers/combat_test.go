package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-org/dnd-game/internal/models"
	"github.com/your-org/dnd-game/internal/testutil"
)

func TestCombatHandler_StartCombat(t *testing.T) {
	tests := []testutil.HTTPTestCase{
		{
			Name:   "successful combat start",
			Method: http.MethodPost,
			Path:   "/api/sessions/1/combat/start",
			Body: map[string]interface{}{
				"participants": []map[string]interface{}{
					{
						"character_id": 1,
						"initiative":   18,
					},
					{
						"character_id": 2,
						"initiative":   15,
					},
					{
						"npc_id":     1,
						"initiative": 12,
					},
				},
			},
			Auth:           true,
			UserID:         1,
			ExpectedStatus: http.StatusCreated,
			Setup: func() {
				// Mock setup would go here
			},
		},
		{
			Name:           "missing participants",
			Method:         http.MethodPost,
			Path:           "/api/sessions/1/combat/start",
			Body:           map[string]interface{}{},
			Auth:           true,
			UserID:         1,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "participants required",
		},
		{
			Name:   "unauthorized user",
			Method: http.MethodPost,
			Path:   "/api/sessions/1/combat/start",
			Body: map[string]interface{}{
				"participants": []map[string]interface{}{},
			},
			Auth:           false,
			ExpectedStatus: http.StatusUnauthorized,
		},
	}

	// Create router with mocked dependencies
	router := setupTestRouter()
	testutil.RunHTTPTestCases(t, router, tests)
}

func TestCombatHandler_StartCombat_Detailed(t *testing.T) {
	t.Run("successful combat initialization", func(t *testing.T) {
		// Setup mocks
		mockCombatService := new(testutil.MockCombatService)
		mockGameService := new(MockGameService)
		mockCharService := new(testutil.MockCharacterService)
		mockWebSocketHub := new(testutil.MockWebSocketHub)
		
		handler := NewCombatHandler(mockCombatService, mockGameService, mockCharService, mockWebSocketHub)
		
		// Test data
		sessionID := int64(1)
		userID := int64(1)
		session := testutil.NewGameSessionBuilder().
			WithID(sessionID).
			WithDM(userID).
			Build()
		
		char1 := testutil.NewCharacterBuilder().
			WithID(1).
			WithUserID(userID).
			WithName("Fighter").
			Build()
		
		char2 := testutil.NewCharacterBuilder().
			WithID(2).
			WithUserID(2).
			WithName("Cleric").
			Build()
		
		participants := []models.CombatParticipant{
			testutil.NewCombatParticipantBuilder().
				WithID("char-1").
				WithName("Fighter").
				WithInitiative(18).
				Build(),
			testutil.NewCombatParticipantBuilder().
				WithID("char-2").
				WithName("Cleric").
				WithInitiative(15).
				Build(),
			testutil.NewCombatParticipantBuilder().
				WithID("npc-1").
				WithName("Goblin").
				AsNPC().
				WithInitiative(12).
				Build(),
		}
		
		combat := testutil.NewCombatBuilder().
			WithID(1).
			WithGameSession(sessionID).
			WithParticipants(participants...).
			Build()
		
		// Setup expectations
		mockGameService.On("GetSession", sessionID).Return(session, nil)
		mockGameService.On("ValidateUserInSession", userID, sessionID).Return(true, nil)
		mockCharService.On("GetByID", int64(1)).Return(char1, nil)
		mockCharService.On("GetByID", int64(2)).Return(char2, nil)
		mockCombatService.On("StartCombat", sessionID, mock.MatchedBy(func(p []models.CombatParticipant) bool {
			return len(p) == 3 && p[0].Initiative == 18
		})).Return(combat, nil)
		mockWebSocketHub.On("SendToSession", sessionID, mock.MatchedBy(func(msg interface{}) bool {
			m, ok := msg.(map[string]interface{})
			return ok && m["type"] == "combat_started"
		}))
		
		// Create request
		client := testutil.NewHTTPTestClient(t).WithUser(userID)
		router := gin.New()
		router.POST("/api/sessions/:id/combat/start", handler.StartCombat)
		client.SetRouter(router)
		
		requestBody := map[string]interface{}{
			"participants": []map[string]interface{}{
				{"character_id": 1, "initiative": 18},
				{"character_id": 2, "initiative": 15},
				{"npc_id": 1, "name": "Goblin", "hp": 7, "ac": 15, "initiative": 12},
			},
		}
		
		// Execute
		resp := client.POST("/api/sessions/1/combat/start", requestBody)
		
		// Assert
		resp.AssertCreated()
		
		var response map[string]interface{}
		resp.DecodeJSON(&response)
		
		require.NotNil(t, response["combat"])
		combatData := response["combat"].(map[string]interface{})
		require.Equal(t, float64(1), combatData["id"])
		require.Equal(t, "active", combatData["status"])
		
		// Verify all expectations met
		mockCombatService.AssertExpectations(t)
		mockGameService.AssertExpectations(t)
		mockCharService.AssertExpectations(t)
		mockWebSocketHub.AssertExpectations(t)
	})

	t.Run("user not DM", func(t *testing.T) {
		mockCombatService := new(testutil.MockCombatService)
		mockGameService := new(MockGameService)
		mockCharService := new(testutil.MockCharacterService)
		mockWebSocketHub := new(testutil.MockWebSocketHub)
		
		handler := NewCombatHandler(mockCombatService, mockGameService, mockCharService, mockWebSocketHub)
		
		session := testutil.NewGameSessionBuilder().
			WithID(1).
			WithDM(2). // Different user is DM
			Build()
		
		mockGameService.On("GetSession", int64(1)).Return(session, nil)
		mockGameService.On("ValidateUserInSession", int64(1), int64(1)).Return(true, nil)
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.POST("/api/sessions/:id/combat/start", handler.StartCombat)
		client.SetRouter(router)
		
		resp := client.POST("/api/sessions/1/combat/start", map[string]interface{}{
			"participants": []interface{}{},
		})
		
		resp.AssertForbidden()
		resp.AssertBodyContains("must be DM")
	})
}

func TestCombatHandler_NextTurn(t *testing.T) {
	t.Run("successful turn advancement", func(t *testing.T) {
		mockCombatService := new(testutil.MockCombatService)
		mockGameService := new(MockGameService)
		mockCharService := new(testutil.MockCharacterService)
		mockWebSocketHub := new(testutil.MockWebSocketHub)
		
		handler := NewCombatHandler(mockCombatService, mockGameService, mockCharService, mockWebSocketHub)
		
		combat := testutil.NewCombatBuilder().
			WithID(1).
			WithGameSession(1).
			Build()
		
		mockCombatService.On("GetCombat", int64(1)).Return(combat, nil)
		mockGameService.On("ValidateUserInSession", int64(1), int64(1)).Return(true, nil)
		mockCombatService.On("NextTurn", int64(1)).Return(nil)
		mockWebSocketHub.On("SendToSession", int64(1), mock.AnythingOfType("map[string]interface {}"))
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.POST("/api/combat/:id/next-turn", handler.NextTurn)
		client.SetRouter(router)
		
		resp := client.POST("/api/combat/1/next-turn", nil)
		
		resp.AssertOK()
		resp.AssertBodyContains("Turn advanced")
		
		mockCombatService.AssertExpectations(t)
		mockWebSocketHub.AssertExpectations(t)
	})
}

func TestCombatHandler_RecordAction(t *testing.T) {
	t.Run("successful attack action", func(t *testing.T) {
		mockCombatService := new(testutil.MockCombatService)
		mockGameService := new(MockGameService)
		mockCharService := new(testutil.MockCharacterService)
		mockAnalytics := new(MockCombatAnalyticsService)
		mockWebSocketHub := new(testutil.MockWebSocketHub)
		
		handler := NewCombatHandler(mockCombatService, mockGameService, mockCharService, mockWebSocketHub)
		handler.analytics = mockAnalytics
		
		combat := testutil.NewCombatBuilder().
			WithID(1).
			WithGameSession(1).
			Build()
		
		actionRequest := map[string]interface{}{
			"action_type": "attack",
			"actor_id":    "char-1",
			"target_id":   "npc-1",
			"dice_roll":   18,
			"damage":      12,
			"damage_type": "slashing",
			"success":     true,
		}
		
		mockCombatService.On("GetCombat", int64(1)).Return(combat, nil)
		mockGameService.On("ValidateUserInSession", int64(1), int64(1)).Return(true, nil)
		mockAnalytics.On("RecordCombatAction", mock.AnythingOfType("*context.valueCtx"), mock.MatchedBy(func(action *models.CombatAction) bool {
			return action.ActionType == "attack" &&
				action.ActorID == "char-1" &&
				action.Damage == 12
		})).Return(nil)
		mockCombatService.On("ApplyDamage", int64(1), "npc-1", 12).Return(nil)
		mockWebSocketHub.On("SendToSession", int64(1), mock.AnythingOfType("map[string]interface {}"))
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.POST("/api/combat/:id/action", handler.RecordAction)
		client.SetRouter(router)
		
		resp := client.POST("/api/combat/1/action", actionRequest)
		
		resp.AssertOK()
		
		mockCombatService.AssertExpectations(t)
		mockAnalytics.AssertExpectations(t)
		mockWebSocketHub.AssertExpectations(t)
	})

	t.Run("healing action", func(t *testing.T) {
		mockCombatService := new(testutil.MockCombatService)
		mockGameService := new(MockGameService)
		mockCharService := new(testutil.MockCharacterService)
		mockAnalytics := new(MockCombatAnalyticsService)
		mockWebSocketHub := new(testutil.MockWebSocketHub)
		
		handler := NewCombatHandler(mockCombatService, mockGameService, mockCharService, mockWebSocketHub)
		handler.analytics = mockAnalytics
		
		combat := testutil.NewCombatBuilder().Build()
		
		actionRequest := map[string]interface{}{
			"action_type": "healing",
			"actor_id":    "char-2",
			"target_id":   "char-1",
			"healing":     8,
			"spell_name":  "Cure Wounds",
			"spell_level": 1,
		}
		
		mockCombatService.On("GetCombat", int64(1)).Return(combat, nil)
		mockGameService.On("ValidateUserInSession", int64(1), int64(1)).Return(true, nil)
		mockAnalytics.On("RecordCombatAction", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("*models.CombatAction")).Return(nil)
		mockCombatService.On("ApplyHealing", int64(1), "char-1", 8).Return(nil)
		mockWebSocketHub.On("SendToSession", mock.AnythingOfType("int64"), mock.AnythingOfType("map[string]interface {}"))
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.POST("/api/combat/:id/action", handler.RecordAction)
		client.SetRouter(router)
		
		resp := client.POST("/api/combat/1/action", actionRequest)
		
		resp.AssertOK()
		
		mockCombatService.AssertExpectations(t)
		mockAnalytics.AssertExpectations(t)
	})
}

func TestCombatHandler_EndCombat(t *testing.T) {
	t.Run("successful combat end", func(t *testing.T) {
		mockCombatService := new(testutil.MockCombatService)
		mockGameService := new(MockGameService)
		mockCharService := new(testutil.MockCharacterService)
		mockAnalytics := new(MockCombatAnalyticsService)
		mockWebSocketHub := new(testutil.MockWebSocketHub)
		
		handler := NewCombatHandler(mockCombatService, mockGameService, mockCharService, mockWebSocketHub)
		handler.analytics = mockAnalytics
		
		session := testutil.NewGameSessionBuilder().
			WithID(1).
			WithDM(1).
			Build()
		
		combat := testutil.NewCombatBuilder().
			WithID(1).
			WithGameSession(1).
			Build()
		
		summary := &models.CombatSummary{
			CombatID:     1,
			Rounds:       5,
			TotalDamage:  120,
			TotalHealing: 35,
		}
		
		mockCombatService.On("GetCombat", int64(1)).Return(combat, nil)
		mockGameService.On("GetSession", int64(1)).Return(session, nil)
		mockGameService.On("ValidateUserInSession", int64(1), int64(1)).Return(true, nil)
		mockCombatService.On("EndCombat", int64(1)).Return(nil)
		mockAnalytics.On("GetCombatSummary", mock.AnythingOfType("*context.valueCtx"), int64(1)).Return(summary, nil)
		mockWebSocketHub.On("SendToSession", int64(1), mock.AnythingOfType("map[string]interface {}"))
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.POST("/api/combat/:id/end", handler.EndCombat)
		client.SetRouter(router)
		
		resp := client.POST("/api/combat/1/end", nil)
		
		resp.AssertOK()
		
		var response map[string]interface{}
		resp.DecodeJSON(&response)
		
		require.NotNil(t, response["summary"])
		summaryData := response["summary"].(map[string]interface{})
		require.Equal(t, float64(5), summaryData["rounds"])
		require.Equal(t, float64(120), summaryData["total_damage"])
		
		mockCombatService.AssertExpectations(t)
		mockAnalytics.AssertExpectations(t)
	})
}

// Mock services
type MockGameService struct {
	mock.Mock
}

func (m *MockGameService) GetSession(id int64) (*models.GameSession, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameService) ValidateUserInSession(userID, sessionID int64) (bool, error) {
	args := m.Called(userID, sessionID)
	return args.Bool(0), args.Error(1)
}

type MockCombatAnalyticsService struct {
	mock.Mock
}

func (m *MockCombatAnalyticsService) RecordCombatAction(ctx context.Context, action *models.CombatAction) error {
	args := m.Called(ctx, action)
	return args.Error(0)
}

func (m *MockCombatAnalyticsService) GetCombatSummary(ctx context.Context, combatID int64) (*models.CombatSummary, error) {
	args := m.Called(ctx, combatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CombatSummary), args.Error(1)
}

// Helper function to setup test router
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add middleware and routes as needed
	// This would typically include auth middleware, error handling, etc.
	
	return router
}