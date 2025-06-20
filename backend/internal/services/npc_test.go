package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// Test constants
const (
	// IDs
	testNPCID      = "npc-1"
	testSessionID  = "session-1"
	testUserID     = "user-1"
	testInvalidID  = "invalid-id"
	testTemplateID = "template-goblin-boss"
	
	// Repository methods
	testMethodCreate  = "Create"
	testMethodUpdate  = "Update"
	testMethodGetByID = "GetByID"
	
	// NPC names and types
	testNPCOrc           = "Orc"
	testNPCGoblinWarrior = "Goblin Warrior"
	testNPCGoblinBoss    = "Goblin Boss"
	testNPCType          = "Humanoid"
	testNPCGeneric       = "Test NPC"
	
	// Damage types
	testDamageSlashing = "slashing"
	testDamageFire     = "fire"
	
	// Errors
	testErrNotFound = "not found"
)

func TestNPCService_CreateNPC(t *testing.T) {
	t.Run("successful NPC creation with all fields", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			Name:            testNPCGoblinWarrior,
			GameSessionID:   testSessionID,
			Type:            testNPCType,
			Size:            "Small",
			MaxHitPoints:    15,
			HitPoints:       15,
			ArmorClass:      13,
			ChallengeRating: 0.25,
			Attributes: models.Attributes{
				Strength:     8,
				Dexterity:    14,
				Constitution: 10,
				Intelligence: 10,
				Wisdom:       8,
				Charisma:     8,
			},
		}

		mockRepo.On(testMethodCreate, mock.Anything, mock.MatchedBy(func(n *models.NPC) bool {
			return n.Name == testNPCGoblinWarrior &&
				n.ExperiencePoints == 50 && // CR 0.25 = 50 XP
				n.HitPoints == 15
		})).Return(nil)

		err := service.CreateNPC(context.Background(), npc)

		require.NoError(t, err)
		require.Equal(t, 50, npc.ExperiencePoints)
		mockRepo.AssertExpectations(t)
	})

	t.Run("sets current HP to max if not specified", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			Name:          testNPCOrc,
			GameSessionID: testSessionID,
			MaxHitPoints:  42,
			HitPoints:     0, // Not specified
			Attributes: models.Attributes{
				Strength:     16,
				Dexterity:    12,
				Constitution: 16,
				Intelligence: 7,
				Wisdom:       11,
				Charisma:     10,
			},
		}

		mockRepo.On(testMethodCreate, mock.Anything, mock.MatchedBy(func(n *models.NPC) bool {
			return n.HitPoints == 42 // Should be set to max
		})).Return(nil)

		err := service.CreateNPC(context.Background(), npc)

		require.NoError(t, err)
		require.Equal(t, 42, npc.HitPoints)
		mockRepo.AssertExpectations(t)
	})

	t.Run("calculates saving throws if not provided", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			Name:          "Dragon Wyrmling",
			GameSessionID: testSessionID,
			MaxHitPoints:  52,
			Attributes: models.Attributes{
				Strength:     15, // +2 modifier
				Dexterity:    10, // +0 modifier
				Constitution: 13, // +1 modifier
				Intelligence: 10, // +0 modifier
				Wisdom:       11, // +0 modifier
				Charisma:     12, // +1 modifier
			},
		}

		mockRepo.On(testMethodCreate, mock.Anything, mock.MatchedBy(func(n *models.NPC) bool {
			return n.SavingThrows.Strength.Modifier == 2 &&
				n.SavingThrows.Constitution.Modifier == 1 &&
				n.SavingThrows.Charisma.Modifier == 1
		})).Return(nil)

		err := service.CreateNPC(context.Background(), npc)

		require.NoError(t, err)
		require.Equal(t, 2, npc.SavingThrows.Strength.Modifier)
		require.Equal(t, 0, npc.SavingThrows.Dexterity.Modifier)
		require.Equal(t, 1, npc.SavingThrows.Constitution.Modifier)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validates required fields", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		tests := []struct {
			name        string
			npc         *models.NPC
			expectedErr string
		}{
			{
				name: "missing name",
				npc: &models.NPC{
					GameSessionID: testSessionID,
					MaxHitPoints:  10,
				},
				expectedErr: "NPC name is required",
			},
			{
				name: "missing game session ID",
				npc: &models.NPC{
					Name:         testNPCGeneric,
					MaxHitPoints: 10,
				},
				expectedErr: "game session ID is required",
			},
			{
				name: "invalid max hit points",
				npc: &models.NPC{
					Name:          testNPCGeneric,
					GameSessionID: testSessionID,
					MaxHitPoints:  0,
				},
				expectedErr: "max hit points must be positive",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := service.CreateNPC(context.Background(), tt.npc)
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			})
		}
	})

	t.Run("corrects invalid attributes", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			Name:          "Weak Creature",
			GameSessionID: testSessionID,
			MaxHitPoints:  5,
			Attributes: models.Attributes{
				Strength:     0,  // Invalid, should be set to 10
				Dexterity:    -1, // Invalid, should be set to 10
				Constitution: 8,
				Intelligence: 10,
				Wisdom:       10,
				Charisma:     10,
			},
		}

		mockRepo.On(testMethodCreate, mock.Anything, mock.MatchedBy(func(n *models.NPC) bool {
			return n.Attributes.Strength == 10 &&
				n.Attributes.Dexterity == 10
		})).Return(nil)

		err := service.CreateNPC(context.Background(), npc)

		require.NoError(t, err)
		require.Equal(t, 10, npc.Attributes.Strength)
		require.Equal(t, 10, npc.Attributes.Dexterity)
		mockRepo.AssertExpectations(t)
	})

	t.Run("calculates correct XP from challenge rating", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		tests := []struct {
			cr       float64
			expected int
		}{
			{0, 10},
			{0.125, 25},
			{0.25, 50},
			{0.5, 100},
			{1, 200},
			{2, 450},
			{5, 1800},
			{10, 5900},
			{20, 25000},
		}

		for _, tt := range tests {
			npc := &models.NPC{
				Name:            "Test Creature",
				GameSessionID:   testSessionID,
				MaxHitPoints:    10,
				ChallengeRating: tt.cr,
			}

			mockRepo.On(testMethodCreate, mock.Anything, mock.MatchedBy(func(n *models.NPC) bool {
				return n.ExperiencePoints == tt.expected
			})).Return(nil).Once()

			err := service.CreateNPC(context.Background(), npc)
			require.NoError(t, err)
			require.Equal(t, tt.expected, npc.ExperiencePoints)
		}

		mockRepo.AssertExpectations(t)
	})
}

