import React, { useState, useEffect } from 'react';

const FactionRelationships = ({ factions, sessionId, onUpdate }) => {
    const [relationshipMatrix, setRelationshipMatrix] = useState({});
    const [selectedRelationship, setSelectedRelationship] = useState(null);

    useEffect(() => {
        buildRelationshipMatrix();
    }, [factions]);

    const buildRelationshipMatrix = () => {
        const matrix = {};
        
        factions.forEach(faction => {
            matrix[faction.id] = {};
            
            if (faction.factionRelationships) {
                Object.entries(faction.factionRelationships).forEach(([targetId, relationship]) => {
                    matrix[faction.id][targetId] = relationship;
                });
            }
        });
        
        setRelationshipMatrix(matrix);
    };

    const getRelationshipColor = (standing) => {
        if (!standing && standing !== 0) return 'neutral';
        if (standing >= 50) return 'ally';
        if (standing <= -50) return 'enemy';
        return 'neutral';
    };

    const getRelationshipSymbol = (standing) => {
        if (!standing && standing !== 0) return '○';
        if (standing >= 75) return '❤️';
        if (standing >= 50) return '🤝';
        if (standing >= 25) return '👍';
        if (standing > -25) return '○';
        if (standing > -50) return '👎';
        if (standing > -75) return '⚔️';
        return '💀';
    };

    const handleRelationshipUpdate = async (faction1Id, faction2Id, change, reason) => {
        try {
            const response = await fetch(
                `/api/v1/factions/${faction1Id}/relationships/${faction2Id}`,
                {
                    method: 'PUT',
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ change, reason })
                }
            );

            if (!response.ok) throw new Error('Failed to update relationship');

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
            console.error('Error updating relationship:', err);
        }
    };

    return (
        <div className="faction-relationships">
            <div className="relationship-matrix">
                <table>
                    <thead>
                        <tr>
                            <th></th>
                            {factions.map(faction => (
                                <th key={faction.id} className="faction-column-header">
                                    <div className="header-content">
                                        <span className="faction-name">{faction.name}</span>
                                    </div>
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {factions.map(faction1 => (
                            <tr key={faction1.id}>
                                <td className="faction-row-header">
                                    <span className="faction-name">{faction1.name}</span>
                                </td>
                                {factions.map(faction2 => {
                                    if (faction1.id === faction2.id) {
                                        return <td key={faction2.id} className="self-cell">-</td>;
                                    }

                                    const relationship = relationshipMatrix[faction1.id]?.[faction2.id];
                                    const standing = relationship?.standing || 0;
                                    const relationshipType = getRelationshipColor(standing);

                                    return (
                                        <td
                                            key={faction2.id}
                                            className={`relationship-cell ${relationshipType}`}
                                            onClick={() => setSelectedRelationship({
                                                faction1,
                                                faction2,
                                                standing,
                                                type: relationship?.type || 'neutral'
                                            })}
                                        >
                                            <div className="relationship-indicator">
                                                <span className="relationship-symbol">
                                                    {getRelationshipSymbol(standing)}
                                                </span>
                                                <span className="relationship-value">
                                                    {standing}
                                                </span>
                                            </div>
                                        </td>
                                    );
                                })}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <div className="relationship-legend">
                <h4>Relationship Legend</h4>
                <div className="legend-items">
                    <div className="legend-item">
                        <span className="symbol">❤️</span>
                        <span>Strong Allies (75+)</span>
                    </div>
                    <div className="legend-item">
                        <span className="symbol">🤝</span>
                        <span>Allies (50-74)</span>
                    </div>
                    <div className="legend-item">
                        <span className="symbol">👍</span>
                        <span>Friendly (25-49)</span>
                    </div>
                    <div className="legend-item">
                        <span className="symbol">○</span>
                        <span>Neutral (-24 to 24)</span>
                    </div>
                    <div className="legend-item">
                        <span className="symbol">👎</span>
                        <span>Unfriendly (-25 to -49)</span>
                    </div>
                    <div className="legend-item">
                        <span className="symbol">⚔️</span>
                        <span>Enemies (-50 to -74)</span>
                    </div>
                    <div className="legend-item">
                        <span className="symbol">💀</span>
                        <span>Bitter Enemies (-75+)</span>
                    </div>
                </div>
            </div>

            {selectedRelationship && (
                <div className="relationship-modifier">
                    <h4>Modify Relationship</h4>
                    <p>
                        <strong>{selectedRelationship.faction1.name}</strong> → 
                        <strong> {selectedRelationship.faction2.name}</strong>
                    </p>
                    <p>Current Standing: {selectedRelationship.standing}</p>
                    
                    <div className="modifier-actions">
                        <button
                            onClick={() => handleRelationshipUpdate(
                                selectedRelationship.faction1.id,
                                selectedRelationship.faction2.id,
                                10,
                                'Diplomatic success'
                            )}
                            className="btn btn-positive"
                        >
                            Improve (+10)
                        </button>
                        <button
                            onClick={() => handleRelationshipUpdate(
                                selectedRelationship.faction1.id,
                                selectedRelationship.faction2.id,
                                -10,
                                'Diplomatic incident'
                            )}
                            className="btn btn-negative"
                        >
                            Worsen (-10)
                        </button>
                        <button
                            onClick={() => handleRelationshipUpdate(
                                selectedRelationship.faction1.id,
                                selectedRelationship.faction2.id,
                                25,
                                'Major alliance'
                            )}
                            className="btn btn-positive"
                        >
                            Major Alliance (+25)
                        </button>
                        <button
                            onClick={() => handleRelationshipUpdate(
                                selectedRelationship.faction1.id,
                                selectedRelationship.faction2.id,
                                -25,
                                'Major conflict'
                            )}
                            className="btn btn-negative"
                        >
                            Major Conflict (-25)
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default FactionRelationships;