import { ApiService } from '../services/api.js';
import authService from '../services/auth.js';

export class EncounterBuilder {
    constructor(container) {
        this.container = container;
        this.api = new ApiService();
        this.currentSessionId = null;
        this.activeEncounter = null;
        this.currentTab = 'generate';
        this.user = authService.getCurrentUser();
        this.isDM = this.user && this.user.role === 'dm';
        
        if (!this.isDM) {
            this.container.innerHTML = '<div class="error-message">You must be a DM to access the Encounter Builder.</div>';
        }
        
        // Don't call async operations in constructor
        // Initialize must be called separately
    }
    
    // Call this method after creating the instance
    async initialize() {
        await this.init();
    }

    async init() {
        await this.loadGameSessions();
        this.render();
        this.attachEventListeners();
    }

    async loadGameSessions() {
        try {
            const sessions = await this.api.gameSession.getSessions();
            this.gameSessions = sessions.filter(s => s.dm_user_id === this.user.id);
        } catch (error) {
            console.error('Failed to load game sessions:', error);
            this.gameSessions = [];
        }
    }

    render() {
        this.container.innerHTML = `
            <div class="encounter-builder">
                <h2>Dynamic Encounter Builder ⚔️</h2>
                
                <div class="session-selector">
                    <label for="game-session">Select Game Session:</label>
                    <select id="game-session">
                        <option value="">-- Select a session --</option>
                        ${this.gameSessions.map(session => `
                            <option value="${session.id}">${session.name}</option>
                        `).join('')}
                    </select>
                </div>

                <div class="encounter-tabs">
                    <button class="tab-button ${this.currentTab === 'generate' ? 'active' : ''}" data-tab="generate">
                        Generate Encounter
                    </button>
                    <button class="tab-button ${this.currentTab === 'active' ? 'active' : ''}" data-tab="active">
                        Active Encounter
                    </button>
                    <button class="tab-button ${this.currentTab === 'history' ? 'active' : ''}" data-tab="history">
                        Encounter History
                    </button>
                    <button class="tab-button ${this.currentTab === 'templates' ? 'active' : ''}" data-tab="templates">
                        Templates
                    </button>
                </div>

                <div class="tab-content" id="tab-content">
                    ${this.renderTabContent()}
                </div>
            </div>
        `;
    }

    renderTabContent() {
        switch(this.currentTab) {
            case 'generate':
                return this.renderGenerateTab();
            case 'active':
                return this.renderActiveTab();
            case 'history':
                return this.renderHistoryTab();
            case 'templates':
                return this.renderTemplatesTab();
            default:
                return '';
        }
    }

    renderGenerateTab() {
        return `
            <div class="generate-encounter">
                <h3>Generate New Encounter</h3>
                
                <form id="generate-encounter-form">
                    <div class="form-section">
                        <h4>Party Composition</h4>
                        <div id="party-members">
                            <div class="party-member">
                                <input type="number" name="level" placeholder="Level" min="1" max="20" required>
                                <input type="text" name="class" placeholder="Class" required>
                                <button type="button" class="remove-member" style="display: none;">Remove</button>
                            </div>
                        </div>
                        <button type="button" id="add-party-member">Add Party Member</button>
                    </div>

                    <div class="form-section">
                        <h4>Encounter Parameters</h4>
                        
                        <div class="form-group">
                            <label for="difficulty">Difficulty:</label>
                            <select id="difficulty" name="difficulty" required>
                                <option value="easy">Easy</option>
                                <option value="medium" selected>Medium</option>
                                <option value="hard">Hard</option>
                                <option value="deadly">Deadly</option>
                            </select>
                        </div>

                        <div class="form-group">
                            <label for="location">Location:</label>
                            <input type="text" id="location" name="location" placeholder="e.g., Ancient forest, Dungeon corridor" required>
                        </div>

                        <div class="form-group">
                            <label for="context">Story Context:</label>
                            <textarea id="context" name="context" rows="3" placeholder="Describe the current story situation..."></textarea>
                        </div>

                        <div class="form-group">
                            <label for="encounter-type">Encounter Type:</label>
                            <select id="encounter-type" name="encounter_type" required>
                                <option value="combat">Combat</option>
                                <option value="ambush">Ambush</option>
                                <option value="guard_post">Guard Post</option>
                                <option value="boss_fight">Boss Fight</option>
                                <option value="puzzle_combat">Puzzle Combat</option>
                                <option value="chase">Chase</option>
                                <option value="negotiation">Negotiation</option>
                            </select>
                        </div>

                        <div class="form-group">
                            <label for="special-conditions">Special Conditions (optional):</label>
                            <textarea id="special-conditions" name="special_conditions" rows="2" placeholder="e.g., Limited visibility, time pressure, protect an NPC"></textarea>
                        </div>
                    </div>

                    <button type="submit" class="primary-button">Generate Encounter</button>
                </form>

                <div id="generated-encounter" class="hidden"></div>
            </div>
        `;
    }

