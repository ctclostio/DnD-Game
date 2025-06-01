import React, { useState, useEffect } from 'react';
import { 
  FaChessBoard, FaBolt, FaClock, FaExclamationTriangle, 
  FaPlay, FaEye, FaFilter, FaPlus, FaInfoCircle 
} from 'react-icons/fa';
import api from '../../services/api';

const ConsequenceTracker = ({ consequences, sessionId, isDM, onRecordAction, onConsequenceUpdate }) => {
  const [filter, setFilter] = useState('all');
  const [sortBy, setSortBy] = useState('severity');
  const [expandedConsequence, setExpandedConsequence] = useState(null);
  const [showActionForm, setShowActionForm] = useState(false);
  const [selectedConsequences, setSelectedConsequences] = useState([]);

  const delayIcons = {
    immediate: <FaBolt />,
    short: <FaClock style={{ color: '#f39c12' }} />,
    medium: <FaClock style={{ color: '#3498db' }} />,
    long: <FaClock style={{ color: '#9b59b6' }} />
  };

  const delayDescriptions = {
    immediate: 'Within this session',
    short: 'Next 1-2 sessions',
    medium: 'Within a week of game time',
    long: 'Months or years later'
  };

  const getSeverityColor = (severity) => {
    if (severity <= 3) return '#2ecc71';
    if (severity <= 6) return '#f39c12';
    if (severity <= 8) return '#e74c3c';
    return '#8e44ad';
  };

  const getSeverityLabel = (severity) => {
    if (severity <= 3) return 'Minor';
    if (severity <= 6) return 'Moderate';
    if (severity <= 8) return 'Major';
    return 'Catastrophic';
  };

  const filteredConsequences = consequences
    .filter(c => {
      if (filter === 'all') return true;
      if (filter === 'pending') return c.status === 'pending';
      if (filter === 'triggered') return c.status === 'triggered';
      return c.delay === filter;
    })
    .sort((a, b) => {
      if (sortBy === 'severity') return b.severity - a.severity;
      if (sortBy === 'delay') {
        const delayOrder = { immediate: 0, short: 1, medium: 2, long: 3 };
        return delayOrder[a.delay] - delayOrder[b.delay];
      }
      return new Date(b.created_at) - new Date(a.created_at);
    });

  const handleTriggerConsequence = async (consequenceId) => {
    if (!isDM) return;
    
    if (!window.confirm('Are you sure you want to trigger this consequence?')) {
      return;
    }

    try {
      await api.post(`/narrative/consequences/${consequenceId}/trigger`, {
        session_id: sessionId
      });
      onConsequenceUpdate();
    } catch (error) {
      console.error('Failed to trigger consequence:', error);
    }
  };

  const handleBatchTrigger = async () => {
    if (!isDM || selectedConsequences.length === 0) return;
    
    if (!window.confirm(`Trigger ${selectedConsequences.length} consequences?`)) {
      return;
    }

    try {
      await Promise.all(
        selectedConsequences.map(id => 
          api.post(`/narrative/consequences/${id}/trigger`, {
            session_id: sessionId
          })
        )
      );
      setSelectedConsequences([]);
      onConsequenceUpdate();
    } catch (error) {
      console.error('Failed to trigger consequences:', error);
    }
  };

  const toggleConsequenceSelection = (id) => {
    setSelectedConsequences(prev => 
      prev.includes(id) 
        ? prev.filter(cId => cId !== id)
        : [...prev, id]
    );
  };

  const ActionRecorder = () => (
    <div className="action-recorder-form">
      <h4>Record Player Action</h4>
      <form onSubmit={(e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        onRecordAction({
          action_type: formData.get('action_type'),
          target_type: formData.get('target_type'),
          target_id: formData.get('target_id'),
          action_description: formData.get('description'),
          moral_weight: formData.get('moral_weight'),
          immediate_result: formData.get('result')
        });
        e.target.reset();
        setShowActionForm(false);
      }}>
        <div className="form-grid">
          <input
            name="action_type"
            placeholder="Action (e.g., kill, save, betray)"
            required
          />
          <input
            name="target_type"
            placeholder="Target type (e.g., npc, faction)"
            required
          />
          <input
            name="target_id"
            placeholder="Target ID/Name"
          />
          <select name="moral_weight" required>
            <option value="">Moral Weight</option>
            <option value="good">Good</option>
            <option value="evil">Evil</option>
            <option value="neutral">Neutral</option>
            <option value="chaotic">Chaotic</option>
            <option value="lawful">Lawful</option>
          </select>
        </div>
        <textarea
          name="description"
          placeholder="Describe the action and its context..."
          rows="3"
          required
        />
        <input
          name="result"
          placeholder="What happened immediately?"
          required
        />
        <div className="form-actions">
          <button type="submit" className="btn-primary">Record Action</button>
          <button 
            type="button" 
            className="btn-secondary"
            onClick={() => setShowActionForm(false)}
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );

  return (
    <div className="consequence-tracker">
      <div className="tracker-header">
        <h3><FaChessBoard /> Consequence Cascade System</h3>
        <div className="header-actions">
          <button 
            className="btn-primary"
            onClick={() => setShowActionForm(!showActionForm)}
          >
            <FaPlus /> Record Action
          </button>
        </div>
      </div>

      {showActionForm && <ActionRecorder />}

      <div className="tracker-controls">
        <div className="filter-controls">
          <label><FaFilter /> Filter:</label>
          <select value={filter} onChange={(e) => setFilter(e.target.value)}>
            <option value="all">All Consequences</option>
            <option value="pending">Pending</option>
            <option value="triggered">Triggered</option>
            <option value="immediate">Immediate</option>
            <option value="short">Short-term</option>
            <option value="medium">Medium-term</option>
            <option value="long">Long-term</option>
          </select>
        </div>
        
        <div className="sort-controls">
          <label>Sort by:</label>
          <select value={sortBy} onChange={(e) => setSortBy(e.target.value)}>
            <option value="severity">Severity</option>
            <option value="delay">Timeline</option>
            <option value="date">Date Created</option>
          </select>
        </div>

        {isDM && selectedConsequences.length > 0 && (
          <button 
            className="btn-batch-trigger"
            onClick={handleBatchTrigger}
          >
            Trigger {selectedConsequences.length} Selected
          </button>
        )}
      </div>

      <div className="consequences-list">
        {filteredConsequences.length === 0 ? (
          <div className="empty-state">
            <FaChessBoard />
            <h4>No Consequences Found</h4>
            <p>Player actions will create ripple effects that appear here</p>
          </div>
        ) : (
          filteredConsequences.map(consequence => (
            <div 
              key={consequence.id} 
              className={`consequence-card ${consequence.status} ${
                selectedConsequences.includes(consequence.id) ? 'selected' : ''
              }`}
            >
              <div className="consequence-header">
                <div className="consequence-meta">
                  {isDM && consequence.status === 'pending' && (
                    <input
                      type="checkbox"
                      checked={selectedConsequences.includes(consequence.id)}
                      onChange={() => toggleConsequenceSelection(consequence.id)}
                    />
                  )}
                  <div 
                    className="severity-indicator"
                    style={{ backgroundColor: getSeverityColor(consequence.severity) }}
                    title={`Severity: ${consequence.severity}/10`}
                  >
                    {consequence.severity}
                  </div>
                  <span className="severity-label">
                    {getSeverityLabel(consequence.severity)}
                  </span>
                  <div className="delay-indicator" title={delayDescriptions[consequence.delay]}>
                    {delayIcons[consequence.delay]}
                    <span>{consequence.delay}</span>
                  </div>
                </div>
                <div className="consequence-status">
                  <span className={`status-badge status-${consequence.status}`}>
                    {consequence.status}
                  </span>
                </div>
              </div>

              <div className="consequence-content">
                <p className="consequence-description">{consequence.description}</p>
                
                {consequence.trigger_type && (
                  <div className="trigger-info">
                    <small>Triggered by: {consequence.trigger_type}</small>
                  </div>
                )}
              </div>

              {consequence.affected_entities && consequence.affected_entities.length > 0 && (
                <div className="affected-entities">
                  <h5>Affected:</h5>
                  <div className="entity-list">
                    {consequence.affected_entities.map((entity, index) => (
                      <div key={index} className="entity-item">
                        <span className="entity-type">{entity.entity_type}:</span>
                        <span className="entity-name">{entity.entity_name}</span>
                        <span className="impact-severity" style={{
                          color: getSeverityColor(entity.impact_severity)
                        }}>
                          ({entity.impact_type})
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <div className="consequence-actions">
                <button
                  className="btn-icon"
                  onClick={() => setExpandedConsequence(
                    expandedConsequence === consequence.id ? null : consequence.id
                  )}
                >
                  <FaEye /> {expandedConsequence === consequence.id ? 'Hide' : 'View'} Details
                </button>
                
                {isDM && consequence.status === 'pending' && (
                  <button
                    className="btn-trigger"
                    onClick={() => handleTriggerConsequence(consequence.id)}
                  >
                    <FaPlay /> Trigger Now
                  </button>
                )}
              </div>

              {expandedConsequence === consequence.id && (
                <div className="consequence-details">
                  {consequence.cascade_effects && consequence.cascade_effects.length > 0 && (
                    <div className="cascade-effects">
                      <h5>Potential Cascade Effects:</h5>
                      {consequence.cascade_effects.map((effect, index) => (
                        <div key={index} className="cascade-item">
                          <div className="cascade-header">
                            <span className="cascade-type">{effect.type}</span>
                            <span className="cascade-probability">
                              {Math.round(effect.probability * 100)}% chance
                            </span>
                          </div>
                          <p>{effect.description}</p>
                          <small>Timeline: {effect.timeline}</small>
                        </div>
                      ))}
                    </div>
                  )}

                  {consequence.metadata?.prevention_methods && (
                    <div className="prevention-methods">
                      <h5>Ways to Prevent/Mitigate:</h5>
                      <ul>
                        {consequence.metadata.prevention_methods.map((method, index) => (
                          <li key={index}>{method}</li>
                        ))}
                      </ul>
                    </div>
                  )}

                  {consequence.actual_trigger_time && (
                    <div className="trigger-time">
                      <small>
                        Triggered at: {new Date(consequence.actual_trigger_time).toLocaleString()}
                      </small>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))
        )}
      </div>

      <div className="consequence-info">
        <FaInfoCircle />
        <p>
          Every action creates ripples. Some consequences happen immediately, 
          while others may not manifest for months of game time. The AI tracks 
          these cascading effects and weaves them into your story when the time is right.
        </p>
      </div>
    </div>
  );
};

export default ConsequenceTracker;