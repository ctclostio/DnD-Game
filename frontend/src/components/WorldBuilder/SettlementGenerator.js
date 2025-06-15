import React, { useState } from 'react';
import { getClickableProps, getSelectableProps } from '../../utils/accessibility';

const SettlementGenerator = ({ onGenerate, onClose, isGenerating }) => {
    const [formData, setFormData] = useState({
        name: '',
        type: 'village',
        region: '',
        populationSize: 'medium',
        dangerLevel: 3,
        ancientInfluence: false,
        specialFeatures: []
    });

    const settlementTypes = [
        { value: 'hamlet', label: 'Hamlet', icon: '🏘️' },
        { value: 'village', label: 'Village', icon: '🏡' },
        { value: 'town', label: 'Town', icon: '🏛️' },
        { value: 'city', label: 'City', icon: '🏰' },
        { value: 'metropolis', label: 'Metropolis', icon: '🌆' },
        { value: 'ruins', label: 'Ruins', icon: '🏚️' }
    ];

    const specialFeatureOptions = [
        'Major Trade Hub',
        'Military Outpost',
        'Religious Center',
        'Magical Academy',
        'Criminal Underground',
        'Ancient Library',
        'Portal Nexus',
        'Cursed Ground',
        'Ley Line Convergence',
        'Sealed Evil'
    ];

    const handleSubmit = (e) => {
        e.preventDefault();
        onGenerate(formData);
    };

    const toggleSpecialFeature = (feature) => {
        setFormData(prev => ({
            ...prev,
            specialFeatures: prev.specialFeatures.includes(feature)
                ? prev.specialFeatures.filter(f => f !== feature)
                : [...prev.specialFeatures, feature]
        }));
    };

    return (
        <div className="modal-overlay" {...getClickableProps(onClose)}>
            <div className="modal-content settlement-generator" {...getClickableProps((e) => e.stopPropagation())}>
                <div className="modal-header">
                    <h3>Generate Settlement</h3>
                    <button className="close-button" onClick={onClose}>×</button>
                </div>

                <form onSubmit={handleSubmit}>
                    <div className="form-group">
                        <label>Settlement Name (Optional)</label>
                        <input
                            type="text"
                            value={formData.name}
                            onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                            placeholder="Leave blank for AI-generated name"
                        />
                    </div>

                    <div className="form-group">
                        <label>Settlement Type</label>
                        <div className="type-selector">
                            {settlementTypes.map(type => (
                                <button
                                    key={type.value}
                                    type="button"
                                    className={`type-option ${formData.type === type.value ? 'selected' : ''}`}
                                    onClick={() => setFormData(prev => ({ ...prev, type: type.value }))}
                                >
                                    <span className="type-icon">{type.icon}</span>
                                    <span className="type-label">{type.label}</span>
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="form-group">
                        <label>Region</label>
                        <input
                            type="text"
                            value={formData.region}
                            onChange={(e) => setFormData(prev => ({ ...prev, region: e.target.value }))}
                            placeholder="e.g., Northern Mountains, Shadowfen Swamp"
                            required
                        />
                    </div>

                    <div className="form-row">
                        <div className="form-group">
                            <label>Population Size</label>
                            <select
                                value={formData.populationSize}
                                onChange={(e) => setFormData(prev => ({ ...prev, populationSize: e.target.value }))}
                            >
                                <option value="small">Small</option>
                                <option value="medium">Medium</option>
                                <option value="large">Large</option>
                            </select>
                        </div>

                        <div className="form-group">
                            <label>Danger Level: {formData.dangerLevel}/10</label>
                            <input
                                type="range"
                                min="1"
                                max="10"
                                value={formData.dangerLevel}
                                onChange={(e) => setFormData(prev => ({ 
                                    ...prev, 
                                    dangerLevel: parseInt(e.target.value) 
                                }))}
                            />
                        </div>
                    </div>

                    <div className="form-group">
                        <label className="checkbox-label">
                            <input
                                type="checkbox"
                                checked={formData.ancientInfluence}
                                onChange={(e) => setFormData(prev => ({ 
                                    ...prev, 
                                    ancientInfluence: e.target.checked 
                                }))}
                            />
                            <span>Strong Ancient Influence</span>
                            <span className="hint">
                                Settlement will have connections to the old world, 
                                ancient ruins, or eldritch influences
                            </span>
                        </label>
                    </div>

                    <div className="form-group">
                        <label>Special Features</label>
                        <div className="feature-grid">
                            {specialFeatureOptions.map(feature => (
                                <button
                                    key={feature}
                                    type="button"
                                    className={`feature-option ${
                                        formData.specialFeatures.includes(feature) ? 'selected' : ''
                                    }`}
                                    onClick={() => toggleSpecialFeature(feature)}
                                >
                                    {feature}
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="generator-actions">
                        <button type="button" onClick={onClose} disabled={isGenerating}>
                            Cancel
                        </button>
                        <button 
                            type="submit" 
                            className="btn btn-primary"
                            disabled={isGenerating || !formData.region}
                        >
                            {isGenerating ? 'Generating...' : 'Generate Settlement'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default SettlementGenerator;