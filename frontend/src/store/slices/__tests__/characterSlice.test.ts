import { configureStore } from '@reduxjs/toolkit';
import characterReducer, {
  addCharacter,
  updateCharacter,
  removeCharacter,
  setCurrentCharacter,
} from '../characterSlice';
import { Character } from '../../../types/game';

describe('characterSlice', () => {
  let store: ReturnType<typeof configureStore>;

  const createMockCharacter = (id: string, name: string): Character => ({
    id,
    name,
    level: 1,
    experience: 0,
    hitPoints: 10,
    maxHitPoints: 10,
    armorClass: 10,
    initiative: 0,
    speed: 30,
    proficiencyBonus: 2,
    race: 'Human',
    class: 'Fighter',
    background: 'Soldier',
    alignment: 'Neutral',
    abilities: {
      strength: 10,
      dexterity: 10,
      constitution: 10,
      intelligence: 10,
      wisdom: 10,
      charisma: 10,
    },
    savingThrows: {
      strength: 0,
      dexterity: 0,
      constitution: 0,
      intelligence: 0,
      wisdom: 0,
      charisma: 0,
    },
    skills: {},
    features: [],
    inventory: [],
    spells: [],
    equipment: {
      armor: null,
      mainHand: null,
      offHand: null,
      accessories: [],
    },
  });

  beforeEach(() => {
    store = configureStore({
      reducer: {
        character: characterReducer,
      },
    });
  });

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const state = store.getState().character;
      
      expect(state).toEqual({
        characters: {
          ids: [],
          entities: {},
        },
        currentCharacterId: null,
        isLoading: {},
        errors: {},
      });
    });
  });

  describe('addCharacter', () => {
    it('should add a new character', () => {
      const character = createMockCharacter('char-1', 'Aragorn');

      store.dispatch(addCharacter(character));

      const state = store.getState().character;
      expect(state.characters.ids).toContain('char-1');
      expect(state.characters.entities['char-1']).toEqual(character);
    });

    it('should add multiple characters', () => {
      const character1 = createMockCharacter('char-1', 'Aragorn');
      const character2 = createMockCharacter('char-2', 'Legolas');

      store.dispatch(addCharacter(character1));
      store.dispatch(addCharacter(character2));

      const state = store.getState().character;
      expect(state.characters.ids).toHaveLength(2);
      expect(state.characters.ids).toEqual(['char-1', 'char-2']);
      expect(state.characters.entities['char-1']).toEqual(character1);
      expect(state.characters.entities['char-2']).toEqual(character2);
    });

    it('should handle duplicate character ids', () => {
      const character1 = createMockCharacter('char-1', 'Aragorn');
      const character2 = createMockCharacter('char-1', 'Aragorn Updated');

      store.dispatch(addCharacter(character1));
      store.dispatch(addCharacter(character2));

      const state = store.getState().character;
      // Should have duplicate id in array (not ideal but testing current behavior)
      expect(state.characters.ids).toEqual(['char-1', 'char-1']);
      // Entity should be overwritten
      expect(state.characters.entities['char-1'].name).toBe('Aragorn Updated');
    });
  });

  describe('updateCharacter', () => {
    beforeEach(() => {
      const character = createMockCharacter('char-1', 'Aragorn');
      store.dispatch(addCharacter(character));
    });

    it('should update character properties', () => {
      store.dispatch(updateCharacter({
        id: 'char-1',
        changes: {
          level: 5,
          experience: 6500,
          hitPoints: 45,
          maxHitPoints: 45,
        },
      }));

      const state = store.getState().character;
      const updatedChar = state.characters.entities['char-1'];
      
      expect(updatedChar.level).toBe(5);
      expect(updatedChar.experience).toBe(6500);
      expect(updatedChar.hitPoints).toBe(45);
      expect(updatedChar.maxHitPoints).toBe(45);
      // Other properties should remain unchanged
      expect(updatedChar.name).toBe('Aragorn');
      expect(updatedChar.race).toBe('Human');
    });

    it('should update nested properties', () => {
      store.dispatch(updateCharacter({
        id: 'char-1',
        changes: {
          abilities: {
            strength: 18,
            dexterity: 14,
            constitution: 16,
            intelligence: 10,
            wisdom: 12,
            charisma: 8,
          },
        },
      }));

      const state = store.getState().character;
      const updatedChar = state.characters.entities['char-1'];
      
      expect(updatedChar.abilities.strength).toBe(18);
      expect(updatedChar.abilities.dexterity).toBe(14);
    });

    it('should not update non-existent character', () => {
      store.dispatch(updateCharacter({
        id: 'non-existent',
        changes: { level: 10 },
      }));

      const state = store.getState().character;
      expect(state.characters.entities['non-existent']).toBeUndefined();
      expect(state.characters.ids).not.toContain('non-existent');
    });

    it('should handle partial updates', () => {
      store.dispatch(updateCharacter({
        id: 'char-1',
        changes: {
          name: 'Strider',
        },
      }));

      const state = store.getState().character;
      const updatedChar = state.characters.entities['char-1'];
      
      expect(updatedChar.name).toBe('Strider');
      // All other properties should remain
      expect(updatedChar.level).toBe(1);
      expect(updatedChar.race).toBe('Human');
      expect(updatedChar.class).toBe('Fighter');
    });
  });

  describe('removeCharacter', () => {
    beforeEach(() => {
      const char1 = createMockCharacter('char-1', 'Aragorn');
      const char2 = createMockCharacter('char-2', 'Legolas');
      const char3 = createMockCharacter('char-3', 'Gimli');
      
      store.dispatch(addCharacter(char1));
      store.dispatch(addCharacter(char2));
      store.dispatch(addCharacter(char3));
    });

    it('should remove a character', () => {
      store.dispatch(removeCharacter('char-2'));

      const state = store.getState().character;
      expect(state.characters.ids).toEqual(['char-1', 'char-3']);
      expect(state.characters.entities['char-2']).toBeUndefined();
      expect(Object.keys(state.characters.entities)).toHaveLength(2);
    });

    it('should clear currentCharacterId if removed character was current', () => {
      store.dispatch(setCurrentCharacter('char-2'));
      
      let state = store.getState().character;
      expect(state.currentCharacterId).toBe('char-2');

      store.dispatch(removeCharacter('char-2'));

      state = store.getState().character;
      expect(state.currentCharacterId).toBeNull();
    });

    it('should not affect currentCharacterId if different character removed', () => {
      store.dispatch(setCurrentCharacter('char-1'));
      store.dispatch(removeCharacter('char-2'));

      const state = store.getState().character;
      expect(state.currentCharacterId).toBe('char-1');
    });

    it('should handle removing non-existent character', () => {
      const initialState = store.getState().character;
      
      store.dispatch(removeCharacter('non-existent'));

      const state = store.getState().character;
      expect(state).toEqual(initialState);
    });

    it('should handle removing all characters', () => {
      store.dispatch(removeCharacter('char-1'));
      store.dispatch(removeCharacter('char-2'));
      store.dispatch(removeCharacter('char-3'));

      const state = store.getState().character;
      expect(state.characters.ids).toEqual([]);
      expect(state.characters.entities).toEqual({});
    });
  });

  describe('setCurrentCharacter', () => {
    beforeEach(() => {
      const char1 = createMockCharacter('char-1', 'Aragorn');
      const char2 = createMockCharacter('char-2', 'Legolas');
      
      store.dispatch(addCharacter(char1));
      store.dispatch(addCharacter(char2));
    });

    it('should set current character', () => {
      store.dispatch(setCurrentCharacter('char-1'));

      const state = store.getState().character;
      expect(state.currentCharacterId).toBe('char-1');
    });

    it('should change current character', () => {
      store.dispatch(setCurrentCharacter('char-1'));
      store.dispatch(setCurrentCharacter('char-2'));

      const state = store.getState().character;
      expect(state.currentCharacterId).toBe('char-2');
    });

    it('should clear current character', () => {
      store.dispatch(setCurrentCharacter('char-1'));
      store.dispatch(setCurrentCharacter(null));

      const state = store.getState().character;
      expect(state.currentCharacterId).toBeNull();
    });

    it('should allow setting non-existent character as current', () => {
      // This might not be ideal behavior but testing current implementation
      store.dispatch(setCurrentCharacter('non-existent'));

      const state = store.getState().character;
      expect(state.currentCharacterId).toBe('non-existent');
    });
  });

  describe('complex scenarios', () => {
    it('should handle a complete character lifecycle', () => {
      // Create character
      const character = createMockCharacter('char-1', 'Hero');
      store.dispatch(addCharacter(character));

      // Set as current
      store.dispatch(setCurrentCharacter('char-1'));

      // Level up
      store.dispatch(updateCharacter({
        id: 'char-1',
        changes: {
          level: 2,
          experience: 300,
          maxHitPoints: 18,
          hitPoints: 18,
        },
      }));

      // Take damage
      store.dispatch(updateCharacter({
        id: 'char-1',
        changes: {
          hitPoints: 10,
        },
      }));

      let state = store.getState().character;
      expect(state.characters.entities['char-1'].hitPoints).toBe(10);
      expect(state.characters.entities['char-1'].maxHitPoints).toBe(18);

      // Remove character
      store.dispatch(removeCharacter('char-1'));

      state = store.getState().character;
      expect(state.characters.ids).toEqual([]);
      expect(state.currentCharacterId).toBeNull();
    });

    it('should handle multiple characters with updates', () => {
      const chars = [
        createMockCharacter('char-1', 'Aragorn'),
        createMockCharacter('char-2', 'Legolas'),
        createMockCharacter('char-3', 'Gimli'),
      ];

      chars.forEach(char => store.dispatch(addCharacter(char)));

      // Update all characters
      chars.forEach((_, index) => {
        store.dispatch(updateCharacter({
          id: `char-${index + 1}`,
          changes: { level: index + 2 },
        }));
      });

      const state = store.getState().character;
      expect(state.characters.entities['char-1'].level).toBe(2);
      expect(state.characters.entities['char-2'].level).toBe(3);
      expect(state.characters.entities['char-3'].level).toBe(4);
    });
  });

  describe('selectors (implicit testing)', () => {
    it('should allow selecting all characters', () => {
      const char1 = createMockCharacter('char-1', 'Aragorn');
      const char2 = createMockCharacter('char-2', 'Legolas');
      
      store.dispatch(addCharacter(char1));
      store.dispatch(addCharacter(char2));

      const state = store.getState().character;
      const allCharacters = state.characters.ids.map(id => state.characters.entities[id]);
      
      expect(allCharacters).toHaveLength(2);
      expect(allCharacters[0].name).toBe('Aragorn');
      expect(allCharacters[1].name).toBe('Legolas');
    });

    it('should allow selecting current character', () => {
      const character = createMockCharacter('char-1', 'Aragorn');
      store.dispatch(addCharacter(character));
      store.dispatch(setCurrentCharacter('char-1'));

      const state = store.getState().character;
      const currentCharacter = state.currentCharacterId 
        ? state.characters.entities[state.currentCharacterId]
        : null;
      
      expect(currentCharacter).toEqual(character);
    });
  });
});