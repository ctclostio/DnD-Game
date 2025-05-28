export class ApiService {
    constructor() {
        this.baseURL = '/api/v1';
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
            ...options,
        };

        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }

    // Character endpoints
    async getCharacters() {
        return this.request('/characters');
    }

    async getCharacter(id) {
        return this.request(`/characters/${id}`);
    }

    async createCharacter(characterData) {
        return this.request('/characters', {
            method: 'POST',
            body: JSON.stringify(characterData),
        });
    }

    async updateCharacter(id, characterData) {
        return this.request(`/characters/${id}`, {
            method: 'PUT',
            body: JSON.stringify(characterData),
        });
    }

    // Dice endpoints
    async rollDice(diceType, purpose = '') {
        return this.request('/dice/roll', {
            method: 'POST',
            body: JSON.stringify({ diceType, purpose }),
        });
    }

    // Game session endpoints
    async createGameSession(sessionData) {
        return this.request('/game/session', {
            method: 'POST',
            body: JSON.stringify(sessionData),
        });
    }

    async getGameSession(id) {
        return this.request(`/game/session/${id}`);
    }

    // Health check
    async healthCheck() {
        return this.request('/health');
    }
}