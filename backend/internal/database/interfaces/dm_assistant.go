package interfaces

import (
	"context"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/google/uuid"
)

// DMAssistantNPCInterface manages NPC operations for the DM assistant
type DMAssistantNPCInterface interface {
	SaveNPC(ctx context.Context, npc *models.AINPC) error
	GetNPCByID(ctx context.Context, id uuid.UUID) (*models.AINPC, error)
	GetNPCsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AINPC, error)
	UpdateNPC(ctx context.Context, npc *models.AINPC) error
	AddNPCDialog(ctx context.Context, npcID uuid.UUID, dialog models.DialogEntry) error
}

// DMAssistantLocationInterface manages location operations
type DMAssistantLocationInterface interface {
	SaveLocation(ctx context.Context, location *models.AILocation) error
	GetLocationByID(ctx context.Context, id uuid.UUID) (*models.AILocation, error)
	GetLocationsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AILocation, error)
	UpdateLocation(ctx context.Context, location *models.AILocation) error
}

// DMAssistantStoryInterface handles story and narration elements
type DMAssistantStoryInterface interface {
	SaveNarration(ctx context.Context, narration *models.AINarration) error
	GetNarrationsByType(ctx context.Context, sessionID uuid.UUID, narrationType string) ([]*models.AINarration, error)
	SaveStoryElement(ctx context.Context, element *models.AIStoryElement) error
	GetUnusedStoryElements(ctx context.Context, sessionID uuid.UUID) ([]*models.AIStoryElement, error)
	MarkStoryElementUsed(ctx context.Context, elementID uuid.UUID) error
}

// DMAssistantHazardInterface manages environmental hazards
type DMAssistantHazardInterface interface {
	SaveEnvironmentalHazard(ctx context.Context, hazard *models.AIEnvironmentalHazard) error
	GetActiveHazardsByLocation(ctx context.Context, locationID uuid.UUID) ([]*models.AIEnvironmentalHazard, error)
	TriggerHazard(ctx context.Context, hazardID uuid.UUID) error
}

// DMAssistantHistoryInterface tracks DM assistant usage
type DMAssistantHistoryInterface interface {
	RecordAssistantRequest(ctx context.Context, request *models.DMAssistantRequest) error
	GetAssistantHistory(ctx context.Context, sessionID uuid.UUID, limit int) ([]*models.DMAssistantRequest, error)
}

// LegacyDMAssistantRepository maintains backward compatibility
// This interface combines all the focused interfaces
// It will be deprecated once all code is updated to use specific interfaces
type LegacyDMAssistantRepository interface {
	DMAssistantNPCInterface
	DMAssistantLocationInterface
	DMAssistantStoryInterface
	DMAssistantHazardInterface
	DMAssistantHistoryInterface
}
