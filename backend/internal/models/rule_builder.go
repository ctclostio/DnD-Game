package models

import (
	"time"
)

// RuleTemplate represents a reusable rule pattern created through the visual builder
type RuleTemplate struct {
	ID               string                 `json:"id" db:"id"`
	Name             string                 `json:"name" db:"name"`
	Description      string                 `json:"description" db:"description"`
	Category         string                 `json:"category" db:"category"` // spell, ability, item, environmental, condition
	CreatedByID      string                 `json:"created_by_id" db:"created_by"`
	IsPublic         bool                   `json:"is_public" db:"is_public"`
	Version          int                    `json:"version" db:"version"`
	LogicGraph       LogicGraph             `json:"logic_graph" db:"logic_graph"`
	Parameters       []RuleParameter        `json:"parameters" db:"parameters"`
	BalanceMetrics   BalanceMetrics         `json:"balance_metrics" db:"balance_metrics"`
	ConditionalRules []ConditionalRule      `json:"conditional_rules" db:"conditional_rules"`
	Tags             []string               `json:"tags" db:"tags"`
	UsageCount       int                    `json:"usage_count" db:"usage_count"`
	ApprovalStatus   string                 `json:"approval_status" db:"approval_status"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// LogicGraph represents the visual node-based logic structure
type LogicGraph struct {
	Nodes       []LogicNode       `json:"nodes"`
	Connections []NodeConnection  `json:"connections"`
	StartNodeID string            `json:"start_node_id"`
	Variables   map[string]Variable `json:"variables"`
}

// LogicNode represents a single node in the visual logic builder
type LogicNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // trigger, condition, action, effect, calculation, variable
	SubType    string                 `json:"subtype"` // specific node functionality
	Position   EditorPosition         `json:"position"`
	Properties map[string]interface{} `json:"properties"`
	Inputs     []NodePort             `json:"inputs"`
	Outputs    []NodePort             `json:"outputs"`
}

// NodePort represents an input or output connection point on a node
type NodePort struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	DataType string `json:"data_type"` // boolean, number, string, entity, array, any
	Required bool   `json:"required"`
	Multiple bool   `json:"multiple"` // Can accept multiple connections
}

// NodeConnection represents a connection between two nodes
type NodeConnection struct {
	ID           string `json:"id"`
	FromNodeID   string `json:"from_node_id"`
	FromPortID   string `json:"from_port_id"`
	ToNodeID     string `json:"to_node_id"`
	ToPortID     string `json:"to_port_id"`
	DataMapping  string `json:"data_mapping"` // How data transforms between nodes
}

// EditorPosition represents x,y coordinates in the visual editor
type EditorPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Variable represents a variable used in the logic graph
type Variable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"default_value"`
	Scope        string      `json:"scope"` // local, character, session, global
}

// RuleParameter represents a customizable parameter for a rule template
type RuleParameter struct {
	Name         string      `json:"name"`
	DisplayName  string      `json:"display_name"`
	Type         string      `json:"type"` // number, string, boolean, choice, entity_reference
	DefaultValue interface{} `json:"default_value"`
	Constraints  Constraints `json:"constraints"`
	Description  string      `json:"description"`
}

// Constraints defines validation rules for parameters
type Constraints struct {
	Min        *float64 `json:"min,omitempty"`
	Max        *float64 `json:"max,omitempty"`
	Choices    []string `json:"choices,omitempty"`
	Pattern    string   `json:"pattern,omitempty"`
	Required   bool     `json:"required"`
	EntityType string   `json:"entity_type,omitempty"` // For entity references
}

// BalanceMetrics contains AI-analyzed balance information
type BalanceMetrics struct {
	PowerLevel          float64                  `json:"power_level"` // 0-10 scale
	ActionEconomy       float64                  `json:"action_economy"` // How many actions it requires/grants
	ResourceCost        float64                  `json:"resource_cost"` // Spell slots, HP, etc.
	ExpectedDamage      DamageExpectation        `json:"expected_damage"`
	UtilityScore        float64                  `json:"utility_score"`
	SynergyPotential    float64                  `json:"synergy_potential"`
	SimulationResults   []SimulationResult       `json:"simulation_results"`
	BalanceSuggestions  []BalanceSuggestion      `json:"balance_suggestions"`
	MetaImpactPrediction MetaImpactPrediction    `json:"meta_impact_prediction"`
}

