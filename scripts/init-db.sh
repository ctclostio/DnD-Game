#!/bin/bash

# Database initialization script for D&D Game
# This script initializes the database and runs migrations in Docker environment

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Wait for PostgreSQL to be ready
wait_for_postgres() {
    print_message $YELLOW "Waiting for PostgreSQL to be ready..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c '\q' 2>/dev/null; then
            print_message $GREEN "✓ PostgreSQL is ready"
            return 0
        fi
        
        print_message $YELLOW "PostgreSQL is not ready yet. Attempt $attempt/$max_attempts..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_message $RED "✗ PostgreSQL failed to start within the expected time"
    return 1
}

# Create database if it doesn't exist
create_database() {
    print_message $YELLOW "Checking if database exists..."
    
    # Check if database exists
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1
    
    if [ $? -eq 0 ]; then
        print_message $GREEN "✓ Database '$DB_NAME' already exists"
    else
        print_message $YELLOW "Creating database '$DB_NAME'..."
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME"
        
        if [ $? -eq 0 ]; then
            print_message $GREEN "✓ Database created successfully"
        else
            print_message $RED "✗ Failed to create database"
            return 1
        fi
    fi
}

# Main execution
print_message $GREEN "=== D&D Game Database Initialization ==="

# Wait for PostgreSQL
if ! wait_for_postgres; then
    exit 1
fi

# Create database
if ! create_database; then
    exit 1
fi

print_message $GREEN "✓ Database initialization completed successfully"

# Note: Migrations will be run by the application on startup