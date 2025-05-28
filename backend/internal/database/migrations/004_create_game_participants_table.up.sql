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

-- Ensure character belongs to the user
ALTER TABLE game_participants
ADD CONSTRAINT check_character_owner
CHECK (
    character_id IS NULL OR
    EXISTS (
        SELECT 1 FROM characters
        WHERE characters.id = character_id
        AND characters.user_id = game_participants.user_id
    )
);