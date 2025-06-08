package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// Builder interface for all test builders
type Builder interface {
	Build() interface{}
}

// UserBuilder provides a fluent interface for creating test users
type UserBuilder struct {
	user models.User
}

// NewUserBuilder creates a new UserBuilder with sensible defaults
func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		user: models.User{
			ID:        "1",
			Username:  "testuser",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

func (b *UserBuilder) WithID(id int64) *UserBuilder {
	b.user.ID = fmt.Sprintf("%d", id)
	return b
}

func (b *UserBuilder) WithUsername(username string) *UserBuilder {
	b.user.Username = username
	return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.user.Email = email
	return b
}

func (b *UserBuilder) WithPassword(password string) *UserBuilder {
	b.user.PasswordHash = password // In real tests, this would be hashed
	return b
}

func (b *UserBuilder) Build() *models.User {
	return &b.user
}

// CharacterBuilder provides a fluent interface for creating test characters
type CharacterBuilder struct {
	character models.Character
}

// NewCharacterBuilder creates a new CharacterBuilder with D&D defaults
func NewCharacterBuilder() *CharacterBuilder {
	return &CharacterBuilder{
		character: models.Character{
			ID:               "1",
			UserID:           "1",
			Name:             "Gandalf",
			Race:             "Human",
			Class:            "Wizard",
			Level:            1,
			ExperiencePoints: 0,
			HitPoints:        10,
			MaxHitPoints:     10,
			ArmorClass:       12,
			Initiative:       2,
			Speed:            30,
			Attributes: models.Attributes{
				Strength:     10,
				Dexterity:    14,
				Constitution: 12,
				Intelligence: 16,
				Wisdom:       14,
				Charisma:     12,
			},
			Skills:         []models.Skill{{Name: "Arcana", Modifier: 5, Proficiency: true}, {Name: "Investigation", Modifier: 5, Proficiency: true}},
			Proficiencies:  models.Proficiencies{Languages: []string{"Common"}, Tools: []string{}, Weapons: []string{"Simple Weapons"}, Armor: []string{"Light Armor"}},
			Equipment:      []models.Item{},
			Spells:         models.SpellData{},
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}
}

func (b *CharacterBuilder) WithID(id int64) *CharacterBuilder {
	b.character.ID = fmt.Sprintf("%d", id)
	return b
}

func (b *CharacterBuilder) WithUserID(userID int64) *CharacterBuilder {
	b.character.UserID = fmt.Sprintf("%d", userID)
	return b
}

func (b *CharacterBuilder) WithName(name string) *CharacterBuilder {
	b.character.Name = name
	return b
}

func (b *CharacterBuilder) WithClass(class string) *CharacterBuilder {
	b.character.Class = class
	return b
}

func (b *CharacterBuilder) WithRace(race string) *CharacterBuilder {
	b.character.Race = race
	return b
}

func (b *CharacterBuilder) WithLevel(level int) *CharacterBuilder {
	b.character.Level = level
	return b
}

func (b *CharacterBuilder) WithAbilities(attributes models.Attributes) *CharacterBuilder {
	b.character.Attributes = attributes
	return b
}

func (b *CharacterBuilder) WithHP(current, max int) *CharacterBuilder {
	b.character.HitPoints = current
	b.character.MaxHitPoints = max
	return b
}

func (b *CharacterBuilder) Build() *models.Character {
	return &b.character
}

// GameSessionBuilder provides a fluent interface for creating test game sessions
type GameSessionBuilder struct {
	session models.GameSession
}

// NewGameSessionBuilder creates a new GameSessionBuilder
func NewGameSessionBuilder() *GameSessionBuilder {
	return &GameSessionBuilder{
		session: models.GameSession{
			ID:          "1",
			Name:        "Test Campaign",
			DMID:        "1",
			Description: "Test session",
			Status:      models.GameStatusActive,
			State:       map[string]interface{}{},
			CreatedAt:   time.Now(),
		},
	}
}

func (b *GameSessionBuilder) WithID(id int64) *GameSessionBuilder {
	b.session.ID = fmt.Sprintf("%d", id)
	return b
}

func (b *GameSessionBuilder) WithName(name string) *GameSessionBuilder {
	b.session.Name = name
	return b
}

func (b *GameSessionBuilder) WithDM(dmID int64) *GameSessionBuilder {
	b.session.DMID = fmt.Sprintf("%d", dmID)
	return b
}

func (b *GameSessionBuilder) WithStatus(status models.GameStatus) *GameSessionBuilder {
	b.session.Status = status
	return b
}

func (b *GameSessionBuilder) Build() *models.GameSession {
	return &b.session
}

// CombatBuilder provides a fluent interface for creating test combat scenarios
type CombatBuilder struct {
	combat models.Combat
}

// NewCombatBuilder creates a new CombatBuilder
func NewCombatBuilder() *CombatBuilder {
	return &CombatBuilder{
		combat: models.Combat{
			ID:              "combat-1",
			GameSessionID:   "session-1",
			Name:            "Test Combat",
			Round:           1,
			CurrentTurn:     0,
			Combatants:      []models.Combatant{},
			TurnOrder:       []string{},
			ActiveEffects:   []models.CombatEffect{},
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}
}

func (b *CombatBuilder) WithID(id string) *CombatBuilder {
	b.combat.ID = id
	return b
}

func (b *CombatBuilder) WithGameSession(sessionID string) *CombatBuilder {
	b.combat.GameSessionID = sessionID
	return b
}

func (b *CombatBuilder) WithParticipants(participants ...models.Combatant) *CombatBuilder {
	b.combat.Combatants = participants
	// Set turn order based on participant IDs
	turnOrder := make([]string, len(participants))
	for i, p := range participants {
		turnOrder[i] = p.ID
	}
	b.combat.TurnOrder = turnOrder
	return b
}

func (b *CombatBuilder) WithRound(round int) *CombatBuilder {
	b.combat.Round = round
	return b
}

func (b *CombatBuilder) Build() *models.Combat {
	return &b.combat
}

// CombatantBuilder for creating combat participants
type CombatantBuilder struct {
	participant models.Combatant
}

// NewCombatantBuilder creates a new participant builder
func NewCombatantBuilder() *CombatantBuilder {
	return &CombatantBuilder{
		participant: models.Combatant{
			ID:           "participant-1",
			Name:         "Fighter",
			Type:         "character",
			Initiative:   15,
			HP:           20,
			MaxHP:        20,
			AC:           16,
			Conditions:   []string{},
			IsActive:     true,
			CharacterID:  1,
		},
	}
}

func (b *CombatantBuilder) WithID(id string) *CombatantBuilder {
	b.participant.ID = id
	return b
}

func (b *CombatantBuilder) WithName(name string) *CombatantBuilder {
	b.participant.Name = name
	return b
}

func (b *CombatantBuilder) WithType(pType string) *CombatantBuilder {
	b.participant.Type = pType
	return b
}

func (b *CombatantBuilder) WithInitiative(init int) *CombatantBuilder {
	b.participant.Initiative = init
	return b
}

func (b *CombatantBuilder) WithHP(current, max int) *CombatantBuilder {
	b.participant.HP = current
	b.participant.MaxHP = max
	return b
}

func (b *CombatantBuilder) AsNPC() *CombatantBuilder {
	b.participant.Type = "npc"
	b.participant.CharacterID = 0
	return b
}

func (b *CombatantBuilder) Build() models.Combatant {
	return b.participant
}

// DiceRollBuilder for creating test dice rolls
type DiceRollBuilder struct {
	roll models.DiceRoll
}

// NewDiceRollBuilder creates a new dice roll builder
func NewDiceRollBuilder() *DiceRollBuilder {
	return &DiceRollBuilder{
		roll: models.DiceRoll{
			ID:            1,
			UserID:        1,
			CharacterID:   1,
			GameSessionID: 1,
			RollType:      "attack",
			DiceNotation:  "1d20+5",
			Result:        18,
			Rolls:         []int{13},
			Modifiers:     5,
			Purpose:       "Attack roll",
			CreatedAt:     time.Now(),
		},
	}
}

func (b *DiceRollBuilder) WithType(rollType string) *DiceRollBuilder {
	b.roll.RollType = rollType
	return b
}

func (b *DiceRollBuilder) WithNotation(notation string) *DiceRollBuilder {
	b.roll.DiceNotation = notation
	return b
}

func (b *DiceRollBuilder) WithResult(result int, rolls []int) *DiceRollBuilder {
	b.roll.Result = result
	b.roll.Rolls = rolls
	return b
}

func (b *DiceRollBuilder) Build() *models.DiceRoll {
	return &b.roll
}

// InventoryItemBuilder for creating test inventory items
type InventoryItemBuilder struct {
	item models.InventoryItem
}

// NewInventoryItemBuilder creates a new inventory item builder
func NewInventoryItemBuilder() *InventoryItemBuilder {
	return &InventoryItemBuilder{
		item: models.InventoryItem{
			ID:          1,
			CharacterID: 1,
			Name:        "Longsword",
			Type:        "weapon",
			Quantity:    1,
			Weight:      3.0,
			Value:       15,
			Properties: map[string]interface{}{
				"damage":     "1d8",
				"damageType": "slashing",
				"versatile":  "1d10",
			},
			Equipped:    true,
			Description: "A standard longsword",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
}

func (b *InventoryItemBuilder) WithName(name string) *InventoryItemBuilder {
	b.item.Name = name
	return b
}

func (b *InventoryItemBuilder) WithType(itemType string) *InventoryItemBuilder {
	b.item.Type = itemType
	return b
}

func (b *InventoryItemBuilder) WithQuantity(qty int) *InventoryItemBuilder {
	b.item.Quantity = qty
	return b
}

func (b *InventoryItemBuilder) AsArmor(ac int) *InventoryItemBuilder {
	b.item.Type = "armor"
	b.item.Properties = map[string]interface{}{
		"armorClass": ac,
		"stealthDisadvantage": false,
	}
	return b
}

func (b *InventoryItemBuilder) AsMagicItem(rarity string) *InventoryItemBuilder {
	b.item.Type = "magic"
	b.item.Properties["rarity"] = rarity
	b.item.Properties["requiresAttunement"] = true
	return b
}

func (b *InventoryItemBuilder) Build() *models.InventoryItem {
	return &b.item
}

// TestScenario provides a complete test scenario with related entities
type TestScenario struct {
	t           *testing.T
	Users       []*models.User
	Characters  []*models.Character
	GameSession *models.GameSession
	Combat      *models.Combat
	Items       []*models.InventoryItem
}

// NewTestScenario creates a complete test scenario
func NewTestScenario(t *testing.T) *TestScenario {
	dm := NewUserBuilder().WithID(1).WithUsername("dm_user").Build()
	player1 := NewUserBuilder().WithID(2).WithUsername("player1").Build()
	player2 := NewUserBuilder().WithID(3).WithUsername("player2").Build()

	dmChar := NewCharacterBuilder().
		WithID(1).
		WithUserID(dm.ID).
		WithName("DM Character").
		Build()

	playerChar1 := NewCharacterBuilder().
		WithID(2).
		WithUserID(player1.ID).
		WithName("Aragorn").
		WithClass("Fighter").
		WithLevel(5).
		Build()

	playerChar2 := NewCharacterBuilder().
		WithID(3).
		WithUserID(player2.ID).
		WithName("Legolas").
		WithClass("Ranger").
		WithLevel(5).
		Build()

	session := NewGameSessionBuilder().
		WithID(1).
		WithDM(dm.ID).
		WithName("Test Adventure").
		Build()

	return &TestScenario{
		t:           t,
		Users:       []*models.User{dm, player1, player2},
		Characters:  []*models.Character{dmChar, playerChar1, playerChar2},
		GameSession: session,
	}
}

// WithCombat adds a combat scenario
func (s *TestScenario) WithCombat() *TestScenario {
	participants := []models.Combatant{
		NewCombatantBuilder().
			WithID("char-2").
			WithName(s.Characters[1].Name).
			WithInitiative(18).
			Build(),
		NewCombatantBuilder().
			WithID("char-3").
			WithName(s.Characters[2].Name).
			WithInitiative(15).
			Build(),
		NewCombatantBuilder().
			WithID("npc-1").
			WithName("Goblin").
			AsNPC().
			WithInitiative(12).
			WithHP(7, 7).
			Build(),
	}

	s.Combat = NewCombatBuilder().
		WithGameSession(s.GameSession.ID).
		WithParticipants(participants...).
		Build()

	return s
}

// WithItems adds inventory items
func (s *TestScenario) WithItems() *TestScenario {
	s.Items = []*models.InventoryItem{
		NewInventoryItemBuilder().
			WithName("Longsword +1").
			AsMagicItem("uncommon").
			Build(),
		NewInventoryItemBuilder().
			WithName("Plate Armor").
			AsArmor(18).
			Build(),
		NewInventoryItemBuilder().
			WithName("Healing Potion").
			WithType("consumable").
			WithQuantity(3).
			Build(),
	}
	return s
}

// AssertValid validates the test scenario
func (s *TestScenario) AssertValid() {
	require.NotNil(s.t, s.Users)
	require.NotNil(s.t, s.Characters)
	require.NotNil(s.t, s.GameSession)
	require.Greater(s.t, len(s.Users), 0)
	require.Greater(s.t, len(s.Characters), 0)
}

// TestContext creates a context with common test values
func TestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "test-request-123")
	ctx = context.WithValue(ctx, "user_id", int64(1))
	return ctx
}