import { test, expect, testData, helpers } from '../fixtures/base';
import { Browser, BrowserContext, Page } from '@playwright/test';

test.describe('Game Session Flow', () => {
  let dmUser: ReturnType<typeof testData.generateUser>;
  let playerUser: ReturnType<typeof testData.generateUser>;
  let dmCharacter: ReturnType<typeof testData.generateCharacter>;
  let playerCharacter: ReturnType<typeof testData.generateCharacter>;

  test.beforeEach(async ({ page, registerPage, characterBuilderPage, dashboardPage }) => {
    // Create DM user and character
    dmUser = testData.generateUser();
    dmCharacter = testData.generateCharacter();
    dmCharacter.name = 'DM_' + dmCharacter.name;
    
    await registerPage.goto();
    await registerPage.register(dmUser.username, dmUser.email, dmUser.password);
    await page.waitForURL(/\/dashboard/);
    
    // Create DM character
    await dashboardPage.createNewCharacter();
    await characterBuilderPage.fillBasicInfo({
      name: dmCharacter.name,
      race: dmCharacter.race,
      class: dmCharacter.class,
      background: dmCharacter.background,
    });
    await characterBuilderPage.setAbilityScores(dmCharacter.abilities);
    await characterBuilderPage.selectSkills(['Athletics']);
    await characterBuilderPage.reviewAndFinish();
    await characterBuilderPage.expectSuccess();
    
    // Create player user and character
    playerUser = testData.generateUser();
    playerCharacter = testData.generateCharacter();
    playerCharacter.name = 'Player_' + playerCharacter.name;
    
    // Clear session and create player
    await page.context().clearCookies();
    await page.evaluate(() => localStorage.clear());
    
    await registerPage.goto();
    await registerPage.register(playerUser.username, playerUser.email, playerUser.password);
    await page.waitForURL(/\/dashboard/);
    
    // Create player character
    await dashboardPage.createNewCharacter();
    await characterBuilderPage.fillBasicInfo({
      name: playerCharacter.name,
      race: playerCharacter.race,
      class: playerCharacter.class,
      background: playerCharacter.background,
    });
    await characterBuilderPage.setAbilityScores(playerCharacter.abilities);
    await characterBuilderPage.selectSkills(['Stealth', 'Perception']);
    await characterBuilderPage.reviewAndFinish();
    await characterBuilderPage.expectSuccess();
  });

  test('DM should create a game session', async ({ page, loginPage, dashboardPage, gameSessionPage }) => {
    // Login as DM
    await page.context().clearCookies();
    await page.evaluate(() => localStorage.clear());
    await loginPage.goto();
    await loginPage.login(dmUser.username, dmUser.password);
    await page.waitForURL(/\/dashboard/);
    
    // Create session
    await dashboardPage.createNewSession();
    await page.waitForURL(/\/session\/create/);
    
    const sessionData = testData.generateSession();
    await gameSessionPage.createSession(sessionData.name, sessionData.description);
    
    // Should redirect to session lobby
    await page.waitForURL(/\/session\/[\w-]+$/);
    
    // Verify session created
    await expect(gameSessionPage.sessionTitle).toContainText(sessionData.name);
    await expect(gameSessionPage.startGameButton).toBeVisible(); // DM has start button
    
    // Check player count
    const playerCount = await gameSessionPage.getPlayerCount();
    expect(playerCount).toBe(1); // Just the DM
  });

  test('Player should join existing session', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    // DM creates session
    await dmPage.goto('/login');
    await dmPage.getByLabel('Username').fill(dmUser.username);
    await dmPage.getByLabel('Password').fill(dmUser.password);
    await dmPage.getByRole('button', { name: /sign in|log in/i }).click();
    await dmPage.waitForURL(/\/dashboard/);
    
    await dmPage.getByRole('button', { name: /create.*session/i }).click();
    await dmPage.waitForURL(/\/session\/create/);
    
    const sessionName = `Session_${Date.now()}`;
    await dmPage.getByLabel(/session name/i).fill(sessionName);
    await dmPage.getByRole('button', { name: /create session/i }).click();
    await dmPage.waitForURL(/\/session\/([\w-]+)$/);
    
    // Get session ID from URL
    const sessionUrl = dmPage.url();
    const sessionId = sessionUrl.split('/').pop()!;
    
    // Player joins session
    await playerPage.goto('/login');
    await playerPage.getByLabel('Username').fill(playerUser.username);
    await playerPage.getByLabel('Password').fill(playerUser.password);
    await playerPage.getByRole('button', { name: /sign in|log in/i }).click();
    await playerPage.waitForURL(/\/dashboard/);
    
    // Join session directly by URL
    await playerPage.goto(`/session/${sessionId}`);
    
    // Select character
    await playerPage.getByLabel(/select character/i).selectOption(playerCharacter.name);
    await playerPage.getByRole('button', { name: /ready/i }).click();
    
    // Verify player joined
    await expect(playerPage.getByText(sessionName)).toBeVisible();
    await expect(playerPage.getByRole('button', { name: /leave/i })).toBeVisible();
    
    // DM should see player joined
    await dmPage.waitForFunction(() => {
      const playerList = document.querySelector('[data-testid="player-list"]');
      return playerList?.textContent?.includes('Player_');
    });
    
    const dmPlayerCount = await dmPage.locator('[data-testid="player-list"] [role="listitem"]').count();
    expect(dmPlayerCount).toBe(2); // DM + Player
    
    await dmContext.close();
    await playerContext.close();
  });

  test('Real-time chat should work between players', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    // Setup session with both users
    await setupMultiplayerSession(dmPage, playerPage, dmUser, playerUser, playerCharacter.name);
    
    // DM sends message
    const dmMessage = 'Welcome to the adventure!';
    await dmPage.getByPlaceholder(/type.*message/i).fill(dmMessage);
    await dmPage.getByRole('button', { name: /send/i }).click();
    
    // Player should see message
    await playerPage.waitForSelector(`text="${dmMessage}"`);
    
    // Player responds
    const playerMessage = 'Ready to play!';
    await playerPage.getByPlaceholder(/type.*message/i).fill(playerMessage);
    await playerPage.getByRole('button', { name: /send/i }).click();
    
    // DM should see message
    await dmPage.waitForSelector(`text="${playerMessage}"`);
    
    // Verify message count
    const dmMessageCount = await dmPage.locator('[data-testid="chat-messages"] [role="listitem"]').count();
    const playerMessageCount = await playerPage.locator('[data-testid="chat-messages"] [role="listitem"]').count();
    
    expect(dmMessageCount).toBe(2);
    expect(playerMessageCount).toBe(2);
    
    await dmContext.close();
    await playerContext.close();
  });

  test('DM should be able to start game', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    // Setup session
    await setupMultiplayerSession(dmPage, playerPage, dmUser, playerUser, playerCharacter.name);
    
    // DM starts game
    await dmPage.getByRole('button', { name: /start game/i }).click();
    
    // Both should see game board
    await dmPage.waitForSelector('[data-testid="game-board"]');
    await playerPage.waitForSelector('[data-testid="game-board"]');
    
    // Verify game started
    const dmGameStarted = await dmPage.locator('[data-testid="game-board"]').isVisible();
    const playerGameStarted = await playerPage.locator('[data-testid="game-board"]').isVisible();
    
    expect(dmGameStarted).toBeTruthy();
    expect(playerGameStarted).toBeTruthy();
    
    await dmContext.close();
    await playerContext.close();
  });

  test('Player should be able to leave session', async ({ page, loginPage, dashboardPage, gameSessionPage }) => {
    // Create session as DM first
    await page.context().clearCookies();
    await loginPage.goto();
    await loginPage.login(dmUser.username, dmUser.password);
    await dashboardPage.createNewSession();
    
    const sessionName = `Leave_Test_${Date.now()}`;
    await gameSessionPage.createSession(sessionName);
    const sessionId = page.url().split('/').pop()!;
    
    // Login as player and join
    await page.context().clearCookies();
    await loginPage.goto();
    await loginPage.login(playerUser.username, playerUser.password);
    await gameSessionPage.joinSession(sessionId);
    await gameSessionPage.selectCharacter(playerCharacter.name);
    
    // Leave session
    await gameSessionPage.leaveSession();
    
    // Should redirect to dashboard
    await page.waitForURL(/\/dashboard/);
    await expect(dashboardPage.welcomeMessage).toBeVisible();
  });

  test('Session should handle disconnection gracefully', async ({ browser }) => {
    const dmContext = await browser.newContext();
    const dmPage = await dmContext.newPage();
    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();
    
    // Setup session
    await setupMultiplayerSession(dmPage, playerPage, dmUser, playerUser, playerCharacter.name);
    
    // Simulate network disconnection for player
    await playerContext.setOffline(true);
    
    // Player should see disconnection message
    await playerPage.waitForSelector('text=/disconnect|offline/i', { timeout: 10000 });
    
    // Restore connection
    await playerContext.setOffline(false);
    
    // Should reconnect automatically
    await playerPage.waitForSelector('[data-testid="player-list"]', { timeout: 10000 });
    
    // Verify still in session
    const playerStillInSession = await playerPage.locator('[data-testid="game-board"], [data-testid="player-list"]').isVisible();
    expect(playerStillInSession).toBeTruthy();
    
    await dmContext.close();
    await playerContext.close();
  });

  test('Session should enforce character selection', async ({ page, loginPage, gameSessionPage }) => {
    // Create session as DM
    await page.context().clearCookies();
    await loginPage.goto();
    await loginPage.login(dmUser.username, dmUser.password);
    await page.goto('/session/create');
    
    const sessionName = `Character_Required_${Date.now()}`;
    await gameSessionPage.createSession(sessionName);
    const sessionId = page.url().split('/').pop()!;
    
    // Create a new user without character
    const noCharUser = testData.generateUser();
    await page.context().clearCookies();
    await page.goto('/register');
    await page.getByLabel('Username').fill(noCharUser.username);
    await page.getByLabel('Email').fill(noCharUser.email);
    await page.getByLabel('Password', { exact: true }).fill(noCharUser.password);
    await page.getByLabel(/confirm password/i).fill(noCharUser.password);
    await page.getByRole('button', { name: /sign up|register|create account/i }).click();
    await page.waitForURL(/\/dashboard/);
    
    // Try to join session without character
    await page.goto(`/session/${sessionId}`);
    
    // Should show character selection or redirect to character creation
    const needsCharacter = await page.locator('text=/create.*character|select.*character/i').isVisible();
    expect(needsCharacter).toBeTruthy();
  });

  test('Session should show participant status', async ({ browser }) => {
    const contexts: BrowserContext[] = [];
    const pages: Page[] = [];
    
    // Create DM and 2 players
    const users = [dmUser, playerUser, { ...testData.generateUser(), character: testData.generateCharacter() }];
    
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();
      contexts.push(context);
      pages.push(page);
      
      if (i === 2) {
        // Create third user
        await page.goto('/register');
        await page.getByLabel('Username').fill(users[2].username);
        await page.getByLabel('Email').fill(users[2].email);
        await page.getByLabel('Password', { exact: true }).fill(users[2].password);
        await page.getByLabel(/confirm password/i).fill(users[2].password);
        await page.getByRole('button', { name: /sign up|register|create account/i }).click();
        await page.waitForURL(/\/dashboard/);
        
        // Create character
        await page.getByRole('button', { name: /create.*character/i }).click();
        await page.getByLabel(/character name/i).fill(users[2].character.name);
        await page.getByLabel(/race/i).selectOption('Elf');
        await page.getByLabel(/class/i).selectOption('Wizard');
        await page.getByLabel(/background/i).selectOption('Sage');
        await page.getByRole('button', { name: /next|continue/i }).click();
        
        // Quick ability scores
        await page.getByRole('button', { name: /roll.*abilities/i }).click();
        await page.getByRole('button', { name: /next|continue/i }).click();
        
        // Skip skills
        await page.getByRole('button', { name: /next|continue/i }).click();
        
        // Finish
        await page.getByRole('button', { name: /finish|create character/i }).click();
        await page.waitForURL(/\/character\/[\w-]+$/);
      }
    }
    
    // DM creates session
    await pages[0].goto('/login');
    await pages[0].getByLabel('Username').fill(dmUser.username);
    await pages[0].getByLabel('Password').fill(dmUser.password);
    await pages[0].getByRole('button', { name: /sign in|log in/i }).click();
    await pages[0].waitForURL(/\/dashboard/);
    
    await pages[0].goto('/session/create');
    const sessionName = `Multi_Player_${Date.now()}`;
    await pages[0].getByLabel(/session name/i).fill(sessionName);
    await pages[0].getByRole('button', { name: /create session/i }).click();
    await pages[0].waitForURL(/\/session\/([\w-]+)$/);
    
    const sessionId = pages[0].url().split('/').pop()!;
    
    // Other players join
    for (let i = 1; i < 3; i++) {
      await pages[i].goto('/login');
      await pages[i].getByLabel('Username').fill(users[i].username);
      await pages[i].getByLabel('Password').fill(users[i].password);
      await pages[i].getByRole('button', { name: /sign in|log in/i }).click();
      await pages[i].waitForURL(/\/dashboard/);
      
      await pages[i].goto(`/session/${sessionId}`);
      const charName = i === 1 ? playerCharacter.name : users[2].character.name;
      await pages[i].getByLabel(/select character/i).selectOption(charName);
      await pages[i].getByRole('button', { name: /ready/i }).click();
    }
    
    // Wait for all players to appear
    await pages[0].waitForFunction(() => {
      const playerList = document.querySelector('[data-testid="player-list"]');
      const playerCount = playerList?.querySelectorAll('[role="listitem"]').length || 0;
      return playerCount === 3;
    });
    
    // Verify all see correct player count
    for (const page of pages) {
      const count = await page.locator('[data-testid="player-list"] [role="listitem"]').count();
      expect(count).toBe(3);
    }
    
    // Cleanup
    for (const context of contexts) {
      await context.close();
    }
  });
});

