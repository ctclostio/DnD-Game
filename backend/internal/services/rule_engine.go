package services

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/google/uuid"
)

// RuleEngine handles compilation and execution of visual logic rules
type RuleEngine struct {
	nodeExecutors map[string]NodeExecutor
	diceRoller    *DiceRollService
	repository    *database.RuleBuilderRepository
}

// NodeExecutor defines the interface for executing different node types
type NodeExecutor interface {
	Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error)
}

// ExecutionState tracks the state during rule execution
type ExecutionState struct {
	Variables      map[string]interface{}
	ExecutedNodes  []string
	StartTime      time.Time
	CurrentEntity  interface{} // The entity executing the rule
	Context        map[string]interface{}
	Errors         []string
	ExecutionPath  []ExecutionStep
}

// ExecutionStep records a step in the execution path for debugging
type ExecutionStep struct {
	NodeID    string
	NodeType  string
	Inputs    map[string]interface{}
	Outputs   map[string]interface{}
	Duration  time.Duration
	Timestamp time.Time
}

// CompiledRule represents a compiled rule ready for execution
type CompiledRule struct {
	TemplateID     string
	Graph          *models.LogicGraph
	ExecutionOrder []string
	Parameters     []models.RuleParameter
	CompiledAt     time.Time
}

// TriggerData represents data that triggered rule execution
type TriggerData struct {
	Type       string
	Source     interface{}
	Target     interface{}
	Properties map[string]interface{}
}

// ExecutionResult represents the result of rule execution
type ExecutionResult struct {
	Success       bool
	Duration      time.Duration
	ExecutedNodes []string
	FinalState    map[string]interface{}
	Errors        []string
	ExecutionPath []ExecutionStep
}

// ValidationResult represents the result of rule validation
type ValidationResult struct {
	IsValid  bool
	Errors   []string
	Warnings []string
}

// NewRuleEngine creates a new rule engine instance
func NewRuleEngine(repository *database.RuleBuilderRepository, diceRoller *DiceRollService) *RuleEngine {
	engine := &RuleEngine{
		nodeExecutors: make(map[string]NodeExecutor),
		diceRoller:    diceRoller,
		repository:    repository,
	}

	// Register all node executors
	engine.registerNodeExecutors()

	return engine
}

// CompileRule validates and optimizes a rule template
func (re *RuleEngine) CompileRule(template *models.RuleTemplate) (*CompiledRule, error) {
	// Validate the logic graph
	if err := re.validateLogicGraph(&template.LogicGraph); err != nil {
		return nil, fmt.Errorf("invalid logic graph: %w", err)
	}

	// Build execution order
	executionOrder, err := re.buildExecutionOrder(&template.LogicGraph)
	if err != nil {
		return nil, fmt.Errorf("failed to build execution order: %w", err)
	}

	// Optimize the graph
	optimizedGraph := re.optimizeGraph(&template.LogicGraph)

	return &CompiledRule{
		TemplateID:     template.ID,
		Graph:          optimizedGraph,
		ExecutionOrder: executionOrder,
		Parameters:     template.Parameters,
		CompiledAt:     time.Now(),
	}, nil
}

