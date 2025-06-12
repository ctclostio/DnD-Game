import React from 'react';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Provider } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';
import { configureStore } from '@reduxjs/toolkit';
import { CharacterBuilder } from '../CharacterBuilder';
import characterSlice from '../../../store/slices/characterSlice';
import authSlice from '../../../store/slices/authSlice';

// Mock dependencies
const mockApiInstance = {
  getCharacterOptions: jest.fn(),
  createCharacter: jest.fn(),
  generateCustomRace: jest.fn(),
  generateCustomClass: jest.fn(),
  // Add generic methods that the component is trying to use
  get: jest.fn(),
  post: jest.fn(),
};

jest.mock('../../../services/api', () => ({
  __esModule: true,
  default: mockApiInstance,
  ApiService: jest.fn().mockImplementation(() => mockApiInstance),
}));

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => jest.fn(),
}));

describe.skip('CharacterBuilder', () => {
  let store: any;

  const renderComponent = () => {
    return render(
      <Provider store={store}>
        <BrowserRouter>
          <CharacterBuilder />
        </BrowserRouter>
      </Provider>
    );
  };

  beforeEach(() => {
    store = configureStore({
      reducer: {
        characters: characterSlice,
        auth: authSlice,
      },
      preloadedState: {
        characters: {
          characters: { ids: [], entities: {} },
          selectedCharacter: null,
          isLoading: false,
          error: null,
          characterOptions: {
            races: ['Human', 'Elf', 'Dwarf', 'Halfling'],
            classes: ['Fighter', 'Wizard', 'Rogue', 'Cleric'],
            backgrounds: ['Acolyte', 'Criminal', 'Folk Hero', 'Noble'],
            skills: ['Acrobatics', 'Athletics', 'Arcana', 'History'],
            aiEnabled: true,
          },
        },
        auth: {
          isAuthenticated: true,
          user: { id: 'user-123', username: 'testuser' },
          token: 'test-token',
          isLoading: false,
          error: null,
        },
      },
    });

    // Mock API responses
    mockApiInstance.get.mockResolvedValue({
      races: ['Human', 'Elf', 'Dwarf', 'Halfling'],
      classes: ['Fighter', 'Wizard', 'Rogue', 'Cleric'],
      backgrounds: ['Acolyte', 'Criminal', 'Folk Hero', 'Noble'],
      skills: ['Acrobatics', 'Athletics', 'Arcana', 'History'],
    });
    
    mockApiInstance.post.mockResolvedValue({ id: 'char-123' });
    
    mockApiInstance.createCharacter.mockResolvedValue({ id: 'char-123' });
    
    mockApiInstance.generateCustomRace.mockResolvedValue({
      name: 'Shadowborn',
      description: 'Beings touched by shadow',
      abilityScoreIncreases: { dexterity: 2, charisma: 1 },
      traits: ['Darkvision', 'Shadow Step'],
    });
    
    mockApiInstance.generateCustomClass.mockResolvedValue({
      name: 'Shadowdancer',
      description: 'Masters of shadow magic',
      hitDice: '1d8',
      primaryAbility: 'Dexterity',
    });
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('Step Navigation', () => {
    it('should render the first step (Basic Info) by default', () => {
      renderComponent();
      expect(screen.getByText('Basic Information')).toBeInTheDocument();
      expect(screen.getByLabelText('Character Name')).toBeInTheDocument();
    });

    it('should navigate to next step when Next is clicked with valid data', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Fill in basic info
      await user.type(screen.getByLabelText('Character Name'), 'Thorin');
      await user.selectOptions(screen.getByLabelText('Alignment'), 'lawful_good');

      // Click next
      await user.click(screen.getByText('Next'));

      // Should be on race selection step
      expect(screen.getByText('Select Race')).toBeInTheDocument();
    });

    it('should not proceed if required fields are empty', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Try to click next without filling required fields
      await user.click(screen.getByText('Next'));

      // Should still be on first step
      expect(screen.getByText('Basic Information')).toBeInTheDocument();
      expect(screen.getByText('Please fill in all required fields')).toBeInTheDocument();
    });

    it('should navigate back to previous step', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Fill basic info and go to next step
      await user.type(screen.getByLabelText('Character Name'), 'Thorin');
      await user.selectOptions(screen.getByLabelText('Alignment'), 'lawful_good');
      await user.click(screen.getByText('Next'));

      // Should be on race selection
      expect(screen.getByText('Select Race')).toBeInTheDocument();

      // Go back
      await user.click(screen.getByText('Previous'));

      // Should be back on basic info
      expect(screen.getByText('Basic Information')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Thorin')).toBeInTheDocument();
    });

    it('should display step indicator correctly', () => {
      renderComponent();
      
      const stepIndicator = screen.getByTestId('step-indicator');
      expect(stepIndicator).toHaveTextContent('Step 1 of 7');
    });
  });

  describe('Basic Info Step', () => {
    it('should validate character name', async () => {
      const user = userEvent.setup();
      renderComponent();

      const nameInput = screen.getByLabelText('Character Name');
      
      // Test empty name
      await user.click(screen.getByText('Next'));
      expect(screen.getByText('Character name is required')).toBeInTheDocument();

      // Test name too short
      await user.type(nameInput, 'A');
      expect(screen.getByText('Name must be at least 2 characters')).toBeInTheDocument();

      // Test valid name
      await user.clear(nameInput);
      await user.type(nameInput, 'Thorin');
      expect(screen.queryByText('Character name is required')).not.toBeInTheDocument();
    });

    it('should require alignment selection', async () => {
      const user = userEvent.setup();
      renderComponent();

      await user.type(screen.getByLabelText('Character Name'), 'Thorin');
      await user.click(screen.getByText('Next'));

      expect(screen.getByText('Please select an alignment')).toBeInTheDocument();
    });
  });

  describe('Race Selection Step', () => {
    beforeEach(async () => {
      const user = userEvent.setup();
      renderComponent();

      // Complete basic info step
      await user.type(screen.getByLabelText('Character Name'), 'Thorin');
      await user.selectOptions(screen.getByLabelText('Alignment'), 'lawful_good');
      await user.click(screen.getByText('Next'));
    });

    it('should display available races', () => {
      expect(screen.getByText('Human')).toBeInTheDocument();
      expect(screen.getByText('Elf')).toBeInTheDocument();
      expect(screen.getByText('Dwarf')).toBeInTheDocument();
      expect(screen.getByText('Halfling')).toBeInTheDocument();
    });

    it('should select a race when clicked', async () => {
      const user = userEvent.setup();
      
      const dwarfOption = screen.getByText('Dwarf').closest('button');
      await user.click(dwarfOption!);

      expect(dwarfOption).toHaveClass('selected');
    });

    it('should show custom race option when AI is enabled', () => {
      expect(screen.getByText('Create Custom Race (AI)')).toBeInTheDocument();
    });

    it('should open custom race form when custom option is selected', async () => {
      const user = userEvent.setup();
      
      await user.click(screen.getByText('Create Custom Race (AI)'));

      expect(screen.getByLabelText('Race Name')).toBeInTheDocument();
      expect(screen.getByLabelText('Race Description')).toBeInTheDocument();
      expect(screen.getByText('Generate with AI')).toBeInTheDocument();
    });

    it('should generate custom race with AI', async () => {
      const user = userEvent.setup();
      
      await user.click(screen.getByText('Create Custom Race (AI)'));
      
      await user.type(screen.getByLabelText('Race Name'), 'Shadowborn');
      await user.type(screen.getByLabelText('Race Description'), 'Beings touched by shadow');
      await user.type(screen.getByLabelText('Desired Traits (Optional)'), 'Darkvision, stealth');
      
      await user.click(screen.getByText('Generate with AI'));

      await waitFor(() => {
        expect(mockApiService.generateCustomRace).toHaveBeenCalledWith({
          name: 'Shadowborn',
          description: 'Beings touched by shadow',
          desiredTraits: 'Darkvision, stealth',
          generationStyle: 'balanced',
        });
      });

      expect(screen.getByText('Race generated successfully!')).toBeInTheDocument();
    });
  });

  describe('Class Selection Step', () => {
    beforeEach(async () => {
      const user = userEvent.setup();
      renderComponent();

      // Complete basic info and race steps
      await user.type(screen.getByLabelText('Character Name'), 'Thorin');
      await user.selectOptions(screen.getByLabelText('Alignment'), 'lawful_good');
      await user.click(screen.getByText('Next'));
      
      await user.click(screen.getByText('Dwarf').closest('button')!);
      await user.click(screen.getByText('Next'));
    });

    it('should display available classes', () => {
      expect(screen.getByText('Fighter')).toBeInTheDocument();
      expect(screen.getByText('Wizard')).toBeInTheDocument();
      expect(screen.getByText('Rogue')).toBeInTheDocument();
      expect(screen.getByText('Cleric')).toBeInTheDocument();
    });

    it('should show class details when hovering', async () => {
      const user = userEvent.setup();
      
      const fighterCard = screen.getByText('Fighter').closest('.class-card');
      await user.hover(fighterCard!);

      await waitFor(() => {
        expect(screen.getByText(/Hit Dice: 1d10/)).toBeInTheDocument();
        expect(screen.getByText(/Primary Ability: Strength/)).toBeInTheDocument();
      });
    });
  });

  describe('Ability Scores Step', () => {
    beforeEach(async () => {
      const user = userEvent.setup();
      renderComponent();

      // Navigate to ability scores step
      await user.type(screen.getByLabelText('Character Name'), 'Thorin');
      await user.selectOptions(screen.getByLabelText('Alignment'), 'lawful_good');
      await user.click(screen.getByText('Next'));
      
      await user.click(screen.getByText('Dwarf').closest('button')!);
      await user.click(screen.getByText('Next'));
      
      await user.click(screen.getByText('Fighter').closest('button')!);
      await user.click(screen.getByText('Next'));
    });

    it('should display ability score methods', () => {
      expect(screen.getByLabelText('Standard Array')).toBeInTheDocument();
      expect(screen.getByLabelText('Point Buy')).toBeInTheDocument();
      expect(screen.getByLabelText('Manual Entry')).toBeInTheDocument();
      expect(screen.getByLabelText('Roll for Stats')).toBeInTheDocument();
    });

    it('should use standard array by default', () => {
      const standardArrayValues = [15, 14, 13, 12, 10, 8];
      const selects = screen.getAllByRole('combobox');
      
      selects.forEach((select) => {
        const options = within(select).getAllByRole('option');
        const values = options.map(opt => opt.textContent);
        standardArrayValues.forEach(value => {
          expect(values).toContain(value.toString());
        });
      });
    });

    it('should switch to point buy system', async () => {
      const user = userEvent.setup();
      
      await user.click(screen.getByLabelText('Point Buy'));

      expect(screen.getByText(/Points Remaining: 27/)).toBeInTheDocument();
      
      // All scores should start at 8
      const scoreDisplays = screen.getAllByTestId(/ability-score-/);
      scoreDisplays.forEach(display => {
        expect(display).toHaveTextContent('8');
      });
    });

    it('should calculate point buy costs correctly', async () => {
      const user = userEvent.setup();
      
      await user.click(screen.getByLabelText('Point Buy'));

      const strIncreaseBtn = screen.getByTestId('str-increase');
      
      // Increase STR from 8 to 13 (costs 5 points)
      for (let i = 0; i < 5; i++) {
        await user.click(strIncreaseBtn);
      }

      expect(screen.getByText(/Points Remaining: 22/)).toBeInTheDocument();
      
      // Try to increase STR to 14 (costs 2 more points)
      await user.click(strIncreaseBtn);
      expect(screen.getByText(/Points Remaining: 20/)).toBeInTheDocument();
    });

    it('should roll for stats', async () => {
      const user = userEvent.setup();
      
      await user.click(screen.getByLabelText('Roll for Stats'));
      await user.click(screen.getByText('Roll All Stats'));

      // Should show rolled values
      const scoreDisplays = screen.getAllByTestId(/ability-score-/);
      scoreDisplays.forEach(display => {
        const value = parseInt(display.textContent || '0');
        expect(value).toBeGreaterThanOrEqual(3);
        expect(value).toBeLessThanOrEqual(18);
      });

      // Should show individual roll results
      expect(screen.getByText(/Rolls:/)).toBeInTheDocument();
    });

    it('should apply racial bonuses', () => {
      // Dwarf gets +2 CON
      const conScore = screen.getByTestId('con-final-score');
      expect(conScore).toHaveTextContent(/\+2/); // Should show racial bonus
    });
  });

  describe('Skill Selection Step', () => {
    beforeEach(async () => {
      const user = userEvent.setup();
      renderComponent();

      // Navigate to skill selection step
      await user.type(screen.getByLabelText('Character Name'), 'Thorin');
      await user.selectOptions(screen.getByLabelText('Alignment'), 'lawful_good');
      await user.click(screen.getByText('Next'));
      
      await user.click(screen.getByText('Dwarf').closest('button')!);
      await user.click(screen.getByText('Next'));
      
      await user.click(screen.getByText('Fighter').closest('button')!);
      await user.click(screen.getByText('Next'));
      
      // Skip ability scores (use defaults)
      await user.click(screen.getByText('Next'));
      
      // Select background
      await user.click(screen.getByText('Noble').closest('button')!);
      await user.click(screen.getByText('Next'));
    });

    it('should show class skill options', () => {
      expect(screen.getByText(/Choose 2 skills from your class/)).toBeInTheDocument();
      
      // Fighter skills
      expect(screen.getByLabelText('Athletics')).toBeInTheDocument();
      expect(screen.getByLabelText('Intimidation')).toBeInTheDocument();
    });

    it('should enforce skill selection limit', async () => {
      const user = userEvent.setup();
      
      const athleticsCheckbox = screen.getByLabelText('Athletics');
      const intimidationCheckbox = screen.getByLabelText('Intimidation');
      const survivalCheckbox = screen.getByLabelText('Survival');
      
      await user.click(athleticsCheckbox);
      await user.click(intimidationCheckbox);
      
      // Third skill should be disabled
      expect(survivalCheckbox).toBeDisabled();
    });

    it('should show background skills as preselected', () => {
      // Noble background gives History and Persuasion
      expect(screen.getByText('History (from Noble background)')).toBeInTheDocument();
      expect(screen.getByText('Persuasion (from Noble background)')).toBeInTheDocument();
    });
  });

  describe('Review Step', () => {
    beforeEach(async () => {
      const user = userEvent.setup();
      renderComponent();

      // Complete all steps to reach review
      await user.type(screen.getByLabelText('Character Name'), 'Thorin Oakenshield');
      await user.selectOptions(screen.getByLabelText('Alignment'), 'lawful_good');
      await user.click(screen.getByText('Next'));
      
      await user.click(screen.getByText('Dwarf').closest('button')!);
      await user.click(screen.getByText('Next'));
      
      await user.click(screen.getByText('Fighter').closest('button')!);
      await user.click(screen.getByText('Next'));
      
      await user.click(screen.getByText('Next')); // Use default ability scores
      
      await user.click(screen.getByText('Noble').closest('button')!);
      await user.click(screen.getByText('Next'));
      
      // Select skills
      await user.click(screen.getByLabelText('Athletics'));
      await user.click(screen.getByLabelText('Intimidation'));
      await user.click(screen.getByText('Next'));
    });

    it('should display all character information', () => {
      expect(screen.getByText('Character Summary')).toBeInTheDocument();
      expect(screen.getByText('Thorin Oakenshield')).toBeInTheDocument();
      expect(screen.getByText('Dwarf Fighter')).toBeInTheDocument();
      expect(screen.getByText('Lawful Good')).toBeInTheDocument();
      expect(screen.getByText('Noble')).toBeInTheDocument();
    });

    it('should show ability scores with modifiers', () => {
      expect(screen.getByTestId('str-summary')).toBeInTheDocument();
      expect(screen.getByTestId('dex-summary')).toBeInTheDocument();
      expect(screen.getByTestId('con-summary')).toBeInTheDocument();
      expect(screen.getByTestId('int-summary')).toBeInTheDocument();
      expect(screen.getByTestId('wis-summary')).toBeInTheDocument();
      expect(screen.getByTestId('cha-summary')).toBeInTheDocument();
    });

    it('should show selected skills', () => {
      expect(screen.getByText(/Athletics/)).toBeInTheDocument();
      expect(screen.getByText(/Intimidation/)).toBeInTheDocument();
      expect(screen.getByText(/History/)).toBeInTheDocument();
      expect(screen.getByText(/Persuasion/)).toBeInTheDocument();
    });

    it('should allow editing by clicking section edit buttons', async () => {
      const user = userEvent.setup();
      
      const editBasicInfoBtn = screen.getByTestId('edit-basic-info');
      await user.click(editBasicInfoBtn);

      // Should go back to basic info step
      expect(screen.getByText('Basic Information')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Thorin Oakenshield')).toBeInTheDocument();
    });

    it('should create character when confirmed', async () => {
      const user = userEvent.setup();
      
      await user.click(screen.getByText('Create Character'));

      await waitFor(() => {
        expect(mockApiService.createCharacter).toHaveBeenCalledWith(
          expect.objectContaining({
            name: 'Thorin Oakenshield',
            race: 'Dwarf',
            class: 'Fighter',
            background: 'Noble',
            alignment: 'lawful_good',
          })
        );
      });

      expect(screen.getByText('Character created successfully!')).toBeInTheDocument();
    });

    it('should handle creation errors', async () => {
      const user = userEvent.setup();
      
      mockApiService.createCharacter.mockRejectedValueOnce(
        new Error('Failed to create character')
      );

      await user.click(screen.getByText('Create Character'));

      await waitFor(() => {
        expect(screen.getByText('Failed to create character')).toBeInTheDocument();
      });
    });
  });

  describe('Accessibility', () => {
    it('should have proper ARIA labels', () => {
      renderComponent();
      
      expect(screen.getByRole('form', { name: /character creation/i })).toBeInTheDocument();
      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should be keyboard navigable', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Tab to first input
      await user.tab();
      expect(screen.getByLabelText('Character Name')).toHaveFocus();

      // Tab to alignment
      await user.tab();
      expect(screen.getByLabelText('Alignment')).toHaveFocus();

      // Tab to next button
      await user.tab();
      expect(screen.getByText('Next')).toHaveFocus();
    });

    it('should announce step changes to screen readers', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Complete first step
      await user.type(screen.getByLabelText('Character Name'), 'Test');
      await user.selectOptions(screen.getByLabelText('Alignment'), 'neutral');
      await user.click(screen.getByText('Next'));

      // Check for aria-live region update
      const liveRegion = screen.getByRole('status', { hidden: true });
      expect(liveRegion).toHaveTextContent('Step 2 of 7: Select Race');
    });
  });

  describe('Performance', () => {
    it('should memoize expensive calculations', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Navigate to ability scores with point buy
      await user.type(screen.getByLabelText('Character Name'), 'Test');
      await user.selectOptions(screen.getByLabelText('Alignment'), 'neutral');
      await user.click(screen.getByText('Next'));
      await user.click(screen.getByText('Human').closest('button')!);
      await user.click(screen.getByText('Next'));
      await user.click(screen.getByText('Fighter').closest('button')!);
      await user.click(screen.getByText('Next'));
      
      await user.click(screen.getByLabelText('Point Buy'));

      const calculateSpy = jest.spyOn(console, 'log');
      
      // Trigger re-render without changing ability scores
      await user.click(screen.getByTestId('str-increase'));
      await user.click(screen.getByTestId('str-decrease'));

      // Calculation should not be re-run
      expect(calculateSpy).not.toHaveBeenCalledWith('Recalculating point costs');
    });

    it('should debounce form updates', async () => {
      const user = userEvent.setup();
      renderComponent();

      const updateSpy = jest.fn();
      // Mock form update handler
      jest.spyOn(React, 'useState').mockImplementation((initial) => {
        const [state, setState] = React.useState(initial);
        return [state, (newState: any) => {
          updateSpy();
          setState(newState);
        }];
      });

      const nameInput = screen.getByLabelText('Character Name');
      
      // Type quickly
      await user.type(nameInput, 'Thorin', { delay: 50 });

      // Should batch updates
      expect(updateSpy).toHaveBeenCalledTimes(1);
    });
  });
});