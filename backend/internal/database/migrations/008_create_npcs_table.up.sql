-- Create NPCs table
CREATE TABLE IF NOT EXISTS npcs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID REFERENCES game_sessions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    size VARCHAR(50) NOT NULL,
    alignment VARCHAR(50),
    armor_class INTEGER NOT NULL DEFAULT 10,
    hit_points INTEGER NOT NULL,
    max_hit_points INTEGER NOT NULL,
    speed JSONB NOT NULL DEFAULT '{"walk": 30}'::jsonb,
    attributes JSONB NOT NULL DEFAULT '{"strength": 10, "dexterity": 10, "constitution": 10, "intelligence": 10, "wisdom": 10, "charisma": 10}'::jsonb,
    saving_throws JSONB DEFAULT '{}',
    skills JSONB DEFAULT '[]',
    damage_resistances TEXT[] DEFAULT '{}',
    damage_immunities TEXT[] DEFAULT '{}',
    condition_immunities TEXT[] DEFAULT '{}',
    senses JSONB DEFAULT '{}',
    languages TEXT[] DEFAULT '{}',
    challenge_rating DECIMAL(4,2) DEFAULT 0,
    experience_points INTEGER DEFAULT 0,
    abilities JSONB DEFAULT '[]',
    actions JSONB DEFAULT '[]',
    legendary_actions INTEGER DEFAULT 0,
    is_template BOOLEAN DEFAULT FALSE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_npcs_game_session_id ON npcs(game_session_id);
CREATE INDEX idx_npcs_created_by ON npcs(created_by);
CREATE INDEX idx_npcs_is_template ON npcs(is_template);
CREATE INDEX idx_npcs_challenge_rating ON npcs(challenge_rating);

-- Create NPC templates table (predefined monsters)
CREATE TABLE IF NOT EXISTS npc_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    source VARCHAR(100) DEFAULT 'custom',
    type VARCHAR(100) NOT NULL,
    size VARCHAR(50) NOT NULL,
    alignment VARCHAR(50),
    armor_class INTEGER NOT NULL DEFAULT 10,
    hit_dice VARCHAR(50) NOT NULL,
    speed JSONB NOT NULL DEFAULT '{"walk": 30}'::jsonb,
    attributes JSONB NOT NULL DEFAULT '{"strength": 10, "dexterity": 10, "constitution": 10, "intelligence": 10, "wisdom": 10, "charisma": 10}'::jsonb,
    saving_throws JSONB DEFAULT '{}',
    skills JSONB DEFAULT '[]',
    damage_resistances TEXT[] DEFAULT '{}',
    damage_immunities TEXT[] DEFAULT '{}',
    condition_immunities TEXT[] DEFAULT '{}',
    senses JSONB DEFAULT '{}',
    languages TEXT[] DEFAULT '{}',
    challenge_rating DECIMAL(4,2) DEFAULT 0,
    abilities JSONB DEFAULT '[]',
    actions JSONB DEFAULT '[]',
    legendary_actions INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for templates
CREATE INDEX idx_npc_templates_name ON npc_templates(name);
CREATE INDEX idx_npc_templates_type ON npc_templates(type);
CREATE INDEX idx_npc_templates_challenge_rating ON npc_templates(challenge_rating);

-- Insert some basic NPC templates
DO $$
DECLARE
    default_walk_speed CONSTANT JSONB := '{"walk": 30}'::jsonb;
    mm_source CONSTANT TEXT := 'MM';
    action_type CONSTANT TEXT := 'action';
    default_attributes CONSTANT JSONB := '{"strength": 10, "dexterity": 10, "constitution": 10, "intelligence": 10, "wisdom": 10, "charisma": 10}'::jsonb;
BEGIN
    INSERT INTO npc_templates (name, source, type, size, alignment, armor_class, hit_dice, speed, attributes, challenge_rating, abilities, actions) VALUES
    ('Goblin', mm_source, 'humanoid', 'small', 'neutral evil', 15, '2d6', default_walk_speed, '{"strength": 8, "dexterity": 14, "constitution": 10, "intelligence": 10, "wisdom": 8, "charisma": 8}', 0.25, 
    '[{"name": "Nimble Escape", "description": "The goblin can take the Disengage or Hide action as a bonus action on each of its turns."}]',
    '[{"name": "Scimitar", "type": "' || action_type || '", "attackBonus": 4, "damage": "1d6+2", "damageType": "slashing"}, {"name": "Shortbow", "type": "' || action_type || '", "attackBonus": 4, "damage": "1d6+2", "damageType": "piercing", "range": "80/320 ft."}]'),

    ('Orc', mm_source, 'humanoid', 'medium', 'chaotic evil', 13, '2d8+6', default_walk_speed, '{"strength": 16, "dexterity": 12, "constitution": 16, "intelligence": 7, "wisdom": 11, "charisma": 10}', 0.5,
    '[{"name": "Aggressive", "description": "As a bonus action, the orc can move up to its speed toward a hostile creature that it can see."}]',
    '[{"name": "Greataxe", "type": "' || action_type || '", "attackBonus": 5, "damage": "1d12+3", "damageType": "slashing"}, {"name": "Javelin", "type": "' || action_type || '", "attackBonus": 5, "damage": "1d6+3", "damageType": "piercing", "range": "30/120 ft."}]'),

    ('Wolf', mm_source, 'beast', 'medium', 'unaligned', 13, '2d8+2', '{"walk": 40}'::jsonb, '{"strength": 12, "dexterity": 15, "constitution": 12, "intelligence": 3, "wisdom": 12, "charisma": 6}', 0.25,
    '[{"name": "Keen Hearing and Smell", "description": "The wolf has advantage on Wisdom (Perception) checks that rely on hearing or smell."}, {"name": "Pack Tactics", "description": "The wolf has advantage on attack rolls against a creature if at least one of the wolf''s allies is within 5 feet of the creature and the ally isn''t incapacitated."}]',
    '[{"name": "Bite", "type": "' || action_type || '", "attackBonus": 4, "damage": "2d4+2", "damageType": "piercing", "description": "If the target is a creature, it must succeed on a DC 11 Strength saving throw or be knocked prone."}]');
END $$;