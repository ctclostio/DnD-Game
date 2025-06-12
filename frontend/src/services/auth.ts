// Authentication service for handling JWT tokens and auth API calls

import { fetchWithCSRF } from '../utils/csrf';

const API_BASE_URL = '/api/v1';

interface User {
  id: string;
  username: string;
  email: string;
  role: 'player' | 'dm' | 'admin';
}

interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

interface LoginRequest {
  username: string;
  password: string;
}

interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

class AuthService {
  private token: string | null;
  private refreshToken: string | null;
  private user: User | null;

  constructor() {
    this.token = localStorage.getItem('access_token');
    this.refreshToken = localStorage.getItem('refresh_token');
    this.user = JSON.parse(localStorage.getItem('user') || 'null');
  }

  // Save auth data to localStorage
  private saveAuthData(authResponse: AuthResponse): void {
    this.token = authResponse.access_token;
    this.refreshToken = authResponse.refresh_token;
    this.user = authResponse.user;

    localStorage.setItem('access_token', authResponse.access_token);
    localStorage.setItem('refresh_token', authResponse.refresh_token);
    localStorage.setItem('user', JSON.stringify(authResponse.user));
    localStorage.setItem('token_expires_at', String(Date.now() + (authResponse.expires_in * 1000)));
  }

  // Clear auth data
  clearAuthData(): void {
    this.token = null;
    this.refreshToken = null;
    this.user = null;

    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
    localStorage.removeItem('token_expires_at');
  }

  // Check if user is authenticated
  isAuthenticated(): boolean {
    return !!this.token;
  }

  // Get current user
  getCurrentUser(): User | null {
    return this.user;
  }

  // Get auth headers
  getAuthHeaders(): HeadersInit {
    return {
      'Authorization': `Bearer ${this.token}`,
      'Content-Type': 'application/json'
    };
  }

  // Register new user
  async register(username: string, email: string, password: string): Promise<AuthResponse> {
    const response = await fetchWithCSRF(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ username, email, password })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Registration failed');
    }

    const data: AuthResponse = await response.json();
    this.saveAuthData(data);
    return data;
  }

  // Login user
  async login(username: string, password: string): Promise<AuthResponse> {
    const response = await fetchWithCSRF(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ username, password })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Login failed');
    }

    const data: AuthResponse = await response.json();
    this.saveAuthData(data);
    return data;
  }

  // Logout user
  async logout(): Promise<void> {
    if (this.token) {
      try {
        await fetchWithCSRF(`${API_BASE_URL}/auth/logout`, {
          method: 'POST',
          headers: this.getAuthHeaders()
        });
      } catch (error) {
        console.error('Logout error:', error);
      }
    }
    
    this.clearAuthData();
  }

  // Refresh access token
  async refreshAccessToken(): Promise<string | null> {
    if (!this.refreshToken) {
      return null;
    }

    try {
      const response = await fetchWithCSRF(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ refresh_token: this.refreshToken })
      });

      if (!response.ok) {
        throw new Error('Token refresh failed');
      }

      const data: AuthResponse = await response.json();
      this.saveAuthData(data);
      return data.access_token;
    } catch (error) {
      console.error('Token refresh error:', error);
      this.clearAuthData();
      return null;
    }
  }

  // Check if token needs refresh
  isTokenExpired(): boolean {
    const expiresAt = localStorage.getItem('token_expires_at');
    if (!expiresAt) return true;
    
    const expiryTime = parseInt(expiresAt);
    const currentTime = Date.now();
    const bufferTime = 60 * 1000; // 1 minute buffer
    
    return currentTime > (expiryTime - bufferTime);
  }

  // Make authenticated request with auto token refresh
  async makeAuthenticatedRequest(url: string, options: RequestInit = {}): Promise<Response> {
    // Check if token needs refresh
    if (this.isTokenExpired()) {
      await this.refreshAccessToken();
    }

    // Make request with auth headers
    const response = await fetchWithCSRF(url, {
      ...options,
      headers: {
        ...this.getAuthHeaders(),
        ...options.headers
      }
    });

    // If unauthorized, try refreshing token and retry once
    if (response.status === 401 && this.refreshToken) {
      const newToken = await this.refreshAccessToken();
      if (newToken) {
        return await fetchWithCSRF(url, {
          ...options,
          headers: {
            ...this.getAuthHeaders(),
            ...options.headers
          }
        });
      }
    }

    return response;
  }

  // Get current user role
  getUserRole(): string | null {
    return this.user?.role || null;
  }

  // Check if user is DM
  isDM(): boolean {
    return this.user?.role === 'dm' || this.user?.role === 'admin';
  }

  // Check if user is admin
  isAdmin(): boolean {
    return this.user?.role === 'admin';
  }
}

// Create singleton instance
const authService = new AuthService();

export default authService;

// Export types for use in other modules
export type { User, AuthResponse, LoginRequest, RegisterRequest };

// Export named functions for compatibility with authSlice
export const login = (username: string, password: string) => authService.login(username, password);
export const register = (username: string, email: string, password: string) => authService.register(username, email, password);
export const logout = () => authService.logout();