import React, { useState, useEffect } from 'react';
import { getClickableProps, getSelectableProps } from '../../utils/accessibility';

const GeneratedContent = ({ content, onReuse }) => {
    const [filter, setFilter] = useState('all');
    const [sortBy, setSortBy] = useState('newest');
    const [selectedItem, setSelectedItem] = useState(null);

    // Content type mapping for icons and labels
    const contentTypes = {
        npc_creation: { icon: '👤', label: 'NPC', color: 'blue' },
        npc_dialogue: { icon: '💬', label: 'Dialogue', color: 'green' },
        location_description: { icon: '📍', label: 'Location', color: 'purple' },
        combat_narration: { icon: '⚔️', label: 'Combat', color: 'red' },
        death_description: { icon: '💀', label: 'Death', color: 'black' },
        plot_twist: { icon: '🎭', label: 'Plot Twist', color: 'orange' },
        story_hook: { icon: '🎣', label: 'Story Hook', color: 'teal' },
        revelation: { icon: '💡', label: 'Revelation', color: 'yellow' },
        complication: { icon: '🌪️', label: 'Complication', color: 'pink' },
        environmental_hazard: { icon: '⚠️', label: 'Hazard', color: 'brown' }
    };

    // Filter content based on selected filter
    const filteredContent = filter === 'all' 
        ? content 
        : content.filter(item => item.type === filter);

    // Sort content based on selected sort option
    const sortedContent = [...filteredContent].sort((a, b) => {
        if (sortBy === 'newest') {
            return new Date(b.createdAt) - new Date(a.createdAt);
        } else if (sortBy === 'oldest') {
            return new Date(a.createdAt) - new Date(b.createdAt);
        } else if (sortBy === 'type') {
            return a.type.localeCompare(b.type);
        }
        return 0;
    });

    // Group content by session or date
    const groupContentByDate = (content) => {
        const groups = {};
        content.forEach(item => {
            const date = new Date(item.createdAt).toDateString();
            if (!groups[date]) {
                groups[date] = [];
            }
            groups[date].push(item);
        });
        return groups;
    };

    const groupedContent = groupContentByDate(sortedContent);

    const formatTimestamp = (timestamp) => {
        const date = new Date(timestamp);
        const now = new Date();
        const diffInHours = (now - date) / (1000 * 60 * 60);
        
        if (diffInHours < 1) {
            const diffInMinutes = Math.floor((now - date) / (1000 * 60));
            return `${diffInMinutes} minutes ago`;
        } else if (diffInHours < 24) {
            return `${Math.floor(diffInHours)} hours ago`;
        } else {
            return date.toLocaleString();
        }
    };

    const getContentPreview = (item) => {
        if (typeof item.data === 'string') {
            return item.data.substring(0, 150) + '...';
        } else if (item.data.description) {
            return item.data.description.substring(0, 150) + '...';
        } else if (item.data.dialogue) {
            return item.data.dialogue.substring(0, 150) + '...';
        } else if (item.data.narration) {
            return item.data.narration.substring(0, 150) + '...';
        }
        return 'No preview available';
    };

    const getContentTitle = (item) => {
        if (item.data.name) return item.data.name;
        if (item.data.title) return item.data.title;
        if (item.data.attackerName && item.data.targetName) {
            return `${item.data.attackerName} vs ${item.data.targetName}`;
        }
        return contentTypes[item.type]?.label || 'Content';
    };

    return (
        <div className="generated-content-panel">
            <div className="panel-header">
                <h3>Generated Content History</h3>
                <div className="content-stats">
                    <span>{content.length} total items</span>
                </div>
            </div>

            {/* Filters and Sort Options */}
            <div className="content-controls">
                <div className="filter-tabs">
                    <button
                        className={`filter-tab ${filter === 'all' ? 'active' : ''}`}
                        onClick={() => setFilter('all')}
                    >
                        All Content
                    </button>
                    {Object.entries(contentTypes).map(([type, config]) => (
                        <button
                            key={type}
                            className={`filter-tab ${filter === type ? 'active' : ''}`}
                            onClick={() => setFilter(type)}
                        >
                            {config.icon} {config.label}
                        </button>
                    ))}
                </div>

                <div className="sort-options">
                    <label>Sort by:</label>
                    <select value={sortBy} onChange={(e) => setSortBy(e.target.value)}>
                        <option value="newest">Newest First</option>
                        <option value="oldest">Oldest First</option>
                        <option value="type">By Type</option>
                    </select>
                </div>
            </div>

            {/* Content List */}
            <div className="content-timeline">
                {Object.entries(groupedContent).map(([date, items]) => (
                    <div key={date} className="date-group">
                        <h4 className="date-header">{date}</h4>
                        <div className="content-items">
                            {items.map(item => (
                                <div
                                    key={item.id}
                                    className={`content-item ${item.type} ${selectedItem?.id === item.id ? 'selected' : ''}`}
                                    {...getSelectableProps(
                                        () => setSelectedItem(item),
                                        selectedItem?.id === item.id
                                    )}
                                >
                                    <div className="item-header">
                                        <span className={`type-icon ${contentTypes[item.type]?.color}`}>
                                            {contentTypes[item.type]?.icon}
                                        </span>
                                        <h5>{getContentTitle(item)}</h5>
                                        <span className="timestamp">
                                            {formatTimestamp(item.createdAt)}
                                        </span>
                                    </div>
                                    <p className="item-preview">
                                        {getContentPreview(item)}
                                    </p>
                                    <div className="item-actions">
                                        <button
                                            className="reuse-btn"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                onReuse(item);
                                            }}
                                        >
                                            📋 Copy
                                        </button>
                                        {item.data.used && (
                                            <span className="used-badge">Used</span>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                ))}
            </div>

            {sortedContent.length === 0 && (
                <div className="empty-state">
                    <p>No generated content yet. Start creating content in the other tabs!</p>
                </div>
            )}

            {/* Selected Item Detail Modal */}
            {selectedItem && (
                <div className="content-detail-modal" {...getClickableProps(() => setSelectedItem(null))}>
                    <div className="modal-content" {...getClickableProps((e) => e.stopPropagation())}>
                        <div className="modal-header">
                            <h3>
                                {contentTypes[selectedItem.type]?.icon} {getContentTitle(selectedItem)}
                            </h3>
                            <button className="close-btn" onClick={() => setSelectedItem(null)}>
                                ×
                            </button>
                        </div>

                        <div className="modal-body">
                            <div className="detail-meta">
                                <span className="detail-type">
                                    Type: {contentTypes[selectedItem.type]?.label}
                                </span>
                                <span className="detail-time">
                                    Created: {new Date(selectedItem.createdAt).toLocaleString()}
                                </span>
                            </div>

                            {/* Render content based on type */}
                            {selectedItem.type === 'npc_creation' && (
                                <div className="npc-details">
                                    <h4>{selectedItem.data.name}</h4>
                                    <p><strong>Race:</strong> {selectedItem.data.race}</p>
                                    <p><strong>Occupation:</strong> {selectedItem.data.occupation}</p>
                                    <p><strong>Description:</strong> {selectedItem.data.description}</p>
                                    <p><strong>Personality:</strong> {selectedItem.data.personalityTraits?.join(', ')}</p>
                                    <p><strong>Voice:</strong> {selectedItem.data.voiceDescription}</p>
                                    <p><strong>Motivations:</strong> {selectedItem.data.motivations}</p>
                                    {selectedItem.data.secrets && (
                                        <p><strong>Secrets:</strong> {selectedItem.data.secrets}</p>
                                    )}
                                </div>
                            )}

                            {selectedItem.type === 'location_description' && (
                                <div className="location-details">
                                    <h4>{selectedItem.data.name}</h4>
                                    <p>{selectedItem.data.description}</p>
                                    {selectedItem.data.atmosphere && (
                                        <p><strong>Atmosphere:</strong> {selectedItem.data.atmosphere}</p>
                                    )}
                                    {selectedItem.data.notableFeatures && (
                                        <div>
                                            <strong>Notable Features:</strong>
                                            <ul>
                                                {selectedItem.data.notableFeatures.map((feature, idx) => (
                                                    <li key={idx}>{feature}</li>
                                                ))}
                                            </ul>
                                        </div>
                                    )}
                                </div>
                            )}

                            {(selectedItem.type === 'combat_narration' || selectedItem.type === 'death_description') && (
                                <div className="combat-details">
                                    <p className="narration-text">{selectedItem.data.narration}</p>
                                    <div className="combat-meta">
                                        <p><strong>Attacker:</strong> {selectedItem.data.attackerName}</p>
                                        <p><strong>Target:</strong> {selectedItem.data.targetName}</p>
                                        <p><strong>Weapon:</strong> {selectedItem.data.weaponOrSpell}</p>
                                        {selectedItem.data.damage > 0 && (
                                            <p><strong>Damage:</strong> {selectedItem.data.damage}</p>
                                        )}
                                    </div>
                                </div>
                            )}

                            {selectedItem.type === 'environmental_hazard' && (
                                <div className="hazard-details">
                                    <h4>{selectedItem.data.name}</h4>
                                    <p>{selectedItem.data.description}</p>
                                    <p><strong>Trigger:</strong> {selectedItem.data.triggerCondition}</p>
                                    <p><strong>Effect:</strong> {selectedItem.data.effectDescription}</p>
                                    <p><strong>DC:</strong> {selectedItem.data.difficultyClass}</p>
                                    {selectedItem.data.damageFormula && (
                                        <p><strong>Damage:</strong> {selectedItem.data.damageFormula}</p>
                                    )}
                                </div>
                            )}

                            {(selectedItem.type === 'plot_twist' || selectedItem.type === 'story_hook' || 
                              selectedItem.type === 'revelation' || selectedItem.type === 'complication') && (
                                <div className="story-details">
                                    <h4>{selectedItem.data.title}</h4>
                                    <p>{selectedItem.data.description}</p>
                                    {selectedItem.data.impactLevel && (
                                        <p><strong>Impact:</strong> {selectedItem.data.impactLevel}</p>
                                    )}
                                    {selectedItem.data.consequences && (
                                        <div>
                                            <strong>Consequences:</strong>
                                            <ul>
                                                {selectedItem.data.consequences.map((consequence, idx) => (
                                                    <li key={idx}>{consequence}</li>
                                                ))}
                                            </ul>
                                        </div>
                                    )}
                                </div>
                            )}

                            {selectedItem.type === 'npc_dialogue' && (
                                <div className="dialogue-details">
                                    <p className="dialogue-text">"{selectedItem.data.dialogue}"</p>
                                    <div className="dialogue-meta">
                                        <p><strong>NPC:</strong> {selectedItem.data.npcName}</p>
                                        <p><strong>Situation:</strong> {selectedItem.data.situation}</p>
                                        {selectedItem.data.emotion && (
                                            <p><strong>Emotion:</strong> {selectedItem.data.emotion}</p>
                                        )}
                                    </div>
                                </div>
                            )}
                        </div>

                        <div className="modal-actions">
                            <button
                                className="copy-btn"
                                onClick={() => {
                                    onReuse(selectedItem);
                                    setSelectedItem(null);
                                }}
                            >
                                Copy to Clipboard
                            </button>
                            <button
                                className="close-action-btn"
                                onClick={() => setSelectedItem(null)}
                            >
                                Close
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default GeneratedContent;