package services

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/google/uuid"
)

// LivingEcosystemService manages the autonomous world simulation
type LivingEcosystemService struct {
	worldRepo      *database.EmergentWorldRepository
	npcRepo        database.NPCRepository
	factionRepo    *database.WorldBuildingRepository
	settlementRepo *database.WorldBuildingRepository
	llm            LLMProvider
	eventEngine    *WorldEventEngineService
}

// NewLivingEcosystemService creates a new living ecosystem service
func NewLivingEcosystemService(
	worldRepo *database.EmergentWorldRepository,
	npcRepo database.NPCRepository,
	factionRepo *database.WorldBuildingRepository,
	settlementRepo *database.WorldBuildingRepository,
	llm LLMProvider,
	eventEngine *WorldEventEngineService,
) *LivingEcosystemService {
	return &LivingEcosystemService{
		worldRepo:      worldRepo,
		npcRepo:        npcRepo,
		factionRepo:    factionRepo,
		settlementRepo: settlementRepo,
		llm:            llm,
		eventEngine:    eventEngine,
	}
}

// SimulateWorldProgress simulates world changes since last update
func (les *LivingEcosystemService) SimulateWorldProgress(ctx context.Context, sessionID string) error {
	// Get world state
	worldState, err := les.worldRepo.GetWorldState(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get world state: %w", err)
	}

	// Calculate time elapsed since last simulation
	timeDelta := time.Since(worldState.LastSimulated)
	if timeDelta < time.Hour {
		// Don't simulate if less than an hour has passed
		return nil
	}

	// Start simulation log
	simLog := &models.SimulationLog{
		ID:             uuid.New().String(),
		SessionID:      sessionID,
		SimulationType: "world_progress",
		StartTime:      time.Now(),
		EventsCreated:  0,
		Details:        make(map[string]interface{}),
		Success:        true,
	}

	// Simulate various aspects of the world
	events := []models.EmergentWorldEvent{}

	// 1. Simulate NPC activities
	npcEvents, err := les.simulateNPCActivities(ctx, sessionID, timeDelta)
	if err != nil {
		simLog.Details["npc_error"] = err.Error()
	} else {
		events = append(events, npcEvents...)
		simLog.Details["npc_events"] = len(npcEvents)
	}

	// 2. Simulate economic changes
	economicEvents, err := les.simulateEconomicChanges(ctx, sessionID, timeDelta)
	if err != nil {
		simLog.Details["economic_error"] = err.Error()
	} else {
		events = append(events, economicEvents...)
		simLog.Details["economic_events"] = len(economicEvents)
	}

	// 3. Simulate political developments
	politicalEvents, err := les.simulatePoliticalDevelopments(ctx, sessionID, timeDelta)
	if err != nil {
		simLog.Details["political_error"] = err.Error()
	} else {
		events = append(events, politicalEvents...)
		simLog.Details["political_events"] = len(politicalEvents)
	}

	// 4. Simulate natural events
	naturalEvents, err := les.simulateNaturalEvents(ctx, sessionID, timeDelta)
	if err != nil {
		simLog.Details["natural_error"] = err.Error()
	} else {
		events = append(events, naturalEvents...)
		simLog.Details["natural_events"] = len(naturalEvents)
	}

	// 5. Simulate cultural evolution
	culturalEvents, err := les.simulateCulturalEvolution(ctx, sessionID, timeDelta)
	if err != nil {
		simLog.Details["cultural_error"] = err.Error()
	} else {
		events = append(events, culturalEvents...)
		simLog.Details["cultural_events"] = len(culturalEvents)
	}

	// Save all events
	for _, event := range events {
		if err := les.worldRepo.CreateWorldEvent(&event); err != nil {
			simLog.Details["save_error"] = err.Error()
			simLog.Success = false
		} else {
			simLog.EventsCreated++
		}
	}

	// Update world state
	worldState.LastSimulated = time.Now()
	worldState.WorldData["last_events"] = events
	if err := les.worldRepo.UpdateWorldState(worldState); err != nil {
		simLog.Details["update_error"] = err.Error()
		simLog.Success = false
	}

	// Save simulation log
	simLog.EndTime = time.Now()
	_ = les.worldRepo.CreateSimulationLog(simLog)

	return nil
}

