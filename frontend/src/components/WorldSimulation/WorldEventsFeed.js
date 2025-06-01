import React, { useState } from 'react';
import { 
  FaScroll, FaUsers, FaCoins, FaSwords, FaTree, 
  FaMagic, FaPlus, FaFilter, FaGlobe
} from 'react-icons/fa';
import api from '../../services/api';

const WorldEventsFeed = ({ events, sessionId, isDM, onEventCreated }) => {
  const [filter, setFilter] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newEvent, setNewEvent] = useState({
    event_type: 'political_opportunity',
    title: '',
    description: '',
    is_player_visible: true,
    affected_entities: []
  });

  const eventIcons = {
    npc_goal_progress: <FaUsers />,
    npc_activity: <FaUsers />,
    economic_event: <FaCoins />,
    political_milestone: <FaScroll />,
    political_opportunity: <FaScroll />,
    faction_interaction: <FaSwords />,
    natural_event: <FaTree />,
    cultural_shift: <FaMagic />,
    player_action: <FaGlobe />
  };

  const eventColors = {
    npc_goal_progress: '#3498db',
    npc_activity: '#2ecc71',
    economic_event: '#f39c12',
    political_milestone: '#9b59b6',
    political_opportunity: '#e74c3c',
    faction_interaction: '#e67e22',
    natural_event: '#1abc9c',
    cultural_shift: '#34495e',
    player_action: '#f1c40f'
  };

  const filteredEvents = filter === 'all' 
    ? events 
    : events.filter(event => event.event_type.includes(filter));

  const createWorldEvent = async () => {
    try {
      const response = await api.post(`/sessions/${sessionId}/world/events`, newEvent);
      onEventCreated(response.data);
      setShowCreateModal(false);
      setNewEvent({
        event_type: 'political_opportunity',
        title: '',
        description: '',
        is_player_visible: true,
        affected_entities: []
      });
    } catch (err) {
      console.error('Failed to create event:', err);
    }
  };

  const getTimeAgo = (timestamp) => {
    const now = new Date();
    const eventTime = new Date(timestamp);
    const diff = now - eventTime;
    
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    
    if (days > 0) return `${days} day${days > 1 ? 's' : ''} ago`;
    if (hours > 0) return `${hours} hour${hours > 1 ? 's' : ''} ago`;
    if (minutes > 0) return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
    return 'Just now';
  };

  return (
    <div className="world-events-feed">
      <div className="feed-controls">
        <div className="filter-buttons">
          <button 
            className={filter === 'all' ? 'active' : ''}
            onClick={() => setFilter('all')}
          >
            All Events
          </button>
          <button 
            className={filter === 'npc' ? 'active' : ''}
            onClick={() => setFilter('npc')}
          >
            <FaUsers /> NPCs
          </button>
          <button 
            className={filter === 'economic' ? 'active' : ''}
            onClick={() => setFilter('economic')}
          >
            <FaCoins /> Economic
          </button>
          <button 
            className={filter === 'political' ? 'active' : ''}
            onClick={() => setFilter('political')}
          >
            <FaScroll /> Political
          </button>
          <button 
            className={filter === 'natural' ? 'active' : ''}
            onClick={() => setFilter('natural')}
          >
            <FaTree /> Natural
          </button>
        </div>

        {isDM && (
          <button 
            className="btn-create-event"
            onClick={() => setShowCreateModal(true)}
          >
            <FaPlus /> Create Event
          </button>
        )}
      </div>

      <div className="events-timeline">
        {filteredEvents.length === 0 ? (
          <div className="empty-state">
            <FaGlobe />
            <p>No world events to display</p>
          </div>
        ) : (
          filteredEvents.map(event => (
            <div 
              key={event.id} 
              className={`event-card ${!event.is_player_visible && isDM ? 'hidden-event' : ''}`}
            >
              <div className="event-icon" style={{ color: eventColors[event.event_type] || '#95a5a6' }}>
                {eventIcons[event.event_type] || <FaGlobe />}
              </div>
              
              <div className="event-content">
                <div className="event-header">
                  <h4>{event.title}</h4>
                  <span className="event-time">{getTimeAgo(event.occurred_at)}</span>
                </div>
                
                <p className="event-description">{event.description}</p>
                
                {event.consequences && event.consequences.length > 0 && (
                  <div className="event-consequences">
                    <strong>Consequences:</strong>
                    <ul>
                      {event.consequences.map((consequence, idx) => (
                        <li key={idx}>
                          {consequence.effect}: {consequence.magnitude > 0 ? '+' : ''}{consequence.magnitude}
                          {consequence.duration && ` (${consequence.duration})`}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
                
                {!event.is_player_visible && isDM && (
                  <div className="hidden-indicator">
                    Hidden from players
                  </div>
                )}
              </div>
            </div>
          ))
        )}
      </div>

      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>Create World Event</h3>
            
            <div className="form-group">
              <label>Event Type</label>
              <select
                value={newEvent.event_type}
                onChange={(e) => setNewEvent({...newEvent, event_type: e.target.value})}
              >
                <option value="political_opportunity">Political Opportunity</option>
                <option value="economic_event">Economic Event</option>
                <option value="natural_disaster">Natural Disaster</option>
                <option value="cultural_shift">Cultural Shift</option>
                <option value="military_conflict">Military Conflict</option>
                <option value="magical_anomaly">Magical Anomaly</option>
              </select>
            </div>

            <div className="form-group">
              <label>Title</label>
              <input
                type="text"
                value={newEvent.title}
                onChange={(e) => setNewEvent({...newEvent, title: e.target.value})}
                placeholder="A dramatic event title..."
              />
            </div>

            <div className="form-group">
              <label>Description</label>
              <textarea
                value={newEvent.description}
                onChange={(e) => setNewEvent({...newEvent, description: e.target.value})}
                placeholder="Describe what happened and its immediate effects..."
                rows="4"
              />
            </div>

            <div className="form-group">
              <label>
                <input
                  type="checkbox"
                  checked={newEvent.is_player_visible}
                  onChange={(e) => setNewEvent({...newEvent, is_player_visible: e.target.checked})}
                />
                Visible to Players
              </label>
            </div>

            <div className="modal-actions">
              <button onClick={() => setShowCreateModal(false)}>Cancel</button>
              <button 
                onClick={createWorldEvent}
                disabled={!newEvent.title || !newEvent.description}
                className="btn-primary"
              >
                Create Event
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default WorldEventsFeed;