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

// Mock implementation of WorldBuildingRepository
type MockWorldBuildingRepositoryImpl struct {
	mock.Mock
}

func (m *MockWorldBuildingRepositoryImpl) CreateSettlement(settlement *models.Settlement) error {
	args := m.Called(settlement)
	return args.Error(0)
}

func (m *MockWorldBuildingRepositoryImpl) GetSettlement(id uuid.UUID) (*models.Settlement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Settlement), args.Error(1)
}

func (m *MockWorldBuildingRepositoryImpl) GetSettlementsByGameSession(gameSessionID uuid.UUID) ([]*models.Settlement, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Settlement), args.Error(1)
}

func (m *MockWorldBuildingRepositoryImpl) CreateSettlementNPC(npc *models.SettlementNPC) error {
	args := m.Called(npc)
	return args.Error(0)
}

func (m *MockWorldBuildingRepositoryImpl) GetSettlementNPCs(settlementID uuid.UUID) ([]models.SettlementNPC, error) {
	args := m.Called(settlementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SettlementNPC), args.Error(1)
}

func (m *MockWorldBuildingRepositoryImpl) CreateSettlementShop(shop *models.SettlementShop) error {
	args := m.Called(shop)
	return args.Error(0)
}

func (m *MockWorldBuildingRepositoryImpl) GetSettlementShops(settlementID uuid.UUID) ([]models.SettlementShop, error) {
	args := m.Called(settlementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SettlementShop), args.Error(1)
}

func (m *MockWorldBuildingRepositoryImpl) CreateFaction(faction *models.Faction) error {
	args := m.Called(faction)
	return args.Error(0)
}

func (m *MockWorldBuildingRepositoryImpl) GetFaction(id uuid.UUID) (*models.Faction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Faction), args.Error(1)
}

func (m *MockWorldBuildingRepositoryImpl) GetFactionsByGameSession(gameSessionID uuid.UUID) ([]*models.Faction, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Faction), args.Error(1)
}

func (m *MockWorldBuildingRepositoryImpl) UpdateFactionRelationship(faction1ID, faction2ID uuid.UUID, standing int, relationType string) error {
	args := m.Called(faction1ID, faction2ID, standing, relationType)
	return args.Error(0)
}

func (m *MockWorldBuildingRepositoryImpl) CreateWorldEvent(event *models.WorldEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockWorldBuildingRepositoryImpl) GetActiveWorldEvents(gameSessionID uuid.UUID) ([]*models.WorldEvent, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.WorldEvent), args.Error(1)
}

func (m *MockWorldBuildingRepositoryImpl) ProgressWorldEvent(eventID uuid.UUID) error {
	args := m.Called(eventID)
	return args.Error(0)
}

func (m *MockWorldBuildingRepositoryImpl) CreateOrUpdateMarket(market *models.Market) error {
	args := m.Called(market)
	return args.Error(0)
}

func (m *MockWorldBuildingRepositoryImpl) GetMarketBySettlement(settlementID uuid.UUID) (*models.Market, error) {
	args := m.Called(settlementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Market), args.Error(1)
}

func (m *MockWorldBuildingRepositoryImpl) CreateTradeRoute(route *models.TradeRoute) error {
	args := m.Called(route)
	return args.Error(0)
}

func (m *MockWorldBuildingRepositoryImpl) GetTradeRoutesBySettlement(settlementID uuid.UUID) ([]*models.TradeRoute, error) {
	args := m.Called(settlementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TradeRoute), args.Error(1)
}

func (m *MockWorldBuildingRepositoryImpl) CreateAncientSite(site *models.AncientSite) error {
	args := m.Called(site)
	return args.Error(0)
}

func (m *MockWorldBuildingRepositoryImpl) GetAncientSitesByGameSession(gameSessionID uuid.UUID) ([]*models.AncientSite, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AncientSite), args.Error(1)
}

