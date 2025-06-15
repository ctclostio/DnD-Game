package testhelpers

import (
	"time"

	"github.com/google/uuid"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// CharacterBuilder provides a fluent interface for building test characters
type CharacterBuilder struct {
	character *models.Character
}

// NewCharacterBuilder creates a new character builder with defaults
func NewCharacterBuilder() *CharacterBuilder {
	return &CharacterBuilder{
		character: &models.Character{
			ID:        uuid.New().String(),
			UserID:    uuid.New().String(),
			Name:      "Test Character",
			Race:      "Human",
			Class:     "Fighter",
			Level:     1,
			HitPoints: 10,
			MaxHP:     10,
			Stats: models.CharacterStats{
				Strength:     10,
				Dexterity:    10,
				Constitution: 10,
				Intelligence: 10,
				Wisdom:       10,
				Charisma:     10,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// WithID sets the character ID
func (b *CharacterBuilder) WithID(id string) *CharacterBuilder {
	b.character.ID = id
	return b
}

// WithUserID sets the user ID
func (b *CharacterBuilder) WithUserID(userID string) *CharacterBuilder {
	b.character.UserID = userID
	return b
}

// WithName sets the character name
func (b *CharacterBuilder) WithName(name string) *CharacterBuilder {
	b.character.Name = name
	return b
}

// WithLevel sets the character level
func (b *CharacterBuilder) WithLevel(level int) *CharacterBuilder {
	b.character.Level = level
	return b
}

// WithClass sets the character class
func (b *CharacterBuilder) WithClass(class string) *CharacterBuilder {
	b.character.Class = class
	return b
}

// WithRace sets the character race
func (b *CharacterBuilder) WithRace(race string) *CharacterBuilder {
	b.character.Race = race
	return b
}

// WithStats sets specific stats
func (b *CharacterBuilder) WithStats(str, dex, con, intel, wis, cha int) *CharacterBuilder {
	b.character.Stats = models.CharacterStats{
		Strength:     str,
		Dexterity:    dex,
		Constitution: con,
		Intelligence: intel,
		Wisdom:       wis,
		Charisma:     cha,
	}
	return b
}

// Build returns the built character
func (b *CharacterBuilder) Build() *models.Character {
	return b.character
}

// GameSessionBuilder provides a fluent interface for building test game sessions
type GameSessionBuilder struct {
	session *models.GameSession
}

// NewGameSessionBuilder creates a new game session builder
func NewGameSessionBuilder() *GameSessionBuilder {
	return &GameSessionBuilder{
		session: &models.GameSession{
			ID:          uuid.New().String(),
			Name:        "Test Session",
			Description: "A test game session",
			DmID:        uuid.New().String(),
			Status:      "active",
			Players:     []models.Player{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
}

// WithID sets the session ID
func (b *GameSessionBuilder) WithID(id string) *GameSessionBuilder {
	b.session.ID = id
	return b
}

// WithDM sets the DM ID
func (b *GameSessionBuilder) WithDM(dmID string) *GameSessionBuilder {
	b.session.DmID = dmID
	return b
}

// WithName sets the session name
func (b *GameSessionBuilder) WithName(name string) *GameSessionBuilder {
	b.session.Name = name
	return b
}

// WithStatus sets the session status
func (b *GameSessionBuilder) WithStatus(status string) *GameSessionBuilder {
	b.session.Status = status
	return b
}

// WithPlayer adds a player to the session
func (b *GameSessionBuilder) WithPlayer(userID, characterID string) *GameSessionBuilder {
	b.session.Players = append(b.session.Players, models.Player{
		UserID:      userID,
		CharacterID: characterID,
		JoinedAt:    time.Now(),
	})
	return b
}

// Build returns the built game session
func (b *GameSessionBuilder) Build() *models.GameSession {
	return b.session
}

// CombatBuilder provides a fluent interface for building test combats
type CombatBuilder struct {
	combat *models.Combat
}

// NewCombatBuilder creates a new combat builder
func NewCombatBuilder() *CombatBuilder {
	return &CombatBuilder{
		combat: &models.Combat{
			ID:            uuid.New().String(),
			GameSessionID: uuid.New().String(),
			Status:        "active",
			CurrentTurn:   0,
			Round:         1,
			Combatants:    []models.Combatant{},
			StartedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
}

// WithID sets the combat ID
func (b *CombatBuilder) WithID(id string) *CombatBuilder {
	b.combat.ID = id
	return b
}

// WithGameSession sets the game session ID
func (b *CombatBuilder) WithGameSession(sessionID string) *CombatBuilder {
	b.combat.GameSessionID = sessionID
	return b
}

// WithStatus sets the combat status
func (b *CombatBuilder) WithStatus(status string) *CombatBuilder {
	b.combat.Status = status
	return b
}

// WithCombatant adds a combatant
func (b *CombatBuilder) WithCombatant(name string, initiative int, isPlayer bool) *CombatBuilder {
	combatantType := models.CombatantTypeNPC
	if isPlayer {
		combatantType = models.CombatantTypePlayer
	}
	
	b.combat.Combatants = append(b.combat.Combatants, models.Combatant{
		ID:         uuid.New().String(),
		Name:       name,
		Type:       combatantType,
		Initiative: initiative,
		HitPoints:  20,
		MaxHP:      20,
		Status:     "active",
	})
	return b
}

// Build returns the built combat
func (b *CombatBuilder) Build() *models.Combat {
	return b.combat
}

// ItemBuilder provides a fluent interface for building test items
type ItemBuilder struct {
	item *models.Item
}

// NewItemBuilder creates a new item builder
func NewItemBuilder() *ItemBuilder {
	return &ItemBuilder{
		item: &models.Item{
			ID:          uuid.New().String(),
			Name:        "Test Item",
			Type:        "equipment",
			Rarity:      "common",
			Description: "A test item",
			Properties:  map[string]interface{}{},
			CreatedAt:   time.Now(),
		},
	}
}

// WithID sets the item ID
func (b *ItemBuilder) WithID(id string) *ItemBuilder {
	b.item.ID = id
	return b
}

// WithName sets the item name
func (b *ItemBuilder) WithName(name string) *ItemBuilder {
	b.item.Name = name
	return b
}

// WithType sets the item type
func (b *ItemBuilder) WithType(itemType string) *ItemBuilder {
	b.item.Type = itemType
	return b
}

// WithRarity sets the item rarity
func (b *ItemBuilder) WithRarity(rarity string) *ItemBuilder {
	b.item.Rarity = rarity
	return b
}

// WithProperty adds a property to the item
func (b *ItemBuilder) WithProperty(key string, value interface{}) *ItemBuilder {
	b.item.Properties[key] = value
	return b
}

// Build returns the built item
func (b *ItemBuilder) Build() *models.Item {
	return b.item
}

// Quick builder functions for common test data

// NewTestCharacter creates a basic test character
func NewTestCharacter(userID string) *models.Character {
	return NewCharacterBuilder().WithUserID(userID).Build()
}

// NewTestGameSession creates a basic test game session
func NewTestGameSession(dmID string) *models.GameSession {
	return NewGameSessionBuilder().WithDM(dmID).Build()
}

// NewTestCombat creates a basic test combat
func NewTestCombat(sessionID string) *models.Combat {
	return NewCombatBuilder().
		WithGameSession(sessionID).
		WithCombatant("Fighter", 15, true).
		WithCombatant("Goblin", 12, false).
		Build()
}

// NewTestItem creates a basic test item
func NewTestItem(name string) *models.Item {
	return NewItemBuilder().WithName(name).Build()
}

// TestIDs provides commonly used test IDs
type TestIDs struct {
	UserID1      string
	UserID2      string
	CharacterID1 string
	CharacterID2 string
	SessionID    string
	CombatID     string
	ItemID       string
}

// NewTestIDs generates a set of test IDs
func NewTestIDs() *TestIDs {
	return &TestIDs{
		UserID1:      uuid.New().String(),
		UserID2:      uuid.New().String(),
		CharacterID1: uuid.New().String(),
		CharacterID2: uuid.New().String(),
		SessionID:    uuid.New().String(),
		CombatID:     uuid.New().String(),
		ItemID:       uuid.New().String(),
	}
}