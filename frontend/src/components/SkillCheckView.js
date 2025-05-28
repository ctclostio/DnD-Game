import { ApiService } from '../services/api.js';

export class SkillCheckView {
    constructor() {
        this.api = new ApiService();
        this.character = null;
        this.checks = null;
        this.selectedCheck = null;
        this.advantage = false;
        this.disadvantage = false;
        this.customDC = null;
    }

    async init(character) {
        this.character = character;
        await this.loadCharacterChecks();
        this.render();
    }

    async loadCharacterChecks() {
        try {
            const response = await this.api.get(`/characters/${this.character.id}/checks`);
            this.checks = response;
        } catch (error) {
            console.error('Error loading character checks:', error);
            this.checks = this.getDefaultChecks();
        }
    }

    getDefaultChecks() {
        // Fallback if API fails
        const abilities = ['strength', 'dexterity', 'constitution', 'intelligence', 'wisdom', 'charisma'];
        const character = this.character;
        
        return {
            savingThrows: abilities.map(ability => ({
                name: ability,
                modifier: Math.floor((character[ability] - 10) / 2),
                proficient: false
            })),
            skills: [
                { name: 'acrobatics', ability: 'dexterity', modifier: Math.floor((character.dexterity - 10) / 2) },
                { name: 'animal handling', ability: 'wisdom', modifier: Math.floor((character.wisdom - 10) / 2) },
                { name: 'arcana', ability: 'intelligence', modifier: Math.floor((character.intelligence - 10) / 2) },
                { name: 'athletics', ability: 'strength', modifier: Math.floor((character.strength - 10) / 2) },
                { name: 'deception', ability: 'charisma', modifier: Math.floor((character.charisma - 10) / 2) },
                { name: 'history', ability: 'intelligence', modifier: Math.floor((character.intelligence - 10) / 2) },
                { name: 'insight', ability: 'wisdom', modifier: Math.floor((character.wisdom - 10) / 2) },
                { name: 'intimidation', ability: 'charisma', modifier: Math.floor((character.charisma - 10) / 2) },
                { name: 'investigation', ability: 'intelligence', modifier: Math.floor((character.intelligence - 10) / 2) },
                { name: 'medicine', ability: 'wisdom', modifier: Math.floor((character.wisdom - 10) / 2) },
                { name: 'nature', ability: 'intelligence', modifier: Math.floor((character.intelligence - 10) / 2) },
                { name: 'perception', ability: 'wisdom', modifier: Math.floor((character.wisdom - 10) / 2) },
                { name: 'performance', ability: 'charisma', modifier: Math.floor((character.charisma - 10) / 2) },
                { name: 'persuasion', ability: 'charisma', modifier: Math.floor((character.charisma - 10) / 2) },
                { name: 'religion', ability: 'intelligence', modifier: Math.floor((character.intelligence - 10) / 2) },
                { name: 'sleight of hand', ability: 'dexterity', modifier: Math.floor((character.dexterity - 10) / 2) },
                { name: 'stealth', ability: 'dexterity', modifier: Math.floor((character.dexterity - 10) / 2) },
                { name: 'survival', ability: 'wisdom', modifier: Math.floor((character.wisdom - 10) / 2) }
            ],
            abilities: abilities.map(ability => ({
                name: ability,
                modifier: Math.floor((character[ability] - 10) / 2)
            }))
        };
    }

    render() {
        const container = document.getElementById('skill-check-view');
        if (!container) return;

        container.innerHTML = `
            <div class="skill-check-container">
                <h2>Skill Checks & Saving Throws</h2>
                
                <div class="check-options">
                    <div class="roll-modifiers">
                        <label class="toggle-option">
                            <input type="checkbox" id="advantage-toggle" ${this.advantage ? 'checked' : ''}>
                            <span>Advantage</span>
                        </label>
                        <label class="toggle-option">
                            <input type="checkbox" id="disadvantage-toggle" ${this.disadvantage ? 'checked' : ''}>
                            <span>Disadvantage</span>
                        </label>
                    </div>
                    
                    <div class="dc-input">
                        <label>DC (Optional):</label>
                        <input type="number" id="dc-input" min="1" max="30" placeholder="10" value="${this.customDC || ''}">
                    </div>
                </div>

                <div class="checks-grid">
                    <div class="check-section">
                        <h3>Saving Throws</h3>
                        <div class="check-list">
                            ${this.renderSavingThrows()}
                        </div>
                    </div>

                    <div class="check-section">
                        <h3>Skills</h3>
                        <div class="check-list skills-list">
                            ${this.renderSkills()}
                        </div>
                    </div>

                    <div class="check-section">
                        <h3>Ability Checks</h3>
                        <div class="check-list">
                            ${this.renderAbilityChecks()}
                        </div>
                    </div>
                </div>

                <div id="roll-result" class="roll-result-panel" style="display: none;">
                    <!-- Roll results will appear here -->
                </div>
            </div>
        `;

        this.attachEventListeners();
    }

