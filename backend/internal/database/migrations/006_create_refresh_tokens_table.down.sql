-- Drop the role column from users table
ALTER TABLE users DROP COLUMN IF EXISTS role;

-- Drop indexes
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;

-- Drop the refresh_tokens table
DROP TABLE IF EXISTS refresh_tokens;