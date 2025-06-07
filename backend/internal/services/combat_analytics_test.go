package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-org/dnd-game/internal/models"
	"github.com/your-org/dnd-game/internal/testutil"
)

func TestCombatAnalytics_RecordCombatAction(t *testing.T) {
	t.Run("successful damage action recording", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)
		mockCombatRepo := new(testutil.MockCombatRepository)
		
		analytics := NewCombatAnalyticsService(mockRepo, mockCombatRepo)
		
		combat := testutil.NewCombatBuilder().
			WithID(1).
			WithGameSession(1).
			Build()
		
		action := &models.CombatAction{
			CombatID:     combat.ID,
			ActorID:      "char-1",
			ActorType:    "character",
			ActionType:   "attack",
			TargetID:     "npc-1",
			TargetType:   "npc",
			DiceRoll:     18,
			Damage:       12,
			DamageType:   "slashing",
			Success:      true,
			Round:        1,
			TurnOrder:    0,
			Description:  "Aragorn attacks Orc with Longsword",
		}
		
		mockCombatRepo.On("GetByID", combat.ID).Return(combat, nil)
		mockRepo.On("RecordAction", mock.MatchedBy(func(a *models.CombatAction) bool {
			return a.CombatID == action.CombatID &&
				a.ActorID == action.ActorID &&
				a.Damage == action.Damage &&
				a.Success == true
		})).Return(nil)
		
		ctx := testutil.TestContext()
		err := analytics.RecordCombatAction(ctx, action)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCombatRepo.AssertExpectations(t)
	})

	t.Run("healing action recording", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)
		mockCombatRepo := new(testutil.MockCombatRepository)
		
		analytics := NewCombatAnalyticsService(mockRepo, mockCombatRepo)
		
		combat := testutil.NewCombatBuilder().Build()
		
		action := &models.CombatAction{
			CombatID:     combat.ID,
			ActorID:      "char-2",
			ActorType:    "character",
			ActionType:   "healing",
			TargetID:     "char-1",
			TargetType:   "character",
			Healing:      8,
			Success:      true,
			Round:        2,
			Description:  "Cleric casts Cure Wounds on Fighter",
		}
		
		mockCombatRepo.On("GetByID", combat.ID).Return(combat, nil)
		mockRepo.On("RecordAction", mock.AnythingOfType("*models.CombatAction")).Return(nil)
		
		ctx := testutil.TestContext()
		err := analytics.RecordCombatAction(ctx, action)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("spell action with save", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)
		mockCombatRepo := new(testutil.MockCombatRepository)
		
		analytics := NewCombatAnalyticsService(mockRepo, mockCombatRepo)
		
		combat := testutil.NewCombatBuilder().Build()
		
		action := &models.CombatAction{
			CombatID:     combat.ID,
			ActorID:      "char-3",
			ActorType:    "character",
			ActionType:   "spell",
			SpellName:    "Fireball",
			SpellLevel:   3,
			SaveDC:       15,
			SaveType:     "dexterity",
			TargetsSaved: []string{"npc-2"},
			TargetsFailed: []string{"npc-1", "npc-3"},
			Damage:       28,
			DamageType:   "fire",
			Round:        3,
			Description:  "Wizard casts Fireball",
		}
		
		mockCombatRepo.On("GetByID", combat.ID).Return(combat, nil)
		mockRepo.On("RecordAction", mock.AnythingOfType("*models.CombatAction")).Return(nil)
		
		ctx := testutil.TestContext()
		err := analytics.RecordCombatAction(ctx, action)
		
		require.NoError(t, err)
	})

	t.Run("invalid combat ID", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)
		mockCombatRepo := new(testutil.MockCombatRepository)
		
		analytics := NewCombatAnalyticsService(mockRepo, mockCombatRepo)
		
		action := &models.CombatAction{
			CombatID: 999,
		}
		
		mockCombatRepo.On("GetByID", int64(999)).Return(nil, models.ErrNotFound)
		
		ctx := testutil.TestContext()
		err := analytics.RecordCombatAction(ctx, action)
		
		require.Error(t, err)
	})
}

