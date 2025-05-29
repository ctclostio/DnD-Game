.PHONY: test test-unit test-integration test-coverage test-verbose clean-test help

# Variables
GOTEST := go test
GOTEST_FLAGS := -v -race
COVERAGE_DIR := ./coverage
COVERAGE_FILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html

# Default target
help:
	@echo "Available targets:"
	@echo "  test              - Run all tests"
	@echo "  test-unit         - Run unit tests only"
	@echo "  test-integration  - Run integration tests only"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  test-verbose      - Run tests with verbose output"
	@echo "  test-short        - Run short tests only"
	@echo "  clean-test        - Clean test cache and coverage files"
	@echo "  help              - Show this help message"

# Run all tests
test:
	@echo "Running all tests..."
	$(GOTEST) $(GOTEST_FLAGS) ./...

# Run unit tests only (excluding integration tests)
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) $(GOTEST_FLAGS) -short ./...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) $(GOTEST_FLAGS) -run Integration ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) $(GOTEST_FLAGS) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print "Total coverage: " $$3}'

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	$(GOTEST) -v -race -count=1 ./...

# Run only short tests
test-short:
	@echo "Running short tests..."
	$(GOTEST) $(GOTEST_FLAGS) -short -timeout 30s ./...

# Test specific package
test-pkg:
	@echo "Running tests for package $(PKG)..."
	$(GOTEST) $(GOTEST_FLAGS) ./$(PKG)/...

# Run tests with specific pattern
test-run:
	@echo "Running tests matching pattern $(RUN)..."
	$(GOTEST) $(GOTEST_FLAGS) -run $(RUN) ./...

# Benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	$(GOTEST) -bench=. -benchmem ./...

# Clean test cache and coverage files
clean-test:
	@echo "Cleaning test cache and coverage files..."
	@go clean -testcache
	@rm -rf $(COVERAGE_DIR)
	@echo "Test cache and coverage files cleaned"

# CI/CD oriented targets
test-ci:
	@echo "Running tests for CI..."
	$(GOTEST) -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...

# Run tests and generate JUnit XML report (requires go-junit-report)
test-junit:
	@echo "Running tests with JUnit output..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v ./... 2>&1 | go-junit-report > $(COVERAGE_DIR)/junit.xml

# Watch mode - requires entr or similar tool
test-watch:
	@echo "Running tests in watch mode..."
	@find . -name '*.go' | entr -c $(MAKE) test-unit