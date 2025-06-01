package database

import (
	"fmt"
	"log"
	"time"

	"github.com/your-username/dnd-game/backend/internal/config"
)

// Initialize creates and initializes the database connection and repositories
func Initialize(cfg *config.Config) (*DB, *Repositories, error) {
	// Create database configuration
	dbConfig := Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		DatabaseName: cfg.Database.DatabaseName,
		SSLMode:      cfg.Database.SSLMode,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		MaxLifetime:  cfg.Database.MaxLifetime,
	}

	// Connect to database with retry logic
	var db *DB
	var err error
	
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = NewConnection(dbConfig)
		if err == nil {
			break
		}
		
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	log.Println("Successfully connected to database")

	// Run migrations
	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")

	// Create repositories
	repos := &Repositories{
		Users:           NewUserRepository(db),
		Characters:      NewCharacterRepository(db),
		GameSessions:    NewGameSessionRepository(db),
		DiceRolls:       NewDiceRollRepository(db),
		NPCs:            NewNPCRepository(db.DB),
		Inventory:       NewInventoryRepository(db.DB),
		RefreshTokens:   NewRefreshTokenRepository(db.DB),
		CustomRaces:     NewCustomRaceRepository(db.DB),
		CustomClasses:   NewCustomClassRepository(db.StdDB()),
		DMAssistant:     NewDMAssistantRepository(db.DB),
		Encounters:      NewEncounterRepository(db.StdDB()),
		Campaign:        NewCampaignRepository(db.DB),
		CombatAnalytics: NewCombatAnalyticsRepository(db.DB),
		WorldBuilding:   NewWorldBuildingRepository(db.DB),
		Narrative:       NewNarrativeRepository(db.DB),
	}

	return db, repos, nil
}

// Ping checks if the database connection is alive
func Ping(db *DB) error {
	return db.Ping()
}