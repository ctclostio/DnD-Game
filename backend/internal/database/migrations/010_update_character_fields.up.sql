-- Add new character fields for complete D&D 5e support
ALTER TABLE characters 
ADD COLUMN IF NOT EXISTS subrace VARCHAR(50),
ADD COLUMN IF NOT EXISTS subclass VARCHAR(50),
ADD COLUMN IF NOT EXISTS background VARCHAR(50),
ADD COLUMN IF NOT EXISTS alignment VARCHAR(50),
ADD COLUMN IF NOT EXISTS temp_hit_points INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS hit_dice VARCHAR(20),
ADD COLUMN IF NOT EXISTS initiative INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS proficiency_bonus INTEGER DEFAULT 2,
ADD COLUMN IF NOT EXISTS saving_throws JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS proficiencies JSONB DEFAULT '{"armor":[],"weapons":[],"tools":[],"languages":[]}',
ADD COLUMN IF NOT EXISTS features JSONB DEFAULT '[]',
ADD COLUMN IF NOT EXISTS resources JSONB DEFAULT '{}';

-- Update spells column to support new spell data structure
ALTER TABLE characters 
DROP COLUMN IF EXISTS spells;

ALTER TABLE characters
ADD COLUMN IF NOT EXISTS spells JSONB DEFAULT '{"spellcastingAbility":"","spellSaveDC":0,"spellAttackBonus":0,"spellSlots":[],"spellsKnown":[],"cantripsKnown":0}';

-- Create index on user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_characters_user_id ON characters(user_id);

-- Create index on name for search functionality
CREATE INDEX IF NOT EXISTS idx_characters_name ON characters(name);