    renderActiveTab() {
        if (!this.activeEncounter) {
            return `
                <div class="no-active-encounter">
                    <p>No active encounter. Generate a new encounter or select one from history.</p>
                </div>
            `;
        }

        return `
            <div class="active-encounter">
                <h3>${this.activeEncounter.name}</h3>
                <p class="encounter-description">${this.activeEncounter.description}</p>
                
                <div class="encounter-stats">
                    <div class="stat-box">
                        <h4>Difficulty</h4>
                        <p>${this.activeEncounter.difficulty}</p>
                    </div>
                    <div class="stat-box">
                        <h4>XP Budget</h4>
                        <p>${this.activeEncounter.xp_budget}</p>
                    </div>
                    <div class="stat-box">
                        <h4>Status</h4>
                        <p>${this.activeEncounter.status}</p>
                    </div>
                </div>

                <div class="encounter-sections">
                    <div class="section enemies-section">
                        <h4>Enemies</h4>
                        <div class="enemies-list">
                            ${this.renderEnemiesList()}
                        </div>
                    </div>

                    <div class="section objectives-section">
                        <h4>Objectives</h4>
                        <div class="objectives-list">
                            ${this.renderObjectivesList()}
                        </div>
                    </div>

                    <div class="section environment-section">
                        <h4>Environmental Features</h4>
                        <div class="environment-features">
                            ${this.renderEnvironmentFeatures()}
                        </div>
                    </div>

                    <div class="section tactics-section">
                        <h4>Tactical Suggestions</h4>
                        <button id="get-tactical-suggestion" class="secondary-button">Get AI Suggestion</button>
                        <div id="tactical-suggestion" class="suggestion-box hidden"></div>
                    </div>
                </div>

                <div class="encounter-controls">
                    <button id="scale-encounter" class="secondary-button">Scale Encounter</button>
                    <button id="trigger-reinforcements" class="secondary-button">Trigger Reinforcements</button>
                    <button id="check-objectives" class="secondary-button">Check Objectives</button>
                    <button id="complete-encounter" class="primary-button">Complete Encounter</button>
                </div>

                <div class="event-log">
                    <h4>Event Log</h4>
                    <div id="event-log-entries"></div>
                    <form id="log-event-form" class="log-event-form">
                        <input type="text" id="event-description" placeholder="Log an event..." required>
                        <button type="submit">Log Event</button>
                    </form>
                </div>
            </div>
        `;
    }

    renderEnemiesList() {
        if (!this.activeEncounter || !this.activeEncounter.enemies) {
            return '<p>No enemies in this encounter.</p>';
        }

        return this.activeEncounter.enemies.map(enemy => `
            <div class="enemy-card ${enemy.status !== 'active' ? 'defeated' : ''}">
                <h5>${enemy.name}</h5>
                <div class="enemy-stats">
                    <span>CR: ${enemy.challenge_rating}</span>
                    <span>HP: ${enemy.hit_points}</span>
                    <span>AC: ${enemy.armor_class}</span>
                </div>
                <div class="enemy-tactics">
                    <strong>Tactics:</strong> ${enemy.tactics || 'Standard combat tactics'}
                </div>
                ${enemy.special_abilities ? `
                    <div class="enemy-abilities">
                        <strong>Special Abilities:</strong> ${enemy.special_abilities}
                    </div>
                ` : ''}
                <div class="enemy-controls">
                    <button class="update-enemy-status" data-enemy-id="${enemy.id}">
                        ${enemy.status === 'active' ? 'Mark Defeated' : 'Mark Active'}
                    </button>
                </div>
            </div>
        `).join('');
    }