// DamageExpectation represents predicted damage output
type DamageExpectation struct {
	MinDamage       float64            `json:"min_damage"`
	MaxDamage       float64            `json:"max_damage"`
	AverageDamage   float64            `json:"average_damage"`
	DamagePerRound  float64            `json:"damage_per_round"`
	DamageTypes     map[string]float64 `json:"damage_types"`
	TargetCount     float64            `json:"target_count"`
}

// SimulationResult represents the outcome of a balance simulation
type SimulationResult struct {
	ScenarioName    string                 `json:"scenario_name"`
	Level           int                    `json:"level"`
	SuccessRate     float64                `json:"success_rate"`
	AverageOutcome  map[string]interface{} `json:"average_outcome"`
	EdgeCases       []string               `json:"edge_cases"`
	ComparisonScore float64                `json:"comparison_score"` // vs similar abilities
}

// BalanceSuggestion represents an AI-generated balance adjustment
type BalanceSuggestion struct {
	Type        string  `json:"type"` // nerf, buff, rework, restriction
	Target      string  `json:"target"` // What aspect to change
	Suggestion  string  `json:"suggestion"`
	Impact      float64 `json:"impact"` // Expected power level change
	Reasoning   string  `json:"reasoning"`
	Priority    string  `json:"priority"` // high, medium, low
}

// MetaImpactPrediction predicts how this rule will affect the game meta
type MetaImpactPrediction struct {
	PopularityScore     float64          `json:"popularity_score"`
	ComboBreaker        bool             `json:"combo_breaker"`
	EnablesCombos       []string         `json:"enables_combos"`
	CounteredBy         []string         `json:"countered_by"`
	Counters            []string         `json:"counters"`
	ExpectedUsageRate   float64          `json:"expected_usage_rate"`
	MetaShiftPotential  float64          `json:"meta_shift_potential"`
}

// ConditionalRule represents a rule that applies under specific conditions
type ConditionalRule struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	ConditionType    string              `json:"condition_type"` // plane, emotion, backstory, environment
	Conditions       []RuleCondition     `json:"conditions"`
	ModifiedLogic    *LogicGraph         `json:"modified_logic"` // Optional override logic
	ParameterOverrides map[string]interface{} `json:"parameter_overrides"`
	Description      string              `json:"description"`
}

// RuleCondition represents a single condition that must be met
type RuleCondition struct {
	Type       string      `json:"type"` // location, character_state, time, narrative
	Operator   string      `json:"operator"` // equals, contains, greater_than, etc.
	Value      interface{} `json:"value"`
	Contextual bool        `json:"contextual"` // If true, value is evaluated at runtime
}

