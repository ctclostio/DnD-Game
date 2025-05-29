package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/your-username/dnd-game/backend/internal/models"
)

func TestNewCombatService(t *testing.T) {
	service := NewCombatService()
	assert.NotNil(t, service)
}

func TestCombatService_StartCombat(t *testing.T) {
	service := NewCombatService()
	ctx := context.Background()

	t.Run("successful combat start", func(t *testing.T) {
		combatants := []models.Combatant{
			{
				ID:         "char-1",
				Name:       "Fighter",
				Type:       models.CombatantTypeCharacter,
				Initiative: 15,
				HP:         45,
				MaxHP:      45,
				AC:         18,
			},
			{
				ID:         "char-2",
				Name:       "Wizard",
				Type:       models.CombatantTypeCharacter,
				Initiative: 12,
				HP:         25,
				MaxHP:      25,
				AC:         13,
			},
			{
				ID:         "npc-1",
				Name:       "Goblin",
				Type:       models.CombatantTypeNPC,
				Initiative: 10,
				HP:         7,
				MaxHP:      7,
				AC:         15,
			},
		}

		combat, err := service.StartCombat(ctx, "session-123", combatants)

		require.NoError(t, err)
		assert.NotNil(t, combat)
		assert.Equal(t, "session-123", combat.GameSessionID)
		assert.NotEmpty(t, combat.ID)
		assert.Equal(t, 1, combat.Round)
		assert.True(t, combat.IsActive)
		assert.Len(t, combat.Combatants, 3)
		assert.Len(t, combat.TurnOrder, 3)
	})

	t.Run("no combatants", func(t *testing.T) {
		combatants := []models.Combatant{}

		combat, err := service.StartCombat(ctx, "session-456", combatants)

		assert.Error(t, err)
		assert.Nil(t, combat)
	})
}

func TestCombatService_GetCombat(t *testing.T) {
	service := NewCombatService()
	ctx := context.Background()

	t.Run("combat exists", func(t *testing.T) {
		combatants := []models.Combatant{
			{
				ID:         "char-1",
				Name:       "Fighter",
				Initiative: 15,
			},
		}

		originalCombat, _ := service.StartCombat(ctx, "session-123", combatants)

		combat, err := service.GetCombat(ctx, originalCombat.ID)

		require.NoError(t, err)
		assert.Equal(t, originalCombat.ID, combat.ID)
	})

	t.Run("combat not found", func(t *testing.T) {
		combat, err := service.GetCombat(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, combat)
	})
}

func TestCombatService_NextTurn(t *testing.T) {
	service := NewCombatService()
	ctx := context.Background()

	t.Run("advance turn", func(t *testing.T) {
		combatants := []models.Combatant{
			{ID: "char-1", Name: "Fighter", Initiative: 15, HP: 30},
			{ID: "char-2", Name: "Wizard", Initiative: 12, HP: 20},
		}

		combat, _ := service.StartCombat(ctx, "session-123", combatants)

		nextCombatant, err := service.NextTurn(ctx, combat.ID)
		require.NoError(t, err)
		assert.NotNil(t, nextCombatant)

		// Get updated combat to check turn advanced
		updatedCombat, _ := service.GetCombat(ctx, combat.ID)
		assert.Equal(t, 1, updatedCombat.CurrentTurn) // Advanced from 0 to 1
	})

	t.Run("combat not found", func(t *testing.T) {
		combatant, err := service.NextTurn(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, combatant)
	})
}

func TestCombatService_ApplyDamage(t *testing.T) {
	service := NewCombatService()
	ctx := context.Background()

	t.Run("apply damage", func(t *testing.T) {
		combatants := []models.Combatant{
			{
				ID:    "char-1",
				Name:  "Fighter",
				HP:    45,
				MaxHP: 45,
			},
		}

		combat, _ := service.StartCombat(ctx, "session-123", combatants)

		damage := []models.Damage{
			{Amount: 15, Type: models.DamageTypeSlashing},
		}

		totalDamage, err := service.ApplyDamage(ctx, combat.ID, "char-1", damage)

		require.NoError(t, err)
		assert.Equal(t, 15, totalDamage)

		// Check HP was reduced
		updatedCombat, _ := service.GetCombat(ctx, combat.ID)
		fighter := updatedCombat.Combatants[0]
		assert.Equal(t, 30, fighter.HP)
	})

	t.Run("combatant not found", func(t *testing.T) {
		combatants := []models.Combatant{
			{ID: "char-1", Name: "Fighter", HP: 30},
		}

		combat, _ := service.StartCombat(ctx, "session-456", combatants)

		damage := []models.Damage{{Amount: 10, Type: models.DamageTypeSlashing}}

		totalDamage, err := service.ApplyDamage(ctx, combat.ID, "nonexistent", damage)

		assert.Error(t, err)
		assert.Equal(t, 0, totalDamage)
	})
}

func TestCombatService_HealCombatant(t *testing.T) {
	service := NewCombatService()
	ctx := context.Background()

	t.Run("heal within max HP", func(t *testing.T) {
		combatants := []models.Combatant{
			{
				ID:    "char-1",
				Name:  "Cleric",
				HP:    20,
				MaxHP: 35,
			},
		}

		combat, _ := service.StartCombat(ctx, "session-123", combatants)

		err := service.HealCombatant(ctx, combat.ID, "char-1", 10)

		require.NoError(t, err)

		// Check HP was increased
		updatedCombat, _ := service.GetCombat(ctx, combat.ID)
		cleric := updatedCombat.Combatants[0]
		assert.Equal(t, 30, cleric.HP)
	})

	t.Run("heal beyond max HP", func(t *testing.T) {
		combatants := []models.Combatant{
			{
				ID:    "char-1",
				Name:  "Paladin",
				HP:    40,
				MaxHP: 45,
			},
		}

		combat, _ := service.StartCombat(ctx, "session-123", combatants)

		err := service.HealCombatant(ctx, combat.ID, "char-1", 20)

		require.NoError(t, err)

		// Check HP was capped at max
		updatedCombat, _ := service.GetCombat(ctx, combat.ID)
		paladin := updatedCombat.Combatants[0]
		assert.Equal(t, 45, paladin.HP)
	})
}

func TestCombatService_EndCombat(t *testing.T) {
	service := NewCombatService()
	ctx := context.Background()

	t.Run("successful end", func(t *testing.T) {
		combatants := []models.Combatant{
			{ID: "char-1", Name: "Fighter", Initiative: 15},
		}

		combat, _ := service.StartCombat(ctx, "session-123", combatants)

		err := service.EndCombat(ctx, combat.ID)

		assert.NoError(t, err)
		
		// Verify combat was ended
		_, err = service.GetCombat(ctx, combat.ID)
		assert.Error(t, err)
	})

	t.Run("combat not found", func(t *testing.T) {
		err := service.EndCombat(ctx, "nonexistent")
		assert.Error(t, err)
	})
}