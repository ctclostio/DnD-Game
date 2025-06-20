package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/config"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// NarrativeEngine orchestrates the AI-powered dynamic storytelling system
type NarrativeEngine struct {
	llm               LLMProvider
	cfg               *config.Config
	ProfileService    *PlayerProfileService
	ConsequenceEngine *ConsequenceEngine
	PerspectiveGen    *PerspectiveGenerator
}

// PlayerProfileService manages player narrative preferences and patterns
type PlayerProfileService struct {
	llm LLMProvider
	cfg *config.Config
}

// ConsequenceEngine calculates and manages ripple effects from player actions
type ConsequenceEngine struct {
	llm LLMProvider
	cfg *config.Config
}

// PerspectiveGenerator creates multiple viewpoints of the same events
type PerspectiveGenerator struct {
	llm LLMProvider
	cfg *config.Config
}

// NewNarrativeEngine creates a new narrative engine instance
func NewNarrativeEngine(cfg *config.Config) (*NarrativeEngine, error) {
	llm := NewLLMProvider(AIConfig{
		Provider: cfg.AI.Provider,
		APIKey:   cfg.AI.APIKey,
		Model:    cfg.AI.Model,
		Enabled:  cfg.AI.Enabled,
	})

	return &NarrativeEngine{
		llm:               llm,
		cfg:               cfg,
		ProfileService:    &PlayerProfileService{llm: llm, cfg: cfg},
		ConsequenceEngine: &ConsequenceEngine{llm: llm, cfg: cfg},
		PerspectiveGen:    &PerspectiveGenerator{llm: llm, cfg: cfg},
	}, nil
}

// AnalyzePlayerDecision updates player profile based on a decision
func (ps *PlayerProfileService) AnalyzePlayerDecision(ctx context.Context, profile *models.NarrativeProfile, decision models.DecisionRecord) (*models.NarrativeProfile, error) {
	if !ps.cfg.AI.Enabled {
		// Simple analysis without AI
		profile.DecisionHistory = append(profile.DecisionHistory, decision)
		return profile, nil
	}

	prompt := fmt.Sprintf(`Analyze this player decision and update their narrative profile:

Current Profile:
- Play Style: %s
- Preferences: %+v
- Recent Decisions: %d recorded

New Decision:
- Context: %s
- Choice Made: %s
- Alternatives: %v
- Consequences: %v

Based on this decision:
1. What does this reveal about the player's storytelling preferences?
2. Are there patterns emerging in their decision-making?
3. What themes resonate with this player?
4. How should future narratives be tailored for maximum engagement?

Provide analysis in JSON format with:
- updated_themes: array of story themes this player enjoys
- updated_tone: array of narrative tones they prefer
- moral_tendency: how they approach moral choices
- engagement_triggers: what hooks them into stories
- play_style_update: any change to their play style classification`,
		profile.PlayStyle,
		profile.Preferences,
		len(profile.DecisionHistory),
		decision.Context,
		decision.Decision,
		decision.Alternatives,
		decision.Consequences,
	)

	response, err := ps.llm.GenerateContent(ctx, prompt, "You are a D&D narrative assistant focusing on player psychology and storytelling.")
	if err != nil {
		return profile, fmt.Errorf("failed to analyze decision: %w", err)
	}

	// Parse AI response and update profile
	var analysis struct {
		UpdatedThemes      []string `json:"updated_themes"`
		UpdatedTone        []string `json:"updated_tone"`
		MoralTendency      string   `json:"moral_tendency"`
		EngagementTriggers []string `json:"engagement_triggers"`
		PlayStyleUpdate    string   `json:"play_style_update"`
	}

	if err := json.Unmarshal([]byte(response), &analysis); err == nil {
		// Update preferences
		profile.Preferences.Themes = mergeUnique(profile.Preferences.Themes, analysis.UpdatedThemes)
		profile.Preferences.Tone = mergeUnique(profile.Preferences.Tone, analysis.UpdatedTone)
		profile.Preferences.MoralAlignment = analysis.MoralTendency

		if analysis.PlayStyleUpdate != "" {
			profile.PlayStyle = analysis.PlayStyleUpdate
		}

		// Update analytics
		if profile.Analytics == nil {
			profile.Analytics = make(map[string]interface{})
		}
		profile.Analytics["engagement_triggers"] = analysis.EngagementTriggers
		profile.Analytics["last_analysis"] = time.Now()
	}

	// Add decision to history
	profile.DecisionHistory = append(profile.DecisionHistory, decision)
	profile.UpdatedAt = time.Now()

	return profile, nil
}

