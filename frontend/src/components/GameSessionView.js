export class GameSessionView {
    constructor(container, api) {
        this.container = container;
        this.api = api;
        this.currentSession = null;
        this.render();
    }

    render() {
        this.container.innerHTML = `
            <div class="game-session-view">
                <h2>Game Session</h2>
                
                <div class="session-actions">
                    <button id="create-session-btn">Create New Session</button>
                    <button id="join-session-btn">Join Session</button>
                </div>
                
                <div id="session-content"></div>
            </div>
        `;

        this.setupEventListeners();
        this.setupChatHandlers();
    }

    setupEventListeners() {
        document.getElementById('create-session-btn').addEventListener('click', () => {
            this.showCreateSessionForm();
        });

        document.getElementById('join-session-btn').addEventListener('click', () => {
            this.showJoinSessionForm();
        });
    }

    setupChatHandlers() {
        const chatInput = document.getElementById('chat-input');
        const chatSend = document.getElementById('chat-send');

        if (chatInput && chatSend) {
            chatSend.addEventListener('click', () => {
                const message = chatInput.value.trim();
                if (message && window.ws) {
                    window.ws.sendChatMessage(message);
                    chatInput.value = '';
                }
            });

            chatInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    const message = e.target.value.trim();
                    if (message && window.ws) {
                        window.ws.sendChatMessage(message);
                        e.target.value = '';
                    }
                }
            });
        }
    }

    showCreateSessionForm() {
        const content = document.getElementById('session-content');
        content.innerHTML = `
            <form id="create-session-form" class="session-form">
                <h3>Create Game Session</h3>
                
                <div class="form-group">
                    <label for="session-name">Session Name</label>
                    <input type="text" id="session-name" required>
                </div>
                
                <div class="form-group">
                    <label for="dm-name">Dungeon Master Name</label>
                    <input type="text" id="dm-name" required>
                </div>
                
                <div class="form-group">
                    <label for="max-players">Max Players</label>
                    <input type="number" id="max-players" min="2" max="8" value="4">
                </div>
                
                <button type="submit">Create Session</button>
                <button type="button" id="cancel-create">Cancel</button>
            </form>
        `;

        document.getElementById('create-session-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.createSession();
        });

        document.getElementById('cancel-create').addEventListener('click', () => {
            this.render();
        });
    }

    showJoinSessionForm() {
        const content = document.getElementById('session-content');
        content.innerHTML = `
            <form id="join-session-form" class="session-form">
                <h3>Join Game Session</h3>
                
                <div class="form-group">
                    <label for="session-id">Session ID</label>
                    <input type="text" id="session-id" required>
                </div>
                
                <div class="form-group">
                    <label for="player-name">Your Name</label>
                    <input type="text" id="player-name" required>
                </div>
                
                <button type="submit">Join Session</button>
                <button type="button" id="cancel-join">Cancel</button>
            </form>
        `;

        document.getElementById('join-session-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.joinSession();
        });

        document.getElementById('cancel-join').addEventListener('click', () => {
            this.render();
        });
    }

    async createSession() {
        const sessionData = {
            name: document.getElementById('session-name').value,
            dungeonMaster: document.getElementById('dm-name').value,
            maxPlayers: parseInt(document.getElementById('max-players').value)
        };

        try {
            const session = await this.api.createGameSession(sessionData);
            this.currentSession = session;
            sessionStorage.setItem('sessionId', session.id);
            this.showGameSession(session, true);
        } catch (error) {
            console.error('Failed to create session:', error);
            alert('Failed to create session. Please try again.');
        }
    }

    async joinSession() {
        const sessionId = document.getElementById('session-id').value;
        const playerName = document.getElementById('player-name').value;

        try {
            const session = await this.api.getGameSession(sessionId);
            this.currentSession = session;
            sessionStorage.setItem('sessionId', session.id);

            // Connect WebSocket
            if (window.ws) {
                window.ws.connect(sessionId, playerName);
            }
            
            this.showGameSession(session, false);
        } catch (error) {
            console.error('Failed to join session:', error);
            alert('Failed to join session. Please check the session ID.');
        }
    }

    showGameSession(session, isDM) {
        const content = document.getElementById('session-content');
        content.innerHTML = `
            <div class="game-session">
                <div class="session-header">
                    <h3>${session.name}</h3>
                    <p>Session ID: <strong>${session.id}</strong></p>
                    <p>DM: ${session.dungeonMaster}</p>
                    <p>Status: <span class="session-status ${session.status}">${session.status}</span></p>
                </div>
                
                <div class="players-section">
                    <h4>Players (${session.players?.length || 0})</h4>
                    <div class="players-list">
                        ${session.players?.map(player => `
                            <div class="player-item ${player.isOnline ? 'online' : 'offline'}">
                                <span class="player-name">${player.name}</span>
                                <span class="player-status">${player.isOnline ? '●' : '○'}</span>
                            </div>
                        `).join('') || '<p>No players yet</p>'}
                    </div>
                </div>
                
                <div class="game-controls">
                    ${isDM ? `
                        <button id="start-game-btn" ${session.status === 'active' ? 'disabled' : ''}>
                            Start Game
                        </button>
                        <button id="pause-game-btn" ${session.status !== 'active' ? 'disabled' : ''}>
                            Pause Game
                        </button>
                    ` : ''}
                    
                    <button id="quick-roll-btn">Quick Roll</button>
                    <button id="leave-session-btn">Leave Session</button>
                </div>
                
                <div class="quick-actions">
                    <h4>Quick Actions</h4>
                    <div class="action-buttons">
                        <button class="action-btn" data-action="attack">Attack Roll</button>
                        <button class="action-btn" data-action="skill">Skill Check</button>
                        <button class="action-btn" data-action="save">Saving Throw</button>
                        <button class="action-btn" data-action="initiative">Initiative</button>
                    </div>
                </div>
            </div>
        `;

        this.setupGameControls(isDM);
    }

    setupGameControls(isDM) {
        // Quick roll button
        document.getElementById('quick-roll-btn')?.addEventListener('click', () => {
            this.showQuickRollDialog();
        });

        // Leave session button
        document.getElementById('leave-session-btn')?.addEventListener('click', () => {
            if (confirm('Are you sure you want to leave this session?')) {
                this.leaveSession();
            }
        });

        // DM controls
        if (isDM) {
            document.getElementById('start-game-btn')?.addEventListener('click', () => {
                this.updateSessionStatus('active');
            });

            document.getElementById('pause-game-btn')?.addEventListener('click', () => {
                this.updateSessionStatus('paused');
            });
        }

        // Quick action buttons
        document.querySelectorAll('.action-btn').forEach(button => {
            button.addEventListener('click', (e) => {
                const action = e.target.dataset.action;
                this.performQuickAction(action);
            });
        });
    }

    showQuickRollDialog() {
        const dialog = document.createElement('div');
        dialog.className = 'quick-roll-dialog';
        dialog.innerHTML = `
            <div class="dialog-content">
                <h4>Quick Roll</h4>
                <select id="quick-roll-type">
                    <option value="1d20">d20</option>
                    <option value="2d6">2d6</option>
                    <option value="1d8">d8</option>
                    <option value="1d6">d6</option>
                    <option value="1d4">d4</option>
                </select>
                <input type="text" id="quick-roll-purpose" placeholder="Purpose (optional)">
                <button id="perform-quick-roll">Roll</button>
                <button id="cancel-quick-roll">Cancel</button>
            </div>
        `;

        document.body.appendChild(dialog);

        document.getElementById('perform-quick-roll').addEventListener('click', async () => {
            const diceType = document.getElementById('quick-roll-type').value;
            const purpose = document.getElementById('quick-roll-purpose').value;
            
            try {
                const result = await this.api.rollDice(diceType, purpose);
                
                // Send to WebSocket if connected
                if (window.ws) {
                    window.ws.sendDiceRoll(diceType, result, purpose);
                }
                
                document.body.removeChild(dialog);
            } catch (error) {
                console.error('Failed to roll dice:', error);
            }
        });

        document.getElementById('cancel-quick-roll').addEventListener('click', () => {
            document.body.removeChild(dialog);
        });
    }

    async performQuickAction(action) {
        let diceType = '1d20';
        let purpose = '';

        switch (action) {
            case 'attack':
                purpose = 'Attack Roll';
                break;
            case 'skill':
                purpose = 'Skill Check';
                break;
            case 'save':
                purpose = 'Saving Throw';
                break;
            case 'initiative':
                purpose = 'Initiative';
                break;
        }

        try {
            const result = await this.api.rollDice(diceType, purpose);
            
            // Send to WebSocket if connected
            if (window.ws) {
                window.ws.sendDiceRoll(diceType, result, purpose);
            }
        } catch (error) {
            console.error('Failed to perform quick action:', error);
        }
    }

    leaveSession() {
        if (window.ws) {
            window.ws.disconnect();
        }
        this.currentSession = null;
        this.render();
    }

    async updateSessionStatus(status) {
        // This would update the session status on the server
        console.debug(`Updating session status to: ${status}`);
        // Implementation would depend on your API
    }
}