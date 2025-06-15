import React from 'react';
import { renderHook, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { useErrorHandler, useFormErrorHandler, useApiErrorHandler } from '../useErrorHandler';
import uiSlice from '../../store/slices/uiSlice';

// Create a mock store
const createMockStore = () => {
  return configureStore({
    reducer: {
      ui: uiSlice,
    },
  });
};

// Wrapper component for Redux Provider
const createWrapper = (store: ReturnType<typeof createMockStore>) => {
  return ({ children }: { children: React.ReactNode }) => (
    <Provider store={store}>{children}</Provider>
  );
};

describe('useErrorHandler', () => {
  let store: ReturnType<typeof createMockStore>;
  let consoleErrorSpy: jest.SpyInstance;

  beforeEach(() => {
    store = createMockStore();
    consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation();
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
  });

  it('should handle Error objects', () => {
    const { result } = renderHook(() => useErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const error = new Error('Test error message');
    
    act(() => {
      result.current.handleError(error);
    });

    const state = store.getState();
    expect(state.ui.notifications).toHaveLength(1);
    expect(state.ui.notifications[0]).toMatchObject({
      type: 'error',
      message: 'Test error message',
    });
  });

  it('should handle string errors', () => {
    const { result } = renderHook(() => useErrorHandler(), {
      wrapper: createWrapper(store),
    });

    act(() => {
      result.current.handleError('String error');
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('String error');
  });

  it('should handle unknown error types', () => {
    const { result } = renderHook(() => useErrorHandler(), {
      wrapper: createWrapper(store),
    });

    act(() => {
      result.current.handleError({ custom: 'error' });
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('An unexpected error occurred');
  });

  it('should use fallback message', () => {
    const { result } = renderHook(() => useErrorHandler('Custom fallback'), {
      wrapper: createWrapper(store),
    });

    act(() => {
      result.current.handleError(null);
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('Custom fallback');
  });

  it('should log to console when enabled', () => {
    const { result } = renderHook(() => useErrorHandler(undefined, true), {
      wrapper: createWrapper(store),
    });

    const error = new Error('Console error');
    
    act(() => {
      result.current.handleError(error);
    });

    expect(consoleErrorSpy).toHaveBeenCalledWith('Error:', error);
  });

  it('should handle async errors', async () => {
    const { result } = renderHook(() => useErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const asyncOperation = async () => {
      throw new Error('Async error');
    };

    await act(async () => {
      await result.current.handleAsyncError(asyncOperation);
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('Async error');
  });

  it('should return result from successful async operation', async () => {
    const { result } = renderHook(() => useErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const asyncOperation = async () => {
      return 'Success result';
    };

    let operationResult;
    await act(async () => {
      operationResult = await result.current.handleAsyncError(asyncOperation);
    });

    expect(operationResult).toBe('Success result');
    const state = store.getState();
    expect(state.ui.notifications).toHaveLength(0);
  });
});

describe('useFormErrorHandler', () => {
  let store: ReturnType<typeof createMockStore>;

  beforeEach(() => {
    store = createMockStore();
  });

  it('should handle validation errors', () => {
    const { result } = renderHook(() => useFormErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const validationError = {
      response: {
        status: 400,
        data: {
          errors: {
            username: 'Username is required',
            email: 'Invalid email format',
          },
        },
      },
    };

    let fieldErrors;
    act(() => {
      fieldErrors = result.current.handleFormError(validationError);
    });

    expect(fieldErrors).toEqual({
      username: 'Username is required',
      email: 'Invalid email format',
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('Please check the form errors');
  });

  it('should handle non-validation errors', () => {
    const { result } = renderHook(() => useFormErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const serverError = {
      response: {
        status: 500,
        data: {
          message: 'Server error',
        },
      },
    };

    let fieldErrors;
    act(() => {
      fieldErrors = result.current.handleFormError(serverError);
    });

    expect(fieldErrors).toEqual({});
    
    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('Server error');
  });

  it('should handle errors without response', () => {
    const { result } = renderHook(() => useFormErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const error = new Error('Network error');

    let fieldErrors;
    act(() => {
      fieldErrors = result.current.handleFormError(error);
    });

    expect(fieldErrors).toEqual({});
    
    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('Network error');
  });

  it('should format field errors correctly', () => {
    const { result } = renderHook(() => useFormErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const validationError = {
      response: {
        status: 400,
        data: {
          errors: {
            'password.confirm': 'Passwords do not match',
            'profile.age': 'Must be a number',
          },
        },
      },
    };

    let fieldErrors;
    act(() => {
      fieldErrors = result.current.handleFormError(validationError);
    });

    expect(fieldErrors).toEqual({
      'password.confirm': 'Passwords do not match',
      'profile.age': 'Must be a number',
    });
  });
});

describe('useApiErrorHandler', () => {
  let store: ReturnType<typeof createMockStore>;

  beforeEach(() => {
    store = createMockStore();
  });

  it('should handle 401 unauthorized errors', () => {
    const { result } = renderHook(() => useApiErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const error = {
      response: {
        status: 401,
      },
    };

    act(() => {
      result.current.handleApiError(error);
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('Your session has expired. Please log in again.');
  });

  it('should handle 403 forbidden errors', () => {
    const { result } = renderHook(() => useApiErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const error = {
      response: {
        status: 403,
      },
    };

    act(() => {
      result.current.handleApiError(error);
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('You do not have permission to perform this action.');
  });

  it('should handle 404 not found errors', () => {
    const { result } = renderHook(() => useApiErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const error = {
      response: {
        status: 404,
      },
    };

    act(() => {
      result.current.handleApiError(error);
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('The requested resource was not found.');
  });

  it('should handle 500 server errors', () => {
    const { result } = renderHook(() => useApiErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const error = {
      response: {
        status: 500,
      },
    };

    act(() => {
      result.current.handleApiError(error);
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('A server error occurred. Please try again later.');
  });

  it('should use custom error message from response', () => {
    const { result } = renderHook(() => useApiErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const error = {
      response: {
        status: 400,
        data: {
          message: 'Custom error message from server',
        },
      },
    };

    act(() => {
      result.current.handleApiError(error);
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('Custom error message from server');
  });

  it('should handle network errors', () => {
    const { result } = renderHook(() => useApiErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const error = {
      request: {},
      message: 'Network Error',
    };

    act(() => {
      result.current.handleApiError(error);
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('Network error. Please check your connection.');
  });

  it('should handle timeout errors', () => {
    const { result } = renderHook(() => useApiErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const error = {
      code: 'ECONNABORTED',
      message: 'timeout of 5000ms exceeded',
    };

    act(() => {
      result.current.handleApiError(error);
    });

    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('Request timed out. Please try again.');
  });

  it('should wrap async API calls', async () => {
    const { result } = renderHook(() => useApiErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const mockApiCall = jest.fn().mockRejectedValue({
      response: {
        status: 403,
      },
    });

    await act(async () => {
      const wrappedCall = result.current.handleAsyncApiCall(mockApiCall);
      await wrappedCall();
    });

    expect(mockApiCall).toHaveBeenCalled();
    const state = store.getState();
    expect(state.ui.notifications[0].message).toBe('You do not have permission to perform this action.');
  });

  it('should return data from successful API calls', async () => {
    const { result } = renderHook(() => useApiErrorHandler(), {
      wrapper: createWrapper(store),
    });

    const mockData = { id: 1, name: 'Test' };
    const mockApiCall = jest.fn().mockResolvedValue(mockData);

    let apiResult;
    await act(async () => {
      const wrappedCall = result.current.handleAsyncApiCall(mockApiCall);
      apiResult = await wrappedCall();
    });

    expect(apiResult).toEqual(mockData);
    const state = store.getState();
    expect(state.ui.notifications).toHaveLength(0);
  });
});