// simulateNPCActivities simulates autonomous NPC actions
func (les *LivingEcosystemService) simulateNPCActivities(ctx context.Context, sessionID string, timeDelta time.Duration) ([]models.EmergentWorldEvent, error) {
	events := []models.EmergentWorldEvent{}

	// Get all NPCs in the session
	npcs, err := les.npcRepo.GetByGameSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Simulate each NPC based on their goals and personality
	for _, npc := range npcs {
		// Get NPC goals
		goals, err := les.worldRepo.GetNPCGoals(npc.ID)
		if err != nil {
			continue
		}

		// Process each active goal
		for _, goal := range goals {
			if goal.Status != constants.StatusActive {
				continue
			}

			// Simulate progress on goal
			event, progress := les.simulateGoalProgress(ctx, npc, goal, timeDelta)
			if event != nil {
				events = append(events, *event)
			}

			// Update goal progress
			goal.Progress = progress
			if progress >= 1.0 {
				goal.Status = constants.StatusCompleted
				now := time.Now()
				goal.CompletedAt = &now
			}
			_ = les.worldRepo.UpdateNPCGoal(&goal)
		}

		// Check if NPC should create new goals
		if len(goals) < 3 && rand.Float64() < 0.3 {
			newGoal := les.generateNPCGoal(ctx, npc)
			if newGoal != nil {
				_ = les.worldRepo.CreateNPCGoal(newGoal)
			}
		}

		// Simulate NPC schedule activities
		scheduleEvents := les.simulateNPCSchedule(ctx, npc, timeDelta)
		events = append(events, scheduleEvents...)
	}

	return events, nil
}

// simulateGoalProgress simulates progress on an NPC goal
func (les *LivingEcosystemService) simulateGoalProgress(ctx context.Context, npc *models.NPC, goal models.NPCGoal, timeDelta time.Duration) (*models.EmergentWorldEvent, float64) {
	// Calculate progress based on goal type and time
	progressRate := 0.1 * (timeDelta.Hours() / 24.0) // Base 10% progress per day

	// Adjust based on NPC stats and goal type
	switch goal.GoalType {
	case "acquire_wealth":
		progressRate *= (1.0 + float64(npc.Attributes.Intelligence)/20.0)
	case "gain_influence":
		progressRate *= (1.0 + float64(npc.Attributes.Charisma)/20.0)
	case "improve_skill":
		progressRate *= (1.0 + float64(npc.Attributes.Wisdom)/20.0)
	case "complete_quest":
		progressRate *= (1.0 + float64(npc.Attributes.Strength+npc.Attributes.Dexterity)/40.0)
	}

	// Add some randomness
	progressRate *= (0.5 + rand.Float64())

	newProgress := math.Min(goal.Progress+progressRate, 1.0)

	// Generate event if significant progress
	if newProgress-goal.Progress > 0.25 {
		event := &models.EmergentWorldEvent{
			ID:          uuid.New().String(),
			SessionID:   npc.GameSessionID,
			EventType:   "npc_goal_progress",
			Title:       fmt.Sprintf("%s Makes Progress", npc.Name),
			Description: fmt.Sprintf("%s has made significant progress on their goal: %s", npc.Name, goal.Description),
			Impact: map[string]interface{}{
				"npc_id":    npc.ID,
				"goal_id":   goal.ID,
				"progress":  newProgress,
				"goal_type": goal.GoalType,
			},
			AffectedEntities: []string{npc.ID},
			IsPlayerVisible:  npc.Type == "ally" || npc.Type == "neutral",
			OccurredAt:       time.Now(),
		}
		return event, newProgress
	}

	return nil, newProgress
}

