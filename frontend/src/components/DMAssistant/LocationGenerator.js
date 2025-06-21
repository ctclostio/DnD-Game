import React, { useState } from 'react';

const LocationGenerator = ({ gameSessionId, savedLocations, onGenerate, isGenerating }) => {
    const [locationForm, setLocationForm] = useState({
        type: 'tavern',
        name: '',
        atmosphere: '',
        timeOfDay: 'day',
        weather: 'clear',
        specialFeatures: []
    });
    const [customFeature, setCustomFeature] = useState('');
    const [selectedLocation, setSelectedLocation] = useState(null);

    const locationTypes = [
        { value: 'tavern', label: 'Tavern', icon: '🍺' },
        { value: 'dungeon', label: 'Dungeon', icon: '🏚️' },
        { value: 'shop', label: 'Shop', icon: '🏪' },
        { value: 'wilderness', label: 'Wilderness', icon: '🌲' },
        { value: 'city', label: 'City', icon: '🏛️' },
        { value: 'temple', label: 'Temple', icon: '⛪' },
        { value: 'castle', label: 'Castle', icon: '🏰' }
    ];

    const atmospheres = [
        'Mysterious and foreboding',
        'Warm and welcoming',
        'Bustling and chaotic',
        'Quiet and serene',
        'Dark and dangerous',
        'Magical and wondrous',
        'Ancient and crumbling',
        'Opulent and luxurious'
    ];

    const weatherOptions = [
        'Clear', 'Rainy', 'Stormy', 'Foggy', 'Snowy', 'Windy', 'Hot', 'Cold'
    ];

    const timeOptions = [
        { value: 'dawn', label: 'Dawn' },
        { value: 'morning', label: 'Morning' },
        { value: 'day', label: 'Day' },
        { value: 'dusk', label: 'Dusk' },
        { value: 'night', label: 'Night' },
        { value: 'midnight', label: 'Midnight' }
    ];

    const quickFeatures = {
        tavern: ['Fighting pit', 'Secret gambling den', 'Mysterious patron', 'Live entertainment'],
        dungeon: ['Trap-filled corridor', 'Ancient altar', 'Prison cells', 'Hidden treasure'],
        shop: ['Rare magical items', 'Eccentric shopkeeper', 'Black market goods', 'Cursed items'],
        wilderness: ['Ancient ruins', 'Bandit camp', 'Sacred grove', 'Monster lair'],
        city: ['Thieves guild', 'Noble district', 'Market square', 'Sewers'],
        temple: ['Sacred relic', 'Healing fountain', 'Library', 'Catacombs'],
        castle: ['Throne room', 'Dungeon', 'Secret passages', 'Treasury']
    };

    const handleGenerateLocation = () => {
        onGenerate('location_description', {
            locationType: locationForm.type,
            locationName: locationForm.name || `${locationForm.type} location`,
            atmosphere: locationForm.atmosphere,
            specialFeatures: locationForm.specialFeatures,
            timeOfDay: locationForm.timeOfDay,
            weather: locationForm.weather
        });
    };

    const addFeature = (feature) => {
        if (!locationForm.specialFeatures.includes(feature)) {
            setLocationForm({
                ...locationForm,
                specialFeatures: [...locationForm.specialFeatures, feature]
            });
        }
    };

    const removeFeature = (feature) => {
        setLocationForm({
            ...locationForm,
            specialFeatures: locationForm.specialFeatures.filter(f => f !== feature)
        });
    };

    const addCustomFeature = () => {
        if (customFeature.trim()) {
            addFeature(customFeature.trim());
            setCustomFeature('');
        }
    };

    return (
        <div className="location-generator-panel">
            <div className="panel-header">
                <h3>Location Generator</h3>
            </div>

            <div className="location-form">
                <div className="location-type-selector">
                    <label>Location Type</label>
                    <div className="type-grid">
                        {locationTypes.map(type => (
                            <button
                                key={type.value}
                                className={`type-btn ${locationForm.type === type.value ? 'selected' : ''}`}
                                onClick={() => setLocationForm({...locationForm, type: type.value})}
                            >
                                <span className="type-icon">{type.icon}</span>
                                <span className="type-label">{type.label}</span>
                            </button>
                        ))}
                    </div>
                </div>

                <div className="form-row">
                    <div className="form-field">
                        <label>Location Name (optional)</label>
                        <input
                            type="text"
                            placeholder={`Name your ${locationForm.type}...`}
                            value={locationForm.name}
                            onChange={(e) => setLocationForm({...locationForm, name: e.target.value})}
                        />
                    </div>
                </div>

                <div className="form-row">
                    <div className="form-field">
                        <label>Atmosphere</label>
                        <select
                            value={locationForm.atmosphere}
                            onChange={(e) => setLocationForm({...locationForm, atmosphere: e.target.value})}
                        >
                            <option value="">Select atmosphere...</option>
                            {atmospheres.map(atm => (
                                <option key={atm} value={atm}>{atm}</option>
                            ))}
                        </select>
                    </div>

                    <div className="form-field">
                        <label>Time of Day</label>
                        <select
                            value={locationForm.timeOfDay}
                            onChange={(e) => setLocationForm({...locationForm, timeOfDay: e.target.value})}
                        >
                            {timeOptions.map(time => (
                                <option key={time.value} value={time.value}>{time.label}</option>
                            ))}
                        </select>
                    </div>

                    <div className="form-field">
                        <label>Weather</label>
                        <select
                            value={locationForm.weather}
                            onChange={(e) => setLocationForm({...locationForm, weather: e.target.value})}
                        >
                            {weatherOptions.map(weather => (
                                <option key={weather} value={weather}>{weather}</option>
                            ))}
                        </select>
                    </div>
                </div>

                <div className="special-features">
                    <label>Special Features</label>
                    <div className="feature-suggestions">
                        {quickFeatures[locationForm.type]?.map(feature => (
                            <button
                                key={feature}
                                className="feature-chip"
                                onClick={() => addFeature(feature)}
                                disabled={locationForm.specialFeatures.includes(feature)}
                            >
                                + {feature}
                            </button>
                        ))}
                    </div>

                    <div className="custom-feature-input">
                        <input
                            type="text"
                            placeholder="Add custom feature..."
                            value={customFeature}
                            onChange={(e) => setCustomFeature(e.target.value)}
                            onKeyPress={(e) => e.key === 'Enter' && addCustomFeature()}
                        />
                        <button onClick={addCustomFeature}>Add</button>
                    </div>

                    {locationForm.specialFeatures.length > 0 && (
                        <div className="selected-features">
                            {locationForm.specialFeatures.map(feature => (
                                <div key={feature} className="feature-tag">
                                    {feature}
                                    <button 
                                        className="remove-btn"
                                        onClick={() => removeFeature(feature)}
                                    >
                                        ×
                                    </button>
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                <button
                    className="generate-location-btn"
                    onClick={handleGenerateLocation}
                    disabled={isGenerating}
                >
                    Generate Location
                </button>
            </div>

            {/* Saved Locations */}
            <div className="saved-locations">
                <h4>Recent Locations</h4>
                <div className="location-list">
                    {savedLocations.map(location => (
                        <div
                            key={location.id}
                            className={`location-card ${selectedLocation?.id === location.id ? 'selected' : ''}`}
                            {...getSelectableProps(() => setSelectedLocation(location), selectedLocation?.id === location.id)}
                        >
                            <div className="location-header">
                                <h5>{location.name}</h5>
                                <span className="location-type">{location.type}</span>
                            </div>
                            <p className="location-description">{location.description}</p>
                            {location.notableFeatures && location.notableFeatures.length > 0 && (
                                <div className="location-features">
                                    {location.notableFeatures.map((feature, idx) => (
                                        <span key={idx} className="feature-badge">{feature}</span>
                                    ))}
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            </div>

            {/* Selected Location Details */}
            {selectedLocation && (
                <div className="location-details">
                    <h4>{selectedLocation.name}</h4>
                    <div className="detail-section">
                        <h5>Description</h5>
                        <p>{selectedLocation.description}</p>
                    </div>
                    
                    {selectedLocation.atmosphere && (
                        <div className="detail-section">
                            <h5>Atmosphere</h5>
                            <p>{selectedLocation.atmosphere}</p>
                        </div>
                    )}

                    {selectedLocation.availableActions && selectedLocation.availableActions.length > 0 && (
                        <div className="detail-section">
                            <h5>Available Actions</h5>
                            <ul>
                                {selectedLocation.availableActions.map((action, idx) => (
                                    <li key={idx}>{action}</li>
                                ))}
                            </ul>
                        </div>
                    )}

                    {selectedLocation.secretsAndHidden && selectedLocation.secretsAndHidden.length > 0 && (
                        <div className="detail-section secrets">
                            <h5>🔒 Secrets & Hidden Elements</h5>
                            {selectedLocation.secretsAndHidden.map((secret, idx) => (
                                <div key={idx} className="secret-item">
                                    <p>{secret.description}</p>
                                    <span className="dc">DC {secret.discoveryDC} to discover</span>
                                    {secret.discoveryHint && (
                                        <p className="hint">Hint: {secret.discoveryHint}</p>
                                    )}
                                </div>
                            ))}
                        </div>
                    )}

                    <div className="location-actions">
                        <button onClick={() => {
                            navigator.clipboard.writeText(
                                `${selectedLocation.name}\n\n${selectedLocation.description}`
                            );
                        }}>
                            Copy Description
                        </button>
                        <button onClick={() => {
                            onGenerate('environmental_hazard', {
                                locationType: selectedLocation.type,
                                locationId: selectedLocation.id,
                                difficulty: 5
                            });
                        }}>
                            Add Hazard
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default LocationGenerator;