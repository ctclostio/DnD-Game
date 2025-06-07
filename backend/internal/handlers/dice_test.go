package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/testutil"
	"github.com/your-username/dnd-game/backend/pkg/dice"
)

func TestDiceHandler_Roll(t *testing.T) {
	t.Run("successful simple roll", func(t *testing.T) {
		mockDiceService := new(MockDiceRollService)
		mockRoller := new(MockDiceRoller)
		
		handler := NewDiceHandler(mockDiceService)
		handler.roller = mockRoller
		
		rollRequest := map[string]interface{}{
			"notation": "1d20+5",
			"purpose":  "Attack roll",
		}
		
		// Mock dice roll result
		mockRoller.On("Roll", "1d20+5").Return(dice.RollResult{
			Total:     23,
			Rolls:     []int{18},
			Modifiers: 5,
			Notation:  "1d20+5",
		}, nil)
		
		mockDiceService.On("RecordRoll", mock.MatchedBy(func(roll *models.DiceRoll) bool {
			return roll.DiceNotation == "1d20+5" &&
				roll.Result == 23 &&
				roll.Purpose == "Attack roll" &&
				len(roll.Rolls) == 1 &&
				roll.Rolls[0] == 18
		})).Return(nil)
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.POST("/api/dice/roll", handler.Roll)
		client.SetRouter(router)
		
		resp := client.POST("/api/dice/roll", rollRequest)
		
		resp.AssertOK()
		
		var response map[string]interface{}
		resp.DecodeJSON(&response)
		
		require.Equal(t, float64(23), response["total"])
		require.Equal(t, "1d20+5", response["notation"])
		rolls := response["rolls"].([]interface{})
		require.Len(t, rolls, 1)
		require.Equal(t, float64(18), rolls[0])
		
		mockDiceService.AssertExpectations(t)
		mockRoller.AssertExpectations(t)
	})

	t.Run("complex roll with multiple dice", func(t *testing.T) {
		mockDiceService := new(MockDiceRollService)
		mockRoller := new(MockDiceRoller)
		
		handler := NewDiceHandler(mockDiceService)
		handler.roller = mockRoller
		
		rollRequest := map[string]interface{}{
			"notation":       "2d6+1d4+3",
			"purpose":        "Damage roll",
			"character_id":   2,
			"game_session_id": 1,
		}
		
		mockRoller.On("Roll", "2d6+1d4+3").Return(dice.RollResult{
			Total:     13,
			Rolls:     []int{4, 5, 1}, // 2d6: 4,5 and 1d4: 1
			Modifiers: 3,
			Notation:  "2d6+1d4+3",
		}, nil)
		
		mockDiceService.On("RecordRoll", mock.AnythingOfType("*models.DiceRoll")).Return(nil)
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.POST("/api/dice/roll", handler.Roll)
		client.SetRouter(router)
		
		resp := client.POST("/api/dice/roll", rollRequest)
		
		resp.AssertOK()
		
		var response map[string]interface{}
		resp.DecodeJSON(&response)
		
		require.Equal(t, float64(13), response["total"])
		rolls := response["rolls"].([]interface{})
		require.Len(t, rolls, 3)
		
		mockDiceService.AssertExpectations(t)
	})

	t.Run("advantage roll", func(t *testing.T) {
		mockDiceService := new(MockDiceRollService)
		mockRoller := new(MockDiceRoller)
		
		handler := NewDiceHandler(mockDiceService)
		handler.roller = mockRoller
		
		rollRequest := map[string]interface{}{
			"notation":  "1d20",
			"purpose":   "Attack with advantage",
			"advantage": true,
		}
		
		// With advantage, roll 2d20 and keep highest
		mockRoller.On("RollWithAdvantage", "1d20").Return(dice.RollResult{
			Total:     18, // Highest of the two rolls
			Rolls:     []int{12, 18},
			Modifiers: 0,
			Notation:  "2d20kh1", // Keep highest 1
		}, nil)
		
		mockDiceService.On("RecordRoll", mock.MatchedBy(func(roll *models.DiceRoll) bool {
			return roll.RollType == "advantage" &&
				len(roll.Rolls) == 2 &&
				roll.Result == 18
		})).Return(nil)
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.POST("/api/dice/roll", handler.Roll)
		client.SetRouter(router)
		
		resp := client.POST("/api/dice/roll", rollRequest)
		
		resp.AssertOK()
		
		var response map[string]interface{}
		resp.DecodeJSON(&response)
		
		require.Equal(t, float64(18), response["total"])
		require.Equal(t, "advantage", response["roll_type"])
		rolls := response["rolls"].([]interface{})
		require.Len(t, rolls, 2)
		
		mockDiceService.AssertExpectations(t)
		mockRoller.AssertExpectations(t)
	})

	t.Run("disadvantage roll", func(t *testing.T) {
		mockDiceService := new(MockDiceRollService)
		mockRoller := new(MockDiceRoller)
		
		handler := NewDiceHandler(mockDiceService)
		handler.roller = mockRoller
		
		rollRequest := map[string]interface{}{
			"notation":     "1d20+3",
			"purpose":      "Stealth check with disadvantage",
			"disadvantage": true,
		}
		
		mockRoller.On("RollWithDisadvantage", "1d20+3").Return(dice.RollResult{
			Total:     8, // Lowest roll (5) + 3
			Rolls:     []int{5, 15},
			Modifiers: 3,
			Notation:  "2d20kl1+3", // Keep lowest 1
		}, nil)
		
		mockDiceService.On("RecordRoll", mock.AnythingOfType("*models.DiceRoll")).Return(nil)
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.POST("/api/dice/roll", handler.Roll)
		client.SetRouter(router)
		
		resp := client.POST("/api/dice/roll", rollRequest)
		
		resp.AssertOK()
		
		var response map[string]interface{}
		resp.DecodeJSON(&response)
		
		require.Equal(t, float64(8), response["total"])
		require.Equal(t, "disadvantage", response["roll_type"])
		
		mockDiceService.AssertExpectations(t)
	})

	t.Run("invalid dice notation", func(t *testing.T) {
		mockDiceService := new(MockDiceRollService)
		mockRoller := new(MockDiceRoller)
		
		handler := NewDiceHandler(mockDiceService)
		handler.roller = mockRoller
		
		rollRequest := map[string]interface{}{
			"notation": "invalid",
		}
		
		mockRoller.On("Roll", "invalid").Return(dice.RollResult{}, dice.ErrInvalidNotation)
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.POST("/api/dice/roll", handler.Roll)
		client.SetRouter(router)
		
		resp := client.POST("/api/dice/roll", rollRequest)
		
		resp.AssertBadRequest()
		resp.AssertBodyContains("Invalid dice notation")
		
		mockRoller.AssertExpectations(t)
	})

	t.Run("missing notation", func(t *testing.T) {
		handler := NewDiceHandler(nil)
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.POST("/api/dice/roll", handler.Roll)
		client.SetRouter(router)
		
		resp := client.POST("/api/dice/roll", map[string]interface{}{})
		
		resp.AssertBadRequest()
		resp.AssertBodyContains("notation required")
	})
}

