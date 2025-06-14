package test

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/database"
)

// NewMockDB creates a mock database for testing.
func NewMockDB(t *testing.T) (*database.DB, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "postgres")

	db := &database.DB{
		DB: sqlxDB,
	}

	cleanup := func() {
		_ = mockDB.Close()
	}

	return db, mock, cleanup
}

// NewTestDB creates an in-memory SQLite database for integration tests.
func NewTestDB(t *testing.T) (*database.DB, func()) {
	// Create in-memory SQLite database.
	sqlxDB, err := sqlx.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	db := &database.DB{
		DB: sqlxDB,
	}

	// Create tables.
	err = createTestTables(db)
	require.NoError(t, err)

	cleanup := func() {
		_ = sqlxDB.Close()
	}

	return db, cleanup
}

// createTestTables creates the necessary tables for testing.
func createTestTables(db *database.DB) error {
	schemas := []string{
		// Users table.
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'player',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP
		)`,

		// Characters table.
		`CREATE TABLE IF NOT EXISTS characters (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			race TEXT NOT NULL,
			class TEXT NOT NULL,
			level INTEGER DEFAULT 1,
			experience INTEGER DEFAULT 0,
			hp INTEGER NOT NULL,
			max_hp INTEGER NOT NULL,
			temp_hp INTEGER DEFAULT 0,
			armor_class INTEGER NOT NULL,
			initiative_bonus INTEGER DEFAULT 0,
			speed INTEGER DEFAULT 30,
			proficiency_bonus INTEGER DEFAULT 2,
			abilities TEXT NOT NULL,
			skills TEXT NOT NULL,
			saving_throws TEXT NOT NULL,
			inventory TEXT,
			spell_slots TEXT,
			known_spells TEXT,
			prepared_spells TEXT,
			features TEXT,
			traits TEXT,
			equipment TEXT,
			backstory TEXT,
			carry_capacity INTEGER DEFAULT 150,
			current_weight REAL DEFAULT 0,
			attunement_slots INTEGER DEFAULT 3,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Game sessions table.
		`CREATE TABLE IF NOT EXISTS game_sessions (
			id TEXT PRIMARY KEY,
			dm_user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			status TEXT DEFAULT 'pending',
			session_state TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			ended_at TIMESTAMP,
			FOREIGN KEY (dm_user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Game participants table.
		`CREATE TABLE IF NOT EXISTS game_participants (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game_session_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			character_id TEXT,
			role TEXT NOT NULL DEFAULT 'player',
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_online BOOLEAN DEFAULT false,
			FOREIGN KEY (game_session_id) REFERENCES game_sessions(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE SET NULL,
			UNIQUE(game_session_id, user_id)
		)`,

		// Dice rolls table.
		`CREATE TABLE IF NOT EXISTS dice_rolls (
			id TEXT PRIMARY KEY,
			game_session_id TEXT,
			user_id TEXT NOT NULL,
			character_id TEXT,
			roll_type TEXT NOT NULL,
			dice_notation TEXT NOT NULL,
			results TEXT NOT NULL,
			total INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (game_session_id) REFERENCES game_sessions(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE SET NULL
		)`,

		// Refresh tokens table.
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			revoked_at TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// NPCs table.
		`CREATE TABLE IF NOT EXISTS npcs (
			id TEXT PRIMARY KEY,
			game_session_id TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			hp INTEGER NOT NULL,
			max_hp INTEGER NOT NULL,
			armor_class INTEGER NOT NULL,
			abilities TEXT NOT NULL,
			challenge_rating REAL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (game_session_id) REFERENCES game_sessions(id) ON DELETE CASCADE
		)`,

		// Items table.
		`CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			rarity TEXT NOT NULL DEFAULT 'common',
			weight REAL DEFAULT 0,
			value INTEGER DEFAULT 0,
			properties TEXT,
			requires_attunement BOOLEAN DEFAULT false,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Character inventory table.
		`CREATE TABLE IF NOT EXISTS character_inventory (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			character_id TEXT NOT NULL,
			item_id TEXT NOT NULL,
			quantity INTEGER DEFAULT 1,
			equipped BOOLEAN DEFAULT false,
			attuned BOOLEAN DEFAULT false,
			custom_name TEXT,
			notes TEXT,
			added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
			FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
		)`,

		// Character currency table.
		`CREATE TABLE IF NOT EXISTS character_currency (
			character_id TEXT PRIMARY KEY,
			copper INTEGER DEFAULT 0,
			silver INTEGER DEFAULT 0,
			electrum INTEGER DEFAULT 0,
			gold INTEGER DEFAULT 0,
			platinum INTEGER DEFAULT 0,
			FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
		)`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// SeedTestData adds test data to the database.
func SeedTestData(t *testing.T, db *database.DB) {
	// Add test users.
	_, err := db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role)
		VALUES 
			('user-1', 'testplayer', 'player@test.com', '$2a$10$YourHashedPasswordHere', 'player'),
			('user-2', 'testdm', 'dm@test.com', '$2a$10$YourHashedPasswordHere', 'dm')
	`)
	require.NoError(t, err)

	// Add test characters.
	_, err = db.Exec(`
		INSERT INTO characters (id, user_id, name, race, class, level, hp, max_hp, armor_class, abilities, skills, saving_throws)
		VALUES 
			('char-1', 'user-1', 'Test Fighter', 'human', 'fighter', 1, 10, 10, 16, '{}', '{}', '{}'),
			('char-2', 'user-1', 'Test Wizard', 'elf', 'wizard', 1, 6, 6, 12, '{}', '{}', '{}')
	`)
	require.NoError(t, err)

	// Add test items.
	_, err = db.Exec(`
		INSERT INTO items (id, name, type, rarity, weight, value, requires_attunement)
		VALUES 
			('item-1', 'Longsword', 'weapon', 'common', 3, 15, false),
			('item-2', 'Chain Mail', 'armor', 'common', 55, 75, false),
			('item-3', 'Ring of Protection', 'ring', 'rare', 0, 5000, true)
	`)
	require.NoError(t, err)
}

// TruncateTables clears all data from tables.
func TruncateTables(t *testing.T, db *database.DB) {
	tables := []string{
		"character_currency",
		"character_inventory",
		"dice_rolls",
		"game_participants",
		"npcs",
		"refresh_tokens",
		"characters",
		"game_sessions",
		"users",
		"items",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(t, err)
	}
}
