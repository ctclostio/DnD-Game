import React, { useState, useEffect } from 'react';
import { getClickableProps, getSelectableProps } from '../../utils/accessibility';

const WorldEventViewer = ({ sessionId }) => {
    const [events, setEvents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [filter, setFilter] = useState('all');
    const [sortBy, setSortBy] = useState('date');
    const [selectedEvent, setSelectedEvent] = useState(null);

    useEffect(() => {
        loadWorldEvents();
    }, [sessionId]);

    const loadWorldEvents = async () => {
        try {
            const response = await fetch(`/api/v1/sessions/${sessionId}/world-events`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                setEvents(data);
            }
        } catch (err) {
            console.error('Failed to load world events:', err);
        } finally {
            setLoading(false);
        }
    };

    const generateWorldEvents = async () => {
        try {
            const response = await fetch(`/api/v1/sessions/${sessionId}/world-events/generate`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (response.ok) {
                const newEvent = await response.json();
                setEvents(prev => [newEvent, ...prev]);
                setSelectedEvent(newEvent);
            }
        } catch (err) {
            console.error('Failed to generate world event:', err);
        }
    };

    const progressEvent = async (eventId) => {
        try {
            const response = await fetch(`/api/v1/world-events/${eventId}/progress`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (response.ok) {
                const updated = await response.json();
                setEvents(prev => prev.map(e => e.id === eventId ? updated : e));
                setSelectedEvent(updated);
            }
        } catch (err) {
            console.error('Failed to progress event:', err);
        }
    };

    const simulateEventProgression = async () => {
        try {
            const response = await fetch(`/api/v1/sessions/${sessionId}/world-events/simulate`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (response.ok) {
                await loadWorldEvents(); // Reload all events
                alert('Event progression simulated');
            }
        } catch (err) {
            console.error('Failed to simulate event progression:', err);
        }
    };

    const getEventTypeIcon = (type) => {
        const icons = {
            political: '👑',
            natural_disaster: '🌋',
            magical_anomaly: '✨',
            ancient_awakening: '🗿',
            planar_intrusion: '🌀',
            divine_intervention: '⚡',
            faction_conflict: '⚔️',
            economic: '💰',
            plague: '☠️',
            prophecy: '📜'
        };
        return icons[type] || '📋';
    };

    const getEventStatusColor = (status) => {
        const colors = {
            brewing: 'brewing',
            active: 'active',
            escalating: 'escalating',
            critical: 'critical',
            resolving: 'resolving',
            resolved: 'resolved',
            dormant: 'dormant'
        };
        return colors[status] || 'default';
    };

    const getEventCategoryColor = (category) => {
        const colors = {
            apocalyptic: 'apocalyptic',
            major: 'major',
            regional: 'regional',
            local: 'local',
            minor: 'minor'
        };
        return colors[category] || 'default';
    };

    const filteredEvents = events
        .filter(event => {
            if (filter === 'all') return true;
            if (filter === 'active') return ['active', 'escalating', 'critical'].includes(event.status);
            if (filter === 'ancient') return event.ancientCause;
            if (filter === 'prophecy') return event.prophecyConnection;
            return event.status === filter;
        })
        .sort((a, b) => {
            if (sortBy === 'date') {
                return new Date(b.createdAt) - new Date(a.createdAt);
            }
            if (sortBy === 'impact') {
                return b.impactLevel - a.impactLevel;
            }
            if (sortBy === 'progress') {
                return b.progression - a.progression;
            }
            return 0;
        });

    if (loading) {
        return <div className="loading">Loading world events...</div>;
    }

    return (
        <div className="world-event-viewer">
            <div className="viewer-header">
                <h3>World Events</h3>
                <div className="header-actions">
                    <button onClick={generateWorldEvents} className="btn btn-primary">
                        + Generate Event
                    </button>
                    <button onClick={simulateEventProgression} className="btn btn-secondary">
                        Simulate Progression
                    </button>
                </div>
            </div>

            <div className="event-controls">
                <div className="filter-controls">
                    <label>Filter:</label>
                    <select value={filter} onChange={(e) => setFilter(e.target.value)}>
                        <option value="all">All Events</option>
                        <option value="active">Active Only</option>
                        <option value="ancient">Ancient-Related</option>
                        <option value="prophecy">Prophecy-Related</option>
                        <option value="resolved">Resolved</option>
                        <option value="dormant">Dormant</option>
                    </select>
                </div>

                <div className="sort-controls">
                    <label>Sort by:</label>
                    <select value={sortBy} onChange={(e) => setSortBy(e.target.value)}>
                        <option value="date">Most Recent</option>
                        <option value="impact">Impact Level</option>
                        <option value="progress">Progress</option>
                    </select>
                </div>
            </div>

            <div className="event-layout">
                <div className="event-list">
                    {filteredEvents.length === 0 ? (
                        <div className="empty-state">
                            <p>No world events yet</p>
                            <p className="hint">Generate events or simulate faction conflicts</p>
                        </div>
                    ) : (
                        filteredEvents.map(event => (
                            <div
                                key={event.id}
                                className={`event-card ${selectedEvent?.id === event.id ? 'selected' : ''} ${getEventStatusColor(event.status)}`}
                                {...getSelectableProps(() => setSelectedEvent(event), selectedEvent?.id === event.id)}
                            >
                                <div className="event-header">
                                    <span className="event-icon">{getEventTypeIcon(event.type)}</span>
                                    <div className="event-info">
                                        <h4>{event.name}</h4>
                                        <span className={`event-category ${getEventCategoryColor(event.category)}`}>
                                            {event.category}
                                        </span>
                                    </div>
                                </div>

                                <div className="event-status">
                                    <span className={`status-badge ${getEventStatusColor(event.status)}`}>
                                        {event.status}
                                    </span>
                                    <div className="progress-bar">
                                        <div 
                                            className="progress-fill"
                                            style={{ width: `${event.progression}%` }}
                                        />
                                    </div>
                                </div>

                                <div className="event-stats">
                                    <div className="stat">
                                        <span>Impact:</span>
                                        <span>{event.impactLevel}/10</span>
                                    </div>
                                    <div className="stat">
                                        <span>Stage:</span>
                                        <span>{event.currentStage}/{event.totalStages || 5}</span>
                                    </div>
                                </div>

                                {event.ancientCause && (
                                    <div className="ancient-marker">🗿 Ancient Origin</div>
                                )}
                                {event.prophecyConnection && (
                                    <div className="prophecy-marker">📜 Prophecy Linked</div>
                                )}
                            </div>
                        ))
                    )}
                </div>

                {selectedEvent && (
                    <div className="event-detail-panel">
                        <div className="detail-header">
                            <h4>{selectedEvent.name}</h4>
                            <span className={`event-type ${selectedEvent.type}`}>
                                {getEventTypeIcon(selectedEvent.type)} {selectedEvent.type.replace('_', ' ')}
                            </span>
                        </div>

                        <div className="detail-section">
                            <h5>Description</h5>
                            <p>{selectedEvent.description}</p>
                        </div>

                        <div className="detail-section">
                            <h5>Public Knowledge</h5>
                            <p>{selectedEvent.publicKnowledge}</p>
                        </div>

                        <div className="detail-section secret">
                            <h5>Hidden Truth (DM Only)</h5>
                            <p>{selectedEvent.hiddenTruth}</p>
                        </div>

                        <div className="detail-section">
                            <h5>Current Effects</h5>
                            <ul>
                                {selectedEvent.currentEffects?.map((effect, index) => (
                                    <li key={index}>{effect}</li>
                                ))}
                            </ul>
                        </div>

                        {selectedEvent.requiredActions?.length > 0 && (
                            <div className="detail-section">
                                <h5>Required Actions</h5>
                                <ul>
                                    {selectedEvent.requiredActions.map((action, index) => (
                                        <li key={index}>{action}</li>
                                    ))}
                                </ul>
                            </div>
                        )}

                        {selectedEvent.possibleOutcomes?.length > 0 && (
                            <div className="detail-section">
                                <h5>Possible Outcomes</h5>
                                <ul>
                                    {selectedEvent.possibleOutcomes.map((outcome, index) => (
                                        <li key={index}>{outcome}</li>
                                    ))}
                                </ul>
                            </div>
                        )}

                        <div className="detail-section">
                            <h5>Affected Regions</h5>
                            <div className="region-list">
                                {selectedEvent.affectedRegions?.map((region, index) => (
                                    <span key={index} className="region-tag">{region}</span>
                                ))}
                            </div>
                        </div>

                        {selectedEvent.involvedFactions?.length > 0 && (
                            <div className="detail-section">
                                <h5>Involved Factions</h5>
                                <div className="faction-list">
                                    {selectedEvent.involvedFactions.map((faction, index) => (
                                        <span key={index} className="faction-tag">{faction}</span>
                                    ))}
                                </div>
                            </div>
                        )}

                        <div className="event-actions">
                            {selectedEvent.status !== 'resolved' && selectedEvent.status !== 'dormant' && (
                                <button
                                    onClick={() => progressEvent(selectedEvent.id)}
                                    className="btn btn-primary"
                                >
                                    Progress Event
                                </button>
                            )}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default WorldEventViewer;