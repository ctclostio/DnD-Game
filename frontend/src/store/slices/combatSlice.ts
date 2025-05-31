import { createSlice } from '@reduxjs/toolkit';

interface CombatState {
  combatActive: boolean;
  initiative: any[];
  currentTurn: number;
  isLoading: boolean;
  error: string | null;
}

const initialState: CombatState = {
  combatActive: false,
  initiative: [],
  currentTurn: 0,
  isLoading: false,
  error: null,
};

const combatSlice = createSlice({
  name: 'combat',
  initialState,
  reducers: {},
});

export default combatSlice.reducer;