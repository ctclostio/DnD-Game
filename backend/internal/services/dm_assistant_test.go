package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services/mocks"
)

// Test constants
const (
	testLocationName = "The Forgotten Crypt"
	testNPCName      = "Thoran Ironforge"
	testTypeHistory  = "*models.DMAssistantHistory"
)

func TestDMAssistantService_ProcessRequest(t *testing.T) {
	t.Run("successful NPC dialog generation", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		userID := uuid.New()
		sessionID := uuid.New()

		req := models.DMAssistantRequest{
			Type:          models.RequestTypeNPCDialog,
			GameSessionID: sessionID.String(),
			Context:       map[string]interface{}{"situation": "Tavern conversation"},
			Parameters: map[string]interface{}{
				"npcName":        "Bartender Bob",
				"npcPersonality": []interface{}{"friendly", "gossipy"},
				"dialogStyle":    "casual",
				"situation":      "player asking about local rumors",
				"playerInput":    "What's the latest gossip?",
			},
		}

		expectedDialog := "Well, funny you should ask! Just this morning..."

		mockAI.On("GenerateNPCDialog", ctx, mock.MatchedBy(func(req *models.NPCDialogRequest) bool {
			return req.NPCName == "Bartender Bob" &&
				len(req.NPCPersonality) == 2 &&
				req.PlayerInput == "What's the latest gossip?"
		})).Return(expectedDialog, nil)

		mockRepo.On("SaveHistory", ctx, mock.MatchedBy(func(h *models.DMAssistantHistory) bool {
			return h.GameSessionID == sessionID &&
				h.UserID == userID &&
				h.RequestType == models.RequestTypeNPCDialog &&
				h.Response == expectedDialog
		})).Return(nil)

		result, err := service.ProcessRequest(ctx, userID, req)

		require.NoError(t, err)
		require.NotNil(t, result)

		dialog, ok := result.(map[string]string)
		require.True(t, ok)
		require.Equal(t, expectedDialog, dialog["dialog"])

		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful location description generation", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		userID := uuid.New()
		sessionID := uuid.New()

		req := models.DMAssistantRequest{
			Type:          models.RequestTypeLocationDesc,
			GameSessionID: sessionID.String(),
			Context:       map[string]interface{}{"situation": "Party entering new area"},
			Parameters: map[string]interface{}{
				"locationType":    "dungeon",
				"locationName":    testLocationName,
				"atmosphere":      "eerie",
				"timeOfDay":       "eternal darkness",
				"specialFeatures": []interface{}{"ancient altar", "mysterious runes"},
			},
		}

		expectedLocation := &models.AILocation{
			ID:          uuid.New(),
			Name:        testLocationName,
			Type:        "dungeon",
			Description: "The air is thick with decay...",
			Atmosphere:  "eerie",
			NPCsPresent: []uuid.UUID{},
		}

		mockAI.On("GenerateLocationDescription", ctx, mock.MatchedBy(func(req *models.LocationDescriptionRequest) bool {
			return req.LocationType == "dungeon" &&
				req.LocationName == testLocationName &&
				len(req.SpecialFeatures) == 2
		})).Return(expectedLocation, nil)

		mockRepo.On("SaveLocation", ctx, mock.MatchedBy(func(loc *models.AILocation) bool {
			return loc.GameSessionID == sessionID &&
				loc.CreatedBy == userID &&
				loc.Name == testLocationName
		})).Return(nil)

		mockRepo.On("SaveHistory", ctx, mock.AnythingOfType(testTypeHistory)).Return(nil)

		result, err := service.ProcessRequest(ctx, userID, req)

		require.NoError(t, err)
		require.NotNil(t, result)

		location, ok := result.(*models.AILocation)
		require.True(t, ok)
		require.Equal(t, testLocationName, location.Name)
		require.Equal(t, sessionID, location.GameSessionID)

		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful combat narration generation", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		userID := uuid.New()
		sessionID := uuid.New()

		req := models.DMAssistantRequest{
			Type:          models.RequestTypeCombatNarration,
			GameSessionID: sessionID.String(),
			Context:       map[string]interface{}{"situation": "Epic battle"},
			Parameters: map[string]interface{}{
				"attackerName":  "Aragorn",
				"targetName":    "Orc Warrior",
				"actionType":    "attack",
				"weaponOrSpell": "Longsword",
				"damage":        float64(15),
				"isHit":         true,
				"isCritical":    true,
				"targetHP":      float64(5),
				"targetMaxHP":   float64(20),
			},
		}

		expectedNarration := "Aragorn's blade finds its mark with devastating precision!"

		mockAI.On("GenerateCombatNarration", ctx, mock.MatchedBy(func(req *models.CombatNarrationRequest) bool {
			return req.AttackerName == "Aragorn" &&
				req.IsCritical == true &&
				req.Damage == 15
		})).Return(expectedNarration, nil)

		mockRepo.On("SaveNarration", ctx, mock.MatchedBy(func(n *models.AINarration) bool {
			return n.GameSessionID == sessionID &&
				n.Type == models.NarrationTypeCombatCritical &&
				n.Narration == expectedNarration
		})).Return(nil)

		mockRepo.On("SaveHistory", ctx, mock.AnythingOfType(testTypeHistory)).Return(nil)

		result, err := service.ProcessRequest(ctx, userID, req)

		require.NoError(t, err)
		require.NotNil(t, result)

		narration, ok := result.(map[string]string)
		require.True(t, ok)
		require.Equal(t, expectedNarration, narration["narration"])

		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful plot twist generation", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		userID := uuid.New()
		sessionID := uuid.New()

		contextMap := map[string]interface{}{
			"situation": "Party just defeated the villain",
		}

		req := models.DMAssistantRequest{
			Type:          models.RequestTypePlotTwist,
			GameSessionID: sessionID.String(),
			Context:       contextMap,
		}

		expectedTwist := &models.AIStoryElement{
			ID:          uuid.New(),
			Type:        models.StoryElementPlotTwist,
			Title:       "The True Puppet Master",
			Description: "The villain was being controlled...",
			ImpactLevel: models.ImpactLevelMajor,
		}

		mockAI.On("GeneratePlotTwist", ctx, contextMap).Return(expectedTwist, nil)

		mockRepo.On("SaveStoryElement", ctx, mock.MatchedBy(func(elem *models.AIStoryElement) bool {
			return elem.GameSessionID == sessionID &&
				elem.CreatedBy == userID &&
				elem.Type == models.StoryElementPlotTwist
		})).Return(nil)

		mockRepo.On("SaveHistory", ctx, mock.AnythingOfType(testTypeHistory)).Return(nil)

		result, err := service.ProcessRequest(ctx, userID, req)

		require.NoError(t, err)
		require.NotNil(t, result)

		twist, ok := result.(*models.AIStoryElement)
		require.True(t, ok)
		require.Equal(t, "The True Puppet Master", twist.Title)

		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful environmental hazard generation", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		userID := uuid.New()
		sessionID := uuid.New()
		locationID := uuid.New()

		req := models.DMAssistantRequest{
			Type:          models.RequestTypeEnvironmentalHazard,
			GameSessionID: sessionID.String(),
			Context:       map[string]interface{}{"situation": "Dangerous dungeon room"},
			Parameters: map[string]interface{}{
				"locationType": "dungeon",
				"difficulty":   float64(3),
				"locationId":   locationID.String(),
			},
		}

		expectedHazard := &models.AIEnvironmentalHazard{
			ID:                uuid.New(),
			Name:              "Poison Dart Trap",
			Description:       "Hidden pressure plates trigger poison darts",
			TriggerCondition:  "pressure",
			EffectDescription: "2d4 poison damage, DC 15 CON save",
			DifficultyClass:   15,
			DamageFormula:     "2d4",
			IsTrap:            true,
			IsActive:          true,
		}

		mockAI.On("GenerateEnvironmentalHazard", ctx, "dungeon", 3).Return(expectedHazard, nil)

		mockRepo.On("SaveEnvironmentalHazard", ctx, mock.MatchedBy(func(h *models.AIEnvironmentalHazard) bool {
			return h.GameSessionID == sessionID &&
				h.CreatedBy == userID &&
				h.LocationID != nil &&
				*h.LocationID == locationID
		})).Return(nil)

		mockRepo.On("SaveHistory", ctx, mock.AnythingOfType(testTypeHistory)).Return(nil)

		result, err := service.ProcessRequest(ctx, userID, req)

		require.NoError(t, err)
		require.NotNil(t, result)

		hazard, ok := result.(*models.AIEnvironmentalHazard)
		require.True(t, ok)
		require.Equal(t, "Poison Dart Trap", hazard.Name)
		require.NotNil(t, hazard.LocationID)
		require.Equal(t, locationID, *hazard.LocationID)

		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid game session ID", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		userID := uuid.New()

		req := models.DMAssistantRequest{
			Type:          models.RequestTypeNPCDialog,
			GameSessionID: "invalid-uuid",
			Context:       map[string]interface{}{},
			Parameters:    map[string]interface{}{},
		}

		_, err := service.ProcessRequest(ctx, userID, req)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid game session ID")
	})

	t.Run("unknown request type", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		userID := uuid.New()
		sessionID := uuid.New()

		req := models.DMAssistantRequest{
			Type:          "unknown_type",
			GameSessionID: sessionID.String(),
			Context:       map[string]interface{}{},
			Parameters:    map[string]interface{}{},
		}

		_, err := service.ProcessRequest(ctx, userID, req)

		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown request type")
	})

	t.Run("missing required parameters", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		userID := uuid.New()
		sessionID := uuid.New()

		req := models.DMAssistantRequest{
			Type:          models.RequestTypeNPCDialog,
			GameSessionID: sessionID.String(),
			Context:       map[string]interface{}{},
			Parameters:    map[string]interface{}{
				// Missing required npcName
			},
		}

		_, err := service.ProcessRequest(ctx, userID, req)

		require.Error(t, err)
		require.Contains(t, err.Error(), "npcName is required")
	})
}

