package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// FactionSystemService manages faction creation, relationships, and conflicts
type FactionSystemService struct {
	llmProvider LLMProvider
	worldRepo   WorldBuildingRepository
}

// NewFactionSystemService creates a new faction system service
func NewFactionSystemService(llmProvider LLMProvider, worldRepo WorldBuildingRepository) *FactionSystemService {
	return &FactionSystemService{
		llmProvider: llmProvider,
		worldRepo:   worldRepo,
	}
}

// CreateFaction generates a new faction with goals and relationships
func (s *FactionSystemService) CreateFaction(ctx context.Context, gameSessionID uuid.UUID, req models.FactionCreationRequest) (*models.Faction, error) {
	// Generate faction data using LLM
	factionData, err := s.generateFactionData(ctx, req)
	if err != nil {
		// Fallback to procedural generation
		return s.generateProceduralFaction(gameSessionID, req), nil
	}

	// Build faction model from generated data
	faction := s.buildFactionModel(gameSessionID, req, factionData)

	// Convert fields to JSONB
	s.populateFactionJSONBFields(faction, factionData)

	// Initialize empty relationships and territory
	faction.FactionRelationships = models.JSONB("{}")
	faction.TerritoryControl = models.JSONB("[]")

	// Save the faction
	if err := s.worldRepo.CreateFaction(faction); err != nil {
		return nil, fmt.Errorf("failed to save faction: %w", err)
	}

	// Generate initial relationships with existing factions
	if err := s.generateInitialRelationships(ctx, faction); err != nil {
		// Non-fatal error
		logger.WithContext(ctx).WithError(err).Warn().Msg("Failed to generate faction relationships")
	}

	return faction, nil
}

// factionData represents the generated faction data from LLM
type factionData struct {
	FoundingDate          string   `json:"foundingDate"`
	PublicGoals           []string `json:"publicGoals"`
	SecretGoals           []string `json:"secretGoals"`
	Motivations           []string `json:"motivations"`
	LeadershipStructure   string   `json:"leadershipStructure"`
	HeadquartersLocation  string   `json:"headquartersLocation"`
	AncientKnowledgeLevel int      `json:"ancientKnowledgeLevel"`
	SeeksAncientPower     bool     `json:"seeksAncientPower"`
	GuardsAncientSecrets  bool     `json:"guardsAncientSecrets"`
	Corrupted             bool     `json:"corrupted"`
	Symbols               struct {
		Sigil  string   `json:"sigil"`
		Colors []string `json:"colors"`
		Motto  string   `json:"motto"`
	} `json:"symbols"`
	Rituals   []string               `json:"rituals"`
	Resources map[string]interface{} `json:"resources"`
	Methods   []string               `json:"methods"`
	Enemies   []string               `json:"enemies"`
}

