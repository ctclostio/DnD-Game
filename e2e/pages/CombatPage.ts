import { Page, Locator } from '@playwright/test';

export class CombatPage {
  readonly page: Page;
  
  // Combat controls
  readonly startCombatButton: Locator;
  readonly endCombatButton: Locator;
  readonly nextTurnButton: Locator;
  readonly addParticipantButton: Locator;
  
  // Initiative tracker
  readonly initiativeList: Locator;
  readonly currentTurnIndicator: Locator;
  readonly roundCounter: Locator;
  
  // Action panel
  readonly attackButton: Locator;
  readonly castSpellButton: Locator;
  readonly moveButton: Locator;
  readonly dashButton: Locator;
  readonly dodgeButton: Locator;
  readonly helpButton: Locator;
  readonly hideButton: Locator;
  readonly readyButton: Locator;
  readonly endTurnButton: Locator;
  
  // Target selection
  readonly targetSelect: Locator;
  readonly confirmActionButton: Locator;
  readonly cancelActionButton: Locator;
  
  // Combat log
  readonly combatLog: Locator;
  readonly diceResults: Locator;
  
  // Health tracking
  readonly healthBars: Locator;
  readonly damageInput: Locator;
  readonly healInput: Locator;
  readonly applyDamageButton: Locator;
  readonly applyHealingButton: Locator;
  
  // Conditions
  readonly conditionSelect: Locator;
  readonly addConditionButton: Locator;
  readonly conditionTags: Locator;

  constructor(page: Page) {
    this.page = page;
    
    // Combat controls
    this.startCombatButton = page.getByRole('button', { name: /start combat/i });
    this.endCombatButton = page.getByRole('button', { name: /end combat/i });
    this.nextTurnButton = page.getByRole('button', { name: /next turn/i });
    this.addParticipantButton = page.getByRole('button', { name: /add.*participant/i });
    
    // Initiative tracker
    this.initiativeList = page.getByTestId('initiative-list');
    this.currentTurnIndicator = page.getByTestId('current-turn');
    this.roundCounter = page.getByTestId('round-counter');
    
    // Action panel
    this.attackButton = page.getByRole('button', { name: /^attack/i });
    this.castSpellButton = page.getByRole('button', { name: /cast spell/i });
    this.moveButton = page.getByRole('button', { name: /^move/i });
    this.dashButton = page.getByRole('button', { name: /dash/i });
    this.dodgeButton = page.getByRole('button', { name: /dodge/i });
    this.helpButton = page.getByRole('button', { name: /help/i });
    this.hideButton = page.getByRole('button', { name: /hide/i });
    this.readyButton = page.getByRole('button', { name: /ready action/i });
    this.endTurnButton = page.getByRole('button', { name: /end turn/i });
    
    // Target selection
    this.targetSelect = page.getByLabel(/select target/i);
    this.confirmActionButton = page.getByRole('button', { name: /confirm/i });
    this.cancelActionButton = page.getByRole('button', { name: /cancel/i });
    
    // Combat log
    this.combatLog = page.getByTestId('combat-log');
    this.diceResults = page.getByTestId('dice-results');
    
    // Health tracking
    this.healthBars = page.getByTestId('health-bar');
    this.damageInput = page.getByLabel(/damage/i);
    this.healInput = page.getByLabel(/healing/i);
    this.applyDamageButton = page.getByRole('button', { name: /apply damage/i });
    this.applyHealingButton = page.getByRole('button', { name: /apply healing/i });
    
    // Conditions
    this.conditionSelect = page.getByLabel(/condition/i);
    this.addConditionButton = page.getByRole('button', { name: /add condition/i });
    this.conditionTags = page.getByTestId('condition-tag');
  }

  async startCombat() {
    await this.startCombatButton.click();
    await this.initiativeList.waitFor({ state: 'visible' });
  }

  async endCombat() {
    await this.endCombatButton.click();
  }

  async performAttack(targetName: string) {
    await this.attackButton.click();
    await this.targetSelect.selectOption(targetName);
    await this.confirmActionButton.click();
    
    // Wait for dice roll animation
    await this.diceResults.waitFor({ state: 'visible' });
  }

  async castSpell(spellName: string, targetName?: string) {
    await this.castSpellButton.click();
    
    // Select spell from modal/dropdown
    const spellOption = this.page.getByText(spellName);
    await spellOption.click();
    
    if (targetName) {
      await this.targetSelect.selectOption(targetName);
    }
    
    await this.confirmActionButton.click();
  }

  async endTurn() {
    await this.endTurnButton.click();
    
    // Wait for turn indicator to update
    await this.page.waitForTimeout(500);
  }

  async nextTurn() {
    await this.nextTurnButton.click();
  }

  async applyDamage(participantName: string, amount: number) {
    const participant = this.initiativeList.getByText(participantName).locator('..');
    await participant.click();
    
    await this.damageInput.fill(amount.toString());
    await this.applyDamageButton.click();
  }

  async applyHealing(participantName: string, amount: number) {
    const participant = this.initiativeList.getByText(participantName).locator('..');
    await participant.click();
    
    await this.healInput.fill(amount.toString());
    await this.applyHealingButton.click();
  }

  async addCondition(participantName: string, condition: string) {
    const participant = this.initiativeList.getByText(participantName).locator('..');
    await participant.click();
    
    await this.conditionSelect.selectOption(condition);
    await this.addConditionButton.click();
  }

  async getCurrentTurn(): Promise<string | null> {
    const currentTurn = await this.currentTurnIndicator.textContent();
    return currentTurn;
  }

  async getRound(): Promise<number> {
    const roundText = await this.roundCounter.textContent();
    const match = roundText?.match(/round (\d+)/i);
    return match ? parseInt(match[1]) : 1;
  }

  async getParticipantHealth(participantName: string): Promise<{ current: number; max: number }> {
    const participant = this.initiativeList.getByText(participantName).locator('..');
    const healthText = await participant.getByTestId('health-display').textContent();
    const match = healthText?.match(/(\d+)\s*\/\s*(\d+)/);
    
    if (match) {
      return {
        current: parseInt(match[1]),
        max: parseInt(match[2]),
      };
    }
    
    return { current: 0, max: 0 };
  }

  async expectCombatLogEntry(text: string) {
    const logEntry = this.combatLog.getByText(text);
    await logEntry.waitFor({ state: 'visible' });
    return true;
  }

  async isCombatActive(): Promise<boolean> {
    return await this.endCombatButton.isVisible();
  }

  async getInitiativeOrder(): Promise<string[]> {
    const participants = await this.initiativeList.getByRole('listitem').all();
    const names: string[] = [];
    
    for (const participant of participants) {
      const name = await participant.getByTestId('participant-name').textContent();
      if (name) names.push(name);
    }
    
    return names;
  }
}