// generateNPCGoal creates a new goal for an NPC based on their personality
func (les *LivingEcosystemService) generateNPCGoal(ctx context.Context, npc *models.NPC) *models.NPCGoal {
	goalTypes := []string{
		"acquire_wealth",
		"gain_influence",
		"improve_skill",
		"complete_quest",
		"build_relationship",
		"seek_knowledge",
		"gain_power",
		"find_artifact",
	}

	// Weight goal types based on NPC personality and stats
	selectedType := goalTypes[rand.Intn(len(goalTypes))]

	// Use AI to generate appropriate goal description
	prompt := fmt.Sprintf(`Generate a specific personal goal for an NPC with these characteristics:
Name: %s
Type: %s
Goal Type: %s
Alignment: %s

Create a specific, achievable goal that fits their personality. Return a brief description (1-2 sentences).`,
		npc.Name, npc.Type, selectedType, npc.Alignment)

	description, err := les.llm.GenerateContent(ctx, prompt, "")
	if err != nil {
		description = fmt.Sprintf("Pursue %s through available means", selectedType)
	}

	return &models.NPCGoal{
		ID:          uuid.New().String(),
		NPCID:       npc.ID,
		GoalType:    selectedType,
		Priority:    rand.Intn(5) + 1,
		Description: description,
		Progress:    0.0,
		Parameters:  make(map[string]interface{}),
		Status:      constants.StatusActive,
		StartedAt:   time.Now(),
	}
}

// simulateNPCSchedule simulates daily NPC activities
func (les *LivingEcosystemService) simulateNPCSchedule(ctx context.Context, npc *models.NPC, timeDelta time.Duration) []models.EmergentWorldEvent {
	events := []models.EmergentWorldEvent{}

	// Get NPC schedule
	schedule, err := les.worldRepo.GetNPCSchedule(npc.ID)
	if err != nil || len(schedule) == 0 {
		// Generate default schedule if none exists
		les.generateDefaultSchedule(npc)
		return events
	}

	// Simulate activities based on time of day
	hoursElapsed := int(timeDelta.Hours())
	for i := 0; i < hoursElapsed && i < 24; i++ {
		hour := (time.Now().Hour() - hoursElapsed + i + 24) % 24
		timeOfDay := getTimeOfDay(hour)

		for _, activity := range schedule {
			if activity.TimeOfDay == timeOfDay {
				// Small chance of activity generating an event
				if rand.Float64() < 0.1 {
					event := les.generateScheduleEvent(ctx, npc, activity)
					if event != nil {
						events = append(events, *event)
					}
				}
			}
		}
	}

	return events
}

// generateScheduleEvent creates an event from a scheduled activity
func (les *LivingEcosystemService) generateScheduleEvent(ctx context.Context, npc *models.NPC, activity models.NPCSchedule) *models.EmergentWorldEvent {
	// Use AI to generate interesting event from routine activity
	prompt := fmt.Sprintf(`Generate a brief interesting event that occurs during this NPC's routine:
NPC: %s
Activity: %s at %s
Location: %s
Time: %s

Create a 1-2 sentence description of something noteworthy that happens. It could be an encounter, discovery, or minor incident.`,
		npc.Name, activity.Activity, activity.Location, activity.Location, activity.TimeOfDay)

	description, err := les.llm.GenerateContent(ctx, prompt, "")
	if err != nil {
		return nil
	}

	return &models.EmergentWorldEvent{
		ID:          uuid.New().String(),
		SessionID:   npc.GameSessionID,
		EventType:   "npc_activity",
		Title:       fmt.Sprintf("%s - %s", npc.Name, activity.Activity),
		Description: description,
		Impact: map[string]interface{}{
			"npc_id":   npc.ID,
			"activity": activity.Activity,
			"location": activity.Location,
		},
		AffectedEntities: []string{npc.ID},
		IsPlayerVisible:  rand.Float64() < 0.3, // 30% chance players hear about it
		OccurredAt:       time.Now(),
	}
}

// generateDefaultSchedule creates a basic schedule for an NPC
func (les *LivingEcosystemService) generateDefaultSchedule(npc *models.NPC) {
	schedules := []models.NPCSchedule{
		{
			ID:         uuid.New().String(),
			NPCID:      npc.ID,
			TimeOfDay:  "morning",
			Activity:   "daily_routine",
			Location:   "home",
			Parameters: map[string]interface{}{},
		},
		{
			ID:         uuid.New().String(),
			NPCID:      npc.ID,
			TimeOfDay:  "afternoon",
			Activity:   "work",
			Location:   "workplace",
			Parameters: map[string]interface{}{},
		},
		{
			ID:         uuid.New().String(),
			NPCID:      npc.ID,
			TimeOfDay:  "evening",
			Activity:   "socializing",
			Location:   "tavern",
			Parameters: map[string]interface{}{},
		},
		{
			ID:         uuid.New().String(),
			NPCID:      npc.ID,
			TimeOfDay:  "night",
			Activity:   "rest",
			Location:   "home",
			Parameters: map[string]interface{}{},
		},
	}

	for _, schedule := range schedules {
		_ = les.worldRepo.CreateNPCSchedule(&schedule)
	}
}

