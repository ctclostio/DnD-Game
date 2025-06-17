package services

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil"
)

// Helper functions to reduce code duplication in mock implementations

func handleErrorResult(args mock.Arguments, index int) error {
	return args.Error(index)
}

func handleSingleResult[T any](args mock.Arguments, valueIndex, errorIndex int) (*T, error) {
	if args.Get(valueIndex) == nil {
		return nil, args.Error(errorIndex)
	}
	return args.Get(valueIndex).(*T), args.Error(errorIndex)
}

func handleSliceResult[T any](args mock.Arguments, valueIndex, errorIndex int) ([]*T, error) {
	if args.Get(valueIndex) == nil {
		return nil, args.Error(errorIndex)
	}
	return args.Get(valueIndex).([]*T), args.Error(errorIndex)
}

func TestCombatAnalytics_TrackCombatAction(t *testing.T) {
	t.Run("successful damage action recording", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)

		analytics := NewCombatAnalyticsService(mockRepo, nil)

		combatID := uuid.New()
		targetID := "npc-1"

		action := &models.CombatActionLog{
			ID:          uuid.New(),
			CombatID:    combatID,
			ActorID:     "char-1",
			ActorType:   "character",
			ActionType:  "attack",
			TargetID:    &targetID,
			RollResults: models.JSONB(`{"attack": 18}`),
			DamageDealt: 12,
			Outcome:     "hit",
			RoundNumber: 1,
			TurnNumber:  0,
			Timestamp:   time.Now(),
		}

		mockRepo.On("CreateCombatAction", mock.MatchedBy(func(a *models.CombatActionLog) bool {
			return a.CombatID == action.CombatID &&
				a.ActorID == action.ActorID &&
				a.DamageDealt == action.DamageDealt &&
				a.Outcome == "hit"
		})).Return(nil)

		ctx := testutil.TestContext()
		err := analytics.TrackCombatAction(ctx, action)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("healing action recording", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)

		analytics := NewCombatAnalyticsService(mockRepo, nil)

		combatID := uuid.New()
		targetID := "char-1"

		action := &models.CombatActionLog{
			ID:          uuid.New(),
			CombatID:    combatID,
			ActorID:     "char-2",
			ActorType:   "character",
			ActionType:  "heal",
			TargetID:    &targetID,
			DamageDealt: 8, // Healing stored as positive damage
			Outcome:     "success",
			RoundNumber: 2,
			Timestamp:   time.Now(),
		}

		mockRepo.On("CreateCombatAction", mock.AnythingOfType("*models.CombatActionLog")).Return(nil)

		ctx := testutil.TestContext()
		err := analytics.TrackCombatAction(ctx, action)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("spell action with save", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)

		analytics := NewCombatAnalyticsService(mockRepo, nil)

		combatID := uuid.New()

		action := &models.CombatActionLog{
			ID:            uuid.New(),
			CombatID:      combatID,
			ActorID:       "char-3",
			ActorType:     "character",
			ActionType:    "spell",
			ResourcesUsed: models.JSONB(`{"spell_level": 3, "spell_name": "Fireball"}`),
			RollResults:   models.JSONB(`{"save_dc": 15}`),
			DamageDealt:   28,
			Outcome:       "hit",
			RoundNumber:   3,
			Timestamp:     time.Now(),
		}

		mockRepo.On("CreateCombatAction", mock.AnythingOfType("*models.CombatActionLog")).Return(nil)

		ctx := testutil.TestContext()
		err := analytics.TrackCombatAction(ctx, action)

		require.NoError(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)

		analytics := NewCombatAnalyticsService(mockRepo, nil)

		action := &models.CombatActionLog{
			ID:          uuid.New(),
			CombatID:    uuid.New(),
			ActorID:     "char-1",
			ActorType:   "character",
			ActionType:  "attack",
			RoundNumber: 1,
		}

		expectedErr := errors.New("database error")
		mockRepo.On("CreateCombatAction", action).Return(expectedErr)

		ctx := testutil.TestContext()
		err := analytics.TrackCombatAction(ctx, action)

		require.Error(t, err)
		require.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCombatAnalytics_FinalizeCombatAnalytics(t *testing.T) {
	t.Run("complete combat analytics", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)

		analytics := NewCombatAnalyticsService(mockRepo, nil)

		combatID := uuid.New()
		sessionID := uuid.New()

		combat := &models.Combat{
			ID:            combatID.String(),
			GameSessionID: sessionID.String(),
			Round:         5,
			Combatants: []models.Combatant{
				{
					ID:    "char-1",
					Name:  "Fighter",
					Type:  models.CombatantTypeCharacter,
					HP:    15,
					MaxHP: 30,
				},
				{
					ID:    "npc-1",
					Name:  "Orc",
					Type:  models.CombatantTypeNPC,
					HP:    0,
					MaxHP: 20,
				},
			},
		}

		targetID := "npc-1"
		actions := []*models.CombatActionLog{
			{
				ID:          uuid.New(),
				CombatID:    combatID,
				ActorID:     "char-1",
				ActorType:   "character",
				ActionType:  "attack",
				TargetID:    &targetID,
				DamageDealt: 20,
				Outcome:     "killing_blow",
				RoundNumber: 5,
			},
		}

		mockRepo.On("GetCombatActions", combatID).Return(actions, nil)
		mockRepo.On("CreateCombatAnalytics", mock.AnythingOfType("*models.CombatAnalytics")).Return(nil)
		mockRepo.On("CreateCombatantAnalytics", mock.AnythingOfType("*models.CombatantAnalytics")).Return(nil).Times(2)
		mockRepo.On("UpdateCombatAnalytics", mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("map[string]interface {}")).Return(nil)

		ctx := testutil.TestContext()
		result, err := analytics.FinalizeCombatAnalytics(ctx, combat, sessionID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Analytics)
		require.Equal(t, 20, result.Analytics.TotalDamageDealt) // Based on single action with 20 damage
		require.Equal(t, 0, result.Analytics.TotalHealingDone)  // No healing actions provided
		require.Equal(t, "char-1", result.Analytics.MVPID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("analytics with no actions", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)

		analytics := NewCombatAnalyticsService(mockRepo, nil)

		combatID := uuid.New()
		sessionID := uuid.New()

		combat := &models.Combat{
			ID:            combatID.String(),
			GameSessionID: sessionID.String(),
			Round:         1,
			Combatants:    []models.Combatant{},
		}

		actions := []*models.CombatActionLog{}

		mockRepo.On("GetCombatActions", combatID).Return(actions, nil)
		mockRepo.On("CreateCombatAnalytics", mock.AnythingOfType("*models.CombatAnalytics")).Return(nil)
		mockRepo.On("UpdateCombatAnalytics", mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("map[string]interface {}")).Return(nil)

		ctx := testutil.TestContext()
		result, err := analytics.FinalizeCombatAnalytics(ctx, combat, sessionID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 0, result.Analytics.TotalDamageDealt)
		require.Equal(t, 0, result.Analytics.TotalHealingDone)
	})
}

