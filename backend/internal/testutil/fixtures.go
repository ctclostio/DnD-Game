package testutil

import (
	"testing"
	"time"

	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/google/uuid"
)

// UserFixture creates a test user
func UserFixture(t *testing.T) *models.User {
	t.Helper()
	return &models.User{
		ID:        uuid.New(),
		Username:  "testuser_" + uuid.New().String()[:8],
		Email:     "test_" + uuid.New().String()[:8] + "@example.com",
		Password:  "hashedpassword123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CharacterFixture creates a test character
func CharacterFixture(t *testing.T, userID uuid.UUID) *models.Character {
	t.Helper()
	return &models.Character{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        "Test Character",
		Race:        "Human",
		Class:       "Fighter",
		Level:       1,
		Experience:  0,
		HitPoints:   10,
		MaxHP:       10,
		ArmorClass:  15,
		Initiative:  2,
		Speed:       30,
		AbilityScores: models.AbilityScores{
			Strength:     16,
			Dexterity:    14,
			Constitution: 15,
			Intelligence: 10,
			Wisdom:       12,
			Charisma:     8,
		},
		Skills: models.Skills{
			Acrobatics:     0,
			AnimalHandling: 1,
			Arcana:         0,
			Athletics:      3,
			Deception:      -1,
			History:        0,
			Insight:        1,
			Intimidation:   -1,
			Investigation:  0,
			Medicine:       1,
			Nature:         0,
			Perception:     1,
			Performance:    -1,
			Persuasion:     -1,
			Religion:       0,
			SleightOfHand:  2,
			Stealth:        2,
			Survival:       1,
		},
		SavingThrows: models.SavingThrows{
			Strength:     3,
			Dexterity:    2,
			Constitution: 2,
			Intelligence: 0,
			Wisdom:       1,
			Charisma:     -1,
		},
		Equipment: []models.Equipment{
			{
				ID:       uuid.New().String(),
				Name:     "Longsword",
				Type:     "Weapon",
				Quantity: 1,
				Weight:   3,
			},
			{
				ID:       uuid.New().String(),
				Name:     "Chain Mail",
				Type:     "Armor",
				Quantity: 1,
				Weight:   55,
			},
		},
		Features: []string{"Fighting Style", "Second Wind"},
		SpellSlots: models.SpellSlots{
			Level1: models.SpellSlotInfo{Max: 0, Used: 0},
			Level2: models.SpellSlotInfo{Max: 0, Used: 0},
			Level3: models.SpellSlotInfo{Max: 0, Used: 0},
			Level4: models.SpellSlotInfo{Max: 0, Used: 0},
			Level5: models.SpellSlotInfo{Max: 0, Used: 0},
			Level6: models.SpellSlotInfo{Max: 0, Used: 0},
			Level7: models.SpellSlotInfo{Max: 0, Used: 0},
			Level8: models.SpellSlotInfo{Max: 0, Used: 0},
			Level9: models.SpellSlotInfo{Max: 0, Used: 0},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GameSessionFixture creates a test game session
func GameSessionFixture(t *testing.T, dmID uuid.UUID) *models.GameSession {
	t.Helper()
	return &models.GameSession{
		ID:          uuid.New(),
		Name:        "Test Campaign",
		Description: "A test campaign for unit testing",
		DmID:        dmID,
		MaxPlayers:  5,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NPCFixture creates a test NPC
func NPCFixture(t *testing.T, sessionID uuid.UUID) *models.NPC {
	t.Helper()
	return &models.NPC{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Name:        "Guard Captain",
		Description: "A grizzled veteran guard",
		Stats: models.NPCStats{
			HitPoints:    20,
			ArmorClass:   16,
			Speed:        30,
			Challenge:    2,
			Abilities:    models.AbilityScores{Strength: 15, Dexterity: 12, Constitution: 14, Intelligence: 10, Wisdom: 11, Charisma: 12},
			SavingThrows: models.SavingThrows{Strength: 4, Dexterity: 1, Constitution: 3, Intelligence: 0, Wisdom: 0, Charisma: 1},
		},
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
func InventoryItemFixture(t *testing.T) *models.InventoryItem {
	t.Helper()
	return &models.InventoryItem{
		ID:          uuid.New(),
		Name:        "Healing Potion",
		Description: "Restores 2d4+2 hit points",
		Type:        "consumable",
		Rarity:      "common",
		Weight:      0.5,
		Value:       50,
		Properties: map[string]interface{}{
			"healing": "2d4+2",
			"uses":    1,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// EncounterFixture creates a test encounter
func EncounterFixture(t *testing.T, sessionID uuid.UUID) *models.Encounter {
	t.Helper()
	return &models.Encounter{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Name:        "Goblin Ambush",
		Description: "Goblins attack from the bushes!",
		Difficulty:  "medium",
		Status:      "prepared",
		Enemies: []models.EncounterEnemy{
			{
				ID:       uuid.New(),
				Name:     "Goblin",
				Quantity: 4,
				CR:       0.25,
			},
		},
		Environment: "forest",
		Rewards: models.EncounterRewards{
			Experience: 200,
			Gold:       50,
			Items:      []string{"Crude weapons", "Goblin ears"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}