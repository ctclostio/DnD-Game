import { createSlice } from '@reduxjs/toolkit';

interface GameSessionState {
  currentSession: any | null;
  sessions: any[];
  isLoading: boolean;
  error: string | null;
}

const initialState: GameSessionState = {
  currentSession: null,
  sessions: [],
  isLoading: false,
  error: null,
};

const gameSessionSlice = createSlice({
  name: 'gameSession',
  initialState,
  reducers: {},
});

export default gameSessionSlice.reducer;