// simulateEconomicChanges simulates economic shifts in the world
func (les *LivingEcosystemService) simulateEconomicChanges(ctx context.Context, sessionID string, timeDelta time.Duration) ([]models.EmergentWorldEvent, error) {
	events := []models.EmergentWorldEvent{}

	// Get settlements
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, err
	}
	settlements, err := les.settlementRepo.GetSettlementsByGameSession(sessionUUID)
	if err != nil {
		return nil, err
	}

	for _, settlement := range settlements {
		// Simulate trade effects
		if rand.Float64() < 0.2*(timeDelta.Hours()/168.0) { // 20% chance per week
			event := les.generateEconomicEvent(ctx, settlement)
			if event != nil {
				events = append(events, *event)
			}
		}

		// Update settlement prosperity based on various factors
		les.updateSettlementProsperity(ctx, settlement, timeDelta)
	}

	return events, nil
}

// generateEconomicEvent creates an economic event for a settlement
func (les *LivingEcosystemService) generateEconomicEvent(ctx context.Context, settlement *models.Settlement) *models.EmergentWorldEvent {
	eventTypes := []string{
		"trade_boom", "trade_disruption", "new_resource", "resource_depletion",
		"merchant_arrival", "market_crash", "guild_formation", "technological_advance",
	}

	eventType := eventTypes[rand.Intn(len(eventTypes))]

	prompt := fmt.Sprintf(`Generate an economic event for a settlement:
Settlement: %s
Population: %d
Event Type: %s
Economy Type: %v

Create a brief description (2-3 sentences) of this economic event and its immediate effects.`,
		settlement.Name, settlement.Population, eventType, settlement.GovernmentType)

	description, err := les.llm.GenerateContent(ctx, prompt, "")
	if err != nil {
		return nil
	}

	// Calculate economic impact
	impact := (rand.Float64() - 0.5) * 0.4 // -20% to +20%
	switch eventType {
	case "trade_boom", "new_resource":
		impact = math.Abs(impact)
	case "market_crash", "resource_depletion":
		impact = -math.Abs(impact)
	}

	return &models.EmergentWorldEvent{
		ID:          uuid.New().String(),
		SessionID:   settlement.GameSessionID.String(),
		EventType:   "economic_" + eventType,
		Title:       fmt.Sprintf("Economic Event in %s", settlement.Name),
		Description: description,
		Impact: map[string]interface{}{
			"settlement_id":   settlement.ID.String(),
			"economic_impact": impact,
			"affected_goods":  les.getAffectedGoods(eventType),
			"duration_days":   rand.Intn(30) + 10,
		},
		AffectedEntities: []string{settlement.ID.String()},
		IsPlayerVisible:  true,
		OccurredAt:       time.Now(),
		Consequences: []models.EventConsequence{
			{
				Type:      "economic",
				Target:    settlement.ID.String(),
				Effect:    "prosperity_change",
				Magnitude: impact,
				Duration:  fmt.Sprintf("%d days", rand.Intn(30)+10),
				Parameters: map[string]interface{}{
					"event_type": eventType,
				},
			},
		},
	}
}

// updateSettlementProsperity updates a settlement's economic status
func (les *LivingEcosystemService) updateSettlementProsperity(ctx context.Context, settlement *models.Settlement, timeDelta time.Duration) {
	// Base prosperity change
	prosperityChange := 0.0

	// Factor in trade routes
	if tradeRoutes, ok := map[string]interface{}{}["trade_routes"].([]interface{}); ok {
		prosperityChange += float64(len(tradeRoutes)) * 0.01
	}

	// Factor in resources
	if resources, ok := map[string]interface{}{}["resources"].([]interface{}); ok {
		prosperityChange += float64(len(resources)) * 0.005
	}

	// Factor in population
	if settlement.Population > 10000 {
		prosperityChange += 0.01
	} else if settlement.Population < 1000 {
		prosperityChange -= 0.01
	}

	// Apply time-based change
	prosperityChange *= timeDelta.Hours() / 168.0 // Weekly rate
	_ = prosperityChange                          // placeholder until persistence implemented

	// Update prosperity
	// TODO: Implement UpdateSettlement method in repository
	// if prosperity, ok := map[string]interface{}{}["prosperity"].(float64); ok {
	// 	map[string]interface{}{}["prosperity"] = math.Max(0, math.Min(1, prosperity+prosperityChange))
	// 	les.settlementRepo.UpdateSettlement(settlement.ID, settlement)
	// }
}