func (m *MockWorldBuildingRepositoryImpl) SimulateEconomicChanges(gameSessionID uuid.UUID) error {
	args := m.Called(gameSessionID)
	return args.Error(0)
}

func TestNewSettlementGeneratorService(t *testing.T) {
	mockLLM := &MockLLMProvider{}
	mockRepo := &MockWorldBuildingRepositoryImpl{}

	service := NewSettlementGeneratorService(mockLLM, mockRepo)

	require.NotNil(t, service)
	require.Equal(t, mockLLM, service.llmProvider)
	require.Equal(t, mockRepo, service.worldRepo)
}

func TestSettlementGeneratorService_GenerateSettlement(t *testing.T) {
	t.Run("successful settlement generation", func(t *testing.T) {
		mockLLM := &MockLLMProvider{}
		mockRepo := &MockWorldBuildingRepositoryImpl{}

		gameSessionID := uuid.New()
		req := models.SettlementGenerationRequest{
			Type:             models.SettlementTown,
			Region:           "Northern Mountains",
			PopulationSize:   "medium",
			DangerLevel:      5,
			AncientInfluence: true,
			SpecialFeatures:  []string{"trade hub", "mining"},
		}

		// Mock AI response for settlement
		settlementResponse := `{
			"name": "Ironhold",
			"description": "A fortified town carved into the mountainside",
			"history": "Founded by dwarven miners who discovered ancient ruins",
			"governmentType": "Guild Council",
			"alignment": "Lawful Neutral",
			"ageCategory": "established",
			"notableLocations": [
				{"name": "The Deep Mine", "description": "Ancient excavation site"}
			],
			"problems": [
				{"title": "Missing Miners", "description": "Workers vanishing in deep tunnels"}
			],
			"secrets": [
				{"title": "Ancient Portal", "description": "Hidden gateway to forgotten realm"}
			],
			"defenses": ["Stone walls", "Mountain position"],
			"primaryExports": ["Iron ore", "Precious gems"],
			"primaryImports": ["Food", "Textiles"],
			"ancientRuinsNearby": true,
			"corruptionLevel": 3,
			"eldritchInfluence": 4
		}`

		// Note: In a real implementation, you'd need to mock multiple LLM calls
		// for NPCs and shops. For this test, we're focusing on the settlement generation.

		// For simplicity in this test, we'll use the settlement response only
		// In a real scenario, you'd need to mock multiple calls differently
		mockLLM.Response = settlementResponse

		// Mock repository calls
		mockRepo.On("CreateSettlement", mock.AnythingOfType("*models.Settlement")).Run(func(args mock.Arguments) {
			settlement := args.Get(0).(*models.Settlement)
			settlement.ID = uuid.New()
		}).Return(nil)

		mockRepo.On("CreateSettlementNPC", mock.AnythingOfType("*models.SettlementNPC")).Return(nil)
		mockRepo.On("CreateSettlementShop", mock.AnythingOfType("*models.SettlementShop")).Return(nil)
		mockRepo.On("CreateOrUpdateMarket", mock.AnythingOfType("*models.Market")).Return(nil)

		service := NewSettlementGeneratorService(mockLLM, mockRepo)

		// Execute
		ctx := testutil.TestContext()
		settlement, err := service.GenerateSettlement(ctx, gameSessionID, req)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, settlement)
		require.Equal(t, "Ironhold", settlement.Name)
		require.Equal(t, models.SettlementTown, settlement.Type)
		require.Equal(t, gameSessionID, settlement.GameSessionID)
		require.Equal(t, 3, settlement.CorruptionLevel)
		require.Equal(t, 4, settlement.EldritchInfluence)
		require.True(t, settlement.AncientRuinsNearby)

		mockRepo.AssertExpectations(t)
	})

	t.Run("LLM error returns error", func(t *testing.T) {
		mockLLM := &MockLLMProvider{
			Error: errors.New("API error"),
		}
		mockRepo := &MockWorldBuildingRepositoryImpl{}

		gameSessionID := uuid.New()
		req := models.SettlementGenerationRequest{
			Type:           models.SettlementVillage,
			Region:         "Forest",
			PopulationSize: "",
			DangerLevel:    3,
		}

		service := NewSettlementGeneratorService(mockLLM, mockRepo)

		ctx := testutil.TestContext()
		settlement, err := service.GenerateSettlement(ctx, gameSessionID, req)

		require.Error(t, err)
		require.Nil(t, settlement)
		require.Contains(t, err.Error(), "failed to generate base settlement")
	})

	t.Run("repository save error", func(t *testing.T) {
		mockLLM := &MockLLMProvider{
			Response: `{"name": "Test Town", "description": "Test"}`,
		}
		mockRepo := &MockWorldBuildingRepositoryImpl{}

		gameSessionID := uuid.New()
		req := models.SettlementGenerationRequest{
			Type: models.SettlementTown,
		}

		mockRepo.On("CreateSettlement", mock.AnythingOfType("*models.Settlement")).Return(errors.New("db error"))

		service := NewSettlementGeneratorService(mockLLM, mockRepo)

		ctx := testutil.TestContext()
		settlement, err := service.GenerateSettlement(ctx, gameSessionID, req)

		require.Error(t, err)
		require.Nil(t, settlement)
		require.Contains(t, err.Error(), "failed to save settlement")
	})
}

