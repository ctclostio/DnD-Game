package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
	"github.com/google/uuid"
)

// EconomicSimulatorService manages market dynamics and trade
type EconomicSimulatorService struct {
	worldRepo WorldBuildingRepository
}

// NewEconomicSimulatorService creates a new economic simulator service
func NewEconomicSimulatorService(worldRepo WorldBuildingRepository) *EconomicSimulatorService {
	return &EconomicSimulatorService{
		worldRepo: worldRepo,
	}
}

// SimulateEconomicCycle updates all market conditions based on various factors
func (s *EconomicSimulatorService) SimulateEconomicCycle(ctx context.Context, gameSessionID uuid.UUID) error {
	// Get all settlements
	settlements, err := s.worldRepo.GetSettlementsByGameSession(gameSessionID)
	if err != nil {
		return err
	}

	// Get active world events
	activeEvents, err := s.worldRepo.GetActiveWorldEvents(gameSessionID)
	if err != nil {
		return err
	}

	// Update each settlement's market
	for _, settlement := range settlements {
		if err := s.updateSettlementMarket(ctx, settlement, activeEvents); err != nil {
			logger.WithContext(ctx).WithError(err).WithField("settlement_name", settlement.Name).Warn().Msg("Failed to update market")
		}
	}

	// Update trade route conditions
	if err := s.updateTradeRoutes(ctx, gameSessionID); err != nil {
		logger.WithContext(ctx).WithError(err).Warn().Msg("Failed to update trade routes")
	}

	// Trigger economic events if conditions warrant
	if s.checkForEconomicCrisis(settlements) {
		s.triggerEconomicEvent(ctx, gameSessionID, "crisis")
	} else if s.checkForEconomicBoom(settlements) {
		s.triggerEconomicEvent(ctx, gameSessionID, "boom")
	}

	return nil
}

// CreateTradeRoute establishes a new trade route between settlements
func (s *EconomicSimulatorService) CreateTradeRoute(ctx context.Context, startSettlementID, endSettlementID uuid.UUID) (*models.TradeRoute, error) {
	// Get both settlements
	startSettlement, err := s.worldRepo.GetSettlement(startSettlementID)
	if err != nil {
		return nil, fmt.Errorf("start settlement not found: %w", err)
	}

	endSettlement, err := s.worldRepo.GetSettlement(endSettlementID)
	if err != nil {
		return nil, fmt.Errorf("end settlement not found: %w", err)
	}

	// Calculate route properties
	distance := s.calculateDistance(startSettlement, endSettlement)
	difficulty := s.calculateRouteDifficulty(startSettlement, endSettlement)
	routeType := s.determineRouteType(startSettlement, endSettlement)

	// Determine traded goods based on settlement economies
	primaryGoods := s.determineTradedGoods(startSettlement, endSettlement)

	route := &models.TradeRoute{
		GameSessionID:      startSettlement.GameSessionID,
		Name:               fmt.Sprintf("%s-%s Trade Route", startSettlement.Name, endSettlement.Name),
		StartSettlementID:  startSettlementID,
		EndSettlementID:    endSettlementID,
		RouteType:          routeType,
		Distance:           distance,
		DifficultyRating:   difficulty,
		BanditThreatLevel:  s.calculateBanditThreat(startSettlement, endSettlement),
		MonsterThreatLevel: s.calculateMonsterThreat(startSettlement, endSettlement),
		AncientHazards:     s.checkForAncientHazards(startSettlement, endSettlement),
		TradeVolume:        s.calculateInitialTradeVolume(startSettlement, endSettlement),
		TariffRate:         0.1, // 10% default
		IsActive:           true,
	}

	// Set traded goods
	goodsJSON, _ := json.Marshal(primaryGoods)
	route.PrimaryGoods = models.JSONB(goodsJSON)

	// Environmental hazards based on terrain
	hazards := s.determineEnvironmentalHazards(startSettlement, endSettlement)
	hazardsJSON, _ := json.Marshal(hazards)
	route.EnvironmentalHazards = models.JSONB(hazardsJSON)

	// Empty disruption events initially
	route.DisruptionEvents = models.JSONB("[]")

	// Save the route
	if err := s.worldRepo.CreateTradeRoute(route); err != nil {
		return nil, fmt.Errorf("failed to create trade route: %w", err)
	}

	// Update settlement trade connections
	s.updateSettlementTradeConnections(startSettlement, endSettlement, route.ID)

	return route, nil
}

