package services

import (
	"context"
	"fmt"

	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/pkg/dice"
)

// NPCService handles NPC-related business logic
type NPCService struct {
	repo   database.NPCRepository
	roller *dice.Roller
}

// NewNPCService creates a new NPC service
func NewNPCService(repo database.NPCRepository) *NPCService {
	return &NPCService{
		repo:   repo,
		roller: dice.NewRoller(),
	}
}

// CreateNPC creates a new NPC
func (s *NPCService) CreateNPC(ctx context.Context, npc *models.NPC) error {
	// Validate NPC
	if npc.Name == "" {
		return fmt.Errorf("NPC name is required")
	}
	if npc.GameSessionID == "" {
		return fmt.Errorf("game session ID is required")
	}
	if npc.MaxHitPoints <= 0 {
		return fmt.Errorf("max hit points must be positive")
	}

	// Set current HP to max if not specified
	if npc.HitPoints == 0 {
		npc.HitPoints = npc.MaxHitPoints
	}

	// Calculate proficiency bonus based on CR
	npc.Attributes = s.ensureValidAttributes(npc.Attributes)

	// Calculate saving throws if not provided
	if !s.hasSavingThrows(npc.SavingThrows) {
		npc.SavingThrows = s.calculateSavingThrows(npc)
	}

	// Calculate XP if not provided
	if npc.ExperiencePoints == 0 {
		npc.ExperiencePoints = s.calculateXPFromCR(npc.ChallengeRating)
	}

	return s.repo.Create(ctx, npc)
}

// GetNPC retrieves an NPC by ID
func (s *NPCService) GetNPC(ctx context.Context, id string) (*models.NPC, error) {
	return s.repo.GetByID(ctx, id)
}

// GetNPCsByGameSession retrieves all NPCs for a game session
func (s *NPCService) GetNPCsByGameSession(ctx context.Context, gameSessionID string) ([]*models.NPC, error) {
	return s.repo.GetByGameSession(ctx, gameSessionID)
}

// UpdateNPC updates an existing NPC
func (s *NPCService) UpdateNPC(ctx context.Context, npc *models.NPC) error {
	// Validate NPC
	if npc.ID == "" {
		return fmt.Errorf("NPC ID is required")
	}

	// Ensure HP doesn't exceed max
	if npc.HitPoints > npc.MaxHitPoints {
		npc.HitPoints = npc.MaxHitPoints
	}

	return s.repo.Update(ctx, npc)
}

