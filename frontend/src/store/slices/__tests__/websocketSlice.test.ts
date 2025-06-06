import { configureStore } from '@reduxjs/toolkit';
import websocketReducer, {
  connected,
  disconnected,
  error,
  updateRoomParticipants,
  leaveRoom,
} from '../websocketSlice';

describe('websocketSlice', () => {
  let store: ReturnType<typeof configureStore>;

  beforeEach(() => {
    store = configureStore({
      reducer: {
        websocket: websocketReducer,
      },
    });
  });

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const state = store.getState().websocket;
      
      expect(state).toEqual({
        connected: false,
        reconnecting: false,
        error: null,
        rooms: {},
      });
    });
  });

  describe('connected', () => {
    it('should set connected state and create room', () => {
      store.dispatch(connected({ roomId: 'session-123' }));

      const state = store.getState().websocket;
      expect(state.connected).toBe(true);
      expect(state.reconnecting).toBe(false);
      expect(state.error).toBeNull();
      expect(state.rooms['session-123']).toEqual({
        connected: true,
        participants: [],
      });
    });

    it('should clear error on connection', () => {
      // Set an error first
      store.dispatch(error({ error: 'Connection failed' }));
      
      // Connect
      store.dispatch(connected({ roomId: 'session-456' }));

      const state = store.getState().websocket;
      expect(state.error).toBeNull();
    });

    it('should handle multiple room connections', () => {
      store.dispatch(connected({ roomId: 'room-1' }));
      store.dispatch(connected({ roomId: 'room-2' }));
      store.dispatch(connected({ roomId: 'room-3' }));

      const state = store.getState().websocket;
      expect(Object.keys(state.rooms)).toHaveLength(3);
      expect(state.rooms['room-1'].connected).toBe(true);
      expect(state.rooms['room-2'].connected).toBe(true);
      expect(state.rooms['room-3'].connected).toBe(true);
    });

    it('should overwrite existing room on reconnection', () => {
      // Initial connection with participants
      store.dispatch(connected({ roomId: 'session-123' }));
      store.dispatch(updateRoomParticipants({ 
        roomId: 'session-123', 
        participants: ['user-1', 'user-2'] 
      }));

      // Reconnect to same room
      store.dispatch(connected({ roomId: 'session-123' }));

      const state = store.getState().websocket;
      // Participants should be reset
      expect(state.rooms['session-123'].participants).toEqual([]);
    });
  });

  describe('disconnected', () => {
    beforeEach(() => {
      // Setup connected state with rooms
      store.dispatch(connected({ roomId: 'room-1' }));
      store.dispatch(connected({ roomId: 'room-2' }));
      store.dispatch(updateRoomParticipants({ 
        roomId: 'room-1', 
        participants: ['user-1', 'user-2'] 
      }));
    });

    it('should set disconnected and reconnecting state', () => {
      store.dispatch(disconnected());

      const state = store.getState().websocket;
      expect(state.connected).toBe(false);
      expect(state.reconnecting).toBe(true);
    });

    it('should disconnect all rooms but preserve them', () => {
      store.dispatch(disconnected());

      const state = store.getState().websocket;
      expect(state.rooms['room-1'].connected).toBe(false);
      expect(state.rooms['room-2'].connected).toBe(false);
      // Participants should be preserved
      expect(state.rooms['room-1'].participants).toEqual(['user-1', 'user-2']);
    });

    it('should handle disconnection with no rooms', () => {
      // Clear rooms first
      store.dispatch(leaveRoom('room-1'));
      store.dispatch(leaveRoom('room-2'));

      store.dispatch(disconnected());

      const state = store.getState().websocket;
      expect(state.connected).toBe(false);
      expect(state.reconnecting).toBe(true);
      expect(state.rooms).toEqual({});
    });
  });

  describe('error', () => {
    it('should set error message', () => {
      const errorMessage = 'WebSocket connection failed';
      store.dispatch(error({ error: errorMessage }));

      const state = store.getState().websocket;
      expect(state.error).toBe(errorMessage);
    });

    it('should overwrite previous error', () => {
      store.dispatch(error({ error: 'First error' }));
      store.dispatch(error({ error: 'Second error' }));

      const state = store.getState().websocket;
      expect(state.error).toBe('Second error');
    });

    it('should not affect connection state', () => {
      store.dispatch(connected({ roomId: 'room-1' }));
      store.dispatch(error({ error: 'Some error occurred' }));

      const state = store.getState().websocket;
      expect(state.connected).toBe(true);
      expect(state.error).toBe('Some error occurred');
    });
  });

  describe('updateRoomParticipants', () => {
    beforeEach(() => {
      store.dispatch(connected({ roomId: 'game-session-1' }));
    });

    it('should update participants for existing room', () => {
      const participants = ['player-1', 'player-2', 'dm-1'];
      store.dispatch(updateRoomParticipants({ 
        roomId: 'game-session-1', 
        participants 
      }));

      const state = store.getState().websocket;
      expect(state.rooms['game-session-1'].participants).toEqual(participants);
    });

    it('should handle empty participants list', () => {
      store.dispatch(updateRoomParticipants({ 
        roomId: 'game-session-1', 
        participants: ['player-1'] 
      }));
      
      store.dispatch(updateRoomParticipants({ 
        roomId: 'game-session-1', 
        participants: [] 
      }));

      const state = store.getState().websocket;
      expect(state.rooms['game-session-1'].participants).toEqual([]);
    });

    it('should not create room if it does not exist', () => {
      store.dispatch(updateRoomParticipants({ 
        roomId: 'non-existent-room', 
        participants: ['user-1'] 
      }));

      const state = store.getState().websocket;
      expect(state.rooms['non-existent-room']).toBeUndefined();
    });

    it('should update participants multiple times', () => {
      store.dispatch(updateRoomParticipants({ 
        roomId: 'game-session-1', 
        participants: ['player-1'] 
      }));
      
      store.dispatch(updateRoomParticipants({ 
        roomId: 'game-session-1', 
        participants: ['player-1', 'player-2'] 
      }));
      
      store.dispatch(updateRoomParticipants({ 
        roomId: 'game-session-1', 
        participants: ['player-1', 'player-2', 'player-3'] 
      }));

      const state = store.getState().websocket;
      expect(state.rooms['game-session-1'].participants).toHaveLength(3);
    });
  });

  describe('leaveRoom', () => {
    beforeEach(() => {
      store.dispatch(connected({ roomId: 'room-1' }));
      store.dispatch(connected({ roomId: 'room-2' }));
      store.dispatch(connected({ roomId: 'room-3' }));
    });

    it('should remove specific room', () => {
      store.dispatch(leaveRoom('room-2'));

      const state = store.getState().websocket;
      expect(state.rooms['room-1']).toBeDefined();
      expect(state.rooms['room-2']).toBeUndefined();
      expect(state.rooms['room-3']).toBeDefined();
    });

    it('should handle leaving non-existent room', () => {
      const roomsBefore = { ...store.getState().websocket.rooms };
      
      store.dispatch(leaveRoom('non-existent'));

      const state = store.getState().websocket;
      expect(state.rooms).toEqual(roomsBefore);
    });

    it('should leave all rooms', () => {
      store.dispatch(leaveRoom('room-1'));
      store.dispatch(leaveRoom('room-2'));
      store.dispatch(leaveRoom('room-3'));

      const state = store.getState().websocket;
      expect(state.rooms).toEqual({});
    });

    it('should not affect connection state', () => {
      store.dispatch(leaveRoom('room-1'));

      const state = store.getState().websocket;
      expect(state.connected).toBe(true);
    });
  });

  describe('complex scenarios', () => {
    it('should handle connection lifecycle', () => {
      // Initial connection
      store.dispatch(connected({ roomId: 'game-123' }));
      store.dispatch(updateRoomParticipants({ 
        roomId: 'game-123', 
        participants: ['user-1', 'user-2'] 
      }));

      let state = store.getState().websocket;
      expect(state.connected).toBe(true);
      expect(state.rooms['game-123'].participants).toHaveLength(2);

      // Connection lost
      store.dispatch(disconnected());
      state = store.getState().websocket;
      expect(state.reconnecting).toBe(true);
      expect(state.rooms['game-123'].connected).toBe(false);

      // Error during reconnection
      store.dispatch(error({ error: 'Failed to reconnect' }));
      state = store.getState().websocket;
      expect(state.error).toBe('Failed to reconnect');

      // Successful reconnection
      store.dispatch(connected({ roomId: 'game-123' }));
      state = store.getState().websocket;
      expect(state.connected).toBe(true);
      expect(state.error).toBeNull();
      expect(state.reconnecting).toBe(false);
    });

    it('should handle multiple room scenario', () => {
      // User in multiple rooms (e.g., game session + chat)
      store.dispatch(connected({ roomId: 'game-session' }));
      store.dispatch(connected({ roomId: 'global-chat' }));
      store.dispatch(connected({ roomId: 'party-chat' }));

      // Update participants in each room
      store.dispatch(updateRoomParticipants({ 
        roomId: 'game-session', 
        participants: ['player-1', 'player-2', 'dm'] 
      }));
      store.dispatch(updateRoomParticipants({ 
        roomId: 'global-chat', 
        participants: Array.from({ length: 50 }, (_, i) => `user-${i}`) 
      }));
      store.dispatch(updateRoomParticipants({ 
        roomId: 'party-chat', 
        participants: ['player-1', 'player-2'] 
      }));

      // Leave party chat
      store.dispatch(leaveRoom('party-chat'));

      // Disconnect affects remaining rooms
      store.dispatch(disconnected());

      const state = store.getState().websocket;
      expect(Object.keys(state.rooms)).toHaveLength(2);
      expect(state.rooms['game-session'].connected).toBe(false);
      expect(state.rooms['global-chat'].connected).toBe(false);
      expect(state.rooms['party-chat']).toBeUndefined();
    });

    it('should handle rapid participant updates', () => {
      store.dispatch(connected({ roomId: 'active-game' }));

      // Simulate rapid joins/leaves
      const updates = [
        ['player-1'],
        ['player-1', 'player-2'],
        ['player-1', 'player-2', 'player-3'],
        ['player-1', 'player-3'], // player-2 left
        ['player-1', 'player-3', 'player-4', 'player-5'],
        ['player-3', 'player-4', 'player-5'], // player-1 left
      ];

      updates.forEach(participants => {
        store.dispatch(updateRoomParticipants({ 
          roomId: 'active-game', 
          participants 
        }));
      });

      const state = store.getState().websocket;
      expect(state.rooms['active-game'].participants).toEqual(['player-3', 'player-4', 'player-5']);
    });

    it('should handle error recovery flow', () => {
      // Connect to room
      store.dispatch(connected({ roomId: 'session-1' }));
      
      // Multiple errors
      store.dispatch(error({ error: 'Network timeout' }));
      store.dispatch(error({ error: 'Server error 500' }));
      store.dispatch(error({ error: 'Authentication failed' }));
      
      let state = store.getState().websocket;
      expect(state.error).toBe('Authentication failed');
      expect(state.connected).toBe(true); // Still shows as connected
      
      // Disconnect due to errors
      store.dispatch(disconnected());
      
      // Clear error on successful reconnection
      store.dispatch(connected({ roomId: 'session-1' }));
      
      state = store.getState().websocket;
      expect(state.error).toBeNull();
      expect(state.connected).toBe(true);
      expect(state.reconnecting).toBe(false);
    });

    it('should maintain state consistency', () => {
      // Setup complex state
      store.dispatch(connected({ roomId: 'room-a' }));
      store.dispatch(connected({ roomId: 'room-b' }));
      store.dispatch(updateRoomParticipants({ 
        roomId: 'room-a', 
        participants: ['user-1', 'user-2'] 
      }));
      store.dispatch(updateRoomParticipants({ 
        roomId: 'room-b', 
        participants: ['user-3', 'user-4'] 
      }));
      store.dispatch(error({ error: 'Minor warning' }));

      const initialState = store.getState().websocket;
      
      // Verify complete state
      expect(initialState).toEqual({
        connected: true,
        reconnecting: false,
        error: 'Minor warning',
        rooms: {
          'room-a': {
            connected: true,
            participants: ['user-1', 'user-2'],
          },
          'room-b': {
            connected: true,
            participants: ['user-3', 'user-4'],
          },
        },
      });
    });
  });
});