func TestNPCService_RollInitiative(t *testing.T) {
	t.Run("successful initiative roll", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			ID:   testNPCID,
			Name: "Quick Rogue",
			Attributes: models.Attributes{
				Dexterity: 18, // +4 modifier
			},
		}

		mockRepo.On(testMethodGetByID, mock.Anything, testNPCID).Return(npc, nil)

		initiative, err := service.RollInitiative(context.Background(), "npc-1")

		require.NoError(t, err)
		// Initiative should be between 5 (1+4) and 24 (20+4)
		require.GreaterOrEqual(t, initiative, 5)
		require.LessOrEqual(t, initiative, 24)
		mockRepo.AssertExpectations(t)
	})

	t.Run("NPC not found", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		mockRepo.On("GetByID", mock.Anything, testInvalidID).
			Return(nil, errors.New(testErrNotFound))

		_, err := service.RollInitiative(context.Background(), "invalid-id")

		require.Error(t, err)
		require.Contains(t, err.Error(), testErrNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestNPCService_ApplyDamage(t *testing.T) {
	t.Run("applies full damage", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			ID:                "npc-1",
			Name:              "Orc",
			HitPoints:         42,
			MaxHitPoints:      42,
			DamageResistances: []string{},
			DamageImmunities:  []string{},
		}

		mockRepo.On(testMethodGetByID, mock.Anything, testNPCID).Return(npc, nil)
		mockRepo.On(testMethodUpdate, mock.Anything, mock.MatchedBy(func(n *models.NPC) bool {
			return n.HitPoints == 27 // 42 - 15
		})).Return(nil)

		err := service.ApplyDamage(context.Background(), "npc-1", 15, "slashing")

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("applies reduced damage with resistance", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			ID:                "npc-1",
			Name:              "Skeleton",
			HitPoints:         13,
			MaxHitPoints:      13,
			DamageResistances: []string{"slashing", "piercing"},
			DamageImmunities:  []string{},
		}

		mockRepo.On(testMethodGetByID, mock.Anything, testNPCID).Return(npc, nil)
		mockRepo.On(testMethodUpdate, mock.Anything, mock.MatchedBy(func(n *models.NPC) bool {
			return n.HitPoints == 5 // 13 - (16/2)
		})).Return(nil)

		err := service.ApplyDamage(context.Background(), "npc-1", 16, "slashing")

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ignores damage with immunity", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			ID:                "npc-1",
			Name:              "Fire Elemental",
			HitPoints:         102,
			MaxHitPoints:      102,
			DamageResistances: []string{},
			DamageImmunities:  []string{"fire", "poison"},
		}

		mockRepo.On(testMethodGetByID, mock.Anything, testNPCID).Return(npc, nil)
		// Should not call Update since no damage is taken

		err := service.ApplyDamage(context.Background(), "npc-1", 50, "fire")

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("prevents HP from going below zero", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			ID:           "npc-1",
			Name:         "Wounded Goblin",
			HitPoints:    3,
			MaxHitPoints: 7,
		}

		mockRepo.On(testMethodGetByID, mock.Anything, testNPCID).Return(npc, nil)
		mockRepo.On(testMethodUpdate, mock.Anything, mock.MatchedBy(func(n *models.NPC) bool {
			return n.HitPoints == 0 // Should be 0, not negative
		})).Return(nil)

		err := service.ApplyDamage(context.Background(), "npc-1", 10, "slashing")

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestNPCService_HealNPC(t *testing.T) {
	tests := []struct {
		name           string
		npc            *models.NPC
		healAmount     int
		expectedHP     int
	}{
		{
			name: "heals NPC without exceeding max HP",
			npc: &models.NPC{
				ID:           "npc-1",
				Name:         "Wounded Orc",
				HitPoints:    20,
				MaxHitPoints: 42,
			},
			healAmount: 15,
			expectedHP: 35, // 20 + 15
		},
		{
			name: "caps healing at max HP",
			npc: &models.NPC{
				ID:           "npc-1",
				Name:         "Nearly Full HP Orc",
				HitPoints:    38,
				MaxHitPoints: 42,
			},
			healAmount: 20,
			expectedHP: 42, // Capped at max
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockNPCRepository)
			service := NewNPCService(mockRepo)

			mockRepo.On("GetByID", mock.Anything, tt.npc.ID).Return(tt.npc, nil)
			mockRepo.On(testMethodUpdate, mock.Anything, mock.MatchedBy(func(n *models.NPC) bool {
				return n.HitPoints == tt.expectedHP
			})).Return(nil)

			err := service.HealNPC(context.Background(), tt.npc.ID, tt.healAmount)

			require.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNPCService_UpdateNPC(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			ID:           "npc-1",
			Name:         "Updated NPC",
			HitPoints:    30,
			MaxHitPoints: 40,
		}

		mockRepo.On(testMethodUpdate, mock.Anything, npc).Return(nil)

		err := service.UpdateNPC(context.Background(), npc)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("caps HP at max during update", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			ID:           "npc-1",
			Name:         "Over-healed NPC",
			HitPoints:    50, // Over max
			MaxHitPoints: 40,
		}

		mockRepo.On(testMethodUpdate, mock.Anything, mock.MatchedBy(func(n *models.NPC) bool {
			return n.HitPoints == 40 // Should be capped
		})).Return(nil)

		err := service.UpdateNPC(context.Background(), npc)

		require.NoError(t, err)
		require.Equal(t, 40, npc.HitPoints)
		mockRepo.AssertExpectations(t)
	})

	t.Run("requires NPC ID", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		npc := &models.NPC{
			Name: "No ID NPC",
		}

		err := service.UpdateNPC(context.Background(), npc)

		require.Error(t, err)
		require.Contains(t, err.Error(), "NPC ID is required")
		mockRepo.AssertNotCalled(t, "Update")
	})
}

