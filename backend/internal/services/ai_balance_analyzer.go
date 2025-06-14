package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/ctclostio/DnD-Game/backend/internal/config"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// AIBalanceAnalyzer uses AI to analyze and balance custom rules.
type AIBalanceAnalyzer struct {
	llm        LLMProvider
	ruleEngine *RuleEngine
	combatSim  *CombatSimulator
	cfg        *config.Config
}

// CombatSimulator runs combat simulations for balance testing.
type CombatSimulator struct {
	combatService *CombatService
}

// NewAIBalanceAnalyzer creates a new balance analyzer instance.
func NewAIBalanceAnalyzer(cfg *config.Config, llm LLMProvider, ruleEngine *RuleEngine, combatService *CombatService) *AIBalanceAnalyzer {
	return &AIBalanceAnalyzer{
		llm:        llm,
		ruleEngine: ruleEngine,
		combatSim:  &CombatSimulator{combatService: combatService},
		cfg:        cfg,
	}
}

// AnalyzeRuleBalance performs comprehensive balance analysis on a rule template.
func (ba *AIBalanceAnalyzer) AnalyzeRuleBalance(ctx context.Context, template *models.RuleTemplate) (*models.BalanceMetrics, error) {
	// Run simulations across different scenarios
	simResults := ba.runSimulations(ctx, template)

	// Calculate damage expectations if applicable
	damageExpectation := ba.calculateDamageExpectation(template)

	// Analyze action economy impact
	actionEconomy := ba.analyzeActionEconomy(template)

	// Calculate resource costs
	resourceCost := ba.calculateResourceCost(template)

	// Determine utility score
	utilityScore := ba.calculateUtilityScore(template)

	// Analyze synergy potential
	synergyPotential := ba.analyzeSynergyPotential(template)

	// Generate AI balance suggestions
	suggestions, err := ba.generateBalanceSuggestions(ctx, template, simResults)
	if err != nil {
		// Non-critical error, continue without suggestions
		suggestions = []models.BalanceSuggestion{}
	}

	// Predict meta impact
	metaImpact := ba.predictMetaImpact(ctx, template, simResults)

	// Calculate overall power level
	powerLevel := ba.calculatePowerLevel(
		damageExpectation,
		actionEconomy,
		resourceCost,
		utilityScore,
		synergyPotential,
		simResults,
	)

	return &models.BalanceMetrics{
		PowerLevel:           powerLevel,
		ActionEconomy:        actionEconomy,
		ResourceCost:         resourceCost,
		ExpectedDamage:       damageExpectation,
		UtilityScore:         utilityScore,
		SynergyPotential:     synergyPotential,
		SimulationResults:    simResults,
		BalanceSuggestions:   suggestions,
		MetaImpactPrediction: metaImpact,
	}, nil
}

// runSimulations executes balance simulations across various scenarios.
func (ba *AIBalanceAnalyzer) runSimulations(ctx context.Context, template *models.RuleTemplate) []models.SimulationResult {
	scenarios := ba.getSimulationScenarios(template.Category)
	results := []models.SimulationResult{}

	for _, scenario := range scenarios {
		result, err := ba.simulateScenario(ctx, template, scenario)
		if err != nil {
			// Log error but continue with other scenarios
			continue
		}
		results = append(results, result)
	}

	return results
}

