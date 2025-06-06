import authService from './auth';
import { fetchWithCSRF } from '../utils/csrf';

interface RequestOptions extends RequestInit {
  headers?: HeadersInit;
}

interface ApiResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
}

export class ApiService {
  private baseURL: string;

  constructor() {
    this.baseURL = '/api/v1';
  }

  async request<T = any>(endpoint: string, options: RequestOptions = {}): Promise<T> {
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
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetchWithCSRF(url, config);
      
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
    return this.request<any[]>('/characters');
  }

  async getCharacter(id: string) {
    return this.request(`/characters/${id}`);
  }

  async createCharacter(characterData: any) {
    return this.request('/characters', {
      method: 'POST',
      body: JSON.stringify(characterData),
    });
  }

  async updateCharacter(id: string, characterData: any) {
    return this.request(`/characters/${id}`, {
      method: 'PUT',
      body: JSON.stringify(characterData),
    });
  }

  async deleteCharacter(id: string) {
    return this.request(`/characters/${id}`, {
      method: 'DELETE',
    });
  }

  // Character creation endpoints
  async getCharacterOptions() {
    return this.request('/characters/options');
  }

  async validateCharacter(characterData: any) {
    return this.request('/characters/validate', {
      method: 'POST',
      body: JSON.stringify(characterData),
    });
  }

  async rollAbilityScores() {
    return this.request('/characters/roll-abilities', {
      method: 'POST',
    });
  }

  // Spell endpoints
  async castSpell(characterId: string, spellData: any) {
    return this.request(`/characters/${characterId}/cast-spell`, {
      method: 'POST',
      body: JSON.stringify(spellData),
    });
  }

  async rest(characterId: string, restType: 'short' | 'long') {
    return this.request(`/characters/${characterId}/rest`, {
      method: 'POST',
      body: JSON.stringify({ type: restType }),
    });
  }

  // Experience endpoints
  async addExperience(characterId: string, amount: number) {
    return this.request(`/characters/${characterId}/add-experience`, {
      method: 'POST',
      body: JSON.stringify({ amount }),
    });
  }

  // Dice endpoints
  async rollDice(notation: string, purpose?: string) {
    return this.request('/dice/roll', {
      method: 'POST',
      body: JSON.stringify({ notation, purpose }),
    });
  }

  // Game session endpoints
  async createGameSession(sessionData: any) {
    return this.request('/game/sessions', {
      method: 'POST',
      body: JSON.stringify(sessionData),
    });
  }

  async getGameSession(id: string) {
    return this.request(`/game/sessions/${id}`);
  }

  async updateGameSession(id: string, sessionData: any) {
    return this.request(`/game/sessions/${id}`, {
      method: 'PUT',
      body: JSON.stringify(sessionData),
    });
  }

  async joinGameSession(id: string) {
    return this.request(`/game/sessions/${id}/join`, {
      method: 'POST',
    });
  }

  async leaveGameSession(id: string) {
    return this.request(`/game/sessions/${id}/leave`, {
      method: 'POST',
    });
  }

  // Combat endpoints
  async startCombat(combatData: any) {
    return this.request('/combat/start', {
      method: 'POST',
      body: JSON.stringify(combatData),
    });
  }

  async getCombat(id: string) {
    return this.request(`/combat/${id}`);
  }

  async getCombatBySession(sessionId: string) {
    return this.request(`/combat/session/${sessionId}`);
  }

  async nextTurn(combatId: string) {
    return this.request(`/combat/${combatId}/next-turn`, {
      method: 'POST',
    });
  }

  async processCombatAction(combatId: string, action: any) {
    return this.request(`/combat/${combatId}/action`, {
      method: 'POST',
      body: JSON.stringify(action),
    });
  }

  async endCombat(combatId: string) {
    return this.request(`/combat/${combatId}/end`, {
      method: 'POST',
    });
  }

  // Skill check endpoints
  async performSkillCheck(checkData: any) {
    return this.request('/skill-check', {
      method: 'POST',
      body: JSON.stringify(checkData),
    });
  }

  async getCharacterChecks(characterId: string) {
    return this.request(`/characters/${characterId}/checks`);
  }

  // NPC endpoints
  async createNPC(npcData: any) {
    return this.request('/npcs', {
      method: 'POST',
      body: JSON.stringify(npcData),
    });
  }

  async getNPC(id: string) {
    return this.request(`/npcs/${id}`);
  }

  async updateNPC(id: string, npcData: any) {
    return this.request(`/npcs/${id}`, {
      method: 'PUT',
      body: JSON.stringify(npcData),
    });
  }

  async deleteNPC(id: string) {
    return this.request(`/npcs/${id}`, {
      method: 'DELETE',
    });
  }

  async getNPCsBySession(sessionId: string) {
    return this.request(`/npcs/session/${sessionId}`);
  }

  async searchNPCs(query: string) {
    return this.request(`/npcs/search?q=${encodeURIComponent(query)}`);
  }

  async getNPCTemplates() {
    return this.request('/npcs/templates');
  }

  // Inventory endpoints
  async getCharacterInventory(characterId: string) {
    return this.request(`/characters/${characterId}/inventory`);
  }

  async addItemToInventory(characterId: string, item: any) {
    return this.request(`/characters/${characterId}/inventory`, {
      method: 'POST',
      body: JSON.stringify(item),
    });
  }

  async removeItemFromInventory(characterId: string, itemId: string, quantity: number) {
    return this.request(`/characters/${characterId}/inventory/remove`, {
      method: 'POST',
      body: JSON.stringify({ itemId, quantity }),
    });
  }

  async equipItem(characterId: string, itemId: string) {
    return this.request(`/characters/${characterId}/inventory/${itemId}/equip`, {
      method: 'POST',
    });
  }

  async unequipItem(characterId: string, itemId: string) {
    return this.request(`/characters/${characterId}/inventory/${itemId}/unequip`, {
      method: 'POST',
    });
  }

  // DM Assistant endpoints
  async generateDMContent(type: string, context: any) {
    return this.request('/dm/assistant/generate', {
      method: 'POST',
      body: JSON.stringify({ type, context }),
    });
  }

  async saveDMNote(sessionId: string, note: any) {
    return this.request(`/dm/assistant/sessions/${sessionId}/notes`, {
      method: 'POST',
      body: JSON.stringify(note),
    });
  }

  async getDMNotes(sessionId: string) {
    return this.request(`/dm/assistant/sessions/${sessionId}/notes`);
  }
}

// Create singleton instance
const apiService = new ApiService();

export default apiService;