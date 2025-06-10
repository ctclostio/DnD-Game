package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services/mocks"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

func TestCustomRaceService_CreateCustomRace(t *testing.T) {
	t.Run("successful race creation with auto-approval", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		
		request := models.CustomRaceRequest{
			Name:        "Shadow Elf",
			Description: "Elves touched by shadow magic, dwelling in twilight realms",
		}
		
		generatedRace := &models.CustomRaceGenerationResult{
			Name:        "Shadow Elf",
			Description: "Elves touched by shadow magic, with pale skin and dark eyes",
			AbilityScoreIncreases: map[string]int{
				"dexterity":    2,
				"intelligence": 1,
			},
			Size:  "Medium",
			Speed: 30,
			Traits: []models.RacialTrait{
				{
					Name:        "Shadow Step",
					Description: "As a bonus action, teleport up to 30 feet to an unoccupied space you can see in dim light or darkness",
				},
				{
					Name:        "Sunlight Sensitivity",
					Description: "Disadvantage on attack rolls and Perception checks in direct sunlight",
				},
			},
			Languages:      []string{"Common", "Elvish"},
			Darkvision:     120,
			Resistances:    []string{"necrotic"},
			BalanceScore:   6,
			BalanceExplanation: "Strong darkvision and teleportation balanced by sunlight weakness",
		}
		
		mockAI.On("GenerateCustomRace", ctx, request).Return(generatedRace, nil)
		
		mockRepo.On("Create", ctx, mock.MatchedBy(func(race *models.CustomRace) bool {
			return race.Name == "Shadow Elf" &&
				race.ApprovalStatus == models.ApprovalStatusApproved && // Auto-approved due to balance score <= 7
				race.CreatedBy == userID &&
				race.TimesUsed == 0 &&
				!race.IsPublic &&
				race.BalanceScore != nil && *race.BalanceScore == 6
		})).Return(nil)
		
		result, err := service.CreateCustomRace(ctx, userID, request)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "Shadow Elf", result.Name)
		require.Equal(t, models.ApprovalStatusApproved, result.ApprovalStatus)
		require.NotNil(t, result.ApprovalNotes)
		require.Contains(t, *result.ApprovalNotes, "Auto-approved")
		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful race creation pending approval", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		
		request := models.CustomRaceRequest{
			Name:        "Godborn",
			Description: "Descendants of deities with immense power",
		}
		
		generatedRace := &models.CustomRaceGenerationResult{
			Name:        "Godborn",
			Description: "Mortals with divine heritage, bearing marks of their celestial lineage",
			AbilityScoreIncreases: map[string]int{
				"strength": 2,
				"wisdom":   2,
			},
			Size:  "Medium",
			Speed: 35,
			Traits: []models.RacialTrait{
				{
					Name:        "Divine Resistance",
					Description: "Resistance to all damage types",
				},
				{
					Name:        "Immortal Blessing",
					Description: "Advantage on all saving throws",
				},
			},
			Languages:       []string{"Common", "Celestial", "Primordial"},
			Immunities:      []string{"poison"},
			BalanceScore:    9, // Too powerful, needs approval
			BalanceExplanation: "Extremely powerful abilities requiring DM oversight",
		}
		
		mockAI.On("GenerateCustomRace", ctx, request).Return(generatedRace, nil)
		
		mockRepo.On("Create", ctx, mock.MatchedBy(func(race *models.CustomRace) bool {
			return race.Name == "Godborn" &&
				race.ApprovalStatus == models.ApprovalStatusPending && // Pending due to high balance score
				race.ApprovalNotes == nil // No auto-approval notes
		})).Return(nil)
		
		result, err := service.CreateCustomRace(ctx, userID, request)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, models.ApprovalStatusPending, result.ApprovalStatus)
		require.Nil(t, result.ApprovalNotes)
		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("AI generation failure", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		
		request := models.CustomRaceRequest{
			Name:        "Test Race",
			Description: "Test description",
		}
		
		mockAI.On("GenerateCustomRace", ctx, request).
			Return(nil, errors.New("AI service unavailable"))
		
		result, err := service.CreateCustomRace(ctx, userID, request)
		
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to generate custom race")
		mockAI.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("repository save failure", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		
		request := models.CustomRaceRequest{
			Name:        "Test Race",
			Description: "Test description",
		}
		
		generatedRace := &models.CustomRaceGenerationResult{
			Name:         "Test Race",
			Description:  "Generated description",
			Size:         "Medium",
			Speed:        30,
			Traits:       []models.RacialTrait{{Name: "Test", Description: "Test trait"}},
			Languages:    []string{"Common"},
			BalanceScore: 5,
		}
		
		mockAI.On("GenerateCustomRace", ctx, request).Return(generatedRace, nil)
		mockRepo.On("Create", ctx, mock.Anything).Return(errors.New("database error"))
		
		result, err := service.CreateCustomRace(ctx, userID, request)
		
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to save custom race")
		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomRaceService_GetCustomRace(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		
		expectedRace := &models.CustomRace{
			ID:   raceID,
			Name: "Shadow Elf",
			AbilityScoreIncreases: map[string]int{
				"dexterity": 2,
			},
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(expectedRace, nil)
		
		result, err := service.GetCustomRace(ctx, raceID)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, expectedRace, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("race not found", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		
		mockRepo.On("GetByID", ctx, raceID).Return(nil, sql.ErrNoRows)
		
		result, err := service.GetCustomRace(ctx, raceID)
		
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "custom race not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		
		mockRepo.On("GetByID", ctx, raceID).Return(nil, errors.New("database error"))
		
		result, err := service.GetCustomRace(ctx, raceID)
		
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to get custom race")
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomRaceService_ApproveCustomRace(t *testing.T) {
	t.Run("successful approval", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		approverID := uuid.New()
		notes := "Well-balanced race, approved for play"
		
		existingRace := &models.CustomRace{
			ID:             raceID,
			Name:           "Shadow Elf",
			ApprovalStatus: models.ApprovalStatusPending,
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(existingRace, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(race *models.CustomRace) bool {
			return race.ID == raceID &&
				race.ApprovalStatus == models.ApprovalStatusApproved &&
				race.ApprovedBy != nil && *race.ApprovedBy == approverID &&
				race.ApprovalNotes != nil && *race.ApprovalNotes == notes
		})).Return(nil)
		
		err := service.ApproveCustomRace(ctx, raceID, approverID, notes)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("race not found", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		approverID := uuid.New()
		
		mockRepo.On("GetByID", ctx, raceID).Return(nil, errors.New("not found"))
		
		err := service.ApproveCustomRace(ctx, raceID, approverID, "notes")
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get custom race")
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomRaceService_RejectCustomRace(t *testing.T) {
	t.Run("successful rejection", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		approverID := uuid.New()
		notes := "Too powerful, breaks game balance"
		
		existingRace := &models.CustomRace{
			ID:             raceID,
			Name:           "Godborn",
			ApprovalStatus: models.ApprovalStatusPending,
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(existingRace, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(race *models.CustomRace) bool {
			return race.ApprovalStatus == models.ApprovalStatusRejected &&
				race.ApprovedBy != nil && *race.ApprovedBy == approverID &&
				race.ApprovalNotes != nil && *race.ApprovalNotes == notes
		})).Return(nil)
		
		err := service.RejectCustomRace(ctx, raceID, approverID, notes)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomRaceService_RequestRevision(t *testing.T) {
	t.Run("successful revision request", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		approverID := uuid.New()
		notes := "Reduce the power of Divine Resistance trait"
		
		existingRace := &models.CustomRace{
			ID:             raceID,
			Name:           "Godborn",
			ApprovalStatus: models.ApprovalStatusPending,
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(existingRace, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(race *models.CustomRace) bool {
			return race.ApprovalStatus == models.ApprovalStatusRevisionNeeded
		})).Return(nil)
		
		err := service.RequestRevision(ctx, raceID, approverID, notes)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomRaceService_MakePublic(t *testing.T) {
	t.Run("creator makes approved race public", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		userID := uuid.New()
		
		existingRace := &models.CustomRace{
			ID:             raceID,
			Name:           "Shadow Elf",
			CreatedBy:      userID,
			ApprovalStatus: models.ApprovalStatusApproved,
			IsPublic:       false,
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(existingRace, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(race *models.CustomRace) bool {
			return race.IsPublic == true
		})).Return(nil)
		
		err := service.MakePublic(ctx, raceID, userID)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("approver makes race public", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		creatorID := uuid.New()
		approverID := uuid.New()
		
		existingRace := &models.CustomRace{
			ID:             raceID,
			Name:           "Shadow Elf",
			CreatedBy:      creatorID,
			ApprovedBy:     &approverID,
			ApprovalStatus: models.ApprovalStatusApproved,
			IsPublic:       false,
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(existingRace, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(race *models.CustomRace) bool {
			return race.IsPublic == true
		})).Return(nil)
		
		err := service.MakePublic(ctx, raceID, approverID)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized user cannot make race public", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		creatorID := uuid.New()
		unauthorizedID := uuid.New()
		
		existingRace := &models.CustomRace{
			ID:             raceID,
			Name:           "Shadow Elf",
			CreatedBy:      creatorID,
			ApprovalStatus: models.ApprovalStatusApproved,
			IsPublic:       false,
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(existingRace, nil)
		
		err := service.MakePublic(ctx, raceID, unauthorizedID)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "unauthorized to make this race public")
		mockRepo.AssertExpectations(t)
	})

	t.Run("cannot make unapproved race public", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		userID := uuid.New()
		
		existingRace := &models.CustomRace{
			ID:             raceID,
			Name:           "Shadow Elf",
			CreatedBy:      userID,
			ApprovalStatus: models.ApprovalStatusPending,
			IsPublic:       false,
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(existingRace, nil)
		
		err := service.MakePublic(ctx, raceID, userID)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "only approved races can be made public")
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomRaceService_ValidateCustomRaceForCharacter(t *testing.T) {
	t.Run("creator can use their own race", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		userID := uuid.New()
		
		race := &models.CustomRace{
			ID:             raceID,
			CreatedBy:      userID,
			ApprovalStatus: models.ApprovalStatusApproved,
			IsPublic:       false,
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(race, nil)
		
		err := service.ValidateCustomRaceForCharacter(ctx, raceID, userID)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("anyone can use public approved race", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		creatorID := uuid.New()
		userID := uuid.New()
		
		race := &models.CustomRace{
			ID:             raceID,
			CreatedBy:      creatorID,
			ApprovalStatus: models.ApprovalStatusApproved,
			IsPublic:       true,
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(race, nil)
		
		err := service.ValidateCustomRaceForCharacter(ctx, raceID, userID)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("cannot use private race of another user", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		creatorID := uuid.New()
		userID := uuid.New()
		
		race := &models.CustomRace{
			ID:             raceID,
			CreatedBy:      creatorID,
			ApprovalStatus: models.ApprovalStatusApproved,
			IsPublic:       false,
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(race, nil)
		
		err := service.ValidateCustomRaceForCharacter(ctx, raceID, userID)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "you don't have permission to use this custom race")
		mockRepo.AssertExpectations(t)
	})

	t.Run("cannot use unapproved race", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		userID := uuid.New()
		
		race := &models.CustomRace{
			ID:             raceID,
			CreatedBy:      userID,
			ApprovalStatus: models.ApprovalStatusPending,
			IsPublic:       false,
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(race, nil)
		
		err := service.ValidateCustomRaceForCharacter(ctx, raceID, userID)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "this custom race has not been approved yet")
		mockRepo.AssertExpectations(t)
	})

	t.Run("race not found", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		userID := uuid.New()
		
		mockRepo.On("GetByID", ctx, raceID).Return(nil, errors.New("not found"))
		
		err := service.ValidateCustomRaceForCharacter(ctx, raceID, userID)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "custom race not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomRaceService_GetCustomRaceStats(t *testing.T) {
	t.Run("successful stats retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		
		race := &models.CustomRace{
			ID:   raceID,
			Name: "Shadow Elf",
			Description: "Elves touched by shadow magic",
			AbilityScoreIncreases: map[string]int{
				"dexterity":    2,
				"intelligence": 1,
			},
			Size:       "Medium",
			Speed:      30,
			Darkvision: 120,
			Languages:  []string{"Common", "Elvish"},
			Resistances: []string{"necrotic"},
			Traits: []models.RacialTrait{
				{
					Name:        "Shadow Step",
					Description: "Teleport in darkness",
				},
			},
		}
		
		mockRepo.On("GetByID", ctx, raceID).Return(race, nil)
		
		stats, err := service.GetCustomRaceStats(ctx, raceID)
		
		require.NoError(t, err)
		require.NotNil(t, stats)
		require.Equal(t, "Shadow Elf", stats["name"])
		require.Equal(t, "Medium", stats["size"])
		require.Equal(t, 30, stats["speed"])
		require.Equal(t, 120, stats["darkvision"])
		require.True(t, stats["isCustom"].(bool))
		require.Equal(t, raceID, stats["customRaceId"])
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomRaceService_GetPendingApproval(t *testing.T) {
	t.Run("retrieve pending races", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		
		pendingRaces := []*models.CustomRace{
			{ID: uuid.New(), Name: "Race 1", ApprovalStatus: models.ApprovalStatusPending},
			{ID: uuid.New(), Name: "Race 2", ApprovalStatus: models.ApprovalStatusPending},
		}
		
		mockRepo.On("GetPendingApproval", ctx).Return(pendingRaces, nil)
		
		result, err := service.GetPendingApproval(ctx)
		
		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Equal(t, pendingRaces, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomRaceService_IncrementUsage(t *testing.T) {
	t.Run("successful usage increment", func(t *testing.T) {
		mockRepo := new(mocks.MockCustomRaceRepository)
		mockAI := new(MockAIRaceGeneratorService)
		
		service := NewCustomRaceService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		raceID := uuid.New()
		
		mockRepo.On("IncrementUsage", ctx, raceID).Return(nil)
		
		err := service.IncrementUsage(ctx, raceID)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

// MockAIRaceGeneratorService is a mock implementation for AI race generation
type MockAIRaceGeneratorService struct {
	mock.Mock
}

func (m *MockAIRaceGeneratorService) GenerateCustomRace(ctx context.Context, request models.CustomRaceRequest) (*models.CustomRaceGenerationResult, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CustomRaceGenerationResult), args.Error(1)
}