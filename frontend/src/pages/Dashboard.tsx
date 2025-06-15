import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { RootState } from '@store/index';

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const user = useSelector((state: RootState) => state.auth.user);

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>Welcome, {user?.username}!</h1>
        <p className="role-badge">{user?.role?.toUpperCase()}</p>
      </div>

      <div className="dashboard-grid">
        <div className="dashboard-card" onClick={() => navigate('/characters')}>
          <h3>My Characters</h3>
          <p>View and manage your D&D characters</p>
        </div>

        <div className="dashboard-card" onClick={() => navigate('/characters/new')}>
          <h3>Create Character</h3>
          <p>Build a new character from scratch</p>
        </div>

        <div className="dashboard-card">
          <h3>Join Game</h3>
          <p>Enter a game session code to join</p>
        </div>

        {user?.role === 'dm' && (
          <>
            <div className="dashboard-card" onClick={() => navigate('/world-builder')}>
              <h3>World Builder</h3>
              <p>Create and manage your campaign world</p>
            </div>

            <div className="dashboard-card" onClick={() => navigate('/dm-tools')}>
              <h3>DM Tools</h3>
              <p>Access DM-specific utilities</p>
            </div>

            <div className="dashboard-card">
              <h3>Create Session</h3>
              <p>Start a new game session</p>
            </div>
          </>
        )}
      </div>

      <div className="recent-activity">
        <h2>Recent Activity</h2>
        <p>No recent activity to display.</p>
      </div>
    </div>
  );
};

export default Dashboard;
