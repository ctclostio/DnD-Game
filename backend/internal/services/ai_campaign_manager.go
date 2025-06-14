package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
	"github.com/google/uuid"
)

type AICampaignManager struct {
	llmProvider LLMProvider
	config      *AIConfig
	logger      *logger.LoggerV2
}

func NewAICampaignManager(provider LLMProvider, config *AIConfig, log *logger.LoggerV2) *AICampaignManager {
	return &AICampaignManager{
		llmProvider: provider,
		config:      config,
		logger:      log,
	}
}

// GenerateStoryArc creates an interconnected story arc based on player actions and context.
func (acm *AICampaignManager) GenerateStoryArc(ctx context.Context, req models.GenerateStoryArcRequest) (*models.GeneratedStoryArc, error) {
	if !acm.config.Enabled {
		return acm.generateDefaultStoryArc(req), nil
	}

	prompt := acm.buildStoryArcPrompt(req)

	systemPrompt := "You are an expert D&D 5e Dungeon Master specializing in creating engaging, interconnected story arcs that adapt to player choices. Generate compelling narratives that feel organic and responsive to player actions. Your response must be valid JSON."

	response, err := acm.llmProvider.GenerateCompletion(ctx, prompt, systemPrompt)
	if err != nil {
		acm.logger.WithContext(ctx).
			Error().
			Err(err).
			Str("arc_type", req.ArcType).
			Str("complexity", req.Complexity).
			Msg("Error generating AI story arc")
		return acm.generateDefaultStoryArc(req), nil
	}

	var arc models.GeneratedStoryArc
	if err := json.Unmarshal([]byte(response), &arc); err != nil {
		acm.logger.WithContext(ctx).
			Error().
			Err(err).
			Msg("Error parsing AI story arc response")
		return acm.generateDefaultStoryArc(req), nil
	}

	return &arc, nil
}

// GenerateSessionRecap creates a "Previously on..." summary for session start
func (acm *AICampaignManager) GenerateSessionRecap(ctx context.Context, memories []*models.SessionMemory) (*models.GeneratedRecap, error) {
	if !acm.config.Enabled || len(memories) == 0 {
		return acm.generateDefaultRecap(memories), nil
	}

	prompt := acm.buildRecapPrompt(memories)

	systemPrompt := "You are a skilled narrator creating engaging 'Previously on...' recaps for D&D sessions. Create dramatic, concise summaries that remind players of key events while building excitement for the upcoming session. Your response must be valid JSON."

	response, err := acm.llmProvider.GenerateCompletion(ctx, prompt, systemPrompt)
	if err != nil {
		acm.logger.WithContext(ctx).
			Error().
			Err(err).
			Int("memory_count", len(memories)).
			Msg("Error generating AI session recap")
		return acm.generateDefaultRecap(memories), nil
	}

	var recap models.GeneratedRecap
	if err := json.Unmarshal([]byte(response), &recap); err != nil {
		acm.logger.WithContext(ctx).
			Error().
			Err(err).
			Msg("Error parsing AI recap response")
		return acm.generateDefaultRecap(memories), nil
	}

	return &recap, nil
}

// GenerateForeshadowing creates subtle hints about future plot elements.
func (acm *AICampaignManager) GenerateForeshadowing(ctx context.Context, req models.GenerateForeshadowingRequest, plotThread *models.PlotThread, storyArc *models.StoryArc) (*models.GeneratedForeshadowing, error) {
	if !acm.config.Enabled {
		return acm.generateDefaultForeshadowing(req), nil
	}

	prompt := acm.buildForeshadowingPrompt(req, plotThread, storyArc)

	systemPrompt := "You are a master of narrative foreshadowing in D&D campaigns. Create subtle hints and clues that will make sense in retrospect without being too obvious. Balance mystery with clarity. Your response must be valid JSON."

	response, err := acm.llmProvider.GenerateCompletion(ctx, prompt, systemPrompt)
	if err != nil {
		acm.logger.WithContext(ctx).
			Error().
			Err(err).
			Str("element_type", req.ElementType).
			Int("subtlety_level", req.SubtletyLevel).
			Msg("Error generating AI foreshadowing")
		return acm.generateDefaultForeshadowing(req), nil
	}

	var foreshadowing models.GeneratedForeshadowing
	if err := json.Unmarshal([]byte(response), &foreshadowing); err != nil {
		acm.logger.WithContext(ctx).
			Error().
			Err(err).
			Msg("Error parsing AI foreshadowing response")
		return acm.generateDefaultForeshadowing(req), nil
	}

	return &foreshadowing, nil
}

