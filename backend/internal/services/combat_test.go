package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/your-username/dnd-game/backend/internal/models"
)

func TestCombatService_StartCombat(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		sessionID     string
		participants  []models.Combatant
		expectedError string
		validate      func(*testing.T, *models.Combat)
	}{
		{
			name:      "successful combat start",
			sessionID: "session-123",
			participants: []models.Combatant{
				{
					ID:         "char-1",
					Name:       "Aragorn",
					Type:       models.CombatantTypeCharacter,
					Initiative: 0, // Will be rolled
					HP:         25,
					MaxHP:      25,
					AC:         16,
				},
				{
					ID:         "char-2",
					Name:       "Legolas",
					Type:       models.CombatantTypeCharacter,
					Initiative: 0,
					HP:         20,
					MaxHP:      20,
					AC:         17,
				},
				{
					ID:         "npc-1",
					Name:       "Goblin",
					Type:       models.CombatantTypeNPC,
					Initiative: 0,
					HP:         7,
					MaxHP:      7,
					AC:         13,
				},
			},
			validate: func(t *testing.T, combat *models.Combat) {
				assert.NotEmpty(t, combat.ID)
				assert.Equal(t, "session-123", combat.SessionID)
				assert.Equal(t, models.CombatStatusActive, combat.Status)
				assert.Len(t, combat.Combatants, 3)
				assert.Equal(t, 1, combat.Round)
				assert.Equal(t, 0, combat.Turn)
				
				// Verify initiative was rolled for all combatants
				for _, combatant := range combat.Combatants {
					assert.Greater(t, combatant.Initiative, 0)
				}
				
				// Verify turn order is sorted by initiative
				for i := 1; i < len(combat.TurnOrder); i++ {
					var prevInit, currInit int
					// Find combatants by ID
					for _, c := range combat.Combatants {
						if c.ID == combat.TurnOrder[i-1] {
							prevInit = c.Initiative
						}
						if c.ID == combat.TurnOrder[i] {
							currInit = c.Initiative
						}
					}
					assert.GreaterOrEqual(t, prevInit, currInit)
				}
			},
		},
		{
			name:      "empty participants",
			sessionID: "session-123",
			participants: []models.Combatant{},
			expectedError: "at least two combatants are required",
		},
		{
			name:      "single participant",
			sessionID: "session-123",
			participants: []models.Combatant{
				{ID: "char-1", Name: "Solo"},
			},
			expectedError: "at least two combatants are required",
		},
		{
			name:      "missing session ID",
			sessionID: "",
			participants: []models.Combatant{
				{ID: "char-1", Name: "Fighter", HP: 10, MaxHP: 10},
				{ID: "char-2", Name: "Mage", HP: 10, MaxHP: 10},
			},
			expectedError: "session ID is required",
		},
		{
			name:      "invalid combatant - missing ID",
			sessionID: "session-123",
			participants: []models.Combatant{
				{ID: "", Name: "No ID", HP: 10, MaxHP: 10},
				{ID: "char-2", Name: "Valid", HP: 10, MaxHP: 10},
			},
			expectedError: "combatant ID is required",
		},
		{
			name:      "invalid combatant - missing name",
			sessionID: "session-123",
			participants: []models.Combatant{
				{ID: "char-1", Name: "", HP: 10, MaxHP: 10},
				{ID: "char-2", Name: "Valid", HP: 10, MaxHP: 10},
			},
			expectedError: "combatant name is required",
		},
		{
			name:      "invalid combatant - zero HP",
			sessionID: "session-123",
			participants: []models.Combatant{
				{ID: "char-1", Name: "Dead Guy", HP: 0, MaxHP: 10},
				{ID: "char-2", Name: "Alive Guy", HP: 10, MaxHP: 10},
			},
			expectedError: "combatant must have positive HP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCombatService(nil, nil, nil)
			combat, err := service.StartCombat(ctx, tt.sessionID, tt.participants)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, combat)
			} else {
				require.NoError(t, err)
				require.NotNil(t, combat)
				if tt.validate != nil {
					tt.validate(t, combat)
				}
			}
		})
	}
}

