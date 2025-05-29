CREATE TABLE IF NOT EXISTS custom_races (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    user_prompt TEXT NOT NULL, -- The original user request
    
    -- Ability Score Improvements
    ability_score_increases JSONB NOT NULL DEFAULT '{}', -- e.g., {"strength": 2, "constitution": 1}
    
    -- Basic attributes
    size VARCHAR(20) NOT NULL CHECK (size IN ('Tiny', 'Small', 'Medium', 'Large', 'Huge', 'Gargantuan')),
    speed INTEGER NOT NULL DEFAULT 30,
    
    -- Features and traits
    traits JSONB NOT NULL DEFAULT '[]', -- Array of trait objects with name and description
    languages JSONB NOT NULL DEFAULT '[]', -- Array of language names
    
    -- Special abilities
    darkvision INTEGER DEFAULT 0, -- Range in feet, 0 means no darkvision
    resistances JSONB DEFAULT '[]', -- Array of damage type resistances
    immunities JSONB DEFAULT '[]', -- Array of damage type immunities
    
    -- Proficiencies
    skill_proficiencies JSONB DEFAULT '[]', -- Array of skill names
    tool_proficiencies JSONB DEFAULT '[]', -- Array of tool names
    weapon_proficiencies JSONB DEFAULT '[]', -- Array of weapon names
    armor_proficiencies JSONB DEFAULT '[]', -- Array of armor types
    
    -- Metadata
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    approval_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (approval_status IN ('pending', 'approved', 'rejected', 'revision_needed')),
    approval_notes TEXT,
    balance_score INTEGER, -- AI-generated balance score (1-10)
    
    -- Usage tracking
    times_used INTEGER DEFAULT 0,
    is_public BOOLEAN DEFAULT false, -- Can other players use this race?
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_custom_races_created_by ON custom_races(created_by);
CREATE INDEX idx_custom_races_approval_status ON custom_races(approval_status);
CREATE INDEX idx_custom_races_is_public ON custom_races(is_public);
CREATE INDEX idx_custom_races_name ON custom_races(name);

-- Update characters table to support custom races
ALTER TABLE characters 
ADD COLUMN custom_race_id UUID REFERENCES custom_races(id) ON DELETE SET NULL,
ADD CONSTRAINT check_race_or_custom_race CHECK (
    (race IS NOT NULL AND custom_race_id IS NULL) OR 
    (race IS NULL AND custom_race_id IS NOT NULL)
);