func TestCombatAnalytics_GetCombatSummary(t *testing.T) {
	t.Run("complete combat summary", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)
		mockCombatRepo := new(testutil.MockCombatRepository)
		
		analytics := NewCombatAnalyticsService(mockRepo, mockCombatRepo)
		
		combat := testutil.NewCombatBuilder().
			WithID(1).
			WithRound(5).
			Build()
		combat.Status = "completed"
		combat.EndTime = time.Now()
		
		summary := &models.CombatSummary{
			CombatID:      combat.ID,
			GameSessionID: combat.GameSessionID,
			Duration:      5 * time.Minute,
			Rounds:        5,
			Participants: []models.ParticipantStats{
				{
					ParticipantID:   "char-1",
					ParticipantType: "character",
					Name:            "Fighter",
					DamageDealt:     65,
					DamageTaken:     22,
					HealingDone:     0,
					HealingReceived: 8,
					AttacksMade:     5,
					AttacksHit:      4,
					AttacksMissed:   1,
					SpellsCast:      0,
					KillCount:       2,
					Unconscious:     0,
				},
				{
					ParticipantID:   "char-2",
					ParticipantType: "character",
					Name:            "Cleric",
					DamageDealt:     30,
					DamageTaken:     15,
					HealingDone:     25,
					HealingReceived: 0,
					AttacksMade:     3,
					AttacksHit:      2,
					SpellsCast:      4,
					KillCount:       1,
				},
			},
			TotalDamage:  95,
			TotalHealing: 25,
			MostDamageDealt: models.ParticipantReference{
				ID:     "char-1",
				Name:   "Fighter",
				Amount: 65,
			},
			MostHealingDone: models.ParticipantReference{
				ID:     "char-2",
				Name:   "Cleric",
				Amount: 25,
			},
			FirstBlood: models.ParticipantReference{
				ID:   "char-1",
				Name: "Fighter",
			},
			Victors: []string{"char-1", "char-2"},
		}
		
		mockCombatRepo.On("GetByID", combat.ID).Return(combat, nil)
		mockRepo.On("GetCombatSummary", combat.ID).Return(summary, nil)
		
		ctx := testutil.TestContext()
		result, err := analytics.GetCombatSummary(ctx, combat.ID)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 5, result.Rounds)
		require.Equal(t, 95, result.TotalDamage)
		require.Equal(t, 25, result.TotalHealing)
		require.Len(t, result.Participants, 2)
		
		// Verify MVP calculations
		require.Equal(t, "Fighter", result.MostDamageDealt.Name)
		require.Equal(t, 65, result.MostDamageDealt.Amount)
		require.Equal(t, "Cleric", result.MostHealingDone.Name)
		
		mockRepo.AssertExpectations(t)
		mockCombatRepo.AssertExpectations(t)
	})

	t.Run("ongoing combat summary", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)
		mockCombatRepo := new(testutil.MockCombatRepository)
		
		analytics := NewCombatAnalyticsService(mockRepo, mockCombatRepo)
		
		combat := testutil.NewCombatBuilder().
			WithID(1).
			WithRound(3).
			Build()
		combat.Status = "active"
		
		summary := &models.CombatSummary{
			CombatID:      combat.ID,
			GameSessionID: combat.GameSessionID,
			Rounds:        3,
			Participants:  []models.ParticipantStats{},
			TotalDamage:   45,
			TotalHealing:  10,
		}
		
		mockCombatRepo.On("GetByID", combat.ID).Return(combat, nil)
		mockRepo.On("GetCombatSummary", combat.ID).Return(summary, nil)
		
		ctx := testutil.TestContext()
		result, err := analytics.GetCombatSummary(ctx, combat.ID)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "active", combat.Status)
		require.Equal(t, 3, result.Rounds)
	})
}