    renderObjectivesList() {
        if (!this.activeEncounter || !this.activeEncounter.objectives) {
            return '<p>No objectives for this encounter.</p>';
        }

        return this.activeEncounter.objectives.map(objective => `
            <div class="objective-card ${objective.is_completed ? 'completed' : ''}">
                <h5>${objective.type.replace('_', ' ').toUpperCase()}</h5>
                <p>${objective.description}</p>
                ${objective.conditions ? `
                    <div class="objective-conditions">
                        <strong>Conditions:</strong> ${JSON.stringify(objective.conditions)}
                    </div>
                ` : ''}
                ${objective.rewards ? `
                    <div class="objective-rewards">
                        <strong>Rewards:</strong> ${JSON.stringify(objective.rewards)}
                    </div>
                ` : ''}
            </div>
        `).join('');
    }

    renderEnvironmentFeatures() {
        if (!this.activeEncounter || !this.activeEncounter.environment_features) {
            return '<p>No special environmental features.</p>';
        }

        const features = this.activeEncounter.environment_features;
        return `
            ${features.hazards ? `
                <div class="feature-section">
                    <strong>Hazards:</strong>
                    <ul>
                        ${features.hazards.map(hazard => `<li>${hazard}</li>`).join('')}
                    </ul>
                </div>
            ` : ''}
            ${features.interactive_elements ? `
                <div class="feature-section">
                    <strong>Interactive Elements:</strong>
                    <ul>
                        ${features.interactive_elements.map(element => `<li>${element}</li>`).join('')}
                    </ul>
                </div>
            ` : ''}
            ${features.cover_positions ? `
                <div class="feature-section">
                    <strong>Cover Positions:</strong>
                    <ul>
                        ${features.cover_positions.map(position => `<li>${position}</li>`).join('')}
                    </ul>
                </div>
            ` : ''}
            ${features.escape_routes ? `
                <div class="feature-section">
                    <strong>Escape Routes:</strong>
                    <ul>
                        ${features.escape_routes.map(route => `<li>${route}</li>`).join('')}
                    </ul>
                </div>
            ` : ''}
        `;
    }

    renderHistoryTab() {
        return `
            <div class="encounter-history">
                <h3>Encounter History</h3>
                <div id="encounter-list" class="encounter-list">
                    <p>Loading encounters...</p>
                </div>
            </div>
        `;
    }

    renderTemplatesTab() {
        return `
            <div class="encounter-templates">
                <h3>Encounter Templates</h3>
                <p>Save and reuse encounter templates for quick setup.</p>
                <div id="templates-list" class="templates-list">
                    <p>Templates feature coming soon...</p>
                </div>
            </div>
        `;
    }

    attachEventListeners() {
        // Session selector
        const sessionSelector = this.container.querySelector('#game-session');
        if (sessionSelector) {
            sessionSelector.addEventListener('change', (e) => {
                this.currentSessionId = e.target.value;
                if (this.currentTab === 'history') {
                    this.loadEncounterHistory();
                }
            });
        }

        // Tab navigation
        this.container.querySelectorAll('.tab-button').forEach(button => {
            button.addEventListener('click', (e) => {
                this.currentTab = e.target.dataset.tab;
                this.render();
                this.attachEventListeners();
                
                if (this.currentTab === 'history' && this.currentSessionId) {
                    this.loadEncounterHistory();
                }
            });
        });

        // Generate encounter form
        const generateForm = this.container.querySelector('#generate-encounter-form');
        if (generateForm) {
            generateForm.addEventListener('submit', (e) => this.handleGenerateEncounter(e));
        }

        // Add party member button
        const addMemberBtn = this.container.querySelector('#add-party-member');
        if (addMemberBtn) {
            addMemberBtn.addEventListener('click', () => this.addPartyMember());
        }

        // Remove party member buttons
        this.container.querySelectorAll('.remove-member').forEach(button => {
            button.addEventListener('click', (e) => this.removePartyMember(e));
        });

        // Active encounter controls
        if (this.activeEncounter) {
            this.attachActiveEncounterListeners();
        }
    }

