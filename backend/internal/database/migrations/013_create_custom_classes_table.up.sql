CREATE TABLE IF NOT EXISTS custom_classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    hit_die INTEGER NOT NULL CHECK (hit_die IN (6, 8, 10, 12)),
    primary_ability VARCHAR(20) NOT NULL,
    saving_throw_proficiencies TEXT[] NOT NULL,
    skill_proficiencies TEXT[] NOT NULL,
    skill_choices INTEGER NOT NULL DEFAULT 2,
    starting_equipment TEXT NOT NULL,
    armor_proficiencies TEXT[] NOT NULL,
    weapon_proficiencies TEXT[] NOT NULL,
    tool_proficiencies TEXT[],
    
    -- Class features stored as JSONB
    class_features JSONB NOT NULL DEFAULT '[]',
    -- Subclass archetypes
    subclass_name VARCHAR(100),
    subclass_level INTEGER DEFAULT 3,
    subclasses JSONB DEFAULT '[]',
    
    -- Spellcasting info (null for non-spellcasters)
    spellcasting_ability VARCHAR(20),
    spell_list TEXT[],
    spells_known_progression INTEGER[],
    cantrips_known_progression INTEGER[],
    spell_slots_progression JSONB,
    ritual_casting BOOLEAN DEFAULT false,
    spellcasting_focus TEXT,
    
    -- Balance and approval
    balance_score INTEGER CHECK (balance_score >= 1 AND balance_score <= 10),
    power_level VARCHAR(20) CHECK (power_level IN ('balanced', 'flavorful', 'powerful')),
    is_approved BOOLEAN DEFAULT false,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP,
    dm_notes TEXT,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_user_class_name UNIQUE (user_id, name)
);

-- Index for faster queries
CREATE INDEX idx_custom_classes_user_id ON custom_classes(user_id);
CREATE INDEX idx_custom_classes_approved ON custom_classes(is_approved);
CREATE INDEX idx_custom_classes_name ON custom_classes(name);

-- Example class_features JSONB structure:
-- [
--   {
--     "level": 1,
--     "name": "Rage",
--     "description": "In battle, you fight with primal ferocity...",
--     "uses_per_rest": "2 + CON modifier",
--     "rest_type": "long"
--   },
--   {
--     "level": 2,
--     "name": "Reckless Attack",
--     "description": "You can throw aside all concern for defense...",
--     "passive": true
--   }
-- ]

-- Example subclasses JSONB structure:
-- [
--   {
--     "name": "Path of the Berserker",
--     "description": "For some barbarians, rage is a means to an end...",
--     "features": [
--       {
--         "level": 3,
--         "name": "Frenzy",
--         "description": "You can go into a frenzy when you rage..."
--       }
--     ]
--   }
-- ]