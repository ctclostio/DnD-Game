package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
	"github.com/google/uuid"
)

// WorldEventEngineService manages world events that happen regardless of party actions.
type WorldEventEngineService struct {
	llmProvider    LLMProvider
	worldRepo      WorldBuildingRepository
	factionService *FactionSystemService
}

// NewWorldEventEngineService creates a new world event engine service.
func NewWorldEventEngineService(llmProvider LLMProvider, worldRepo WorldBuildingRepository, factionService *FactionSystemService) *WorldEventEngineService {
	return &WorldEventEngineService{
		llmProvider:    llmProvider,
		worldRepo:      worldRepo,
		factionService: factionService,
	}
}

// GenerateWorldEvent creates a new world event based on current world state.
func (s *WorldEventEngineService) GenerateWorldEvent(ctx context.Context, gameSessionID uuid.UUID, eventType models.WorldEventType) (*models.WorldEvent, error) {
	// Get current world state for context.
	settlements, _ := s.worldRepo.GetSettlementsByGameSession(gameSessionID)
	factions, _ := s.worldRepo.GetFactionsByGameSession(gameSessionID)
	activeEvents, _ := s.worldRepo.GetActiveWorldEvents(gameSessionID)

	systemPrompt := `You are creating world events for a dark fantasy setting where ancient evils shape history.
Events should feel like natural consequences of the world's deep history and current tensions.
Some events might awaken old powers, fulfill prophecies, or reveal long-buried secrets.`

	userPrompt := fmt.Sprintf(`Generate a %s world event with these conditions:
Current settlements: %d
Active factions: %d  
Ongoing events: %d

Create an event that:
1. Has clear causes rooted in the world's history or current politics
2. Affects multiple settlements or factions
3. Creates opportunities for adventure
4. Has stages that will unfold over time
5. Can be influenced but not easily stopped

For ancient-related events, consider:
- Seals weakening on old prisons
- Artifacts activating after eons
- Prophecies beginning to manifest
- Reality growing thin in certain places

Respond in JSON format:
{
  "name": "event name",
  "description": "detailed description",
  "cause": "what triggered this",
  "severity": "minor/moderate/major/catastrophic",
  "duration": "expected timeline",
  "ancientCause": boolean,
  "awakensAncientEvil": boolean,
  "prophecyRelated": boolean,
  "affectedRegions": ["region names"],
  "stages": [
    {"stage": 1, "description": "initial signs"},
    {"stage": 2, "description": "escalation"},
    {"stage": 3, "description": "climax"}
  ],
  "resolutionConditions": ["how it might end"],
  "consequences": {
    "ifResolved": "positive outcome",
    "ifIgnored": "negative outcome"
  },
  "economicImpacts": {
    "trade": "impact description",
    "prices": "what changes"
  },
  "politicalImpacts": {
    "factions": "how factions react",
    "stability": "regional stability effects"
  },
  "adventureHooks": ["hook 1", "hook 2", "hook 3"]
}`, eventType, len(settlements), len(factions), len(activeEvents))

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		// Fallback to procedural generation.
		return s.generateProceduralEvent(gameSessionID, eventType), nil
	}

	var eventData struct {
		Name               string   `json:"name"`
		Description        string   `json:"description"`
		Cause              string   `json:"cause"`
		Severity           string   `json:"severity"`
		Duration           string   `json:"duration"`
		AncientCause       bool     `json:"ancientCause"`
		AwakensAncientEvil bool     `json:"awakensAncientEvil"`
		ProphecyRelated    bool     `json:"prophecyRelated"`
		AffectedRegions    []string `json:"affectedRegions"`
		Stages             []struct {
			Stage       int    `json:"stage"`
			Description string `json:"description"`
		} `json:"stages"`
		ResolutionConditions []string               `json:"resolutionConditions"`
		Consequences         map[string]string      `json:"consequences"`
		EconomicImpacts      map[string]interface{} `json:"economicImpacts"`
		PoliticalImpacts     map[string]interface{} `json:"politicalImpacts"`
		AdventureHooks       []string               `json:"adventureHooks"`
	}

	if err := json.Unmarshal([]byte(response), &eventData); err != nil {
		return s.generateProceduralEvent(gameSessionID, eventType), nil
	}

	// Map severity string to enum.
	severityMap := map[string]models.WorldEventSeverity{
		"minor":        models.SeverityMinor,
		"moderate":     models.SeverityModerate,
		"major":        models.SeverityMajor,
		"catastrophic": models.SeverityCatastrophic,
	}

	severity := severityMap[strings.ToLower(eventData.Severity)]
	if severity == "" {
		severity = models.SeverityModerate
	}

	event := &models.WorldEvent{
		GameSessionID:      gameSessionID,
		Name:               eventData.Name,
		Type:               eventType,
		Severity:           severity,
		Description:        eventData.Description,
		Cause:              eventData.Cause,
		StartDate:          "Current",
		Duration:           eventData.Duration,
		IsActive:           true,
		IsResolved:         false,
		AncientCause:       eventData.AncientCause,
		AwakensAncientEvil: eventData.AwakensAncientEvil,
		ProphecyRelated:    eventData.ProphecyRelated,
		CurrentStage:       1,
		PartyAware:         false,
		PartyInvolved:      false,
	}

	// Convert to JSONB.
	affectedRegions, _ := json.Marshal(eventData.AffectedRegions)
	event.AffectedRegions = models.JSONB(affectedRegions)

	stages, _ := json.Marshal(eventData.Stages)
	event.Stages = models.JSONB(stages)

	resolutionConditions, _ := json.Marshal(eventData.ResolutionConditions)
	event.ResolutionConditions = models.JSONB(resolutionConditions)

	consequences, _ := json.Marshal(eventData.Consequences)
	event.Consequences = models.JSONB(consequences)

	economicImpacts, _ := json.Marshal(eventData.EconomicImpacts)
	event.EconomicImpacts = models.JSONB(economicImpacts)

	politicalImpacts, _ := json.Marshal(eventData.PoliticalImpacts)
	event.PoliticalImpacts = models.JSONB(politicalImpacts)

	// Determine affected settlements and factions.
	event.AffectedSettlements = s.determineAffectedSettlements(settlements, eventData.AffectedRegions)
	event.AffectedFactions = s.determineAffectedFactions(factions, eventType)

	// Empty party actions initially.
	event.PartyActions = models.JSONB("[]")

	// Save the event.
	if err := s.worldRepo.CreateWorldEvent(event); err != nil {
		return nil, fmt.Errorf("failed to save world event: %w", err)
	}

	// Trigger immediate effects.
	s.applyEventEffects(ctx, event)

	return event, nil
}

