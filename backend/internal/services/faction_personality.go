package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// FactionPersonalityService manages AI-driven faction personalities
type FactionPersonalityService struct {
	worldRepo   *database.EmergentWorldRepository
	factionRepo *database.WorldBuildingRepository
	llm         LLMProvider
}

// NewFactionPersonalityService creates a new faction personality service
func NewFactionPersonalityService(
	worldRepo *database.EmergentWorldRepository,
	factionRepo *database.WorldBuildingRepository,
	llm LLMProvider,
) *FactionPersonalityService {
	return &FactionPersonalityService{
		worldRepo:   worldRepo,
		factionRepo: factionRepo,
		llm:         llm,
	}
}

// InitializeFactionPersonality creates an AI personality for a faction
func (fps *FactionPersonalityService) InitializeFactionPersonality(ctx context.Context, faction *models.Faction) (*models.FactionPersonality, error) {
	// Generate personality traits based on faction type and description
	traits := fps.generatePersonalityTraits(faction)
	values := fps.generateFactionValues(faction)

	// Use AI to enhance personality
	prompt := fmt.Sprintf(`Analyze this faction and provide personality insights:
Faction: %s
Type: %s
Description: %s
Generated Traits: %v
Generated Values: %v

Provide a JSON response with:
1. "mood": Current faction mood/temperament
2. "decision_style": How they make decisions
3. "communication_style": How they interact with others
4. "core_motivations": What drives them (array of 3-5 items)
5. "fears": What they fear or avoid (array of 2-3 items)
6. "negotiation_approach": Their diplomatic style`,
		faction.Name, faction.Type, faction.Description, traits, values)

	response, err := fps.llm.GenerateContent(ctx, prompt, "You are a D&D faction personality generator focused on creating complex organizational behaviors.")
	if err != nil {
		// Use defaults if AI fails
		response = `{
			"mood": "neutral",
			"decision_style": "pragmatic",
			"communication_style": "formal",
			"core_motivations": ["power", "security", "prosperity"],
			"fears": ["extinction", "irrelevance"],
			"negotiation_approach": "balanced"
		}`
	}

	// Parse AI response
	var aiInsights map[string]interface{}
	_ = json.Unmarshal([]byte(response), &aiInsights)

	personality := &models.FactionPersonality{
		ID:               uuid.New().String(),
		FactionID:        faction.ID.String(),
		Traits:           traits,
		Values:           values,
		Memories:         []models.FactionMemory{},
		CurrentMood:      aiInsights["mood"].(string),
		DecisionWeights:  fps.generateDecisionWeights(traits, values),
		LearningData:     aiInsights,
		LastLearningTime: time.Now(),
	}

	// Save personality
	if err := fps.worldRepo.CreateFactionPersonality(personality); err != nil {
		return nil, err
	}

	return personality, nil
}

// generateRandomMap creates a map with random float64 values for given keys
func generateRandomMap(keys ...string) map[string]float64 {
	result := make(map[string]float64, len(keys))
	for _, key := range keys {
		result[key] = rand.Float64()
	}
	return result
}

// generatePersonalityTraits creates base personality traits
func (fps *FactionPersonalityService) generatePersonalityTraits(faction *models.Faction) map[string]float64 {
	traits := generateRandomMap(
		"aggressive", "diplomatic", "isolationist", "expansionist",
		"traditional", "progressive", "mercantile", "militaristic",
		"scholarly", "religious", "pragmatic", "idealistic",
		"xenophobic", "cosmopolitan", "authoritarian", "libertarian",
	)

	// Adjust based on faction type
	switch faction.Type {
	case models.FactionReligious:
		traits["religious"] += 0.5
		traits["idealistic"] += 0.3
	case models.FactionPolitical:
		traits["traditional"] += 0.3
		traits["authoritarian"] += 0.2
	case models.FactionCriminal:
		traits["aggressive"] += 0.3
		traits["pragmatic"] += 0.4
	case models.FactionMerchant:
		traits["mercantile"] += 0.5
		traits["diplomatic"] += 0.3
	case models.FactionMilitary:
		traits["aggressive"] += 0.5
		traits["militaristic"] += 0.4
	case models.FactionCult:
		traits["religious"] += 0.4
		traits["isolationist"] += 0.3
		traits["idealistic"] += 0.2
	case models.FactionAncientOrder:
		traits["scholarly"] += 0.5
		traits["progressive"] += 0.3
		traits["traditional"] += 0.2
	}

	// Normalize traits
	for k, v := range traits {
		traits[k] = math.Max(0, math.Min(1, v))
	}

	return traits
}

