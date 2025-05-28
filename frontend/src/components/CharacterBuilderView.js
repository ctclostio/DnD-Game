import { apiService } from '../services/api.js';

export class CharacterBuilderView {
    constructor(container) {
        this.container = container;
        this.currentStep = 0;
        this.characterData = {
            name: '',
            race: '',
            subrace: '',
            class: '',
            background: '',
            alignment: 'True Neutral',
            abilityScoreMethod: 'standard_array',
            abilityScores: {},
            selectedSkills: []
        };
        this.options = {};
        this.isCustomMode = false;
    }

    async render() {
        // Load available options
        try {
            this.options = await apiService.get('/characters/options');
        } catch (error) {
            console.error('Failed to load character options:', error);
        }

        this.container.innerHTML = `
            <div class="character-builder">
                <div class="builder-header">
                    <h1>Create Your Character</h1>
                    <div class="mode-toggle">
                        <label class="toggle">
                            <input type="checkbox" id="customModeToggle">
                            <span class="slider"></span>
                        </label>
                        <span>Custom/Homebrew Mode ${this.options.aiEnabled ? '(AI Powered)' : '(Basic)'}</span>
                    </div>
                </div>
                
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${(this.currentStep + 1) * 14.28}%"></div>
                </div>
                
                <div class="builder-content" id="builderContent">
                    ${this.renderCurrentStep()}
                </div>
                
                <div class="builder-navigation">
                    <button class="btn-secondary" id="prevBtn" ${this.currentStep === 0 ? 'disabled' : ''}>
                        Previous
                    </button>
                    <button class="btn-primary" id="nextBtn">
                        ${this.currentStep === 6 ? 'Create Character' : 'Next'}
                    </button>
                </div>
            </div>
        `;

        this.attachEventListeners();
    }

    renderCurrentStep() {
        if (this.isCustomMode) {
            return this.renderCustomMode();
        }

        const steps = [
            this.renderBasicInfo,
            this.renderRaceSelection,
            this.renderClassSelection,
            this.renderAbilityScores,
            this.renderBackgroundSelection,
            this.renderSkillSelection,
            this.renderReview
        ];

        return steps[this.currentStep].call(this);
    }

    renderCustomMode() {
        return `
            <div class="custom-character-form">
                <h2>Create a Custom Character</h2>
                <p>Describe your character concept and let ${this.options.aiEnabled ? 'AI' : 'the system'} help bring them to life!</p>
                
                <div class="form-group">
                    <label for="characterName">Character Name</label>
                    <input type="text" id="characterName" value="${this.characterData.name}" 
                           placeholder="Enter character name" required>
                </div>
                
                <div class="form-group">
                    <label for="characterConcept">Character Concept</label>
                    <textarea id="characterConcept" rows="4" 
                              placeholder="Describe your character's concept, background, or unique traits..."
                              required></textarea>
                </div>
                
                <div class="form-group">
                    <label for="customRuleset">Ruleset (Optional)</label>
                    <input type="text" id="customRuleset" placeholder="D&D 5e (default) or custom ruleset">
                </div>
                
                <div class="form-row">
                    <div class="form-group">
                        <label for="customRace">Preferred Race (Optional)</label>
                        <input type="text" id="customRace" placeholder="Any race or custom">
                    </div>
                    
                    <div class="form-group">
                        <label for="customClass">Preferred Class (Optional)</label>
                        <input type="text" id="customClass" placeholder="Any class or custom">
                    </div>
                </div>
                
                <div class="ai-disclaimer">
                    ${this.options.aiEnabled ? 
                        '<p class="info">✨ AI will generate a unique character based on your concept!</p>' :
                        '<p class="warning">⚠️ AI is not configured. A basic custom character will be created.</p>'
                    }
                </div>
            </div>
        `;
    }