// SimulateEventProgression advances all active events.
func (s *WorldEventEngineService) SimulateEventProgression(ctx context.Context, gameSessionID uuid.UUID) error {
	activeEvents, err := s.worldRepo.GetActiveWorldEvents(gameSessionID)
	if err != nil {
		return err
	}

	for _, event := range activeEvents {
		// Check if event should progress.
		if s.shouldEventProgress(event) {
			if err := s.progressEvent(ctx, event); err != nil {
				logger.WithContext(ctx).WithError(err).WithField("event_name", event.Name).Warn().Msg("Failed to progress event")
			}
		}

		// Check resolution conditions.
		if s.checkResolutionConditions(ctx, event) {
			if err := s.resolveEvent(ctx, event); err != nil {
				logger.WithContext(ctx).WithError(err).WithField("event_name", event.Name).Warn().Msg("Failed to resolve event")
			}
		}
	}

	// Chance to generate new events.
	if rand.Float32() < 0.2 {
		eventTypes := []models.WorldEventType{
			models.EventPolitical,
			models.EventEconomic,
			models.EventNatural,
			models.EventSupernatural,
			models.EventAncientAwakening,
			models.EventPlanar,
		}

		eventType := eventTypes[rand.Intn(len(eventTypes))]

		// Higher chance for ancient events in corrupted worlds.
		corruptionLevel := s.calculateWorldCorruption(gameSessionID)
		if corruptionLevel > 5 && rand.Float32() < 0.5 {
			eventType = models.EventAncientAwakening
		}

		_, _ = s.GenerateWorldEvent(ctx, gameSessionID, eventType)
	}

	return nil
}

