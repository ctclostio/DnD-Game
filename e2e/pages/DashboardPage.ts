import { Page, Locator } from '@playwright/test';

export class DashboardPage {
  readonly page: Page;
  readonly welcomeMessage: Locator;
  readonly createCharacterButton: Locator;
  readonly joinSessionButton: Locator;
  readonly createSessionButton: Locator;
  readonly characterList: Locator;
  readonly sessionList: Locator;
  readonly profileMenu: Locator;
  readonly logoutButton: Locator;
  readonly notificationBell: Locator;

  constructor(page: Page) {
    this.page = page;
    this.welcomeMessage = page.getByText(/welcome|dashboard/i);
    this.createCharacterButton = page.getByRole('button', { name: /create.*character/i });
    this.joinSessionButton = page.getByRole('button', { name: /join.*session/i });
    this.createSessionButton = page.getByRole('button', { name: /create.*session/i });
    this.characterList = page.getByTestId('character-list');
    this.sessionList = page.getByTestId('session-list');
    this.profileMenu = page.getByRole('button', { name: /profile|account/i });
    this.logoutButton = page.getByRole('button', { name: /log.*out|sign.*out/i });
    this.notificationBell = page.getByRole('button', { name: /notifications/i });
  }

  async goto() {
    await this.page.goto('/dashboard');
  }

  async waitForLoad() {
    await this.welcomeMessage.waitFor({ state: 'visible' });
  }

  async createNewCharacter() {
    await this.createCharacterButton.click();
  }

  async joinSession(sessionId?: string) {
    if (sessionId) {
      const sessionCard = this.page.getByTestId(`session-${sessionId}`);
      await sessionCard.getByRole('button', { name: /join/i }).click();
    } else {
      await this.joinSessionButton.click();
    }
  }

  async createNewSession() {
    await this.createSessionButton.click();
  }

  async selectCharacter(characterName: string) {
    const characterCard = this.page.getByText(characterName).locator('..');
    await characterCard.click();
  }

  async logout() {
    await this.profileMenu.click();
    await this.logoutButton.click();
  }

  async getCharacterCount() {
    const characters = await this.characterList.getByRole('listitem').all();
    return characters.length;
  }

  async getSessionCount() {
    const sessions = await this.sessionList.getByRole('listitem').all();
    return sessions.length;
  }

  async expectNotification(message: string) {
    const notification = this.page.getByText(message);
    await notification.waitFor({ state: 'visible' });
    return true;
  }
}