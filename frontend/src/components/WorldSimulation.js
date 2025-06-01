import React, { useState, useEffect } from 'react';
import { Tab, Tabs, TabList, TabPanel } from 'react-tabs';
import 'react-tabs/style/react-tabs.css';
import {
  FaGlobe, FaHistory, FaUsers, FaScroll, FaChartLine,
  FaClock, FaBrain, FaTheaterMasks, FaLanguage, FaExclamationCircle
} from 'react-icons/fa';
import api from '../services/api';
import WorldEventsFeed from './WorldSimulation/WorldEventsFeed';
import NPCAutonomy from './WorldSimulation/NPCAutonomy';
import FactionPersonalities from './WorldSimulation/FactionPersonalities';
import CultureExplorer from './WorldSimulation/CultureExplorer';
import SimulationControls from './WorldSimulation/SimulationControls';
import WorldTimeline from './WorldSimulation/WorldTimeline';
import '../styles/world-simulation.css';

const WorldSimulation = ({ sessionId, isDM }) => {
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState(0);
  const [worldState, setWorldState] = useState(null);
  const [recentEvents, setRecentEvents] = useState([]);
  const [simulationLogs, setSimulationLogs] = useState([]);
  const [error, setError] = useState(null);
  const [lastSimulated, setLastSimulated] = useState(null);

  // Load world state on mount
  useEffect(() => {
    loadWorldState();
    loadRecentEvents();
    if (isDM) {
      loadSimulationLogs();
    }
  }, [sessionId]);

  const loadWorldState = async () => {
    try {
      const response = await api.get(`/sessions/${sessionId}/world/state`);
      setWorldState(response.data);
      setLastSimulated(new Date(response.data.last_simulated));
    } catch (err) {
      console.error('Failed to load world state:', err);
    }
  };

  const loadRecentEvents = async () => {
    try {
      const response = await api.get(`/sessions/${sessionId}/world/events`, {
        params: { limit: 20, visible: !isDM }
      });
      setRecentEvents(response.data);
    } catch (err) {
      console.error('Failed to load world events:', err);
    }
  };

  const loadSimulationLogs = async () => {
    try {
      const response = await api.get(`/sessions/${sessionId}/world/logs`);
      setSimulationLogs(response.data);
    } catch (err) {
      console.error('Failed to load simulation logs:', err);
    }
  };

  const simulateWorld = async () => {
    if (!isDM) return;

    setLoading(true);
    setError(null);

    try {
      const response = await api.post(`/sessions/${sessionId}/world/simulate`);
      setRecentEvents(response.data.events);
      await loadWorldState();
      await loadSimulationLogs();
    } catch (err) {
      setError('Failed to simulate world progress');
      console.error('Simulation error:', err);
    } finally {
      setLoading(false);
    }
  };

  const getTimeSinceSimulation = () => {
    if (!lastSimulated) return 'Never';
    
    const now = new Date();
    const diff = now - lastSimulated;
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const days = Math.floor(hours / 24);
    
    if (days > 0) return `${days} day${days > 1 ? 's' : ''} ago`;
    if (hours > 0) return `${hours} hour${hours > 1 ? 's' : ''} ago`;
    return 'Recently';
  };

  return (
    <div className="world-simulation">
      <div className="world-simulation-header">
        <div className="header-info">
          <h2><FaGlobe /> Living World</h2>
          <p className="subtitle">
            The world continues to evolve, even in your absence
          </p>
        </div>

        <div className="simulation-status">
          <div className="status-item">
            <FaClock />
            <span>Last Simulated: {getTimeSinceSimulation()}</span>
          </div>
          {worldState && (
            <div className="status-item">
              <FaChartLine />
              <span>World Time: Day {Math.floor((new Date() - new Date(worldState.created_at)) / (1000 * 60 * 60 * 24))}</span>
            </div>
          )}
        </div>

        {isDM && (
          <SimulationControls
            onSimulate={simulateWorld}
            loading={loading}
            lastSimulated={lastSimulated}
          />
        )}
      </div>

      {error && (
        <div className="error-banner">
          <FaExclamationCircle /> {error}
          <button onClick={() => setError(null)}>&times;</button>
        </div>
      )}

      <Tabs selectedIndex={activeTab} onSelect={setActiveTab}>
        <TabList>
          <Tab><FaHistory /> World Events</Tab>
          <Tab><FaUsers /> Living NPCs</Tab>
          <Tab><FaBrain /> Faction Politics</Tab>
          <Tab><FaTheaterMasks /> Cultures</Tab>
          {isDM && <Tab><FaScroll /> Simulation Logs</Tab>}
        </TabList>

        <TabPanel>
          <div className="world-events-panel">
            <WorldEventsFeed 
              events={recentEvents}
              sessionId={sessionId}
              isDM={isDM}
              onEventCreated={(event) => setRecentEvents([event, ...recentEvents])}
            />
            <WorldTimeline
              events={recentEvents}
              sessionId={sessionId}
            />
          </div>
        </TabPanel>

        <TabPanel>
          <NPCAutonomy
            sessionId={sessionId}
            isDM={isDM}
          />
        </TabPanel>

        <TabPanel>
          <FactionPersonalities
            sessionId={sessionId}
            isDM={isDM}
          />
        </TabPanel>

        <TabPanel>
          <CultureExplorer
            sessionId={sessionId}
            isDM={isDM}
          />
        </TabPanel>

        {isDM && (
          <TabPanel>
            <div className="simulation-logs">
              <h3>Simulation History</h3>
              {simulationLogs.length === 0 ? (
                <p className="empty-state">No simulations run yet</p>
              ) : (
                <div className="logs-list">
                  {simulationLogs.map(log => (
                    <div key={log.id} className={`log-entry ${log.success ? 'success' : 'error'}`}>
                      <div className="log-header">
                        <span className="log-type">{log.simulation_type}</span>
                        <span className="log-time">
                          {new Date(log.start_time).toLocaleString()}
                        </span>
                      </div>
                      <div className="log-details">
                        <div className="log-stat">
                          <span>Events Created:</span>
                          <strong>{log.events_created}</strong>
                        </div>
                        <div className="log-stat">
                          <span>Duration:</span>
                          <strong>
                            {Math.round((new Date(log.end_time) - new Date(log.start_time)) / 1000)}s
                          </strong>
                        </div>
                        {log.error_message && (
                          <div className="log-error">
                            Error: {log.error_message}
                          </div>
                        )}
                      </div>
                      {log.details && Object.keys(log.details).length > 0 && (
                        <div className="log-breakdown">
                          {Object.entries(log.details).map(([key, value]) => (
                            <div key={key} className="breakdown-item">
                              <span>{key.replace(/_/g, ' ')}:</span>
                              <strong>{typeof value === 'object' ? JSON.stringify(value) : value}</strong>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </TabPanel>
        )}
      </Tabs>
    </div>
  );
};

export default WorldSimulation;