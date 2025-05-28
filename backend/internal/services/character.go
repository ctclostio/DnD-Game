package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/models"
)

type CharacterService struct {
	repo database.CharacterRepository
}

func NewCharacterService(repo database.CharacterRepository) *CharacterService {
	return &CharacterService{
		repo: repo,
	}
}

func (s *CharacterService) GetAllCharacters(ctx context.Context, userID string) ([]*models.Character, error) {
	// If userID is provided, get characters for that user
	if userID != "" {
		return s.repo.GetByUserID(ctx, userID)
	}
	// Otherwise, return an empty list (we don't allow listing all characters)
	return []*models.Character{}, nil
}

func (s *CharacterService) GetCharacterByID(ctx context.Context, id string) (*models.Character, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CharacterService) CreateCharacter(ctx context.Context, char *models.Character) error {
	// Validate character
	if char.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if char.Name == "" {
		return fmt.Errorf("character name is required")
	}
	if char.Race == "" {
		return fmt.Errorf("character race is required")
	}
	if char.Class == "" {
		return fmt.Errorf("character class is required")
	}
	
	// Set default values
	if char.Level == 0 {
		char.Level = 1
	}
	if char.MaxHitPoints == 0 {
		char.MaxHitPoints = 10 + getModifier(char.Attributes.Constitution)
	}
	char.HitPoints = char.MaxHitPoints
	
	// Set default armor class if not provided
	if char.ArmorClass == 0 {
		char.ArmorClass = 10 + getModifier(char.Attributes.Dexterity)
	}
	
	// Set default speed if not provided
	if char.Speed == 0 {
		char.Speed = 30 // Default speed in feet
	}
	
	// Set carry capacity based on strength
	char.CarryCapacity = CalculateCarryCapacity(char.Attributes.Strength)
	
	// Set default attunement slots
	if char.AttunementSlotsMax == 0 {
		char.AttunementSlotsMax = 3
	}
	
	return s.repo.Create(ctx, char)
}

func (s *CharacterService) UpdateCharacter(ctx context.Context, char *models.Character) error {
	// Validate character ID
	if char.ID == "" {
		return fmt.Errorf("character ID is required")
	}
	
	// Check if character exists
	existing, err := s.repo.GetByID(ctx, char.ID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}
	
	// Preserve the user ID and created at timestamp
	char.UserID = existing.UserID
	char.CreatedAt = existing.CreatedAt
	
	return s.repo.Update(ctx, char)
}

