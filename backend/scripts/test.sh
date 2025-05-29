#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BACKEND_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"

# Change to backend directory
cd "$BACKEND_DIR"

echo -e "${GREEN}Running D&D Game Backend Tests${NC}"
echo "=================================="

# Function to run tests
run_tests() {
    local test_type=$1
    local test_command=$2
    
    echo -e "\n${YELLOW}Running $test_type tests...${NC}"
    if eval "$test_command"; then
        echo -e "${GREEN}✓ $test_type tests passed${NC}"
        return 0
    else
        echo -e "${RED}✗ $test_type tests failed${NC}"
        return 1
    fi
}

# Check if specific test type is requested
TEST_TYPE=${1:-all}

case $TEST_TYPE in
    unit)
        run_tests "Unit" "go test -v -race -short ./..."
        ;;
    integration)
        run_tests "Integration" "go test -v -race -tags=integration ./..."
        ;;
    coverage)
        echo -e "\n${YELLOW}Generating test coverage report...${NC}"
        mkdir -p coverage
        go test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
        go tool cover -html=coverage/coverage.out -o coverage/coverage.html
        echo -e "${GREEN}Coverage report generated: coverage/coverage.html${NC}"
        ;;
    auth)
        run_tests "Auth" "go test -v ./internal/auth/..."
        ;;
    services)
        run_tests "Services" "go test -v -short ./internal/services/..."
        ;;
    handlers)
        run_tests "Handlers" "go test -v -short ./internal/handlers/..."
        ;;
    dice)
        run_tests "Dice" "go test -v ./pkg/dice/..."
        ;;
    all)
        failed=0
        
        # Run all test suites
        run_tests "Auth" "go test -v -race ./internal/auth/..." || failed=1
        run_tests "Dice" "go test -v -race ./pkg/dice/..." || failed=1
        run_tests "Unit" "go test -v -race -short ./..." || failed=1
        
        if [ $failed -eq 0 ]; then
            echo -e "\n${GREEN}All tests passed!${NC}"
        else
            echo -e "\n${RED}Some tests failed!${NC}"
            exit 1
        fi
        ;;
    *)
        echo "Usage: $0 [unit|integration|coverage|auth|services|handlers|dice|all]"
        echo ""
        echo "Options:"
        echo "  unit         Run unit tests only"
        echo "  integration  Run integration tests only"
        echo "  coverage     Generate test coverage report"
        echo "  auth         Run auth package tests"
        echo "  services     Run services package tests"
        echo "  handlers     Run handlers package tests"
        echo "  dice         Run dice package tests"
        echo "  all          Run all tests (default)"
        exit 1
        ;;
esac