func TestDiceHandler_GetHistory(t *testing.T) {
	t.Run("get roll history for session", func(t *testing.T) {
		mockDiceService := new(MockDiceRollService)
		
		handler := NewDiceHandler(mockDiceService)
		
		sessionID := int64(1)
		rolls := []*models.DiceRoll{
			testutil.NewDiceRollBuilder().
				WithType("attack").
				WithNotation("1d20+5").
				WithResult(23, []int{18}).
				Build(),
			testutil.NewDiceRollBuilder().
				WithType("damage").
				WithNotation("2d6+3").
				WithResult(11, []int{4, 4}).
				Build(),
			testutil.NewDiceRollBuilder().
				WithType("saving_throw").
				WithNotation("1d20+2").
				WithResult(15, []int{13}).
				Build(),
		}
		
		mockDiceService.On("GetSessionHistory", sessionID, 20).Return(rolls, nil)
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.GET("/api/sessions/:id/dice/history", handler.GetHistory)
		client.SetRouter(router)
		
		resp := client.GET("/api/sessions/1/dice/history")
		
		resp.AssertOK()
		
		var response map[string]interface{}
		resp.DecodeJSON(&response)
		
		history := response["rolls"].([]interface{})
		require.Len(t, history, 3)
		
		// Verify rolls are returned in order
		firstRoll := history[0].(map[string]interface{})
		require.Equal(t, "attack", firstRoll["roll_type"])
		require.Equal(t, float64(23), firstRoll["result"])
		
		mockDiceService.AssertExpectations(t)
	})

	t.Run("get history with custom limit", func(t *testing.T) {
		mockDiceService := new(MockDiceRollService)
		
		handler := NewDiceHandler(mockDiceService)
		
		mockDiceService.On("GetSessionHistory", int64(1), 5).Return([]*models.DiceRoll{}, nil)
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.GET("/api/sessions/:id/dice/history", handler.GetHistory)
		client.SetRouter(router)
		
		resp := client.GET("/api/sessions/1/dice/history?limit=5")
		
		resp.AssertOK()
		
		mockDiceService.AssertExpectations(t)
	})
}