    renderBasicInfo() {
        return `
            <div class="step-content">
                <h2>Basic Information</h2>
                
                <div class="form-group">
                    <label for="characterName">Character Name</label>
                    <input type="text" id="characterName" value="${this.characterData.name}" 
                           placeholder="Enter character name" required>
                </div>
                
                <div class="form-group">
                    <label for="alignment">Alignment</label>
                    <select id="alignment" value="${this.characterData.alignment}">
                        <option value="Lawful Good">Lawful Good</option>
                        <option value="Neutral Good">Neutral Good</option>
                        <option value="Chaotic Good">Chaotic Good</option>
                        <option value="Lawful Neutral">Lawful Neutral</option>
                        <option value="True Neutral" selected>True Neutral</option>
                        <option value="Chaotic Neutral">Chaotic Neutral</option>
                        <option value="Lawful Evil">Lawful Evil</option>
                        <option value="Neutral Evil">Neutral Evil</option>
                        <option value="Chaotic Evil">Chaotic Evil</option>
                    </select>
                </div>
            </div>
        `;
    }

    renderRaceSelection() {
        return `
            <div class="step-content">
                <h2>Choose Your Race</h2>
                <div class="selection-grid">
                    ${this.options.races.map(race => `
                        <div class="selection-card ${this.characterData.race === race ? 'selected' : ''}" 
                             data-race="${race}">
                            <h3>${this.formatName(race)}</h3>
                            <p class="race-description">Click to select</p>
                        </div>
                    `).join('')}
                </div>
                
                ${this.renderSubraceSelection()}
            </div>
        `;
    }

    renderSubraceSelection() {
        // This would be populated based on selected race
        if (!this.characterData.race) return '';
        
        // For now, just return empty - would load subraces from race data
        return `<div id="subraceSelection"></div>`;
    }

