import React, { useState } from 'react';

const CombatNarrator = ({ gameSessionId, currentCombat, onGenerate, isGenerating }) => {
    const [combatForm, setCombatForm] = useState({
        attackerName: '',
        targetName: '',
        weaponOrSpell: '',
        damage: 0,
        isHit: true,
        isCritical: false,
        targetHP: 100,
        targetMaxHP: 100
    });
    const [narrationStyle, setNarrationStyle] = useState('dramatic');
    const [recentNarrations, setRecentNarrations] = useState([]);

    const narrationStyles = [
        { value: 'dramatic', label: 'Dramatic & Epic', icon: '🎭' },
        { value: 'gritty', label: 'Gritty & Realistic', icon: '⚔️' },
        { value: 'heroic', label: 'Heroic & Inspiring', icon: '🦸' },
        { value: 'humorous', label: 'Light & Humorous', icon: '😄' },
        { value: 'dark', label: 'Dark & Brutal', icon: '💀' }
    ];

    const quickActions = [
        { type: 'melee', weapons: ['Sword', 'Axe', 'Hammer', 'Dagger', 'Fists'] },
        { type: 'ranged', weapons: ['Bow', 'Crossbow', 'Thrown Dagger', 'Javelin'] },
        { type: 'magic', weapons: ['Fireball', 'Lightning Bolt', 'Eldritch Blast', 'Sacred Flame'] }
    ];

    const handleGenerateNarration = () => {
        const params = {
            ...combatForm,
            actionType: narrationStyle,
            damage: parseInt(combatForm.damage),
            targetHP: parseInt(combatForm.targetHP),
            targetMaxHP: parseInt(combatForm.targetMaxHP)
        };

        // Determine if this is a death blow
        if (params.targetHP <= 0) {
            onGenerate('death_description', params);
        } else {
            onGenerate('combat_narration', params);
        }

        // Add to recent narrations
        setRecentNarrations(prev => [{
            ...params,
            timestamp: new Date(),
            id: Date.now()
        }, ...prev].slice(0, 10));
    };

    const quickFillCombat = (attacker, target, weapon) => {
        setCombatForm({
            ...combatForm,
            attackerName: attacker || combatForm.attackerName,
            targetName: target || combatForm.targetName,
            weaponOrSpell: weapon || combatForm.weaponOrSpell
        });
    };

    const calculateHealthPercentage = () => {
        if (combatForm.targetMaxHP === 0) return 100;
        return Math.round((combatForm.targetHP / combatForm.targetMaxHP) * 100);
    };

    const getHealthStatus = () => {
        const percentage = calculateHealthPercentage();
        if (percentage > 75) return { text: 'Healthy', color: 'green' };
        if (percentage > 50) return { text: 'Injured', color: 'yellow' };
        if (percentage > 25) return { text: 'Bloodied', color: 'orange' };
        if (percentage > 0) return { text: 'Critical', color: 'red' };
        return { text: 'Defeated', color: 'black' };
    };

    return (
        <div className="combat-narrator-panel">
            <div className="panel-header">
                <h3>Combat Narrator</h3>
                <div className="narration-style-selector">
                    {narrationStyles.map(style => (
                        <button
                            key={style.value}
                            className={`style-btn ${narrationStyle === style.value ? 'active' : ''}`}
                            onClick={() => setNarrationStyle(style.value)}
                            title={style.label}
                        >
                            {style.icon}
                        </button>
                    ))}
                </div>
            </div>

            {/* Quick Combat Setup from Current Combat */}
            {currentCombat && currentCombat.combatants && (
                <div className="current-combat-quick-fill">
                    <h4>Current Combat Participants</h4>
                    <div className="combatant-grid">
                        {currentCombat.combatants.map(combatant => (
                            <div key={combatant.id} className="combatant-quick">
                                <button
                                    onClick={() => quickFillCombat(combatant.name, null, null)}
                                    className="attacker-btn"
                                >
                                    ⚔️ {combatant.name}
                                </button>
                                <button
                                    onClick={() => quickFillCombat(null, combatant.name, null)}
                                    className="target-btn"
                                >
                                    🎯 Target
                                </button>
                                <span className="hp-display">
                                    {combatant.currentHP}/{combatant.maxHP} HP
                                </span>
                            </div>
                        ))}
                    </div>
                </div>
            )}

            <div className="combat-form">
                <div className="form-row">
                    <div className="form-field">
                        <label>Attacker</label>
                        <input
                            type="text"
                            placeholder="Character/Monster name"
                            value={combatForm.attackerName}
                            onChange={(e) => setCombatForm({...combatForm, attackerName: e.target.value})}
                        />
                    </div>
                    <div className="form-field">
                        <label>Target</label>
                        <input
                            type="text"
                            placeholder="Character/Monster name"
                            value={combatForm.targetName}
                            onChange={(e) => setCombatForm({...combatForm, targetName: e.target.value})}
                        />
                    </div>
                </div>

                <div className="weapon-selection">
                    <label>Weapon/Spell</label>
                    <input
                        type="text"
                        placeholder="e.g., Longsword, Fireball"
                        value={combatForm.weaponOrSpell}
                        onChange={(e) => setCombatForm({...combatForm, weaponOrSpell: e.target.value})}
                    />
                    <div className="quick-weapons">
                        {quickActions.map(category => (
                            <div key={category.type} className="weapon-category">
                                <span className="category-label">{category.type}:</span>
                                {category.weapons.map(weapon => (
                                    <button
                                        key={weapon}
                                        className="weapon-chip"
                                        onClick={() => setCombatForm({...combatForm, weaponOrSpell: weapon})}
                                    >
                                        {weapon}
                                    </button>
                                ))}
                            </div>
                        ))}
                    </div>
                </div>

                <div className="combat-outcome">
                    <div className="outcome-toggles">
                        <label className="toggle-label">
                            <input
                                type="checkbox"
                                checked={combatForm.isHit}
                                onChange={(e) => setCombatForm({...combatForm, isHit: e.target.checked})}
                            />
                            <span>Attack Hits</span>
                        </label>
                        <label className="toggle-label">
                            <input
                                type="checkbox"
                                checked={combatForm.isCritical}
                                onChange={(e) => setCombatForm({...combatForm, isCritical: e.target.checked})}
                                disabled={!combatForm.isHit}
                            />
                            <span>Critical Hit!</span>
                        </label>
                    </div>

                    {combatForm.isHit && (
                        <div className="damage-input">
                            <label>Damage Dealt</label>
                            <input
                                type="number"
                                min="0"
                                value={combatForm.damage}
                                onChange={(e) => setCombatForm({...combatForm, damage: e.target.value})}
                            />
                            <div className="damage-presets">
                                {[5, 10, 15, 20, 30, 50].map(dmg => (
                                    <button
                                        key={dmg}
                                        onClick={() => setCombatForm({...combatForm, damage: dmg})}
                                        className="damage-preset"
                                    >
                                        {dmg}
                                    </button>
                                ))}
                            </div>
                        </div>
                    )}
                </div>

                <div className="target-health">
                    <label>Target Health</label>
                    <div className="health-inputs">
                        <input
                            type="number"
                            min="0"
                            placeholder="Current HP"
                            value={combatForm.targetHP}
                            onChange={(e) => setCombatForm({...combatForm, targetHP: e.target.value})}
                        />
                        <span>/</span>
                        <input
                            type="number"
                            min="1"
                            placeholder="Max HP"
                            value={combatForm.targetMaxHP}
                            onChange={(e) => setCombatForm({...combatForm, targetMaxHP: e.target.value})}
                        />
                    </div>
                    <div className="health-bar">
                        <div 
                            className={`health-fill ${getHealthStatus().color}`}
                            style={{ width: `${calculateHealthPercentage()}%` }}
                        />
                    </div>
                    <span className={`health-status ${getHealthStatus().color}`}>
                        {getHealthStatus().text} ({calculateHealthPercentage()}%)
                    </span>
                </div>

                <button
                    className="generate-narration-btn"
                    onClick={handleGenerateNarration}
                    disabled={isGenerating || !combatForm.attackerName || !combatForm.targetName}
                >
                    {combatForm.targetHP <= 0 ? 'Generate Death Description' : 'Generate Combat Narration'}
                </button>
            </div>

            {/* Recent Narrations */}
            {recentNarrations.length > 0 && (
                <div className="recent-narrations">
                    <h4>Recent Narrations</h4>
                    <div className="narration-list">
                        {recentNarrations.map(narration => (
                            <div key={narration.id} className="narration-item">
                                <div className="narration-header">
                                    <span className="combatants">
                                        {narration.attackerName} → {narration.targetName}
                                    </span>
                                    <span className="damage">
                                        {narration.isHit ? `${narration.damage} damage` : 'Miss'}
                                    </span>
                                </div>
                                <button
                                    className="reuse-btn"
                                    onClick={() => setCombatForm(narration)}
                                >
                                    Reuse
                                </button>
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

export default CombatNarrator;