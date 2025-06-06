import { Page, Locator } from '@playwright/test';

export class GameSessionPage {
  readonly page: Page;
  
  // Session creation
  readonly sessionNameInput: Locator;
  readonly sessionDescriptionInput: Locator;
  readonly campaignSelect: Locator;
  readonly createButton: Locator;
  
  // Session lobby
  readonly sessionTitle: Locator;
  readonly playerList: Locator;
  readonly chatInput: Locator;
  readonly sendButton: Locator;
  readonly chatMessages: Locator;
  readonly startGameButton: Locator;
  readonly inviteButton: Locator;
  readonly leaveButton: Locator;
  
  // Character selection
  readonly characterSelect: Locator;
  readonly readyButton: Locator;
  readonly characterPreview: Locator;
  
  // In-game
  readonly gameBoard: Locator;
  readonly initiativeTracker: Locator;
  readonly actionButtons: Locator;
  readonly diceRoller: Locator;

  constructor(page: Page) {
    this.page = page;
    
    // Session creation
    this.sessionNameInput = page.getByLabel(/session name/i);
    this.sessionDescriptionInput = page.getByLabel(/description/i);
    this.campaignSelect = page.getByLabel(/campaign/i);
    this.createButton = page.getByRole('button', { name: /create session/i });
    
    // Session lobby
    this.sessionTitle = page.getByRole('heading', { level: 1 });
    this.playerList = page.getByTestId('player-list');
    this.chatInput = page.getByPlaceholder(/type.*message/i);
    this.sendButton = page.getByRole('button', { name: /send/i });
    this.chatMessages = page.getByTestId('chat-messages');
    this.startGameButton = page.getByRole('button', { name: /start game/i });
    this.inviteButton = page.getByRole('button', { name: /invite/i });
    this.leaveButton = page.getByRole('button', { name: /leave/i });
    
    // Character selection
    this.characterSelect = page.getByLabel(/select character/i);
    this.readyButton = page.getByRole('button', { name: /ready/i });
    this.characterPreview = page.getByTestId('character-preview');
    
    // In-game
    this.gameBoard = page.getByTestId('game-board');
    this.initiativeTracker = page.getByTestId('initiative-tracker');
    this.actionButtons = page.getByTestId('action-buttons');
    this.diceRoller = page.getByTestId('dice-roller');
  }

  async createSession(name: string, description?: string, campaignId?: string) {
    await this.sessionNameInput.fill(name);
    
    if (description) {
      await this.sessionDescriptionInput.fill(description);
    }
    
    if (campaignId) {
      await this.campaignSelect.selectOption(campaignId);
    }
    
    await this.createButton.click();
  }

  async joinSession(sessionId: string) {
    await this.page.goto(`/session/${sessionId}`);
  }

  async selectCharacter(characterName: string) {
    await this.characterSelect.selectOption(characterName);
    await this.readyButton.click();
  }

  async sendChatMessage(message: string) {
    await this.chatInput.fill(message);
    await this.sendButton.click();
  }

  async waitForPlayer(playerName: string) {
    const player = this.playerList.getByText(playerName);
    await player.waitFor({ state: 'visible' });
  }

  async startGame() {
    await this.startGameButton.click();
    await this.gameBoard.waitFor({ state: 'visible' });
  }

  async leaveSession() {
    await this.leaveButton.click();
  }

  async getPlayerCount(): Promise<number> {
    const players = await this.playerList.getByRole('listitem').all();
    return players.length;
  }

  async getChatMessageCount(): Promise<number> {
    const messages = await this.chatMessages.getByRole('listitem').all();
    return messages.length;
  }

  async expectChatMessage(message: string) {
    const chatMessage = this.chatMessages.getByText(message);
    await chatMessage.waitFor({ state: 'visible' });
    return true;
  }

  async isGameStarted(): Promise<boolean> {
    try {
      await this.gameBoard.waitFor({ state: 'visible', timeout: 1000 });
      return true;
    } catch {
      return false;
    }
  }

  async isDM(): Promise<boolean> {
    // DM has access to start game button
    return await this.startGameButton.isVisible();
  }
}