// simulateScenario runs a single simulation scenario.
func (ba *AIBalanceAnalyzer) simulateScenario(ctx context.Context, template *models.RuleTemplate, scenario SimulationScenario) (models.SimulationResult, error) {
	successCount := 0
	outcomes := []map[string]interface{}{}
	edgeCases := []string{}

	// Run multiple iterations
	iterations := 1000
	for i := 0; i < iterations; i++ {
		outcome, success, edgeCase := ba.runSingleSimulation(ctx, template, scenario)
		if success {
			successCount++
		}
		outcomes = append(outcomes, outcome)
		if edgeCase != "" && !containsInBalancer(edgeCases, edgeCase) {
			edgeCases = append(edgeCases, edgeCase)
		}
	}

	// Calculate average outcome
	avgOutcome := ba.calculateAverageOutcome(outcomes)

	// Compare to baseline abilities
	comparisonScore := ba.compareToBaseline(template, scenario, avgOutcome)

	return models.SimulationResult{
		ScenarioName:    scenario.Name,
		Level:           scenario.Level,
		SuccessRate:     float64(successCount) / float64(iterations),
		AverageOutcome:  avgOutcome,
		EdgeCases:       edgeCases,
		ComparisonScore: comparisonScore,
	}, nil
}

// calculateDamageExpectation analyzes potential damage output.
func (ba *AIBalanceAnalyzer) calculateDamageExpectation(template *models.RuleTemplate) models.DamageExpectation {
	expectation := models.DamageExpectation{
		DamageTypes: make(map[string]float64),
	}

	// Analyze logic graph for damage nodes
	for _, node := range template.LogicGraph.Nodes {
		if node.Type == models.NodeTypeActionDamage {
			dice, _ := node.Properties["damage_dice"].(string)
			damageType, _ := node.Properties["damage_type"].(string)

			// Parse dice notation to get average damage
			avgDamage := ba.calculateAverageDiceRoll(dice)
			expectation.DamageTypes[damageType] += avgDamage
			expectation.AverageDamage += avgDamage
		}
	}

	// Calculate min/max based on dice
	expectation.MinDamage = expectation.AverageDamage * 0.5
	expectation.MaxDamage = expectation.AverageDamage * 1.5

	// Estimate targets affected
	expectation.TargetCount = ba.estimateTargetCount(template)

	// Calculate damage per round
	actionCost := ba.analyzeActionEconomy(template)
	if actionCost > 0 {
		expectation.DamagePerRound = expectation.AverageDamage / actionCost
	}

	return expectation
}

