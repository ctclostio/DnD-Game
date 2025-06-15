package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// ConditionalRealitySystem manages context-aware rule modifications
type ConditionalRealitySystem struct {
	contextManager *ContextManager
	ruleEngine     *RuleEngine
	modifiers      map[string][]RuleModifier
}

// ContextManager tracks active contexts in the game world
type ContextManager struct {
	activeContexts map[string][]models.ConditionalContext
	subscribers    map[string][]ContextSubscriber
}

// ContextSubscriber interface for entities that respond to context changes
type ContextSubscriber interface {
	OnContextChange(ctx context.Context, contextType string, newContext models.ConditionalContext)
}

// RuleModifier defines how a rule changes under specific conditions
type RuleModifier struct {
	ConditionType string
	Conditions    []models.RuleCondition
	Modifications ModificationSet
}

// ModificationSet contains all modifications to apply
type ModificationSet struct {
	NodeOverrides      map[string]map[string]interface{} // nodeID -> property overrides
	ParameterOverrides map[string]interface{}
	DisabledNodes      []string
	AdditionalNodes    []models.LogicNode
	Description        string
}

// NewConditionalRealitySystem creates a new conditional reality system
func NewConditionalRealitySystem(ruleEngine *RuleEngine) *ConditionalRealitySystem {
	return &ConditionalRealitySystem{
		contextManager: &ContextManager{
			activeContexts: make(map[string][]models.ConditionalContext),
			subscribers:    make(map[string][]ContextSubscriber),
		},
		ruleEngine: ruleEngine,
		modifiers:  make(map[string][]RuleModifier),
	}
}

// RegisterContext adds a new active context to the system
func (crs *ConditionalRealitySystem) RegisterContext(ctx context.Context, sessionID, contextType string, contextValue interface{}) error {
	context := models.ConditionalContext{
		ID:          uuid.New().String(),
		SessionID:   sessionID,
		ContextType: contextType,
		ContextValue: map[string]interface{}{
			"value": contextValue,
		},
		IsActive:  true,
		StartedAt: time.Now(),
	}

	// Add to active contexts
	if crs.contextManager.activeContexts[sessionID] == nil {
		crs.contextManager.activeContexts[sessionID] = []models.ConditionalContext{}
	}
	crs.contextManager.activeContexts[sessionID] = append(crs.contextManager.activeContexts[sessionID], context)

	// Notify subscribers
	crs.contextManager.notifySubscribers(ctx, sessionID, contextType, context)

	return nil
}

// GetActiveContexts returns all active contexts for a session
func (crs *ConditionalRealitySystem) GetActiveContexts(sessionID string) []models.ConditionalContext {
	return crs.contextManager.activeContexts[sessionID]
}

// ApplyConditionalRules modifies a rule based on active contexts
func (crs *ConditionalRealitySystem) ApplyConditionalRules(
	template *models.RuleTemplate,
	instance *models.RuleInstance,
	activeContexts []models.ConditionalContext,
) (*models.RuleTemplate, error) {
	// Create a copy of the template to modify
	modifiedTemplate := *template

	// Check each conditional rule in the template
	for _, conditionalRule := range template.ConditionalRules {
		if crs.conditionsMet(conditionalRule.Conditions, activeContexts, instance) {
			// Apply modifications
			if conditionalRule.ModifiedLogic != nil {
				modifiedTemplate.LogicGraph = *conditionalRule.ModifiedLogic
			}

			// Apply parameter overrides
			for param, value := range conditionalRule.ParameterOverrides {
				instance.ParameterValues[param] = value
			}
		}
	}

	// Apply global modifiers based on context types
	for _, context := range activeContexts {
		modifiers := crs.getModifiersForContext(context.ContextType, context.ContextValue)
		for _, modifier := range modifiers {
			crs.applyModifier(&modifiedTemplate, modifier)
		}
	}

	return &modifiedTemplate, nil
}

// conditionsMet checks if all conditions are satisfied
func (crs *ConditionalRealitySystem) conditionsMet(
	conditions []models.RuleCondition,
	activeContexts []models.ConditionalContext,
	instance *models.RuleInstance,
) bool {
	for _, condition := range conditions {
		if !crs.evaluateCondition(condition, activeContexts, instance) {
			return false
		}
	}
	return true
}

