package services

import (
	"context"
	"time"

	"github.com/your-username/dnd-game/backend/internal/auth"
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

// JWTManagerInterface defines the JWT manager contract
type JWTManagerInterface interface {
	GenerateTokenPair(userID, username, email, role string) (*auth.TokenPair, error)
	ValidateToken(tokenString string, expectedType auth.TokenType) (*auth.Claims, error)
	RefreshToken(refreshToken string) (*auth.TokenPair, error)
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

func (e BaseEvent) Type() string         { return e.EventType }
func (e BaseEvent) Timestamp() time.Time { return e.EventTime }
func (e BaseEvent) Data() interface{}    { return e.EventData }

// Common event types
const (
	EventCombatStarted          = "combat.started"
	EventCombatEnded            = "combat.ended"
	EventCharacterLeveled       = "character.leveled"
	EventQuestCompleted         = "quest.completed"
	EventFactionRelationChanged = "faction.relation.changed"
)

// AIRaceGeneratorInterface defines the AI race generation contract
type AIRaceGeneratorInterface interface {
	GenerateCustomRace(ctx context.Context, request models.CustomRaceRequest) (*models.CustomRaceGenerationResult, error)
}

// AICampaignManagerInterface defines the AI campaign management contract
type AICampaignManagerInterface interface {
	GenerateStoryArc(ctx context.Context, req models.GenerateStoryArcRequest) (*models.GeneratedStoryArc, error)
	GenerateSessionRecap(ctx context.Context, memories []*models.SessionMemory) (*models.GeneratedRecap, error)
	GenerateForeshadowing(ctx context.Context, req models.GenerateForeshadowingRequest, plotThread *models.PlotThread, storyArc *models.StoryArc) (*models.GeneratedForeshadowing, error)
}

// AIDMAssistantInterface defines the AI DM assistant contract
type AIDMAssistantInterface interface {
	GenerateNPCDialogue(ctx context.Context, req models.NPCDialogueRequest) (string, error)
	GenerateLocationDescription(ctx context.Context, req models.LocationDescriptionRequest) (*models.AILocation, error)
	GenerateCombatNarration(ctx context.Context, req models.CombatNarrationRequest) (string, error)
	GeneratePlotTwist(ctx context.Context, currentContext map[string]interface{}) (*models.AIStoryElement, error)
	GenerateEnvironmentalHazard(ctx context.Context, locationType string, difficulty int) (*models.AIEnvironmentalHazard, error)
	GenerateNPC(ctx context.Context, role string, context map[string]interface{}) (*models.AINPC, error)
}