    renderSavingThrows() {
        return this.checks.savingThrows.map(save => `
            <div class="check-item ${save.proficient ? 'proficient' : ''}" 
                 data-check-type="save" 
                 data-ability="${save.name}"
                 data-modifier="${save.modifier}">
                <span class="check-name">${this.capitalizeFirst(save.name)} Save</span>
                <span class="check-modifier">${save.modifier >= 0 ? '+' : ''}${save.modifier}</span>
                ${save.proficient ? '<span class="proficiency-marker">●</span>' : ''}
            </div>
        `).join('');
    }

    renderSkills() {
        return this.checks.skills.map(skill => `
            <div class="check-item ${skill.proficient ? 'proficient' : ''}" 
                 data-check-type="skill" 
                 data-skill="${skill.name}"
                 data-ability="${skill.ability}"
                 data-modifier="${skill.modifier}">
                <span class="check-name">${this.capitalizeFirst(skill.name)}</span>
                <span class="ability-tag">(${skill.ability.substring(0, 3).toUpperCase()})</span>
                <span class="check-modifier">${skill.modifier >= 0 ? '+' : ''}${skill.modifier}</span>
                ${skill.proficient ? '<span class="proficiency-marker">●</span>' : ''}
            </div>
        `).join('');
    }

    renderAbilityChecks() {
        return this.checks.abilities.map(ability => `
            <div class="check-item" 
                 data-check-type="ability" 
                 data-ability="${ability.name}"
                 data-modifier="${ability.modifier}">
                <span class="check-name">${this.capitalizeFirst(ability.name)}</span>
                <span class="check-modifier">${ability.modifier >= 0 ? '+' : ''}${ability.modifier}</span>
            </div>
        `).join('');
    }

    attachEventListeners() {
        // Toggle advantage/disadvantage
        document.getElementById('advantage-toggle')?.addEventListener('change', (e) => {
            this.advantage = e.target.checked;
            if (this.advantage && this.disadvantage) {
                this.disadvantage = false;
                document.getElementById('disadvantage-toggle').checked = false;
            }
        });

        document.getElementById('disadvantage-toggle')?.addEventListener('change', (e) => {
            this.disadvantage = e.target.checked;
            if (this.disadvantage && this.advantage) {
                this.advantage = false;
                document.getElementById('advantage-toggle').checked = false;
            }
        });

        // DC input
        document.getElementById('dc-input')?.addEventListener('change', (e) => {
            this.customDC = e.target.value ? parseInt(e.target.value) : null;
        });

        // Check items
        document.querySelectorAll('.check-item').forEach(item => {
            item.addEventListener('click', () => this.performCheck(item));
        });
    }

    async performCheck(checkElement) {
        const checkType = checkElement.dataset.checkType;
        const ability = checkElement.dataset.ability;
        const skill = checkElement.dataset.skill;
        const modifier = parseInt(checkElement.dataset.modifier);

        try {
            const response = await this.api.post('/skill-check', {
                characterId: this.character.id,
                checkType: checkType,
                skill: skill,
                ability: ability,
                modifier: modifier,
                advantage: this.advantage,
                disadvantage: this.disadvantage,
                dc: this.customDC
            });

            this.displayResult(response, checkElement);
        } catch (error) {
            console.error('Error performing check:', error);
            this.displayError('Failed to perform check');
        }
    }

    displayResult(result, checkElement) {
        const resultPanel = document.getElementById('roll-result');
        const checkName = checkElement.querySelector('.check-name').textContent;
        
        let resultClass = '';
        if (result.criticalSuccess) resultClass = 'critical-success';
        else if (result.criticalFailure) resultClass = 'critical-failure';
        else if (result.dc && result.success) resultClass = 'success';
        else if (result.dc && !result.success) resultClass = 'failure';

        resultPanel.innerHTML = `
            <div class="result-content ${resultClass}">
                <h4>${checkName}</h4>
                <div class="roll-details">
                    <div class="dice-result">
                        <span class="die-roll">${result.roll}</span>
                        ${result.advantage || result.disadvantage ? `
                            <span class="all-rolls">(${result.allRolls.join(', ')})</span>
                        ` : ''}
                    </div>
                    <div class="modifier-display">
                        ${result.modifier >= 0 ? '+' : ''}${result.modifier}
                    </div>
                    <div class="total-result">
                        = ${result.total}
                    </div>
                </div>
                ${result.dc ? `
                    <div class="dc-result">
                        DC ${result.dc}: <strong>${result.success ? 'Success!' : 'Failure'}</strong>
                    </div>
                ` : ''}
                ${result.criticalSuccess ? '<div class="critical-text">Critical Success!</div>' : ''}
                ${result.criticalFailure ? '<div class="critical-text">Critical Failure!</div>' : ''}
                ${result.advantage ? '<div class="roll-type">Advantage</div>' : ''}
                ${result.disadvantage ? '<div class="roll-type">Disadvantage</div>' : ''}
            </div>
        `;

        resultPanel.style.display = 'block';
        
        // Auto-hide after 10 seconds
        setTimeout(() => {
            resultPanel.style.display = 'none';
        }, 10000);
    }

    displayError(message) {
        const resultPanel = document.getElementById('roll-result');
        resultPanel.innerHTML = `
            <div class="result-content error">
                <p>${message}</p>
            </div>
        `;
        resultPanel.style.display = 'block';
    }

    capitalizeFirst(str) {
        return str.charAt(0).toUpperCase() + str.slice(1);
    }
}