func TestCombatService_GetCombatState(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		combatID      string
		expectedError string
		validate      func(*testing.T, *models.Combat)
	}{
		{
			name:     "get existing combat",
			combatID: "combat-123",
			validate: func(t *testing.T, combat *models.Combat) {
				assert.Equal(t, "combat-123", combat.ID)
				assert.NotNil(t, combat.Combatants)
				assert.NotNil(t, combat.TurnOrder)
			},
		},
		{
			name:          "combat not found",
			combatID:      "nonexistent",
			expectedError: "combat not found",
		},
		{
			name:          "empty combat ID",
			combatID:      "",
			expectedError: "combat ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCombatService(nil, nil, nil)
			
			// Pre-populate combat for valid test case
			if tt.name == "get existing combat" {
				testCombat := &models.Combat{
					ID:        "combat-123",
					SessionID: "session-123",
					Status:    models.CombatStatusActive,
					Combatants: []models.Combatant{
						{ID: "char-1", Name: "Fighter"},
					},
					TurnOrder: []string{"char-1"},
				}
				// In real implementation, this would be stored in a repository
				service.SetCombatState(testCombat)
			}

			combat, err := service.GetCombatState(ctx, tt.combatID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, combat)
			} else {
				require.NoError(t, err)
				require.NotNil(t, combat)
				if tt.validate != nil {
					tt.validate(t, combat)
				}
			}
		})
	}
}

func TestCombatService_ExecuteAction(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		combatID      string
		action        models.CombatAction
		setupCombat   func(*CombatService)
		expectedError string
		validate      func(*testing.T, *models.Combat)
	}{
		{
			name:     "successful attack action",
			combatID: "combat-123",
			action: models.CombatAction{
				ActionType:  models.ActionTypeAttack,
				ActorID:    "char-1",
				TargetID:   "npc-1",
				WeaponName: "Longsword",
				AttackBonus: 5,
				DamageDice: "1d8",
				DamageBonus: 3,
				DamageType: "slashing",
			},
			setupCombat: func(service *CombatService) {
				// Setup existing combat state
				combat := &models.Combat{
					ID:        "combat-123",
					SessionID: "session-123",
					Status:    models.CombatStatusActive,
					Round:     1,
					CurrentTurn: 0,
					Combatants: []models.Combatant{
						{
							ID:    "char-1",
							Name:  "Fighter",
							HP:    25,
							MaxHP: 25,
							AC:    16,
						},
						{
							ID:    "npc-1",
							Name:  "Goblin",
							HP:    7,
							MaxHP: 7,
							AC:    13,
						},
					},
					TurnOrder: []string{"char-1", "npc-1"},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Verify action was recorded
				assert.NotEmpty(t, combat.ActionHistory)
				lastAction := combat.ActionHistory[len(combat.ActionHistory)-1]
				assert.Equal(t, models.ActionTypeAttack, lastAction.ActionType)
				assert.Equal(t, "char-1", lastAction.ActorID)
				assert.Equal(t, "npc-1", lastAction.TargetID)
				
				// Verify damage was applied if hit
				if lastAction.Hit {
					// Find goblin in combatants slice
					var goblin *models.Combatant
					for i := range combat.Combatants {
						if combat.Combatants[i].ID == "npc-1" {
							goblin = &combat.Combatants[i]
							break
						}
					}
					require.NotNil(t, goblin)
					assert.Less(t, goblin.HP, 7)
				}
			},
		},
		{
			name:     "cast spell action",
			combatID: "combat-123",
			action: models.CombatAction{
				ActionType: models.ActionTypeCastSpell,
				ActorID:   "char-2",
				TargetID:  "npc-1",
				SpellName: "Fire Bolt",
				SpellLevel: 0,
				SpellDC:   14,
				SpellAttackBonus: 6,
				SpellDamage: "1d10",
				SpellDamageType: "fire",
			},
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID:        "combat-123",
					SessionID: "session-123",
					Status:    models.CombatStatusActive,
					Round:     1,
					CurrentTurn: 1,
					Combatants: []models.Combatant{
						{
							ID:    "char-2",
							Name:  "Wizard",
							HP:    15,
							MaxHP: 15,
							AC:    12,
						},
						{
							ID:    "npc-1",
							Name:  "Goblin",
							HP:    7,
							MaxHP: 7,
							AC:    13,
						},
					},
					TurnOrder: []string{"npc-1", "char-2"},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				assert.NotEmpty(t, combat.ActionHistory)
				lastAction := combat.ActionHistory[len(combat.ActionHistory)-1]
				assert.Equal(t, models.ActionTypeCastSpell, lastAction.ActionType)
				assert.Equal(t, "Fire Bolt", lastAction.SpellName)
			},
		},
		{
			name:     "movement action",
			combatID: "combat-123",
			action: models.CombatAction{
				ActionType:  models.ActionTypeMove,
				ActorID:     "char-1",
				Movement:    30,
				NewPosition: models.Position{X: 5, Y: 10},
			},
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID:        "combat-123",
					SessionID: "session-123",
					Status:    models.CombatStatusActive,
					Combatants: []models.Combatant{
						{
							ID:       "char-1",
							Name:     "Fighter",
							Position: models.Position{X: 0, Y: 0},
						},
					},
					TurnOrder: []string{"char-1"},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Find fighter in combatants slice
				var fighter *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						fighter = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, fighter)
				assert.Equal(t, 5, fighter.Position.X)
				assert.Equal(t, 10, fighter.Position.Y)
			},
		},
		{
			name:     "dodge action",
			combatID: "combat-123",
			action: models.CombatAction{
				ActionType: models.ActionTypeDodge,
				ActorID: "char-1",
			},
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID:        "combat-123",
					SessionID: "session-123",
					Status:    models.CombatStatusActive,
					Combatants: []models.Combatant{
						{ID: "char-1", Name: "Fighter"},
					},
					TurnOrder: []string{"char-1"},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Find fighter in combatants slice
				var fighter *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						fighter = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, fighter)
				assert.Contains(t, fighter.Conditions, models.ConditionDodging)
			},
		},
		{
			name:     "end turn action",
			combatID: "combat-123",
			action: models.CombatAction{
				ActionType: models.ActionTypeEndTurn,
				ActorID: "char-1",
			},
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID:        "combat-123",
					SessionID: "session-123",
					Status:    models.CombatStatusActive,
					Round:     1,
					CurrentTurn: 0,
					Combatants: []models.Combatant{
						{ID: "char-1", Name: "Fighter"},
						{ID: "char-2", Name: "Wizard"},
					},
					TurnOrder: []string{"char-1", "char-2"},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Turn should advance
				assert.Equal(t, 1, combat.Turn)
				// Still round 1 since we haven't completed all turns
				assert.Equal(t, 1, combat.Round)
			},
		},
		{
			name:          "combat not found",
			combatID:      "nonexistent",
			action:        models.CombatAction{ActionType: models.ActionTypeAttack},
			expectedError: "combat not found",
		},
		{
			name:     "not actor's turn",
			combatID: "combat-123",
			action: models.CombatAction{
				ActionType: models.ActionTypeAttack,
				ActorID: "char-2", // Wrong character
			},
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID:        "combat-123",
					SessionID: "session-123",
					Status:    models.CombatStatusActive,
					CurrentTurn: 0,
					Combatants: []models.Combatant{
						{ID: "char-1", Name: "Fighter"},
						{ID: "char-2", Name: "Wizard"},
					},
					TurnOrder: []string{"char-1", "char-2"}, // char-1's turn
				}
				service.SetCombatState(combat)
			},
			expectedError: "not this combatant's turn",
		},
		{
			name:     "combat already ended",
			combatID: "combat-123",
			action: models.CombatAction{
				ActionType: models.ActionTypeAttack,
				ActorID: "char-1",
			},
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID:     "combat-123",
					Status: models.CombatStatusCompleted,
				}
				service.SetCombatState(combat)
			},
			expectedError: "combat has already ended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCombatService(nil, nil, nil)
			
			if tt.setupCombat != nil {
				tt.setupCombat(service)
			}

			combat, err := service.ExecuteAction(ctx, tt.combatID, tt.action)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, combat)
				if tt.validate != nil {
					tt.validate(t, combat)
				}
			}
		})
	}
}