// ExecuteRule executes a compiled rule with given parameters
func (re *RuleEngine) ExecuteRule(ctx context.Context, compiled *CompiledRule, instance *models.RuleInstance, trigger TriggerData) (*ExecutionResult, error) {
	state := &ExecutionState{
		Variables:     make(map[string]interface{}),
		ExecutedNodes: []string{},
		StartTime:     time.Now(),
		Context:       make(map[string]interface{}),
		Errors:        []string{},
		ExecutionPath: []ExecutionStep{},
	}

	// Initialize variables from instance parameters
	for name, value := range instance.ParameterValues {
		state.Variables[name] = value
	}

	// Add trigger data to context
	state.Context["trigger"] = trigger
	state.Context["instance"] = instance

	// Execute nodes in order
	for _, nodeID := range compiled.ExecutionOrder {
		node := re.findNode(compiled.Graph, nodeID)
		if node == nil {
			continue
		}

		// Check if node should execute based on incoming connections
		inputs, shouldExecute := re.gatherNodeInputs(node, compiled.Graph, state)
		if !shouldExecute {
			continue
		}

		// Execute the node
		stepStart := time.Now()
		outputs, err := re.executeNode(ctx, node, inputs, state)
		
		step := ExecutionStep{
			NodeID:    node.ID,
			NodeType:  node.Type,
			Inputs:    inputs,
			Outputs:   outputs,
			Duration:  time.Since(stepStart),
			Timestamp: stepStart,
		}
		state.ExecutionPath = append(state.ExecutionPath, step)

		if err != nil {
			state.Errors = append(state.Errors, fmt.Sprintf("node %s error: %v", node.ID, err))
			if re.isCriticalError(err) {
				return nil, err
			}
			continue
		}

		// Store outputs for connected nodes
		state.Context[node.ID] = outputs
		state.ExecutedNodes = append(state.ExecutedNodes, node.ID)
	}

	return &ExecutionResult{
		Success:       len(state.Errors) == 0,
		Duration:      time.Since(state.StartTime),
		ExecutedNodes: state.ExecutedNodes,
		FinalState:    state.Variables,
		Errors:        state.Errors,
		ExecutionPath: state.ExecutionPath,
	}, nil
}

// validateLogicGraph ensures the graph is valid and executable
func (re *RuleEngine) validateLogicGraph(graph *models.LogicGraph) error {
	// Check for start node
	if graph.StartNodeID == "" {
		return fmt.Errorf("no start node defined")
	}

	// Validate all connections
	nodeMap := make(map[string]*models.LogicNode)
	for i := range graph.Nodes {
		node := &graph.Nodes[i]
		nodeMap[node.ID] = node
	}

	for _, conn := range graph.Connections {
		// Validate source node and port
		fromNode, ok := nodeMap[conn.FromNodeID]
		if !ok {
			return fmt.Errorf("connection references non-existent node: %s", conn.FromNodeID)
		}

		// Validate target node and port
		toNode, ok := nodeMap[conn.ToNodeID]
		if !ok {
			return fmt.Errorf("connection references non-existent node: %s", conn.ToNodeID)
		}

		// Validate ports exist
		if !re.hasPort(fromNode.Outputs, conn.FromPortID) {
			return fmt.Errorf("invalid output port %s on node %s", conn.FromPortID, conn.FromNodeID)
		}

		if !re.hasPort(toNode.Inputs, conn.ToPortID) {
			return fmt.Errorf("invalid input port %s on node %s", conn.ToPortID, conn.ToNodeID)
		}
	}

	// Check for cycles
	if re.hasCycles(graph) {
		return fmt.Errorf("logic graph contains cycles")
	}

	return nil
}

// buildExecutionOrder creates a topologically sorted execution order
func (re *RuleEngine) buildExecutionOrder(graph *models.LogicGraph) ([]string, error) {
	// Build adjacency list
	adjacency := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, node := range graph.Nodes {
		adjacency[node.ID] = []string{}
		inDegree[node.ID] = 0
	}

	for _, conn := range graph.Connections {
		adjacency[conn.FromNodeID] = append(adjacency[conn.FromNodeID], conn.ToNodeID)
		inDegree[conn.ToNodeID]++
	}

	// Topological sort using Kahn's algorithm
	queue := []string{}
	
	// Start with nodes that have no incoming connections
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	var order []string
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		order = append(order, nodeID)

		for _, neighbor := range adjacency[nodeID] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(order) != len(graph.Nodes) {
		return nil, fmt.Errorf("failed to create execution order - graph may contain cycles")
	}

	return order, nil
}

