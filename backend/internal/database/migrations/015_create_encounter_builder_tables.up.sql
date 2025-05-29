-- Create encounters table
CREATE TABLE IF NOT EXISTS encounters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    location VARCHAR(255),
    encounter_type VARCHAR(50) NOT NULL CHECK (encounter_type IN ('combat', 'social', 'exploration', 'puzzle', 'hybrid')),
    difficulty VARCHAR(20) NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard', 'deadly')),
    challenge_rating DECIMAL(4,2),
    
    -- Context for AI generation
    narrative_context TEXT,
    environmental_features TEXT[],
    story_hooks TEXT[],
    
    -- Party information snapshot
    party_level INTEGER NOT NULL,
    party_size INTEGER NOT NULL,
    party_composition JSONB NOT NULL, -- Store class breakdown
    
    -- Enemy/Challenge information
    enemies JSONB NOT NULL DEFAULT '[]',
    total_xp INTEGER,
    adjusted_xp INTEGER,
    
    -- Tactical information
    enemy_tactics JSONB,
    environmental_hazards JSONB,
    terrain_features JSONB,
    
    -- Non-combat options
    social_solutions JSONB,
    stealth_options JSONB,
    environmental_solutions JSONB,
    
    -- Dynamic scaling
    scaling_options JSONB,
    reinforcement_waves JSONB,
    escape_routes JSONB,
    
    -- Status
    status VARCHAR(20) DEFAULT 'planned' CHECK (status IN ('planned', 'active', 'completed', 'abandoned')),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    outcome TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create encounter_enemies table for detailed enemy tracking
CREATE TABLE IF NOT EXISTS encounter_enemies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    encounter_id UUID NOT NULL REFERENCES encounters(id) ON DELETE CASCADE,
    npc_id UUID REFERENCES ai_npcs(id),
    
    -- Enemy details
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    size VARCHAR(20),
    challenge_rating DECIMAL(4,2),
    hit_points INTEGER NOT NULL,
    armor_class INTEGER NOT NULL,
    
    -- Combat stats
    stats JSONB NOT NULL,
    abilities JSONB NOT NULL,
    actions JSONB NOT NULL,
    legendary_actions JSONB,
    
    -- AI behavior
    personality_traits TEXT[],
    ideal TEXT,
    bond TEXT,
    flaw TEXT,
    tactics TEXT,
    morale_threshold INTEGER DEFAULT 50, -- HP percentage when they might flee
    
    -- Positioning and status
    initial_position JSONB,
    current_position JSONB,
    conditions TEXT[],
    is_alive BOOLEAN DEFAULT true,
    fled BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create encounter_objectives table
CREATE TABLE IF NOT EXISTS encounter_objectives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    encounter_id UUID NOT NULL REFERENCES encounters(id) ON DELETE CASCADE,
    
    -- Objective details
    type VARCHAR(50) NOT NULL CHECK (type IN ('defeat_all', 'survive_rounds', 'protect_npc', 'reach_location', 'retrieve_item', 'solve_puzzle', 'negotiate', 'escape', 'custom')),
    description TEXT NOT NULL,
    success_conditions JSONB NOT NULL,
    failure_conditions JSONB,
    
    -- Rewards
    xp_reward INTEGER,
    gold_reward INTEGER,
    item_rewards JSONB,
    story_rewards TEXT[],
    
    -- Status
    is_completed BOOLEAN DEFAULT false,
    is_failed BOOLEAN DEFAULT false,
    completed_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create encounter_events table for tracking what happens
CREATE TABLE IF NOT EXISTS encounter_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    encounter_id UUID NOT NULL REFERENCES encounters(id) ON DELETE CASCADE,
    round_number INTEGER NOT NULL,
    
    -- Event details
    event_type VARCHAR(50) NOT NULL,
    actor_type VARCHAR(20) CHECK (actor_type IN ('player', 'enemy', 'environment', 'system')),
    actor_id UUID,
    actor_name VARCHAR(255),
    
    description TEXT NOT NULL,
    mechanical_effect JSONB,
    
    -- For tactical suggestions
    ai_suggestion TEXT,
    suggestion_used BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create encounter_templates table for reusable encounters
CREATE TABLE IF NOT EXISTS encounter_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by UUID REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    tags TEXT[],
    
    -- Template data
    encounter_type VARCHAR(50) NOT NULL,
    min_level INTEGER NOT NULL,
    max_level INTEGER NOT NULL,
    environment_types TEXT[],
    
    -- Enemy templates
    enemy_groups JSONB NOT NULL,
    scaling_formula JSONB,
    
    -- Tactical templates
    tactical_notes TEXT,
    environmental_features JSONB,
    objective_options JSONB,
    
    is_public BOOLEAN DEFAULT false,
    times_used INTEGER DEFAULT 0,
    rating DECIMAL(3,2),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_encounters_game_session ON encounters(game_session_id);
CREATE INDEX idx_encounters_status ON encounters(status);
CREATE INDEX idx_encounter_enemies_encounter ON encounter_enemies(encounter_id);
CREATE INDEX idx_encounter_objectives_encounter ON encounter_objectives(encounter_id);
CREATE INDEX idx_encounter_events_encounter ON encounter_events(encounter_id);
CREATE INDEX idx_encounter_templates_public ON encounter_templates(is_public);
CREATE INDEX idx_encounter_templates_levels ON encounter_templates(min_level, max_level);

-- Example enemies JSONB structure:
-- [
--   {
--     "id": "uuid",
--     "name": "Goblin Scout",
--     "type": "goblin",
--     "cr": 0.25,
--     "quantity": 3,
--     "role": "skirmisher",
--     "tactics": "Hit and run attacks, use cover"
--   }
-- ]

-- Example scaling_options JSONB structure:
-- {
--   "easy": {
--     "remove_enemies": ["weakest"],
--     "reduce_hp": 20,
--     "lower_damage": 2
--   },
--   "hard": {
--     "add_enemies": ["goblin_archer", "goblin_archer"],
--     "increase_hp": 20,
--     "add_legendary_actions": 1
--   },
--   "environmental": {
--     "add_hazard": "falling_rocks",
--     "difficult_terrain": true
--   }
-- }