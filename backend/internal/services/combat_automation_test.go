package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Using MockCombatAnalyticsRepository from combat_analytics_test.go

// Test helpers
func createTestCombatAutomationService() (*CombatAutomationService, *MockCombatAnalyticsRepository, *mocks.MockCharacterRepository, *mocks.MockNPCRepository) {
	mockCombatRepo := new(MockCombatAnalyticsRepository)
	mockCharRepo := new(mocks.MockCharacterRepository)
	mockNPCRepo := new(mocks.MockNPCRepository)

	service := NewCombatAutomationService(mockCombatRepo, mockCharRepo, mockNPCRepo)
	return service, mockCombatRepo, mockCharRepo, mockNPCRepo
}

func createTestCharacters(count int, level int) []*models.Character {
	chars := make([]*models.Character, count)
	for i := 0; i < count; i++ {
		chars[i] = &models.Character{
			ID:    uuid.New().String(),
			Name:  fmt.Sprintf("Character %d", i+1),
			Level: level,
			Attributes: models.Attributes{
				Strength:     14,
				Dexterity:    16,
				Constitution: 14,
				Intelligence: 12,
				Wisdom:       13,
				Charisma:     10,
			},
			MaxHitPoints: 10 + (level * 6),
			HitPoints:    10 + (level * 6),
			Class:        "Fighter",
		}
	}
	return chars
}

// Tests

