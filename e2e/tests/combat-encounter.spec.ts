import { test, expect, testData } from '../fixtures/base';
import { Browser, BrowserContext, Page } from '@playwright/test';

test.describe('Combat Encounter Flow', () => {
  let dmUser: ReturnType<typeof testData.generateUser>;
  let playerUser: ReturnType<typeof testData.generateUser>;
  let dmCharacter: ReturnType<typeof testData.generateCharacter>;
  let playerCharacter: ReturnType<typeof testData.generateCharacter>;
  let sessionId: string;

  test.beforeEach(async ({ browser }) => {
    // Create users and characters
    dmUser = testData.generateUser();
    playerUser = testData.generateUser();
    dmCharacter = testData.generateCharacter();
    playerCharacter = testData.generateCharacter();
    
    dmCharacter.name = 'DM_' + dmCharacter.name;
    playerCharacter.name = 'Hero_' + playerCharacter.name;
    
    // Setup DM and player with characters
    const setupContext = await browser.newContext();
    const setupPage = await setupContext.newPage();
    
    // Create DM
    await createUserWithCharacter(setupPage, dmUser, dmCharacter);
    
    // Create Player
    await createUserWithCharacter(setupPage, playerUser, playerCharacter);
    
    // Create session as DM
    await setupPage.goto('/login');
    await setupPage.getByLabel('Username').fill(dmUser.username);
    await setupPage.getByLabel('Password').fill(dmUser.password);
    await setupPage.getByRole('button', { name: /sign in|log in/i }).click();
    await setupPage.waitForURL(/\/dashboard/);
    
    await setupPage.goto('/session/create');
    const sessionName = `Combat_Test_${Date.now()}`;
    await setupPage.getByLabel(/session name/i).fill(sessionName);
    await setupPage.getByRole('button', { name: /create session/i }).click();
    await setupPage.waitForURL(/\/session\/([\w-]+)$/);
    
    sessionId = setupPage.url().split('/').pop()!;
    
    await setupContext.close();
  });

  test('DM should initiate combat encounter', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    // Setup session with both users
    await joinSession(dmPage, dmUser, sessionId);
    await joinSession(playerPage, playerUser, sessionId, playerCharacter.name);
    
    // Wait for player to appear
    await dmPage.waitForSelector(`text="${playerCharacter.name}"`);
    
    // Start game
    await dmPage.getByRole('button', { name: /start game/i }).click();
    await dmPage.waitForSelector('[data-testid="game-board"]');
    await playerPage.waitForSelector('[data-testid="game-board"]');
    
    // DM starts combat
    await dmPage.getByRole('button', { name: /start combat/i }).click();
    
    // Should show initiative tracker
    await dmPage.waitForSelector('[data-testid="initiative-list"]');
    await playerPage.waitForSelector('[data-testid="initiative-list"]');
    
    // Verify combat started
    const dmCombatActive = await dmPage.getByRole('button', { name: /end combat/i }).isVisible();
    const playerSeesInitiative = await playerPage.locator('[data-testid="initiative-list"]').isVisible();
    
    expect(dmCombatActive).toBeTruthy();
    expect(playerSeesInitiative).toBeTruthy();
    
    await dmContext.close();
    await playerContext.close();
  });

  test('Combat should follow initiative order', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    await setupCombat(dmPage, playerPage, dmUser, playerUser, sessionId, playerCharacter.name);
    
    // Add NPC enemy
    await dmPage.getByRole('button', { name: /add.*participant/i }).click();
    await dmPage.getByLabel(/name/i).fill('Goblin');
    await dmPage.getByLabel(/hit points/i).fill('7');
    await dmPage.getByLabel(/armor class/i).fill('13');
    await dmPage.getByLabel(/initiative/i).fill('12');
    await dmPage.getByRole('button', { name: /add|confirm/i }).click();
    
    // Get initial turn order
    const initialOrder = await dmPage.locator('[data-testid="initiative-list"] [role="listitem"]').allTextContents();
    
    // First participant should be current
    const currentTurn = await dmPage.locator('[data-testid="current-turn"]').textContent();
    expect(currentTurn).toContain(initialOrder[0]);
    
    // Advance turn
    await dmPage.getByRole('button', { name: /next turn/i }).click();
    
    // Should move to next participant
    const nextTurn = await dmPage.locator('[data-testid="current-turn"]').textContent();
    expect(nextTurn).toContain(initialOrder[1]);
    
    await dmContext.close();
    await playerContext.close();
  });

  test('Player should perform attack action', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    await setupCombat(dmPage, playerPage, dmUser, playerUser, sessionId, playerCharacter.name);
    
    // Add enemy
    await dmPage.getByRole('button', { name: /add.*participant/i }).click();
    await dmPage.getByLabel(/name/i).fill('Orc');
    await dmPage.getByLabel(/hit points/i).fill('15');
    await dmPage.getByLabel(/armor class/i).fill('13');
    await dmPage.getByLabel(/initiative/i).fill('8'); // Lower than player
    await dmPage.getByRole('button', { name: /add|confirm/i }).click();
    
    // Wait for it to be player's turn
    const currentPlayer = await playerPage.locator('[data-testid="current-turn"]').textContent();
    
    if (!currentPlayer?.includes(playerCharacter.name)) {
      // Advance turns until it's player's turn
      while (true) {
        const current = await dmPage.locator('[data-testid="current-turn"]').textContent();
        if (current?.includes(playerCharacter.name)) break;
        await dmPage.getByRole('button', { name: /next turn/i }).click();
        await dmPage.waitForTimeout(500);
      }
    }
    
    // Player attacks
    await playerPage.getByRole('button', { name: /^attack/i }).click();
    await playerPage.getByLabel(/select target/i).selectOption('Orc');
    await playerPage.getByRole('button', { name: /confirm/i }).click();
    
    // Wait for dice roll
    await playerPage.waitForSelector('[data-testid="dice-results"]');
    
    // Should see combat log entry
    const combatLogEntry = playerPage.locator('[data-testid="combat-log"]').getByText(/attacks Orc/i);
    await expect(combatLogEntry).toBeVisible();
    
    // Both should see the same log
    const dmLogEntry = dmPage.locator('[data-testid="combat-log"]').getByText(/attacks Orc/i);
    await expect(dmLogEntry).toBeVisible();
    
    await dmContext.close();
    await playerContext.close();
  });

  test('DM should apply damage and healing', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    await setupCombat(dmPage, playerPage, dmUser, playerUser, sessionId, playerCharacter.name);
    
    // Get initial HP
    const initialHP = await getParticipantHP(playerPage, playerCharacter.name);
    
    // DM applies damage to player
    await dmPage.locator(`text="${playerCharacter.name}"`).locator('..').click();
    await dmPage.getByLabel(/damage/i).fill('5');
    await dmPage.getByRole('button', { name: /apply damage/i }).click();
    
    // Verify HP reduced
    await playerPage.waitForTimeout(500);
    const damagedHP = await getParticipantHP(playerPage, playerCharacter.name);
    expect(damagedHP.current).toBe(initialHP.current - 5);
    
    // DM heals player
    await dmPage.locator(`text="${playerCharacter.name}"`).locator('..').click();
    await dmPage.getByLabel(/healing/i).fill('3');
    await dmPage.getByRole('button', { name: /apply healing/i }).click();
    
    // Verify HP increased
    await playerPage.waitForTimeout(500);
    const healedHP = await getParticipantHP(playerPage, playerCharacter.name);
    expect(healedHP.current).toBe(damagedHP.current + 3);
    
    await dmContext.close();
    await playerContext.close();
  });

  test('Combat should handle conditions', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    await setupCombat(dmPage, playerPage, dmUser, playerUser, sessionId, playerCharacter.name);
    
    // DM applies condition to player
    await dmPage.locator(`text="${playerCharacter.name}"`).locator('..').click();
    await dmPage.getByLabel(/condition/i).selectOption('Poisoned');
    await dmPage.getByRole('button', { name: /add condition/i }).click();
    
    // Both should see condition
    await playerPage.waitForSelector('[data-testid="condition-tag"]:has-text("Poisoned")');
    const playerSeesCondition = await playerPage.locator('[data-testid="condition-tag"]').getByText('Poisoned').isVisible();
    const dmSeesCondition = await dmPage.locator('[data-testid="condition-tag"]').getByText('Poisoned').isVisible();
    
    expect(playerSeesCondition).toBeTruthy();
    expect(dmSeesCondition).toBeTruthy();
    
    await dmContext.close();
    await playerContext.close();
  });

  test('Combat should track rounds and turns', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    await setupCombat(dmPage, playerPage, dmUser, playerUser, sessionId, playerCharacter.name);
    
    // Add multiple participants
    const enemies = ['Goblin', 'Orc', 'Wolf'];
    for (const enemy of enemies) {
      await dmPage.getByRole('button', { name: /add.*participant/i }).click();
      await dmPage.getByLabel(/name/i).fill(enemy);
      await dmPage.getByLabel(/hit points/i).fill('10');
      await dmPage.getByLabel(/armor class/i).fill('12');
      await dmPage.getByLabel(/initiative/i).fill(String(Math.floor(Math.random() * 20) + 1));
      await dmPage.getByRole('button', { name: /add|confirm/i }).click();
    }
    
    // Check initial round
    let round = await dmPage.locator('[data-testid="round-counter"]').textContent();
    expect(round).toContain('1');
    
    // Get participant count
    const participantCount = await dmPage.locator('[data-testid="initiative-list"] [role="listitem"]').count();
    
    // Advance through all turns to complete round
    for (let i = 0; i < participantCount; i++) {
      await dmPage.getByRole('button', { name: /next turn|end turn/i }).click();
      await dmPage.waitForTimeout(200);
    }
    
    // Should be round 2
    round = await dmPage.locator('[data-testid="round-counter"]').textContent();
    expect(round).toContain('2');
    
    await dmContext.close();
    await playerContext.close();
  });

  test('Combat should end properly', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    await setupCombat(dmPage, playerPage, dmUser, playerUser, sessionId, playerCharacter.name);
    
    // End combat
    await dmPage.getByRole('button', { name: /end combat/i }).click();
    
    // Confirm if needed
    const confirmButton = dmPage.getByRole('button', { name: /confirm|yes/i });
    if (await confirmButton.isVisible({ timeout: 1000 })) {
      await confirmButton.click();
    }
    
    // Should no longer show combat UI
    await dmPage.waitForSelector('[data-testid="initiative-list"]', { state: 'hidden' });
    await playerPage.waitForSelector('[data-testid="initiative-list"]', { state: 'hidden' });
    
    // Start combat button should be visible again
    const startCombatVisible = await dmPage.getByRole('button', { name: /start combat/i }).isVisible();
    expect(startCombatVisible).toBeTruthy();
    
    await dmContext.close();
    await playerContext.close();
  });

  test('Combat should handle player defeat', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    await setupCombat(dmPage, playerPage, dmUser, playerUser, sessionId, playerCharacter.name);
    
    // Get player's max HP
    const playerHP = await getParticipantHP(dmPage, playerCharacter.name);
    
    // Apply lethal damage
    await dmPage.locator(`text="${playerCharacter.name}"`).locator('..').click();
    await dmPage.getByLabel(/damage/i).fill(String(playerHP.current + 10));
    await dmPage.getByRole('button', { name: /apply damage/i }).click();
    
    // Player should be at 0 HP with unconscious condition
    await playerPage.waitForTimeout(500);
    const defeatedHP = await getParticipantHP(playerPage, playerCharacter.name);
    expect(defeatedHP.current).toBe(0);
    
    // Should have unconscious condition
    const unconsciousCondition = await playerPage.locator('[data-testid="condition-tag"]').getByText(/unconscious/i).isVisible();
    expect(unconsciousCondition).toBeTruthy();
    
    await dmContext.close();
    await playerContext.close();
  });

  test('Combat actions should consume resources', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    // Create a spellcaster
    const wizardUser = testData.generateUser();
    const wizardChar = testData.generateCharacter();
    wizardChar.name = 'Wizard_' + wizardChar.name;
    wizardChar.class = 'Wizard';
    
    const setupPage = await browser.newPage();
    await createUserWithCharacter(setupPage, wizardUser, wizardChar);
    await setupPage.close();
    
    // Join session as wizard
    await playerPage.goto('/login');
    await playerPage.getByLabel('Username').fill(wizardUser.username);
    await playerPage.getByLabel('Password').fill(wizardUser.password);
    await playerPage.getByRole('button', { name: /sign in|log in/i }).click();
    await playerPage.waitForURL(/\/dashboard/);
    
    await playerPage.goto(`/session/${sessionId}`);
    await playerPage.getByLabel(/select character/i).selectOption(wizardChar.name);
    await playerPage.getByRole('button', { name: /ready/i }).click();
    
    // DM starts combat
    await joinSession(dmPage, dmUser, sessionId);
    await dmPage.waitForSelector(`text="${wizardChar.name}"`);
    await dmPage.getByRole('button', { name: /start game/i }).click();
    await dmPage.waitForSelector('[data-testid="game-board"]');
    await dmPage.getByRole('button', { name: /start combat/i }).click();
    
    // Add enemy
    await dmPage.getByRole('button', { name: /add.*participant/i }).click();
    await dmPage.getByLabel(/name/i).fill('Target Dummy');
    await dmPage.getByLabel(/hit points/i).fill('20');
    await dmPage.getByLabel(/armor class/i).fill('10');
    await dmPage.getByLabel(/initiative/i).fill('1');
    await dmPage.getByRole('button', { name: /add|confirm/i }).click();
    
    // Wait for wizard's turn
    while (true) {
      const current = await playerPage.locator('[data-testid="current-turn"]').textContent();
      if (current?.includes(wizardChar.name)) break;
      await dmPage.getByRole('button', { name: /next turn/i }).click();
      await dmPage.waitForTimeout(500);
    }
    
    // Cast spell (should consume spell slot)
    const initialSlots = await playerPage.locator('[data-testid="spell-slots-1"]').textContent();
    
    await playerPage.getByRole('button', { name: /cast spell/i }).click();
    await playerPage.getByText(/magic missile/i).click();
    await playerPage.getByLabel(/select target/i).selectOption('Target Dummy');
    await playerPage.getByRole('button', { name: /confirm/i }).click();
    
    // Check spell slot consumed
    await playerPage.waitForTimeout(500);
    const remainingSlots = await playerPage.locator('[data-testid="spell-slots-1"]').textContent();
    expect(remainingSlots).not.toBe(initialSlots);
    
    await dmContext.close();
    await playerContext.close();
  });
});