    renderClassSelection() {
        return `
            <div class="step-content">
                <h2>Choose Your Class</h2>
                <div class="selection-grid">
                    ${this.options.classes.map(cls => `
                        <div class="selection-card ${this.characterData.class === cls ? 'selected' : ''}" 
                             data-class="${cls}">
                            <h3>${this.formatName(cls)}</h3>
                            <p class="class-description">Click to select</p>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    renderAbilityScores() {
        return `
            <div class="step-content">
                <h2>Determine Ability Scores</h2>
                
                <div class="ability-method-selector">
                    <label>Method:</label>
                    <select id="abilityMethod" value="${this.characterData.abilityScoreMethod}">
                        <option value="standard_array">Standard Array (15,14,13,12,10,8)</option>
                        <option value="point_buy">Point Buy (27 points)</option>
                        <option value="roll_4d6">Roll 4d6 Drop Lowest</option>
                        <option value="custom">Custom Entry</option>
                    </select>
                    ${this.characterData.abilityScoreMethod === 'roll_4d6' ? 
                        '<button class="btn-small" id="rollDiceBtn">Roll Dice</button>' : ''}
                </div>
                
                <div class="ability-scores-grid">
                    ${['strength', 'dexterity', 'constitution', 'intelligence', 'wisdom', 'charisma'].map(ability => `
                        <div class="ability-score-item">
                            <label>${this.formatName(ability)}</label>
                            <input type="number" id="${ability}Score" 
                                   value="${this.characterData.abilityScores[ability] || 10}"
                                   min="3" max="20">
                            <span class="modifier">+${Math.floor((this.characterData.abilityScores[ability] - 10) / 2) || 0}</span>
                        </div>
                    `).join('')}
                </div>
                
                ${this.renderAbilityMethodInfo()}
            </div>
        `;
    }

    renderAbilityMethodInfo() {
        const methods = {
            'standard_array': 'Assign these values to your abilities: 15, 14, 13, 12, 10, 8',
            'point_buy': 'You have 27 points to spend. Each ability starts at 8.',
            'roll_4d6': 'Roll 4d6 and drop the lowest die for each ability.',
            'custom': 'Enter your ability scores directly.'
        };
        
        return `<p class="method-info">${methods[this.characterData.abilityScoreMethod]}</p>`;
    }

    renderBackgroundSelection() {
        return `
            <div class="step-content">
                <h2>Choose Your Background</h2>
                <div class="selection-grid">
                    ${this.options.backgrounds.map(bg => `
                        <div class="selection-card ${this.characterData.background === bg ? 'selected' : ''}" 
                             data-background="${bg}">
                            <h3>${this.formatName(bg)}</h3>
                            <p class="background-description">Click to select</p>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    renderSkillSelection() {
        // This would be populated based on class and background
        return `
            <div class="step-content">
                <h2>Choose Your Skills</h2>
                <p>Select skills based on your class and background proficiencies.</p>
                <div class="skills-list">
                    <!-- Skills would be populated here based on class/background -->
                </div>
            </div>
        `;
    }

    renderReview() {
        return `
            <div class="step-content">
                <h2>Review Your Character</h2>
                <div class="character-summary">
                    <h3>${this.characterData.name || 'Unnamed Character'}</h3>
                    <p>${this.formatName(this.characterData.race)} ${this.formatName(this.characterData.class)}</p>
                    <p>Background: ${this.formatName(this.characterData.background)}</p>
                    <p>Alignment: ${this.characterData.alignment}</p>
                    
                    <div class="ability-summary">
                        <h4>Ability Scores</h4>
                        ${Object.entries(this.characterData.abilityScores).map(([ability, score]) => `
                            <div class="ability-item">
                                <span>${ability.toUpperCase().substring(0, 3)}</span>
                                <span>${score}</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            </div>
        `;
    }

    attachEventListeners() {
        // Mode toggle
        const modeToggle = this.container.querySelector('#customModeToggle');
        if (modeToggle) {
            modeToggle.addEventListener('change', (e) => {
                this.isCustomMode = e.target.checked;
                this.currentStep = 0;
                this.updateContent();
            });
        }

        // Navigation buttons
        const prevBtn = this.container.querySelector('#prevBtn');
        const nextBtn = this.container.querySelector('#nextBtn');

        prevBtn.addEventListener('click', () => this.previousStep());
        nextBtn.addEventListener('click', () => this.nextStep());

        // Form inputs
        this.attachFormListeners();

        // Selection cards
        this.attachSelectionListeners();
    }

    attachFormListeners() {
        // Character name
        const nameInput = this.container.querySelector('#characterName');
        if (nameInput) {
            nameInput.addEventListener('input', (e) => {
                this.characterData.name = e.target.value;
            });
        }

        // Alignment
        const alignmentSelect = this.container.querySelector('#alignment');
        if (alignmentSelect) {
            alignmentSelect.addEventListener('change', (e) => {
                this.characterData.alignment = e.target.value;
            });
        }

        // Ability scores
        const abilities = ['strength', 'dexterity', 'constitution', 'intelligence', 'wisdom', 'charisma'];
        abilities.forEach(ability => {
            const input = this.container.querySelector(`#${ability}Score`);
            if (input) {
                input.addEventListener('input', (e) => {
                    this.characterData.abilityScores[ability] = parseInt(e.target.value) || 10;
                    this.updateAbilityModifier(ability);
                });
            }
        });

        // Ability method
        const methodSelect = this.container.querySelector('#abilityMethod');
        if (methodSelect) {
            methodSelect.addEventListener('change', (e) => {
                this.characterData.abilityScoreMethod = e.target.value;
                this.updateContent();
            });
        }

        // Roll dice button
        const rollBtn = this.container.querySelector('#rollDiceBtn');
        if (rollBtn) {
            rollBtn.addEventListener('click', () => this.rollAbilityScores());
        }
    }

