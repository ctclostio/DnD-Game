package testutil

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// Regular expression for valid SQL identifiers (table and column names)
var validSQLIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// isValidTableName validates table names to prevent SQL injection
func isValidTableName(name string) bool {
	// Whitelist of known table names in the test database
	validTables := map[string]bool{
		"users":           true,
		"refresh_tokens":  true,
		"game_sessions":   true,
		"characters":      true,
		"combat_sessions": true,
		"combat_logs":     true,
		"inventories":     true,
		"items":           true,
		"factions":        true,
		"settlements":     true,
		"cultures":        true,
		"world_events":    true,
	}

	return validTables[name] && validSQLIdentifier.MatchString(name)
}

// isValidColumnName validates column names to prevent SQL injection
func isValidColumnName(name string) bool {
	// Basic validation: must be a valid SQL identifier
	return validSQLIdentifier.MatchString(name) && len(name) <= 64
}

// NewMockDB creates a new mock database for testing
func NewMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	return sqlxDB, mock
}

// SetupTestDB creates an in-memory SQLite database for integration tests
func SetupTestDB(t *testing.T) *sqlx.DB {
	// Use shared cache mode for better concurrency support
	db, err := sqlx.Open("sqlite3", ":memory:?cache=shared&mode=rwc")
	require.NoError(t, err)

	// Set connection pool settings for better concurrency
	db.SetMaxOpenConns(1) // SQLite can only have one writer at a time
	db.SetMaxIdleConns(1)

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Create test schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'player',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash TEXT NOT NULL UNIQUE,
		token_id TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		revoked_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS characters (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id),
		name TEXT NOT NULL,
		race TEXT NOT NULL,
		subrace TEXT,
		class TEXT NOT NULL,
		subclass TEXT,
		background TEXT,
		alignment TEXT,
		level INTEGER DEFAULT 1,
		experience_points INTEGER DEFAULT 0,
		hit_points INTEGER NOT NULL,
		max_hit_points INTEGER NOT NULL,
		temp_hit_points INTEGER DEFAULT 0,
		hit_dice TEXT,
		armor_class INTEGER NOT NULL,
		initiative INTEGER DEFAULT 0,
		speed INTEGER DEFAULT 30,
		proficiency_bonus INTEGER DEFAULT 2,
		attributes JSONB,
		saving_throws JSONB,
		skills JSONB,
		proficiencies JSONB,
		features JSONB,
		equipment JSONB,
		spells JSONB,
		resources JSONB,
		carry_capacity REAL DEFAULT 0,
		current_weight REAL DEFAULT 0,
		attunement_slots_used INTEGER DEFAULT 0,
		attunement_slots_max INTEGER DEFAULT 3,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS items (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		rarity TEXT DEFAULT 'common',
		weight REAL DEFAULT 0,
		value INTEGER DEFAULT 0,
		properties JSONB DEFAULT '{}',
		requires_attunement BOOLEAN DEFAULT FALSE,
		attunement_requirements TEXT,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS character_inventory (
		id TEXT PRIMARY KEY,
		character_id TEXT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
		item_id TEXT NOT NULL REFERENCES items(id),
		quantity INTEGER DEFAULT 1,
		equipped BOOLEAN DEFAULT FALSE,
		attuned BOOLEAN DEFAULT FALSE,
		custom_properties JSONB DEFAULT '{}',
		notes TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT unique_character_item UNIQUE(character_id, item_id)
	);

	CREATE TABLE IF NOT EXISTS character_currency (
		character_id TEXT PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
		copper INTEGER DEFAULT 0,
		silver INTEGER DEFAULT 0,
		electrum INTEGER DEFAULT 0,
		gold INTEGER DEFAULT 0,
		platinum INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS game_sessions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		dm_user_id TEXT NOT NULL REFERENCES users(id),
		code TEXT UNIQUE NOT NULL,
		max_players INTEGER DEFAULT 6,
		is_public BOOLEAN DEFAULT FALSE,
		requires_invite BOOLEAN DEFAULT FALSE,
		allowed_character_level INTEGER DEFAULT 0,
		is_active BOOLEAN DEFAULT TRUE,
		status TEXT NOT NULL DEFAULT 'active',
		session_state TEXT DEFAULT '{}',
		started_at TIMESTAMP,
		ended_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS game_participants (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
		user_id TEXT NOT NULL REFERENCES users(id),
		character_id TEXT REFERENCES characters(id),
		joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_online BOOLEAN DEFAULT FALSE,
		CONSTRAINT unique_session_user UNIQUE(session_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS dice_rolls (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id),
		character_id TEXT REFERENCES characters(id),
		session_id TEXT REFERENCES game_sessions(id),
		dice_type TEXT NOT NULL,
		result JSONB NOT NULL,
		purpose TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS npcs (
		id TEXT PRIMARY KEY,
		game_session_id TEXT REFERENCES game_sessions(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		size TEXT NOT NULL,
		alignment TEXT,
		armor_class INTEGER NOT NULL DEFAULT 10,
		hit_points INTEGER NOT NULL,
		max_hit_points INTEGER NOT NULL,
		speed JSONB DEFAULT '{"walk": 30}',
		attributes JSONB NOT NULL DEFAULT '{"strength": 10, "dexterity": 10, "constitution": 10, "intelligence": 10, "wisdom": 10, "charisma": 10}',
		saving_throws JSONB DEFAULT '{}',
		skills JSONB DEFAULT '[]',
		damage_resistances JSONB DEFAULT '[]',
		damage_immunities JSONB DEFAULT '[]',
		condition_immunities JSONB DEFAULT '[]',
		senses JSONB DEFAULT '{}',
		languages JSONB DEFAULT '[]',
		challenge_rating REAL DEFAULT 0,
		experience_points INTEGER DEFAULT 0,
		abilities JSONB DEFAULT '[]',
		actions JSONB DEFAULT '[]',
		legendary_actions INTEGER DEFAULT 0,
		is_template BOOLEAN DEFAULT FALSE,
		created_by TEXT REFERENCES users(id) ON DELETE SET NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

// CleanupDB closes the database connection
func CleanupDB(db *sqlx.DB) {
	if db != nil {
		_ = db.Close()
	}
}

// TruncateTables clears all data from tables for test isolation
func TruncateTables(t *testing.T, db *sqlx.DB) {
	tables := []string{
		"dice_rolls",
		"npcs",
		"game_participants",
		"game_sessions",
		"character_currency",
		"character_inventory",
		"items",
		"characters",
		"refresh_tokens",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(t, err)
	}
}

// SeedTestUser creates a test user
func SeedTestUser(t *testing.T, db *sqlx.DB, id, username, email, role string) {
	query := `
		INSERT INTO users (id, username, email, password_hash, role)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := db.Exec(db.Rebind(query), id, username, email, "hashed_password", role)
	require.NoError(t, err)
}

// SeedTestCharacter creates a test character
func SeedTestCharacter(t *testing.T, db *sqlx.DB, id, userID, name string) {
	query := `
		INSERT INTO characters (
			id, user_id, name, race, subrace, class, subclass, background, alignment, level,
			experience_points, hit_points, max_hit_points, armor_class, speed,
			attributes, saving_throws, skills, equipment, spells
		) VALUES (
			$1, $2, $3, 'Human', NULL, 'Fighter', NULL, 'Soldier', 'Lawful Good', 1,
			0, 10, 10, 15, 30,
			'{"strength":16,"dexterity":14,"constitution":14,"intelligence":10,"wisdom":12,"charisma":8}',
			'{}', '[]', '[]', '{}'
		)
	`
	_, err := db.Exec(query, id, userID, name)
	require.NoError(t, err)
}

// SeedTestItem creates a test item
func SeedTestItem(t *testing.T, db *sqlx.DB, id, name, itemType string, value int) {
	query := `
		INSERT INTO items (id, name, type, value, weight, properties)
		VALUES ($1, $2, $3, $4, 1.0, '{}')
	`
	_, err := db.Exec(query, id, name, itemType, value)
	require.NoError(t, err)
}

// AssertRowExists checks if a row exists in a table
// Uses a whitelist of allowed table and column names to prevent SQL injection
func AssertRowExists(t *testing.T, db *sqlx.DB, table, column, value string) {
	// Validate table and column names to prevent SQL injection
	if !isValidTableName(table) {
		t.Fatalf("Invalid table name: %s", table)
	}
	if !isValidColumnName(column) {
		t.Fatalf("Invalid column name: %s", column)
	}

	var count int
	// Safe to use fmt.Sprintf after validation
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", table, column)
	err := db.Get(&count, query, value)
	require.NoError(t, err)
	require.Equal(t, 1, count, "Expected row with %s='%s' in table %s", column, value, table)
}

// AssertRowNotExists checks if a row does not exist in a table
// Uses a whitelist of allowed table and column names to prevent SQL injection
func AssertRowNotExists(t *testing.T, db *sqlx.DB, table, column, value string) {
	// Validate table and column names to prevent SQL injection
	if !isValidTableName(table) {
		t.Fatalf("Invalid table name: %s", table)
	}
	if !isValidColumnName(column) {
		t.Fatalf("Invalid column name: %s", column)
	}

	var count int
	// Safe to use fmt.Sprintf after validation
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", table, column)
	err := db.Get(&count, query, value)
	require.NoError(t, err)
	require.Equal(t, 0, count, "Expected no row with %s='%s' in table %s", column, value, table)
}