func TestCombatService_ApplyDamage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		combatID      string
		targetID      string
		damage        int
		damageType    string
		setupCombat   func(*CombatService)
		expectedError string
		validate      func(*testing.T, *models.Combat)
	}{
		{
			name:       "normal damage application",
			combatID:   "combat-123",
			targetID:   "char-1",
			damage:     8,
			damageType: "slashing",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:    "char-1",
							Name:  "Fighter",
							HP:    25,
							MaxHP: 25,
						},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Find fighter in combatants slice
				var fighter *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						fighter = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, fighter)
				assert.Equal(t, 17, fighter.HP) // 25 - 8
			},
		},
		{
			name:       "damage reduces HP to zero",
			combatID:   "combat-123",
			targetID:   "char-1",
			damage:     30,
			damageType: "fire",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:    "char-1",
							Name:  "Fighter",
							HP:    25,
							MaxHP: 25,
						},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Find fighter in combatants slice
				var fighter *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						fighter = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, fighter)
				assert.Equal(t, 0, fighter.HP)
				assert.Contains(t, fighter.Conditions, models.ConditionUnconscious)
			},
		},
		{
			name:       "damage with resistance",
			combatID:   "combat-123",
			targetID:   "char-1",
			damage:     10,
			damageType: "fire",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:               "char-1",
							Name:             "Tiefling",
							HP:               20,
							MaxHP:            20,
							DamageResistances: []string{"fire"},
						},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Find tiefling in combatants slice
				var tiefling *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						tiefling = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, tiefling)
				assert.Equal(t, 15, tiefling.HP) // 20 - (10/2)
			},
		},
		{
			name:       "damage with immunity",
			combatID:   "combat-123",
			targetID:   "char-1",
			damage:     15,
			damageType: "poison",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:              "char-1",
							Name:            "Construct",
							HP:              30,
							MaxHP:           30,
							DamageImmunities: []string{"poison"},
						},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Find construct in combatants slice
				var construct *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						construct = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, construct)
				assert.Equal(t, 30, construct.HP) // No damage
			},
		},
		{
			name:       "damage with vulnerability",
			combatID:   "combat-123",
			targetID:   "char-1",
			damage:     6,
			damageType: "radiant",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:                   "char-1",
							Name:                 "Shadow",
							HP:                   15,
							MaxHP:                15,
							DamageVulnerabilities: []string{"radiant"},
						},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Find shadow in combatants slice
				var shadow *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						shadow = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, shadow)
				assert.Equal(t, 3, shadow.HP) // 15 - (6*2)
			},
		},
		{
			name:          "target not found",
			combatID:      "combat-123",
			targetID:      "nonexistent",
			damage:        10,
			damageType:    "slashing",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID:         "combat-123",
					Combatants: []models.Combatant{},
				}
				service.SetCombatState(combat)
			},
			expectedError: "target not found in combat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCombatService(nil, nil, nil)
			
			if tt.setupCombat != nil {
				tt.setupCombat(service)
			}

			combat, err := service.ApplyDamage(ctx, tt.combatID, tt.targetID, tt.damage, tt.damageType)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, combat)
				if tt.validate != nil {
					tt.validate(t, combat)
				}
			}
		})
	}
}