// registerNodeExecutors sets up all the node type executors
func (re *RuleEngine) registerNodeExecutors() {
	// Trigger executors
	re.nodeExecutors[models.NodeTypeTriggerAction] = &ActionTriggerExecutor{}
	re.nodeExecutors[models.NodeTypeTriggerDamage] = &DamageTriggerExecutor{}
	re.nodeExecutors[models.NodeTypeTriggerTime] = &TimeTriggerExecutor{}

	// Condition executors
	re.nodeExecutors[models.NodeTypeConditionCheck] = &ConditionCheckExecutor{}
	re.nodeExecutors[models.NodeTypeConditionCompare] = &CompareExecutor{}
	re.nodeExecutors[models.NodeTypeConditionRoll] = &RollCheckExecutor{diceRoller: re.diceRoller}

	// Action executors
	re.nodeExecutors[models.NodeTypeActionDamage] = &DamageActionExecutor{}
	re.nodeExecutors[models.NodeTypeActionHeal] = &HealActionExecutor{}
	re.nodeExecutors[models.NodeTypeActionEffect] = &EffectActionExecutor{}
	re.nodeExecutors[models.NodeTypeActionResource] = &ResourceActionExecutor{}

	// Calculation executors
	re.nodeExecutors[models.NodeTypeCalcMath] = &MathExecutor{}
	re.nodeExecutors[models.NodeTypeCalcRandom] = &RandomExecutor{diceRoller: re.diceRoller}
}

// Helper methods

func (re *RuleEngine) executeNode(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	executor, ok := re.nodeExecutors[node.Type]
	if !ok {
		return nil, fmt.Errorf("no executor for node type: %s", node.Type)
	}

	return executor.Execute(ctx, node, inputs, state)
}

func (re *RuleEngine) findNode(graph *models.LogicGraph, nodeID string) *models.LogicNode {
	for i := range graph.Nodes {
		if graph.Nodes[i].ID == nodeID {
			return &graph.Nodes[i]
		}
	}
	return nil
}

func (re *RuleEngine) hasPort(ports []models.NodePort, portID string) bool {
	for _, port := range ports {
		if port.ID == portID {
			return true
		}
	}
	return false
}

func (re *RuleEngine) hasCycles(graph *models.LogicGraph) bool {
	// Simple cycle detection using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for _, node := range graph.Nodes {
		if re.hasCyclesDFS(node.ID, graph, visited, recStack) {
			return true
		}
	}

	return false
}

func (re *RuleEngine) hasCyclesDFS(nodeID string, graph *models.LogicGraph, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	// Find all connections from this node
	for _, conn := range graph.Connections {
		if conn.FromNodeID == nodeID {
			if !visited[conn.ToNodeID] {
				if re.hasCyclesDFS(conn.ToNodeID, graph, visited, recStack) {
					return true
				}
			} else if recStack[conn.ToNodeID] {
				return true
			}
		}
	}

	recStack[nodeID] = false
	return false
}

func (re *RuleEngine) gatherNodeInputs(node *models.LogicNode, graph *models.LogicGraph, state *ExecutionState) (map[string]interface{}, bool) {
	inputs := make(map[string]interface{})
	
	// Find all connections to this node
	for _, conn := range graph.Connections {
		if conn.ToNodeID == node.ID {
			// Get output from source node
			sourceOutputs, ok := state.Context[conn.FromNodeID].(map[string]interface{})
			if !ok {
				continue
			}

			// Map the specific output port to input port
			if value, ok := sourceOutputs[conn.FromPortID]; ok {
				inputs[conn.ToPortID] = value
			}
		}
	}

	// Check if all required inputs are present
	for _, inputPort := range node.Inputs {
		if inputPort.Required {
			if _, ok := inputs[inputPort.ID]; !ok {
				return inputs, false
			}
		}
	}

	return inputs, true
}

func (re *RuleEngine) optimizeGraph(graph *models.LogicGraph) *models.LogicGraph {
	// Simple optimizations
	// TODO: Implement graph optimizations like constant folding, dead code elimination
	return graph
}

func (re *RuleEngine) isCriticalError(err error) bool {
	// Determine if an error should stop execution
	return strings.Contains(err.Error(), "critical") || strings.Contains(err.Error(), "fatal")
}

// Supporting types

// Node Executor Implementations

// MathExecutor handles mathematical operations
type MathExecutor struct{}

