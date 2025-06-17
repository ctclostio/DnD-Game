package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil"
)

// Helper functions to reduce code duplication in mock implementations

func mockErrorReturn(args mock.Arguments, index int) error {
	return args.Error(index)
}

func mockSingleReturn[T any](args mock.Arguments, valueIndex, errorIndex int) (*T, error) {
	if args.Get(valueIndex) == nil {
		return nil, args.Error(errorIndex)
	}
	return args.Get(valueIndex).(*T), args.Error(errorIndex)
}

func mockSliceReturn[T any](args mock.Arguments, valueIndex, errorIndex int) ([]*T, error) {
	if args.Get(valueIndex) == nil {
		return nil, args.Error(errorIndex)
	}
	return args.Get(valueIndex).([]*T), args.Error(errorIndex)
}

// MockWorldBuildingRepository implements all methods of WorldBuildingRepository
type MockWorldBuildingRepository struct {
	mock.Mock
}

// Settlement operations
func (m *MockWorldBuildingRepository) CreateSettlement(settlement *models.Settlement) error {
	args := m.Called(settlement)
	return mockErrorReturn(args, 0)
}

func (m *MockWorldBuildingRepository) GetSettlement(id uuid.UUID) (*models.Settlement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Settlement), args.Error(1)
}

func (m *MockWorldBuildingRepository) GetSettlementsByGameSession(gameSessionID uuid.UUID) ([]*models.Settlement, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Settlement), args.Error(1)
}

// NPC operations
func (m *MockWorldBuildingRepository) CreateSettlementNPC(npc *models.SettlementNPC) error {
	args := m.Called(npc)
	return args.Error(0)
}

func (m *MockWorldBuildingRepository) GetSettlementNPCs(settlementID uuid.UUID) ([]models.SettlementNPC, error) {
	args := m.Called(settlementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SettlementNPC), args.Error(1)
}

// Shop operations
func (m *MockWorldBuildingRepository) CreateSettlementShop(shop *models.SettlementShop) error {
	args := m.Called(shop)
	return args.Error(0)
}

func (m *MockWorldBuildingRepository) GetSettlementShops(settlementID uuid.UUID) ([]models.SettlementShop, error) {
	args := m.Called(settlementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SettlementShop), args.Error(1)
}

// Faction operations
func (m *MockWorldBuildingRepository) CreateFaction(faction *models.Faction) error {
	args := m.Called(faction)
	return args.Error(0)
}

func (m *MockWorldBuildingRepository) GetFaction(id uuid.UUID) (*models.Faction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Faction), args.Error(1)
}

func (m *MockWorldBuildingRepository) GetFactionsByGameSession(gameSessionID uuid.UUID) ([]*models.Faction, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Faction), args.Error(1)
}

func (m *MockWorldBuildingRepository) UpdateFactionRelationship(faction1ID, faction2ID uuid.UUID, standing int, relationType string) error {
	args := m.Called(faction1ID, faction2ID, standing, relationType)
	return args.Error(0)
}

// World Event operations
func (m *MockWorldBuildingRepository) CreateWorldEvent(event *models.WorldEvent) error {
	args := m.Called(event)
	return mockErrorReturn(args, 0)
}

func (m *MockWorldBuildingRepository) GetActiveWorldEvents(gameSessionID uuid.UUID) ([]*models.WorldEvent, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.WorldEvent), args.Error(1)
}

func (m *MockWorldBuildingRepository) ProgressWorldEvent(eventID uuid.UUID) error {
	args := m.Called(eventID)
	return mockErrorReturn(args, 0)
}

func (m *MockWorldBuildingRepository) GetWorldEventByID(eventID uuid.UUID) (*models.WorldEvent, error) {
	args := m.Called(eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WorldEvent), args.Error(1)
}

func (m *MockWorldBuildingRepository) UpdateWorldEvent(event *models.WorldEvent) error {
	args := m.Called(event)
	return mockErrorReturn(args, 0)
}

// Market operations
func (m *MockWorldBuildingRepository) CreateOrUpdateMarket(market *models.Market) error {
	args := m.Called(market)
	return mockErrorReturn(args, 0)
}

func (m *MockWorldBuildingRepository) GetMarketBySettlement(settlementID uuid.UUID) (*models.Market, error) {
	args := m.Called(settlementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Market), args.Error(1)
}

