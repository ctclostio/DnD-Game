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
        const { type, data } = message;
        
        // Handle system messages
        switch (type) {
            case 'chat':
                this.displayChatMessage(data);
                break;
            case 'dice_roll':
                this.displayDiceRoll(data);
                break;
            case 'player_joined':
                this.displaySystemMessage(`${data.playerName} joined the game`);
                break;
            case 'player_left':
                this.displaySystemMessage(`${data.playerName} left the game`);
                break;
        }

        // Call custom handlers
        const handlers = this.messageHandlers.get(type) || [];
        handlers.forEach(handler => handler(data));
    }

    on(messageType, handler) {
        if (!this.messageHandlers.has(messageType)) {
            this.messageHandlers.set(messageType, []);
        }
        this.messageHandlers.get(messageType).push(handler);
    }

    off(messageType, handler) {
        const handlers = this.messageHandlers.get(messageType) || [];
        const index = handlers.indexOf(handler);
        if (index > -1) {
            handlers.splice(index, 1);
        }
    }

    sendMessage(type, data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            const message = {
                type,
                roomId: this.roomId,
                playerId: this.user.id,
                username: this.user.username,
                role: this.user.role,
                data,
            };
            this.ws.send(JSON.stringify(message));
        }
    }

    sendChatMessage(text) {
        this.sendMessage('chat', {
            text,
            timestamp: new Date().toISOString(),
        });
    }

    sendDiceRoll(diceType, result, purpose) {
        this.sendMessage('dice_roll', {
            diceType,
            result,
            purpose,
            timestamp: new Date().toISOString(),
        });
    }

    displayChatMessage(data) {
        const chatMessages = document.getElementById('chat-messages');
        const messageDiv = document.createElement('div');
        messageDiv.className = 'chat-message';
        messageDiv.innerHTML = `
            <span class="username">${data.playerName}:</span> ${data.text}
        `;
        chatMessages.appendChild(messageDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    displayDiceRoll(data) {
        const chatMessages = document.getElementById('chat-messages');
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