func (e *MathExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	a, ok := inputs["a"].(float64)
	if !ok {
		if aInt, ok := inputs["a"].(int); ok {
			a = float64(aInt)
		} else {
			return nil, fmt.Errorf("invalid input 'a'")
		}
	}

	b, ok := inputs["b"].(float64)
	if !ok {
		if bInt, ok := inputs["b"].(int); ok {
			b = float64(bInt)
		} else {
			return nil, fmt.Errorf("invalid input 'b'")
		}
	}

	operation, _ := node.Properties["operation"].(string)
	var result float64

	switch operation {
	case "+":
		result = a + b
	case "-":
		result = a - b
	case "*":
		result = a * b
	case "/":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	case "^":
		result = math.Pow(a, b)
	case "min":
		result = math.Min(a, b)
	case "max":
		result = math.Max(a, b)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}

	return map[string]interface{}{
		"result": result,
	}, nil
}

// RandomExecutor handles dice rolls and random number generation
type RandomExecutor struct {
	diceRoller *DiceRollService
}

func (e *RandomExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	diceNotation, _ := node.Properties["dice_notation"].(string)
	if diceNotation == "" {
		// Random number between min and max
		min, _ := node.Properties["min"].(float64)
		max, _ := node.Properties["max"].(float64)
		
		if max <= min {
			return nil, fmt.Errorf("max must be greater than min")
		}

		result := min + rand.Float64()*(max-min)
		return map[string]interface{}{
			"result": result,
		}, nil
	}

	// Parse dice notation
	result, details, err := e.parseDiceNotation(diceNotation)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"result": result,
		"details": details,
	}, nil
}

func (e *RandomExecutor) parseDiceNotation(notation string) (int, map[string]interface{}, error) {
	// Simple dice notation parser (e.g., "2d6+3")
	re := regexp.MustCompile(`(\d+)d(\d+)([+-]\d+)?`)
	matches := re.FindStringSubmatch(notation)
	
	if len(matches) < 3 {
		return 0, nil, fmt.Errorf("invalid dice notation: %s", notation)
	}

	count, _ := strconv.Atoi(matches[1])
	sides, _ := strconv.Atoi(matches[2])
	modifier := 0
	
	if len(matches) > 3 && matches[3] != "" {
		modifier, _ = strconv.Atoi(matches[3])
	}

	total := 0
	rolls := []int{}
	
	for i := 0; i < count; i++ {
		roll := rand.Intn(sides) + 1
		rolls = append(rolls, roll)
		total += roll
	}
	
	total += modifier

	details := map[string]interface{}{
		"rolls":    rolls,
		"modifier": modifier,
		"dice":     notation,
	}

	return total, details, nil
}

// ConditionCheckExecutor handles if/else branching
type ConditionCheckExecutor struct{}

func (e *ConditionCheckExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	condition, ok := inputs["condition"].(bool)
	if !ok {
		return nil, fmt.Errorf("condition input must be boolean")
	}

	outputs := make(map[string]interface{})
	
	if condition {
		outputs["true"] = true
	} else {
		outputs["false"] = true
	}

	return outputs, nil
}

// CompareExecutor handles number comparisons
type CompareExecutor struct{}

func (e *CompareExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	a := toFloat64(inputs["a"])
	b := toFloat64(inputs["b"])
	operator, _ := node.Properties["operator"].(string)

	var result bool
	switch operator {
	case ">":
		result = a > b
	case ">=":
		result = a >= b
	case "<":
		result = a < b
	case "<=":
		result = a <= b
	case "==":
		result = a == b
	case "!=":
		result = a != b
	default:
		return nil, fmt.Errorf("unknown operator: %s", operator)
	}

	return map[string]interface{}{
		"result": result,
	}, nil
}

// Helper function to convert interface{} to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// Placeholder implementations for other executors
type ActionTriggerExecutor struct{}
func (e *ActionTriggerExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	return map[string]interface{}{"trigger": true}, nil
}

type DamageTriggerExecutor struct{}
func (e *DamageTriggerExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	return map[string]interface{}{"trigger": true, "damage_amount": 0}, nil
}

type TimeTriggerExecutor struct{}
func (e *TimeTriggerExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	return map[string]interface{}{"trigger": true}, nil
}

type RollCheckExecutor struct{ diceRoller *DiceRollService }
func (e *RollCheckExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	return map[string]interface{}{"success": true, "roll_total": 15}, nil
}

type DamageActionExecutor struct{}
func (e *DamageActionExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	return map[string]interface{}{"out": true}, nil
}

