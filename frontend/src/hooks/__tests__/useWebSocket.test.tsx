import { renderHook } from '@testing-library/react';
import { useWebSocket, useWebSocketChat } from '../useWebSocket';
import wsService, {
  connectWebSocket,
  disconnectWebSocket,
  onMessage,
  getConnectionState,
} from '../../services/websocket';

// Mock the websocket service
jest.mock('../../services/websocket');

const mockedWsService = wsService as jest.Mocked<typeof wsService>;
const mockedConnect = connectWebSocket as jest.Mock;
const mockedDisconnect = disconnectWebSocket as jest.Mock;
const mockedOnMessage = onMessage as jest.Mock;
const mockedGetConnectionState = getConnectionState as jest.Mock;

describe('useWebSocket', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  it('should connect on mount and disconnect on unmount', () => {
    const onConnect = jest.fn();
    const onDisconnect = jest.fn();
    const { unmount } = renderHook(() =>
      useWebSocket({
        roomId: 'test-room',
        onConnect,
        onDisconnect,
      })
    );

    expect(mockedConnect).toHaveBeenCalledWith('test-room');

    unmount();
    expect(mockedDisconnect).toHaveBeenCalled();
    expect(onDisconnect).toHaveBeenCalled();
  });

  it('should call onConnect when a "connected" message is received', () => {
    const onConnect = jest.fn();
    mockedOnMessage.mockImplementation((callback) => {
      callback({ type: 'connected' });
      return jest.fn(); // return cleanup function
    });

    renderHook(() => useWebSocket({ roomId: 'test-room', onConnect }));
    expect(onConnect).toHaveBeenCalled();
  });

  it('should call onError when an "error" message is received', () => {
    const onError = jest.fn();
    const errorMessage = 'Test Error';
    mockedOnMessage.mockImplementation((callback) => {
      callback({ type: 'error', data: { message: errorMessage } });
      return jest.fn();
    });

    renderHook(() => useWebSocket({ roomId: 'test-room', onError }));
    expect(onError).toHaveBeenCalledWith(new Error(errorMessage));
  });

  it('should send a message', () => {
    const { result } = renderHook(() => useWebSocket({ roomId: 'test-room' }));
    const message = { type: 'test', data: { value: 123 } };
    result.current.sendMessage(message.type, message.data);
    expect(mockedWsService.sendMessage).toHaveBeenCalledWith(message.type, message.data);
  });

  it('should return the current connection state', async () => {
    mockedGetConnectionState.mockReturnValue('connected');
    const { result } = renderHook(() => useWebSocket({ roomId: 'test-room' }));
    
    // Initially disconnected
    expect(result.current.isConnected).toBe(false);
    expect(result.current.connectionState).toBe('disconnected');
    
    // Wait for the interval to update the state
    await new Promise(resolve => setTimeout(resolve, 1100));
    
    expect(result.current.isConnected).toBe(true);
    expect(result.current.connectionState).toBe('connected');
  });
});

describe('useWebSocketChat', () => {
  it('should send a chat message', async () => {
    mockedGetConnectionState.mockReturnValue('connected');
    const { result } = renderHook(() => useWebSocketChat('test-room'));
    
    // Wait for connection state to update
    await new Promise(resolve => setTimeout(resolve, 1100));
    
    const message = 'Hello, world!';
    result.current.sendChatMessage(message);
    expect(mockedWsService.sendMessage).toHaveBeenCalledWith('chat', { message });
  });

  it('should add a received chat message to the state', () => {
    // Mock Date.now to return a predictable value
    const mockDateNow = jest.spyOn(Date, 'now').mockReturnValue(1);
    
    const chatMessage = {
      type: 'chat',
      data: { username: 'test', message: 'hello' },
    };
    mockedOnMessage.mockImplementation((callback) => {
      callback(chatMessage);
      return jest.fn();
    });
    const { result } = renderHook(() => useWebSocketChat('test-room'));
    expect(result.current.messages).toEqual([{
      id: '1',
      type: 'chat',
      data: { username: 'test', message: 'hello' },
    }]);
    
    // Restore Date.now
    mockDateNow.mockRestore();
  });
});