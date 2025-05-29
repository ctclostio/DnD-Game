#!/bin/bash

echo "Starting D&D Game Application..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if PostgreSQL is running
if ! docker ps | grep -q dnd-postgres; then
    echo -e "${YELLOW}Starting PostgreSQL database...${NC}"
    docker run --name dnd-postgres \
        -e POSTGRES_USER=dndgame \
        -e POSTGRES_PASSWORD=dndgamepass \
        -e POSTGRES_DB=dndgame \
        -p 5432:5432 \
        -d postgres:16-alpine
    
    # Wait for database to be ready
    echo "Waiting for database to be ready..."
    sleep 5
fi

# Start backend
echo -e "${GREEN}Starting backend server...${NC}"
cd backend
# Export environment variables from .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi
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