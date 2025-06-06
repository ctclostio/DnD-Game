import { test, expect, testData } from '../fixtures/base';

test.describe('Character Creation Flow', () => {
  let testUser: ReturnType<typeof testData.generateUser>;

  test.beforeEach(async ({ registerPage, page }) => {
    // Create and login test user
    testUser = testData.generateUser();
    await registerPage.goto();
    await registerPage.register(testUser.username, testUser.email, testUser.password);
    await page.waitForURL(/\/dashboard/);
  });

  test('should create a character using step-by-step wizard', async ({ 
    dashboardPage, 
    characterBuilderPage,
    page 
  }) => {
    const character = testData.generateCharacter();
    
    // Start character creation
    await dashboardPage.createNewCharacter();
    await page.waitForURL(/\/character\/new/);
    
    // Step 1: Basic Info
    await expect(characterBuilderPage.stepIndicator).toContainText('Step 1');
    await characterBuilderPage.fillBasicInfo({
      name: character.name,
      race: character.race,
      class: character.class,
      background: character.background,
      alignment: 'Neutral Good',
    });
    
    // Step 2: Ability Scores
    await expect(characterBuilderPage.stepIndicator).toContainText('Step 2');
    await characterBuilderPage.setAbilityScores(character.abilities);
    
    // Step 3: Skills
    await expect(characterBuilderPage.stepIndicator).toContainText('Step 3');
    const fighterSkills = ['Athletics', 'Intimidation'];
    await characterBuilderPage.selectSkills(fighterSkills);
    
    // Step 4: Review
    await expect(characterBuilderPage.stepIndicator).toContainText('Step 4');
    await characterBuilderPage.reviewAndFinish();
    
    // Should redirect to character page
    await characterBuilderPage.expectSuccess();
    
    // Verify character appears in dashboard
    await page.goto('/dashboard');
    await dashboardPage.waitForLoad();
    const characterCount = await dashboardPage.getCharacterCount();
    expect(characterCount).toBeGreaterThan(0);
  });

  test('should create character with rolled ability scores', async ({ 
    dashboardPage, 
    characterBuilderPage,
    page 
  }) => {
    const character = testData.generateCharacter();
    
    await dashboardPage.createNewCharacter();
    
    // Basic Info
    await characterBuilderPage.fillBasicInfo({
      name: character.name,
      race: 'Elf',
      class: 'Wizard',
      background: 'Sage',
    });
    
    // Roll ability scores
    await characterBuilderPage.rollAbilityScores();
    
    // Skills
    const wizardSkills = ['Arcana', 'Investigation'];
    await characterBuilderPage.selectSkills(wizardSkills);
    
    // Review and finish
    await characterBuilderPage.reviewAndFinish();
    await characterBuilderPage.expectSuccess();
  });

  test('should allow navigation between steps', async ({ 
    dashboardPage, 
    characterBuilderPage,
    page 
  }) => {
    await dashboardPage.createNewCharacter();
    
    // Fill first step
    await characterBuilderPage.fillBasicInfo({
      name: 'Test Character',
      race: 'Dwarf',
      class: 'Cleric',
      background: 'Acolyte',
    });
    
    // Should be on step 2
    let currentStep = await characterBuilderPage.getCurrentStep();
    expect(currentStep).toBe(2);
    
    // Go back to step 1
    await characterBuilderPage.backButton.click();
    currentStep = await characterBuilderPage.getCurrentStep();
    expect(currentStep).toBe(1);
    
    // Verify data is preserved
    const nameValue = await characterBuilderPage.nameInput.inputValue();
    expect(nameValue).toBe('Test Character');
    
    // Go forward again
    await characterBuilderPage.nextButton.click();
    currentStep = await characterBuilderPage.getCurrentStep();
    expect(currentStep).toBe(2);
  });

  test('should validate required fields', async ({ 
    dashboardPage, 
    characterBuilderPage,
    page 
  }) => {
    await dashboardPage.createNewCharacter();
    
    // Try to proceed without filling required fields
    await characterBuilderPage.nextButton.click();
    
    // Should show validation errors
    await expect(characterBuilderPage.nameInput).toHaveAttribute('aria-invalid', 'true');
    
    // Fill only name and try again
    await characterBuilderPage.nameInput.fill('Incomplete Character');
    await characterBuilderPage.nextButton.click();
    
    // Should still show errors for other fields
    await expect(characterBuilderPage.raceSelect).toHaveAttribute('aria-invalid', 'true');
    await expect(characterBuilderPage.classSelect).toHaveAttribute('aria-invalid', 'true');
  });

  test('should enforce ability score limits', async ({ 
    dashboardPage, 
    characterBuilderPage,
    page 
  }) => {
    await dashboardPage.createNewCharacter();
    
    // Fill basic info
    await characterBuilderPage.fillBasicInfo({
      name: 'Score Test',
      race: 'Human',
      class: 'Fighter',
      background: 'Soldier',
    });
    
    // Try to set invalid ability scores
    await characterBuilderPage.strengthInput.fill('25'); // Above max
    await characterBuilderPage.dexterityInput.fill('0'); // Below min
    
    // Values should be clamped
    await characterBuilderPage.nextButton.click();
    
    // Go back to verify
    await characterBuilderPage.backButton.click();
    
    const strengthValue = await characterBuilderPage.strengthInput.inputValue();
    const dexterityValue = await characterBuilderPage.dexterityInput.inputValue();
    
    expect(parseInt(strengthValue)).toBeLessThanOrEqual(20);
    expect(parseInt(dexterityValue)).toBeGreaterThanOrEqual(3);
  });

  test('should apply racial bonuses correctly', async ({ 
    dashboardPage, 
    characterBuilderPage,
    page 
  }) => {
    await dashboardPage.createNewCharacter();
    
    // Select Human (gets +1 to all abilities)
    await characterBuilderPage.fillBasicInfo({
      name: 'Human Test',
      race: 'Human',
      class: 'Fighter',
      background: 'Soldier',
    });
    
    // Set base scores
    const baseScores = {
      strength: 15,
      dexterity: 14,
      constitution: 13,
      intelligence: 12,
      wisdom: 11,
      charisma: 10,
    };
    
    await characterBuilderPage.setAbilityScores(baseScores);
    
    // Continue to review
    await characterBuilderPage.selectSkills(['Athletics']);
    
    // Check final scores in review
    const reviewText = await characterBuilderPage.reviewSection.textContent();
    
    // Human gets +1 to all, so should see 16, 15, 14, 13, 12, 11
    expect(reviewText).toContain('16'); // Strength
    expect(reviewText).toContain('15'); // Dexterity
  });

  test('should limit skill selection based on class', async ({ 
    dashboardPage, 
    characterBuilderPage,
    page 
  }) => {
    await dashboardPage.createNewCharacter();
    
    // Create a Rogue (can choose 4 skills)
    await characterBuilderPage.fillBasicInfo({
      name: 'Skill Test',
      race: 'Halfling',
      class: 'Rogue',
      background: 'Criminal',
    });
    
    await characterBuilderPage.setAbilityScores({
      strength: 10,
      dexterity: 16,
      constitution: 12,
      intelligence: 14,
      wisdom: 13,
      charisma: 11,
    });
    
    // Try to select more than allowed skills
    const rogueSkills = ['Stealth', 'Sleight of Hand', 'Investigation', 'Perception', 'Deception'];
    
    for (const skill of rogueSkills) {
      const checkbox = characterBuilderPage.page.getByRole('checkbox', { name: new RegExp(skill, 'i') });
      await checkbox.check();
    }
    
    // Should only allow 4 skills to be selected
    const checkedBoxes = await characterBuilderPage.skillCheckboxes.filter({ hasText: /checked/i }).count();
    expect(checkedBoxes).toBeLessThanOrEqual(4);
  });

  test('should save character to database', async ({ 
    dashboardPage, 
    characterBuilderPage,
    page 
  }) => {
    const character = testData.generateCharacter();
    
    await dashboardPage.createNewCharacter();
    
    // Create character
    await characterBuilderPage.fillBasicInfo({
      name: character.name,
      race: character.race,
      class: character.class,
      background: character.background,
    });
    
    await characterBuilderPage.setAbilityScores(character.abilities);
    await characterBuilderPage.selectSkills(['Athletics', 'Survival']);
    await characterBuilderPage.reviewAndFinish();
    
    // Wait for save
    await page.waitForResponse(response => 
      response.url().includes('/api/characters') && response.status() === 201
    );
    
    // Verify character is saved
    await page.goto('/dashboard');
    await dashboardPage.selectCharacter(character.name);
    
    // Should navigate to character sheet
    await page.waitForURL(/\/character\/[\w-]+$/);
    await expect(page.getByText(character.name)).toBeVisible();
  });

  test('should handle API errors gracefully', async ({ 
    dashboardPage, 
    characterBuilderPage,
    page 
  }) => {
    // Intercept API call to simulate error
    await page.route('**/api/characters', route => {
      route.fulfill({
        status: 500,
        body: JSON.stringify({ error: 'Server error' }),
      });
    });
    
    const character = testData.generateCharacter();
    
    await dashboardPage.createNewCharacter();
    
    // Fill all steps
    await characterBuilderPage.fillBasicInfo({
      name: character.name,
      race: character.race,
      class: character.class,
      background: character.background,
    });
    
    await characterBuilderPage.setAbilityScores(character.abilities);
    await characterBuilderPage.selectSkills(['Athletics']);
    
    // Try to finish
    await characterBuilderPage.reviewAndFinish();
    
    // Should show error message
    const errorMessage = page.getByRole('alert');
    await expect(errorMessage).toBeVisible();
    await expect(errorMessage).toContainText('error');
    
    // Should not redirect
    expect(page.url()).toContain('/character/new');
  });

  test('should create custom race character', async ({ 
    dashboardPage, 
    characterBuilderPage,
    page 
  }) => {
    await dashboardPage.createNewCharacter();
    
    // Select custom race option
    await characterBuilderPage.raceSelect.selectOption('custom');
    
    // Fill custom race form (if it appears)
    const customRaceDialog = page.getByRole('dialog', { name: /custom race/i });
    if (await customRaceDialog.isVisible()) {
      await page.getByLabel(/race name/i).fill('Dragonkin');
      await page.getByLabel(/size/i).selectOption('Medium');
      await page.getByLabel(/speed/i).fill('30');
      await page.getByRole('button', { name: /save/i }).click();
    }
    
    // Continue with character creation
    await characterBuilderPage.nameInput.fill('Custom Race Hero');
    await characterBuilderPage.classSelect.selectOption('Fighter');
    await characterBuilderPage.backgroundSelect.selectOption('Soldier');
    await characterBuilderPage.nextButton.click();
    
    await characterBuilderPage.setAbilityScores({
      strength: 16,
      dexterity: 14,
      constitution: 15,
      intelligence: 10,
      wisdom: 12,
      charisma: 8,
    });
    
    await characterBuilderPage.selectSkills(['Athletics']);
    await characterBuilderPage.reviewAndFinish();
    
    await characterBuilderPage.expectSuccess();
  });
});