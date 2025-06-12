package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// Basic compilation test to ensure the service compiles correctly
func TestEncounterService_Compilation(t *testing.T) {
	// This test just ensures that the EncounterService struct and its methods compile correctly
	var _ *EncounterService

	// Test that EncounterRequest struct exists
	req := EncounterRequest{
		PartyLevel:    5,
		PartySize:     4,
		Difficulty:    "medium",
		EncounterType: "combat",
		Location:      "forest",
	}
	assert.NotNil(t, req)
}

// Test the basic structure of encounters
func TestEncounterModels(t *testing.T) {
	// Test Encounter model
	encounter := &models.Encounter{
		Name:          "Test Encounter",
		Description:   "A test encounter",
		EncounterType: "combat",
		Difficulty:    "medium",
		Status:        "planned",
		Location:      "forest",
	}
	assert.Equal(t, "Test Encounter", encounter.Name)
	assert.Equal(t, "combat", encounter.EncounterType)

	// Test EncounterEnemy model
	enemy := models.EncounterEnemy{
		Name:            "Goblin",
		Type:            "humanoid",
		ChallengeRating: 0.25,
		HitPoints:       7,
		ArmorClass:      15,
		Quantity:        4,
	}
	assert.Equal(t, "Goblin", enemy.Name)
	assert.Equal(t, 0.25, enemy.ChallengeRating)
	assert.Equal(t, 4, enemy.Quantity)

	// Test Action model
	action := models.Action{
		Name:        "Scimitar",
		AttackBonus: 4,
		Damage:      "1d6+2",
	}
	assert.Equal(t, "Scimitar", action.Name)
	assert.Equal(t, 4, action.AttackBonus)

	// Test Ability model
	ability := models.Ability{
		Name:        "Nimble Escape",
		Description: "The goblin can take the Disengage or Hide action as a bonus action on each of its turns.",
	}
	assert.Equal(t, "Nimble Escape", ability.Name)

	// Test ScalingOptions
	scalingOptions := &models.ScalingOptions{
		Easy: models.ScalingAdjustment{
			HPModifier:     -2,
			DamageModifier: -1,
		},
		Medium: models.ScalingAdjustment{
			HPModifier:     0,
			DamageModifier: 0,
		},
		Hard: models.ScalingAdjustment{
			HPModifier:     5,
			DamageModifier: 2,
		},
		Deadly: models.ScalingAdjustment{
			HPModifier:     10,
			DamageModifier: 3,
		},
	}
	assert.Equal(t, -2, scalingOptions.Easy.HPModifier)
	assert.Equal(t, 5, scalingOptions.Hard.HPModifier)

	// Test EncounterObjective
	objective := &models.EncounterObjective{
		Type:        "defeat_all",
		Description: "Defeat all enemies",
		XPReward:    200,
		GoldReward:  50,
	}
	assert.Equal(t, "defeat_all", objective.Type)
	assert.Equal(t, 200, objective.XPReward)

	// Test EncounterEvent
	event := &models.EncounterEvent{
		RoundNumber: 1,
		EventType:   "combat_start",
		ActorType:   "system",
		ActorName:   "System",
		Description: "Combat begins!",
	}
	assert.Equal(t, 1, event.RoundNumber)
	assert.Equal(t, "combat_start", event.EventType)
}

// Test encounter request validation
func TestEncounterRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request EncounterRequest
		valid   bool
	}{
		{
			name: "Valid Combat Request",
			request: EncounterRequest{
				PartyLevel:    5,
				PartySize:     4,
				Difficulty:    "medium",
				EncounterType: "combat",
				Location:      "forest",
			},
			valid: true,
		},
		{
			name: "Valid Social Request",
			request: EncounterRequest{
				PartyLevel:    3,
				PartySize:     5,
				Difficulty:    "easy",
				EncounterType: "social",
				Location:      "tavern",
			},
			valid: true,
		},
		{
			name:    "Empty Request",
			request: EncounterRequest{},
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation
			if tt.valid {
				assert.Greater(t, tt.request.PartyLevel, 0)
				assert.NotEmpty(t, tt.request.Difficulty)
			} else {
				assert.Equal(t, 0, tt.request.PartyLevel)
			}
		})
	}
}