// NotifyPartyOfEvent makes the party aware of an event.
func (s *WorldEventEngineService) NotifyPartyOfEvent(eventID uuid.UUID) error {
	// TODO: Implement repository method to update party_aware.
	// query := `UPDATE world_events SET party_aware = true WHERE id = $1`.
	return nil
}

// RecordPartyAction records party intervention in an event.
func (s *WorldEventEngineService) RecordPartyAction(eventID uuid.UUID, action string) error {
	// Get the event.
	// Add action to party_actions array.
	// Update party_involved to true.
	// This would be implemented properly.
	return nil
}

// Helper methods.
func (s *WorldEventEngineService) shouldEventProgress(event *models.WorldEvent) bool {
	// Events progress based on various factors.
	// For now, simple random chance.
	progressChance := 0.3

	// Major events progress faster.
	if event.Severity == models.SeverityMajor || event.Severity == models.SeverityCatastrophic {
		progressChance = 0.5
	}

	// Ancient events are more inevitable.
	if event.AncientCause {
		progressChance += 0.2
	}

	return rand.Float64() < progressChance
}

func (s *WorldEventEngineService) progressEvent(ctx context.Context, event *models.WorldEvent) error {
	// Progress to next stage.
	err := s.worldRepo.ProgressWorldEvent(event.ID)
	if err != nil {
		return err
	}

	event.CurrentStage++

	// Apply new stage effects.
	s.applyStageEffects(ctx, event)

	// Generate related events for major progressions.
	if event.CurrentStage == 2 && (event.Severity == models.SeverityMajor || event.Severity == models.SeverityCatastrophic) {
		s.generateCascadeEvents(ctx, event)
	}

	return nil
}

func (s *WorldEventEngineService) checkResolutionConditions(ctx context.Context, event *models.WorldEvent) bool {
	// Check if resolution conditions are met.
	// This is simplified - full implementation would check actual conditions.
	// Party intervention can help resolve events.
	if event.PartyInvolved {
		var partyActions []interface{}
		_ = json.Unmarshal([]byte(event.PartyActions), &partyActions)
		if len(partyActions) >= 3 {
			return rand.Float32() < 0.7 // High chance if party is actively involved
		}
	}

	// Random chance for natural resolution.
	resolutionChance := 0.1
	if event.Severity == models.SeverityMinor {
		resolutionChance = 0.3
	}

	return rand.Float64() < resolutionChance
}

func (s *WorldEventEngineService) resolveEvent(ctx context.Context, event *models.WorldEvent) error {
	// Mark event as resolved.
	// TODO: Implement repository method to mark event as resolved.
	// query := `UPDATE world_events SET is_resolved = true, is_active = false WHERE id = $1`.
	// Apply resolution consequences.
	var consequences map[string]string
	_ = json.Unmarshal([]byte(event.Consequences), &consequences)

	outcome := "ifIgnored"
	if event.PartyInvolved {
		outcome = "ifResolved"
	}

	// Generate resolution event.
	resolutionEvent := &models.WorldEvent{
		GameSessionID: event.GameSessionID,
		Name:          fmt.Sprintf("Resolution: %s", event.Name),
		Type:          event.Type,
		Severity:      models.SeverityMinor,
		Description:   consequences[outcome],
		Cause:         fmt.Sprintf("Resolution of %s", event.Name),
		StartDate:     "Current",
		Duration:      "Immediate",
		IsActive:      true,
		IsResolved:    true,
	}

	// Copy affected areas.
	resolutionEvent.AffectedRegions = event.AffectedRegions
	resolutionEvent.AffectedSettlements = event.AffectedSettlements
	resolutionEvent.AffectedFactions = event.AffectedFactions

	// Empty other fields.
	resolutionEvent.EconomicImpacts = models.JSONB("{}")
	resolutionEvent.PoliticalImpacts = models.JSONB("{}")
	resolutionEvent.Stages = models.JSONB("[]")
	resolutionEvent.ResolutionConditions = models.JSONB("[]")
	resolutionEvent.Consequences = models.JSONB("{}")
	resolutionEvent.PartyActions = models.JSONB("[]")

	return s.worldRepo.CreateWorldEvent(resolutionEvent)
}

