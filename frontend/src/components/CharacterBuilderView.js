import { ApiService } from '../services/api.js';

const apiService = new ApiService();

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
                    <div class="mode-toggle" style="margin-top: 8px;">
                        <label class="toggle">
                            <input type="checkbox" id="guidedModeToggle" ${this.isGuidedMode ? 'checked' : ''}>
                            <span class="slider"></span>
                        </label>
                        <span>Guided Mode <span style="font-size:0.9em;color:#888;">(Beginner Friendly)</span></span>
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
                    ${this.isGuidedMode && !this.isCustomMode ? `<button class="btn-quickbuild" id="quickBuildBtn" style="margin-left:12px;">Quick Build</button>` : ''}
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
                    <div class="selection-card custom-race-card ${this.characterData.race === 'custom' ? 'selected' : ''}" 
                         data-race="custom">
                        <h3>🎨 Custom Race</h3>
                        <p class="race-description">Create your own unique race</p>
                    </div>
                </div>
                
                ${this.characterData.race === 'custom' ? this.renderCustomRaceForm() : this.renderSubraceSelection()}
            </div>
        `;
    }

    renderSubraceSelection() {
        // This would be populated based on selected race
        if (!this.characterData.race) return '';
        
        // For now, just return empty - would load subraces from race data
        return `<div id="subraceSelection"></div>`;
    }

    renderCustomRaceForm() {
        return `
            <div class="custom-race-form">
                <h3>Create Your Custom Race</h3>
                <p class="ai-notice">🤖 Our AI will help balance your custom race and generate appropriate traits!</p>
                
                <div class="form-group">
                    <label for="customRaceName">Race Name</label>
                    <input type="text" id="customRaceName" 
                           value="${this.characterData.customRaceData?.name || ''}"
                           placeholder="e.g., Celestial Tiefling, Frostborn Dwarf">
                </div>
                
                <div class="form-group">
                    <label for="customRaceDescription">Description</label>
                    <textarea id="customRaceDescription" rows="4"
                              placeholder="Describe your race's appearance, culture, origins, and any unique features...">${this.characterData.customRaceData?.description || ''}</textarea>
                </div>
                
                <div class="form-group">
                    <label for="customRaceTraits">Desired Traits (Optional)</label>
                    <textarea id="customRaceTraits" rows="3"
                              placeholder="List any specific abilities or traits you'd like (e.g., darkvision, natural armor, elemental resistance)">${this.characterData.customRaceData?.desiredTraits || ''}</textarea>
                </div>
                
                <div class="ai-generation-options">
                    <h4>Generation Style</h4>
                    <div class="radio-group">
                        <label>
                            <input type="radio" name="generationStyle" value="balanced" checked>
                            Balanced (Standard D&D power level)
                        </label>
                        <label>
                            <input type="radio" name="generationStyle" value="flavorful">
                            Flavorful (Focus on unique abilities)
                        </label>
                        <label>
                            <input type="radio" name="generationStyle" value="powerful">
                            Powerful (Slightly stronger, for experienced players)
                        </label>
                    </div>
                </div>
                
                <button class="btn-primary generate-race-btn" id="generateCustomRace">
                    🎲 Generate Custom Race
                </button>
                
                <div id="generatedRacePreview" class="race-preview" style="display: none;">
                    <!-- Generated race details will appear here -->
                </div>
            </div>
        `;
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
                    <div class="selection-card custom-class-card ${this.characterData.class === 'custom' ? 'selected' : ''}" 
                         data-class="custom">
                        <h3>🎯 Custom Class</h3>
                        <p class="class-description">Create your own unique class</p>
                    </div>
                </div>
                
                ${this.characterData.class === 'custom' ? this.renderCustomClassForm() : ''}
            </div>
        `;
    }

    renderCustomClassForm() {
        return `
            <div class="custom-class-form">
                <h3>Create Your Custom Class</h3>
                <p class="ai-notice">🤖 Our AI will design a balanced class with unique features and abilities!</p>
                
                <div class="form-group">
                    <label for="customClassName">Class Name</label>
                    <input type="text" id="customClassName" 
                           value="${this.characterData.customClassData?.name || ''}"
                           placeholder="e.g., Shadow Dancer, Spell Blade, Beast Master">
                </div>
                
                <div class="form-group">
                    <label for="customClassDescription">Description</label>
                    <textarea id="customClassDescription" rows="4"
                              placeholder="Describe your class's role, combat style, source of power, and unique features...">${this.characterData.customClassData?.description || ''}</textarea>
                </div>
                
                <div class="form-row">
                    <div class="form-group">
                        <label for="customClassRole">Primary Role</label>
                        <select id="customClassRole">
                            <option value="">Select a role...</option>
                            <option value="tank">Tank (Defender)</option>
                            <option value="damage">Damage Dealer</option>
                            <option value="healer">Healer/Support</option>
                            <option value="controller">Controller/Utility</option>
                            <option value="hybrid">Hybrid/Versatile</option>
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label for="customClassStyle">Class Style</label>
                        <select id="customClassStyle">
                            <option value="balanced">Balanced</option>
                            <option value="flavorful">Flavorful & Unique</option>
                            <option value="powerful">Powerful (Advanced)</option>
                        </select>
                    </div>
                </div>
                
                <div class="form-group">
                    <label for="customClassFeatures">Desired Features (Optional)</label>
                    <textarea id="customClassFeatures" rows="3"
                              placeholder="List any specific abilities or mechanics you'd like (e.g., pet companion, spell-sword hybrid, rage mechanics)">${this.characterData.customClassData?.features || ''}</textarea>
                </div>
                
                <button class="btn-primary generate-class-btn" id="generateCustomClass">
                    🎲 Generate Custom Class
                </button>
                
                <div id="generatedClassPreview" class="class-preview" style="display: none;">
                    <!-- Generated class details will appear here -->
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
        // Guided mode toggle
        const guidedToggle = this.container.querySelector('#guidedModeToggle');
        if (guidedToggle) {
            guidedToggle.addEventListener('change', (e) => {
                this.isGuidedMode = e.target.checked;
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
                this.updateContent(); // Re-render to show custom race form if needed
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

        // Generate custom race button
        const generateRaceBtn = this.container.querySelector('#generateCustomRace');
        if (generateRaceBtn) {
            generateRaceBtn.addEventListener('click', () => this.generateCustomRace());
        }

        // Generate custom class button
        const generateClassBtn = this.container.querySelector('#generateCustomClass');
        if (generateClassBtn) {
            generateClassBtn.addEventListener('click', () => this.generateCustomClass());
        }
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
        // Attach quick build button if in guided mode
        if (this.isGuidedMode && !this.isCustomMode) {
            const quickBuildBtn = this.container.querySelector('#quickBuildBtn');
            if (quickBuildBtn) {
                quickBuildBtn.addEventListener('click', () => this.quickBuild());
            }
        }
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
            // Prepare character data, handling custom race and class
            const characterData = {...this.characterData};
            if (this.characterData.race === 'custom' && this.characterData.customRaceId) {
                characterData.customRaceId = this.characterData.customRaceId;
                // Don't send regular race when using custom race
                delete characterData.race;
            }
            if (this.characterData.class === 'custom' && this.characterData.customClassId) {
                characterData.customClassId = this.characterData.customClassId;
                // Don't send regular class when using custom class
                delete characterData.class;
            }
            
            const response = await apiService.post('/characters/create', characterData);
            console.log('Character created:', response);
            // Redirect to character sheet or show success message
            window.location.hash = '#/characters/' + response.id;
        } catch (error) {
            console.error('Failed to create character:', error);
            alert('Failed to create character. Please try again.');
        }
    }

    async generateCustomClass() {
        const nameInput = this.container.querySelector('#customClassName');
        const descriptionInput = this.container.querySelector('#customClassDescription');
        const roleInput = this.container.querySelector('#customClassRole');
        const styleInput = this.container.querySelector('#customClassStyle');
        const featuresInput = this.container.querySelector('#customClassFeatures');

        if (!nameInput.value || !descriptionInput.value) {
            alert('Please provide a class name and description');
            return;
        }

        const generateBtn = this.container.querySelector('#generateCustomClass');
        const previewDiv = this.container.querySelector('#generatedClassPreview');
        
        generateBtn.disabled = true;
        generateBtn.textContent = '🎲 Generating...';

        try {
            const response = await apiService.post('/characters/custom-classes/generate', {
                name: nameInput.value,
                description: descriptionInput.value,
                role: roleInput.value,
                style: styleInput.value || 'balanced',
                features: featuresInput.value
            });

            // Store the generated class data
            this.characterData.customClassData = response;
            this.characterData.customClassId = response.id;

            // Display the preview
            previewDiv.innerHTML = `
                <h4>Generated Class: ${response.name}</h4>
                <div class="class-stats">
                    <p><strong>Hit Die:</strong> d${response.hitDie}</p>
                    <p><strong>Primary Ability:</strong> ${response.primaryAbility}</p>
                    <p><strong>Saving Throws:</strong> ${response.savingThrowProficiencies.join(', ')}</p>
                    <p><strong>Skills:</strong> Choose ${response.skillChoices} from ${response.skillProficiencies.join(', ')}</p>
                    <p><strong>Armor:</strong> ${response.armorProficiencies.join(', ')}</p>
                    <p><strong>Weapons:</strong> ${response.weaponProficiencies.join(', ')}</p>
                    ${response.toolProficiencies?.length ? `<p><strong>Tools:</strong> ${response.toolProficiencies.join(', ')}</p>` : ''}
                    
                    <h5>Class Features (Level 1-5):</h5>
                    <ul>
                        ${response.classFeatures
                            .filter(f => f.level <= 5)
                            .map(feature => `
                                <li>
                                    <strong>Level ${feature.level} - ${feature.name}:</strong> 
                                    ${feature.description}
                                    ${feature.usesPerRest ? ` (${feature.usesPerRest} per ${feature.restType} rest)` : ''}
                                </li>
                            `).join('')}
                    </ul>
                    
                    ${response.spellcastingAbility ? `
                        <h5>Spellcasting:</h5>
                        <p><strong>Ability:</strong> ${response.spellcastingAbility}</p>
                        <p><strong>Spell List:</strong> ${response.spellList.join(', ')}</p>
                        ${response.ritualCasting ? '<p><strong>Ritual Casting:</strong> Yes</p>' : ''}
                        <p><strong>Focus:</strong> ${response.spellcastingFocus}</p>
                    ` : ''}
                    
                    ${response.subclassName ? `
                        <p><strong>Subclass:</strong> Choose your ${response.subclassName} at level ${response.subclassLevel}</p>
                    ` : ''}
                </div>
                <p class="balance-score">Balance Score: ${response.balanceScore}/10</p>
                ${response.dmNotes ? `<p class="dm-notes"><strong>DM Notes:</strong> ${response.dmNotes}</p>` : ''}
            `;
            previewDiv.style.display = 'block';

        } catch (error) {
            console.error('Failed to generate custom class:', error);
            alert('Failed to generate custom class. Please try again.');
        } finally {
            generateBtn.disabled = false;
            generateBtn.textContent = '🎲 Generate Custom Class';
        }
    }

    async generateCustomRace() {
        const nameInput = this.container.querySelector('#customRaceName');
        const descriptionInput = this.container.querySelector('#customRaceDescription');
        const traitsInput = this.container.querySelector('#customRaceTraits');
        const generationStyle = this.container.querySelector('input[name="generationStyle"]:checked');

        if (!nameInput.value || !descriptionInput.value) {
            alert('Please provide a race name and description');
            return;
        }

        const generateBtn = this.container.querySelector('#generateCustomRace');
        const previewDiv = this.container.querySelector('#generatedRacePreview');
        
        generateBtn.disabled = true;
        generateBtn.textContent = '🎲 Generating...';

        try {
            const response = await apiService.post('/characters/custom-races/generate', {
                name: nameInput.value,
                description: descriptionInput.value,
                desiredTraits: traitsInput.value,
                style: generationStyle?.value || 'balanced'
            });

            // Store the generated race data
            this.characterData.customRaceData = response;
            this.characterData.customRaceId = response.id;

            // Display the preview
            previewDiv.innerHTML = `
                <h4>Generated Race: ${response.name}</h4>
                <div class="race-stats">
                    <p><strong>Size:</strong> ${response.size}</p>
                    <p><strong>Speed:</strong> ${response.speed} feet</p>
                    <p><strong>Ability Score Increases:</strong></p>
                    <ul>
                        ${Object.entries(response.abilityScoreIncreases).map(([ability, value]) => 
                            `<li>${this.formatName(ability)}: +${value}</li>`
                        ).join('')}
                    </ul>
                    <p><strong>Traits:</strong></p>
                    <ul>
                        ${response.racialTraits.map(trait => 
                            `<li><strong>${trait.name}:</strong> ${trait.description}</li>`
                        ).join('')}
                    </ul>
                    ${response.languages ? `<p><strong>Languages:</strong> ${response.languages.join(', ')}</p>` : ''}
                    ${response.proficiencies ? `<p><strong>Proficiencies:</strong> ${response.proficiencies.join(', ')}</p>` : ''}
                </div>
                <p class="balance-score">Balance Score: ${response.balanceScore}/10</p>
                ${response.dmNotes ? `<p class="dm-notes"><strong>DM Notes:</strong> ${response.dmNotes}</p>` : ''}
            `;
            previewDiv.style.display = 'block';

        } catch (error) {
            console.error('Failed to generate custom race:', error);
            alert('Failed to generate custom race. Please try again.');
        } finally {
            generateBtn.disabled = false;
            generateBtn.textContent = '🎲 Generate Custom Race';
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
    // Guided Mode: Quick Build for new players
    quickBuild() {
        // Example: pick a beginner-friendly race/class/background
        this.characterData.race = this.options.races?.includes('human') ? 'human' : this.options.races[0];
        this.characterData.class = this.options.classes?.includes('fighter') ? 'fighter' : this.options.classes[0];
        this.characterData.background = this.options.backgrounds?.includes('folk-hero') ? 'folk-hero' : this.options.backgrounds[0];
        this.characterData.abilityScoreMethod = 'standard_array';
        this.characterData.abilityScores = {
            strength: 15, dexterity: 14, constitution: 13, intelligence: 12, wisdom: 10, charisma: 8
        };
        this.currentStep = 6; // Jump to review
        this.updateContent();
    }
}