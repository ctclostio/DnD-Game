package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// diceRollRepository implements DiceRollRepository interface
type diceRollRepository struct {
	db *DB
}

// NewDiceRollRepository creates a new dice roll repository
func NewDiceRollRepository(db *DB) DiceRollRepository {
	return &diceRollRepository{db: db}
}

// scanDiceRoll is a helper to scan a single DiceRoll row
func (r *diceRollRepository) scanDiceRoll(row RowScanner) (*models.DiceRoll, error) {
	var roll models.DiceRoll
	err := row.Scan(
		&roll.ID, &roll.GameSessionID, &roll.UserID, &roll.DiceType,
		&roll.Count, &roll.Modifier, pq.Array(&roll.Results),
		&roll.Total, &roll.Purpose, &roll.RollNotation, &roll.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to scan dice roll: %w", err)
	}
	return &roll, nil
}

// Create creates a new dice roll
func (r *diceRollRepository) Create(ctx context.Context, roll *models.DiceRoll) error {
	query := `
		INSERT INTO dice_rolls (
			game_session_id, user_id, dice_type, count, modifier,
			results, total, purpose, roll_notation
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, timestamp`

	err := r.db.QueryRowContextRebind(ctx, query,
		roll.GameSessionID, roll.UserID, roll.DiceType, roll.Count,
		roll.Modifier, pq.Array(roll.Results), roll.Total,
		roll.Purpose, roll.RollNotation).
		Scan(&roll.ID, &roll.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to create dice roll: %w", err)
	}

	return nil
}

// GetByID retrieves a dice roll by ID
func (r *diceRollRepository) GetByID(ctx context.Context, id string) (*models.DiceRoll, error) {
	var roll models.DiceRoll
	query := `
		SELECT id, game_session_id, user_id, dice_type, count, modifier,
			   results, total, purpose, roll_notation, timestamp
		FROM dice_rolls
		WHERE id = ?`

	err := r.db.QueryRowContextRebind(ctx, query, id).Scan(
		&roll.ID, &roll.GameSessionID, &roll.UserID, &roll.DiceType,
		&roll.Count, &roll.Modifier, pq.Array(&roll.Results),
		&roll.Total, &roll.Purpose, &roll.RollNotation, &roll.Timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("dice roll not found")
		}
		return nil, fmt.Errorf("failed to get dice roll by id: %w", err)
	}

	return &roll, nil
}

// GetByGameSession retrieves dice rolls for a game session
func (r *diceRollRepository) GetByGameSession(ctx context.Context, sessionID string, offset, limit int) ([]*models.DiceRoll, error) {
	query := `
		SELECT id, game_session_id, user_id, dice_type, count, modifier,
			   results, total, purpose, roll_notation, timestamp
		FROM dice_rolls
		WHERE game_session_id = ?
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContextRebind(ctx, query, sessionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get dice rolls by game session: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return ScanRowsGeneric(rows, r.scanDiceRoll)
}

// GetByUser retrieves dice rolls for a user
func (r *diceRollRepository) GetByUser(ctx context.Context, userID string, offset, limit int) ([]*models.DiceRoll, error) {
	query := `
		SELECT id, game_session_id, user_id, dice_type, count, modifier,
			   results, total, purpose, roll_notation, timestamp
		FROM dice_rolls
		WHERE user_id = ?
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContextRebind(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get dice rolls by user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return ScanRowsGeneric(rows, r.scanDiceRoll)
}

// GetByGameSessionAndUser retrieves dice rolls for a specific user in a game session
func (r *diceRollRepository) GetByGameSessionAndUser(ctx context.Context, sessionID, userID string, offset, limit int) ([]*models.DiceRoll, error) {
	query := `
		SELECT id, game_session_id, user_id, dice_type, count, modifier,
			   results, total, purpose, roll_notation, timestamp
		FROM dice_rolls
		WHERE game_session_id = ? AND user_id = ?
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContextRebind(ctx, query, sessionID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get dice rolls by game session and user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return ScanRowsGeneric(rows, r.scanDiceRoll)
}

// Delete deletes a dice roll
func (r *diceRollRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM dice_rolls WHERE id = ?`

	result, err := r.db.ExecContextRebind(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete dice roll: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dice roll not found")
	}

	return nil
}
