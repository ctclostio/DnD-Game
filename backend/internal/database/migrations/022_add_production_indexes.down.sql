-- Rollback Migration: Remove production performance indexes
-- Purpose: Remove the indexes added for production optimization
-- Date: 2025-06-15

-- ==========================================
-- 1. DROP COMPOSITE INDEXES
-- ==========================================

DROP INDEX IF EXISTS idx_game_participants_user_session;
DROP INDEX IF EXISTS idx_characters_user_id_id;
DROP INDEX IF EXISTS idx_game_participants_session_online;

-- ==========================================
-- 2. DROP PARTIAL INDEXES
-- ==========================================

DROP INDEX IF EXISTS idx_game_sessions_active;
DROP INDEX IF EXISTS idx_refresh_tokens_valid;
DROP INDEX IF EXISTS idx_game_participants_online_only;

-- ==========================================
-- 3. DROP JSONB INDEXES
-- ==========================================

DROP INDEX IF EXISTS idx_characters_attributes_gin;
DROP INDEX IF EXISTS idx_characters_skills_gin;
DROP INDEX IF EXISTS idx_characters_spells_gin;

-- ==========================================
-- 4. DROP TIME-BASED INDEXES
-- ==========================================

DROP INDEX IF EXISTS idx_dice_rolls_session_time;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_game_sessions_updated_at;

-- ==========================================
-- 5. DROP CAMPAIGN MANAGEMENT INDEXES
-- ==========================================

DROP INDEX IF EXISTS idx_timeline_session_event_date;
DROP INDEX IF EXISTS idx_npc_relationships_target;

-- ==========================================
-- 6. DROP INVENTORY INDEXES
-- ==========================================

DROP INDEX IF EXISTS idx_inventory_character_equipped;

-- ==========================================
-- 7. DROP CASE-INSENSITIVE INDEXES
-- ==========================================

DROP INDEX IF EXISTS idx_users_username_lower;
DROP INDEX IF EXISTS idx_characters_name_lower;

-- ==========================================
-- 8. DROP COVERING INDEXES
-- ==========================================

DROP INDEX IF EXISTS idx_users_auth_covering;
DROP INDEX IF EXISTS idx_characters_list_covering;