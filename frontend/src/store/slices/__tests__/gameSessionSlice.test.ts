import { configureStore } from '@reduxjs/toolkit';
import gameSessionReducer, {
  fetchSessions,
  createSession,
  joinSession,
  leaveSession,
  updateSession,
  sessionUpdated,
  playerJoined,
  playerLeft,
  combatStarted,
  combatEnded,
  setCurrentSession,
  setConnected,
  setConnectionError,
  clearSessions,
} from '../gameSessionSlice';
import { GameSession, CombatParticipant } from '../../../types/game';
import apiService from '../../../services/api';

// Mock the API module
jest.mock('../../../services/api', () => ({
  __esModule: true,
  default: {
    getGameSessions: jest.fn(),
    createGameSession: jest.fn(),
    joinGameSession: jest.fn(),
    leaveGameSession: jest.fn(),
    updateGameSession: jest.fn(),
  },
  // Also export individual functions for backward compatibility
  getGameSessions: jest.fn(),
  createGameSession: jest.fn(),
  joinGameSession: jest.fn(),
  leaveGameSession: jest.fn(),
  updateGameSession: jest.fn(),
}));

describe('gameSessionSlice', () => {
  let store: ReturnType<typeof configureStore>;

  const createMockSession = (id: string, name: string): GameSession => ({
    id,
    name,
    dmId: 'dm-1',
    playerIds: ['player-1', 'player-2'],
    campaignId: 'campaign-1',
    combatActive: false,
    currentRound: undefined,
    combatHistory: [],
    sessionNotes: 'Test session notes',
    sharedResources: {},
    mapData: undefined,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  });

  const createMockParticipant = (id: string, name: string): CombatParticipant => ({
    id,
    name,
    initiative: 10,
    initiativeModifier: 2,
    armorClass: 15,
    hitPointsMax: 20,
    hitPointsCurrent: 20,
    temporaryHitPoints: 0,
    conditions: [],
    isPlayer: true,
    isActive: true,
    hasActed: false,
    hasBonusActed: false,
    hasReacted: false,
    movementUsed: 0,
    movementMax: 30,
    concentrating: false,
  });

  beforeEach(() => {
    jest.clearAllMocks();
    store = configureStore({
      reducer: {
        gameSession: gameSessionReducer,
      },
    });
  });

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const state = store.getState().gameSession;
      
      expect(state).toEqual({
        sessions: {
          ids: [],
          entities: {},
        },
        currentSessionId: null,
        isConnected: false,
        connectionError: null,
        isLoading: {},
        errors: {},
      });
    });
  });

  describe('fetchSessions', () => {
    it('should fetch sessions successfully', async () => {
      const mockSessions = [
        createMockSession('session-1', 'Adventure 1'),
        createMockSession('session-2', 'Adventure 2'),
      ];
      
      (apiService.getGameSessions as jest.Mock).mockResolvedValue(mockSessions);

      await store.dispatch(fetchSessions());

      const state = store.getState().gameSession;
      expect(state.sessions.ids).toEqual(['session-1', 'session-2']);
      expect(state.sessions.entities['session-1']).toEqual(mockSessions[0]);
      expect(state.sessions.entities['session-2']).toEqual(mockSessions[1]);
      expect(state.isLoading.fetchSessions).toBe(false);
      expect(state.errors.fetchSessions).toBeNull();
    });

    it('should handle fetch sessions failure', async () => {
      const errorMessage = 'Network error';
      (apiService.getGameSessions as jest.Mock).mockRejectedValue(new Error(errorMessage));

      await store.dispatch(fetchSessions());

      const state = store.getState().gameSession;
      expect(state.sessions.ids).toEqual([]);
      expect(state.isLoading.fetchSessions).toBe(false);
      expect(state.errors.fetchSessions).toBe(errorMessage);
    });

    it('should set loading state while fetching', () => {
      (apiService.getGameSessions as jest.Mock).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      store.dispatch(fetchSessions());

      const state = store.getState().gameSession;
      expect(state.isLoading.fetchSessions).toBe(true);
      expect(state.errors.fetchSessions).toBeNull();
    });
  });

  describe('createSession', () => {
    it('should create a new session', async () => {
      const sessionData = { name: 'New Adventure', campaignId: 'campaign-1' };
      const mockSession = createMockSession('session-3', 'New Adventure');
      
      (apiService.createGameSession as jest.Mock).mockResolvedValue(mockSession);

      await store.dispatch(createSession(sessionData));

      const state = store.getState().gameSession;
      expect(state.sessions.ids).toContain('session-3');
      expect(state.sessions.entities['session-3']).toEqual(mockSession);
      expect(state.currentSessionId).toBe('session-3');
      expect(state.isLoading.createSession).toBe(false);
    });

    it('should handle create session failure', async () => {
      const sessionData = { name: 'New Adventure' };
      const errorMessage = 'Failed to create session';
      
      (apiService.createGameSession as jest.Mock).mockRejectedValue(new Error(errorMessage));

      await store.dispatch(createSession(sessionData));

      const state = store.getState().gameSession;
      expect(state.errors.createSession).toBe(errorMessage);
      expect(state.isLoading.createSession).toBe(false);
      expect(state.currentSessionId).toBeNull();
    });
  });

  describe('joinSession', () => {
    it('should join an existing session', async () => {
      const mockSession = createMockSession('session-1', 'Adventure 1');
      
      (apiService.joinGameSession as jest.Mock).mockResolvedValue(mockSession);

      await store.dispatch(joinSession('session-1'));

      const state = store.getState().gameSession;
      expect(state.sessions.ids).toContain('session-1');
      expect(state.sessions.entities['session-1']).toEqual(mockSession);
      expect(state.currentSessionId).toBe('session-1');
    });

    it('should update existing session when joining', async () => {
      // Add initial session
      const initialSession = createMockSession('session-1', 'Adventure 1');
      store.dispatch(sessionUpdated(initialSession));

      // Join with updated data
      const updatedSession = {
        ...initialSession,
        playerIds: [...initialSession.playerIds, 'player-3'],
      };
      
      (apiService.joinGameSession as jest.Mock).mockResolvedValue(updatedSession);

      await store.dispatch(joinSession('session-1'));

      const state = store.getState().gameSession;
      expect(state.sessions.entities['session-1'].playerIds).toContain('player-3');
    });

    it('should handle join session failure', async () => {
      const errorMessage = 'Session full';
      
      (apiService.joinGameSession as jest.Mock).mockRejectedValue(new Error(errorMessage));

      await store.dispatch(joinSession('session-1'));

      const state = store.getState().gameSession;
      expect(state.errors.joinSession).toBe(errorMessage);
      expect(state.currentSessionId).toBeNull();
    });
  });

  describe('leaveSession', () => {
    beforeEach(async () => {
      const mockSession = createMockSession('session-1', 'Adventure 1');
      (apiService.joinGameSession as jest.Mock).mockResolvedValue(mockSession);
      await store.dispatch(joinSession('session-1'));
    });

    it('should leave current session', async () => {
      (apiService.leaveGameSession as jest.Mock).mockResolvedValue(undefined);

      await store.dispatch(leaveSession('session-1'));

      const state = store.getState().gameSession;
      expect(state.currentSessionId).toBeNull();
      expect(state.isLoading.leaveSession).toBe(false);
    });

    it('should not change current session if leaving different session', async () => {
      (apiService.leaveGameSession as jest.Mock).mockResolvedValue(undefined);

      await store.dispatch(leaveSession('session-2'));

      const state = store.getState().gameSession;
      expect(state.currentSessionId).toBe('session-1');
    });

    it('should handle leave session failure', async () => {
      const errorMessage = 'Cannot leave session';
      
      (apiService.leaveGameSession as jest.Mock).mockRejectedValue(new Error(errorMessage));

      await store.dispatch(leaveSession('session-1'));

      const state = store.getState().gameSession;
      expect(state.errors.leaveSession).toBe(errorMessage);
      expect(state.currentSessionId).toBe('session-1'); // Should remain in session
    });
  });

  describe('updateSession', () => {
    beforeEach(() => {
      const mockSession = createMockSession('session-1', 'Adventure 1');
      store.dispatch(sessionUpdated(mockSession));
    });

    it('should update session data', async () => {
      const updates = { sessionNotes: 'Updated notes' };
      const updatedSession = {
        ...createMockSession('session-1', 'Adventure 1'),
        ...updates,
      };
      
      (apiService.updateGameSession as jest.Mock).mockResolvedValue(updatedSession);

      await store.dispatch(updateSession({ sessionId: 'session-1', updates }));

      const state = store.getState().gameSession;
      expect(state.sessions.entities['session-1'].sessionNotes).toBe('Updated notes');
    });

    it('should handle update session failure', async () => {
      const updates = { sessionNotes: 'Updated notes' };
      const errorMessage = 'Update failed';
      
      (apiService.updateGameSession as jest.Mock).mockRejectedValue(new Error(errorMessage));

      await store.dispatch(updateSession({ sessionId: 'session-1', updates }));

      const state = store.getState().gameSession;
      expect(state.errors.updateSession).toBe(errorMessage);
    });
  });

  describe('sync actions', () => {
    beforeEach(async () => {
      // First, we need to populate the store with sessions
      const mockSessions = [
        createMockSession('session-1', 'Adventure 1'),
        createMockSession('session-2', 'Adventure 2'),
      ];
      
      (apiService.getGameSessions as jest.Mock).mockResolvedValue(mockSessions);
      await store.dispatch(fetchSessions());
    });

    describe('sessionUpdated', () => {
      it('should update existing session', () => {
        const updatedSession = {
          ...createMockSession('session-1', 'Adventure 1'),
          name: 'Updated Adventure',
        };

        store.dispatch(sessionUpdated(updatedSession));

        const state = store.getState().gameSession;
        expect(state.sessions.entities['session-1'].name).toBe('Updated Adventure');
      });

      it('should not add non-existent session', () => {
        const newSession = createMockSession('session-3', 'New Adventure');

        store.dispatch(sessionUpdated(newSession));

        const state = store.getState().gameSession;
        expect(state.sessions.ids).not.toContain('session-3');
      });
    });

    describe('playerJoined', () => {
      it('should add player to session', () => {
        store.dispatch(playerJoined({ sessionId: 'session-1', playerId: 'player-3' }));

        const state = store.getState().gameSession;
        expect(state.sessions.entities['session-1'].playerIds).toContain('player-3');
      });

      it('should not add duplicate player', () => {
        store.dispatch(playerJoined({ sessionId: 'session-1', playerId: 'player-1' }));

        const state = store.getState().gameSession;
        const playerIds = state.sessions.entities['session-1'].playerIds;
        expect(playerIds.filter(id => id === 'player-1')).toHaveLength(1);
      });

      it('should handle non-existent session', () => {
        store.dispatch(playerJoined({ sessionId: 'non-existent', playerId: 'player-3' }));

        const state = store.getState().gameSession;
        expect(state.sessions.ids).not.toContain('non-existent');
      });
    });

    describe('playerLeft', () => {
      it('should remove player from session', () => {
        store.dispatch(playerLeft({ sessionId: 'session-1', playerId: 'player-2' }));

        const state = store.getState().gameSession;
        expect(state.sessions.entities['session-1'].playerIds).not.toContain('player-2');
        expect(state.sessions.entities['session-1'].playerIds).toHaveLength(1);
      });

      it('should handle non-existent player', () => {
        const initialPlayerIds = store.getState().gameSession.sessions.entities['session-1'].playerIds;
        
        store.dispatch(playerLeft({ sessionId: 'session-1', playerId: 'non-existent' }));

        const state = store.getState().gameSession;
        expect(state.sessions.entities['session-1'].playerIds).toEqual(initialPlayerIds);
      });
    });

    describe('combatStarted', () => {
      it('should set combat active', () => {
        const participants = [
          createMockParticipant('char-1', 'Hero 1'),
          createMockParticipant('char-2', 'Hero 2'),
        ];

        store.dispatch(combatStarted({ sessionId: 'session-1', participants }));

        const state = store.getState().gameSession;
        expect(state.sessions.entities['session-1'].combatActive).toBe(true);
      });
    });

    describe('combatEnded', () => {
      it('should set combat inactive', () => {
        // First start combat
        store.dispatch(combatStarted({ 
          sessionId: 'session-1', 
          participants: [createMockParticipant('char-1', 'Hero')] 
        }));

        store.dispatch(combatEnded('session-1'));

        const state = store.getState().gameSession;
        expect(state.sessions.entities['session-1'].combatActive).toBe(false);
      });
    });

    describe('setCurrentSession', () => {
      it('should set current session', () => {
        store.dispatch(setCurrentSession('session-2'));

        const state = store.getState().gameSession;
        expect(state.currentSessionId).toBe('session-2');
      });

      it('should clear current session', () => {
        store.dispatch(setCurrentSession('session-1'));
        store.dispatch(setCurrentSession(null));

        const state = store.getState().gameSession;
        expect(state.currentSessionId).toBeNull();
      });
    });

    describe('connection actions', () => {
      it('should set connected state', () => {
        store.dispatch(setConnected(true));

        const state = store.getState().gameSession;
        expect(state.isConnected).toBe(true);
        expect(state.connectionError).toBeNull();
      });

      it('should clear error when connected', () => {
        store.dispatch(setConnectionError('Connection lost'));
        store.dispatch(setConnected(true));

        const state = store.getState().gameSession;
        expect(state.isConnected).toBe(true);
        expect(state.connectionError).toBeNull();
      });

      it('should set connection error', () => {
        store.dispatch(setConnected(true));
        store.dispatch(setConnectionError('WebSocket error'));

        const state = store.getState().gameSession;
        expect(state.isConnected).toBe(false);
        expect(state.connectionError).toBe('WebSocket error');
      });
    });

    describe('clearSessions', () => {
      it('should clear all sessions and current session', () => {
        store.dispatch(setCurrentSession('session-1'));
        store.dispatch(clearSessions());

        const state = store.getState().gameSession;
        expect(state.sessions.ids).toEqual([]);
        expect(state.sessions.entities).toEqual({});
        expect(state.currentSessionId).toBeNull();
      });
    });
  });

  describe('complex scenarios', () => {
    it('should properly handle playerJoined action', async () => {
      // First create a session
      const mockSession = createMockSession('test-session', 'Test Session');
      (apiService.createGameSession as jest.Mock).mockResolvedValue(mockSession);
      
      await store.dispatch(createSession({ name: 'Test Session' }));
      
      // Check initial state
      let state = store.getState().gameSession;
      expect(state.sessions.entities['test-session']).toBeDefined();
      expect(state.sessions.entities['test-session'].playerIds).toEqual(['player-1', 'player-2']);
      
      // Add a player
      store.dispatch(playerJoined({ sessionId: 'test-session', playerId: 'new-player' }));
      
      // Check updated state
      state = store.getState().gameSession;
      expect(state.sessions.entities['test-session'].playerIds).toEqual(['player-1', 'player-2', 'new-player']);
    });

    it('should handle complete session lifecycle', async () => {
      // Create session
      const sessionData = { name: 'Epic Campaign' };
      const mockSession = createMockSession('session-1', 'Epic Campaign');
      (apiService.createGameSession as jest.Mock).mockResolvedValue(mockSession);
      
      await store.dispatch(createSession(sessionData));

      // Player joins
      store.dispatch(playerJoined({ sessionId: 'session-1', playerId: 'player-3' }));

      // Start combat
      const participants = [
        createMockParticipant('char-1', 'Hero'),
        createMockParticipant('char-2', 'Monster'),
      ];
      store.dispatch(combatStarted({ sessionId: 'session-1', participants }));

      // Update session
      // Get current session state to preserve playerIds changes
      const currentSessionState = store.getState().gameSession.sessions.entities['session-1'];
      const updatedSession = {
        ...currentSessionState,
        sessionNotes: 'Combat in progress',
      };
      (apiService.updateGameSession as jest.Mock).mockResolvedValue(updatedSession);
      await store.dispatch(updateSession({ 
        sessionId: 'session-1', 
        updates: { sessionNotes: 'Combat in progress' } 
      }));

      // End combat
      store.dispatch(combatEnded('session-1'));

      // Player leaves
      store.dispatch(playerLeft({ sessionId: 'session-1', playerId: 'player-2' }));

      // Verify final state
      const state = store.getState().gameSession;
      const session = state.sessions.entities['session-1'];
      
      expect(session).toBeDefined();
      expect(session.playerIds).toContain('player-3');
      expect(session.playerIds).not.toContain('player-2');
      expect(session.combatActive).toBe(false);
      expect(session.sessionNotes).toBe('Combat in progress');
    });

    it('should handle multiple concurrent sessions', async () => {
      const sessions = [
        createMockSession('session-1', 'Campaign 1'),
        createMockSession('session-2', 'Campaign 2'),
        createMockSession('session-3', 'Campaign 3'),
      ];

      (apiService.getGameSessions as jest.Mock).mockResolvedValue(sessions);
      await store.dispatch(fetchSessions());

      // Join one session
      (apiService.joinGameSession as jest.Mock).mockResolvedValue(sessions[1]);
      await store.dispatch(joinSession('session-2'));

      // Update different sessions
      store.dispatch(playerJoined({ sessionId: 'session-1', playerId: 'new-player' }));
      store.dispatch(combatStarted({ 
        sessionId: 'session-3', 
        participants: [createMockParticipant('char-1', 'Hero')] 
      }));

      const state = store.getState().gameSession;
      
      expect(state.currentSessionId).toBe('session-2');
      expect(state.sessions.entities['session-1'].playerIds).toContain('new-player');
      expect(state.sessions.entities['session-3'].combatActive).toBe(true);
      expect(state.sessions.entities['session-2'].combatActive).toBe(false);
    });

    it('should handle connection loss and recovery', () => {
      // Establish connection
      store.dispatch(setConnected(true));
      
      // Lose connection
      store.dispatch(setConnectionError('Connection lost'));
      
      let state = store.getState().gameSession;
      expect(state.isConnected).toBe(false);
      expect(state.connectionError).toBe('Connection lost');
      
      // Recover connection
      store.dispatch(setConnected(true));
      
      state = store.getState().gameSession;
      expect(state.isConnected).toBe(true);
      expect(state.connectionError).toBeNull();
    });
  });
});