func TestCombatAutomationService_AutoResolveCombat(t *testing.T) {
	sessionID := uuid.New()
	characters := createTestCharacters(4, 5)

	tests := []struct {
		name           string
		request        models.AutoResolveRequest
		setupMocks     func(*MockCombatAnalyticsRepository)
		expectError    bool
		validateResult func(*testing.T, *models.AutoCombatResolution)
	}{
		{
			name: "Easy Encounter - Decisive Victory",
			request: models.AutoResolveRequest{
				EncounterDifficulty: "easy",
				EnemyTypes: []models.EnemyInfo{
					{Name: "goblin", Count: 4, CR: "1/4"},
				},
				TerrainType:  "open_field",
				UseResources: true,
			},
			setupMocks: func(repo *MockCombatAnalyticsRepository) {
				repo.On("CreateAutoCombatResolution", mock.AnythingOfType("*models.AutoCombatResolution")).
					Return(nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *models.AutoCombatResolution) {
				assert.NotNil(t, result)
				assert.Equal(t, sessionID, result.GameSessionID)
				assert.Equal(t, "easy", result.EncounterDifficulty)
				assert.Equal(t, "quick", result.ResolutionType)
				assert.True(t, result.ExperienceAwarded > 0)
				assert.NotEmpty(t, result.NarrativeSummary)

				// Check that resources were used
				var resources map[string]interface{}
				err := json.Unmarshal([]byte(result.PartyResourcesUsed), &resources)
				assert.NoError(t, err)
				assert.Contains(t, resources, "hp_lost")
			},
		},
		{
			name: "Hard Encounter - Multiple Enemy Types",
			request: models.AutoResolveRequest{
				EncounterDifficulty: "hard",
				EnemyTypes: []models.EnemyInfo{
					{Name: "orc", Count: 3, CR: "1/2"},
					{Name: "ogre", Count: 1, CR: "2"},
				},
				TerrainType:  "forest",
				UseResources: true,
			},
			setupMocks: func(repo *MockCombatAnalyticsRepository) {
				repo.On("CreateAutoCombatResolution", mock.AnythingOfType("*models.AutoCombatResolution")).
					Return(nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *models.AutoCombatResolution) {
				assert.NotNil(t, result)
				assert.Equal(t, "hard", result.EncounterDifficulty)
				assert.True(t, result.RoundsSimulated >= 2)
			},
		},
		{
			name: "No Resource Usage",
			request: models.AutoResolveRequest{
				EncounterDifficulty: "medium",
				EnemyTypes: []models.EnemyInfo{
					{Name: "skeleton", Count: 6, CR: "1/4"},
				},
				TerrainType:  "dungeon",
				UseResources: false,
			},
			setupMocks: func(repo *MockCombatAnalyticsRepository) {
				repo.On("CreateAutoCombatResolution", mock.AnythingOfType("*models.AutoCombatResolution")).
					Return(nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *models.AutoCombatResolution) {
				var resources map[string]interface{}
				err := json.Unmarshal([]byte(result.PartyResourcesUsed), &resources)
				assert.NoError(t, err)
				assert.Contains(t, resources, "hp_lost")
				assert.NotContains(t, resources, "spell_slots_used")
			},
		},
		{
			name: "Repository Error",
			request: models.AutoResolveRequest{
				EncounterDifficulty: "easy",
				EnemyTypes: []models.EnemyInfo{
					{Name: "rat", Count: 10, CR: "0"},
				},
			},
			setupMocks: func(repo *MockCombatAnalyticsRepository) {
				repo.On("CreateAutoCombatResolution", mock.AnythingOfType("*models.AutoCombatResolution")).
					Return(errors.New("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockCombatRepo, _, _ := createTestCombatAutomationService()

			if tt.setupMocks != nil {
				tt.setupMocks(mockCombatRepo)
			}

			result, err := service.AutoResolveCombat(context.Background(), sessionID, characters, tt.request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			mockCombatRepo.AssertExpectations(t)
		})
	}
}

func TestCombatAutomationService_SmartInitiative(t *testing.T) {
	sessionID := uuid.New()

	tests := []struct {
		name            string
		request         models.SmartInitiativeRequest
		initiativeRules map[uuid.UUID]*models.SmartInitiativeRule
		setupMocks      func(*MockCombatAnalyticsRepository)
		expectError     bool
		validateResult  func(*testing.T, []models.InitiativeEntry)
	}{
		{
			name: "Basic Initiative Roll",
			request: models.SmartInitiativeRequest{
				CombatID: uuid.New(),
				Combatants: []models.InitiativeCombatant{
					{
						ID:                uuid.New().String(),
						Type:              "character",
						Name:              "Fighter",
						DexterityModifier: 2,
					},
					{
						ID:                uuid.New().String(),
						Type:              "npc",
						Name:              "Goblin",
						DexterityModifier: 1,
					},
				},
			},
			setupMocks: func(repo *MockCombatAnalyticsRepository) {
				repo.On("GetInitiativeRule", sessionID, mock.Anything).Return(nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, entries []models.InitiativeEntry) {
				assert.Len(t, entries, 2)
				// Check that entries are sorted (highest first)
				assert.GreaterOrEqual(t, entries[0].Initiative, entries[1].Initiative)

				// Bonuses should be 2 and 1 in any order
				bonuses := []int{entries[0].Bonus, entries[1].Bonus}
				sort.Ints(bonuses)
				assert.Equal(t, []int{1, 2}, bonuses)
			},
		},
		{
			name: "Initiative with Alert Feat",
			request: models.SmartInitiativeRequest{
				CombatID: uuid.New(),
				Combatants: []models.InitiativeCombatant{
					{
						ID:                uuid.New().String(),
						Type:              "character",
						Name:              "Rogue",
						DexterityModifier: 4,
					},
				},
			},
			setupMocks: func(repo *MockCombatAnalyticsRepository) {
				combatantID := mock.AnythingOfType("string")
				rule := &models.SmartInitiativeRule{
					BaseInitiativeBonus: 2,
					AlertFeat:           true,
				}
				repo.On("GetInitiativeRule", sessionID, combatantID).Return(rule, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, entries []models.InitiativeEntry) {
				assert.Len(t, entries, 1)
				// Base 4 + Alert 5 + Bonus 2 = 11
				assert.Equal(t, 11, entries[0].Bonus)
			},
		},
		{
			name: "Initiative with Advantage",
			request: models.SmartInitiativeRequest{
				CombatID: uuid.New(),
				Combatants: []models.InitiativeCombatant{
					{
						ID:                uuid.New().String(),
						Type:              "character",
						Name:              "Barbarian",
						DexterityModifier: 1,
					},
				},
			},
			setupMocks: func(repo *MockCombatAnalyticsRepository) {
				rule := &models.SmartInitiativeRule{
					AdvantageOnInitiative: true,
				}
				repo.On("GetInitiativeRule", sessionID, mock.Anything).Return(rule, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, entries []models.InitiativeEntry) {
				assert.Len(t, entries, 1)
				// Just check that it doesn't error with advantage
				assert.True(t, entries[0].Initiative >= entries[0].Bonus)
			},
		},
		{
			name: "Special Priority Rules",
			request: models.SmartInitiativeRequest{
				CombatID: uuid.New(),
				Combatants: []models.InitiativeCombatant{
					{
						ID:                uuid.New().String(),
						Type:              "npc",
						Name:              "Boss",
						DexterityModifier: 3,
					},
				},
			},
			setupMocks: func(repo *MockCombatAnalyticsRepository) {
				specialRules := json.RawMessage(`{"priority": 10}`)
				rule := &models.SmartInitiativeRule{
					SpecialRules: models.JSONB(specialRules),
				}
				repo.On("GetInitiativeRule", sessionID, mock.Anything).Return(rule, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, entries []models.InitiativeEntry) {
				assert.Len(t, entries, 1)
				// Priority should boost initiative significantly
				assert.Greater(t, entries[0].Initiative, 100)
			},
		},
		{
			name: "Multiple Combatants with Mixed Rules",
			request: models.SmartInitiativeRequest{
				CombatID: uuid.New(),
				Combatants: []models.InitiativeCombatant{
					{
						ID:                uuid.New().String(),
						Type:              "character",
						Name:              "Wizard",
						DexterityModifier: 0,
					},
					{
						ID:                uuid.New().String(),
						Type:              "character",
						Name:              "Ranger",
						DexterityModifier: 3,
					},
					{
						ID:                uuid.New().String(),
						Type:              "npc",
						Name:              "Orc",
						DexterityModifier: -1,
					},
				},
			},
			setupMocks: func(repo *MockCombatAnalyticsRepository) {
				// First combatant has no special rules
				repo.On("GetInitiativeRule", sessionID, mock.Anything).Return(nil, nil).Once()

				// Second combatant has alert feat
				rule := &models.SmartInitiativeRule{AlertFeat: true}
				repo.On("GetInitiativeRule", sessionID, mock.Anything).Return(rule, nil).Once()

				// Third combatant has no special rules
				repo.On("GetInitiativeRule", sessionID, mock.Anything).Return(nil, nil).Once()
			},
			expectError: false,
			validateResult: func(t *testing.T, entries []models.InitiativeEntry) {
				assert.Len(t, entries, 3)
				// Should be sorted by initiative
				for i := 0; i < len(entries)-1; i++ {
					assert.GreaterOrEqual(t, entries[i].Initiative, entries[i+1].Initiative)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockCombatRepo, _, _ := createTestCombatAutomationService()

			if tt.setupMocks != nil {
				tt.setupMocks(mockCombatRepo)
			}

			entries, err := service.SmartInitiative(context.Background(), sessionID, tt.request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, entries)
				}
			}

			mockCombatRepo.AssertExpectations(t)
		})
	}
}

func TestCombatAutomationService_BattleMapOperations(t *testing.T) {
	sessionID := uuid.New()
	mapID := uuid.New()

	battleMap := &models.BattleMap{
		ID:                  mapID,
		GameSessionID:       sessionID,
		LocationDescription: "Goblin Cave",
		GridSizeX:           20,
		GridSizeY:           15,
	}

	t.Run("SaveBattleMap", func(t *testing.T) {
		service, mockCombatRepo, _, _ := createTestCombatAutomationService()

		mockCombatRepo.On("CreateBattleMap", battleMap).Return(nil)

		err := service.SaveBattleMap(context.Background(), battleMap)
		assert.NoError(t, err)

		mockCombatRepo.AssertExpectations(t)
	})

	t.Run("SaveBattleMap_Error", func(t *testing.T) {
		service, mockCombatRepo, _, _ := createTestCombatAutomationService()

		mockCombatRepo.On("CreateBattleMap", battleMap).Return(errors.New("database error"))

		err := service.SaveBattleMap(context.Background(), battleMap)
		assert.Error(t, err)

		mockCombatRepo.AssertExpectations(t)
	})

	t.Run("GetBattleMap", func(t *testing.T) {
		service, mockCombatRepo, _, _ := createTestCombatAutomationService()

		mockCombatRepo.On("GetBattleMap", mapID).Return(battleMap, nil)

		result, err := service.GetBattleMap(context.Background(), mapID)
		assert.NoError(t, err)
		assert.Equal(t, battleMap, result)

		mockCombatRepo.AssertExpectations(t)
	})

	t.Run("GetBattleMapsBySession", func(t *testing.T) {
		service, mockCombatRepo, _, _ := createTestCombatAutomationService()

		maps := []*models.BattleMap{battleMap}
		mockCombatRepo.On("GetBattleMapsBySession", sessionID).Return(maps, nil)

		result, err := service.GetBattleMapsBySession(context.Background(), sessionID)
		assert.NoError(t, err)
		assert.Equal(t, maps, result)

		mockCombatRepo.AssertExpectations(t)
	})
}

func TestCombatAutomationService_SetInitiativeRule(t *testing.T) {
	rule := &models.SmartInitiativeRule{
		ID:                    uuid.New(),
		GameSessionID:         uuid.New(),
		EntityID:              uuid.New().String(),
		BaseInitiativeBonus:   2,
		AlertFeat:             true,
		AdvantageOnInitiative: false,
	}

	t.Run("Success", func(t *testing.T) {
		service, mockCombatRepo, _, _ := createTestCombatAutomationService()

		mockCombatRepo.On("CreateOrUpdateInitiativeRule", rule).Return(nil)

		err := service.SetInitiativeRule(context.Background(), rule)
		assert.NoError(t, err)

		mockCombatRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		service, mockCombatRepo, _, _ := createTestCombatAutomationService()

		mockCombatRepo.On("CreateOrUpdateInitiativeRule", rule).Return(errors.New("database error"))

		err := service.SetInitiativeRule(context.Background(), rule)
		assert.Error(t, err)

		mockCombatRepo.AssertExpectations(t)
	})
}

func TestCombatAutomationService_GetAutoResolutionsBySession(t *testing.T) {
	sessionID := uuid.New()
	resolutions := []*models.AutoCombatResolution{
		{
			ID:            uuid.New(),
			GameSessionID: sessionID,
			Outcome:       "victory",
		},
	}

	service, mockCombatRepo, _, _ := createTestCombatAutomationService()

	mockCombatRepo.On("GetAutoCombatResolutionsBySession", sessionID).Return(resolutions, nil)

	result, err := service.GetAutoResolutionsBySession(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.Equal(t, resolutions, result)

	mockCombatRepo.AssertExpectations(t)
}

// Helper method tests

func TestCombatAutomationService_CalculateAveragePartyLevel(t *testing.T) {
	service, _, _, _ := createTestCombatAutomationService()

	tests := []struct {
		name       string
		characters []*models.Character
		expected   float64
	}{
		{
			name:       "Empty Party",
			characters: []*models.Character{},
			expected:   1,
		},
		{
			name:       "Single Character",
			characters: createTestCharacters(1, 5),
			expected:   5,
		},
		{
			name:       "Multiple Characters Same Level",
			characters: createTestCharacters(4, 10),
			expected:   10,
		},
		{
			name: "Mixed Levels",
			characters: []*models.Character{
				{Level: 5},
				{Level: 6},
				{Level: 7},
				{Level: 8},
			},
			expected: 6.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateAveragePartyLevel(tt.characters)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCombatAutomationService_ParseCR(t *testing.T) {
	service, _, _, _ := createTestCombatAutomationService()

	tests := []struct {
		cr       string
		expected float64
	}{
		{"1/8", 0.125},
		{"1/4", 0.25},
		{"1/2", 0.5},
		{"1", 1.0},
		{"5", 5.0},
		{"10", 10.0},
		{"20", 20.0},
		{"30", 30.0},
		{"invalid", 0.0},
		{"", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.cr, func(t *testing.T) {
			result := service.parseCR(tt.cr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCombatAutomationService_CalculateEncounterCR(t *testing.T) {
	service, _, _, _ := createTestCombatAutomationService()

	tests := []struct {
		name     string
		enemies  []models.EnemyInfo
		expected float64
	}{
		{
			name: "Single Enemy",
			enemies: []models.EnemyInfo{
				{Name: "goblin", Count: 1, CR: "1/4"},
			},
			expected: 0.25,
		},
		{
			name: "Multiple Same Enemy",
			enemies: []models.EnemyInfo{
				{Name: "goblin", Count: 4, CR: "1/4"},
			},
			expected: 1.0,
		},
		{
			name: "Multiple Enemy Types",
			enemies: []models.EnemyInfo{
				{Name: "goblin", Count: 2, CR: "1/4"},
				{Name: "hobgoblin", Count: 1, CR: "1/2"},
			},
			expected: 1.2, // (0.5 + 0.5) * 1.2 multiple enemy bonus
		},
		{
			name: "High CR Enemies",
			enemies: []models.EnemyInfo{
				{Name: "dragon", Count: 1, CR: "20"},
				{Name: "giant", Count: 2, CR: "10"},
			},
			expected: 48.0, // (20 + 20) * 1.2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateEncounterCR(tt.enemies)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests

func BenchmarkCombatAutomationService_AutoResolveCombat(b *testing.B) {
	service, mockCombatRepo, _, _ := createTestCombatAutomationService()
	sessionID := uuid.New()
	characters := createTestCharacters(4, 5)

	mockCombatRepo.On("CreateAutoCombatResolution", mock.AnythingOfType("*models.AutoCombatResolution")).
		Return(nil)

	request := models.AutoResolveRequest{
		EncounterDifficulty: "medium",
		EnemyTypes: []models.EnemyInfo{
			{Name: "orc", Count: 5, CR: "1/2"},
		},
		TerrainType:  "forest",
		UseResources: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.AutoResolveCombat(context.Background(), sessionID, characters, request)
	}
}

func BenchmarkCombatAutomationService_SmartInitiative(b *testing.B) {
	service, mockCombatRepo, _, _ := createTestCombatAutomationService()
	sessionID := uuid.New()

	// Setup mocks to return quickly
	mockCombatRepo.On("GetInitiativeRule", sessionID, mock.Anything).Return(nil, nil)

	request := models.SmartInitiativeRequest{
		CombatID: uuid.New(),
		Combatants: []models.InitiativeCombatant{
			{ID: uuid.New().String(), Type: "character", Name: "Fighter", DexterityModifier: 2},
			{ID: uuid.New().String(), Type: "character", Name: "Wizard", DexterityModifier: 1},
			{ID: uuid.New().String(), Type: "character", Name: "Rogue", DexterityModifier: 4},
			{ID: uuid.New().String(), Type: "character", Name: "Cleric", DexterityModifier: 0},
			{ID: uuid.New().String(), Type: "npc", Name: "Goblin1", DexterityModifier: 2},
			{ID: uuid.New().String(), Type: "npc", Name: "Goblin2", DexterityModifier: 2},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.SmartInitiative(context.Background(), sessionID, request)
	}
}
