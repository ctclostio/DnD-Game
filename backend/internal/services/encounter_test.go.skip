package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// Mock implementations
type MockEncounterRepository struct {
	mock.Mock
}

func (m *MockEncounterRepository) Create(encounter *models.Encounter) error {
	args := m.Called(encounter)
	return args.Error(0)
}

func (m *MockEncounterRepository) GetByID(encounterID string) (*models.Encounter, error) {
	args := m.Called(encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Encounter), args.Error(1)
}

func (m *MockEncounterRepository) GetByGameSession(gameSessionID string) ([]*models.Encounter, error) {
	args := m.Called(gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Encounter), args.Error(1)
}

func (m *MockEncounterRepository) StartEncounter(encounterID string) error {
	args := m.Called(encounterID)
	return args.Error(0)
}

func (m *MockEncounterRepository) CompleteEncounter(encounterID string, outcome string) error {
	args := m.Called(encounterID, outcome)
	return args.Error(0)
}

func (m *MockEncounterRepository) CreateEvent(event *models.EncounterEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockEncounterRepository) GetObjectives(encounterID string) ([]*models.EncounterObjective, error) {
	args := m.Called(encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.EncounterObjective), args.Error(1)
}

type MockAIEncounterBuilder struct {
	mock.Mock
}

func (m *MockAIEncounterBuilder) GenerateEncounter(ctx context.Context, req EncounterRequest) (*models.Encounter, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Encounter), args.Error(1)
}

type MockCombatServiceForEncounter struct {
	mock.Mock
}

// Test helpers
func createTestEncounterService() (*EncounterService, *MockEncounterRepository, *MockAIEncounterBuilder, *MockCombatServiceForEncounter) {
	mockRepo := new(MockEncounterRepository)
	mockBuilder := new(MockAIEncounterBuilder)
	mockCombat := new(MockCombatServiceForEncounter)
	
	// Convert mock to actual repository type for the service
	service := NewEncounterService(
		(*database.EncounterRepository)(mockRepo),
		(*AIEncounterBuilder)(mockBuilder),
		(*CombatService)(mockCombat),
	)
	
	return service, mockRepo, mockBuilder, mockCombat
}

func createTestEncounter() *models.Encounter {
	return &models.Encounter{
		ID:            uuid.New().String(),
		GameSessionID: uuid.New().String(),
		Name:          "Goblin Ambush",
		Description:   "A group of goblins attacks the party",
		EncounterType: "combat",
		Difficulty:    "medium",
		Status:        "planned",
		Terrain:       "forest",
		Enemies: []models.Enemy{
			{
				Name:     "Goblin",
				Type:     "humanoid",
				CR:       "1/4",
				HP:       7,
				AC:       15,
				Count:    4,
				Attacks:  []models.Attack{{Name: "Scimitar", Bonus: 4, Damage: "1d6+2"}},
				Abilities: []string{"Nimble Escape"},
			},
		},
		TotalXP:      200,
		AdjustedXP:   300,
		ThreatRating: 0.75,
	}
}

// Tests for GenerateEncounter

