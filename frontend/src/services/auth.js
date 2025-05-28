// Authentication service for handling JWT tokens and auth API calls

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

class AuthService {
  constructor() {
    this.token = localStorage.getItem('access_token');
    this.refreshToken = localStorage.getItem('refresh_token');
    this.user = JSON.parse(localStorage.getItem('user') || 'null');
  }

  // Save auth data to localStorage
  saveAuthData(authResponse) {
    this.token = authResponse.access_token;
    this.refreshToken = authResponse.refresh_token;
    this.user = authResponse.user;

    localStorage.setItem('access_token', authResponse.access_token);
    localStorage.setItem('refresh_token', authResponse.refresh_token);
    localStorage.setItem('user', JSON.stringify(authResponse.user));
    localStorage.setItem('token_expires_at', Date.now() + (authResponse.expires_in * 1000));
  }

  // Clear auth data
  clearAuthData() {
    this.token = null;
    this.refreshToken = null;
    this.user = null;

    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
    localStorage.removeItem('token_expires_at');
  }

  // Check if user is authenticated
  isAuthenticated() {
    return !!this.token;
  }

  // Get current user
  getCurrentUser() {
    return this.user;
  }

  // Get auth headers
  getAuthHeaders() {
    return {
      'Authorization': `Bearer ${this.token}`,
      'Content-Type': 'application/json'
    };
  }

  // Register new user
  async register(username, email, password) {
    const response = await fetch(`${API_BASE_URL}/auth/register`, {
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

    const data = await response.json();
    this.saveAuthData(data);
    return data;
  }

  // Login user
  async login(username, password) {
    const response = await fetch(`${API_BASE_URL}/auth/login`, {
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

    const data = await response.json();
    this.saveAuthData(data);
    return data;
  }

  // Logout user
  async logout() {
    try {
      await fetch(`${API_BASE_URL}/auth/logout`, {
        method: 'POST',
        headers: this.getAuthHeaders()
      });
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      this.clearAuthData();
    }
  }

  // Refresh access token
  async refreshAccessToken() {
    if (!this.refreshToken) {
      throw new Error('No refresh token available');
    }

    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ refresh_token: this.refreshToken })
    });

    if (!response.ok) {
      this.clearAuthData();
      throw new Error('Token refresh failed');
    }

    const data = await response.json();
    this.saveAuthData(data);
    return data;
  }

  // Check if token needs refresh
  shouldRefreshToken() {
    const expiresAt = localStorage.getItem('token_expires_at');
    if (!expiresAt) return true;

    // Refresh if token expires in less than 5 minutes
    const fiveMinutes = 5 * 60 * 1000;
    return Date.now() > (parseInt(expiresAt) - fiveMinutes);
  }

  // Make authenticated API request with automatic token refresh
  async makeAuthenticatedRequest(url, options = {}) {
    // Check if token needs refresh
    if (this.shouldRefreshToken() && this.refreshToken) {
      try {
        await this.refreshAccessToken();
      } catch (error) {
        console.error('Token refresh failed:', error);
        throw error;
      }
    }

    // Make request with auth headers
    const response = await fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        ...this.getAuthHeaders()
      }
    });

    // If unauthorized, try to refresh token and retry
    if (response.status === 401 && this.refreshToken) {
      try {
        await this.refreshAccessToken();
        
        // Retry the request
        return await fetch(url, {
          ...options,
          headers: {
            ...options.headers,
            ...this.getAuthHeaders()
          }
        });
      } catch (error) {
        console.error('Token refresh failed:', error);
        this.clearAuthData();
        window.location.href = '/login';
        throw error;
      }
    }

    return response;
  }
}

// Export singleton instance
export default new AuthService();