// Mock localStorage before any imports
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

import { configureStore } from '@reduxjs/toolkit';
import authReducer, { login, register, logout, clearError } from '../authSlice';
import * as authService from '../../../services/auth';

// Mock auth service
jest.mock('../../../services/auth');

// Test constants - these are not real passwords
const TEST_PASSWORD = 'testPass123'; // NOSONAR - test password for unit tests
const TEST_PASSWORD_2 = 'testP@ss'; // NOSONAR - test password for unit tests  
const TEST_PASSWORD_3 = '123456'; // NOSONAR - test password for unit tests

describe('authSlice', () => {
  let store: ReturnType<typeof configureStore>;

  beforeEach(() => {
    jest.clearAllMocks();
    localStorageMock.getItem.mockReturnValue(null);
    
    store = configureStore({
      reducer: {
        auth: authReducer,
      },
    });
  });

  describe('initial state', () => {
    it('should initialize with no user when localStorage is empty', () => {
      const state = store.getState().auth;
      
      expect(state).toEqual({
        user: null,
        token: null,
        isLoading: false,
        error: null,
      });
    });

    it('should initialize with user from localStorage', () => {
      const savedUser = { id: '1', username: 'testuser', email: 'test@example.com', role: 'player' };
      const savedToken = 'saved-token';
      
      // Set up localStorage before requiring the module
      const mockGetItem = (key: string) => {
        if (key === 'user') return JSON.stringify(savedUser);
        if (key === 'token') return savedToken;
        return null;
      };
      localStorageMock.getItem.mockImplementation(mockGetItem);

      // Clear module cache and re-import
      jest.resetModules();
      const authSlice = require('../authSlice').default;

      // Create new store to test initialization
      const newStore = configureStore({
        reducer: {
          auth: authSlice,
        },
      });

      const state = newStore.getState().auth;
      expect(state.user).toEqual(savedUser);
      expect(state.token).toBe(savedToken);
    });
  });

  describe('login', () => {
    const loginCredentials = { username: 'testuser', password: TEST_PASSWORD };
    const mockResponse = {
      user: { id: '1', username: 'testuser', email: 'test@example.com', role: 'player' as const },
      token: 'mock-token',
    };

    it('should handle successful login', async () => {
      (authService.login as jest.Mock).mockResolvedValue(mockResponse);

      await store.dispatch(login(loginCredentials));

      const state = store.getState().auth;
      expect(state.user).toEqual(mockResponse.user);
      expect(state.token).toBe(mockResponse.token);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();

      // Check localStorage
      expect(localStorageMock.setItem).toHaveBeenCalledWith('user', JSON.stringify(mockResponse.user));
      expect(localStorageMock.setItem).toHaveBeenCalledWith('token', mockResponse.token);
    });

    it('should handle login pending state', () => {
      const delayedPromise = new Promise(resolve => {
        setTimeout(() => resolve(mockResponse), 100);
      });
      (authService.login as jest.Mock).mockReturnValue(delayedPromise);

      store.dispatch(login(loginCredentials));

      const state = store.getState().auth;
      expect(state.isLoading).toBe(true);
      expect(state.error).toBeNull();
    });

    it('should handle login failure', async () => {
      const errorMessage = 'Invalid credentials';
      (authService.login as jest.Mock).mockRejectedValue(new Error(errorMessage));

      await store.dispatch(login(loginCredentials));

      const state = store.getState().auth;
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();
      expect(state.isLoading).toBe(false);
      expect(state.error).toBe(errorMessage);

      // localStorage should not be updated
      expect(localStorageMock.setItem).not.toHaveBeenCalled();
    });

    it('should handle login failure with no error message', async () => {
      (authService.login as jest.Mock).mockRejectedValue({});

      await store.dispatch(login(loginCredentials));

      const state = store.getState().auth;
      expect(state.error).toBe('Login failed');
    });
  });

  describe('register', () => {
    const registerData = { 
      username: 'newuser', 
      email: 'new@example.com', 
      password: TEST_PASSWORD 
    };
    const mockResponse = {
      user: { id: '2', username: 'newuser', email: 'new@example.com', role: 'player' as const },
      token: 'new-token',
    };

    it('should handle successful registration', async () => {
      (authService.register as jest.Mock).mockResolvedValue(mockResponse);

      await store.dispatch(register(registerData));

      const state = store.getState().auth;
      expect(state.user).toEqual(mockResponse.user);
      expect(state.token).toBe(mockResponse.token);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();

      // Check localStorage
      expect(localStorageMock.setItem).toHaveBeenCalledWith('user', JSON.stringify(mockResponse.user));
      expect(localStorageMock.setItem).toHaveBeenCalledWith('token', mockResponse.token);
    });

    it('should handle registration pending state', () => {
      const delayedPromise = new Promise(resolve => {
        setTimeout(() => resolve(mockResponse), 100);
      });
      (authService.register as jest.Mock).mockReturnValue(delayedPromise);

      store.dispatch(register(registerData));

      const state = store.getState().auth;
      expect(state.isLoading).toBe(true);
      expect(state.error).toBeNull();
    });

    it('should handle registration failure', async () => {
      const errorMessage = 'Username already exists';
      (authService.register as jest.Mock).mockRejectedValue(new Error(errorMessage));

      await store.dispatch(register(registerData));

      const state = store.getState().auth;
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();
      expect(state.isLoading).toBe(false);
      expect(state.error).toBe(errorMessage);
    });

    it('should handle registration failure with no error message', async () => {
      (authService.register as jest.Mock).mockRejectedValue({});

      await store.dispatch(register(registerData));

      const state = store.getState().auth;
      expect(state.error).toBe('Registration failed');
    });
  });

  describe('logout', () => {
    it('should handle successful logout', async () => {
      // Set initial logged-in state
      const initialUser = { id: '1', username: 'testuser', email: 'test@example.com', role: 'player' as const };
      await store.dispatch(login({
        username: 'testuser',
        password: TEST_PASSWORD,
      }));
      (authService.login as jest.Mock).mockResolvedValue({
        user: initialUser,
        token: 'test-token',
      });

      // Mock logout
      (authService.logout as jest.Mock).mockResolvedValue(undefined);

      await store.dispatch(logout());

      const state = store.getState().auth;
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();

      // Check localStorage
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('user');
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('token');
    });

    it('should call logout service', async () => {
      (authService.logout as jest.Mock).mockResolvedValue(undefined);

      await store.dispatch(logout());

      expect(authService.logout).toHaveBeenCalled();
    });

    it('should handle logout even if service fails', async () => {
      (authService.logout as jest.Mock).mockRejectedValue(new Error('Logout failed'));

      // Should not throw
      await expect(store.dispatch(logout())).resolves.toBeDefined();

      // State should still be cleared
      const state = store.getState().auth;
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();
    });
  });

  describe('clearError', () => {
    it('should clear error state', async () => {
      // First, create an error
      (authService.login as jest.Mock).mockRejectedValue(new Error('Test error'));
      await store.dispatch(login({ username: 'test', password: TEST_PASSWORD_2 }));

      let state = store.getState().auth;
      expect(state.error).toBe('Test error');

      // Clear error
      store.dispatch(clearError());

      state = store.getState().auth;
      expect(state.error).toBeNull();
    });
  });

  describe('role-based user types', () => {
    it('should handle player role', async () => {
      const playerResponse = {
        user: { id: '1', username: 'player1', email: 'player@example.com', role: 'player' as const },
        token: 'player-token',
      };

      (authService.login as jest.Mock).mockResolvedValue(playerResponse);
      await store.dispatch(login({ username: 'player1', password: TEST_PASSWORD_2 }));

      const state = store.getState().auth;
      expect(state.user?.role).toBe('player');
    });

    it('should handle dm role', async () => {
      const dmResponse = {
        user: { id: '2', username: 'dm1', email: 'dm@example.com', role: 'dm' as const },
        token: 'dm-token',
      };

      (authService.login as jest.Mock).mockResolvedValue(dmResponse);
      await store.dispatch(login({ username: 'dm1', password: TEST_PASSWORD_2 }));

      const state = store.getState().auth;
      expect(state.user?.role).toBe('dm');
    });

    it('should handle admin role', async () => {
      const adminResponse = {
        user: { id: '3', username: 'admin1', email: 'admin@example.com', role: 'admin' as const },
        token: 'admin-token',
      };

      (authService.login as jest.Mock).mockResolvedValue(adminResponse);
      await store.dispatch(login({ username: 'admin1', password: TEST_PASSWORD_2 }));

      const state = store.getState().auth;
      expect(state.user?.role).toBe('admin');
    });
  });

  describe('concurrent actions', () => {
    it('should handle multiple login attempts', async () => {
      const response1 = {
        user: { id: '1', username: 'user1', email: 'user1@example.com', role: 'player' as const },
        token: 'token1',
      };
      const response2 = {
        user: { id: '2', username: 'user2', email: 'user2@example.com', role: 'player' as const },
        token: 'token2',
      };

      (authService.login as jest.Mock)
        .mockResolvedValueOnce(response1)
        .mockResolvedValueOnce(response2);

      // Dispatch multiple logins
      const promise1 = store.dispatch(login({ username: 'user1', password: TEST_PASSWORD }));
      const promise2 = store.dispatch(login({ username: 'user2', password: TEST_PASSWORD_2 }));

      await Promise.all([promise1, promise2]);

      // Should have the last successful login
      const state = store.getState().auth;
      expect(state.user?.username).toBe('user2');
      expect(state.token).toBe('token2');
    });
  });
});