func TestSettlementGeneratorService_HelperFunctions(t *testing.T) {
	service := &SettlementGeneratorService{}

	t.Run("determinePopulationSize", func(t *testing.T) {
		tests := []struct {
			settlementType models.SettlementType
			expected       string
		}{
			{models.SettlementHamlet, "small"},
			{models.SettlementVillage, "small"},
			{models.SettlementTown, "medium"},
			{models.SettlementCity, "large"},
			{models.SettlementMetropolis, "large"},
		}

		for _, tt := range tests {
			result := service.determinePopulationSize(tt.settlementType)
			require.Equal(t, tt.expected, result)
		}
	})

	t.Run("calculatePopulation", func(t *testing.T) {
		tests := []struct {
			name           string
			settlementType models.SettlementType
			size           string
			minExpected    int
			maxExpected    int
		}{
			{"small hamlet", models.SettlementHamlet, "small", 20, 30},
			{"medium town", models.SettlementTown, "medium", 800, 1200},
			{"large city", models.SettlementCity, "large", 8000, 12000},
			{"ruins", models.SettlementRuins, "small", 0, 0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Run multiple times to check randomness
				for i := 0; i < 10; i++ {
					result := service.calculatePopulation(tt.settlementType, tt.size)
					require.GreaterOrEqual(t, result, tt.minExpected)
					require.LessOrEqual(t, result, tt.maxExpected)
				}
			})
		}
	})

	t.Run("calculateNPCCount", func(t *testing.T) {
		tests := []struct {
			settlementType models.SettlementType
			minExpected    int
			maxExpected    int
		}{
			{models.SettlementHamlet, 3, 5},
			{models.SettlementVillage, 5, 7},
			{models.SettlementTown, 8, 10},
			{models.SettlementCity, 12, 14},
			{models.SettlementMetropolis, 20, 22},
			{models.SettlementRuins, 1, 3},
		}

		for _, tt := range tests {
			result := service.calculateNPCCount(tt.settlementType)
			require.GreaterOrEqual(t, result, tt.minExpected)
			require.LessOrEqual(t, result, tt.maxExpected)
		}
	})

	t.Run("calculateShopCount", func(t *testing.T) {
		tests := []struct {
			settlementType models.SettlementType
			minExpected    int
			maxExpected    int
		}{
			{models.SettlementHamlet, 1, 2},
			{models.SettlementVillage, 2, 3},
			{models.SettlementTown, 4, 5},
			{models.SettlementCity, 8, 9},
			{models.SettlementMetropolis, 15, 16},
			{models.SettlementRuins, 0, 1},
		}

		for _, tt := range tests {
			result := service.calculateShopCount(tt.settlementType)
			require.GreaterOrEqual(t, result, tt.minExpected)
			require.LessOrEqual(t, result, tt.maxExpected)
		}
	})

	t.Run("calculateWealthLevel", func(t *testing.T) {
		tests := []struct {
			name           string
			settlementType models.SettlementType
			population     int
			minExpected    int
			maxExpected    int
		}{
			{"poor hamlet", models.SettlementHamlet, 50, 1, 4},
			{"average town", models.SettlementTown, 1000, 4, 6},
			{"wealthy city", models.SettlementCity, 15000, 7, 10},
			{"ruins", models.SettlementRuins, 0, 1, 2},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Run multiple times due to randomness
				for i := 0; i < 10; i++ {
					result := service.calculateWealthLevel(tt.settlementType, tt.population)
					require.GreaterOrEqual(t, result, tt.minExpected)
					require.LessOrEqual(t, result, tt.maxExpected)
					require.GreaterOrEqual(t, result, 1)
					require.LessOrEqual(t, result, 10)
				}
			})
		}
	})

	t.Run("inferTerrainType", func(t *testing.T) {
		tests := []struct {
			region   string
			expected string
		}{
			{"Northern Mountains", "mountainous"},
			{"Dark Forest", "forest"},
			{"Scorching Desert", "desert"},
			{"Coastal Trading Post", "coastal"},
			{"Murky Swamplands", "swamp"},
			{"Rolling Plains", "plains"},
			{"Random Region", "varied"},
		}

		for _, tt := range tests {
			result := service.inferTerrainType(tt.region)
			require.Equal(t, tt.expected, result)
		}
	})

	t.Run("inferClimate", func(t *testing.T) {
		tests := []struct {
			region   string
			expected string
		}{
			{"Northern Wastes", "cold"},
			{"Southern Jungles", "tropical"},
			{"Desert Expanse", "arid"},
			{"Swamp Delta", "humid"},
			{"Central Valley", "temperate"},
		}

		for _, tt := range tests {
			result := service.inferClimate(tt.region)
			require.Equal(t, tt.expected, result)
		}
	})

	t.Run("getNPCRoles", func(t *testing.T) {
		tests := []struct {
			settlementType models.SettlementType
			expectedRoles  []string
		}{
			{
				models.SettlementHamlet,
				[]string{"farmer", "hunter"},
			},
			{
				models.SettlementCity,
				[]string{"noble", "guildmaster", "magistrate", "spy", "cultist"},
			},
			{
				models.SettlementRuins,
				[]string{"hermit", "scavenger", "mad prophet"},
			},
		}

		baseRoles := []string{"merchant", "guard", "innkeeper", "blacksmith", "priest"}

		for _, tt := range tests {
			result := service.getNPCRoles(tt.settlementType)

			// Check base roles are included
			for _, role := range baseRoles {
				require.Contains(t, result, role)
			}

			// Check specific roles are included
			for _, role := range tt.expectedRoles {
				require.Contains(t, result, role)
			}
		}
	})

	t.Run("getShopTypes", func(t *testing.T) {
		tests := []struct {
			settlementType models.SettlementType
			expectedShops  []string
		}{
			{
				models.SettlementVillage,
				[]string{"weaponsmith", "inn"},
			},
			{
				models.SettlementCity,
				[]string{"magic", "jeweler", "library", "herbalist"},
			},
			{
				models.SettlementMetropolis,
				[]string{"enchanter", "artificer", "grand bazaar", "auction house"},
			},
		}

		baseShops := []string{"general", "tavern"}

		for _, tt := range tests {
			result := service.getShopTypes(tt.settlementType)

			// Check base shops are included
			for _, shop := range baseShops {
				require.Contains(t, result, shop)
			}

			// Check specific shops are included
			for _, shop := range tt.expectedShops {
				require.Contains(t, result, shop)
			}
		}
	})
}

