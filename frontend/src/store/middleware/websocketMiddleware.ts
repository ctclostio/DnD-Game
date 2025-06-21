import { Middleware } from '@reduxjs/toolkit';
import { WebSocketMessage } from '../../types/state';

// WebSocket action types
export const WS_CONNECT = 'websocket/connect';
export const WS_DISCONNECT = 'websocket/disconnect';
export const WS_SEND_MESSAGE = 'websocket/sendMessage';
export const WS_RECEIVE_MESSAGE = 'websocket/receiveMessage';

// WebSocket instance storage
let socket: WebSocket | null = null;
let reconnectTimer: NodeJS.Timeout | null = null;
let pingInterval: NodeJS.Timeout | null = null;

const RECONNECT_DELAY = 3000;
const PING_INTERVAL = 30000;

// Helper function to handle message types
const handleMessageType = (message: WebSocketMessage, store: any) => {
  switch (message.type) {
    case 'combat.update':
      store.dispatch({
        type: 'combat/updateFromServer',
        payload: message.data,
      });
      break;
      
    case 'character.update':
      store.dispatch({
        type: 'character/updateFromServer',
        payload: message.data,
      });
      break;
      
    case 'chat.message':
      store.dispatch({
        type: 'chat/addMessage',
        payload: message.data,
      });
      break;
      
    case 'player.joined':
    case 'player.left':
      store.dispatch({
        type: 'gameSession/updatePlayers',
        payload: message.data,
      });
      break;
      
    default:
      // Generic message handling
      store.dispatch({
        type: WS_RECEIVE_MESSAGE,
        payload: message,
      });
  }
};

// Helper function to start ping interval
const startPingInterval = () => {
  pingInterval = setInterval(() => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type: 'ping' }));
    }
  }, PING_INTERVAL);
};

// Helper function to clear intervals
const clearIntervals = () => {
  if (pingInterval) {
    clearInterval(pingInterval);
    pingInterval = null;
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
};

// Helper function to setup reconnect
const setupReconnect = (store: any, url: string, roomId: string, token: string) => {
  if (!reconnectTimer) {
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null;
      store.dispatch({
        type: WS_CONNECT,
        payload: { url, roomId, token },
      });
    }, RECONNECT_DELAY);
  }
};

// Helper function to setup WebSocket handlers
const setupWebSocketHandlers = (ws: WebSocket, store: any, url: string, roomId: string, token: string) => {
  ws.onopen = () => {
    console.log('WebSocket connected');
    store.dispatch({ type: 'websocket/connected', payload: { roomId } });
    startPingInterval();
  };
  
  ws.onmessage = (event) => {
    try {
      const message: WebSocketMessage = JSON.parse(event.data);
      handleMessageType(message, store);
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  };
  
  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
    store.dispatch({
      type: 'websocket/error',
      payload: { error: 'Connection error' },
    });
  };
  
  ws.onclose = () => {
    console.log('WebSocket disconnected');
    store.dispatch({ type: 'websocket/disconnected' });
    clearIntervals();
    setupReconnect(store, url, roomId, token);
  };
};

// Helper function to handle combat action interception
const interceptCombatAction = (action: any, store: any) => {
  if (!action.type.startsWith('combat/') || 
      !socket || 
      socket.readyState !== WebSocket.OPEN ||
      action.type.includes('FromServer')) {
    return;
  }
  
  const state = store.getState();
  if (!state.combat.sessionId) return;
  
  const message: WebSocketMessage = {
    type: 'combat.action',
    roomId: state.combat.sessionId,
    data: {
      action: action.type,
      payload: action.payload,
    },
    timestamp: Date.now(),
  };
  socket.send(JSON.stringify(message));
};

// Helper function to handle connect action
const handleConnect = (payload: any, store: any) => {
  const { url, roomId, token } = payload;
  
  // Close existing connection
  if (socket) {
    socket.close();
  }
  
  // Create new WebSocket connection
  socket = new WebSocket(`${url}?room=${roomId}&token=${token}`);
  setupWebSocketHandlers(socket, store, url, roomId, token);
};

// Helper function to handle disconnect action
const handleDisconnect = () => {
  if (socket) {
    socket.close();
    socket = null;
  }
  clearIntervals();
};

// Helper function to handle send message action
const handleSendMessage = (payload: any) => {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    console.warn('WebSocket not connected');
    return;
  }
  
  const message: WebSocketMessage = {
    type: payload.type,
    roomId: payload.roomId,
    data: payload.data,
    timestamp: Date.now(),
  };
  socket.send(JSON.stringify(message));
};

export const websocketMiddleware: Middleware = (store) => (next) => (action: any) => {
  switch (action.type) {
    case WS_CONNECT:
      handleConnect(action.payload, store);
      break;
      
    case WS_DISCONNECT:
      handleDisconnect();
      break;
      
    case WS_SEND_MESSAGE:
      handleSendMessage(action.payload);
      break;
  }
  
  // Intercept combat actions to send via WebSocket
  interceptCombatAction(action, store);
  
  return next(action);
};

// Action creators
export const wsConnect = (url: string, roomId: string, token: string) => ({
  type: WS_CONNECT,
  payload: { url, roomId, token },
});

export const wsDisconnect = () => ({
  type: WS_DISCONNECT,
});

export const wsSendMessage = (type: string, roomId: string, data: Record<string, unknown>) => ({
  type: WS_SEND_MESSAGE,
  payload: { type, roomId, data },
});