// generateFactionValues creates core values
func (fps *FactionPersonalityService) generateFactionValues(_ *models.Faction) map[string]float64 {
	values := generateRandomMap(
		"honor", "wealth", "knowledge", "power",
		"freedom", "order", "tradition", "innovation",
		"faith", "nature", "justice", "loyalty",
		"independence", "unity", "glory", "survival",
	)

	// Ensure some values are prioritized
	topValues := 3 + rand.Intn(3)
	for i := 0; i < topValues; i++ {
		keys := make([]string, 0, len(values))
		for k := range values {
			keys = append(keys, k)
		}
		selectedKey := keys[rand.Intn(len(keys))]
		values[selectedKey] = 0.7 + rand.Float64()*0.3
	}

	return values
}

// generateDecisionWeights creates weights for different decision factors
func (fps *FactionPersonalityService) generateDecisionWeights(traits, values map[string]float64) map[string]float64 {
	weights := map[string]float64{
		"economic_benefit":    values["wealth"]*0.5 + traits["mercantile"]*0.5,
		"military_advantage":  values["power"]*0.5 + traits["militaristic"]*0.5,
		"diplomatic_gain":     traits["diplomatic"]*0.7 + values["unity"]*0.3,
		"territorial_gain":    traits["expansionist"]*0.8 + values["power"]*0.2,
		"cultural_impact":     values["tradition"]*0.5 + values["innovation"]*0.5,
		"religious_alignment": values["faith"]*0.8 + traits["religious"]*0.2,
		"knowledge_gain":      values["knowledge"]*0.7 + traits["scholarly"]*0.3,
		"security_increase":   values["survival"]*0.6 + values["order"]*0.4,
		"reputation_change":   values["honor"]*0.6 + values["glory"]*0.4,
		"alliance_strength":   values["loyalty"]*0.5 + traits["diplomatic"]*0.5,
	}

	// Normalize weights
	total := 0.0
	for _, w := range weights {
		total += w
	}
	if total > 0 {
		for k := range weights {
			weights[k] /= total
		}
	}

	return weights
}

// RecordMemory adds a significant event to faction memory
func (fps *FactionPersonalityService) RecordMemory(_ context.Context, factionID string, event *models.WorldEvent) error {
	personality, err := fps.worldRepo.GetFactionPersonality(factionID)
	if err != nil {
		return err
	}

	// Determine impact based on event
	impact := fps.calculateEventImpact(personality, event)

	memory := models.FactionMemory{
		ID:           uuid.New().String(),
		EventType:    string(event.Type),
		Description:  event.Description,
		Impact:       impact,
		Participants: []string{},
		Context:      map[string]interface{}{"economicImpacts": event.EconomicImpacts, "politicalImpacts": event.PoliticalImpacts},
		Timestamp:    event.CreatedAt,
		Decay:        0.95, // Memories fade slowly
	}

	personality.Memories = append(personality.Memories, memory)

	// Limit memories to most recent/impactful 100
	if len(personality.Memories) > 100 {
		// Sort by impact and recency
		personality.Memories = fps.pruneMemories(personality.Memories)
	}

	// Update learning data based on memory
	fps.updateLearningFromMemory(personality, memory)

	return fps.worldRepo.UpdateFactionPersonality(personality)
}

