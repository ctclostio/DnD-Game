package testutil

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// AssertHelpers provides custom assertion methods for D&D game entities
type AssertHelpers struct {
	t *testing.T
}

// NewAssertHelpers creates a new assertion helper
func NewAssertHelpers(t *testing.T) *AssertHelpers {
	return &AssertHelpers{t: t}
}

// AssertCharacterValid asserts a character has valid D&D properties
func (a *AssertHelpers) AssertCharacterValid(char *models.Character) {
	require.NotNil(a.t, char)
	require.NotEmpty(a.t, char.Name)
	require.NotEmpty(a.t, char.Race)
	require.NotEmpty(a.t, char.Class)
	require.Greater(a.t, char.Level, 0)
	require.LessOrEqual(a.t, char.Level, 20)
	require.GreaterOrEqual(a.t, char.HitPoints, 0)
	require.LessOrEqual(a.t, char.HitPoints, char.MaxHitPoints)
	require.Greater(a.t, char.MaxHitPoints, 0)
	require.GreaterOrEqual(a.t, char.ArmorClass, 0)

	// Validate ability scores
	a.AssertAbilityScoresValid(char.Attributes)
}

// AssertAbilityScoresValid asserts ability scores are within D&D bounds
func (a *AssertHelpers) AssertAbilityScoresValid(attributes models.Attributes) {
	scores := []int{
		attributes.Strength,
		attributes.Dexterity,
		attributes.Constitution,
		attributes.Intelligence,
		attributes.Wisdom,
		attributes.Charisma,
	}

	for _, score := range scores {
		require.GreaterOrEqual(a.t, score, 1, "Ability score must be at least 1")
		require.LessOrEqual(a.t, score, 30, "Ability score cannot exceed 30")
	}
}

// AssertDiceRollValid asserts a dice roll result is valid
func (a *AssertHelpers) AssertDiceRollValid(roll *models.DiceRoll) {
	require.NotNil(a.t, roll)
	require.NotEmpty(a.t, roll.RollNotation)
	require.NotEmpty(a.t, roll.Purpose)
	require.Greater(a.t, len(roll.Results), 0)

	// Validate each individual roll
	for _, r := range roll.Results {
		require.Greater(a.t, r, 0, "Dice roll must be positive")
	}

	// Validate result matches rolls + modifiers
	sum := 0
	for _, r := range roll.Results {
		sum += r
	}
	require.Equal(a.t, sum+roll.Modifier, roll.Total)
}

// AssertCombatValid asserts combat state is valid
func (a *AssertHelpers) AssertCombatValid(combat *models.Combat) {
	require.NotNil(a.t, combat)
	require.Greater(a.t, combat.Round, 0)
	require.GreaterOrEqual(a.t, combat.CurrentTurn, 0)
	require.Less(a.t, combat.CurrentTurn, len(combat.TurnOrder))

	// Validate turn order
	for i, combatantID := range combat.TurnOrder {
		require.NotEmpty(a.t, combatantID, "Turn order entry %d must have combatant ID", i)
	}

	// Validate combatants
	for i := range combat.Combatants {
		combatant := &combat.Combatants[i]
		require.NotEmpty(a.t, combatant.ID, "Combatant %d must have ID", i)
		require.NotEmpty(a.t, combatant.Name, "Combatant %d must have name", i)
		require.GreaterOrEqual(a.t, combatant.HP, 0)
		require.LessOrEqual(a.t, combatant.HP, combatant.MaxHP)
	}

	// Ensure turn order contains valid combatant IDs
	combatantIDMap := make(map[string]bool)
	for i := range combat.Combatants {
		combatantIDMap[combat.Combatants[i].ID] = true
	}

	for _, id := range combat.TurnOrder {
		require.True(a.t, combatantIDMap[id], "Turn order contains invalid combatant ID: %s", id)
	}
}

