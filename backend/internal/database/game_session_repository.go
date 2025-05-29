package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/your-username/dnd-game/backend/internal/models"
)

// gameSessionRepository implements GameSessionRepository interface
type gameSessionRepository struct {
	db *DB
}

// NewGameSessionRepository creates a new game session repository
func NewGameSessionRepository(db *DB) GameSessionRepository {
	return &gameSessionRepository{db: db}
}

// Create creates a new game session
func (r *gameSessionRepository) Create(ctx context.Context, session *models.GameSession) error {
	query := `
		INSERT INTO game_sessions (name, dm_user_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query, session.Name, session.DMID, session.Status).
		Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create game session: %w", err)
	}

	return nil
}

// GetByID retrieves a game session by ID with participants
func (r *gameSessionRepository) GetByID(ctx context.Context, id string) (*models.GameSession, error) {
	var session models.GameSession
	query := `
		SELECT id, name, dm_user_id, status, created_at, updated_at
		FROM game_sessions
		WHERE id = $1`

	err := r.db.GetContext(ctx, &session, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game session not found")
		}
		return nil, fmt.Errorf("failed to get game session by id: %w", err)
	}

	// TODO: Load participants if needed
	// Get participants
	// participants, err := r.GetParticipants(ctx, id)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get participants: %w", err)
	// }

	return &session, nil
}

// GetByDMUserID retrieves all game sessions for a DM
func (r *gameSessionRepository) GetByDMUserID(ctx context.Context, dmUserID string) ([]*models.GameSession, error) {
	var sessions []*models.GameSession
	query := `
		SELECT id, name, dm_user_id, status, created_at, updated_at
		FROM game_sessions
		WHERE dm_user_id = $1
		ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &sessions, query, dmUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game sessions by dm user id: %w", err)
	}

	// TODO: Load participants if needed
	// Get participants for each session
	// for _, session := range sessions {
	// 	participants, err := r.GetParticipants(ctx, session.ID)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to get participants for session %s: %w", session.ID, err)
	// 	}
	// }

	return sessions, nil
}

// GetByParticipantUserID retrieves all game sessions where user is a participant
func (r *gameSessionRepository) GetByParticipantUserID(ctx context.Context, userID string) ([]*models.GameSession, error) {
	var sessions []*models.GameSession
	query := `
		SELECT DISTINCT gs.id, gs.name, gs.dm_user_id, gs.status, gs.created_at, gs.updated_at
		FROM game_sessions gs
		JOIN game_participants gp ON gs.id = gp.game_session_id
		WHERE gp.user_id = $1 OR gs.dm_user_id = $1
		ORDER BY gs.created_at DESC`

	err := r.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game sessions by participant user id: %w", err)
	}

	// TODO: Load participants if needed
	// Get participants for each session
	// for _, session := range sessions {
	// 	participants, err := r.GetParticipants(ctx, session.ID)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to get participants for session %s: %w", session.ID, err)
	// 	}
	// }

	return sessions, nil
}

// Update updates a game session
func (r *gameSessionRepository) Update(ctx context.Context, session *models.GameSession) error {
	query := `
		UPDATE game_sessions
		SET name = $2, status = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, session.ID, session.Name, session.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("game session not found")
		}
		return fmt.Errorf("failed to update game session: %w", err)
	}

	return nil
}

// Delete deletes a game session
func (r *gameSessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM game_sessions WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete game session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("game session not found")
	}

	return nil
}

// List retrieves a paginated list of game sessions
func (r *gameSessionRepository) List(ctx context.Context, offset, limit int) ([]*models.GameSession, error) {
	var sessions []*models.GameSession
	query := `
		SELECT id, name, dm_user_id, status, created_at, updated_at
		FROM game_sessions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	err := r.db.SelectContext(ctx, &sessions, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list game sessions: %w", err)
	}

	return sessions, nil
}

// AddParticipant adds a participant to a game session
func (r *gameSessionRepository) AddParticipant(ctx context.Context, sessionID, userID string, characterID *string) error {
	query := `
		INSERT INTO game_participants (game_session_id, user_id, character_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (game_session_id, user_id) DO UPDATE
		SET character_id = EXCLUDED.character_id`

	_, err := r.db.ExecContext(ctx, query, sessionID, userID, characterID)
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

// RemoveParticipant removes a participant from a game session
func (r *gameSessionRepository) RemoveParticipant(ctx context.Context, sessionID, userID string) error {
	query := `
		DELETE FROM game_participants
		WHERE game_session_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

// GetParticipants retrieves all participants for a game session
func (r *gameSessionRepository) GetParticipants(ctx context.Context, sessionID string) ([]*models.GameParticipant, error) {
	query := `
		SELECT 
			gp.game_session_id, gp.user_id, gp.character_id, gp.is_online, gp.joined_at,
			u.id, u.username, u.email, u.created_at, u.updated_at
		FROM game_participants gp
		JOIN users u ON gp.user_id = u.id
		WHERE gp.game_session_id = $1
		ORDER BY gp.joined_at`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	defer rows.Close()

	var participants []*models.GameParticipant
	for rows.Next() {
		var p models.GameParticipant
		var u models.User
		
		err := rows.Scan(
			&p.SessionID, &p.UserID, &p.CharacterID, &p.IsOnline, &p.JoinedAt,
			&u.ID, &u.Username, &u.Email, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		p.User = &u
		participants = append(participants, &p)
	}

	return participants, nil
}

// UpdateParticipantOnlineStatus updates the online status of a participant
func (r *gameSessionRepository) UpdateParticipantOnlineStatus(ctx context.Context, sessionID, userID string, isOnline bool) error {
	query := `
		UPDATE game_participants
		SET is_online = $3
		WHERE game_session_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, sessionID, userID, isOnline)
	if err != nil {
		return fmt.Errorf("failed to update participant online status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}