func TestSettlementGeneratorService_GenerateMarketConditions(t *testing.T) {
	service := &SettlementGeneratorService{}

	t.Run("basic market", func(t *testing.T) {
		settlement := &models.Settlement{
			ID:                 uuid.New(),
			WealthLevel:        5,
			CorruptionLevel:    3,
			AncientRuinsNearby: false,
		}

		market := service.generateMarketConditions(settlement)

		require.NotNil(t, market)
		require.Equal(t, settlement.ID, market.SettlementID)
		require.Equal(t, 1.0, market.FoodPriceModifier)
		// CommonGoodsModifier can be 0.9, 1.0, or 1.3 due to random economic conditions
		require.InDelta(t, 1.0, market.CommonGoodsModifier, 0.31)
		require.False(t, market.BlackMarketActive)
		require.False(t, market.ArtifactDealerPresent)
	})

	t.Run("poor settlement", func(t *testing.T) {
		settlement := &models.Settlement{
			ID:          uuid.New(),
			WealthLevel: 2,
		}

		market := service.generateMarketConditions(settlement)

		require.Greater(t, market.CommonGoodsModifier, 1.0)
		require.Greater(t, market.MagicalItemsModifier, 1.0)
	})

	t.Run("wealthy settlement", func(t *testing.T) {
		settlement := &models.Settlement{
			ID:          uuid.New(),
			WealthLevel: 8,
		}

		market := service.generateMarketConditions(settlement)

		// Wealthy settlements normally have lower prices (modifier < 1.0)
		// but can have higher prices during economic depression
		if market.EconomicDepression {
			// During depression, prices can be higher
			require.Greater(t, market.CommonGoodsModifier, 1.0)
		} else if market.EconomicBoom {
			// During boom, prices should be even lower
			require.Less(t, market.CommonGoodsModifier, 0.9)
		} else {
			// Normal conditions for wealthy settlement
			require.Less(t, market.CommonGoodsModifier, 1.0)
		}
		
		// Magical items are not affected by economic conditions in the implementation
		require.Less(t, market.MagicalItemsModifier, 1.0)
	})

	t.Run("corrupted settlement", func(t *testing.T) {
		settlement := &models.Settlement{
			ID:              uuid.New(),
			WealthLevel:     5,
			CorruptionLevel: 7,
		}

		market := service.generateMarketConditions(settlement)

		require.True(t, market.BlackMarketActive)
	})

	t.Run("settlement near ruins", func(t *testing.T) {
		settlement := &models.Settlement{
			ID:                 uuid.New(),
			WealthLevel:        5,
			AncientRuinsNearby: true,
		}

		// Run multiple times due to randomness
		hasArtifactDealer := false
		for i := 0; i < 20; i++ {
			market := service.generateMarketConditions(settlement)
			if market.ArtifactDealerPresent {
				hasArtifactDealer = true
				require.Less(t, market.AncientArtifactsModifier, 2.0)
				break
			}
		}
		require.True(t, hasArtifactDealer, "Should have artifact dealer at least once in 20 runs")
	})
}