func TestCombatService_ApplyHealing(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		combatID      string
		targetID      string
		healing       int
		setupCombat   func(*CombatService)
		expectedError string
		validate      func(*testing.T, *models.Combat)
	}{
		{
			name:     "normal healing",
			combatID: "combat-123",
			targetID: "char-1",
			healing:  10,
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:    "char-1",
							Name:  "Fighter",
							HP:    15,
							MaxHP: 25,
						},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Find fighter in combatants slice
				var fighter *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						fighter = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, fighter)
				assert.Equal(t, 25, fighter.HP) // Healed to max
			},
		},
		{
			name:     "healing unconscious character",
			combatID: "combat-123",
			targetID: "char-1",
			healing:  5,
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:         "char-1",
							Name:       "Fighter",
							HP:         0,
							MaxHP:      25,
							Conditions: []models.Condition{models.ConditionUnconscious},
						},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Find fighter in combatants slice
				var fighter *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						fighter = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, fighter)
				assert.Equal(t, 5, fighter.HP)
				assert.NotContains(t, fighter.Conditions, models.ConditionUnconscious)
			},
		},
		{
			name:     "healing at max HP",
			combatID: "combat-123",
			targetID: "char-1",
			healing:  10,
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:    "char-1",
							Name:  "Fighter",
							HP:    25,
							MaxHP: 25,
						},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat) {
				// Find fighter in combatants slice
				var fighter *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						fighter = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, fighter)
				assert.Equal(t, 25, fighter.HP) // Still at max
			},
		},
		{
			name:          "invalid healing amount",
			combatID:      "combat-123",
			targetID:      "char-1",
			healing:       -5,
			expectedError: "healing must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCombatService(nil, nil, nil)
			
			if tt.setupCombat != nil {
				tt.setupCombat(service)
			}

			combat, err := service.ApplyHealing(ctx, tt.combatID, tt.targetID, tt.healing)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, combat)
				if tt.validate != nil {
					tt.validate(t, combat)
				}
			}
		})
	}
}