func TestCombatAnalytics_GetCombatAnalytics(t *testing.T) {
	t.Run("retrieve analytics for combat", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)

		analytics := NewCombatAnalyticsService(mockRepo, nil)

		combatID := uuid.New()
		sessionID := uuid.New()

		expectedAnalytics := &models.CombatAnalytics{
			ID:               uuid.New(),
			CombatID:         combatID,
			GameSessionID:    sessionID,
			CombatDuration:   10,
			TotalDamageDealt: 150,
			TotalHealingDone: 45,
			MVPID:            "char-1",
			MVPType:          "character",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		mockRepo.On("GetCombatAnalytics", combatID).Return(expectedAnalytics, nil)

		ctx := testutil.TestContext()
		result, err := analytics.GetCombatAnalytics(ctx, combatID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, combatID, result.CombatID)
		require.Equal(t, 150, result.TotalDamageDealt)
		require.Equal(t, 45, result.TotalHealingDone)
		require.Equal(t, "char-1", result.MVPID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("analytics not found", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)

		analytics := NewCombatAnalyticsService(mockRepo, nil)

		combatID := uuid.New()

		mockRepo.On("GetCombatAnalytics", combatID).Return(nil, models.ErrNotFound)

		ctx := testutil.TestContext()
		result, err := analytics.GetCombatAnalytics(ctx, combatID)

		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, models.ErrNotFound, err)
	})
}