// GeneratePersonalizedNarrative creates a narrative tailored to a specific player
func (ne *NarrativeEngine) GeneratePersonalizedNarrative(ctx context.Context, baseEvent *models.NarrativeEvent, profile *models.NarrativeProfile, backstory []models.BackstoryElement) (*models.PersonalizedNarrative, error) {
	if baseEvent == nil {
		return nil, fmt.Errorf("base event cannot be nil")
	}
	if !ne.cfg.AI.Enabled {
		return ne.createBasicNarrative(baseEvent, profile)
	}

	// Select relevant backstory elements
	relevantBackstory := ne.selectRelevantBackstory(backstory, baseEvent)

	// Generate AI response
	response, err := ne.generateNarrativePrompt(ctx, baseEvent, profile, relevantBackstory)
	if err != nil {
		return nil, err
	}

	// Parse response
	narrativeData, err := ne.parseNarrativeResponse(response)
	if err != nil {
		return nil, err
	}

	// Build and return narrative
	return ne.buildPersonalizedNarrative(narrativeData, baseEvent, profile)
}

func (ne *NarrativeEngine) createBasicNarrative(baseEvent *models.NarrativeEvent, profile *models.NarrativeProfile) (*models.PersonalizedNarrative, error) {
	return &models.PersonalizedNarrative{
		ID:          uuid.New().String(),
		BaseEventID: baseEvent.ID,
		CharacterID: profile.CharacterID,
		GeneratedAt: time.Now(),
	}, nil
}

func (ne *NarrativeEngine) generateNarrativePrompt(ctx context.Context, baseEvent *models.NarrativeEvent, profile *models.NarrativeProfile, relevantBackstory []models.BackstoryElement) (string, error) {
	prompt := fmt.Sprintf(`Create a personalized narrative for this player based on their profile and the current event:

Base Event:
- Type: %s
- Description: %s
- Location: %s
- Key Participants: %v

Player Profile:
- Preferred Themes: %v
- Preferred Tone: %v
- Play Style: %s
- Moral Alignment: %s
- Recent Decisions: %d

Relevant Backstory Elements:
%s

Create a personalized version of this event that:
1. Incorporates elements from their backstory naturally
2. Presents choices aligned with their play style
3. Uses their preferred narrative tone
4. Creates hooks based on their demonstrated preferences
5. Weaves in consequences from their past decisions where relevant

Provide the narrative in JSON format with:
- personalized_description: the event description tailored to this player
- narrative_hooks: array of {type, content, relevance, backstory_id}
- backstory_callbacks: array of {backstory_element_id, integration_type, narrative_text, subtlety}
- moral_choices: array of choices that align with their tendencies
- emotional_resonance: float 0-1
- predicted_engagement: float 0-1`,
		baseEvent.Type,
		baseEvent.Description,
		baseEvent.Location,
		baseEvent.Participants,
		profile.Preferences.Themes,
		profile.Preferences.Tone,
		profile.PlayStyle,
		profile.Preferences.MoralAlignment,
		len(profile.DecisionHistory),
		formatBackstoryElements(relevantBackstory),
	)

	response, err := ne.llm.GenerateContent(ctx, prompt, "You are a D&D narrative engine that creates dynamic storylines based on player actions.")
	if err != nil {
		return "", fmt.Errorf("failed to generate personalized narrative: %w", err)
	}

	return response, nil
}

