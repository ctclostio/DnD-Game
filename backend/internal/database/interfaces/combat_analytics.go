package interfaces

import (
	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// CombatAnalyticsInterface handles core combat analytics data
type CombatAnalyticsInterface interface {
	CreateCombatAnalytics(analytics *models.CombatAnalytics) error
	GetCombatAnalytics(combatID uuid.UUID) (*models.CombatAnalytics, error)
	GetCombatAnalyticsBySession(sessionID uuid.UUID) ([]*models.CombatAnalytics, error)
	UpdateCombatAnalytics(id uuid.UUID, updates map[string]interface{}) error
}

// CombatantAnalyticsInterface handles individual combatant performance data
type CombatantAnalyticsInterface interface {
	CreateCombatantAnalytics(analytics *models.CombatantAnalytics) error
	GetCombatantAnalytics(combatAnalyticsID uuid.UUID) ([]*models.CombatantAnalytics, error)
	UpdateCombatantAnalytics(id uuid.UUID, updates map[string]interface{}) error
}

// AutoCombatInterface handles automated combat resolution
type AutoCombatInterface interface {
	CreateAutoCombatResolution(resolution *models.AutoCombatResolution) error
	GetAutoCombatResolution(id uuid.UUID) (*models.AutoCombatResolution, error)
	GetAutoCombatResolutionsBySession(sessionID uuid.UUID) ([]*models.AutoCombatResolution, error)
}

// BattleMapInterface manages combat battle maps
type BattleMapInterface interface {
	CreateBattleMap(battleMap *models.BattleMap) error
	GetBattleMap(id uuid.UUID) (*models.BattleMap, error)
	GetBattleMapByCombat(combatID uuid.UUID) (*models.BattleMap, error)
	GetBattleMapsBySession(sessionID uuid.UUID) ([]*models.BattleMap, error)
	UpdateBattleMap(id uuid.UUID, updates map[string]interface{}) error
}

// InitiativeRuleInterface manages smart initiative rules
type InitiativeRuleInterface interface {
	CreateOrUpdateInitiativeRule(rule *models.SmartInitiativeRule) error
	GetInitiativeRule(sessionID uuid.UUID, entityID string) (*models.SmartInitiativeRule, error)
	GetInitiativeRulesBySession(sessionID uuid.UUID) ([]*models.SmartInitiativeRule, error)
}

// CombatActionLogInterface handles combat action logging
type CombatActionLogInterface interface {
	CreateCombatAction(action *models.CombatActionLog) error
	GetCombatActions(combatID uuid.UUID) ([]*models.CombatActionLog, error)
	GetCombatActionsByRound(combatID uuid.UUID, roundNumber int) ([]*models.CombatActionLog, error)
}

// CombatHistoryInterface manages combat history and summaries
type CombatHistoryInterface interface {
	// NOTE: CombatHistory model does not exist - only CombatSummary
	// CreateCombatHistory(history *models.CombatHistory) error
	// GetCombatHistory(combatID uuid.UUID) (*models.CombatHistory, error)
	// GetCombatHistoriesBySession(sessionID uuid.UUID) ([]*models.CombatHistory, error)
	CreateCombatSummary(summary *models.CombatSummary) error
	GetCombatSummary(combatID uuid.UUID) (*models.CombatSummary, error)
}

// CombatAnimationInterface handles combat animation presets
// NOTE: AnimationPreset model does not exist - commenting out until implemented
// type CombatAnimationInterface interface {
// 	CreateAnimationPreset(preset *models.AnimationPreset) error
// 	GetAnimationPreset(id uuid.UUID) (*models.AnimationPreset, error)
// 	GetAnimationPresets() ([]*models.AnimationPreset, error)
// }

// CombatStrategyInterface manages AI combat strategies
// NOTE: CombatStrategy model does not exist - commenting out until implemented
// type CombatStrategyInterface interface {
// 	CreateCombatStrategy(strategy *models.CombatStrategy) error
// 	GetCombatStrategy(id uuid.UUID) (*models.CombatStrategy, error)
// 	GetCombatStrategiesByType(entityType string) ([]*models.CombatStrategy, error)
// 	UpdateCombatStrategy(id uuid.UUID, updates map[string]interface{}) error
// }

// CombatPredictionInterface handles combat outcome predictions
// NOTE: CombatPrediction model does not exist - commenting out until implemented
// type CombatPredictionInterface interface {
// 	SaveCombatPrediction(prediction *models.CombatPrediction) error
// 	GetCombatPrediction(combatID uuid.UUID) (*models.CombatPrediction, error)
// 	UpdatePredictionResult(combatID uuid.UUID, actualOutcome string, actualDuration int) error
// }

// LegacyCombatAnalyticsRepository maintains backward compatibility
// This interface combines all the focused interfaces
// It will be deprecated once all code is updated to use specific interfaces
type LegacyCombatAnalyticsRepository interface {
	CombatAnalyticsInterface
	CombatantAnalyticsInterface
	AutoCombatInterface
	BattleMapInterface
	InitiativeRuleInterface
	CombatActionLogInterface
	CombatHistoryInterface
	// CombatAnimationInterface // Commented out - model doesn't exist
	// CombatStrategyInterface  // Commented out - model doesn't exist
	// CombatPredictionInterface // Commented out - model doesn't exist
}
