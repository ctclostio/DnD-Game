.PHONY: all build run test clean install-deps

# Variables
BACKEND_DIR = backend
FRONTEND_DIR = frontend
SERVER_BINARY = $(BACKEND_DIR)/cmd/server/server
GO = go
NPM = npm

# Default target
all: install-deps build

# Install dependencies
install-deps: install-backend-deps install-frontend-deps

install-backend-deps:
	@echo "Installing backend dependencies..."
	@cd $(BACKEND_DIR) && $(GO) mod download

install-frontend-deps:
	@echo "Installing frontend dependencies..."
	@cd $(FRONTEND_DIR) && $(NPM) install

# Build targets
build: build-backend build-frontend

build-backend:
	@echo "Building backend..."
	@cd $(BACKEND_DIR)/cmd/server && $(GO) build -o server

build-frontend:
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && $(NPM) run build

# Run targets
run: build
	@echo "Starting server..."
	@cd $(BACKEND_DIR)/cmd/server && ./server

run-backend:
	@echo "Starting backend in development mode..."
	@cd $(BACKEND_DIR)/cmd/server && $(GO) run main.go

run-frontend:
	@echo "Starting frontend in development mode..."
	@cd $(FRONTEND_DIR) && $(NPM) start

# Development mode - run both frontend and backend
dev:
	@echo "Starting development servers..."
	@make -j 2 run-backend run-frontend

# Test targets
test: test-backend test-frontend

test-backend:
	@echo "Running backend tests..."
	@cd $(BACKEND_DIR) && $(GO) test -v ./...

test-frontend:
	@echo "Running frontend tests..."
	@cd $(FRONTEND_DIR) && $(NPM) test

# Lint targets
lint: lint-backend lint-frontend

lint-backend:
	@echo "Linting backend code..."
	@cd $(BACKEND_DIR) && golangci-lint run

lint-frontend:
	@echo "Linting frontend code..."
	@cd $(FRONTEND_DIR) && $(NPM) run lint

# Clean targets
clean: clean-backend clean-frontend

clean-backend:
	@echo "Cleaning backend..."
	@rm -f $(SERVER_BINARY)
	@cd $(BACKEND_DIR) && $(GO) clean

clean-frontend:
	@echo "Cleaning frontend..."
	@rm -rf $(FRONTEND_DIR)/build
	@rm -rf $(FRONTEND_DIR)/node_modules

# Docker targets
docker-build:
	@echo "Building Docker images..."
	@docker-compose build

docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-logs:
	@docker-compose logs -f

# Database targets (for future use)
db-migrate:
	@echo "Running database migrations..."
	# Add migration commands here when database is implemented

db-seed:
	@echo "Seeding database..."
	# Add seed commands here when database is implemented

# Utility targets
fmt:
	@echo "Formatting Go code..."
	@cd $(BACKEND_DIR) && $(GO) fmt ./...

vet:
	@echo "Running go vet..."
	@cd $(BACKEND_DIR) && $(GO) vet ./...

# Help target
help:
	@echo "Available targets:"
	@echo "  make all          - Install deps and build everything"
	@echo "  make build        - Build both backend and frontend"
	@echo "  make run          - Build and run the application"
	@echo "  make dev          - Run in development mode (hot reload)"
	@echo "  make test         - Run all tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-up    - Start with Docker"
	@echo "  make docker-down  - Stop Docker containers"
	@echo "  make help         - Show this help message"