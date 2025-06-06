import { createElement, clearElement, createInput, createSelect, appendChildren } from '../utils/dom.js';

export class CharacterView {
    constructor(container, api) {
        this.container = container;
        this.api = api;
        this.currentCharacter = null;
        this.render();
    }

    render() {
        clearElement(this.container);
        
        const characterView = createElement('div', { className: 'character-view' });
        
        const heading = createElement('h2', { textContent: 'Character Management' });
        
        const actions = createElement('div', { className: 'character-actions' });
        const createBtn = createElement('button', {
            id: 'create-character-btn',
            textContent: 'Create New Character',
            events: { click: () => this.showCharacterForm() }
        });
        const loadBtn = createElement('button', {
            id: 'load-character-btn',
            textContent: 'Load Character',
            events: { click: () => this.showCharacterList() }
        });
        
        actions.appendChild(createBtn);
        actions.appendChild(loadBtn);
        
        const content = createElement('div', { id: 'character-content' });
        
        characterView.appendChild(heading);
        characterView.appendChild(actions);
        characterView.appendChild(content);
        
        this.container.appendChild(characterView);
    }

    showCharacterForm(character = null) {
        const content = document.getElementById('character-content');
        clearElement(content);
        
        const form = createElement('form', {
            id: 'character-form',
            className: 'character-form',
            events: {
                submit: (e) => {
                    e.preventDefault();
                    this.saveCharacter(character?.id);
                }
            }
        });
        
        const heading = createElement('h3', {
            textContent: `${character ? 'Edit' : 'Create'} Character`
        });
        form.appendChild(heading);
        
        // Name field
        const nameGroup = createElement('div', { className: 'form-group' });
        const nameLabel = createElement('label', {
            textContent: 'Name',
            attributes: { for: 'char-name' }
        });
        const nameInput = createInput({
            type: 'text',
            id: 'char-name',
            value: character?.name || '',
            required: true
        });
        nameGroup.appendChild(nameLabel);
        nameGroup.appendChild(nameInput);
        form.appendChild(nameGroup);
        
        // Race field
        const raceGroup = createElement('div', { className: 'form-group' });
        const raceLabel = createElement('label', {
            textContent: 'Race',
            attributes: { for: 'char-race' }
        });
        const raceSelect = createSelect({
            id: 'char-race',
            required: true,
            options: [
                { value: '', text: 'Select Race' },
                { value: 'human', text: 'Human', selected: character?.race === 'human' },
                { value: 'elf', text: 'Elf', selected: character?.race === 'elf' },
                { value: 'dwarf', text: 'Dwarf', selected: character?.race === 'dwarf' },
                { value: 'halfling', text: 'Halfling', selected: character?.race === 'halfling' },
                { value: 'dragonborn', text: 'Dragonborn', selected: character?.race === 'dragonborn' },
                { value: 'tiefling', text: 'Tiefling', selected: character?.race === 'tiefling' }
            ]
        });
        raceGroup.appendChild(raceLabel);
        raceGroup.appendChild(raceSelect);
        form.appendChild(raceGroup);
        
        // Class field
        const classGroup = createElement('div', { className: 'form-group' });
        const classLabel = createElement('label', {
            textContent: 'Class',
            attributes: { for: 'char-class' }
        });
        const classSelect = createSelect({
            id: 'char-class',
            required: true,
            options: [
                { value: '', text: 'Select Class' },
                { value: 'fighter', text: 'Fighter', selected: character?.class === 'fighter' },
                { value: 'wizard', text: 'Wizard', selected: character?.class === 'wizard' },
                { value: 'cleric', text: 'Cleric', selected: character?.class === 'cleric' },
                { value: 'rogue', text: 'Rogue', selected: character?.class === 'rogue' },
                { value: 'ranger', text: 'Ranger', selected: character?.class === 'ranger' },
                { value: 'paladin', text: 'Paladin', selected: character?.class === 'paladin' }
            ]
        });
        classGroup.appendChild(classLabel);
        classGroup.appendChild(classSelect);
        form.appendChild(classGroup);
        
        // Attributes
        const attributesDiv = createElement('div', { className: 'attributes' });
        const attributes = [
            { name: 'STR', id: 'attr-str', value: character?.attributes?.strength || 10 },
            { name: 'DEX', id: 'attr-dex', value: character?.attributes?.dexterity || 10 },
            { name: 'CON', id: 'attr-con', value: character?.attributes?.constitution || 10 },
            { name: 'INT', id: 'attr-int', value: character?.attributes?.intelligence || 10 },
            { name: 'WIS', id: 'attr-wis', value: character?.attributes?.wisdom || 10 },
            { name: 'CHA', id: 'attr-cha', value: character?.attributes?.charisma || 10 }
        ];
        
        attributes.forEach(attr => {
            const box = createElement('div', { className: 'attribute-box' });
            const heading = createElement('h4', { textContent: attr.name });
            const input = createInput({
                type: 'number',
                id: attr.id,
                min: 3,
                max: 18,
                value: attr.value,
                required: true
            });
            box.appendChild(heading);
            box.appendChild(input);
            attributesDiv.appendChild(box);
        });
        form.appendChild(attributesDiv);
        
        // Buttons
        const submitBtn = createElement('button', {
            type: 'submit',
            textContent: `${character ? 'Update' : 'Create'} Character`
        });
        const cancelBtn = createElement('button', {
            type: 'button',
            id: 'cancel-btn',
            textContent: 'Cancel',
            events: {
                click: () => this.showCharacterSheet(this.currentCharacter)
            }
        });
        form.appendChild(submitBtn);
        form.appendChild(cancelBtn);
        
        content.appendChild(form);
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
            clearElement(content);
            
            if (characters.length === 0) {
                const message = createElement('p', {
                    textContent: 'No characters found. Create your first character!'
                });
                content.appendChild(message);
                return;
            }
            
            const heading = createElement('h3', { textContent: 'Your Characters' });
            content.appendChild(heading);
            
            const list = createElement('div', { className: 'character-list' });
            
            characters.forEach(char => {
                const card = createElement('div', {
                    className: 'character-card',
                    events: {
                        click: () => this.loadCharacter(char.id)
                    }
                });
                
                const name = createElement('h4', { textContent: char.name });
                const info = createElement('p', {
                    textContent: `Level ${char.level} ${char.race} ${char.class}`
                });
                
                card.appendChild(name);
                card.appendChild(info);
                list.appendChild(card);
            });
            
            content.appendChild(list);
        } catch (error) {
            console.error('Failed to load characters:', error);
            const content = document.getElementById('character-content');
            clearElement(content);
            const errorMsg = createElement('p', {
                textContent: 'Failed to load characters. Please try again.'
            });
            content.appendChild(errorMsg);
        }
    }

    async loadCharacter(characterId) {
        try {
            const character = await this.api.getCharacter(characterId);
            this.currentCharacter = character;
            this.showCharacterSheet(character);
        } catch (error) {
            console.error('Failed to load character:', error);
            alert('Failed to load character. Please try again.');
        }
    }

    showCharacterSheet(character) {
        if (!character) {
            this.render();
            return;
        }

        const content = document.getElementById('character-content');
        clearElement(content);
        
        const sheet = createElement('div', { className: 'character-sheet' });
        
        // Header
        const header = createElement('div', { className: 'character-header' });
        const name = createElement('h3', { textContent: character.name });
        const info = createElement('p', {
            textContent: `Level ${character.level} ${character.race} ${character.class}`
        });
        const editBtn = createElement('button', {
            textContent: 'Edit',
            events: {
                click: () => this.showCharacterForm(character)
            }
        });
        
        header.appendChild(name);
        header.appendChild(info);
        header.appendChild(editBtn);
        sheet.appendChild(header);
        
        // Stats
        const statsSection = createElement('div', { className: 'stats-section' });
        const statsHeading = createElement('h4', { textContent: 'Stats' });
        statsSection.appendChild(statsHeading);
        
        const stats = createElement('div', { className: 'stats' });
        
        // HP
        const hpStat = createElement('div', { className: 'stat' });
        const hpLabel = createElement('span', { textContent: 'HP:' });
        const hpValue = createElement('span', {
            textContent: `${character.currentHP || character.maxHP}/${character.maxHP}`
        });
        hpStat.appendChild(hpLabel);
        hpStat.appendChild(hpValue);
        stats.appendChild(hpStat);
        
        // AC
        const acStat = createElement('div', { className: 'stat' });
        const acLabel = createElement('span', { textContent: 'AC:' });
        const acValue = createElement('span', { textContent: character.armorClass });
        acStat.appendChild(acLabel);
        acStat.appendChild(acValue);
        stats.appendChild(acStat);
        
        // Speed
        const speedStat = createElement('div', { className: 'stat' });
        const speedLabel = createElement('span', { textContent: 'Speed:' });
        const speedValue = createElement('span', { textContent: character.speed });
        speedStat.appendChild(speedLabel);
        speedStat.appendChild(speedValue);
        stats.appendChild(speedStat);
        
        statsSection.appendChild(stats);
        sheet.appendChild(statsSection);
        
        // Attributes
        const attrSection = createElement('div', { className: 'attributes-section' });
        const attrHeading = createElement('h4', { textContent: 'Attributes' });
        attrSection.appendChild(attrHeading);
        
        const attrGrid = createElement('div', { className: 'attributes-grid' });
        
        Object.entries(character.attributes).forEach(([key, value]) => {
            const attrBox = createElement('div', { className: 'attribute-display' });
            const attrName = createElement('div', {
                className: 'attr-name',
                textContent: key.toUpperCase().slice(0, 3)
            });
            const attrValue = createElement('div', {
                className: 'attr-value',
                textContent: value
            });
            const modifier = Math.floor((value - 10) / 2);
            const attrMod = createElement('div', {
                className: 'attr-modifier',
                textContent: modifier >= 0 ? `+${modifier}` : `${modifier}`
            });
            
            attrBox.appendChild(attrName);
            attrBox.appendChild(attrValue);
            attrBox.appendChild(attrMod);
            attrGrid.appendChild(attrBox);
        });
        
        attrSection.appendChild(attrGrid);
        sheet.appendChild(attrSection);
        
        // Skills
        if (character.skills && character.skills.length > 0) {
            const skillsSection = createElement('div', { className: 'skills-section' });
            const skillsHeading = createElement('h4', { textContent: 'Skills' });
            skillsSection.appendChild(skillsHeading);
            
            const skillsList = createElement('div', { className: 'skills-list' });
            character.skills.forEach(skill => {
                const skillItem = createElement('div', {
                    className: 'skill-item',
                    textContent: skill
                });
                skillsList.appendChild(skillItem);
            });
            
            skillsSection.appendChild(skillsList);
            sheet.appendChild(skillsSection);
        }
        
        content.appendChild(sheet);
    }
}