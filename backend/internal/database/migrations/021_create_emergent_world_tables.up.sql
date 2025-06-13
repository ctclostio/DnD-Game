-- World State table for tracking the living world
CREATE TABLE IF NOT EXISTS world_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    world_time TIMESTAMP WITH TIME ZONE NOT NULL,
    last_simulated TIMESTAMP WITH TIME ZONE NOT NULL,
    world_data JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for world_states
CREATE INDEX idx_world_states_session ON world_states(session_id);
CREATE INDEX idx_world_states_active ON world_states(is_active);

-- NPC Goals for autonomous behavior
CREATE TABLE IF NOT EXISTS npc_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    npc_id UUID NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    goal_type VARCHAR(50) NOT NULL,
    priority INTEGER DEFAULT 1,
    description TEXT,
    progress DECIMAL(3,2) DEFAULT 0.0,
    parameters JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for npc_goals
CREATE INDEX idx_npc_goals_npc ON npc_goals(npc_id);
CREATE INDEX idx_npc_goals_status ON npc_goals(status);
CREATE INDEX idx_npc_goals_priority ON npc_goals(priority DESC);

-- NPC Schedules for daily routines
CREATE TABLE IF NOT EXISTS npc_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    npc_id UUID NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    time_of_day VARCHAR(20) NOT NULL,
    activity VARCHAR(100) NOT NULL,
    location VARCHAR(255),
    parameters JSONB DEFAULT '{}'
);

-- Create indexes for npc_schedules
CREATE INDEX idx_npc_schedules_npc ON npc_schedules(npc_id);
CREATE INDEX idx_npc_schedules_time ON npc_schedules(time_of_day);

-- Faction Personalities for AI-driven behavior
CREATE TABLE IF NOT EXISTS faction_personalities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    faction_id UUID NOT NULL REFERENCES factions(id) ON DELETE CASCADE,
    traits JSONB NOT NULL,
    values JSONB NOT NULL,
    memories JSONB DEFAULT '[]',
    current_mood VARCHAR(50) DEFAULT 'neutral',
    decision_weights JSONB DEFAULT '{}',
    learning_data JSONB DEFAULT '{}',
    last_learning_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(faction_id)
);

-- Create indexes for faction_personalities
CREATE INDEX idx_faction_personalities_faction ON faction_personalities(faction_id);

-- Faction Agendas for long-term goals
CREATE TABLE IF NOT EXISTS faction_agendas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    faction_id UUID NOT NULL REFERENCES factions(id) ON DELETE CASCADE,
    agenda_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority INTEGER DEFAULT 1,
    stages JSONB DEFAULT '[]',
    progress DECIMAL(3,2) DEFAULT 0.0,
    status VARCHAR(20) DEFAULT 'active',
    parameters JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for faction_agendas
CREATE INDEX idx_faction_agendas_faction ON faction_agendas(faction_id);
CREATE INDEX idx_faction_agendas_status ON faction_agendas(status);
CREATE INDEX idx_faction_agendas_priority ON faction_agendas(priority DESC);

-- Procedural Cultures table
CREATE TABLE IF NOT EXISTS procedural_cultures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    language JSONB NOT NULL,
    customs JSONB DEFAULT '[]',
    art_style JSONB NOT NULL,
    belief_system JSONB NOT NULL,
    values JSONB DEFAULT '{}',
    taboos JSONB DEFAULT '[]',
    greetings JSONB DEFAULT '{}',
    architecture JSONB NOT NULL,
    cuisine JSONB DEFAULT '[]',
    music_style JSONB NOT NULL,
    clothing_style JSONB NOT NULL,
    naming_conventions JSONB NOT NULL,
    social_structure JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for procedural_cultures
CREATE INDEX idx_cultures_name ON procedural_cultures(name);
CREATE INDEX idx_cultures_session ON procedural_cultures((metadata->>'session_id'));

-- World Events for tracking significant happenings
CREATE TABLE IF NOT EXISTS world_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    impact JSONB DEFAULT '{}',
    affected_entities JSONB DEFAULT '[]',
    consequences JSONB DEFAULT '[]',
    is_player_visible BOOLEAN DEFAULT TRUE,
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for world_events
CREATE INDEX idx_world_events_session ON world_events(session_id);
CREATE INDEX idx_world_events_type ON world_events(event_type);
CREATE INDEX idx_world_events_occurred ON world_events(occurred_at DESC);
CREATE INDEX idx_world_events_visible ON world_events(is_player_visible);

-- Simulation Logs for tracking world simulation activities
CREATE TABLE IF NOT EXISTS simulation_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    simulation_type VARCHAR(50) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    events_created INTEGER DEFAULT 0,
    details JSONB DEFAULT '{}',
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT
);

-- Create indexes for simulation_logs
CREATE INDEX idx_simulation_logs_session ON simulation_logs(session_id);
CREATE INDEX idx_simulation_logs_type ON simulation_logs(simulation_type);
CREATE INDEX idx_simulation_logs_time ON simulation_logs(start_time DESC);

-- Cultural Interactions table for tracking player influence on cultures
CREATE TABLE IF NOT EXISTS cultural_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    culture_id UUID NOT NULL REFERENCES procedural_cultures(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL, -- Player or NPC ID
    actor_type VARCHAR(20) NOT NULL, -- player, npc
    interaction_type VARCHAR(50) NOT NULL,
    approach VARCHAR(50) NOT NULL,
    impact JSONB DEFAULT '{}',
    response JSONB DEFAULT '{}',
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for cultural_interactions
CREATE INDEX idx_cultural_interactions_culture ON cultural_interactions(culture_id);
CREATE INDEX idx_cultural_interactions_actor ON cultural_interactions(actor_id, actor_type);
CREATE INDEX idx_cultural_interactions_time ON cultural_interactions(occurred_at DESC);

-- Faction Memories for detailed event tracking
CREATE TABLE IF NOT EXISTS faction_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    faction_id UUID NOT NULL REFERENCES factions(id) ON DELETE CASCADE,
    memory_type VARCHAR(50) NOT NULL,
    description TEXT,
    impact DECIMAL(3,2),
    participants JSONB DEFAULT '[]',
    context JSONB DEFAULT '{}',
    decay_rate DECIMAL(3,2) DEFAULT 0.95,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for faction_memories
CREATE INDEX idx_faction_memories_faction ON faction_memories(faction_id);
CREATE INDEX idx_faction_memories_type ON faction_memories(memory_type);
CREATE INDEX idx_faction_memories_impact ON faction_memories(impact DESC);

-- Create triggers for updated_at
CREATE OR REPLACE FUNCTION update_emergent_world_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_world_states_updated_at
    BEFORE UPDATE ON world_states
    FOR EACH ROW
    EXECUTE FUNCTION update_emergent_world_updated_at();