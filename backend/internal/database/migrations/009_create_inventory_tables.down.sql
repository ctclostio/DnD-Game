-- Drop indexes
DROP INDEX IF EXISTS idx_character_inventory_attuned;
DROP INDEX IF EXISTS idx_character_inventory_equipped;
DROP INDEX IF EXISTS idx_character_inventory_character;
DROP INDEX IF EXISTS idx_items_rarity;
DROP INDEX IF EXISTS idx_items_type;

-- Remove columns from characters table
ALTER TABLE characters 
DROP COLUMN IF EXISTS carry_capacity,
DROP COLUMN IF EXISTS current_weight,
DROP COLUMN IF EXISTS attunement_slots_used,
DROP COLUMN IF EXISTS attunement_slots_max;

-- Drop tables
DROP TABLE IF EXISTS character_currency;
DROP TABLE IF EXISTS character_inventory;
DROP TABLE IF EXISTS items;