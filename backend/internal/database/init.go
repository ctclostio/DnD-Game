package database

import (
	"fmt"
	"time"

	"github.com/your-username/dnd-game/backend/internal/config"
	"github.com/your-username/dnd-game/backend/pkg/logger"
)

// Initialize creates and initializes the database connection and repositories
func Initialize(cfg *config.Config) (*DB, *Repositories, error) {
	return InitializeWithLogging(cfg, nil)
}

// InitializeWithLogging creates and initializes the database connection and repositories with optional logging
func InitializeWithLogging(cfg *config.Config, log *logger.LoggerV2) (*DB, *Repositories, error) {
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

		if log != nil {
			log.Error().
				Err(err).
				Int("attempt", i+1).
				Int("max_retries", maxRetries).
				Msg("Failed to connect to database")
		}
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	if log != nil {
		log.Info().
			Str("host", cfg.Database.Host).
			Int("port", cfg.Database.Port).
			Str("database", cfg.Database.DatabaseName).
			Msg("Successfully connected to database")
	}

	// Run migrations
	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	if log != nil {
		log.Info().
			Msg("Database migrations completed successfully")
	}

	// Set logger on database connection if provided
	if log != nil {
		db.SetLogger(log)
	}

	// Create repositories
	repos := &Repositories{
		Users:           NewUserRepository(db),
		Characters:      NewCharacterRepository(db),
		GameSessions:    NewGameSessionRepository(db),
		DiceRolls:       NewDiceRollRepository(db),
		NPCs:            NewNPCRepository(db.DB),
		Inventory:       NewInventoryRepository(db),
		RefreshTokens:   NewRefreshTokenRepository(db.DB),
		CustomRaces:     NewCustomRaceRepository(db.DB),
		CustomClasses:   NewCustomClassRepository(db),
		DMAssistant:     NewDMAssistantRepository(db.DB),
		Encounters:      NewEncounterRepository(db),
		Campaign:        NewCampaignRepository(db.DB),
		CombatAnalytics: NewCombatAnalyticsRepository(db.DB),
		WorldBuilding:   NewWorldBuildingRepository(db),
		Narrative:       NewNarrativeRepository(db.DB),
		RuleBuilder:     NewRuleBuilderRepository(db),
	}

	return db, repos, nil
}

// Ping checks if the database connection is alive
func Ping(db *DB) error {
	return db.Ping()
}