// calculateEventImpact determines how much an event affects a faction
func (fps *FactionPersonalityService) calculateEventImpact(personality *models.FactionPersonality, event *models.WorldEvent) float64 {
	impact := 0.0

	// Base impact from event type
	impactMap := map[string]float64{
		"faction_interaction": 0.5,
		"political_milestone": 0.7,
		"economic_event":      0.4,
		"military_conflict":   0.8,
		"diplomatic_success":  0.6,
		"cultural_shift":      0.3,
		"natural_disaster":    0.5,
		"player_action":       0.9,
	}

	if baseImpact, ok := impactMap[string(event.Type)]; ok {
		impact = baseImpact
	} else {
		impact = 0.3
	}

	// Modify based on faction values
	// Note: EconomicImpacts and PoliticalImpacts are JSONB fields
	// Would need proper JSONB handling here
	impact += 0.2 * personality.Values["wealth"]
	impact += 0.2 * personality.Values["power"]

	// Normalize to -1 to 1 range
	impact = math.Max(-1, math.Min(1, impact))

	return impact
}

// MakeFactionDecision uses AI personality to make strategic decisions
func (fps *FactionPersonalityService) MakeFactionDecision(ctx context.Context, factionID string, decision *models.FactionDecision) (*models.FactionDecisionResult, error) {
	if decision == nil {
		return nil, fmt.Errorf("decision cannot be nil")
	}
	personality, err := fps.worldRepo.GetFactionPersonality(factionID)
	if err != nil {
		return nil, err
	}

	factionUUID, err := uuid.Parse(factionID)
	if err != nil {
		return nil, fmt.Errorf("invalid faction ID: %w", err)
	}

	faction, err := fps.factionRepo.GetFaction(factionUUID)
	if err != nil {
		return nil, err
	}

	// Analyze options based on personality
	optionScores := make(map[string]float64)
	for _, option := range decision.Options {
		score := fps.scoreOption(personality, option)
		optionScores[option.ID] = score
	}

	// Get recent memories that might influence decision
	relevantMemories := fps.getRelevantMemories(personality, decision)

	// Use AI to make nuanced decision
	prompt := fmt.Sprintf(`A faction must make a strategic decision:

Faction: %s
Personality Traits: %v
Core Values: %v
Current Mood: %s
Decision Weights: %v

Decision Context: %s

Options:
%s

Recent Relevant Memories:
%s

Based on this faction's personality and history, which option would they choose and why? 
Provide a JSON response with:
1. "chosen_option": The ID of the chosen option
2. "reasoning": A brief explanation (2-3 sentences) of why this choice aligns with their personality
3. "confidence": How certain they are (0.0-1.0)
4. "mood_change": How this decision affects their mood
5. "relationship_impacts": Map of faction_id to relationship change (-10 to +10)`,
		faction.Name,
		personality.Traits,
		personality.Values,
		personality.CurrentMood,
		personality.DecisionWeights,
		decision.Context,
		fps.formatOptions(decision.Options, optionScores),
		fps.formatMemories(relevantMemories))

	response, err := fps.llm.GenerateContent(ctx, prompt, "You are a D&D faction decision-making AI that analyzes faction personalities and makes strategic choices.")
	if err != nil {
		// Fallback to highest scored option
		return fps.makeDefaultDecision(personality, decision, optionScores), nil
	}

	// Parse AI response
	var aiDecision map[string]interface{}
	if err := json.Unmarshal([]byte(response), &aiDecision); err != nil {
		return fps.makeDefaultDecision(personality, decision, optionScores), nil
	}

	// Update personality based on decision
	if moodChange, ok := aiDecision["mood_change"].(string); ok {
		personality.CurrentMood = moodChange
	}

	// Record decision as memory
	decisionMemory := models.FactionMemory{
		ID:           uuid.New().String(),
		EventType:    "strategic_decision",
		Description:  fmt.Sprintf("Chose: %s - %s", aiDecision["chosen_option"], aiDecision["reasoning"]),
		Impact:       0.5,
		Participants: []string{factionID},
		Context:      map[string]interface{}{"decision": decision, "result": aiDecision},
		Timestamp:    time.Now(),
		Decay:        0.9,
	}
	personality.Memories = append(personality.Memories, decisionMemory)
	_ = fps.worldRepo.UpdateFactionPersonality(personality)

	chosenOption := ""
	if opt, ok := aiDecision["chosen_option"].(string); ok {
		chosenOption = opt
	}
	reasoning := ""
	if r, ok := aiDecision["reasoning"].(string); ok {
		reasoning = r
	}

	return &models.FactionDecisionResult{
		DecisionID:    decision.ID,
		Success:       true,
		Consequences:  []string{fmt.Sprintf("Chose option: %s", chosenOption)},
		ImpactMetrics: map[string]interface{}{"chosen_option": chosenOption, "reasoning": reasoning},
		NextActions:   []string{},
	}, nil
}

