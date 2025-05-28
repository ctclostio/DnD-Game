#!/bin/bash

echo "Starting D&D Game Application (Test Mode)..."

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting backend server...${NC}"
cd backend
go run cmd/server/main.go &
BACKEND_PID=$!

# Wait for backend to start
sleep 3

echo -e "${YELLOW}Starting frontend server...${NC}"
cd ../frontend
npm start &
FRONTEND_PID=$!

echo -e "${GREEN}Application starting...${NC}"
echo -e "${GREEN}Frontend will open at: http://localhost:3000${NC}"
echo -e "${GREEN}Backend API at: http://localhost:8080${NC}"
echo ""
echo "Press Ctrl+C to stop all services"

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Shutting down...${NC}"
    kill $BACKEND_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    exit
}

# Set trap for cleanup
trap cleanup INT TERM

# Wait for processes
wait