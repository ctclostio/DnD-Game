package testutil

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// UserFixture creates a test user
func UserFixture(t *testing.T) *models.User {
	t.Helper()
	return &models.User{
		ID:           uuid.New().String(),
		Username:     "testuser_" + uuid.New().String()[:8],
		Email:        "test_" + uuid.New().String()[:8] + "@example.com",
		PasswordHash: "hashedpassword123",
		Role:         "player",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// CharacterFixture creates a test character
func CharacterFixture(t *testing.T, userID string) *models.Character {
	t.Helper()
	return &models.Character{
		ID:               uuid.New().String(),
		UserID:           userID,
		Name:             "Test Character",
		Race:             "Human",
		Class:            "Fighter",
		Level:            1,
		ExperiencePoints: 0,
		HitPoints:        10,
		MaxHitPoints:     10,
		ArmorClass:       15,
		Initiative:       2,
		Speed:            30,
		Attributes: models.Attributes{
			Strength:     16,
			Dexterity:    14,
			Constitution: 15,
			Intelligence: 10,
			Wisdom:       12,
			Charisma:     8,
		},
		Skills: []models.Skill{
			{Name: "Athletics", Modifier: 3, Proficiency: true},
			{Name: "Intimidation", Modifier: -1, Proficiency: false},
		},
		SavingThrows: models.SavingThrows{
			Strength:     models.SavingThrow{Modifier: 3, Proficiency: true},
			Dexterity:    models.SavingThrow{Modifier: 2, Proficiency: false},
			Constitution: models.SavingThrow{Modifier: 2, Proficiency: true},
			Intelligence: models.SavingThrow{Modifier: 0, Proficiency: false},
			Wisdom:       models.SavingThrow{Modifier: 1, Proficiency: false},
			Charisma:     models.SavingThrow{Modifier: -1, Proficiency: false},
		},
		Equipment:        []models.Item{},
		ProficiencyBonus: 2,
		Features: []models.Feature{
			{Name: "Fighting Style", Description: "Choose a fighting style", Level: 1, Source: "Fighter"},
			{Name: "Second Wind", Description: "Regain hit points", Level: 1, Source: "Fighter"},
		},
		Spells: models.SpellData{
			SpellSlots:  []models.SpellSlot{},
			SpellsKnown: []models.Spell{},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GameSessionFixture creates a test game session
func GameSessionFixture(t *testing.T, dmID string) *models.GameSession {
	t.Helper()
	return &models.GameSession{
		ID:          uuid.New().String(),
		Name:        "Test Campaign",
		Description: "A test campaign for unit testing",
		DMID:        dmID,
		Status:      models.GameStatusActive,
		State:       map[string]interface{}{},
		CreatedAt:   time.Now(),
	}
}

// NPCFixture creates a test NPC
func NPCFixture(t *testing.T, sessionID string) *models.NPC {
	t.Helper()
	return &models.NPC{
		ID:           uuid.New().String(),
		Name:         "Guard Captain",
		Type:         "humanoid",
		Size:         "medium",
		Alignment:    "lawful neutral",
		ArmorClass:   16,
		HitPoints:    20,
		MaxHitPoints: 20,
		Speed:        map[string]int{"walk": 30},
		Attributes: models.Attributes{
			Strength:     15,
			Dexterity:    12,
			Constitution: 14,
			Intelligence: 10,
			Wisdom:       11,
			Charisma:     12,
		},
		SavingThrows: models.SavingThrows{
			Strength:     models.SavingThrow{Modifier: 4, Proficiency: true},
			Dexterity:    models.SavingThrow{Modifier: 1, Proficiency: false},
			Constitution: models.SavingThrow{Modifier: 3, Proficiency: true},
			Intelligence: models.SavingThrow{Modifier: 0, Proficiency: false},
			Wisdom:       models.SavingThrow{Modifier: 0, Proficiency: false},
			Charisma:     models.SavingThrow{Modifier: 1, Proficiency: false},
		},
		ChallengeRating:  2,
		ExperiencePoints: 450,
		GameSessionID:    sessionID,
		Actions: []models.NPCAction{
			{
				Name:        "Longsword",
				Description: "Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 6 (1d8 + 2) slashing damage.",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CampaignFixture creates a test campaign
// TODO: Campaign model doesn't exist yet
/*
func CampaignFixture(t *testing.T, dmID uuid.UUID) *models.Campaign {
	t.Helper()
	return &models.Campaign{
		ID:          uuid.New(),
		Name:        "The Lost Mines",
		Description: "Adventure in the Sword Coast",
		DmID:        dmID,
		Status:      "active",
		Settings: models.CampaignSettings{
			World:        "Forgotten Realms",
			StartingLevel: 1,
			MaxLevel:     5,
		},
		Milestones: []models.CampaignMilestone{
			{
				ID:          uuid.New(),
				Title:       "Arrival at Phandalin",
				Description: "The party arrives at the frontier town",
				Completed:   false,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
}
*/

// DiceRollFixture creates a test dice roll
func DiceRollFixture(t *testing.T, userID, sessionID uuid.UUID) *models.DiceRoll {
	t.Helper()
	return &models.DiceRoll{
		ID:            uuid.New().String(),
		GameSessionID: sessionID.String(),
		UserID:        userID.String(),
		DiceType:      "d20",
		Count:         1,
		Modifier:      3,
		Results:       []int{15},
		Total:         18,
		Purpose:       "attack",
		RollNotation:  "1d20+3",
		Timestamp:     time.Now(),
	}
}

// InventoryItemFixture creates a test inventory item
func InventoryItemFixture(t *testing.T, characterID string) *models.InventoryItem {
	t.Helper()
	return &models.InventoryItem{
		ID:          uuid.New().String(),
		CharacterID: characterID,
		ItemID:      "item-1",
		Quantity:    1,
		Equipped:    false,
		Attuned:     false,
		Item: &models.Item{
			ID:          "item-1",
			Name:        "Healing Potion",
			Type:        models.ItemTypeConsumable,
			Rarity:      models.ItemRarityCommon,
			Weight:      0.5,
			Value:       50,
			Description: "Restores 2d4+2 hit points",
			Properties: models.ItemProperties{
				"healing": "2d4+2",
				"uses":    1,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// EncounterFixture creates a test encounter
func EncounterFixture(t *testing.T, sessionID string) *models.Encounter {
	t.Helper()
	return &models.Encounter{
		ID:            uuid.New().String(),
		GameSessionID: sessionID,
		Name:          "Goblin Ambush",
		Description:   "Goblins attack from the bushes!",
		Difficulty:    "medium",
		Status:        "prepared",
		Enemies: []models.EncounterEnemy{
			{
				ID:              uuid.New().String(),
				Name:            "Goblin",
				Quantity:        4,
				ChallengeRating: 0.25,
			},
		},
		TotalXP:    200,
		AdjustedXP: 200,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}