// scoreOption calculates how well an option aligns with faction personality
func (fps *FactionPersonalityService) scoreOption(personality *models.FactionPersonality, option models.DecisionOption) float64 {
	score := 0.0

	// Score based on benefits
	score += float64(len(option.Benefits)) * 0.5

	// Adjust based on risks
	riskLevel := float64(len(option.Risks))
	// Risk-averse personalities penalize high-risk options
	if personality.Traits["pragmatic"] > 0.6 {
		score -= riskLevel * 0.3
	}
	// Risk-taking personalities favor high-risk options
	if personality.Traits["aggressive"] > 0.6 {
		score += riskLevel * 0.1
	}

	// Consider requirements
	requirementsPenalty := 0.0
	for _, value := range option.Requirements {
		// Penalize if requirements are high relative to personality
		if reqValue, ok := value.(float64); ok {
			requirementsPenalty += reqValue * 0.1
		}
	}
	score -= requirementsPenalty

	return score
}

// getRelevantMemories retrieves memories that might influence a decision
func (fps *FactionPersonalityService) getRelevantMemories(personality *models.FactionPersonality, decision *models.FactionDecision) []models.FactionMemory {
	if decision == nil {
		return []models.FactionMemory{}
	}

	relevant := []models.FactionMemory{}
	for _, memory := range personality.Memories {
		if fps.isMemoryRelevant(memory, decision) {
			relevant = append(relevant, memory)
		}
	}

	// Return most impactful memories
	if len(relevant) > 5 {
		relevant = relevant[:5]
	}

	return relevant
}

// isMemoryRelevant checks if a memory is relevant to a decision
func (fps *FactionPersonalityService) isMemoryRelevant(memory models.FactionMemory, decision *models.FactionDecision) bool {
	// Check memory decay first
	if !fps.isMemoryActive(memory) {
		return false
	}

	// Check faction involvement
	if fps.isFactionInvolved(memory, decision.FactionID) {
		return true
	}

	// Check event type relevance
	return fps.isEventTypeRelevant(memory.EventType, decision.DecisionType)
}

// isMemoryActive checks if a memory has decayed too much
func (fps *FactionPersonalityService) isMemoryActive(memory models.FactionMemory) bool {
	age := time.Since(memory.Timestamp).Hours() / 24.0 // Days
	decayedImpact := memory.Impact * math.Pow(memory.Decay, age)
	return decayedImpact >= 0.1
}

// isFactionInvolved checks if a faction is a participant in a memory
func (fps *FactionPersonalityService) isFactionInvolved(memory models.FactionMemory, factionID string) bool {
	for _, participant := range memory.Participants {
		if participant == factionID {
			return true
		}
	}
	return false
}

// isEventTypeRelevant checks if an event type matches the decision type
func (fps *FactionPersonalityService) isEventTypeRelevant(eventType, decisionType string) bool {
	switch decisionType {
	case constants.ApproachDiplomatic:
		return eventType == "faction_interaction"
	case constants.ApproachMilitary:
		return eventType == "military_conflict"
	default:
		return false
	}
}

