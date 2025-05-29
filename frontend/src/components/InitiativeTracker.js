import React, { useState, useEffect } from 'react';
import { startCombat, nextTurn, getCombatBySession, processCombatAction } from '../services/api';
import '../styles/initiative-tracker.css';

const InitiativeTracker = ({ gameSessionId, characters, npcs, onCombatUpdate }) => {
    const [combat, setCombat] = useState(null);
    const [selectedCombatants, setSelectedCombatants] = useState([]);
    const [initiativeInputs, setInitiativeInputs] = useState({});
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [turnTimer, setTurnTimer] = useState(null);
    const [elapsedTime, setElapsedTime] = useState(0);

    useEffect(() => {
        if (gameSessionId) {
            checkActiveCombat();
        }
    }, [gameSessionId]);

    useEffect(() => {
        let interval;
        if (combat?.isActive && turnTimer) {
            interval = setInterval(() => {
                setElapsedTime(prev => prev + 1);
            }, 1000);
        }
        return () => clearInterval(interval);
    }, [combat?.isActive, turnTimer]);

    const checkActiveCombat = async () => {
        try {
            const response = await getCombatBySession(gameSessionId);
            if (response.data) {
                setCombat(response.data);
                if (onCombatUpdate) onCombatUpdate(response.data);
            }
        } catch (err) {
            // No active combat
        }
    };

    const handleAddCombatant = (entity, type) => {
        const combatant = {
            id: entity.id,
            name: entity.name,
            type: type,
            initiative: initiativeInputs[entity.id] || 0,
            hp: entity.hitPoints || entity.hp,
            maxHp: entity.maxHitPoints || entity.maxHp,
            ac: entity.armorClass || entity.ac,
            speed: entity.speed || 30
        };
        
        setSelectedCombatants([...selectedCombatants, combatant]);
    };

    const handleRemoveCombatant = (id) => {
        setSelectedCombatants(selectedCombatants.filter(c => c.id !== id));
    };

    const handleInitiativeChange = (id, value) => {
        setInitiativeInputs({
            ...initiativeInputs,
            [id]: parseInt(value) || 0
        });
    };

    const handleStartCombat = async () => {
        if (selectedCombatants.length < 2) {
            setError('At least 2 combatants required to start combat');
            return;
        }

        setLoading(true);
        try {
            const combatants = selectedCombatants.map(c => ({
                characterId: c.type === 'character' ? c.id : undefined,
                npcId: c.type === 'npc' ? c.id : undefined,
                name: c.name,
                initiative: c.initiative,
                hp: c.hp,
                maxHp: c.maxHp,
                ac: c.ac,
                speed: c.speed
            }));

            const response = await startCombat(gameSessionId, combatants);
            setCombat(response.data);
            setSelectedCombatants([]);
            setInitiativeInputs({});
            setError('');
            setTurnTimer(new Date());
            setElapsedTime(0);
            
            if (onCombatUpdate) onCombatUpdate(response.data);
        } catch (err) {
            setError('Failed to start combat');
        } finally {
            setLoading(false);
        }
    };

    const handleNextTurn = async () => {
        if (!combat) return;

        setLoading(true);
        try {
            const response = await nextTurn(combat.id);
            setCombat(response.data.combat);
            setTurnTimer(new Date());
            setElapsedTime(0);
            
            if (onCombatUpdate) onCombatUpdate(response.data.combat);
        } catch (err) {
            setError('Failed to advance turn');
        } finally {
            setLoading(false);
        }
    };

    const handleEndCombat = async () => {
        if (!combat) return;
        
        if (window.confirm('Are you sure you want to end combat?')) {
            try {
                // End combat logic here
                setCombat(null);
                setTurnTimer(null);
                setElapsedTime(0);
                if (onCombatUpdate) onCombatUpdate(null);
            } catch (err) {
                setError('Failed to end combat');
            }
        }
    };

    const getCurrentCombatant = () => {
        if (!combat || !combat.turnOrder.length) return null;
        const currentId = combat.turnOrder[combat.currentTurn];
        return combat.combatants.find(c => c.id === currentId);
    };

    const formatTime = (seconds) => {
        const mins = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return `${mins}:${secs.toString().padStart(2, '0')}`;
    };

    if (!combat) {
        return (
            <div className="initiative-tracker">
                <h3>Combat Setup</h3>
                {error && <div className="error">{error}</div>}
                
                <div className="combatant-selection">
                    <div className="available-combatants">
                        <h4>Characters</h4>
                        {characters.map(char => (
                            <div key={char.id} className="combatant-option">
                                <span>{char.name}</span>
                                <input
                                    type="number"
                                    placeholder="Init"
                                    value={initiativeInputs[char.id] || ''}
                                    onChange={(e) => handleInitiativeChange(char.id, e.target.value)}
                                    className="initiative-input"
                                />
                                <button onClick={() => handleAddCombatant(char, 'character')}>
                                    Add
                                </button>
                            </div>
                        ))}
                        
                        <h4>NPCs</h4>
                        {npcs.map(npc => (
                            <div key={npc.id} className="combatant-option">
                                <span>{npc.name}</span>
                                <input
                                    type="number"
                                    placeholder="Init"
                                    value={initiativeInputs[npc.id] || ''}
                                    onChange={(e) => handleInitiativeChange(npc.id, e.target.value)}
                                    className="initiative-input"
                                />
                                <button onClick={() => handleAddCombatant(npc, 'npc')}>
                                    Add
                                </button>
                            </div>
                        ))}
                    </div>
                    
                    <div className="selected-combatants">
                        <h4>Selected Combatants ({selectedCombatants.length})</h4>
                        {selectedCombatants.map(combatant => (
                            <div key={combatant.id} className="selected-combatant">
                                <span>{combatant.name}</span>
                                <span className="initiative-badge">Init: {combatant.initiative}</span>
                                <button onClick={() => handleRemoveCombatant(combatant.id)}>
                                    Remove
                                </button>
                            </div>
                        ))}
                    </div>
                </div>
                
                <button 
                    className="start-combat-btn"
                    onClick={handleStartCombat}
                    disabled={loading || selectedCombatants.length < 2}
                >
                    Start Combat
                </button>
            </div>
        );
    }

    const currentCombatant = getCurrentCombatant();

    return (
        <div className="initiative-tracker active">
            <div className="combat-header">
                <h3>Combat - Round {combat.round}</h3>
                <div className="combat-controls">
                    <div className="turn-timer">
                        Turn Time: {formatTime(elapsedTime)}
                    </div>
                    <button onClick={handleNextTurn} disabled={loading}>
                        Next Turn
                    </button>
                    <button onClick={handleEndCombat} className="end-combat">
                        End Combat
                    </button>
                </div>
            </div>
            
            {error && <div className="error">{error}</div>}
            
            <div className="initiative-order">
                {combat.combatants
                    .sort((a, b) => b.initiative - a.initiative)
                    .map((combatant, index) => {
                        const isActive = currentCombatant?.id === combatant.id;
                        const isDead = combatant.hp <= 0;
                        
                        return (
                            <div 
                                key={combatant.id}
                                className={`combatant-card ${isActive ? 'active' : ''} ${isDead ? 'dead' : ''}`}
                            >
                                <div className="combatant-header">
                                    <div className="combatant-info">
                                        <span className="combatant-name">{combatant.name}</span>
                                        <span className="initiative-value">Init: {combatant.initiative}</span>
                                    </div>
                                    {isActive && <span className="current-turn">CURRENT TURN</span>}
                                </div>
                                
                                <div className="combatant-stats">
                                    <div className="stat">
                                        <span className="stat-label">HP:</span>
                                        <span className={`stat-value ${combatant.hp < combatant.maxHp / 4 ? 'critical' : ''}`}>
                                            {combatant.hp}/{combatant.maxHp}
                                        </span>
                                        {combatant.tempHp > 0 && <span className="temp-hp">+{combatant.tempHp}</span>}
                                    </div>
                                    <div className="stat">
                                        <span className="stat-label">AC:</span>
                                        <span className="stat-value">{combatant.ac}</span>
                                    </div>
                                    <div className="stat">
                                        <span className="stat-label">Speed:</span>
                                        <span className="stat-value">{combatant.movement}/{combatant.speed}ft</span>
                                    </div>
                                </div>
                                
                                <div className="action-economy">
                                    <div className={`action-marker ${combatant.actions > 0 ? 'available' : 'used'}`}>
                                        Action
                                    </div>
                                    <div className={`action-marker ${combatant.bonusActions > 0 ? 'available' : 'used'}`}>
                                        Bonus
                                    </div>
                                    <div className={`action-marker ${combatant.reactions > 0 ? 'available' : 'used'}`}>
                                        Reaction
                                    </div>
                                </div>
                                
                                {combatant.conditions && combatant.conditions.length > 0 && (
                                    <div className="conditions">
                                        {combatant.conditions.map((condition, idx) => (
                                            <span key={idx} className={`condition ${condition.name}`}>
                                                {condition.name}
                                            </span>
                                        ))}
                                    </div>
                                )}
                                
                                {combatant.isConcentrating && (
                                    <div className="concentration">
                                        Concentrating: {combatant.concentrationSpell}
                                    </div>
                                )}
                            </div>
                        );
                    })}
            </div>
        </div>
    );
};

export default InitiativeTracker;