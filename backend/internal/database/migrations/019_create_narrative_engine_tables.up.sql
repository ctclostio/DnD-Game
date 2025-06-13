-- Narrative Profile table for tracking player preferences
CREATE TABLE IF NOT EXISTS narrative_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    preferences JSONB NOT NULL DEFAULT '{}',
    decision_history JSONB NOT NULL DEFAULT '[]',
    play_style VARCHAR(50),
    analytics JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, character_id)
);

-- Backstory elements that can be woven into narratives
CREATE TABLE IF NOT EXISTS backstory_elements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- origin, trauma, goal, relationship, secret
    content TEXT NOT NULL,
    weight FLOAT DEFAULT 1.0,
    used BOOLEAN DEFAULT FALSE,
    usage_count INTEGER DEFAULT 0,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_backstory_character ON backstory_elements (character_id);
CREATE INDEX idx_backstory_type ON backstory_elements (type);
CREATE INDEX idx_backstory_used ON backstory_elements (used);

-- Personalized narratives generated for specific players
CREATE TABLE IF NOT EXISTS personalized_narratives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_event_id UUID NOT NULL,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    personalized_hooks JSONB NOT NULL DEFAULT '[]',
    backstory_callbacks JSONB NOT NULL DEFAULT '[]',
    emotional_resonance FLOAT DEFAULT 0.0,
    predicted_impact JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_personalized_character ON personalized_narratives (character_id);
CREATE INDEX idx_personalized_event ON personalized_narratives (base_event_id);

-- Consequence events tracking ripple effects
CREATE TABLE IF NOT EXISTS consequence_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trigger_action_id UUID NOT NULL,
    trigger_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    severity INTEGER CHECK (severity >= 1 AND severity <= 10),
    delay VARCHAR(50) NOT NULL, -- immediate, short, medium, long
    actual_trigger_time TIMESTAMP WITH TIME ZONE,
    affected_entities JSONB NOT NULL DEFAULT '[]',
    cascade_effects JSONB NOT NULL DEFAULT '[]',
    status VARCHAR(50) DEFAULT 'pending', -- pending, triggered, resolved, prevented
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_consequence_trigger ON consequence_events (trigger_action_id);
CREATE INDEX idx_consequence_status ON consequence_events (status);
CREATE INDEX idx_consequence_trigger_time ON consequence_events (actual_trigger_time);

-- World events that can have multiple perspectives
CREATE TABLE IF NOT EXISTS narrative_world_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    location VARCHAR(255),
    event_timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    participants TEXT[] DEFAULT '{}',
    witnesses TEXT[] DEFAULT '{}',
    immediate_effects TEXT[] DEFAULT '{}',
    player_involvement JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_narrative_world_event_type ON narrative_world_events (type);
CREATE INDEX idx_narrative_world_event_timestamp ON narrative_world_events (event_timestamp);
CREATE INDEX idx_narrative_world_event_status ON narrative_world_events (status);

-- Different perspectives on the same event
CREATE TABLE IF NOT EXISTS perspective_narratives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES narrative_world_events(id) ON DELETE CASCADE,
    perspective_type VARCHAR(50) NOT NULL, -- npc, faction, deity, historical
    source_id UUID NOT NULL,
    source_name VARCHAR(255) NOT NULL,
    narrative TEXT NOT NULL,
    bias VARCHAR(50) DEFAULT 'neutral', -- positive, negative, neutral, conflicted
    truth_level FLOAT DEFAULT 1.0 CHECK (truth_level >= 0 AND truth_level <= 1),
    hidden_details TEXT[] DEFAULT '{}',
    contradictions JSONB DEFAULT '[]',
    emotional_tone VARCHAR(50),
    cultural_context JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_perspective_event ON perspective_narratives (event_id);
CREATE INDEX idx_perspective_source ON perspective_narratives (source_id);
CREATE INDEX idx_perspective_type ON perspective_narratives (perspective_type);

-- Narrative memory for AI context
CREATE TABLE IF NOT EXISTS narrative_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES game_sessions(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    memory_type VARCHAR(50) NOT NULL, -- decision, consequence, relationship, discovery
    content TEXT NOT NULL,
    emotional_weight FLOAT DEFAULT 0.0,
    connections UUID[] DEFAULT '{}',
    active BOOLEAN DEFAULT TRUE,
    last_referenced TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    reference_count INTEGER DEFAULT 0,
    tags TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_memory_session ON narrative_memories (session_id);
CREATE INDEX idx_memory_character ON narrative_memories (character_id);
CREATE INDEX idx_memory_type ON narrative_memories (memory_type);
CREATE INDEX idx_memory_active ON narrative_memories (active);

-- Player actions that can trigger consequences
CREATE TABLE IF NOT EXISTS player_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    action_type VARCHAR(100) NOT NULL,
    target_type VARCHAR(100),
    target_id UUID,
    action_description TEXT NOT NULL,
    moral_weight VARCHAR(50), -- good, evil, neutral, chaotic, lawful
    immediate_result TEXT,
    potential_consequences INTEGER DEFAULT 0,
    action_timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_action_session ON player_actions (session_id);
CREATE INDEX idx_action_character ON player_actions (character_id);
CREATE INDEX idx_action_timestamp ON player_actions (action_timestamp);

-- Narrative threads connecting events across time
CREATE TABLE IF NOT EXISTS narrative_threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    thread_type VARCHAR(50) NOT NULL, -- main_quest, side_quest, character_arc, world_event
    status VARCHAR(50) DEFAULT 'active', -- active, dormant, resolved, abandoned
    connected_events UUID[] DEFAULT '{}',
    key_participants UUID[] DEFAULT '{}',
    tension_level FLOAT DEFAULT 0.5,
    resolution_proximity FLOAT DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_thread_status ON narrative_threads (status);
CREATE INDEX idx_thread_type ON narrative_threads (thread_type);

-- Track relationships between entities for narrative purposes
CREATE TABLE IF NOT EXISTS narrative_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity1_type VARCHAR(50) NOT NULL,
    entity1_id UUID NOT NULL,
    entity2_type VARCHAR(50) NOT NULL,
    entity2_id UUID NOT NULL,
    relationship_type VARCHAR(100) NOT NULL,
    strength FLOAT DEFAULT 0.0, -- -1 to 1 (hostile to allied)
    history JSONB DEFAULT '[]',
    last_interaction TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(entity1_id, entity2_id, relationship_type)
);

CREATE INDEX idx_relationship_entities ON narrative_relationships (entity1_id, entity2_id);
CREATE INDEX idx_relationship_type ON narrative_relationships (relationship_type);

-- Create update trigger for updated_at columns
CREATE OR REPLACE FUNCTION update_narrative_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_narrative_profiles_updated_at
    BEFORE UPDATE ON narrative_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_narrative_updated_at();

CREATE TRIGGER update_narrative_threads_updated_at
    BEFORE UPDATE ON narrative_threads
    FOR EACH ROW
    EXECUTE FUNCTION update_narrative_updated_at();

CREATE TRIGGER update_narrative_relationships_updated_at
    BEFORE UPDATE ON narrative_relationships
    FOR EACH ROW
    EXECUTE FUNCTION update_narrative_updated_at();