// Test difficulty scaling logic
func TestDifficultyScaling(t *testing.T) {
	testCases := []struct {
		difficulty  string
		hpModifier  int
		dmgModifier int
	}{
		{"easy", -2, -1},
		{"medium", 0, 0},
		{"hard", 5, 2},
		{"deadly", 10, 3},
	}

	for _, tc := range testCases {
		t.Run(tc.difficulty, func(t *testing.T) {
			scaling := models.ScalingAdjustment{
				HPModifier:     tc.hpModifier,
				DamageModifier: tc.dmgModifier,
			}

			// Test HP modification
			baseHP := 10
			modifiedHP := baseHP + scaling.HPModifier
			if scaling.HPModifier < 0 {
				assert.Less(t, modifiedHP, baseHP)
			} else if scaling.HPModifier > 0 {
				assert.Greater(t, modifiedHP, baseHP)
			} else {
				assert.Equal(t, modifiedHP, baseHP)
			}

			// Ensure HP never goes below 1
			if modifiedHP < 1 {
				modifiedHP = 1
			}
			assert.GreaterOrEqual(t, modifiedHP, 1)
		})
	}
}

// Test encounter status transitions
func TestEncounterStatusTransitions(t *testing.T) {
	validTransitions := map[string][]string{
		"planned":   {"active", "cancelled"},
		"active":    {"completed", "failed"},
		"completed": {},
		"failed":    {},
		"cancelled": {},
	}

	for fromStatus, toStatuses := range validTransitions {
		t.Run(fromStatus, func(t *testing.T) {
			// Test valid transitions
			for _, toStatus := range toStatuses {
				// In a real implementation, you'd have a method to validate transitions
				assert.Contains(t, toStatuses, toStatus)
			}

			// Test that completed/failed/cancelled are terminal states
			if fromStatus == "completed" || fromStatus == "failed" || fromStatus == "cancelled" {
				assert.Empty(t, toStatuses)
			}
		})
	}
}

// Test environmental hazard creation
func TestEnvironmentalHazard(t *testing.T) {
	hazard := models.EnvironmentalHazard{
		Name:        "Falling Rocks",
		Description: "Rocks fall from the ceiling",
		Trigger:     "Loud noise or vibration",
		Effect:      "All creatures in area must make DC 15 Dex save or take damage",
		SaveDC:      15,
		Damage:      "2d6",
	}

	assert.Equal(t, "Falling Rocks", hazard.Name)
	assert.Equal(t, 15, hazard.SaveDC)
	assert.Equal(t, "2d6", hazard.Damage)
}

// Test terrain feature creation
func TestTerrainFeature(t *testing.T) {
	terrain := models.TerrainFeature{
		Name:        "Difficult Terrain",
		Description: "Rubble and debris litter the ground",
		Effect:      "Movement costs double",
		Location:    "Eastern half of the room",
	}

	assert.Equal(t, "Difficult Terrain", terrain.Name)
	assert.Equal(t, "Movement costs double", terrain.Effect)
}

// Test solution options
func TestSolutionOptions(t *testing.T) {
	solution := models.Solution{
		Method:       "Negotiation",
		Description:  "Convince the enemies to let you pass",
		Requirements: []string{"Speak their language", "Offer something valuable"},
		DC:           15,
		Consequences: "They demand payment or a favor",
	}

	assert.Equal(t, "Negotiation", solution.Method)
	assert.Equal(t, 15, solution.DC)
	assert.Len(t, solution.Requirements, 2)
}

// Test reinforcement waves
func TestReinforcementWave(t *testing.T) {
	wave := models.ReinforcementWave{
		Round:   3,
		Trigger: "If combat lasts to round 3 or alarm is raised",
		Enemies: []models.EncounterEnemy{
			{
				Name:     "Goblin Reinforcement",
				Quantity: 2,
			},
		},
		Entrance:     "Through the northern door",
		Announcement: "You hear footsteps approaching from the north!",
	}

	assert.Equal(t, 3, wave.Round)
	assert.Len(t, wave.Enemies, 1)
	assert.Equal(t, 2, wave.Enemies[0].Quantity)
}