func TestSettlementGeneratorService_ProceduralGenerators(t *testing.T) {
	service := &SettlementGeneratorService{}

	t.Run("generateProceduralNPC", func(t *testing.T) {
		settlement := &models.Settlement{
			ID:              uuid.New(),
			CorruptionLevel: 7,
		}

		npc := service.generateProceduralNPC(settlement, "merchant")

		require.NotNil(t, npc)
		require.Equal(t, settlement.ID, npc.SettlementID)
		require.Equal(t, "merchant", npc.Role)
		require.Equal(t, "merchant", npc.Occupation)
		require.NotEmpty(t, npc.Name)
		require.NotEmpty(t, npc.Race)
		require.GreaterOrEqual(t, npc.Level, 1)
		require.LessOrEqual(t, npc.Level, 5)

		// Check JSONB fields are initialized
		require.NotNil(t, npc.PersonalityTraits)
		require.NotNil(t, npc.Ideals)
		require.NotNil(t, npc.Bonds)
		require.NotNil(t, npc.Flaws)
	})

	t.Run("generateProceduralShop", func(t *testing.T) {
		settlement := &models.Settlement{
			ID:                 uuid.New(),
			CorruptionLevel:    8,
			AncientRuinsNearby: true,
		}

		shop := service.generateProceduralShop(settlement, "weaponsmith")

		require.NotNil(t, shop)
		require.Equal(t, settlement.ID, shop.SettlementID)
		require.Equal(t, "weaponsmith", shop.Type)
		require.NotEmpty(t, shop.Name)
		require.GreaterOrEqual(t, shop.QualityLevel, 3)
		require.LessOrEqual(t, shop.QualityLevel, 7)
		require.GreaterOrEqual(t, shop.PriceModifier, 0.9)
		require.LessOrEqual(t, shop.PriceModifier, 1.2)

		// Check JSONB fields are initialized
		require.NotNil(t, shop.AvailableItems)
		require.NotNil(t, shop.SpecialItems)
		require.NotNil(t, shop.CurrentRumors)
	})
}

