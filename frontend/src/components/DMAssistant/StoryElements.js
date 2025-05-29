import React, { useState } from 'react';

const StoryElements = ({ gameSessionId, storyElements, onGenerate, onUseElement, isGenerating }) => {
    const [currentContext, setCurrentContext] = useState({
        mainPlot: '',
        recentEvents: '',
        keyNPCs: '',
        playerGoals: ''
    });
    const [filter, setFilter] = useState('all');
    const [selectedElement, setSelectedElement] = useState(null);

    const elementTypes = [
        { value: 'plot_twist', label: 'Plot Twist', icon: '🎭', color: 'purple' },
        { value: 'story_hook', label: 'Story Hook', icon: '🎣', color: 'blue' },
        { value: 'revelation', label: 'Revelation', icon: '💡', color: 'yellow' },
        { value: 'complication', label: 'Complication', icon: '🌪️', color: 'red' }
    ];

    const impactLevels = {
        'minor': { label: 'Minor', color: 'green' },
        'moderate': { label: 'Moderate', color: 'yellow' },
        'major': { label: 'Major', color: 'orange' },
        'campaign-changing': { label: 'Campaign Changing', color: 'red' }
    };

    const handleGenerateElement = (type) => {
        const context = {
            mainPlot: currentContext.mainPlot,
            recentEvents: currentContext.recentEvents,
            keyNPCs: currentContext.keyNPCs,
            playerGoals: currentContext.playerGoals,
            sessionId: gameSessionId
        };

        onGenerate(type, {}, context);
    };

    const filteredElements = filter === 'all' 
        ? storyElements 
        : storyElements.filter(el => el.type === filter);

    const unusedElements = filteredElements.filter(el => !el.used);
    const usedElements = filteredElements.filter(el => el.used);

    return (
        <div className="story-elements-panel">
            <div className="panel-header">
                <h3>Story Elements & Plot Twists</h3>
                <div className="element-stats">
                    <span className="stat">
                        <strong>{unusedElements.length}</strong> Available
                    </span>
                    <span className="stat">
                        <strong>{usedElements.length}</strong> Used
                    </span>
                </div>
            </div>

            {/* Campaign Context */}
            <div className="campaign-context">
                <h4>Current Campaign Context</h4>
                <p className="context-help">
                    Provide context to generate more relevant story elements
                </p>
                
                <div className="context-form">
                    <div className="context-field">
                        <label>Main Plot/Quest</label>
                        <textarea
                            placeholder="e.g., The party is investigating disappearances in the northern villages..."
                            value={currentContext.mainPlot}
                            onChange={(e) => setCurrentContext({...currentContext, mainPlot: e.target.value})}
                            rows={2}
                        />
                    </div>

                    <div className="context-field">
                        <label>Recent Events</label>
                        <textarea
                            placeholder="e.g., Just discovered the mayor is involved, fought cultists in the woods..."
                            value={currentContext.recentEvents}
                            onChange={(e) => setCurrentContext({...currentContext, recentEvents: e.target.value})}
                            rows={2}
                        />
                    </div>

                    <div className="context-row">
                        <div className="context-field">
                            <label>Key NPCs</label>
                            <input
                                type="text"
                                placeholder="e.g., Mayor Blackwood, Sage Elara, Captain Morris"
                                value={currentContext.keyNPCs}
                                onChange={(e) => setCurrentContext({...currentContext, keyNPCs: e.target.value})}
                            />
                        </div>

                        <div className="context-field">
                            <label>Player Goals</label>
                            <input
                                type="text"
                                placeholder="e.g., Find missing villagers, expose corruption"
                                value={currentContext.playerGoals}
                                onChange={(e) => setCurrentContext({...currentContext, playerGoals: e.target.value})}
                            />
                        </div>
                    </div>
                </div>

                {/* Generate Buttons */}
                <div className="generate-elements">
                    {elementTypes.map(type => (
                        <button
                            key={type.value}
                            className={`generate-element-btn ${type.color}`}
                            onClick={() => handleGenerateElement(type.value)}
                            disabled={isGenerating}
                        >
                            <span className="element-icon">{type.icon}</span>
                            Generate {type.label}
                        </button>
                    ))}
                </div>
            </div>

            {/* Filter Tabs */}
            <div className="element-filters">
                <button
                    className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
                    onClick={() => setFilter('all')}
                >
                    All Elements
                </button>
                {elementTypes.map(type => (
                    <button
                        key={type.value}
                        className={`filter-btn ${filter === type.value ? 'active' : ''}`}
                        onClick={() => setFilter(type.value)}
                    >
                        {type.icon} {type.label}
                    </button>
                ))}
            </div>

            {/* Story Elements List */}
            <div className="story-elements-container">
                {unusedElements.length > 0 && (
                    <div className="element-section">
                        <h4>Available Story Elements</h4>
                        <div className="element-grid">
                            {unusedElements.map(element => (
                                <div
                                    key={element.id}
                                    className={`story-element-card ${element.type} ${selectedElement?.id === element.id ? 'selected' : ''}`}
                                    onClick={() => setSelectedElement(element)}
                                >
                                    <div className="element-header">
                                        <span className="element-type">
                                            {elementTypes.find(t => t.value === element.type)?.icon}
                                        </span>
                                        <h5>{element.title}</h5>
                                        <span className={`impact-badge ${element.impactLevel}`}>
                                            {impactLevels[element.impactLevel]?.label}
                                        </span>
                                    </div>
                                    <p className="element-preview">
                                        {element.description.substring(0, 100)}...
                                    </p>
                                    {element.suggestedTiming && (
                                        <div className="timing-hint">
                                            <span className="timing-icon">⏰</span>
                                            {element.suggestedTiming}
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {usedElements.length > 0 && (
                    <div className="element-section used">
                        <h4>Used Story Elements</h4>
                        <div className="element-list">
                            {usedElements.map(element => (
                                <div key={element.id} className="used-element">
                                    <span className="element-type">
                                        {elementTypes.find(t => t.value === element.type)?.icon}
                                    </span>
                                    <span className="element-title">{element.title}</span>
                                    <span className="used-date">
                                        Used: {new Date(element.usedAt).toLocaleDateString()}
                                    </span>
                                </div>
                            ))}
                        </div>
                    </div>
                )}
            </div>

            {/* Selected Element Details */}
            {selectedElement && (
                <div className="element-details-modal">
                    <div className="modal-content">
                        <div className="modal-header">
                            <h3>{selectedElement.title}</h3>
                            <button 
                                className="close-btn"
                                onClick={() => setSelectedElement(null)}
                            >
                                ×
                            </button>
                        </div>

                        <div className="element-full-details">
                            <div className="detail-row">
                                <span className="detail-label">Type:</span>
                                <span className="detail-value">
                                    {elementTypes.find(t => t.value === selectedElement.type)?.label}
                                </span>
                            </div>

                            <div className="detail-row">
                                <span className="detail-label">Impact:</span>
                                <span className={`impact-badge ${selectedElement.impactLevel}`}>
                                    {impactLevels[selectedElement.impactLevel]?.label}
                                </span>
                            </div>

                            <div className="detail-section">
                                <h4>Description</h4>
                                <p>{selectedElement.description}</p>
                            </div>

                            {selectedElement.suggestedTiming && (
                                <div className="detail-section">
                                    <h4>Suggested Timing</h4>
                                    <p>{selectedElement.suggestedTiming}</p>
                                </div>
                            )}

                            {selectedElement.prerequisites && selectedElement.prerequisites.length > 0 && (
                                <div className="detail-section">
                                    <h4>Prerequisites</h4>
                                    <ul>
                                        {selectedElement.prerequisites.map((prereq, idx) => (
                                            <li key={idx}>{prereq}</li>
                                        ))}
                                    </ul>
                                </div>
                            )}

                            {selectedElement.consequences && selectedElement.consequences.length > 0 && (
                                <div className="detail-section">
                                    <h4>Potential Consequences</h4>
                                    <ul>
                                        {selectedElement.consequences.map((consequence, idx) => (
                                            <li key={idx}>{consequence}</li>
                                        ))}
                                    </ul>
                                </div>
                            )}

                            {selectedElement.foreshadowingHints && selectedElement.foreshadowingHints.length > 0 && (
                                <div className="detail-section">
                                    <h4>Foreshadowing Hints</h4>
                                    <ul className="foreshadowing-list">
                                        {selectedElement.foreshadowingHints.map((hint, idx) => (
                                            <li key={idx}>
                                                <span className="hint-icon">💭</span>
                                                {hint}
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            )}

                            {!selectedElement.used && (
                                <div className="element-actions">
                                    <button
                                        className="use-element-btn"
                                        onClick={() => {
                                            onUseElement(selectedElement.id);
                                            setSelectedElement(null);
                                        }}
                                    >
                                        Mark as Used
                                    </button>
                                    <button
                                        className="copy-btn"
                                        onClick={() => {
                                            navigator.clipboard.writeText(
                                                `${selectedElement.title}\n\n${selectedElement.description}`
                                            );
                                        }}
                                    >
                                        Copy to Clipboard
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default StoryElements;