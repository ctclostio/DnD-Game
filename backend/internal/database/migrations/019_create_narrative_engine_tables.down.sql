-- Drop triggers
DROP TRIGGER IF EXISTS update_narrative_relationships_updated_at ON narrative_relationships;
DROP TRIGGER IF EXISTS update_narrative_threads_updated_at ON narrative_threads;
DROP TRIGGER IF EXISTS update_narrative_profiles_updated_at ON narrative_profiles;

-- Drop function
DROP FUNCTION IF EXISTS update_narrative_updated_at();

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS narrative_relationships;
DROP TABLE IF EXISTS narrative_threads;
DROP TABLE IF EXISTS player_actions;
DROP TABLE IF EXISTS narrative_memories;
DROP TABLE IF EXISTS perspective_narratives;
DROP TABLE IF EXISTS world_events;
DROP TABLE IF EXISTS consequence_events;
DROP TABLE IF EXISTS personalized_narratives;
DROP TABLE IF EXISTS backstory_elements;
DROP TABLE IF EXISTS narrative_profiles;