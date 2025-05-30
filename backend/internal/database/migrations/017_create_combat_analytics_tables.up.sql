-- Combat Analytics and Automation Tables

-- Combat Analytics table - stores detailed combat statistics
CREATE TABLE IF NOT EXISTS combat_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    combat_id UUID NOT NULL,
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    combat_duration INTEGER NOT NULL, -- Duration in rounds
    total_damage_dealt INTEGER DEFAULT 0,
    total_healing_done INTEGER DEFAULT 0,
    killing_blows JSONB DEFAULT '[]'::jsonb, -- Array of {dealer_id, target_id, damage}
    combat_summary JSONB DEFAULT '{}'::jsonb, -- AI-generated summary
    mvp_id VARCHAR(255), -- Character/NPC who dealt most damage
    mvp_type VARCHAR(50), -- 'character' or 'npc'
    tactical_rating INTEGER, -- 1-10 rating of tactical effectiveness
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Combatant Analytics table - individual performance metrics
CREATE TABLE IF NOT EXISTS combatant_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    combat_analytics_id UUID NOT NULL REFERENCES combat_analytics(id) ON DELETE CASCADE,
    combatant_id VARCHAR(255) NOT NULL,
    combatant_type VARCHAR(50) NOT NULL, -- 'character' or 'npc'
    combatant_name VARCHAR(255) NOT NULL,
    damage_dealt INTEGER DEFAULT 0,
    damage_taken INTEGER DEFAULT 0,
    healing_done INTEGER DEFAULT 0,
    healing_received INTEGER DEFAULT 0,
    attacks_made INTEGER DEFAULT 0,
    attacks_hit INTEGER DEFAULT 0,
    attacks_missed INTEGER DEFAULT 0,
    critical_hits INTEGER DEFAULT 0,
    critical_misses INTEGER DEFAULT 0,
    saves_made INTEGER DEFAULT 0,
    saves_failed INTEGER DEFAULT 0,
    rounds_survived INTEGER DEFAULT 0,
    final_hp INTEGER,
    conditions_suffered JSONB DEFAULT '[]'::jsonb, -- Array of conditions
    abilities_used JSONB DEFAULT '[]'::jsonb, -- Array of abilities/spells
    tactical_decisions JSONB DEFAULT '[]'::jsonb, -- AI analysis of decisions
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Auto-Combat Resolutions table - for quick combat resolution
CREATE TABLE IF NOT EXISTS auto_combat_resolutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    encounter_difficulty VARCHAR(50) NOT NULL, -- trivial, easy, medium, hard, deadly
    party_composition JSONB NOT NULL, -- Array of character info
    enemy_composition JSONB NOT NULL, -- Array of enemy info
    resolution_type VARCHAR(50) NOT NULL, -- 'quick', 'simulated', 'detailed'
    outcome VARCHAR(50) NOT NULL, -- 'victory', 'defeat', 'retreat'
    rounds_simulated INTEGER,
    party_resources_used JSONB DEFAULT '{}'::jsonb, -- HP lost, spell slots, etc.
    loot_generated JSONB DEFAULT '[]'::jsonb,
    experience_awarded INTEGER DEFAULT 0,
    narrative_summary TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Battle Maps table - stores generated tactical maps
CREATE TABLE IF NOT EXISTS battle_maps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    combat_id UUID,
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    location_description TEXT NOT NULL,
    map_type VARCHAR(50) NOT NULL, -- 'dungeon', 'outdoor', 'urban', 'special'
    grid_size_x INTEGER NOT NULL DEFAULT 20,
    grid_size_y INTEGER NOT NULL DEFAULT 20,
    terrain_features JSONB NOT NULL, -- Array of terrain objects
    obstacle_positions JSONB DEFAULT '[]'::jsonb, -- Array of obstacles
    cover_positions JSONB DEFAULT '[]'::jsonb, -- Array of cover locations
    hazard_zones JSONB DEFAULT '[]'::jsonb, -- Array of hazardous areas
    spawn_points JSONB DEFAULT '[]'::jsonb, -- Suggested spawn locations
    tactical_notes JSONB DEFAULT '[]'::jsonb, -- AI-generated tactical advice
    visual_theme VARCHAR(100), -- For frontend rendering
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Smart Initiative table - stores initiative bonuses and special rules
CREATE TABLE IF NOT EXISTS smart_initiative_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    entity_id VARCHAR(255) NOT NULL, -- Character or NPC ID
    entity_type VARCHAR(50) NOT NULL, -- 'character' or 'npc'
    base_initiative_bonus INTEGER DEFAULT 0,
    advantage_on_initiative BOOLEAN DEFAULT FALSE,
    alert_feat BOOLEAN DEFAULT FALSE, -- +5 to initiative
    special_rules JSONB DEFAULT '{}'::jsonb, -- Custom rules
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(game_session_id, entity_id)
);

-- Combat Action Log table - detailed action tracking for analytics
CREATE TABLE IF NOT EXISTS combat_action_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    combat_id UUID NOT NULL,
    round_number INTEGER NOT NULL,
    turn_number INTEGER NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    actor_type VARCHAR(50) NOT NULL,
    action_type VARCHAR(100) NOT NULL, -- 'attack', 'spell', 'ability', 'move', etc.
    target_id VARCHAR(255),
    target_type VARCHAR(50),
    roll_results JSONB DEFAULT '{}'::jsonb, -- Attack rolls, damage rolls, etc.
    outcome VARCHAR(100), -- 'hit', 'miss', 'critical', etc.
    damage_dealt INTEGER DEFAULT 0,
    conditions_applied JSONB DEFAULT '[]'::jsonb,
    resources_used JSONB DEFAULT '{}'::jsonb, -- Spell slots, abilities, etc.
    position_data JSONB DEFAULT '{}'::jsonb, -- Movement and positioning
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_combat_analytics_session ON combat_analytics(game_session_id);
CREATE INDEX idx_combat_analytics_combat ON combat_analytics(combat_id);
CREATE INDEX idx_combatant_analytics_combat ON combatant_analytics(combat_analytics_id);
CREATE INDEX idx_auto_resolutions_session ON auto_combat_resolutions(game_session_id);
CREATE INDEX idx_battle_maps_session ON battle_maps(game_session_id);
CREATE INDEX idx_battle_maps_combat ON battle_maps(combat_id);
CREATE INDEX idx_smart_initiative_session ON smart_initiative_rules(game_session_id);
CREATE INDEX idx_combat_action_log_combat ON combat_action_log(combat_id);
CREATE INDEX idx_combat_action_log_actor ON combat_action_log(actor_id);

-- Trigger to update combat_analytics updated_at
CREATE OR REPLACE FUNCTION update_combat_analytics_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_combat_analytics_timestamp_trigger
BEFORE UPDATE ON combat_analytics
FOR EACH ROW
EXECUTE FUNCTION update_combat_analytics_timestamp();