import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { WebSocketState } from '../../types/state';

const initialState: WebSocketState = {
  connected: false,
  reconnecting: false,
  error: null,
  rooms: {},
};

const websocketSlice = createSlice({
  name: 'websocket',
  initialState,
  reducers: {
    connected: (state, action: PayloadAction<{ roomId: string }>) => {
      state.connected = true;
      state.reconnecting = false;
      state.error = null;
      state.rooms[action.payload.roomId] = {
        connected: true,
        participants: [],
      };
    },
    
    disconnected: (state) => {
      state.connected = false;
      state.reconnecting = true;
      Object.keys(state.rooms).forEach(roomId => {
        state.rooms[roomId].connected = false;
      });
    },
    
    error: (state, action: PayloadAction<{ error: string }>) => {
      state.error = action.payload.error;
    },
    
    updateRoomParticipants: (state, action: PayloadAction<{ roomId: string; participants: string[] }>) => {
      if (state.rooms[action.payload.roomId]) {
        state.rooms[action.payload.roomId].participants = action.payload.participants;
      }
    },
    
    leaveRoom: (state, action: PayloadAction<string>) => {
      delete state.rooms[action.payload];
    },
  },
});

export const { connected, disconnected, error, updateRoomParticipants, leaveRoom } = websocketSlice.actions;
export default websocketSlice.reducer;