// UpdateFactionMood adjusts faction mood based on recent events
func (fps *FactionPersonalityService) UpdateFactionMood(_ context.Context, factionID string) error {
	personality, err := fps.worldRepo.GetFactionPersonality(factionID)
	if err != nil {
		return err
	}

	// Calculate mood from recent memories
	recentImpact := 0.0
	memoryCount := 0

	for _, memory := range personality.Memories {
		age := time.Since(memory.Timestamp).Hours() / 24.0
		if age < 30 { // Last 30 days
			decayedImpact := memory.Impact * math.Pow(memory.Decay, age)
			recentImpact += decayedImpact
			memoryCount++
		}
	}

	if memoryCount > 0 {
		avgImpact := recentImpact / float64(memoryCount)

		// Determine new mood based on average impact
		if avgImpact > 0.5 {
			personality.CurrentMood = "triumphant"
		} else if avgImpact > 0.2 {
			personality.CurrentMood = "confident"
		} else if avgImpact > -0.2 {
			personality.CurrentMood = "cautious"
		} else if avgImpact > -0.5 {
			personality.CurrentMood = "worried"
		} else {
			personality.CurrentMood = "desperate"
		}
	}

	return fps.worldRepo.UpdateFactionPersonality(personality)
}

// LearnFromInteraction updates faction personality based on player interactions
func (fps *FactionPersonalityService) LearnFromInteraction(_ context.Context, factionID string, interaction models.PlayerInteraction) error {
	personality, err := fps.worldRepo.GetFactionPersonality(factionID)
	if err != nil {
		return err
	}

	// Update learning data
	if personality.LearningData == nil {
		personality.LearningData = make(map[string]interface{})
	}

	// Track interaction patterns
	// Extract outcome from context
	outcome := outcomeNeutral
	if outcomeVal, ok := interaction.Context["outcome"].(string); ok {
		outcome = outcomeVal
	}

	interactions, _ := personality.LearningData["player_interactions"].([]interface{})
	interactions = append(interactions, map[string]interface{}{
		"type":      interaction.Type,
		"outcome":   outcome,
		"timestamp": time.Now(),
	})

	// Keep last 50 interactions
	if len(interactions) > 50 {
		interactions = interactions[len(interactions)-50:]
	}
	personality.LearningData["player_interactions"] = interactions

	// Adjust personality traits based on successful interactions
	if outcome == "positive" {
		switch interaction.Type {
		case constants.ApproachDiplomatic:
			personality.Traits["diplomatic"] = math.Min(1.0, personality.Traits["diplomatic"]+0.02)
			personality.Traits["aggressive"] = math.Max(0.0, personality.Traits["aggressive"]-0.01)
		case constants.ActionTrade:
			personality.Traits["mercantile"] = math.Min(1.0, personality.Traits["mercantile"]+0.02)
		case constants.ApproachMilitary:
			personality.Traits["militaristic"] = math.Min(1.0, personality.Traits["militaristic"]+0.02)
		}
	}

	personality.LastLearningTime = time.Now()

	return fps.worldRepo.UpdateFactionPersonality(personality)
}

// Helper functions

func (fps *FactionPersonalityService) pruneMemories(memories []models.FactionMemory) []models.FactionMemory {
	// Keep most impactful and recent memories
	// Simple implementation - in production would use better algorithm
	if len(memories) <= 100 {
		return memories
	}

	// Keep last 50 and top 50 by impact
	recent := memories[len(memories)-50:]

	// Sort remaining by impact (simplified)
	remaining := memories[:len(memories)-50]
	// Would implement proper sorting here

	combined := append(remaining[:50], recent...)
	return combined
}

