package models

import (
	"time"

	"github.com/google/uuid"
)

// Settlement types
type SettlementType string

const (
	SettlementHamlet     SettlementType = "hamlet"
	SettlementVillage    SettlementType = "village"
	SettlementTown       SettlementType = "town"
	SettlementCity       SettlementType = "city"
	SettlementMetropolis SettlementType = "metropolis"
	SettlementRuins      SettlementType = "ruins"
)

// Settlement represents a populated location in the world
type Settlement struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	GameSessionID  uuid.UUID      `json:"gameSessionId" db:"game_session_id"`
	Name           string         `json:"name" db:"name"`
	Type           SettlementType `json:"type" db:"type"`
	Population     int            `json:"population" db:"population"`
	AgeCategory    string         `json:"ageCategory" db:"age_category"`
	Description    string         `json:"description" db:"description"`
	History        string         `json:"history" db:"history"`
	GovernmentType string         `json:"governmentType" db:"government_type"`
	Alignment      string         `json:"alignment" db:"alignment"`
	DangerLevel    int            `json:"dangerLevel" db:"danger_level"`
	CorruptionLevel int           `json:"corruptionLevel" db:"corruption_level"`

	// Location
	Region       string    `json:"region" db:"region"`
	Coordinates  JSONB     `json:"coordinates" db:"coordinates"`
	TerrainType  string    `json:"terrainType" db:"terrain_type"`
	Climate      string    `json:"climate" db:"climate"`

	// Economic data
	WealthLevel     int       `json:"wealthLevel" db:"wealth_level"`
	PrimaryExports  JSONB     `json:"primaryExports" db:"primary_exports"`
	PrimaryImports  JSONB     `json:"primaryImports" db:"primary_imports"`
	TradeRoutes     JSONB     `json:"tradeRoutes" db:"trade_routes"`

	// Ancient connections
	AncientRuinsNearby  bool `json:"ancientRuinsNearby" db:"ancient_ruins_nearby"`
	EldritchInfluence   int  `json:"eldritchInfluence" db:"eldritch_influence"`
	LeyLineConnection   bool `json:"leyLineConnection" db:"ley_line_connection"`

	// Features
	NotableLocations JSONB     `json:"notableLocations" db:"notable_locations"`
	Defenses         JSONB     `json:"defenses" db:"defenses"`
	Problems         JSONB     `json:"problems" db:"problems"`
	Secrets          JSONB     `json:"secrets" db:"secrets"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`

	// Related data (not stored in main table)
	NPCs  []SettlementNPC  `json:"npcs,omitempty"`
	Shops []SettlementShop `json:"shops,omitempty"`
}

