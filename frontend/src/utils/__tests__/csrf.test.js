import { getCSRFToken, addCSRFHeader, fetchWithCSRF } from '../csrf';

describe('CSRF utilities', () => {
  // Mock document.cookie
  let cookieValue = '';
  beforeEach(() => {
    // Clear cookies
    cookieValue = '';
    Object.defineProperty(document, 'cookie', {
      get: jest.fn(() => cookieValue),
      set: jest.fn((value) => {
        cookieValue = value;
      }),
      configurable: true,
    });
    // Reset fetch mock
    global.fetch.mockClear();
  });

  describe('getCSRFToken', () => {
    it('should extract CSRF token from cookie', () => {
      cookieValue = 'csrf_token=test-token-123';
      expect(getCSRFToken()).toBe('test-token-123');
    });

    it('should return null if no CSRF token cookie', () => {
      cookieValue = 'other_cookie=value';
      expect(getCSRFToken()).toBeNull();
    });

    it('should handle multiple cookies', () => {
      cookieValue = 'session=abc; csrf_token=my-token; user=123';
      expect(getCSRFToken()).toBe('my-token');
    });
  });

  describe('addCSRFHeader', () => {
    beforeEach(() => {
      cookieValue = 'csrf_token=test-csrf-token';
    });

    it('should add CSRF header to object', () => {
      const headers = {};
      const result = addCSRFHeader(headers);
      
      expect(result['X-CSRF-Token']).toBe('test-csrf-token');
    });

    it('should add CSRF header to Headers instance', () => {
      const headers = new Headers();
      const result = addCSRFHeader(headers);
      
      expect(result.get('X-CSRF-Token')).toBe('test-csrf-token');
    });

    it('should not add header if no token available', () => {
      cookieValue = '';
      const headers = {};
      const result = addCSRFHeader(headers);
      
      expect(result['X-CSRF-Token']).toBeUndefined();
    });

    it('should preserve existing headers', () => {
      const headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer token'
      };
      const result = addCSRFHeader(headers);
      
      expect(result['Content-Type']).toBe('application/json');
      expect(result['Authorization']).toBe('Bearer token');
      expect(result['X-CSRF-Token']).toBe('test-csrf-token');
    });
  });

  describe('fetchWithCSRF', () => {
    beforeEach(() => {
      cookieValue = 'csrf_token=csrf-123';
      global.fetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ success: true }),
        text: () => Promise.resolve('OK')
      });
    });

    it('should add CSRF header for POST requests', async () => {
      await fetchWithCSRF('/api/test', { method: 'POST' });
      
      expect(global.fetch).toHaveBeenCalledWith('/api/test', {
        method: 'POST',
        headers: { 'X-CSRF-Token': 'csrf-123' },
        credentials: 'same-origin'
      });
    });

    it('should add CSRF header for PUT requests', async () => {
      await fetchWithCSRF('/api/test', { method: 'PUT' });
      
      expect(global.fetch).toHaveBeenCalledWith('/api/test', {
        method: 'PUT',
        headers: { 'X-CSRF-Token': 'csrf-123' },
        credentials: 'same-origin'
      });
    });

    it('should add CSRF header for DELETE requests', async () => {
      await fetchWithCSRF('/api/test', { method: 'DELETE' });
      
      expect(global.fetch).toHaveBeenCalledWith('/api/test', {
        method: 'DELETE',
        headers: { 'X-CSRF-Token': 'csrf-123' },
        credentials: 'same-origin'
      });
    });

    it('should not add CSRF header for GET requests', async () => {
      await fetchWithCSRF('/api/test', { method: 'GET' });
      
      expect(global.fetch).toHaveBeenCalledWith('/api/test', {
        method: 'GET',
        credentials: 'same-origin'
      });
    });

    it('should retry on CSRF failure', async () => {
      // First call fails with 403
      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 403,
        text: () => Promise.resolve('CSRF token invalid')
      });
      
      // Token refresh succeeds
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200
      });
      
      // Retry succeeds
      global.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ success: true })
      });

      const response = await fetchWithCSRF('/api/test', { method: 'POST' });
      
      expect(global.fetch).toHaveBeenCalledTimes(3);
      expect(response.ok).toBe(true);
    });

    it('should preserve custom headers', async () => {
      await fetchWithCSRF('/api/test', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Custom-Header': 'value'
        }
      });
      
      expect(global.fetch).toHaveBeenCalledWith('/api/test', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Custom-Header': 'value',
          'X-CSRF-Token': 'csrf-123'
        },
        credentials: 'same-origin'
      });
    });
  });
});