func TestDMAssistantService_CreateNPC(t *testing.T) {
	t.Run("successful NPC creation", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()
		role := "merchant"
		context := map[string]interface{}{
			"location": "marketplace",
			"goods":    "exotic weapons",
		}

		expectedNPC := &models.AINPC{
			ID:                uuid.New(),
			Name:              testNPCName,
			Occupation:        role,
			Appearance:        "A gruff dwarven weaponsmith",
			PersonalityTraits: []string{"gruff", "honest", "skilled"},
			Motivations:       "Former royal armorer...",
		}

		mockAI.On("GenerateNPC", ctx, role, context).Return(expectedNPC, nil)

		mockRepo.On("SaveNPC", ctx, mock.MatchedBy(func(npc *models.AINPC) bool {
			return npc.GameSessionID == sessionID &&
				npc.CreatedBy == userID &&
				npc.Name == testNPCName
		})).Return(nil)

		result, err := service.CreateNPC(ctx, sessionID, userID, role, context)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, testNPCName, result.Name)
		require.Equal(t, sessionID, result.GameSessionID)
		require.Equal(t, userID, result.CreatedBy)
		require.NotZero(t, result.CreatedAt)

		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("AI generation failure", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()

		mockAI.On("GenerateNPC", ctx, "merchant", mock.Anything).
			Return(nil, errors.New("AI service unavailable"))

		_, err := service.CreateNPC(ctx, sessionID, userID, "merchant", nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to generate NPC")

		mockAI.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "SaveNPC")
	})

	t.Run("repository save failure", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()

		npc := &models.AINPC{
			ID:   uuid.New(),
			Name: "Test NPC",
		}

		mockAI.On("GenerateNPC", ctx, "merchant", mock.Anything).Return(npc, nil)
		mockRepo.On("SaveNPC", ctx, mock.Anything).Return(errors.New("database error"))

		_, err := service.CreateNPC(ctx, sessionID, userID, "merchant", nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to save NPC")

		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestDMAssistantService_UpdateNPCDialog(t *testing.T) {
	t.Run("successful dialog update", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		npcID := uuid.New()
		dialog := "The dragon? Aye, I've seen it fly over the mountains."
		context := "Player asking about dragon sightings"

		mockRepo.On("AddNPCDialog", ctx, npcID, mock.MatchedBy(func(entry models.DialogEntry) bool {
			return entry.Dialog == dialog &&
				entry.Context == context &&
				!entry.Timestamp.IsZero()
		})).Return(nil)

		err := service.UpdateNPCDialog(ctx, npcID, dialog, context)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestDMAssistantService_GetUnusedStoryElements(t *testing.T) {
	t.Run("retrieve unused story elements", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		sessionID := uuid.New()

		expectedElements := []*models.AIStoryElement{
			{
				ID:    uuid.New(),
				Type:  models.StoryElementPlotTwist,
				Title: "Hidden Betrayal",
			},
			{
				ID:    uuid.New(),
				Type:  models.StoryElementRevelation,
				Title: "Ancient Prophecy",
			},
		}

		mockRepo.On("GetUnusedStoryElements", ctx, sessionID).Return(expectedElements, nil)

		elements, err := service.GetUnusedStoryElements(ctx, sessionID)

		require.NoError(t, err)
		require.Len(t, elements, 2)
		require.Equal(t, "Hidden Betrayal", elements[0].Title)

		mockRepo.AssertExpectations(t)
	})
}

func TestDMAssistantService_TriggerHazard(t *testing.T) {
	t.Run("trigger environmental hazard", func(t *testing.T) {
		mockRepo := new(mocks.MockDMAssistantRepository)
		mockAI := new(mocks.MockAIDMAssistantService)

		service := NewDMAssistantService(mockRepo, mockAI)

		ctx := context.Background()
		hazardID := uuid.New()

		mockRepo.On("TriggerHazard", ctx, hazardID).Return(nil)

		err := service.TriggerHazard(ctx, hazardID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
