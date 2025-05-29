-- Remove custom race support from characters table
ALTER TABLE characters DROP CONSTRAINT IF EXISTS check_race_or_custom_race;
ALTER TABLE characters DROP COLUMN IF EXISTS custom_race_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_custom_races_created_by;
DROP INDEX IF EXISTS idx_custom_races_approval_status;
DROP INDEX IF EXISTS idx_custom_races_is_public;
DROP INDEX IF EXISTS idx_custom_races_name;

-- Drop the custom races table
DROP TABLE IF EXISTS custom_races;