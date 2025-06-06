import { Page, Locator } from '@playwright/test';

export class LoginPage {
  readonly page: Page;
  readonly usernameInput: Locator;
  readonly passwordInput: Locator;
  readonly loginButton: Locator;
  readonly registerLink: Locator;
  readonly errorMessage: Locator;
  readonly rememberMeCheckbox: Locator;

  constructor(page: Page) {
    this.page = page;
    this.usernameInput = page.getByLabel('Username');
    this.passwordInput = page.getByLabel('Password');
    this.loginButton = page.getByRole('button', { name: /sign in|log in/i });
    this.registerLink = page.getByRole('link', { name: /sign up|register/i });
    this.errorMessage = page.getByRole('alert');
    this.rememberMeCheckbox = page.getByLabel(/remember me/i);
  }

  async goto() {
    await this.page.goto('/login');
  }

  async login(username: string, password: string, rememberMe = false) {
    await this.usernameInput.fill(username);
    await this.passwordInput.fill(password);
    
    if (rememberMe) {
      await this.rememberMeCheckbox.check();
    }
    
    await this.loginButton.click();
  }

  async expectError(message: string) {
    await this.errorMessage.waitFor({ state: 'visible' });
    await this.page.waitForTimeout(100); // Small delay for text to appear
    const errorText = await this.errorMessage.textContent();
    return errorText?.includes(message);
  }

  async navigateToRegister() {
    await this.registerLink.click();
  }
}