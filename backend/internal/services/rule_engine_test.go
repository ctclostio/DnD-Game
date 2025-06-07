package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-org/dnd-game/internal/models"
	"github.com/your-org/dnd-game/internal/testutil"
)

func TestRuleEngine_EvaluateRule(t *testing.T) {
	t.Run("simple damage modifier rule", func(t *testing.T) {
		mockRepo := new(MockRuleBuilderRepository)
		engine := NewRuleEngine(mockRepo)
		
		rule := &models.CustomRule{
			ID:          1,
			Name:        "Fire Vulnerability",
			Type:        "damage_modifier",
			Trigger:     "on_damage_taken",
			Conditions:  []models.RuleCondition{
				{
					Type:     "damage_type",
					Operator: "equals",
					Value:    "fire",
				},
			},
			Effects: []models.RuleEffect{
				{
					Type:       "multiply_damage",
					Value:      2.0,
					TargetType: "self",
				},
			},
			Priority: 100,
			Active:   true,
		}
		
		context := models.RuleContext{
			TriggerType: "on_damage_taken",
			Actor:       models.RuleActor{ID: "char-1", Type: "character"},
			Target:      models.RuleActor{ID: "char-1", Type: "character"},
			Values: map[string]interface{}{
				"damage":      10,
				"damage_type": "fire",
			},
		}
		
		mockRepo.On("GetActiveRules", "damage_modifier").Return([]*models.CustomRule{rule}, nil)
		
		ctx := testutil.TestContext()
		result, err := engine.EvaluateRule(ctx, "damage_modifier", context)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 20, result.ModifiedValues["damage"]) // 10 * 2
		require.Contains(t, result.AppliedRules, "Fire Vulnerability")
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("conditional AC bonus rule", func(t *testing.T) {
		mockRepo := new(MockRuleBuilderRepository)
		engine := NewRuleEngine(mockRepo)
		
		rule := &models.CustomRule{
			ID:      2,
			Name:    "Shield of Faith Bonus",
			Type:    "ac_modifier",
			Trigger: "on_attack_received",
			Conditions: []models.RuleCondition{
				{
					Type:     "has_condition",
					Operator: "contains",
					Value:    "shield_of_faith",
				},
			},
			Effects: []models.RuleEffect{
				{
					Type:       "add_ac",
					Value:      2,
					TargetType: "self",
				},
			},
			Priority: 50,
			Active:   true,
		}
		
		context := models.RuleContext{
			TriggerType: "on_attack_received",
			Target:      models.RuleActor{
				ID:         "char-1",
				Type:       "character",
				Conditions: []string{"shield_of_faith", "blessed"},
			},
			Values: map[string]interface{}{
				"base_ac": 16,
			},
		}
		
		mockRepo.On("GetActiveRules", "ac_modifier").Return([]*models.CustomRule{rule}, nil)
		
		ctx := testutil.TestContext()
		result, err := engine.EvaluateRule(ctx, "ac_modifier", context)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 18, result.ModifiedValues["ac"]) // 16 + 2
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("complex multi-condition rule", func(t *testing.T) {
		mockRepo := new(MockRuleBuilderRepository)
		engine := NewRuleEngine(mockRepo)
		
		rule := &models.CustomRule{
			ID:      3,
			Name:    "Sneak Attack",
			Type:    "damage_modifier",
			Trigger: "on_attack",
			Conditions: []models.RuleCondition{
				{
					Type:     "weapon_property",
					Operator: "contains",
					Value:    "finesse",
				},
				{
					Type:     "or",
					SubConditions: []models.RuleCondition{
						{
							Type:     "has_advantage",
							Operator: "equals",
							Value:    true,
						},
						{
							Type:     "ally_adjacent_to_target",
							Operator: "equals",
							Value:    true,
						},
					},
				},
			},
			Effects: []models.RuleEffect{
				{
					Type:       "add_dice",
					Value:      "3d6",
					TargetType: "damage",
				},
			},
			Priority: 75,
			Active:   true,
		}
		
		context := models.RuleContext{
			TriggerType: "on_attack",
			Actor:       models.RuleActor{ID: "rogue-1", Type: "character"},
			Target:      models.RuleActor{ID: "enemy-1", Type: "npc"},
			Values: map[string]interface{}{
				"weapon_properties": []string{"finesse", "light"},
				"has_advantage":     true,
				"base_damage":       "1d6+3",
			},
		}
		
		mockRepo.On("GetActiveRules", "damage_modifier").Return([]*models.CustomRule{rule}, nil)
		
		ctx := testutil.TestContext()
		result, err := engine.EvaluateRule(ctx, "damage_modifier", context)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "1d6+3+3d6", result.ModifiedValues["damage_formula"])
		require.Contains(t, result.AppliedRules, "Sneak Attack")
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("no applicable rules", func(t *testing.T) {
		mockRepo := new(MockRuleBuilderRepository)
		engine := NewRuleEngine(mockRepo)
		
		context := models.RuleContext{
			TriggerType: "on_attack",
			Values:      map[string]interface{}{},
		}
		
		mockRepo.On("GetActiveRules", "damage_modifier").Return([]*models.CustomRule{}, nil)
		
		ctx := testutil.TestContext()
		result, err := engine.EvaluateRule(ctx, "damage_modifier", context)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Empty(t, result.AppliedRules)
		require.Empty(t, result.ModifiedValues)
		
		mockRepo.AssertExpectations(t)
	})
}

func TestRuleEngine_ValidateRule(t *testing.T) {
	engine := &RuleEngine{}
	
	tests := []struct {
		name    string
		rule    *models.CustomRule
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid damage modifier rule",
			rule: &models.CustomRule{
				Name:     "Test Rule",
				Type:     "damage_modifier",
				Trigger:  "on_damage_taken",
				Conditions: []models.RuleCondition{
					{
						Type:     "damage_type",
						Operator: "equals",
						Value:    "fire",
					},
				},
				Effects: []models.RuleEffect{
					{
						Type:       "multiply_damage",
						Value:      1.5,
						TargetType: "self",
					},
				},
				Priority: 50,
			},
			wantErr: false,
		},
		{
			name: "invalid rule type",
			rule: &models.CustomRule{
				Name:    "Invalid Rule",
				Type:    "invalid_type",
				Trigger: "on_attack",
			},
			wantErr: true,
			errMsg:  "invalid rule type",
		},
		{
			name: "missing conditions",
			rule: &models.CustomRule{
				Name:       "No Conditions",
				Type:       "damage_modifier",
				Trigger:    "on_attack",
				Conditions: []models.RuleCondition{},
				Effects: []models.RuleEffect{
					{Type: "add_damage", Value: 5},
				},
			},
			wantErr: true,
			errMsg:  "at least one condition required",
		},
		{
			name: "missing effects",
			rule: &models.CustomRule{
				Name:    "No Effects",
				Type:    "damage_modifier",
				Trigger: "on_attack",
				Conditions: []models.RuleCondition{
					{Type: "always", Operator: "equals", Value: true},
				},
				Effects: []models.RuleEffect{},
			},
			wantErr: true,
			errMsg:  "at least one effect required",
		},
		{
			name: "invalid operator",
			rule: &models.CustomRule{
				Name:    "Bad Operator",
				Type:    "damage_modifier",
				Trigger: "on_attack",
				Conditions: []models.RuleCondition{
					{
						Type:     "level",
						Operator: "invalid_op",
						Value:    5,
					},
				},
				Effects: []models.RuleEffect{
					{Type: "add_damage", Value: 5},
				},
			},
			wantErr: true,
			errMsg:  "invalid operator",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateRule(tt.rule)
			
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRuleEngine_ProcessRuleChain(t *testing.T) {
	t.Run("multiple rules with priority ordering", func(t *testing.T) {
		mockRepo := new(MockRuleBuilderRepository)
		engine := NewRuleEngine(mockRepo)
		
		rules := []*models.CustomRule{
			{
				ID:       1,
				Name:     "Base Damage Bonus",
				Priority: 10, // Lower priority, applies first
				Conditions: []models.RuleCondition{
					{Type: "always", Operator: "equals", Value: true},
				},
				Effects: []models.RuleEffect{
					{Type: "add_damage", Value: 2},
				},
			},
			{
				ID:       2,
				Name:     "Critical Hit Multiplier",
				Priority: 90, // Higher priority, applies last
				Conditions: []models.RuleCondition{
					{Type: "is_critical", Operator: "equals", Value: true},
				},
				Effects: []models.RuleEffect{
					{Type: "multiply_damage", Value: 2.0},
				},
			},
		}
		
		context := models.RuleContext{
			TriggerType: "on_attack",
			Values: map[string]interface{}{
				"damage":      10,
				"is_critical": true,
			},
		}
		
		mockRepo.On("GetActiveRules", "damage_modifier").Return(rules, nil)
		
		ctx := testutil.TestContext()
		result, err := engine.EvaluateRule(ctx, "damage_modifier", context)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		// Base damage 10 + 2 = 12, then * 2 = 24
		require.Equal(t, 24, result.ModifiedValues["damage"])
		require.Len(t, result.AppliedRules, 2)
		require.Equal(t, "Base Damage Bonus", result.AppliedRules[0])
		require.Equal(t, "Critical Hit Multiplier", result.AppliedRules[1])
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("conflicting rules resolution", func(t *testing.T) {
		mockRepo := new(MockRuleBuilderRepository)
		engine := NewRuleEngine(mockRepo)
		
		rules := []*models.CustomRule{
			{
				ID:       1,
				Name:     "Fire Resistance",
				Priority: 50,
				Conditions: []models.RuleCondition{
					{Type: "damage_type", Operator: "equals", Value: "fire"},
				},
				Effects: []models.RuleEffect{
					{Type: "multiply_damage", Value: 0.5},
				},
			},
			{
				ID:       2,
				Name:     "Elemental Vulnerability",
				Priority: 60, // Higher priority overrides
				Conditions: []models.RuleCondition{
					{Type: "has_condition", Operator: "contains", Value: "elemental_curse"},
					{Type: "damage_type", Operator: "in", Value: []string{"fire", "cold", "lightning"}},
				},
				Effects: []models.RuleEffect{
					{Type: "multiply_damage", Value: 2.0},
				},
			},
		}
		
		context := models.RuleContext{
			TriggerType: "on_damage_taken",
			Target: models.RuleActor{
				Conditions: []string{"elemental_curse"},
			},
			Values: map[string]interface{}{
				"damage":      20,
				"damage_type": "fire",
			},
		}
		
		mockRepo.On("GetActiveRules", "damage_modifier").Return(rules, nil)
		
		ctx := testutil.TestContext()
		result, err := engine.EvaluateRule(ctx, "damage_modifier", context)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		// Both rules apply: 20 * 0.5 * 2.0 = 20 (they cancel out)
		require.Equal(t, 20, result.ModifiedValues["damage"])
		require.Len(t, result.AppliedRules, 2)
		
		mockRepo.AssertExpectations(t)
	})
}

func TestRuleEngine_EvaluateCondition(t *testing.T) {
	engine := &RuleEngine{}
	
	tests := []struct {
		name      string
		condition models.RuleCondition
		context   models.RuleContext
		expected  bool
	}{
		{
			name: "simple equals condition",
			condition: models.RuleCondition{
				Type:     "level",
				Operator: "equals",
				Value:    5,
			},
			context: models.RuleContext{
				Actor: models.RuleActor{Level: 5},
			},
			expected: true,
		},
		{
			name: "greater than condition",
			condition: models.RuleCondition{
				Type:     "hp_percentage",
				Operator: "less_than",
				Value:    0.5,
			},
			context: models.RuleContext{
				Values: map[string]interface{}{
					"hp_percentage": 0.3,
				},
			},
			expected: true,
		},
		{
			name: "contains condition for array",
			condition: models.RuleCondition{
				Type:     "has_condition",
				Operator: "contains",
				Value:    "poisoned",
			},
			context: models.RuleContext{
				Actor: models.RuleActor{
					Conditions: []string{"poisoned", "exhausted"},
				},
			},
			expected: true,
		},
		{
			name: "in condition",
			condition: models.RuleCondition{
				Type:     "class",
				Operator: "in",
				Value:    []string{"Fighter", "Barbarian", "Paladin"},
			},
			context: models.RuleContext{
				Actor: models.RuleActor{Class: "Fighter"},
			},
			expected: true,
		},
		{
			name: "complex OR condition",
			condition: models.RuleCondition{
				Type: "or",
				SubConditions: []models.RuleCondition{
					{Type: "level", Operator: "greater_than", Value: 10},
					{Type: "has_feat", Operator: "contains", Value: "expert"},
				},
			},
			context: models.RuleContext{
				Actor: models.RuleActor{
					Level: 8,
					Feats: []string{"expert", "lucky"},
				},
			},
			expected: true, // Second condition is true
		},
		{
			name: "complex AND condition",
			condition: models.RuleCondition{
				Type: "and",
				SubConditions: []models.RuleCondition{
					{Type: "level", Operator: "greater_than_or_equal", Value: 5},
					{Type: "class", Operator: "equals", Value: "Rogue"},
				},
			},
			context: models.RuleContext{
				Actor: models.RuleActor{
					Level: 6,
					Class: "Rogue",
				},
			},
			expected: true, // Both conditions are true
		},
		{
			name: "failed AND condition",
			condition: models.RuleCondition{
				Type: "and",
				SubConditions: []models.RuleCondition{
					{Type: "level", Operator: "greater_than", Value: 5},
					{Type: "class", Operator: "equals", Value: "Wizard"},
				},
			},
			context: models.RuleContext{
				Actor: models.RuleActor{
					Level: 3, // Fails first condition
					Class: "Wizard",
				},
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.evaluateCondition(tt.condition, tt.context)
			require.Equal(t, tt.expected, result)
		})
	}
}

// Mock repository for rule builder
type MockRuleBuilderRepository struct {
	mock.Mock
}

func (m *MockRuleBuilderRepository) CreateRule(rule *models.CustomRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockRuleBuilderRepository) GetRule(id int64) (*models.CustomRule, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CustomRule), args.Error(1)
}

func (m *MockRuleBuilderRepository) GetActiveRules(ruleType string) ([]*models.CustomRule, error) {
	args := m.Called(ruleType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CustomRule), args.Error(1)
}

func (m *MockRuleBuilderRepository) UpdateRule(rule *models.CustomRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockRuleBuilderRepository) DeleteRule(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}