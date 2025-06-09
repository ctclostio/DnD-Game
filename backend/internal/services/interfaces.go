package services

import (
	"context"
	"time"
	
	"github.com/your-username/dnd-game/backend/internal/models"
)

// Service interfaces to prevent circular dependencies
// These interfaces define contracts that services can depend on without creating circular imports

// CombatServiceInterface defines the combat service contract
type CombatServiceInterface interface {
	StartCombat(ctx context.Context, sessionID string, participants []models.Combatant) (*models.Combat, error)
	GetCombatState(ctx context.Context, combatID string) (*models.Combat, error)
	ExecuteAction(ctx context.Context, combatID string, action models.CombatAction) (*models.Combat, error)
	EndCombat(ctx context.Context, combatID string) error
	ApplyDamage(ctx context.Context, combatID, targetID string, damage int, damageType string) (*models.Combat, error)
	ApplyHealing(ctx context.Context, combatID, targetID string, healing int) (*models.Combat, error)
	DeathSavingThrow(ctx context.Context, combatID, characterID string) (*models.Combat, *models.DeathSaveResult, error)
	SetCombatState(combat *models.Combat)
}

// RuleEngineInterface defines the rule engine contract
type RuleEngineInterface interface {
	EvaluateRule(ctx context.Context, rule *models.Rule, context map[string]interface{}) (bool, error)
	ExecuteAction(ctx context.Context, action *models.RuleAction, context map[string]interface{}) error
	ValidateRule(rule *models.Rule) error
}

// DiceRollServiceInterface defines the dice rolling contract
type DiceRollServiceInterface interface {
	Roll(notation string) (*models.DiceRollResult, error)
	RollWithAdvantage(notation string, advantage bool) (*models.DiceRollResult, error)
}

// FactionSystemInterface defines the faction system contract
type FactionSystemInterface interface {
	GetFaction(ctx context.Context, factionID string) (*models.Faction, error)
	UpdateFactionRelationship(ctx context.Context, faction1ID, faction2ID string, change float64) error
	GetFactionRelationship(ctx context.Context, faction1ID, faction2ID string) (float64, error)
}

// EventBus for decoupled communication between services
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType string, handler EventHandler) error
}

// Event represents a domain event
type Event interface {
	Type() string
	Timestamp() time.Time
	Data() interface{}
}

// EventHandler processes events
type EventHandler func(ctx context.Context, event Event) error

// BaseEvent provides common event functionality
type BaseEvent struct {
	EventType string      `json:"type"`
	EventTime time.Time   `json:"timestamp"`
	EventData interface{} `json:"data"`
}

func (e BaseEvent) Type() string      { return e.EventType }
func (e BaseEvent) Timestamp() time.Time { return e.EventTime }
func (e BaseEvent) Data() interface{} { return e.EventData }

// Common event types
const (
	EventCombatStarted   = "combat.started"
	EventCombatEnded     = "combat.ended"
	EventCharacterLeveled = "character.leveled"
	EventQuestCompleted  = "quest.completed"
	EventFactionRelationChanged = "faction.relation.changed"
)