// CalculateItemPrice determines the price of an item in a specific market
func (s *EconomicSimulatorService) CalculateItemPrice(settlementID uuid.UUID, basePrice float64, itemType string) (float64, error) {
	market, err := s.worldRepo.GetMarketBySettlement(settlementID)
	if err != nil {
		// Default pricing if no market data
		return basePrice, nil
	}

	// Apply type-specific modifiers
	modifier := market.CommonGoodsModifier
	switch itemType {
	case "food", "rations":
		modifier = market.FoodPriceModifier
	case "weapon":
		modifier = market.WeaponsArmorModifier
	case "armor":
		modifier = market.WeaponsArmorModifier
	case "magic":
		modifier = market.MagicalItemsModifier
	case "artifact":
		modifier = market.AncientArtifactsModifier
	}

	// Check if item is in high demand or surplus
	var highDemand []string
	var surplus []string
	_ = json.Unmarshal([]byte(market.HighDemandItems), &highDemand)
	_ = json.Unmarshal([]byte(market.SurplusItems), &surplus)

	for _, item := range highDemand {
		if item == itemType {
			modifier *= 1.5 // 50% more expensive
			break
		}
	}

	for _, item := range surplus {
		if item == itemType {
			modifier *= 0.7 // 30% cheaper
			break
		}
	}

	// Economic conditions
	if market.EconomicBoom {
		modifier *= 0.9 // Prices slightly lower in boom
	}
	if market.EconomicDepression {
		modifier *= 1.3 // Prices higher in depression
	}

	// Add some random market fluctuation (±5%)
	fluctuation := 0.95 + (rand.Float64() * 0.1)
	modifier *= fluctuation

	return basePrice * modifier, nil
}

// DisruptTradeRoute applies a disruption to a trade route
func (s *EconomicSimulatorService) DisruptTradeRoute(routeID uuid.UUID, disruptionType string, severity int) error {
	// This would update the trade route with disruption events
	// and affect connected settlement markets
	return nil
}

// Helper methods

func (s *EconomicSimulatorService) updateSettlementMarket(ctx context.Context, settlement *models.Settlement, activeEvents []*models.WorldEvent) error {
	market, err := s.worldRepo.GetMarketBySettlement(settlement.ID)
	if err != nil {
		// Create new market if none exists
		market = &models.Market{
			SettlementID:             settlement.ID,
			FoodPriceModifier:        1.0,
			CommonGoodsModifier:      1.0,
			WeaponsArmorModifier:     1.0,
			MagicalItemsModifier:     1.0,
			AncientArtifactsModifier: 2.0,
		}
	}

	// Base adjustments based on settlement properties
	s.applySettlementFactors(market, settlement)

	// Apply world event effects
	s.applyEventEffects(market, settlement, activeEvents)

	// Supply and demand simulation
	s.simulateSupplyDemand(market, settlement)

	// Trade route effects
	s.applyTradeRouteEffects(market, settlement)

	// Random market events
	if rand.Float32() < 0.1 {
		s.applyRandomMarketEvent(market)
	}

	// Save updated market
	return s.worldRepo.CreateOrUpdateMarket(market)
}

func (s *EconomicSimulatorService) applySettlementFactors(market *models.Market, settlement *models.Settlement) {
	// Wealth affects general prices based on settlement level
	wealthFactor := float64(settlement.WealthLevel) / 5.0
	if wealthFactor < 0.5 {
		wealthFactor = 0.5
	} else if wealthFactor > 1.5 {
		wealthFactor = 1.5
	}
	_ = wealthFactor // currently unused but kept for future balancing logic

	// Poor settlements have higher prices due to scarcity
	if settlement.WealthLevel < 3 {
		market.CommonGoodsModifier *= 1.1
		market.FoodPriceModifier *= 1.2
	}

	// Corruption affects black market
	if settlement.CorruptionLevel > 5 {
		market.BlackMarketActive = true
		// Black market reduces "official" prices slightly
		market.CommonGoodsModifier *= 0.95
	}

	// Ancient ruins affect artifact trade
	if settlement.AncientRuinsNearby {
		market.ArtifactDealerPresent = rand.Float32() < 0.7
		market.AncientArtifactsModifier *= 0.8 // More common
		market.MagicalItemsModifier *= 0.9
	}

	// Population affects supply/demand
	if settlement.Population > 10000 {
		// Large cities have more stable prices
		market.CommonGoodsModifier = (market.CommonGoodsModifier + 1.0) / 2.0
		market.FoodPriceModifier = (market.FoodPriceModifier + 1.0) / 2.0
	}
}

