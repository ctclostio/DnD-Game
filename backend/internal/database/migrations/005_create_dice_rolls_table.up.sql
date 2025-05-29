CREATE TABLE IF NOT EXISTS dice_rolls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    dice_type VARCHAR(10) NOT NULL CHECK (dice_type IN ('d4', 'd6', 'd8', 'd10', 'd12', 'd20', 'd100')),
    count INTEGER NOT NULL DEFAULT 1 CHECK (count > 0),
    modifier INTEGER DEFAULT 0,
    results INTEGER[] NOT NULL,
    total INTEGER NOT NULL,
    purpose VARCHAR(50),
    roll_notation VARCHAR(50) NOT NULL, -- e.g., "2d20+5"
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_dice_rolls_game_session_id ON dice_rolls(game_session_id);
CREATE INDEX idx_dice_rolls_user_id ON dice_rolls(user_id);
CREATE INDEX idx_dice_rolls_timestamp ON dice_rolls(timestamp);
CREATE INDEX idx_dice_rolls_purpose ON dice_rolls(purpose);

-- Note: User participation validation should be done at the application level
-- PostgreSQL doesn't support subqueries in CHECK constraints