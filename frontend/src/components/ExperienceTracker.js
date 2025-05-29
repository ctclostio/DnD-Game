import { ApiService } from '../services/api.js';

const api = new ApiService();

export class ExperienceTracker {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.character = null;
        this.xpThresholds = {
            1: 300,
            2: 900,
            3: 2700,
            4: 6500,
            5: 14000,
            6: 23000,
            7: 34000,
            8: 48000,
            9: 64000,
            10: 85000,
            11: 100000,
            12: 120000,
            13: 140000,
            14: 165000,
            15: 195000,
            16: 225000,
            17: 265000,
            18: 305000,
            19: 355000,
            20: 999999
        };
    }

    setCharacter(character) {
        this.character = character;
        this.render();
    }

    render() {
        if (!this.character) {
            this.container.innerHTML = '<p>No character selected</p>';
            return;
        }

        const xpForNext = this.xpThresholds[this.character.level] || 999999;
        const xpProgress = this.character.experiencePoints;
        const xpNeeded = xpForNext - xpProgress;
        const progressPercent = Math.min((xpProgress / xpForNext) * 100, 100);

        this.container.innerHTML = `
            <div class="experience-tracker">
                <div class="level-info">
                    <h3>Level ${this.character.level}</h3>
                    <p class="character-class">${this.character.class}</p>
                </div>
                
                <div class="xp-progress">
                    <div class="xp-bar">
                        <div class="xp-fill" style="width: ${progressPercent}%"></div>
                        <div class="xp-text">
                            ${xpProgress.toLocaleString()} / ${xpForNext.toLocaleString()} XP
                        </div>
                    </div>
                    <p class="xp-needed">${xpNeeded.toLocaleString()} XP to level ${this.character.level + 1}</p>
                </div>

                <div class="add-xp-section">
                    <h4>Award Experience</h4>
                    <div class="xp-input-group">
                        <input type="number" id="xp-amount" placeholder="XP Amount" min="1">
                        <button class="btn btn-primary" onclick="experienceTracker.addExperience()">
                            Add XP
                        </button>
                    </div>
                    <div class="quick-xp-buttons">
                        <button class="btn btn-sm" onclick="experienceTracker.quickAddXP(50)">+50</button>
                        <button class="btn btn-sm" onclick="experienceTracker.quickAddXP(100)">+100</button>
                        <button class="btn btn-sm" onclick="experienceTracker.quickAddXP(250)">+250</button>
                        <button class="btn btn-sm" onclick="experienceTracker.quickAddXP(500)">+500</button>
                        <button class="btn btn-sm" onclick="experienceTracker.quickAddXP(1000)">+1000</button>
                    </div>
                </div>

                <div class="level-features">
                    <h4>Current Features</h4>
                    <ul class="features-list">
                        <li>Proficiency Bonus: +${this.character.proficiencyBonus}</li>
                        <li>Hit Points: ${this.character.hitPoints}/${this.character.maxHitPoints}</li>
                        ${this.character.spells?.spellcastingAbility ? `
                            <li>Spell Save DC: ${this.character.spells.spellSaveDC}</li>
                            <li>Spell Attack Bonus: +${this.character.spells.spellAttackBonus}</li>
                        ` : ''}
                    </ul>
                </div>

                ${this.getNextLevelPreview()}
            </div>
        `;
    }

    getNextLevelPreview() {
        if (this.character.level >= 20) {
            return '<p class="max-level">Maximum level reached!</p>';
        }

        const nextLevel = this.character.level + 1;
        const newProfBonus = Math.floor((nextLevel - 1) / 4) + 2;

        return `
            <div class="next-level-preview">
                <h4>At Level ${nextLevel}:</h4>
                <ul>
                    ${newProfBonus > this.character.proficiencyBonus ? 
                        `<li>Proficiency Bonus increases to +${newProfBonus}</li>` : ''}
                    <li>Hit Points increase (based on class Hit Die)</li>
                    ${this.character.spells?.spellcastingAbility ? 
                        '<li>Spell slots may increase</li>' : ''}
                    <li>New class features may be unlocked</li>
                </ul>
            </div>
        `;
    }

    quickAddXP(amount) {
        document.getElementById('xp-amount').value = amount;
        this.addExperience();
    }

    async addExperience() {
        const xpInput = document.getElementById('xp-amount');
        const xpAmount = parseInt(xpInput.value);

        if (!xpAmount || xpAmount <= 0) {
            alert('Please enter a valid XP amount');
            return;
        }

        try {
            const response = await api.post(`/characters/${this.character.id}/add-experience`, {
                experience: xpAmount
            });

            const oldLevel = this.character.level;
            this.character = response.character;
            
            // Check if leveled up
            if (this.character.level > oldLevel) {
                this.showLevelUpNotification(oldLevel, this.character.level);
            }

            this.render();
            
            // Clear input
            xpInput.value = '';

            // Notify other components
            if (window.characterView) {
                window.characterView.updateCharacter(this.character);
            }
            if (window.spellSlotManager && this.character.spells) {
                window.spellSlotManager.setCharacter(this.character);
            }
        } catch (error) {
            console.error('Failed to add experience:', error);
            alert('Failed to add experience');
        }
    }

    showLevelUpNotification(oldLevel, newLevel) {
        const notification = document.createElement('div');
        notification.className = 'level-up-notification';
        notification.innerHTML = `
            <h2>Level Up!</h2>
            <p>Congratulations! You've reached level ${newLevel}!</p>
            <button onclick="this.parentElement.remove()">Dismiss</button>
        `;
        document.body.appendChild(notification);

        // Auto-remove after 5 seconds
        setTimeout(() => notification.remove(), 5000);
    }
}

// Create global instance
window.experienceTracker = new ExperienceTracker('experience-container');