// simulatePoliticalDevelopments simulates political changes and faction activities
func (les *LivingEcosystemService) simulatePoliticalDevelopments(ctx context.Context, sessionID string, timeDelta time.Duration) ([]models.EmergentWorldEvent, error) {
	events := []models.EmergentWorldEvent{}

	// Get factions
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, err
	}
	factions, err := les.factionRepo.GetFactionsByGameSession(sessionUUID)
	if err != nil {
		return nil, err
	}

	// Process faction agendas
	for _, faction := range factions {
		// Get faction personality
		personality, err := les.worldRepo.GetFactionPersonality(faction.ID.String())
		if err != nil {
			continue
		}

		// Process active agendas
		agendas, err := les.worldRepo.GetFactionAgendas(faction.ID.String())
		if err != nil {
			continue
		}

		for _, agenda := range agendas {
			if agenda.Status != constants.StatusActive {
				continue
			}

			// Simulate agenda progress
			event := les.simulateAgendaProgress(ctx, faction, personality, &agenda, timeDelta)
			if event != nil {
				events = append(events, *event)
			}

			// Update agenda
			_ = les.worldRepo.UpdateFactionAgenda(&agenda)
		}

		// Check for new political opportunities
		if rand.Float64() < 0.15*(timeDelta.Hours()/168.0) { // 15% chance per week
			event := les.generatePoliticalOpportunity(ctx, faction, personality)
			if event != nil {
				events = append(events, *event)
			}
		}
	}

	// Simulate faction interactions
	interactionEvents := les.simulateFactionInteractions(ctx, factions, timeDelta)
	events = append(events, interactionEvents...)

	return events, nil
}

// simulateAgendaProgress advances a faction's political agenda
func (les *LivingEcosystemService) simulateAgendaProgress(ctx context.Context, faction *models.Faction, personality *models.FactionPersonality, agenda *models.FactionAgenda, timeDelta time.Duration) *models.EmergentWorldEvent {
	// Calculate progress based on faction traits and resources
	progressRate := 0.05 * (timeDelta.Hours() / 168.0) // Base 5% per week

	// Modify based on faction traits
	if aggressive, ok := personality.Traits["aggressive"]; ok {
		progressRate *= (1.0 + aggressive*0.5)
	}
	if diplomatic, ok := personality.Traits["diplomatic"]; ok {
		progressRate *= (1.0 + diplomatic*0.3)
	}

	// Check current stage
	for i, stage := range agenda.Stages {
		if !stage.IsComplete {
			// Random chance to complete stage
			if rand.Float64() < progressRate {
				stage.IsComplete = true
				now := time.Now()
				stage.CompletedAt = &now
				agenda.Stages[i] = stage
				agenda.Progress = float64(i+1) / float64(len(agenda.Stages))

				// Generate completion event
				return &models.EmergentWorldEvent{
					ID:        uuid.New().String(),
					SessionID: faction.GameSessionID.String(),
					EventType: "political_milestone",
					Title:     fmt.Sprintf("%s Advances Agenda", faction.Name),
					Description: fmt.Sprintf("%s has completed a key milestone in their agenda '%s': %s",
						faction.Name, agenda.Title, stage.Description),
					Impact: map[string]interface{}{
						"faction_id":   faction.ID,
						"agenda_id":    agenda.ID,
						"stage_name":   stage.Name,
						"new_progress": agenda.Progress,
					},
					AffectedEntities: []string{faction.ID.String()},
					IsPlayerVisible:  true,
					OccurredAt:       time.Now(),
				}
			}
			break
		}
	}

	// Check if agenda is complete
	if agenda.Progress >= 1.0 {
		agenda.Status = constants.StatusCompleted
	}

	return nil
}

