import authService, { User } from './auth';
import { createElement } from '../utils/dom';

interface WebSocketMessage {
  type: string;
  data: Record<string, unknown>;
}

interface ChatMessage {
  username: string;
  message: string;
}

interface DiceRollMessage {
  playerName: string;
  diceType: string;
  purpose?: string;
  result: {
    total: number;
    rolls?: number[];
  };
}

type MessageHandler = (message: WebSocketMessage) => void;
type CleanupFunction = () => void;

export class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectInterval: number = 5000;
  private messageHandlers: Map<number, MessageHandler> = new Map();
  private roomId: string | null = null;
  private user: User | null = null;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private isIntentionalDisconnect: boolean = false;
  private maxReconnectAttempts: number = 10;
  private reconnectAttempts: number = 0;
  private cleanupFunctions: Set<CleanupFunction> = new Set();

  connect(roomId: string): void {
    // Clean up any existing connection
    this.cleanup();
    
    if (!authService.isAuthenticated()) {
      console.error('User must be authenticated to connect to WebSocket');
      return;
    }

    this.roomId = roomId;
    this.user = authService.getCurrentUser();
    this.isIntentionalDisconnect = false;
    this.reconnectAttempts = 0;
    
    // Connect without token in URL
    let wsUrl: string;
    
    if (process.env.NODE_ENV === 'development') {
      // In development, use hardcoded localhost
      wsUrl = `ws://localhost:8080/ws?room=${roomId}`;
    } else {
      // In production, use relative URL
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host; // includes port if present
      wsUrl = `${protocol}//${host}/ws?room=${roomId}`;
    }
    
    try {
      this.ws = new WebSocket(wsUrl);
      this.setupEventHandlers();
    } catch (error) {
      console.error('WebSocket connection failed:', error);
      this.scheduleReconnect();
    }
  }

  private setupEventHandlers(): void {
    if (!this.ws) return;

    let isAuthenticated = false;

    const handleOpen = () => {
      console.log('WebSocket connected, waiting for authentication...');
      this.reconnectAttempts = 0; // Reset on successful connection
    };

    const handleMessage = (event: MessageEvent) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        
        // Handle authentication flow
        if (message.type === 'auth_required') {
          // Send authentication token
          const token = localStorage.getItem('access_token');
          if (token) {
            this.ws?.send(JSON.stringify({
              type: 'auth',
              token: token,
              room: this.roomId
            }));
          } else {
            console.error('No access token available');
            this.disconnect();
          }
          return;
        }
        
        if (message.type === 'auth_success') {
          console.log('WebSocket authenticated successfully');
          isAuthenticated = true;
          // Now send the join message
          this.sendMessage('join', { 
            roomId: this.roomId, 
            userId: this.user?.id,
            username: this.user?.username,
            role: this.user?.role 
          });
          return;
        }
        
        if (message.type === 'error') {
          console.error('WebSocket error:', message.data);
          if (message.data?.error === 'Invalid token') {
            // Token is invalid, try to refresh
            this.handleTokenRefresh().catch(error => console.error("Token refresh failed:", error));
          }
          return;
        }
        
        // Only handle other messages if authenticated
        if (isAuthenticated) {
          this.handleMessage(message);
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    const handleError = (error: Event) => {
      console.error('WebSocket error:', error);
    };

    const handleClose = () => {
      console.log('WebSocket disconnected');
      isAuthenticated = false;
      
      // Only attempt reconnect if it wasn't intentional and under the limit
      if (!this.isIntentionalDisconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.scheduleReconnect();
      }
    };

    // Add event listeners
    this.ws.addEventListener('open', handleOpen);
    this.ws.addEventListener('message', handleMessage);
    this.ws.addEventListener('error', handleError);
    this.ws.addEventListener('close', handleClose);

    // Store cleanup function
    this.cleanupFunctions.add(() => {
      if (this.ws) {
        this.ws.removeEventListener('open', handleOpen);
        this.ws.removeEventListener('message', handleMessage);
        this.ws.removeEventListener('error', handleError);
        this.ws.removeEventListener('close', handleClose);
      }
    });
  }

  private async handleTokenRefresh(): Promise<void> {
    try {
      const newToken = await authService.refreshAccessToken();
      if (newToken) {
        console.log('Token refreshed, reconnecting...');
      } else {
        console.error('Token refresh failed');
        this.disconnect();
      }
    } catch (error) {
      console.error('Error refreshing token:', error);
      this.disconnect();
    }
  }

  private handleMessage(message: WebSocketMessage): void {
    // Notify all registered handlers
    this.messageHandlers.forEach(handler => {
      try {
        handler(message);
      } catch (error) {
        console.error('Message handler error:', error);
      }
    });

    // Handle specific message types
    switch (message.type) {
      case 'chat':
        if (this.isChatMessage(message.data)) {
          this.displayChatMessage(message.data);
        }
        break;
      case 'dice_roll':
        if (this.isDiceRollMessage(message.data)) {
          this.displayDiceRoll(message.data);
        }
        break;
      case 'player_joined':
        this.displaySystemMessage(`${message.data.username} joined the game`);
        break;
      case 'player_left':
        this.displaySystemMessage(`${message.data.username} left the game`);
        break;
    }
  }

  sendMessage(type: string, data: Record<string, unknown>): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }));
    } else {
      console.error('WebSocket is not connected');
    }
  }

  sendChatMessage(text: string): void {
    this.sendMessage('chat', { 
      message: text,
      username: this.user?.username 
    });
  }

  sendDiceRoll(diceType: string, purpose?: string): void {
    this.sendMessage('dice_roll', {
      diceType,
      purpose,
      playerName: this.user?.username
    });
  }

  onMessage(handler: MessageHandler): CleanupFunction {
    const id = Date.now() + Math.random();
    this.messageHandlers.set(id, handler);
    
    // Return cleanup function
    const cleanup = () => {
      this.messageHandlers.delete(id);
    };
    
    this.cleanupFunctions.add(cleanup);
    return cleanup;
  }

  private displayChatMessage(data: ChatMessage): void {
    const chatMessages = document.getElementById('chat-messages');
    if (!chatMessages) return;
    
    const messageDiv = createElement('div', { className: 'chat-message' });
    
    const usernameSpan = createElement('span', {
      className: 'username',
      textContent: data.username + ':'
    });
    
    const messageSpan = createElement('span', {
      className: 'message-text',
      textContent: data.message
    });
    
    messageDiv.appendChild(usernameSpan);
    messageDiv.appendChild(document.createTextNode(' '));
    messageDiv.appendChild(messageSpan);
    
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
  }

  private displayDiceRoll(data: DiceRollMessage): void {
    const chatMessages = document.getElementById('chat-messages');
    if (!chatMessages) return;
    
    const messageDiv = createElement('div', { 
      className: 'chat-message dice-roll-message' 
    });
    
    const usernameSpan = createElement('span', {
      className: 'username',
      textContent: data.playerName
    });
    
    const rollText = document.createTextNode(` rolled ${data.diceType}: `);
    
    const resultSpan = createElement('strong', {
      textContent: String(data.result.total)
    });
    
    messageDiv.appendChild(usernameSpan);
    messageDiv.appendChild(rollText);
    messageDiv.appendChild(resultSpan);
    
    if (data.purpose) {
      const purposeText = document.createTextNode(` (${data.purpose})`);
      messageDiv.appendChild(purposeText);
    }
    
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
  }

  private displaySystemMessage(text: string): void {
    const chatMessages = document.getElementById('chat-messages');
    if (!chatMessages) return;
    
    const messageDiv = createElement('div', { 
      className: 'chat-message system-message' 
    });
    
    const emText = createElement('em', {
      textContent: text
    });
    
    messageDiv.appendChild(emText);
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
  }

  private scheduleReconnect(): void {
    // Clear any existing timer
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    
    this.reconnectAttempts++;
    
    // Exponential backoff with max delay
    const delay = Math.min(
      this.reconnectInterval * Math.pow(2, this.reconnectAttempts - 1),
      30000 // Max 30 seconds
    );
    
    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts} in ${delay}ms`);
    
    this.reconnectTimer = setTimeout(() => {
      if (this.roomId && authService.isAuthenticated() && !this.isIntentionalDisconnect) {
        console.log('Attempting to reconnect...');
        this.connect(this.roomId);
      }
    }, delay);
    
    // Store cleanup for the timer
    this.cleanupFunctions.add(() => {
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer);
        this.reconnectTimer = null;
      }
    });
  }

  disconnect(): void {
    this.isIntentionalDisconnect = true;
    this.cleanup();
  }

  private cleanup(): void {
    // Execute all cleanup functions
    this.cleanupFunctions.forEach(cleanup => cleanup());
    this.cleanupFunctions.clear();
    
    // Clear reconnect timer
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    
    // Close WebSocket
    if (this.ws) {
      if (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING) {
        this.ws.close();
      }
      this.ws = null;
    }
    
    // Clear handlers
    this.messageHandlers.clear();
    
    // Reset state
    this.roomId = null;
    this.user = null;
  }

  // Get connection state
  getConnectionState(): 'connected' | 'connecting' | 'disconnected' {
    if (!this.ws) return 'disconnected';
    
    switch (this.ws.readyState) {
      case WebSocket.CONNECTING:
        return 'connecting';
      case WebSocket.OPEN:
        return 'connected';
      default:
        return 'disconnected';
    }
  }

  // Check if connected
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
  private isChatMessage(data: unknown): data is ChatMessage {
    return (
      typeof data === 'object' &&
      data !== null &&
      'username' in data &&
      'message' in data
    );
  }

  private isDiceRollMessage(data: unknown): data is DiceRollMessage {
    return (
      typeof data === 'object' &&
      data !== null &&
      'playerName' in data &&
      'diceType' in data &&
      'result' in data
    );
  }
}

// Create singleton instance
const wsService = new WebSocketService();

// Export convenience functions
export const connectWebSocket = (roomId: string): void => wsService.connect(roomId);
export const disconnectWebSocket = (): void => wsService.disconnect();
export const sendMessage = (type: string, data: Record<string, unknown>): void => wsService.sendMessage(type, data);
export const sendChatMessage = (text: string): void => wsService.sendChatMessage(text);
export const sendDiceRoll = (diceType: string, purpose?: string): void => wsService.sendDiceRoll(diceType, purpose);
export const onMessage = (handler: MessageHandler): CleanupFunction => wsService.onMessage(handler);
export const getConnectionState = (): 'connected' | 'connecting' | 'disconnected' => wsService.getConnectionState();
export const isConnected = (): boolean => wsService.isConnected();

export default wsService;