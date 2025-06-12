package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/google/uuid"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// WorldBuildingRepository interface for world building data operations
type WorldBuildingRepository interface {
	// Settlement operations
	CreateSettlement(settlement *models.Settlement) error
	GetSettlement(id uuid.UUID) (*models.Settlement, error)
	GetSettlementsByGameSession(gameSessionID uuid.UUID) ([]*models.Settlement, error)

	// NPC operations
	CreateSettlementNPC(npc *models.SettlementNPC) error
	GetSettlementNPCs(settlementID uuid.UUID) ([]models.SettlementNPC, error)

	// Shop operations
	CreateSettlementShop(shop *models.SettlementShop) error
	GetSettlementShops(settlementID uuid.UUID) ([]models.SettlementShop, error)

	// Faction operations
	CreateFaction(faction *models.Faction) error
	GetFaction(id uuid.UUID) (*models.Faction, error)
	GetFactionsByGameSession(gameSessionID uuid.UUID) ([]*models.Faction, error)
	UpdateFactionRelationship(faction1ID, faction2ID uuid.UUID, standing int, relationType string) error

	// World Event operations
	CreateWorldEvent(event *models.WorldEvent) error
	GetActiveWorldEvents(gameSessionID uuid.UUID) ([]*models.WorldEvent, error)
	ProgressWorldEvent(eventID uuid.UUID) error

	// Market operations
	CreateOrUpdateMarket(market *models.Market) error
	GetMarketBySettlement(settlementID uuid.UUID) (*models.Market, error)

	// Trade Route operations
	CreateTradeRoute(route *models.TradeRoute) error
	GetTradeRoutesBySettlement(settlementID uuid.UUID) ([]*models.TradeRoute, error)

	// Ancient Site operations
	CreateAncientSite(site *models.AncientSite) error
	GetAncientSitesByGameSession(gameSessionID uuid.UUID) ([]*models.AncientSite, error)

	// Economic simulation
	SimulateEconomicChanges(gameSessionID uuid.UUID) error
}

// SettlementGeneratorService handles AI-powered settlement generation
type SettlementGeneratorService struct {
	llmProvider LLMProvider
	worldRepo   WorldBuildingRepository
}

// NewSettlementGeneratorService creates a new settlement generator service
func NewSettlementGeneratorService(llmProvider LLMProvider, worldRepo WorldBuildingRepository) *SettlementGeneratorService {
	return &SettlementGeneratorService{
		llmProvider: llmProvider,
		worldRepo:   worldRepo,
	}
}

// GenerateSettlement creates a complete settlement with NPCs, shops, and plot hooks
func (s *SettlementGeneratorService) GenerateSettlement(ctx context.Context, gameSessionID uuid.UUID, req models.SettlementGenerationRequest) (*models.Settlement, error) {
	// Determine settlement size if not specified
	if req.PopulationSize == "" {
		req.PopulationSize = s.determinePopulationSize(req.Type)
	}

	// Generate the base settlement
	settlement, err := s.generateBaseSettlement(ctx, gameSessionID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate base settlement: %w", err)
	}

	// Save the settlement
	if err := s.worldRepo.CreateSettlement(settlement); err != nil {
		return nil, fmt.Errorf("failed to save settlement: %w", err)
	}

	// Generate NPCs
	npcCount := s.calculateNPCCount(settlement.Type)
	for i := 0; i < npcCount; i++ {
		npc, err := s.generateNPC(ctx, settlement)
		if err != nil {
			continue
		}
		if err := s.worldRepo.CreateSettlementNPC(npc); err == nil {
			settlement.NPCs = append(settlement.NPCs, *npc)
		}
	}

	// Generate shops
	shopCount := s.calculateShopCount(settlement.Type)
	for i := 0; i < shopCount; i++ {
		shop, err := s.generateShop(ctx, settlement)
		if err != nil {
			continue
		}
		if err := s.worldRepo.CreateSettlementShop(shop); err == nil {
			settlement.Shops = append(settlement.Shops, *shop)
		}
	}

	// Generate initial market conditions
	market := s.generateMarketConditions(settlement)
	s.worldRepo.CreateOrUpdateMarket(market)

	return settlement, nil
}