// generatePoliticalOpportunity creates new political events for factions
func (les *LivingEcosystemService) generatePoliticalOpportunity(ctx context.Context, faction *models.Faction, personality *models.FactionPersonality) *models.EmergentWorldEvent {
	opportunities := []string{
		"alliance_proposal", "trade_agreement", "territorial_claim",
		"diplomatic_summit", "military_buildup", "espionage_discovered",
		"succession_crisis", "popular_uprising", "religious_movement",
	}

	opportunity := opportunities[rand.Intn(len(opportunities))]

	prompt := fmt.Sprintf(`Generate a political opportunity or crisis for a faction:
Faction: %s
Type: %s
Opportunity: %s
Faction Traits: %v
Current Relations: %v

Create a compelling description (2-3 sentences) of this political development.`,
		faction.Name, faction.Type, opportunity, personality.Traits, faction.FactionRelationships)

	description, err := les.llm.GenerateContent(ctx, prompt, "")
	if err != nil {
		return nil
	}

	return &models.EmergentWorldEvent{
		ID:          uuid.New().String(),
		SessionID:   faction.GameSessionID.String(),
		EventType:   "political_opportunity",
		Title:       fmt.Sprintf("%s - %s", faction.Name, opportunity),
		Description: description,
		Impact: map[string]interface{}{
			"faction_id":      faction.ID,
			"opportunity":     opportunity,
			"response_needed": true,
			"deadline_days":   rand.Intn(14) + 7,
		},
		AffectedEntities: []string{faction.ID.String()},
		IsPlayerVisible:  true,
		OccurredAt:       time.Now(),
	}
}

// simulateFactionInteractions simulates diplomatic and conflict interactions
func (les *LivingEcosystemService) simulateFactionInteractions(ctx context.Context, factions []*models.Faction, timeDelta time.Duration) []models.EmergentWorldEvent {
	events := []models.EmergentWorldEvent{}

	// Check each faction pair
	for i := 0; i < len(factions); i++ {
		for j := i + 1; j < len(factions); j++ {
			faction1 := factions[i]
			faction2 := factions[j]

			// Get current relationship
			relationship := les.getFactionRelationship(faction1, faction2)

			// Chance of interaction based on relationship
			interactionChance := 0.1 * (timeDelta.Hours() / 168.0) // Base 10% per week
			if relationship < -50 {
				interactionChance *= 2 // More likely if hostile
			} else if relationship > 50 {
				interactionChance *= 1.5 // Somewhat more likely if friendly
			}

			if rand.Float64() < interactionChance {
				event := les.generateFactionInteraction(ctx, faction1, faction2, relationship)
				if event != nil {
					events = append(events, *event)
				}
			}
		}
	}

	return events
}

// simulateNaturalEvents generates natural world events
func (les *LivingEcosystemService) simulateNaturalEvents(ctx context.Context, sessionID string, timeDelta time.Duration) ([]models.EmergentWorldEvent, error) {
	events := []models.EmergentWorldEvent{}

	// Chance of natural events
	eventChance := 0.2 * (timeDelta.Hours() / 168.0) // 20% chance per week

	if rand.Float64() < eventChance {
		eventTypes := []string{
			"weather_extreme", "natural_disaster", "celestial_event",
			"monster_migration", "magical_anomaly", "resource_discovery",
			"plague_outbreak", "bumper_harvest", "divine_manifestation",
		}

		eventType := eventTypes[rand.Intn(len(eventTypes))]

		prompt := fmt.Sprintf(`Generate a natural or supernatural event:
Event Type: %s
World Setting: Fantasy D&D

Create a dramatic description (2-3 sentences) of this event and its immediate impact on the region.`, eventType)

		description, err := les.llm.GenerateContent(ctx, prompt, "")
		if err != nil {
			return events, nil
		}

		// Determine affected area
		affectedEntities := les.determineAffectedEntities(ctx, sessionID, eventType)

		event := models.EmergentWorldEvent{
			ID:          uuid.New().String(),
			SessionID:   sessionID,
			EventType:   "natural_" + eventType,
			Title:       les.generateEventTitle(eventType),
			Description: description,
			Impact: map[string]interface{}{
				"severity": rand.Intn(5) + 1,
				"duration": fmt.Sprintf("%d days", rand.Intn(30)+1),
				"area":     les.getEventArea(eventType),
			},
			AffectedEntities: affectedEntities,
			IsPlayerVisible:  true,
			OccurredAt:       time.Now(),
		}

		events = append(events, event)
	}

	return events, nil
}

