package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// diceRollRepository implements DiceRollRepository interface
type diceRollRepository struct {
	db *DB
}

// NewDiceRollRepository creates a new dice roll repository
func NewDiceRollRepository(db *DB) DiceRollRepository {
	return &diceRollRepository{db: db}
}

// Create creates a new dice roll
func (r *diceRollRepository) Create(ctx context.Context, roll *models.DiceRoll) error {
	query := `
		INSERT INTO dice_rolls (
			game_session_id, user_id, dice_type, count, modifier,
			results, total, purpose, roll_notation
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, timestamp`

	err := r.db.QueryRowContext(ctx, query,
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
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
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
		WHERE game_session_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, sessionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get dice rolls by game session: %w", err)
	}
	defer rows.Close()

	var rolls []*models.DiceRoll
	for rows.Next() {
		var roll models.DiceRoll
		err := rows.Scan(
			&roll.ID, &roll.GameSessionID, &roll.UserID, &roll.DiceType,
			&roll.Count, &roll.Modifier, pq.Array(&roll.Results),
			&roll.Total, &roll.Purpose, &roll.RollNotation, &roll.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dice roll: %w", err)
		}
		rolls = append(rolls, &roll)
	}

	return rolls, nil
}

// GetByUser retrieves dice rolls for a user
func (r *diceRollRepository) GetByUser(ctx context.Context, userID string, offset, limit int) ([]*models.DiceRoll, error) {
	query := `
		SELECT id, game_session_id, user_id, dice_type, count, modifier,
			   results, total, purpose, roll_notation, timestamp
		FROM dice_rolls
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get dice rolls by user: %w", err)
	}
	defer rows.Close()

	var rolls []*models.DiceRoll
	for rows.Next() {
		var roll models.DiceRoll
		err := rows.Scan(
			&roll.ID, &roll.GameSessionID, &roll.UserID, &roll.DiceType,
			&roll.Count, &roll.Modifier, pq.Array(&roll.Results),
			&roll.Total, &roll.Purpose, &roll.RollNotation, &roll.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dice roll: %w", err)
		}
		rolls = append(rolls, &roll)
	}

	return rolls, nil
}

// GetByGameSessionAndUser retrieves dice rolls for a specific user in a game session
func (r *diceRollRepository) GetByGameSessionAndUser(ctx context.Context, sessionID, userID string, offset, limit int) ([]*models.DiceRoll, error) {
	query := `
		SELECT id, game_session_id, user_id, dice_type, count, modifier,
			   results, total, purpose, roll_notation, timestamp
		FROM dice_rolls
		WHERE game_session_id = $1 AND user_id = $2
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, query, sessionID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get dice rolls by game session and user: %w", err)
	}
	defer rows.Close()

	var rolls []*models.DiceRoll
	for rows.Next() {
		var roll models.DiceRoll
		err := rows.Scan(
			&roll.ID, &roll.GameSessionID, &roll.UserID, &roll.DiceType,
			&roll.Count, &roll.Modifier, pq.Array(&roll.Results),
			&roll.Total, &roll.Purpose, &roll.RollNotation, &roll.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dice roll: %w", err)
		}
		rolls = append(rolls, &roll)
	}

	return rolls, nil
}

// Delete deletes a dice roll
func (r *diceRollRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM dice_rolls WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
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