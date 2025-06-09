package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// Mock implementations for testing
type MockCampaignRepository struct {
	mock.Mock
}

func (m *MockCampaignRepository) CreateStoryArc(arc *models.StoryArc) error {
	args := m.Called(arc)
	return args.Error(0)
}

func (m *MockCampaignRepository) GetStoryArc(id uuid.UUID) (*models.StoryArc, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.StoryArc), args.Error(1)
}

func (m *MockCampaignRepository) GetStoryArcsBySession(sessionID uuid.UUID) ([]*models.StoryArc, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.StoryArc), args.Error(1)
}

func (m *MockCampaignRepository) UpdateStoryArc(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockCampaignRepository) DeleteStoryArc(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCampaignRepository) CreateSessionMemory(memory *models.SessionMemory) error {
	args := m.Called(memory)
	return args.Error(0)
}

func (m *MockCampaignRepository) GetSessionMemory(id uuid.UUID) (*models.SessionMemory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SessionMemory), args.Error(1)
}

func (m *MockCampaignRepository) GetSessionMemories(sessionID uuid.UUID, limit int) ([]*models.SessionMemory, error) {
	args := m.Called(sessionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SessionMemory), args.Error(1)
}

func (m *MockCampaignRepository) GetLatestSessionMemory(sessionID uuid.UUID) (*models.SessionMemory, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SessionMemory), args.Error(1)
}

func (m *MockCampaignRepository) UpdateSessionMemory(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockCampaignRepository) CreatePlotThread(thread *models.PlotThread) error {
	args := m.Called(thread)
	return args.Error(0)
}

func (m *MockCampaignRepository) GetPlotThread(id uuid.UUID) (*models.PlotThread, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlotThread), args.Error(1)
}

func (m *MockCampaignRepository) GetPlotThreadsBySession(sessionID uuid.UUID) ([]*models.PlotThread, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PlotThread), args.Error(1)
}

func (m *MockCampaignRepository) GetActivePlotThreads(sessionID uuid.UUID) ([]*models.PlotThread, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PlotThread), args.Error(1)
}

func (m *MockCampaignRepository) UpdatePlotThread(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockCampaignRepository) DeletePlotThread(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCampaignRepository) CreateForeshadowingElement(element *models.ForeshadowingElement) error {
	args := m.Called(element)
	return args.Error(0)
}

func (m *MockCampaignRepository) GetForeshadowingElement(id uuid.UUID) (*models.ForeshadowingElement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ForeshadowingElement), args.Error(1)
}

func (m *MockCampaignRepository) GetUnrevealedForeshadowing(sessionID uuid.UUID) ([]*models.ForeshadowingElement, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ForeshadowingElement), args.Error(1)
}

func (m *MockCampaignRepository) RevealForeshadowing(id uuid.UUID, sessionNumber int) error {
	args := m.Called(id, sessionNumber)
	return args.Error(0)
}

func (m *MockCampaignRepository) CreateTimelineEvent(event *models.CampaignTimeline) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockCampaignRepository) GetTimelineEvents(sessionID uuid.UUID, startDate, endDate time.Time) ([]*models.CampaignTimeline, error) {
	args := m.Called(sessionID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CampaignTimeline), args.Error(1)
}

func (m *MockCampaignRepository) CreateOrUpdateNPCRelationship(relationship *models.NPCRelationship) error {
	args := m.Called(relationship)
	return args.Error(0)
}

func (m *MockCampaignRepository) GetNPCRelationships(sessionID uuid.UUID, npcID uuid.UUID) ([]*models.NPCRelationship, error) {
	args := m.Called(sessionID, npcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NPCRelationship), args.Error(1)
}

func (m *MockCampaignRepository) UpdateRelationshipScore(sessionID, npcID, targetID uuid.UUID, scoreDelta int) error {
	args := m.Called(sessionID, npcID, targetID, scoreDelta)
	return args.Error(0)
}

// Mock Game Session Repository
type MockGameSessionRepository struct {
	mock.Mock
}

// Mock AI Campaign Manager
type MockAICampaignManager struct {
	mock.Mock
}

func (m *MockAICampaignManager) GenerateStoryArc(ctx context.Context, req models.GenerateStoryArcRequest) (*models.GeneratedStoryArc, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GeneratedStoryArc), args.Error(1)
}

