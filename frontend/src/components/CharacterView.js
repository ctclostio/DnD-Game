export class CharacterView {
    constructor(container, api) {
        this.container = container;
        this.api = api;
        this.currentCharacter = null;
        this.render();
    }

    render() {
        this.container.innerHTML = `
            <div class="character-view">
                <h2>Character Management</h2>
                <div class="character-actions">
                    <button id="create-character-btn">Create New Character</button>
                    <button id="load-character-btn">Load Character</button>
                </div>
                <div id="character-content"></div>
            </div>
        `;

        this.setupEventListeners();
    }

    setupEventListeners() {
        document.getElementById('create-character-btn').addEventListener('click', () => {
            this.showCharacterForm();
        });

        document.getElementById('load-character-btn').addEventListener('click', () => {
            this.showCharacterList();
        });
    }

    showCharacterForm(character = null) {
        const content = document.getElementById('character-content');
        content.innerHTML = `
            <form id="character-form" class="character-form">
                <h3>${character ? 'Edit' : 'Create'} Character</h3>
                
                <div class="form-group">
                    <label for="char-name">Name</label>
                    <input type="text" id="char-name" value="${character?.name || ''}" required>
                </div>

                <div class="form-group">
                    <label for="char-race">Race</label>
                    <select id="char-race" required>
                        <option value="">Select Race</option>
                        <option value="human" ${character?.race === 'human' ? 'selected' : ''}>Human</option>
                        <option value="elf" ${character?.race === 'elf' ? 'selected' : ''}>Elf</option>
                        <option value="dwarf" ${character?.race === 'dwarf' ? 'selected' : ''}>Dwarf</option>
                        <option value="halfling" ${character?.race === 'halfling' ? 'selected' : ''}>Halfling</option>
                        <option value="dragonborn" ${character?.race === 'dragonborn' ? 'selected' : ''}>Dragonborn</option>
                        <option value="tiefling" ${character?.race === 'tiefling' ? 'selected' : ''}>Tiefling</option>
                    </select>
                </div>

                <div class="form-group">
                    <label for="char-class">Class</label>
                    <select id="char-class" required>
                        <option value="">Select Class</option>
                        <option value="fighter" ${character?.class === 'fighter' ? 'selected' : ''}>Fighter</option>
                        <option value="wizard" ${character?.class === 'wizard' ? 'selected' : ''}>Wizard</option>
                        <option value="cleric" ${character?.class === 'cleric' ? 'selected' : ''}>Cleric</option>
                        <option value="rogue" ${character?.class === 'rogue' ? 'selected' : ''}>Rogue</option>
                        <option value="ranger" ${character?.class === 'ranger' ? 'selected' : ''}>Ranger</option>
                        <option value="paladin" ${character?.class === 'paladin' ? 'selected' : ''}>Paladin</option>
                    </select>
                </div>

                <div class="attributes">
                    <div class="attribute-box">
                        <h4>STR</h4>
                        <input type="number" id="attr-str" min="3" max="18" value="${character?.attributes?.strength || 10}" required>
                    </div>
                    <div class="attribute-box">
                        <h4>DEX</h4>
                        <input type="number" id="attr-dex" min="3" max="18" value="${character?.attributes?.dexterity || 10}" required>
                    </div>
                    <div class="attribute-box">
                        <h4>CON</h4>
                        <input type="number" id="attr-con" min="3" max="18" value="${character?.attributes?.constitution || 10}" required>
                    </div>
                    <div class="attribute-box">
                        <h4>INT</h4>
                        <input type="number" id="attr-int" min="3" max="18" value="${character?.attributes?.intelligence || 10}" required>
                    </div>
                    <div class="attribute-box">
                        <h4>WIS</h4>
                        <input type="number" id="attr-wis" min="3" max="18" value="${character?.attributes?.wisdom || 10}" required>
                    </div>
                    <div class="attribute-box">
                        <h4>CHA</h4>
                        <input type="number" id="attr-cha" min="3" max="18" value="${character?.attributes?.charisma || 10}" required>
                    </div>
                </div>

                <button type="submit">${character ? 'Update' : 'Create'} Character</button>
                <button type="button" id="cancel-btn">Cancel</button>
            </form>
        `;

        document.getElementById('character-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.saveCharacter(character?.id);
        });

        document.getElementById('cancel-btn').addEventListener('click', () => {
            this.showCharacterSheet(this.currentCharacter);
        });
    }

    async saveCharacter(characterId) {
        const characterData = {
            name: document.getElementById('char-name').value,
            race: document.getElementById('char-race').value,
            class: document.getElementById('char-class').value,
            attributes: {
                strength: parseInt(document.getElementById('attr-str').value),
                dexterity: parseInt(document.getElementById('attr-dex').value),
                constitution: parseInt(document.getElementById('attr-con').value),
                intelligence: parseInt(document.getElementById('attr-int').value),
                wisdom: parseInt(document.getElementById('attr-wis').value),
                charisma: parseInt(document.getElementById('attr-cha').value),
            }
        };

        try {
            let character;
            if (characterId) {
                character = await this.api.updateCharacter(characterId, characterData);
            } else {
                character = await this.api.createCharacter(characterData);
            }
            this.currentCharacter = character;
            this.showCharacterSheet(character);
        } catch (error) {
            console.error('Failed to save character:', error);
            alert('Failed to save character. Please try again.');
        }
    }

    async showCharacterList() {
        try {
            const characters = await this.api.getCharacters();
            const content = document.getElementById('character-content');
            
            if (characters.length === 0) {
                content.innerHTML = '<p>No characters found. Create your first character!</p>';
                return;
            }

            content.innerHTML = `
                <div class="character-list">
                    <h3>Select a Character</h3>
                    <div class="character-cards">
                        ${characters.map(char => `
                            <div class="character-card" data-id="${char.id}">
                                <h4>${char.name}</h4>
                                <p>${char.race} ${char.class}</p>
                                <p>Level ${char.level}</p>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;

            document.querySelectorAll('.character-card').forEach(card => {
                card.addEventListener('click', async (e) => {
                    const characterId = e.currentTarget.dataset.id;
                    const character = await this.api.getCharacter(characterId);
                    this.currentCharacter = character;
                    this.showCharacterSheet(character);
                });
            });
        } catch (error) {
            console.error('Failed to load characters:', error);
            alert('Failed to load characters. Please try again.');
        }
    }

    showCharacterSheet(character) {
        if (!character) {
            document.getElementById('character-content').innerHTML = '<p>No character selected</p>';
            return;
        }

        const content = document.getElementById('character-content');
        content.innerHTML = `
            <div class="character-sheet">
                <div class="character-header">
                    <div>
                        <h3>${character.name}</h3>
                        <p>${character.race} ${character.class} - Level ${character.level}</p>
                    </div>
                    <div>
                        <button id="edit-character-btn">Edit</button>
                        <button id="skill-check-btn">Skill Checks</button>
                    </div>
                </div>

                <div class="character-stats">
                    <div class="stat-box">
                        <h4>Hit Points</h4>
                        <p>${character.hitPoints} / ${character.maxHitPoints}</p>
                    </div>
                    <div class="stat-box">
                        <h4>Armor Class</h4>
                        <p>${character.armorClass}</p>
                    </div>
                    <div class="stat-box">
                        <h4>Speed</h4>
                        <p>${character.speed} ft</p>
                    </div>
                </div>

                <div class="attributes">
                    <div class="attribute-box">
                        <h4>STR</h4>
                        <div class="attribute-value">${character.strength || 10}</div>
                        <div class="attribute-modifier">${this.getModifier(character.strength || 10) >= 0 ? '+' : ''}${this.getModifier(character.strength || 10)}</div>
                    </div>
                    <div class="attribute-box">
                        <h4>DEX</h4>
                        <div class="attribute-value">${character.dexterity || 10}</div>
                        <div class="attribute-modifier">${this.getModifier(character.dexterity || 10) >= 0 ? '+' : ''}${this.getModifier(character.dexterity || 10)}</div>
                    </div>
                    <div class="attribute-box">
                        <h4>CON</h4>
                        <div class="attribute-value">${character.constitution || 10}</div>
                        <div class="attribute-modifier">${this.getModifier(character.constitution || 10) >= 0 ? '+' : ''}${this.getModifier(character.constitution || 10)}</div>
                    </div>
                    <div class="attribute-box">
                        <h4>INT</h4>
                        <div class="attribute-value">${character.intelligence || 10}</div>
                        <div class="attribute-modifier">${this.getModifier(character.intelligence || 10) >= 0 ? '+' : ''}${this.getModifier(character.intelligence || 10)}</div>
                    </div>
                    <div class="attribute-box">
                        <h4>WIS</h4>
                        <div class="attribute-value">${character.wisdom || 10}</div>
                        <div class="attribute-modifier">${this.getModifier(character.wisdom || 10) >= 0 ? '+' : ''}${this.getModifier(character.wisdom || 10)}</div>
                    </div>
                    <div class="attribute-box">
                        <h4>CHA</h4>
                        <div class="attribute-value">${character.charisma || 10}</div>
                        <div class="attribute-modifier">${this.getModifier(character.charisma || 10) >= 0 ? '+' : ''}${this.getModifier(character.charisma || 10)}</div>
                    </div>
                </div>

                <div class="skills-section">
                    <h4>Skills</h4>
                    <div class="skills-list">
                        ${character.skills?.map(skill => `
                            <div class="skill-item">
                                <span>${skill.name}</span>
                                <span>${skill.modifier >= 0 ? '+' : ''}${skill.modifier}</span>
                            </div>
                        `).join('') || '<p>No skills recorded</p>'}
                    </div>
                </div>

                <div id="experience-container"></div>
                
                <div id="spell-slot-container"></div>
                
                <div id="skill-check-view" style="display: none;"></div>
            </div>
        `;

        document.getElementById('edit-character-btn').addEventListener('click', () => {
            this.showCharacterForm(character);
        });
        
        document.getElementById('skill-check-btn').addEventListener('click', () => {
            // Toggle skill check view
            const skillCheckView = document.getElementById('skill-check-view');
            if (skillCheckView.style.display === 'none') {
                skillCheckView.style.display = 'block';
                // Initialize skill check view if available
                if (window.skillCheckView) {
                    window.skillCheckView.init(character);
                }
            } else {
                skillCheckView.style.display = 'none';
            }
        });

        // Initialize experience tracker
        if (window.experienceTracker) {
            window.experienceTracker.setCharacter(character);
        }

        // Initialize spell slot manager if character has spell data
        if (window.spellSlotManager && character.spells) {
            window.spellSlotManager.setCharacter(character);
        }
    }

    updateCharacter(character) {
        this.currentCharacter = character;
        this.showCharacterSheet(character);
    }

    getModifier(abilityScore) {
        return Math.floor((abilityScore - 10) / 2);
    }
}