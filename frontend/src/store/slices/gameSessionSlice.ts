import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import { GameSessionState, EntityState } from '../../types/state';
import { GameSession, CombatParticipant } from '../../types/game';
import apiService from '../../services/api';

const initialState: GameSessionState = {
  sessions: {
    ids: [],
    entities: {},
  },
  currentSessionId: null,
  isConnected: false,
  connectionError: null,
  isLoading: {},
  errors: {},
};

// Helper function to create entity state
function createEntityState<T>(items: T[], getId: (item: T) => string): EntityState<T> {
  const ids = items.map(getId);
  const entities = items.reduce((acc, item) => {
    acc[getId(item)] = item;
    return acc;
  }, {} as { [id: string]: T });
  
  return { ids, entities };
}

// Async thunks
export const fetchSessions = createAsyncThunk(
  'gameSession/fetchSessions',
  async () => {
    const response = await apiService.getGameSessions();
    return response;
  }
);

export const createSession = createAsyncThunk(
  'gameSession/create',
  async (sessionData: { name: string; campaignId?: string }) => {
    const response = await apiService.createGameSession(sessionData);
    return response;
  }
);

export const joinSession = createAsyncThunk(
  'gameSession/join',
  async (sessionId: string) => {
    const response = await apiService.joinGameSession(sessionId);
    return response;
  }
);

export const leaveSession = createAsyncThunk(
  'gameSession/leave',
  async (sessionId: string) => {
    await apiService.leaveGameSession(sessionId);
    return sessionId;
  }
);

export const updateSession = createAsyncThunk(
  'gameSession/update',
  async ({ sessionId, updates }: { sessionId: string; updates: Partial<GameSession> }) => {
    const response = await apiService.updateGameSession(sessionId, updates);
    return response;
  }
);

const gameSessionSlice = createSlice({
  name: 'gameSession',
  initialState,
  reducers: {
    // Sync actions for WebSocket updates
    sessionUpdated: (state, action: PayloadAction<GameSession>) => {
      const session = action.payload;
      if (state.sessions.entities[session.id]) {
        state.sessions.entities[session.id] = session;
      }
    },
    
    playerJoined: (state, action: PayloadAction<{ sessionId: string; playerId: string }>) => {
      const { sessionId, playerId } = action.payload;
      const session = state.sessions.entities[sessionId];
      if (session) {
        if (!session.playerIds.includes(playerId)) {
          // Create a new array to ensure Immer detects the change
          session.playerIds = [...session.playerIds, playerId];
        }
      }
    },
    
    playerLeft: (state, action: PayloadAction<{ sessionId: string; playerId: string }>) => {
      const { sessionId, playerId } = action.payload;
      const session = state.sessions.entities[sessionId];
      if (session) {
        session.playerIds = session.playerIds.filter(id => id !== playerId);
      }
    },
    
    combatStarted: (state, action: PayloadAction<{ sessionId: string; participants: CombatParticipant[] }>) => {
      const { sessionId } = action.payload;
      const session = state.sessions.entities[sessionId];
      if (session) {
        session.combatActive = true;
      }
    },
    
    combatEnded: (state, action: PayloadAction<string>) => {
      const sessionId = action.payload;
      const session = state.sessions.entities[sessionId];
      if (session) {
        session.combatActive = false;
      }
    },
    
    setCurrentSession: (state, action: PayloadAction<string | null>) => {
      state.currentSessionId = action.payload;
    },
    
    setConnected: (state, action: PayloadAction<boolean>) => {
      state.isConnected = action.payload;
      if (action.payload) {
        state.connectionError = null;
      }
    },
    
    setConnectionError: (state, action: PayloadAction<string>) => {
      state.connectionError = action.payload;
      state.isConnected = false;
    },
    
    clearSessions: (state) => {
      state.sessions = {
        ids: [],
        entities: {},
      };
      state.currentSessionId = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch sessions
      .addCase(fetchSessions.pending, (state) => {
        state.isLoading.fetchSessions = true;
        state.errors.fetchSessions = null;
      })
      .addCase(fetchSessions.fulfilled, (state, action) => {
        state.isLoading.fetchSessions = false;
        state.sessions = createEntityState(action.payload, s => s.id);
      })
      .addCase(fetchSessions.rejected, (state, action) => {
        state.isLoading.fetchSessions = false;
        state.errors.fetchSessions = action.error.message || 'Failed to fetch sessions';
      })
      
      // Create session
      .addCase(createSession.pending, (state) => {
        state.isLoading.createSession = true;
        state.errors.createSession = null;
      })
      .addCase(createSession.fulfilled, (state, action) => {
        state.isLoading.createSession = false;
        const session = action.payload;
        state.sessions.ids.push(session.id);
        state.sessions.entities[session.id] = session;
        state.currentSessionId = session.id;
      })
      .addCase(createSession.rejected, (state, action) => {
        state.isLoading.createSession = false;
        state.errors.createSession = action.error.message || 'Failed to create session';
      })
      
      // Join session
      .addCase(joinSession.pending, (state) => {
        state.isLoading.joinSession = true;
        state.errors.joinSession = null;
      })
      .addCase(joinSession.fulfilled, (state, action) => {
        state.isLoading.joinSession = false;
        const session = action.payload;
        if (!state.sessions.ids.includes(session.id)) {
          state.sessions.ids.push(session.id);
        }
        state.sessions.entities[session.id] = session;
        state.currentSessionId = session.id;
      })
      .addCase(joinSession.rejected, (state, action) => {
        state.isLoading.joinSession = false;
        state.errors.joinSession = action.error.message || 'Failed to join session';
      })
      
      // Leave session
      .addCase(leaveSession.pending, (state) => {
        state.isLoading.leaveSession = true;
        state.errors.leaveSession = null;
      })
      .addCase(leaveSession.fulfilled, (state, action) => {
        state.isLoading.leaveSession = false;
        const sessionId = action.payload;
        if (state.currentSessionId === sessionId) {
          state.currentSessionId = null;
        }
      })
      .addCase(leaveSession.rejected, (state, action) => {
        state.isLoading.leaveSession = false;
        state.errors.leaveSession = action.error.message || 'Failed to leave session';
      })
      
      // Update session
      .addCase(updateSession.pending, (state) => {
        state.isLoading.updateSession = true;
        state.errors.updateSession = null;
      })
      .addCase(updateSession.fulfilled, (state, action) => {
        state.isLoading.updateSession = false;
        const session = action.payload;
        state.sessions.entities[session.id] = session;
      })
      .addCase(updateSession.rejected, (state, action) => {
        state.isLoading.updateSession = false;
        state.errors.updateSession = action.error.message || 'Failed to update session';
      });
  },
});

export const {
  sessionUpdated,
  playerJoined,
  playerLeft,
  combatStarted,
  combatEnded,
  setCurrentSession,
  setConnected,
  setConnectionError,
  clearSessions,
} = gameSessionSlice.actions;

export default gameSessionSlice.reducer;
