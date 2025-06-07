package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

// Mock implementations
type MockWorldBuildingRepository struct {
	mock.Mock
}

func (m *MockWorldBuildingRepository) GetSettlementsByGameSession(gameSessionID uuid.UUID) ([]*models.Settlement, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Settlement), args.Error(1)
}

func (m *MockWorldBuildingRepository) GetFactionsByGameSession(gameSessionID uuid.UUID) ([]*models.Faction, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Faction), args.Error(1)
}

func (m *MockWorldBuildingRepository) GetActiveWorldEvents(gameSessionID uuid.UUID) ([]*models.WorldEvent, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.WorldEvent), args.Error(1)
}

func (m *MockWorldBuildingRepository) CreateWorldEvent(event *models.WorldEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockWorldBuildingRepository) UpdateWorldEvent(event *models.WorldEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockWorldBuildingRepository) GetWorldEventByID(eventID uuid.UUID) (*models.WorldEvent, error) {
	args := m.Called(eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WorldEvent), args.Error(1)
}

type MockFactionSystemService struct {
	mock.Mock
}

func TestNewWorldEventEngineService(t *testing.T) {
	mockLLM := &MockLLMProvider{}
	mockRepo := &MockWorldBuildingRepository{}
	mockFaction := &MockFactionSystemService{}

	service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

	require.NotNil(t, service)
	require.Equal(t, mockLLM, service.llmProvider)
	require.Equal(t, mockRepo, service.worldRepo)
	require.Equal(t, mockFaction, service.factionService)
}