func TestSettlementGeneratorService_ConcurrentGeneration(t *testing.T) {
	mockLLM := &MockLLMProvider{
		Response: `{"name": "Concurrent Town", "description": "Test town"}`,
	}
	mockRepo := &MockWorldBuildingRepositoryImpl{}

	// Mock all repository calls to succeed
	mockRepo.On("CreateSettlement", mock.AnythingOfType("*models.Settlement")).Return(nil)
	mockRepo.On("CreateSettlementNPC", mock.AnythingOfType("*models.SettlementNPC")).Return(nil).Maybe()
	mockRepo.On("CreateSettlementShop", mock.AnythingOfType("*models.SettlementShop")).Return(nil).Maybe()
	mockRepo.On("CreateOrUpdateMarket", mock.AnythingOfType("*models.Market")).Return(nil)

	service := NewSettlementGeneratorService(mockLLM, mockRepo)

	// Run multiple generations concurrently
	const numGoroutines = 5
	errors := make(chan error, numGoroutines)
	settlements := make(chan *models.Settlement, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			gameSessionID := uuid.New()
			req := models.SettlementGenerationRequest{
				Type:   models.SettlementTown,
				Region: "Test Region",
			}

			ctx := context.Background()
			settlement, err := service.GenerateSettlement(ctx, gameSessionID, req)
			if err != nil {
				errors <- err
			} else {
				settlements <- settlement
			}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-errors:
			require.NoError(t, err)
		case settlement := <-settlements:
			require.NotNil(t, settlement)
			successCount++
		}
	}

	require.Equal(t, numGoroutines, successCount)
}

