-- Drop indexes
DROP INDEX IF EXISTS idx_dm_assistant_history_user;
DROP INDEX IF EXISTS idx_dm_assistant_history_session;
DROP INDEX IF EXISTS idx_ai_environmental_hazards_location;
DROP INDEX IF EXISTS idx_ai_story_elements_unused;
DROP INDEX IF EXISTS idx_ai_story_elements_session;
DROP INDEX IF EXISTS idx_ai_narrations_type;
DROP INDEX IF EXISTS idx_ai_narrations_session;
DROP INDEX IF EXISTS idx_ai_locations_type;
DROP INDEX IF EXISTS idx_ai_locations_session;
DROP INDEX IF EXISTS idx_ai_npcs_recurring;
DROP INDEX IF EXISTS idx_ai_npcs_session;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS dm_assistant_history;
DROP TABLE IF EXISTS ai_environmental_hazards;
DROP TABLE IF EXISTS ai_story_elements;
DROP TABLE IF EXISTS ai_narrations;
DROP TABLE IF EXISTS ai_locations;
DROP TABLE IF EXISTS ai_npcs;