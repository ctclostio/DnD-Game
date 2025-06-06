import { Page, Locator } from '@playwright/test';

export class RegisterPage {
  readonly page: Page;
  readonly usernameInput: Locator;
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly confirmPasswordInput: Locator;
  readonly registerButton: Locator;
  readonly loginLink: Locator;
  readonly errorMessage: Locator;
  readonly successMessage: Locator;
  readonly termsCheckbox: Locator;

  constructor(page: Page) {
    this.page = page;
    this.usernameInput = page.getByLabel('Username');
    this.emailInput = page.getByLabel('Email');
    this.passwordInput = page.getByLabel('Password', { exact: true });
    this.confirmPasswordInput = page.getByLabel(/confirm password/i);
    this.registerButton = page.getByRole('button', { name: /sign up|register|create account/i });
    this.loginLink = page.getByRole('link', { name: /sign in|log in/i });
    this.errorMessage = page.getByRole('alert');
    this.successMessage = page.getByText(/registration successful|account created/i);
    this.termsCheckbox = page.getByLabel(/terms|agree/i);
  }

  async goto() {
    await this.page.goto('/register');
  }

  async register(username: string, email: string, password: string, confirmPassword?: string) {
    await this.usernameInput.fill(username);
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.confirmPasswordInput.fill(confirmPassword || password);
    
    // Check terms if present
    if (await this.termsCheckbox.isVisible()) {
      await this.termsCheckbox.check();
    }
    
    await this.registerButton.click();
  }

  async expectError(message: string) {
    await this.errorMessage.waitFor({ state: 'visible' });
    const errorText = await this.errorMessage.textContent();
    return errorText?.includes(message);
  }

  async expectSuccess() {
    await this.successMessage.waitFor({ state: 'visible' });
    return true;
  }

  async navigateToLogin() {
    await this.loginLink.click();
  }
}