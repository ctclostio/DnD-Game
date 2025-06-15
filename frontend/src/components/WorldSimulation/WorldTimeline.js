import React, { useState, useEffect } from 'react';
import { FaCalendarAlt, FaFilter } from 'react-icons/fa';
import { getClickableProps, getSelectableProps } from '../../utils/accessibility';

const WorldTimeline = ({ events, sessionId }) => {
  const [groupedEvents, setGroupedEvents] = useState({});
  const [expandedDays, setExpandedDays] = useState(new Set());
  const [timeRange, setTimeRange] = useState('week'); // week, month, all

  useEffect(() => {
    groupEventsByDay();
  }, [events, timeRange]);

  const groupEventsByDay = () => {
    const now = new Date();
    const cutoffDate = getCutoffDate(now, timeRange);
    
    const grouped = {};
    
    events
      .filter(event => new Date(event.occurred_at) >= cutoffDate)
      .forEach(event => {
        const date = new Date(event.occurred_at);
        const dayKey = date.toLocaleDateString();
        
        if (!grouped[dayKey]) {
          grouped[dayKey] = {
            date: date,
            events: []
          };
        }
        
        grouped[dayKey].events.push(event);
      });

    // Sort events within each day
    Object.values(grouped).forEach(day => {
      day.events.sort((a, b) => 
        new Date(b.occurred_at) - new Date(a.occurred_at)
      );
    });

    setGroupedEvents(grouped);
  };

  const getCutoffDate = (now, range) => {
    const cutoff = new Date(now);
    
    switch (range) {
      case 'week':
        cutoff.setDate(cutoff.getDate() - 7);
        break;
      case 'month':
        cutoff.setMonth(cutoff.getMonth() - 1);
        break;
      case 'all':
        cutoff.setFullYear(cutoff.getFullYear() - 10);
        break;
    }
    
    return cutoff;
  };

  const toggleDay = (dayKey) => {
    const newExpanded = new Set(expandedDays);
    if (newExpanded.has(dayKey)) {
      newExpanded.delete(dayKey);
    } else {
      newExpanded.add(dayKey);
    }
    setExpandedDays(newExpanded);
  };

  const getEventIcon = (eventType) => {
    const icons = {
      npc_goal_progress: '🎯',
      npc_activity: '👤',
      economic_event: '💰',
      political_milestone: '🏛️',
      political_opportunity: '⚖️',
      faction_interaction: '🤝',
      natural_event: '🌍',
      cultural_shift: '🎭',
      player_action: '🎮'
    };
    return icons[eventType] || '📜';
  };

  const formatDayHeader = (date) => {
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);
    
    if (date.toDateString() === today.toDateString()) {
      return 'Today';
    } else if (date.toDateString() === yesterday.toDateString()) {
      return 'Yesterday';
    } else {
      const options = { weekday: 'long', month: 'short', day: 'numeric' };
      return date.toLocaleDateString(undefined, options);
    }
  };

  const sortedDays = Object.entries(groupedEvents)
    .sort(([,a], [,b]) => b.date - a.date);

  return (
    <div className="world-timeline">
      <div className="timeline-header">
        <h3><FaCalendarAlt /> World Timeline</h3>
        <div className="timeline-filters">
          <button 
            className={timeRange === 'week' ? 'active' : ''}
            onClick={() => setTimeRange('week')}
          >
            Past Week
          </button>
          <button 
            className={timeRange === 'month' ? 'active' : ''}
            onClick={() => setTimeRange('month')}
          >
            Past Month
          </button>
          <button 
            className={timeRange === 'all' ? 'active' : ''}
            onClick={() => setTimeRange('all')}
          >
            All Time
          </button>
        </div>
      </div>

      <div className="timeline-content">
        {sortedDays.length === 0 ? (
          <div className="empty-state">
            <FaCalendarAlt />
            <p>No events in this time period</p>
          </div>
        ) : (
          sortedDays.map(([dayKey, dayData]) => (
            <div key={dayKey} className="timeline-day">
              <div 
                className="day-header"
                {...getClickableProps(() => toggleDay(dayKey))}
              >
                <div className="day-info">
                  <h4>{formatDayHeader(dayData.date)}</h4>
                  <span className="event-count">{dayData.events.length} events</span>
                </div>
                <div className="day-summary">
                  {dayData.events.slice(0, 3).map((event, idx) => (
                    <span key={idx} className="event-preview">
                      {getEventIcon(event.event_type)}
                    </span>
                  ))}
                  {dayData.events.length > 3 && (
                    <span className="more-events">+{dayData.events.length - 3}</span>
                  )}
                </div>
              </div>

              {expandedDays.has(dayKey) && (
                <div className="day-events">
                  {dayData.events.map(event => (
                    <div key={event.id} className="timeline-event">
                      <div className="event-time">
                        {new Date(event.occurred_at).toLocaleTimeString([], {
                          hour: '2-digit',
                          minute: '2-digit'
                        })}
                      </div>
                      <div className="event-icon">
                        {getEventIcon(event.event_type)}
                      </div>
                      <div className="event-details">
                        <h5>{event.title}</h5>
                        <p>{event.description}</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default WorldTimeline;