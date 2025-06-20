import React, { useState, useEffect } from 'react';
import { 
  FaFlag, FaBrain, FaHandshake, FaChessKing, 
  FaScroll, FaBalanceScale, FaMemory, FaLightbulb
} from 'react-icons/fa';
import api from '../../services/api';
import { Radar } from 'react-chartjs-2';
import { getClickableProps, getSelectableProps } from '../../utils/accessibility';

const FactionPersonalities = ({ sessionId, isDM }) => {
  const [factions, setFactions] = useState([]);
  const [selectedFaction, setSelectedFaction] = useState(null);
  const [personality, setPersonality] = useState(null);
  const [agendas, setAgendas] = useState([]);
  const [, setLoading] = useState(false);
  const [showDecisionModal, setShowDecisionModal] = useState(false);
  const [decision, setDecision] = useState({
    type: 'diplomatic',
    context: '',
    options: []
  });

  useEffect(() => {
    loadFactions();
  }, [sessionId]);

  useEffect(() => {
    if (selectedFaction) {
      loadFactionDetails(selectedFaction.id);
    }
  }, [selectedFaction]);

  const loadFactions = async () => {
    try {
      const response = await api.get(`/sessions/${sessionId}/factions`);
      setFactions(response.data);
      if (response.data.length > 0 && !selectedFaction) {
        setSelectedFaction(response.data[0]);
      }
    } catch (err) {
      console.error('Failed to load factions:', err);
    }
  };

  const loadFactionDetails = async (factionId) => {
    setLoading(true);
    try {
      const [personalityRes, agendasRes] = await Promise.all([
        api.get(`/factions/${factionId}/personality`),
        api.get(`/factions/${factionId}/agendas`)
      ]);
      setPersonality(personalityRes.data);
      setAgendas(agendasRes.data);
    } catch (err) {
      // If personality doesn't exist, initialize it
      if (err.response?.status === 404 && isDM) {
        await initializePersonality(factionId);
      }
    } finally {
      setLoading(false);
    }
  };

  const initializePersonality = async (factionId) => {
    try {
      const response = await api.post(`/factions/${factionId}/personality`);
      setPersonality(response.data);
    } catch (err) {
      console.error('Failed to initialize personality:', err);
    }
  };

  const triggerDecision = async () => {
    if (!selectedFaction || !decision.context || decision.options.length < 2) return;

    try {
      const response = await api.post(`/factions/${selectedFaction.id}/decide`, decision);
      alert(`Decision made: ${response.data.reasoning}`);
      setShowDecisionModal(false);
      loadFactionDetails(selectedFaction.id);
    } catch (err) {
      console.error('Failed to make decision:', err);
    }
  };

  const getMoodIcon = (mood) => {
    const icons = {
      triumphant: '😄',
      confident: '😊',
      cautious: '😐',
      worried: '😟',
      desperate: '😰',
      neutral: '😶'
    };
    return icons[mood] || '😶';
  };

  const getTraitRadarData = () => {
    if (!personality) return null;

    const traits = personality.traits || {};
    const labels = Object.keys(traits).slice(0, 8); // Limit to 8 for readability
    const data = labels.map(trait => traits[trait] * 100);

    return {
      labels: labels.map(label => label.charAt(0).toUpperCase() + label.slice(1)),
      datasets: [{
        label: 'Personality Traits',
        data: data,
        backgroundColor: 'rgba(155, 89, 182, 0.2)',
        borderColor: 'rgba(155, 89, 182, 1)',
        pointBackgroundColor: 'rgba(155, 89, 182, 1)',
        pointBorderColor: '#fff',
        pointHoverBackgroundColor: '#fff',
        pointHoverBorderColor: 'rgba(155, 89, 182, 1)'
      }]
    };
  };

  const getValueBars = () => {
    if (!personality) return [];
    
    const values = personality.values || {};
    return Object.entries(values)
      .sort(([,a], [,b]) => b - a)
      .slice(0, 6);
  };

  const addDecisionOption = () => {
    setDecision({
      ...decision,
      options: [...decision.options, {
        id: `opt_${Date.now()}`,
        description: '',
        outcomes: {},
        requirements: {},
        risks: {}
      }]
    });
  };

  return (
    <div className="faction-personalities">
      <div className="faction-list">
        <h3>Faction AI Personalities</h3>
        <div className="faction-cards">
          {factions.map(faction => (
            <button 
              key={faction.id}
              className={`faction-card ${selectedFaction?.id === faction.id ? 'selected' : ''}`}
              onClick={() => setSelectedFaction(faction)}
              type="button"
              aria-pressed={selectedFaction?.id === faction.id}
              style={{ borderColor: faction.color || '#95a5a6' }}
            >
              <div className="faction-banner" style={{ backgroundColor: faction.color || '#95a5a6' }}>
                <FaFlag />
              </div>
              <h4>{faction.name}</h4>
              <p className="faction-type">{faction.type}</p>
              <div className="faction-stats">
                <span>Power: {faction.power || 50}</span>
                <span>Wealth: {faction.wealth || 50}</span>
              </div>
            </button>
          ))}
        </div>
      </div>

      {selectedFaction && personality && (
        <div className="faction-details">
          <div className="personality-header">
            <h3>{selectedFaction.name}</h3>
            <div className="mood-indicator">
              <span className="mood-icon">{getMoodIcon(personality.current_mood)}</span>
              <span className="mood-text">{personality.current_mood}</span>
            </div>
          </div>

          <div className="personality-content">
            <div className="traits-section">
              <h4><FaBrain /> Personality Traits</h4>
              {getTraitRadarData() && (
                <div className="radar-chart">
                  <Radar 
                    data={getTraitRadarData()} 
                    options={{
                      scale: {
                        ticks: { beginAtZero: true, max: 100 }
                      },
                      responsive: true,
                      maintainAspectRatio: false
                    }}
                  />
                </div>
              )}
            </div>

            <div className="values-section">
              <h4><FaBalanceScale /> Core Values</h4>
              <div className="values-bars">
                {getValueBars().map(([value, score]) => (
                  <div key={value} className="value-bar">
                    <div className="value-label">{value}</div>
                    <div className="value-progress">
                      <div 
                        className="value-fill"
                        style={{ width: `${score * 100}%` }}
                      />
                    </div>
                    <div className="value-score">{Math.round(score * 100)}</div>
                  </div>
                ))}
              </div>
            </div>

            <div className="agendas-section">
              <h4><FaScroll /> Political Agendas</h4>
              {agendas.length === 0 ? (
                <p className="empty-state">No active agendas</p>
              ) : (
                <div className="agendas-list">
                  {agendas.map(agenda => (
                    <div key={agenda.id} className="agenda-card">
                      <div className="agenda-header">
                        <h5>{agenda.title}</h5>
                        <span className={`agenda-status ${agenda.status}`}>
                          {agenda.status}
                        </span>
                      </div>
                      <p className="agenda-description">{agenda.description}</p>
                      <div className="agenda-progress">
                        <div className="progress-bar">
                          <div 
                            className="progress-fill"
                            style={{ width: `${agenda.progress * 100}%` }}
                          />
                        </div>
                        <span>{Math.round(agenda.progress * 100)}%</span>
                      </div>
                      <div className="agenda-stages">
                        {agenda.stages && agenda.stages.map((stage, idx) => (
                          <div 
                            key={idx} 
                            className={`stage ${stage.is_complete ? 'complete' : ''}`}
                          >
                            <span className="stage-number">{idx + 1}</span>
                            <span className="stage-name">{stage.name}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="memories-section">
              <h4><FaMemory /> Recent Memories</h4>
              {personality.memories && personality.memories.length > 0 ? (
                <div className="memories-list">
                  {personality.memories.slice(0, 5).map((memory, idx) => (
                    <div key={idx} className="memory-item">
                      <div className={`memory-impact ${memory.impact > 0 ? 'positive' : 'negative'}`}>
                        {memory.impact > 0 ? '+' : ''}{(memory.impact * 100).toFixed(0)}
                      </div>
                      <div className="memory-content">
                        <p>{memory.description}</p>
                        <span className="memory-type">{memory.event_type}</span>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="empty-state">No significant memories yet</p>
              )}
            </div>

            {isDM && (
              <div className="faction-actions">
                <button 
                  className="btn-trigger-decision"
                  onClick={() => setShowDecisionModal(true)}
                >
                  <FaLightbulb /> Trigger Decision
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {showDecisionModal && (
        <div className="modal-overlay" {...getClickableProps(() => setShowDecisionModal(false))}>
          <div className="modal-content large" {...getClickableProps(e => e.stopPropagation())}>
            <h3>Faction Decision Point</h3>
            
            <div className="form-group">
              <label>Decision Type</label>
              <select
                value={decision.type}
                onChange={(e) => setDecision({...decision, type: e.target.value})}
              >
                <option value="diplomatic">Diplomatic</option>
                <option value="military">Military</option>
                <option value="economic">Economic</option>
                <option value="strategic">Strategic</option>
              </select>
            </div>

            <div className="form-group">
              <label>Context</label>
              <textarea
                value={decision.context}
                onChange={(e) => setDecision({...decision, context: e.target.value})}
                placeholder="Describe the situation requiring a decision..."
                rows="3"
              />
            </div>

            <div className="form-group">
              <label>Decision Options</label>
              {decision.options.map((option, idx) => (
                <div key={option.id} className="decision-option">
                  <input
                    type="text"
                    value={option.description}
                    onChange={(e) => {
                      const newOptions = [...decision.options];
                      newOptions[idx].description = e.target.value;
                      setDecision({...decision, options: newOptions});
                    }}
                    placeholder={`Option ${idx + 1} description...`}
                  />
                </div>
              ))}
              <button onClick={addDecisionOption} className="btn-add-option">
                Add Option
              </button>
            </div>

            <div className="modal-actions">
              <button onClick={() => setShowDecisionModal(false)}>Cancel</button>
              <button 
                onClick={triggerDecision}
                disabled={!decision.context || decision.options.length < 2}
                className="btn-primary"
              >
                Let AI Decide
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default FactionPersonalities;