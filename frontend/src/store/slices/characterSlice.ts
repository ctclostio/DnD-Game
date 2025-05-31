import { createSlice } from '@reduxjs/toolkit';

interface CharacterState {
  characters: any[];
  currentCharacter: any | null;
  isLoading: boolean;
  error: string | null;
}

const initialState: CharacterState = {
  characters: [],
  currentCharacter: null,
  isLoading: false,
  error: null,
};

const characterSlice = createSlice({
  name: 'character',
  initialState,
  reducers: {},
});

export default characterSlice.reducer;