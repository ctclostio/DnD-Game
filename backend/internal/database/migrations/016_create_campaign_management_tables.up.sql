-- Campaign Management Tables

-- Story Arcs table
CREATE TABLE IF NOT EXISTS story_arcs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    arc_type VARCHAR(50) NOT NULL, -- main_quest, side_quest, character_arc, etc.
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, completed, abandoned, foreshadowed
    parent_arc_id UUID REFERENCES story_arcs(id),
    importance_level INTEGER DEFAULT 5, -- 1-10 scale
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    metadata JSONB DEFAULT '{}'::jsonb -- For AI-generated elements, connections, etc.
);

-- Session Memories table
CREATE TABLE IF NOT EXISTS session_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    session_number INTEGER NOT NULL,
    session_date TIMESTAMP NOT NULL,
    recap_summary TEXT, -- AI-generated summary
    key_events JSONB DEFAULT '[]'::jsonb, -- Array of important events
    npcs_encountered JSONB DEFAULT '[]'::jsonb, -- NPCs met with context
    decisions_made JSONB DEFAULT '[]'::jsonb, -- Player choices and outcomes
    items_acquired JSONB DEFAULT '[]'::jsonb, -- Loot and rewards
    locations_visited JSONB DEFAULT '[]'::jsonb, -- Places explored
    combat_encounters JSONB DEFAULT '[]'::jsonb, -- Battle summaries
    plot_developments JSONB DEFAULT '[]'::jsonb, -- Story progression
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Plot Threads table
CREATE TABLE IF NOT EXISTS plot_threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    story_arc_id UUID REFERENCES story_arcs(id) ON DELETE CASCADE,
    thread_type VARCHAR(50) NOT NULL, -- mystery, conflict, relationship, prophecy, etc.
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, resolved, dormant, abandoned
    tension_level INTEGER DEFAULT 5, -- 1-10, how urgent/important
    introduced_session INTEGER,
    resolved_session INTEGER,
    related_npcs JSONB DEFAULT '[]'::jsonb, -- NPC IDs involved
    related_locations JSONB DEFAULT '[]'::jsonb, -- Location IDs involved
    foreshadowing_hints JSONB DEFAULT '[]'::jsonb, -- Hints to drop
    resolution_conditions JSONB DEFAULT '{}'::jsonb, -- What needs to happen
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Foreshadowing Elements table
CREATE TABLE IF NOT EXISTS foreshadowing_elements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    plot_thread_id UUID REFERENCES plot_threads(id) ON DELETE CASCADE,
    story_arc_id UUID REFERENCES story_arcs(id) ON DELETE CASCADE,
    element_type VARCHAR(50) NOT NULL, -- prophecy, rumor, symbol, dream, omen, etc.
    content TEXT NOT NULL,
    subtlety_level INTEGER DEFAULT 5, -- 1-10, how obvious
    revealed BOOLEAN DEFAULT FALSE,
    revealed_session INTEGER,
    placement_suggestions JSONB DEFAULT '[]'::jsonb, -- Where/when to introduce
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Campaign Timeline table (for tracking chronological events)
CREATE TABLE IF NOT EXISTS campaign_timeline (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    session_memory_id UUID REFERENCES session_memories(id) ON DELETE CASCADE,
    event_date TIMESTAMP NOT NULL, -- In-game date/time
    real_session_date TIMESTAMP NOT NULL, -- When it was played
    event_type VARCHAR(50) NOT NULL, -- combat, roleplay, discovery, decision, etc.
    event_title VARCHAR(255) NOT NULL,
    event_description TEXT,
    impact_level INTEGER DEFAULT 5, -- 1-10, how significant
    related_arcs JSONB DEFAULT '[]'::jsonb, -- Story arc IDs affected
    related_threads JSONB DEFAULT '[]'::jsonb, -- Plot thread IDs affected
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- NPC Relationships table (track how NPCs relate to party and each other)
CREATE TABLE IF NOT EXISTS npc_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    npc_id UUID NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    target_type VARCHAR(50) NOT NULL, -- character, npc, faction
    target_id UUID NOT NULL, -- ID of character, NPC, or faction
    relationship_type VARCHAR(50) NOT NULL, -- ally, enemy, neutral, rival, etc.
    relationship_score INTEGER DEFAULT 0, -- -100 to 100
    last_interaction_session INTEGER,
    interaction_history JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(game_session_id, npc_id, target_id)
);

-- Create indexes for better performance
CREATE INDEX idx_story_arcs_session ON story_arcs(game_session_id);
CREATE INDEX idx_story_arcs_parent ON story_arcs(parent_arc_id);
CREATE INDEX idx_story_arcs_status ON story_arcs(status);
CREATE INDEX idx_session_memories_session ON session_memories(game_session_id);
CREATE INDEX idx_session_memories_date ON session_memories(session_date);
CREATE INDEX idx_plot_threads_session ON plot_threads(game_session_id);
CREATE INDEX idx_plot_threads_arc ON plot_threads(story_arc_id);
CREATE INDEX idx_plot_threads_status ON plot_threads(status);
CREATE INDEX idx_foreshadowing_session ON foreshadowing_elements(game_session_id);
CREATE INDEX idx_foreshadowing_revealed ON foreshadowing_elements(revealed);
CREATE INDEX idx_timeline_session ON campaign_timeline(game_session_id);
CREATE INDEX idx_timeline_dates ON campaign_timeline(event_date, real_session_date);
CREATE INDEX idx_npc_relationships_session ON npc_relationships(game_session_id);
CREATE INDEX idx_npc_relationships_npc ON npc_relationships(npc_id);