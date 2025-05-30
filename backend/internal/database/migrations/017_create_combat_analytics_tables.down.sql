-- Drop Combat Analytics and Automation Tables
DROP TRIGGER IF EXISTS update_combat_analytics_timestamp_trigger ON combat_analytics;
DROP FUNCTION IF EXISTS update_combat_analytics_timestamp();

DROP TABLE IF EXISTS combat_action_log CASCADE;
DROP TABLE IF EXISTS smart_initiative_rules CASCADE;
DROP TABLE IF EXISTS battle_maps CASCADE;
DROP TABLE IF EXISTS auto_combat_resolutions CASCADE;
DROP TABLE IF EXISTS combatant_analytics CASCADE;
DROP TABLE IF EXISTS combat_analytics CASCADE;