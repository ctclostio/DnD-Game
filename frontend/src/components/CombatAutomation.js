import React, { useState, useEffect } from 'react';
import BattleMapViewer from './BattleMapViewer';
import CombatAnalyticsView from './CombatAnalyticsView';
import { 
    autoResolveCombat, 
    generateBattleMap, 
    smartInitiative,
    getCombatAnalytics,
    getCombatHistory 
} from '../services/api';
import { getClickableProps, getSelectableProps } from '../utils/accessibility';
import '../styles/combat-automation.css';

const CombatAutomation = ({ gameSessionId, characters, npcs, isDM }) => {
    const [activeTab, setActiveTab] = useState('quick-combat');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    
    // Quick Combat state
    const [quickCombatForm, setQuickCombatForm] = useState({
        encounterDifficulty: 'medium',
        enemyTypes: [{ name: '', cr: '1', count: 1 }],
        terrainType: 'open',
        useResources: true
    });
    const [combatResolution, setCombatResolution] = useState(null);
    
    // Smart Initiative state
    const [initiativeList, setInitiativeList] = useState([]);
    
    // Battle Map state
    const [battleMapForm, setBattleMapForm] = useState({
        locationDescription: '',
        mapType: 'outdoor',
        desiredSize: 'medium',
        includeHazards: false,
        terrainComplexity: 'moderate'
    });
    const [currentBattleMap, setCurrentBattleMap] = useState(null);
    const [battleMaps, setBattleMaps] = useState([]);
    
    // Combat Analytics state
    const [combatHistory, setCombatHistory] = useState([]);
    const [selectedCombat, setSelectedCombat] = useState(null);

    useEffect(() => {
        if (gameSessionId) {
            loadCombatHistory();
        }
    }, [gameSessionId]);

    const loadCombatHistory = async () => {
        try {
            const history = await getCombatHistory(gameSessionId);
            setCombatHistory(history);
        } catch (err) {
            console.error('Failed to load combat history:', err);
        }
    };

    const handleQuickCombat = async () => {
        setLoading(true);
        setError(null);
        try {
            const resolution = await autoResolveCombat(gameSessionId, quickCombatForm);
            setCombatResolution(resolution);
            // Refresh combat history
            await loadCombatHistory();
        } catch (err) {
            setError('Failed to resolve combat');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const handleSmartInitiative = async () => {
        setLoading(true);
        setError(null);
        try {
            // Build combatants list
            const combatants = [
                ...characters.map(char => ({
                    id: char.id,
                    type: 'character',
                    name: char.name,
                    dexterity_modifier: char.abilities?.dexterity?.modifier || 0
                })),
                ...npcs.map(npc => ({
                    id: npc.id,
                    type: 'npc',
                    name: npc.name,
                    dexterity_modifier: npc.dexterity_modifier || 0
                }))
            ];

            const initiatives = await smartInitiative(gameSessionId, {
                combat_id: 'temp-' + Date.now(), // Temporary ID
                combatants
            });
            
            setInitiativeList(initiatives);
        } catch (err) {
            setError('Failed to roll initiative');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const handleGenerateBattleMap = async () => {
        setLoading(true);
        setError(null);
        try {
            const battleMap = await generateBattleMap(gameSessionId, battleMapForm);
            setCurrentBattleMap(battleMap);
            setBattleMaps([battleMap, ...battleMaps]);
        } catch (err) {
            setError('Failed to generate battle map');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const renderQuickCombat = () => (
        <div className="quick-combat-tab">
            <h3>Quick Combat Resolution</h3>
            <p>Resolve minor encounters without playing out every turn.</p>
            
            <div className="combat-form">
                <div className="form-group">
                    <label>Encounter Difficulty</label>
                    <select 
                        value={quickCombatForm.encounterDifficulty}
                        onChange={(e) => setQuickCombatForm({
                            ...quickCombatForm,
                            encounterDifficulty: e.target.value
                        })}
                    >
                        <option value="trivial">Trivial</option>
                        <option value="easy">Easy</option>
                        <option value="medium">Medium</option>
                        <option value="hard">Hard</option>
                        <option value="deadly">Deadly</option>
                    </select>
                </div>

                <div className="form-group">
                    <label>Enemy Types</label>
                    {quickCombatForm.enemyTypes.map((enemy, idx) => (
                        <div key={idx} className="enemy-row">
                            <input
                                type="text"
                                placeholder="Enemy name"
                                value={enemy.name}
                                onChange={(e) => {
                                    const newEnemies = [...quickCombatForm.enemyTypes];
                                    newEnemies[idx].name = e.target.value;
                                    setQuickCombatForm({ ...quickCombatForm, enemyTypes: newEnemies });
                                }}
                            />
                            <input
                                type="text"
                                placeholder="CR"
                                value={enemy.cr}
                                onChange={(e) => {
                                    const newEnemies = [...quickCombatForm.enemyTypes];
                                    newEnemies[idx].cr = e.target.value;
                                    setQuickCombatForm({ ...quickCombatForm, enemyTypes: newEnemies });
                                }}
                                className="cr-input"
                            />
                            <input
                                type="number"
                                placeholder="Count"
                                value={enemy.count}
                                min="1"
                                onChange={(e) => {
                                    const newEnemies = [...quickCombatForm.enemyTypes];
                                    newEnemies[idx].count = parseInt(e.target.value) || 1;
                                    setQuickCombatForm({ ...quickCombatForm, enemyTypes: newEnemies });
                                }}
                                className="count-input"
                            />
                            {quickCombatForm.enemyTypes.length > 1 && (
                                <button
                                    onClick={() => {
                                        const newEnemies = quickCombatForm.enemyTypes.filter((_, i) => i !== idx);
                                        setQuickCombatForm({ ...quickCombatForm, enemyTypes: newEnemies });
                                    }}
                                    className="remove-btn"
                                >
                                    ×
                                </button>
                            )}
                        </div>
                    ))}
                    <button
                        onClick={() => setQuickCombatForm({
                            ...quickCombatForm,
                            enemyTypes: [...quickCombatForm.enemyTypes, { name: '', cr: '1', count: 1 }]
                        })}
                        className="add-enemy-btn"
                    >
                        Add Enemy Type
                    </button>
                </div>

                <div className="form-row">
                    <div className="form-group">
                        <label>Terrain Type</label>
                        <select
                            value={quickCombatForm.terrainType}
                            onChange={(e) => setQuickCombatForm({
                                ...quickCombatForm,
                                terrainType: e.target.value
                            })}
                        >
                            <option value="open">Open Field</option>
                            <option value="forest">Forest</option>
                            <option value="dungeon">Dungeon</option>
                            <option value="urban">Urban</option>
                            <option value="mountain">Mountain</option>
                        </select>
                    </div>

                    <div className="form-group">
                        <label>
                            <input
                                type="checkbox"
                                checked={quickCombatForm.useResources}
                                onChange={(e) => setQuickCombatForm({
                                    ...quickCombatForm,
                                    useResources: e.target.checked
                                })}
                            />
                            Use spell slots and abilities
                        </label>
                    </div>
                </div>

                <button
                    onClick={handleQuickCombat}
                    disabled={loading || !quickCombatForm.enemyTypes[0].name}
                    className="resolve-btn"
                >
                    Resolve Combat
                </button>
            </div>

            {combatResolution && (
                <div className="combat-resolution">
                    <h4>Combat Resolution</h4>
                    <div className={`outcome ${combatResolution.outcome}`}>
                        <span className="outcome-label">Outcome:</span> 
                        <span className="outcome-value">{combatResolution.outcome.replace('_', ' ')}</span>
                    </div>
                    
                    <p className="narrative">{combatResolution.narrative_summary}</p>
                    
                    <div className="resolution-details">
                        <div className="detail-section">
                            <h5>Combat Summary</h5>
                            <ul>
                                <li>Duration: {combatResolution.rounds_simulated} rounds</li>
                                <li>Experience Gained: {combatResolution.experience_awarded} XP</li>
                                {combatResolution.party_resources_used?.hp_lost && (
                                    <li>Total HP Lost: {combatResolution.party_resources_used.hp_lost}</li>
                                )}
                            </ul>
                        </div>

                        {combatResolution.loot_generated && combatResolution.loot_generated.length > 0 && (
                            <div className="detail-section">
                                <h5>Loot</h5>
                                <ul>
                                    {combatResolution.loot_generated.map((item, idx) => (
                                        <li key={idx}>
                                            {item.type === 'currency' 
                                                ? `${item.amount} ${item.currency}`
                                                : `${item.name} (${item.rarity})`
                                            }
                                        </li>
                                    ))}
                                </ul>
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    );

    const renderSmartInitiative = () => (
        <div className="smart-initiative-tab">
            <h3>Smart Initiative Tracker</h3>
            <p>Automatically roll initiative for all combatants with bonuses applied.</p>
            
            <button
                onClick={handleSmartInitiative}
                disabled={loading || (characters.length === 0 && npcs.length === 0)}
                className="roll-initiative-btn"
            >
                Roll Initiative for All
            </button>

            {initiativeList.length > 0 && (
                <div className="initiative-results">
                    <h4>Initiative Order</h4>
                    <table className="initiative-table">
                        <thead>
                            <tr>
                                <th>Order</th>
                                <th>Name</th>
                                <th>Type</th>
                                <th>Initiative</th>
                                <th>Roll</th>
                                <th>Bonus</th>
                            </tr>
                        </thead>
                        <tbody>
                            {initiativeList.map((entry, idx) => (
                                <tr key={entry.id} className={entry.type}>
                                    <td className="order">{idx + 1}</td>
                                    <td className="name">{entry.name}</td>
                                    <td className="type">{entry.type}</td>
                                    <td className="total">{entry.initiative}</td>
                                    <td className="roll">d20: {entry.roll}</td>
                                    <td className="bonus">+{entry.bonus}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );

    const renderBattleMaps = () => (
        <div className="battle-maps-tab">
            <h3>Dynamic Battle Maps</h3>
            {isDM ? (
                <div className="map-generator">
                    <h4>Generate New Map</h4>
                    <div className="map-form">
                        <div className="form-group">
                            <label>Location Description</label>
                            <textarea
                                value={battleMapForm.locationDescription}
                                onChange={(e) => setBattleMapForm({
                                    ...battleMapForm,
                                    locationDescription: e.target.value
                                })}
                                placeholder="A dark forest clearing with ancient stone ruins..."
                                rows="3"
                            />
                        </div>

                        <div className="form-row">
                            <div className="form-group">
                                <label>Map Type</label>
                                <select
                                    value={battleMapForm.mapType}
                                    onChange={(e) => setBattleMapForm({
                                        ...battleMapForm,
                                        mapType: e.target.value
                                    })}
                                >
                                    <option value="outdoor">Outdoor</option>
                                    <option value="dungeon">Dungeon</option>
                                    <option value="urban">Urban</option>
                                    <option value="special">Special</option>
                                </select>
                            </div>

                            <div className="form-group">
                                <label>Size</label>
                                <select
                                    value={battleMapForm.desiredSize}
                                    onChange={(e) => setBattleMapForm({
                                        ...battleMapForm,
                                        desiredSize: e.target.value
                                    })}
                                >
                                    <option value="small">Small (15x15)</option>
                                    <option value="medium">Medium (20x20)</option>
                                    <option value="large">Large (30x30)</option>
                                    <option value="huge">Huge (40x40)</option>
                                </select>
                            </div>

                            <div className="form-group">
                                <label>Complexity</label>
                                <select
                                    value={battleMapForm.terrainComplexity}
                                    onChange={(e) => setBattleMapForm({
                                        ...battleMapForm,
                                        terrainComplexity: e.target.value
                                    })}
                                >
                                    <option value="simple">Simple</option>
                                    <option value="moderate">Moderate</option>
                                    <option value="complex">Complex</option>
                                </select>
                            </div>
                        </div>

                        <div className="form-group">
                            <label>
                                <input
                                    type="checkbox"
                                    checked={battleMapForm.includeHazards}
                                    onChange={(e) => setBattleMapForm({
                                        ...battleMapForm,
                                        includeHazards: e.target.checked
                                    })}
                                />
                                Include environmental hazards
                            </label>
                        </div>

                        <button
                            onClick={handleGenerateBattleMap}
                            disabled={loading || !battleMapForm.locationDescription}
                            className="generate-map-btn"
                        >
                            Generate Battle Map
                        </button>
                    </div>
                </div>
            ) : (
                <p className="dm-only">Only the DM can generate battle maps</p>
            )}

            {currentBattleMap && (
                <div className="battle-map-viewer">
                    <h4>Current Battle Map</h4>
                    <BattleMapViewer battleMap={currentBattleMap} />
                </div>
            )}
        </div>
    );

    const renderCombatAnalytics = () => (
        <div className="combat-analytics-tab">
            <h3>Combat Analytics</h3>
            <div className="analytics-container">
                <div className="combat-list">
                    <h4>Combat History</h4>
                    {combatHistory.length === 0 ? (
                        <p className="empty-state">No combat history yet</p>
                    ) : (
                        <div className="history-list">
                            {combatHistory.combat_analytics?.map(combat => (
                                <div
                                    key={combat.id}
                                    className={`combat-entry ${selectedCombat?.id === combat.id ? 'selected' : ''}`}
                                    {...getSelectableProps(() => setSelectedCombat(combat), selectedCombat?.id === combat.id)}
                                >
                                    <div className="combat-header">
                                        <span className="combat-date">
                                            {new Date(combat.created_at).toLocaleDateString()}
                                        </span>
                                        <span className="combat-duration">
                                            {combat.combat_duration} rounds
                                        </span>
                                    </div>
                                    <div className="combat-stats">
                                        <span>Damage: {combat.total_damage_dealt}</span>
                                        <span>Healing: {combat.total_healing_done}</span>
                                        {combat.mvp_id && <span className="mvp">MVP</span>}
                                    </div>
                                </div>
                            ))}
                            
                            {combatHistory.auto_resolutions?.map(resolution => (
                                <div
                                    key={resolution.id}
                                    className="combat-entry auto-resolved"
                                >
                                    <div className="combat-header">
                                        <span className="combat-date">
                                            {new Date(resolution.created_at).toLocaleDateString()}
                                        </span>
                                        <span className="resolution-type">Quick Combat</span>
                                    </div>
                                    <div className="combat-stats">
                                        <span>Outcome: {resolution.outcome}</span>
                                        <span>XP: {resolution.experience_awarded}</span>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                {selectedCombat && (
                    <div className="analytics-details">
                        <CombatAnalyticsView combatId={selectedCombat.combat_id} />
                    </div>
                )}
            </div>
        </div>
    );

    return (
        <div className="combat-automation">
            <div className="automation-tabs">
                <button
                    className={activeTab === 'quick-combat' ? 'active' : ''}
                    onClick={() => setActiveTab('quick-combat')}
                >
                    Quick Combat
                </button>
                <button
                    className={activeTab === 'smart-initiative' ? 'active' : ''}
                    onClick={() => setActiveTab('smart-initiative')}
                >
                    Smart Initiative
                </button>
                <button
                    className={activeTab === 'battle-maps' ? 'active' : ''}
                    onClick={() => setActiveTab('battle-maps')}
                >
                    Battle Maps
                </button>
                <button
                    className={activeTab === 'analytics' ? 'active' : ''}
                    onClick={() => setActiveTab('analytics')}
                >
                    Analytics
                </button>
            </div>

            {error && <div className="error-message">{error}</div>}

            <div className="tab-content">
                {loading && <div className="loading">Processing...</div>}
                {!loading && (
                    <>
                        {activeTab === 'quick-combat' && renderQuickCombat()}
                        {activeTab === 'smart-initiative' && renderSmartInitiative()}
                        {activeTab === 'battle-maps' && renderBattleMaps()}
                        {activeTab === 'analytics' && renderCombatAnalytics()}
                    </>
                )}
            </div>
        </div>
    );
};

export default CombatAutomation;