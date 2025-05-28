export class DiceRollerView {
    constructor(container, api) {
        this.container = container;
        this.api = api;
        this.rollHistory = [];
        this.render();
    }

    render() {
        this.container.innerHTML = `
            <div class="dice-roller">
                <h2>Dice Roller</h2>
                
                <div class="dice-options">
                    <button class="dice-button" data-dice="1d4">d4</button>
                    <button class="dice-button" data-dice="1d6">d6</button>
                    <button class="dice-button" data-dice="1d8">d8</button>
                    <button class="dice-button" data-dice="1d10">d10</button>
                    <button class="dice-button" data-dice="1d12">d12</button>
                    <button class="dice-button" data-dice="1d20">d20</button>
                    <button class="dice-button" data-dice="1d100">d100</button>
                    <button class="dice-button" data-dice="2d6">2d6</button>
                </div>

                <div class="custom-roll">
                    <h3>Custom Roll</h3>
                    <div class="form-group">
                        <input type="text" id="custom-dice" placeholder="e.g., 3d6+2" />
                        <button id="custom-roll-btn">Roll</button>
                    </div>
                    <div class="form-group">
                        <label for="roll-purpose">Purpose (optional)</label>
                        <input type="text" id="roll-purpose" placeholder="e.g., Attack roll, Damage" />
                    </div>
                </div>

                <div class="modifiers">
                    <h3>Quick Modifiers</h3>
                    <div class="modifier-buttons">
                        <button class="modifier-btn" data-mod="-5">-5</button>
                        <button class="modifier-btn" data-mod="-2">-2</button>
                        <button class="modifier-btn" data-mod="-1">-1</button>
                        <button class="modifier-btn" data-mod="+1">+1</button>
                        <button class="modifier-btn" data-mod="+2">+2</button>
                        <button class="modifier-btn" data-mod="+5">+5</button>
                    </div>
                </div>

                <div id="roll-result"></div>

                <div class="roll-history">
                    <h3>Roll History</h3>
                    <div id="history-list"></div>
                </div>
            </div>
        `;

        this.setupEventListeners();
    }

    setupEventListeners() {
        // Standard dice buttons
        document.querySelectorAll('.dice-button').forEach(button => {
            button.addEventListener('click', (e) => {
                const diceType = e.target.dataset.dice;
                this.rollDice(diceType);
            });
        });

        // Custom roll button
        document.getElementById('custom-roll-btn').addEventListener('click', () => {
            const customDice = document.getElementById('custom-dice').value;
            if (customDice) {
                this.rollDice(customDice);
            }
        });

        // Enter key for custom roll
        document.getElementById('custom-dice').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                const customDice = e.target.value;
                if (customDice) {
                    this.rollDice(customDice);
                }
            }
        });

        // Modifier buttons
        document.querySelectorAll('.modifier-btn').forEach(button => {
            button.addEventListener('click', (e) => {
                const modifier = e.target.dataset.mod;
                const customDiceInput = document.getElementById('custom-dice');
                const currentValue = customDiceInput.value;
                
                // Remove existing modifier if present
                const baseValue = currentValue.replace(/[+-]\d+$/, '');
                customDiceInput.value = baseValue + modifier;
            });
        });
    }

    async rollDice(diceType) {
        const purpose = document.getElementById('roll-purpose').value;
        
        try {
            const result = await this.api.rollDice(diceType, purpose);
            this.displayResult(result);
            this.addToHistory(result);
            
            // Clear custom input after successful roll
            document.getElementById('custom-dice').value = '';
        } catch (error) {
            console.error('Failed to roll dice:', error);
            this.displayError('Invalid dice notation. Try something like "2d6+3"');
        }
    }

    displayResult(result) {
        const resultDiv = document.getElementById('roll-result');
        const isNatural20 = result.request.diceType.includes('d20') && result.results.includes(20);
        const isNatural1 = result.request.diceType.includes('d20') && result.results.includes(1);
        
        resultDiv.innerHTML = `
            <div class="roll-result ${isNatural20 ? 'natural-20' : ''} ${isNatural1 ? 'natural-1' : ''}">
                <h3>Roll Result</h3>
                <div class="dice-type">${result.request.diceType}</div>
                ${result.request.purpose ? `<div class="roll-purpose">${result.request.purpose}</div>` : ''}
                
                <div class="dice-display">
                    ${result.results.map(die => `
                        <div class="die">${die}</div>
                    `).join('')}
                </div>
                
                ${result.modifier !== 0 ? `
                    <div class="modifier-display">
                        Modifier: ${result.modifier >= 0 ? '+' : ''}${result.modifier}
                    </div>
                ` : ''}
                
                <div class="total-result">
                    Total: <span class="total-value">${result.total}</span>
                </div>
                
                ${isNatural20 ? '<div class="special-roll">Critical Success!</div>' : ''}
                ${isNatural1 ? '<div class="special-roll">Critical Failure!</div>' : ''}
            </div>
        `;

        // Animate dice
        resultDiv.querySelectorAll('.die').forEach((die, index) => {
            die.style.animationDelay = `${index * 0.1}s`;
        });
    }

    displayError(message) {
        const resultDiv = document.getElementById('roll-result');
        resultDiv.innerHTML = `
            <div class="roll-result error">
                <p>${message}</p>
            </div>
        `;
    }

    addToHistory(result) {
        this.rollHistory.unshift({
            ...result,
            timestamp: new Date().toLocaleTimeString()
        });

        // Keep only last 10 rolls
        if (this.rollHistory.length > 10) {
            this.rollHistory.pop();
        }

        this.updateHistoryDisplay();
    }

    updateHistoryDisplay() {
        const historyList = document.getElementById('history-list');
        
        if (this.rollHistory.length === 0) {
            historyList.innerHTML = '<p>No rolls yet</p>';
            return;
        }

        historyList.innerHTML = this.rollHistory.map(roll => `
            <div class="history-item">
                <span class="history-time">${roll.timestamp}</span>
                <span class="history-dice">${roll.request.diceType}</span>
                <span class="history-result">${roll.total}</span>
                ${roll.request.purpose ? `<span class="history-purpose">(${roll.request.purpose})</span>` : ''}
            </div>
        `).join('');
    }
}