func TestWorldEventEngineService_GenerateWorldEvent(t *testing.T) {
	t.Run("successful event generation", func(t *testing.T) {
		// Setup mocks
		mockLLM := &MockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &MockFactionSystemService{}

		gameSessionID := uuid.New()

		// Mock repository responses
		settlements := []*models.Settlement{
			{ID: uuid.New(), Name: "Winterhold"},
			{ID: uuid.New(), Name: "Solitude"},
		}
		factions := []*models.Faction{
			{ID: uuid.New(), Name: "Mages Guild"},
			{ID: uuid.New(), Name: "Thieves Guild"},
		}
		activeEvents := []*models.WorldEvent{
			{ID: uuid.New(), Name: "Dragon Attack"},
		}

		mockRepo.On("GetSettlementsByGameSession", gameSessionID).Return(settlements, nil)
		mockRepo.On("GetFactionsByGameSession", gameSessionID).Return(factions, nil)
		mockRepo.On("GetActiveWorldEvents", gameSessionID).Return(activeEvents, nil)

		// Mock LLM response
		aiEvent := map[string]interface{}{
			"name":        "The Great Plague",
			"description": "A mysterious disease spreads across the land",
			"cause":       "Ancient curse awakened",
			"severity":    "major",
			"duration":    "3 months",
			"affectedRegions": []string{"Northern Kingdoms", "Eastern Provinces"},
			"affectedSettlements": map[string]string{
				"Winterhold": "severe",
				"Solitude":   "moderate",
			},
			"affectedFactions": map[string]string{
				"Mages Guild":   "researching cure",
				"Thieves Guild": "exploiting chaos",
			},
			"stages": []map[string]interface{}{
				{
					"number":      1,
					"name":        "Initial Outbreak",
					"description": "First cases appear",
					"duration":    "2 weeks",
				},
			},
			"economicImpacts": map[string]interface{}{
				"trade":      -30,
				"population": -15,
			},
			"ancientCause":      true,
			"prophecyRelated":   true,
			"partyOpportunities": []string{
				"Find the cure",
				"Discover the source",
				"Protect settlements",
			},
		}

		aiResponse, _ := json.Marshal(aiEvent)
		mockLLM.Response = string(aiResponse)

		mockRepo.On("CreateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(nil)

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		// Execute
		ctx := testutil.TestContext()
		event, err := service.GenerateWorldEvent(ctx, gameSessionID, models.EventPolitical)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, event)
		require.Equal(t, "The Great Plague", event.Name)
		require.Equal(t, models.EventPolitical, event.Type)
		require.Equal(t, gameSessionID, event.GameSessionID)
		require.True(t, event.IsActive)
		require.False(t, event.PartyAware)

		mockRepo.AssertExpectations(t)
	})

	t.Run("LLM provider error", func(t *testing.T) {
		mockLLM := &MockLLMProvider{
			Error: errors.New("API error"),
		}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &MockFactionSystemService{}

		gameSessionID := uuid.New()

		mockRepo.On("GetSettlementsByGameSession", gameSessionID).Return([]*models.Settlement{}, nil)
		mockRepo.On("GetFactionsByGameSession", gameSessionID).Return([]*models.Faction{}, nil)
		mockRepo.On("GetActiveWorldEvents", gameSessionID).Return([]*models.WorldEvent{}, nil)

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		ctx := testutil.TestContext()
		event, err := service.GenerateWorldEvent(ctx, gameSessionID, models.EventNatural)

		require.Error(t, err)
		require.Nil(t, event)
		require.Contains(t, err.Error(), "failed to generate world event")
	})

	t.Run("invalid AI response", func(t *testing.T) {
		mockLLM := &MockLLMProvider{
			Response: "invalid json",
		}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &MockFactionSystemService{}

		gameSessionID := uuid.New()

		mockRepo.On("GetSettlementsByGameSession", gameSessionID).Return([]*models.Settlement{}, nil)
		mockRepo.On("GetFactionsByGameSession", gameSessionID).Return([]*models.Faction{}, nil)
		mockRepo.On("GetActiveWorldEvents", gameSessionID).Return([]*models.WorldEvent{}, nil)

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		ctx := testutil.TestContext()
		event, err := service.GenerateWorldEvent(ctx, gameSessionID, models.EventMagical)

		require.Error(t, err)
		require.Nil(t, event)
		require.Contains(t, err.Error(), "failed to parse AI response")
	})

	t.Run("repository error", func(t *testing.T) {
		mockLLM := &MockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &MockFactionSystemService{}

		gameSessionID := uuid.New()

		mockRepo.On("GetSettlementsByGameSession", gameSessionID).Return(nil, errors.New("db error"))
		mockRepo.On("GetFactionsByGameSession", gameSessionID).Return([]*models.Faction{}, nil)
		mockRepo.On("GetActiveWorldEvents", gameSessionID).Return([]*models.WorldEvent{}, nil)

		// Still generate event despite repo error for settlements
		aiEvent := map[string]interface{}{
			"name":        "Test Event",
			"description": "Test",
			"severity":    "minor",
		}
		aiResponse, _ := json.Marshal(aiEvent)
		mockLLM.Response = string(aiResponse)

		mockRepo.On("CreateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(errors.New("create error"))

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		ctx := testutil.TestContext()
		event, err := service.GenerateWorldEvent(ctx, gameSessionID, models.EventPolitical)

		require.Error(t, err)
		require.Nil(t, event)
		require.Contains(t, err.Error(), "failed to save world event")
	})
}

func TestWorldEventEngineService_SimulateEventProgression(t *testing.T) {
	t.Run("progress active events", func(t *testing.T) {
		mockLLM := &MockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &MockFactionSystemService{}

		gameSessionID := uuid.New()

		// Create test events
		event1 := &models.WorldEvent{
			ID:            uuid.New(),
			GameSessionID: gameSessionID,
			Name:          "Event 1",
			IsActive:      true,
			StartDate:     "Day 1",
			Duration:      "1 week",
			CurrentStage:  1,
			Stages: models.JSONB{
				"1": map[string]interface{}{"name": "Stage 1"},
				"2": map[string]interface{}{"name": "Stage 2"},
			},
		}

		event2 := &models.WorldEvent{
			ID:            uuid.New(),
			GameSessionID: gameSessionID,
			Name:          "Event 2",
			IsActive:      true,
			IsResolved:    false,
		}

		activeEvents := []*models.WorldEvent{event1, event2}

		mockRepo.On("GetActiveWorldEvents", gameSessionID).Return(activeEvents, nil)
		mockRepo.On("GetWorldEventByID", event1.ID).Return(event1, nil)
		mockRepo.On("GetWorldEventByID", event2.ID).Return(event2, nil)
		mockRepo.On("UpdateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(nil)

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		ctx := testutil.TestContext()
		err := service.SimulateEventProgression(ctx, gameSessionID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no active events", func(t *testing.T) {
		mockLLM := &MockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &MockFactionSystemService{}

		gameSessionID := uuid.New()

		mockRepo.On("GetActiveWorldEvents", gameSessionID).Return([]*models.WorldEvent{}, nil)

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		ctx := testutil.TestContext()
		err := service.SimulateEventProgression(ctx, gameSessionID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestWorldEventEngineService_NotifyPartyOfEvent(t *testing.T) {
	t.Run("notify party", func(t *testing.T) {
		mockLLM := &MockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &MockFactionSystemService{}

		eventID := uuid.New()
		event := &models.WorldEvent{
			ID:         eventID,
			Name:       "Test Event",
			PartyAware: false,
		}

		mockRepo.On("GetWorldEventByID", eventID).Return(event, nil)
		mockRepo.On("UpdateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(nil)

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		err := service.NotifyPartyOfEvent(eventID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		mockLLM := &MockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &MockFactionSystemService{}

		eventID := uuid.New()

		mockRepo.On("GetWorldEventByID", eventID).Return(nil, errors.New("not found"))

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		err := service.NotifyPartyOfEvent(eventID)

		require.Error(t, err)
		require.Contains(t, err.Error(), "event not found")
	})
}

func TestWorldEventEngineService_RecordPartyAction(t *testing.T) {
	t.Run("record action", func(t *testing.T) {
		mockLLM := &MockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &MockFactionSystemService{}

		eventID := uuid.New()
		event := &models.WorldEvent{
			ID:           eventID,
			Name:         "Test Event",
			PartyActions: models.JSONB{},
		}

		mockRepo.On("GetWorldEventByID", eventID).Return(event, nil)
		mockRepo.On("UpdateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(nil)

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		err := service.RecordPartyAction(eventID, "Investigated the source")

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestWorldEventEngineService_shouldEventProgress(t *testing.T) {
	service := &WorldEventEngineService{}

	tests := []struct {
		name     string
		event    *models.WorldEvent
		expected bool
	}{
		{
			name: "resolved event should not progress",
			event: &models.WorldEvent{
				IsResolved: true,
			},
			expected: false,
		},
		{
			name: "inactive event should not progress",
			event: &models.WorldEvent{
				IsActive:   false,
				IsResolved: false,
			},
			expected: false,
		},
		{
			name: "active unresolved event should progress",
			event: &models.WorldEvent{
				IsActive:   true,
				IsResolved: false,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.shouldEventProgress(tt.event)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestWorldEventEngineService_determineAffectedSettlements(t *testing.T) {
	service := &WorldEventEngineService{}

	settlements := []*models.Settlement{
		{
			ID:       uuid.New(),
			Name:     "Northern City",
			Location: models.Location{Region: "North"},
		},
		{
			ID:       uuid.New(),
			Name:     "Eastern Town",
			Location: models.Location{Region: "East"},
		},
		{
			ID:       uuid.New(),
			Name:     "Southern Village",
			Location: models.Location{Region: "South"},
		},
	}

	t.Run("affect specific regions", func(t *testing.T) {
		affectedRegions := []string{"North", "East"}

		result := service.determineAffectedSettlements(settlements, affectedRegions)

		affected := result.(map[string]interface{})
		require.Len(t, affected, 2)
		require.Contains(t, affected, "Northern City")
		require.Contains(t, affected, "Eastern Town")
		require.NotContains(t, affected, "Southern Village")
	})

	t.Run("empty regions affects all", func(t *testing.T) {
		affectedRegions := []string{}

		result := service.determineAffectedSettlements(settlements, affectedRegions)

		affected := result.(map[string]interface{})
		require.Len(t, affected, 3)
	})
}

func TestWorldEventEngineService_determineAffectedFactions(t *testing.T) {
	service := &WorldEventEngineService{}

	factions := []*models.Faction{
		{
			ID:   uuid.New(),
			Name: "Merchant Guild",
			Type: models.FactionTypeGuild,
		},
		{
			ID:   uuid.New(),
			Name: "Noble House",
			Type: models.FactionTypeNoble,
		},
		{
			ID:   uuid.New(),
			Name: "Religious Order",
			Type: models.FactionTypeReligious,
		},
	}

	t.Run("economic event affects merchants", func(t *testing.T) {
		result := service.determineAffectedFactions(factions, models.EventEconomic)

		affected := result.(map[string]interface{})
		require.Contains(t, affected, "Merchant Guild")
		require.Contains(t, affected, "Noble House") // Nobles also affected by economic events
	})

	t.Run("religious event", func(t *testing.T) {
		result := service.determineAffectedFactions(factions, models.EventReligious)

		affected := result.(map[string]interface{})
		require.Contains(t, affected, "Religious Order")
	})
}

func TestWorldEventEngineService_calculateWorldCorruption(t *testing.T) {
	mockLLM := &MockLLMProvider{}
	mockRepo := &MockWorldBuildingRepository{}
	mockFaction := &MockFactionSystemService{}

	gameSessionID := uuid.New()

	// Mock active events with different corruption levels
	activeEvents := []*models.WorldEvent{
		{
			ID:              uuid.New(),
			AncientCause:    true, // +10
			AwakensAncientEvil: true, // +15
		},
		{
			ID:           uuid.New(),
			AncientCause: true, // +10
		},
		{
			ID: uuid.New(), // No corruption
		},
	}

	mockRepo.On("GetActiveWorldEvents", gameSessionID).Return(activeEvents, nil)

	service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

	corruption := service.calculateWorldCorruption(gameSessionID)

	require.Equal(t, 35, corruption) // 10 + 15 + 10
}

func TestWorldEventEngineService_generateProceduralEvent(t *testing.T) {
	service := &WorldEventEngineService{}

	gameSessionID := uuid.New()

	t.Run("natural disaster event", func(t *testing.T) {
		event := service.generateProceduralEvent(gameSessionID, models.EventNatural)

		require.NotNil(t, event)
		require.Equal(t, gameSessionID, event.GameSessionID)
		require.Equal(t, models.EventNatural, event.Type)
		require.NotEmpty(t, event.Name)
		require.NotEmpty(t, event.Description)
		require.True(t, event.IsActive)
	})

	t.Run("political event", func(t *testing.T) {
		event := service.generateProceduralEvent(gameSessionID, models.EventPolitical)

		require.NotNil(t, event)
		require.Equal(t, models.EventPolitical, event.Type)
		require.Contains(t, []models.WorldEventSeverity{
			models.SeverityMinor,
			models.SeverityModerate,
		}, event.Severity)
	})
}

// Integration test
func TestWorldEventEngineService_Integration(t *testing.T) {
	t.Run("complete event lifecycle", func(t *testing.T) {
		mockLLM := &MockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &MockFactionSystemService{}

		gameSessionID := uuid.New()

		// Setup world state
		settlements := []*models.Settlement{
			{ID: uuid.New(), Name: "Capital City"},
		}
		factions := []*models.Faction{
			{ID: uuid.New(), Name: "Royal Court"},
		}

		mockRepo.On("GetSettlementsByGameSession", gameSessionID).Return(settlements, nil)
		mockRepo.On("GetFactionsByGameSession", gameSessionID).Return(factions, nil)
		mockRepo.On("GetActiveWorldEvents", gameSessionID).Return([]*models.WorldEvent{}, nil).Once()

		// Generate event
		aiEvent := map[string]interface{}{
			"name":        "Royal Succession Crisis",
			"description": "The king has died without a clear heir",
			"severity":    "major",
			"stages": []map[string]interface{}{
				{"number": 1, "name": "Initial Chaos"},
				{"number": 2, "name": "Faction War"},
				{"number": 3, "name": "Resolution"},
			},
		}
		aiResponse, _ := json.Marshal(aiEvent)
		mockLLM.Response = string(aiResponse)

		var createdEvent *models.WorldEvent
		mockRepo.On("CreateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Run(func(args mock.Arguments) {
			createdEvent = args.Get(0).(*models.WorldEvent)
			createdEvent.ID = uuid.New() // Simulate DB assigning ID
		}).Return(nil)

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		// Generate event
		ctx := testutil.TestContext()
		event, err := service.GenerateWorldEvent(ctx, gameSessionID, models.EventPolitical)

		require.NoError(t, err)
		require.NotNil(t, event)

		// Simulate progression
		mockRepo.On("GetActiveWorldEvents", gameSessionID).Return([]*models.WorldEvent{createdEvent}, nil)
		mockRepo.On("GetWorldEventByID", createdEvent.ID).Return(createdEvent, nil)
		mockRepo.On("UpdateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(nil)

		err = service.SimulateEventProgression(ctx, gameSessionID)
		require.NoError(t, err)

		// Notify party
		err = service.NotifyPartyOfEvent(createdEvent.ID)
		require.NoError(t, err)

		// Record party action
		err = service.RecordPartyAction(createdEvent.ID, "Supported the legitimate heir")
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkWorldEventEngineService_GenerateWorldEvent(b *testing.B) {
	mockLLM := &MockLLMProvider{
		Response: `{"name":"Test Event","description":"Test","severity":"minor"}`,
	}
	mockRepo := &MockWorldBuildingRepository{}
	mockFaction := &MockFactionSystemService{}

	gameSessionID := uuid.New()

	mockRepo.On("GetSettlementsByGameSession", gameSessionID).Return([]*models.Settlement{}, nil)
	mockRepo.On("GetFactionsByGameSession", gameSessionID).Return([]*models.Faction{}, nil)
	mockRepo.On("GetActiveWorldEvents", gameSessionID).Return([]*models.WorldEvent{}, nil)
	mockRepo.On("CreateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(nil)

	service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateWorldEvent(ctx, gameSessionID, models.EventNatural)
	}
}

func BenchmarkWorldEventEngineService_determineAffectedSettlements(b *testing.B) {
	service := &WorldEventEngineService{}

	settlements := make([]*models.Settlement, 100)
	for i := 0; i < 100; i++ {
		settlements[i] = &models.Settlement{
			ID:       uuid.New(),
			Name:     "Settlement",
			Location: models.Location{Region: "North"},
		}
	}

	affectedRegions := []string{"North", "South", "East"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.determineAffectedSettlements(settlements, affectedRegions)
	}
}