func (s *EconomicSimulatorService) applyEventEffects(market *models.Market, settlement *models.Settlement, events []*models.WorldEvent) {
	for _, event := range events {
		// Check if settlement is affected
		var affectedSettlements []string
		_ = json.Unmarshal([]byte(event.AffectedSettlements), &affectedSettlements)

		isAffected := false
		for _, affectedID := range affectedSettlements {
			if affectedID == settlement.ID.String() {
				isAffected = true
				break
			}
		}

		if !isAffected {
			continue
		}

		// Apply event-specific effects
		switch event.Type {
		case models.EventEconomic:
			if event.Severity == models.SeverityMajor {
				market.EconomicDepression = true
				market.CommonGoodsModifier *= 1.4
				market.FoodPriceModifier *= 1.6
			}
		case models.EventNatural:
			// Natural disasters affect food supply
			market.FoodPriceModifier *= 1.3
		case models.EventPolitical:
			// Political instability affects trade
			market.CommonGoodsModifier *= 1.1
			market.WeaponsArmorModifier *= 1.2 // Increased demand for protection
		case models.EventAncientAwakening:
			// Ancient events increase demand for magical protection
			market.MagicalItemsModifier *= 0.8     // Cheaper due to desperation
			market.AncientArtifactsModifier *= 1.5 // But artifacts are riskier
		}
	}
}

func (s *EconomicSimulatorService) simulateSupplyDemand(market *models.Market, settlement *models.Settlement) {
	highDemand := []string{}
	surplus := []string{}

	// Based on settlement type and conditions
	switch settlement.Type {
	case models.SettlementCity, models.SettlementMetropolis:
		// Cities need food imports
		highDemand = append(highDemand, "food", "raw materials")
		surplus = append(surplus, "manufactured goods", "services")
	case models.SettlementVillage, models.SettlementHamlet:
		// Rural areas produce food
		surplus = append(surplus, "food", "raw materials")
		highDemand = append(highDemand, "tools", "manufactured goods")
	}

	// Danger creates demand for weapons
	if settlement.DangerLevel > 5 {
		highDemand = append(highDemand, "weapons", "armor")
	}

	// Corruption creates demand for certain items
	if settlement.CorruptionLevel > 5 {
		highDemand = append(highDemand, "holy water", "silver weapons")
	}

	demandJSON, _ := json.Marshal(highDemand)
	market.HighDemandItems = models.JSONB(demandJSON)

	surplusJSON, _ := json.Marshal(surplus)
	market.SurplusItems = models.JSONB(surplusJSON)
}

func (s *EconomicSimulatorService) applyTradeRouteEffects(market *models.Market, settlement *models.Settlement) {
	routes, err := s.worldRepo.GetTradeRoutesBySettlement(settlement.ID)
	if err != nil || len(routes) == 0 {
		return
	}

	// More trade routes = more stable prices
	tradeStabilization := 1.0 - (float64(len(routes)) * 0.05)
	if tradeStabilization < 0.7 {
		tradeStabilization = 0.7
	}

	// Move prices toward baseline
	market.CommonGoodsModifier = market.CommonGoodsModifier*tradeStabilization + 1.0*(1-tradeStabilization)
	market.FoodPriceModifier = market.FoodPriceModifier*tradeStabilization + 1.0*(1-tradeStabilization)

	// Active trade routes reduce scarcity
	activeRoutes := 0
	for _, route := range routes {
		if route.IsActive {
			activeRoutes++
		}
	}

	if activeRoutes > 2 {
		market.EconomicBoom = rand.Float32() < 0.3
	}
}

func (s *EconomicSimulatorService) applyRandomMarketEvent(market *models.Market) {
	events := []string{
		"merchant_caravan", "shortage", "surplus", "speculation", "new_discovery",
	}

	event := events[rand.Intn(len(events))]

	switch event {
	case "merchant_caravan":
		// Temporary price reduction
		market.CommonGoodsModifier *= 0.9
	case "shortage":
		// Random item shortage
		itemTypes := []string{"food", "weapons", "potions", "materials"}
		shortage := itemTypes[rand.Intn(len(itemTypes))]
		var highDemand []string
		_ = json.Unmarshal([]byte(market.HighDemandItems), &highDemand)
		highDemand = append(highDemand, shortage)
		demandJSON, _ := json.Marshal(highDemand)
		market.HighDemandItems = models.JSONB(demandJSON)
	case "surplus":
		// Random item surplus
		market.CommonGoodsModifier *= 0.95
	case "speculation":
		// Price volatility
		market.CommonGoodsModifier *= 0.9 + rand.Float64()*0.3
	case "new_discovery":
		// New source of goods
		if market.ArtifactDealerPresent {
			market.AncientArtifactsModifier *= 0.9
		}
	}
}