    attachActiveEncounterListeners() {
        // Tactical suggestion
        const tacticalBtn = this.container.querySelector('#get-tactical-suggestion');
        if (tacticalBtn) {
            tacticalBtn.addEventListener('click', () => this.getTacticalSuggestion());
        }

        // Scale encounter
        const scaleBtn = this.container.querySelector('#scale-encounter');
        if (scaleBtn) {
            scaleBtn.addEventListener('click', () => this.showScaleDialog());
        }

        // Trigger reinforcements
        const reinforcementsBtn = this.container.querySelector('#trigger-reinforcements');
        if (reinforcementsBtn) {
            reinforcementsBtn.addEventListener('click', () => this.triggerReinforcements());
        }

        // Check objectives
        const checkObjectivesBtn = this.container.querySelector('#check-objectives');
        if (checkObjectivesBtn) {
            checkObjectivesBtn.addEventListener('click', () => this.checkObjectives());
        }

        // Complete encounter
        const completeBtn = this.container.querySelector('#complete-encounter');
        if (completeBtn) {
            completeBtn.addEventListener('click', () => this.completeEncounter());
        }

        // Log event form
        const logEventForm = this.container.querySelector('#log-event-form');
        if (logEventForm) {
            logEventForm.addEventListener('submit', (e) => this.logEvent(e));
        }

        // Update enemy status buttons
        this.container.querySelectorAll('.update-enemy-status').forEach(button => {
            button.addEventListener('click', (e) => this.updateEnemyStatus(e));
        });

        // Load event log
        this.loadEventLog();
    }

    addPartyMember() {
        const partyMembers = this.container.querySelector('#party-members');
        const newMember = document.createElement('div');
        newMember.className = 'party-member';
        newMember.innerHTML = `
            <input type="number" name="level" placeholder="Level" min="1" max="20" required>
            <input type="text" name="class" placeholder="Class" required>
            <button type="button" class="remove-member">Remove</button>
        `;
        
        partyMembers.appendChild(newMember);
        
        // Attach remove listener
        newMember.querySelector('.remove-member').addEventListener('click', (e) => this.removePartyMember(e));
        
        // Show remove buttons if more than one member
        if (partyMembers.children.length > 1) {
            partyMembers.querySelectorAll('.remove-member').forEach(btn => btn.style.display = 'inline-block');
        }
    }

    removePartyMember(e) {
        const partyMembers = this.container.querySelector('#party-members');
        e.target.closest('.party-member').remove();
        
        // Hide remove buttons if only one member left
        if (partyMembers.children.length === 1) {
            partyMembers.querySelector('.remove-member').style.display = 'none';
        }
    }

    async handleGenerateEncounter(e) {
        e.preventDefault();
        
        if (!this.currentSessionId) {
            alert('Please select a game session first.');
            return;
        }

        const formData = new FormData(e.target);
        const partyMembers = [];
        
        // Collect party members
        this.container.querySelectorAll('.party-member').forEach(member => {
            const level = member.querySelector('input[name="level"]').value;
            const className = member.querySelector('input[name="class"]').value;
            if (level && className) {
                partyMembers.push({ level: parseInt(level), class: className });
            }
        });

        const requestData = {
            party_composition: partyMembers,
            difficulty: formData.get('difficulty'),
            location: formData.get('location'),
            story_context: formData.get('context'),
            encounter_type: formData.get('encounter_type'),
            special_conditions: formData.get('special_conditions')
        };

        try {
            const encounter = await this.api.encounters.generate(requestData);
            this.activeEncounter = encounter;
            this.showGeneratedEncounter(encounter);
        } catch (error) {
            console.error('Failed to generate encounter:', error);
            alert('Failed to generate encounter. Please try again.');
        }
    }

    showGeneratedEncounter(encounter) {
        const container = this.container.querySelector('#generated-encounter');
        container.classList.remove('hidden');
        
        container.innerHTML = `
            <div class="generated-encounter-preview">
                <h4>${encounter.name}</h4>
                <p>${encounter.description}</p>
                
                <div class="encounter-summary">
                    <div class="summary-item">
                        <strong>Difficulty:</strong> ${encounter.difficulty}
                    </div>
                    <div class="summary-item">
                        <strong>XP Budget:</strong> ${encounter.xp_budget}
                    </div>
                    <div class="summary-item">
                        <strong>Enemies:</strong> ${encounter.enemies ? encounter.enemies.length : 0}
                    </div>
                </div>

                <div class="encounter-actions">
                    <button id="start-encounter" class="primary-button">Start Encounter</button>
                    <button id="regenerate-encounter" class="secondary-button">Regenerate</button>
                </div>
            </div>
        `;

        // Attach listeners for the new buttons
        container.querySelector('#start-encounter').addEventListener('click', () => this.startEncounter());
        container.querySelector('#regenerate-encounter').addEventListener('click', () => {
            this.container.querySelector('#generate-encounter-form').dispatchEvent(new Event('submit'));
        });
    }