func TestNPCService_SearchNPCs(t *testing.T) {
	t.Run("search by filter", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		filter := models.NPCSearchFilter{
			GameSessionID: testSessionID,
			Type:          "Humanoid",
			MinCR:         1,
			MaxCR:         5,
		}

		expectedNPCs := []*models.NPC{
			{ID: "npc-1", Name: "Orc", ChallengeRating: 1},
			{ID: "npc-2", Name: "Hobgoblin Captain", ChallengeRating: 3},
		}

		mockRepo.On("Search", mock.Anything, &filter).Return(expectedNPCs, nil)

		npcs, err := service.SearchNPCs(context.Background(), &filter)

		require.NoError(t, err)
		require.Len(t, npcs, 2)
		require.Equal(t, "Orc", npcs[0].Name)
		mockRepo.AssertExpectations(t)
	})
}

func TestNPCService_CreateFromTemplate(t *testing.T) {
	t.Run("create from template", func(t *testing.T) {
		mockRepo := new(MockNPCRepository)
		service := NewNPCService(mockRepo)

		expectedNPC := &models.NPC{
			ID:            "new-npc-1",
			Name:          testNPCGoblinBoss,
			GameSessionID: testSessionID,
			MaxHitPoints:  21,
		}

		mockRepo.On("CreateFromTemplate",
			mock.Anything,
			testTemplateID,
			testSessionID,
			testUserID,
		).Return(expectedNPC, nil)

		npc, err := service.CreateFromTemplate(context.Background(), testTemplateID, testSessionID, testUserID)

		require.NoError(t, err)
		require.NotNil(t, npc)
		require.Equal(t, testNPCGoblinBoss, npc.Name)
		mockRepo.AssertExpectations(t)
	})
}