// AssertInventoryItemValid asserts an inventory item is valid
func (a *AssertHelpers) AssertInventoryItemValid(item *models.InventoryItem) {
	require.NotNil(a.t, item)
	require.NotNil(a.t, item.Item, "InventoryItem must have an Item")
	require.NotEmpty(a.t, item.Item.Name)
	require.NotEmpty(a.t, item.Item.Type)
	require.Greater(a.t, item.Quantity, 0)
	require.GreaterOrEqual(a.t, item.Item.Weight, 0.0)
	require.GreaterOrEqual(a.t, item.Item.Value, 0)

	// Validate type-specific properties
	switch item.Item.Type {
	case models.ItemTypeWeapon:
		a.AssertWeaponPropertiesValid(item.Item.Properties)
	case models.ItemTypeArmor:
		a.AssertArmorPropertiesValid(item.Item.Properties)
	case models.ItemTypeMagic:
		a.AssertMagicItemPropertiesValid(item.Item.Properties)
	case models.ItemTypeConsumable, models.ItemTypeTool, models.ItemTypeOther:
		// These item types may have various properties or none at all
		// No specific validation required
	}
}

// AssertWeaponPropertiesValid validates weapon-specific properties
func (a *AssertHelpers) AssertWeaponPropertiesValid(props map[string]interface{}) {
	damage, ok := props["damage"].(string)
	require.True(a.t, ok, "Weapon must have damage property")
	require.NotEmpty(a.t, damage)

	damageType, ok := props["damageType"].(string)
	require.True(a.t, ok, "Weapon must have damage type")
	require.Contains(a.t, []string{
		"slashing", "piercing", "bludgeoning",
		"fire", "cold", "lightning", "acid",
		"poison", "psychic", "necrotic", "radiant",
	}, damageType)
}

// AssertArmorPropertiesValid validates armor-specific properties
func (a *AssertHelpers) AssertArmorPropertiesValid(props map[string]interface{}) {
	ac, ok := props["armorClass"].(int)
	if !ok {
		// Try float64 (JSON unmarshaling)
		acFloat, ok := props["armorClass"].(float64)
		require.True(a.t, ok, "Armor must have armorClass property")
		ac = int(acFloat)
	}
	require.Greater(a.t, ac, 0)
}

// AssertMagicItemPropertiesValid validates magic item properties
func (a *AssertHelpers) AssertMagicItemPropertiesValid(props map[string]interface{}) {
	rarity, ok := props["rarity"].(string)
	require.True(a.t, ok, "Magic item must have rarity")
	require.Contains(a.t, []string{
		"common", "uncommon", "rare", "very rare", "legendary", "artifact",
	}, rarity)
}

// AssertSpellSlotsValid validates spell slot structure
func (a *AssertHelpers) AssertSpellSlotsValid(slots []models.SpellSlot) {
	for _, slot := range slots {
		require.GreaterOrEqual(a.t, slot.Remaining, 0)
		require.LessOrEqual(a.t, slot.Remaining, slot.Total)
		require.GreaterOrEqual(a.t, slot.Total, 0)

		// Validate spell level
		require.GreaterOrEqual(a.t, slot.Level, 1)
		require.LessOrEqual(a.t, slot.Level, 9)
	}
}

// AssertErrorCode asserts an error has the expected code
func (a *AssertHelpers) AssertErrorCode(err error, _ string) {
	require.Error(a.t, err)
	// This would check against your custom error type
	// For now, we'll just check the error exists
}

// AssertWithinDuration asserts two times are within a duration of each other
func (a *AssertHelpers) AssertWithinDuration(expected, actual time.Time, delta time.Duration) {
	require.WithinDuration(a.t, expected, actual, delta)
}

// AssertJSONEqual asserts two values are equal when marshaled to JSON
func (a *AssertHelpers) AssertJSONEqual(expected, actual interface{}) {
	expectedJSON, err := json.Marshal(expected)
	require.NoError(a.t, err)

	actualJSON, err := json.Marshal(actual)
	require.NoError(a.t, err)

	require.JSONEq(a.t, string(expectedJSON), string(actualJSON))
}

