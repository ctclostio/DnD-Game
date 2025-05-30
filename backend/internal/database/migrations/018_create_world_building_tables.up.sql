-- World Building Tables for Ancient World Setting

-- Settlements table
CREATE TABLE IF NOT EXISTS settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- hamlet, village, town, city, metropolis, ruins
    population INTEGER NOT NULL,
    age_category VARCHAR(50) NOT NULL, -- ancient, old, established, recent, new
    description TEXT,
    history TEXT, -- Ancient history and connections to the old world
    government_type VARCHAR(100),
    alignment VARCHAR(50),
    danger_level INTEGER DEFAULT 1 CHECK (danger_level BETWEEN 1 AND 10),
    corruption_level INTEGER DEFAULT 0 CHECK (corruption_level BETWEEN 0 AND 10), -- Influence of ancient evils
    
    -- Location data
    region VARCHAR(255),
    coordinates JSONB, -- {x: 0, y: 0} on world map
    terrain_type VARCHAR(100),
    climate VARCHAR(100),
    
    -- Economic data
    wealth_level INTEGER DEFAULT 3 CHECK (wealth_level BETWEEN 1 AND 10),
    primary_exports JSONB DEFAULT '[]'::jsonb,
    primary_imports JSONB DEFAULT '[]'::jsonb,
    trade_routes JSONB DEFAULT '[]'::jsonb, -- Connected settlement IDs
    
    -- Ancient connections
    ancient_ruins_nearby BOOLEAN DEFAULT false,
    eldritch_influence INTEGER DEFAULT 0 CHECK (eldritch_influence BETWEEN 0 AND 10),
    ley_line_connection BOOLEAN DEFAULT false,
    
    -- Notable features
    notable_locations JSONB DEFAULT '[]'::jsonb,
    defenses JSONB DEFAULT '[]'::jsonb,
    problems JSONB DEFAULT '[]'::jsonb, -- Current issues/plot hooks
    secrets JSONB DEFAULT '[]'::jsonb, -- Hidden knowledge about ancient times
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Settlement NPCs
CREATE TABLE IF NOT EXISTS settlement_npcs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    settlement_id UUID NOT NULL REFERENCES settlements(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    race VARCHAR(100),
    class VARCHAR(100),
    level INTEGER DEFAULT 1,
    role VARCHAR(255), -- mayor, guard captain, innkeeper, merchant, etc.
    occupation VARCHAR(255),
    personality_traits JSONB DEFAULT '[]'::jsonb,
    ideals JSONB DEFAULT '[]'::jsonb,
    bonds JSONB DEFAULT '[]'::jsonb,
    flaws JSONB DEFAULT '[]'::jsonb,
    
    -- Ancient connections
    ancient_knowledge BOOLEAN DEFAULT false,
    corruption_touched BOOLEAN DEFAULT false,
    secret_agenda TEXT,
    true_age INTEGER, -- Some NPCs might be far older than they appear
    
    -- Relationships
    faction_affiliations JSONB DEFAULT '[]'::jsonb,
    relationships JSONB DEFAULT '{}'::jsonb, -- {npc_id: relationship_type}
    
    -- Mechanical data
    stats JSONB DEFAULT '{}'::jsonb,
    skills JSONB DEFAULT '{}'::jsonb,
    inventory JSONB DEFAULT '[]'::jsonb,
    plot_hooks JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Settlement Shops
CREATE TABLE IF NOT EXISTS settlement_shops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    settlement_id UUID NOT NULL REFERENCES settlements(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL, -- general, weaponsmith, armorer, alchemist, magic, tavern, inn, temple
    owner_npc_id UUID REFERENCES settlement_npcs(id) ON DELETE SET NULL,
    quality_level INTEGER DEFAULT 3 CHECK (quality_level BETWEEN 1 AND 10),
    price_modifier DECIMAL(3,2) DEFAULT 1.0, -- Multiplier for base prices
    
    -- Inventory
    available_items JSONB DEFAULT '[]'::jsonb, -- Item IDs and quantities
    special_items JSONB DEFAULT '[]'::jsonb, -- Rare/unique items
    can_craft BOOLEAN DEFAULT false,
    crafting_specialties JSONB DEFAULT '[]'::jsonb,
    
    -- Special features
    black_market BOOLEAN DEFAULT false,
    ancient_artifacts BOOLEAN DEFAULT false, -- Deals in items from the old world
    faction_discount JSONB DEFAULT '{}'::jsonb, -- {faction_id: discount_percent}
    
    reputation_required INTEGER DEFAULT 0,
    operating_hours JSONB DEFAULT '{}'::jsonb,
    current_rumors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Factions
CREATE TABLE IF NOT EXISTS factions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL, -- religious, political, criminal, merchant, military, cult, ancient_order
    description TEXT,
    founding_date VARCHAR(100), -- Could be "Lost to time" for ancient factions
    
    -- Goals and motivations
    public_goals JSONB DEFAULT '[]'::jsonb,
    secret_goals JSONB DEFAULT '[]'::jsonb,
    motivations JSONB DEFAULT '[]'::jsonb,
    
    -- Ancient connections
    ancient_knowledge_level INTEGER DEFAULT 0 CHECK (ancient_knowledge_level BETWEEN 0 AND 10),
    seeks_ancient_power BOOLEAN DEFAULT false,
    guards_ancient_secrets BOOLEAN DEFAULT false,
    corrupted BOOLEAN DEFAULT false,
    
    -- Resources and power
    influence_level INTEGER DEFAULT 3 CHECK (influence_level BETWEEN 1 AND 10),
    military_strength INTEGER DEFAULT 3 CHECK (military_strength BETWEEN 1 AND 10),
    economic_power INTEGER DEFAULT 3 CHECK (economic_power BETWEEN 1 AND 10),
    magical_resources INTEGER DEFAULT 1 CHECK (magical_resources BETWEEN 1 AND 10),
    
    -- Organizational data
    leadership_structure VARCHAR(255),
    headquarters_location VARCHAR(255),
    member_count INTEGER DEFAULT 0,
    territory_control JSONB DEFAULT '[]'::jsonb, -- Settlement IDs
    
    -- Relationships (stored as JSONB for flexibility)
    faction_relationships JSONB DEFAULT '{}'::jsonb, -- {faction_id: {standing: -100 to 100, type: 'ally/enemy/neutral'}}
    
    symbols JSONB DEFAULT '{}'::jsonb, -- Heraldry, colors, etc.
    rituals JSONB DEFAULT '[]'::jsonb,
    resources JSONB DEFAULT '{}'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Faction Members
CREATE TABLE IF NOT EXISTS faction_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    faction_id UUID NOT NULL REFERENCES factions(id) ON DELETE CASCADE,
    npc_id UUID REFERENCES settlement_npcs(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    rank VARCHAR(100) NOT NULL,
    reputation INTEGER DEFAULT 0,
    join_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    secret_member BOOLEAN DEFAULT false,
    special_role TEXT,
    
    CONSTRAINT faction_member_entity CHECK (
        (npc_id IS NOT NULL AND character_id IS NULL) OR 
        (npc_id IS NULL AND character_id IS NOT NULL)
    )
);

-- World Events
CREATE TABLE IF NOT EXISTS world_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL, -- political, economic, natural, supernatural, ancient_awakening, planar
    severity VARCHAR(50) NOT NULL, -- minor, moderate, major, catastrophic
    
    description TEXT,
    cause TEXT, -- What triggered this event
    
    -- Timing
    start_date VARCHAR(100) NOT NULL, -- In-game date
    duration VARCHAR(100), -- Expected duration
    is_active BOOLEAN DEFAULT true,
    is_resolved BOOLEAN DEFAULT false,
    
    -- Ancient world connections
    ancient_cause BOOLEAN DEFAULT false, -- Caused by ancient powers/artifacts
    awakens_ancient_evil BOOLEAN DEFAULT false,
    prophecy_related BOOLEAN DEFAULT false,
    
    -- Effects
    affected_regions JSONB DEFAULT '[]'::jsonb,
    affected_settlements JSONB DEFAULT '[]'::jsonb,
    affected_factions JSONB DEFAULT '[]'::jsonb,
    economic_impacts JSONB DEFAULT '{}'::jsonb,
    political_impacts JSONB DEFAULT '{}'::jsonb,
    
    -- Progression
    current_stage INTEGER DEFAULT 1,
    stages JSONB DEFAULT '[]'::jsonb, -- Array of stage descriptions
    resolution_conditions JSONB DEFAULT '[]'::jsonb,
    consequences JSONB DEFAULT '{}'::jsonb,
    
    -- Player interaction
    party_aware BOOLEAN DEFAULT false,
    party_involved BOOLEAN DEFAULT false,
    party_actions JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Economic Data
CREATE TABLE IF NOT EXISTS markets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    settlement_id UUID NOT NULL REFERENCES settlements(id) ON DELETE CASCADE,
    
    -- Base prices (multipliers of standard prices)
    food_price_modifier DECIMAL(3,2) DEFAULT 1.0,
    common_goods_modifier DECIMAL(3,2) DEFAULT 1.0,
    weapons_armor_modifier DECIMAL(3,2) DEFAULT 1.0,
    magical_items_modifier DECIMAL(3,2) DEFAULT 1.0,
    ancient_artifacts_modifier DECIMAL(3,2) DEFAULT 2.0, -- Always expensive
    
    -- Supply and demand
    high_demand_items JSONB DEFAULT '[]'::jsonb,
    surplus_items JSONB DEFAULT '[]'::jsonb,
    banned_items JSONB DEFAULT '[]'::jsonb,
    
    -- Special market conditions
    black_market_active BOOLEAN DEFAULT false,
    artifact_dealer_present BOOLEAN DEFAULT false,
    economic_boom BOOLEAN DEFAULT false,
    economic_depression BOOLEAN DEFAULT false,
    
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Trade Routes
CREATE TABLE IF NOT EXISTS trade_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    start_settlement_id UUID NOT NULL REFERENCES settlements(id) ON DELETE CASCADE,
    end_settlement_id UUID NOT NULL REFERENCES settlements(id) ON DELETE CASCADE,
    
    route_type VARCHAR(50) NOT NULL, -- land, sea, air, underground, planar
    distance INTEGER NOT NULL, -- in days of travel
    difficulty_rating INTEGER DEFAULT 3 CHECK (difficulty_rating BETWEEN 1 AND 10),
    
    -- Hazards
    bandit_threat_level INTEGER DEFAULT 0 CHECK (bandit_threat_level BETWEEN 0 AND 10),
    monster_threat_level INTEGER DEFAULT 0 CHECK (monster_threat_level BETWEEN 0 AND 10),
    ancient_hazards BOOLEAN DEFAULT false, -- Old curses, awakened guardians, etc.
    environmental_hazards JSONB DEFAULT '[]'::jsonb,
    
    -- Economics
    trade_volume INTEGER DEFAULT 3 CHECK (trade_volume BETWEEN 1 AND 10),
    primary_goods JSONB DEFAULT '[]'::jsonb,
    tariff_rate DECIMAL(3,2) DEFAULT 0.1,
    
    -- Control
    controlling_faction_id UUID REFERENCES factions(id) ON DELETE SET NULL,
    protection_fee DECIMAL(10,2) DEFAULT 0,
    
    is_active BOOLEAN DEFAULT true,
    disruption_events JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Ancient Sites (for world building context)
CREATE TABLE IF NOT EXISTS ancient_sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    true_name VARCHAR(255), -- Name in the old tongue
    type VARCHAR(100) NOT NULL, -- temple, fortress, city, prison, seal, portal
    age_category VARCHAR(50) NOT NULL, -- first_age, second_age, etc.
    
    location_description TEXT,
    nearest_settlement_id UUID REFERENCES settlements(id) ON DELETE SET NULL,
    coordinates JSONB,
    
    -- State
    exploration_level INTEGER DEFAULT 0 CHECK (exploration_level BETWEEN 0 AND 100),
    corruption_level INTEGER DEFAULT 5 CHECK (corruption_level BETWEEN 0 AND 10),
    structural_integrity INTEGER DEFAULT 3 CHECK (structural_integrity BETWEEN 0 AND 10),
    
    -- Dangers and treasures
    guardian_type VARCHAR(255),
    guardian_defeated BOOLEAN DEFAULT false,
    seals_intact BOOLEAN DEFAULT true,
    treasures JSONB DEFAULT '[]'::jsonb,
    artifacts JSONB DEFAULT '[]'::jsonb,
    forbidden_knowledge JSONB DEFAULT '[]'::jsonb,
    
    -- Effects on world
    ley_line_nexus BOOLEAN DEFAULT false,
    reality_weakness INTEGER DEFAULT 0 CHECK (reality_weakness BETWEEN 0 AND 10),
    planar_connections JSONB DEFAULT '[]'::jsonb,
    
    -- History
    original_purpose TEXT,
    fall_description TEXT,
    prophecies JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_settlements_game_session ON settlements(game_session_id);
CREATE INDEX idx_settlement_npcs_settlement ON settlement_npcs(settlement_id);
CREATE INDEX idx_settlement_shops_settlement ON settlement_shops(settlement_id);
CREATE INDEX idx_factions_game_session ON factions(game_session_id);
CREATE INDEX idx_world_events_game_session ON world_events(game_session_id);
CREATE INDEX idx_world_events_active ON world_events(is_active);
CREATE INDEX idx_trade_routes_settlements ON trade_routes(start_settlement_id, end_settlement_id);
CREATE INDEX idx_ancient_sites_game_session ON ancient_sites(game_session_id);
CREATE INDEX idx_ancient_sites_nearest_settlement ON ancient_sites(nearest_settlement_id);