// evaluateCondition checks a single condition
func (crs *ConditionalRealitySystem) evaluateCondition(
	condition models.RuleCondition,
	activeContexts []models.ConditionalContext,
	instance *models.RuleInstance,
) bool {
	switch condition.Type {
	case "location":
		return crs.evaluateLocationCondition(condition, activeContexts)
	case "character_state":
		return crs.evaluateCharacterStateCondition(condition, instance)
	case "time":
		return crs.evaluateTimeCondition(condition)
	case "narrative":
		return crs.evaluateNarrativeCondition(condition, activeContexts)
	case "plane":
		return crs.evaluatePlaneCondition(condition, activeContexts)
	case "emotion":
		return crs.evaluateEmotionCondition(condition, instance)
	case "environment":
		return crs.evaluateEnvironmentCondition(condition, activeContexts)
	default:
		return false
	}
}

// Specific condition evaluators

func (crs *ConditionalRealitySystem) evaluateLocationCondition(condition models.RuleCondition, contexts []models.ConditionalContext) bool {
	for _, ctx := range contexts {
		if ctx.ContextType == models.ConditionTypeEnvironment {
			location, ok := ctx.ContextValue["location"].(string)
			if !ok {
				continue
			}

			switch condition.Operator {
			case constants.OperatorEquals:
				return location == condition.Value.(string)
			case "contains":
				return strings.Contains(location, condition.Value.(string))
			case "in":
				locations, ok := condition.Value.([]string)
				if !ok {
					return false
				}
				for _, loc := range locations {
					if location == loc {
						return true
					}
				}
			}
		}
	}
	return false
}

func (crs *ConditionalRealitySystem) evaluatePlaneCondition(condition models.RuleCondition, contexts []models.ConditionalContext) bool {
	for _, ctx := range contexts {
		if ctx.ContextType == models.ConditionTypePlane {
			plane, ok := ctx.ContextValue["plane"].(string)
			if !ok {
				continue
			}

			expectedPlane, ok := condition.Value.(string)
			if !ok {
				return false
			}

			return plane == expectedPlane
		}
	}
	return false
}

func (crs *ConditionalRealitySystem) evaluateEmotionCondition(condition models.RuleCondition, instance *models.RuleInstance) bool {
	// Check character's emotional state
	if emotionalState, ok := instance.State["emotional_state"].(map[string]interface{}); ok {
		emotion, ok := emotionalState["current"].(string)
		if !ok {
			return false
		}

		switch condition.Operator {
		case constants.OperatorEquals:
			return emotion == condition.Value.(string)
		case "intensity_above":
			intensity, _ := emotionalState["intensity"].(float64)
			threshold, _ := condition.Value.(float64)
			return intensity > threshold
		}
	}
	return false
}

func (crs *ConditionalRealitySystem) evaluateCharacterStateCondition(_ models.RuleCondition, instance *models.RuleInstance) bool {
	// This would check character conditions like HP, status effects, etc.
	// Simplified for now
	return true
}

func (crs *ConditionalRealitySystem) evaluateTimeCondition(condition models.RuleCondition) bool {
	// Check time-based conditions
	now := time.Now()

	switch condition.Operator {
	case "hour_of_day":
		expectedHour, ok := condition.Value.(int)
		if !ok {
			return false
		}
		return now.Hour() == expectedHour
	case "time_of_day":
		timeOfDay, ok := condition.Value.(string)
		if !ok {
			return false
		}
		hour := now.Hour()
		switch timeOfDay {
		case "dawn":
			return hour >= 5 && hour < 7
		case "day":
			return hour >= 7 && hour < 18
		case "dusk":
			return hour >= 18 && hour < 20
		case constants.TimeNight:
			return hour >= 20 || hour < 5
		}
	}
	return false
}

func (crs *ConditionalRealitySystem) evaluateNarrativeCondition(condition models.RuleCondition, contexts []models.ConditionalContext) bool {
	// Check narrative conditions like story progress, completed quests, etc.
	for _, ctx := range contexts {
		if ctx.ContextType == models.ConditionTypeNarrative {
			if storyProgress, ok := ctx.ContextValue["story_progress"].(map[string]interface{}); ok {
				// Check various narrative conditions
				switch condition.Operator {
				case "quest_completed":
					questName, _ := condition.Value.(string)
					completed, _ := storyProgress[questName].(bool)
					return completed
				case "chapter_reached":
					currentChapter, _ := storyProgress["current_chapter"].(int)
					requiredChapter, _ := condition.Value.(int)
					return currentChapter >= requiredChapter
				}
			}
		}
	}
	return false
}

