-- Drop indexes
DROP INDEX IF EXISTS idx_encounter_templates_levels;
DROP INDEX IF EXISTS idx_encounter_templates_public;
DROP INDEX IF EXISTS idx_encounter_events_encounter;
DROP INDEX IF EXISTS idx_encounter_objectives_encounter;
DROP INDEX IF EXISTS idx_encounter_enemies_encounter;
DROP INDEX IF EXISTS idx_encounters_status;
DROP INDEX IF EXISTS idx_encounters_game_session;

-- Drop tables
DROP TABLE IF EXISTS encounter_templates;
DROP TABLE IF EXISTS encounter_events;
DROP TABLE IF EXISTS encounter_objectives;
DROP TABLE IF EXISTS encounter_enemies;
DROP TABLE IF EXISTS encounters;