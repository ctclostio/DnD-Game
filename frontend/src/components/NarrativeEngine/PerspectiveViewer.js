import React, { useState, useEffect } from 'react';
import { 
  FaTheaterMasks, FaEye, FaUser, FaGlobe, FaScroll,
  FaExclamationTriangle, FaPlus, FaRandom, FaQuestionCircle
} from 'react-icons/fa';
import api from '../../services/api';
import { getClickableProps, getSelectableProps } from '../../utils/accessibility';

const PerspectiveViewer = ({ sessionId, characterId, isDM, onCreateEvent }) => {
  const [worldEvents, setWorldEvents] = useState([]);
  const [selectedEvent, setSelectedEvent] = useState(null);
  const [perspectives, setPerspectives] = useState([]);
  const [personalizedNarrative, setPersonalizedNarrative] = useState(null);
  const [loading, setLoading] = useState(false);
  const [showEventCreator, setShowEventCreator] = useState(false);
  const [compareMode, setCompareMode] = useState(false);
  const [comparedPerspectives, setComparedPerspectives] = useState([]);

  useEffect(() => {
    fetchRecentEvents();
  }, [sessionId]);

  const fetchRecentEvents = async () => {
    try {
      setLoading(true);
      // In a real implementation, this would fetch world events for the session
      // For now, we'll use a placeholder
      const events = [];
      setWorldEvents(events);
    } catch (error) {
      console.error('Failed to fetch events:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchPerspectives = async (eventId) => {
    try {
      setLoading(true);
      const response = await api.get(`/narrative/event/${eventId}/perspectives`);
      setPerspectives(response.data);
    } catch (error) {
      console.error('Failed to fetch perspectives:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchPersonalizedNarrative = async (eventId) => {
    if (!characterId) return;
    
    try {
      const response = await api.post(`/narrative/event/${eventId}/personalize/${characterId}`);
      setPersonalizedNarrative(response.data);
    } catch (error) {
      console.error('Failed to fetch personalized narrative:', error);
    }
  };

  const handleEventSelect = async (event) => {
    setSelectedEvent(event);
    await fetchPerspectives(event.id);
    await fetchPersonalizedNarrative(event.id);
  };

  const handleCreateEvent = async (eventData) => {
    try {
      const event = await onCreateEvent(eventData);
      setWorldEvents([event, ...worldEvents]);
      setShowEventCreator(false);
      handleEventSelect(event);
    } catch (error) {
      console.error('Failed to create event:', error);
    }
  };

  const toggleCompare = (perspectiveId) => {
    setComparedPerspectives(prev => {
      if (prev.includes(perspectiveId)) {
        return prev.filter(id => id !== perspectiveId);
      }
      if (prev.length >= 2) {
        return [prev[1], perspectiveId];
      }
      return [...prev, perspectiveId];
    });
  };

  const getTruthIndicator = (truthLevel) => {
    if (truthLevel >= 0.8) return { color: '#2ecc71', label: 'Highly Accurate' };
    if (truthLevel >= 0.6) return { color: '#3498db', label: 'Mostly True' };
    if (truthLevel >= 0.4) return { color: '#f39c12', label: 'Partially True' };
    if (truthLevel >= 0.2) return { color: '#e74c3c', label: 'Mostly False' };
    return { color: '#c0392b', label: 'Deceptive' };
  };

  const getBiasIcon = (bias) => {
    const icons = {
      positive: '😊',
      negative: '😠',
      neutral: '😐',
      conflicted: '🤔'
    };
    return icons[bias] || '❓';
  };

  // Helper function to process form data and reduce complexity
  const processEventFormData = (formData) => {
    const splitAndTrim = (value) => {
      return value.split(',').map(item => item.trim()).filter(item => item);
    };
    
    return {
      type: formData.get('type'),
      name: formData.get('name'),
      description: formData.get('description'),
      location: formData.get('location'),
      participants: splitAndTrim(formData.get('participants')),
      witnesses: splitAndTrim(formData.get('witnesses')),
      immediate_effects: splitAndTrim(formData.get('effects')),
      generate_perspectives: true,
      perspective_sources: [
        { type: 'participant', name: 'Direct Participant' },
        { type: 'witness', name: 'Bystander' },
        { type: 'faction', name: 'Local Authority' },
        { type: 'historical', name: 'Future Historian' }
      ]
    };
  };

  const handleEventFormSubmit = (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    const eventData = processEventFormData(formData);
    handleCreateEvent(eventData);
  };

  const EventCreator = () => (
    <div className="event-creator">
      <h4>Create World Event</h4>
      <form onSubmit={handleEventFormSubmit}>
        <div className="form-group">
          <label>Event Type:</label>
          <select name="type" required>
            <option value="">Select type...</option>
            <option value="battle">Battle</option>
            <option value="discovery">Discovery</option>
            <option value="betrayal">Betrayal</option>
            <option value="celebration">Celebration</option>
            <option value="disaster">Disaster</option>
            <option value="political">Political Event</option>
            <option value="supernatural">Supernatural Occurrence</option>
          </select>
        </div>

        <div className="form-group">
          <label>Event Name:</label>
          <input
            name="name"
            placeholder="e.g., The Battle of Crimson Bridge"
            required
          />
        </div>

        <div className="form-group">
          <label>Description:</label>
          <textarea
            name="description"
            placeholder="Describe what happened..."
            rows="4"
            required
          />
        </div>

        <div className="form-group">
          <label>Location:</label>
          <input
            name="location"
            placeholder="Where did this occur?"
            required
          />
        </div>

        <div className="form-group">
          <label>Key Participants (comma-separated):</label>
          <input
            name="participants"
            placeholder="e.g., Lord Blackwood, The Red Company, Player Party"
          />
        </div>

        <div className="form-group">
          <label>Witnesses (comma-separated):</label>
          <input
            name="witnesses"
            placeholder="e.g., Town Guards, Local Merchants"
          />
        </div>

        <div className="form-group">
          <label>Immediate Effects (comma-separated):</label>
          <input
            name="effects"
            placeholder="e.g., Bridge destroyed, Lord injured, Trade route blocked"
          />
        </div>

        <div className="form-actions">
          <button type="submit" className="btn-primary">
            Create Event & Generate Perspectives
          </button>
          <button 
            type="button" 
            className="btn-secondary"
            onClick={() => setShowEventCreator(false)}
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );

  return (
    <div className="perspective-viewer">
      <div className="viewer-header">
        <h3><FaTheaterMasks /> Multi-Perspective Storytelling</h3>
        {isDM && (
          <button 
            className="btn-primary"
            onClick={() => setShowEventCreator(!showEventCreator)}
          >
            <FaPlus /> Create Event
          </button>
        )}
      </div>

      {showEventCreator && <EventCreator />}

      <div className="viewer-content">
        {!selectedEvent ? (
          <div className="event-selector">
            <h4>Select or Create an Event</h4>
            {worldEvents.length === 0 ? (
              <div className="empty-state">
                <FaGlobe />
                <p>No world events yet. Create one to see multiple perspectives!</p>
              </div>
            ) : (
              <div className="event-list">
                {worldEvents.map(event => (
                  <div 
                    key={event.id} 
                    className="event-card"
                    {...getClickableProps(() => handleEventSelect(event))}
                  >
                    <h5>{event.name}</h5>
                    <p>{event.description}</p>
                    <div className="event-meta">
                      <span><FaGlobe /> {event.location}</span>
                      <span>{event.type}</span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        ) : (
          <div className="perspectives-display">
            <div className="selected-event">
              <h4>{selectedEvent.name}</h4>
              <p className="event-description">{selectedEvent.description}</p>
              <div className="event-details">
                <span><FaGlobe /> {selectedEvent.location}</span>
                <span><FaUser /> {selectedEvent.participants?.length || 0} participants</span>
                <span><FaEye /> {selectedEvent.witnesses?.length || 0} witnesses</span>
              </div>
              <button 
                className="btn-secondary"
                onClick={() => {
                  setSelectedEvent(null);
                  setPerspectives([]);
                  setPersonalizedNarrative(null);
                  setComparedPerspectives([]);
                }}
              >
                Back to Events
              </button>
            </div>

            {personalizedNarrative && (
              <div className="personalized-narrative">
                <h4><FaUser /> Your Character's Perspective</h4>
                <div className="narrative-content">
                  <p className="personalized-description">
                    {personalizedNarrative.metadata?.personalized_description || selectedEvent.description}
                  </p>
                  
                  {personalizedNarrative.personalized_hooks?.length > 0 && (
                    <div className="narrative-hooks">
                      <h5>Personal Connections:</h5>
                      {personalizedNarrative.personalized_hooks.map((hook, index) => (
                        <div key={index} className="hook-item">
                          <span className="hook-type">{hook.type}:</span>
                          <span className="hook-content">{hook.content}</span>
                          <span className="hook-relevance" style={{
                            opacity: 0.5 + (hook.relevance * 0.5)
                          }}>
                            {Math.round(hook.relevance * 100)}% relevant
                          </span>
                        </div>
                      ))}
                    </div>
                  )}

                  {personalizedNarrative.backstory_callbacks?.length > 0 && (
                    <div className="backstory-connections">
                      <h5>Backstory Echoes:</h5>
                      {personalizedNarrative.backstory_callbacks.map((callback, index) => (
                        <div key={index} className="backstory-echo">
                          <p>{callback.narrative_text}</p>
                          <small>Connection type: {callback.integration_type}</small>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            )}

            <div className="perspectives-controls">
              <h4>Different Perspectives</h4>
              {perspectives.length > 1 && (
                <button 
                  className={`btn-compare ${compareMode ? 'active' : ''}`}
                  onClick={() => setCompareMode(!compareMode)}
                >
                  {compareMode ? 'Exit Compare Mode' : 'Compare Perspectives'}
                </button>
              )}
            </div>

            {loading ? (
              <div className="loading-state">
                <div className="loading-spinner"></div>
                <p>Loading perspectives...</p>
              </div>
            ) : perspectives.length === 0 ? (
              <div className="empty-state">
                <FaTheaterMasks />
                <p>No perspectives generated yet</p>
              </div>
            ) : (
              <div className={`perspectives-grid ${compareMode ? 'compare-mode' : ''}`}>
                {perspectives.map(perspective => {
                  const truthIndicator = getTruthIndicator(perspective.truth_level);
                  const isCompared = comparedPerspectives.includes(perspective.id);
                  
                  return (
                    <div 
                      key={perspective.id} 
                      className={`perspective-card ${perspective.bias} ${isCompared ? 'compared' : ''}`}
                    >
                      <div className="perspective-header">
                        <div className="perspective-source">
                          <span className="source-type">{perspective.perspective_type}</span>
                          <span className="source-name">{perspective.source_name}</span>
                        </div>
                        <div className="perspective-meta">
                          <span className="bias-indicator" title={`Bias: ${perspective.bias}`}>
                            {getBiasIcon(perspective.bias)}
                          </span>
                          <span 
                            className="truth-indicator"
                            style={{ color: truthIndicator.color }}
                            title={truthIndicator.label}
                          >
                            {Math.round(perspective.truth_level * 100)}%
                          </span>
                        </div>
                      </div>

                      <div className="perspective-narrative">
                        <p>{perspective.narrative}</p>
                      </div>

                      {perspective.hidden_details?.length > 0 && (
                        <details className="hidden-details">
                          <summary>
                            <FaQuestionCircle /> What they're not saying...
                          </summary>
                          <ul>
                            {perspective.hidden_details.map((detail, index) => (
                              <li key={index}>{detail}</li>
                            ))}
                          </ul>
                        </details>
                      )}

                      {perspective.contradictions?.length > 0 && (
                        <div className="contradictions">
                          <h6><FaExclamationTriangle /> Contradictions:</h6>
                          {perspective.contradictions.map((contradiction, index) => (
                            <div key={index} className="contradiction-item">
                              <small>{contradiction.conflicting_detail}</small>
                            </div>
                          ))}
                        </div>
                      )}

                      <div className="perspective-footer">
                        <span className="emotional-tone">
                          Tone: {perspective.emotional_tone || 'neutral'}
                        </span>
                        {compareMode && (
                          <button
                            className="btn-compare-toggle"
                            onClick={() => toggleCompare(perspective.id)}
                          >
                            {isCompared ? 'Remove' : 'Add to Compare'}
                          </button>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}

            {compareMode && comparedPerspectives.length === 2 && (
              <div className="comparison-analysis">
                <h4>Perspective Comparison</h4>
                <div className="comparison-content">
                  {(() => {
                    const p1 = perspectives.find(p => p.id === comparedPerspectives[0]);
                    const p2 = perspectives.find(p => p.id === comparedPerspectives[1]);
                    
                    if (!p1 || !p2) return null;
                    
                    return (
                      <>
                        <div className="comparison-header">
                          <div>
                            <strong>{p1.source_name}</strong>
                            <span className="truth-level" style={{ 
                              color: getTruthIndicator(p1.truth_level).color 
                            }}>
                              {Math.round(p1.truth_level * 100)}% accurate
                            </span>
                          </div>
                          <span>vs</span>
                          <div>
                            <strong>{p2.source_name}</strong>
                            <span className="truth-level" style={{ 
                              color: getTruthIndicator(p2.truth_level).color 
                            }}>
                              {Math.round(p2.truth_level * 100)}% accurate
                            </span>
                          </div>
                        </div>
                        
                        <div className="comparison-insights">
                          <h5>Key Differences:</h5>
                          <ul>
                            <li>Bias: {p1.bias} vs {p2.bias}</li>
                            <li>Emotional tone: {p1.emotional_tone} vs {p2.emotional_tone}</li>
                            <li>Truth discrepancy: {Math.abs(p1.truth_level - p2.truth_level * 100).toFixed(0)}%</li>
                          </ul>
                          
                          {(p1.hidden_details?.length > 0 || p2.hidden_details?.length > 0) && (
                            <div className="hidden-comparison">
                              <h5>Information Gaps:</h5>
                              <p>
                                {p1.source_name} omits {p1.hidden_details?.length || 0} details.
                                {p2.source_name} omits {p2.hidden_details?.length || 0} details.
                              </p>
                            </div>
                          )}
                        </div>
                      </>
                    );
                  })()}
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      <div className="perspective-info">
        <FaQuestionCircle />
        <p>
          Every event is seen differently through different eyes. NPCs, factions, 
          and even the gods themselves may have their own version of the truth. 
          Use these perspectives to uncover hidden motivations and deeper truths.
        </p>
      </div>
    </div>
  );
};

export default PerspectiveViewer;