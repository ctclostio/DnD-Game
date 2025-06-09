package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

func TestDMAssistantService_ProcessRequest(t *testing.T) {
	t.Run("successful NPC dialogue generation", func(t *testing.T) {
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		sessionID := uuid.New()
		
		req := models.DMAssistantRequest{
			Type:          models.RequestTypeNPCDialogue,
			GameSessionID: sessionID.String(),
			Context:       "Tavern conversation",
			Parameters: map[string]interface{}{
				"npcName":        "Bartender Bob",
				"npcPersonality": []interface{}{"friendly", "gossipy"},
				"dialogueStyle":  "casual",
				"situation":      "player asking about local rumors",
				"playerInput":    "What's the latest gossip?",
			},
		}
		
		expectedDialogue := "Well, funny you should ask! Just this morning..."
		
		mockAI.On("GenerateNPCDialogue", ctx, mock.MatchedBy(func(req models.NPCDialogueRequest) bool {
			return req.NPCName == "Bartender Bob" &&
				len(req.NPCPersonality) == 2 &&
				req.PlayerInput == "What's the latest gossip?"
		})).Return(expectedDialogue, nil)
		
		mockRepo.On("SaveHistory", ctx, mock.MatchedBy(func(h *models.DMAssistantHistory) bool {
			return h.GameSessionID == sessionID &&
				h.UserID == userID &&
				h.RequestType == models.RequestTypeNPCDialogue &&
				h.Response == expectedDialogue
		})).Return(nil)
		
		result, err := service.ProcessRequest(ctx, userID, req)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		
		dialogue, ok := result.(map[string]string)
		require.True(t, ok)
		require.Equal(t, expectedDialogue, dialogue["dialogue"])
		
		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful location description generation", func(t *testing.T) {
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		sessionID := uuid.New()
		
		req := models.DMAssistantRequest{
			Type:          models.RequestTypeLocationDesc,
			GameSessionID: sessionID.String(),
			Context:       "Party entering new area",
			Parameters: map[string]interface{}{
				"locationType":    "dungeon",
				"locationName":    "The Forgotten Crypt",
				"atmosphere":      "eerie",
				"timeOfDay":       "eternal darkness",
				"specialFeatures": []interface{}{"ancient altar", "mysterious runes"},
			},
		}
		
		expectedLocation := &models.AILocation{
			ID:           uuid.New(),
			Name:         "The Forgotten Crypt",
			Type:         "dungeon",
			Description:  "The air is thick with decay...",
			Atmosphere:   "eerie",
			NPCs:         []uuid.UUID{},
		}
		
		mockAI.On("GenerateLocationDescription", ctx, mock.MatchedBy(func(req models.LocationDescriptionRequest) bool {
			return req.LocationType == "dungeon" &&
				req.LocationName == "The Forgotten Crypt" &&
				len(req.SpecialFeatures) == 2
		})).Return(expectedLocation, nil)
		
		mockRepo.On("SaveLocation", ctx, mock.MatchedBy(func(loc *models.AILocation) bool {
			return loc.GameSessionID == sessionID &&
				loc.CreatedBy == userID &&
				loc.Name == "The Forgotten Crypt"
		})).Return(nil)
		
		mockRepo.On("SaveHistory", ctx, mock.AnythingOfType("*models.DMAssistantHistory")).Return(nil)
		
		result, err := service.ProcessRequest(ctx, userID, req)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		
		location, ok := result.(*models.AILocation)
		require.True(t, ok)
		require.Equal(t, "The Forgotten Crypt", location.Name)
		require.Equal(t, sessionID, location.GameSessionID)
		
		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful combat narration generation", func(t *testing.T) {
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		sessionID := uuid.New()
		
		req := models.DMAssistantRequest{
			Type:          models.RequestTypeCombatNarration,
			GameSessionID: sessionID.String(),
			Context:       "Epic battle",
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
		
		mockAI.On("GenerateCombatNarration", ctx, mock.MatchedBy(func(req models.CombatNarrationRequest) bool {
			return req.AttackerName == "Aragorn" &&
				req.IsCritical == true &&
				req.Damage == 15
		})).Return(expectedNarration, nil)
		
		mockRepo.On("SaveNarration", ctx, mock.MatchedBy(func(n *models.AINarration) bool {
			return n.GameSessionID == sessionID &&
				n.Type == models.NarrationTypeCombatCritical &&
				n.Narration == expectedNarration
		})).Return(nil)
		
		mockRepo.On("SaveHistory", ctx, mock.AnythingOfType("*models.DMAssistantHistory")).Return(nil)
		
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
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		sessionID := uuid.New()
		
		req := models.DMAssistantRequest{
			Type:          models.RequestTypePlotTwist,
			GameSessionID: sessionID.String(),
			Context:       "Party just defeated the villain",
		}
		
		expectedTwist := &models.AIStoryElement{
			ID:          uuid.New(),
			Type:        models.StoryElementPlotTwist,
			Title:       "The True Puppet Master",
			Description: "The villain was being controlled...",
			Impact:      "major",
			Tags:        []string{"betrayal", "revelation"},
		}
		
		mockAI.On("GeneratePlotTwist", ctx, req.Context).Return(expectedTwist, nil)
		
		mockRepo.On("SaveStoryElement", ctx, mock.MatchedBy(func(elem *models.AIStoryElement) bool {
			return elem.GameSessionID == sessionID &&
				elem.CreatedBy == userID &&
				elem.Type == models.StoryElementPlotTwist
		})).Return(nil)
		
		mockRepo.On("SaveHistory", ctx, mock.AnythingOfType("*models.DMAssistantHistory")).Return(nil)
		
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
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		sessionID := uuid.New()
		locationID := uuid.New()
		
		req := models.DMAssistantRequest{
			Type:          models.RequestTypeEnvironmentalHazard,
			GameSessionID: sessionID.String(),
			Context:       "Dangerous dungeon room",
			Parameters: map[string]interface{}{
				"locationType": "dungeon",
				"difficulty":   float64(3),
				"locationId":   locationID.String(),
			},
		}
		
		expectedHazard := &models.AIEnvironmentalHazard{
			ID:          uuid.New(),
			Name:        "Poison Dart Trap",
			Type:        "trap",
			Description: "Hidden pressure plates trigger poison darts",
			Trigger:     "pressure",
			Effect:      "2d4 poison damage, DC 15 CON save",
			DC:          15,
			Damage:      "2d4",
			IsActive:    true,
		}
		
		mockAI.On("GenerateEnvironmentalHazard", ctx, "dungeon", 3).Return(expectedHazard, nil)
		
		mockRepo.On("SaveEnvironmentalHazard", ctx, mock.MatchedBy(func(h *models.AIEnvironmentalHazard) bool {
			return h.GameSessionID == sessionID &&
				h.CreatedBy == userID &&
				h.LocationID != nil &&
				*h.LocationID == locationID
		})).Return(nil)
		
		mockRepo.On("SaveHistory", ctx, mock.AnythingOfType("*models.DMAssistantHistory")).Return(nil)
		
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
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		
		req := models.DMAssistantRequest{
			Type:          models.RequestTypeNPCDialogue,
			GameSessionID: "invalid-uuid",
			Context:       "Test",
			Parameters:    map[string]interface{}{},
		}
		
		_, err := service.ProcessRequest(ctx, userID, req)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid game session ID")
	})

	t.Run("unknown request type", func(t *testing.T) {
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		sessionID := uuid.New()
		
		req := models.DMAssistantRequest{
			Type:          "unknown_type",
			GameSessionID: sessionID.String(),
			Context:       "Test",
			Parameters:    map[string]interface{}{},
		}
		
		_, err := service.ProcessRequest(ctx, userID, req)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown request type")
	})

	t.Run("missing required parameters", func(t *testing.T) {
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		userID := uuid.New()
		sessionID := uuid.New()
		
		req := models.DMAssistantRequest{
			Type:          models.RequestTypeNPCDialogue,
			GameSessionID: sessionID.String(),
			Context:       "Test",
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
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		sessionID := uuid.New()
		userID := uuid.New()
		role := "merchant"
		context := map[string]interface{}{
			"location": "marketplace",
			"goods":    "exotic weapons",
		}
		
		expectedNPC := &models.AINPC{
			ID:          uuid.New(),
			Name:        "Thoran Ironforge",
			Role:        role,
			Description: "A gruff dwarven weaponsmith",
			Personality: []string{"gruff", "honest", "skilled"},
			Backstory:   "Former royal armorer...",
		}
		
		mockAI.On("GenerateNPC", ctx, role, context).Return(expectedNPC, nil)
		
		mockRepo.On("SaveNPC", ctx, mock.MatchedBy(func(npc *models.AINPC) bool {
			return npc.GameSessionID == sessionID &&
				npc.CreatedBy == userID &&
				npc.Name == "Thoran Ironforge"
		})).Return(nil)
		
		result, err := service.CreateNPC(ctx, sessionID, userID, role, context)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "Thoran Ironforge", result.Name)
		require.Equal(t, sessionID, result.GameSessionID)
		require.Equal(t, userID, result.CreatedBy)
		require.NotZero(t, result.CreatedAt)
		
		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("AI generation failure", func(t *testing.T) {
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
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
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
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

func TestDMAssistantService_UpdateNPCDialogue(t *testing.T) {
	t.Run("successful dialogue update", func(t *testing.T) {
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		npcID := uuid.New()
		dialogue := "The dragon? Aye, I've seen it fly over the mountains."
		context := "Player asking about dragon sightings"
		
		mockRepo.On("AddNPCDialogue", ctx, npcID, mock.MatchedBy(func(entry models.DialogueEntry) bool {
			return entry.Dialogue == dialogue &&
				entry.Context == context &&
				!entry.Timestamp.IsZero()
		})).Return(nil)
		
		err := service.UpdateNPCDialogue(ctx, npcID, dialogue, context)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestDMAssistantService_GetUnusedStoryElements(t *testing.T) {
	t.Run("retrieve unused story elements", func(t *testing.T) {
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		sessionID := uuid.New()
		
		expectedElements := []*models.AIStoryElement{
			{
				ID:    uuid.New(),
				Type:  models.StoryElementPlotTwist,
				Title: "Hidden Betrayal",
			},
			{
				ID:    uuid.New(),
				Type:  models.StoryElementLoreReveal,
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
		mockRepo := new(MockDMAssistantRepository)
		mockAI := new(MockAIDMAssistantService)
		
		service := NewDMAssistantService(mockRepo, mockAI)
		
		ctx := testutil.TestContext()
		hazardID := uuid.New()
		
		mockRepo.On("TriggerHazard", ctx, hazardID).Return(nil)
		
		err := service.TriggerHazard(ctx, hazardID)
		
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

// Mock implementations
type MockDMAssistantRepository struct {
	mock.Mock
}

func (m *MockDMAssistantRepository) SaveHistory(ctx context.Context, history *models.DMAssistantHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockDMAssistantRepository) SaveNPC(ctx context.Context, npc *models.AINPC) error {
	args := m.Called(ctx, npc)
	return args.Error(0)
}

func (m *MockDMAssistantRepository) GetNPCByID(ctx context.Context, id uuid.UUID) (*models.AINPC, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AINPC), args.Error(1)
}

func (m *MockDMAssistantRepository) GetNPCsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AINPC, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AINPC), args.Error(1)
}

func (m *MockDMAssistantRepository) AddNPCDialogue(ctx context.Context, npcID uuid.UUID, entry models.DialogueEntry) error {
	args := m.Called(ctx, npcID, entry)
	return args.Error(0)
}

func (m *MockDMAssistantRepository) SaveLocation(ctx context.Context, location *models.AILocation) error {
	args := m.Called(ctx, location)
	return args.Error(0)
}

func (m *MockDMAssistantRepository) GetLocationByID(ctx context.Context, id uuid.UUID) (*models.AILocation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AILocation), args.Error(1)
}

func (m *MockDMAssistantRepository) GetLocationsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AILocation, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AILocation), args.Error(1)
}

func (m *MockDMAssistantRepository) SaveStoryElement(ctx context.Context, element *models.AIStoryElement) error {
	args := m.Called(ctx, element)
	return args.Error(0)
}

func (m *MockDMAssistantRepository) GetUnusedStoryElements(ctx context.Context, sessionID uuid.UUID) ([]*models.AIStoryElement, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AIStoryElement), args.Error(1)
}

func (m *MockDMAssistantRepository) MarkStoryElementUsed(ctx context.Context, elementID uuid.UUID) error {
	args := m.Called(ctx, elementID)
	return args.Error(0)
}

func (m *MockDMAssistantRepository) SaveNarration(ctx context.Context, narration *models.AINarration) error {
	args := m.Called(ctx, narration)
	return args.Error(0)
}

func (m *MockDMAssistantRepository) SaveEnvironmentalHazard(ctx context.Context, hazard *models.AIEnvironmentalHazard) error {
	args := m.Called(ctx, hazard)
	return args.Error(0)
}

func (m *MockDMAssistantRepository) GetActiveHazardsByLocation(ctx context.Context, locationID uuid.UUID) ([]*models.AIEnvironmentalHazard, error) {
	args := m.Called(ctx, locationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AIEnvironmentalHazard), args.Error(1)
}

func (m *MockDMAssistantRepository) TriggerHazard(ctx context.Context, hazardID uuid.UUID) error {
	args := m.Called(ctx, hazardID)
	return args.Error(0)
}

type MockAIDMAssistantService struct {
	mock.Mock
}

func (m *MockAIDMAssistantService) GenerateNPCDialogue(ctx context.Context, req models.NPCDialogueRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockAIDMAssistantService) GenerateLocationDescription(ctx context.Context, req models.LocationDescriptionRequest) (*models.AILocation, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AILocation), args.Error(1)
}

func (m *MockAIDMAssistantService) GenerateCombatNarration(ctx context.Context, req models.CombatNarrationRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockAIDMAssistantService) GeneratePlotTwist(ctx context.Context, context string) (*models.AIStoryElement, error) {
	args := m.Called(ctx, context)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AIStoryElement), args.Error(1)
}

func (m *MockAIDMAssistantService) GenerateEnvironmentalHazard(ctx context.Context, locationType string, difficulty int) (*models.AIEnvironmentalHazard, error) {
	args := m.Called(ctx, locationType, difficulty)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AIEnvironmentalHazard), args.Error(1)
}

func (m *MockAIDMAssistantService) GenerateNPC(ctx context.Context, role string, context map[string]interface{}) (*models.AINPC, error) {
	args := m.Called(ctx, role, context)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AINPC), args.Error(1)
}