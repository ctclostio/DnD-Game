-- Migration: Add production performance indexes
-- Purpose: Optimize query performance for common access patterns in production
-- Date: 2025-06-15

-- ==========================================
-- 1. COMPOSITE INDEXES FOR JOIN OPERATIONS
-- ==========================================

-- For finding user's active game sessions (as player or DM)
CREATE INDEX IF NOT EXISTS idx_game_participants_user_session 
ON game_participants(user_id, game_session_id);

-- For character ownership validation
CREATE INDEX IF NOT EXISTS idx_characters_user_id_id 
ON characters(user_id, id);

-- For session participant listings with online status
CREATE INDEX IF NOT EXISTS idx_game_participants_session_online 
ON game_participants(game_session_id, is_online);

-- ==========================================
-- 2. PARTIAL INDEXES FOR FILTERED QUERIES
-- ==========================================

-- For active sessions only (reduces index size)
CREATE INDEX IF NOT EXISTS idx_game_sessions_active 
ON game_sessions(dm_user_id, created_at DESC) 
WHERE status = 'active';

-- For valid (non-expired, non-revoked) refresh tokens
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_valid 
ON refresh_tokens(user_id, expires_at) 
WHERE revoked_at IS NULL;

-- For online participants only
CREATE INDEX IF NOT EXISTS idx_game_participants_online_only 
ON game_participants(game_session_id) 
WHERE is_online = true;

-- ==========================================
-- 3. JSONB INDEXES FOR COMPLEX QUERIES
-- ==========================================

-- For character attribute searches (PostgreSQL only)
-- Note: These will be skipped in SQLite environments
DO $$ 
BEGIN
    IF current_database() != 'sqlite' THEN
        -- For character attribute searches (e.g., finding high-strength characters)
        CREATE INDEX IF NOT EXISTS idx_characters_attributes_gin 
        ON characters USING GIN (attributes);
        
        -- For skill searches
        CREATE INDEX IF NOT EXISTS idx_characters_skills_gin 
        ON characters USING GIN (skills);
        
        -- For spell searches
        CREATE INDEX IF NOT EXISTS idx_characters_spells_gin 
        ON characters USING GIN (spells);
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        -- Skip if not PostgreSQL
        NULL;
END $$;

-- ==========================================
-- 4. TIME-BASED INDEXES FOR ANALYTICS
-- ==========================================

-- For dice roll analytics by session and time
CREATE INDEX IF NOT EXISTS idx_dice_rolls_session_time 
ON dice_rolls(game_session_id, timestamp DESC);

-- For user activity tracking
CREATE INDEX IF NOT EXISTS idx_users_created_at 
ON users(created_at DESC);

-- For session history queries
CREATE INDEX IF NOT EXISTS idx_game_sessions_updated_at 
ON game_sessions(updated_at DESC);

-- ==========================================
-- 5. CAMPAIGN MANAGEMENT INDEXES
-- ==========================================

-- For efficient timeline queries (if campaign_timeline exists)
CREATE INDEX IF NOT EXISTS idx_timeline_session_event_date 
ON campaign_timeline(game_session_id, event_date DESC);

-- For NPC relationship lookups (if npc_relationships exists)
CREATE INDEX IF NOT EXISTS idx_npc_relationships_target 
ON npc_relationships(target_id, target_type);

-- ==========================================
-- 6. INVENTORY AND ITEM QUERIES
-- ==========================================

-- For character inventory lookups (if character_inventory exists)
CREATE INDEX IF NOT EXISTS idx_inventory_character_equipped 
ON character_inventory(character_id, equipped);

-- ==========================================
-- 7. CASE-INSENSITIVE SEARCH INDEXES
-- ==========================================

-- For case-insensitive username searches
CREATE INDEX IF NOT EXISTS idx_users_username_lower 
ON users(LOWER(username));

-- For case-insensitive character name searches
CREATE INDEX IF NOT EXISTS idx_characters_name_lower 
ON characters(LOWER(name));

-- ==========================================
-- 8. COVERING INDEXES FOR COMMON QUERIES
-- ==========================================

-- For user authentication queries (reduces table lookups)
CREATE INDEX IF NOT EXISTS idx_users_auth_covering 
ON users(username, id, password_hash);

-- For character list queries
CREATE INDEX IF NOT EXISTS idx_characters_list_covering 
ON characters(user_id, name, level, class, race);

-- ==========================================
-- ANALYZE TABLES FOR QUERY PLANNER
-- ==========================================

-- Update table statistics for query planner optimization
ANALYZE users;
ANALYZE characters;
ANALYZE game_sessions;
ANALYZE game_participants;
ANALYZE dice_rolls;
ANALYZE refresh_tokens;

-- Note: Additional ANALYZE commands should be added for other tables as needed