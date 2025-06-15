package interfaces

import (
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/google/uuid"
)

// StoryArcInterface manages story arc operations
type StoryArcInterface interface {
	CreateStoryArc(arc *models.StoryArc) error
	GetStoryArc(id uuid.UUID) (*models.StoryArc, error)
	GetStoryArcsBySession(sessionID uuid.UUID) ([]*models.StoryArc, error)
	UpdateStoryArc(id uuid.UUID, updates map[string]interface{}) error
	DeleteStoryArc(id uuid.UUID) error
}

// SessionMemoryInterface tracks session memories
type SessionMemoryInterface interface {
	CreateSessionMemory(memory *models.SessionMemory) error
	GetSessionMemory(id uuid.UUID) (*models.SessionMemory, error)
	GetSessionMemories(sessionID uuid.UUID, limit int) ([]*models.SessionMemory, error)
	GetLatestSessionMemory(sessionID uuid.UUID) (*models.SessionMemory, error)
	UpdateSessionMemory(id uuid.UUID, updates map[string]interface{}) error
}

// PlotThreadInterface manages plot threads
type PlotThreadInterface interface {
	CreatePlotThread(thread *models.PlotThread) error
	GetPlotThread(id uuid.UUID) (*models.PlotThread, error)
	GetPlotThreadsBySession(sessionID uuid.UUID) ([]*models.PlotThread, error)
	GetActivePlotThreads(sessionID uuid.UUID) ([]*models.PlotThread, error)
	UpdatePlotThread(id uuid.UUID, updates map[string]interface{}) error
	DeletePlotThread(id uuid.UUID) error
}

// ForeshadowingInterface handles foreshadowing elements
type ForeshadowingInterface interface {
	CreateForeshadowingElement(element *models.ForeshadowingElement) error
	GetForeshadowingElement(id uuid.UUID) (*models.ForeshadowingElement, error)
	GetUnrevealedForeshadowing(sessionID uuid.UUID) ([]*models.ForeshadowingElement, error)
	RevealForeshadowing(id uuid.UUID, sessionNumber int) error
}

// CampaignTimelineInterface manages campaign timeline
type CampaignTimelineInterface interface {
	CreateTimelineEvent(event *models.CampaignTimeline) error
	GetTimelineEvents(sessionID uuid.UUID) ([]*models.CampaignTimeline, error)
	UpdateTimelineEvent(id uuid.UUID, updates map[string]interface{}) error
}

// NPCRelationshipInterface tracks NPC relationships
type NPCRelationshipInterface interface {
	CreateNPCRelationship(relationship *models.NPCRelationship) error
	GetNPCRelationships(npcID uuid.UUID) ([]*models.NPCRelationship, error)
	UpdateNPCRelationship(id uuid.UUID, newStatus string) error
}

// LegacyCampaignRepository maintains backward compatibility
// This interface combines all the focused interfaces
// It will be deprecated once all code is updated to use specific interfaces
type LegacyCampaignRepository interface {
	StoryArcInterface
	SessionMemoryInterface
	PlotThreadInterface
	ForeshadowingInterface
	CampaignTimelineInterface
	NPCRelationshipInterface
}