type narrativeResponseData struct {
	PersonalizedDescription string `json:"personalized_description"`
	NarrativeHooks          []struct {
		Type        string  `json:"type"`
		Content     string  `json:"content"`
		Relevance   float64 `json:"relevance"`
		BackstoryID string  `json:"backstory_id"`
	} `json:"narrative_hooks"`
	BackstoryCallbacks []struct {
		BackstoryElementID string `json:"backstory_element_id"`
		IntegrationType    string `json:"integration_type"`
		NarrativeText      string `json:"narrative_text"`
		Subtlety           int    `json:"subtlety"`
	} `json:"backstory_callbacks"`
	MoralChoices        []string `json:"moral_choices"`
	EmotionalResonance  float64  `json:"emotional_resonance"`
	PredictedEngagement float64  `json:"predicted_engagement"`
}

func (ne *NarrativeEngine) parseNarrativeResponse(response string) (*narrativeResponseData, error) {
	var narrativeData narrativeResponseData
	if err := json.Unmarshal([]byte(response), &narrativeData); err != nil {
		return nil, fmt.Errorf("failed to parse narrative data: %w", err)
	}
	return &narrativeData, nil
}

func (ne *NarrativeEngine) buildPersonalizedNarrative(data *narrativeResponseData, baseEvent *models.NarrativeEvent, profile *models.NarrativeProfile) (*models.PersonalizedNarrative, error) {
	narrative := &models.PersonalizedNarrative{
		ID:                 uuid.New().String(),
		BaseEventID:        baseEvent.ID,
		CharacterID:        profile.CharacterID,
		EmotionalResonance: data.EmotionalResonance,
		GeneratedAt:        time.Now(),
		Metadata: map[string]interface{}{
			"personalized_description": data.PersonalizedDescription,
			"moral_choices":            data.MoralChoices,
			"predicted_engagement":     data.PredictedEngagement,
		},
	}

	// Convert hooks
	for _, hook := range data.NarrativeHooks {
		narrative.PersonalizedHooks = append(narrative.PersonalizedHooks, models.NarrativeHook{
			Type:        hook.Type,
			Content:     hook.Content,
			Relevance:   hook.Relevance,
			BackstoryID: hook.BackstoryID,
		})
	}

	// Convert backstory callbacks
	for _, callback := range data.BackstoryCallbacks {
		narrative.BackstoryCallbacks = append(narrative.BackstoryCallbacks, models.BackstoryIntegration{
			BackstoryElementID: callback.BackstoryElementID,
			IntegrationType:    callback.IntegrationType,
			NarrativeText:      callback.NarrativeText,
			Subtlety:           callback.Subtlety,
		})
	}

	// Calculate predicted impacts
	narrative.PredictedImpact = ne.calculatePredictedImpacts(profile, data.EmotionalResonance)

	return narrative, nil
}

// CalculateConsequences determines the ripple effects of a player action
func (ce *ConsequenceEngine) CalculateConsequences(ctx context.Context, action models.PlayerAction, worldState map[string]interface{}) ([]models.ConsequenceEvent, error) {
	if !ce.cfg.AI.Enabled {
		return []models.ConsequenceEvent{}, nil
	}

	// Generate AI response
	response, err := ce.generateConsequencePrompt(ctx, action, worldState)
	if err != nil {
		return nil, err
	}

	// Parse response
	consequenceData, err := ce.parseConsequenceResponse(response)
	if err != nil {
		return nil, err
	}

	// Build consequences
	consequences := ce.buildConsequenceEvents(consequenceData, action)

	// Sort by priority
	ce.sortConsequencesByPriority(consequences)

	return consequences, nil
}