// AnalyzeSessionForMemory processes a session's events to create a structured memory.
func (acm *AICampaignManager) AnalyzeSessionForMemory(ctx context.Context, events []interface{}) (*models.SessionMemory, error) {
	if !acm.config.Enabled {
		return acm.createBasicSessionMemory(events), nil
	}

	prompt := acm.buildSessionAnalysisPrompt(events)

	systemPrompt := "You are an expert at analyzing D&D game sessions and extracting key information. Identify important events, decisions, NPCs, items, and plot developments to create a comprehensive session memory. Your response must be valid JSON."

	response, err := acm.llmProvider.GenerateCompletion(ctx, prompt, systemPrompt)
	if err != nil {
		acm.logger.WithContext(ctx).
			Error().
			Err(err).
			Int("event_count", len(events)).
			Msg("Error analyzing AI session")
		return acm.createBasicSessionMemory(events), nil
	}

	var memory models.SessionMemory
	if err := json.Unmarshal([]byte(response), &memory); err != nil {
		acm.logger.WithContext(ctx).
			Error().
			Err(err).
			Msg("Error parsing AI session analysis")
		return acm.createBasicSessionMemory(events), nil
	}

	memory.ID = uuid.New()
	memory.CreatedAt = time.Now()
	memory.UpdatedAt = time.Now()

	return &memory, nil
}

// Helper methods for building prompts.
func (acm *AICampaignManager) buildStoryArcPrompt(req models.GenerateStoryArcRequest) string {
	return fmt.Sprintf(`Generate a compelling D&D 5e story arc with the following parameters:

Context: %s
Player Goals: %s
Arc Type: %s
Complexity: %s

Create a story arc that:
1. Connects naturally to the current campaign context
2. Incorporates player goals and interests
3. Has clear milestones and progression
4. Includes meaningful conflicts and stakes
5. Provides opportunities for character growth
6. Can branch based on player choices

Format the response as JSON with the following structure:
{
  "title": "Arc title",
  "description": "Detailed arc description",
  "arc_type": "%s",
  "importance_level": 1-10,
  "key_milestones": [
    {"title": "Milestone name", "description": "What happens", "trigger": "What causes this"}
  ],
  "potential_conflicts": [
    {"type": "combat/social/moral", "description": "Conflict details", "stakes": "What's at risk"}
  ],
  "npcs_involved": [
    {"name": "NPC name", "role": "Their role", "motivation": "What they want"}
  ],
  "expected_duration": "Estimated sessions",
  "possible_resolutions": [
    {"type": "victory/compromise/failure", "description": "How it ends", "consequences": "What happens next"}
  ],
  "connections": [
    {"to_arc": "Related arc", "relationship": "How they connect"}
  ]
}`,
		req.Context,
		strings.Join(req.PlayerGoals, ", "),
		req.ArcType,
		req.Complexity,
		req.ArcType)
}

func (acm *AICampaignManager) buildRecapPrompt(memories []*models.SessionMemory) string {
	sessionsText := make([]string, 0, 10)
	for _, memory := range memories {
		sessionInfo := fmt.Sprintf("Session %d (%s):\n", memory.SessionNumber, memory.SessionDate.Format("Jan 2"))
		if memory.RecapSummary != "" {
			sessionInfo += memory.RecapSummary
		} else {
			// Extract key events from JSONB.
			sessionInfo += "Key events from this session"
		}
		sessionsText = append(sessionsText, sessionInfo)
	}

	return fmt.Sprintf(`Create an engaging "Previously on..." recap based on these recent sessions:

%s

Generate a dramatic recap that:
1. Summarizes the most important events concisely
2. Highlights unresolved plot threads
3. Mentions key NPCs and their current status
4. Builds excitement for the upcoming session
5. Ends with a cliffhanger or hook

Format as JSON:
{
  "summary": "Main recap narrative (2-3 paragraphs)",
  "key_events": ["Event 1", "Event 2", ...],
  "unresolved_threads": ["Thread 1", "Thread 2", ...],
  "npc_updates": [
    {"name": "NPC name", "update": "Current status"}
  ],
  "cliffhanger": "Exciting session opener",
  "next_session_hooks": ["Hook 1", "Hook 2", ...]
}`, strings.Join(sessionsText, "\n\n"))
}