func (s *WorldEventEngineService) applyEventEffects(ctx context.Context, event *models.WorldEvent) {
	// Apply economic impacts.
	s.applyEconomicImpacts(ctx, event)

	// Apply political impacts.
	s.applyPoliticalImpacts(ctx, event)
}

func (s *WorldEventEngineService) applyStageEffects(ctx context.Context, event *models.WorldEvent) {
	// Apply effects specific to the current stage.
	// This could trigger new events, modify settlements, etc.

	// Ancient awakening events get worse over time.
	if event.Type == models.EventAncientAwakening {
		var affectedSettlements []string
		_ = json.Unmarshal([]byte(event.AffectedSettlements), &affectedSettlements)

		for _, settlementIDStr := range affectedSettlements {
			settlementID, err := uuid.Parse(settlementIDStr)
			if err != nil {
				continue
			}

			settlement, err := s.worldRepo.GetSettlement(settlementID)
			if err != nil || settlement == nil {
				continue
			}

			// Increase corruption.
			settlement.CorruptionLevel += event.CurrentStage
			if settlement.CorruptionLevel > 10 {
				settlement.CorruptionLevel = 10
			}

			// Update would be done through repository.
		}
	}
}

func (s *WorldEventEngineService) generateCascadeEvents(ctx context.Context, parentEvent *models.WorldEvent) {
	// Major events can trigger secondary events.
	if parentEvent.Type == models.EventAncientAwakening {
		// Ancient awakenings might trigger supernatural events elsewhere.
		cascadeEvent := &models.WorldEvent{
			GameSessionID:        parentEvent.GameSessionID,
			Name:                 fmt.Sprintf("Ripple Effect: %s", parentEvent.Name),
			Type:                 models.EventSupernatural,
			Severity:             models.SeverityModerate,
			Description:          "Strange phenomena manifest in response to ancient stirrings",
			Cause:                parentEvent.Name,
			StartDate:            "Current",
			Duration:             "Weeks",
			IsActive:             true,
			AncientCause:         true,
			ProphecyRelated:      parentEvent.ProphecyRelated,
			CurrentStage:         1,
			AffectedRegions:      models.JSONB("[]"),
			AffectedSettlements:  models.JSONB("[]"),
			AffectedFactions:     models.JSONB("[]"),
			EconomicImpacts:      models.JSONB("{}"),
			PoliticalImpacts:     models.JSONB("{}"),
			Stages:               models.JSONB("[]"),
			ResolutionConditions: models.JSONB("[]"),
			Consequences:         models.JSONB("{}"),
			PartyActions:         models.JSONB("[]"),
		}

		_ = s.worldRepo.CreateWorldEvent(cascadeEvent)
	}
}

func (s *WorldEventEngineService) determineAffectedSettlements(settlements []*models.Settlement, affectedRegions []string) models.JSONB {
	affectedSettlementIDs := []string{}

	for _, settlement := range settlements {
		for _, region := range affectedRegions {
			if strings.Contains(strings.ToLower(settlement.Region), strings.ToLower(region)) {
				affectedSettlementIDs = append(affectedSettlementIDs, settlement.ID.String())
				break
			}
		}
	}

	// If no specific matches, affect random settlements.
	if len(affectedSettlementIDs) == 0 && len(settlements) > 0 {
		numAffected := 1 + rand.Intn(3)
		if numAffected > len(settlements) {
			numAffected = len(settlements)
		}

		for i := 0; i < numAffected; i++ {
			settlement := settlements[rand.Intn(len(settlements))]
			affectedSettlementIDs = append(affectedSettlementIDs, settlement.ID.String())
		}
	}

	result, _ := json.Marshal(affectedSettlementIDs)
	return models.JSONB(result)
}