func TestCombatAnalytics_GetCombatantAnalytics(t *testing.T) {
	t.Run("retrieve combatant analytics", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)

		analytics := NewCombatAnalyticsService(mockRepo, nil)

		analyticsID := uuid.New()

		expectedCombatants := []*models.CombatantAnalytics{
			{
				ID:                uuid.New(),
				CombatAnalyticsID: analyticsID,
				CombatantID:       "char-1",
				CombatantType:     "character",
				CombatantName:     "Fighter",
				DamageDealt:       65,
				DamageTaken:       22,
				HealingDone:       0,
				HealingReceived:   8,
				AttacksMade:       5,
				AttacksHit:        4,
				AttacksMissed:     1,
				CriticalHits:      1,
				FinalHP:           15,
				RoundsSurvived:    5,
			},
			{
				ID:                uuid.New(),
				CombatAnalyticsID: analyticsID,
				CombatantID:       "char-2",
				CombatantType:     "character",
				CombatantName:     "Cleric",
				DamageDealt:       30,
				DamageTaken:       15,
				HealingDone:       25,
				HealingReceived:   0,
				AttacksMade:       3,
				AttacksHit:        2,
				FinalHP:           20,
				RoundsSurvived:    5,
			},
		}

		mockRepo.On("GetCombatantAnalytics", analyticsID).Return(expectedCombatants, nil)

		ctx := testutil.TestContext()
		result, err := analytics.GetCombatantAnalytics(ctx, analyticsID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 2)
		require.Equal(t, "char-1", result[0].CombatantID)
		require.Equal(t, 65, result[0].DamageDealt)
		require.Equal(t, "char-2", result[1].CombatantID)
		require.Equal(t, 25, result[1].HealingDone)

		mockRepo.AssertExpectations(t)
	})
}

func TestCombatAnalytics_GetCombatAnalyticsBySession(t *testing.T) {
	t.Run("retrieve all analytics for session", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)

		analytics := NewCombatAnalyticsService(mockRepo, nil)

		sessionID := uuid.New()

		expectedAnalytics := []*models.CombatAnalytics{
			{
				ID:               uuid.New(),
				CombatID:         uuid.New(),
				GameSessionID:    sessionID,
				CombatDuration:   5,
				TotalDamageDealt: 95,
				TotalHealingDone: 25,
				MVPID:            "char-1",
				MVPType:          "character",
			},
			{
				ID:               uuid.New(),
				CombatID:         uuid.New(),
				GameSessionID:    sessionID,
				CombatDuration:   8,
				TotalDamageDealt: 120,
				TotalHealingDone: 35,
				MVPID:            "char-2",
				MVPType:          "character",
			},
		}

		mockRepo.On("GetCombatAnalyticsBySession", sessionID).Return(expectedAnalytics, nil)

		ctx := testutil.TestContext()
		result, err := analytics.GetCombatAnalyticsBySession(ctx, sessionID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 2)
		require.Equal(t, sessionID, result[0].GameSessionID)
		require.Equal(t, 95, result[0].TotalDamageDealt)
		require.Equal(t, sessionID, result[1].GameSessionID)
		require.Equal(t, 120, result[1].TotalDamageDealt)

		mockRepo.AssertExpectations(t)
	})
}

// Mock repository for combat analytics
type MockCombatAnalyticsRepository struct {
	mock.Mock
}

func (m *MockCombatAnalyticsRepository) CreateCombatAction(action *models.CombatActionLog) error {
	args := m.Called(action)
	return args.Error(0)
}

func (m *MockCombatAnalyticsRepository) GetCombatActions(combatID uuid.UUID) ([]*models.CombatActionLog, error) {
	args := m.Called(combatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CombatActionLog), args.Error(1)
}

func (m *MockCombatAnalyticsRepository) CreateCombatAnalytics(analytics *models.CombatAnalytics) error {
	args := m.Called(analytics)
	return handleErrorResult(args, 0)
}

