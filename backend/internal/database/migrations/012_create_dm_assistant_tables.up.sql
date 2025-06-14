-- Table for storing NPC templates and generated NPCs
CREATE TABLE IF NOT EXISTS ai_npcs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID REFERENCES game_sessions(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    race VARCHAR(50),
    occupation VARCHAR(100),
    personality_traits JSONB NOT NULL DEFAULT '[]', -- Array of personality descriptors
    appearance TEXT,
    voice_description TEXT,
    motivations TEXT,
    secrets TEXT,
    dialog_style TEXT, -- How they speak (formal, slang, accent, etc.)
    relationship_to_party TEXT,
    stat_block JSONB, -- Optional combat stats if needed
    generated_dialog JSONB DEFAULT '[]', -- History of generated dialog
    created_by UUID REFERENCES users(id),
    is_recurring BOOLEAN DEFAULT false,
    last_seen_session UUID,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table for storing generated locations
CREATE TABLE IF NOT EXISTS ai_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID REFERENCES game_sessions(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(50) NOT NULL, -- tavern, dungeon, shop, wilderness, etc.
    description TEXT NOT NULL,
    atmosphere TEXT,
    notable_features JSONB DEFAULT '[]',
    npcs_present JSONB DEFAULT '[]', -- References to ai_npcs
    available_actions JSONB DEFAULT '[]', -- What players can do here
    secrets_and_hidden JSONB DEFAULT '[]',
    environmental_effects TEXT,
    created_by UUID REFERENCES users(id),
    parent_location_id UUID REFERENCES ai_locations(id), -- For nested locations
    is_discovered BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table for combat narrations and dramatic moments
CREATE TABLE IF NOT EXISTS ai_narrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID REFERENCES game_sessions(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- combat_hit, combat_miss, death, critical, dramatic_moment
    context JSONB NOT NULL, -- Contains relevant info like damage, attacker, target, etc.
    narration TEXT NOT NULL,
    intensity_level INTEGER DEFAULT 5 CHECK (intensity_level BETWEEN 1 AND 10),
    tags JSONB DEFAULT '[]', -- Tags like "brutal", "heroic", "comedic"
    created_by UUID REFERENCES users(id),
    used_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table for plot twists and story hooks
CREATE TABLE IF NOT EXISTS ai_story_elements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID REFERENCES game_sessions(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- plot_twist, story_hook, revelation, complication
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    context JSONB NOT NULL, -- Current story state that led to this
    impact_level VARCHAR(20) CHECK (impact_level IN ('minor', 'moderate', 'major', 'campaign-changing')),
    suggested_timing TEXT,
    prerequisites JSONB DEFAULT '[]', -- Conditions that should be met
    consequences JSONB DEFAULT '[]', -- What happens after
    foreshadowing_hints JSONB DEFAULT '[]',
    created_by UUID REFERENCES users(id),
    used BOOLEAN DEFAULT false,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table for environmental hazards and challenges
CREATE TABLE IF NOT EXISTS ai_environmental_hazards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID REFERENCES game_sessions(id) ON DELETE CASCADE,
    location_id UUID REFERENCES ai_locations(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    trigger_condition TEXT, -- What causes this to activate
    effect_description TEXT NOT NULL,
    mechanical_effects JSONB NOT NULL, -- Game mechanics (damage, saves, etc.)
    difficulty_class INTEGER,
    damage_formula VARCHAR(50), -- e.g., "2d6 fire"
    avoidance_hints TEXT,
    is_trap BOOLEAN DEFAULT false,
    is_natural BOOLEAN DEFAULT true,
    reset_condition TEXT, -- For traps that reset
    created_by UUID REFERENCES users(id),
    is_active BOOLEAN DEFAULT true,
    triggered_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table for DM assistant conversation history
CREATE TABLE IF NOT EXISTS dm_assistant_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID REFERENCES game_sessions(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    request_type VARCHAR(50) NOT NULL, -- npc_dialog, location, combat_narration, etc.
    request_context JSONB NOT NULL,
    prompt TEXT NOT NULL,
    response TEXT NOT NULL,
    feedback VARCHAR(20), -- positive, negative, neutral
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_ai_npcs_session ON ai_npcs(game_session_id);
CREATE INDEX idx_ai_npcs_recurring ON ai_npcs(is_recurring) WHERE is_recurring = true;
CREATE INDEX idx_ai_locations_session ON ai_locations(game_session_id);
CREATE INDEX idx_ai_locations_type ON ai_locations(type);
CREATE INDEX idx_ai_narrations_session ON ai_narrations(game_session_id);
CREATE INDEX idx_ai_narrations_type ON ai_narrations(type);
CREATE INDEX idx_ai_story_elements_session ON ai_story_elements(game_session_id);
CREATE INDEX idx_ai_story_elements_unused ON ai_story_elements(used) WHERE used = false;
CREATE INDEX idx_ai_environmental_hazards_location ON ai_environmental_hazards(location_id);
CREATE INDEX idx_dm_assistant_history_session ON dm_assistant_history(game_session_id);
CREATE INDEX idx_dm_assistant_history_user ON dm_assistant_history(user_id);