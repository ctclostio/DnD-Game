package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ctclostio/DnD-Game/backend/internal/database/interfaces"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// CombatAnalyticsService demonstrates the refactored approach using focused interfaces
// Before: Depended on CombatAnalyticsRepository with 46 methods
// After: Depends only on the specific interfaces it needs
type CombatAnalyticsService struct {
	// Only include the interfaces this service actually uses
	analytics  interfaces.CombatAnalyticsInterface
	combatants interfaces.CombatantAnalyticsInterface
	history    interfaces.CombatHistoryInterface
	
	// Other dependencies
	combatService CombatServiceInterface
}

// NewCombatAnalyticsService creates a new service with focused dependencies
func NewCombatAnalyticsService(
	analytics interfaces.CombatAnalyticsInterface,
	combatants interfaces.CombatantAnalyticsInterface,
	history interfaces.CombatHistoryInterface,
	combatService CombatServiceInterface,
) *CombatAnalyticsService {
	return &CombatAnalyticsService{
		analytics:     analytics,
		combatants:    combatants,
		history:       history,
		combatService: combatService,
	}
}

// AnalyzeCombat demonstrates using only the needed interfaces
func (s *CombatAnalyticsService) AnalyzeCombat(ctx context.Context, combatID uuid.UUID) (*models.CombatAnalytics, error) {
	// Get combat state from combat service
	combat, err := s.combatService.GetCombatState(ctx, combatID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get combat state: %w", err)
	}

	// Create analytics using focused interface
	analytics := &models.CombatAnalytics{
		ID:            uuid.New(),
		CombatID:      combatID,
		GameSessionID: uuid.MustParse(combat.GameSessionID),
		StartTime:     combat.StartedAt,
		EndTime:       combat.EndedAt,
		TotalRounds:   combat.Round,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save analytics - only using the analytics interface
	if err := s.analytics.CreateCombatAnalytics(analytics); err != nil {
		return nil, fmt.Errorf("failed to create combat analytics: %w", err)
	}

	// Analyze combatants - only using the combatants interface
	for i := range combat.Combatants {
		combatantAnalytics := s.analyzeCombatant(&combat.Combatants[i], analytics.ID)
		if err := s.combatants.CreateCombatantAnalytics(combatantAnalytics); err != nil {
			// Log error but continue
			continue
		}
	}

	// Create history record - only using the history interface
	history := &models.CombatHistory{
		ID:        uuid.New(),
		CombatID:  combatID,
		SessionID: analytics.GameSessionID,
		Summary:   s.generateSummary(combat, analytics),
		CreatedAt: time.Now(),
	}
	
	if err := s.history.CreateCombatHistory(history); err != nil {
		// Non-critical error, log but continue
	}

	return analytics, nil
}

// GetCombatAnalytics retrieves analytics for a combat
func (s *CombatAnalyticsService) GetCombatAnalytics(ctx context.Context, combatID uuid.UUID) (*models.CombatReport, error) {
	// Get analytics
	analytics, err := s.analytics.GetCombatAnalytics(combatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat analytics: %w", err)
	}

	// Get combatant analytics
	combatantAnalytics, err := s.combatants.GetCombatantAnalytics(analytics.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combatant analytics: %w", err)
	}

	// Get history
	history, err := s.history.GetCombatHistory(combatID)
	if err != nil {
		// History is optional, continue without it
		history = nil
	}

	// Build report
	report := &models.CombatReport{
		Analytics:  analytics,
		Combatants: combatantAnalytics,
		History:    history,
	}

	return report, nil
}

// Helper methods
func (s *CombatAnalyticsService) analyzeCombatant(combatant *models.Combatant, analyticsID uuid.UUID) *models.CombatantAnalytics {
	return &models.CombatantAnalytics{
		ID:                 uuid.New(),
		CombatAnalyticsID:  analyticsID,
		CombatantID:        combatant.ID,
		CombatantName:      combatant.Name,
		CombatantType:      string(combatant.Type),
		DamageDealt:        combatant.DamageDealt,
		DamageTaken:        combatant.MaxHP - combatant.HitPoints,
		HealingDone:        combatant.HealingDone,
		HealingReceived:    combatant.HealingReceived,
		KnockoutCount:      combatant.KnockoutCount,
		CriticalHits:       combatant.CriticalHits,
		CriticalMisses:     combatant.CriticalMisses,
		ActionsPerformed:   combatant.ActionsPerformed,
		SpellsCast:         combatant.SpellsCast,
		SurvivalStatus:     s.determineSurvivalStatus(combatant),
		CreatedAt:          time.Now(),
	}
}

func (s *CombatAnalyticsService) generateSummary(combat *models.Combat, analytics *models.CombatAnalytics) string {
	return fmt.Sprintf(
		"Combat lasted %d rounds with %d participants",
		combat.Round,
		len(combat.Combatants),
	)
}

func (s *CombatAnalyticsService) determineSurvivalStatus(combatant *models.Combatant) string {
	if combatant.HitPoints <= 0 {
		return "defeated"
	}
	if combatant.Fled {
		return "fled"
	}
	return "survived"
}

// Example of how to create the service with a legacy repository
// This allows gradual migration
func NewCombatAnalyticsServiceFromLegacy(
	legacyRepo interfaces.LegacyCombatAnalyticsRepository,
	combatService CombatServiceInterface,
) *CombatAnalyticsService {
	// The legacy repository implements all the focused interfaces
	// So we can use it for each focused dependency
	return &CombatAnalyticsService{
		analytics:     legacyRepo, // Implements CombatAnalyticsInterface
		combatants:    legacyRepo, // Implements CombatantAnalyticsInterface
		history:       legacyRepo, // Implements CombatHistoryInterface
		combatService: combatService,
	}
}

// Example showing a service that only needs battle map functionality
type BattleMapService struct {
	// This service only depends on what it needs
	battleMaps interfaces.BattleMapInterface
}

func NewBattleMapService(battleMaps interfaces.BattleMapInterface) *BattleMapService {
	return &BattleMapService{
		battleMaps: battleMaps,
	}
}

func (s *BattleMapService) CreateBattleMap(ctx context.Context, combatID uuid.UUID, config models.BattleMapConfig) (*models.BattleMap, error) {
	battleMap := &models.BattleMap{
		ID:       uuid.New(),
		CombatID: combatID,
		Width:    config.Width,
		Height:   config.Height,
		TileSize: config.TileSize,
		Terrain:  config.Terrain,
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	// Only uses the battle map interface, not the entire 46-method repository
	if err := s.battleMaps.CreateBattleMap(battleMap); err != nil {
		return nil, fmt.Errorf("failed to create battle map: %w", err)
	}

	return battleMap, nil
}