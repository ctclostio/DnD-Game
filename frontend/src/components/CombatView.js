import { ApiService } from '../services/api.js';
import { WebSocketService } from '../services/websocket.js';

export class CombatView {
    constructor(container, api) {
        this.container = container;
        this.api = api || new ApiService();
        this.ws = new WebSocketService();
        
        this.sessionId = new URLSearchParams(window.location.search).get('sessionId') || 'default-session';
        this.combat = null;
        this.selectedCombatant = null;
        this.actionInProgress = false;
        this.combatLog = [];
        this.myCharacters = [];
        
        this.init();
    }
    
    async init() {
        await this.loadMyCharacters();
        await this.loadCombat();
        
        // Subscribe to combat updates
        this.ws.onMessage((message) => {
            if (message.type === 'combat') {
                this.handleCombatUpdate(message.data);
            }
        });
        
        this.render();
    }
    
    async loadCombat() {
        try {
            const response = await this.api.get(`/combat/session/${this.sessionId}`);
            this.combat = response;
            this.render();
        } catch (error) {
            console.error('Failed to load combat:', error);
        }
    }
    
    async loadMyCharacters() {
        try {
            const response = await this.api.get('/characters');
            this.myCharacters = response;
        } catch (error) {
            console.error('Failed to load characters:', error);
        }
    }
    
    handleCombatUpdate(update) {
        if (update.combat) {
            this.combat = update.combat;
        }
        
        if (update.message) {
            this.combatLog.unshift({
                message: update.message,
                timestamp: new Date(),
                type: update.type
            });
            
            // Keep only last 20 entries
            if (this.combatLog.length > 20) {
                this.combatLog = this.combatLog.slice(0, 20);
            }
        }
        
        this.render();
    }
    
    getCurrentCombatant() {
        if (!this.combat || !this.combat.combatants || this.combat.turnOrder.length === 0) return null;
        return this.combat.combatants.find(c => c.id === this.combat.turnOrder[this.combat.currentTurn]);
    }
    
    canControlCombatant(combatant) {
        if (!combatant || !this.myCharacters) return false;
        return this.myCharacters.some(char => char.id === combatant.characterId);
    }
    
    async handleAttack(targetId) {
        if (!this.selectedCombatant || this.actionInProgress) return;
        
        this.actionInProgress = true;
        try {
            await this.api.post(`/combat/${this.combat.id}/action`, {
                action: 'attack',
                actorId: this.selectedCombatant.id,
                targetId: targetId,
            });
        } catch (error) {
            console.error('Attack failed:', error);
        } finally {
            this.actionInProgress = false;
        }
    }
    
    async handleMove() {
        if (!this.selectedCombatant || this.actionInProgress) return;
        
        this.actionInProgress = true;
        try {
            await this.api.post(`/combat/${this.combat.id}/action`, {
                action: 'move',
                actorId: this.selectedCombatant.id,
            });
        } catch (error) {
            console.error('Move failed:', error);
        } finally {
            this.actionInProgress = false;
        }
    }
    
    async handleDash() {
        if (!this.selectedCombatant || this.actionInProgress) return;
        
        this.actionInProgress = true;
        try {
            await this.api.post(`/combat/${this.combat.id}/action`, {
                action: 'dash',
                actorId: this.selectedCombatant.id,
            });
        } catch (error) {
            console.error('Dash failed:', error);
        } finally {
            this.actionInProgress = false;
        }
    }
    
    async handleDodge() {
        if (!this.selectedCombatant || this.actionInProgress) return;
        
        this.actionInProgress = true;
        try {
            await this.api.post(`/combat/${this.combat.id}/action`, {
                action: 'dodge',
                actorId: this.selectedCombatant.id,
            });
        } catch (error) {
            console.error('Dodge failed:', error);
        } finally {
            this.actionInProgress = false;
        }
    }
    
    async handleDeathSave() {
        if (!this.selectedCombatant || this.actionInProgress) return;
        
        this.actionInProgress = true;
        try {
            await this.api.post(`/combat/${this.combat.id}/action`, {
                action: 'deathSave',
                actorId: this.selectedCombatant.id,
            });
        } catch (error) {
            console.error('Death save failed:', error);
        } finally {
            this.actionInProgress = false;
        }
    }
    