// generateFactionData generates faction data using LLM
func (s *FactionSystemService) generateFactionData(ctx context.Context, req models.FactionCreationRequest) (*factionData, error) {
	systemPrompt := s.getFactionSystemPrompt()
	userPrompt := s.getFactionUserPrompt(req)

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var data factionData
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// getFactionSystemPrompt returns the system prompt for faction generation
func (s *FactionSystemService) getFactionSystemPrompt() string {
	return `You are creating factions for a dark fantasy world where ancient powers still shape events.
Factions might seek forbidden knowledge, guard against ancient evils, or be corrupted by them.
Create complex organizations with both public faces and hidden agendas.`
}

// getFactionUserPrompt returns the user prompt for faction generation
func (s *FactionSystemService) getFactionUserPrompt(req models.FactionCreationRequest) string {
	return fmt.Sprintf(`Create a %s faction with these parameters:
Name: %s
Description: %s
Goals: %v
Ancient Ties: %v

Generate a detailed faction including:
1. Founding history (when and why it was created)
2. Public goals (3-4 that they openly pursue)
3. Secret goals (2-3 hidden agendas)
4. Core motivations driving the faction
5. Leadership structure
6. Resources and methods
7. Symbols and identifying marks
8. Rituals or traditions
9. Connection to ancient powers (if any)

Consider:
- Ancient orders might guard forbidden knowledge
- Cults might seek to awaken old evils
- Political factions might unknowingly serve ancient agendas
- Merchant guilds might traffic in artifacts

Respond in JSON format:
{
  "foundingDate": "description of when founded",
  "publicGoals": ["goal1", "goal2", "goal3"],
  "secretGoals": ["hidden goal1", "hidden goal2"],
  "motivations": ["core drive1", "core drive2"],
  "leadershipStructure": "how the faction is organized",
  "headquartersLocation": "where they operate from",
  "ancientKnowledgeLevel": 0-10,
  "seeksAncientPower": boolean,
  "guardsAncientSecrets": boolean,
  "corrupted": boolean,
  "symbols": {
    "sigil": "description",
    "colors": ["color1", "color2"],
    "motto": "faction motto"
  },
  "rituals": ["ritual1", "ritual2"],
  "resources": {
    "wealth": "description of financial resources",
    "connections": "political/social capital",
    "specialAssets": "unique resources"
  },
  "methods": ["how they operate"],
  "enemies": ["natural enemies or rivals"]
}`,
		req.Type, req.Name, req.Description, req.Goals, req.AncientTies)
}

// buildFactionModel builds a faction model from generated data
func (s *FactionSystemService) buildFactionModel(gameSessionID uuid.UUID, req models.FactionCreationRequest, data *factionData) *models.Faction {
	// Calculate power levels
	influenceLevel := s.calculateInfluenceLevel(req.Type, data.Resources)
	militaryStrength := s.calculateMilitaryStrength(req.Type, data.Resources)
	economicPower := s.calculateEconomicPower(req.Type, data.Resources)
	magicalResources := s.calculateMagicalResources(req.Type, data.AncientKnowledgeLevel)

	return &models.Faction{
		GameSessionID:         gameSessionID,
		Name:                  req.Name,
		Type:                  req.Type,
		Description:           req.Description,
		FoundingDate:          data.FoundingDate,
		AncientKnowledgeLevel: data.AncientKnowledgeLevel,
		SeeksAncientPower:     data.SeeksAncientPower,
		GuardsAncientSecrets:  data.GuardsAncientSecrets,
		Corrupted:             data.Corrupted,
		InfluenceLevel:        influenceLevel,
		MilitaryStrength:      militaryStrength,
		EconomicPower:         economicPower,
		MagicalResources:      magicalResources,
		LeadershipStructure:   data.LeadershipStructure,
		HeadquartersLocation:  data.HeadquartersLocation,
		MemberCount:           s.estimateMemberCount(req.Type, influenceLevel),
	}
}

// populateFactionJSONBFields converts data fields to JSONB
func (s *FactionSystemService) populateFactionJSONBFields(faction *models.Faction, data *factionData) {
	publicGoals, _ := json.Marshal(data.PublicGoals)
	faction.PublicGoals = models.JSONB(publicGoals)

	secretGoals, _ := json.Marshal(data.SecretGoals)
	faction.SecretGoals = models.JSONB(secretGoals)

	motivations, _ := json.Marshal(data.Motivations)
	faction.Motivations = models.JSONB(motivations)

	symbols, _ := json.Marshal(data.Symbols)
	faction.Symbols = models.JSONB(symbols)

	rituals, _ := json.Marshal(data.Rituals)
	faction.Rituals = models.JSONB(rituals)

	resources, _ := json.Marshal(data.Resources)
	faction.Resources = models.JSONB(resources)
}

// getCurrentStanding extracts the current standing from faction relationships
func (s *FactionSystemService) getCurrentStanding(relationships map[string]interface{}, faction2ID uuid.UUID) int {
	rel, exists := relationships[faction2ID.String()]
	if !exists {
		return 0
	}
	
	relMap, ok := rel.(map[string]interface{})
	if !ok {
		return 0
	}
	
	standing, ok := relMap["standing"].(float64)
	if !ok {
		return 0
	}
	
	return int(standing)
}

// clampStanding ensures standing is within valid range [-100, 100]
func (s *FactionSystemService) clampStanding(standing int) int {
	if standing > 100 {
		return 100
	}
	if standing < -100 {
		return -100
	}
	return standing
}

// determineRelationType converts standing value to relationship type
func (s *FactionSystemService) determineRelationType(standing int) string {
	if standing >= 50 {
		return constants.RelationAlly
	}
	if standing <= -50 {
		return constants.RelationEnemy
	}
	return constants.RelationNeutral
}

// UpdateFactionRelationship changes the relationship between two factions
func (s *FactionSystemService) UpdateFactionRelationship(_ context.Context, faction1ID, faction2ID uuid.UUID, change int, _ string) error {
	// Get current relationship
	faction1, err := s.worldRepo.GetFaction(faction1ID)
	if err != nil {
		return err
	}

	var relationships map[string]interface{}
	_ = json.Unmarshal([]byte(faction1.FactionRelationships), &relationships)

	currentStanding := s.getCurrentStanding(relationships, faction2ID)
	newStanding := s.clampStanding(currentStanding + change)
	relationType := s.determineRelationType(newStanding)

	// Update the relationship
	return s.worldRepo.UpdateFactionRelationship(faction1ID, faction2ID, newStanding, relationType)
}

// SimulateFactionConflicts processes faction interactions and conflicts
func (s *FactionSystemService) SimulateFactionConflicts(ctx context.Context, gameSessionID uuid.UUID) ([]*models.WorldEvent, error) {
	factions, err := s.worldRepo.GetFactionsByGameSession(gameSessionID)
	if err != nil {
		return nil, err
	}

	var events []*models.WorldEvent

	// Check each faction pair for potential conflicts
	events = s.processFactionPairs(ctx, gameSessionID, factions, events)

	// Check for faction-specific events
	events = s.processFactionSpecificEvents(ctx, gameSessionID, factions, events)

	return events, nil
}

// processFactionPairs checks faction pairs for conflicts and alliances
func (s *FactionSystemService) processFactionPairs(ctx context.Context, gameSessionID uuid.UUID, factions []*models.Faction, events []*models.WorldEvent) []*models.WorldEvent {
	for i := 0; i < len(factions); i++ {
		for j := i + 1; j < len(factions); j++ {
			events = s.processFactionPair(ctx, gameSessionID, factions[i], factions[j], events)
		}
	}
	return events
}

// processFactionPair processes interactions between two factions
func (s *FactionSystemService) processFactionPair(ctx context.Context, gameSessionID uuid.UUID, faction1, faction2 *models.Faction, events []*models.WorldEvent) []*models.WorldEvent {
	standing := s.getFactionStanding(faction1, faction2.ID)
	
	// Check for conflict
	if event := s.checkAndGenerateConflict(ctx, gameSessionID, faction1, faction2, standing); event != nil {
		events = append(events, event)
	}
	
	// Check for alliance
	if event := s.checkAndGenerateAlliance(ctx, gameSessionID, faction1, faction2, standing); event != nil {
		events = append(events, event)
	}
	
	return events
}

// checkAndGenerateConflict checks conditions and generates conflict event if appropriate
func (s *FactionSystemService) checkAndGenerateConflict(ctx context.Context, gameSessionID uuid.UUID, faction1, faction2 *models.Faction, standing int) *models.WorldEvent {
	if standing < -50 && rand.Float32() < 0.3 {
		event := s.generateConflictEvent(ctx, faction1, faction2)
		if event != nil {
			event.GameSessionID = gameSessionID
			if err := s.worldRepo.CreateWorldEvent(event); err == nil {
				return event
			}
		}
	}
	return nil
}

// checkAndGenerateAlliance checks conditions and generates alliance event if appropriate
func (s *FactionSystemService) checkAndGenerateAlliance(ctx context.Context, gameSessionID uuid.UUID, faction1, faction2 *models.Faction, standing int) *models.WorldEvent {
	if standing > 50 && rand.Float32() < 0.2 {
		event := s.generateAllianceEvent(ctx, faction1, faction2)
		if event != nil {
			event.GameSessionID = gameSessionID
			if err := s.worldRepo.CreateWorldEvent(event); err == nil {
				return event
			}
		}
	}
	return nil
}

// processFactionSpecificEvents generates events for individual factions
func (s *FactionSystemService) processFactionSpecificEvents(ctx context.Context, gameSessionID uuid.UUID, factions []*models.Faction, events []*models.WorldEvent) []*models.WorldEvent {
	for _, faction := range factions {
		if rand.Float32() < 0.1 {
			event := s.generateFactionEvent(ctx, faction)
			if event != nil {
				event.GameSessionID = gameSessionID
				if err := s.worldRepo.CreateWorldEvent(event); err == nil {
					events = append(events, event)
				}
			}
		}
	}
	return events
}

// Helper methods

func (s *FactionSystemService) generateInitialRelationships(_ context.Context, newFaction *models.Faction) error {
	// Get other factions in the same game session
	otherFactions, err := s.worldRepo.GetFactionsByGameSession(newFaction.GameSessionID)
	if err != nil {
		return err
	}

	for _, otherFaction := range otherFactions {
		if otherFaction.ID == newFaction.ID {
			continue
		}

		// Calculate initial standing based on faction types and goals
		standing := s.calculateInitialStanding(newFaction, otherFaction)

		// Add some randomness
		standing += rand.Intn(20) - 10

		relationType := constants.RelationNeutral
		if standing >= 50 {
			relationType = constants.RelationAlly
		} else if standing <= -50 {
			relationType = constants.RelationEnemy
		}

		// Update both factions' relationships
		_ = s.worldRepo.UpdateFactionRelationship(newFaction.ID, otherFaction.ID, standing, relationType)
	}

	return nil
}

func (s *FactionSystemService) calculateInitialStanding(faction1, faction2 *models.Faction) int {
	standing := 0

	// Same type factions might compete
	if faction1.Type == faction2.Type {
		standing -= 20
	}

	// Natural oppositions
	oppositions := map[models.FactionType]models.FactionType{
		models.FactionReligious: models.FactionCult,
		models.FactionCriminal:  models.FactionMilitary,
		models.FactionCult:      models.FactionReligious,
	}

	if oppositions[faction1.Type] == faction2.Type || oppositions[faction2.Type] == faction1.Type {
		standing -= 50
	}

	// Both seek ancient power - conflict
	if faction1.SeeksAncientPower && faction2.SeeksAncientPower {
		standing -= 30
	}

	// One guards, one seeks - major conflict
	if (faction1.GuardsAncientSecrets && faction2.SeeksAncientPower) ||
		(faction2.GuardsAncientSecrets && faction1.SeeksAncientPower) {
		standing -= 70
	}

	// Both corrupted - might ally
	if faction1.Corrupted && faction2.Corrupted {
		standing += 30
	}

	// Merchant factions generally neutral to positive
	if faction1.Type == models.FactionMerchant || faction2.Type == models.FactionMerchant {
		standing += 20
	}

	return standing
}

func (s *FactionSystemService) getFactionStanding(faction *models.Faction, otherFactionID uuid.UUID) int {
	var relationships map[string]interface{}
	_ = json.Unmarshal([]byte(faction.FactionRelationships), &relationships)

	if rel, exists := relationships[otherFactionID.String()]; exists {
		if relMap, ok := rel.(map[string]interface{}); ok {
			if standing, ok := relMap["standing"].(float64); ok {
				return int(standing)
			}
		}
	}

	return 0 // Neutral by default
}

func (s *FactionSystemService) generateConflictEvent(_ context.Context, faction1, faction2 *models.Faction) *models.WorldEvent {
	conflicts := []string{
		"trade war", "territorial dispute", "ideological conflict",
		"resource competition", "assassination attempt", "sabotage",
	}

	conflictType := conflicts[rand.Intn(len(conflicts))]

	event := &models.WorldEvent{
		Name:     fmt.Sprintf("%s between %s and %s", conflictType, faction1.Name, faction2.Name),
		Type:     models.EventPolitical,
		Severity: models.SeverityModerate,
		Description: fmt.Sprintf("Tensions between %s and %s have escalated into open %s",
			faction1.Name, faction2.Name, conflictType),
		Cause:     "Longstanding grievances and incompatible goals",
		StartDate: "Current",
		Duration:  "Ongoing",
		IsActive:  true,
	}

	// If either faction is corrupted or deals with ancient powers, it might escalate
	if faction1.Corrupted || faction2.Corrupted {
		event.AncientCause = true
		event.Severity = models.SeverityMajor
	}

	// Set affected factions
	affectedFactions := []string{faction1.ID.String(), faction2.ID.String()}
	affectedFactionsJSON, _ := json.Marshal(affectedFactions)
	event.AffectedFactions = models.JSONB(affectedFactionsJSON)

	// Empty arrays for other fields
	event.AffectedRegions = models.JSONB("[]")
	event.AffectedSettlements = models.JSONB("[]")
	event.EconomicImpacts = models.JSONB("{}")
	event.PoliticalImpacts = models.JSONB("{}")
	event.Stages = models.JSONB("[]")
	event.ResolutionConditions = models.JSONB("[]")
	event.Consequences = models.JSONB("{}")
	event.PartyActions = models.JSONB("[]")

	return event
}

func (s *FactionSystemService) generateAllianceEvent(_ context.Context, faction1, faction2 *models.Faction) *models.WorldEvent {
	alliances := []string{
		"trade agreement", "mutual defense pact", "intelligence sharing",
		"resource exchange", "joint venture", "marriage alliance",
	}

	allianceType := alliances[rand.Intn(len(alliances))]

	event := &models.WorldEvent{
		Name:        fmt.Sprintf("%s: %s and %s", allianceType, faction1.Name, faction2.Name),
		Type:        models.EventPolitical,
		Severity:    models.SeverityMinor,
		Description: fmt.Sprintf("%s and %s have formed a %s", faction1.Name, faction2.Name, allianceType),
		Cause:       "Mutual interests and shared goals",
		StartDate:   "Current",
		Duration:    "Indefinite",
		IsActive:    true,
	}

	// Set affected factions
	affectedFactions := []string{faction1.ID.String(), faction2.ID.String()}
	affectedFactionsJSON, _ := json.Marshal(affectedFactions)
	event.AffectedFactions = models.JSONB(affectedFactionsJSON)

	// Empty arrays for other fields
	event.AffectedRegions = models.JSONB("[]")
	event.AffectedSettlements = models.JSONB("[]")
	event.EconomicImpacts = models.JSONB("{}")
	event.PoliticalImpacts = models.JSONB("{}")
	event.Stages = models.JSONB("[]")
	event.ResolutionConditions = models.JSONB("[]")
	event.Consequences = models.JSONB("{}")
	event.PartyActions = models.JSONB("[]")

	return event
}

func (s *FactionSystemService) generateFactionEvent(_ context.Context, faction *models.Faction) *models.WorldEvent {
	// Different event types based on faction characteristics
	if faction.Corrupted && rand.Float32() < 0.5 {
		return s.generateCorruptionEvent(faction)
	}

	if faction.SeeksAncientPower && rand.Float32() < 0.4 {
		return s.generateAncientPowerEvent(faction)
	}

	// Default to internal faction events
	events := []string{
		"leadership change", "schism", "major discovery",
		"recruitment drive", "internal purge", "expansion",
	}

	eventType := events[rand.Intn(len(events))]

	event := &models.WorldEvent{
		Name:        fmt.Sprintf("%s: %s", faction.Name, eventType),
		Type:        models.EventPolitical,
		Severity:    models.SeverityMinor,
		Description: fmt.Sprintf("%s is undergoing %s", faction.Name, eventType),
		Cause:       "Internal faction dynamics",
		StartDate:   "Current",
		Duration:    "Several weeks",
		IsActive:    true,
	}

	// Set affected faction
	affectedFactions := []string{faction.ID.String()}
	affectedFactionsJSON, _ := json.Marshal(affectedFactions)
	event.AffectedFactions = models.JSONB(affectedFactionsJSON)

	// Empty arrays for other fields
	event.AffectedRegions = models.JSONB("[]")
	event.AffectedSettlements = models.JSONB("[]")
	event.EconomicImpacts = models.JSONB("{}")
	event.PoliticalImpacts = models.JSONB("{}")
	event.Stages = models.JSONB("[]")
	event.ResolutionConditions = models.JSONB("[]")
	event.Consequences = models.JSONB("{}")
	event.PartyActions = models.JSONB("[]")

	return event
}

func (s *FactionSystemService) generateCorruptionEvent(faction *models.Faction) *models.WorldEvent {
	return &models.WorldEvent{
		Name:                 fmt.Sprintf("Dark Revelation: %s", faction.Name),
		Type:                 models.EventSupernatural,
		Severity:             models.SeverityMajor,
		Description:          fmt.Sprintf("The corruption within %s has manifested in terrifying ways", faction.Name),
		Cause:                "Ancient corruption reaching critical mass",
		StartDate:            "Current",
		Duration:             "Unknown",
		IsActive:             true,
		AncientCause:         true,
		AwakensAncientEvil:   rand.Float32() < 0.3,
		AffectedFactions:     models.JSONB(fmt.Sprintf(`["%s"]`, faction.ID)),
		AffectedRegions:      models.JSONB("[]"),
		AffectedSettlements:  models.JSONB("[]"),
		EconomicImpacts:      models.JSONB("{}"),
		PoliticalImpacts:     models.JSONB("{}"),
		Stages:               models.JSONB("[]"),
		ResolutionConditions: models.JSONB("[]"),
		Consequences:         models.JSONB("{}"),
		PartyActions:         models.JSONB("[]"),
	}
}

func (s *FactionSystemService) generateAncientPowerEvent(faction *models.Faction) *models.WorldEvent {
	return &models.WorldEvent{
		Name:                 fmt.Sprintf("Ancient Discovery: %s", faction.Name),
		Type:                 models.EventSupernatural,
		Severity:             models.SeverityModerate,
		Description:          fmt.Sprintf("%s has uncovered significant ancient artifacts or knowledge", faction.Name),
		Cause:                "Faction research and exploration",
		StartDate:            "Current",
		Duration:             "Ongoing",
		IsActive:             true,
		AncientCause:         true,
		ProphecyRelated:      rand.Float32() < 0.4,
		AffectedFactions:     models.JSONB(fmt.Sprintf(`["%s"]`, faction.ID)),
		AffectedRegions:      models.JSONB("[]"),
		AffectedSettlements:  models.JSONB("[]"),
		EconomicImpacts:      models.JSONB("{}"),
		PoliticalImpacts:     models.JSONB("{}"),
		Stages:               models.JSONB("[]"),
		ResolutionConditions: models.JSONB("[]"),
		Consequences:         models.JSONB("{}"),
		PartyActions:         models.JSONB("[]"),
	}
}

// Power calculation methods

func (s *FactionSystemService) calculateInfluenceLevel(factionType models.FactionType, resources map[string]interface{}) int {
	baseInfluence := map[models.FactionType]int{
		models.FactionReligious:    6,
		models.FactionPolitical:    8,
		models.FactionCriminal:     4,
		models.FactionMerchant:     5,
		models.FactionMilitary:     7,
		models.FactionCult:         3,
		models.FactionAncientOrder: 5,
	}

	influence := baseInfluence[factionType]

	// Adjust based on resources
	if connections, ok := resources["connections"].(string); ok && connections != "" {
		influence++
	}

	// Add randomness
	influence += rand.Intn(3) - 1

	if influence < 1 {
		influence = 1
	}
	if influence > 10 {
		influence = 10
	}

	return influence
}

func (s *FactionSystemService) calculateMilitaryStrength(factionType models.FactionType, _ map[string]interface{}) int {
	baseMilitary := map[models.FactionType]int{
		models.FactionReligious:    3,
		models.FactionPolitical:    5,
		models.FactionCriminal:     4,
		models.FactionMerchant:     2,
		models.FactionMilitary:     9,
		models.FactionCult:         3,
		models.FactionAncientOrder: 4,
	}

	strength := baseMilitary[factionType]

	// Add randomness
	strength += rand.Intn(3) - 1

	if strength < 1 {
		strength = 1
	}
	if strength > 10 {
		strength = 10
	}

	return strength
}

func (s *FactionSystemService) calculateEconomicPower(factionType models.FactionType, resources map[string]interface{}) int {
	baseEconomic := map[models.FactionType]int{
		models.FactionReligious:    5,
		models.FactionPolitical:    6,
		models.FactionCriminal:     6,
		models.FactionMerchant:     9,
		models.FactionMilitary:     4,
		models.FactionCult:         2,
		models.FactionAncientOrder: 3,
	}

	power := baseEconomic[factionType]

	// Adjust based on resources
	if wealth, ok := resources["wealth"].(string); ok && wealth != "" {
		power++
	}

	// Add randomness
	power += rand.Intn(3) - 1

	if power < 1 {
		power = 1
	}
	if power > 10 {
		power = 10
	}

	return power
}

func (s *FactionSystemService) calculateMagicalResources(factionType models.FactionType, ancientKnowledgeLevel int) int {
	baseMagical := map[models.FactionType]int{
		models.FactionReligious:    4,
		models.FactionPolitical:    2,
		models.FactionCriminal:     1,
		models.FactionMerchant:     2,
		models.FactionMilitary:     1,
		models.FactionCult:         6,
		models.FactionAncientOrder: 7,
	}

	magical := baseMagical[factionType]

	// Ancient knowledge greatly increases magical resources
	magical += ancientKnowledgeLevel / 2

	// Add randomness
	magical += rand.Intn(2)

	if magical < 1 {
		magical = 1
	}
	if magical > 10 {
		magical = 10
	}

	return magical
}

func (s *FactionSystemService) estimateMemberCount(factionType models.FactionType, influenceLevel int) int {
	baseMembers := map[models.FactionType]int{
		models.FactionReligious:    500,
		models.FactionPolitical:    200,
		models.FactionCriminal:     100,
		models.FactionMerchant:     300,
		models.FactionMilitary:     1000,
		models.FactionCult:         50,
		models.FactionAncientOrder: 100,
	}

	members := baseMembers[factionType]
	members = int(float64(members) * (float64(influenceLevel) / 5.0))

	// Add variance
	variance := rand.Float64()*0.4 + 0.8 // 0.8 to 1.2
	members = int(float64(members) * variance)

	if members < 10 {
		members = 10
	}

	return members
}

// Procedural fallback

func (s *FactionSystemService) generateProceduralFaction(gameSessionID uuid.UUID, req models.FactionCreationRequest) *models.Faction {
	faction := &models.Faction{
		GameSessionID:         gameSessionID,
		Name:                  req.Name,
		Type:                  req.Type,
		Description:           req.Description,
		FoundingDate:          "Lost to history",
		AncientKnowledgeLevel: rand.Intn(5),
		SeeksAncientPower:     req.AncientTies && rand.Float32() < 0.5,
		GuardsAncientSecrets:  req.AncientTies && rand.Float32() < 0.5,
		Corrupted:             req.AncientTies && rand.Float32() < 0.3,
		InfluenceLevel:        rand.Intn(5) + 3,
		MilitaryStrength:      rand.Intn(5) + 3,
		EconomicPower:         rand.Intn(5) + 3,
		MagicalResources:      rand.Intn(5) + 1,
		LeadershipStructure:   "Hierarchical",
		HeadquartersLocation:  "Unknown",
		MemberCount:           rand.Intn(500) + 100,
	}

	// Set goals from request
	publicGoals, _ := json.Marshal(req.Goals)
	faction.PublicGoals = models.JSONB(publicGoals)

	// Generate secret goals
	secretGoals := []string{"Gain more power", "Eliminate rivals"}
	if req.AncientTies {
		secretGoals = append(secretGoals, "Uncover ancient secrets")
	}
	secretGoalsJSON, _ := json.Marshal(secretGoals)
	faction.SecretGoals = models.JSONB(secretGoalsJSON)

	// Empty fields
	faction.Motivations = models.JSONB(`["survival", "dominance"]`)
	faction.FactionRelationships = models.JSONB("{}")
	faction.TerritoryControl = models.JSONB("[]")
	faction.Symbols = models.JSONB("{}")
	faction.Rituals = models.JSONB("[]")
	faction.Resources = models.JSONB("{}")

	return faction
}