func (s *CharacterService) DeleteCharacter(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Helper function to calculate ability modifier
func getModifier(ability int) int {
	return (ability - 10) / 2
}

// CalculateCarryCapacity calculates carry capacity based on strength
func CalculateCarryCapacity(strength int) float64 {
	return float64(strength * 15)
}

// CalculateHitPoints calculates hit points based on class and constitution
func (s *CharacterService) CalculateHitPoints(class string, level int, constitution int) int {
	// Base hit points by class (simplified)
	baseHP := map[string]int{
		"fighter":  10,
		"wizard":   6,
		"rogue":    8,
		"cleric":   8,
		"ranger":   10,
		"paladin":  10,
		"barbarian": 12,
		"bard":     8,
		"druid":    8,
		"monk":     8,
		"sorcerer": 6,
		"warlock":  8,
	}
	
	base, ok := baseHP[class]
	if !ok {
		base = 8 // Default
	}
	
	conMod := getModifier(constitution)
	// First level gets full hit die + con mod
	// Additional levels get average hit die + con mod
	hitPoints := base + conMod
	if level > 1 {
		hitPoints += (level - 1) * ((base/2 + 1) + conMod)
	}
	
	return hitPoints
}

// InitializeSpellSlots sets up spell slots based on class and level
func (s *CharacterService) InitializeSpellSlots(class string, level int) []models.SpellSlot {
	// Spell slots by level for full casters (wizard, cleric, druid, bard, sorcerer)
	fullCasterSlots := map[int][]int{
		1:  {2, 0, 0, 0, 0, 0, 0, 0, 0},
		2:  {3, 0, 0, 0, 0, 0, 0, 0, 0},
		3:  {4, 2, 0, 0, 0, 0, 0, 0, 0},
		4:  {4, 3, 0, 0, 0, 0, 0, 0, 0},
		5:  {4, 3, 2, 0, 0, 0, 0, 0, 0},
		6:  {4, 3, 3, 0, 0, 0, 0, 0, 0},
		7:  {4, 3, 3, 1, 0, 0, 0, 0, 0},
		8:  {4, 3, 3, 2, 0, 0, 0, 0, 0},
		9:  {4, 3, 3, 3, 1, 0, 0, 0, 0},
		10: {4, 3, 3, 3, 2, 0, 0, 0, 0},
		11: {4, 3, 3, 3, 2, 1, 0, 0, 0},
		12: {4, 3, 3, 3, 2, 1, 0, 0, 0},
		13: {4, 3, 3, 3, 2, 1, 1, 0, 0},
		14: {4, 3, 3, 3, 2, 1, 1, 0, 0},
		15: {4, 3, 3, 3, 2, 1, 1, 1, 0},
		16: {4, 3, 3, 3, 2, 1, 1, 1, 0},
		17: {4, 3, 3, 3, 2, 1, 1, 1, 1},
		18: {4, 3, 3, 3, 3, 1, 1, 1, 1},
		19: {4, 3, 3, 3, 3, 2, 1, 1, 1},
		20: {4, 3, 3, 3, 3, 2, 2, 1, 1},
	}

	// Half casters (ranger, paladin) get spells at level 2
	halfCasterSlots := map[int][]int{
		1:  {0, 0, 0, 0, 0, 0, 0, 0, 0},
		2:  {2, 0, 0, 0, 0, 0, 0, 0, 0},
		3:  {3, 0, 0, 0, 0, 0, 0, 0, 0},
		4:  {3, 0, 0, 0, 0, 0, 0, 0, 0},
		5:  {4, 2, 0, 0, 0, 0, 0, 0, 0},
		6:  {4, 2, 0, 0, 0, 0, 0, 0, 0},
		7:  {4, 3, 0, 0, 0, 0, 0, 0, 0},
		8:  {4, 3, 0, 0, 0, 0, 0, 0, 0},
		9:  {4, 3, 2, 0, 0, 0, 0, 0, 0},
		10: {4, 3, 2, 0, 0, 0, 0, 0, 0},
		11: {4, 3, 3, 0, 0, 0, 0, 0, 0},
		12: {4, 3, 3, 0, 0, 0, 0, 0, 0},
		13: {4, 3, 3, 1, 0, 0, 0, 0, 0},
		14: {4, 3, 3, 1, 0, 0, 0, 0, 0},
		15: {4, 3, 3, 2, 0, 0, 0, 0, 0},
		16: {4, 3, 3, 2, 0, 0, 0, 0, 0},
		17: {4, 3, 3, 3, 1, 0, 0, 0, 0},
		18: {4, 3, 3, 3, 1, 0, 0, 0, 0},
		19: {4, 3, 3, 3, 2, 0, 0, 0, 0},
		20: {4, 3, 3, 3, 2, 0, 0, 0, 0},
	}

	// Warlock has unique spell slot progression (Pact Magic)
	warlockSlots := map[int][]int{
		1:  {1, 0, 0, 0, 0, 0, 0, 0, 0},
		2:  {2, 0, 0, 0, 0, 0, 0, 0, 0},
		3:  {0, 2, 0, 0, 0, 0, 0, 0, 0},
		4:  {0, 2, 0, 0, 0, 0, 0, 0, 0},
		5:  {0, 0, 2, 0, 0, 0, 0, 0, 0},
		6:  {0, 0, 2, 0, 0, 0, 0, 0, 0},
		7:  {0, 0, 0, 2, 0, 0, 0, 0, 0},
		8:  {0, 0, 0, 2, 0, 0, 0, 0, 0},
		9:  {0, 0, 0, 0, 2, 0, 0, 0, 0},
		10: {0, 0, 0, 0, 2, 0, 0, 0, 0},
		11: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		12: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		13: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		14: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		15: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		16: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		17: {0, 0, 0, 0, 4, 0, 0, 0, 0},
		18: {0, 0, 0, 0, 4, 0, 0, 0, 0},
		19: {0, 0, 0, 0, 4, 0, 0, 0, 0},
		20: {0, 0, 0, 0, 4, 0, 0, 0, 0},
	}

	var slotsTable map[int][]int
	switch class {
	case "wizard", "cleric", "druid", "bard", "sorcerer":
		slotsTable = fullCasterSlots
	case "ranger", "paladin":
		slotsTable = halfCasterSlots
	case "warlock":
		slotsTable = warlockSlots
	default:
		// Non-casters have no spell slots
		return []models.SpellSlot{}
	}

	slots := []models.SpellSlot{}
	if slotCounts, ok := slotsTable[level]; ok {
		for i, count := range slotCounts {
			if count > 0 {
				slots = append(slots, models.SpellSlot{
					Level:     i + 1,
					Total:     count,
					Remaining: count,
				})
			}
		}
	}

	return slots
}

// UseSpellSlot consumes a spell slot of the specified level
func (s *CharacterService) UseSpellSlot(ctx context.Context, characterID string, slotLevel int) error {
	char, err := s.GetCharacterByID(ctx, characterID)
	if err != nil {
		return err
	}

	for i := range char.Spells.SpellSlots {
		if char.Spells.SpellSlots[i].Level == slotLevel {
			if char.Spells.SpellSlots[i].Remaining > 0 {
				char.Spells.SpellSlots[i].Remaining--
				return s.UpdateCharacter(ctx, char)
			}
			return fmt.Errorf("no remaining spell slots of level %d", slotLevel)
		}
	}

	return fmt.Errorf("character does not have spell slots of level %d", slotLevel)
}

// RestoreSpellSlots restores spell slots (short or long rest)
func (s *CharacterService) RestoreSpellSlots(ctx context.Context, characterID string, restType string) error {
	char, err := s.GetCharacterByID(ctx, characterID)
	if err != nil {
		return err
	}

	switch restType {
	case "short":
		// Warlocks recover all spell slots on short rest
		if char.Class == "warlock" {
			for i := range char.Spells.SpellSlots {
				char.Spells.SpellSlots[i].Remaining = char.Spells.SpellSlots[i].Total
			}
		}
		// Other classes might have features that restore slots on short rest
		// This could be expanded based on specific class features
	case "long":
		// All classes recover all spell slots on long rest
		for i := range char.Spells.SpellSlots {
			char.Spells.SpellSlots[i].Remaining = char.Spells.SpellSlots[i].Total
		}
	default:
		return fmt.Errorf("invalid rest type: %s", restType)
	}

	return s.UpdateCharacter(ctx, char)
}

// Experience and Level Management

// AddExperience adds XP to a character and handles level up if needed
func (s *CharacterService) AddExperience(ctx context.Context, characterID string, xp int) error {
	char, err := s.GetCharacterByID(ctx, characterID)
	if err != nil {
		return err
	}

	char.ExperiencePoints += xp
	
	// Check for level up
	newLevel := s.calculateLevelFromXP(char.ExperiencePoints)
	if newLevel > char.Level {
		return s.levelUp(ctx, char, newLevel)
	}

	return s.UpdateCharacter(ctx, char)
}

// calculateLevelFromXP determines character level based on XP
func (s *CharacterService) calculateLevelFromXP(xp int) int {
	xpThresholds := []int{
		0,      // Level 1
		300,    // Level 2
		900,    // Level 3
		2700,   // Level 4
		6500,   // Level 5
		14000,  // Level 6
		23000,  // Level 7
		34000,  // Level 8
		48000,  // Level 9
		64000,  // Level 10
		85000,  // Level 11
		100000, // Level 12
		120000, // Level 13
		140000, // Level 14
		165000, // Level 15
		195000, // Level 16
		225000, // Level 17
		265000, // Level 18
		305000, // Level 19
		355000, // Level 20
	}

	level := 1
	for i, threshold := range xpThresholds {
		if xp >= threshold {
			level = i + 1
		} else {
			break
		}
	}

	if level > 20 {
		level = 20
	}

	return level
}

// GetXPForNextLevel returns the XP needed for the next level
func (s *CharacterService) GetXPForNextLevel(level int) int {
	xpThresholds := map[int]int{
		1:  300,
		2:  900,
		3:  2700,
		4:  6500,
		5:  14000,
		6:  23000,
		7:  34000,
		8:  48000,
		9:  64000,
		10: 85000,
		11: 100000,
		12: 120000,
		13: 140000,
		14: 165000,
		15: 195000,
		16: 225000,
		17: 265000,
		18: 305000,
		19: 355000,
		20: 999999, // Max level
	}

	if xp, ok := xpThresholds[level]; ok {
		return xp
	}
	return 999999
}

// levelUp handles character level progression
func (s *CharacterService) levelUp(ctx context.Context, char *models.Character, newLevel int) error {
	oldLevel := char.Level
	char.Level = newLevel

	// Update proficiency bonus
	char.ProficiencyBonus = ((char.Level - 1) / 4) + 2

	// Calculate HP increase for each level gained
	for level := oldLevel + 1; level <= newLevel; level++ {
		hpIncrease := s.calculateHPIncrease(char.Class, char.Attributes.Constitution)
		char.MaxHitPoints += hpIncrease
		char.HitPoints += hpIncrease // Also increase current HP
	}

	// Update spell slots for spellcasters
	if char.Spells.SpellcastingAbility != "" {
		char.Spells.SpellSlots = s.InitializeSpellSlots(char.Class, char.Level)
		
		// Update spell save DC and attack bonus
		abilityMod := s.getSpellcastingAbilityModifier(char)
		char.Spells.SpellSaveDC = 8 + char.ProficiencyBonus + abilityMod
		char.Spells.SpellAttackBonus = char.ProficiencyBonus + abilityMod
	}

	// Update saving throws
	s.updateSavingThrows(char)

	// TODO: Add class features based on new level
	// TODO: Update skill modifiers

	return s.UpdateCharacter(ctx, char)
}

// calculateHPIncrease calculates HP gained on level up
func (s *CharacterService) calculateHPIncrease(class string, constitution int) int {
	// Average hit die value by class
	hitDieAverage := map[string]int{
		"fighter":   6, // 1d10 average
		"wizard":    4, // 1d6 average
		"rogue":     5, // 1d8 average
		"cleric":    5, // 1d8 average
		"ranger":    6, // 1d10 average
		"paladin":   6, // 1d10 average
		"barbarian": 7, // 1d12 average
		"bard":      5, // 1d8 average
		"druid":     5, // 1d8 average
		"monk":      5, // 1d8 average
		"sorcerer":  4, // 1d6 average
		"warlock":   5, // 1d8 average
	}

	average, ok := hitDieAverage[class]
	if !ok {
		average = 5 // Default to d8
	}

	conMod := getModifier(constitution)
	return average + conMod
}

// getSpellcastingAbilityModifier returns the modifier for the character's spellcasting ability
func (s *CharacterService) getSpellcastingAbilityModifier(char *models.Character) int {
	switch strings.ToLower(char.Spells.SpellcastingAbility) {
	case "intelligence":
		return getModifier(char.Attributes.Intelligence)
	case "wisdom":
		return getModifier(char.Attributes.Wisdom)
	case "charisma":
		return getModifier(char.Attributes.Charisma)
	default:
		return 0
	}
}

// updateSavingThrows recalculates saving throws with new proficiency bonus
func (s *CharacterService) updateSavingThrows(char *models.Character) {
	// Recalculate all saving throws
	char.SavingThrows.Strength.Modifier = getModifier(char.Attributes.Strength)
	if char.SavingThrows.Strength.Proficiency {
		char.SavingThrows.Strength.Modifier += char.ProficiencyBonus
	}

	char.SavingThrows.Dexterity.Modifier = getModifier(char.Attributes.Dexterity)
	if char.SavingThrows.Dexterity.Proficiency {
		char.SavingThrows.Dexterity.Modifier += char.ProficiencyBonus
	}

	char.SavingThrows.Constitution.Modifier = getModifier(char.Attributes.Constitution)
	if char.SavingThrows.Constitution.Proficiency {
		char.SavingThrows.Constitution.Modifier += char.ProficiencyBonus
	}

	char.SavingThrows.Intelligence.Modifier = getModifier(char.Attributes.Intelligence)
	if char.SavingThrows.Intelligence.Proficiency {
		char.SavingThrows.Intelligence.Modifier += char.ProficiencyBonus
	}

	char.SavingThrows.Wisdom.Modifier = getModifier(char.Attributes.Wisdom)
	if char.SavingThrows.Wisdom.Proficiency {
		char.SavingThrows.Wisdom.Modifier += char.ProficiencyBonus
	}

	char.SavingThrows.Charisma.Modifier = getModifier(char.Attributes.Charisma)
	if char.SavingThrows.Charisma.Proficiency {
		char.SavingThrows.Charisma.Modifier += char.ProficiencyBonus
	}
}

// GetCharacter gets a character by ID (alias for GetCharacterByID)
func (s *CharacterService) GetCharacter(ctx context.Context, id string) (*models.Character, error) {
	return s.GetCharacterByID(ctx, id)
}