package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/models"
)

type CampaignService struct {
	campaignRepo   database.CampaignRepository
	gameRepo       database.GameSessionRepository
	aiManager      AICampaignManagerInterface
}

func NewCampaignService(
	campaignRepo database.CampaignRepository,
	gameRepo database.GameSessionRepository,
	aiManager AICampaignManagerInterface,
) *CampaignService {
	return &CampaignService{
		campaignRepo: campaignRepo,
		gameRepo:     gameRepo,
		aiManager:    aiManager,
	}
}

// Story Arc Management

func (cs *CampaignService) CreateStoryArc(ctx context.Context, sessionID uuid.UUID, req models.CreateStoryArcRequest) (*models.StoryArc, error) {
	arc := &models.StoryArc{
		ID:              uuid.New(),
		GameSessionID:   sessionID,
		Title:           req.Title,
		Description:     req.Description,
		ArcType:         req.ArcType,
		Status:          "active",
		ParentArcID:     req.ParentArcID,
		ImportanceLevel: req.ImportanceLevel,
		Metadata:        models.JSONB(`{}`),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if arc.ImportanceLevel == 0 {
		arc.ImportanceLevel = 5
	}

	if err := cs.campaignRepo.CreateStoryArc(arc); err != nil {
		return nil, fmt.Errorf("failed to create story arc: %w", err)
	}

	return arc, nil
}

func (cs *CampaignService) GenerateStoryArc(ctx context.Context, sessionID uuid.UUID, req models.GenerateStoryArcRequest) (*models.StoryArc, error) {
	// Generate the arc using AI
	generated, err := cs.aiManager.GenerateStoryArc(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate story arc: %w", err)
	}

	// Convert generated arc to database model
	metadata, _ := json.Marshal(generated)
	arc := &models.StoryArc{
		ID:              uuid.New(),
		GameSessionID:   sessionID,
		Title:           generated.Title,
		Description:     generated.Description,
		ArcType:         generated.ArcType,
		Status:          "active",
		ImportanceLevel: generated.ImportanceLevel,
		Metadata:        models.JSONB(metadata),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := cs.campaignRepo.CreateStoryArc(arc); err != nil {
		return nil, fmt.Errorf("failed to save generated story arc: %w", err)
	}

	return arc, nil
}

func (cs *CampaignService) GetStoryArcs(ctx context.Context, sessionID uuid.UUID) ([]*models.StoryArc, error) {
	return cs.campaignRepo.GetStoryArcsBySession(sessionID)
}

func (cs *CampaignService) UpdateStoryArc(ctx context.Context, arcID uuid.UUID, req models.UpdateStoryArcRequest) error {
	updates := make(map[string]interface{})
	
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status == "completed" {
			now := time.Now()
			updates["resolved_at"] = now
		}
	}
	if req.ImportanceLevel != nil {
		updates["importance_level"] = *req.ImportanceLevel
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	return cs.campaignRepo.UpdateStoryArc(arcID, updates)
}

// Session Memory Management

func (cs *CampaignService) CreateSessionMemory(ctx context.Context, sessionID uuid.UUID, req models.CreateSessionMemoryRequest) (*models.SessionMemory, error) {
	// Convert request data to JSONB
	keyEventsJSON, _ := json.Marshal(req.KeyEvents)
	npcsJSON, _ := json.Marshal(req.NPCsEncountered)
	decisionsJSON, _ := json.Marshal(req.DecisionsMade)
	itemsJSON, _ := json.Marshal(req.ItemsAcquired)
	locationsJSON, _ := json.Marshal(req.LocationsVisited)

	memory := &models.SessionMemory{
		ID:               uuid.New(),
		GameSessionID:    sessionID,
		SessionNumber:    req.SessionNumber,
		SessionDate:      req.SessionDate,
		KeyEvents:        models.JSONB(keyEventsJSON),
		NPCsEncountered:  models.JSONB(npcsJSON),
		DecisionsMade:    models.JSONB(decisionsJSON),
		ItemsAcquired:    models.JSONB(itemsJSON),
		LocationsVisited: models.JSONB(locationsJSON),
		CombatEncounters: models.JSONB(`[]`),
		PlotDevelopments: models.JSONB(`[]`),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Generate AI recap if we have the data
	if len(req.KeyEvents) > 0 {
		recap := cs.generateRecapFromEvents(req)
		memory.RecapSummary = recap
	}

	if err := cs.campaignRepo.CreateSessionMemory(memory); err != nil {
		return nil, fmt.Errorf("failed to create session memory: %w", err)
	}

	return memory, nil
}

func (cs *CampaignService) GetSessionMemories(ctx context.Context, sessionID uuid.UUID, limit int) ([]*models.SessionMemory, error) {
	if limit <= 0 {
		limit = 10
	}
	return cs.campaignRepo.GetSessionMemories(sessionID, limit)
}

func (cs *CampaignService) GenerateRecap(ctx context.Context, sessionID uuid.UUID, sessionCount int) (*models.GeneratedRecap, error) {
	if sessionCount <= 0 {
		sessionCount = 3
	}

	// Get recent session memories
	memories, err := cs.campaignRepo.GetSessionMemories(sessionID, sessionCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get session memories: %w", err)
	}

	if len(memories) == 0 {
		return &models.GeneratedRecap{
			Summary: "This is the beginning of your adventure...",
			KeyEvents: []string{"The party gathers for the first time"},
		}, nil
	}

	// Generate recap using AI
	return cs.aiManager.GenerateSessionRecap(ctx, memories)
}

// Plot Thread Management

func (cs *CampaignService) CreatePlotThread(ctx context.Context, sessionID uuid.UUID, thread *models.PlotThread) error {
	thread.ID = uuid.New()
	thread.GameSessionID = sessionID
	thread.CreatedAt = time.Now()
	thread.UpdatedAt = time.Now()

	if thread.Status == "" {
		thread.Status = "active"
	}
	if thread.TensionLevel == 0 {
		thread.TensionLevel = 5
	}

	return cs.campaignRepo.CreatePlotThread(thread)
}

func (cs *CampaignService) GetPlotThreads(ctx context.Context, sessionID uuid.UUID, activeOnly bool) ([]*models.PlotThread, error) {
	if activeOnly {
		return cs.campaignRepo.GetActivePlotThreads(sessionID)
	}
	return cs.campaignRepo.GetPlotThreadsBySession(sessionID)
}

func (cs *CampaignService) UpdatePlotThread(ctx context.Context, threadID uuid.UUID, updates map[string]interface{}) error {
	return cs.campaignRepo.UpdatePlotThread(threadID, updates)
}

// Foreshadowing Management

func (cs *CampaignService) GenerateForeshadowing(ctx context.Context, sessionID uuid.UUID, req models.GenerateForeshadowingRequest) (*models.ForeshadowingElement, error) {
	var plotThread *models.PlotThread
	var storyArc *models.StoryArc
	var err error

	// Get the associated plot thread or story arc for context
	if req.PlotThreadID != nil {
		plotThread, err = cs.campaignRepo.GetPlotThread(*req.PlotThreadID)
		if err != nil {
			return nil, fmt.Errorf("failed to get plot thread: %w", err)
		}
	}
	if req.StoryArcID != nil {
		storyArc, err = cs.campaignRepo.GetStoryArc(*req.StoryArcID)
		if err != nil {
			return nil, fmt.Errorf("failed to get story arc: %w", err)
		}
	}

	// Generate foreshadowing using AI
	generated, err := cs.aiManager.GenerateForeshadowing(ctx, req, plotThread, storyArc)
	if err != nil {
		return nil, fmt.Errorf("failed to generate foreshadowing: %w", err)
	}

	// Create database entry
	placementJSON, _ := json.Marshal(generated.PlacementSuggestions)
	element := &models.ForeshadowingElement{
		ID:                   uuid.New(),
		GameSessionID:        sessionID,
		PlotThreadID:         req.PlotThreadID,
		StoryArcID:           req.StoryArcID,
		ElementType:          generated.ElementType,
		Content:              generated.Content,
		SubtletyLevel:        generated.SubtletyLevel,
		Revealed:             false,
		PlacementSuggestions: models.JSONB(placementJSON),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := cs.campaignRepo.CreateForeshadowingElement(element); err != nil {
		return nil, fmt.Errorf("failed to save foreshadowing element: %w", err)
	}

	return element, nil
}

func (cs *CampaignService) GetUnrevealedForeshadowing(ctx context.Context, sessionID uuid.UUID) ([]*models.ForeshadowingElement, error) {
	return cs.campaignRepo.GetUnrevealedForeshadowing(sessionID)
}

func (cs *CampaignService) RevealForeshadowing(ctx context.Context, elementID uuid.UUID, sessionNumber int) error {
	return cs.campaignRepo.RevealForeshadowing(elementID, sessionNumber)
}

// Timeline Management

func (cs *CampaignService) AddTimelineEvent(ctx context.Context, event *models.CampaignTimeline) error {
	event.ID = uuid.New()
	event.CreatedAt = time.Now()
	
	if event.ImpactLevel == 0 {
		event.ImpactLevel = 5
	}

	return cs.campaignRepo.CreateTimelineEvent(event)
}

func (cs *CampaignService) GetTimeline(ctx context.Context, sessionID uuid.UUID, startDate, endDate time.Time) ([]*models.CampaignTimeline, error) {
	return cs.campaignRepo.GetTimelineEvents(sessionID, startDate, endDate)
}

// NPC Relationship Management

func (cs *CampaignService) UpdateNPCRelationship(ctx context.Context, relationship *models.NPCRelationship) error {
	relationship.ID = uuid.New()
	relationship.CreatedAt = time.Now()
	relationship.UpdatedAt = time.Now()
	
	return cs.campaignRepo.CreateOrUpdateNPCRelationship(relationship)
}

func (cs *CampaignService) GetNPCRelationships(ctx context.Context, sessionID, npcID uuid.UUID) ([]*models.NPCRelationship, error) {
	return cs.campaignRepo.GetNPCRelationships(sessionID, npcID)
}

func (cs *CampaignService) AdjustRelationshipScore(ctx context.Context, sessionID, npcID, targetID uuid.UUID, scoreDelta int) error {
	return cs.campaignRepo.UpdateRelationshipScore(sessionID, npcID, targetID, scoreDelta)
}

// Helper methods

func (cs *CampaignService) generateRecapFromEvents(req models.CreateSessionMemoryRequest) string {
	recap := fmt.Sprintf("In session %d, the party ", req.SessionNumber)
	
	if len(req.KeyEvents) > 0 {
		recap += "experienced significant events. "
	}
	
	if len(req.NPCsEncountered) > 0 {
		recap += fmt.Sprintf("They encountered %d NPCs. ", len(req.NPCsEncountered))
	}
	
	if len(req.DecisionsMade) > 0 {
		recap += "Important decisions were made that will shape future events. "
	}
	
	if len(req.ItemsAcquired) > 0 {
		recap += fmt.Sprintf("The party acquired %d new items. ", len(req.ItemsAcquired))
	}
	
	if len(req.LocationsVisited) > 0 {
		recap += fmt.Sprintf("They explored %d locations. ", len(req.LocationsVisited))
	}
	
	return recap
}