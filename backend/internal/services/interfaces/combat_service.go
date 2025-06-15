package interfaces

import (
	"context"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// CombatInitiationInterface handles combat start/stop operations
type CombatInitiationInterface interface {
	StartCombat(ctx context.Context, sessionID string, participants []models.Combatant) (*models.Combat, error)
	EndCombat(ctx context.Context, combatID string) error
}

// CombatStateInterface manages combat state queries
type CombatStateInterface interface {
	GetCombatState(ctx context.Context, combatID string) (*models.Combat, error)
	SetCombatState(combat *models.Combat)
}

// CombatActionInterface handles combat actions
type CombatActionInterface interface {
	ExecuteAction(ctx context.Context, combatID string, action *models.CombatAction) (*models.Combat, error)
}

// CombatDamageInterface manages damage and healing
type CombatDamageInterface interface {
	ApplyDamage(ctx context.Context, combatID, targetID string, damage int, damageType string) (*models.Combat, error)
	ApplyHealing(ctx context.Context, combatID, targetID string, healing int) (*models.Combat, error)
}

// DeathSaveInterface handles death saving throws
type DeathSaveInterface interface {
	DeathSavingThrow(ctx context.Context, combatID, characterID string) (*models.Combat, *models.DeathSaveResult, error)
}

// LegacyCombatServiceInterface maintains backward compatibility
// This interface combines all the focused interfaces
// It will be deprecated once all code is updated to use specific interfaces
type LegacyCombatServiceInterface interface {
	CombatInitiationInterface
	CombatStateInterface
	CombatActionInterface
	CombatDamageInterface
	DeathSaveInterface
}
