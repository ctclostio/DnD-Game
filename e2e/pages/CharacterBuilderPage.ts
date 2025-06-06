import { Page, Locator } from '@playwright/test';

export class CharacterBuilderPage {
  readonly page: Page;
  // Step indicators
  readonly stepIndicator: Locator;
  readonly nextButton: Locator;
  readonly backButton: Locator;
  readonly finishButton: Locator;
  
  // Basic Info Step
  readonly nameInput: Locator;
  readonly raceSelect: Locator;
  readonly classSelect: Locator;
  readonly backgroundSelect: Locator;
  readonly alignmentSelect: Locator;
  
  // Ability Scores Step
  readonly strengthInput: Locator;
  readonly dexterityInput: Locator;
  readonly constitutionInput: Locator;
  readonly intelligenceInput: Locator;
  readonly wisdomInput: Locator;
  readonly charismaInput: Locator;
  readonly rollAbilitiesButton: Locator;
  readonly pointBuyToggle: Locator;
  
  // Skills Step
  readonly skillCheckboxes: Locator;
  readonly expertiseSelects: Locator;
  
  // Review Step
  readonly reviewSection: Locator;
  readonly editButtons: Locator;

  constructor(page: Page) {
    this.page = page;
    
    // Navigation
    this.stepIndicator = page.getByTestId('step-indicator');
    this.nextButton = page.getByRole('button', { name: /next|continue/i });
    this.backButton = page.getByRole('button', { name: /back|previous/i });
    this.finishButton = page.getByRole('button', { name: /finish|create character/i });
    
    // Basic Info
    this.nameInput = page.getByLabel(/character name/i);
    this.raceSelect = page.getByLabel(/race/i);
    this.classSelect = page.getByLabel(/class/i);
    this.backgroundSelect = page.getByLabel(/background/i);
    this.alignmentSelect = page.getByLabel(/alignment/i);
    
    // Ability Scores
    this.strengthInput = page.getByLabel(/strength/i);
    this.dexterityInput = page.getByLabel(/dexterity/i);
    this.constitutionInput = page.getByLabel(/constitution/i);
    this.intelligenceInput = page.getByLabel(/intelligence/i);
    this.wisdomInput = page.getByLabel(/wisdom/i);
    this.charismaInput = page.getByLabel(/charisma/i);
    this.rollAbilitiesButton = page.getByRole('button', { name: /roll.*abilities/i });
    this.pointBuyToggle = page.getByLabel(/point buy/i);
    
    // Skills
    this.skillCheckboxes = page.getByRole('checkbox');
    this.expertiseSelects = page.getByLabel(/expertise/i);
    
    // Review
    this.reviewSection = page.getByTestId('character-review');
    this.editButtons = page.getByRole('button', { name: /edit/i });
  }

  async goto() {
    await this.page.goto('/character/new');
  }

  async fillBasicInfo(character: {
    name: string;
    race: string;
    class: string;
    background: string;
    alignment?: string;
  }) {
    await this.nameInput.fill(character.name);
    await this.raceSelect.selectOption(character.race);
    await this.classSelect.selectOption(character.class);
    await this.backgroundSelect.selectOption(character.background);
    
    if (character.alignment) {
      await this.alignmentSelect.selectOption(character.alignment);
    }
    
    await this.nextButton.click();
  }

  async setAbilityScores(scores: {
    strength: number;
    dexterity: number;
    constitution: number;
    intelligence: number;
    wisdom: number;
    charisma: number;
  }) {
    await this.strengthInput.fill(scores.strength.toString());
    await this.dexterityInput.fill(scores.dexterity.toString());
    await this.constitutionInput.fill(scores.constitution.toString());
    await this.intelligenceInput.fill(scores.intelligence.toString());
    await this.wisdomInput.fill(scores.wisdom.toString());
    await this.charismaInput.fill(scores.charisma.toString());
    
    await this.nextButton.click();
  }

  async rollAbilityScores() {
    await this.rollAbilitiesButton.click();
    await this.page.waitForTimeout(500); // Wait for animation
    await this.nextButton.click();
  }

  async selectSkills(skills: string[]) {
    for (const skill of skills) {
      const checkbox = this.page.getByRole('checkbox', { name: new RegExp(skill, 'i') });
      await checkbox.check();
    }
    
    await this.nextButton.click();
  }

  async reviewAndFinish() {
    // Wait for review section to load
    await this.reviewSection.waitFor({ state: 'visible' });
    
    // Verify all sections are present
    await this.page.getByText(/ability scores/i).waitFor();
    await this.page.getByText(/skills/i).waitFor();
    
    await this.finishButton.click();
  }

  async getCurrentStep(): Promise<number> {
    const stepText = await this.stepIndicator.textContent();
    const match = stepText?.match(/step (\d+)/i);
    return match ? parseInt(match[1]) : 1;
  }

  async goToStep(stepNumber: number) {
    const currentStep = await this.getCurrentStep();
    
    if (stepNumber > currentStep) {
      for (let i = currentStep; i < stepNumber; i++) {
        await this.nextButton.click();
      }
    } else if (stepNumber < currentStep) {
      for (let i = currentStep; i > stepNumber; i--) {
        await this.backButton.click();
      }
    }
  }

  async expectSuccess() {
    // Wait for redirect to character page or success message
    await this.page.waitForURL(/\/character\/[^\/]+$/);
    return true;
  }
}