// Integration test
func TestSettlementGeneratorService_Integration(t *testing.T) {
	t.Run("complete settlement generation flow", func(t *testing.T) {
		mockLLM := &MockLLMProvider{}
		mockRepo := &MockWorldBuildingRepositoryImpl{}

		gameSessionID := uuid.New()
		req := models.SettlementGenerationRequest{
			Type:             models.SettlementCity,
			Region:           "Ancient Coastline",
			PopulationSize:   "large",
			DangerLevel:      7,
			AncientInfluence: true,
			SpecialFeatures:  []string{"port", "ancient temple", "thieves guild"},
		}

		// Complex settlement response
		settlementResponse := `{
			"name": "Port Dagon",
			"description": "A sprawling port city built atop sunken ruins",
			"history": "Founded by sailors who discovered the sunken temple of an elder god",
			"governmentType": "Merchant Council with Secret Cult Influence",
			"alignment": "Chaotic Neutral",
			"ageCategory": "ancient",
			"notableLocations": [
				{"name": "The Sunken Temple", "description": "Partially submerged temple to forgotten sea god"},
				{"name": "Smuggler's Wharf", "description": "Black market hub"},
				{"name": "The Leviathan's Rest", "description": "Inn frequented by deep sea fishermen"}
			],
			"problems": [
				{"title": "Cult Activity", "description": "Dagon cultists kidnapping citizens"},
				{"title": "Sea Monster Attacks", "description": "Ships vanishing in the harbor"}
			],
			"secrets": [
				{"title": "The Dreaming Deep", "description": "Ancient entity slumbers beneath the city"},
				{"title": "Hybrid Citizens", "description": "Some residents have Deep One ancestry"}
			],
			"defenses": ["Sea walls", "Harbor chain", "Mercenary fleet"],
			"primaryExports": ["Fish", "Pearls", "Ancient artifacts"],
			"primaryImports": ["Grain", "Weapons", "Fresh water"],
			"ancientRuinsNearby": true,
			"corruptionLevel": 8,
			"eldritchInfluence": 9
		}`

		// Set response
		mockLLM.Response = settlementResponse

		// Mock repository
		var createdSettlement *models.Settlement
		mockRepo.On("CreateSettlement", mock.AnythingOfType("*models.Settlement")).Run(func(args mock.Arguments) {
			createdSettlement = args.Get(0).(*models.Settlement)
			createdSettlement.ID = uuid.New()
		}).Return(nil)

		mockRepo.On("CreateSettlementNPC", mock.AnythingOfType("*models.SettlementNPC")).Return(nil).Maybe()
		mockRepo.On("CreateSettlementShop", mock.AnythingOfType("*models.SettlementShop")).Return(nil).Maybe()
		mockRepo.On("CreateOrUpdateMarket", mock.AnythingOfType("*models.Market")).Return(nil)

		service := NewSettlementGeneratorService(mockLLM, mockRepo)

		// Execute
		ctx := testutil.TestContext()
		settlement, err := service.GenerateSettlement(ctx, gameSessionID, req)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, settlement)
		require.Equal(t, "Port Dagon", settlement.Name)
		require.Equal(t, models.SettlementCity, settlement.Type)
		require.Equal(t, 8, settlement.CorruptionLevel)
		require.Equal(t, 9, settlement.EldritchInfluence)
		require.True(t, settlement.AncientRuinsNearby)
		require.Equal(t, "coastal", settlement.TerrainType)

		// Verify JSONB fields
		var notableLocations []map[string]string
		err = json.Unmarshal(settlement.NotableLocations, &notableLocations)
		require.NoError(t, err)
		require.Len(t, notableLocations, 3)

		mockRepo.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkSettlementGeneratorService_GenerateSettlement(b *testing.B) {
	mockLLM := &MockLLMProvider{
		Response: `{"name": "Benchmark Town", "description": "Test"}`,
	}
	mockRepo := &MockWorldBuildingRepositoryImpl{}

	mockRepo.On("CreateSettlement", mock.AnythingOfType("*models.Settlement")).Return(nil)
	mockRepo.On("CreateSettlementNPC", mock.AnythingOfType("*models.SettlementNPC")).Return(nil).Maybe()
	mockRepo.On("CreateSettlementShop", mock.AnythingOfType("*models.SettlementShop")).Return(nil).Maybe()
	mockRepo.On("CreateOrUpdateMarket", mock.AnythingOfType("*models.Market")).Return(nil)

	service := NewSettlementGeneratorService(mockLLM, mockRepo)

	gameSessionID := uuid.New()
	req := models.SettlementGenerationRequest{
		Type:   models.SettlementTown,
		Region: "Test",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateSettlement(ctx, gameSessionID, req)
	}
}

func BenchmarkSettlementGeneratorService_CalculatePopulation(b *testing.B) {
	service := &SettlementGeneratorService{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.calculatePopulation(models.SettlementCity, "large")
	}
}

func BenchmarkSettlementGeneratorService_GenerateMarketConditions(b *testing.B) {
	service := &SettlementGeneratorService{}
	settlement := &models.Settlement{
		ID:                 uuid.New(),
		WealthLevel:        5,
		CorruptionLevel:    5,
		AncientRuinsNearby: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.generateMarketConditions(settlement)
	}
}