func TestDiceHandler_GetStatistics(t *testing.T) {
	t.Run("get dice statistics for character", func(t *testing.T) {
		mockDiceService := new(MockDiceRollService)
		
		handler := NewDiceHandler(mockDiceService)
		
		stats := &models.DiceStatistics{
			CharacterID:   1,
			TotalRolls:    150,
			AverageRoll:   10.5,
			NaturalOnes:   7,
			NaturalTwenties: 8,
			RollsByType: map[string]int{
				"attack":       75,
				"damage":       40,
				"saving_throw": 20,
				"ability_check": 15,
			},
			LuckIndex: 1.02, // Slightly lucky
			CommonDice: map[string]int{
				"d20": 95,
				"d6":  35,
				"d8":  15,
				"d4":  5,
			},
		}
		
		mockDiceService.On("GetCharacterStatistics", int64(1)).Return(stats, nil)
		
		client := testutil.NewHTTPTestClient(t).WithUser(1)
		router := gin.New()
		router.GET("/api/characters/:id/dice/stats", handler.GetStatistics)
		client.SetRouter(router)
		
		resp := client.GET("/api/characters/1/dice/stats")
		
		resp.AssertOK()
		
		var response map[string]interface{}
		resp.DecodeJSON(&response)
		
		statsData := response["statistics"].(map[string]interface{})
		require.Equal(t, float64(150), statsData["total_rolls"])
		require.Equal(t, float64(10.5), statsData["average_roll"])
		require.Equal(t, float64(1.02), statsData["luck_index"])
		
		rollTypes := statsData["rolls_by_type"].(map[string]interface{})
		require.Equal(t, float64(75), rollTypes["attack"])
		
		mockDiceService.AssertExpectations(t)
	})
}

// Mock services
type MockDiceRollService struct {
	mock.Mock
}

func (m *MockDiceRollService) RecordRoll(roll *models.DiceRoll) error {
	args := m.Called(roll)
	return args.Error(0)
}

func (m *MockDiceRollService) GetSessionHistory(sessionID int64, limit int) ([]*models.DiceRoll, error) {
	args := m.Called(sessionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollService) GetCharacterStatistics(characterID int64) (*models.DiceStatistics, error) {
	args := m.Called(characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiceStatistics), args.Error(1)
}

type MockDiceRoller struct {
	mock.Mock
}

func (m *MockDiceRoller) Roll(notation string) (dice.RollResult, error) {
	args := m.Called(notation)
	return args.Get(0).(dice.RollResult), args.Error(1)
}

func (m *MockDiceRoller) RollWithAdvantage(notation string) (dice.RollResult, error) {
	args := m.Called(notation)
	return args.Get(0).(dice.RollResult), args.Error(1)
}

func (m *MockDiceRoller) RollWithDisadvantage(notation string) (dice.RollResult, error) {
	args := m.Called(notation)
	return args.Get(0).(dice.RollResult), args.Error(1)
}