func TestEncounterService_GenerateEncounter(t *testing.T) {
	gameSessionID := uuid.New().String()
	userID := uuid.New().String()

	tests := []struct {
		name          string
		request       EncounterRequest
		generatedEnc  *models.Encounter
		generateError error
		saveError     error
		expectError   bool
		validate      func(*testing.T, *models.Encounter)
	}{
		{
			name: "Successful Encounter Generation",
			request: EncounterRequest{
				PartyLevel:    5,
				PartySize:     4,
				Difficulty:    "medium",
				EncounterType: "combat",
				Environment:   "forest",
			},
			generatedEnc: createTestEncounter(),
			expectError:  false,
			validate: func(t *testing.T, enc *models.Encounter) {
				assert.Equal(t, gameSessionID, enc.GameSessionID)
				assert.Equal(t, userID, enc.CreatedBy)
				assert.Equal(t, "planned", enc.Status)
			},
		},
		{
			name: "AI Generation Error",
			request: EncounterRequest{
				PartyLevel: 5,
				Difficulty: "hard",
			},
			generateError: errors.New("AI service unavailable"),
			expectError:   true,
		},
		{
			name: "Repository Save Error",
			request: EncounterRequest{
				PartyLevel: 5,
				Difficulty: "easy",
			},
			generatedEnc: createTestEncounter(),
			saveError:    errors.New("database error"),
			expectError:  true,
		},
		{
			name: "Social Encounter Generation",
			request: EncounterRequest{
				PartyLevel:    3,
				PartySize:     5,
				Difficulty:    "medium",
				EncounterType: "social",
				Environment:   "tavern",
			},
			generatedEnc: &models.Encounter{
				ID:            uuid.New().String(),
				Name:          "Mysterious Stranger",
				EncounterType: "social",
				Description:   "A hooded figure approaches with information",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, mockBuilder, _ := createTestEncounterService()

			// Setup mocks
			mockBuilder.On("GenerateEncounter", mock.Anything, tt.request).
				Return(tt.generatedEnc, tt.generateError)

			if tt.generatedEnc != nil && tt.generateError == nil {
				mockRepo.On("Create", mock.AnythingOfType("*models.Encounter")).
					Return(tt.saveError)
			}

			// Execute
			encounter, err := service.GenerateEncounter(context.Background(), tt.request, gameSessionID, userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, encounter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, encounter)
				if tt.validate != nil {
					tt.validate(t, encounter)
				}
			}

			mockBuilder.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

// Tests for GetEncounter

func TestEncounterService_GetEncounter(t *testing.T) {
	encounterID := uuid.New().String()
	testEncounter := createTestEncounter()
	testEncounter.ID = encounterID

	objectives := []*models.EncounterObjective{
		{
			ID:          uuid.New().String(),
			EncounterID: encounterID,
			Type:        "defeat_enemies",
			Description: "Defeat all goblins",
			IsRequired:  true,
		},
	}

	tests := []struct {
		name           string
		encounterID    string
		repoEncounter  *models.Encounter
		repoError      error
		objectives     []*models.EncounterObjective
		objectivesErr  error
		expectError    bool
	}{
		{
			name:          "Successful Retrieval with Objectives",
			encounterID:   encounterID,
			repoEncounter: testEncounter,
			objectives:    objectives,
			expectError:   false,
		},
		{
			name:          "Successful Retrieval without Objectives",
			encounterID:   encounterID,
			repoEncounter: testEncounter,
			objectives:    nil,
			objectivesErr: errors.New("no objectives found"),
			expectError:   false,
		},
		{
			name:        "Encounter Not Found",
			encounterID: "non-existent",
			repoError:   errors.New("not found"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := createTestEncounterService()

			// Setup mocks
			mockRepo.On("GetByID", tt.encounterID).
				Return(tt.repoEncounter, tt.repoError)

			if tt.repoEncounter != nil && tt.repoError == nil {
				mockRepo.On("GetObjectives", tt.encounterID).
					Return(tt.objectives, tt.objectivesErr)
			}

			// Execute
			encounter, err := service.GetEncounter(context.Background(), tt.encounterID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, encounter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, encounter)
				assert.Equal(t, tt.encounterID, encounter.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Tests for GetEncountersBySession

func TestEncounterService_GetEncountersBySession(t *testing.T) {
	gameSessionID := uuid.New().String()
	encounters := []*models.Encounter{
		createTestEncounter(),
		createTestEncounter(),
	}

	tests := []struct {
		name          string
		gameSessionID string
		repoResult    []*models.Encounter
		repoError     error
		expectError   bool
		expectCount   int
	}{
		{
			name:          "Multiple Encounters",
			gameSessionID: gameSessionID,
			repoResult:    encounters,
			expectError:   false,
			expectCount:   2,
		},
		{
			name:          "No Encounters",
			gameSessionID: gameSessionID,
			repoResult:    []*models.Encounter{},
			expectError:   false,
			expectCount:   0,
		},
		{
			name:          "Repository Error",
			gameSessionID: gameSessionID,
			repoError:     errors.New("database error"),
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := createTestEncounterService()

			mockRepo.On("GetByGameSession", tt.gameSessionID).
				Return(tt.repoResult, tt.repoError)

			result, err := service.GetEncountersBySession(context.Background(), tt.gameSessionID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Tests for StartEncounter

func TestEncounterService_StartEncounter(t *testing.T) {
	tests := []struct {
		name          string
		encounterID   string
		encounter     *models.Encounter
		getError      error
		startError    error
		expectError   bool
		errorContains string
	}{
		{
			name:        "Successful Start",
			encounterID: "enc1",
			encounter: &models.Encounter{
				ID:     "enc1",
				Name:   "Test Encounter",
				Status: "planned",
			},
			expectError: false,
		},
		{
			name:        "Encounter Not Found",
			encounterID: "non-existent",
			getError:    errors.New("not found"),
			expectError: true,
			errorContains: "encounter not found",
		},
		{
			name:        "Already Started",
			encounterID: "enc2",
			encounter: &models.Encounter{
				ID:     "enc2",
				Status: "active",
			},
			expectError:   true,
			errorContains: "already started",
		},
		{
			name:        "Already Completed",
			encounterID: "enc3",
			encounter: &models.Encounter{
				ID:     "enc3",
				Status: "completed",
			},
			expectError:   true,
			errorContains: "already started",
		},
		{
			name:        "Repository Start Error",
			encounterID: "enc4",
			encounter: &models.Encounter{
				ID:     "enc4",
				Status: "planned",
			},
			startError:    errors.New("database error"),
			expectError:   true,
			errorContains: "failed to start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := createTestEncounterService()

			// Setup mocks
			mockRepo.On("GetByID", tt.encounterID).
				Return(tt.encounter, tt.getError)

			if tt.encounter != nil && tt.encounter.Status == "planned" && tt.getError == nil {
				mockRepo.On("StartEncounter", tt.encounterID).
					Return(tt.startError)

				if tt.startError == nil {
					mockRepo.On("CreateEvent", mock.AnythingOfType("*models.EncounterEvent")).
						Return(nil)
				}
			}

			// Execute
			err := service.StartEncounter(context.Background(), tt.encounterID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Tests for CompleteEncounter

func TestEncounterService_CompleteEncounter(t *testing.T) {
	tests := []struct {
		name          string
		encounterID   string
		outcome       string
		completeError error
		expectError   bool
	}{
		{
			name:        "Successful Victory",
			encounterID: "enc1",
			outcome:     "victory",
			expectError: false,
		},
		{
			name:        "Successful Defeat",
			encounterID: "enc2",
			outcome:     "defeat",
			expectError: false,
		},
		{
			name:        "Successful Retreat",
			encounterID: "enc3",
			outcome:     "retreat",
			expectError: false,
		},
		{
			name:          "Repository Error",
			encounterID:   "enc4",
			outcome:       "victory",
			completeError: errors.New("database error"),
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := createTestEncounterService()

			mockRepo.On("CompleteEncounter", tt.encounterID, tt.outcome).
				Return(tt.completeError)

			if tt.completeError == nil {
				mockRepo.On("CreateEvent", mock.AnythingOfType("*models.EncounterEvent")).
					Return(nil)
			}

			err := service.CompleteEncounter(context.Background(), tt.encounterID, tt.outcome)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Tests for ScaleEncounter

func TestEncounterService_ScaleEncounter(t *testing.T) {
	baseEncounter := &models.Encounter{
		ID:         "enc1",
		Name:       "Scalable Encounter",
		Difficulty: "medium",
		Enemies: []models.Enemy{
			{Name: "Goblin", Count: 4, HP: 7},
		},
		ScalingOptions: &models.ScalingOptions{
			Easy: models.ScalingAdjustment{
				EnemyCountMultiplier: 0.5,
				EnemyHPMultiplier:    0.8,
			},
			Medium: models.ScalingAdjustment{
				EnemyCountMultiplier: 1.0,
				EnemyHPMultiplier:    1.0,
			},
			Hard: models.ScalingAdjustment{
				EnemyCountMultiplier: 1.5,
				EnemyHPMultiplier:    1.2,
			},
			Deadly: models.ScalingAdjustment{
				EnemyCountMultiplier: 2.0,
				EnemyHPMultiplier:    1.5,
			},
		},
	}

	tests := []struct {
		name          string
		encounterID   string
		newDifficulty string
		encounter     *models.Encounter
		getError      error
		expectError   bool
		errorContains string
	}{
		{
			name:          "Scale to Easy",
			encounterID:   "enc1",
			newDifficulty: "easy",
			encounter:     baseEncounter,
			expectError:   false,
		},
		{
			name:          "Scale to Hard",
			encounterID:   "enc1",
			newDifficulty: "hard",
			encounter:     baseEncounter,
			expectError:   false,
		},
		{
			name:          "Scale to Deadly",
			encounterID:   "enc1",
			newDifficulty: "deadly",
			encounter:     baseEncounter,
			expectError:   false,
		},
		{
			name:          "Invalid Difficulty",
			encounterID:   "enc1",
			newDifficulty: "extreme",
			encounter:     baseEncounter,
			expectError:   true,
			errorContains: "invalid difficulty",
		},
		{
			name:        "Encounter Not Found",
			encounterID: "non-existent",
			getError:    errors.New("not found"),
			expectError: true,
			errorContains: "encounter not found",
		},
		{
			name:          "No Scaling Options",
			encounterID:   "enc2",
			newDifficulty: "hard",
			encounter: &models.Encounter{
				ID:             "enc2",
				ScalingOptions: nil,
			},
			expectError:   true,
			errorContains: "no scaling options",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := createTestEncounterService()

			mockRepo.On("GetByID", tt.encounterID).
				Return(tt.encounter, tt.getError)

			result, err := service.ScaleEncounter(context.Background(), tt.encounterID, tt.newDifficulty)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Edge case tests

func TestEncounterService_GenerateEncounter_EdgeCases(t *testing.T) {
	t.Run("Empty GameSessionID", func(t *testing.T) {
		service, mockRepo, mockBuilder, _ := createTestEncounterService()
		
		req := EncounterRequest{PartyLevel: 5}
		encounter := createTestEncounter()
		
		mockBuilder.On("GenerateEncounter", mock.Anything, req).Return(encounter, nil)
		mockRepo.On("Create", mock.AnythingOfType("*models.Encounter")).Return(nil)
		
		result, err := service.GenerateEncounter(context.Background(), req, "", "user1")
		
		assert.NoError(t, err)
		assert.Equal(t, "", result.GameSessionID)
	})

	t.Run("Empty UserID", func(t *testing.T) {
		service, mockRepo, mockBuilder, _ := createTestEncounterService()
		
		req := EncounterRequest{PartyLevel: 5}
		encounter := createTestEncounter()
		
		mockBuilder.On("GenerateEncounter", mock.Anything, req).Return(encounter, nil)
		mockRepo.On("Create", mock.AnythingOfType("*models.Encounter")).Return(nil)
		
		result, err := service.GenerateEncounter(context.Background(), req, "session1", "")
		
		assert.NoError(t, err)
		assert.Equal(t, "", result.CreatedBy)
	})
}

func TestEncounterService_CreateEvent_Validation(t *testing.T) {
	service, mockRepo, _, _ := createTestEncounterService()
	
	// Test that event descriptions are properly formatted
	encounter := &models.Encounter{
		ID:   "enc1",
		Name: "Test & Special <Characters>",
		Status: "planned",
	}
	
	mockRepo.On("GetByID", "enc1").Return(encounter, nil)
	mockRepo.On("StartEncounter", "enc1").Return(nil)
	mockRepo.On("CreateEvent", mock.MatchedBy(func(event *models.EncounterEvent) bool {
		return event.Description == "Encounter 'Test & Special <Characters>' has begun!"
	})).Return(nil)
	
	err := service.StartEncounter(context.Background(), "enc1")
	assert.NoError(t, err)
	
	mockRepo.AssertExpectations(t)
}

// Concurrent operation tests

func TestEncounterService_ConcurrentOperations(t *testing.T) {
	service, mockRepo, mockBuilder, _ := createTestEncounterService()
	
	// Setup mocks for concurrent calls
	mockBuilder.On("GenerateEncounter", mock.Anything, mock.Anything).
		Return(createTestEncounter(), nil)
	mockRepo.On("Create", mock.Anything).Return(nil)
	
	// Run multiple goroutines
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			req := EncounterRequest{
				PartyLevel: id,
				Difficulty: "medium",
			}
			_, err := service.GenerateEncounter(
				context.Background(),
				req,
				fmt.Sprintf("session%d", id),
				fmt.Sprintf("user%d", id),
			)
			assert.NoError(t, err)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}
	
	mockBuilder.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Benchmark tests

func BenchmarkEncounterService_GenerateEncounter(b *testing.B) {
	service, mockRepo, mockBuilder, _ := createTestEncounterService()
	
	encounter := createTestEncounter()
	req := EncounterRequest{
		PartyLevel: 5,
		PartySize:  4,
		Difficulty: "medium",
	}
	
	mockBuilder.On("GenerateEncounter", mock.Anything, req).Return(encounter, nil)
	mockRepo.On("Create", mock.Anything).Return(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateEncounter(context.Background(), req, "session1", "user1")
	}
}

func BenchmarkEncounterService_ScaleEncounter(b *testing.B) {
	service, mockRepo, _, _ := createTestEncounterService()
	
	encounter := &models.Encounter{
		ID: "enc1",
		ScalingOptions: &models.ScalingOptions{
			Hard: models.ScalingAdjustment{
				EnemyCountMultiplier: 1.5,
				EnemyHPMultiplier:    1.2,
			},
		},
		Enemies: []models.Enemy{
			{Name: "Goblin", Count: 4, HP: 7},
			{Name: "Hobgoblin", Count: 2, HP: 11},
		},
	}
	
	mockRepo.On("GetByID", "enc1").Return(encounter, nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ScaleEncounter(context.Background(), "enc1", "hard")
	}
}