// Trade Route operations
func (m *MockWorldBuildingRepository) CreateTradeRoute(route *models.TradeRoute) error {
	args := m.Called(route)
	return args.Error(0)
}

func (m *MockWorldBuildingRepository) GetTradeRoutesBySettlement(settlementID uuid.UUID) ([]*models.TradeRoute, error) {
	args := m.Called(settlementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TradeRoute), args.Error(1)
}

// Ancient Site operations
func (m *MockWorldBuildingRepository) CreateAncientSite(site *models.AncientSite) error {
	args := m.Called(site)
	return args.Error(0)
}

func (m *MockWorldBuildingRepository) GetAncientSitesByGameSession(gameSessionID uuid.UUID) ([]*models.AncientSite, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AncientSite), args.Error(1)
}

// Economic simulation
func (m *MockWorldBuildingRepository) SimulateEconomicChanges(gameSessionID uuid.UUID) error {
	args := m.Called(gameSessionID)
	return args.Error(0)
}

// TestMockLLMProvider is a test-specific mock for LLMProvider that doesn't conflict with the one in llm_providers.go
type TestMockLLMProvider struct {
	mock.Mock
	Response string
	Error    error
}

func (m *TestMockLLMProvider) GenerateCompletion(_ context.Context, _ string, systemPrompt string) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}
	return m.Response, nil
}

func (m *TestMockLLMProvider) GenerateContent(ctx context.Context, prompt, system string) (string, error) {
	return m.GenerateCompletion(ctx, prompt, system)
}

func (m *TestMockLLMProvider) GenerateJSON(ctx context.Context, prompt, system string, _ interface{}) (string, error) {
	return m.GenerateCompletion(ctx, prompt, system)
}

func (m *TestMockLLMProvider) StreamContent(_ context.Context, _, system string) (<-chan string, <-chan error) {
	content := make(chan string)
	errors := make(chan error)
	close(content)
	close(errors)
	return content, errors
}

func TestNewWorldEventEngineService(t *testing.T) {
	mockLLM := &TestMockLLMProvider{}
	mockRepo := &MockWorldBuildingRepository{}
	mockFaction := &FactionSystemService{
		llmProvider: mockLLM,
		worldRepo:   mockRepo,
	}

	service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

	require.NotNil(t, service)
	require.Equal(t, mockLLM, service.llmProvider)
	require.Equal(t, mockRepo, service.worldRepo)
	require.Equal(t, mockFaction, service.factionService)
}