// Helper function to setup multiplayer session
async function setupMultiplayerSession(
  dmPage: Page,
  playerPage: Page,
  dmUser: any,
  playerUser: any,
  playerCharacterName: string
) {
  // DM creates session
  await dmPage.goto('/login');
  await dmPage.getByLabel('Username').fill(dmUser.username);
  await dmPage.getByLabel('Password').fill(dmUser.password);
  await dmPage.getByRole('button', { name: /sign in|log in/i }).click();
  await dmPage.waitForURL(/\/dashboard/);
  
  await dmPage.goto('/session/create');
  const sessionName = `Test_Session_${Date.now()}`;
  await dmPage.getByLabel(/session name/i).fill(sessionName);
  await dmPage.getByRole('button', { name: /create session/i }).click();
  await dmPage.waitForURL(/\/session\/([\w-]+)$/);
  
  const sessionId = dmPage.url().split('/').pop()!;
  
  // Player joins
  await playerPage.goto('/login');
  await playerPage.getByLabel('Username').fill(playerUser.username);
  await playerPage.getByLabel('Password').fill(playerUser.password);
  await playerPage.getByRole('button', { name: /sign in|log in/i }).click();
  await playerPage.waitForURL(/\/dashboard/);
  
  await playerPage.goto(`/session/${sessionId}`);
  await playerPage.getByLabel(/select character/i).selectOption(playerCharacterName);
  await playerPage.getByRole('button', { name: /ready/i }).click();
  
  // Wait for both to be ready
  await dmPage.waitForSelector(`text="${playerCharacterName}"`);
  await playerPage.waitForSelector(`text="${sessionName}"`);
}