func TestCombatAnalytics_GetPlayerStatistics(t *testing.T) {
	t.Run("comprehensive player statistics", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)
		
		analytics := NewCombatAnalyticsService(mockRepo, nil)
		
		charID := int64(1)
		stats := &models.PlayerCombatStats{
			CharacterID:      charID,
			TotalCombats:     25,
			CombatsWon:       20,
			CombatsLost:      3,
			CombatsFled:      2,
			TotalDamageDealt: 1250,
			TotalDamageTaken: 480,
			TotalHealingDone: 320,
			TotalHealingReceived: 150,
			TotalKills:       45,
			TimesUnconscious: 3,
			FavoriteTarget: models.TargetInfo{
				Type:        "monster_type",
				Name:        "Orc",
				TimesKilled: 12,
			},
			MostUsedWeapon: models.WeaponInfo{
				Name:   "Longsword +1",
				Uses:   156,
				Damage: 680,
			},
			MostCastSpell: models.SpellInfo{
				Name:  "Cure Wounds",
				Casts: 28,
			},
			CombatRole: "damage_dealer",
			AverageStats: models.AverageStats{
				DamagePerRound:   8.5,
				HealingPerRound:  2.1,
				AccuracyRate:     0.75,
				CriticalHitRate:  0.08,
				SaveSuccessRate:  0.65,
			},
		}
		
		mockRepo.On("GetPlayerStats", charID).Return(stats, nil)
		
		ctx := testutil.TestContext()
		result, err := analytics.GetPlayerStatistics(ctx, charID)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 25, result.TotalCombats)
		require.Equal(t, 20, result.CombatsWon)
		require.Equal(t, 1250, result.TotalDamageDealt)
		require.Equal(t, "damage_dealer", result.CombatRole)
		require.Equal(t, 0.75, result.AverageStats.AccuracyRate)
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("new player with no combat history", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)
		
		analytics := NewCombatAnalyticsService(mockRepo, nil)
		
		charID := int64(999)
		stats := &models.PlayerCombatStats{
			CharacterID:  charID,
			TotalCombats: 0,
			CombatRole:   "unknown",
			AverageStats: models.AverageStats{},
		}
		
		mockRepo.On("GetPlayerStats", charID).Return(stats, nil)
		
		ctx := testutil.TestContext()
		result, err := analytics.GetPlayerStatistics(ctx, charID)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 0, result.TotalCombats)
		require.Equal(t, "unknown", result.CombatRole)
	})
}

func TestCombatAnalytics_GetSessionStatistics(t *testing.T) {
	t.Run("session with multiple combats", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)
		
		analytics := NewCombatAnalyticsService(mockRepo, nil)
		
		sessionID := int64(1)
		stats := &models.SessionCombatStats{
			GameSessionID:    sessionID,
			TotalCombats:     5,
			AverageDuration:  7 * time.Minute,
			AverageRounds:    6.2,
			TotalDamageDealt: 450,
			TotalHealingDone: 120,
			PlayerDeaths:     1,
			MonsterDeaths:    18,
			CombatsByType: map[string]int{
				"random_encounter": 3,
				"boss_fight":       1,
				"story_combat":     1,
			},
			DifficultyBreakdown: map[string]int{
				"easy":   2,
				"medium": 2,
				"hard":   1,
			},
			MostDangerousEnemy: models.EnemyInfo{
				Name:        "Ancient Red Dragon",
				DamageDealt: 180,
				PlayerKills: 1,
			},
		}
		
		mockRepo.On("GetSessionStats", sessionID).Return(stats, nil)
		
		ctx := testutil.TestContext()
		result, err := analytics.GetSessionStatistics(ctx, sessionID)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 5, result.TotalCombats)
		require.Equal(t, 7*time.Minute, result.AverageDuration)
		require.Equal(t, 6.2, result.AverageRounds)
		require.Equal(t, "Ancient Red Dragon", result.MostDangerousEnemy.Name)
		
		mockRepo.AssertExpectations(t)
	})
}