func (ce *ConsequenceEngine) generateConsequencePrompt(ctx context.Context, action models.PlayerAction, worldState map[string]interface{}) (string, error) {
	prompt := fmt.Sprintf(`Analyze this player action and determine its consequences across the game world:

Action Details:
- Type: %s
- Target: %s (ID: %s)
- Description: %s
- Moral Weight: %s
- Immediate Result: %s

World State Context:
%s

Generate a cascade of consequences considering:
1. Immediate effects (within the session)
2. Short-term ripples (next few sessions)  
3. Medium-term changes (weeks of game time)
4. Long-term impacts (months/years of game time)

Consider effects on:
- NPCs and their relationships
- Factions and politics
- Economic systems
- Geographic/environmental changes
- Future plot opportunities
- Other player characters

For each consequence, assess:
- Severity (1-10 scale)
- Probability of occurrence
- What could prevent or mitigate it
- Secondary effects it might trigger

Provide consequences in JSON format as an array of:
{
  "description": "what happens",
  "trigger_type": "type of consequence",
  "severity": 1-10,
  "delay": "immediate/short/medium/long",
  "affected_entities": [{"entity_type", "entity_id", "entity_name", "impact_type", "impact_severity", "description"}],
  "cascade_effects": [{"type", "description", "probability", "timeline"}],
  "prevention_methods": ["ways to prevent or mitigate"]
}`,
		action.ActionType,
		action.TargetType,
		action.TargetID,
		action.ActionDescription,
		action.MoralWeight,
		action.ImmediateResult,
		formatWorldState(worldState),
	)

	response, err := ce.llm.GenerateContent(ctx, prompt, "You are a D&D narrative assistant that analyzes and tracks story consequences.")
	if err != nil {
		return "", fmt.Errorf("failed to calculate consequences: %w", err)
	}

	return response, nil
}

type consequenceResponseData struct {
	Description      string `json:"description"`
	TriggerType      string `json:"trigger_type"`
	Severity         int    `json:"severity"`
	Delay            string `json:"delay"`
	AffectedEntities []struct {
		EntityType     string `json:"entity_type"`
		EntityID       string `json:"entity_id"`
		EntityName     string `json:"entity_name"`
		ImpactType     string `json:"impact_type"`
		ImpactSeverity int    `json:"impact_severity"`
		Description    string `json:"description"`
	} `json:"affected_entities"`
	CascadeEffects []struct {
		Type        string  `json:"type"`
		Description string  `json:"description"`
		Probability float64 `json:"probability"`
		Timeline    string  `json:"timeline"`
	} `json:"cascade_effects"`
	PreventionMethods []string `json:"prevention_methods"`
}

func (ce *ConsequenceEngine) parseConsequenceResponse(response string) ([]consequenceResponseData, error) {
	var consequenceData []consequenceResponseData
	if err := json.Unmarshal([]byte(response), &consequenceData); err != nil {
		return nil, fmt.Errorf("failed to parse consequence data: %w", err)
	}
	return consequenceData, nil
}

func (ce *ConsequenceEngine) buildConsequenceEvents(dataList []consequenceResponseData, action models.PlayerAction) []models.ConsequenceEvent {
	consequences := make([]models.ConsequenceEvent, 0, len(dataList))
	
	for i := range dataList {
		consequence := ce.buildSingleConsequence(&dataList[i], action)
		consequences = append(consequences, consequence)
	}
	
	return consequences
}

func (ce *ConsequenceEngine) buildSingleConsequence(data *consequenceResponseData, action models.PlayerAction) models.ConsequenceEvent {
	consequence := models.ConsequenceEvent{
		ID:              uuid.New().String(),
		TriggerActionID: action.ID,
		TriggerType:     data.TriggerType,
		Description:     data.Description,
		Severity:        data.Severity,
		Delay:           data.Delay,
		Status:          "pending",
		CreatedAt:       time.Now(),
		Metadata: map[string]interface{}{
			"prevention_methods": data.PreventionMethods,
		},
	}

	// Convert affected entities
	for _, entity := range data.AffectedEntities {
		consequence.AffectedEntities = append(consequence.AffectedEntities, models.AffectedEntity{
			EntityType:     entity.EntityType,
			EntityID:       entity.EntityID,
			EntityName:     entity.EntityName,
			ImpactType:     entity.ImpactType,
			ImpactSeverity: entity.ImpactSeverity,
			Description:    entity.Description,
		})
	}

	// Convert cascade effects
	for _, cascade := range data.CascadeEffects {
		consequence.CascadeEffects = append(consequence.CascadeEffects, models.CascadeEffect{
			ID:          uuid.New().String(),
			Type:        cascade.Type,
			Description: cascade.Description,
			Probability: cascade.Probability,
			Timeline:    cascade.Timeline,
			Triggered:   false,
		})
	}

	return consequence
}