func TestCombatService_EndCombat(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		combatID      string
		setupCombat   func(*CombatService)
		expectedError string
		validate      func(*testing.T, error)
	}{
		{
			name:     "successful combat end",
			combatID: "combat-123",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID:        "combat-123",
					SessionID: "session-123",
					Status:    models.CombatStatusActive,
					Combatants: []models.Combatant{
						{ID: "char-1", Name: "Fighter", HP: 25},
						{ID: "npc-1", Name: "Goblin", HP: 0},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:          "combat not found",
			combatID:      "nonexistent",
			expectedError: "combat not found",
		},
		{
			name:     "combat already ended",
			combatID: "combat-123",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID:     "combat-123",
					Status: models.CombatStatusCompleted,
				}
				service.SetCombatState(combat)
			},
			expectedError: "combat has already ended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCombatService(nil, nil, nil)
			
			if tt.setupCombat != nil {
				tt.setupCombat(service)
			}

			err := service.EndCombat(ctx, tt.combatID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				if tt.validate != nil {
					tt.validate(t, err)
				}
			}
		})
	}
}

func TestCombatService_DeathSavingThrow(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		combatID      string
		characterID   string
		setupCombat   func(*CombatService)
		expectedError string
		validate      func(*testing.T, *models.Combat, *models.DeathSaveResult)
	}{
		{
			name:        "successful death save",
			combatID:    "combat-123",
			characterID: "char-1",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:               "char-1",
							Name:             "Fighter",
							HP:               0,
							MaxHP:            25,
							Conditions:       []models.Condition{models.ConditionUnconscious},
							DeathSaveSuccesses: 0,
							DeathSaveFailures:  0,
						},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat, result *models.DeathSaveResult) {
				assert.NotNil(t, result)
				// Find fighter in combatants slice
				var fighter *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						fighter = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, fighter)
				
				if result.Roll >= 10 {
					assert.Greater(t, fighter.DeathSaveSuccesses, 0)
				} else {
					assert.Greater(t, fighter.DeathSaveFailures, 0)
				}
				
				// Check for critical results
				if result.Roll == 20 {
					assert.Equal(t, 1, fighter.HP)
					assert.NotContains(t, fighter.Conditions, models.ConditionUnconscious)
				} else if result.Roll == 1 {
					assert.Equal(t, 2, fighter.DeathSaveFailures)
				}
			},
		},
		{
			name:        "third success stabilizes",
			combatID:    "combat-123",
			characterID: "char-1",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:                 "char-1",
							Name:               "Fighter",
							HP:                 0,
							MaxHP:              25,
							Conditions:         []models.Condition{models.ConditionUnconscious},
							DeathSaveSuccesses: 2,
							DeathSaveFailures:  1,
						},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat, result *models.DeathSaveResult) {
				// Find fighter in combatants slice
				var fighter *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						fighter = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, fighter)
				
				if result.Roll >= 10 {
					assert.Equal(t, 0, fighter.DeathSaveSuccesses)
					assert.Equal(t, 0, fighter.DeathSaveFailures)
					assert.Contains(t, fighter.Conditions, models.ConditionStable)
				}
			},
		},
		{
			name:        "third failure causes death",
			combatID:    "combat-123",
			characterID: "char-1",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:                 "char-1",
							Name:               "Fighter",
							HP:                 0,
							MaxHP:              25,
							Conditions:         []models.Condition{models.ConditionUnconscious},
							DeathSaveSuccesses: 1,
							DeathSaveFailures:  2,
						},
					},
				}
				service.SetCombatState(combat)
			},
			validate: func(t *testing.T, combat *models.Combat, result *models.DeathSaveResult) {
				// Find fighter in combatants slice
				var fighter *models.Combatant
				for i := range combat.Combatants {
					if combat.Combatants[i].ID == "char-1" {
						fighter = &combat.Combatants[i]
						break
					}
				}
				require.NotNil(t, fighter)
				
				if result.Roll < 10 {
					assert.Contains(t, fighter.Conditions, models.ConditionDead)
				}
			},
		},
		{
			name:          "character not at 0 HP",
			combatID:      "combat-123",
			characterID:   "char-1",
			setupCombat: func(service *CombatService) {
				combat := &models.Combat{
					ID: "combat-123",
					Combatants: []models.Combatant{
						{
							ID:    "char-1",
							Name:  "Fighter",
							HP:    10,
							MaxHP: 25,
						},
					},
				}
				service.SetCombatState(combat)
			},
			expectedError: "character is not unconscious",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCombatService(nil, nil, nil)
			
			if tt.setupCombat != nil {
				tt.setupCombat(service)
			}

			combat, result, err := service.DeathSavingThrow(ctx, tt.combatID, tt.characterID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, combat)
				if tt.validate != nil {
					tt.validate(t, combat, result)
				}
			}
		})
	}
}