func (m *MockAICampaignManager) GenerateSessionRecap(ctx context.Context, memories []*models.SessionMemory) (*models.GeneratedRecap, error) {
	args := m.Called(ctx, memories)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GeneratedRecap), args.Error(1)
}

func (m *MockAICampaignManager) GenerateForeshadowing(ctx context.Context, req models.GenerateForeshadowingRequest, plotThread *models.PlotThread, storyArc *models.StoryArc) (*models.GeneratedForeshadowing, error) {
	args := m.Called(ctx, req, plotThread, storyArc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GeneratedForeshadowing), args.Error(1)
}

// Test helpers
func createTestCampaignService(repo *MockCampaignRepository, gameRepo *MockGameSessionRepository, aiManager *MockAICampaignManager) *CampaignService {
	return NewCampaignService(repo, gameRepo, aiManager)
}

// Tests for Story Arc Management

func TestCampaignService_CreateStoryArc(t *testing.T) {
	tests := []struct {
		name        string
		sessionID   uuid.UUID
		request     models.CreateStoryArcRequest
		setupMocks  func(*MockCampaignRepository)
		expectError bool
		validate    func(*testing.T, *models.StoryArc)
	}{
		{
			name:      "Successful Story Arc Creation",
			sessionID: uuid.New(),
			request: models.CreateStoryArcRequest{
				Title:           "The Lost Mines",
				Description:     "Adventurers seek a legendary mine",
				ArcType:         "main",
				ImportanceLevel: 8,
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("CreateStoryArc", mock.AnythingOfType("*models.StoryArc")).Return(nil)
			},
			expectError: false,
			validate: func(t *testing.T, arc *models.StoryArc) {
				assert.Equal(t, "The Lost Mines", arc.Title)
				assert.Equal(t, "main", arc.ArcType)
				assert.Equal(t, 8, arc.ImportanceLevel)
				assert.Equal(t, "active", arc.Status)
			},
		},
		{
			name:      "Default Importance Level",
			sessionID: uuid.New(),
			request: models.CreateStoryArcRequest{
				Title:       "Side Quest",
				Description: "A minor adventure",
				ArcType:     "side",
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("CreateStoryArc", mock.AnythingOfType("*models.StoryArc")).Return(nil)
			},
			expectError: false,
			validate: func(t *testing.T, arc *models.StoryArc) {
				assert.Equal(t, 5, arc.ImportanceLevel) // Default value
			},
		},
		{
			name:      "Repository Error",
			sessionID: uuid.New(),
			request: models.CreateStoryArcRequest{
				Title: "Failed Arc",
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("CreateStoryArc", mock.AnythingOfType("*models.StoryArc")).
					Return(errors.New("database error"))
			},
			expectError: true,
		},
		{
			name:      "With Parent Arc",
			sessionID: uuid.New(),
			request: models.CreateStoryArcRequest{
				Title:       "Chapter 2",
				Description: "Continuation of the main quest",
				ArcType:     "main",
				ParentArcID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("CreateStoryArc", mock.AnythingOfType("*models.StoryArc")).Return(nil)
			},
			expectError: false,
			validate: func(t *testing.T, arc *models.StoryArc) {
				assert.NotNil(t, arc.ParentArcID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCampaignRepository)
			mockGameRepo := new(MockGameSessionRepository)
			mockAI := new(MockAICampaignManager)
			
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			service := createTestCampaignService(mockRepo, mockGameRepo, mockAI)
			arc, err := service.CreateStoryArc(context.Background(), tt.sessionID, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, arc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, arc)
				assert.NotEqual(t, uuid.Nil, arc.ID)
				assert.Equal(t, tt.sessionID, arc.GameSessionID)
				if tt.validate != nil {
					tt.validate(t, arc)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignService_GenerateStoryArc(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     uuid.UUID
		request       models.GenerateStoryArcRequest
		generatedArc  *models.GeneratedStoryArc
		generateError error
		repoError     error
		expectError   bool
	}{
		{
			name:      "Successful Generation",
			sessionID: uuid.New(),
			request: models.GenerateStoryArcRequest{
				Theme:      "Dark Fantasy",
				Length:     "Medium",
				Complexity: "High",
			},
			generatedArc: &models.GeneratedStoryArc{
				Title:           "The Shadow's Curse",
				Description:     "A dark curse spreads across the land",
				ArcType:         "main",
				ImportanceLevel: 9,
			},
			expectError: false,
		},
		{
			name:      "AI Generation Error",
			sessionID: uuid.New(),
			request: models.GenerateStoryArcRequest{
				Theme: "Epic",
			},
			generateError: errors.New("AI service unavailable"),
			expectError:   true,
		},
		{
			name:      "Repository Save Error",
			sessionID: uuid.New(),
			request:   models.GenerateStoryArcRequest{},
			generatedArc: &models.GeneratedStoryArc{
				Title: "Generated Arc",
			},
			repoError:   errors.New("database error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCampaignRepository)
			mockGameRepo := new(MockGameSessionRepository)
			mockAI := new(MockAICampaignManager)

			mockAI.On("GenerateStoryArc", mock.Anything, tt.request).
				Return(tt.generatedArc, tt.generateError)

			if tt.generatedArc != nil && tt.generateError == nil {
				mockRepo.On("CreateStoryArc", mock.AnythingOfType("*models.StoryArc")).
					Return(tt.repoError)
			}

			service := createTestCampaignService(mockRepo, mockGameRepo, mockAI)
			arc, err := service.GenerateStoryArc(context.Background(), tt.sessionID, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, arc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, arc)
				assert.Equal(t, tt.generatedArc.Title, arc.Title)
				assert.Equal(t, tt.generatedArc.Description, arc.Description)
			}

			mockRepo.AssertExpectations(t)
			mockAI.AssertExpectations(t)
		})
	}
}

func TestCampaignService_GetStoryArcs(t *testing.T) {
	sessionID := uuid.New()
	mockArcs := []*models.StoryArc{
		{
			ID:    uuid.New(),
			Title: "Arc 1",
		},
		{
			ID:    uuid.New(),
			Title: "Arc 2",
		},
	}

	tests := []struct {
		name        string
		setupMocks  func(*MockCampaignRepository)
		expectError bool
		expectCount int
	}{
		{
			name: "Successful Retrieval",
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("GetStoryArcsBySession", sessionID).Return(mockArcs, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name: "Repository Error",
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("GetStoryArcsBySession", sessionID).Return(nil, errors.New("database error"))
			},
			expectError: true,
			expectCount: 0,
		},
		{
			name: "Empty Result",
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("GetStoryArcsBySession", sessionID).Return([]*models.StoryArc{}, nil)
			},
			expectError: false,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCampaignRepository)
			mockGameRepo := new(MockGameSessionRepository)
			mockAI := new(MockAICampaignManager)

			tt.setupMocks(mockRepo)

			service := createTestCampaignService(mockRepo, mockGameRepo, mockAI)
			arcs, err := service.GetStoryArcs(context.Background(), sessionID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, arcs, tt.expectCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignService_UpdateStoryArc(t *testing.T) {
	arcID := uuid.New()
	completedStatus := "completed"
	newTitle := "Updated Title"
	newImportance := 10

	tests := []struct {
		name        string
		request     models.UpdateStoryArcRequest
		setupMocks  func(*MockCampaignRepository)
		expectError bool
		validate    func(*testing.T, map[string]interface{})
	}{
		{
			name: "Update Multiple Fields",
			request: models.UpdateStoryArcRequest{
				Title:           &newTitle,
				Status:          &completedStatus,
				ImportanceLevel: &newImportance,
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("UpdateStoryArc", arcID, mock.MatchedBy(func(updates map[string]interface{}) bool {
					return updates["title"] == newTitle &&
						updates["status"] == completedStatus &&
						updates["importance_level"] == newImportance &&
						updates["resolved_at"] != nil
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name: "Update Title Only",
			request: models.UpdateStoryArcRequest{
				Title: &newTitle,
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("UpdateStoryArc", arcID, mock.MatchedBy(func(updates map[string]interface{}) bool {
					return updates["title"] == newTitle && len(updates) == 1
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name: "Update Status to Completed",
			request: models.UpdateStoryArcRequest{
				Status: &completedStatus,
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("UpdateStoryArc", arcID, mock.MatchedBy(func(updates map[string]interface{}) bool {
					_, hasResolvedAt := updates["resolved_at"]
					return updates["status"] == completedStatus && hasResolvedAt
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:    "Repository Error",
			request: models.UpdateStoryArcRequest{Title: &newTitle},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("UpdateStoryArc", arcID, mock.Anything).Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCampaignRepository)
			mockGameRepo := new(MockGameSessionRepository)
			mockAI := new(MockAICampaignManager)

			tt.setupMocks(mockRepo)

			service := createTestCampaignService(mockRepo, mockGameRepo, mockAI)
			err := service.UpdateStoryArc(context.Background(), arcID, tt.request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Tests for Session Memory Management

func TestCampaignService_CreateSessionMemory(t *testing.T) {
	sessionID := uuid.New()
	sessionDate := time.Now()

	tests := []struct {
		name        string
		request     models.CreateSessionMemoryRequest
		setupMocks  func(*MockCampaignRepository)
		expectError bool
		validate    func(*testing.T, *models.SessionMemory)
	}{
		{
			name: "Complete Session Memory",
			request: models.CreateSessionMemoryRequest{
				SessionNumber: 5,
				SessionDate:   sessionDate,
				KeyEvents: []string{
					"Party defeated the goblin king",
					"Found the magical artifact",
				},
				NPCsEncountered: []models.NPCEncounter{
					{NPCID: uuid.New(), Name: "Gandor the Wise"},
				},
				DecisionsMade: []models.Decision{
					{Description: "Chose to spare the goblin king"},
				},
				ItemsAcquired: []models.ItemAcquired{
					{ItemID: uuid.New(), Name: "Sword of Truth"},
				},
				LocationsVisited: []string{"Goblin Cave", "Ancient Temple"},
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("CreateSessionMemory", mock.AnythingOfType("*models.SessionMemory")).Return(nil)
			},
			expectError: false,
			validate: func(t *testing.T, memory *models.SessionMemory) {
				assert.Equal(t, 5, memory.SessionNumber)
				assert.Equal(t, sessionID, memory.GameSessionID)
				assert.NotEmpty(t, memory.RecapSummary) // Should generate recap
			},
		},
		{
			name: "Empty Events - No Recap",
			request: models.CreateSessionMemoryRequest{
				SessionNumber: 1,
				SessionDate:   sessionDate,
				KeyEvents:     []string{},
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("CreateSessionMemory", mock.AnythingOfType("*models.SessionMemory")).Return(nil)
			},
			expectError: false,
			validate: func(t *testing.T, memory *models.SessionMemory) {
				assert.Empty(t, memory.RecapSummary) // No recap for empty events
			},
		},
		{
			name: "Repository Error",
			request: models.CreateSessionMemoryRequest{
				SessionNumber: 1,
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("CreateSessionMemory", mock.AnythingOfType("*models.SessionMemory")).
					Return(errors.New("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCampaignRepository)
			mockGameRepo := new(MockGameSessionRepository)
			mockAI := new(MockAICampaignManager)

			tt.setupMocks(mockRepo)

			service := createTestCampaignService(mockRepo, mockGameRepo, mockAI)
			memory, err := service.CreateSessionMemory(context.Background(), sessionID, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, memory)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, memory)
				assert.NotEqual(t, uuid.Nil, memory.ID)
				if tt.validate != nil {
					tt.validate(t, memory)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignService_GetSessionMemories(t *testing.T) {
	sessionID := uuid.New()
	mockMemories := []*models.SessionMemory{
		{ID: uuid.New(), SessionNumber: 3},
		{ID: uuid.New(), SessionNumber: 2},
		{ID: uuid.New(), SessionNumber: 1},
	}

	tests := []struct {
		name         string
		limit        int
		expectedCall int
		setupMocks   func(*MockCampaignRepository)
		expectError  bool
		expectCount  int
	}{
		{
			name:         "With Specified Limit",
			limit:        5,
			expectedCall: 5,
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("GetSessionMemories", sessionID, 5).Return(mockMemories[:2], nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:         "Default Limit",
			limit:        0,
			expectedCall: 10, // Default
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("GetSessionMemories", sessionID, 10).Return(mockMemories, nil)
			},
			expectError: false,
			expectCount: 3,
		},
		{
			name:         "Negative Limit Uses Default",
			limit:        -5,
			expectedCall: 10,
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("GetSessionMemories", sessionID, 10).Return(mockMemories, nil)
			},
			expectError: false,
			expectCount: 3,
		},
		{
			name:         "Repository Error",
			limit:        10,
			expectedCall: 10,
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("GetSessionMemories", sessionID, 10).Return(nil, errors.New("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCampaignRepository)
			mockGameRepo := new(MockGameSessionRepository)
			mockAI := new(MockAICampaignManager)

			tt.setupMocks(mockRepo)

			service := createTestCampaignService(mockRepo, mockGameRepo, mockAI)
			memories, err := service.GetSessionMemories(context.Background(), sessionID, tt.limit)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, memories, tt.expectCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignService_GenerateRecap(t *testing.T) {
	sessionID := uuid.New()
	mockMemories := []*models.SessionMemory{
		{
			ID:            uuid.New(),
			SessionNumber: 3,
			KeyEvents:     models.JSONB(`["Event 1", "Event 2"]`),
		},
		{
			ID:            uuid.New(),
			SessionNumber: 2,
			KeyEvents:     models.JSONB(`["Event 3"]`),
		},
	}

	tests := []struct {
		name         string
		sessionCount int
		memories     []*models.SessionMemory
		memoryError  error
		aiRecap      *models.GeneratedRecap
		aiError      error
		expectError  bool
		validate     func(*testing.T, *models.GeneratedRecap)
	}{
		{
			name:         "Successful AI Recap Generation",
			sessionCount: 3,
			memories:     mockMemories,
			aiRecap: &models.GeneratedRecap{
				Summary:   "The party continues their epic quest...",
				KeyEvents: []string{"Defeated goblin king", "Found artifact"},
			},
			expectError: false,
			validate: func(t *testing.T, recap *models.GeneratedRecap) {
				assert.Equal(t, "The party continues their epic quest...", recap.Summary)
				assert.Len(t, recap.KeyEvents, 2)
			},
		},
		{
			name:         "No Memories - Default Recap",
			sessionCount: 3,
			memories:     []*models.SessionMemory{},
			expectError:  false,
			validate: func(t *testing.T, recap *models.GeneratedRecap) {
				assert.Equal(t, "This is the beginning of your adventure...", recap.Summary)
				assert.Contains(t, recap.KeyEvents, "The party gathers for the first time")
			},
		},
		{
			name:         "Default Session Count",
			sessionCount: 0,
			memories:     mockMemories,
			aiRecap: &models.GeneratedRecap{
				Summary: "Recent adventures...",
			},
			expectError: false,
		},
		{
			name:         "Repository Error",
			sessionCount: 3,
			memoryError:  errors.New("database error"),
			expectError:  true,
		},
		{
			name:        "AI Error",
			sessionCount: 3,
			memories:    mockMemories,
			aiError:     errors.New("AI service error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCampaignRepository)
			mockGameRepo := new(MockGameSessionRepository)
			mockAI := new(MockAICampaignManager)

			expectedLimit := tt.sessionCount
			if expectedLimit <= 0 {
				expectedLimit = 3
			}

			mockRepo.On("GetSessionMemories", sessionID, expectedLimit).
				Return(tt.memories, tt.memoryError)

			if tt.memoryError == nil && len(tt.memories) > 0 {
				mockAI.On("GenerateSessionRecap", mock.Anything, tt.memories).
					Return(tt.aiRecap, tt.aiError)
			}

			service := createTestCampaignService(mockRepo, mockGameRepo, mockAI)
			recap, err := service.GenerateRecap(context.Background(), sessionID, tt.sessionCount)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, recap)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, recap)
				if tt.validate != nil {
					tt.validate(t, recap)
				}
			}

			mockRepo.AssertExpectations(t)
			mockAI.AssertExpectations(t)
		})
	}
}

// Tests for Plot Thread Management

func TestCampaignService_CreatePlotThread(t *testing.T) {
	sessionID := uuid.New()

	tests := []struct {
		name        string
		thread      *models.PlotThread
		setupMocks  func(*MockCampaignRepository)
		expectError bool
		validate    func(*testing.T, *models.PlotThread)
	}{
		{
			name: "Complete Plot Thread",
			thread: &models.PlotThread{
				Title:        "The Missing Prince",
				Description:  "The prince has vanished",
				Status:       "active",
				TensionLevel: 8,
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("CreatePlotThread", mock.AnythingOfType("*models.PlotThread")).Return(nil)
			},
			expectError: false,
			validate: func(t *testing.T, thread *models.PlotThread) {
				assert.NotEqual(t, uuid.Nil, thread.ID)
				assert.Equal(t, sessionID, thread.GameSessionID)
				assert.Equal(t, "active", thread.Status)
				assert.Equal(t, 8, thread.TensionLevel)
			},
		},
		{
			name: "Default Values",
			thread: &models.PlotThread{
				Title:       "Mystery Plot",
				Description: "Something mysterious",
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("CreatePlotThread", mock.AnythingOfType("*models.PlotThread")).Return(nil)
			},
			expectError: false,
			validate: func(t *testing.T, thread *models.PlotThread) {
				assert.Equal(t, "active", thread.Status)      // Default
				assert.Equal(t, 5, thread.TensionLevel)        // Default
			},
		},
		{
			name: "Repository Error",
			thread: &models.PlotThread{
				Title: "Failed Thread",
			},
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("CreatePlotThread", mock.AnythingOfType("*models.PlotThread")).
					Return(errors.New("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCampaignRepository)
			mockGameRepo := new(MockGameSessionRepository)
			mockAI := new(MockAICampaignManager)

			tt.setupMocks(mockRepo)

			service := createTestCampaignService(mockRepo, mockGameRepo, mockAI)
			err := service.CreatePlotThread(context.Background(), sessionID, tt.thread)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.thread)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignService_GetPlotThreads(t *testing.T) {
	sessionID := uuid.New()
	activeThreads := []*models.PlotThread{
		{ID: uuid.New(), Title: "Active 1", Status: "active"},
		{ID: uuid.New(), Title: "Active 2", Status: "active"},
	}
	allThreads := append(activeThreads, &models.PlotThread{
		ID: uuid.New(), Title: "Resolved", Status: "resolved",
	})

	tests := []struct {
		name        string
		activeOnly  bool
		setupMocks  func(*MockCampaignRepository)
		expectError bool
		expectCount int
	}{
		{
			name:       "Get Active Threads Only",
			activeOnly: true,
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("GetActivePlotThreads", sessionID).Return(activeThreads, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:       "Get All Threads",
			activeOnly: false,
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("GetPlotThreadsBySession", sessionID).Return(allThreads, nil)
			},
			expectError: false,
			expectCount: 3,
		},
		{
			name:       "Repository Error - Active",
			activeOnly: true,
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("GetActivePlotThreads", sessionID).Return(nil, errors.New("database error"))
			},
			expectError: true,
		},
		{
			name:       "Repository Error - All",
			activeOnly: false,
			setupMocks: func(repo *MockCampaignRepository) {
				repo.On("GetPlotThreadsBySession", sessionID).Return(nil, errors.New("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCampaignRepository)
			mockGameRepo := new(MockGameSessionRepository)
			mockAI := new(MockAICampaignManager)

			tt.setupMocks(mockRepo)

			service := createTestCampaignService(mockRepo, mockGameRepo, mockAI)
			threads, err := service.GetPlotThreads(context.Background(), sessionID, tt.activeOnly)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, threads, tt.expectCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Benchmarks

func BenchmarkCampaignService_CreateStoryArc(b *testing.B) {
	mockRepo := new(MockCampaignRepository)
	mockGameRepo := new(MockGameSessionRepository)
	mockAI := new(MockAICampaignManager)
	
	mockRepo.On("CreateStoryArc", mock.AnythingOfType("*models.StoryArc")).Return(nil)
	
	service := createTestCampaignService(mockRepo, mockGameRepo, mockAI)
	sessionID := uuid.New()
	request := models.CreateStoryArcRequest{
		Title:       "Benchmark Arc",
		Description: "A test story arc",
		ArcType:     "main",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CreateStoryArc(context.Background(), sessionID, request)
	}
}

func BenchmarkCampaignService_GenerateRecap(b *testing.B) {
	mockRepo := new(MockCampaignRepository)
	mockGameRepo := new(MockGameSessionRepository)
	mockAI := new(MockAICampaignManager)
	
	memories := []*models.SessionMemory{
		{ID: uuid.New(), SessionNumber: 1},
		{ID: uuid.New(), SessionNumber: 2},
	}
	
	mockRepo.On("GetSessionMemories", mock.Anything, 3).Return(memories, nil)
	mockAI.On("GenerateSessionRecap", mock.Anything, memories).Return(&models.GeneratedRecap{
		Summary: "Test recap",
	}, nil)
	
	service := createTestCampaignService(mockRepo, mockGameRepo, mockAI)
	sessionID := uuid.New()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateRecap(context.Background(), sessionID, 3)
	}
}