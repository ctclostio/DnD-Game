package services

import (
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// SpellSlotCalculator interface for different spell slot progression systems
type SpellSlotCalculator interface {
	GetSpellSlots(level int) []models.SpellSlot
}

// SpellSlotRegistry manages different spell slot calculators
type SpellSlotRegistry struct {
	calculators map[string]SpellSlotCalculator
}

// NewSpellSlotRegistry creates a new registry with all spell slot calculators
func NewSpellSlotRegistry() *SpellSlotRegistry {
	return &SpellSlotRegistry{
		calculators: map[string]SpellSlotCalculator{
			// Full casters
			constants.ClassWizard:   NewFullCasterCalculator(),
			constants.ClassCleric:   NewFullCasterCalculator(),
			constants.ClassDruid:    NewFullCasterCalculator(),
			constants.ClassBard:     NewFullCasterCalculator(),
			constants.ClassSorcerer: NewFullCasterCalculator(),
			
			// Half casters
			constants.ClassRanger:  NewHalfCasterCalculator(),
			constants.ClassPaladin: NewHalfCasterCalculator(),
			
			// Warlock (Pact Magic)
			constants.ClassWarlock: NewWarlockCalculator(),
			
			// Non-casters
			constants.ClassFighter:  NewNonCasterCalculator(),
			constants.ClassRogue:    NewNonCasterCalculator(),
			constants.ClassBarbarian: NewNonCasterCalculator(),
			constants.ClassMonk:     NewNonCasterCalculator(),
		},
	}
}

// GetCalculator returns the appropriate calculator for a class
func (r *SpellSlotRegistry) GetCalculator(class string) SpellSlotCalculator {
	if calc, exists := r.calculators[class]; exists {
		return calc
	}
	return NewNonCasterCalculator()
}

// BaseSpellSlotCalculator provides common functionality
type BaseSpellSlotCalculator struct {
	progression SpellSlotProgression
}

// GetSpellSlots returns spell slots for a given level
func (c *BaseSpellSlotCalculator) GetSpellSlots(level int) []models.SpellSlot {
	if level < 1 || level > 20 {
		return []models.SpellSlot{}
	}
	
	slotCounts := c.progression.GetSlots(level)
	return c.convertToSpellSlots(slotCounts)
}

// convertToSpellSlots converts slot counts to SpellSlot models
func (c *BaseSpellSlotCalculator) convertToSpellSlots(slotCounts []int) []models.SpellSlot {
	slots := []models.SpellSlot{}
	
	for level, count := range slotCounts {
		if count > 0 {
			slots = append(slots, models.SpellSlot{
				Level:     level + 1, // 0-indexed to 1-indexed
				Total:     count,
				Remaining: count,
			})
		}
	}
	
	return slots
}

// SpellSlotProgression defines the interface for spell slot progression data
type SpellSlotProgression interface {
	GetSlots(level int) []int
}

// FullCasterProgression implements spell slot progression for full casters
type FullCasterProgression struct{}

func (p *FullCasterProgression) GetSlots(level int) []int {
	// Using a more maintainable structure
	// Format: [1st, 2nd, 3rd, 4th, 5th, 6th, 7th, 8th, 9th]
	progressions := [][]int{
		{2, 0, 0, 0, 0, 0, 0, 0, 0}, // Level 1
		{3, 0, 0, 0, 0, 0, 0, 0, 0}, // Level 2
		{4, 2, 0, 0, 0, 0, 0, 0, 0}, // Level 3
		{4, 3, 0, 0, 0, 0, 0, 0, 0}, // Level 4
		{4, 3, 2, 0, 0, 0, 0, 0, 0}, // Level 5
		{4, 3, 3, 0, 0, 0, 0, 0, 0}, // Level 6
		{4, 3, 3, 1, 0, 0, 0, 0, 0}, // Level 7
		{4, 3, 3, 2, 0, 0, 0, 0, 0}, // Level 8
		{4, 3, 3, 3, 1, 0, 0, 0, 0}, // Level 9
		{4, 3, 3, 3, 2, 0, 0, 0, 0}, // Level 10
		{4, 3, 3, 3, 2, 1, 0, 0, 0}, // Level 11
		{4, 3, 3, 3, 2, 1, 0, 0, 0}, // Level 12
		{4, 3, 3, 3, 2, 1, 1, 0, 0}, // Level 13
		{4, 3, 3, 3, 2, 1, 1, 0, 0}, // Level 14
		{4, 3, 3, 3, 2, 1, 1, 1, 0}, // Level 15
		{4, 3, 3, 3, 2, 1, 1, 1, 0}, // Level 16
		{4, 3, 3, 3, 2, 1, 1, 1, 1}, // Level 17
		{4, 3, 3, 3, 3, 1, 1, 1, 1}, // Level 18
		{4, 3, 3, 3, 3, 2, 1, 1, 1}, // Level 19
		{4, 3, 3, 3, 3, 2, 2, 1, 1}, // Level 20
	}
	
	if level >= 1 && level <= 20 {
		return progressions[level-1]
	}
	return []int{0, 0, 0, 0, 0, 0, 0, 0, 0}
}

// HalfCasterProgression implements spell slot progression for half casters
type HalfCasterProgression struct{}

func (p *HalfCasterProgression) GetSlots(level int) []int {
	// Half casters get roughly half the spell slots, starting at level 2
	// This could be loaded from a configuration file
	progressions := map[int][]int{
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
	
	if slots, exists := progressions[level]; exists {
		return slots
	}
	return []int{0, 0, 0, 0, 0, 0, 0, 0, 0}
}

// WarlockProgression implements Pact Magic progression
type WarlockProgression struct{}

func (p *WarlockProgression) GetSlots(level int) []int {
	// Warlocks have a unique progression - fewer slots but they refresh on short rest
	// All slots are of the highest level available
	switch level {
	case 1, 2:
		return []int{level, 0, 0, 0, 0, 0, 0, 0, 0} // 1st level slots
	case 3, 4:
		return []int{0, 2, 0, 0, 0, 0, 0, 0, 0} // 2nd level slots
	case 5, 6:
		return []int{0, 0, 2, 0, 0, 0, 0, 0, 0} // 3rd level slots
	case 7, 8:
		return []int{0, 0, 0, 2, 0, 0, 0, 0, 0} // 4th level slots
	case 9, 10:
		return []int{0, 0, 0, 0, 2, 0, 0, 0, 0} // 5th level slots
	case 11, 12, 13, 14, 15, 16:
		return []int{0, 0, 0, 0, 3, 0, 0, 0, 0} // 5th level slots (3)
	case 17, 18, 19, 20:
		return []int{0, 0, 0, 0, 4, 0, 0, 0, 0} // 5th level slots (4)
	default:
		return []int{0, 0, 0, 0, 0, 0, 0, 0, 0}
	}
}

// Calculator implementations
type FullCasterCalculator struct {
	BaseSpellSlotCalculator
}

func NewFullCasterCalculator() *FullCasterCalculator {
	return &FullCasterCalculator{
		BaseSpellSlotCalculator{
			progression: &FullCasterProgression{},
		},
	}
}

type HalfCasterCalculator struct {
	BaseSpellSlotCalculator
}

func NewHalfCasterCalculator() *HalfCasterCalculator {
	return &HalfCasterCalculator{
		BaseSpellSlotCalculator{
			progression: &HalfCasterProgression{},
		},
	}
}

type WarlockCalculator struct {
	BaseSpellSlotCalculator
}

func NewWarlockCalculator() *WarlockCalculator {
	return &WarlockCalculator{
		BaseSpellSlotCalculator{
			progression: &WarlockProgression{},
		},
	}
}

type NonCasterCalculator struct{}

func NewNonCasterCalculator() *NonCasterCalculator {
	return &NonCasterCalculator{}
}

func (c *NonCasterCalculator) GetSpellSlots(level int) []models.SpellSlot {
	return []models.SpellSlot{}
}

// RefactoredInitializeSpellSlots is the new, cleaner version
func RefactoredInitializeSpellSlots(registry *SpellSlotRegistry, class string, level int) []models.SpellSlot {
	calculator := registry.GetCalculator(class)
	return calculator.GetSpellSlots(level)
}

// Usage example for CharacterService
type CharacterServiceV2 struct {
	spellSlotRegistry *SpellSlotRegistry
	// other dependencies...
}

func NewCharacterServiceV2() *CharacterServiceV2 {
	return &CharacterServiceV2{
		spellSlotRegistry: NewSpellSlotRegistry(),
	}
}

// InitializeSpellSlots - refactored version
func (s *CharacterServiceV2) InitializeSpellSlots(class string, level int) []models.SpellSlot {
	return RefactoredInitializeSpellSlots(s.spellSlotRegistry, class, level)
}

// Additional benefit: Easy to add new classes or modify progressions
func (r *SpellSlotRegistry) RegisterClass(class string, calculator SpellSlotCalculator) {
	r.calculators[class] = calculator
}

// Example: Adding a homebrew class with custom progression
type CustomSpellProgression struct {
	slots [][]int
}

func (p *CustomSpellProgression) GetSlots(level int) []int {
	if level >= 1 && level <= len(p.slots) {
		return p.slots[level-1]
	}
	return []int{0, 0, 0, 0, 0, 0, 0, 0, 0}
}

// Configuration-based approach (even better for maintainability)
type SpellSlotConfig struct {
	ClassProgressions map[string][][]int `json:"class_progressions"`
}

// LoadFromConfig loads spell slot progressions from a configuration file
func LoadSpellSlotRegistryFromConfig(config *SpellSlotConfig) *SpellSlotRegistry {
	registry := &SpellSlotRegistry{
		calculators: make(map[string]SpellSlotCalculator),
	}
	
	for class, progression := range config.ClassProgressions {
		customProg := &CustomSpellProgression{slots: progression}
		calc := &BaseSpellSlotCalculator{progression: customProg}
		registry.RegisterClass(class, calc)
	}
	
	return registry
}