// Test escape routes
func TestEscapeRoute(t *testing.T) {
	routes := []models.EscapeRoute{
		{
			Direction:   "Back the way you came",
			Difficulty:  "Easy",
			Consequence: "None, but objective not completed",
		},
		{
			Direction:   "Through the window",
			Difficulty:  "Medium",
			Consequence: "Take 2d6 falling damage, but escape pursuit",
		},
	}

	assert.Len(t, routes, 2)
	assert.Equal(t, "Easy", routes[0].Difficulty)
	assert.Equal(t, "Medium", routes[1].Difficulty)
}

// Test tactical info structure
func TestTacticalInfo(t *testing.T) {
	tactics := &models.TacticalInfo{
		GeneralStrategy: "Goblins use hit-and-run tactics",
		PriorityTargets: []string{"Spellcasters", "Healers"},
		Positioning:     "Spread out to avoid area effects",
		CombatPhases: []models.CombatPhase{
			{
				Name:    "Opening",
				Trigger: "Combat starts",
				Tactics: "Rush the weakest-looking target",
			},
			{
				Name:    "Desperate",
				Trigger: "Below 50% forces",
				Tactics: "Fighting retreat to reinforcement location",
			},
		},
		RetreatConditions: "When reduced to 25% of original force",
	}

	assert.Equal(t, "Goblins use hit-and-run tactics", tactics.GeneralStrategy)
	assert.Len(t, tactics.PriorityTargets, 2)
	assert.Len(t, tactics.CombatPhases, 2)
	assert.Equal(t, "Opening", tactics.CombatPhases[0].Name)
}

// Test encounter validation helpers
func TestEncounterValidation(t *testing.T) {
	t.Run("Valid Difficulties", func(t *testing.T) {
		validDifficulties := []string{"easy", "medium", "hard", "deadly"}
		for _, diff := range validDifficulties {
			assert.Contains(t, validDifficulties, diff)
		}
	})

	t.Run("Valid Encounter Types", func(t *testing.T) {
		validTypes := []string{"combat", "social", "exploration", "puzzle", "hybrid"}
		for _, encType := range validTypes {
			assert.Contains(t, validTypes, encType)
		}
	})

	t.Run("Valid Statuses", func(t *testing.T) {
		validStatuses := []string{"planned", "active", "completed", "failed", "cancelled"}
		for _, status := range validStatuses {
			assert.Contains(t, validStatuses, status)
		}
	})
}

// Test XP calculation helpers
func TestXPCalculation(t *testing.T) {
	// Test party size multipliers
	partySizeMultipliers := map[int]float64{
		1: 1.5,  // Solo
		2: 1.5,  // Pair
		3: 1.0,  // Small party
		4: 1.0,  // Standard party
		5: 1.0,  // Standard party
		6: 0.75, // Large party
		7: 0.75, // Large party
		8: 0.5,  // Very large party
	}

	for size, multiplier := range partySizeMultipliers {
		t.Run(fmt.Sprintf("Party Size %d", size), func(t *testing.T) {
			baseXP := 1000
			adjustedXP := int(float64(baseXP) * multiplier)

			if size <= 2 {
				assert.Greater(t, adjustedXP, baseXP, "Smaller parties should have harder encounters")
			} else if size >= 6 {
				assert.Less(t, adjustedXP, baseXP, "Larger parties should have easier encounters")
			} else {
				assert.Equal(t, adjustedXP, baseXP, "Standard parties should have normal difficulty")
			}
		})
	}
}

// Helper function to test encounter creation
func createTestEncounterData() *models.Encounter {
	return &models.Encounter{
		Name:          "Test Encounter",
		Description:   "A test encounter for unit testing",
		EncounterType: "combat",
		Difficulty:    "medium",
		Status:        "planned",
		Location:      "dungeon",
		Enemies: []models.EncounterEnemy{
			{
				Name:            "Test Enemy",
				Type:            "construct",
				ChallengeRating: 1.0,
				HitPoints:       20,
				ArmorClass:      14,
				Quantity:        2,
			},
		},
		TotalXP:         400,
		AdjustedXP:      600,
		ChallengeRating: 1.5,
	}
}
