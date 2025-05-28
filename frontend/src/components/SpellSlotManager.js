import { api } from '../services/api.js';

export class SpellSlotManager {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.character = null;
    }

    setCharacter(character) {
        this.character = character;
        this.render();
    }

    render() {
        if (!this.character || !this.character.spells) {
            this.container.innerHTML = '<p>No spell data available</p>';
            return;
        }

        const { spells } = this.character;
        
        this.container.innerHTML = `
            <div class="spell-slot-manager">
                <h3>Spell Slots</h3>
                ${this.renderSpellSlots(spells.spellSlots)}
                <div class="rest-buttons">
                    <button class="btn btn-secondary" onclick="spellSlotManager.takeRest('short')">
                        Short Rest
                    </button>
                    <button class="btn btn-primary" onclick="spellSlotManager.takeRest('long')">
                        Long Rest
                    </button>
                </div>
                ${this.renderKnownSpells(spells.spellsKnown)}
            </div>
        `;
    }

    renderSpellSlots(spellSlots) {
        if (!spellSlots || spellSlots.length === 0) {
            return '<p>No spell slots available</p>';
        }

        return `
            <div class="spell-slots-grid">
                ${spellSlots.map(slot => `
                    <div class="spell-slot-level">
                        <h4>Level ${slot.level}</h4>
                        <div class="spell-slot-boxes">
                            ${this.renderSlotBoxes(slot)}
                        </div>
                        <p>${slot.remaining}/${slot.total}</p>
                    </div>
                `).join('')}
            </div>
        `;
    }

    renderSlotBoxes(slot) {
        const boxes = [];
        for (let i = 0; i < slot.total; i++) {
            const isUsed = i >= slot.remaining;
            boxes.push(`
                <div class="spell-slot-box ${isUsed ? 'used' : 'available'}" 
                     title="${isUsed ? 'Used' : 'Available'}">
                </div>
            `);
        }
        return boxes.join('');
    }

    renderKnownSpells(spellsKnown) {
        if (!spellsKnown || spellsKnown.length === 0) {
            return '';
        }

        const spellsByLevel = this.groupSpellsByLevel(spellsKnown);
        
        return `
            <div class="known-spells">
                <h3>Known Spells</h3>
                ${Object.entries(spellsByLevel).map(([level, spells]) => `
                    <div class="spell-level-group">
                        <h4>${level === '0' ? 'Cantrips' : `Level ${level}`}</h4>
                        <div class="spell-list">
                            ${spells.map(spell => `
                                <div class="spell-item">
                                    <span class="spell-name">${spell.name}</span>
                                    ${spell.level > 0 ? `
                                        <button class="btn btn-sm btn-cast" 
                                                onclick="spellSlotManager.castSpell(${spell.level})"
                                                ${!this.canCastSpell(spell.level) ? 'disabled' : ''}>
                                            Cast
                                        </button>
                                    ` : ''}
                                </div>
                            `).join('')}
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    groupSpellsByLevel(spells) {
        const grouped = {};
        spells.forEach(spell => {
            const level = spell.level.toString();
            if (!grouped[level]) {
                grouped[level] = [];
            }
            grouped[level].push(spell);
        });
        return grouped;
    }

    canCastSpell(level) {
        if (!this.character.spells.spellSlots) return false;
        const slot = this.character.spells.spellSlots.find(s => s.level === level);
        return slot && slot.remaining > 0;
    }

    async castSpell(level) {
        if (!this.canCastSpell(level)) {
            alert('No spell slots remaining for this level');
            return;
        }

        try {
            const response = await api.post(`/characters/${this.character.id}/cast-spell`, {
                spellLevel: level
            });
            
            this.character = response;
            this.render();
            
            // Notify other components
            if (window.characterView) {
                window.characterView.updateCharacter(this.character);
            }
        } catch (error) {
            console.error('Failed to cast spell:', error);
            alert('Failed to cast spell');
        }
    }

    async takeRest(restType) {
        try {
            const response = await api.post(`/characters/${this.character.id}/rest`, {
                restType: restType
            });
            
            this.character = response;
            this.render();
            
            // Notify other components
            if (window.characterView) {
                window.characterView.updateCharacter(this.character);
            }
            
            alert(`${restType === 'long' ? 'Long' : 'Short'} rest completed!`);
        } catch (error) {
            console.error('Failed to rest:', error);
            alert('Failed to complete rest');
        }
    }
}

// Create global instance
window.spellSlotManager = new SpellSlotManager('spell-slot-container');