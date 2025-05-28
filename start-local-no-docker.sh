#!/bin/bash

echo "Starting D&D Game Application (No Docker)..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}NOTE: This script assumes you have PostgreSQL installed locally${NC}"
echo -e "${YELLOW}If not, install it with: brew install postgresql${NC}"
echo ""

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo -e "${RED}PostgreSQL is not running!${NC}"
    echo "Please start PostgreSQL first:"
    echo "  brew services start postgresql"
    echo "Or run PostgreSQL manually:"
    echo "  postgres -D /usr/local/var/postgres"
    exit 1
fi

# Create database if it doesn't exist
echo -e "${YELLOW}Setting up database...${NC}"
createdb dndgame 2>/dev/null || echo "Database already exists"
psql dndgame -c "CREATE USER dndgame WITH PASSWORD 'dndgamepass';" 2>/dev/null || echo "User already exists"
psql dndgame -c "GRANT ALL PRIVILEGES ON DATABASE dndgame TO dndgame;" 2>/dev/null || echo "Privileges already granted"

# Start backend
echo -e "${GREEN}Starting backend server...${NC}"
cd backend
go run cmd/server/main.go &
BACKEND_PID=$!

# Wait for backend to start
sleep 3

# Start frontend
echo -e "${GREEN}Starting frontend server...${NC}"
cd ../frontend
npm start &
FRONTEND_PID=$!

echo -e "${GREEN}Application started!${NC}"
echo "Frontend: http://localhost:3000"
echo "Backend: http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for interrupt
trap "echo -e '\n${YELLOW}Shutting down...${NC}'; kill $BACKEND_PID $FRONTEND_PID; exit" INT
wait