    async handleNextTurn() {
        if (this.actionInProgress) return;
        
        this.actionInProgress = true;
        try {
            await this.api.post(`/combat/${this.combat.id}/next-turn`);
        } catch (error) {
            console.error('Next turn failed:', error);
        } finally {
            this.actionInProgress = false;
        }
    }
    
    renderCombatant(combatant) {
        const currentCombatant = this.getCurrentCombatant();
        const isCurrent = currentCombatant?.id === combatant.id;
        const canControl = this.canControlCombatant(combatant);
        const isDead = combatant.hp <= 0 && combatant.deathSaves.isDead;
        const isUnconscious = combatant.hp <= 0 && !combatant.deathSaves.isDead && !combatant.deathSaves.isStable;
        const isStable = combatant.deathSaves.isStable;
        
        const hpPercent = Math.max(0, (combatant.hp / combatant.maxHp) * 100);
        
        return `
            <div class="combatant-card ${isCurrent ? 'current-turn' : ''} ${isDead ? 'dead' : ''} ${isUnconscious ? 'unconscious' : ''}"
                 data-combatant-id="${combatant.id}">
                <div class="combatant-header">
                    <h3>${combatant.name}</h3>
                    <span class="initiative">Initiative: ${combatant.initiative}</span>
                </div>
                
                <div class="combatant-stats">
                    <div class="hp-bar">
                        <div class="hp-fill" style="width: ${hpPercent}%"></div>
                        <span class="hp-text">${combatant.hp} / ${combatant.maxHp} HP</span>
                        ${combatant.tempHp > 0 ? `<span class="temp-hp">+${combatant.tempHp} Temp</span>` : ''}
                    </div>
                    
                    <div class="combat-info">
                        <span>AC: ${combatant.ac}</span>
                        <span>Speed: ${combatant.speed} ft</span>
                    </div>
                    
                    ${combatant.hp <= 0 && !isDead ? `
                        <div class="death-saves">
                            <span>Death Saves: </span>
                            <span class="successes">${'● '.repeat(combatant.deathSaves.successes || 0)}</span>
                            <span class="failures">${'● '.repeat(combatant.deathSaves.failures || 0)}</span>
                            ${isStable ? '<span class="stable">STABLE</span>' : ''}
                        </div>
                    ` : ''}
                    
                    <div class="conditions">
                        ${combatant.conditions ? combatant.conditions.map(condition => 
                            `<span class="condition ${condition}">${condition}</span>`
                        ).join('') : ''}
                        ${combatant.isConcentrating ? 
                            `<span class="condition concentration">Concentrating: ${combatant.concentrationSpell}</span>` : ''}
                    </div>
                    
                    <div class="action-economy">
                        <span class="${combatant.actions > 0 ? 'available' : 'used'}">
                            Actions: ${combatant.actions}
                        </span>
                        <span class="${combatant.bonusActions > 0 ? 'available' : 'used'}">
                            Bonus: ${combatant.bonusActions}
                        </span>
                        <span class="${combatant.reactions > 0 ? 'available' : 'used'}">
                            Reaction: ${combatant.reactions}
                        </span>
                        <span>Movement: ${combatant.movement} ft</span>
                    </div>
                </div>
            </div>
        `;
    }
    