// Helper functions
async function createUserWithCharacter(page: Page, user: any, character: any) {
  await page.goto('/register');
  await page.getByLabel('Username').fill(user.username);
  await page.getByLabel('Email').fill(user.email);
  await page.getByLabel('Password', { exact: true }).fill(user.password);
  await page.getByLabel(/confirm password/i).fill(user.password);
  await page.getByRole('button', { name: /sign up|register|create account/i }).click();
  await page.waitForURL(/\/dashboard/);
  
  // Create character
  await page.getByRole('button', { name: /create.*character/i }).click();
  await page.getByLabel(/character name/i).fill(character.name);
  await page.getByLabel(/race/i).selectOption(character.race);
  await page.getByLabel(/class/i).selectOption(character.class);
  await page.getByLabel(/background/i).selectOption(character.background);
  await page.getByRole('button', { name: /next|continue/i }).click();
  
  // Quick setup - roll abilities
  await page.getByRole('button', { name: /roll.*abilities/i }).click();
  await page.getByRole('button', { name: /next|continue/i }).click();
  
  // Skip skills
  await page.getByRole('button', { name: /next|continue/i }).click();
  
  // Finish
  await page.getByRole('button', { name: /finish|create character/i }).click();
  await page.waitForURL(/\/character\/[\w-]+$/);
  
  // Logout
  await page.context().clearCookies();
  await page.evaluate(() => localStorage.clear());
}