func (s *WorldEventEngineService) determineAffectedFactions(factions []*models.Faction, eventType models.WorldEventType) models.JSONB {
	affectedFactionIDs := []string{}

	// Certain event types affect specific faction types.
	targetFactionTypes := map[models.WorldEventType][]models.FactionType{
		models.EventPolitical:        {models.FactionPolitical, models.FactionMilitary},
		models.EventEconomic:         {models.FactionMerchant, models.FactionCriminal},
		models.EventSupernatural:     {models.FactionCult, models.FactionAncientOrder, models.FactionReligious},
		models.EventAncientAwakening: {models.FactionAncientOrder, models.FactionCult},
	}

	if targetTypes, exists := targetFactionTypes[eventType]; exists {
		for _, faction := range factions {
			for _, targetType := range targetTypes {
				if faction.Type == targetType {
					affectedFactionIDs = append(affectedFactionIDs, faction.ID.String())
					break
				}
			}
		}
	}

	// If no specific matches, affect random factions.
	if len(affectedFactionIDs) == 0 && len(factions) > 0 {
		numAffected := 1 + rand.Intn(2)
		if numAffected > len(factions) {
			numAffected = len(factions)
		}

		for i := 0; i < numAffected; i++ {
			faction := factions[rand.Intn(len(factions))]
			affectedFactionIDs = append(affectedFactionIDs, faction.ID.String())
		}
	}

	result, _ := json.Marshal(affectedFactionIDs)
	return models.JSONB(result)
}

func (s *WorldEventEngineService) calculateWorldCorruption(gameSessionID uuid.UUID) int {
	// Calculate overall world corruption level.
	settlements, err := s.worldRepo.GetSettlementsByGameSession(gameSessionID)
	if err != nil || len(settlements) == 0 {
		return 0
	}

	totalCorruption := 0
	for _, settlement := range settlements {
		totalCorruption += settlement.CorruptionLevel
	}

	return totalCorruption / len(settlements)
}

// Procedural event generation.
func (s *WorldEventEngineService) generateProceduralEvent(gameSessionID uuid.UUID, eventType models.WorldEventType) *models.WorldEvent {
	eventNames := map[models.WorldEventType][]string{
		models.EventPolitical:        {"Border Dispute", "Succession Crisis", "Trade Embargo"},
		models.EventEconomic:         {"Market Crash", "Resource Shortage", "Trade Boom"},
		models.EventNatural:          {"Plague Outbreak", "Severe Storm", "Earthquake"},
		models.EventSupernatural:     {"Strange Portents", "Magical Anomaly", "Undead Rising"},
		models.EventAncientAwakening: {"Seal Weakening", "Artifact Activation", "Prophecy Manifests"},
		models.EventPlanar:           {"Planar Rift", "Dimensional Instability", "Outsider Incursion"},
	}

	names := eventNames[eventType]
	if names == nil {
		names = []string{"Unknown Event"}
	}

	severities := []models.WorldEventSeverity{
		models.SeverityMinor,
		models.SeverityModerate,
		models.SeverityMajor,
	}

	// Ancient events tend to be more severe.
	if eventType == models.EventAncientAwakening || eventType == models.EventPlanar {
		severities = []models.WorldEventSeverity{
			models.SeverityModerate,
			models.SeverityMajor,
			models.SeverityCatastrophic,
		}
	}

	event := &models.WorldEvent{
		GameSessionID:      gameSessionID,
		Name:               names[rand.Intn(len(names))],
		Type:               eventType,
		Severity:           severities[rand.Intn(len(severities))],
		Description:        "A significant event affecting the region",
		Cause:              "Natural progression of world events",
		StartDate:          "Current",
		Duration:           "Several weeks",
		IsActive:           true,
		CurrentStage:       1,
		AncientCause:       eventType == models.EventAncientAwakening,
		AwakensAncientEvil: eventType == models.EventAncientAwakening && rand.Float32() < 0.3,
		ProphecyRelated:    rand.Float32() < 0.2,
	}

	// Empty arrays.
	event.AffectedRegions = models.JSONB("[]")
	event.AffectedSettlements = models.JSONB("[]")
	event.AffectedFactions = models.JSONB("[]")
	event.EconomicImpacts = models.JSONB("{}")
	event.PoliticalImpacts = models.JSONB("{}")
	event.Stages = models.JSONB(`[{"stage": 1, "description": "Initial signs"}, {"stage": 2, "description": "Escalation"}, {"stage": 3, "description": "Resolution"}]`)
	event.ResolutionConditions = models.JSONB(`["Time passes", "Heroes intervene"]`)
	event.Consequences = models.JSONB(`{"ifResolved": "Crisis averted", "ifIgnored": "Situation worsens"}`)
	event.PartyActions = models.JSONB("[]")

	return event
}

