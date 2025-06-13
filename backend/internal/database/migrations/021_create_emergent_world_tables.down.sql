-- Drop triggers
DROP TRIGGER IF EXISTS update_world_states_updated_at ON world_states;
DROP FUNCTION IF EXISTS update_emergent_world_updated_at();

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS faction_memories;
DROP TABLE IF EXISTS cultural_interactions;
DROP TABLE IF EXISTS simulation_logs;
DROP TABLE IF EXISTS emergent_world_events;
DROP TABLE IF EXISTS procedural_cultures;
DROP TABLE IF EXISTS faction_agendas;
DROP TABLE IF EXISTS faction_personalities;
DROP TABLE IF EXISTS npc_schedules;
DROP TABLE IF EXISTS npc_goals;
DROP TABLE IF EXISTS world_states;