func (crs *ConditionalRealitySystem) evaluateEnvironmentCondition(condition models.RuleCondition, contexts []models.ConditionalContext) bool {
	for _, ctx := range contexts {
		if ctx.ContextType == models.ConditionTypeEnvironment || ctx.ContextType == models.ConditionTypeWeather {
			// Check environmental conditions
			switch condition.Operator {
			case "weather_is":
				weather, _ := ctx.ContextValue["weather"].(string)
				return weather == condition.Value.(string)
			case "terrain_type":
				terrain, _ := ctx.ContextValue["terrain"].(string)
				return terrain == condition.Value.(string)
			case "light_level":
				lightLevel, _ := ctx.ContextValue["light_level"].(string)
				return lightLevel == condition.Value.(string)
			}
		}
	}
	return false
}

// getModifiersForContext returns modifiers that apply to a specific context
func (crs *ConditionalRealitySystem) getModifiersForContext(contextType string, contextValue map[string]interface{}) []RuleModifier {
	// Predefined modifiers for different contexts
	modifiers := []RuleModifier{}

	switch contextType {
	case models.ConditionTypePlane:
		plane, _ := contextValue["plane"].(string)
		modifiers = append(modifiers, crs.getPlaneModifiers(plane)...)
	case models.ConditionTypeWeather:
		weather, _ := contextValue["weather"].(string)
		modifiers = append(modifiers, crs.getWeatherModifiers(weather)...)
	case models.ConditionTypeEmotion:
		emotion, _ := contextValue["emotion"].(string)
		intensity, _ := contextValue["intensity"].(float64)
		modifiers = append(modifiers, crs.getEmotionModifiers(emotion, intensity)...)
	}

	return modifiers
}

// Plane-specific modifiers
func (crs *ConditionalRealitySystem) getPlaneModifiers(plane string) []RuleModifier {
	modifiers := []RuleModifier{}

	switch plane {
	case "Feywild":
		// Magic is chaotic in the Feywild
		modifiers = append(modifiers, RuleModifier{
			ConditionType: models.ConditionTypePlane,
			Modifications: ModificationSet{
				NodeOverrides: map[string]map[string]interface{}{
					"*": { // Apply to all damage nodes
						"wild_magic_chance": 0.2,
					},
				},
				Description: "Magic behaves unpredictably in the Feywild",
			},
		})
	case "Shadowfell":
		// Necrotic damage enhanced, radiant weakened
		modifiers = append(modifiers, RuleModifier{
			ConditionType: models.ConditionTypePlane,
			Modifications: ModificationSet{
				NodeOverrides: map[string]map[string]interface{}{
					"damage_nodes": {
						"damage_modifier_necrotic": 1.5,
						"damage_modifier_radiant":  0.5,
					},
				},
				Description: "The Shadowfell empowers death magic",
			},
		})
	case "Elemental Plane of Fire":
		// Fire damage enhanced, cold weakened
		modifiers = append(modifiers, RuleModifier{
			ConditionType: models.ConditionTypePlane,
			Modifications: ModificationSet{
				NodeOverrides: map[string]map[string]interface{}{
					"damage_nodes": {
						"damage_modifier_fire": 2.0,
						"damage_modifier_cold": 0.25,
					},
				},
				Description: "Fire reigns supreme in this plane",
			},
		})
	}

	return modifiers
}

// Weather-specific modifiers
func (crs *ConditionalRealitySystem) getWeatherModifiers(weather string) []RuleModifier {
	modifiers := []RuleModifier{}

	switch weather {
	case "storm":
		// Lightning damage enhanced, ranged attacks hindered
		modifiers = append(modifiers, RuleModifier{
			ConditionType: models.ConditionTypeWeather,
			Modifications: ModificationSet{
				NodeOverrides: map[string]map[string]interface{}{
					"damage_nodes": {
						"damage_modifier_lightning": 1.5,
					},
					"attack_nodes": {
						"ranged_penalty": -2,
					},
				},
				Description: "Storm conditions affect combat",
			},
		})
	case "fog":
		// Visibility reduced, stealth enhanced
		modifiers = append(modifiers, RuleModifier{
			ConditionType: models.ConditionTypeWeather,
			Modifications: ModificationSet{
				ParameterOverrides: map[string]interface{}{
					"visibility_range": 30, // feet
					"stealth_bonus":    5,
				},
				Description: "Heavy fog obscures vision",
			},
		})
	}

	return modifiers
}

