import { Middleware } from '@reduxjs/toolkit';
import { WebSocketMessage } from '../../types/state';

// WebSocket action types
export const WS_CONNECT = 'websocket/connect';
export const WS_DISCONNECT = 'websocket/disconnect';
export const WS_SEND_MESSAGE = 'websocket/sendMessage';
export const WS_RECEIVE_MESSAGE = 'websocket/receiveMessage';

// WebSocket instance storage
let socket: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let pingInterval: ReturnType<typeof setInterval> | null = null;

const RECONNECT_DELAY = 3000;
const PING_INTERVAL = 30000;

export const websocketMiddleware: Middleware = (store) => {
  return (next) => (action: any) => {
    switch (action.type) {
      case WS_CONNECT:
        const { url, roomId, token } = action.payload;
        
        // Close existing connection
        if (socket) {
          socket.close();
        }
        
        // Create new WebSocket connection
        socket = new WebSocket(`${url}?room=${roomId}&token=${token}`);
        
        socket.onopen = () => {
          console.debug('WebSocket connected');
          store.dispatch({ type: 'websocket/connected', payload: { roomId } });
          
          // Start ping interval
          pingInterval = setInterval(() => {
            if (socket && socket.readyState === WebSocket.OPEN) {
              socket.send(JSON.stringify({ type: 'ping' }));
            }
          }, PING_INTERVAL);
        };
        
        socket.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data);
            
            // Handle different message types
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
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };
        
        socket.onerror = (error) => {
          console.error('WebSocket error:', error);
          store.dispatch({
            type: 'websocket/error',
            payload: { error: 'Connection error' },
          });
        };
        
        socket.onclose = () => {
          console.debug('WebSocket disconnected');
          store.dispatch({ type: 'websocket/disconnected' });
          
          // Clear ping interval
          if (pingInterval) {
            clearInterval(pingInterval);
            pingInterval = null;
          }
          
          // Attempt to reconnect
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
        break;
        
      case WS_DISCONNECT:
        if (socket) {
          socket.close();
          socket = null;
        }
        if (reconnectTimer) {
          clearTimeout(reconnectTimer);
          reconnectTimer = null;
        }
        if (pingInterval) {
          clearInterval(pingInterval);
          pingInterval = null;
        }
        break;
        
      case WS_SEND_MESSAGE:
        if (socket && socket.readyState === WebSocket.OPEN) {
          const message: WebSocketMessage = {
            type: action.payload.type,
            roomId: action.payload.roomId,
            data: action.payload.data,
            timestamp: Date.now(),
          };
          socket.send(JSON.stringify(message));
        } else {
          console.warn('WebSocket not connected');
        }
        break;
    }
    
    // Intercept combat actions to send via WebSocket
    if (action.type.startsWith('combat/') && 
        socket && 
        socket.readyState === WebSocket.OPEN &&
        !action.type.includes('FromServer')) {
      
      const state = store.getState();
      if (state.combat.sessionId) {
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
      }
    }
    
    return next(action);
  };
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
