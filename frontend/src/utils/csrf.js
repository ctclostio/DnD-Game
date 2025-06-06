/**
 * CSRF token management utilities
 */

/**
 * Get CSRF token from cookie
 * @returns {string|null} The CSRF token or null if not found
 */
export function getCSRFToken() {
    const name = 'csrf_token=';
    const decodedCookie = decodeURIComponent(document.cookie);
    const cookies = decodedCookie.split(';');
    
    for (let cookie of cookies) {
        cookie = cookie.trim();
        if (cookie.indexOf(name) === 0) {
            return cookie.substring(name.length);
        }
    }
    
    return null;
}

/**
 * Add CSRF token to request headers
 * @param {Headers|Object} headers - The headers object to modify
 * @returns {Headers|Object} The modified headers
 */
export function addCSRFHeader(headers = {}) {
    const token = getCSRFToken();
    
    if (token) {
        if (headers instanceof Headers) {
            headers.set('X-CSRF-Token', token);
        } else {
            headers['X-CSRF-Token'] = token;
        }
    }
    
    return headers;
}

/**
 * Fetch with CSRF protection
 * @param {string} url - The URL to fetch
 * @param {Object} options - Fetch options
 * @returns {Promise<Response>} The fetch response
 */
export async function fetchWithCSRF(url, options = {}) {
    // Only add CSRF for state-changing methods
    const method = options.method || 'GET';
    const needsCSRF = ['POST', 'PUT', 'DELETE', 'PATCH'].includes(method.toUpperCase());
    
    if (needsCSRF) {
        options.headers = addCSRFHeader(options.headers || {});
    }
    
    // Ensure cookies are sent
    options.credentials = options.credentials || 'same-origin';
    
    const response = await fetch(url, options);
    
    // If CSRF token is invalid, try to get a new one
    if (response.status === 403 && needsCSRF) {
        const text = await response.text();
        if (text.includes('CSRF')) {
            // Get a new token by making a GET request
            await fetch('/api/v1/csrf-token', { credentials: 'same-origin' });
            
            // Retry the original request
            options.headers = addCSRFHeader(options.headers || {});
            return fetch(url, options);
        }
    }
    
    return response;
}