import React, { useState } from 'react';
import FactionCreator from './FactionCreator';
import FactionRelationships from './FactionRelationships';

const FactionManager = ({ sessionId, factions, settlements, onUpdate }) => {
    const [selectedFaction, setSelectedFaction] = useState(null);
    const [showCreator, setShowCreator] = useState(false);
    const [viewMode, setViewMode] = useState('list'); // list or relationships

    const handleCreateFaction = async (factionData) => {
        try {
            const response = await fetch(`/api/v1/sessions/${sessionId}/factions`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(factionData)
            });

            if (!response.ok) throw new Error('Failed to create faction');
            
            const newFaction = await response.json();
            onUpdate([...factions, newFaction]);
            setShowCreator(false);
        } catch (err) {
            console.error('Error creating faction:', err);
            alert('Failed to create faction');
        }
    };

    const handleSimulateConflicts = async () => {
        try {
            const response = await fetch(`/api/v1/sessions/${sessionId}/factions/simulate-conflicts`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (!response.ok) throw new Error('Failed to simulate conflicts');
            
            const events = await response.json();
            alert(`Generated ${events.length} faction events`);
            
            // Reload factions to get updated relationships
            const factionsResponse = await fetch(`/api/v1/sessions/${sessionId}/factions`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });
            
            if (factionsResponse.ok) {
                const updatedFactions = await factionsResponse.json();
                onUpdate(updatedFactions);
            }
        } catch (err) {
            console.error('Error simulating conflicts:', err);
        }
    };

    const getFactionIcon = (type) => {
        const icons = {
            religious: '⛪',
            political: '👑',
            criminal: '🗡️',
            merchant: '💰',
            military: '⚔️',
            cult: '🌑',
            ancient_order: '🗿'
        };
        return icons[type] || '🏛️';
    };

    const getFactionPowerLevel = (faction) => {
        const totalPower = faction.influenceLevel + faction.militaryStrength + 
                          faction.economicPower + faction.magicalResources;
        return Math.round(totalPower / 4);
    };

    return (
        <div className="faction-manager">
            <div className="manager-header">
                <h3>Factions</h3>
                <div className="header-actions">
                    <button
                        className={`view-toggle ${viewMode === 'list' ? 'active' : ''}`}
                        onClick={() => setViewMode('list')}
                    >
                        List View
                    </button>
                    <button
                        className={`view-toggle ${viewMode === 'relationships' ? 'active' : ''}`}
                        onClick={() => setViewMode('relationships')}
                    >
                        Relationships
                    </button>
                    <button onClick={handleSimulateConflicts} className="btn btn-secondary">
                        Simulate Conflicts
                    </button>
                    <button onClick={() => setShowCreator(true)} className="btn btn-primary">
                        + Create Faction
                    </button>
                </div>
            </div>

            {viewMode === 'list' ? (
                <div className="faction-list-view">
                    <div className="faction-grid">
                        {factions.map(faction => (
                            <div
                                key={faction.id}
                                className={`faction-card ${selectedFaction?.id === faction.id ? 'selected' : ''}`}
                                onClick={() => setSelectedFaction(faction)}
                            >
                                <div className="faction-header">
                                    <span className="faction-icon">{getFactionIcon(faction.type)}</span>
                                    <div className="faction-info">
                                        <h4>{faction.name}</h4>
                                        <span className="faction-type">{faction.type}</span>
                                    </div>
                                </div>

                                <div className="faction-power">
                                    <div className="power-bar">
                                        <div 
                                            className="power-fill"
                                            style={{ width: `${getFactionPowerLevel(faction) * 10}%` }}
                                        />
                                    </div>
                                    <span className="power-label">
                                        Power: {getFactionPowerLevel(faction)}/10
                                    </span>
                                </div>

                                <div className="faction-stats">
                                    <div className="stat" title="Influence">
                                        <span>👥</span> {faction.influenceLevel}
                                    </div>
                                    <div className="stat" title="Military">
                                        <span>⚔️</span> {faction.militaryStrength}
                                    </div>
                                    <div className="stat" title="Economic">
                                        <span>💰</span> {faction.economicPower}
                                    </div>
                                    <div className="stat" title="Magical">
                                        <span>✨</span> {faction.magicalResources}
                                    </div>
                                </div>

                                {faction.corrupted && (
                                    <div className="corruption-badge">🌑 Corrupted</div>
                                )}
                                {faction.seeksAncientPower && (
                                    <div className="ancient-badge">🗿 Seeks Ancient Power</div>
                                )}
                                {faction.guardsAncientSecrets && (
                                    <div className="guardian-badge">🛡️ Guards Secrets</div>
                                )}

                                <div className="faction-details">
                                    <p className="faction-description">{faction.description}</p>
                                    <p className="member-count">
                                        Members: {faction.memberCount.toLocaleString()}
                                    </p>
                                </div>
                            </div>
                        ))}
                    </div>

                    {selectedFaction && (
                        <div className="faction-detail-panel">
                            <h4>{selectedFaction.name} Details</h4>
                            
                            <div className="detail-section">
                                <h5>Public Goals</h5>
                                <ul>
                                    {selectedFaction.publicGoals?.map((goal, index) => (
                                        <li key={index}>{goal}</li>
                                    ))}
                                </ul>
                            </div>

                            <div className="detail-section secret">
                                <h5>Secret Goals (DM Only)</h5>
                                <ul>
                                    {selectedFaction.secretGoals?.map((goal, index) => (
                                        <li key={index}>{goal}</li>
                                    ))}
                                </ul>
                            </div>

                            <div className="detail-section">
                                <h5>Resources</h5>
                                <p><strong>Leadership:</strong> {selectedFaction.leadershipStructure}</p>
                                <p><strong>Headquarters:</strong> {selectedFaction.headquartersLocation}</p>
                                <p><strong>Founded:</strong> {selectedFaction.foundingDate}</p>
                            </div>

                            {selectedFaction.ancientKnowledgeLevel > 0 && (
                                <div className="detail-section ancient">
                                    <h5>Ancient Connections</h5>
                                    <p>Knowledge Level: {selectedFaction.ancientKnowledgeLevel}/10</p>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            ) : (
                <FactionRelationships
                    factions={factions}
                    sessionId={sessionId}
                    onUpdate={onUpdate}
                />
            )}

            {showCreator && (
                <FactionCreator
                    onClose={() => setShowCreator(false)}
                    onSubmit={handleCreateFaction}
                />
            )}
        </div>
    );
};

export default FactionManager;