func (s *EconomicSimulatorService) updateTradeRoutes(ctx context.Context, gameSessionID uuid.UUID) error {
	// This would update all trade routes based on current conditions
	// Check for disruptions, update trade volumes, etc.
	return nil
}

func (s *EconomicSimulatorService) checkForEconomicCrisis(settlements []*models.Settlement) bool {
	// Check if overall economic conditions warrant a crisis event
	depressionCount := 0

	for _, settlement := range settlements {
		market, err := s.worldRepo.GetMarketBySettlement(settlement.ID)
		if err != nil {
			continue
		}

		if market.EconomicDepression {
			depressionCount++
		}
	}

	// If more than half of settlements are in depression
	return float64(depressionCount) > float64(len(settlements))*0.5
}

func (s *EconomicSimulatorService) checkForEconomicBoom(settlements []*models.Settlement) bool {
	// Check if overall economic conditions warrant a boom event
	boomCount := 0

	for _, settlement := range settlements {
		market, err := s.worldRepo.GetMarketBySettlement(settlement.ID)
		if err != nil {
			continue
		}

		if market.EconomicBoom {
			boomCount++
		}
	}

	// If more than half of settlements are booming
	return float64(boomCount) > float64(len(settlements))*0.5
}

func (s *EconomicSimulatorService) triggerEconomicEvent(ctx context.Context, gameSessionID uuid.UUID, eventType string) {
	// This would create a world event for economic conditions
	// Would integrate with WorldEventEngine
}

// Trade route calculation helpers

func (s *EconomicSimulatorService) calculateDistance(start, end *models.Settlement) int {
	// Simplified distance calculation based on coordinates
	var startCoords, endCoords map[string]int
	_ = json.Unmarshal([]byte(start.Coordinates), &startCoords)
	_ = json.Unmarshal([]byte(end.Coordinates), &endCoords)

	dx := float64(endCoords["x"] - startCoords["x"])
	dy := float64(endCoords["y"] - startCoords["y"])

	distance := int(math.Sqrt(dx*dx + dy*dy))

	// Convert to travel days (roughly 20 miles per day)
	travelDays := distance / 20
	if travelDays < 1 {
		travelDays = 1
	}

	return travelDays
}

func (s *EconomicSimulatorService) calculateRouteDifficulty(start, end *models.Settlement) int {
	difficulty := 3 // Base difficulty

	// Terrain affects difficulty
	terrains := []string{start.TerrainType, end.TerrainType}
	for _, terrain := range terrains {
		switch terrain {
		case "mountainous":
			difficulty += 2
		case "swamp":
			difficulty += 2
		case "desert":
			difficulty += 1
		case "forest":
			difficulty += 1
		}
	}

	// Average danger level
	avgDanger := (start.DangerLevel + end.DangerLevel) / 2
	difficulty += avgDanger / 3

	if difficulty > 10 {
		difficulty = 10
	}

	return difficulty
}

func (s *EconomicSimulatorService) determineRouteType(start, end *models.Settlement) string {
	// Determine based on terrain
	if start.TerrainType == "coastal" || end.TerrainType == "coastal" {
		return "sea"
	}

	if start.TerrainType == "mountainous" && end.TerrainType == "mountainous" {
		return "mountain pass"
	}

	// Check for special conditions
	if start.EldritchInfluence > 7 || end.EldritchInfluence > 7 {
		return "planar" // Trade through other dimensions!
	}

	return "land"
}

func (s *EconomicSimulatorService) determineTradedGoods(start, end *models.Settlement) []string {
	goods := []string{}

	// Get what each settlement exports/imports
	var startExports, startImports, endExports, endImports []string
	_ = json.Unmarshal([]byte(start.PrimaryExports), &startExports)
	_ = json.Unmarshal([]byte(start.PrimaryImports), &startImports)
	_ = json.Unmarshal([]byte(end.PrimaryExports), &endExports)
	_ = json.Unmarshal([]byte(end.PrimaryImports), &endImports)

	// Match exports to imports
	for _, export := range startExports {
		for _, import_ := range endImports {
			if export == import_ {
				goods = append(goods, export)
			}
		}
	}

	for _, export := range endExports {
		for _, import_ := range startImports {
			if export == import_ {
				goods = append(goods, export)
			}
		}
	}

	// Default goods if no matches
	if len(goods) == 0 {
		goods = []string{"general goods", "raw materials"}
	}

	return goods
}