// Emotion-specific modifiers
func (crs *ConditionalRealitySystem) getEmotionModifiers(emotion string, intensity float64) []RuleModifier {
	modifiers := []RuleModifier{}

	switch emotion {
	case "rage":
		// Damage increased, defense decreased
		modifiers = append(modifiers, RuleModifier{
			ConditionType: models.ConditionTypeEmotion,
			Modifications: ModificationSet{
				ParameterOverrides: map[string]interface{}{
					"damage_bonus":    intensity * 2,
					"ac_penalty":      intensity,
					"critical_chance": 0.05 * intensity,
				},
				Description: fmt.Sprintf("Rage (intensity %.1f) affects combat", intensity),
			},
		})
	case "fear":
		// Disadvantage on attacks, movement restricted
		modifiers = append(modifiers, RuleModifier{
			ConditionType: models.ConditionTypeEmotion,
			Modifications: ModificationSet{
				ParameterOverrides: map[string]interface{}{
					"attack_disadvantage": true,
					"movement_penalty":    intensity * 10,
				},
				DisabledNodes: []string{"aggressive_action_nodes"},
				Description:   "Fear inhibits aggressive actions",
			},
		})
	case "determination":
		// Bonus to saves, resistance to conditions
		modifiers = append(modifiers, RuleModifier{
			ConditionType: models.ConditionTypeEmotion,
			Modifications: ModificationSet{
				ParameterOverrides: map[string]interface{}{
					"save_bonus":           intensity * 2,
					"condition_resistance": 0.2 * intensity,
				},
				Description: "Determination strengthens resolve",
			},
		})
	}

	return modifiers
}

// applyModifier applies a modifier to a rule template
func (crs *ConditionalRealitySystem) applyModifier(template *models.RuleTemplate, modifier RuleModifier) {
	// Apply node overrides
	for nodePattern, overrides := range modifier.Modifications.NodeOverrides {
		for i := range template.LogicGraph.Nodes {
			node := &template.LogicGraph.Nodes[i]

			// Check if node matches pattern
			if nodePattern == "*" || strings.Contains(node.Type, nodePattern) {
				// Apply property overrides
				if node.Properties == nil {
					node.Properties = make(map[string]interface{})
				}
				for prop, value := range overrides {
					node.Properties[prop] = value
				}
			}
		}
	}

	// Apply parameter overrides
	for param, value := range modifier.Modifications.ParameterOverrides {
		// Find and update parameter
		for i := range template.Parameters {
			if template.Parameters[i].Name == param {
				template.Parameters[i].DefaultValue = value
				break
			}
		}
	}

	// Disable specified nodes
	for _, nodeID := range modifier.Modifications.DisabledNodes {
		for i := range template.LogicGraph.Nodes {
			if template.LogicGraph.Nodes[i].ID == nodeID {
				// Mark node as disabled
				template.LogicGraph.Nodes[i].Properties["disabled"] = true
			}
		}
	}

	// Add additional nodes
	template.LogicGraph.Nodes = append(template.LogicGraph.Nodes, modifier.Modifications.AdditionalNodes...)
}

// Subscribe adds a subscriber for context changes
func (cm *ContextManager) Subscribe(sessionID string, subscriber ContextSubscriber) {
	if cm.subscribers[sessionID] == nil {
		cm.subscribers[sessionID] = []ContextSubscriber{}
	}
	cm.subscribers[sessionID] = append(cm.subscribers[sessionID], subscriber)
}

// notifySubscribers notifies all subscribers of a context change
func (cm *ContextManager) notifySubscribers(ctx context.Context, sessionID string, contextType string, newContext models.ConditionalContext) {
	subscribers := cm.subscribers[sessionID]
	for _, subscriber := range subscribers {
		go subscriber.OnContextChange(ctx, contextType, newContext)
	}
}

// PredefinedContexts contains common game contexts

var PredefinedPlanes = []string{
	"Material Plane",
	"Feywild",
	"Shadowfell",
	"Elemental Plane of Fire",
	"Elemental Plane of Water",
	"Elemental Plane of Earth",
	"Elemental Plane of Air",
	"Ethereal Plane",
	"Astral Plane",
	"Nine Hells",
	"The Abyss",
	"Mechanus",
	"Limbo",
}

var PredefinedWeatherConditions = []string{
	"clear",
	"rain",
	"storm",
	"snow",
	"fog",
	"wind",
	"hail",
	"magical_storm",
}

var PredefinedEmotionalStates = []string{
	"calm",
	"rage",
	"fear",
	"joy",
	"sorrow",
	"determination",
	"confusion",
	"despair",
	"hope",
}
