import React, { useState, useEffect } from 'react';
import { getCampaignData, generateStoryArc, createSessionMemory, generateRecap, generateForeshadowing } from '../services/api';
import '../styles/campaign-manager.css';

const CampaignManager = ({ gameSessionId, isDM }) => {
    const [activeTab, setActiveTab] = useState('arcs');
    const [storyArcs, setStoryArcs] = useState([]);
    const [sessionMemories, setSessionMemories] = useState([]);
    const [plotThreads, setPlotThreads] = useState([]);
    const [foreshadowing, setForeshadowing] = useState([]);
    const [timeline, setTimeline] = useState([]);
    const [recap, setRecap] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    // Forms state
    const [arcForm, setArcForm] = useState({
        context: '',
        playerGoals: '',
        arcType: 'main_quest',
        complexity: 'moderate'
    });

    const [memoryForm, setMemoryForm] = useState({
        sessionNumber: '',
        sessionDate: new Date().toISOString().split('T')[0],
        keyEvents: [{ time: '', description: '', impact: '' }],
        npcsEncountered: [''],
        decisionsMode: [{ context: '', choice: '', outcome: '' }],
        itemsAcquired: [''],
        locationsVisited: ['']
    });

    useEffect(() => {
        if (gameSessionId) {
            loadCampaignData();
        }
    }, [gameSessionId]);

    const loadCampaignData = async () => {
        setLoading(true);
        try {
            const data = await getCampaignData(gameSessionId);
            setStoryArcs(data.storyArcs || []);
            setSessionMemories(data.sessionMemories || []);
            setPlotThreads(data.plotThreads || []);
            setForeshadowing(data.foreshadowing || []);
            setTimeline(data.timeline || []);
        } catch (err) {
            setError('Failed to load campaign data');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const handleGenerateStoryArc = async () => {
        setLoading(true);
        try {
            const playerGoalsArray = arcForm.playerGoals.split('\n').filter(g => g.trim());
            const arc = await generateStoryArc(gameSessionId, {
                ...arcForm,
                playerGoals: playerGoalsArray
            });
            setStoryArcs([arc, ...storyArcs]);
            setArcForm({
                context: '',
                playerGoals: '',
                arcType: 'main_quest',
                complexity: 'moderate'
            });
        } catch (err) {
            setError('Failed to generate story arc');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateSessionMemory = async () => {
        setLoading(true);
        try {
            // Filter out empty entries
            const filteredMemory = {
                ...memoryForm,
                keyEvents: memoryForm.keyEvents.filter(e => e.description),
                npcsEncountered: memoryForm.npcsEncountered.filter(n => n.trim()),
                decisionsMode: memoryForm.decisionsMode.filter(d => d.choice),
                itemsAcquired: memoryForm.itemsAcquired.filter(i => i.trim()),
                locationsVisited: memoryForm.locationsVisited.filter(l => l.trim())
            };

            const memory = await createSessionMemory(gameSessionId, filteredMemory);
            setSessionMemories([memory, ...sessionMemories]);
            // Reset form
            setMemoryForm({
                sessionNumber: '',
                sessionDate: new Date().toISOString().split('T')[0],
                keyEvents: [{ time: '', description: '', impact: '' }],
                npcsEncountered: [''],
                decisionsMode: [{ context: '', choice: '', outcome: '' }],
                itemsAcquired: [''],
                locationsVisited: ['']
            });
        } catch (err) {
            setError('Failed to create session memory');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const handleGenerateRecap = async () => {
        setLoading(true);
        try {
            const generatedRecap = await generateRecap(gameSessionId, { sessionCount: 3 });
            setRecap(generatedRecap);
        } catch (err) {
            setError('Failed to generate recap');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const handleGenerateForeshadowing = async (plotThreadId, storyArcId) => {
        setLoading(true);
        try {
            const element = await generateForeshadowing(gameSessionId, {
                plotThreadId,
                storyArcId,
                elementType: 'rumor',
                subtletyLevel: 5
            });
            setForeshadowing([element, ...foreshadowing]);
        } catch (err) {
            setError('Failed to generate foreshadowing');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const renderStoryArcs = () => (
        <div className="story-arcs-tab">
            {isDM && (
                <div className="arc-generator">
                    <h3>Generate New Story Arc</h3>
                    <div className="form-group">
                        <label>Campaign Context</label>
                        <textarea
                            value={arcForm.context}
                            onChange={(e) => setArcForm({ ...arcForm, context: e.target.value })}
                            placeholder="Current state of the campaign..."
                            rows="3"
                        />
                    </div>
                    <div className="form-group">
                        <label>Player Goals (one per line)</label>
                        <textarea
                            value={arcForm.playerGoals}
                            onChange={(e) => setArcForm({ ...arcForm, playerGoals: e.target.value })}
                            placeholder="Find the lost artifact&#10;Defeat the evil wizard&#10;Save the kingdom"
                            rows="3"
                        />
                    </div>
                    <div className="form-row">
                        <div className="form-group">
                            <label>Arc Type</label>
                            <select
                                value={arcForm.arcType}
                                onChange={(e) => setArcForm({ ...arcForm, arcType: e.target.value })}
                            >
                                <option value="main_quest">Main Quest</option>
                                <option value="side_quest">Side Quest</option>
                                <option value="character_arc">Character Arc</option>
                                <option value="faction_conflict">Faction Conflict</option>
                            </select>
                        </div>
                        <div className="form-group">
                            <label>Complexity</label>
                            <select
                                value={arcForm.complexity}
                                onChange={(e) => setArcForm({ ...arcForm, complexity: e.target.value })}
                            >
                                <option value="simple">Simple</option>
                                <option value="moderate">Moderate</option>
                                <option value="complex">Complex</option>
                            </select>
                        </div>
                    </div>
                    <button 
                        onClick={handleGenerateStoryArc} 
                        disabled={loading || !arcForm.context}
                        className="generate-btn"
                    >
                        Generate Story Arc
                    </button>
                </div>
            )}

            <div className="arcs-list">
                <h3>Active Story Arcs</h3>
                {storyArcs.length === 0 ? (
                    <p className="empty-state">No story arcs yet. Generate one to get started!</p>
                ) : (
                    storyArcs.map(arc => (
                        <div key={arc.id} className={`arc-card ${arc.status}`}>
                            <div className="arc-header">
                                <h4>{arc.title}</h4>
                                <span className={`arc-type ${arc.arc_type}`}>{arc.arc_type.replace('_', ' ')}</span>
                                <span className="importance">Importance: {arc.importance_level}/10</span>
                            </div>
                            <p>{arc.description}</p>
                            {arc.metadata && arc.metadata.key_milestones && (
                                <div className="milestones">
                                    <h5>Key Milestones</h5>
                                    <ul>
                                        {arc.metadata.key_milestones.map((milestone, idx) => (
                                            <li key={idx}>
                                                <strong>{milestone.title}:</strong> {milestone.description}
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            )}
                            <div className="arc-actions">
                                <button 
                                    onClick={() => handleGenerateForeshadowing(null, arc.id)}
                                    className="action-btn"
                                >
                                    Generate Foreshadowing
                                </button>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );

    const renderSessionMemories = () => (
        <div className="session-memories-tab">
            <div className="recap-section">
                <h3>Session Recap Generator</h3>
                <button onClick={handleGenerateRecap} disabled={loading} className="generate-btn">
                    Generate "Previously On..." Recap
                </button>
                {recap && (
                    <div className="recap-display">
                        <h4>Previously on your adventure...</h4>
                        <p className="recap-summary">{recap.summary}</p>
                        {recap.key_events && recap.key_events.length > 0 && (
                            <div className="key-events">
                                <h5>Key Events:</h5>
                                <ul>
                                    {recap.key_events.map((event, idx) => (
                                        <li key={idx}>{event}</li>
                                    ))}
                                </ul>
                            </div>
                        )}
                        {recap.cliffhanger && (
                            <p className="cliffhanger">
                                <strong>And now...</strong> {recap.cliffhanger}
                            </p>
                        )}
                    </div>
                )}
            </div>

            {isDM && (
                <div className="memory-creator">
                    <h3>Record Session Memory</h3>
                    <div className="form-row">
                        <div className="form-group">
                            <label>Session Number</label>
                            <input
                                type="number"
                                value={memoryForm.sessionNumber}
                                onChange={(e) => setMemoryForm({ ...memoryForm, sessionNumber: e.target.value })}
                                placeholder="1"
                            />
                        </div>
                        <div className="form-group">
                            <label>Session Date</label>
                            <input
                                type="date"
                                value={memoryForm.sessionDate}
                                onChange={(e) => setMemoryForm({ ...memoryForm, sessionDate: e.target.value })}
                            />
                        </div>
                    </div>

                    <div className="form-group">
                        <label>Key Events</label>
                        {memoryForm.keyEvents.map((event, idx) => (
                            <div key={idx} className="event-entry">
                                <input
                                    type="text"
                                    placeholder="When"
                                    value={event.time}
                                    onChange={(e) => {
                                        const newEvents = [...memoryForm.keyEvents];
                                        newEvents[idx].time = e.target.value;
                                        setMemoryForm({ ...memoryForm, keyEvents: newEvents });
                                    }}
                                />
                                <input
                                    type="text"
                                    placeholder="What happened"
                                    value={event.description}
                                    onChange={(e) => {
                                        const newEvents = [...memoryForm.keyEvents];
                                        newEvents[idx].description = e.target.value;
                                        setMemoryForm({ ...memoryForm, keyEvents: newEvents });
                                    }}
                                />
                                <input
                                    type="text"
                                    placeholder="Impact"
                                    value={event.impact}
                                    onChange={(e) => {
                                        const newEvents = [...memoryForm.keyEvents];
                                        newEvents[idx].impact = e.target.value;
                                        setMemoryForm({ ...memoryForm, keyEvents: newEvents });
                                    }}
                                />
                            </div>
                        ))}
                        <button 
                            onClick={() => setMemoryForm({
                                ...memoryForm,
                                keyEvents: [...memoryForm.keyEvents, { time: '', description: '', impact: '' }]
                            })}
                            className="add-btn"
                        >
                            Add Event
                        </button>
                    </div>

                    <button 
                        onClick={handleCreateSessionMemory} 
                        disabled={loading || !memoryForm.sessionNumber}
                        className="save-btn"
                    >
                        Save Session Memory
                    </button>
                </div>
            )}

            <div className="memories-list">
                <h3>Session History</h3>
                {sessionMemories.map(memory => (
                    <div key={memory.id} className="memory-card">
                        <div className="memory-header">
                            <h4>Session {memory.session_number}</h4>
                            <span className="date">{new Date(memory.session_date).toLocaleDateString()}</span>
                        </div>
                        {memory.recap_summary && <p className="summary">{memory.recap_summary}</p>}
                    </div>
                ))}
            </div>
        </div>
    );

    const renderPlotThreads = () => (
        <div className="plot-threads-tab">
            <h3>Active Plot Threads</h3>
            {plotThreads.length === 0 ? (
                <p className="empty-state">No active plot threads</p>
            ) : (
                <div className="threads-grid">
                    {plotThreads.map(thread => (
                        <div key={thread.id} className={`thread-card ${thread.status}`}>
                            <div className="thread-header">
                                <h4>{thread.title}</h4>
                                <span className="thread-type">{thread.thread_type}</span>
                                <div className="tension-meter">
                                    <span>Tension: </span>
                                    <div className="meter">
                                        <div 
                                            className="meter-fill" 
                                            style={{ width: `${thread.tension_level * 10}%` }}
                                        />
                                    </div>
                                </div>
                            </div>
                            <p>{thread.description}</p>
                            {isDM && (
                                <button 
                                    onClick={() => handleGenerateForeshadowing(thread.id, null)}
                                    className="action-btn"
                                >
                                    Add Foreshadowing
                                </button>
                            )}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );

    const renderForeshadowing = () => (
        <div className="foreshadowing-tab">
            <h3>Foreshadowing Elements</h3>
            {!isDM ? (
                <p className="restricted">Only the DM can view foreshadowing elements</p>
            ) : (
                <div className="foreshadowing-grid">
                    {foreshadowing.filter(f => !f.revealed).map(element => (
                        <div key={element.id} className="foreshadowing-card">
                            <div className="element-header">
                                <span className="element-type">{element.element_type}</span>
                                <span className="subtlety">Subtlety: {element.subtlety_level}/10</span>
                            </div>
                            <p className="content">{element.content}</p>
                            {element.placement_suggestions && (
                                <div className="suggestions">
                                    <h5>Placement Ideas:</h5>
                                    <ul>
                                        {element.placement_suggestions.map((suggestion, idx) => (
                                            <li key={idx}>{suggestion}</li>
                                        ))}
                                    </ul>
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );

    const renderTimeline = () => (
        <div className="timeline-tab">
            <h3>Campaign Timeline</h3>
            <div className="timeline">
                {timeline.map(event => (
                    <div key={event.id} className={`timeline-event impact-${event.impact_level}`}>
                        <div className="event-date">
                            {new Date(event.event_date).toLocaleDateString()}
                        </div>
                        <div className="event-content">
                            <h4>{event.event_title}</h4>
                            <p>{event.event_description}</p>
                            <span className="event-type">{event.event_type}</span>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );

    return (
        <div className="campaign-manager">
            <div className="tabs">
                <button 
                    className={activeTab === 'arcs' ? 'active' : ''} 
                    onClick={() => setActiveTab('arcs')}
                >
                    Story Arcs
                </button>
                <button 
                    className={activeTab === 'memories' ? 'active' : ''} 
                    onClick={() => setActiveTab('memories')}
                >
                    Session Memories
                </button>
                <button 
                    className={activeTab === 'threads' ? 'active' : ''} 
                    onClick={() => setActiveTab('threads')}
                >
                    Plot Threads
                </button>
                <button 
                    className={activeTab === 'foreshadowing' ? 'active' : ''} 
                    onClick={() => setActiveTab('foreshadowing')}
                >
                    Foreshadowing
                </button>
                <button 
                    className={activeTab === 'timeline' ? 'active' : ''} 
                    onClick={() => setActiveTab('timeline')}
                >
                    Timeline
                </button>
            </div>

            {error && <div className="error-message">{error}</div>}

            <div className="tab-content">
                {loading && <div className="loading">Loading...</div>}
                {!loading && (
                    <>
                        {activeTab === 'arcs' && renderStoryArcs()}
                        {activeTab === 'memories' && renderSessionMemories()}
                        {activeTab === 'threads' && renderPlotThreads()}
                        {activeTab === 'foreshadowing' && renderForeshadowing()}
                        {activeTab === 'timeline' && renderTimeline()}
                    </>
                )}
            </div>
        </div>
    );
};

export default CampaignManager;