#!/bin/bash

# Frontend Test Runner Script
# Provides various testing options for the frontend

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to display usage
usage() {
    echo "Frontend Test Runner"
    echo "Usage: $0 [option]"
    echo ""
    echo "Options:"
    echo "  all          Run all tests"
    echo "  watch        Run tests in watch mode"
    echo "  coverage     Run tests with coverage report"
    echo "  check        Run tests with coverage threshold check"
    echo "  debug        Run tests in debug mode"
    echo "  file <path>  Run tests for a specific file"
    echo "  pattern <p>  Run tests matching a pattern"
    echo "  update       Update snapshots"
    echo "  failing      Show only failing tests"
    echo "  help         Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 all"
    echo "  $0 file src/components/CharacterBuilder/__tests__/CharacterBuilder.test.tsx"
    echo "  $0 pattern useWebSocket"
}

# Function to run tests
run_tests() {
    case "$1" in
        all)
            echo -e "${GREEN}Running all tests...${NC}"
            npm test -- --passWithNoTests
            ;;
        watch)
            echo -e "${GREEN}Running tests in watch mode...${NC}"
            npm test -- --watch
            ;;
        coverage)
            echo -e "${GREEN}Running tests with coverage...${NC}"
            npm test -- --coverage --passWithNoTests
            ;;
        check)
            echo -e "${GREEN}Running tests with coverage threshold check...${NC}"
            npm run test:coverage:check
            ;;
        debug)
            echo -e "${GREEN}Running tests in debug mode...${NC}"
            echo -e "${YELLOW}Open chrome://inspect in Chrome to debug${NC}"
            npm run test:debug
            ;;
        file)
            if [ -z "$2" ]; then
                echo -e "${RED}Error: Please provide a file path${NC}"
                usage
                exit 1
            fi
            echo -e "${GREEN}Running tests for file: $2${NC}"
            npm test -- "$2" --passWithNoTests
            ;;
        pattern)
            if [ -z "$2" ]; then
                echo -e "${RED}Error: Please provide a pattern${NC}"
                usage
                exit 1
            fi
            echo -e "${GREEN}Running tests matching pattern: $2${NC}"
            npm test -- -t "$2" --passWithNoTests
            ;;
        update)
            echo -e "${GREEN}Updating test snapshots...${NC}"
            npm test -- -u --passWithNoTests
            ;;
        failing)
            echo -e "${GREEN}Running only failing tests...${NC}"
            npm test -- --onlyFailures --passWithNoTests
            ;;
        help)
            usage
            ;;
        *)
            echo -e "${RED}Error: Unknown option '$1'${NC}"
            usage
            exit 1
            ;;
    esac
}

# Main execution
if [ $# -eq 0 ]; then
    usage
    exit 0
fi

# Navigate to frontend directory
cd "$(dirname "$0")"

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}node_modules not found. Running npm install...${NC}"
    npm install
fi

# Run the requested test command
run_tests "$@"