func (ce *ConsequenceEngine) sortConsequencesByPriority(consequences []models.ConsequenceEvent) {
	sort.Slice(consequences, func(i, j int) bool {
		if consequences[i].Delay != consequences[j].Delay {
			return getDelayPriority(consequences[i].Delay) < getDelayPriority(consequences[j].Delay)
		}
		return consequences[i].Severity > consequences[j].Severity
	})
}

// GenerateMultiplePerspectives creates different viewpoints of the same event
func (pg *PerspectiveGenerator) GenerateMultiplePerspectives(ctx context.Context, event *models.NarrativeEvent, sources []models.PerspectiveSource) ([]models.PerspectiveNarrative, error) {
	if event == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}
	if !pg.cfg.AI.Enabled {
		return pg.createNeutralPerspective(event), nil
	}

	perspectives := make([]models.PerspectiveNarrative, 0, len(sources))

	for _, source := range sources {
		perspective, err := pg.generateSinglePerspective(ctx, event, source)
		if err != nil {
			continue // Skip this perspective on error
		}
		if perspective != nil {
			perspectives = append(perspectives, *perspective)
		}
	}

	// Find contradictions between perspectives
	pg.findContradictions(perspectives)

	return perspectives, nil
}

func (pg *PerspectiveGenerator) createNeutralPerspective(event *models.NarrativeEvent) []models.PerspectiveNarrative {
	return []models.PerspectiveNarrative{{
		ID:              uuid.New().String(),
		EventID:         event.ID,
		PerspectiveType: "neutral",
		SourceName:      "Observer",
		Narrative:       event.Description,
		Bias:            "neutral",
		TruthLevel:      1.0,
		CreatedAt:       time.Now(),
	}}
}

func (pg *PerspectiveGenerator) generateSinglePerspective(ctx context.Context, event *models.NarrativeEvent, source models.PerspectiveSource) (*models.PerspectiveNarrative, error) {
	// Generate prompt
	prompt := pg.buildPerspectivePrompt(event, source)

	// Get AI response
	response, err := pg.llm.GenerateContent(ctx, prompt, "You are a D&D perspective generator that creates authentic character viewpoints.")
	if err != nil {
		return nil, err
	}

	// Parse response
	perspectiveData, err := pg.parsePerspectiveResponse(response)
	if err != nil {
		return nil, err
	}

	// Build perspective
	return pg.buildPerspective(perspectiveData, event, source), nil
}

func (pg *PerspectiveGenerator) buildPerspectivePrompt(event *models.NarrativeEvent, source models.PerspectiveSource) string {
	return fmt.Sprintf(`Generate a perspective on this event from a specific viewpoint:

Event Details:
- Type: %s
- Description: %s
- Location: %s
- Participants: %v
- Immediate Effects: %v

Perspective Source:
- Type: %s
- Name: %s
- Background: %s
- Motivations: %v
- Relationships: %v
- Cultural Context: %v

Generate a narrative that:
1. Tells the event from this source's perspective
2. Reflects their biases, beliefs, and motivations
3. May omit, emphasize, or misinterpret details based on their viewpoint
4. Includes cultural and personal filters
5. Shows how their relationships color their interpretation

Consider:
- What would they focus on?
- What might they not notice or understand?
- How would their goals affect their telling?
- What "spin" would they put on events?
- What details might they hide or fabricate?

Provide the perspective in JSON format:
{
  "narrative": "the event as told by this source",
  "bias": "positive/negative/neutral/conflicted",
  "truth_level": 0.0-1.0,
  "emphasized_details": ["details they focus on"],
  "omitted_details": ["details they ignore or hide"],
  "misinterpretations": ["things they get wrong"],
  "hidden_agenda": "what they're really thinking",
  "emotional_tone": "how they feel about it",
  "cultural_filters": ["cultural biases affecting their view"]
}`,
		event.Type,
		event.Description,
		event.Location,
		event.Participants,
		event.ImmediateEffects,
		source.Type,
		source.Name,
		source.Background,
		source.Motivations,
		source.Relationships,
		source.CulturalContext,
	)
}

