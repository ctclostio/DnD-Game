package testhelpers

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// TestDB wraps a mock database for testing
type TestDB struct {
	DB   *sqlx.DB
	Mock sqlmock.Sqlmock
}

// NewTestDB creates a new test database with sqlmock
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()
	
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err, "Failed to create mock database")
	
	db := sqlx.NewDb(mockDB, "sqlmock")
	
	return &TestDB{
		DB:   db,
		Mock: mock,
	}
}

// Close closes the test database
func (tdb *TestDB) Close() error {
	return tdb.DB.Close()
}

// AssertExpectationsMet verifies all expectations were met
func (tdb *TestDB) AssertExpectationsMet(t *testing.T) {
	t.Helper()
	
	err := tdb.Mock.ExpectationsWereMet()
	require.NoError(t, err, "Database expectations were not met")
}

// ExpectBegin sets up an expectation for a transaction begin
func (tdb *TestDB) ExpectBegin() {
	tdb.Mock.ExpectBegin()
}

// ExpectCommit sets up an expectation for a transaction commit
func (tdb *TestDB) ExpectCommit() {
	tdb.Mock.ExpectCommit()
}

// ExpectRollback sets up an expectation for a transaction rollback
func (tdb *TestDB) ExpectRollback() {
	tdb.Mock.ExpectRollback()
}

// ExpectQuery sets up an expectation for a SELECT query
func (tdb *TestDB) ExpectQuery(query string) *sqlmock.ExpectedQuery {
	// Convert ? placeholders to \? for regex matching
	return tdb.Mock.ExpectQuery(ConvertPlaceholders(query))
}

// ExpectExec sets up an expectation for an INSERT/UPDATE/DELETE query
func (tdb *TestDB) ExpectExec(query string) *sqlmock.ExpectedExec {
	// Convert ? placeholders to \? for regex matching
	return tdb.Mock.ExpectExec(ConvertPlaceholders(query))
}

// ConvertPlaceholders converts ? placeholders to \? for regex matching
func ConvertPlaceholders(query string) string {
	// This is a simple conversion - in production you might want a more robust solution
	result := ""
	for i, char := range query {
		if char == '?' && (i == 0 || query[i-1] != '\\') {
			result += `\?`
		} else {
			result += string(char)
		}
	}
	return result
}

// TestTx wraps a mock transaction
type TestTx struct {
	*sql.Tx
	Mock sqlmock.Sqlmock
}

// NewTestTx creates a test transaction
func NewTestTx(t *testing.T) (*TestTx, *sql.DB) {
	t.Helper()
	
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	
	return &TestTx{Tx: tx, Mock: mock}, db
}

// Common database test scenarios

// SetupGetByIDSuccess sets up a successful GetByID query
func (tdb *TestDB) SetupGetByIDSuccess(table string, columns []string, result interface{}) {
	query := `SELECT .* FROM ` + table + ` WHERE id = \?`
	rows := sqlmock.NewRows(columns)
	
	// Add row data based on result type
	// This is simplified - you'd expand based on your models
	rows.AddRow(result)
	
	tdb.Mock.ExpectQuery(query).WillReturnRows(rows)
}

// SetupGetByIDNotFound sets up a GetByID query that returns no rows
func (tdb *TestDB) SetupGetByIDNotFound(table string) {
	query := `SELECT .* FROM ` + table + ` WHERE id = \?`
	tdb.Mock.ExpectQuery(query).WillReturnError(sql.ErrNoRows)
}

// SetupCreateSuccess sets up a successful INSERT
func (tdb *TestDB) SetupCreateSuccess(table string) {
	query := `INSERT INTO ` + table
	tdb.Mock.ExpectExec(query).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

// SetupUpdateSuccess sets up a successful UPDATE
func (tdb *TestDB) SetupUpdateSuccess(table string, rowsAffected int64) {
	query := `UPDATE ` + table
	tdb.Mock.ExpectExec(query).
		WillReturnResult(sqlmock.NewResult(0, rowsAffected))
}

// SetupDeleteSuccess sets up a successful DELETE
func (tdb *TestDB) SetupDeleteSuccess(table string) {
	query := `DELETE FROM ` + table
	tdb.Mock.ExpectExec(query).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

// TestRepository provides common test repository functionality
type TestRepository struct {
	DB      *TestDB
	Context context.Context
}

// NewTestRepository creates a new test repository
func NewTestRepository(t *testing.T) *TestRepository {
	return &TestRepository{
		DB:      NewTestDB(t),
		Context: context.Background(),
	}
}

// Cleanup cleans up the test repository
func (tr *TestRepository) Cleanup(t *testing.T) {
	t.Helper()
	
	tr.DB.AssertExpectationsMet(t)
	err := tr.DB.Close()
	require.NoError(t, err)
}

// WithContext returns a new test repository with the given context
func (tr *TestRepository) WithContext(ctx context.Context) *TestRepository {
	return &TestRepository{
		DB:      tr.DB,
		Context: ctx,
	}
}

// Common SQL patterns for D&D game

// CharacterColumns returns standard character table columns
func CharacterColumns() []string {
	return []string{
		"id", "user_id", "name", "race", "class", "level",
		"hit_points", "max_hp", "armor_class", "initiative_bonus",
		"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma",
		"created_at", "updated_at",
	}
}

// GameSessionColumns returns standard game session table columns  
func GameSessionColumns() []string {
	return []string{
		"id", "name", "description", "dm_id", "status",
		"created_at", "updated_at",
	}
}

// CombatColumns returns standard combat table columns
func CombatColumns() []string {
	return []string{
		"id", "game_session_id", "status", "current_turn", "round",
		"started_at", "ended_at", "created_at", "updated_at",
	}
}

// ItemColumns returns standard item table columns
func ItemColumns() []string {
	return []string{
		"id", "name", "type", "rarity", "description", "properties",
		"created_at", "updated_at",
	}
}

// UserColumns returns standard user table columns
func UserColumns() []string {
	return []string{
		"id", "username", "email", "password_hash", "role",
		"created_at", "updated_at",
	}
}