func TestWorldEventEngineService_GenerateWorldEvent(t *testing.T) {
	t.Run("successful event generation", func(t *testing.T) {
		// Setup mocks
		mockLLM := &TestMockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &FactionSystemService{
			llmProvider: mockLLM,
			worldRepo:   mockRepo,
		}

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
			"name":            "The Great Plague",
			"description":     "A mysterious disease spreads across the land",
			"cause":           "Ancient curse awakened",
			"severity":        "major",
			"duration":        "3 months",
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
			"ancientCause":    true,
			"prophecyRelated": true,
			"partyOpportunities": []string{
				"Find the cure",
				"Discover the source",
				"Protect settlements",
			},
		}

		aiResponse, _ := json.Marshal(aiEvent)
		mockLLM.Response = string(aiResponse)

		mockRepo.On("CreateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(nil)

		// Mock market operations for applyEventEffects
		mockRepo.On("GetMarketBySettlement", mock.AnythingOfType("uuid.UUID")).Return(nil, nil).Maybe()
		mockRepo.On("CreateOrUpdateMarket", mock.AnythingOfType("*models.Market")).Return(nil).Maybe()

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
		mockLLM := &TestMockLLMProvider{
			Error: errors.New("API error"),
		}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &FactionSystemService{
			llmProvider: mockLLM,
			worldRepo:   mockRepo,
		}

		gameSessionID := uuid.New()

		mockRepo.On("GetSettlementsByGameSession", gameSessionID).Return([]*models.Settlement{}, nil)
		mockRepo.On("GetFactionsByGameSession", gameSessionID).Return([]*models.Faction{}, nil)
		mockRepo.On("GetActiveWorldEvents", gameSessionID).Return([]*models.WorldEvent{}, nil)

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		ctx := testutil.TestContext()
		event, err := service.GenerateWorldEvent(ctx, gameSessionID, models.EventNatural)

		require.NoError(t, err) // Should fallback to procedural generation
		require.NotNil(t, event)
		require.Equal(t, models.EventNatural, event.Type)
	})

	t.Run("invalid AI response", func(t *testing.T) {
		mockLLM := &TestMockLLMProvider{
			Response: "invalid json",
		}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &FactionSystemService{
			llmProvider: mockLLM,
			worldRepo:   mockRepo,
		}

		gameSessionID := uuid.New()

		mockRepo.On("GetSettlementsByGameSession", gameSessionID).Return([]*models.Settlement{}, nil)
		mockRepo.On("GetFactionsByGameSession", gameSessionID).Return([]*models.Faction{}, nil)
		mockRepo.On("GetActiveWorldEvents", gameSessionID).Return([]*models.WorldEvent{}, nil)

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		ctx := testutil.TestContext()
		event, err := service.GenerateWorldEvent(ctx, gameSessionID, models.EventSupernatural)

		require.NoError(t, err) // Should fallback to procedural generation
		require.NotNil(t, event)
		require.Equal(t, models.EventSupernatural, event.Type)
	})

	t.Run("repository error", func(t *testing.T) {
		mockLLM := &TestMockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &FactionSystemService{
			llmProvider: mockLLM,
			worldRepo:   mockRepo,
		}

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
		mockLLM := &TestMockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &FactionSystemService{
			llmProvider: mockLLM,
			worldRepo:   mockRepo,
		}

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
			Stages: func() models.JSONB {
				data, _ := json.Marshal(map[string]interface{}{
					"1": map[string]interface{}{"name": "Stage 1"},
					"2": map[string]interface{}{"name": "Stage 2"},
				})
				return models.JSONB(data)
			}(),
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
		// The service processes events but doesn't necessarily update them
		// due to random chance in shouldEventProgress
		mockRepo.On("GetWorldEventByID", mock.AnythingOfType("uuid.UUID")).Return(event1, nil).Maybe()
		mockRepo.On("UpdateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(nil).Maybe()
		mockRepo.On("ProgressWorldEvent", mock.AnythingOfType("uuid.UUID")).Return(nil).Maybe()
		// resolveEvent creates a resolution event
		mockRepo.On("CreateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(nil).Maybe()
		// Market operations for applyEventEffects
		mockRepo.On("GetMarketBySettlement", mock.AnythingOfType("uuid.UUID")).Return(nil, nil).Maybe()
		mockRepo.On("CreateOrUpdateMarket", mock.AnythingOfType("*models.Market")).Return(nil).Maybe()
		// SimulateEventProgression may generate new events
		mockRepo.On("GetSettlementsByGameSession", gameSessionID).Return([]*models.Settlement{}, nil).Maybe()
		mockRepo.On("GetFactionsByGameSession", gameSessionID).Return([]*models.Faction{}, nil).Maybe()

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		ctx := testutil.TestContext()
		err := service.SimulateEventProgression(ctx, gameSessionID)

		require.NoError(t, err)
		// Don't assert specific expectations since progression is random
	})

	t.Run("no active events", func(t *testing.T) {
		mockLLM := &TestMockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &FactionSystemService{
			llmProvider: mockLLM,
			worldRepo:   mockRepo,
		}

		gameSessionID := uuid.New()

		mockRepo.On("GetActiveWorldEvents", gameSessionID).Return([]*models.WorldEvent{}, nil)
		// SimulateEventProgression has a 20% chance to generate new events
		// which calls calculateWorldCorruption
		mockRepo.On("GetSettlementsByGameSession", gameSessionID).Return([]*models.Settlement{}, nil).Maybe()
		mockRepo.On("GetFactionsByGameSession", gameSessionID).Return([]*models.Faction{}, nil).Maybe()
		// If it generates a new event
		mockRepo.On("CreateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(nil).Maybe()
		mockRepo.On("GetMarketBySettlement", mock.AnythingOfType("uuid.UUID")).Return(nil, nil).Maybe()
		mockRepo.On("CreateOrUpdateMarket", mock.AnythingOfType("*models.Market")).Return(nil).Maybe()

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		ctx := testutil.TestContext()
		err := service.SimulateEventProgression(ctx, gameSessionID)

		require.NoError(t, err)
	})
}

func TestWorldEventEngineService_NotifyPartyOfEvent(t *testing.T) {
	t.Run("notify party", func(t *testing.T) {
		mockLLM := &TestMockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &FactionSystemService{
			llmProvider: mockLLM,
			worldRepo:   mockRepo,
		}

		eventID := uuid.New()

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		// The method is a stub that just returns nil
		err := service.NotifyPartyOfEvent(eventID)

		require.NoError(t, err)
	})

	t.Run("event not found", func(t *testing.T) {
		mockLLM := &TestMockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &FactionSystemService{
			llmProvider: mockLLM,
			worldRepo:   mockRepo,
		}

		eventID := uuid.New()

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		// The method is a stub that always returns nil
		err := service.NotifyPartyOfEvent(eventID)

		require.NoError(t, err)
	})
}

func TestWorldEventEngineService_RecordPartyAction(t *testing.T) {
	t.Run("record action", func(t *testing.T) {
		mockLLM := &TestMockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &FactionSystemService{
			llmProvider: mockLLM,
			worldRepo:   mockRepo,
		}

		eventID := uuid.New()

		service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

		// The method is a stub that just returns nil
		err := service.RecordPartyAction(eventID, "Investigated the source")

		require.NoError(t, err)
	})
}

func TestWorldEventEngineService_shouldEventProgress(t *testing.T) {
	service := &WorldEventEngineService{}

	// Since shouldEventProgress uses random values, we'll test it multiple times
	// and check for expected behavior patterns

	t.Run("minor event progression", func(t *testing.T) {
		event := &models.WorldEvent{
			IsActive:     true,
			IsResolved:   false,
			Severity:     models.SeverityMinor,
			AncientCause: false,
		}

		// Run multiple times to get a sense of the probability
		progressCount := 0
		runs := 100
		for i := 0; i < runs; i++ {
			if service.shouldEventProgress(event) {
				progressCount++
			}
		}

		// Should progress roughly 30% of the time (±15% for test stability)
		progressRate := float64(progressCount) / float64(runs)
		require.True(t, progressRate >= 0.15 && progressRate <= 0.45,
			"Expected progress rate around 30%%, got %.2f%%", progressRate*100)
	})

	t.Run("major event progression", func(t *testing.T) {
		event := &models.WorldEvent{
			IsActive:     true,
			IsResolved:   false,
			Severity:     models.SeverityMajor,
			AncientCause: false,
		}

		// Run multiple times to get a sense of the probability
		progressCount := 0
		runs := 100
		for i := 0; i < runs; i++ {
			if service.shouldEventProgress(event) {
				progressCount++
			}
		}

		// Should progress roughly 50% of the time (±15% for test stability)
		progressRate := float64(progressCount) / float64(runs)
		require.True(t, progressRate >= 0.35 && progressRate <= 0.65,
			"Expected progress rate around 50%%, got %.2f%%", progressRate*100)
	})

	t.Run("ancient major event progression", func(t *testing.T) {
		event := &models.WorldEvent{
			IsActive:     true,
			IsResolved:   false,
			Severity:     models.SeverityMajor,
			AncientCause: true,
		}

		// Run multiple times to get a sense of the probability
		progressCount := 0
		runs := 100
		for i := 0; i < runs; i++ {
			if service.shouldEventProgress(event) {
				progressCount++
			}
		}

		// Should progress roughly 70% of the time (50% + 20%) (±15% for test stability)
		progressRate := float64(progressCount) / float64(runs)
		require.True(t, progressRate >= 0.55 && progressRate <= 0.85,
			"Expected progress rate around 70%%, got %.2f%%", progressRate*100)
	})
}

func TestWorldEventEngineService_determineAffectedSettlements(t *testing.T) {
	service := &WorldEventEngineService{}

	settlements := []*models.Settlement{
		{
			ID:     uuid.New(),
			Name:   "Northern City",
			Region: "North",
		},
		{
			ID:     uuid.New(),
			Name:   "Eastern Town",
			Region: "East",
		},
		{
			ID:     uuid.New(),
			Name:   "Southern Village",
			Region: "South",
		},
	}

	t.Run("affect specific regions", func(t *testing.T) {
		affectedRegions := []string{"North", "East"}

		result := service.determineAffectedSettlements(settlements, affectedRegions)

		var affectedIDs []string
		err := json.Unmarshal([]byte(result), &affectedIDs)
		require.NoError(t, err)
		require.Len(t, affectedIDs, 2)
	})

	t.Run("empty regions affects all", func(t *testing.T) {
		affectedRegions := []string{}

		result := service.determineAffectedSettlements(settlements, affectedRegions)

		var affectedIDs []string
		err := json.Unmarshal([]byte(result), &affectedIDs)
		require.NoError(t, err)
		require.True(t, len(affectedIDs) >= 1 && len(affectedIDs) <= 3) // Random selection
	})
}

func TestWorldEventEngineService_determineAffectedFactions(t *testing.T) {
	service := &WorldEventEngineService{}

	factions := []*models.Faction{
		{
			ID:   uuid.New(),
			Name: "Merchant Guild",
			Type: models.FactionMerchant,
		},
		{
			ID:   uuid.New(),
			Name: "Noble House",
			Type: models.FactionPolitical,
		},
		{
			ID:   uuid.New(),
			Name: "Religious Order",
			Type: models.FactionReligious,
		},
	}

	t.Run("economic event affects merchants", func(t *testing.T) {
		result := service.determineAffectedFactions(factions, models.EventEconomic)

		var affectedIDs []string
		err := json.Unmarshal([]byte(result), &affectedIDs)
		require.NoError(t, err)
		// EventEconomic affects FactionMerchant and FactionCriminal types
		// We have FactionMerchant, so at least one should be affected
		require.True(t, len(affectedIDs) >= 1)
	})

	t.Run("supernatural event", func(t *testing.T) {
		result := service.determineAffectedFactions(factions, models.EventSupernatural)

		var affectedIDs []string
		err := json.Unmarshal([]byte(result), &affectedIDs)
		require.NoError(t, err)
		// EventSupernatural affects FactionCult, FactionAncientOrder, FactionReligious types
		// We have FactionTypeReligious, so it should be affected if the types match
		require.True(t, len(affectedIDs) >= 1)
	})
}

func TestWorldEventEngineService_calculateWorldCorruption(t *testing.T) {
	mockLLM := &TestMockLLMProvider{}
	mockRepo := &MockWorldBuildingRepository{}
	mockFaction := &FactionSystemService{
		llmProvider: mockLLM,
		worldRepo:   mockRepo,
	}

	gameSessionID := uuid.New()

	// Mock settlements with different corruption levels
	settlements := []*models.Settlement{
		{
			ID:              uuid.New(),
			Name:            "Corrupted City",
			CorruptionLevel: 80,
		},
		{
			ID:              uuid.New(),
			Name:            "Tainted Town",
			CorruptionLevel: 40,
		},
		{
			ID:              uuid.New(),
			Name:            "Pure Village",
			CorruptionLevel: 0,
		},
	}

	mockRepo.On("GetSettlementsByGameSession", gameSessionID).Return(settlements, nil)

	service := NewWorldEventEngineService(mockLLM, mockRepo, mockFaction)

	corruption := service.calculateWorldCorruption(gameSessionID)

	// Average corruption: (80 + 40 + 0) / 3 = 40
	require.Equal(t, 40, corruption)
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
			models.SeverityMajor,
		}, event.Severity)
	})
}