func (acm *AICampaignManager) buildForeshadowingPrompt(req models.GenerateForeshadowingRequest, plotThread *models.PlotThread, storyArc *models.StoryArc) string {
	var context string
	if plotThread != nil {
		context = fmt.Sprintf("Plot Thread: %s\nDescription: %s\nType: %s",
			plotThread.Title, plotThread.Description, plotThread.ThreadType)
	} else if storyArc != nil {
		context = fmt.Sprintf("Story Arc: %s\nDescription: %s\nType: %s",
			storyArc.Title, storyArc.Description, storyArc.ArcType)
	}

	return fmt.Sprintf(`Generate foreshadowing for the following narrative element:

%s

Element Type: %s
Subtlety Level: %d/10 (1=very obvious, 10=extremely subtle)

Create foreshadowing that:
1. Hints at future events without revealing too much
2. Feels natural in the game world
3. Can be discovered through various means
4. Rewards attentive players
5. Makes sense in retrospect

Format as JSON:
{
  "content": "The foreshadowing element description",
  "element_type": "%s",
  "subtlety_level": %d,
  "placement_suggestions": [
    "Where/how to introduce this hint"
  ],
  "reveal_timing": "When this should become clear",
  "connection_hints": [
    "How this connects to the larger narrative"
  ]
}`, context, req.ElementType, req.SubtletyLevel, req.ElementType, req.SubtletyLevel)
}

func (acm *AICampaignManager) buildSessionAnalysisPrompt(events []interface{}) string {
	eventsJSON, _ := json.Marshal(events)

	return fmt.Sprintf(`Analyze this D&D session's events and extract key information:

Events: %s

Extract and organize:
1. Key story events with their impact
2. Important NPCs encountered
3. Player decisions and their outcomes
4. Items and treasures acquired
5. Locations visited
6. Combat encounters
7. Plot developments

Format as JSON matching the SessionMemory structure:
{
  "recap_summary": "Narrative summary of the session",
  "key_events": [{"time": "When", "description": "What happened", "impact": "Why it matters"}],
  "npcs_encountered": ["NPC names and brief context"],
  "decisions_made": [{"context": "Situation", "choice": "What was chosen", "outcome": "Result"}],
  "items_acquired": ["Item descriptions"],
  "locations_visited": ["Location names and significance"],
  "combat_encounters": ["Battle summaries"],
  "plot_developments": ["How the story advanced"]
}`, string(eventsJSON))
}

// Default/fallback generators.
func (acm *AICampaignManager) generateDefaultStoryArc(req models.GenerateStoryArcRequest) *models.GeneratedStoryArc {
	// Simple title case conversion.
	titleCased := strings.ToUpper(req.ArcType[:1]) + strings.ToLower(req.ArcType[1:])
	return &models.GeneratedStoryArc{
		Title:           fmt.Sprintf("The %s Quest", titleCased),
		Description:     "A new adventure awaits the party",
		ArcType:         req.ArcType,
		ImportanceLevel: 5,
		KeyMilestones: []models.Milestone{
			{Title: "Discovery", Description: "The party learns of the quest", Trigger: "Investigation or NPC interaction"},
			{Title: "Challenge", Description: "The party faces obstacles", Trigger: "Progress on the quest"},
			{Title: "Resolution", Description: "The quest concludes", Trigger: "Overcoming the final challenge"},
		},
		ExpectedDuration: "3-5 sessions",
	}
}

func (acm *AICampaignManager) generateDefaultRecap(memories []*models.SessionMemory) *models.GeneratedRecap {
	var events []string
	if len(memories) > 0 {
		events = append(events, fmt.Sprintf("Last session (#%d), the party continued their adventure",
			memories[0].SessionNumber))
	}

	return &models.GeneratedRecap{
		Summary:           "Previously, the party embarked on their journey through the realm...",
		KeyEvents:         events,
		UnresolvedThreads: []string{"The party's current quest remains unfinished"},
		Cliffhanger:       "As you gather once more, adventure awaits...",
	}
}

func (acm *AICampaignManager) generateDefaultForeshadowing(req models.GenerateForeshadowingRequest) *models.GeneratedForeshadowing {
	return &models.GeneratedForeshadowing{
		Content:       "A mysterious sign appears, hinting at things to come",
		ElementType:   req.ElementType,
		SubtletyLevel: req.SubtletyLevel,
		PlacementSuggestions: []string{
			"During exploration",
			"In NPC dialogue",
			"As environmental detail",
		},
		RevealTiming: "When the time is right",
	}
}

func (acm *AICampaignManager) createBasicSessionMemory(events []interface{}) *models.SessionMemory {
	memory := &models.SessionMemory{
		ID:               uuid.New(),
		RecapSummary:     "The party continued their adventure",
		KeyEvents:        models.JSONB(`[]`),
		NPCsEncountered:  models.JSONB(`[]`),
		DecisionsMade:    models.JSONB(`[]`),
		ItemsAcquired:    models.JSONB(`[]`),
		LocationsVisited: models.JSONB(`[]`),
		CombatEncounters: models.JSONB(`[]`),
		PlotDevelopments: models.JSONB(`[]`),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Basic analysis of events could be added here.
	return memory
}