type perspectiveResponseData struct {
	Narrative          string   `json:"narrative"`
	Bias               string   `json:"bias"`
	TruthLevel         float64  `json:"truth_level"`
	EmphasizedDetails  []string `json:"emphasized_details"`
	OmittedDetails     []string `json:"omitted_details"`
	Misinterpretations []string `json:"misinterpretations"`
	HiddenAgenda       string   `json:"hidden_agenda"`
	EmotionalTone      string   `json:"emotional_tone"`
	CulturalFilters    []string `json:"cultural_filters"`
}

func (pg *PerspectiveGenerator) parsePerspectiveResponse(response string) (*perspectiveResponseData, error) {
	var data perspectiveResponseData
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (pg *PerspectiveGenerator) buildPerspective(data *perspectiveResponseData, event *models.NarrativeEvent, source models.PerspectiveSource) *models.PerspectiveNarrative {
	return &models.PerspectiveNarrative{
		ID:              uuid.New().String(),
		EventID:         event.ID,
		PerspectiveType: source.Type,
		SourceID:        source.ID,
		SourceName:      source.Name,
		Narrative:       data.Narrative,
		Bias:            data.Bias,
		TruthLevel:      data.TruthLevel,
		EmotionalTone:   data.EmotionalTone,
		HiddenDetails:   append(data.OmittedDetails, data.HiddenAgenda),
		CreatedAt:       time.Now(),
		CulturalContext: map[string]interface{}{
			"filters":       data.CulturalFilters,
			"emphasized":    data.EmphasizedDetails,
			"misunderstood": data.Misinterpretations,
		},
	}
}

// Helper functions

func (ne *NarrativeEngine) selectRelevantBackstory(backstory []models.BackstoryElement, event *models.NarrativeEvent) []models.BackstoryElement {
	if event == nil {
		return []models.BackstoryElement{}
	}

	relevant := ne.scoreBackstoryElements(backstory, event)
	return ne.selectTopBackstoryElements(relevant, 3)
}

func (ne *NarrativeEngine) scoreBackstoryElements(backstory []models.BackstoryElement, event *models.NarrativeEvent) []models.BackstoryElement {
	relevant := make([]models.BackstoryElement, 0)
	eventTags := extractEventTags(event)

	for i := range backstory {
		element := &backstory[i]
		if ne.shouldSkipBackstoryElement(element) {
			continue
		}

		relevanceScore := ne.calculateRelevanceScore(element, eventTags)
		if relevanceScore > 0 {
			element.Weight = relevanceScore
			relevant = append(relevant, *element)
		}
	}

	return relevant
}

func (ne *NarrativeEngine) shouldSkipBackstoryElement(element *models.BackstoryElement) bool {
	return element.Used && element.UsageCount > 2
}

func (ne *NarrativeEngine) calculateRelevanceScore(element *models.BackstoryElement, eventTags []string) float64 {
	score := 0.0
	for _, tag := range element.Tags {
		for _, eventTag := range eventTags {
			if strings.Contains(strings.ToLower(tag), strings.ToLower(eventTag)) {
				score += 1.0
			}
		}
	}
	return score
}

func (ne *NarrativeEngine) selectTopBackstoryElements(elements []models.BackstoryElement, limit int) []models.BackstoryElement {
	// Sort by relevance
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Weight > elements[j].Weight
	})

	if len(elements) > limit {
		return elements[:limit]
	}
	return elements
}

