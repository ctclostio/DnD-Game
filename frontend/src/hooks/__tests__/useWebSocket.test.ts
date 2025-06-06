import { renderHook, act, waitFor } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { useWebSocket, useWebSocketChat } from '../useWebSocket';
import websocketSlice from '../../store/slices/websocketSlice';

// Mock WebSocket
let mockWebSocket: any;
let wsInstances: any[] = [];

beforeEach(() => {
  wsInstances = [];
  
  global.WebSocket = jest.fn().mockImplementation((url) => {
    mockWebSocket = {
      url,
      readyState: WebSocket.CONNECTING,
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
      onopen: null,
      onclose: null,
      onerror: null,
      onmessage: null,
    };
    
    // Simulate event listeners
    const eventListeners: { [key: string]: Function[] } = {};
    
    mockWebSocket.addEventListener = jest.fn((event: string, handler: Function) => {
      if (!eventListeners[event]) {
        eventListeners[event] = [];
      }
      eventListeners[event].push(handler);
    });
    
    mockWebSocket.removeEventListener = jest.fn((event: string, handler: Function) => {
      if (eventListeners[event]) {
        eventListeners[event] = eventListeners[event].filter(h => h !== handler);
      }
    });
    
    mockWebSocket.dispatchEvent = jest.fn((event: any) => {
      const handlers = eventListeners[event.type] || [];
      handlers.forEach(handler => handler(event));
      
      // Also call direct handlers
      if (event.type === 'open' && mockWebSocket.onopen) {
        mockWebSocket.onopen(event);
      } else if (event.type === 'close' && mockWebSocket.onclose) {
        mockWebSocket.onclose(event);
      } else if (event.type === 'error' && mockWebSocket.onerror) {
        mockWebSocket.onerror(event);
      } else if (event.type === 'message' && mockWebSocket.onmessage) {
        mockWebSocket.onmessage(event);
      }
    });
    
    wsInstances.push(mockWebSocket);
    return mockWebSocket;
  });
});

afterEach(() => {
  jest.clearAllMocks();
});

// Create a mock store
const createMockStore = () => {
  return configureStore({
    reducer: {
      websocket: websocketSlice,
    },
  });
};

// Wrapper component for Redux Provider
const createWrapper = (store: ReturnType<typeof createMockStore>) => {
  return ({ children }: { children: React.ReactNode }) => (
    <Provider store={store}>{children}</Provider>
  );
};

