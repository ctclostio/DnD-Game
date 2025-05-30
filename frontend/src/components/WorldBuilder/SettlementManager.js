import React, { useState } from 'react';
import SettlementDetails from './SettlementDetails';
import SettlementGenerator from './SettlementGenerator';

const SettlementManager = ({ sessionId, settlements, onUpdate }) => {
    const [selectedSettlement, setSelectedSettlement] = useState(null);
    const [isGenerating, setIsGenerating] = useState(false);
    const [showGenerator, setShowGenerator] = useState(false);

    const handleGenerateSettlement = async (generationData) => {
        setIsGenerating(true);
        
        try {
            const response = await fetch(`/api/v1/sessions/${sessionId}/settlements/generate`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(generationData)
            });

            if (!response.ok) throw new Error('Failed to generate settlement');
            
            const newSettlement = await response.json();
            
            // Update settlements list
            const updatedSettlements = [...settlements, newSettlement];
            onUpdate(updatedSettlements);
            
            // Select the new settlement
            setSelectedSettlement(newSettlement);
            setShowGenerator(false);
        } catch (err) {
            console.error('Error generating settlement:', err);
            alert('Failed to generate settlement');
        } finally {
            setIsGenerating(false);
        }
    };

    const getSettlementTypeIcon = (type) => {
        const icons = {
            hamlet: '🏘️',
            village: '🏡',
            town: '🏛️',
            city: '🏰',
            metropolis: '🌆',
            ruins: '🏚️'
        };
        return icons[type] || '🏘️';
    };

    const getCorruptionColor = (level) => {
        if (level <= 2) return 'low';
        if (level <= 5) return 'medium';
        if (level <= 8) return 'high';
        return 'extreme';
    };

    return (
        <div className="settlement-manager">
            <div className="manager-header">
                <h3>Settlements</h3>
                <button 
                    className="btn btn-primary"
                    onClick={() => setShowGenerator(true)}
                    disabled={isGenerating}
                >
                    + Generate Settlement
                </button>
            </div>

            <div className="settlement-layout">
                <div className="settlement-list">
                    {settlements.length === 0 ? (
                        <div className="empty-state">
                            <p>No settlements yet</p>
                            <p className="hint">Generate your first settlement to begin</p>
                        </div>
                    ) : (
                        settlements.map(settlement => (
                            <div
                                key={settlement.id}
                                className={`settlement-card ${selectedSettlement?.id === settlement.id ? 'selected' : ''}`}
                                onClick={() => setSelectedSettlement(settlement)}
                            >
                                <div className="settlement-header">
                                    <span className="settlement-icon">
                                        {getSettlementTypeIcon(settlement.type)}
                                    </span>
                                    <div className="settlement-info">
                                        <h4>{settlement.name}</h4>
                                        <span className="settlement-type">{settlement.type}</span>
                                    </div>
                                </div>
                                
                                <div className="settlement-stats">
                                    <div className="stat">
                                        <span className="icon">👥</span>
                                        <span>{settlement.population.toLocaleString()}</span>
                                    </div>
                                    <div className="stat">
                                        <span className="icon">⚠️</span>
                                        <span>Danger: {settlement.dangerLevel}/10</span>
                                    </div>
                                    <div className={`stat corruption-${getCorruptionColor(settlement.corruptionLevel)}`}>
                                        <span className="icon">🌑</span>
                                        <span>Corruption: {settlement.corruptionLevel}/10</span>
                                    </div>
                                </div>

                                {settlement.ancientRuinsNearby && (
                                    <div className="ancient-marker">
                                        <span className="icon">🗿</span>
                                        Ancient ruins nearby
                                    </div>
                                )}
                            </div>
                        ))
                    )}
                </div>

                <div className="settlement-detail-panel">
                    {selectedSettlement ? (
                        <SettlementDetails
                            settlement={selectedSettlement}
                            sessionId={sessionId}
                            onUpdate={(updated) => {
                                const updatedSettlements = settlements.map(s =>
                                    s.id === updated.id ? updated : s
                                );
                                onUpdate(updatedSettlements);
                                setSelectedSettlement(updated);
                            }}
                        />
                    ) : (
                        <div className="no-selection">
                            <p>Select a settlement to view details</p>
                        </div>
                    )}
                </div>
            </div>

            {showGenerator && (
                <SettlementGenerator
                    onGenerate={handleGenerateSettlement}
                    onClose={() => setShowGenerator(false)}
                    isGenerating={isGenerating}
                />
            )}
        </div>
    );
};

export default SettlementManager;