// RuleInstance represents an active instance of a rule in play
type RuleInstance struct {
	ID               string                 `json:"id" db:"id"`
	TemplateID       string                 `json:"template_id" db:"template_id"`
	OwnerID          string                 `json:"owner_id" db:"owner_id"` // Character, item, or location
	OwnerType        string                 `json:"owner_type" db:"owner_type"`
	SessionID        string                 `json:"session_id" db:"session_id"`
	ParameterValues  map[string]interface{} `json:"parameter_values" db:"parameter_values"`
	ActiveConditions []string               `json:"active_conditions" db:"active_conditions"`
	State            map[string]interface{} `json:"state" db:"state"` // Runtime state
	IsActive         bool                   `json:"is_active" db:"is_active"`
	ActivatedAt      *time.Time             `json:"activated_at,omitempty" db:"activated_at"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty" db:"expires_at"`
	UsageCount       int                    `json:"usage_count" db:"usage_count"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}

// NodeType constants
const (
	// Trigger nodes - what starts the rule
	NodeTypeTriggerAction      = "trigger_action"
	NodeTypeTriggerTime        = "trigger_time"
	NodeTypeTriggerCondition   = "trigger_condition"
	NodeTypeTriggerDamage      = "trigger_damage"
	NodeTypeTriggerMovement    = "trigger_movement"
	
	// Condition nodes - decision making
	NodeTypeConditionCheck     = "condition_check"
	NodeTypeConditionCompare   = "condition_compare"
	NodeTypeConditionRoll      = "condition_roll"
	NodeTypeConditionState     = "condition_state"
	
	// Action nodes - what happens
	NodeTypeActionDamage       = "action_damage"
	NodeTypeActionHeal         = "action_heal"
	NodeTypeActionEffect       = "action_effect"
	NodeTypeActionMove         = "action_move"
	NodeTypeActionResource     = "action_resource"
	NodeTypeActionRoll         = "action_roll"
	
	// Calculation nodes - math and logic
	NodeTypeCalcMath          = "calc_math"
	NodeTypeCalcRandom        = "calc_random"
	NodeTypeCalcAggregate     = "calc_aggregate"
	
	// Flow control
	NodeTypeFlowSplit         = "flow_split"
	NodeTypeFlowMerge         = "flow_merge"
	NodeTypeFlowLoop          = "flow_loop"
	NodeTypeFlowDelay         = "flow_delay"
)

// ConditionType constants for Conditional Reality System
const (
	ConditionTypePlane        = "plane"
	ConditionTypeEmotion      = "emotion"
	ConditionTypeBackstory    = "backstory"
	ConditionTypeEnvironment  = "environment"
	ConditionTypeNarrative    = "narrative"
	ConditionTypeRelationship = "relationship"
	ConditionTypeTime         = "time"
	ConditionTypeWeather      = "weather"
)

// ActiveRule represents an active instance of a rule in play
type ActiveRule struct {
	ID            string                 `json:"id" db:"id"`
	TemplateID    string                 `json:"template_id" db:"template_id"`
	GameSessionID string                 `json:"game_session_id" db:"game_session_id"`
	CharacterID   string                 `json:"character_id" db:"character_id"`
	CompiledLogic interface{}            `json:"compiled_logic" db:"compiled_logic"`
	Parameters    map[string]interface{} `json:"parameters" db:"parameters"`
	IsActive      bool                   `json:"is_active" db:"is_active"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// RuleExecution represents a record of rule execution
type RuleExecution struct {
	ID              string                 `json:"id" db:"id"`
	RuleID          string                 `json:"rule_id" db:"rule_id"`
	GameSessionID   string                 `json:"game_session_id" db:"game_session_id"`
	CharacterID     string                 `json:"character_id" db:"character_id"`
	TriggerContext  map[string]interface{} `json:"trigger_context" db:"trigger_context"`
	ExecutionResult map[string]interface{} `json:"execution_result" db:"execution_result"`
	Success         bool                   `json:"success" db:"success"`
	ErrorMessage    string                 `json:"error_message" db:"error_message"`
	ExecutedAt      time.Time              `json:"executed_at" db:"executed_at"`
}

// NodeTemplate represents a template for creating logic nodes
type NodeTemplate struct {
	ID                string                 `json:"id" db:"id"`
	NodeType          string                 `json:"node_type" db:"node_type"`
	Subtype           string                 `json:"subtype" db:"subtype"`
	Category          string                 `json:"category" db:"category"`
	Name              string                 `json:"name" db:"name"`
	Description       string                 `json:"description" db:"description"`
	Icon              string                 `json:"icon" db:"icon"`
	Color             string                 `json:"color" db:"color"`
	InputPorts        []NodePort             `json:"input_ports" db:"input_ports"`
	OutputPorts       []NodePort             `json:"output_ports" db:"output_ports"`
	DefaultProperties map[string]interface{} `json:"default_properties" db:"default_properties"`
}

// ConditionalContext represents an active conditional context
type ConditionalContext struct {
	ID           string                 `json:"id"`
	SessionID    string                 `json:"session_id"`
	ContextType  string                 `json:"context_type"`
	ContextValue map[string]interface{} `json:"context_value"`
	IsActive     bool                   `json:"is_active"`
	StartedAt    time.Time              `json:"started_at"`
	EndedAt      *time.Time             `json:"ended_at,omitempty"`
}

// ConditionalModifier represents a modifier that applies under specific conditions
type ConditionalModifier struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	ContextType  string                 `json:"context_type"`
	ContextValue string                 `json:"context_value"`
	Modifiers    map[string]interface{} `json:"modifiers"`
	Priority     int                    `json:"priority"`
	Description  string                 `json:"description"`
}

// LevelRange represents a range of character levels
type LevelRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}