// DeleteNPC deletes an NPC
func (s *NPCService) DeleteNPC(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// SearchNPCs searches for NPCs based on filter criteria
func (s *NPCService) SearchNPCs(ctx context.Context, filter models.NPCSearchFilter) ([]*models.NPC, error) {
	return s.repo.Search(ctx, filter)
}

// GetTemplates retrieves all NPC templates
func (s *NPCService) GetTemplates(ctx context.Context) ([]*models.NPCTemplate, error) {
	return s.repo.GetTemplates(ctx)
}

// CreateFromTemplate creates a new NPC from a template
func (s *NPCService) CreateFromTemplate(ctx context.Context, templateID, gameSessionID, createdBy string) (*models.NPC, error) {
	return s.repo.CreateFromTemplate(ctx, templateID, gameSessionID, createdBy)
}

// RollInitiative rolls initiative for an NPC
func (s *NPCService) RollInitiative(ctx context.Context, npcID string) (int, error) {
	npc, err := s.GetNPC(ctx, npcID)
	if err != nil {
		return 0, err
	}

	// Roll 1d20 + Dexterity modifier
	dexMod := s.getAbilityModifier(npc.Attributes.Dexterity)
	roll, err := s.roller.Roll("1d20")
	if err != nil {
		return 0, err
	}

	return roll.Total + dexMod, nil
}

// ApplyDamage applies damage to an NPC
func (s *NPCService) ApplyDamage(ctx context.Context, npcID string, damage int, damageType string) error {
	npc, err := s.GetNPC(ctx, npcID)
	if err != nil {
		return err
	}

	// Check for damage immunity
	for _, immunity := range npc.DamageImmunities {
		if immunity == damageType {
			return nil // No damage taken
		}
	}

	// Check for damage resistance
	finalDamage := damage
	for _, resistance := range npc.DamageResistances {
		if resistance == damageType {
			finalDamage = damage / 2
			break
		}
	}

	// Apply damage
	npc.HitPoints -= finalDamage
	if npc.HitPoints < 0 {
		npc.HitPoints = 0
	}

	return s.UpdateNPC(ctx, npc)
}

// HealNPC heals an NPC
func (s *NPCService) HealNPC(ctx context.Context, npcID string, healing int) error {
	npc, err := s.GetNPC(ctx, npcID)
	if err != nil {
		return err
	}

	// Apply healing
	npc.HitPoints += healing
	if npc.HitPoints > npc.MaxHitPoints {
		npc.HitPoints = npc.MaxHitPoints
	}

	return s.UpdateNPC(ctx, npc)
}

// Helper functions

func (s *NPCService) ensureValidAttributes(attrs models.Attributes) models.Attributes {
	// Ensure all attributes are at least 1
	if attrs.Strength < 1 {
		attrs.Strength = 10
	}
	if attrs.Dexterity < 1 {
		attrs.Dexterity = 10
	}
	if attrs.Constitution < 1 {
		attrs.Constitution = 10
	}
	if attrs.Intelligence < 1 {
		attrs.Intelligence = 10
	}
	if attrs.Wisdom < 1 {
		attrs.Wisdom = 10
	}
	if attrs.Charisma < 1 {
		attrs.Charisma = 10
	}
	return attrs
}

func (s *NPCService) hasSavingThrows(st models.SavingThrows) bool {
	// Check if any saving throw has been set (modifier != 0 or has proficiency)
	return st.Strength.Modifier != 0 || st.Strength.Proficiency ||
		st.Dexterity.Modifier != 0 || st.Dexterity.Proficiency ||
		st.Constitution.Modifier != 0 || st.Constitution.Proficiency ||
		st.Intelligence.Modifier != 0 || st.Intelligence.Proficiency ||
		st.Wisdom.Modifier != 0 || st.Wisdom.Proficiency ||
		st.Charisma.Modifier != 0 || st.Charisma.Proficiency
}

func (s *NPCService) calculateSavingThrows(npc *models.NPC) models.SavingThrows {
	// TODO: Add proficiency bonus when implementing proficient saves
	// profBonus := s.getProficiencyBonusFromCR(npc.ChallengeRating)

	return models.SavingThrows{
		Strength: models.SavingThrow{
			Modifier:    s.getAbilityModifier(npc.Attributes.Strength),
			Proficiency: false,
		},
		Dexterity: models.SavingThrow{
			Modifier:    s.getAbilityModifier(npc.Attributes.Dexterity),
			Proficiency: false,
		},
		Constitution: models.SavingThrow{
			Modifier:    s.getAbilityModifier(npc.Attributes.Constitution),
			Proficiency: false,
		},
		Intelligence: models.SavingThrow{
			Modifier:    s.getAbilityModifier(npc.Attributes.Intelligence),
			Proficiency: false,
		},
		Wisdom: models.SavingThrow{
			Modifier:    s.getAbilityModifier(npc.Attributes.Wisdom),
			Proficiency: false,
		},
		Charisma: models.SavingThrow{
			Modifier:    s.getAbilityModifier(npc.Attributes.Charisma),
			Proficiency: false,
		},
	}
}

func (s *NPCService) getAbilityModifier(score int) int {
	return (score - 10) / 2
}

func (s *NPCService) getProficiencyBonusFromCR(cr float64) int {
	// CR to proficiency bonus mapping
	if cr <= 4 {
		return 2
	} else if cr <= 8 {
		return 3
	} else if cr <= 12 {
		return 4
	} else if cr <= 16 {
		return 5
	} else if cr <= 20 {
		return 6
	} else if cr <= 24 {
		return 7
	} else if cr <= 28 {
		return 8
	}
	return 9
}

func (s *NPCService) calculateXPFromCR(cr float64) int {
	// D&D 5e XP by Challenge Rating
	xpByCR := map[float64]int{
		0:     10,
		0.125: 25,
		0.25:  50,
		0.5:   100,
		1:     200,
		2:     450,
		3:     700,
		4:     1100,
		5:     1800,
		6:     2300,
		7:     2900,
		8:     3900,
		9:     5000,
		10:    5900,
		11:    7200,
		12:    8400,
		13:    10000,
		14:    11500,
		15:    13000,
		16:    15000,
		17:    18000,
		18:    20000,
		19:    22000,
		20:    25000,
	}

	if xp, ok := xpByCR[cr]; ok {
		return xp
	}
	return 0
}
