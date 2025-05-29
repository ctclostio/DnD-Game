import authService from './auth';

export class ApiService {
    constructor() {
        this.baseURL = '/api/v1';
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        
        // Use authenticated request if user is logged in
        if (authService.isAuthenticated()) {
            const response = await authService.makeAuthenticatedRequest(url, {
                ...options,
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers,
                },
            });

            if (!response.ok) {
                const error = await response.json().catch(() => ({ error: 'Request failed' }));
                throw new Error(error.error || `HTTP error! status: ${response.status}`);
            }

            return await response.json();
        }

        // Non-authenticated request
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
                const error = await response.json().catch(() => ({ error: 'Request failed' }));
                throw new Error(error.error || `HTTP error! status: ${response.status}`);
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
        return this.request(`/game/sessions/${id}`);
    }

    // Health check
    async healthCheck() {
        return this.request('/health');
    }

    // Combat endpoints
    async startCombat(gameSessionId, combatants) {
        return this.request('/combat/start', {
            method: 'POST',
            body: JSON.stringify({ gameSessionId, combatants })
        });
    }

    async getCombat(combatId) {
        return this.request(`/combat/${combatId}`);
    }

    async getCombatBySession(sessionId) {
        return this.request(`/combat/session/${sessionId}`);
    }

    async nextTurn(combatId) {
        return this.request(`/combat/${combatId}/next-turn`, {
            method: 'POST'
        });
    }

    async processCombatAction(combatId, action) {
        return this.request(`/combat/${combatId}/action`, {
            method: 'POST',
            body: JSON.stringify(action)
        });
    }

    async endCombat(combatId) {
        return this.request(`/combat/${combatId}/end`, {
            method: 'POST'
        });
    }

    // Inventory endpoints
    async getCharacterInventory(characterId) {
        return this.request(`/characters/${characterId}/inventory`);
    }

    async addItemToInventory(characterId, itemId, quantity = 1) {
        return this.request(`/characters/${characterId}/inventory`, {
            method: 'POST',
            body: JSON.stringify({ item_id: itemId, quantity })
        });
    }

    async removeItemFromInventory(characterId, itemId, quantity = 1) {
        return this.request(`/characters/${characterId}/inventory/remove`, {
            method: 'POST',
            body: JSON.stringify({ item_id: itemId, quantity })
        });
    }

    async equipItem(characterId, itemId) {
        return this.request(`/characters/${characterId}/inventory/${itemId}/equip`, {
            method: 'POST'
        });
    }

    async unequipItem(characterId, itemId) {
        return this.request(`/characters/${characterId}/inventory/${itemId}/unequip`, {
            method: 'POST'
        });
    }

    async attuneItem(characterId, itemId) {
        return this.request(`/characters/${characterId}/inventory/${itemId}/attune`, {
            method: 'POST'
        });
    }

    async unattuneItem(characterId, itemId) {
        return this.request(`/characters/${characterId}/inventory/${itemId}/unattune`, {
            method: 'POST'
        });
    }

    async getCharacterCurrency(characterId) {
        return this.request(`/characters/${characterId}/currency`);
    }

    async updateCharacterCurrency(characterId, currencyChange) {
        return this.request(`/characters/${characterId}/currency`, {
            method: 'PUT',
            body: JSON.stringify(currencyChange)
        });
    }

    async purchaseItem(characterId, itemId, quantity = 1) {
        return this.request(`/characters/${characterId}/inventory/purchase`, {
            method: 'POST',
            body: JSON.stringify({ item_id: itemId, quantity })
        });
    }

    async sellItem(characterId, itemId, quantity = 1) {
        return this.request(`/characters/${characterId}/inventory/sell`, {
            method: 'POST',
            body: JSON.stringify({ item_id: itemId, quantity })
        });
    }

    async getCharacterWeight(characterId) {
        return this.request(`/characters/${characterId}/weight`);
    }

    async createItem(item) {
        return this.request('/items', {
            method: 'POST',
            body: JSON.stringify(item)
        });
    }

    async getItemsByType(type) {
        return this.request(`/items?type=${type}`);
    }

    // NPC endpoints
    async getNPCsBySession(sessionId) {
        return this.request(`/npcs/session/${sessionId}`);
    }

    async createNPC(npcData) {
        return this.request('/npcs', {
            method: 'POST',
            body: JSON.stringify(npcData)
        });
    }

    async updateNPC(npcId, npcData) {
        return this.request(`/npcs/${npcId}`, {
            method: 'PUT',
            body: JSON.stringify(npcData)
        });
    }

    async deleteNPC(npcId) {
        return this.request(`/npcs/${npcId}`, {
            method: 'DELETE'
        });
    }
}

// Create and export singleton instance
const apiService = new ApiService();

// Export individual methods for convenience
export const startCombat = (gameSessionId, combatants) => apiService.startCombat(gameSessionId, combatants);
export const getCombat = (combatId) => apiService.getCombat(combatId);
export const getCombatBySession = (sessionId) => apiService.getCombatBySession(sessionId);
export const nextTurn = (combatId) => apiService.nextTurn(combatId);
export const processCombatAction = (combatId, action) => apiService.processCombatAction(combatId, action);
export const endCombat = (combatId) => apiService.endCombat(combatId);
export const getCharacterInventory = (characterId) => apiService.getCharacterInventory(characterId);
export const addItemToInventory = (characterId, itemId, quantity) => apiService.addItemToInventory(characterId, itemId, quantity);
export const removeItemFromInventory = (characterId, itemId, quantity) => apiService.removeItemFromInventory(characterId, itemId, quantity);
export const equipItem = (characterId, itemId) => apiService.equipItem(characterId, itemId);
export const unequipItem = (characterId, itemId) => apiService.unequipItem(characterId, itemId);
export const attuneItem = (characterId, itemId) => apiService.attuneItem(characterId, itemId);
export const unattuneItem = (characterId, itemId) => apiService.unattuneItem(characterId, itemId);
export const getCharacterCurrency = (characterId) => apiService.getCharacterCurrency(characterId);
export const updateCharacterCurrency = (characterId, currencyChange) => apiService.updateCharacterCurrency(characterId, currencyChange);
export const purchaseItem = (characterId, itemId, quantity) => apiService.purchaseItem(characterId, itemId, quantity);
export const sellItem = (characterId, itemId, quantity) => apiService.sellItem(characterId, itemId, quantity);
export const getCharacterWeight = (characterId) => apiService.getCharacterWeight(characterId);
export const createItem = (item) => apiService.createItem(item);
export const getItemsByType = (type) => apiService.getItemsByType(type);
export const getNPCsBySession = (sessionId) => apiService.getNPCsBySession(sessionId);
export const createNPC = (npcData) => apiService.createNPC(npcData);
export const updateNPC = (npcId, npcData) => apiService.updateNPC(npcId, npcData);
export const deleteNPC = (npcId) => apiService.deleteNPC(npcId);
export const getCharacters = () => apiService.getCharacters();
export const getGameSession = (id) => apiService.getGameSession(id);

export default apiService;