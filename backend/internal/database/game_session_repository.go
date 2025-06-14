package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
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
	// Generate a unique ID if not provided
	if session.ID == "" {
		session.ID = fmt.Sprintf("session-%s-%d", session.Code, time.Now().UnixNano())
	}

	// Set timestamps
	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now

	// Convert state to JSON string for storage
	stateJSON := constants.EmptyJSON
	if session.State != nil && len(session.State) > 0 {
		// In production, handle JSON marshaling properly
		stateJSON = constants.EmptyJSON
	}

	query := `
		INSERT INTO game_sessions (
			id, name, description, dm_user_id, code, is_active, 
			status, session_state, max_players, is_public, 
			requires_invite, allowed_character_level, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query),
		session.ID, session.Name, session.Description, session.DMID,
		session.Code, session.IsActive, string(session.Status),
		stateJSON, session.MaxPlayers, session.IsPublic,
		session.RequiresInvite, session.AllowedCharacterLevel,
		session.CreatedAt, session.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create game session: %w", err)
	}

	return nil
}

// GetByID retrieves a game session by ID with participants
func (r *gameSessionRepository) GetByID(ctx context.Context, id string) (*models.GameSession, error) {
	var session models.GameSession
	var stateJSON string

	query := `
		SELECT id, name, description, dm_user_id, code, is_active,
		       status, session_state, max_players, is_public, requires_invite,
		       allowed_character_level, created_at, updated_at, started_at, ended_at
		FROM game_sessions
		WHERE id = ?`

	err := r.db.QueryRowContext(ctx, r.db.Rebind(query), id).Scan(
		&session.ID, &session.Name, &session.Description, &session.DMID,
		&session.Code, &session.IsActive, &session.Status, &stateJSON,
		&session.MaxPlayers, &session.IsPublic, &session.RequiresInvite,
		&session.AllowedCharacterLevel, &session.CreatedAt, &session.UpdatedAt,
		&session.StartedAt, &session.EndedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game session not found")
		}
		return nil, fmt.Errorf("failed to get game session by id: %w", err)
	}

	// Parse state JSON (in production, handle this properly)
	if stateJSON != "" && stateJSON != "{}" {
		// Parse JSON into session.State
		session.State = make(map[string]interface{})
	} else {
		session.State = make(map[string]interface{})
	}

	return &session, nil
}

// GetByDMUserID retrieves all game sessions for a DM
func (r *gameSessionRepository) GetByDMUserID(ctx context.Context, dmUserID string) ([]*models.GameSession, error) {
	sessions := make([]*models.GameSession, 0, 20)
	query := `
		SELECT id, name, description, dm_user_id, code, is_active,
		       status, session_state, max_players, is_public, requires_invite,
		       allowed_character_level, created_at, updated_at, started_at, ended_at
		FROM game_sessions
		WHERE dm_user_id = ?
		ORDER BY created_at DESC`

	// For SQLite compatibility, we need to handle this differently
	rows, err := r.db.QueryContext(ctx, r.db.Rebind(query), dmUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game sessions by dm user id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var session models.GameSession
		var stateJSON string

		err := rows.Scan(
			&session.ID, &session.Name, &session.Description, &session.DMID,
			&session.Code, &session.IsActive, &session.Status, &stateJSON,
			&session.MaxPlayers, &session.IsPublic, &session.RequiresInvite,
			&session.AllowedCharacterLevel, &session.CreatedAt, &session.UpdatedAt,
			&session.StartedAt, &session.EndedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game session: %w", err)
		}

		// Initialize state
		session.State = make(map[string]interface{})

		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetByParticipantUserID retrieves all game sessions where user is a participant
func (r *gameSessionRepository) GetByParticipantUserID(ctx context.Context, userID string) ([]*models.GameSession, error) {
	sessions := make([]*models.GameSession, 0, 20)
	query := `
		SELECT DISTINCT gs.id, gs.name, gs.description, gs.dm_user_id, gs.code, gs.is_active,
		       gs.status, gs.session_state, gs.max_players, gs.is_public, gs.requires_invite,
		       gs.allowed_character_level, gs.created_at, gs.updated_at, gs.started_at, gs.ended_at
		FROM game_sessions gs
		LEFT JOIN game_participants gp ON gs.id = gp.session_id
		WHERE gp.user_id = ? OR gs.dm_user_id = ?
		ORDER BY gs.created_at DESC`

	rows, err := r.db.QueryContext(ctx, r.db.Rebind(query), userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game sessions by participant user id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var session models.GameSession
		var stateJSON string

		err := rows.Scan(
			&session.ID, &session.Name, &session.Description, &session.DMID,
			&session.Code, &session.IsActive, &session.Status, &stateJSON,
			&session.MaxPlayers, &session.IsPublic, &session.RequiresInvite,
			&session.AllowedCharacterLevel, &session.CreatedAt, &session.UpdatedAt,
			&session.StartedAt, &session.EndedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game session: %w", err)
		}

		// Initialize state
		session.State = make(map[string]interface{})

		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// Update updates a game session
func (r *gameSessionRepository) Update(ctx context.Context, session *models.GameSession) error {
	query := `
		UPDATE game_sessions
		SET name = ?, description = ?, status = ?, is_active = ?, max_players = ?, 
		    is_public = ?, requires_invite = ?, allowed_character_level = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := r.db.ExecContextRebind(ctx, query,
		session.Name, session.Description, session.Status, session.IsActive,
		session.MaxPlayers, session.IsPublic, session.RequiresInvite,
		session.AllowedCharacterLevel, session.ID)
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
	query := `DELETE FROM game_sessions WHERE id = ?`

	result, err := r.db.ExecContextRebind(ctx, query, id)
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
	sessions := make([]*models.GameSession, 0, limit)
	query := `
		SELECT id, name, description, dm_user_id, code, is_active,
		       status, session_state, max_players, is_public, requires_invite,
		       allowed_character_level, created_at, updated_at, started_at, ended_at
		FROM game_sessions
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, r.db.Rebind(query), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list game sessions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var session models.GameSession
		var stateJSON string

		err := rows.Scan(
			&session.ID, &session.Name, &session.Description, &session.DMID,
			&session.Code, &session.IsActive, &session.Status, &stateJSON,
			&session.MaxPlayers, &session.IsPublic, &session.RequiresInvite,
			&session.AllowedCharacterLevel, &session.CreatedAt, &session.UpdatedAt,
			&session.StartedAt, &session.EndedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game session: %w", err)
		}

		// Initialize state
		session.State = make(map[string]interface{})

		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// AddParticipant adds a participant to a game session
func (r *gameSessionRepository) AddParticipant(ctx context.Context, sessionID, userID string, characterID *string) error {
	// Generate an ID for the participant
	participantID := fmt.Sprintf("participant-%s-%s-%d", userID, sessionID, time.Now().UnixNano())

	query := `
		INSERT INTO game_participants (id, session_id, user_id, character_id, is_online)
		VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), participantID, sessionID, userID, characterID, false)
	if err != nil {
		// Check if it's a duplicate entry error
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("user already in session")
		}
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

// RemoveParticipant removes a participant from a game session
func (r *gameSessionRepository) RemoveParticipant(ctx context.Context, sessionID, userID string) error {
	query := `
		DELETE FROM game_participants
		WHERE session_id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, r.db.Rebind(query), sessionID, userID)
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
			gp.session_id, gp.user_id, gp.character_id, gp.is_online, gp.joined_at
		FROM game_participants gp
		WHERE gp.session_id = ?
		ORDER BY gp.joined_at`

	rows, err := r.db.QueryContext(ctx, r.db.Rebind(query), sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	defer rows.Close()

	var participants []*models.GameParticipant
	for rows.Next() {
		var p models.GameParticipant

		err := rows.Scan(
			&p.SessionID, &p.UserID, &p.CharacterID, &p.IsOnline, &p.JoinedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		participants = append(participants, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating participants: %w", err)
	}

	return participants, nil
}

// UpdateParticipantOnlineStatus updates the online status of a participant
func (r *gameSessionRepository) UpdateParticipantOnlineStatus(ctx context.Context, sessionID, userID string, isOnline bool) error {
	query := `
		UPDATE game_participants
		SET is_online = ?
		WHERE session_id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, r.db.Rebind(query), isOnline, sessionID, userID)
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
