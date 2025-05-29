import authService from './auth';

export class WebSocketService {
    constructor() {
        this.ws = null;
        this.reconnectInterval = 5000;
        this.messageHandlers = new Map();
        this.roomId = null;
        this.user = null;
    }

    connect(roomId) {
        if (!authService.isAuthenticated()) {
            console.error('User must be authenticated to connect to WebSocket');
            return;
        }

        this.roomId = roomId;
        this.user = authService.getCurrentUser();
        
        // Include token in WebSocket URL for authentication
        const token = localStorage.getItem('access_token');
        const wsUrl = `ws://localhost:8080/ws?room=${roomId}&token=${token}`;
        
        try {
            this.ws = new WebSocket(wsUrl);
            this.setupEventHandlers();
        } catch (error) {
            console.error('WebSocket connection failed:', error);
            this.scheduleReconnect();
        }
    }

    setupEventHandlers() {
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.sendMessage('join', { 
                roomId: this.roomId, 
                userId: this.user.id,
                username: this.user.username,
                role: this.user.role 
            });
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            this.scheduleReconnect();
        };
    }

    handleMessage(message) {
        // Notify all registered handlers
        this.messageHandlers.forEach(handler => {
            handler(message);
        });

        // Handle specific message types if needed
        switch (message.type) {
            case 'chat':
                this.displayChatMessage(message.data);
                break;
            case 'dice_roll':
                this.displayDiceRoll(message.data);
                break;
            case 'user_joined':
                this.displaySystemMessage(`${message.data.username} joined the session`);
                break;
            case 'user_left':
                this.displaySystemMessage(`${message.data.username} left the session`);
                break;
            case 'combat':
                // Combat updates are handled by registered handlers
                break;
            default:
                console.log('Unknown message type:', message.type);
        }
    }

    sendMessage(type, data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            const message = {
                type,
                roomId: this.roomId,
                data
            };
            this.ws.send(JSON.stringify(message));
        } else {
            console.error('WebSocket is not connected');
        }
    }

    sendChatMessage(text) {
        this.sendMessage('chat', { 
            message: text,
            username: this.user.username 
        });
    }

    sendDiceRoll(diceType, purpose) {
        this.sendMessage('dice_roll', {
            diceType,
            purpose,
            playerName: this.user.username
        });
    }

    onMessage(handler) {
        const id = Date.now();
        this.messageHandlers.set(id, handler);
        return () => this.messageHandlers.delete(id);
    }

    displayChatMessage(data) {
        const chatMessages = document.getElementById('chat-messages');
        if (!chatMessages) return;
        
        const messageDiv = document.createElement('div');
        messageDiv.className = 'chat-message';
        messageDiv.innerHTML = `
            <span class="username">${data.username}:</span>
            <span class="message-text">${data.message}</span>
        `;
        chatMessages.appendChild(messageDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    displayDiceRoll(data) {
        const chatMessages = document.getElementById('chat-messages');
        if (!chatMessages) return;
        
        const messageDiv = document.createElement('div');
        messageDiv.className = 'chat-message dice-roll-message';
        messageDiv.innerHTML = `
            <span class="username">${data.playerName}</span> rolled ${data.diceType}: 
            <strong>${data.result.total}</strong> 
            ${data.purpose ? `(${data.purpose})` : ''}
        `;
        chatMessages.appendChild(messageDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    displaySystemMessage(text) {
        const chatMessages = document.getElementById('chat-messages');
        if (!chatMessages) return;
        
        const messageDiv = document.createElement('div');
        messageDiv.className = 'chat-message system-message';
        messageDiv.innerHTML = `<em>${text}</em>`;
        chatMessages.appendChild(messageDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    scheduleReconnect() {
        setTimeout(() => {
            if (this.roomId && authService.isAuthenticated()) {
                console.log('Attempting to reconnect...');
                this.connect(this.roomId);
            }
        }, this.reconnectInterval);
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}

// Create singleton instance
const wsService = new WebSocketService();

// Export convenience functions
export const connectWebSocket = (roomId) => wsService.connect(roomId);
export const disconnectWebSocket = () => wsService.disconnect();
export const sendMessage = (type, data) => wsService.sendMessage(type, data);
export const sendChatMessage = (text) => wsService.sendChatMessage(text);
export const sendDiceRoll = (diceType, purpose) => wsService.sendDiceRoll(diceType, purpose);
export const onMessage = (handler) => wsService.onMessage(handler);

export default wsService;