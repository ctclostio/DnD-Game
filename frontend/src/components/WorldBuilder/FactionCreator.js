import React, { useState } from 'react';

const FactionCreator = ({ onClose, onSubmit }) => {
    const [formData, setFormData] = useState({
        name: '',
        type: 'political',
        description: '',
        publicGoals: [''],
        secretGoals: [''],
        leadershipStructure: '',
        headquartersLocation: '',
        memberCount: 100,
        influenceLevel: 5,
        militaryStrength: 5,
        economicPower: 5,
        magicalResources: 5,
        ancientKnowledgeLevel: 0,
        corrupted: false,
        seeksAncientPower: false,
        guardsAncientSecrets: false,
        foundingDate: ''
    });

    const factionTypes = [
        { value: 'religious', label: 'Religious', icon: '⛪' },
        { value: 'political', label: 'Political', icon: '👑' },
        { value: 'criminal', label: 'Criminal', icon: '🗡️' },
        { value: 'merchant', label: 'Merchant', icon: '💰' },
        { value: 'military', label: 'Military', icon: '⚔️' },
        { value: 'cult', label: 'Cult', icon: '🌑' },
        { value: 'ancient_order', label: 'Ancient Order', icon: '🗿' }
    ];

    const handleArrayChange = (field, index, value) => {
        setFormData(prev => ({
            ...prev,
            [field]: prev[field].map((item, i) => i === index ? value : item)
        }));
    };

    const addArrayItem = (field) => {
        setFormData(prev => ({
            ...prev,
            [field]: [...prev[field], '']
        }));
    };

    const removeArrayItem = (field, index) => {
        setFormData(prev => ({
            ...prev,
            [field]: prev[field].filter((_, i) => i !== index)
        }));
    };

    const handleSubmit = (e) => {
        e.preventDefault();
        
        // Filter out empty goals
        const cleanedData = {
            ...formData,
            publicGoals: formData.publicGoals.filter(g => g.trim()),
            secretGoals: formData.secretGoals.filter(g => g.trim())
        };
        
        onSubmit(cleanedData);
    };

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal-content faction-creator" onClick={(e) => e.stopPropagation()}>
                <div className="modal-header">
                    <h3>Create New Faction</h3>
                    <button className="close-button" onClick={onClose}>×</button>
                </div>

                <form onSubmit={handleSubmit}>
                    <div className="form-row">
                        <div className="form-group">
                            <label>Faction Name*</label>
                            <input
                                type="text"
                                value={formData.name}
                                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                                placeholder="e.g., The Order of the Eternal Flame"
                                required
                            />
                        </div>

                        <div className="form-group">
                            <label>Faction Type</label>
                            <div className="type-selector">
                                {factionTypes.map(type => (
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
                    </div>

                    <div className="form-group">
                        <label>Description*</label>
                        <textarea
                            value={formData.description}
                            onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                            placeholder="Describe the faction's history, purpose, and role in the world..."
                            rows={4}
                            required
                        />
                    </div>

                    <div className="form-row">
                        <div className="form-group">
                            <label>Public Goals</label>
                            {formData.publicGoals.map((goal, index) => (
                                <div key={index} className="array-input">
                                    <input
                                        type="text"
                                        value={goal}
                                        onChange={(e) => handleArrayChange('publicGoals', index, e.target.value)}
                                        placeholder="What this faction openly claims to pursue"
                                    />
                                    {formData.publicGoals.length > 1 && (
                                        <button
                                            type="button"
                                            onClick={() => removeArrayItem('publicGoals', index)}
                                            className="remove-button"
                                        >
                                            ×
                                        </button>
                                    )}
                                </div>
                            ))}
                            <button
                                type="button"
                                onClick={() => addArrayItem('publicGoals')}
                                className="add-button"
                            >
                                + Add Goal
                            </button>
                        </div>

                        <div className="form-group">
                            <label>Secret Goals (DM Only)</label>
                            {formData.secretGoals.map((goal, index) => (
                                <div key={index} className="array-input">
                                    <input
                                        type="text"
                                        value={goal}
                                        onChange={(e) => handleArrayChange('secretGoals', index, e.target.value)}
                                        placeholder="Hidden agendas and true motivations"
                                    />
                                    {formData.secretGoals.length > 1 && (
                                        <button
                                            type="button"
                                            onClick={() => removeArrayItem('secretGoals', index)}
                                            className="remove-button"
                                        >
                                            ×
                                        </button>
                                    )}
                                </div>
                            ))}
                            <button
                                type="button"
                                onClick={() => addArrayItem('secretGoals')}
                                className="add-button"
                            >
                                + Add Secret Goal
                            </button>
                        </div>
                    </div>

                    <div className="form-row">
                        <div className="form-group">
                            <label>Leadership Structure*</label>
                            <input
                                type="text"
                                value={formData.leadershipStructure}
                                onChange={(e) => setFormData(prev => ({ ...prev, leadershipStructure: e.target.value }))}
                                placeholder="e.g., Council of Elders, Merchant Prince, High Priest"
                                required
                            />
                        </div>

                        <div className="form-group">
                            <label>Headquarters Location</label>
                            <input
                                type="text"
                                value={formData.headquartersLocation}
                                onChange={(e) => setFormData(prev => ({ ...prev, headquartersLocation: e.target.value }))}
                                placeholder="Where is their main base of operations?"
                            />
                        </div>
                    </div>

                    <div className="form-row">
                        <div className="form-group">
                            <label>Member Count</label>
                            <input
                                type="number"
                                value={formData.memberCount}
                                onChange={(e) => setFormData(prev => ({ 
                                    ...prev, 
                                    memberCount: parseInt(e.target.value) || 0 
                                }))}
                                min="1"
                            />
                        </div>

                        <div className="form-group">
                            <label>Founding Date</label>
                            <input
                                type="text"
                                value={formData.foundingDate}
                                onChange={(e) => setFormData(prev => ({ ...prev, foundingDate: e.target.value }))}
                                placeholder="e.g., 200 years ago, During the Age of Shadows"
                            />
                        </div>
                    </div>

                    <div className="power-stats">
                        <h4>Faction Power Levels</h4>
                        
                        <div className="stat-slider">
                            <label>Influence Level: {formData.influenceLevel}/10</label>
                            <input
                                type="range"
                                min="0"
                                max="10"
                                value={formData.influenceLevel}
                                onChange={(e) => setFormData(prev => ({ 
                                    ...prev, 
                                    influenceLevel: parseInt(e.target.value) 
                                }))}
                            />
                        </div>

                        <div className="stat-slider">
                            <label>Military Strength: {formData.militaryStrength}/10</label>
                            <input
                                type="range"
                                min="0"
                                max="10"
                                value={formData.militaryStrength}
                                onChange={(e) => setFormData(prev => ({ 
                                    ...prev, 
                                    militaryStrength: parseInt(e.target.value) 
                                }))}
                            />
                        </div>

                        <div className="stat-slider">
                            <label>Economic Power: {formData.economicPower}/10</label>
                            <input
                                type="range"
                                min="0"
                                max="10"
                                value={formData.economicPower}
                                onChange={(e) => setFormData(prev => ({ 
                                    ...prev, 
                                    economicPower: parseInt(e.target.value) 
                                }))}
                            />
                        </div>

                        <div className="stat-slider">
                            <label>Magical Resources: {formData.magicalResources}/10</label>
                            <input
                                type="range"
                                min="0"
                                max="10"
                                value={formData.magicalResources}
                                onChange={(e) => setFormData(prev => ({ 
                                    ...prev, 
                                    magicalResources: parseInt(e.target.value) 
                                }))}
                            />
                        </div>

                        <div className="stat-slider ancient">
                            <label>Ancient Knowledge: {formData.ancientKnowledgeLevel}/10</label>
                            <input
                                type="range"
                                min="0"
                                max="10"
                                value={formData.ancientKnowledgeLevel}
                                onChange={(e) => setFormData(prev => ({ 
                                    ...prev, 
                                    ancientKnowledgeLevel: parseInt(e.target.value) 
                                }))}
                            />
                        </div>
                    </div>

                    <div className="ancient-connections">
                        <h4>Ancient Connections</h4>
                        
                        <label className="checkbox-label">
                            <input
                                type="checkbox"
                                checked={formData.corrupted}
                                onChange={(e) => setFormData(prev => ({ 
                                    ...prev, 
                                    corrupted: e.target.checked 
                                }))}
                            />
                            <span>🌑 Corrupted</span>
                            <span className="hint">Touched by ancient darkness</span>
                        </label>

                        <label className="checkbox-label">
                            <input
                                type="checkbox"
                                checked={formData.seeksAncientPower}
                                onChange={(e) => setFormData(prev => ({ 
                                    ...prev, 
                                    seeksAncientPower: e.target.checked 
                                }))}
                            />
                            <span>🗿 Seeks Ancient Power</span>
                            <span className="hint">Actively pursuing forgotten knowledge</span>
                        </label>

                        <label className="checkbox-label">
                            <input
                                type="checkbox"
                                checked={formData.guardsAncientSecrets}
                                onChange={(e) => setFormData(prev => ({ 
                                    ...prev, 
                                    guardsAncientSecrets: e.target.checked 
                                }))}
                            />
                            <span>🛡️ Guards Ancient Secrets</span>
                            <span className="hint">Protects forbidden knowledge</span>
                        </label>
                    </div>

                    <div className="creator-actions">
                        <button type="button" onClick={onClose}>
                            Cancel
                        </button>
                        <button 
                            type="submit" 
                            className="btn btn-primary"
                            disabled={!formData.name || !formData.description || !formData.leadershipStructure}
                        >
                            Create Faction
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default FactionCreator;