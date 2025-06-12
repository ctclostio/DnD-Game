package database

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrations embed.FS

// RunMigrations runs database migrations
func RunMigrations(db *DB) error {
	// Create source from embedded files
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create database driver
	driver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// RollbackMigration rolls back the last migration
func RollbackMigration(db *DB) error {
	// Create source from embedded files
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create database driver
	driver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Rollback one migration
	err = m.Steps(-1)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// GetMigrationVersion returns the current migration version
func GetMigrationVersion(db *DB) (uint, bool, error) {
	// Create source from embedded files
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create database driver
	driver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m.Version()
}