// generateBalanceSuggestions uses AI to suggest balance adjustments.
func (ba *AIBalanceAnalyzer) generateBalanceSuggestions(ctx context.Context, template *models.RuleTemplate, simResults []models.SimulationResult) ([]models.BalanceSuggestion, error) {
	if !ba.cfg.AI.Enabled {
		return ba.generateDefaultSuggestions(template, simResults), nil
	}

	// Prepare context for AI
	prompt := fmt.Sprintf(`Analyze this custom D&D rule for balance issues and suggest adjustments:

Rule Name: %s
Category: %s
Description: %s

Simulation Results Summary:
%s

Logic Graph Analysis:
- Number of nodes: %d
- Triggers: %s
- Actions: %s
- Conditions: %s

Current Balance Metrics:
- Overall Power Level: %.2f/10
- Resource Cost: %.2f
- Action Economy: %.2f actions

Please provide balance suggestions in JSON format:
[
  {
    "type": "nerf/buff/rework/restriction",
    "target": "specific aspect to change",
    "suggestion": "detailed suggestion",
    "impact": estimated power level change (-5 to +5),
    "reasoning": "why this change is needed",
    "priority": "high/medium/low"
  }
]

Consider:
1. How this compares to similar official abilities
2. Potential for abuse or broken combinations
3. Fun factor vs balance
4. Clarity and ease of use
5. Thematic appropriateness`,
		template.Name,
		template.Category,
		template.Description,
		ba.summarizeSimResults(simResults),
		len(template.LogicGraph.Nodes),
		ba.listNodeTypes(template.LogicGraph, "trigger"),
		ba.listNodeTypes(template.LogicGraph, "action"),
		ba.listNodeTypes(template.LogicGraph, "condition"),
		template.BalanceMetrics.PowerLevel,
		template.BalanceMetrics.ResourceCost,
		template.BalanceMetrics.ActionEconomy,
	)

	systemPrompt := "You are a D&D 5th edition game balance expert. Analyze the provided rule and suggest balance adjustments."
	response, err := ba.llm.GenerateContent(ctx, prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var suggestions []models.BalanceSuggestion
	if err := json.Unmarshal([]byte(response), &suggestions); err != nil {
		// Try to parse as best as possible
		return ba.generateDefaultSuggestions(template, simResults), nil
	}

	return suggestions, nil
}

// predictMetaImpact predicts how the rule will affect the game meta.
func (ba *AIBalanceAnalyzer) predictMetaImpact(ctx context.Context, template *models.RuleTemplate, simResults []models.SimulationResult) models.MetaImpactPrediction {
	prediction := models.MetaImpactPrediction{
		EnablesCombos: []string{},
		CounteredBy:   []string{},
		Counters:      []string{},
	}

	// Analyze potential combos
	if template.Category == "spell" || template.Category == "ability" {
		// Check for action surge potential
		if ba.analyzeActionEconomy(template) < 1 {
			prediction.EnablesCombos = append(prediction.EnablesCombos, "Action Surge combos")
		}

		// Check for metamagic potential
		if ba.hasSpellProperties(template) {
			prediction.EnablesCombos = append(prediction.EnablesCombos, "Metamagic enhancement")
		}
	}

	// Predict counters based on damage types and effects
	damageTypes := ba.getDamageTypes(template)
	for damageType := range damageTypes {
		switch damageType {
		case "fire":
			prediction.CounteredBy = append(prediction.CounteredBy, "Fire resistance/immunity")
		case "psychic":
			prediction.CounteredBy = append(prediction.CounteredBy, "Mind blank, Psychic resistance")
		}
	}

	// Check if it's a counter to existing strategies
	if ba.hasCounterspellProperties(template) {
		prediction.Counters = append(prediction.Counters, "Spellcasting strategies")
		prediction.ComboBreaker = true
	}

	// Calculate popularity and usage predictions
	prediction.PopularityScore = ba.calculatePopularityScore(template, simResults)
	prediction.ExpectedUsageRate = prediction.PopularityScore * 0.8

	// Meta shift potential based on power and uniqueness
	uniqueness := ba.calculateUniqueness(template)
	prediction.MetaShiftPotential = (template.BalanceMetrics.PowerLevel/10 + uniqueness) / 2

	return prediction
}

// Helper methods

func (ba *AIBalanceAnalyzer) getSimulationScenarios(category string) []SimulationScenario {
	baseScenarios := []SimulationScenario{
		{Name: "Level 1 Solo", Level: 1, PartySize: 1, EnemyCount: 1},
		{Name: "Level 5 Party", Level: 5, PartySize: 4, EnemyCount: 4},
		{Name: "Level 10 Boss Fight", Level: 10, PartySize: 4, EnemyCount: 1},
		{Name: "Level 15 Horde", Level: 15, PartySize: 4, EnemyCount: 8},
		{Name: "Level 20 Epic", Level: 20, PartySize: 4, EnemyCount: 2},
	}

	// Add category-specific scenarios
	switch category {
	case "spell":
		baseScenarios = append(baseScenarios, SimulationScenario{
			Name: "Counterspell Scenario", Level: 10, PartySize: 4, SpecialConditions: []string{"enemy_counterspell"},
		})
	case "environmental":
		baseScenarios = append(baseScenarios, SimulationScenario{
			Name: "Hazardous Terrain", Level: 8, PartySize: 4, SpecialConditions: []string{"difficult_terrain", "environmental_damage"},
		})
	}

	return baseScenarios
}

func (ba *AIBalanceAnalyzer) calculatePowerLevel(
	damage models.DamageExpectation,
	actionEconomy float64,
	resourceCost float64,
	utilityScore float64,
	synergyPotential float64,
	simResults []models.SimulationResult,
) float64 {
	// Weight different factors
	weights := map[string]float64{
		"damage":     0.3,
		"action":     0.2,
		"resource":   0.15,
		"utility":    0.2,
		"synergy":    0.1,
		"simulation": 0.05,
	}

	// Normalize damage (assume 50 damage per round is max expected)
	normalizedDamage := math.Min(damage.DamagePerRound/50, 1.0) * 10

	// Normalize action economy (lower is better, 0 = free action)
	normalizedAction := (2 - math.Min(actionEconomy, 2)) / 2 * 10

	// Normalize resource cost (lower is better)
	normalizedResource := (1 - math.Min(resourceCost, 1)) * 10

	// Calculate average simulation success
	avgSimSuccess := 0.0
	for _, result := range simResults {
		avgSimSuccess += result.ComparisonScore
	}
	if len(simResults) > 0 {
		avgSimSuccess /= float64(len(simResults))
	}

	powerLevel := normalizedDamage*weights["damage"] +
		normalizedAction*weights["action"] +
		normalizedResource*weights["resource"] +
		utilityScore*weights["utility"] +
		synergyPotential*weights["synergy"] +
		avgSimSuccess*weights["simulation"]

	return math.Min(math.Max(powerLevel, 0), 10)
}

func (ba *AIBalanceAnalyzer) analyzeActionEconomy(template *models.RuleTemplate) float64 {
	// Analyze trigger nodes to determine action cost
	actionCost := 1.0 // Default to 1 action

	for _, node := range template.LogicGraph.Nodes {
		if strings.HasPrefix(node.Type, "trigger_") {
			triggerType, _ := node.Properties["trigger_type"].(string)
			switch triggerType {
			case constants.ActionTypeBonusAction:
				actionCost = 0.5
			case "reaction":
				actionCost = 0.3
			case "free":
				actionCost = 0
			case "full_round":
				actionCost = 2
			}
		}
	}

	return actionCost
}

func (ba *AIBalanceAnalyzer) calculateResourceCost(template *models.RuleTemplate) float64 {
	cost := 0.0

	// Check for spell slot usage
	for _, param := range template.Parameters {
		if param.Name == "spell_slot_level" {
			if level, ok := param.DefaultValue.(float64); ok {
				cost += level * 0.2 // Each spell level adds 0.2 to cost
			}
		}
	}

	// Check for other resource consumption
	for _, node := range template.LogicGraph.Nodes {
		if node.Type == models.NodeTypeActionResource {
			amount, _ := node.Properties["amount"].(float64)
			resourceType, _ := node.Properties["resource"].(string)

			switch resourceType {
			case "hit_points":
				cost += amount / 50 // Normalize HP cost
			case "hit_dice":
				cost += 0.3 // Hit dice are valuable
			case "exhaustion":
				cost += 0.5 // Exhaustion is very costly
			}
		}
	}

	return math.Min(cost, 1.0)
}

func (ba *AIBalanceAnalyzer) calculateUtilityScore(template *models.RuleTemplate) float64 {
	score := 0.0

	// Award points for different utility effects
	utilityEffects := map[string]float64{
		"movement":      2.0,
		"invisibility":  3.0,
		"teleportation": 4.0,
		"flight":        4.0,
		"healing":       2.5,
		"buff":          2.0,
		"debuff":        2.5,
		"control":       3.0,
		"summon":        3.5,
		"transform":     4.0,
	}

	// Scan nodes for utility effects
	for _, node := range template.LogicGraph.Nodes {
		effectType, _ := node.Properties["effect_type"].(string)
		if points, ok := utilityEffects[effectType]; ok {
			score += points
		}
	}

	return math.Min(score, 10.0)
}

func (ba *AIBalanceAnalyzer) analyzeSynergyPotential(template *models.RuleTemplate) float64 {
	potential := 5.0 // Base synergy

	// Check for concentration (limits synergy)
	if ba.requiresConcentration(template) {
		potential -= 2.0
	}

	// Check for combo-enabling properties
	if ba.grantsAdvantage(template) {
		potential += 1.5
	}

	if ba.grantsExtraActions(template) {
		potential += 2.0
	}

	// Check for stacking potential
	if ba.isStackable(template) {
		potential += 1.0
	}

	return math.Min(math.Max(potential, 0), 10)
}

func (ba *AIBalanceAnalyzer) generateDefaultSuggestions(template *models.RuleTemplate, simResults []models.SimulationResult) []models.BalanceSuggestion {
	suggestions := []models.BalanceSuggestion{}

	// Check if overpowered
	if template.BalanceMetrics.PowerLevel > 7.5 {
		suggestions = append(suggestions, models.BalanceSuggestion{
			Type:       "nerf",
			Target:     "overall power",
			Suggestion: "Consider reducing damage output or adding resource costs",
			Impact:     -2.0,
			Reasoning:  "Power level significantly exceeds comparable abilities",
			Priority:   "high",
		})
	}

	// Check if underpowered
	if template.BalanceMetrics.PowerLevel < 3.0 {
		suggestions = append(suggestions, models.BalanceSuggestion{
			Type:       "buff",
			Target:     "effectiveness",
			Suggestion: "Consider increasing damage, reducing action cost, or adding utility",
			Impact:     2.0,
			Reasoning:  "Power level is too low to be worth using",
			Priority:   "medium",
		})
	}

	return suggestions
}

// Utility functions

func (ba *AIBalanceAnalyzer) calculateAverageDiceRoll(diceNotation string) float64 {
	// Simple average calculation for dice
	// Format: XdY+Z
	if diceNotation == "" {
		return 0
	}

	// TODO: Implement proper dice notation parsing
	// For now, return a placeholder
	return 10.5
}

func (ba *AIBalanceAnalyzer) estimateTargetCount(template *models.RuleTemplate) float64 {
	// Check for area effects
	for _, node := range template.LogicGraph.Nodes {
		if areaSize, ok := node.Properties["area_size"].(float64); ok {
			// Estimate targets based on area
			return math.Ceil(areaSize / 5) // 5ft per target estimate
		}
	}
	return 1.0 // Single target default
}

func (ba *AIBalanceAnalyzer) summarizeSimResults(results []models.SimulationResult) string {
	summary := ""
	for _, result := range results {
		summary += fmt.Sprintf("- %s: %.1f%% success rate\n", result.ScenarioName, result.SuccessRate*100)
	}
	return summary
}

func (ba *AIBalanceAnalyzer) listNodeTypes(graph models.LogicGraph, prefix string) string {
	types := []string{}
	for _, node := range graph.Nodes {
		if strings.HasPrefix(node.Type, prefix) {
			types = append(types, node.SubType)
		}
	}
	return strings.Join(types, ", ")
}

func containsInBalancer(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Additional helper methods for meta prediction
func (ba *AIBalanceAnalyzer) hasSpellProperties(template *models.RuleTemplate) bool {
	return template.Category == "spell" || strings.Contains(template.Description, "spell")
}

func (ba *AIBalanceAnalyzer) getDamageTypes(template *models.RuleTemplate) map[string]bool {
	types := make(map[string]bool)
	for _, node := range template.LogicGraph.Nodes {
		if damageType, ok := node.Properties["damage_type"].(string); ok {
			types[damageType] = true
		}
	}
	return types
}

func (ba *AIBalanceAnalyzer) hasCounterspellProperties(template *models.RuleTemplate) bool {
	for _, node := range template.LogicGraph.Nodes {
		if effect, ok := node.Properties["effect_type"].(string); ok {
			if strings.Contains(effect, "counter") || strings.Contains(effect, "dispel") {
				return true
			}
		}
	}
	return false
}

func (ba *AIBalanceAnalyzer) calculatePopularityScore(template *models.RuleTemplate, simResults []models.SimulationResult) float64 {
	// Base popularity on power, fun factor, and ease of use
	power := template.BalanceMetrics.PowerLevel / 10

	// Fun factor based on variety of effects
	funFactor := float64(len(template.LogicGraph.Nodes)) / 20

	// Ease of use (fewer parameters = easier)
	easeOfUse := 1.0 - (float64(len(template.Parameters)) / 10)

	return (power*0.4 + funFactor*0.3 + easeOfUse*0.3) * 10
}

func (ba *AIBalanceAnalyzer) calculateUniqueness(template *models.RuleTemplate) float64 {
	// Calculate based on unusual node combinations
	nodeTypes := make(map[string]int)
	for _, node := range template.LogicGraph.Nodes {
		nodeTypes[node.Type]++
	}

	// More variety = more unique
	return math.Min(float64(len(nodeTypes))/5, 1.0)
}

func (ba *AIBalanceAnalyzer) requiresConcentration(template *models.RuleTemplate) bool {
	for _, param := range template.Parameters {
		if param.Name == "concentration" {
			return true
		}
	}
	return false
}

func (ba *AIBalanceAnalyzer) grantsAdvantage(template *models.RuleTemplate) bool {
	for _, node := range template.LogicGraph.Nodes {
		if effect, ok := node.Properties["effect"].(string); ok {
			if strings.Contains(effect, "advantage") {
				return true
			}
		}
	}
	return false
}

func (ba *AIBalanceAnalyzer) grantsExtraActions(template *models.RuleTemplate) bool {
	for _, node := range template.LogicGraph.Nodes {
		if node.Type == models.NodeTypeActionResource {
			if resource, ok := node.Properties["resource"].(string); ok {
				return resource == "action" || resource == constants.ActionTypeBonusAction
			}
		}
	}
	return false
}

func (ba *AIBalanceAnalyzer) isStackable(template *models.RuleTemplate) bool {
	// Check if effects can stack
	for _, param := range template.Parameters {
		if param.Name == "stacking" && param.DefaultValue == false {
			return false
		}
	}
	return true
}

func (ba *AIBalanceAnalyzer) runSingleSimulation(ctx context.Context, template *models.RuleTemplate, scenario SimulationScenario) (map[string]interface{}, bool, string) {
	// Placeholder for actual simulation
	// In a real implementation, this would create mock combat scenarios
	// and test the rule's effectiveness

	outcome := map[string]interface{}{
		"damage_dealt":     15.5,
		"actions_used":     1,
		"resources_spent":  1,
		"targets_affected": 1,
	}

	success := true
	edgeCase := ""

	return outcome, success, edgeCase
}

func (ba *AIBalanceAnalyzer) calculateAverageOutcome(outcomes []map[string]interface{}) map[string]interface{} {
	if len(outcomes) == 0 {
		return map[string]interface{}{}
	}

	// Calculate averages for numeric values
	avgOutcome := make(map[string]interface{})
	totals := make(map[string]float64)
	counts := make(map[string]int)

	for _, outcome := range outcomes {
		for key, value := range outcome {
			if num, ok := value.(float64); ok {
				totals[key] += num
				counts[key]++
			}
		}
	}

	for key, total := range totals {
		if count := counts[key]; count > 0 {
			avgOutcome[key] = total / float64(count)
		}
	}

	return avgOutcome
}

func (ba *AIBalanceAnalyzer) compareToBaseline(template *models.RuleTemplate, scenario SimulationScenario, outcome map[string]interface{}) float64 {
	// Compare to baseline abilities at this level
	// This would normally reference a database of baseline abilities

	baselineDamage := float64(scenario.Level) * 5.0 // Simplified baseline
	actualDamage, _ := outcome["damage_dealt"].(float64)

	// Score from 0-10 based on comparison
	ratio := actualDamage / baselineDamage
	score := math.Min(ratio*5, 10)

	return score
}

// SimulationScenario defines a test scenario for balance simulation
type SimulationScenario struct {
	Name              string
	Level             int
	PartySize         int
	EnemyCount        int
	SpecialConditions []string
}