func TestNPCService_GetAbilityModifier(t *testing.T) {
	service := &NPCService{}

	tests := []struct {
		score    int
		expected int
	}{
		{1, -4}, // (1-10)/2 = -9/2 = -4 (rounds towards zero)
		{3, -3}, // (3-10)/2 = -7/2 = -3 (rounds towards zero)
		{6, -2},
		{8, -1},
		{10, 0},
		{11, 0},
		{12, 1},
		{14, 2},
		{16, 3},
		{18, 4},
		{20, 5},
		{30, 10},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("score_%d", tt.score), func(t *testing.T) {
			modifier := service.getAbilityModifier(tt.score)
			require.Equal(t, tt.expected, modifier)
		})
	}
}

func TestNPCService_GetProficiencyBonusFromCR(t *testing.T) {
	service := &NPCService{}

	tests := []struct {
		cr       float64
		expected int
	}{
		{0, 2},
		{0.25, 2},
		{1, 2},
		{4, 2},
		{5, 3},
		{8, 3},
		{9, 4},
		{12, 4},
		{13, 5},
		{16, 5},
		{17, 6},
		{20, 6},
		{21, 7},
		{24, 7},
		{25, 8},
		{28, 8},
		{29, 9},
		{30, 9},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("cr_%.2f", tt.cr), func(t *testing.T) {
			bonus := service.getProficiencyBonusFromCR(tt.cr)
			require.Equal(t, tt.expected, bonus)
		})
	}
}

// Mock implementations
type MockNPCRepository struct {
	mock.Mock
}

func (m *MockNPCRepository) Create(ctx context.Context, npc *models.NPC) error {
	args := m.Called(ctx, npc)
	return args.Error(0)
}

func (m *MockNPCRepository) GetByID(ctx context.Context, id string) (*models.NPC, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NPC), args.Error(1)
}

func (m *MockNPCRepository) GetByGameSession(ctx context.Context, gameSessionID string) ([]*models.NPC, error) {
	args := m.Called(ctx, gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NPC), args.Error(1)
}

func (m *MockNPCRepository) Update(ctx context.Context, npc *models.NPC) error {
	args := m.Called(ctx, npc)
	return args.Error(0)
}

func (m *MockNPCRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNPCRepository) Search(ctx context.Context, filter *models.NPCSearchFilter) ([]*models.NPC, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NPC), args.Error(1)
}

func (m *MockNPCRepository) GetTemplates(ctx context.Context) ([]*models.NPCTemplate, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NPCTemplate), args.Error(1)
}

func (m *MockNPCRepository) CreateFromTemplate(ctx context.Context, templateID, gameSessionID, createdBy string) (*models.NPC, error) {
	args := m.Called(ctx, templateID, gameSessionID, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NPC), args.Error(1)
}

func (m *MockNPCRepository) GetTemplateByID(ctx context.Context, id string) (*models.NPCTemplate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NPCTemplate), args.Error(1)
}