func (m *MockCombatAnalyticsRepository) GetCombatAnalytics(combatID uuid.UUID) (*models.CombatAnalytics, error) {
	args := m.Called(combatID)
	return handleSingleResult[models.CombatAnalytics](args, 0, 1)
}

func (m *MockCombatAnalyticsRepository) UpdateCombatAnalytics(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return handleErrorResult(args, 0)
}

func (m *MockCombatAnalyticsRepository) CreateCombatantAnalytics(analytics *models.CombatantAnalytics) error {
	args := m.Called(analytics)
	return handleErrorResult(args, 0)
}

func (m *MockCombatAnalyticsRepository) GetCombatantAnalytics(analyticsID uuid.UUID) ([]*models.CombatantAnalytics, error) {
	args := m.Called(analyticsID)
	return handleSliceResult[models.CombatantAnalytics](args, 0, 1)
}

func (m *MockCombatAnalyticsRepository) GetCombatAnalyticsBySession(sessionID uuid.UUID) ([]*models.CombatAnalytics, error) {
	args := m.Called(sessionID)
	return handleSliceResult[models.CombatAnalytics](args, 0, 1)
}

// Mock combat service
type MockCombatService struct {
	mock.Mock
}

// Add missing methods for CombatAnalyticsRepository
func (m *MockCombatAnalyticsRepository) UpdateCombatantAnalytics(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return handleErrorResult(args, 0)
}

func (m *MockCombatAnalyticsRepository) CreateAutoCombatResolution(resolution *models.AutoCombatResolution) error {
	args := m.Called(resolution)
	return handleErrorResult(args, 0)
}

func (m *MockCombatAnalyticsRepository) GetAutoCombatResolution(id uuid.UUID) (*models.AutoCombatResolution, error) {
	args := m.Called(id)
	return handleSingleResult[models.AutoCombatResolution](args, 0, 1)
}

func (m *MockCombatAnalyticsRepository) GetAutoCombatResolutionsBySession(sessionID uuid.UUID) ([]*models.AutoCombatResolution, error) {
	args := m.Called(sessionID)
	return handleSliceResult[models.AutoCombatResolution](args, 0, 1)
}

func (m *MockCombatAnalyticsRepository) CreateBattleMap(battleMap *models.BattleMap) error {
	args := m.Called(battleMap)
	return handleErrorResult(args, 0)
}

func (m *MockCombatAnalyticsRepository) GetBattleMap(id uuid.UUID) (*models.BattleMap, error) {
	args := m.Called(id)
	return handleSingleResult[models.BattleMap](args, 0, 1)
}

func (m *MockCombatAnalyticsRepository) GetBattleMapByCombat(combatID uuid.UUID) (*models.BattleMap, error) {
	args := m.Called(combatID)
	return handleSingleResult[models.BattleMap](args, 0, 1)
}

func (m *MockCombatAnalyticsRepository) GetBattleMapsBySession(sessionID uuid.UUID) ([]*models.BattleMap, error) {
	args := m.Called(sessionID)
	return handleSliceResult[models.BattleMap](args, 0, 1)
}

func (m *MockCombatAnalyticsRepository) UpdateBattleMap(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return handleErrorResult(args, 0)
}

func (m *MockCombatAnalyticsRepository) CreateOrUpdateInitiativeRule(rule *models.SmartInitiativeRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockCombatAnalyticsRepository) GetInitiativeRule(sessionID uuid.UUID, entityID string) (*models.SmartInitiativeRule, error) {
	args := m.Called(sessionID, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SmartInitiativeRule), args.Error(1)
}

func (m *MockCombatAnalyticsRepository) GetInitiativeRulesBySession(sessionID uuid.UUID) ([]*models.SmartInitiativeRule, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SmartInitiativeRule), args.Error(1)
}

func (m *MockCombatAnalyticsRepository) GetCombatActionsByRound(combatID uuid.UUID, roundNumber int) ([]*models.CombatActionLog, error) {
	args := m.Called(combatID, roundNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CombatActionLog), args.Error(1)
}
