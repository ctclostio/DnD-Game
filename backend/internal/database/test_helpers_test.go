package database

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/stretchr/testify/require"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) (*DB, func()) {
	// Create in-memory SQLite database
	sqlxDB, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create schema - you'll need to add your actual schema here
	schema := `
	CREATE TABLE IF NOT EXISTS settlements (
		id TEXT PRIMARY KEY,
		game_session_id TEXT NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		population INTEGER NOT NULL,
		wealth_level INTEGER NOT NULL,
		region TEXT,
		danger_level INTEGER,
		corruption_level INTEGER,
		alignment TEXT,
		government_type TEXT,
		terrain_type TEXT,
		climate TEXT,
		age_category TEXT,
		description TEXT,
		history TEXT,
		coordinates TEXT,
		primary_exports TEXT DEFAULT '[]',
		primary_imports TEXT DEFAULT '[]',
		trade_routes TEXT DEFAULT '[]',
		ancient_ruins_nearby BOOLEAN DEFAULT FALSE,
		eldritch_influence INTEGER DEFAULT 0,
		ley_line_connection BOOLEAN DEFAULT FALSE,
		notable_locations TEXT DEFAULT '[]',
		defenses TEXT DEFAULT '[]',
		problems TEXT DEFAULT '[]',
		secrets TEXT DEFAULT '[]',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS markets (
		id TEXT PRIMARY KEY,
		settlement_id TEXT NOT NULL,
		food_price_modifier REAL DEFAULT 1.0,
		common_goods_modifier REAL DEFAULT 1.0,
		weapons_armor_modifier REAL DEFAULT 1.0,
		magical_items_modifier REAL DEFAULT 1.0,
		available_items TEXT,
		seasonal_factors TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (settlement_id) REFERENCES settlements(id)
	);

	CREATE TABLE IF NOT EXISTS inventories (
		id TEXT PRIMARY KEY,
		character_id TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL,
		last_login TIMESTAMP,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS game_sessions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		dm_user_id TEXT NOT NULL,
		session_code TEXT UNIQUE NOT NULL,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (dm_user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS characters (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		race TEXT NOT NULL,
		class TEXT NOT NULL,
		level INTEGER DEFAULT 1,
		experience_points INTEGER DEFAULT 0,
		ability_scores TEXT,
		skills TEXT,
		proficiencies TEXT,
		hit_points INTEGER,
		max_hit_points INTEGER,
		temporary_hit_points INTEGER DEFAULT 0,
		armor_class INTEGER,
		initiative_modifier INTEGER,
		speed INTEGER DEFAULT 30,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	_, err = sqlxDB.Exec(schema)
	require.NoError(t, err)

	db := &DB{DB: sqlxDB}

	cleanup := func() {
		sqlxDB.Close()
	}

	return db, cleanup
}