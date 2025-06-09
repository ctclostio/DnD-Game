package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Config holds the database configuration
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// DB wraps the database connection
type DB struct {
	*sqlx.DB
	config Config
}

// StdDB returns the underlying *sql.DB
func (db *DB) StdDB() *sql.DB {
	return db.DB.DB
}

// NewConnection creates a new database connection
func NewConnection(cfg Config) (*DB, error) {
	// Construct DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DatabaseName, cfg.SSLMode)

	// Open database connection
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.MaxLifetime)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		DB:     db,
		config: cfg,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// WithTx executes a function within a database transaction
func (db *DB) WithTx(fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %v, unable to rollback: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDB returns the underlying sqlx.DB instance
func (db *DB) GetDB() *sqlx.DB {
	return db.DB
}

// QueryRowContext executes a query with context that is expected to return at most one row
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

// ExecContext executes a query with context without returning any rows
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

// GetConfig returns the database configuration
func (db *DB) GetConfig() Config {
	return db.config
}

// Exec executes a query without returning any rows
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.Exec(query, args...)
}

// Query executes a query that returns rows
func (db *DB) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	return db.DB.Queryx(query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (db *DB) QueryRow(query string, args ...interface{}) *sqlx.Row {
	return db.DB.QueryRowx(query, args...)
}

// Get executes a query and scans the result into dest
func (db *DB) Get(dest interface{}, query string, args ...interface{}) error {
	return db.DB.Get(dest, query, args...)
}

// Select executes a query and scans the results into dest
func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return db.DB.Select(dest, query, args...)
}

// Rebind transforms a query from QUESTION to the DB driver's bindvar type
// This allows us to write queries with ? placeholders that work with both SQLite and PostgreSQL
func (db *DB) Rebind(query string) string {
	return db.DB.Rebind(query)
}

// ExecRebind executes a query after rebinding placeholders
func (db *DB) ExecRebind(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.Exec(db.Rebind(query), args...)
}

// QueryRowRebind executes a query after rebinding placeholders
func (db *DB) QueryRowRebind(query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRow(db.Rebind(query), args...)
}

// ExecContextRebind executes a query with context after rebinding placeholders
func (db *DB) ExecContextRebind(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, db.Rebind(query), args...)
}

// QueryRowContextRebind executes a query with context after rebinding placeholders
func (db *DB) QueryRowContextRebind(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, db.Rebind(query), args...)
}

// QueryContextRebind executes a query with context after rebinding placeholders
func (db *DB) QueryContextRebind(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, db.Rebind(query), args...)
}