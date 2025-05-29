-- Drop the index
DROP INDEX IF EXISTS idx_characters_custom_class_id;

-- Remove custom_class_id column from characters
ALTER TABLE characters
DROP COLUMN IF EXISTS custom_class_id;