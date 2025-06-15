import React, { useState, useEffect } from 'react';
import { api } from '../../services/api';
import { getSelectableProps, getClickableProps } from '../../utils/accessibility';

const EnvironmentalHazards = ({ gameSessionId, currentLocation, onGenerate, isGenerating }) => {
    const [hazardForm, setHazardForm] = useState({
        locationType: currentLocation?.type || 'dungeon',
        difficulty: 5,
        hazardType: 'trap'
    });
    const [activeHazards, setActiveHazards] = useState([]);
    const [selectedHazard, setSelectedHazard] = useState(null);

    useEffect(() => {
        if (currentLocation?.id) {
            loadLocationHazards(currentLocation.id);
        }
    }, [currentLocation]);

    const loadLocationHazards = async (locationId) => {
        try {
            const response = await api.get(`/dm-assistant/locations/${locationId}/hazards`);
            setActiveHazards(response.data);
        } catch (error) {
            console.error('Error loading hazards:', error);
        }
    };

    const locationTypes = [
        { value: 'dungeon', label: 'Dungeon', hazards: ['pit trap', 'dart trap', 'collapsing ceiling'] },
        { value: 'wilderness', label: 'Wilderness', hazards: ['quicksand', 'avalanche', 'wild magic'] },
        { value: 'city', label: 'City', hazards: ['crumbling building', 'sewer gas', 'pickpockets'] },
        { value: 'temple', label: 'Temple', hazards: ['divine test', 'cursed altar', 'guardian statue'] },
        { value: 'castle', label: 'Castle', hazards: ['murder holes', 'portcullis trap', 'hidden blades'] }
    ];

    const difficultyLevels = [
        { value: 1, label: 'Trivial', dc: '8-10', damage: '1d4' },
        { value: 3, label: 'Easy', dc: '10-12', damage: '1d6' },
        { value: 5, label: 'Moderate', dc: '13-15', damage: '2d6' },
        { value: 7, label: 'Hard', dc: '16-18', damage: '3d6' },
        { value: 9, label: 'Deadly', dc: '19-21', damage: '4d6+' }
    ];

    const handleGenerateHazard = () => {
        const params = {
            locationType: hazardForm.locationType,
            difficulty: hazardForm.difficulty,
            hazardType: hazardForm.hazardType
        };

        if (currentLocation?.id) {
            params.locationId = currentLocation.id;
        }

        onGenerate('environmental_hazard', params);
    };

    const handleTriggerHazard = async (hazardId) => {
        try {
            await api.post(`/dm-assistant/hazards/${hazardId}/trigger`);
            // Update local state
            setActiveHazards(prev => prev.map(h => 
                h.id === hazardId 
                    ? { ...h, triggeredCount: h.triggeredCount + 1 }
                    : h
            ));
        } catch (error) {
            console.error('Error triggering hazard:', error);
        }
    };

    const getHazardIcon = (hazard) => {
        if (hazard.isTrap) return '🪤';
        if (hazard.isNatural) return '🌿';
        if (hazard.damageFormula?.includes('fire')) return '🔥';
        if (hazard.damageFormula?.includes('cold')) return '❄️';
        if (hazard.damageFormula?.includes('poison')) return '☠️';
        if (hazard.damageFormula?.includes('psychic')) return '🧠';
        return '⚠️';
    };

    const getDifficultyColor = (dc) => {
        if (dc <= 10) return 'green';
        if (dc <= 15) return 'yellow';
        if (dc <= 18) return 'orange';
        return 'red';
    };

    return (
        <div className="environmental-hazards-panel">
            <div className="panel-header">
                <h3>Environmental Hazards</h3>
                {currentLocation && (
                    <span className="current-location">
                        📍 {currentLocation.name}
                    </span>
                )}
            </div>

            {/* Hazard Generator */}
            <div className="hazard-generator">
                <h4>Generate New Hazard</h4>
                
                <div className="generator-form">
                    <div className="form-row">
                        <div className="form-field">
                            <label>Location Type</label>
                            <select
                                value={hazardForm.locationType}
                                onChange={(e) => setHazardForm({...hazardForm, locationType: e.target.value})}
                            >
                                {locationTypes.map(type => (
                                    <option key={type.value} value={type.value}>
                                        {type.label}
                                    </option>
                                ))}
                            </select>
                        </div>

                        <div className="form-field">
                            <label>Hazard Type</label>
                            <select
                                value={hazardForm.hazardType}
                                onChange={(e) => setHazardForm({...hazardForm, hazardType: e.target.value})}
                            >
                                <option value="trap">Trap</option>
                                <option value="natural">Natural Hazard</option>
                                <option value="magical">Magical Effect</option>
                                <option value="environmental">Environmental</option>
                            </select>
                        </div>
                    </div>

                    <div className="difficulty-selector">
                        <label>Difficulty Level: {hazardForm.difficulty}</label>
                        <input
                            type="range"
                            min="1"
                            max="10"
                            value={hazardForm.difficulty}
                            onChange={(e) => setHazardForm({...hazardForm, difficulty: parseInt(e.target.value)})}
                            className="difficulty-slider"
                        />
                        <div className="difficulty-labels">
                            {difficultyLevels.map(level => (
                                <div
                                    key={level.value}
                                    className={`difficulty-label ${hazardForm.difficulty >= level.value ? 'active' : ''}`}
                                >
                                    <span>{level.label}</span>
                                    <small>DC {level.dc}</small>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Quick Hazard Ideas */}
                    <div className="quick-hazards">
                        <label>Quick Ideas:</label>
                        <div className="hazard-chips">
                            {locationTypes
                                .find(t => t.value === hazardForm.locationType)
                                ?.hazards.map(hazard => (
                                    <button
                                        key={hazard}
                                        className="hazard-chip"
                                        onClick={() => {
                                            // This could pre-fill some context
                                            handleGenerateHazard();
                                        }}
                                    >
                                        {hazard}
                                    </button>
                                ))}
                        </div>
                    </div>

                    <button
                        className="generate-hazard-btn"
                        onClick={handleGenerateHazard}
                        disabled={isGenerating}
                    >
                        Generate Hazard
                    </button>
                </div>
            </div>

            {/* Active Hazards */}
            {activeHazards.length > 0 && (
                <div className="active-hazards">
                    <h4>Active Hazards in Location</h4>
                    <div className="hazard-grid">
                        {activeHazards.map(hazard => (
                            <div
                                key={hazard.id}
                                className={`hazard-card ${selectedHazard?.id === hazard.id ? 'selected' : ''} ${!hazard.isActive ? 'inactive' : ''}`}
                                {...getSelectableProps(
                                    () => setSelectedHazard(hazard),
                                    selectedHazard?.id === hazard.id
                                )}
                            >
                                <div className="hazard-header">
                                    <span className="hazard-icon">{getHazardIcon(hazard)}</span>
                                    <h5>{hazard.name}</h5>
                                    <span className={`dc-badge ${getDifficultyColor(hazard.difficultyClass)}`}>
                                        DC {hazard.difficultyClass}
                                    </span>
                                </div>
                                <p className="hazard-desc">{hazard.description}</p>
                                <div className="hazard-stats">
                                    <span className="damage">
                                        <span className="stat-icon">⚔️</span>
                                        {hazard.damageFormula}
                                    </span>
                                    {hazard.triggeredCount > 0 && (
                                        <span className="triggered">
                                            Triggered {hazard.triggeredCount}x
                                        </span>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* Selected Hazard Details */}
            {selectedHazard && (
                <div className="hazard-details">
                    <div className="detail-header">
                        <h4>{selectedHazard.name}</h4>
                        <div className="hazard-badges">
                            {selectedHazard.isTrap && <span className="badge trap">Trap</span>}
                            {selectedHazard.isNatural && <span className="badge natural">Natural</span>}
                            <span className={`badge dc ${getDifficultyColor(selectedHazard.difficultyClass)}`}>
                                DC {selectedHazard.difficultyClass}
                            </span>
                        </div>
                    </div>

                    <div className="detail-sections">
                        <div className="detail-section">
                            <h5>Description</h5>
                            <p>{selectedHazard.description}</p>
                        </div>

                        <div className="detail-section">
                            <h5>Trigger</h5>
                            <p>{selectedHazard.triggerCondition}</p>
                        </div>

                        <div className="detail-section">
                            <h5>Effect</h5>
                            <p>{selectedHazard.effectDescription}</p>
                        </div>

                        {selectedHazard.avoidanceHints && (
                            <div className="detail-section hints">
                                <h5>Detection/Avoidance</h5>
                                <p>{selectedHazard.avoidanceHints}</p>
                            </div>
                        )}

                        <div className="mechanical-effects">
                            <h5>Mechanical Effects</h5>
                            <div className="effect-grid">
                                {selectedHazard.mechanicalEffects.save && (
                                    <div className="effect-item">
                                        <span className="effect-label">Save:</span>
                                        <span>{selectedHazard.mechanicalEffects.save}</span>
                                    </div>
                                )}
                                <div className="effect-item">
                                    <span className="effect-label">DC:</span>
                                    <span>{selectedHazard.difficultyClass}</span>
                                </div>
                                {selectedHazard.damageFormula && (
                                    <div className="effect-item">
                                        <span className="effect-label">Damage:</span>
                                        <span>{selectedHazard.damageFormula}</span>
                                    </div>
                                )}
                                {selectedHazard.mechanicalEffects.damageType && (
                                    <div className="effect-item">
                                        <span className="effect-label">Type:</span>
                                        <span>{selectedHazard.mechanicalEffects.damageType}</span>
                                    </div>
                                )}
                            </div>
                            {selectedHazard.mechanicalEffects.additionalEffects && (
                                <p className="additional-effects">
                                    {selectedHazard.mechanicalEffects.additionalEffects}
                                </p>
                            )}
                        </div>

                        {selectedHazard.resetCondition && (
                            <div className="detail-section">
                                <h5>Reset Condition</h5>
                                <p>{selectedHazard.resetCondition}</p>
                            </div>
                        )}
                    </div>

                    <div className="hazard-actions">
                        <button
                            className="trigger-btn"
                            onClick={() => handleTriggerHazard(selectedHazard.id)}
                            disabled={!selectedHazard.isActive}
                        >
                            Trigger Hazard
                        </button>
                        <button
                            className="copy-btn"
                            onClick={() => {
                                const text = `${selectedHazard.name}\n\n${selectedHazard.description}\n\nTrigger: ${selectedHazard.triggerCondition}\nEffect: ${selectedHazard.effectDescription}\nDC ${selectedHazard.difficultyClass} | Damage: ${selectedHazard.damageFormula}`;
                                navigator.clipboard.writeText(text);
                            }}
                        >
                            Copy to Clipboard
                        </button>
                    </div>
                </div>
            )}

            {/* Hazard Templates */}
            <div className="hazard-templates">
                <h4>Common Hazard Templates</h4>
                <div className="template-grid">
                    <div 
                        className="template-card" 
                        {...getClickableProps(() => {
                            setHazardForm({ locationType: 'dungeon', difficulty: 5, hazardType: 'trap' });
                            handleGenerateHazard();
                        })}>
                        <span className="template-icon">🪤</span>
                        <span>Pit Trap</span>
                    </div>
                    <div 
                        className="template-card" 
                        {...getClickableProps(() => {
                            setHazardForm({ locationType: 'dungeon', difficulty: 7, hazardType: 'trap' });
                            handleGenerateHazard();
                        })}>
                        <span className="template-icon">🎯</span>
                        <span>Dart Trap</span>
                    </div>
                    <div 
                        className="template-card" 
                        {...getClickableProps(() => {
                            setHazardForm({ locationType: 'wilderness', difficulty: 6, hazardType: 'natural' });
                            handleGenerateHazard();
                        })}>
                        <span className="template-icon">🌊</span>
                        <span>Quicksand</span>
                    </div>
                    <div 
                        className="template-card" 
                        {...getClickableProps(() => {
                            setHazardForm({ locationType: 'temple', difficulty: 8, hazardType: 'magical' });
                            handleGenerateHazard();
                        })}>
                        <span className="template-icon">⚡</span>
                        <span>Lightning Ward</span>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default EnvironmentalHazards;