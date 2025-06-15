package interfaces

import (
	"context"
	
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// AIDialogGeneratorInterface handles NPC dialog generation
type AIDialogGeneratorInterface interface {
	GenerateNPCDialog(ctx context.Context, req models.NPCDialogRequest) (string, error)
}

// AILocationGeneratorInterface manages location descriptions
type AILocationGeneratorInterface interface {
	GenerateLocationDescription(ctx context.Context, req models.LocationDescriptionRequest) (*models.AILocation, error)
}

// AICombatNarratorInterface provides combat narration
type AICombatNarratorInterface interface {
	GenerateCombatNarration(ctx context.Context, req models.CombatNarrationRequest) (string, error)
}

// AIStoryGeneratorInterface creates story elements
type AIStoryGeneratorInterface interface {
	GeneratePlotTwist(ctx context.Context, currentContext map[string]interface{}) (*models.AIStoryElement, error)
}

// AIHazardGeneratorInterface creates environmental hazards
type AIHazardGeneratorInterface interface {
	GenerateEnvironmentalHazard(ctx context.Context, locationType string, difficulty int) (*models.AIEnvironmentalHazard, error)
}

// AINPCGeneratorInterface creates NPCs
type AINPCGeneratorInterface interface {
	GenerateNPC(ctx context.Context, role string, context map[string]interface{}) (*models.AINPC, error)
}

// LegacyAIDMAssistantInterface maintains backward compatibility
// This interface combines all the focused interfaces
// It will be deprecated once all code is updated to use specific interfaces
type LegacyAIDMAssistantInterface interface {
	AIDialogGeneratorInterface
	AILocationGeneratorInterface
	AICombatNarratorInterface
	AIStoryGeneratorInterface
	AIHazardGeneratorInterface
	AINPCGeneratorInterface
}