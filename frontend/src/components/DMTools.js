import { api } from '../services/api.js';

export class DMTools {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.gameSessionId = null;
        this.npcs = [];
        this.templates = [];
        this.selectedNPC = null;
    }

    setGameSession(gameSessionId) {
        this.gameSessionId = gameSessionId;
        this.loadNPCs();
        this.loadTemplates();
    }

    async loadNPCs() {
        if (!this.gameSessionId) return;

        try {
            const response = await api.get(`/npcs/session/${this.gameSessionId}`);
            this.npcs = response;
            this.render();
        } catch (error) {
            console.error('Failed to load NPCs:', error);
        }
    }

    async loadTemplates() {
        try {
            const response = await api.get('/npcs/templates');
            this.templates = response;
        } catch (error) {
            console.error('Failed to load templates:', error);
        }
    }

    render() {
        if (!this.gameSessionId) {
            this.container.innerHTML = '<p>No game session selected</p>';
            return;
        }

        this.container.innerHTML = `
            <div class="dm-tools">
                <div class="dm-tools-header">
                    <h2>Dungeon Master Tools</h2>
                    <button class="btn btn-primary" onclick="dmTools.showCreateNPCForm()">
                        Add NPC/Monster
                    </button>
                </div>

                <div class="dm-tools-content">
                    <div class="npc-list-section">
                        <h3>NPCs & Monsters</h3>
                        ${this.renderNPCList()}
                    </div>

                    <div class="npc-details-section">
                        ${this.selectedNPC ? this.renderNPCDetails(this.selectedNPC) : '<p>Select an NPC to view details</p>'}
                    </div>
                </div>

                <div id="npc-modal" class="modal hidden"></div>
            </div>
        `;
    }

    renderNPCList() {
        if (this.npcs.length === 0) {
            return '<p class="no-npcs">No NPCs in this session yet</p>';
        }

        const groupedNPCs = this.groupNPCsByType();
        
        return Object.entries(groupedNPCs).map(([type, npcs]) => `
            <div class="npc-type-group">
                <h4>${this.formatType(type)}</h4>
                <div class="npc-cards">
                    ${npcs.map(npc => `
                        <div class="npc-card ${this.selectedNPC?.id === npc.id ? 'selected' : ''}" 
                             onclick="dmTools.selectNPC('${npc.id}')">
                            <div class="npc-card-header">
                                <span class="npc-name">${npc.name}</span>
                                <span class="npc-cr">CR ${npc.challengeRating}</span>
                            </div>
                            <div class="npc-card-body">
                                <div class="npc-hp">
                                    <div class="hp-bar">
                                        <div class="hp-fill" style="width: ${(npc.hitPoints / npc.maxHitPoints) * 100}%"></div>
                                    </div>
                                    <span class="hp-text">${npc.hitPoints}/${npc.maxHitPoints} HP</span>
                                </div>
                                <div class="npc-info">
                                    <span>${npc.size} ${npc.type}</span>
                                    <span>AC ${npc.armorClass}</span>
                                </div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `).join('');
    }

    renderNPCDetails(npc) {
        return `
            <div class="npc-details">
                <div class="npc-details-header">
                    <h3>${npc.name}</h3>
                    <div class="npc-actions">
                        <button class="btn btn-sm" onclick="dmTools.editNPC('${npc.id}')">Edit</button>
                        <button class="btn btn-sm btn-danger" onclick="dmTools.deleteNPC('${npc.id}')">Delete</button>
                    </div>
                </div>

                <div class="npc-stats">
                    <div class="stat-row">
                        <span>${npc.size} ${npc.type}, ${npc.alignment || 'unaligned'}</span>
                    </div>
                    <div class="stat-row">
                        <strong>Armor Class:</strong> ${npc.armorClass}
                    </div>
                    <div class="stat-row">
                        <strong>Hit Points:</strong> 
                        <span class="hp-current">${npc.hitPoints}/${npc.maxHitPoints}</span>
                        <div class="hp-controls">
                            <button onclick="dmTools.quickDamage('${npc.id}')">Damage</button>
                            <button onclick="dmTools.quickHeal('${npc.id}')">Heal</button>
                        </div>
                    </div>
                    <div class="stat-row">
                        <strong>Speed:</strong> ${this.formatSpeed(npc.speed)}
                    </div>
                </div>

                <div class="ability-scores">
                    <h4>Ability Scores</h4>
                    <div class="abilities-grid">
                        ${this.renderAbilityScores(npc.attributes)}
                    </div>
                </div>

                ${npc.savingThrows ? this.renderSavingThrows(npc.savingThrows) : ''}
                ${npc.skills?.length > 0 ? this.renderSkills(npc.skills) : ''}
                ${this.renderResistancesAndImmunities(npc)}
                ${npc.senses ? this.renderSenses(npc.senses) : ''}
                ${npc.languages?.length > 0 ? `<p><strong>Languages:</strong> ${npc.languages.join(', ')}</p>` : ''}
                <p><strong>Challenge:</strong> ${npc.challengeRating} (${npc.experiencePoints} XP)</p>

                ${npc.abilities?.length > 0 ? this.renderAbilities(npc.abilities) : ''}
                ${npc.actions?.length > 0 ? this.renderActions(npc.actions) : ''}

                <div class="combat-actions">
                    <button class="btn btn-primary" onclick="dmTools.rollInitiative('${npc.id}')">
                        Roll Initiative
                    </button>
                    <button class="btn" onclick="dmTools.addToCombat('${npc.id}')">
                        Add to Combat
                    </button>
                </div>
            </div>
        `;
    }

    renderAbilityScores(attributes) {
        return Object.entries(attributes).map(([ability, score]) => `
            <div class="ability-score">
                <div class="ability-name">${ability.toUpperCase().slice(0, 3)}</div>
                <div class="ability-value">${score}</div>
                <div class="ability-modifier">${this.getModifierString(score)}</div>
            </div>
        `).join('');
    }

    renderSavingThrows(saves) {
        const proficientSaves = Object.entries(saves)
            .filter(([_, save]) => save.proficiency)
            .map(([ability, save]) => `${ability.slice(0, 3)} +${save.modifier}`);

        return proficientSaves.length > 0 ? 
            `<p><strong>Saving Throws:</strong> ${proficientSaves.join(', ')}</p>` : '';
    }

    renderSkills(skills) {
        return `
            <p><strong>Skills:</strong> ${skills.map(skill => 
                `${skill.name} +${skill.modifier}`
            ).join(', ')}</p>
        `;
    }

    renderResistancesAndImmunities(npc) {
        let html = '';
        if (npc.damageResistances?.length > 0) {
            html += `<p><strong>Damage Resistances:</strong> ${npc.damageResistances.join(', ')}</p>`;
        }
        if (npc.damageImmunities?.length > 0) {
            html += `<p><strong>Damage Immunities:</strong> ${npc.damageImmunities.join(', ')}</p>`;
        }
        if (npc.conditionImmunities?.length > 0) {
            html += `<p><strong>Condition Immunities:</strong> ${npc.conditionImmunities.join(', ')}</p>`;
        }
        return html;
    }

    renderSenses(senses) {
        const sensesList = Object.entries(senses)
            .map(([sense, range]) => `${sense} ${range} ft.`)
            .join(', ');
        return `<p><strong>Senses:</strong> ${sensesList}</p>`;
    }

    renderAbilities(abilities) {
        return `
            <div class="npc-abilities">
                <h4>Abilities</h4>
                ${abilities.map(ability => `
                    <div class="ability-block">
                        <p><strong>${ability.name}.</strong> ${ability.description}</p>
                    </div>
                `).join('')}
            </div>
        `;
    }

    renderActions(actions) {
        return `
            <div class="npc-actions">
                <h4>Actions</h4>
                ${actions.map(action => `
                    <div class="action-block">
                        <p><strong>${action.name}.</strong> 
                        ${action.attackBonus ? `<em>Melee/Ranged Attack:</em> +${action.attackBonus} to hit, ` : ''}
                        ${action.range ? `reach/range ${action.range}, ` : ''}
                        ${action.damage ? `${action.damage} ${action.damageType} damage. ` : ''}
                        ${action.description || ''}</p>
                    </div>
                `).join('')}
            </div>
        `;
    }

    showCreateNPCForm() {
        const modal = document.getElementById('npc-modal');
        modal.classList.remove('hidden');
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Add NPC/Monster</h3>
                    <button class="close-btn" onclick="dmTools.closeModal()">×</button>
                </div>
                <div class="modal-body">
                    <div class="create-options">
                        <button class="btn btn-primary" onclick="dmTools.showTemplateSelection()">
                            Choose from Templates
                        </button>
                        <button class="btn" onclick="dmTools.showCustomNPCForm()">
                            Create Custom NPC
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    showTemplateSelection() {
        const modal = document.getElementById('npc-modal');
        const groupedTemplates = {};
        
        // Group templates by CR
        this.templates.forEach(template => {
            const cr = template.challengeRating;
            if (!groupedTemplates[cr]) {
                groupedTemplates[cr] = [];
            }
            groupedTemplates[cr].push(template);
        });

        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Choose Template</h3>
                    <button class="close-btn" onclick="dmTools.closeModal()">×</button>
                </div>
                <div class="modal-body">
                    <div class="template-list">
                        ${Object.entries(groupedTemplates)
                            .sort(([a], [b]) => parseFloat(a) - parseFloat(b))
                            .map(([cr, templates]) => `
                                <div class="template-group">
                                    <h4>Challenge Rating ${cr}</h4>
                                    <div class="template-cards">
                                        ${templates.map(template => `
                                            <div class="template-card" onclick="dmTools.createFromTemplate('${template.id}')">
                                                <h5>${template.name}</h5>
                                                <p>${template.type} - ${template.size}</p>
                                                <p>AC ${template.armorClass}, HP ${template.hitDice}</p>
                                            </div>
                                        `).join('')}
                                    </div>
                                </div>
                            `).join('')}
                    </div>
                </div>
            </div>
        `;
    }

    showCustomNPCForm() {
        const modal = document.getElementById('npc-modal');
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Create Custom NPC</h3>
                    <button class="close-btn" onclick="dmTools.closeModal()">×</button>
                </div>
                <div class="modal-body">
                    <form id="custom-npc-form">
                        <div class="form-row">
                            <div class="form-group">
                                <label>Name</label>
                                <input type="text" name="name" required>
                            </div>
                            <div class="form-group">
                                <label>Type</label>
                                <select name="type" required>
                                    <option value="beast">Beast</option>
                                    <option value="humanoid">Humanoid</option>
                                    <option value="undead">Undead</option>
                                    <option value="construct">Construct</option>
                                    <option value="dragon">Dragon</option>
                                    <option value="elemental">Elemental</option>
                                    <option value="fey">Fey</option>
                                    <option value="fiend">Fiend</option>
                                    <option value="giant">Giant</option>
                                    <option value="monstrosity">Monstrosity</option>
                                    <option value="ooze">Ooze</option>
                                    <option value="plant">Plant</option>
                                </select>
                            </div>
                        </div>

                        <div class="form-row">
                            <div class="form-group">
                                <label>Size</label>
                                <select name="size" required>
                                    <option value="tiny">Tiny</option>
                                    <option value="small">Small</option>
                                    <option value="medium" selected>Medium</option>
                                    <option value="large">Large</option>
                                    <option value="huge">Huge</option>
                                    <option value="gargantuan">Gargantuan</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label>Alignment</label>
                                <input type="text" name="alignment" placeholder="e.g., neutral evil">
                            </div>
                        </div>

                        <div class="form-row">
                            <div class="form-group">
                                <label>Armor Class</label>
                                <input type="number" name="armorClass" value="10" min="1" required>
                            </div>
                            <div class="form-group">
                                <label>Hit Points</label>
                                <input type="number" name="maxHitPoints" value="10" min="1" required>
                            </div>
                            <div class="form-group">
                                <label>Challenge Rating</label>
                                <input type="number" name="challengeRating" value="0.25" min="0" step="0.125">
                            </div>
                        </div>

                        <h4>Ability Scores</h4>
                        <div class="ability-scores-form">
                            <div class="form-group">
                                <label>STR</label>
                                <input type="number" name="strength" value="10" min="1" max="30">
                            </div>
                            <div class="form-group">
                                <label>DEX</label>
                                <input type="number" name="dexterity" value="10" min="1" max="30">
                            </div>
                            <div class="form-group">
                                <label>CON</label>
                                <input type="number" name="constitution" value="10" min="1" max="30">
                            </div>
                            <div class="form-group">
                                <label>INT</label>
                                <input type="number" name="intelligence" value="10" min="1" max="30">
                            </div>
                            <div class="form-group">
                                <label>WIS</label>
                                <input type="number" name="wisdom" value="10" min="1" max="30">
                            </div>
                            <div class="form-group">
                                <label>CHA</label>
                                <input type="number" name="charisma" value="10" min="1" max="30">
                            </div>
                        </div>

                        <button type="submit" class="btn btn-primary">Create NPC</button>
                    </form>
                </div>
            </div>
        `;

        document.getElementById('custom-npc-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.createCustomNPC(new FormData(e.target));
        });
    }

    async createFromTemplate(templateId) {
        try {
            await api.post('/npcs/create-from-template', {
                templateId: templateId,
                gameSessionId: this.gameSessionId
            });
            
            this.closeModal();
            await this.loadNPCs();
        } catch (error) {
            console.error('Failed to create NPC from template:', error);
            alert('Failed to create NPC from template');
        }
    }

    async createCustomNPC(formData) {
        const npc = {
            gameSessionId: this.gameSessionId,
            name: formData.get('name'),
            type: formData.get('type'),
            size: formData.get('size'),
            alignment: formData.get('alignment'),
            armorClass: parseInt(formData.get('armorClass')),
            maxHitPoints: parseInt(formData.get('maxHitPoints')),
            hitPoints: parseInt(formData.get('maxHitPoints')),
            challengeRating: parseFloat(formData.get('challengeRating')),
            speed: { walk: 30 },
            attributes: {
                strength: parseInt(formData.get('strength')),
                dexterity: parseInt(formData.get('dexterity')),
                constitution: parseInt(formData.get('constitution')),
                intelligence: parseInt(formData.get('intelligence')),
                wisdom: parseInt(formData.get('wisdom')),
                charisma: parseInt(formData.get('charisma'))
            },
            abilities: [],
            actions: []
        };

        try {
            await api.post('/npcs', npc);
            this.closeModal();
            await this.loadNPCs();
        } catch (error) {
            console.error('Failed to create NPC:', error);
            alert('Failed to create NPC');
        }
    }

    async selectNPC(npcId) {
        this.selectedNPC = this.npcs.find(npc => npc.id === npcId);
        this.render();
    }

    async deleteNPC(npcId) {
        if (!confirm('Are you sure you want to delete this NPC?')) {
            return;
        }

        try {
            await api.delete(`/npcs/${npcId}`);
            this.npcs = this.npcs.filter(npc => npc.id !== npcId);
            if (this.selectedNPC?.id === npcId) {
                this.selectedNPC = null;
            }
            this.render();
        } catch (error) {
            console.error('Failed to delete NPC:', error);
            alert('Failed to delete NPC');
        }
    }

    async quickDamage(npcId) {
        const amount = prompt('Enter damage amount:');
        if (!amount || isNaN(amount)) return;

        const damageType = prompt('Enter damage type (optional):', 'slashing');

        try {
            const response = await api.post(`/npcs/${npcId}/action/damage`, {
                amount: parseInt(amount),
                damageType: damageType || 'untyped'
            });

            // Update local NPC data
            const index = this.npcs.findIndex(npc => npc.id === npcId);
            if (index !== -1) {
                this.npcs[index] = response;
                if (this.selectedNPC?.id === npcId) {
                    this.selectedNPC = response;
                }
            }
            this.render();
        } catch (error) {
            console.error('Failed to apply damage:', error);
            alert('Failed to apply damage');
        }
    }

    async quickHeal(npcId) {
        const amount = prompt('Enter healing amount:');
        if (!amount || isNaN(amount)) return;

        try {
            const response = await api.post(`/npcs/${npcId}/action/heal`, {
                amount: parseInt(amount)
            });

            // Update local NPC data
            const index = this.npcs.findIndex(npc => npc.id === npcId);
            if (index !== -1) {
                this.npcs[index] = response;
                if (this.selectedNPC?.id === npcId) {
                    this.selectedNPC = response;
                }
            }
            this.render();
        } catch (error) {
            console.error('Failed to heal NPC:', error);
            alert('Failed to heal NPC');
        }
    }

    async rollInitiative(npcId) {
        try {
            const response = await api.post(`/npcs/${npcId}/action/initiative`);
            alert(`Initiative rolled: ${response.initiative}`);
        } catch (error) {
            console.error('Failed to roll initiative:', error);
            alert('Failed to roll initiative');
        }
    }

    closeModal() {
        document.getElementById('npc-modal').classList.add('hidden');
    }

    // Helper methods

    groupNPCsByType() {
        return this.npcs.reduce((groups, npc) => {
            const type = npc.type || 'other';
            if (!groups[type]) {
                groups[type] = [];
            }
            groups[type].push(npc);
            return groups;
        }, {});
    }

    formatType(type) {
        return type.charAt(0).toUpperCase() + type.slice(1) + 's';
    }

    formatSpeed(speed) {
        if (!speed) return '30 ft.';
        return Object.entries(speed)
            .map(([type, value]) => `${type === 'walk' ? '' : type + ' '}${value} ft.`)
            .join(', ');
    }

    getModifierString(score) {
        const modifier = Math.floor((score - 10) / 2);
        return modifier >= 0 ? `+${modifier}` : `${modifier}`;
    }

    addToCombat(npcId) {
        // This would integrate with the combat system
        alert('Add to combat functionality coming soon!');
    }

    editNPC(npcId) {
        // This would show an edit form
        alert('Edit NPC functionality coming soon!');
    }
}

// Create global instance
window.dmTools = new DMTools('dm-tools-container');