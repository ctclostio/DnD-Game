import { useEffect, useRef, useCallback, useState } from 'react';
import wsService, { 
  connectWebSocket, 
  disconnectWebSocket, 
  onMessage,
  getConnectionState 
} from '../services/websocket';

interface UseWebSocketOptions {
  roomId: string;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Error) => void;
  autoReconnect?: boolean;
}

interface UseWebSocketReturn {
  isConnected: boolean;
  connectionState: 'connected' | 'connecting' | 'disconnected';
  sendMessage: (type: string, data: any) => void;
  disconnect: () => void;
  reconnect: () => void;
}

export function useWebSocket({
  roomId,
  onConnect,
  onDisconnect,
  onError,
  autoReconnect = true,
}: UseWebSocketOptions): UseWebSocketReturn {
  const [connectionState, setConnectionState] = useState<'connected' | 'connecting' | 'disconnected'>('disconnected');
  const cleanupRef = useRef<(() => void) | null>(null);
  const isMountedRef = useRef(true);
  
  // Monitor connection state
  useEffect(() => {
    const checkConnection = setInterval(() => {
      if (isMountedRef.current) {
        const state = getConnectionState();
        setConnectionState(state);
      }
    }, 1000);
    
    return () => clearInterval(checkConnection);
  }, []);

  // Handle WebSocket connection
  useEffect(() => {
    isMountedRef.current = true;
    
    // Connect to WebSocket
    connectWebSocket(roomId);
    
    // Set up message handler for connection events
    const cleanup = onMessage((message) => {
      if (!isMountedRef.current) return;
      
      if (message.type === 'connected' && onConnect) {
        onConnect();
      } else if (message.type === 'error' && onError) {
        onError(new Error(message.data?.message || 'WebSocket error'));
      }
    });
    
    cleanupRef.current = cleanup;
    
    // Cleanup on unmount
    return () => {
      isMountedRef.current = false;
      
      // Clean up message handler
      if (cleanupRef.current) {
        cleanupRef.current();
        cleanupRef.current = null;
      }
      
      // Disconnect WebSocket
      disconnectWebSocket();
      
      if (onDisconnect) {
        onDisconnect();
      }
    };
  }, [roomId]); // Only reconnect if roomId changes

  const sendMessage = useCallback((type: string, data: any) => {
    wsService.sendMessage(type, data);
  }, []);

  const disconnect = useCallback(() => {
    disconnectWebSocket();
  }, []);

  const reconnect = useCallback(() => {
    connectWebSocket(roomId);
  }, [roomId]);

  return {
    isConnected: connectionState === 'connected',
    connectionState,
    sendMessage,
    disconnect,
    reconnect,
  };
}

// Hook for WebSocket chat functionality
export function useWebSocketChat(roomId: string) {
  const [messages, setMessages] = useState<Array<{ id: string; type: string; data: any }>>([]);
  const messagesRef = useRef(messages);
  
  // Keep ref in sync with state
  useEffect(() => {
    messagesRef.current = messages;
  }, [messages]);

  const { isConnected, connectionState, sendMessage } = useWebSocket({
    roomId,
    onConnect: () => {
      console.log('Chat connected');
    },
    onDisconnect: () => {
      console.log('Chat disconnected');
    },
  });

  useEffect(() => {
    // Subscribe to messages
    const cleanup = onMessage((message) => {
      if (['chat', 'dice_roll', 'system'].includes(message.type)) {
        setMessages(prev => [...prev, {
          id: Date.now().toString(),
          type: message.type,
          data: message.data,
        }]);
      }
    });

    return cleanup;
  }, []);

  const sendChatMessage = useCallback((text: string) => {
    if (isConnected) {
      sendMessage('chat', { message: text });
    }
  }, [isConnected, sendMessage]);

  const sendDiceRoll = useCallback((diceType: string, purpose?: string) => {
    if (isConnected) {
      sendMessage('dice_roll', { diceType, purpose });
    }
  }, [isConnected, sendMessage]);

  const clearMessages = useCallback(() => {
    setMessages([]);
  }, []);

  return {
    messages,
    sendChatMessage,
    sendDiceRoll,
    clearMessages,
    isConnected,
    connectionState,
  };
}