#!/bin/bash

# Database migration script for D&D Game

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-dndgame}"
DB_PASSWORD="${DB_PASSWORD:-dndgamepass}"
DB_NAME="${DB_NAME:-dndgame}"
DB_SSLMODE="${DB_SSLMODE:-disable}"

# Migration directory
MIGRATION_DIR="../backend/internal/database/migrations"

# Function to print colored output
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if PostgreSQL is running
check_postgres() {
    print_message $YELLOW "Checking PostgreSQL connection..."
    
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c '\q' 2>/dev/null
    
    if [ $? -eq 0 ]; then
        print_message $GREEN "✓ PostgreSQL is running and accessible"
        return 0
    else
        print_message $RED "✗ Cannot connect to PostgreSQL"
        return 1
    fi
}

# Function to run migrations
run_migrations() {
    print_message $YELLOW "Running database migrations..."
    
    # Build the database URL
    DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
    
    # Run migrations using golang-migrate
    migrate -path $MIGRATION_DIR -database $DATABASE_URL up
    
    if [ $? -eq 0 ]; then
        print_message $GREEN "✓ Migrations completed successfully"
        return 0
    else
        print_message $RED "✗ Migration failed"
        return 1
    fi
}

# Function to rollback migrations
rollback_migrations() {
    local steps=${1:-1}
    print_message $YELLOW "Rolling back $steps migration(s)..."
    
    # Build the database URL
    DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
    
    # Rollback migrations
    migrate -path $MIGRATION_DIR -database $DATABASE_URL down $steps
    
    if [ $? -eq 0 ]; then
        print_message $GREEN "✓ Rollback completed successfully"
        return 0
    else
        print_message $RED "✗ Rollback failed"
        return 1
    fi
}

# Function to show migration status
migration_status() {
    print_message $YELLOW "Current migration status:"
    
    # Build the database URL
    DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
    
    # Show version
    migrate -path $MIGRATION_DIR -database $DATABASE_URL version
}

# Function to create a new migration
create_migration() {
    local name=$1
    if [ -z "$name" ]; then
        print_message $RED "Error: Migration name is required"
        echo "Usage: $0 create <migration_name>"
        exit 1
    fi
    
    print_message $YELLOW "Creating new migration: $name"
    
    # Create migration files
    migrate create -ext sql -dir $MIGRATION_DIR -seq $name
    
    if [ $? -eq 0 ]; then
        print_message $GREEN "✓ Migration files created successfully"
        return 0
    else
        print_message $RED "✗ Failed to create migration files"
        return 1
    fi
}

# Main script logic
case "$1" in
    up)
        check_postgres && run_migrations
        ;;
    down)
        check_postgres && rollback_migrations ${2:-1}
        ;;
    status)
        check_postgres && migration_status
        ;;
    create)
        create_migration "$2"
        ;;
    *)
        echo "D&D Game Database Migration Tool"
        echo ""
        echo "Usage: $0 {up|down|status|create} [options]"
        echo ""
        echo "Commands:"
        echo "  up              Run all pending migrations"
        echo "  down [n]        Rollback n migrations (default: 1)"
        echo "  status          Show current migration version"
        echo "  create <name>   Create a new migration with the given name"
        echo ""
        echo "Environment variables:"
        echo "  DB_HOST         Database host (default: localhost)"
        echo "  DB_PORT         Database port (default: 5432)"
        echo "  DB_USER         Database user (default: dndgame)"
        echo "  DB_PASSWORD     Database password (default: dndgamepass)"
        echo "  DB_NAME         Database name (default: dndgame)"
        echo "  DB_SSLMODE      SSL mode (default: disable)"
        exit 1
        ;;
esac