async function joinSession(page: Page, user: any, sessionId: string, characterName?: string) {
  await page.goto('/login');
  await page.getByLabel('Username').fill(user.username);
  await page.getByLabel('Password').fill(user.password);
  await page.getByRole('button', { name: /sign in|log in/i }).click();
  await page.waitForURL(/\/dashboard/);
  
  await page.goto(`/session/${sessionId}`);
  
  if (characterName) {
    await page.getByLabel(/select character/i).selectOption(characterName);
    await page.getByRole('button', { name: /ready/i }).click();
  }
}

async function setupCombat(dmPage: Page, playerPage: Page, dmUser: any, playerUser: any, sessionId: string, playerCharacterName: string) {
  await joinSession(dmPage, dmUser, sessionId);
  await joinSession(playerPage, playerUser, sessionId, playerCharacterName);
  
  // Wait for player to appear
  await dmPage.waitForSelector(`text="${playerCharacterName}"`);
  
  // Start game and combat
  await dmPage.getByRole('button', { name: /start game/i }).click();
  await dmPage.waitForSelector('[data-testid="game-board"]');
  await playerPage.waitForSelector('[data-testid="game-board"]');
  
  await dmPage.getByRole('button', { name: /start combat/i }).click();
  await dmPage.waitForSelector('[data-testid="initiative-list"]');
  await playerPage.waitForSelector('[data-testid="initiative-list"]');
}

async function getParticipantHP(page: Page, participantName: string): Promise<{ current: number; max: number }> {
  const participant = page.locator(`[data-testid="initiative-list"]`).getByText(participantName).locator('..');
  const healthText = await participant.getByTestId('health-display').textContent();
  // Use a safe regex pattern that avoids backtracking vulnerabilities
  // Match only spaces (not all whitespace) with strict limits
  const match = healthText?.match(/(\d+)[ ]{0,3}\/[ ]{0,3}(\d+)/);
  
  if (match) {
    return {
      current: parseInt(match[1]),
      max: parseInt(match[2]),
    };
  }
  
  return { current: 0, max: 0 };
}