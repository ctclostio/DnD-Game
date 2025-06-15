package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ctclostio/DnD-Game/backend/internal/database/interfaces"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/testhelpers"
)

// Focused mock for CombatAnalyticsInterface - only 4 methods to mock
type MockCombatAnalyticsInterface struct {
	mock.Mock
}

func (m *MockCombatAnalyticsInterface) CreateCombatAnalytics(analytics *models.CombatAnalytics) error {
	args := m.Called(analytics)
	return args.Error(0)
}

func (m *MockCombatAnalyticsInterface) GetCombatAnalytics(combatID uuid.UUID) (*models.CombatAnalytics, error) {
	args := m.Called(combatID)
	if analytics := args.Get(0); analytics != nil {
		return analytics.(*models.CombatAnalytics), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCombatAnalyticsInterface) GetCombatAnalyticsBySession(sessionID uuid.UUID) ([]*models.CombatAnalytics, error) {
	args := m.Called(sessionID)
	if analytics := args.Get(0); analytics != nil {
		return analytics.([]*models.CombatAnalytics), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCombatAnalyticsInterface) UpdateCombatAnalytics(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

// Focused mock for CombatantAnalyticsInterface - only 3 methods to mock
type MockCombatantAnalyticsInterface struct {
	mock.Mock
}

func (m *MockCombatantAnalyticsInterface) CreateCombatantAnalytics(analytics *models.CombatantAnalytics) error {
	args := m.Called(analytics)
	return args.Error(0)
}

func (m *MockCombatantAnalyticsInterface) GetCombatantAnalytics(combatAnalyticsID uuid.UUID) ([]*models.CombatantAnalytics, error) {
	args := m.Called(combatAnalyticsID)
	if analytics := args.Get(0); analytics != nil {
		return analytics.([]*models.CombatantAnalytics), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCombatantAnalyticsInterface) UpdateCombatantAnalytics(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

// Test demonstrating improved testability with focused interfaces
func TestCombatAnalyticsService_AnalyzeCombat_Refactored(t *testing.T) {
	// Setup - Notice how much simpler this is!
	mockAnalytics := new(MockCombatAnalyticsInterface)
	mockCombatants := new(MockCombatantAnalyticsInterface)
	mockHistory := new(MockCombatHistoryInterface) // Would implement 5 methods
	mockCombatService := new(MockCombatService)
	
	service := NewCombatAnalyticsService(
		mockAnalytics,
		mockCombatants,
		mockHistory,
		mockCombatService,
	)
	
	// Test data
	combatID := uuid.New()
	sessionID := uuid.New()
	
	testCombat := &models.Combat{
		ID:            combatID.String(),
		GameSessionID: sessionID.String(),
		Round:         5,
		StartedAt:     time.Now().Add(-10 * time.Minute),
		EndedAt:       &time.Time{},
		Combatants: []models.Combatant{
			{
				ID:               uuid.New().String(),
				Name:             "Fighter",
				Type:             models.CombatantTypePlayer,
				HitPoints:        15,
				MaxHP:            20,
				DamageDealt:      25,
				HealingDone:      0,
				ActionsPerformed: 5,
			},
			{
				ID:               uuid.New().String(),
				Name:             "Goblin",
				Type:             models.CombatantTypeMonster,
				HitPoints:        0,
				MaxHP:            10,
				DamageDealt:      5,
				ActionsPerformed: 3,
			},
		},
	}
	
	// Setup expectations - Only mock what we actually use!
	mockCombatService.On("GetCombatState", mock.Anything, combatID.String()).
		Return(testCombat, nil)
	
	mockAnalytics.On("CreateCombatAnalytics", mock.MatchedBy(func(a *models.CombatAnalytics) bool {
		return a.CombatID == combatID && a.TotalRounds == 5
	})).Return(nil)
	
	mockCombatants.On("CreateCombatantAnalytics", mock.AnythingOfType("*models.CombatantAnalytics")).
		Return(nil).Times(2)
	
	mockHistory.On("CreateCombatHistory", mock.AnythingOfType("*models.CombatHistory")).
		Return(nil)
	
	// Execute
	analytics, err := service.AnalyzeCombat(context.Background(), combatID)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, analytics)
	assert.Equal(t, combatID, analytics.CombatID)
	assert.Equal(t, 5, analytics.TotalRounds)
	
	// Verify all expectations met
	mockCombatService.AssertExpectations(t)
	mockAnalytics.AssertExpectations(t)
	mockCombatants.AssertExpectations(t)
	mockHistory.AssertExpectations(t)
}

// Comparison: Before refactoring (would need to mock 46 methods)
func TestCombatAnalyticsService_OldApproach(t *testing.T) {
	t.Skip("This shows the old approach - DON'T USE")
	
	// Before: Had to create a mock with 46 methods
	// mockRepo := new(MockCombatAnalyticsRepository)
	
	// Even for a simple test, you'd need to setup expectations for methods you don't use:
	// mockRepo.On("CreateBattleMap", mock.Anything).Return(nil).Maybe()
	// mockRepo.On("GetBattleMap", mock.Anything).Return(nil, errors.New("not used")).Maybe()
	// mockRepo.On("UpdateBattleMap", mock.Anything, mock.Anything).Return(nil).Maybe()
	// mockRepo.On("CreateAnimationPreset", mock.Anything).Return(nil).Maybe()
	// mockRepo.On("GetAnimationPreset", mock.Anything).Return(nil, errors.New("not used")).Maybe()
	// ... 41 more mock setups!
	
	// This made tests:
	// 1. Hard to write (which methods do I need to mock?)
	// 2. Brittle (changing interface breaks all tests)
	// 3. Unclear (what is this test actually testing?)
	// 4. Slow (setting up 46 mocks takes time)
}

// Example: Testing a service that only needs battle maps
func TestBattleMapService_CreateBattleMap_Refactored(t *testing.T) {
	// Only need to mock 5 battle map methods, not 46!
	mockBattleMaps := new(MockBattleMapInterface)
	service := NewBattleMapService(mockBattleMaps)
	
	combatID := uuid.New()
	config := models.BattleMapConfig{
		Width:    20,
		Height:   20,
		TileSize: 5,
		Terrain:  "forest",
	}
	
	// Setup expectation - clean and focused
	mockBattleMaps.On("CreateBattleMap", mock.MatchedBy(func(bm *models.BattleMap) bool {
		return bm.CombatID == combatID && 
			   bm.Width == 20 && 
			   bm.Height == 20 && 
			   bm.Terrain == "forest"
	})).Return(nil)
	
	// Execute
	battleMap, err := service.CreateBattleMap(context.Background(), combatID, config)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, battleMap)
	assert.Equal(t, 20, battleMap.Width)
	assert.Equal(t, 20, battleMap.Height)
	assert.Equal(t, "forest", battleMap.Terrain)
	
	mockBattleMaps.AssertExpectations(t)
}

// Example using test helpers with focused interfaces
func TestCombatAnalyticsService_WithTestHelpers(t *testing.T) {
	// Create mocks using test helpers
	mockAnalytics := new(MockCombatAnalyticsInterface)
	mockCombatants := new(MockCombatantAnalyticsInterface)
	mockHistory := new(MockCombatHistoryInterface)
	mockCombatService := new(MockCombatService)
	
	// Use test builders for test data
	testCombat := testhelpers.NewCombatBuilder().
		WithID(uuid.New().String()).
		WithGameSession(uuid.New().String()).
		WithStatus("completed").
		WithCombatant("Fighter", 15, true).
		WithCombatant("Goblin", 12, false).
		Build()
	
	// Setup mock calls using helpers
	testhelpers.SetupMockCalls(&mockCombatService.Mock, []testhelpers.MockCall{
		testhelpers.DataCall("GetCombatState", testCombat, mock.Anything, testCombat.ID),
	})
	
	testhelpers.SetupMockCalls(&mockAnalytics.Mock, []testhelpers.MockCall{
		testhelpers.SuccessCall("CreateCombatAnalytics", mock.AnythingOfType("*models.CombatAnalytics")),
	})
	
	// The test is now focused on behavior, not mock setup
	service := NewCombatAnalyticsService(
		mockAnalytics,
		mockCombatants,
		mockHistory,
		mockCombatService,
	)
	
	analytics, err := service.AnalyzeCombat(context.Background(), uuid.MustParse(testCombat.ID))
	
	assert.NoError(t, err)
	assert.NotNil(t, analytics)
	
	// Verify all mocks
	testhelpers.AssertMockExpectations(t, 
		mockAnalytics,
		mockCombatants,
		mockHistory,
		mockCombatService,
	)
}

// Benchmark showing performance improvement
func BenchmarkMockCreation(b *testing.B) {
	b.Run("Focused Interface Mock", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Creating a mock for 4-method interface
			mock := new(MockCombatAnalyticsInterface)
			mock.On("CreateCombatAnalytics", mock.Anything).Return(nil)
			_ = mock
		}
	})
	
	b.Run("Legacy Interface Mock", func(b *testing.B) {
		b.Skip("Would create mock with 46 methods - much slower")
		// This would be significantly slower due to:
		// 1. More memory allocation
		// 2. More method registration
		// 3. More reflection overhead
	})
}

// Mock for combat history (simplified for example)
type MockCombatHistoryInterface struct {
	mock.Mock
}

func (m *MockCombatHistoryInterface) CreateCombatHistory(history *models.CombatHistory) error {
	args := m.Called(history)
	return args.Error(0)
}

func (m *MockCombatHistoryInterface) GetCombatHistory(combatID uuid.UUID) (*models.CombatHistory, error) {
	args := m.Called(combatID)
	if history := args.Get(0); history != nil {
		return history.(*models.CombatHistory), args.Error(1)
	}
	return nil, args.Error(1)
}

// ... implement remaining 3 methods