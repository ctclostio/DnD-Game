import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { CharacterState } from '../../types/state';
import { Character } from '../../types/game';

const initialState: CharacterState = {
  characters: {
    ids: [],
    entities: {},
  },
  currentCharacterId: null,
  isLoading: {},
  errors: {},
};

const characterSlice = createSlice({
  name: 'character',
  initialState,
  reducers: {
    addCharacter: (state, action: PayloadAction<Character>) => {
      const character = action.payload;
      state.characters.ids.push(character.id);
      state.characters.entities[character.id] = character;
    },
    updateCharacter: (state, action: PayloadAction<{ id: string; changes: Partial<Character> }>) => {
      const { id, changes } = action.payload;
      if (state.characters.entities[id]) {
        state.characters.entities[id] = {
          ...state.characters.entities[id],
          ...changes,
        };
      }
    },
    removeCharacter: (state, action: PayloadAction<string>) => {
      const id = action.payload;
      state.characters.ids = state.characters.ids.filter(cid => cid !== id);
      delete state.characters.entities[id];
      if (state.currentCharacterId === id) {
        state.currentCharacterId = null;
      }
    },
    setCurrentCharacter: (state, action: PayloadAction<string | null>) => {
      state.currentCharacterId = action.payload;
    },
  },
});

export const { addCharacter, updateCharacter, removeCharacter, setCurrentCharacter } = characterSlice.actions;
export default characterSlice.reducer;