// SettlementNPC represents a non-player character in a settlement
type SettlementNPC struct {
	ID               uuid.UUID `json:"id" db:"id"`
	SettlementID     uuid.UUID `json:"settlementId" db:"settlement_id"`
	Name             string    `json:"name" db:"name"`
	Race             string    `json:"race" db:"race"`
	Class            string    `json:"class" db:"class"`
	Level            int       `json:"level" db:"level"`
	Role             string    `json:"role" db:"role"`
	Occupation       string    `json:"occupation" db:"occupation"`
	PersonalityTraits JSONB    `json:"personalityTraits" db:"personality_traits"`
	Ideals           JSONB     `json:"ideals" db:"ideals"`
	Bonds            JSONB     `json:"bonds" db:"bonds"`
	Flaws            JSONB     `json:"flaws" db:"flaws"`

	// Ancient connections
	AncientKnowledge   bool   `json:"ancientKnowledge" db:"ancient_knowledge"`
	CorruptionTouched  bool   `json:"corruptionTouched" db:"corruption_touched"`
	SecretAgenda       string `json:"secretAgenda" db:"secret_agenda"`
	TrueAge            *int   `json:"trueAge" db:"true_age"`

	// Relationships
	FactionAffiliations JSONB `json:"factionAffiliations" db:"faction_affiliations"`
	Relationships       JSONB `json:"relationships" db:"relationships"`

	// Mechanical data
	Stats      JSONB     `json:"stats" db:"stats"`
	Skills     JSONB     `json:"skills" db:"skills"`
	Inventory  JSONB     `json:"inventory" db:"inventory"`
	PlotHooks  JSONB     `json:"plotHooks" db:"plot_hooks"`

	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

// SettlementShop represents a shop or service in a settlement
type SettlementShop struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	SettlementID   uuid.UUID  `json:"settlementId" db:"settlement_id"`
	Name           string     `json:"name" db:"name"`
	Type           string     `json:"type" db:"type"`
	OwnerNPCID     *uuid.UUID `json:"ownerNpcId" db:"owner_npc_id"`
	QualityLevel   int        `json:"qualityLevel" db:"quality_level"`
	PriceModifier  float64    `json:"priceModifier" db:"price_modifier"`

	// Inventory
	AvailableItems      JSONB `json:"availableItems" db:"available_items"`
	SpecialItems        JSONB `json:"specialItems" db:"special_items"`
	CanCraft            bool  `json:"canCraft" db:"can_craft"`
	CraftingSpecialties JSONB `json:"craftingSpecialties" db:"crafting_specialties"`

	// Special features
	BlackMarket        bool  `json:"blackMarket" db:"black_market"`
	AncientArtifacts   bool  `json:"ancientArtifacts" db:"ancient_artifacts"`
	FactionDiscount    JSONB `json:"factionDiscount" db:"faction_discount"`

	ReputationRequired int       `json:"reputationRequired" db:"reputation_required"`
	OperatingHours     JSONB     `json:"operatingHours" db:"operating_hours"`
	CurrentRumors      JSONB     `json:"currentRumors" db:"current_rumors"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	// Related data
	Owner *SettlementNPC `json:"owner,omitempty"`
}

// FactionType represents the type of faction
type FactionType string

const (
	FactionReligious    FactionType = "religious"
	FactionPolitical    FactionType = "political"
	FactionCriminal     FactionType = "criminal"
	FactionMerchant     FactionType = "merchant"
	FactionMilitary     FactionType = "military"
	FactionCult         FactionType = "cult"
	FactionAncientOrder FactionType = "ancient_order"
)

// Faction represents an organization with goals and influence
type Faction struct {
	ID            uuid.UUID   `json:"id" db:"id"`
	GameSessionID uuid.UUID   `json:"gameSessionId" db:"game_session_id"`
	Name          string      `json:"name" db:"name"`
	Type          FactionType `json:"type" db:"type"`
	Description   string      `json:"description" db:"description"`
	FoundingDate  string      `json:"foundingDate" db:"founding_date"`

	// Goals
	PublicGoals  JSONB `json:"publicGoals" db:"public_goals"`
	SecretGoals  JSONB `json:"secretGoals" db:"secret_goals"`
	Motivations  JSONB `json:"motivations" db:"motivations"`

	// Ancient connections
	AncientKnowledgeLevel int  `json:"ancientKnowledgeLevel" db:"ancient_knowledge_level"`
	SeeksAncientPower     bool `json:"seeksAncientPower" db:"seeks_ancient_power"`
	GuardsAncientSecrets  bool `json:"guardsAncientSecrets" db:"guards_ancient_secrets"`
	Corrupted             bool `json:"corrupted" db:"corrupted"`

	// Power levels
	InfluenceLevel    int `json:"influenceLevel" db:"influence_level"`
	MilitaryStrength  int `json:"militaryStrength" db:"military_strength"`
	EconomicPower     int `json:"economicPower" db:"economic_power"`
	MagicalResources  int `json:"magicalResources" db:"magical_resources"`

	// Organization
	LeadershipStructure  string `json:"leadershipStructure" db:"leadership_structure"`
	HeadquartersLocation string `json:"headquartersLocation" db:"headquarters_location"`
	MemberCount          int    `json:"memberCount" db:"member_count"`
	TerritoryControl     JSONB  `json:"territoryControl" db:"territory_control"`

	// Relationships
	FactionRelationships JSONB `json:"factionRelationships" db:"faction_relationships"`

	Symbols   JSONB `json:"symbols" db:"symbols"`
	Rituals   JSONB `json:"rituals" db:"rituals"`
	Resources JSONB `json:"resources" db:"resources"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// WorldEventType represents the type of world event
type WorldEventType string

const (
	EventPolitical         WorldEventType = "political"
	EventEconomic          WorldEventType = "economic"
	EventNatural           WorldEventType = "natural"
	EventSupernatural      WorldEventType = "supernatural"
	EventAncientAwakening  WorldEventType = "ancient_awakening"
	EventPlanar            WorldEventType = "planar"
)

// WorldEventSeverity represents how severe an event is
type WorldEventSeverity string

const (
	SeverityMinor        WorldEventSeverity = "minor"
	SeverityModerate     WorldEventSeverity = "moderate"
	SeverityMajor        WorldEventSeverity = "major"
	SeverityCatastrophic WorldEventSeverity = "catastrophic"
)

// WorldEvent represents a significant event happening in the world
type WorldEvent struct {
	ID            uuid.UUID          `json:"id" db:"id"`
	GameSessionID uuid.UUID          `json:"gameSessionId" db:"game_session_id"`
	Name          string             `json:"name" db:"name"`
	Type          WorldEventType     `json:"type" db:"type"`
	Severity      WorldEventSeverity `json:"severity" db:"severity"`

	Description string `json:"description" db:"description"`
	Cause       string `json:"cause" db:"cause"`

	// Timing
	StartDate  string `json:"startDate" db:"start_date"`
	Duration   string `json:"duration" db:"duration"`
	IsActive   bool   `json:"isActive" db:"is_active"`
	IsResolved bool   `json:"isResolved" db:"is_resolved"`

	// Ancient connections
	AncientCause       bool `json:"ancientCause" db:"ancient_cause"`
	AwakensAncientEvil bool `json:"awakensAncientEvil" db:"awakens_ancient_evil"`
	ProphecyRelated    bool `json:"prophecyRelated" db:"prophecy_related"`

	// Effects
	AffectedRegions     JSONB `json:"affectedRegions" db:"affected_regions"`
	AffectedSettlements JSONB `json:"affectedSettlements" db:"affected_settlements"`
	AffectedFactions    JSONB `json:"affectedFactions" db:"affected_factions"`
	EconomicImpacts     JSONB `json:"economicImpacts" db:"economic_impacts"`
	PoliticalImpacts    JSONB `json:"politicalImpacts" db:"political_impacts"`

	// Progression
	CurrentStage         int   `json:"currentStage" db:"current_stage"`
	Stages              JSONB `json:"stages" db:"stages"`
	ResolutionConditions JSONB `json:"resolutionConditions" db:"resolution_conditions"`
	Consequences         JSONB `json:"consequences" db:"consequences"`

	// Player interaction
	PartyAware    bool  `json:"partyAware" db:"party_aware"`
	PartyInvolved bool  `json:"partyInvolved" db:"party_involved"`
	PartyActions  JSONB `json:"partyActions" db:"party_actions"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Market represents economic conditions in a settlement
type Market struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SettlementID uuid.UUID `json:"settlementId" db:"settlement_id"`

	// Price modifiers
	FoodPriceModifier           float64 `json:"foodPriceModifier" db:"food_price_modifier"`
	CommonGoodsModifier         float64 `json:"commonGoodsModifier" db:"common_goods_modifier"`
	WeaponsArmorModifier        float64 `json:"weaponsArmorModifier" db:"weapons_armor_modifier"`
	MagicalItemsModifier        float64 `json:"magicalItemsModifier" db:"magical_items_modifier"`
	AncientArtifactsModifier    float64 `json:"ancientArtifactsModifier" db:"ancient_artifacts_modifier"`

	// Supply and demand
	HighDemandItems JSONB `json:"highDemandItems" db:"high_demand_items"`
	SurplusItems    JSONB `json:"surplusItems" db:"surplus_items"`
	BannedItems     JSONB `json:"bannedItems" db:"banned_items"`

	// Special conditions
	BlackMarketActive      bool `json:"blackMarketActive" db:"black_market_active"`
	ArtifactDealerPresent  bool `json:"artifactDealerPresent" db:"artifact_dealer_present"`
	EconomicBoom           bool `json:"economicBoom" db:"economic_boom"`
	EconomicDepression     bool `json:"economicDepression" db:"economic_depression"`

	LastUpdated time.Time `json:"lastUpdated" db:"last_updated"`
}

// TradeRoute represents a connection between settlements
type TradeRoute struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	GameSessionID     uuid.UUID  `json:"gameSessionId" db:"game_session_id"`
	Name              string     `json:"name" db:"name"`
	StartSettlementID uuid.UUID  `json:"startSettlementId" db:"start_settlement_id"`
	EndSettlementID   uuid.UUID  `json:"endSettlementId" db:"end_settlement_id"`

	RouteType        string `json:"routeType" db:"route_type"`
	Distance         int    `json:"distance" db:"distance"`
	DifficultyRating int    `json:"difficultyRating" db:"difficulty_rating"`

	// Threats
	BanditThreatLevel  int   `json:"banditThreatLevel" db:"bandit_threat_level"`
	MonsterThreatLevel int   `json:"monsterThreatLevel" db:"monster_threat_level"`
	AncientHazards     bool  `json:"ancientHazards" db:"ancient_hazards"`
	EnvironmentalHazards JSONB `json:"environmentalHazards" db:"environmental_hazards"`

	// Economics
	TradeVolume   int     `json:"tradeVolume" db:"trade_volume"`
	PrimaryGoods  JSONB   `json:"primaryGoods" db:"primary_goods"`
	TariffRate    float64 `json:"tariffRate" db:"tariff_rate"`

	// Control
	ControllingFactionID *uuid.UUID `json:"controllingFactionId" db:"controlling_faction_id"`
	ProtectionFee        float64    `json:"protectionFee" db:"protection_fee"`

	IsActive         bool      `json:"isActive" db:"is_active"`
	DisruptionEvents JSONB     `json:"disruptionEvents" db:"disruption_events"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// AncientSite represents a location from the old world
type AncientSite struct {
	ID            uuid.UUID `json:"id" db:"id"`
	GameSessionID uuid.UUID `json:"gameSessionId" db:"game_session_id"`
	Name          string    `json:"name" db:"name"`
	TrueName      *string   `json:"trueName" db:"true_name"`
	Type          string    `json:"type" db:"type"`
	AgeCategory   string    `json:"ageCategory" db:"age_category"`

	LocationDescription  string     `json:"locationDescription" db:"location_description"`
	NearestSettlementID  *uuid.UUID `json:"nearestSettlementId" db:"nearest_settlement_id"`
	Coordinates          JSONB      `json:"coordinates" db:"coordinates"`

	// State
	ExplorationLevel    int `json:"explorationLevel" db:"exploration_level"`
	CorruptionLevel     int `json:"corruptionLevel" db:"corruption_level"`
	StructuralIntegrity int `json:"structuralIntegrity" db:"structural_integrity"`

	// Dangers and treasures
	GuardianType      *string `json:"guardianType" db:"guardian_type"`
	GuardianDefeated  bool    `json:"guardianDefeated" db:"guardian_defeated"`
	SealsIntact       bool    `json:"sealsIntact" db:"seals_intact"`
	Treasures         JSONB   `json:"treasures" db:"treasures"`
	Artifacts         JSONB   `json:"artifacts" db:"artifacts"`
	ForbiddenKnowledge JSONB  `json:"forbiddenKnowledge" db:"forbidden_knowledge"`

	// World effects
	LeyLineNexus      bool  `json:"leyLineNexus" db:"ley_line_nexus"`
	RealityWeakness   int   `json:"realityWeakness" db:"reality_weakness"`
	PlanarConnections JSONB `json:"planarConnections" db:"planar_connections"`

	// History
	OriginalPurpose string `json:"originalPurpose" db:"original_purpose"`
	FallDescription string `json:"fallDescription" db:"fall_description"`
	Prophecies      JSONB  `json:"prophecies" db:"prophecies"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Request/Response types for API

// SettlementGenerationRequest for generating a new settlement
type SettlementGenerationRequest struct {
	Name              string         `json:"name,omitempty"`
	Type              SettlementType `json:"type"`
	Region            string         `json:"region"`
	PopulationSize    string         `json:"populationSize,omitempty"` // small, medium, large
	DangerLevel       int            `json:"dangerLevel,omitempty"`
	AncientInfluence  bool           `json:"ancientInfluence"`
	SpecialFeatures   []string       `json:"specialFeatures,omitempty"`
}

// FactionCreationRequest for creating a new faction
type FactionCreationRequest struct {
	Name        string      `json:"name"`
	Type        FactionType `json:"type"`
	Description string      `json:"description"`
	Goals       []string    `json:"goals"`
	AncientTies bool        `json:"ancientTies"`
}

// WorldEventCreationRequest for creating world events
type WorldEventCreationRequest struct {
	Type              WorldEventType     `json:"type"`
	Severity          WorldEventSeverity `json:"severity"`
	AffectedRegions   []string          `json:"affectedRegions"`
	AncientConnection bool              `json:"ancientConnection"`
}