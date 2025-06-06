import { test as base } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { RegisterPage } from '../pages/RegisterPage';
import { DashboardPage } from '../pages/DashboardPage';
import { CharacterBuilderPage } from '../pages/CharacterBuilderPage';
import { GameSessionPage } from '../pages/GameSessionPage';
import { CombatPage } from '../pages/CombatPage';

// Define custom fixtures
type MyFixtures = {
  loginPage: LoginPage;
  registerPage: RegisterPage;
  dashboardPage: DashboardPage;
  characterBuilderPage: CharacterBuilderPage;
  gameSessionPage: GameSessionPage;
  combatPage: CombatPage;
};

// Extend base test with our fixtures
export const test = base.extend<MyFixtures>({
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page));
  },
  
  registerPage: async ({ page }, use) => {
    await use(new RegisterPage(page));
  },
  
  dashboardPage: async ({ page }, use) => {
    await use(new DashboardPage(page));
  },
  
  characterBuilderPage: async ({ page }, use) => {
    await use(new CharacterBuilderPage(page));
  },
  
  gameSessionPage: async ({ page }, use) => {
    await use(new GameSessionPage(page));
  },
  
  combatPage: async ({ page }, use) => {
    await use(new CombatPage(page));
  },
});

export { expect } from '@playwright/test';

// Test data generators
export const testData = {
  generateUser: () => ({
    username: `testuser_${Date.now()}`,
    email: `test_${Date.now()}@example.com`,
    password: 'TestPassword123!',
  }),
  
  generateCharacter: () => ({
    name: `Hero_${Date.now()}`,
    race: 'Human',
    class: 'Fighter',
    background: 'Soldier',
    abilities: {
      strength: 16,
      dexterity: 14,
      constitution: 15,
      intelligence: 10,
      wisdom: 12,
      charisma: 8,
    },
  }),
  
  generateSession: () => ({
    name: `Adventure_${Date.now()}`,
    description: 'A test adventure session',
  }),
};

// Common helpers
export const helpers = {
  waitForAPIResponse: async (page: any, endpoint: string) => {
    return page.waitForResponse((response: any) => 
      response.url().includes(endpoint) && response.status() === 200
    );
  },
  
  waitForWebSocket: async (page: any) => {
    return page.waitForFunction(() => {
      // Check if WebSocket is connected in the Redux store
      const state = (window as any).__REDUX_STORE__?.getState();
      return state?.websocket?.connected === true;
    });
  },
};