func (s *EconomicSimulatorService) calculateBanditThreat(start, end *models.Settlement) int {
	// Based on route characteristics
	threat := 0

	// Low settlement law enforcement
	avgWealth := (start.WealthLevel + end.WealthLevel) / 2
	if avgWealth > 5 {
		threat += 2 // Wealthy routes attract bandits
	}

	// Danger levels
	avgDanger := (start.DangerLevel + end.DangerLevel) / 2
	threat += avgDanger / 2

	// Corruption enables banditry
	avgCorruption := (start.CorruptionLevel + end.CorruptionLevel) / 2
	if avgCorruption > 5 {
		threat += 2
	}

	if threat > 10 {
		threat = 10
	}

	return threat
}

func (s *EconomicSimulatorService) calculateMonsterThreat(start, end *models.Settlement) int {
	threat := 0

	// Wilderness areas
	if start.Type == models.SettlementHamlet || end.Type == models.SettlementHamlet {
		threat += 2
	}

	// Ancient influence attracts monsters
	avgEldritch := (start.EldritchInfluence + end.EldritchInfluence) / 2
	threat += avgEldritch / 2

	// Corrupted areas
	avgCorruption := (start.CorruptionLevel + end.CorruptionLevel) / 2
	threat += avgCorruption / 3

	if threat > 10 {
		threat = 10
	}

	return threat
}

func (s *EconomicSimulatorService) checkForAncientHazards(start, end *models.Settlement) bool {
	// Ancient hazards present if either settlement has strong ancient connections
	return start.AncientRuinsNearby || end.AncientRuinsNearby ||
		start.EldritchInfluence > 5 || end.EldritchInfluence > 5 ||
		start.LeyLineConnection || end.LeyLineConnection
}

func (s *EconomicSimulatorService) calculateInitialTradeVolume(start, end *models.Settlement) int {
	// Based on settlement sizes and wealth
	volume := 3 // Base volume

	// Larger settlements = more trade
	avgPop := (start.Population + end.Population) / 2
	if avgPop > 10000 {
		volume += 3
	} else if avgPop > 5000 {
		volume += 2
	} else if avgPop > 1000 {
		volume += 1
	}

	// Wealth increases trade
	avgWealth := (start.WealthLevel + end.WealthLevel) / 2
	volume += avgWealth / 3

	if volume > 10 {
		volume = 10
	}

	return volume
}

func (s *EconomicSimulatorService) determineEnvironmentalHazards(start, end *models.Settlement) []string {
	hazards := []string{}

	// Terrain-based hazards
	terrains := []string{start.TerrainType, end.TerrainType}
	for _, terrain := range terrains {
		switch terrain {
		case "mountainous":
			hazards = append(hazards, "avalanches", "altitude sickness")
		case "swamp":
			hazards = append(hazards, "disease", "quicksand")
		case "desert":
			hazards = append(hazards, "sandstorms", "dehydration")
		case "forest":
			hazards = append(hazards, "getting lost", "wild animals")
		}
	}

	// Climate hazards
	climates := []string{start.Climate, end.Climate}
	for _, climate := range climates {
		switch climate {
		case "cold":
			hazards = append(hazards, "blizzards", "frostbite")
		case constants.ClimateTropical:
			hazards = append(hazards, "monsoons", "tropical diseases")
		case "arid":
			hazards = append(hazards, "extreme heat", "water scarcity")
		}
	}

	return hazards
}

func (s *EconomicSimulatorService) updateSettlementTradeConnections(start, end *models.Settlement, routeID uuid.UUID) {
	// Update start settlement's trade routes
	var startRoutes []string
	_ = json.Unmarshal([]byte(start.TradeRoutes), &startRoutes)
	startRoutes = append(startRoutes, routeID.String())
	startRoutesJSON, _ := json.Marshal(startRoutes)
	start.TradeRoutes = models.JSONB(startRoutesJSON)
	// Would update through repository

	// Update end settlement's trade routes
	var endRoutes []string
	_ = json.Unmarshal([]byte(end.TradeRoutes), &endRoutes)
	endRoutes = append(endRoutes, routeID.String())
	endRoutesJSON, _ := json.Marshal(endRoutes)
	end.TradeRoutes = models.JSONB(endRoutesJSON)
	// Would update through repository
}
