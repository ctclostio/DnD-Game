// Test configuration with secure defaults
export const testConfig = {
  // Use environment variable for test password, with a secure default
  // This avoids hard-coding passwords while maintaining test functionality
  defaultTestPassword: process.env.E2E_TEST_PASSWORD || generateSecureTestPassword(),
  
  // Other test configuration
  baseUrl: process.env.BASE_URL || 'http://localhost:3000',
  apiUrl: process.env.API_URL || 'http://localhost:8080',
  testTimeout: 30000,
};

// Generate a secure test password if not provided via environment
function generateSecureTestPassword(): string {
  // Use a combination of timestamp and random string for uniqueness
  const timestamp = Date.now().toString(36);
  const randomPart = Math.random().toString(36).substring(2, 8);
  return `Test_${timestamp}_${randomPart}!`;
}