    async startEncounter() {
        if (!this.activeEncounter) return;

        try {
            await this.api.encounters.start(this.activeEncounter.id);
            this.activeEncounter.status = 'in_progress';
            this.currentTab = 'active';
            this.render();
            this.attachEventListeners();
        } catch (error) {
            console.error('Failed to start encounter:', error);
            alert('Failed to start encounter.');
        }
    }

    async getTacticalSuggestion() {
        if (!this.activeEncounter) return;

        const suggestionBox = this.container.querySelector('#tactical-suggestion');
        suggestionBox.innerHTML = '<p>Getting tactical suggestion...</p>';
        suggestionBox.classList.remove('hidden');

        try {
            const response = await this.api.encounters.getTacticalSuggestion(this.activeEncounter.id, {
                current_round: 1, // You might want to track this
                party_status: 'healthy' // You might want to get actual party status
            });
            
            suggestionBox.innerHTML = `
                <h5>AI Tactical Suggestion</h5>
                <p>${response.suggestion}</p>
                <small>Generated at ${new Date().toLocaleTimeString()}</small>
            `;
        } catch (error) {
            console.error('Failed to get tactical suggestion:', error);
            suggestionBox.innerHTML = '<p class="error">Failed to get suggestion.</p>';
        }
    }

    showScaleDialog() {
        const dialog = document.createElement('div');
        dialog.className = 'modal-overlay';
        dialog.innerHTML = `
            <div class="modal-content">
                <h3>Scale Encounter</h3>
                <form id="scale-form">
                    <div class="form-group">
                        <label for="scale-direction">Scale Direction:</label>
                        <select id="scale-direction" name="direction" required>
                            <option value="up">Scale Up (Harder)</option>
                            <option value="down">Scale Down (Easier)</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="scale-factor">Scale Factor:</label>
                        <input type="number" id="scale-factor" name="factor" min="0.5" max="2" step="0.1" value="1.2" required>
                    </div>
                    <div class="modal-actions">
                        <button type="submit" class="primary-button">Apply Scaling</button>
                        <button type="button" class="secondary-button" id="cancel-scale">Cancel</button>
                    </div>
                </form>
            </div>
        `;
        
        document.body.appendChild(dialog);
        
        dialog.querySelector('#scale-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            
            try {
                await this.api.encounters.scale(this.activeEncounter.id, {
                    direction: formData.get('direction'),
                    factor: parseFloat(formData.get('factor'))
                });
                
                // Reload encounter
                await this.loadActiveEncounter();
                dialog.remove();
            } catch (error) {
                console.error('Failed to scale encounter:', error);
                alert('Failed to scale encounter.');
            }
        });
        
