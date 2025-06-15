package testhelpers

import (
	"encoding/json"
	"testing"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertValidCharacter validates a character has required fields
func AssertValidCharacter(t *testing.T, char *models.Character) {
	t.Helper()

	require.NotNil(t, char, "Character should not be nil")
	assert.NotEmpty(t, char.ID, "Character ID should not be empty")
	assert.NotEmpty(t, char.UserID, "Character UserID should not be empty")
	assert.NotEmpty(t, char.Name, "Character Name should not be empty")
	assert.NotEmpty(t, char.Race, "Character Race should not be empty")
	assert.NotEmpty(t, char.Class, "Character Class should not be empty")
	assert.Greater(t, char.Level, 0, "Character Level should be greater than 0")
	assert.GreaterOrEqual(t, char.HitPoints, 0, "Character HitPoints should not be negative")
	assert.Greater(t, char.MaxHitPoints, 0, "Character MaxHitPoints should be greater than 0")

	// Validate attributes
	AssertValidAttributes(t, &char.Attributes)
}

// AssertValidAttributes validates character attributes
func AssertValidAttributes(t *testing.T, attrs *models.Attributes) {
	t.Helper()

	assert.GreaterOrEqual(t, attrs.Strength, 1, "Strength should be at least 1")
	assert.LessOrEqual(t, attrs.Strength, 30, "Strength should not exceed 30")
	assert.GreaterOrEqual(t, attrs.Dexterity, 1, "Dexterity should be at least 1")
	assert.LessOrEqual(t, attrs.Dexterity, 30, "Dexterity should not exceed 30")
	assert.GreaterOrEqual(t, attrs.Constitution, 1, "Constitution should be at least 1")
	assert.LessOrEqual(t, attrs.Constitution, 30, "Constitution should not exceed 30")
	assert.GreaterOrEqual(t, attrs.Intelligence, 1, "Intelligence should be at least 1")
	assert.LessOrEqual(t, attrs.Intelligence, 30, "Intelligence should not exceed 30")
	assert.GreaterOrEqual(t, attrs.Wisdom, 1, "Wisdom should be at least 1")
	assert.LessOrEqual(t, attrs.Wisdom, 30, "Wisdom should not exceed 30")
	assert.GreaterOrEqual(t, attrs.Charisma, 1, "Charisma should be at least 1")
	assert.LessOrEqual(t, attrs.Charisma, 30, "Charisma should not exceed 30")
}

// AssertValidGameSession validates a game session
func AssertValidGameSession(t *testing.T, session *models.GameSession) {
	t.Helper()

	require.NotNil(t, session, "GameSession should not be nil")
	assert.NotEmpty(t, session.ID, "GameSession ID should not be empty")
	assert.NotEmpty(t, session.Name, "GameSession Name should not be empty")
	assert.NotEmpty(t, session.DMID, "GameSession DMID should not be empty")
	assert.NotEmpty(t, session.Status, "GameSession Status should not be empty")
	assert.Contains(t, []string{"pending", "active", "paused", "completed"}, session.Status,
		"GameSession Status should be valid")
}

// AssertValidCombat validates a combat instance
func AssertValidCombat(t *testing.T, combat *models.Combat) {
	t.Helper()

	require.NotNil(t, combat, "Combat should not be nil")
	assert.NotEmpty(t, combat.ID, "Combat ID should not be empty")
	assert.NotEmpty(t, combat.GameSessionID, "Combat GameSessionID should not be empty")
	assert.NotEmpty(t, combat.Status, "Combat Status should not be empty")
	assert.GreaterOrEqual(t, combat.Round, 1, "Combat Round should be at least 1")
	assert.GreaterOrEqual(t, combat.CurrentTurn, 0, "Combat CurrentTurn should not be negative")

	if len(combat.Combatants) > 0 {
		assert.Less(t, combat.CurrentTurn, len(combat.Combatants),
			"CurrentTurn should be within combatants range")

		for i, combatant := range combat.Combatants {
			AssertValidCombatant(t, &combatant, i)
		}
	}
}

// AssertValidCombatant validates a combatant
func AssertValidCombatant(t *testing.T, combatant *models.Combatant, index int) {
	t.Helper()

	assert.NotEmpty(t, combatant.ID, "Combatant[%d] ID should not be empty", index)
	assert.NotEmpty(t, combatant.Name, "Combatant[%d] Name should not be empty", index)
	assert.NotEmpty(t, combatant.Type, "Combatant[%d] Type should not be empty", index)
	assert.Contains(t, []models.CombatantType{models.CombatantTypeCharacter, models.CombatantTypeNPC},
		combatant.Type, "Combatant[%d] Type should be valid", index)
	assert.GreaterOrEqual(t, combatant.HP, 0, "Combatant[%d] HP should not be negative", index)
	assert.Greater(t, combatant.MaxHP, 0, "Combatant[%d] MaxHP should be greater than 0", index)
	assert.LessOrEqual(t, combatant.HP, combatant.MaxHP,
		"Combatant[%d] HP should not exceed MaxHP", index)
}

// AssertValidCombatAction validates a combat action
func AssertValidCombatAction(t *testing.T, action *models.CombatAction) {
	t.Helper()

	require.NotNil(t, action, "CombatAction should not be nil")
	assert.NotEmpty(t, action.CombatID, "CombatAction CombatID should not be empty")
	assert.NotEmpty(t, action.ActorID, "CombatAction ActorID should not be empty")
	assert.NotEmpty(t, action.ActionType, "CombatAction ActionType should not be empty")
	assert.Contains(t, []string{"attack", "spell", "move", "dash", "dodge", "help", "hide", "ready", "custom"},
		action.ActionType, "CombatAction ActionType should be valid")

	if action.ActionType == "attack" || action.ActionType == "spell" {
		assert.NotEmpty(t, action.TargetID, "CombatAction TargetID should not be empty for attacks/spells")
	}
}

// AssertValidItem validates an item
func AssertValidItem(t *testing.T, item *models.Item) {
	t.Helper()

	require.NotNil(t, item, "Item should not be nil")
	assert.NotEmpty(t, item.ID, "Item ID should not be empty")
	assert.NotEmpty(t, item.Name, "Item Name should not be empty")
	assert.NotEmpty(t, item.Type, "Item Type should not be empty")
	assert.NotEmpty(t, item.Rarity, "Item Rarity should not be empty")
	assert.Contains(t, []string{"common", "uncommon", "rare", "very_rare", "legendary", "artifact"},
		item.Rarity, "Item Rarity should be valid")
}

// AssertValidInventory validates inventory items
func AssertValidInventory(t *testing.T, items []models.Item) {
	t.Helper()

	require.NotNil(t, items, "Inventory items should not be nil")
	for i, item := range items {
		AssertValidItem(t, &item)
		assert.Greater(t, i+1, 0, "Item index should be valid")
	}
}

// AssertValidDiceRoll validates a dice roll result
func AssertValidDiceRoll(t *testing.T, roll *models.DiceRoll) {
	t.Helper()

	require.NotNil(t, roll, "DiceRoll should not be nil")
	assert.NotEmpty(t, roll.ID, "DiceRoll ID should not be empty")
	assert.NotEmpty(t, roll.RollNotation, "DiceRoll RollNotation should not be empty")
	assert.NotEmpty(t, roll.Purpose, "DiceRoll Purpose should not be empty")
	assert.GreaterOrEqual(t, roll.Total, 0, "DiceRoll Total should not be negative")
	assert.NotEmpty(t, roll.Results, "DiceRoll should have at least one result")

	// Validate individual results
	for i, r := range roll.Results {
		assert.GreaterOrEqual(t, r, 1, "Result[%d] should be at least 1", i)
		// Infer max value from dice type
		var maxValue int
		switch roll.DiceType {
		case "d4":
			maxValue = 4
		case "d6":
			maxValue = 6
		case "d8":
			maxValue = 8
		case "d10":
			maxValue = 10
		case "d12":
			maxValue = 12
		case "d20":
			maxValue = 20
		case "d100":
			maxValue = 100
		default:
			maxValue = 20
		}
		assert.LessOrEqual(t, r, maxValue, "Result[%d] should not exceed dice max", i)
	}
}

// AssertValidUser validates a user
func AssertValidUser(t *testing.T, user *models.User) {
	t.Helper()

	require.NotNil(t, user, "User should not be nil")
	assert.NotEmpty(t, user.ID, "User ID should not be empty")
	assert.NotEmpty(t, user.Username, "User Username should not be empty")
	assert.NotEmpty(t, user.Email, "User Email should not be empty")
	assert.NotEmpty(t, user.Role, "User Role should not be empty")
	assert.Contains(t, []string{"player", "dm", "admin"}, user.Role, "User Role should be valid")
}

// AssertValidEncounter validates an encounter
func AssertValidEncounter(t *testing.T, encounter *models.Encounter) {
	t.Helper()

	require.NotNil(t, encounter, "Encounter should not be nil")
	assert.NotEmpty(t, encounter.ID, "Encounter ID should not be empty")
	assert.NotEmpty(t, encounter.Name, "Encounter Name should not be empty")
	assert.NotEmpty(t, encounter.EncounterType, "Encounter Type should not be empty")
	assert.NotEmpty(t, encounter.Difficulty, "Encounter Difficulty should not be empty")
	assert.Contains(t, []string{"easy", "medium", "hard", "deadly"}, encounter.Difficulty,
		"Encounter Difficulty should be valid")
	assert.GreaterOrEqual(t, encounter.TotalXP, 0, "Encounter TotalXP should not be negative")
}

// AssertEqualJSON compares two objects via JSON marshaling (useful for complex structs)
func AssertEqualJSON(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err, "Failed to marshal expected value")

	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err, "Failed to marshal actual value")

	assert.JSONEq(t, string(expectedJSON), string(actualJSON), msgAndArgs...)
}

// AssertContainsString checks if a string slice contains a value
func AssertContainsString(t *testing.T, slice []string, value string, msgAndArgs ...interface{}) {
	t.Helper()

	assert.Contains(t, slice, value, msgAndArgs...)
}

// AssertUUID validates a string is a valid UUID
func AssertUUID(t *testing.T, id string, fieldName string) {
	t.Helper()

	assert.NotEmpty(t, id, "%s should not be empty", fieldName)
	assert.Regexp(t, "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", id,
		"%s should be a valid UUID", fieldName)
}