    attachSelectionListeners() {
        // Race selection
        this.container.querySelectorAll('[data-race]').forEach(card => {
            card.addEventListener('click', () => {
                this.characterData.race = card.dataset.race;
                this.updateSelectionCards('[data-race]', card);
            });
        });

        // Class selection
        this.container.querySelectorAll('[data-class]').forEach(card => {
            card.addEventListener('click', () => {
                this.characterData.class = card.dataset.class;
                this.updateSelectionCards('[data-class]', card);
            });
        });

        // Background selection
        this.container.querySelectorAll('[data-background]').forEach(card => {
            card.addEventListener('click', () => {
                this.characterData.background = card.dataset.background;
                this.updateSelectionCards('[data-background]', card);
            });
        });
    }

    updateSelectionCards(selector, selectedCard) {
        this.container.querySelectorAll(selector).forEach(card => {
            card.classList.remove('selected');
        });
        selectedCard.classList.add('selected');
    }

    updateAbilityModifier(ability) {
        const score = this.characterData.abilityScores[ability];
        const modifier = Math.floor((score - 10) / 2);
        const modifierSpan = this.container.querySelector(`#${ability}Score`).nextElementSibling;
        if (modifierSpan) {
            modifierSpan.textContent = modifier >= 0 ? `+${modifier}` : `${modifier}`;
        }
    }

    async rollAbilityScores() {
        try {
            const response = await apiService.post('/characters/roll-abilities', {
                method: this.characterData.abilityScoreMethod
            });

            this.characterData.abilityScores = response.scores;
            this.updateContent();
        } catch (error) {
            console.error('Failed to roll ability scores:', error);
        }
    }

    updateContent() {
        const content = this.container.querySelector('#builderContent');
        content.innerHTML = this.renderCurrentStep();
        this.attachFormListeners();
        this.attachSelectionListeners();
        this.updateProgressBar();
    }

    updateProgressBar() {
        const progressFill = this.container.querySelector('.progress-fill');
        const progress = this.isCustomMode ? 100 : ((this.currentStep + 1) / 7) * 100;
        progressFill.style.width = `${progress}%`;
    }

    previousStep() {
        if (this.currentStep > 0) {
            this.currentStep--;
            this.updateContent();
            this.updateNavigationButtons();
        }
    }

    async nextStep() {
        if (this.isCustomMode) {
            await this.createCustomCharacter();
        } else if (this.currentStep < 6) {
            this.currentStep++;
            this.updateContent();
            this.updateNavigationButtons();
        } else {
            await this.createCharacter();
        }
    }

    updateNavigationButtons() {
        const prevBtn = this.container.querySelector('#prevBtn');
        const nextBtn = this.container.querySelector('#nextBtn');

        prevBtn.disabled = this.currentStep === 0;
        nextBtn.textContent = this.currentStep === 6 ? 'Create Character' : 'Next';
    }

    async createCharacter() {
        try {
            const response = await apiService.post('/characters/create', this.characterData);
            console.log('Character created:', response);
            // Redirect to character sheet or show success message
            window.location.hash = '#/characters/' + response.id;
        } catch (error) {
            console.error('Failed to create character:', error);
            alert('Failed to create character. Please try again.');
        }
    }

    async createCustomCharacter() {
        const nameInput = this.container.querySelector('#characterName');
        const conceptInput = this.container.querySelector('#characterConcept');
        const rulesetInput = this.container.querySelector('#customRuleset');
        const raceInput = this.container.querySelector('#customRace');
        const classInput = this.container.querySelector('#customClass');

        const customData = {
            name: nameInput.value,
            concept: conceptInput.value,
            ruleset: rulesetInput.value || 'D&D 5e',
            race: raceInput.value,
            class: classInput.value,
            level: 1
        };

        try {
            const response = await apiService.post('/characters/create-custom', customData);
            console.log('Custom character created:', response);
            window.location.hash = '#/characters/' + response.id;
        } catch (error) {
            console.error('Failed to create custom character:', error);
            alert('Failed to create custom character. Please try again.');
        }
    }

    formatName(name) {
        return name.split('-').map(word => 
            word.charAt(0).toUpperCase() + word.slice(1)
        ).join(' ');
    }
}