        dialog.querySelector('#cancel-scale').addEventListener('click', () => dialog.remove());
    }

    async triggerReinforcements() {
        if (!this.activeEncounter) return;

        if (confirm('Trigger reinforcement wave?')) {
            try {
                const response = await this.api.encounters.triggerReinforcements(this.activeEncounter.id, {
                    wave_number: 1 // You might want to track this
                });
                
                alert(`Reinforcements arrived! ${response.enemies_added} new enemies join the battle.`);
                await this.loadActiveEncounter();
            } catch (error) {
                console.error('Failed to trigger reinforcements:', error);
                alert('Failed to trigger reinforcements.');
            }
        }
    }

    async checkObjectives() {
        if (!this.activeEncounter) return;

        try {
            const response = await this.api.encounters.checkObjectives(this.activeEncounter.id);
            
            let message = 'Objective Status:\n\n';
            response.objectives.forEach(obj => {
                message += `${obj.type}: ${obj.is_completed ? '✓ Completed' : '○ In Progress'}\n`;
            });
            
            if (response.encounter_complete) {
                message += '\nAll objectives completed! Encounter can be finished.';
            }
            
            alert(message);
        } catch (error) {
            console.error('Failed to check objectives:', error);
            alert('Failed to check objectives.');
        }
    }

    async completeEncounter() {
        if (!this.activeEncounter) return;

        if (confirm('Complete this encounter?')) {
            try {
                await this.api.encounters.complete(this.activeEncounter.id, {
                    outcome: 'victory', // You might want to let DM choose
                    casualties: 0,
                    resources_used: {}
                });
                
                this.activeEncounter = null;
                this.currentTab = 'history';
                this.render();
                this.attachEventListeners();
                this.loadEncounterHistory();
            } catch (error) {
                console.error('Failed to complete encounter:', error);
                alert('Failed to complete encounter.');
            }
        }
    }

    async logEvent(e) {
        e.preventDefault();
        if (!this.activeEncounter) return;

        const description = e.target.event_description.value;
        
        try {
            await this.api.encounters.logEvent(this.activeEncounter.id, {
                event_type: 'custom',
                description: description,
                metadata: {}
            });
            
            e.target.reset();
            await this.loadEventLog();
        } catch (error) {
            console.error('Failed to log event:', error);
        }
    }

    async loadEventLog() {
        if (!this.activeEncounter) return;

        try {
            const events = await this.api.encounters.getEvents(this.activeEncounter.id);
            const logContainer = this.container.querySelector('#event-log-entries');
            
            logContainer.innerHTML = events.map(event => `
                <div class="event-entry">
                    <span class="event-time">${new Date(event.created_at).toLocaleTimeString()}</span>
                    <span class="event-type">[${event.event_type}]</span>
                    <span class="event-description">${event.description}</span>
                </div>
            `).join('');
        } catch (error) {
            console.error('Failed to load event log:', error);
        }
    }

    async updateEnemyStatus(e) {
        const enemyId = e.target.dataset.enemyId;
        const enemy = this.activeEncounter.enemies.find(e => e.id === enemyId);
        if (!enemy) return;

        const newStatus = enemy.status === 'active' ? 'defeated' : 'active';
        
        try {
            await this.api.encounters.updateEnemyStatus(this.activeEncounter.id, enemyId, {
                status: newStatus
            });
            
            enemy.status = newStatus;
            this.render();
            this.attachEventListeners();
        } catch (error) {
            console.error('Failed to update enemy status:', error);
        }
    }

    async loadActiveEncounter() {
        if (!this.activeEncounter) return;

        try {
            this.activeEncounter = await this.api.encounters.get(this.activeEncounter.id);
            this.render();
            this.attachEventListeners();
        } catch (error) {
            console.error('Failed to load active encounter:', error);
        }
    }

    async loadEncounterHistory() {
        if (!this.currentSessionId) return;

        const listContainer = this.container.querySelector('#encounter-list');
        listContainer.innerHTML = '<p>Loading...</p>';

        try {
            const encounters = await this.api.encounters.getBySession(this.currentSessionId);
            
            if (encounters.length === 0) {
                listContainer.innerHTML = '<p>No encounters found for this session.</p>';
                return;
            }

            listContainer.innerHTML = encounters.map(encounter => `
                <div class="encounter-history-item">
                    <h4>${encounter.name}</h4>
                    <p>${encounter.description}</p>
                    <div class="encounter-meta">
                        <span>Difficulty: ${encounter.difficulty}</span>
                        <span>Status: ${encounter.status}</span>
                        <span>Created: ${new Date(encounter.created_at).toLocaleDateString()}</span>
                    </div>
                    ${encounter.status !== 'completed' ? `
                        <button class="resume-encounter" data-encounter-id="${encounter.id}">
                            Resume
                        </button>
                    ` : ''}
                </div>
            `).join('');

            // Attach resume listeners
            listContainer.querySelectorAll('.resume-encounter').forEach(button => {
                button.addEventListener('click', async (e) => {
                    const encounterId = e.target.dataset.encounterId;
                    this.activeEncounter = await this.api.encounters.get(encounterId);
                    this.currentTab = 'active';
                    this.render();
                    this.attachEventListeners();
                });
            });
        } catch (error) {
            console.error('Failed to load encounter history:', error);
            listContainer.innerHTML = '<p class="error">Failed to load encounters.</p>';
        }
    }
}