// simulateCulturalEvolution simulates gradual cultural changes
func (les *LivingEcosystemService) simulateCulturalEvolution(ctx context.Context, sessionID string, timeDelta time.Duration) ([]models.EmergentWorldEvent, error) {
	events := []models.EmergentWorldEvent{}

	// Get cultures
	cultures, err := les.worldRepo.GetCulturesBySession(sessionID)
	if err != nil {
		return nil, err
	}

	for _, culture := range cultures {
		// Small chance of cultural shift
		if rand.Float64() < 0.05*(timeDelta.Hours()/720.0) { // 5% chance per month
			event := les.generateCulturalShift(ctx, culture)
			if event != nil {
				events = append(events, *event)
			}
		}
	}

	return events, nil
}

// Helper functions

func getTimeOfDay(hour int) string {
	switch {
	case hour >= 6 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 18:
		return "afternoon"
	case hour >= 18 && hour < 22:
		return "evening"
	default:
		return "night"
	}
}

func (les *LivingEcosystemService) getAffectedGoods(eventType string) []string {
	goodsMap := map[string][]string{
		"trade_boom":            {"all"},
		"trade_disruption":      {"luxury_goods", "exotic_materials"},
		"new_resource":          {"raw_materials", "crafting_supplies"},
		"resource_depletion":    {"basic_resources", "food"},
		"merchant_arrival":      {"rare_items", "foreign_goods"},
		"market_crash":          {"all"},
		"guild_formation":       {"crafted_goods", "services"},
		"technological_advance": {"tools", "weapons", "armor"},
	}

	if goods, ok := goodsMap[eventType]; ok {
		return goods
	}
	return []string{"general_goods"}
}

func (les *LivingEcosystemService) getFactionRelationship(faction1, faction2 *models.Faction) float64 {
	// TODO: Parse FactionRelationships JSONB to get standing
	// For now, return neutral relationship
	return 0.0
}

func (les *LivingEcosystemService) generateFactionInteraction(ctx context.Context, faction1, faction2 *models.Faction, relationship float64) *models.EmergentWorldEvent {
	var interactionTypes []string

	if relationship < -25 {
		interactionTypes = []string{"border_skirmish", "trade_embargo", "diplomatic_protest", "spy_captured"}
	} else if relationship > 25 {
		interactionTypes = []string{"trade_deal", "military_cooperation", "cultural_exchange", "royal_marriage"}
	} else {
		interactionTypes = []string{"diplomatic_meeting", "trade_negotiation", "border_dispute", "information_exchange"}
	}

	interaction := interactionTypes[rand.Intn(len(interactionTypes))]

	prompt := fmt.Sprintf(`Generate a faction interaction event:
Faction 1: %s
Faction 2: %s
Interaction Type: %s
Current Relationship: %.0f (scale: -100 hostile to +100 allied)

Create a description (2-3 sentences) of this interaction and its outcome.`,
		faction1.Name, faction2.Name, interaction, relationship)

	description, err := les.llm.GenerateContent(ctx, prompt, "")
	if err != nil {
		return nil
	}

	// Calculate relationship change
	relationshipChange := 0.0
	switch interaction {
	case "border_skirmish", "trade_embargo", "spy_captured":
		relationshipChange = -(rand.Float64()*10 + 5)
	case "trade_deal", "military_cooperation", "royal_marriage":
		relationshipChange = rand.Float64()*10 + 5
	case "cultural_exchange":
		relationshipChange = rand.Float64()*5 + 2
	default:
		relationshipChange = (rand.Float64() - 0.5) * 10
	}

	return &models.EmergentWorldEvent{
		ID:          uuid.New().String(),
		SessionID:   faction1.GameSessionID.String(),
		EventType:   "faction_interaction",
		Title:       fmt.Sprintf("%s between %s and %s", interaction, faction1.Name, faction2.Name),
		Description: description,
		Impact: map[string]interface{}{
			"faction1_id":         faction1.ID.String(),
			"faction2_id":         faction2.ID.String(),
			"interaction_type":    interaction,
			"relationship_change": relationshipChange,
		},
		AffectedEntities: []string{faction1.ID.String(), faction2.ID.String()},
		IsPlayerVisible:  true,
		OccurredAt:       time.Now(),
		Consequences: []models.EventConsequence{
			{
				Type:      "diplomatic",
				Target:    faction1.ID.String() + "_" + faction2.ID.String(),
				Effect:    "relationship_change",
				Magnitude: relationshipChange,
				Duration:  "permanent",
				Parameters: map[string]interface{}{
					"interaction": interaction,
				},
			},
		},
	}
}

