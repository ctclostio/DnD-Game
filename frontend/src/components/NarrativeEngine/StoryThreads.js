import React, { useState } from 'react';
import { 
  FaComments, FaLink, FaFire, FaSnowflake, FaCheckCircle,
  FaBan, FaPlus, FaEdit, FaChartLine, FaUsers
} from 'react-icons/fa';
import api from '../../services/api';

const StoryThreads = ({ threads, sessionId, isDM, onThreadUpdate }) => {
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [selectedThread, setSelectedThread] = useState(null);
  const [newThread, setNewThread] = useState({
    name: '',
    description: '',
    thread_type: 'side_quest',
    tension_level: 0.5,
    key_participants: []
  });
  const [participantInput, setParticipantInput] = useState('');

  const threadTypes = {
    main_quest: { icon: '⚔️', label: 'Main Quest', color: '#e74c3c' },
    side_quest: { icon: '📜', label: 'Side Quest', color: '#3498db' },
    character_arc: { icon: '🎭', label: 'Character Arc', color: '#9b59b6' },
    world_event: { icon: '🌍', label: 'World Event', color: '#2ecc71' },
    mystery: { icon: '🔍', label: 'Mystery', color: '#f39c12' },
    relationship: { icon: '💕', label: 'Relationship', color: '#e91e63' }
  };

  const getStatusIcon = (status) => {
    const icons = {
      active: <FaFire style={{ color: '#e74c3c' }} />,
      dormant: <FaSnowflake style={{ color: '#3498db' }} />,
      resolved: <FaCheckCircle style={{ color: '#2ecc71' }} />,
      abandoned: <FaBan style={{ color: '#95a5a6' }} />
    };
    return icons[status] || null;
  };

  const getTensionBar = (tensionLevel) => {
    const percentage = tensionLevel * 100;
    const color = tensionLevel > 0.7 ? '#e74c3c' : tensionLevel > 0.4 ? '#f39c12' : '#3498db';
    
    return (
      <div className="tension-bar">
        <div 
          className="tension-fill" 
          style={{ 
            width: `${percentage}%`,
            backgroundColor: color
          }}
        />
      </div>
    );
  };

  const getResolutionProximity = (proximity) => {
    if (proximity >= 0.8) return { text: 'Climax Approaching', class: 'climax' };
    if (proximity >= 0.6) return { text: 'Building to Resolution', class: 'building' };
    if (proximity >= 0.4) return { text: 'Mid Development', class: 'developing' };
    if (proximity >= 0.2) return { text: 'Early Stages', class: 'early' };
    return { text: 'Just Beginning', class: 'beginning' };
  };

  const handleCreateThread = async (e) => {
    e.preventDefault();
    
    try {
      const threadData = {
        ...newThread,
        metadata: { session_id: sessionId }
      };
      
      await api.post('/narrative/threads', threadData);
      onThreadUpdate();
      setShowCreateForm(false);
      setNewThread({
        name: '',
        description: '',
        thread_type: 'side_quest',
        tension_level: 0.5,
        key_participants: []
      });
      setParticipantInput('');
    } catch (error) {
      console.error('Failed to create thread:', error);
    }
  };

  const handleUpdateThreadStatus = async (threadId, newStatus) => {
    if (!isDM) return;
    
    try {
      await api.put(`/narrative/threads/${threadId}`, { status: newStatus });
      onThreadUpdate();
    } catch (error) {
      console.error('Failed to update thread status:', error);
    }
  };

  const handleAddParticipant = (e) => {
    if (e.key === 'Enter' && participantInput.trim()) {
      e.preventDefault();
      if (!newThread.key_participants.includes(participantInput.trim())) {
        setNewThread({
          ...newThread,
          key_participants: [...newThread.key_participants, participantInput.trim()]
        });
      }
      setParticipantInput('');
    }
  };

  const removeParticipant = (participant) => {
    setNewThread({
      ...newThread,
      key_participants: newThread.key_participants.filter(p => p !== participant)
    });
  };

  const activeThreads = threads.filter(t => t.status === 'active');
  const dormantThreads = threads.filter(t => t.status === 'dormant');
  const resolvedThreads = threads.filter(t => t.status === 'resolved' || t.status === 'abandoned');

  return (
    <div className="story-threads">
      <div className="threads-header">
        <h3><FaComments /> Narrative Threads</h3>
        {isDM && (
          <button 
            className="btn-primary"
            onClick={() => setShowCreateForm(!showCreateForm)}
          >
            <FaPlus /> Create Thread
          </button>
        )}
      </div>

      {showCreateForm && (
        <form className="thread-form" onSubmit={handleCreateThread}>
          <h4>Create Narrative Thread</h4>
          
          <div className="form-group">
            <label>Thread Name:</label>
            <input
              value={newThread.name}
              onChange={(e) => setNewThread({ ...newThread, name: e.target.value })}
              placeholder="e.g., The Missing Crown Prince"
              required
            />
          </div>

          <div className="form-group">
            <label>Type:</label>
            <div className="type-selector">
              {Object.entries(threadTypes).map(([type, info]) => (
                <button
                  key={type}
                  type="button"
                  className={`type-button ${newThread.thread_type === type ? 'active' : ''}`}
                  onClick={() => setNewThread({ ...newThread, thread_type: type })}
                  style={{ borderColor: newThread.thread_type === type ? info.color : '#ddd' }}
                >
                  <span className="type-icon">{info.icon}</span>
                  <span>{info.label}</span>
                </button>
              ))}
            </div>
          </div>

          <div className="form-group">
            <label>Description:</label>
            <textarea
              value={newThread.description}
              onChange={(e) => setNewThread({ ...newThread, description: e.target.value })}
              placeholder="Describe the narrative thread..."
              rows="3"
            />
          </div>

          <div className="form-group">
            <label>Key Participants (press Enter to add):</label>
            <div className="participant-input-container">
              <input
                type="text"
                value={participantInput}
                onChange={(e) => setParticipantInput(e.target.value)}
                onKeyDown={handleAddParticipant}
                placeholder="Add participant..."
              />
              <div className="participant-list">
                {newThread.key_participants.map(participant => (
                  <span key={participant} className="participant-tag">
                    {participant}
                    <button type="button" onClick={() => removeParticipant(participant)}>
                      &times;
                    </button>
                  </span>
                ))}
              </div>
            </div>
          </div>

          <div className="form-group">
            <label>
              Initial Tension Level:
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                value={newThread.tension_level}
                onChange={(e) => setNewThread({ 
                  ...newThread, 
                  tension_level: parseFloat(e.target.value) 
                })}
              />
              <span className="tension-value">{(newThread.tension_level * 100).toFixed(0)}%</span>
            </label>
          </div>

          <div className="form-actions">
            <button type="submit" className="btn-primary">Create Thread</button>
            <button 
              type="button" 
              className="btn-secondary"
              onClick={() => {
                setShowCreateForm(false);
                setNewThread({
                  name: '',
                  description: '',
                  thread_type: 'side_quest',
                  tension_level: 0.5,
                  key_participants: []
                });
                setParticipantInput('');
              }}
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      <div className="threads-overview">
        <div className="thread-stats">
          <div className="stat-card">
            <FaFire />
            <span className="stat-value">{activeThreads.length}</span>
            <span className="stat-label">Active</span>
          </div>
          <div className="stat-card">
            <FaSnowflake />
            <span className="stat-value">{dormantThreads.length}</span>
            <span className="stat-label">Dormant</span>
          </div>
          <div className="stat-card">
            <FaCheckCircle />
            <span className="stat-value">{resolvedThreads.length}</span>
            <span className="stat-label">Resolved</span>
          </div>
        </div>
      </div>

      <div className="threads-sections">
        {activeThreads.length > 0 && (
          <div className="thread-section">
            <h4><FaFire /> Active Threads</h4>
            <div className="thread-list">
              {activeThreads.map(thread => {
                const typeInfo = threadTypes[thread.thread_type];
                const proximity = getResolutionProximity(thread.resolution_proximity);
                
                return (
                  <div 
                    key={thread.id} 
                    className="thread-card active"
                    onClick={() => setSelectedThread(
                      selectedThread?.id === thread.id ? null : thread
                    )}
                  >
                    <div className="thread-header">
                      <div className="thread-title">
                        <span className="thread-icon" style={{ color: typeInfo.color }}>
                          {typeInfo.icon}
                        </span>
                        <h5>{thread.name}</h5>
                      </div>
                      <div className="thread-status">
                        {getStatusIcon(thread.status)}
                      </div>
                    </div>

                    <p className="thread-description">{thread.description}</p>

                    <div className="thread-metrics">
                      <div className="metric">
                        <label>Tension:</label>
                        {getTensionBar(thread.tension_level)}
                        <span>{(thread.tension_level * 100).toFixed(0)}%</span>
                      </div>
                      <div className="metric">
                        <label>Progress:</label>
                        <span className={`proximity-badge ${proximity.class}`}>
                          {proximity.text}
                        </span>
                      </div>
                    </div>

                    {thread.key_participants?.length > 0 && (
                      <div className="thread-participants">
                        <FaUsers />
                        {thread.key_participants.slice(0, 3).map(p => (
                          <span key={p} className="participant">{p}</span>
                        ))}
                        {thread.key_participants.length > 3 && (
                          <span className="more">+{thread.key_participants.length - 3} more</span>
                        )}
                      </div>
                    )}

                    {selectedThread?.id === thread.id && isDM && (
                      <div className="thread-actions">
                        <button 
                          className="btn-action"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleUpdateThreadStatus(thread.id, 'dormant');
                          }}
                        >
                          <FaSnowflake /> Make Dormant
                        </button>
                        <button 
                          className="btn-action success"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleUpdateThreadStatus(thread.id, 'resolved');
                          }}
                        >
                          <FaCheckCircle /> Resolve
                        </button>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {dormantThreads.length > 0 && (
          <div className="thread-section">
            <h4><FaSnowflake /> Dormant Threads</h4>
            <div className="thread-list dormant">
              {dormantThreads.map(thread => {
                const typeInfo = threadTypes[thread.thread_type];
                
                return (
                  <div key={thread.id} className="thread-card dormant">
                    <div className="thread-header">
                      <div className="thread-title">
                        <span className="thread-icon" style={{ color: typeInfo.color }}>
                          {typeInfo.icon}
                        </span>
                        <h5>{thread.name}</h5>
                      </div>
                      {isDM && (
                        <button 
                          className="btn-reactivate"
                          onClick={() => handleUpdateThreadStatus(thread.id, 'active')}
                          title="Reactivate Thread"
                        >
                          <FaFire />
                        </button>
                      )}
                    </div>
                    <p className="thread-description">{thread.description}</p>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {resolvedThreads.length > 0 && (
          <div className="thread-section">
            <h4><FaCheckCircle /> Resolved Threads</h4>
            <div className="thread-list resolved">
              {resolvedThreads.slice(0, 5).map(thread => {
                const typeInfo = threadTypes[thread.thread_type];
                
                return (
                  <div key={thread.id} className="thread-card resolved">
                    <div className="thread-header">
                      <div className="thread-title">
                        <span className="thread-icon" style={{ color: typeInfo.color }}>
                          {typeInfo.icon}
                        </span>
                        <h5>{thread.name}</h5>
                      </div>
                      <span className="resolution-date">
                        {thread.resolved_at && 
                          new Date(thread.resolved_at).toLocaleDateString()
                        }
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>

      <div className="threads-info">
        <FaLink />
        <p>
          Narrative threads connect events across sessions, creating a living story. 
          The AI tracks tension levels and weaves these threads into gameplay, 
          bringing them to the forefront when dramatically appropriate.
        </p>
      </div>
    </div>
  );
};

export default StoryThreads;