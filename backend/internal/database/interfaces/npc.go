package interfaces

import (
	"context"
	
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// NPCInterface handles core NPC operations
type NPCInterface interface {
	Create(ctx context.Context, npc *models.NPC) error
	GetByID(ctx context.Context, id string) (*models.NPC, error)
	GetByGameSession(ctx context.Context, gameSessionID string) ([]*models.NPC, error)
	Update(ctx context.Context, npc *models.NPC) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, filter models.NPCSearchFilter) ([]*models.NPC, error)
}

// NPCTemplateInterface manages NPC template operations
type NPCTemplateInterface interface {
	GetTemplates(ctx context.Context) ([]*models.NPCTemplate, error)
	GetTemplateByID(ctx context.Context, id string) (*models.NPCTemplate, error)
	CreateFromTemplate(ctx context.Context, templateID, gameSessionID, createdBy string) (*models.NPC, error)
}

// LegacyNPCRepository maintains backward compatibility
// This interface combines all the focused interfaces
// It will be deprecated once all code is updated to use specific interfaces
type LegacyNPCRepository interface {
	NPCInterface
	NPCTemplateInterface
}