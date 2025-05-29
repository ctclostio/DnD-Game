-- Rollback character field updates
ALTER TABLE characters 
DROP COLUMN IF EXISTS subrace,
DROP COLUMN IF EXISTS subclass,
DROP COLUMN IF EXISTS background,
DROP COLUMN IF EXISTS alignment,
DROP COLUMN IF EXISTS temp_hit_points,
DROP COLUMN IF EXISTS hit_dice,
DROP COLUMN IF EXISTS initiative,
DROP COLUMN IF EXISTS proficiency_bonus,
DROP COLUMN IF EXISTS saving_throws,
DROP COLUMN IF EXISTS proficiencies,
DROP COLUMN IF EXISTS features,
DROP COLUMN IF EXISTS resources;

-- Restore original spells column
ALTER TABLE characters 
DROP COLUMN IF EXISTS spells;

ALTER TABLE characters
ADD COLUMN IF NOT EXISTS spells JSONB DEFAULT '[]';

-- Drop indexes
DROP INDEX IF EXISTS idx_characters_user_id;
DROP INDEX IF EXISTS idx_characters_name;