// Add missing constant.
const (
	EventReligious models.WorldEventType = "religious"
)

// applyEconomicImpacts handles economic effects of world events.
func (s *WorldEventEngineService) applyEconomicImpacts(ctx context.Context, event *models.WorldEvent) {
	var economicImpacts map[string]interface{}
	if err := json.Unmarshal([]byte(event.EconomicImpacts), &economicImpacts); err != nil || len(economicImpacts) == 0 {
		return
	}

	// Update market conditions in affected settlements.
	var affectedSettlements []string
	if err := json.Unmarshal([]byte(event.AffectedSettlements), &affectedSettlements); err != nil {
		return
	}

	for _, settlementIDStr := range affectedSettlements {
		settlementID, err := uuid.Parse(settlementIDStr)
		if err != nil {
			continue
		}

		s.updateMarketForSettlement(ctx, settlementID, event)
	}
}

// updateMarketForSettlement updates market conditions based on event.
func (s *WorldEventEngineService) updateMarketForSettlement(ctx context.Context, settlementID uuid.UUID, event *models.WorldEvent) {
	market, err := s.worldRepo.GetMarketBySettlement(settlementID)
	if err != nil || market == nil {
		return
	}

	// Apply price modifiers based on event type.
	if event.Type == models.EventEconomic && event.Severity == models.SeverityMajor {
		market.CommonGoodsModifier *= 1.5
		market.FoodPriceModifier *= 1.8
	}

	_ = s.worldRepo.CreateOrUpdateMarket(market)
}

// applyPoliticalImpacts handles political effects of world events.
func (s *WorldEventEngineService) applyPoliticalImpacts(ctx context.Context, event *models.WorldEvent) {
	var politicalImpacts map[string]interface{}
	if err := json.Unmarshal([]byte(event.PoliticalImpacts), &politicalImpacts); err != nil || len(politicalImpacts) == 0 {
		return
	}

	// Update faction relationships.
	s.updateFactionRelationships(ctx, event)
}

// updateFactionRelationships handles faction relationship changes from events.
func (s *WorldEventEngineService) updateFactionRelationships(ctx context.Context, event *models.WorldEvent) {
	var affectedFactions []string
	if err := json.Unmarshal([]byte(event.AffectedFactions), &affectedFactions); err != nil {
		return
	}

	// Conflicts might worsen relationships.
	if event.Type != models.EventPolitical || len(affectedFactions) < 2 {
		return
	}

	faction1ID, err1 := uuid.Parse(affectedFactions[0])
	faction2ID, err2 := uuid.Parse(affectedFactions[1])

	if err1 != nil || err2 != nil || faction1ID == uuid.Nil || faction2ID == uuid.Nil {
		return
	}

	change := -10
	if event.Severity == models.SeverityMajor {
		change = -25
	}
	_ = s.factionService.UpdateFactionRelationship(ctx, faction1ID, faction2ID, change, "world event")
}