func (ne *NarrativeEngine) calculatePredictedImpacts(profile *models.NarrativeProfile, resonance float64) []models.PredictedImpact {
	impacts := []models.PredictedImpact{
		{
			Type:        "emotional",
			Description: "Strong emotional engagement based on backstory integration",
			Likelihood:  resonance,
			Magnitude:   resonance * 0.8,
		},
	}

	// Add more impacts based on profile
	if profile.PlayStyle == "combat-focused" && resonance > 0.6 {
		impacts = append(impacts, models.PredictedImpact{
			Type:        "mechanical",
			Description: "Increased motivation in upcoming combat encounters",
			Likelihood:  0.7,
			Magnitude:   0.5,
		})
	}

	if len(profile.Preferences.Themes) > 0 && resonance > 0.7 {
		impacts = append(impacts, models.PredictedImpact{
			Type:        "story_progression",
			Description: "High likelihood of pursuing related story threads",
			Likelihood:  0.85,
			Magnitude:   0.9,
		})
	}

	return impacts
}

func (pg *PerspectiveGenerator) findContradictions(perspectives []models.PerspectiveNarrative) {
	// Compare perspectives to find contradictions
	for i := 0; i < len(perspectives); i++ {
		for j := i + 1; j < len(perspectives); j++ {
			// Simple contradiction detection based on bias and truth level
			if math.Abs(perspectives[i].TruthLevel-perspectives[j].TruthLevel) > 0.3 {
				contradiction := models.Contradiction{
					OtherPerspectiveID: perspectives[j].ID,
					ConflictingDetail:  "Truth level discrepancy",
					ThisVersion:        fmt.Sprintf("Truth level: %.2f", perspectives[i].TruthLevel),
					OtherVersion:       fmt.Sprintf("Truth level: %.2f", perspectives[j].TruthLevel),
					TruthValue:         "both_partial",
				}
				perspectives[i].Contradictions = append(perspectives[i].Contradictions, contradiction)
			}
		}
	}
}

// Utility functions

func mergeUnique(existing, newItems []string) []string {
	seen := make(map[string]bool)
	for _, item := range existing {
		seen[item] = true
	}

	result := append([]string{}, existing...)
	for _, item := range newItems {
		if !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}

	return result
}

func formatBackstoryElements(elements []models.BackstoryElement) string {
	if len(elements) == 0 {
		return "No relevant backstory elements"
	}

	formatted := make([]string, 0, 10)
	for i := range elements {
		element := &elements[i]
		formatted = append(formatted, fmt.Sprintf("- [%s] %s (Weight: %.2f)", element.Type, element.Content, element.Weight))
	}

	return strings.Join(formatted, "\n")
}

func formatWorldState(state map[string]interface{}) string {
	// Format world state for AI context
	formatted, _ := json.MarshalIndent(state, "", "  ")
	return string(formatted)
}

func extractEventTags(event *models.NarrativeEvent) []string {
	if event == nil {
		return []string{}
	}
	// Extract relevant tags from event for matching
	tags := []string{event.Type}
	tags = append(tags, strings.Fields(event.Name)...)

	// Add location-based tags
	if event.Location != "" {
		tags = append(tags, strings.Fields(event.Location)...)
	}

	return tags
}

func getDelayPriority(delay string) int {
	priorities := map[string]int{
		"immediate": 1,
		"short":     2,
		"medium":    3,
		"long":      4,
	}

	if priority, ok := priorities[delay]; ok {
		return priority
	}
	return 5
}

// PerspectiveSource represents an entity that can have a perspective
type PerspectiveSource struct {
	ID              string
	Type            string // "npc", "faction", "deity", "historical"
	Name            string
	Background      string
	Motivations     []string
	Relationships   map[string]string
	CulturalContext map[string]interface{}
}
