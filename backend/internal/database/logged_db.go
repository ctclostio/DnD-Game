package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/your-username/dnd-game/backend/pkg/logger"
)

// LoggedDB wraps DB with query logging
type LoggedDB struct {
	*DB
	logger *logger.LoggerV2
}

// NewLoggedDB creates a new LoggedDB wrapper
func NewLoggedDB(db *DB, logger *logger.LoggerV2) *LoggedDB {
	return &LoggedDB{
		DB:     db,
		logger: logger,
	}
}

// logQuery logs a database query with timing and context
func (ldb *LoggedDB) logQuery(ctx context.Context, query string, args []interface{}, err error, duration time.Duration) {
	log := ldb.logger.WithContext(ctx)

	// Truncate query for logging
	truncatedQuery := truncateQuery(query)

	event := log.Debug().
		Str("query", truncatedQuery).
		Dur("duration", duration).
		Int64("duration_ms", duration.Milliseconds()).
		Int("args_count", len(args))

	// Add first few args for debugging (be careful not to log sensitive data)
	if len(args) > 0 && len(args) <= 3 {
		event = event.Interface("args", args)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			event.Msg("Database query returned no rows")
		} else {
			event.Err(err).Msg("Database query failed")
		}
	} else {
		event.Msg("Database query executed")
	}
}

// QueryRowContext executes a query with context that is expected to return at most one row
func (ldb *LoggedDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := ldb.DB.QueryRowContext(ctx, query, args...)
	// Note: We can't get the error here until Scan is called
	ldb.logQuery(ctx, query, args, nil, time.Since(start))
	return row
}

// ExecContext executes a query with context without returning any rows
func (ldb *LoggedDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := ldb.DB.ExecContext(ctx, query, args...)
	ldb.logQuery(ctx, query, args, err, time.Since(start))
	return result, err
}

// QueryContext executes a query with context that returns rows
func (ldb *LoggedDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := ldb.DB.QueryContext(ctx, query, args...)
	ldb.logQuery(ctx, query, args, err, time.Since(start))
	return rows, err
}

// GetContext executes a query with context and scans the result into dest
func (ldb *LoggedDB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := ldb.DB.GetContext(ctx, dest, query, args...)
	ldb.logQuery(ctx, query, args, err, time.Since(start))
	return err
}

// SelectContext executes a query with context and scans the results into dest
func (ldb *LoggedDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := ldb.DB.SelectContext(ctx, dest, query, args...)
	ldb.logQuery(ctx, query, args, err, time.Since(start))
	return err
}

// QueryRowContextRebind executes a query with context after rebinding placeholders
func (ldb *LoggedDB) QueryRowContextRebind(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := ldb.DB.QueryRowContextRebind(ctx, query, args...)
	ldb.logQuery(ctx, query, args, nil, time.Since(start))
	return row
}

// ExecContextRebind executes a query with context after rebinding placeholders
func (ldb *LoggedDB) ExecContextRebind(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := ldb.DB.ExecContextRebind(ctx, query, args...)
	ldb.logQuery(ctx, query, args, err, time.Since(start))
	return result, err
}

// QueryContextRebind executes a query with context after rebinding placeholders
func (ldb *LoggedDB) QueryContextRebind(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := ldb.DB.QueryContextRebind(ctx, query, args...)
	ldb.logQuery(ctx, query, args, err, time.Since(start))
	return rows, err
}

// WithTx executes a function within a database transaction with logging
func (ldb *LoggedDB) WithTx(fn func(*sqlx.Tx) error) error {
	start := time.Now()
	ctx := context.Background()

	log := ldb.logger.WithContext(ctx)
	log.Debug().Msg("Beginning database transaction")

	err := ldb.DB.WithTx(fn)

	duration := time.Since(start)
	if err != nil {
		log.Error().
			Err(err).
			Dur("duration", duration).
			Msg("Database transaction failed")
	} else {
		log.Debug().
			Dur("duration", duration).
			Msg("Database transaction committed")
	}

	return err
}

// WithTxContext executes a function within a database transaction with context and logging
func (ldb *LoggedDB) WithTxContext(ctx context.Context, fn func(*sqlx.Tx) error) error {
	start := time.Now()

	log := ldb.logger.WithContext(ctx)
	log.Debug().Msg("Beginning database transaction")

	tx, err := ldb.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return err
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Error().
				Err(rbErr).
				Str("original_error", err.Error()).
				Msg("Failed to rollback transaction")
			return err
		}
		log.Debug().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("Database transaction rolled back")
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("Failed to commit transaction")
		return err
	}

	log.Debug().
		Dur("duration", time.Since(start)).
		Msg("Database transaction committed")

	return nil
}

// truncateQuery truncates long queries for logging
func truncateQuery(query string) string {
	const maxLength = 500
	if len(query) > maxLength {
		return query[:maxLength] + "..."
	}
	return query
}
