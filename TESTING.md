# D&D Game - Testing Guide

This guide covers how to run all tests in the D&D Game application, including unit tests, integration tests, and end-to-end tests.

## Quick Start

```bash
# Backend tests
cd backend && go test ./...

# Frontend tests  
cd frontend && npm test

# E2E tests
cd frontend && npm run test:e2e
```

## Table of Contents
- [Backend Tests](#backend-tests)
- [Frontend Tests](#frontend-tests)
- [End-to-End Tests](#end-to-end-tests)
- [Running All Tests](#running-all-tests)
- [Test Coverage](#test-coverage)
- [CI/CD Testing](#cicd-testing)

## Backend Tests

### Running Backend Tests

Navigate to the backend directory:
```bash
cd backend
```

#### Run all tests:
```bash
go test ./...
```

#### Run tests with verbose output:
```bash
go test -v ./...
```

#### Run tests with coverage:
```bash
go test -cover ./...
```

#### Generate coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### Run specific package tests:
```bash
# Service tests
go test ./internal/services/...

# Handler tests
go test ./internal/handlers/...

# Auth tests
go test ./internal/auth/...
```

#### Run tests with race detection:
```bash
go test -race ./...
```

#### Run benchmarks:
```bash
go test -bench=. ./...
```

### Backend Test Structure
- **Unit Tests**: `*_test.go` files alongside source files
- **Service Tests**: `/internal/services/*_test.go`
- **Handler Tests**: `/internal/handlers/*_test.go`
- **Integration Tests**: Files ending with `_integration_test.go`


## Frontend Tests

### Running Frontend Tests

Navigate to the frontend directory:
```bash
cd frontend
```

#### Run all tests:
```bash
npm test
```

#### Run tests in watch mode:
```bash
npm run test:watch
```

#### Run tests with coverage:
```bash
npm run test:coverage
```

#### Run tests with coverage threshold check:
```bash
npm run test:coverage:check
```

#### Run tests in CI mode:
```bash
npm run test:ci
```

#### Debug tests:
```bash
npm run test:debug
```

### Frontend Test Structure
- **Component Tests**: `/src/components/__tests__/`
- **Hook Tests**: `/src/hooks/__tests__/`
- **Redux Tests**: `/src/store/slices/__tests__/`
- **Utility Tests**: `/src/utils/__tests__/`

### Test File Patterns
- `*.test.ts` or `*.test.tsx` - Test files
- `*.spec.ts` or `*.spec.tsx` - Specification tests
- `__tests__/` directories - Test folders


## End-to-End Tests

### Setup E2E Tests

First, ensure Playwright browsers are installed:
```bash
cd frontend
npx playwright install
```

### Running E2E Tests

#### Run all E2E tests:
```bash
npm run test:e2e
```

#### Run E2E tests with UI mode (recommended for development):
```bash
npm run test:e2e:ui
```

#### Debug E2E tests:
```bash
npm run test:e2e:debug
```

#### Run E2E tests in headed mode (see browser):
```bash
npm run test:e2e:headed
```

#### View test report after run:
```bash
npm run test:e2e:report
```

### E2E Test Structure
- **Test Files**: `/e2e/tests/*.spec.ts`
- **Page Objects**: `/e2e/pages/*.ts`
- **Fixtures**: `/e2e/fixtures/*.ts`

### E2E Test Suites
1. **Authentication Flow** (`auth.spec.ts`)
   - User registration
   - Login/logout
   - Protected routes

2. **Character Creation** (`character-creation.spec.ts`)
   - Step-by-step wizard
   - Validation
   - Character persistence

3. **Game Session** (`game-session.spec.ts`)
   - Creating sessions
   - Joining sessions
   - Real-time chat
   - Multiplayer functionality

4. **Combat Encounters** (`combat-encounter.spec.ts`)
   - Combat initialization
   - Turn order
   - Actions and damage
   - Conditions

## Running All Tests

### Full Test Suite

From the project root:
```bash
# Run all backend tests
cd backend && go test ./... && cd ..

# Run all frontend tests
cd frontend && npm test && cd ..

# Run all E2E tests
cd frontend && npm run test:e2e && cd ..
```

### Using Make (if available)

From the project root:
```bash
# Run all tests
make test

# Run backend tests only
make test-backend

# Run frontend tests only
make test-frontend

# Run E2E tests only
make test-e2e
```

## Test Coverage

### Coverage Goals
- **Backend**: 80% coverage target
- **Frontend**: 70% coverage target
- **Critical paths**: 90%+ coverage recommended

### Viewing Coverage Reports

#### Backend Coverage:
```bash
cd backend
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Frontend Coverage:
```bash
cd frontend
npm run test:coverage
# Coverage report will be in frontend/coverage/lcov-report/index.html
```

### Coverage by Component

Current test coverage includes:
- ✅ Backend Services (User, Game Session, Inventory, Character, Combat)
- ✅ API Handlers (Auth, Character, Game, Inventory)
- ✅ Frontend Hooks (all custom hooks)
- ✅ Redux Slices (auth, character, combat, gameSession, dmTools, ui, websocket)
- ✅ E2E Critical User Journeys

## CI/CD Testing

### GitHub Actions

Tests run automatically on:
- Pull requests
- Pushes to main branch

### Local CI Simulation

Run tests as they would in CI:
```bash
# Backend
cd backend
go test -v -race -cover ./...

# Frontend
cd frontend
npm run test:ci
npm run test:e2e -- --reporter=list
```

## Troubleshooting

### Common Issues

1. **E2E Tests Failing - Browsers Not Installed**
   ```bash
   npx playwright install
   ```

2. **Port Already in Use**
   - Ensure no other services are running on ports 3000 (frontend) or 8080 (backend)

3. **Database Connection Issues**
   - Ensure PostgreSQL is running
   - Check `.env` configuration

4. **Slow Tests**
   - Run tests in parallel: `go test -parallel 4 ./...`
   - Use focused test runs during development

### Debug Tips

1. **Backend Tests**
   - Use `t.Logf()` for debug output
   - Run single test: `go test -run TestName`

2. **Frontend Tests**
   - Use `screen.debug()` to see component state
   - Run single test file: `npm test -- MyComponent.test.tsx`

3. **E2E Tests**
   - Use `--debug` flag for step-by-step execution
   - Take screenshots: Tests automatically capture screenshots on failure

## Best Practices

1. **Write tests alongside new features**
2. **Run tests before committing**
3. **Keep tests focused and independent**
4. **Use meaningful test descriptions**
5. **Mock external dependencies**
6. **Test both success and failure cases**
7. **Maintain test data generators**

## Test Maintenance

### Updating Tests
- Update tests when modifying features
- Remove obsolete tests
- Keep page objects synchronized with UI changes
- Update test data generators as needed

### Adding New Tests
1. Follow existing patterns
2. Use appropriate helpers and utilities
3. Add to relevant test suites
4. Ensure new tests run in CI

---

For more information about specific test implementations, see the test files in the respective directories.