describe('useWebSocket', () => {
  let store: ReturnType<typeof createMockStore>;

  beforeEach(() => {
    store = createMockStore();
  });

  it('should initialize with disconnected state', () => {
    const { result } = renderHook(() => useWebSocket(), {
      wrapper: createWrapper(store),
    });

    expect(result.current.isConnected).toBe(false);
    expect(result.current.connectionState).toBe('disconnected');
  });

  it('should connect to WebSocket', async () => {
    const { result } = renderHook(() => useWebSocket(), {
      wrapper: createWrapper(store),
    });

    act(() => {
      result.current.connect();
    });

    expect(global.WebSocket).toHaveBeenCalledWith(expect.stringContaining('ws://'));
    expect(result.current.connectionState).toBe('connecting');

    // Simulate connection open
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.dispatchEvent(new Event('open'));
    });

    await waitFor(() => {
      expect(result.current.isConnected).toBe(true);
      expect(result.current.connectionState).toBe('connected');
    });
  });

  it('should handle connection errors', async () => {
    const { result } = renderHook(() => useWebSocket(), {
      wrapper: createWrapper(store),
    });

    act(() => {
      result.current.connect();
    });

    // Simulate connection error
    act(() => {
      mockWebSocket.dispatchEvent(new Event('error'));
    });

    await waitFor(() => {
      expect(result.current.connectionState).toBe('error');
    });
  });

  it('should send messages when connected', async () => {
    const { result } = renderHook(() => useWebSocket(), {
      wrapper: createWrapper(store),
    });

    act(() => {
      result.current.connect();
    });

    // Simulate connection open
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.dispatchEvent(new Event('open'));
    });

    const message = { type: 'test', data: 'hello' };
    
    act(() => {
      result.current.send(message);
    });

    expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify(message));
  });

  it('should not send messages when disconnected', () => {
    const { result } = renderHook(() => useWebSocket(), {
      wrapper: createWrapper(store),
    });

    const message = { type: 'test', data: 'hello' };
    
    act(() => {
      result.current.send(message);
    });

    expect(mockWebSocket).toBeUndefined();
  });

  it('should subscribe to and receive messages', async () => {
    const { result } = renderHook(() => useWebSocket(), {
      wrapper: createWrapper(store),
    });

    const messageHandler = jest.fn();
    
    act(() => {
      result.current.connect();
      result.current.subscribe('test-message', messageHandler);
    });

    // Simulate connection open
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.dispatchEvent(new Event('open'));
    });

    // Simulate receiving a message
    const messageData = { type: 'test-message', payload: { value: 42 } };
    const messageEvent = new MessageEvent('message', {
      data: JSON.stringify(messageData),
    });

    act(() => {
      mockWebSocket.dispatchEvent(messageEvent);
    });

    await waitFor(() => {
      expect(messageHandler).toHaveBeenCalledWith(messageData);
    });
  });

  it('should unsubscribe from messages', async () => {
    const { result } = renderHook(() => useWebSocket(), {
      wrapper: createWrapper(store),
    });

    const messageHandler = jest.fn();
    
    act(() => {
      result.current.connect();
      const unsubscribe = result.current.subscribe('test-message', messageHandler);
      unsubscribe();
    });

    // Simulate connection and message
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.dispatchEvent(new Event('open'));
      
      const messageEvent = new MessageEvent('message', {
        data: JSON.stringify({ type: 'test-message', payload: {} }),
      });
      mockWebSocket.dispatchEvent(messageEvent);
    });

    expect(messageHandler).not.toHaveBeenCalled();
  });

  it('should disconnect and cleanup', async () => {
    const { result } = renderHook(() => useWebSocket(), {
      wrapper: createWrapper(store),
    });

    act(() => {
      result.current.connect();
    });

    // Simulate connection open
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.dispatchEvent(new Event('open'));
    });

    act(() => {
      result.current.disconnect();
    });

    expect(mockWebSocket.close).toHaveBeenCalled();
    
    // Simulate close event
    act(() => {
      mockWebSocket.readyState = WebSocket.CLOSED;
      mockWebSocket.dispatchEvent(new CloseEvent('close'));
    });

    await waitFor(() => {
      expect(result.current.isConnected).toBe(false);
      expect(result.current.connectionState).toBe('disconnected');
    });
  });

  it('should handle auto-reconnect', async () => {
    jest.useFakeTimers();
    
    const { result } = renderHook(() => useWebSocket({ autoReconnect: true }), {
      wrapper: createWrapper(store),
    });

    act(() => {
      result.current.connect();
    });

    // Simulate connection open then unexpected close
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.dispatchEvent(new Event('open'));
    });

    act(() => {
      mockWebSocket.readyState = WebSocket.CLOSED;
      mockWebSocket.dispatchEvent(new CloseEvent('close', { code: 1006 }));
    });

    // Fast forward to trigger reconnect
    act(() => {
      jest.advanceTimersByTime(3000);
    });

    // Should attempt to reconnect
    expect(global.WebSocket).toHaveBeenCalledTimes(2);

    jest.useRealTimers();
  });

  it('should not auto-reconnect when disabled', async () => {
    jest.useFakeTimers();
    
    const { result } = renderHook(() => useWebSocket({ autoReconnect: false }), {
      wrapper: createWrapper(store),
    });

    act(() => {
      result.current.connect();
    });

    // Simulate connection close
    act(() => {
      mockWebSocket.readyState = WebSocket.CLOSED;
      mockWebSocket.dispatchEvent(new CloseEvent('close', { code: 1006 }));
    });

    // Fast forward
    act(() => {
      jest.advanceTimersByTime(5000);
    });

    // Should not attempt to reconnect
    expect(global.WebSocket).toHaveBeenCalledTimes(1);

    jest.useRealTimers();
  });
});