// Integration test
func TestWorldEventEngineService_Integration(t *testing.T) {
	t.Run("complete event lifecycle", func(t *testing.T) {
		mockLLM := &TestMockLLMProvider{}
		mockRepo := &MockWorldBuildingRepository{}
		mockFaction := &FactionSystemService{
			llmProvider: mockLLM,
			worldRepo:   mockRepo,
		}

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
		mockRepo.On("GetWorldEventByID", createdEvent.ID).Return(createdEvent, nil).Maybe()
		mockRepo.On("UpdateWorldEvent", mock.AnythingOfType("*models.WorldEvent")).Return(nil).Maybe()
		mockRepo.On("ProgressWorldEvent", mock.AnythingOfType("uuid.UUID")).Return(nil).Maybe()

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
	mockLLM := &TestMockLLMProvider{
		Response: `{"name":"Test Event","description":"Test","severity":"minor"}`,
	}
	mockRepo := &MockWorldBuildingRepository{}
	mockFaction := &FactionSystemService{
		llmProvider: mockLLM,
		worldRepo:   mockRepo,
	}

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
			ID:     uuid.New(),
			Name:   "Settlement",
			Region: "North",
		}
	}

	affectedRegions := []string{"North", "South", "East"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.determineAffectedSettlements(settlements, affectedRegions)
	}
}
