package database

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// WorldBuildingRepository handles all world building data operations
type WorldBuildingRepository struct {
	db *DB
}

// NewWorldBuildingRepository creates a new world building repository
func NewWorldBuildingRepository(db *DB) *WorldBuildingRepository {
	return &WorldBuildingRepository{db: db}
}

// Settlement operations

// CreateSettlement creates a new settlement
func (r *WorldBuildingRepository) CreateSettlement(settlement *models.Settlement) error {
	settlement.ID = uuid.New()
	settlement.CreatedAt = time.Now()
	settlement.UpdatedAt = time.Now()

	// Marshal JSONB fields
	coordinates, _ := json.Marshal(settlement.Coordinates)
	primaryExports, _ := json.Marshal(settlement.PrimaryExports)
	primaryImports, _ := json.Marshal(settlement.PrimaryImports)
	tradeRoutes, _ := json.Marshal(settlement.TradeRoutes)
	notableLocations, _ := json.Marshal(settlement.NotableLocations)
	defenses, _ := json.Marshal(settlement.Defenses)
	problems, _ := json.Marshal(settlement.Problems)
	secrets, _ := json.Marshal(settlement.Secrets)

	query := `
		INSERT INTO settlements (
			id, game_session_id, name, type, population, age_category,
			description, history, government_type, alignment, danger_level, corruption_level,
			region, coordinates, terrain_type, climate,
			wealth_level, primary_exports, primary_imports, trade_routes,
			ancient_ruins_nearby, eldritch_influence, ley_line_connection,
			notable_locations, defenses, problems, secrets,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?
		)`

	_, err := r.db.ExecRebind(query,
		settlement.ID, settlement.GameSessionID, settlement.Name, settlement.Type,
		settlement.Population, settlement.AgeCategory, settlement.Description,
		settlement.History, settlement.GovernmentType, settlement.Alignment,
		settlement.DangerLevel, settlement.CorruptionLevel,
		settlement.Region, coordinates, settlement.TerrainType, settlement.Climate,
		settlement.WealthLevel, primaryExports, primaryImports, tradeRoutes,
		settlement.AncientRuinsNearby, settlement.EldritchInfluence, settlement.LeyLineConnection,
		notableLocations, defenses, problems, secrets,
		settlement.CreatedAt, settlement.UpdatedAt,
	)

	return err
}