func (les *LivingEcosystemService) determineAffectedEntities(ctx context.Context, sessionID string, eventType string) []string {
	// For now, return empty - in full implementation would determine based on event type and location
	return []string{}
}

func (les *LivingEcosystemService) generateEventTitle(eventType string) string {
	titles := map[string]string{
		"weather_extreme":      "Extreme Weather Event",
		"natural_disaster":     "Natural Disaster Strikes",
		"celestial_event":      "Celestial Phenomenon",
		"monster_migration":    "Monster Migration",
		"magical_anomaly":      "Magical Anomaly Detected",
		"resource_discovery":   "New Resource Discovered",
		"plague_outbreak":      "Disease Outbreak",
		"bumper_harvest":       "Exceptional Harvest",
		"divine_manifestation": "Divine Manifestation",
	}

	if title, ok := titles[eventType]; ok {
		return title
	}
	return "Unusual Event"
}

func (les *LivingEcosystemService) getEventArea(eventType string) string {
	areas := map[string]string{
		"weather_extreme":      "regional",
		"natural_disaster":     "local",
		"celestial_event":      "global",
		"monster_migration":    "regional",
		"magical_anomaly":      "local",
		"resource_discovery":   "local",
		"plague_outbreak":      "regional",
		"bumper_harvest":       "local",
		"divine_manifestation": "local",
	}

	if area, ok := areas[eventType]; ok {
		return area
	}
	return "local"
}

func (les *LivingEcosystemService) generateCulturalShift(ctx context.Context, culture *models.ProceduralCulture) *models.EmergentWorldEvent {
	shiftTypes := []string{
		"artistic_renaissance", "religious_reform", "linguistic_evolution",
		"social_movement", "technological_adoption", "cultural_fusion",
		"traditional_revival", "philosophical_shift",
	}

	shiftType := shiftTypes[rand.Intn(len(shiftTypes))]

	prompt := fmt.Sprintf(`Generate a cultural evolution event:
Culture: %s
Shift Type: %s
Current Values: %v

Describe this cultural shift and its impact on society (2-3 sentences).`,
		culture.Name, shiftType, culture.Values)

	description, err := les.llm.GenerateContent(ctx, prompt, "")
	if err != nil {
		return nil
	}

	return &models.EmergentWorldEvent{
		ID:          uuid.New().String(),
		SessionID:   culture.Metadata["session_id"].(string),
		EventType:   "cultural_shift",
		Title:       fmt.Sprintf("%s in %s Culture", shiftType, culture.Name),
		Description: description,
		Impact: map[string]interface{}{
			"culture_id":       culture.ID,
			"shift_type":       shiftType,
			"affected_aspects": les.getAffectedCulturalAspects(shiftType),
		},
		AffectedEntities: []string{culture.ID},
		IsPlayerVisible:  true,
		OccurredAt:       time.Now(),
	}
}

func (les *LivingEcosystemService) getAffectedCulturalAspects(shiftType string) []string {
	aspectsMap := map[string][]string{
		"artistic_renaissance":   {"art_style", "music_style", "architecture"},
		"religious_reform":       {"belief_system", "customs", "holy_days"},
		"linguistic_evolution":   {"language", "naming_conventions", "idioms"},
		"social_movement":        {"social_structure", "values", "taboos"},
		"technological_adoption": {"architecture", "art_style", "customs"},
		"cultural_fusion":        {"cuisine", "clothing_style", "language"},
		"traditional_revival":    {"customs", "belief_system", "art_style"},
		"philosophical_shift":    {"values", "belief_system", "social_structure"},
	}

	if aspects, ok := aspectsMap[shiftType]; ok {
		return aspects
	}
	return []string{"general"}
}
