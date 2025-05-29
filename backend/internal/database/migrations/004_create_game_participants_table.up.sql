CREATE TABLE IF NOT EXISTS game_participants (
    game_session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE SET NULL,
    is_online BOOLEAN DEFAULT false,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (game_session_id, user_id)
);

-- Create indexes
CREATE INDEX idx_game_participants_user_id ON game_participants(user_id);
CREATE INDEX idx_game_participants_character_id ON game_participants(character_id);
CREATE INDEX idx_game_participants_is_online ON game_participants(is_online);

-- Note: Character ownership validation should be done at the application level
-- PostgreSQL doesn't support subqueries in CHECK constraints