#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

# Load test environment
if [ -f .env.test ]; then
    export $(cat .env.test | grep -v '^#' | xargs)
fi

# Function to print colored output
print_status() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    print_status "Checking dependencies..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    
    if ! command -v make &> /dev/null; then
        print_error "Make is not installed"
        exit 1
    fi
    
    print_success "All dependencies are installed"
}

# Clean test cache
clean_test_cache() {
    print_status "Cleaning test cache..."
    go clean -testcache
    rm -rf coverage/
    print_success "Test cache cleaned"
}

# Run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    if make test-unit; then
        print_success "Unit tests passed"
        return 0
    else
        print_error "Unit tests failed"
        return 1
    fi
}

# Run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    # Check if database is available
    if ! pg_isready -h "$DATABASE_HOST" -p "$DATABASE_PORT" &> /dev/null; then
        print_warning "PostgreSQL is not running. Skipping integration tests."
        return 0
    fi
    
    if make test-integration; then
        print_success "Integration tests passed"
        return 0
    else
        print_error "Integration tests failed"
        return 1
    fi
}

# Run tests with coverage
run_coverage() {
    print_status "Running tests with coverage..."
    if make test-coverage; then
        print_success "Coverage report generated"
        return 0
    else
        print_error "Coverage generation failed"
        return 1
    fi
}

# Main test runner
main() {
    local test_type="${1:-all}"
    local exit_code=0
    
    echo "======================================"
    echo "D&D Game Backend Test Runner"
    echo "======================================"
    
    check_dependencies
    
    case "$test_type" in
        "unit")
            clean_test_cache
            run_unit_tests || exit_code=$?
            ;;
        "integration")
            clean_test_cache
            run_integration_tests || exit_code=$?
            ;;
        "coverage")
            clean_test_cache
            run_coverage || exit_code=$?
            ;;
        "all")
            clean_test_cache
            run_unit_tests || exit_code=$?
            run_integration_tests || exit_code=$?
            if [ $exit_code -eq 0 ]; then
                run_coverage || exit_code=$?
            fi
            ;;
        *)
            print_error "Unknown test type: $test_type"
            echo "Usage: $0 [unit|integration|coverage|all]"
            exit 1
            ;;
    esac
    
    echo "======================================"
    if [ $exit_code -eq 0 ]; then
        print_success "All tests completed successfully!"
    else
        print_error "Some tests failed. Please check the output above."
    fi
    echo "======================================"
    
    exit $exit_code
}

# Run main function
main "$@"