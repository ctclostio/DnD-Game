import React, { useState } from 'react';
import { FaPlay, FaCog, FaClock, FaExclamationTriangle } from 'react-icons/fa';

const SimulationControls = ({ onSimulate, loading, lastSimulated }) => {
  const [showSettings, setShowSettings] = useState(false);
  const [autoSimulate, setAutoSimulate] = useState(false);
  const [simulationSpeed, setSimulationSpeed] = useState('normal');

  const canSimulate = () => {
    if (!lastSimulated) return true;
    
    const hoursSinceLastSimulation = (new Date() - lastSimulated) / (1000 * 60 * 60);
    
    // Require at least 1 hour between simulations
    return hoursSinceLastSimulation >= 1;
  };

  const getSimulationWarning = () => {
    if (!lastSimulated) return null;
    
    const hoursSinceLastSimulation = (new Date() - lastSimulated) / (1000 * 60 * 60);
    
    if (hoursSinceLastSimulation < 1) {
      const minutesRemaining = Math.ceil((1 - hoursSinceLastSimulation) * 60);
      return `Please wait ${minutesRemaining} minute${minutesRemaining > 1 ? 's' : ''} before simulating again`;
    }
    
    if (hoursSinceLastSimulation > 24) {
      return 'Warning: Large time gap may result in significant world changes';
    }
    
    return null;
  };

  return (
    <div className="simulation-controls">
      <button 
        className="btn-simulate"
        onClick={onSimulate}
        disabled={loading || !canSimulate()}
      >
        {loading ? (
          <>
            <FaClock className="spinning" /> Simulating...
          </>
        ) : (
          <>
            <FaPlay /> Simulate World
          </>
        )}
      </button>

      <button 
        className="btn-settings"
        onClick={() => setShowSettings(!showSettings)}
      >
        <FaCog />
      </button>

      {getSimulationWarning() && (
        <div className="simulation-warning">
          <FaExclamationTriangle /> {getSimulationWarning()}
        </div>
      )}

      {showSettings && (
        <div className="simulation-settings">
          <h4>Simulation Settings</h4>
          
          <div className="setting-group">
            <label>
              <input
                type="checkbox"
                checked={autoSimulate}
                onChange={(e) => setAutoSimulate(e.target.checked)}
              />
              Auto-simulate when players are offline
            </label>
          </div>

          <div className="setting-group">
            <label>Simulation Speed</label>
            <select
              value={simulationSpeed}
              onChange={(e) => setSimulationSpeed(e.target.value)}
            >
              <option value="slow">Slow (1 day = 1 real day)</option>
              <option value="normal">Normal (1 day = 4 hours)</option>
              <option value="fast">Fast (1 day = 1 hour)</option>
            </select>
          </div>

          <div className="setting-info">
            <p>Current world speed: <strong>{simulationSpeed}</strong></p>
            <p>Events will occur based on the selected time scale</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default SimulationControls;