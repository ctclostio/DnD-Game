import { test, expect, testData } from '../fixtures/base';

test.describe('Authentication Flow', () => {
  test.describe('User Registration', () => {
    test('should register a new user successfully', async ({ registerPage, dashboardPage, page }) => {
      const user = testData.generateUser();
      
      await registerPage.goto();
      
      // Fill registration form
      await registerPage.register(user.username, user.email, user.password);
      
      // Should redirect to dashboard after successful registration
      await page.waitForURL(/\/dashboard/);
      await dashboardPage.waitForLoad();
      
      // Verify welcome message
      await expect(dashboardPage.welcomeMessage).toBeVisible();
      await expect(dashboardPage.welcomeMessage).toContainText(user.username);
    });

    test('should show error for duplicate username', async ({ registerPage }) => {
      const user = testData.generateUser();
      
      await registerPage.goto();
      
      // Register first time
      await registerPage.register(user.username, user.email, user.password);
      
      // Try to register again with same username
      const newEmail = `different_${user.email}`;
      await registerPage.goto();
      await registerPage.register(user.username, newEmail, user.password);
      
      // Should show error
      const hasError = await registerPage.expectError('username already exists');
      expect(hasError).toBeTruthy();
    });

    test('should validate password requirements', async ({ registerPage }) => {
      const user = testData.generateUser();
      
      await registerPage.goto();
      
      // Try weak password
      await registerPage.register(user.username, user.email, 'weak');
      
      // Should show validation error
      const hasError = await registerPage.expectError('password');
      expect(hasError).toBeTruthy();
    });

    test('should validate email format', async ({ registerPage }) => {
      const user = testData.generateUser();
      
      await registerPage.goto();
      
      // Try invalid email
      await registerPage.register(user.username, 'invalid-email', user.password);
      
      // Should show validation error
      const hasError = await registerPage.expectError('email');
      expect(hasError).toBeTruthy();
    });

    test('should navigate to login page', async ({ registerPage, page }) => {
      await registerPage.goto();
      await registerPage.navigateToLogin();
      
      await page.waitForURL(/\/login/);
      expect(page.url()).toContain('/login');
    });
  });

  test.describe('User Login', () => {
    let testUser: ReturnType<typeof testData.generateUser>;

    test.beforeEach(async ({ registerPage, page }) => {
      // Create a test user
      testUser = testData.generateUser();
      await registerPage.goto();
      await registerPage.register(testUser.username, testUser.email, testUser.password);
      
      // Logout
      await page.context().clearCookies();
      await page.evaluate(() => localStorage.clear());
    });

    test('should login with valid credentials', async ({ loginPage, dashboardPage, page }) => {
      await loginPage.goto();
      
      // Login
      await loginPage.login(testUser.username, testUser.password);
      
      // Should redirect to dashboard
      await page.waitForURL(/\/dashboard/);
      await dashboardPage.waitForLoad();
      
      // Verify logged in
      await expect(dashboardPage.welcomeMessage).toContainText(testUser.username);
    });

    test('should remember user when checkbox is checked', async ({ loginPage, page }) => {
      await loginPage.goto();
      
      // Login with remember me
      await loginPage.login(testUser.username, testUser.password, true);
      
      // Close and reopen browser context
      await page.context().close();
      const newContext = await page.context().browser()!.newContext();
      const newPage = await newContext.newPage();
      
      // Should still be logged in
      await newPage.goto('/dashboard');
      await expect(newPage.getByText(testUser.username)).toBeVisible();
      
      await newContext.close();
    });

    test('should show error for invalid credentials', async ({ loginPage }) => {
      await loginPage.goto();
      
      // Try invalid password
      await loginPage.login(testUser.username, 'wrongpassword');
      
      // Should show error
      const hasError = await loginPage.expectError('Invalid credentials');
      expect(hasError).toBeTruthy();
    });

    test('should show error for non-existent user', async ({ loginPage }) => {
      await loginPage.goto();
      
      // Try non-existent user
      await loginPage.login('nonexistentuser', 'password');
      
      // Should show error
      const hasError = await loginPage.expectError('Invalid credentials');
      expect(hasError).toBeTruthy();
    });

    test('should navigate to register page', async ({ loginPage, page }) => {
      await loginPage.goto();
      await loginPage.navigateToRegister();
      
      await page.waitForURL(/\/register/);
      expect(page.url()).toContain('/register');
    });
  });

  test.describe('Logout', () => {
    let testUser: ReturnType<typeof testData.generateUser>;

    test.beforeEach(async ({ registerPage, page }) => {
      // Create and login test user
      testUser = testData.generateUser();
      await registerPage.goto();
      await registerPage.register(testUser.username, testUser.email, testUser.password);
      await page.waitForURL(/\/dashboard/);
    });

    test('should logout successfully', async ({ dashboardPage, page }) => {
      await dashboardPage.logout();
      
      // Should redirect to login
      await page.waitForURL(/\/login/);
      
      // Try to access protected route
      await page.goto('/dashboard');
      
      // Should redirect back to login
      await page.waitForURL(/\/login/);
    });

    test('should clear session data on logout', async ({ dashboardPage, page }) => {
      await dashboardPage.logout();
      
      // Check localStorage is cleared
      const token = await page.evaluate(() => localStorage.getItem('token'));
      expect(token).toBeNull();
      
      // Check cookies are cleared
      const cookies = await page.context().cookies();
      const sessionCookie = cookies.find(c => c.name === 'session');
      expect(sessionCookie).toBeUndefined();
    });
  });

  test.describe('Protected Routes', () => {
    test('should redirect to login when accessing protected routes', async ({ page }) => {
      // Clear any existing session
      await page.context().clearCookies();
      await page.evaluate(() => localStorage.clear());
      
      // Try to access protected routes
      const protectedRoutes = [
        '/dashboard',
        '/character/new',
        '/session/create',
        '/combat',
      ];
      
      for (const route of protectedRoutes) {
        await page.goto(route);
        await page.waitForURL(/\/login/);
        expect(page.url()).toContain('/login');
      }
    });

    test('should redirect to original route after login', async ({ loginPage, page, registerPage }) => {
      // Create test user
      const user = testData.generateUser();
      await registerPage.goto();
      await registerPage.register(user.username, user.email, user.password);
      
      // Logout
      await page.context().clearCookies();
      await page.evaluate(() => localStorage.clear());
      
      // Try to access character builder
      await page.goto('/character/new');
      
      // Should redirect to login
      await page.waitForURL(/\/login/);
      
      // Login
      await loginPage.login(user.username, user.password);
      
      // Should redirect back to character builder
      await page.waitForURL(/\/character\/new/);
      expect(page.url()).toContain('/character/new');
    });
  });
});