// DiceRollAssertions provides dice-specific assertions
type DiceRollAssertions struct {
	t *testing.T
}

// NewDiceRollAssertions creates dice roll assertions
func NewDiceRollAssertions(t *testing.T) *DiceRollAssertions {
	return &DiceRollAssertions{t: t}
}

// AssertD20Roll asserts a d20 roll is valid
func (d *DiceRollAssertions) AssertD20Roll(roll int) {
	require.GreaterOrEqual(d.t, roll, 1)
	require.LessOrEqual(d.t, roll, 20)
}

// AssertAdvantageRoll asserts an advantage roll (2d20 keep highest)
func (d *DiceRollAssertions) AssertAdvantageRoll(rolls []int, result int) {
	require.Len(d.t, rolls, 2)
	d.AssertD20Roll(rolls[0])
	d.AssertD20Roll(rolls[1])

	expected := rolls[0]
	if rolls[1] > expected {
		expected = rolls[1]
	}
	require.Equal(d.t, expected, result)
}

// AssertDisadvantageRoll asserts a disadvantage roll (2d20 keep lowest)
func (d *DiceRollAssertions) AssertDisadvantageRoll(rolls []int, result int) {
	require.Len(d.t, rolls, 2)
	d.AssertD20Roll(rolls[0])
	d.AssertD20Roll(rolls[1])

	expected := rolls[0]
	if rolls[1] < expected {
		expected = rolls[1]
	}
	require.Equal(d.t, expected, result)
}

// CombatAssertions provides combat-specific assertions
type CombatAssertions struct {
	t *testing.T
}

// NewCombatAssertions creates combat assertions
func NewCombatAssertions(t *testing.T) *CombatAssertions {
	return &CombatAssertions{t: t}
}

// AssertDamageApplied asserts damage was correctly applied
func (c *CombatAssertions) AssertDamageApplied(
	originalHP, damage, currentHP int,
	hasResistance, hasVulnerability bool,
) {
	expectedDamage := damage
	if hasResistance {
		expectedDamage = damage / 2
	} else if hasVulnerability {
		expectedDamage = damage * 2
	}

	expectedHP := originalHP - expectedDamage
	if expectedHP < 0 {
		expectedHP = 0
	}

	require.Equal(c.t, expectedHP, currentHP)
}

// AssertConditionApplied asserts a condition was applied
func (c *CombatAssertions) AssertConditionApplied(
	combatant *models.Combatant,
	condition string,
) {
	require.NotNil(c.t, combatant)
	require.Contains(c.t, combatant.Condition, condition)
}

// AssertInitiativeOrder asserts combat order is correct
func (c *CombatAssertions) AssertInitiativeOrder(combatants []models.Combatant) {
	require.Greater(c.t, len(combatants), 0)

	for i := 1; i < len(combatants); i++ {
		assert.GreaterOrEqual(c.t,
			combatants[i-1].Initiative,
			combatants[i].Initiative,
			"Initiative order should be descending")
	}
}

// Quick assertion functions for common checks

// RequireValidCharacter requires a character to be valid
func RequireValidCharacter(t *testing.T, char *models.Character) {
	NewAssertHelpers(t).AssertCharacterValid(char)
}

// RequireValidDiceRoll requires a dice roll to be valid
func RequireValidDiceRoll(t *testing.T, roll *models.DiceRoll) {
	NewAssertHelpers(t).AssertDiceRollValid(roll)
}

// RequireValidCombat requires combat state to be valid
func RequireValidCombat(t *testing.T, combat *models.Combat) {
	NewAssertHelpers(t).AssertCombatValid(combat)
}

// RequireValidInventoryItem requires an inventory item to be valid
func RequireValidInventoryItem(t *testing.T, item *models.InventoryItem) {
	NewAssertHelpers(t).AssertInventoryItemValid(item)
}