func (s *SettlementGeneratorService) generateBaseSettlement(ctx context.Context, gameSessionID uuid.UUID, req models.SettlementGenerationRequest) (*models.Settlement, error) {
	systemPrompt := `You are a world-building AI for a dark fantasy setting where ancient evils and forgotten powers shape the world. 
The world has existed for eons, with layers of fallen civilizations and sleeping horrors beneath the surface.
Generate detailed, atmospheric settlements that feel lived-in and connected to this ancient world.`

	userPrompt := fmt.Sprintf(`Generate a %s settlement with the following parameters:
Type: %s
Region: %s
Population Size: %s
Danger Level: %d/10
Ancient Influence: %v
Special Features: %v

Create a detailed settlement including:
1. Name (evocative and fitting the dark fantasy theme)
2. History (including connections to ancient times)
3. Government structure
4. Notable locations (3-5 interesting places)
5. Current problems/tensions (2-3 plot hooks)
6. Secrets (1-2 hidden truths about ancient connections)
7. Defenses
8. Economic focus (exports/imports)

The response should emphasize the weight of history and potential ancient threats.

Respond in JSON format:
{
  "name": "settlement name",
  "description": "atmospheric description",
  "history": "detailed history with ancient connections",
  "governmentType": "type of government",
  "alignment": "general alignment",
  "ageCategory": "ancient/old/established/recent/new",
  "notableLocations": [
    {"name": "location", "description": "details"}
  ],
  "problems": [
    {"title": "problem", "description": "details"}
  ],
  "secrets": [
    {"title": "secret", "description": "details"}
  ],
  "defenses": ["list of defenses"],
  "primaryExports": ["main exports"],
  "primaryImports": ["main imports"],
  "ancientRuinsNearby": boolean,
  "corruptionLevel": 0-10,
  "eldritchInfluence": 0-10
}`,
		req.Type, req.Type, req.Region, req.PopulationSize,
		req.DangerLevel, req.AncientInfluence, strings.Join(req.SpecialFeatures, ", "))

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var generatedData struct {
		Name             string `json:"name"`
		Description      string `json:"description"`
		History          string `json:"history"`
		GovernmentType   string `json:"governmentType"`
		Alignment        string `json:"alignment"`
		AgeCategory      string `json:"ageCategory"`
		NotableLocations []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"notableLocations"`
		Problems []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"problems"`
		Secrets []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"secrets"`
		Defenses           []string `json:"defenses"`
		PrimaryExports     []string `json:"primaryExports"`
		PrimaryImports     []string `json:"primaryImports"`
		AncientRuinsNearby bool     `json:"ancientRuinsNearby"`
		CorruptionLevel    int      `json:"corruptionLevel"`
		EldritchInfluence  int      `json:"eldritchInfluence"`
	}

	if err := json.Unmarshal([]byte(response), &generatedData); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Calculate population based on type and size
	population := s.calculatePopulation(req.Type, req.PopulationSize)

	// Build the settlement model
	settlement := &models.Settlement{
		GameSessionID:      gameSessionID,
		Name:               generatedData.Name,
		Type:               req.Type,
		Population:         population,
		AgeCategory:        generatedData.AgeCategory,
		Description:        generatedData.Description,
		History:            generatedData.History,
		GovernmentType:     generatedData.GovernmentType,
		Alignment:          generatedData.Alignment,
		DangerLevel:        req.DangerLevel,
		CorruptionLevel:    generatedData.CorruptionLevel,
		Region:             req.Region,
		TerrainType:        s.inferTerrainType(req.Region),
		Climate:            s.inferClimate(req.Region),
		WealthLevel:        s.calculateWealthLevel(req.Type, population),
		AncientRuinsNearby: generatedData.AncientRuinsNearby,
		EldritchInfluence:  generatedData.EldritchInfluence,
		LeyLineConnection:  rand.Float32() < 0.3 && generatedData.EldritchInfluence > 5,
	}

	// Convert arrays to JSONB
	notableLocations, _ := json.Marshal(generatedData.NotableLocations)
	settlement.NotableLocations = models.JSONB(notableLocations)

	problems, _ := json.Marshal(generatedData.Problems)
	settlement.Problems = models.JSONB(problems)

	secrets, _ := json.Marshal(generatedData.Secrets)
	settlement.Secrets = models.JSONB(secrets)

	defenses, _ := json.Marshal(generatedData.Defenses)
	settlement.Defenses = models.JSONB(defenses)

	exports, _ := json.Marshal(generatedData.PrimaryExports)
	settlement.PrimaryExports = models.JSONB(exports)

	imports, _ := json.Marshal(generatedData.PrimaryImports)
	settlement.PrimaryImports = models.JSONB(imports)

	// Empty trade routes initially
	settlement.TradeRoutes = models.JSONB("[]")

	// Generate coordinates (simplified - would be more complex in production)
	coords := map[string]int{
		"x": rand.Intn(1000),
		"y": rand.Intn(1000),
	}
	coordsJSON, _ := json.Marshal(coords)
	settlement.Coordinates = models.JSONB(coordsJSON)

	return settlement, nil
}

func (s *SettlementGeneratorService) generateNPC(ctx context.Context, settlement *models.Settlement) (*models.SettlementNPC, error) {
	// Determine NPC role based on settlement needs
	roles := s.getNPCRoles(settlement.Type)
	role := roles[rand.Intn(len(roles))]

	systemPrompt := `You are creating NPCs for a dark fantasy world where ancient powers still influence daily life.
Create interesting, complex characters with hidden depths and potential connections to the old world.`

	userPrompt := fmt.Sprintf(`Create an NPC for the settlement of %s (%s).
Settlement has corruption level %d and eldritch influence %d.
NPC Role: %s

Generate:
1. Name and basic details (race, class, level 1-10)
2. Personality traits (2-3)
3. Ideals, bonds, and flaws (1-2 each)
4. Potential ancient connections or secrets
5. Plot hooks involving this NPC

Consider that some NPCs might be:
- Secretly ancient beings in disguise
- Touched by corruption
- Guardians of old knowledge
- Cultists or their opponents

Respond in JSON format:
{
  "name": "full name",
  "race": "race",
  "class": "class or profession",
  "level": 1-10,
  "occupation": "specific job",
  "personalityTraits": ["trait1", "trait2"],
  "ideals": ["ideal1"],
  "bonds": ["bond1"],
  "flaws": ["flaw1"],
  "ancientKnowledge": boolean,
  "corruptionTouched": boolean,
  "secretAgenda": "hidden goal if any",
  "plotHooks": ["hook1", "hook2"],
  "trueAge": null or number if secretly ancient
}`,
		settlement.Name, settlement.Type, settlement.CorruptionLevel,
		settlement.EldritchInfluence, role)

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		// Fallback to procedural generation
		return s.generateProceduralNPC(settlement, role), nil
	}

	var npcData struct {
		Name              string   `json:"name"`
		Race              string   `json:"race"`
		Class             string   `json:"class"`
		Level             int      `json:"level"`
		Occupation        string   `json:"occupation"`
		PersonalityTraits []string `json:"personalityTraits"`
		Ideals            []string `json:"ideals"`
		Bonds             []string `json:"bonds"`
		Flaws             []string `json:"flaws"`
		AncientKnowledge  bool     `json:"ancientKnowledge"`
		CorruptionTouched bool     `json:"corruptionTouched"`
		SecretAgenda      string   `json:"secretAgenda"`
		PlotHooks         []string `json:"plotHooks"`
		TrueAge           *int     `json:"trueAge"`
	}

	if err := json.Unmarshal([]byte(response), &npcData); err != nil {
		return s.generateProceduralNPC(settlement, role), nil
	}

	npc := &models.SettlementNPC{
		SettlementID:      settlement.ID,
		Name:              npcData.Name,
		Race:              npcData.Race,
		Class:             npcData.Class,
		Level:             npcData.Level,
		Role:              role,
		Occupation:        npcData.Occupation,
		AncientKnowledge:  npcData.AncientKnowledge,
		CorruptionTouched: npcData.CorruptionTouched,
		SecretAgenda:      npcData.SecretAgenda,
		TrueAge:           npcData.TrueAge,
	}

	// Convert arrays to JSONB
	traits, _ := json.Marshal(npcData.PersonalityTraits)
	npc.PersonalityTraits = models.JSONB(traits)

	ideals, _ := json.Marshal(npcData.Ideals)
	npc.Ideals = models.JSONB(ideals)

	bonds, _ := json.Marshal(npcData.Bonds)
	npc.Bonds = models.JSONB(bonds)

	flaws, _ := json.Marshal(npcData.Flaws)
	npc.Flaws = models.JSONB(flaws)

	plotHooks, _ := json.Marshal(npcData.PlotHooks)
	npc.PlotHooks = models.JSONB(plotHooks)

	// Empty fields
	npc.FactionAffiliations = models.JSONB("[]")
	npc.Relationships = models.JSONB("{}")
	npc.Stats = models.JSONB("{}")
	npc.Skills = models.JSONB("{}")
	npc.Inventory = models.JSONB("[]")

	return npc, nil
}

func (s *SettlementGeneratorService) generateShop(ctx context.Context, settlement *models.Settlement) (*models.SettlementShop, error) {
	// Determine shop type based on settlement
	shopTypes := s.getShopTypes(settlement.Type)
	shopType := shopTypes[rand.Intn(len(shopTypes))]

	systemPrompt := `You are creating shops for a dark fantasy world. Some shops might deal in forbidden goods,
ancient artifacts, or have connections to the old powers. Create atmospheric, memorable establishments.`

	userPrompt := fmt.Sprintf(`Create a %s shop for the settlement of %s.
The settlement has wealth level %d and %v ancient ruins nearby.

Generate:
1. Shop name (evocative and memorable)
2. Quality level (1-10)
3. Special features (black market, ancient artifacts, etc.)
4. Current rumors or plot hooks (2-3)
5. Notable items or services

Consider:
- Shops near ancient ruins might trade in artifacts
- Corrupted areas might have black markets
- Some shops might be fronts for cults

Respond in JSON format:
{
  "name": "shop name",
  "qualityLevel": 1-10,
  "priceModifier": 0.5-2.0,
  "blackMarket": boolean,
  "ancientArtifacts": boolean,
  "specialItems": ["item1", "item2"],
  "currentRumors": ["rumor1", "rumor2"],
  "craftingSpecialties": ["specialty1"] or []
}`,
		shopType, settlement.Name, settlement.WealthLevel, settlement.AncientRuinsNearby)

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		// Fallback to procedural generation
		return s.generateProceduralShop(settlement, shopType), nil
	}

	var shopData struct {
		Name                string   `json:"name"`
		QualityLevel        int      `json:"qualityLevel"`
		PriceModifier       float64  `json:"priceModifier"`
		BlackMarket         bool     `json:"blackMarket"`
		AncientArtifacts    bool     `json:"ancientArtifacts"`
		SpecialItems        []string `json:"specialItems"`
		CurrentRumors       []string `json:"currentRumors"`
		CraftingSpecialties []string `json:"craftingSpecialties"`
	}

	if err := json.Unmarshal([]byte(response), &shopData); err != nil {
		return s.generateProceduralShop(settlement, shopType), nil
	}

	shop := &models.SettlementShop{
		SettlementID:       settlement.ID,
		Name:               shopData.Name,
		Type:               shopType,
		QualityLevel:       shopData.QualityLevel,
		PriceModifier:      shopData.PriceModifier,
		BlackMarket:        shopData.BlackMarket,
		AncientArtifacts:   shopData.AncientArtifacts,
		CanCraft:           len(shopData.CraftingSpecialties) > 0,
		ReputationRequired: 0,
	}

	// Convert arrays to JSONB
	specialItems, _ := json.Marshal(shopData.SpecialItems)
	shop.SpecialItems = models.JSONB(specialItems)

	rumors, _ := json.Marshal(shopData.CurrentRumors)
	shop.CurrentRumors = models.JSONB(rumors)

	crafting, _ := json.Marshal(shopData.CraftingSpecialties)
	shop.CraftingSpecialties = models.JSONB(crafting)

	// Empty fields
	shop.AvailableItems = models.JSONB("[]")
	shop.FactionDiscount = models.JSONB("{}")
	shop.OperatingHours = models.JSONB(`{"open": "dawn", "close": "dusk"}`)

	return shop, nil
}

// Helper functions

func (s *SettlementGeneratorService) determinePopulationSize(settlementType models.SettlementType) string {
	switch settlementType {
	case models.SettlementHamlet:
		return "small"
	case models.SettlementVillage:
		return "small"
	case models.SettlementTown:
		return "medium"
	case models.SettlementCity:
		return "large"
	case models.SettlementMetropolis:
		return "large"
	default:
		return "small"
	}
}

func (s *SettlementGeneratorService) calculatePopulation(settlementType models.SettlementType, size string) int {
	basePopulation := map[models.SettlementType]int{
		models.SettlementHamlet:     50,
		models.SettlementVillage:    200,
		models.SettlementTown:       1000,
		models.SettlementCity:       5000,
		models.SettlementMetropolis: 25000,
		models.SettlementRuins:      0,
	}

	sizeMultiplier := map[string]float64{
		"small":  0.5,
		"medium": 1.0,
		"large":  2.0,
	}

	base := basePopulation[settlementType]
	multiplier := sizeMultiplier[size]

	// Add some randomness
	variance := rand.Float64()*0.4 + 0.8 // 0.8 to 1.2

	return int(float64(base) * multiplier * variance)
}

func (s *SettlementGeneratorService) calculateNPCCount(settlementType models.SettlementType) int {
	npcCounts := map[models.SettlementType]int{
		models.SettlementHamlet:     3,
		models.SettlementVillage:    5,
		models.SettlementTown:       8,
		models.SettlementCity:       12,
		models.SettlementMetropolis: 20,
		models.SettlementRuins:      1,
	}
	return npcCounts[settlementType] + rand.Intn(3)
}

func (s *SettlementGeneratorService) calculateShopCount(settlementType models.SettlementType) int {
	shopCounts := map[models.SettlementType]int{
		models.SettlementHamlet:     1,
		models.SettlementVillage:    2,
		models.SettlementTown:       4,
		models.SettlementCity:       8,
		models.SettlementMetropolis: 15,
		models.SettlementRuins:      0,
	}
	return shopCounts[settlementType] + rand.Intn(2)
}

func (s *SettlementGeneratorService) calculateWealthLevel(settlementType models.SettlementType, population int) int {
	baseWealth := map[models.SettlementType]int{
		models.SettlementHamlet:     2,
		models.SettlementVillage:    3,
		models.SettlementTown:       5,
		models.SettlementCity:       7,
		models.SettlementMetropolis: 9,
		models.SettlementRuins:      1,
	}

	wealth := baseWealth[settlementType]
	// Larger populations tend to be wealthier
	if population > 10000 {
		wealth++
	}

	// Add some randomness
	wealth += rand.Intn(3) - 1 // -1 to +1

	if wealth < 1 {
		wealth = 1
	}
	if wealth > 10 {
		wealth = 10
	}

	return wealth
}

func (s *SettlementGeneratorService) inferTerrainType(region string) string {
	// Simple inference based on region name
	region = strings.ToLower(region)

	switch {
	case strings.Contains(region, "mountain"):
		return "mountainous"
	case strings.Contains(region, "forest"):
		return "forest"
	case strings.Contains(region, "desert"):
		return "desert"
	case strings.Contains(region, "coast"):
		return "coastal"
	case strings.Contains(region, "swamp"):
		return "swamp"
	case strings.Contains(region, "plain"):
		return "plains"
	default:
		return "varied"
	}
}

func (s *SettlementGeneratorService) inferClimate(region string) string {
	region = strings.ToLower(region)

	switch {
	case strings.Contains(region, "north"):
		return "cold"
	case strings.Contains(region, "south"):
		return "tropical"
	case strings.Contains(region, "desert"):
		return "arid"
	case strings.Contains(region, "swamp"):
		return "humid"
	default:
		return "temperate"
	}
}

func (s *SettlementGeneratorService) getNPCRoles(settlementType models.SettlementType) []string {
	baseRoles := []string{"merchant", "guard", "innkeeper", "blacksmith", "priest"}

	additionalRoles := map[models.SettlementType][]string{
		models.SettlementHamlet:     {"farmer", "hunter"},
		models.SettlementVillage:    {"healer", "elder", "miller"},
		models.SettlementTown:       {"mayor", "captain", "scholar", "alchemist"},
		models.SettlementCity:       {"noble", "guildmaster", "magistrate", "spy", "cultist"},
		models.SettlementMetropolis: {"archmage", "high priest", "crime lord", "diplomat"},
		models.SettlementRuins:      {"hermit", "scavenger", "mad prophet"},
	}

	return append(baseRoles, additionalRoles[settlementType]...)
}

func (s *SettlementGeneratorService) getShopTypes(settlementType models.SettlementType) []string {
	baseShops := []string{"general", "tavern"}

	additionalShops := map[models.SettlementType][]string{
		models.SettlementHamlet:     {},
		models.SettlementVillage:    {"weaponsmith", "inn"},
		models.SettlementTown:       {"armorer", "alchemist", "temple"},
		models.SettlementCity:       {"magic", "jeweler", "library", "herbalist"},
		models.SettlementMetropolis: {"enchanter", "artificer", "grand bazaar", "auction house"},
		models.SettlementRuins:      {},
	}

	return append(baseShops, additionalShops[settlementType]...)
}

func (s *SettlementGeneratorService) generateMarketConditions(settlement *models.Settlement) *models.Market {
	market := &models.Market{
		SettlementID:             settlement.ID,
		FoodPriceModifier:        1.0,
		CommonGoodsModifier:      1.0,
		WeaponsArmorModifier:     1.0,
		MagicalItemsModifier:     1.0,
		AncientArtifactsModifier: 2.0,
	}

	// Adjust based on settlement characteristics
	if settlement.WealthLevel < 3 {
		market.CommonGoodsModifier *= 1.2
		market.MagicalItemsModifier *= 1.5
	} else if settlement.WealthLevel > 7 {
		market.CommonGoodsModifier *= 0.9
		market.MagicalItemsModifier *= 0.8
	}

	// Ancient ruins increase artifact availability but also price
	if settlement.AncientRuinsNearby {
		market.ArtifactDealerPresent = rand.Float32() < 0.6
		market.AncientArtifactsModifier *= 0.8 // Slightly cheaper due to supply
	}

	// Corruption affects black market
	if settlement.CorruptionLevel > 5 {
		market.BlackMarketActive = true
	}

	// Random economic conditions
	if rand.Float32() < 0.1 {
		market.EconomicBoom = true
		market.CommonGoodsModifier *= 0.9
	} else if rand.Float32() < 0.1 {
		market.EconomicDepression = true
		market.CommonGoodsModifier *= 1.3
	}

	// Empty arrays for now
	market.HighDemandItems = models.JSONB("[]")
	market.SurplusItems = models.JSONB("[]")
	market.BannedItems = models.JSONB("[]")

	return market
}

// Procedural fallback generators

func (s *SettlementGeneratorService) generateProceduralNPC(settlement *models.Settlement, role string) *models.SettlementNPC {
	names := []string{"Aldric", "Mira", "Thorne", "Elara", "Grimm", "Lyssa", "Darius", "Nyx"}
	races := []string{"human", "dwarf", "elf", "halfling", "tiefling", "half-orc"}

	npc := &models.SettlementNPC{
		SettlementID:      settlement.ID,
		Name:              names[rand.Intn(len(names))],
		Race:              races[rand.Intn(len(races))],
		Class:             "commoner",
		Level:             rand.Intn(5) + 1,
		Role:              role,
		Occupation:        role,
		AncientKnowledge:  rand.Float32() < 0.1,
		CorruptionTouched: settlement.CorruptionLevel > 5 && rand.Float32() < 0.2,
	}

	// Generate basic traits
	traits := []string{"gruff", "friendly", "suspicious", "weary", "jovial", "secretive"}
	selectedTraits := []string{traits[rand.Intn(len(traits))]}
	traitsJSON, _ := json.Marshal(selectedTraits)
	npc.PersonalityTraits = models.JSONB(traitsJSON)

	// Empty other fields
	npc.Ideals = models.JSONB("[]")
	npc.Bonds = models.JSONB("[]")
	npc.Flaws = models.JSONB("[]")
	npc.FactionAffiliations = models.JSONB("[]")
	npc.Relationships = models.JSONB("{}")
	npc.Stats = models.JSONB("{}")
	npc.Skills = models.JSONB("{}")
	npc.Inventory = models.JSONB("[]")
	npc.PlotHooks = models.JSONB("[]")

	return npc
}

func (s *SettlementGeneratorService) generateProceduralShop(settlement *models.Settlement, shopType string) *models.SettlementShop {
	prefixes := []string{"The", "Ye Olde", "The Ancient", "The Rusty", "The Golden"}
	suffixes := map[string][]string{
		"general":     {"Trading Post", "Emporium", "Supply"},
		"tavern":      {"Inn", "Alehouse", "Rest"},
		"weaponsmith": {"Forge", "Arms", "Steel"},
		"armorer":     {"Protection", "Plate", "Defense"},
		"alchemist":   {"Cauldron", "Elixirs", "Potions"},
		"magic":       {"Arcanum", "Mysteries", "Enchantments"},
	}

	shopSuffixes := suffixes[shopType]
	if shopSuffixes == nil {
		shopSuffixes = []string{"Shop", "Store", "Market"}
	}

	shop := &models.SettlementShop{
		SettlementID:     settlement.ID,
		Name:             fmt.Sprintf("%s %s", prefixes[rand.Intn(len(prefixes))], shopSuffixes[rand.Intn(len(shopSuffixes))]),
		Type:             shopType,
		QualityLevel:     rand.Intn(5) + 3,
		PriceModifier:    0.9 + rand.Float64()*0.3,
		BlackMarket:      settlement.CorruptionLevel > 6 && rand.Float32() < 0.3,
		AncientArtifacts: settlement.AncientRuinsNearby && rand.Float32() < 0.4,
	}

	// Empty arrays
	shop.AvailableItems = models.JSONB("[]")
	shop.SpecialItems = models.JSONB("[]")
	shop.CraftingSpecialties = models.JSONB("[]")
	shop.FactionDiscount = models.JSONB("{}")
	shop.OperatingHours = models.JSONB(`{"open": "dawn", "close": "dusk"}`)
	shop.CurrentRumors = models.JSONB("[]")

	return shop
}
