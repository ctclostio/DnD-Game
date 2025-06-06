# Testing Strategy and Guidelines

## Overview

This document outlines the testing strategy for the D&D Game application, including testing standards, tools, and best practices for both backend (Go) and frontend (React/TypeScript) code.

## Testing Philosophy

- **Test Pyramid**: We follow the testing pyramid approach with many unit tests, fewer integration tests, and selective E2E tests
- **Test Coverage**: Target 80% coverage for backend, 70% for frontend
- **TDD Encouraged**: Write tests before implementation when adding new features
- **Fast Feedback**: Tests should run quickly to encourage frequent execution

## Backend Testing (Go)

### Test Structure

```
backend/
├── internal/
│   ├── services/
│   │   ├── character.go
│   │   ├── character_test.go       # Unit tests
│   │   └── ...
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── auth_test.go            # Unit tests
│   │   └── auth_integration_test.go # Integration tests
│   └── testutil/
│       ├── fixtures.go             # Test data factories
│       ├── database.go             # Test database setup
│       └── mocks/                  # Mock implementations
```

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run with coverage
make test-coverage

# Run specific package tests
go test ./internal/services/...

# Run with race detector
go test -race ./...

# Run verbose
go test -v ./...
```

### Writing Unit Tests

```go
func TestCharacterService_Create(t *testing.T) {
    tests := []struct {
        name          string
        input         *models.Character
        setupMock     func(*mocks.MockCharacterRepository)
        expectedError string
        validate      func(*testing.T, *models.Character)
    }{
        {
            name: "successful creation",
            input: testutil.CharacterFixture(t, userID),
            setupMock: func(m *mocks.MockCharacterRepository) {
                m.On("Create", mock.Anything).Return(nil)
            },
            validate: func(t *testing.T, c *models.Character) {
                assert.NotEqual(t, uuid.Nil, c.ID)
                assert.Equal(t, 10, c.HitPoints)
            },
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Writing Integration Tests

```go
func TestAuthHandler_Login_Integration(t *testing.T) {
    // Skip if running short tests
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Set up test database
    db := testutil.SetupTestDB(t)
    defer db.Close()

    // Create test server
    server := setupTestServer(db)

    // Test implementation...
}
```

### Test Fixtures and Helpers

```go
// Use fixtures for consistent test data
user := testutil.UserFixture(t)
character := testutil.CharacterFixture(t, user.ID)

// Use test builders for complex objects
character := testutil.NewCharacterBuilder().
    WithName("Aragorn").
    WithClass("Ranger").
    WithLevel(10).
    Build()
```

### Mocking Guidelines

- Use interfaces for dependencies
- Generate mocks using `mockery` or write manual mocks
- Keep mocks in `testutil/mocks` directory
- Mock external services (database, APIs, etc.)

```go
// Example mock setup
mockRepo := new(mocks.MockCharacterRepository)
mockRepo.On("GetByID", characterID).Return(character, nil)

service := services.NewCharacterService(mockRepo)
// Test service methods...

mockRepo.AssertExpectations(t)
```

## Frontend Testing (React/TypeScript)

### Test Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── CharacterBuilder/
│   │   │   ├── CharacterBuilder.tsx
│   │   │   └── __tests__/
│   │   │       └── CharacterBuilder.test.tsx
│   │   └── __tests__/
│   │       └── CombatView.test.js
│   ├── hooks/
│   │   └── __tests__/
│   ├── services/
│   │   └── __tests__/
│   └── setupTests.js
```

### Running Tests

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run in watch mode
npm run test:watch

# Run specific test file
npm test CharacterBuilder

# Update snapshots
npm test -- -u
```

### Writing Component Tests

```tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { CharacterBuilder } from '../CharacterBuilder';

describe('CharacterBuilder', () => {
  it('should create character successfully', async () => {
    const user = userEvent.setup();
    render(<CharacterBuilder />);

    // Fill form
    await user.type(screen.getByLabelText('Character Name'), 'Thorin');
    await user.selectOptions(screen.getByLabelText('Race'), 'Dwarf');
    
    // Submit
    await user.click(screen.getByText('Create Character'));

    // Assert
    await waitFor(() => {
      expect(screen.getByText('Character created!')).toBeInTheDocument();
    });
  });
});
```

### Testing Hooks

```tsx
import { renderHook, act } from '@testing-library/react';
import { useDebounce } from '../useDebounce';

describe('useDebounce', () => {
  it('should debounce value changes', async () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'initial', delay: 500 } }
    );

    expect(result.current).toBe('initial');

    // Update value
    rerender({ value: 'updated', delay: 500 });
    
    // Value shouldn't change immediately
    expect(result.current).toBe('initial');

    // Wait for debounce
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 600));
    });

    expect(result.current).toBe('updated');
  });
});
```

### Testing API Services

```typescript
import { ApiService } from '../api';

// Mock fetch
global.fetch = jest.fn();

describe('ApiService', () => {
  beforeEach(() => {
    fetch.mockClear();
  });

  it('should create character', async () => {
    fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: '123', name: 'Thorin' }),
    });

    const api = new ApiService();
    const result = await api.createCharacter({ name: 'Thorin' });

    expect(fetch).toHaveBeenCalledWith(
      '/api/characters',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'Thorin' }),
      })
    );
    expect(result.id).toBe('123');
  });
});
```

## E2E Testing (Playwright)

### Test Structure

```
frontend/
├── e2e/
│   ├── auth.spec.ts
│   ├── character-creation.spec.ts
│   ├── combat.spec.ts
│   └── fixtures/
│       └── test-data.ts
```

### Writing E2E Tests

```typescript
import { test, expect } from '@playwright/test';

test.describe('Character Creation', () => {
  test('should create a new character', async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('[name="username"]', 'testuser');
    await page.fill('[name="password"]', 'password');
    await page.click('button[type="submit"]');

    // Navigate to character builder
    await page.click('text=Create Character');

    // Fill character details
    await page.fill('[name="characterName"]', 'Aragorn');
    await page.selectOption('[name="race"]', 'Human');
    await page.selectOption('[name="class"]', 'Ranger');

    // Complete creation
    await page.click('text=Create Character');

    // Verify success
    await expect(page.locator('text=Character created!')).toBeVisible();
  });
});
```

## Test Data Management

### Backend Test Data

- Use transactions for test isolation
- Clean up after each test
- Use factories for consistent test data

```go
func TestWithTransaction(t *testing.T) {
    db := testutil.SetupTestDB(t)
    
    // Start transaction
    tx, err := db.Begin()
    require.NoError(t, err)
    defer tx.Rollback() // Always rollback
    
    // Run test with transaction...
}
```

### Frontend Test Data

- Use mock service worker (MSW) for API mocking
- Keep test data close to tests
- Use factories for complex objects

```typescript
// test-factories.ts
export const createMockCharacter = (overrides = {}) => ({
  id: 'char-123',
  name: 'Test Character',
  level: 1,
  class: 'Fighter',
  ...overrides,
});
```

## CI/CD Integration

Tests run automatically on:
- Every push to main/develop branches
- Every pull request
- Can be triggered manually

### GitHub Actions Workflow

- Linting (backend & frontend)
- Unit tests with coverage
- Integration tests
- E2E tests
- Security scanning
- Build verification

## Best Practices

### General

1. **Test Naming**: Use descriptive names that explain what is being tested
   ```go
   // Good
   func TestCharacterService_Create_WithInvalidLevel_ReturnsError(t *testing.T)
   
   // Bad
   func TestCreate2(t *testing.T)
   ```

2. **Arrange-Act-Assert**: Structure tests clearly
   ```go
   // Arrange
   character := testutil.CharacterFixture(t)
   mockRepo.On("Create", character).Return(nil)
   
   // Act
   err := service.Create(character)
   
   // Assert
   assert.NoError(t, err)
   ```

3. **Test One Thing**: Each test should verify a single behavior

4. **Independent Tests**: Tests should not depend on each other

5. **Fast Tests**: Keep unit tests fast (< 100ms per test)

### Backend Specific

1. Use table-driven tests for multiple scenarios
2. Mock external dependencies
3. Use `t.Parallel()` for independent tests
4. Test error cases thoroughly
5. Use `testify` assertions for clarity

### Frontend Specific

1. Test user interactions, not implementation
2. Use Testing Library queries correctly
3. Avoid testing implementation details
4. Mock API calls appropriately
5. Test accessibility features

## Debugging Tests

### Backend

```bash
# Run single test with verbose output
go test -v -run TestCharacterService_Create

# Debug with delve
dlv test ./internal/services

# Show test coverage in browser
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Frontend

```bash
# Debug specific test
npm test -- --no-coverage CharacterBuilder --verbose

# Run tests in debug mode
node --inspect-brk node_modules/.bin/jest --runInBand

# Interactive mode
npm test -- --watch
```

## Performance Testing

### Load Testing

Use `k6` or `vegeta` for API load testing:

```javascript
// k6-script.js
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  vus: 10,
  duration: '30s',
};

export default function() {
  let response = http.get('http://localhost:8080/api/health');
  check(response, {
    'status is 200': (r) => r.status === 200,
  });
}
```

### Benchmark Tests

```go
func BenchmarkCharacterCreate(b *testing.B) {
    service := setupService()
    character := testutil.CharacterFixture(b)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.Create(character)
    }
}
```

## Troubleshooting

### Common Issues

1. **Flaky Tests**: 
   - Add proper waits for async operations
   - Mock time-dependent functionality
   - Ensure proper cleanup

2. **Slow Tests**:
   - Use test database pooling
   - Parallelize independent tests
   - Mock expensive operations

3. **Coverage Gaps**:
   - Run coverage reports regularly
   - Focus on critical paths
   - Test edge cases

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
- [Jest Documentation](https://jestjs.io/docs/getting-started)
- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Testify Assertions](https://github.com/stretchr/testify)