// GetSettlement retrieves a settlement by ID
func (r *WorldBuildingRepository) GetSettlement(id uuid.UUID) (*models.Settlement, error) {
	var settlement models.Settlement
	var coordinates, primaryExports, primaryImports, tradeRoutes,
		notableLocations, defenses, problems, secrets []byte

	query := `
		SELECT id, game_session_id, name, type, population, age_category,
			description, history, government_type, alignment, danger_level, corruption_level,
			region, coordinates, terrain_type, climate,
			wealth_level, primary_exports, primary_imports, trade_routes,
			ancient_ruins_nearby, eldritch_influence, ley_line_connection,
			notable_locations, defenses, problems, secrets,
			created_at, updated_at
		FROM settlements WHERE id = ?`

	err := r.db.QueryRowRebind(query, id).Scan(
		&settlement.ID, &settlement.GameSessionID, &settlement.Name, &settlement.Type,
		&settlement.Population, &settlement.AgeCategory, &settlement.Description,
		&settlement.History, &settlement.GovernmentType, &settlement.Alignment,
		&settlement.DangerLevel, &settlement.CorruptionLevel,
		&settlement.Region, &coordinates, &settlement.TerrainType, &settlement.Climate,
		&settlement.WealthLevel, &primaryExports, &primaryImports, &tradeRoutes,
		&settlement.AncientRuinsNearby, &settlement.EldritchInfluence, &settlement.LeyLineConnection,
		&notableLocations, &defenses, &problems, &secrets,
		&settlement.CreatedAt, &settlement.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSONB fields
	settlement.Coordinates = models.JSONB(coordinates)
	settlement.PrimaryExports = models.JSONB(primaryExports)
	settlement.PrimaryImports = models.JSONB(primaryImports)
	settlement.TradeRoutes = models.JSONB(tradeRoutes)
	settlement.NotableLocations = models.JSONB(notableLocations)
	settlement.Defenses = models.JSONB(defenses)
	settlement.Problems = models.JSONB(problems)
	settlement.Secrets = models.JSONB(secrets)

	// Load related NPCs and shops
	settlement.NPCs, _ = r.GetSettlementNPCs(id)
	settlement.Shops, _ = r.GetSettlementShops(id)

	return &settlement, nil
}

// GetSettlementsByGameSession retrieves all settlements for a game session
func (r *WorldBuildingRepository) GetSettlementsByGameSession(gameSessionID uuid.UUID) ([]*models.Settlement, error) {
	query := `
		SELECT id, name, type, population, region, danger_level, corruption_level
		FROM settlements 
		WHERE game_session_id = ?
		ORDER BY population DESC`

	rows, err := r.db.Query(query, gameSessionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	settlements := make([]*models.Settlement, 0, 10)
	for rows.Next() {
		var s models.Settlement
		err := rows.Scan(
			&s.ID, &s.Name, &s.Type, &s.Population,
			&s.Region, &s.DangerLevel, &s.CorruptionLevel,
		)
		if err != nil {
			continue
		}
		settlements = append(settlements, &s)
	}

	return settlements, nil
}

// UpdateSettlement updates a settlement's data
func (r *WorldBuildingRepository) UpdateSettlement(settlement *models.Settlement) error {
	settlement.UpdatedAt = time.Now()

	// Marshal JSONB fields
	coordinates, _ := json.Marshal(settlement.Coordinates)
	primaryExports, _ := json.Marshal(settlement.PrimaryExports)
	primaryImports, _ := json.Marshal(settlement.PrimaryImports)
	tradeRoutes, _ := json.Marshal(settlement.TradeRoutes)
	notableLocations, _ := json.Marshal(settlement.NotableLocations)
	defenses, _ := json.Marshal(settlement.Defenses)
	problems, _ := json.Marshal(settlement.Problems)
	secrets, _ := json.Marshal(settlement.Secrets)

	query := `
		UPDATE settlements SET
			name = ?, type = ?, population = ?, age_category = ?,
			description = ?, history = ?, government_type = ?, alignment = ?,
			danger_level = ?, corruption_level = ?, region = ?, coordinates = ?,
			terrain_type = ?, climate = ?, wealth_level = ?,
			primary_exports = ?, primary_imports = ?, trade_routes = ?,
			ancient_ruins_nearby = ?, eldritch_influence = ?, ley_line_connection = ?,
			notable_locations = ?, defenses = ?, problems = ?, secrets = ?,
			updated_at = ?
		WHERE id = ?`

	_, err := r.db.ExecRebind(query,
		settlement.Name, settlement.Type, settlement.Population, settlement.AgeCategory,
		settlement.Description, settlement.History, settlement.GovernmentType, settlement.Alignment,
		settlement.DangerLevel, settlement.CorruptionLevel, settlement.Region, coordinates,
		settlement.TerrainType, settlement.Climate, settlement.WealthLevel,
		primaryExports, primaryImports, tradeRoutes,
		settlement.AncientRuinsNearby, settlement.EldritchInfluence, settlement.LeyLineConnection,
		notableLocations, defenses, problems, secrets,
		settlement.UpdatedAt, settlement.ID,
	)

	return err
}

// UpdateSettlementProsperity updates only the wealth level of a settlement
func (r *WorldBuildingRepository) UpdateSettlementProsperity(settlementID uuid.UUID, wealthLevel int) error {
	query := `
		UPDATE settlements 
		SET wealth_level = ?, updated_at = ?
		WHERE id = ?`

	_, err := r.db.ExecRebind(query, wealthLevel, time.Now(), settlementID)
	return err
}

// NPC operations

// CreateSettlementNPC creates a new NPC in a settlement
func (r *WorldBuildingRepository) CreateSettlementNPC(npc *models.SettlementNPC) error {
	npc.ID = uuid.New()
	npc.CreatedAt = time.Now()

	// Marshal JSONB fields
	personalityTraits, _ := json.Marshal(npc.PersonalityTraits)
	ideals, _ := json.Marshal(npc.Ideals)
	bonds, _ := json.Marshal(npc.Bonds)
	flaws, _ := json.Marshal(npc.Flaws)
	factionAffiliations, _ := json.Marshal(npc.FactionAffiliations)
	relationships, _ := json.Marshal(npc.Relationships)
	stats, _ := json.Marshal(npc.Stats)
	skills, _ := json.Marshal(npc.Skills)
	inventory, _ := json.Marshal(npc.Inventory)
	plotHooks, _ := json.Marshal(npc.PlotHooks)

	query := `
		INSERT INTO settlement_npcs (
			id, settlement_id, name, race, class, level, role, occupation,
			personality_traits, ideals, bonds, flaws,
			ancient_knowledge, corruption_touched, secret_agenda, true_age,
			faction_affiliations, relationships, stats, skills, inventory, plot_hooks,
			created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`

	_, err := r.db.ExecRebind(query,
		npc.ID, npc.SettlementID, npc.Name, npc.Race, npc.Class, npc.Level,
		npc.Role, npc.Occupation, personalityTraits, ideals, bonds, flaws,
		npc.AncientKnowledge, npc.CorruptionTouched, npc.SecretAgenda, npc.TrueAge,
		factionAffiliations, relationships, stats, skills, inventory, plotHooks,
		npc.CreatedAt,
	)

	return err
}

// GetSettlementNPCs retrieves all NPCs in a settlement
func (r *WorldBuildingRepository) GetSettlementNPCs(settlementID uuid.UUID) ([]models.SettlementNPC, error) {
	query := `
		SELECT id, name, race, class, level, role, occupation,
			ancient_knowledge, corruption_touched
		FROM settlement_npcs
		WHERE settlement_id = ?`

	rows, err := r.db.Query(query, settlementID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	npcs := make([]models.SettlementNPC, 0, 10)
	for rows.Next() {
		var npc models.SettlementNPC
		err := rows.Scan(
			&npc.ID, &npc.Name, &npc.Race, &npc.Class, &npc.Level,
			&npc.Role, &npc.Occupation, &npc.AncientKnowledge, &npc.CorruptionTouched,
		)
		if err != nil {
			continue
		}
		npc.SettlementID = settlementID
		npcs = append(npcs, npc)
	}

	return npcs, nil
}

// Shop operations

// CreateSettlementShop creates a new shop in a settlement
func (r *WorldBuildingRepository) CreateSettlementShop(shop *models.SettlementShop) error {
	shop.ID = uuid.New()
	shop.CreatedAt = time.Now()

	// Marshal JSONB fields
	availableItems, _ := json.Marshal(shop.AvailableItems)
	specialItems, _ := json.Marshal(shop.SpecialItems)
	craftingSpecialties, _ := json.Marshal(shop.CraftingSpecialties)
	factionDiscount, _ := json.Marshal(shop.FactionDiscount)
	operatingHours, _ := json.Marshal(shop.OperatingHours)
	currentRumors, _ := json.Marshal(shop.CurrentRumors)

	query := `
		INSERT INTO settlement_shops (
			id, settlement_id, name, type, owner_npc_id, quality_level, price_modifier,
			available_items, special_items, can_craft, crafting_specialties,
			black_market, ancient_artifacts, faction_discount,
			reputation_required, operating_hours, current_rumors, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`

	_, err := r.db.ExecRebind(query,
		shop.ID, shop.SettlementID, shop.Name, shop.Type, shop.OwnerNPCID,
		shop.QualityLevel, shop.PriceModifier, availableItems, specialItems,
		shop.CanCraft, craftingSpecialties, shop.BlackMarket, shop.AncientArtifacts,
		factionDiscount, shop.ReputationRequired, operatingHours, currentRumors,
		shop.CreatedAt,
	)

	return err
}

// GetSettlementShops retrieves all shops in a settlement
func (r *WorldBuildingRepository) GetSettlementShops(settlementID uuid.UUID) ([]models.SettlementShop, error) {
	query := `
		SELECT id, name, type, owner_npc_id, quality_level, price_modifier,
			black_market, ancient_artifacts
		FROM settlement_shops
		WHERE settlement_id = ?`

	rows, err := r.db.Query(query, settlementID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	shops := make([]models.SettlementShop, 0, 10)
	for rows.Next() {
		var shop models.SettlementShop
		err := rows.Scan(
			&shop.ID, &shop.Name, &shop.Type, &shop.OwnerNPCID,
			&shop.QualityLevel, &shop.PriceModifier,
			&shop.BlackMarket, &shop.AncientArtifacts,
		)
		if err != nil {
			continue
		}
		shop.SettlementID = settlementID
		shops = append(shops, shop)
	}

	return shops, nil
}

// Faction operations

// CreateFaction creates a new faction
func (r *WorldBuildingRepository) CreateFaction(faction *models.Faction) error {
	faction.ID = uuid.New()
	faction.CreatedAt = time.Now()
	faction.UpdatedAt = time.Now()

	// Marshal JSONB fields
	publicGoals, _ := json.Marshal(faction.PublicGoals)
	secretGoals, _ := json.Marshal(faction.SecretGoals)
	motivations, _ := json.Marshal(faction.Motivations)
	territoryControl, _ := json.Marshal(faction.TerritoryControl)
	factionRelationships, _ := json.Marshal(faction.FactionRelationships)
	symbols, _ := json.Marshal(faction.Symbols)
	rituals, _ := json.Marshal(faction.Rituals)
	resources, _ := json.Marshal(faction.Resources)

	query := `
		INSERT INTO factions (
			id, game_session_id, name, type, description, founding_date,
			public_goals, secret_goals, motivations,
			ancient_knowledge_level, seeks_ancient_power, guards_ancient_secrets, corrupted,
			influence_level, military_strength, economic_power, magical_resources,
			leadership_structure, headquarters_location, member_count, territory_control,
			faction_relationships, symbols, rituals, resources,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`

	_, err := r.db.ExecRebind(query,
		faction.ID, faction.GameSessionID, faction.Name, faction.Type,
		faction.Description, faction.FoundingDate,
		publicGoals, secretGoals, motivations,
		faction.AncientKnowledgeLevel, faction.SeeksAncientPower,
		faction.GuardsAncientSecrets, faction.Corrupted,
		faction.InfluenceLevel, faction.MilitaryStrength, faction.EconomicPower,
		faction.MagicalResources, faction.LeadershipStructure, faction.HeadquartersLocation,
		faction.MemberCount, territoryControl, factionRelationships,
		symbols, rituals, resources,
		faction.CreatedAt, faction.UpdatedAt,
	)

	return err
}

// GetFaction retrieves a faction by ID
func (r *WorldBuildingRepository) GetFaction(id uuid.UUID) (*models.Faction, error) {
	var faction models.Faction
	var publicGoals, secretGoals, motivations, territoryControl,
		factionRelationships, symbols, rituals, resources []byte

	query := `
		SELECT id, game_session_id, name, type, description, founding_date,
			public_goals, secret_goals, motivations,
			ancient_knowledge_level, seeks_ancient_power, guards_ancient_secrets, corrupted,
			influence_level, military_strength, economic_power, magical_resources,
			leadership_structure, headquarters_location, member_count, territory_control,
			faction_relationships, symbols, rituals, resources,
			created_at, updated_at
		FROM factions WHERE id = ?`

	err := r.db.QueryRowRebind(query, id).Scan(
		&faction.ID, &faction.GameSessionID, &faction.Name, &faction.Type,
		&faction.Description, &faction.FoundingDate,
		&publicGoals, &secretGoals, &motivations,
		&faction.AncientKnowledgeLevel, &faction.SeeksAncientPower,
		&faction.GuardsAncientSecrets, &faction.Corrupted,
		&faction.InfluenceLevel, &faction.MilitaryStrength, &faction.EconomicPower,
		&faction.MagicalResources, &faction.LeadershipStructure, &faction.HeadquartersLocation,
		&faction.MemberCount, &territoryControl, &factionRelationships,
		&symbols, &rituals, &resources,
		&faction.CreatedAt, &faction.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSONB fields
	faction.PublicGoals = models.JSONB(publicGoals)
	faction.SecretGoals = models.JSONB(secretGoals)
	faction.Motivations = models.JSONB(motivations)
	faction.TerritoryControl = models.JSONB(territoryControl)
	faction.FactionRelationships = models.JSONB(factionRelationships)
	faction.Symbols = models.JSONB(symbols)
	faction.Rituals = models.JSONB(rituals)
	faction.Resources = models.JSONB(resources)

	return &faction, nil
}

// GetFactionsByGameSession retrieves all factions for a game session
func (r *WorldBuildingRepository) GetFactionsByGameSession(gameSessionID uuid.UUID) ([]*models.Faction, error) {
	query := `
		SELECT id, name, type, influence_level, corrupted
		FROM factions 
		WHERE game_session_id = ?
		ORDER BY influence_level DESC`

	rows, err := r.db.Query(query, gameSessionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	factions := make([]*models.Faction, 0, 10)
	for rows.Next() {
		var f models.Faction
		err := rows.Scan(&f.ID, &f.Name, &f.Type, &f.InfluenceLevel, &f.Corrupted)
		if err != nil {
			continue
		}
		f.GameSessionID = gameSessionID
		factions = append(factions, &f)
	}

	return factions, nil
}

// UpdateFactionRelationship updates the relationship between two factions
func (r *WorldBuildingRepository) UpdateFactionRelationship(faction1ID, faction2ID uuid.UUID, standing int, relationType string) error {
	// Get faction 1
	faction1, err := r.GetFaction(faction1ID)
	if err != nil {
		return err
	}

	// Update relationships
	var relationships map[string]interface{}
	_ = json.Unmarshal([]byte(faction1.FactionRelationships), &relationships)
	if relationships == nil {
		relationships = make(map[string]interface{})
	}

	relationships[faction2ID.String()] = map[string]interface{}{
		"standing": standing,
		"type":     relationType,
	}

	updatedRelationships, _ := json.Marshal(relationships)

	// Update in database
	query := `UPDATE factions SET faction_relationships = ?, updated_at = ? WHERE id = ?`
	_, err = r.db.ExecRebind(query, updatedRelationships, time.Now(), faction1ID)

	// Also update faction 2's relationship with faction 1
	if err == nil {
		faction2, err := r.GetFaction(faction2ID)
		if err == nil {
			var relationships2 map[string]interface{}
			_ = json.Unmarshal([]byte(faction2.FactionRelationships), &relationships2)
			if relationships2 == nil {
				relationships2 = make(map[string]interface{})
			}

			relationships2[faction1ID.String()] = map[string]interface{}{
				"standing": standing,
				"type":     relationType,
			}

			updatedRelationships2, _ := json.Marshal(relationships2)
			_, _ = r.db.ExecRebind(query, updatedRelationships2, time.Now(), faction2ID)
		}
	}

	return err
}

// World Event operations

// CreateWorldEvent creates a new world event
func (r *WorldBuildingRepository) CreateWorldEvent(event *models.WorldEvent) error {
	event.ID = uuid.New()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	// Marshal JSONB fields
	affectedRegions, _ := json.Marshal(event.AffectedRegions)
	affectedSettlements, _ := json.Marshal(event.AffectedSettlements)
	affectedFactions, _ := json.Marshal(event.AffectedFactions)
	economicImpacts, _ := json.Marshal(event.EconomicImpacts)
	politicalImpacts, _ := json.Marshal(event.PoliticalImpacts)
	stages, _ := json.Marshal(event.Stages)
	resolutionConditions, _ := json.Marshal(event.ResolutionConditions)
	consequences, _ := json.Marshal(event.Consequences)
	partyActions, _ := json.Marshal(event.PartyActions)

	query := `
		INSERT INTO world_events (
			id, game_session_id, name, type, severity, description, cause,
			start_date, duration, is_active, is_resolved,
			ancient_cause, awakens_ancient_evil, prophecy_related,
			affected_regions, affected_settlements, affected_factions,
			economic_impacts, political_impacts,
			current_stage, stages, resolution_conditions, consequences,
			party_aware, party_involved, party_actions,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`

	_, err := r.db.ExecRebind(query,
		event.ID, event.GameSessionID, event.Name, event.Type, event.Severity,
		event.Description, event.Cause, event.StartDate, event.Duration,
		event.IsActive, event.IsResolved, event.AncientCause, event.AwakensAncientEvil,
		event.ProphecyRelated, affectedRegions, affectedSettlements, affectedFactions,
		economicImpacts, politicalImpacts, event.CurrentStage, stages,
		resolutionConditions, consequences, event.PartyAware, event.PartyInvolved,
		partyActions, event.CreatedAt, event.UpdatedAt,
	)

	return err
}

// GetActiveWorldEvents retrieves all active world events for a game session
func (r *WorldBuildingRepository) GetActiveWorldEvents(gameSessionID uuid.UUID) ([]*models.WorldEvent, error) {
	query := `
		SELECT id, name, type, severity, description, current_stage,
			party_aware, party_involved
		FROM world_events 
		WHERE game_session_id = ? AND is_active = true
		ORDER BY severity DESC, created_at DESC`

	rows, err := r.db.Query(query, gameSessionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	events := make([]*models.WorldEvent, 0, 10)
	for rows.Next() {
		var e models.WorldEvent
		err := rows.Scan(
			&e.ID, &e.Name, &e.Type, &e.Severity, &e.Description,
			&e.CurrentStage, &e.PartyAware, &e.PartyInvolved,
		)
		if err != nil {
			continue
		}
		e.GameSessionID = gameSessionID
		e.IsActive = true
		events = append(events, &e)
	}

	return events, nil
}

// ProgressWorldEvent advances a world event to the next stage
func (r *WorldBuildingRepository) ProgressWorldEvent(eventID uuid.UUID) error {
	query := `
		UPDATE world_events 
		SET current_stage = current_stage + 1, updated_at = ?
		WHERE id = ?`

	_, err := r.db.ExecRebind(query, time.Now(), eventID)
	return err
}

// Market operations

// CreateOrUpdateMarket creates or updates market conditions for a settlement
func (r *WorldBuildingRepository) CreateOrUpdateMarket(market *models.Market) error {
	if market.ID == uuid.Nil {
		market.ID = uuid.New()
	}
	market.LastUpdated = time.Now()

	// Marshal JSONB fields
	highDemandItems, _ := json.Marshal(market.HighDemandItems)
	surplusItems, _ := json.Marshal(market.SurplusItems)
	bannedItems, _ := json.Marshal(market.BannedItems)

	query := `
		INSERT INTO markets (
			id, settlement_id, food_price_modifier, common_goods_modifier,
			weapons_armor_modifier, magical_items_modifier, ancient_artifacts_modifier,
			high_demand_items, surplus_items, banned_items,
			black_market_active, artifact_dealer_present,
			economic_boom, economic_depression, last_updated
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		) ON CONFLICT (settlement_id) DO UPDATE SET
			food_price_modifier = ?, common_goods_modifier = ?,
			weapons_armor_modifier = ?, magical_items_modifier = ?,
			ancient_artifacts_modifier = ?, high_demand_items = ?,
			surplus_items = ?, banned_items = ?, black_market_active = ?,
			artifact_dealer_present = ?, economic_boom = ?,
			economic_depression = ?, last_updated = ?`

	_, err := r.db.ExecRebind(query,
		market.ID, market.SettlementID,
		market.FoodPriceModifier, market.CommonGoodsModifier,
		market.WeaponsArmorModifier, market.MagicalItemsModifier,
		market.AncientArtifactsModifier, highDemandItems, surplusItems,
		bannedItems, market.BlackMarketActive, market.ArtifactDealerPresent,
		market.EconomicBoom, market.EconomicDepression, market.LastUpdated,
	)

	return err
}

// GetMarketBySettlement retrieves market conditions for a settlement
func (r *WorldBuildingRepository) GetMarketBySettlement(settlementID uuid.UUID) (*models.Market, error) {
	var market models.Market
	var highDemandItems, surplusItems, bannedItems []byte

	query := `
		SELECT id, settlement_id, food_price_modifier, common_goods_modifier,
			weapons_armor_modifier, magical_items_modifier, ancient_artifacts_modifier,
			high_demand_items, surplus_items, banned_items,
			black_market_active, artifact_dealer_present,
			economic_boom, economic_depression, last_updated
		FROM markets WHERE settlement_id = ?`

	err := r.db.QueryRowRebind(query, settlementID).Scan(
		&market.ID, &market.SettlementID,
		&market.FoodPriceModifier, &market.CommonGoodsModifier,
		&market.WeaponsArmorModifier, &market.MagicalItemsModifier,
		&market.AncientArtifactsModifier, &highDemandItems, &surplusItems,
		&bannedItems, &market.BlackMarketActive, &market.ArtifactDealerPresent,
		&market.EconomicBoom, &market.EconomicDepression, &market.LastUpdated,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSONB fields
	market.HighDemandItems = models.JSONB(highDemandItems)
	market.SurplusItems = models.JSONB(surplusItems)
	market.BannedItems = models.JSONB(bannedItems)

	return &market, nil
}

// Trade Route operations

// CreateTradeRoute creates a new trade route
func (r *WorldBuildingRepository) CreateTradeRoute(route *models.TradeRoute) error {
	route.ID = uuid.New()
	route.CreatedAt = time.Now()
	route.UpdatedAt = time.Now()

	// Marshal JSONB fields
	environmentalHazards, _ := json.Marshal(route.EnvironmentalHazards)
	primaryGoods, _ := json.Marshal(route.PrimaryGoods)
	disruptionEvents, _ := json.Marshal(route.DisruptionEvents)

	query := `
		INSERT INTO trade_routes (
			id, game_session_id, name, start_settlement_id, end_settlement_id,
			route_type, distance, difficulty_rating,
			bandit_threat_level, monster_threat_level, ancient_hazards, environmental_hazards,
			trade_volume, primary_goods, tariff_rate,
			controlling_faction_id, protection_fee,
			is_active, disruption_events,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?
		)`

	_, err := r.db.ExecRebind(query,
		route.ID, route.GameSessionID, route.Name,
		route.StartSettlementID, route.EndSettlementID,
		route.RouteType, route.Distance, route.DifficultyRating,
		route.BanditThreatLevel, route.MonsterThreatLevel,
		route.AncientHazards, environmentalHazards,
		route.TradeVolume, primaryGoods, route.TariffRate,
		route.ControllingFactionID, route.ProtectionFee,
		route.IsActive, disruptionEvents,
		route.CreatedAt, route.UpdatedAt,
	)

	return err
}

// GetTradeRoutesBySettlement retrieves all trade routes connected to a settlement
func (r *WorldBuildingRepository) GetTradeRoutesBySettlement(settlementID uuid.UUID) ([]*models.TradeRoute, error) {
	query := `
		SELECT id, name, start_settlement_id, end_settlement_id,
			route_type, distance, difficulty_rating, is_active
		FROM trade_routes 
		WHERE (start_settlement_id = ? OR end_settlement_id = ?) AND is_active = true`

	rows, err := r.db.Query(query, settlementID, settlementID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	routes := make([]*models.TradeRoute, 0, 10)
	for rows.Next() {
		var route models.TradeRoute
		err := rows.Scan(
			&route.ID, &route.Name, &route.StartSettlementID, &route.EndSettlementID,
			&route.RouteType, &route.Distance, &route.DifficultyRating, &route.IsActive,
		)
		if err != nil {
			continue
		}
		routes = append(routes, &route)
	}

	return routes, nil
}

// Ancient Site operations

// CreateAncientSite creates a new ancient site
func (r *WorldBuildingRepository) CreateAncientSite(site *models.AncientSite) error {
	site.ID = uuid.New()
	site.CreatedAt = time.Now()
	site.UpdatedAt = time.Now()

	// Marshal JSONB fields
	coordinates, _ := json.Marshal(site.Coordinates)
	treasures, _ := json.Marshal(site.Treasures)
	artifacts, _ := json.Marshal(site.Artifacts)
	forbiddenKnowledge, _ := json.Marshal(site.ForbiddenKnowledge)
	planarConnections, _ := json.Marshal(site.PlanarConnections)
	prophecies, _ := json.Marshal(site.Prophecies)

	query := `
		INSERT INTO ancient_sites (
			id, game_session_id, name, true_name, type, age_category,
			location_description, nearest_settlement_id, coordinates,
			exploration_level, corruption_level, structural_integrity,
			guardian_type, guardian_defeated, seals_intact,
			treasures, artifacts, forbidden_knowledge,
			ley_line_nexus, reality_weakness, planar_connections,
			original_purpose, fall_description, prophecies,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`

	_, err := r.db.ExecRebind(query,
		site.ID, site.GameSessionID, site.Name, site.TrueName, site.Type,
		site.AgeCategory, site.LocationDescription, site.NearestSettlementID,
		coordinates, site.ExplorationLevel, site.CorruptionLevel,
		site.StructuralIntegrity, site.GuardianType, site.GuardianDefeated,
		site.SealsIntact, treasures, artifacts, forbiddenKnowledge,
		site.LeyLineNexus, site.RealityWeakness, planarConnections,
		site.OriginalPurpose, site.FallDescription, prophecies,
		site.CreatedAt, site.UpdatedAt,
	)

	return err
}

// GetAncientSitesByGameSession retrieves all ancient sites for a game session
func (r *WorldBuildingRepository) GetAncientSitesByGameSession(gameSessionID uuid.UUID) ([]*models.AncientSite, error) {
	query := `
		SELECT id, name, type, age_category, corruption_level, exploration_level
		FROM ancient_sites 
		WHERE game_session_id = ?
		ORDER BY corruption_level DESC`

	rows, err := r.db.Query(query, gameSessionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	sites := make([]*models.AncientSite, 0, 10)
	for rows.Next() {
		var site models.AncientSite
		err := rows.Scan(
			&site.ID, &site.Name, &site.Type, &site.AgeCategory,
			&site.CorruptionLevel, &site.ExplorationLevel,
		)
		if err != nil {
			continue
		}
		site.GameSessionID = gameSessionID
		sites = append(sites, &site)
	}

	return sites, nil
}

// SimulateEconomicChanges updates market prices based on world events and trade disruptions
func (r *WorldBuildingRepository) SimulateEconomicChanges(gameSessionID uuid.UUID) error {
	// This would be called periodically to update market conditions
	// based on active world events, trade route disruptions, etc.

	events, err := r.getActiveEconomicEvents(gameSessionID)
	if err != nil {
		return err
	}

	for _, event := range events {
		r.applyEventEconomicImpacts(event)
	}

	return nil
}

// economicEvent represents an event with economic impacts
type economicEvent struct {
	ID      uuid.UUID
	Impacts map[string]interface{}
}

// getActiveEconomicEvents retrieves all active events with economic impacts
func (r *WorldBuildingRepository) getActiveEconomicEvents(gameSessionID uuid.UUID) ([]*economicEvent, error) {
	query := `
		SELECT id, economic_impacts 
		FROM world_events 
		WHERE game_session_id = ? AND is_active = true`

	rows, err := r.db.Query(query, gameSessionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []*economicEvent
	for rows.Next() {
		event, err := r.scanEconomicEvent(rows)
		if err != nil {
			continue // Skip invalid events
		}
		events = append(events, event)
	}

	return events, nil
}

// scanEconomicEvent scans a single economic event from the database
func (r *WorldBuildingRepository) scanEconomicEvent(row interface{ Scan(...interface{}) error }) (*economicEvent, error) {
	var eventID uuid.UUID
	var economicImpacts []byte

	if err := row.Scan(&eventID, &economicImpacts); err != nil {
		return nil, err
	}

	var impacts map[string]interface{}
	if err := json.Unmarshal(economicImpacts, &impacts); err != nil {
		return nil, err
	}

	return &economicEvent{
		ID:      eventID,
		Impacts: impacts,
	}, nil
}

// applyEventEconomicImpacts applies economic impacts from an event to affected settlements
func (r *WorldBuildingRepository) applyEventEconomicImpacts(event *economicEvent) {
	settlementIDs := r.extractAffectedSettlements(event.Impacts)
	modifier := r.extractPriceModifier(event.Impacts)
	
	if modifier == 0 {
		return // No price changes to apply
	}

	for _, settlementID := range settlementIDs {
		r.updateSettlementMarket(settlementID, modifier)
	}
}

// extractAffectedSettlements extracts settlement IDs from event impacts
func (r *WorldBuildingRepository) extractAffectedSettlements(impacts map[string]interface{}) []uuid.UUID {
	settlementIDs, ok := impacts["affected_settlements"].([]interface{})
	if !ok {
		return nil
	}

	var validIDs []uuid.UUID
	for _, id := range settlementIDs {
		if sidStr, ok := id.(string); ok {
			if sid, err := uuid.Parse(sidStr); err == nil {
				validIDs = append(validIDs, sid)
			}
		}
	}

	return validIDs
}

// extractPriceModifier extracts the price modifier from event impacts
func (r *WorldBuildingRepository) extractPriceModifier(impacts map[string]interface{}) float64 {
	if modifier, ok := impacts["price_modifier"].(float64); ok {
		return modifier
	}
	return 0
}

// updateSettlementMarket updates a settlement's market with the given price modifier
func (r *WorldBuildingRepository) updateSettlementMarket(settlementID uuid.UUID, modifier float64) {
	market, err := r.GetMarketBySettlement(settlementID)
	if err != nil || market == nil {
		return // Skip if market not found
	}

	// Apply economic modifiers
	market.CommonGoodsModifier *= modifier
	market.FoodPriceModifier *= modifier
	
	_ = r.CreateOrUpdateMarket(market)
}
