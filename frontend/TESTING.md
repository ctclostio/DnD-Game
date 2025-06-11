# Frontend Testing Guide

## Overview
The frontend uses Jest and React Testing Library for unit and integration testing.

## Running Tests

### Using npm scripts
```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run tests with coverage threshold check
npm run test:coverage:check

# Run tests in CI mode
npm run test:ci

# Debug tests
npm run test:debug
```

### Using the test runner script
```bash
# Run all tests
./test-runner.sh all

# Run tests in watch mode
./test-runner.sh watch

# Run tests with coverage
./test-runner.sh coverage

# Run tests for a specific file
./test-runner.sh file src/components/CharacterBuilder/__tests__/CharacterBuilder.test.tsx

# Run tests matching a pattern
./test-runner.sh pattern "useWebSocket"

# Update snapshots
./test-runner.sh update

# Show help
./test-runner.sh help
```

## Test Structure

### Unit Tests
- Located next to the component/module being tested
- Use `__tests__` folder or `.test.ts(x)` suffix
- Focus on isolated component behavior

### Integration Tests
- Test component interactions
- Test Redux store integration
- Test API integration

## Testing Guidelines

### Component Tests
```tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import MyComponent from './MyComponent';

describe('MyComponent', () => {
  it('should render correctly', () => {
    const store = configureStore({ reducer: {} });
    render(
      <Provider store={store}>
        <MyComponent />
      </Provider>
    );
    
    expect(screen.getByText('Expected Text')).toBeInTheDocument();
  });
});
```

### Hook Tests
```tsx
import { renderHook, act } from '@testing-library/react';
import useMyHook from './useMyHook';

describe('useMyHook', () => {
  it('should update state', () => {
    const { result } = renderHook(() => useMyHook());
    
    act(() => {
      result.current.updateValue('new value');
    });
    
    expect(result.current.value).toBe('new value');
  });
});
```

### Redux Slice Tests
```ts
import reducer, { action } from './mySlice';

describe('mySlice', () => {
  it('should handle action', () => {
    const initialState = { value: 0 };
    const newState = reducer(initialState, action(5));
    expect(newState.value).toBe(5);
  });
});
```

## Mocking

### API Mocking
```ts
jest.mock('../services/api', () => ({
  fetchData: jest.fn().mockResolvedValue({ data: 'mocked' })
}));
```

### WebSocket Mocking
```ts
jest.mock('../services/websocket', () => ({
  getWebSocketService: jest.fn().mockReturnValue({
    connect: jest.fn(),
    disconnect: jest.fn(),
    sendMessage: jest.fn()
  })
}));
```

## Coverage Requirements
- Minimum 60% coverage for all metrics
- Target 70%+ for critical components
- Run `npm run test:coverage:check` to verify

## Debugging Tests

### VS Code
1. Add breakpoint in test or component
2. Run "Jest: Debug" from command palette
3. Or use the test runner: `npm run test:debug`

### Chrome DevTools
1. Run `npm run test:debug`
2. Open `chrome://inspect`
3. Click "inspect" on the Node process

## Common Issues

### Module Resolution
If you get module resolution errors:
```js
// jest.config.js
moduleNameMapper: {
  '^@/(.*)$': '<rootDir>/src/$1'
}
```

### TypeScript Errors
Ensure `tsconfig.json` includes:
```json
{
  "compilerOptions": {
    "jsx": "react",
    "esModuleInterop": true
  }
}
```

### React 18+ Issues
For React 18+ compatibility:
```js
// setupTests.js
import '@testing-library/jest-dom';
global.IS_REACT_ACT_ENVIRONMENT = true;
```

## Best Practices
1. Write tests alongside implementation
2. Use descriptive test names
3. Test user behavior, not implementation details
4. Keep tests isolated and independent
5. Use data-testid sparingly
6. Mock external dependencies
7. Test error states and edge cases