    render() {
        if (!this.combat) {
            this.container.innerHTML = '<div class="combat-view loading">Loading combat...</div>';
            return;
        }
        
        const currentCombatant = this.getCurrentCombatant();
        const isMyTurn = currentCombatant && this.canControlCombatant(currentCombatant);
        
        const sortedCombatants = [...this.combat.combatants].sort((a, b) => {
            const aIndex = this.combat.turnOrder.indexOf(a.id);
            const bIndex = this.combat.turnOrder.indexOf(b.id);
            return aIndex - bIndex;
        });
        
        this.container.innerHTML = `
            <div class="combat-view">
                <div class="combat-header">
                    <h2>Combat - Round ${this.combat.round}</h2>
                    <div class="turn-info">
                        ${currentCombatant ? `<span>Current Turn: ${currentCombatant.name}</span>` : ''}
                    </div>
                </div>

                <div class="combat-layout">
                    <div class="combatants-section">
                        <h3>Combatants</h3>
                        <div class="combatants-list">
                            ${sortedCombatants.map(c => this.renderCombatant(c)).join('')}
                        </div>
                    </div>

                    <div class="actions-section">
                        <h3>Actions</h3>
                        ${isMyTurn && this.selectedCombatant && this.selectedCombatant.id === currentCombatant.id ? `
                            <div class="action-buttons">
                                ${this.selectedCombatant.hp > 0 ? `
                                    <div class="action-group">
                                        <h4>Combat Actions</h4>
                                        <div class="targets">
                                            ${this.combat.combatants
                                                .filter(c => c.id !== this.selectedCombatant.id && c.hp > 0)
                                                .map(target => `
                                                    <button class="attack-button" data-target-id="${target.id}"
                                                        ${this.actionInProgress || this.selectedCombatant.actions <= 0 ? 'disabled' : ''}>
                                                        Attack ${target.name}
                                                    </button>
                                                `).join('')}
                                        </div>
                                    </div>
                                    
                                    <div class="action-group">
                                        <h4>Other Actions</h4>
                                        <button class="move-button"
                                            ${this.actionInProgress || this.selectedCombatant.movement <= 0 ? 'disabled' : ''}>
                                            Move (${this.selectedCombatant.movement} ft remaining)
                                        </button>
                                        <button class="dash-button"
                                            ${this.actionInProgress || this.selectedCombatant.actions <= 0 ? 'disabled' : ''}>
                                            Dash
                                        </button>
                                        <button class="dodge-button"
                                            ${this.actionInProgress || this.selectedCombatant.actions <= 0 ? 'disabled' : ''}>
                                            Dodge
                                        </button>
                                    </div>
                                ` : `
                                    <div class="action-group">
                                        <h4>Death Save</h4>
                                        <button class="death-save-button"
                                            ${this.actionInProgress || this.selectedCombatant.deathSaves.isStable || this.selectedCombatant.deathSaves.isDead ? 'disabled' : ''}>
                                            Roll Death Save
                                        </button>
                                    </div>
                                `}
                                
                                <button class="end-turn-button" ${this.actionInProgress ? 'disabled' : ''}>
                                    End Turn
                                </button>
                            </div>
                        ` : `
                            <div class="waiting-message">
                                ${currentCombatant ? `Waiting for ${currentCombatant.name}'s turn...` : 'Waiting...'}
                            </div>
                        `}
                    </div>

                    <div class="combat-log-section">
                        <h3>Combat Log</h3>
                        <div class="combat-log">
                            ${this.combatLog.map(entry => `
                                <div class="log-entry ${entry.type}">
                                    <span class="timestamp">${entry.timestamp.toLocaleTimeString()}</span>
                                    <span class="message">${entry.message}</span>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                </div>
            </div>
        `;
        
        this.attachEventListeners();
    }
    
    attachEventListeners() {
        // Combatant selection
        this.container.querySelectorAll('.combatant-card').forEach(card => {
            card.addEventListener('click', (e) => {
                const combatantId = e.currentTarget.dataset.combatantId;
                const combatant = this.combat.combatants.find(c => c.id === combatantId);
                if (this.canControlCombatant(combatant)) {
                    this.selectedCombatant = combatant;
                    this.render();
                }
            });
        });
        
        // Attack buttons
        this.container.querySelectorAll('.attack-button').forEach(button => {
            button.addEventListener('click', (e) => {
                const targetId = e.target.dataset.targetId;
                this.handleAttack(targetId);
            });
        });
        
        // Other action buttons
        const moveBtn = this.container.querySelector('.move-button');
        if (moveBtn) moveBtn.addEventListener('click', () => this.handleMove());
        
        const dashBtn = this.container.querySelector('.dash-button');
        if (dashBtn) dashBtn.addEventListener('click', () => this.handleDash());
        
        const dodgeBtn = this.container.querySelector('.dodge-button');
        if (dodgeBtn) dodgeBtn.addEventListener('click', () => this.handleDodge());
        
        const deathSaveBtn = this.container.querySelector('.death-save-button');
        if (deathSaveBtn) deathSaveBtn.addEventListener('click', () => this.handleDeathSave());
        
        const endTurnBtn = this.container.querySelector('.end-turn-button');
        if (endTurnBtn) endTurnBtn.addEventListener('click', () => this.handleNextTurn());
    }
}