type HealActionExecutor struct{}
func (e *HealActionExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	return map[string]interface{}{"out": true}, nil
}

type EffectActionExecutor struct{}
func (e *EffectActionExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	return map[string]interface{}{"out": true}, nil
}

type ResourceActionExecutor struct{}
func (e *ResourceActionExecutor) Execute(ctx context.Context, node *models.LogicNode, inputs map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	return map[string]interface{}{"out": true}, nil
}

// Repository wrapper methods

// GetRuleTemplates gets rule templates with filters
func (re *RuleEngine) GetRuleTemplates(userID, category string, isPublic bool) ([]models.RuleTemplate, error) {
	return re.repository.GetRuleTemplates(userID, category, isPublic)
}

// GetRuleTemplate gets a rule template by ID
func (re *RuleEngine) GetRuleTemplate(templateID string) (*models.RuleTemplate, error) {
	return re.repository.GetRuleTemplate(templateID)
}

// CreateRuleTemplate creates a new rule template
func (re *RuleEngine) CreateRuleTemplate(template *models.RuleTemplate) (*models.RuleTemplate, error) {
	// Generate ID if not provided
	if template.ID == "" {
		template.ID = uuid.New().String()
	}
	
	// Set timestamps
	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now
	
	// Validate and compile the template
	_, err := re.CompileRule(template)
	if err != nil {
		return nil, fmt.Errorf("invalid rule template: %w", err)
	}
	
	// Save to repository
	if err := re.repository.CreateRuleTemplate(template); err != nil {
		return nil, err
	}
	
	return template, nil
}

// UpdateRuleTemplate updates a rule template
func (re *RuleEngine) UpdateRuleTemplate(templateID string, updates map[string]interface{}) (*models.RuleTemplate, error) {
	// Update in repository
	if err := re.repository.UpdateRuleTemplate(templateID, updates); err != nil {
		return nil, err
	}
	
	// Return updated template
	return re.repository.GetRuleTemplate(templateID)
}

// DeleteRuleTemplate deletes a rule template
func (re *RuleEngine) DeleteRuleTemplate(templateID string) error {
	return re.repository.DeleteRuleTemplate(templateID)
}

// GetNodeTemplates gets all available node templates
func (re *RuleEngine) GetNodeTemplates() ([]models.NodeTemplate, error) {
	return re.repository.GetNodeTemplates()
}

// ActivateRule creates an active instance of a rule
func (re *RuleEngine) ActivateRule(templateID, gameSessionID, characterID string, parameters map[string]interface{}) (*models.ActiveRule, error) {
	template, err := re.repository.GetRuleTemplate(templateID)
	if err != nil {
		return nil, err
	}
	
	// Compile the rule
	compiled, err := re.CompileRule(template)
	if err != nil {
		return nil, err
	}
	
	activeRule := &models.ActiveRule{
		ID:            uuid.New().String(),
		TemplateID:    templateID,
		GameSessionID: gameSessionID,
		CharacterID:   characterID,
		CompiledLogic: compiled,
		Parameters:    parameters,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	if err := re.repository.CreateActiveRule(activeRule); err != nil {
		return nil, err
	}
	
	// Increment usage count
	re.repository.IncrementUsageCount(templateID)
	
	return activeRule, nil
}

// GetActiveRules gets active rules for a session or character
func (re *RuleEngine) GetActiveRules(gameSessionID, characterID string) ([]models.ActiveRule, error) {
	return re.repository.GetActiveRules(gameSessionID, characterID)
}

// DeactivateRule deactivates an active rule
func (re *RuleEngine) DeactivateRule(ruleID string) error {
	return re.repository.DeactivateRule(ruleID)
}


// ValidateRule validates a logic graph
func (re *RuleEngine) ValidateRule(graph models.LogicGraph) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}
	
	// Validate graph structure
	if err := re.validateLogicGraph(&graph); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
	}
	
	return result, nil
}

// GetRuleExecutionHistory gets rule execution history
func (re *RuleEngine) GetRuleExecutionHistory(gameSessionID, characterID string, limit int) ([]models.RuleExecution, error) {
	return re.repository.GetRuleExecutionHistory(gameSessionID, characterID, limit)
}