describe('useWebSocketChat', () => {
  let store: ReturnType<typeof createMockStore>;

  beforeEach(() => {
    store = createMockStore();
  });

  it('should connect to room and handle chat messages', async () => {
    const { result } = renderHook(() => useWebSocketChat('room-123'), {
      wrapper: createWrapper(store),
    });

    expect(result.current.messages).toEqual([]);
    expect(result.current.isConnected).toBe(false);

    // Wait for connection
    await waitFor(() => {
      expect(global.WebSocket).toHaveBeenCalled();
    });

    // Simulate connection open
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.dispatchEvent(new Event('open'));
    });

    // Should send join room message
    await waitFor(() => {
      expect(mockWebSocket.send).toHaveBeenCalledWith(
        JSON.stringify({ type: 'join_room', roomId: 'room-123' })
      );
    });

    // Simulate receiving a chat message
    const chatMessage = {
      type: 'chat_message',
      payload: {
        id: 'msg-1',
        userId: 'user-1',
        username: 'Player1',
        message: 'Hello everyone!',
        timestamp: new Date().toISOString(),
      },
    };

    act(() => {
      const messageEvent = new MessageEvent('message', {
        data: JSON.stringify(chatMessage),
      });
      mockWebSocket.dispatchEvent(messageEvent);
    });

    expect(result.current.messages).toHaveLength(1);
    expect(result.current.messages[0]).toMatchObject({
      id: 'msg-1',
      userId: 'user-1',
      username: 'Player1',
      message: 'Hello everyone!',
      type: 'chat',
    });
  });

  it('should send chat messages', async () => {
    const { result } = renderHook(() => useWebSocketChat('room-123'), {
      wrapper: createWrapper(store),
    });

    // Simulate connection
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.dispatchEvent(new Event('open'));
    });

    act(() => {
      result.current.sendMessage('Test message');
    });

    expect(mockWebSocket.send).toHaveBeenCalledWith(
      JSON.stringify({
        type: 'chat_message',
        roomId: 'room-123',
        message: 'Test message',
      })
    );
  });

  it('should handle dice roll messages', async () => {
    const { result } = renderHook(() => useWebSocketChat('room-123'), {
      wrapper: createWrapper(store),
    });

    // Simulate connection
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.dispatchEvent(new Event('open'));
    });

    const diceRollMessage = {
      type: 'dice_roll',
      payload: {
        id: 'roll-1',
        userId: 'user-1',
        username: 'Player1',
        roll: '2d6+3',
        result: { total: 10, rolls: [4, 3], modifier: 3 },
        timestamp: new Date().toISOString(),
      },
    };

    act(() => {
      const messageEvent = new MessageEvent('message', {
        data: JSON.stringify(diceRollMessage),
      });
      mockWebSocket.dispatchEvent(messageEvent);
    });

    expect(result.current.messages).toHaveLength(1);
    expect(result.current.messages[0].type).toBe('dice');
    expect(result.current.messages[0].rollResult).toEqual({
      total: 10,
      rolls: [4, 3],
      modifier: 3,
    });
  });

  it('should handle system messages', async () => {
    const { result } = renderHook(() => useWebSocketChat('room-123'), {
      wrapper: createWrapper(store),
    });

    // Simulate connection
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.dispatchEvent(new Event('open'));
    });

    const systemMessage = {
      type: 'system_message',
      payload: {
        id: 'sys-1',
        message: 'Player2 has joined the game',
        timestamp: new Date().toISOString(),
      },
    };

    act(() => {
      const messageEvent = new MessageEvent('message', {
        data: JSON.stringify(systemMessage),
      });
      mockWebSocket.dispatchEvent(messageEvent);
    });

    expect(result.current.messages).toHaveLength(1);
    expect(result.current.messages[0].type).toBe('system');
    expect(result.current.messages[0].message).toBe('Player2 has joined the game');
  });

  it('should clear messages', async () => {
    const { result } = renderHook(() => useWebSocketChat('room-123'), {
      wrapper: createWrapper(store),
    });

    // Add some messages
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.dispatchEvent(new Event('open'));
      
      const messageEvent = new MessageEvent('message', {
        data: JSON.stringify({
          type: 'chat_message',
          payload: {
            id: 'msg-1',
            userId: 'user-1',
            username: 'Player1',
            message: 'Test',
            timestamp: new Date().toISOString(),
          },
        }),
      });
      mockWebSocket.dispatchEvent(messageEvent);
    });

    expect(result.current.messages).toHaveLength(1);

    act(() => {
      result.current.clearMessages();
    });

    expect(result.current.messages).toEqual([]);
  });

  it('should handle connection errors gracefully', async () => {
    const { result } = renderHook(() => useWebSocketChat('room-123'), {
      wrapper: createWrapper(store),
    });

    // Simulate connection error
    act(() => {
      mockWebSocket.dispatchEvent(new Event('error'));
    });

    // Should still be able to track messages locally
    expect(result.current.messages).toEqual([]);
    expect(result.current.isConnected).toBe(false);
    
    // Sending should fail gracefully
    act(() => {
      result.current.sendMessage('Test');
    });
    
    // No crash, just no send
    expect(mockWebSocket.send).not.toHaveBeenCalled();
  });
});