func (fps *FactionPersonalityService) updateLearningFromMemory(personality *models.FactionPersonality, memory models.FactionMemory) {
	// Update learning patterns based on memory type and impact
	patterns, _ := personality.LearningData["event_patterns"].(map[string]interface{})
	if patterns == nil {
		patterns = make(map[string]interface{})
	}

	// Track event outcomes
	eventOutcomes, _ := patterns[memory.EventType].([]float64)
	eventOutcomes = append(eventOutcomes, memory.Impact)

	// Keep last 20 outcomes per event type
	if len(eventOutcomes) > 20 {
		eventOutcomes = eventOutcomes[len(eventOutcomes)-20:]
	}

	patterns[memory.EventType] = eventOutcomes
	personality.LearningData["event_patterns"] = patterns
}

func (fps *FactionPersonalityService) formatOptions(options []models.DecisionOption, scores map[string]float64) string {
	result := ""
	for _, opt := range options {
		result += fmt.Sprintf("- Option %s: %s (Score: %.2f)\n  Benefits: %v\n  Risks: %v\n\n",
			opt.ID, opt.Description, scores[opt.ID], opt.Benefits, opt.Risks)
	}
	return result
}

func (fps *FactionPersonalityService) formatMemories(memories []models.FactionMemory) string {
	if len(memories) == 0 {
		return "No relevant recent memories"
	}

	result := ""
	for _, mem := range memories {
		age := time.Since(mem.Timestamp).Hours() / 24.0
		result += fmt.Sprintf("- %s (%.0f days ago, Impact: %.2f): %s\n",
			mem.EventType, age, mem.Impact, mem.Description)
	}
	return result
}

func (fps *FactionPersonalityService) makeDefaultDecision(_ *models.FactionPersonality, decision *models.FactionDecision, scores map[string]float64) *models.FactionDecisionResult {
	if decision == nil {
		return &models.FactionDecisionResult{
			Success:      false,
			Consequences: []string{"Invalid decision input"},
		}
	}
	// Find highest scored option
	var bestOption string
	bestScore := -999.0
	for id, score := range scores {
		if score > bestScore {
			bestScore = score
			bestOption = id
		}
	}

	return &models.FactionDecisionResult{
		DecisionID:   decision.ID,
		Success:      true,
		Consequences: []string{fmt.Sprintf("Chose option: %s", bestOption)},
		ImpactMetrics: map[string]interface{}{
			"chosen_option": bestOption,
			"reasoning":     "Chose option that best aligns with faction values and goals",
			"confidence":    0.7,
		},
		NextActions: []string{},
	}
}

//lint:ignore U1000 retained for future feature work
func (fps *FactionPersonalityService) parseRelationshipImpacts(impacts interface{}) map[string]float64 {
	result := make(map[string]float64)

	if impactMap, ok := impacts.(map[string]interface{}); ok {
		for k, v := range impactMap {
			if val, ok := v.(float64); ok {
				result[k] = val
			}
		}
	}

	return result
}

// Additional types for decision making

type FactionDecision struct {
	ID               string           `json:"id"`
	Type             string           `json:"type"`
	Context          string           `json:"context"`
	Options          []DecisionOption `json:"options"`
	InvolvedEntities []string         `json:"involved_entities"`
	Deadline         *time.Time       `json:"deadline,omitempty"`
}

type DecisionOption struct {
	ID           string                 `json:"id"`
	Description  string                 `json:"description"`
	Outcomes     map[string]interface{} `json:"outcomes"`
	Requirements map[string]interface{} `json:"requirements"`
	Risks        map[string]interface{} `json:"risks"`
}

type FactionDecisionResult struct {
	ChosenOptionID      string             `json:"chosen_option_id"`
	Reasoning           string             `json:"reasoning"`
	Confidence          float64            `json:"confidence"`
	RelationshipImpacts map[string]float64 `json:"relationship_impacts"`
	Timestamp           time.Time          `json:"timestamp"`
}

type PlayerInteraction struct {
	Type    string                 `json:"type"`
	Outcome string                 `json:"outcome"`
	Details map[string]interface{} `json:"details"`
}