func TestCombatAnalytics_AnalyzeCombatTrends(t *testing.T) {
	t.Run("analyze player combat trends", func(t *testing.T) {
		mockRepo := new(MockCombatAnalyticsRepository)
		
		analytics := NewCombatAnalyticsService(mockRepo, nil)
		
		charID := int64(1)
		timeRange := 30 * 24 * time.Hour // 30 days
		
		trends := &models.CombatTrends{
			CharacterID: charID,
			TimeRange:   timeRange,
			DamageDealtTrend: models.TrendData{
				Direction:  "increasing",
				Percentage: 15.5,
				DataPoints: []models.DataPoint{
					{Time: time.Now().Add(-30 * 24 * time.Hour), Value: 30},
					{Time: time.Now().Add(-20 * 24 * time.Hour), Value: 35},
					{Time: time.Now().Add(-10 * 24 * time.Hour), Value: 40},
					{Time: time.Now(), Value: 45},
				},
			},
			AccuracyTrend: models.TrendData{
				Direction:  "stable",
				Percentage: 2.1,
				DataPoints: []models.DataPoint{
					{Time: time.Now().Add(-30 * 24 * time.Hour), Value: 0.73},
					{Time: time.Now().Add(-20 * 24 * time.Hour), Value: 0.74},
					{Time: time.Now().Add(-10 * 24 * time.Hour), Value: 0.75},
					{Time: time.Now(), Value: 0.75},
				},
			},
			SurvivabilityTrend: models.TrendData{
				Direction:  "improving",
				Percentage: 25.0,
				DataPoints: []models.DataPoint{
					{Time: time.Now().Add(-30 * 24 * time.Hour), Value: 0.60},
					{Time: time.Now().Add(-20 * 24 * time.Hour), Value: 0.70},
					{Time: time.Now().Add(-10 * 24 * time.Hour), Value: 0.75},
					{Time: time.Now(), Value: 0.80},
				},
			},
			Recommendations: []string{
				"Your damage output is increasing steadily",
				"Consider using more defensive tactics to improve survivability",
				"Your accuracy is consistent - good job!",
			},
		}
		
		mockRepo.On("AnalyzeTrends", charID, timeRange).Return(trends, nil)
		
		ctx := testutil.TestContext()
		result, err := analytics.AnalyzeCombatTrends(ctx, charID, timeRange)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "increasing", result.DamageDealtTrend.Direction)
		require.Equal(t, 15.5, result.DamageDealtTrend.Percentage)
		require.Len(t, result.Recommendations, 3)
		
		mockRepo.AssertExpectations(t)
	})
}

// Mock repository for combat analytics
type MockCombatAnalyticsRepository struct {
	mock.Mock
}

func (m *MockCombatAnalyticsRepository) RecordAction(action *models.CombatAction) error {
	args := m.Called(action)
	return args.Error(0)
}

func (m *MockCombatAnalyticsRepository) GetCombatSummary(combatID int64) (*models.CombatSummary, error) {
	args := m.Called(combatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CombatSummary), args.Error(1)
}

func (m *MockCombatAnalyticsRepository) GetPlayerStats(characterID int64) (*models.PlayerCombatStats, error) {
	args := m.Called(characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerCombatStats), args.Error(1)
}

func (m *MockCombatAnalyticsRepository) GetSessionStats(sessionID int64) (*models.SessionCombatStats, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SessionCombatStats), args.Error(1)
}

func (m *MockCombatAnalyticsRepository) AnalyzeTrends(characterID int64, timeRange time.Duration) (*models.CombatTrends, error) {
	args := m.Called(characterID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CombatTrends), args.Error(1)
}