import React, { useState, useEffect } from 'react';
import { 
  FaUser, FaTarget, FaClock, FaMapMarkerAlt, 
  FaChartLine, FaPlus, FaEdit, FaCheckCircle
} from 'react-icons/fa';
import api from '../../services/api';
import { getSelectableProps, getClickableProps } from '../../utils/accessibility';

const NPCAutonomy = ({ sessionId, isDM }) => {
  const [npcs, setNpcs] = useState([]);
  const [selectedNPC, setSelectedNPC] = useState(null);
  const [npcGoals, setNpcGoals] = useState([]);
  const [npcSchedule, setNpcSchedule] = useState([]);
  const [loading, setLoading] = useState(false);
  const [showGoalModal, setShowGoalModal] = useState(false);
  const [newGoal, setNewGoal] = useState({
    goal_type: 'acquire_wealth',
    priority: 3,
    description: ''
  });

  useEffect(() => {
    loadNPCs();
  }, [sessionId]);

  useEffect(() => {
    if (selectedNPC) {
      loadNPCDetails(selectedNPC.id);
    }
  }, [selectedNPC]);

  const loadNPCs = async () => {
    try {
      const response = await api.get(`/npcs/session/${sessionId}`);
      setNpcs(response.data);
      if (response.data.length > 0 && !selectedNPC) {
        setSelectedNPC(response.data[0]);
      }
    } catch (err) {
      console.error('Failed to load NPCs:', err);
    }
  };

  const loadNPCDetails = async (npcId) => {
    setLoading(true);
    try {
      const [goalsRes, scheduleRes] = await Promise.all([
        api.get(`/npcs/${npcId}/goals`),
        api.get(`/npcs/${npcId}/schedule`)
      ]);
      setNpcGoals(goalsRes.data);
      setNpcSchedule(scheduleRes.data);
    } catch (err) {
      console.error('Failed to load NPC details:', err);
    } finally {
      setLoading(false);
    }
  };

  const createGoal = async () => {
    try {
      const response = await api.post(`/npcs/${selectedNPC.id}/goals`, newGoal);
      setNpcGoals([...npcGoals, response.data]);
      setShowGoalModal(false);
      setNewGoal({
        goal_type: 'acquire_wealth',
        priority: 3,
        description: ''
      });
    } catch (err) {
      console.error('Failed to create goal:', err);
    }
  };

  const getGoalIcon = (goalType) => {
    const icons = {
      acquire_wealth: '💰',
      gain_influence: '👑',
      improve_skill: '📚',
      complete_quest: '⚔️',
      build_relationship: '🤝',
      seek_knowledge: '🔮',
      gain_power: '💪',
      find_artifact: '🏺'
    };
    return icons[goalType] || '🎯';
  };

  const getTimeOfDayIcon = (timeOfDay) => {
    const icons = {
      morning: '🌅',
      afternoon: '☀️',
      evening: '🌆',
      night: '🌙'
    };
    return icons[timeOfDay] || '🕐';
  };

  const getGoalStatusColor = (status) => {
    const colors = {
      active: '#3498db',
      completed: '#27ae60',
      failed: '#e74c3c',
      abandoned: '#95a5a6'
    };
    return colors[status] || '#34495e';
  };

  return (
    <div className="npc-autonomy">
      <div className="npc-list">
        <h3>Autonomous NPCs</h3>
        <div className="npc-cards">
          {npcs.map(npc => (
            <div 
              key={npc.id}
              className={`npc-card ${selectedNPC?.id === npc.id ? 'selected' : ''}`}
              {...getSelectableProps(
                () => setSelectedNPC(npc),
                selectedNPC?.id === npc.id
              )}
            >
              <div className="npc-avatar">
                <FaUser />
              </div>
              <div className="npc-info">
                <h4>{npc.name}</h4>
                <p className="npc-type">{npc.type}</p>
                <p className="npc-role">{npc.role || 'Commoner'}</p>
              </div>
            </div>
          ))}
        </div>
      </div>

      {selectedNPC && (
        <div className="npc-details">
          <div className="detail-header">
            <h3>{selectedNPC.name}'s Autonomous Life</h3>
            <p className="npc-description">{selectedNPC.description}</p>
          </div>

          <div className="npc-goals">
            <div className="section-header">
              <h4><FaTarget /> Personal Goals</h4>
              {isDM && (
                <button 
                  className="btn-add"
                  onClick={() => setShowGoalModal(true)}
                >
                  <FaPlus /> Add Goal
                </button>
              )}
            </div>

            {loading ? (
              <div className="loading">Loading goals...</div>
            ) : npcGoals.length === 0 ? (
              <p className="empty-state">No personal goals set</p>
            ) : (
              <div className="goals-list">
                {npcGoals.map(goal => (
                  <div key={goal.id} className="goal-card">
                    <div className="goal-icon">{getGoalIcon(goal.goal_type)}</div>
                    <div className="goal-content">
                      <div className="goal-header">
                        <h5>{goal.description}</h5>
                        <span 
                          className="goal-status"
                          style={{ color: getGoalStatusColor(goal.status) }}
                        >
                          {goal.status}
                        </span>
                      </div>
                      <div className="goal-progress">
                        <div className="progress-bar">
                          <div 
                            className="progress-fill"
                            style={{ 
                              width: `${goal.progress * 100}%`,
                              backgroundColor: getGoalStatusColor(goal.status)
                            }}
                          />
                        </div>
                        <span className="progress-text">{Math.round(goal.progress * 100)}%</span>
                      </div>
                      <div className="goal-meta">
                        <span className="priority">Priority: {goal.priority}</span>
                        <span className="goal-type">{goal.goal_type.replace(/_/g, ' ')}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="npc-schedule">
            <h4><FaClock /> Daily Schedule</h4>
            {npcSchedule.length === 0 ? (
              <p className="empty-state">No schedule defined</p>
            ) : (
              <div className="schedule-timeline">
                {npcSchedule.map(schedule => (
                  <div key={schedule.id} className="schedule-item">
                    <div className="time-icon">
                      {getTimeOfDayIcon(schedule.time_of_day)}
                    </div>
                    <div className="schedule-content">
                      <div className="schedule-time">{schedule.time_of_day}</div>
                      <div className="schedule-activity">{schedule.activity}</div>
                      <div className="schedule-location">
                        <FaMapMarkerAlt /> {schedule.location}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="npc-stats">
            <h4><FaChartLine /> Character Stats</h4>
            <div className="stats-grid">
              <div className="stat">
                <span className="stat-label">Level</span>
                <span className="stat-value">{selectedNPC.level}</span>
              </div>
              <div className="stat">
                <span className="stat-label">HP</span>
                <span className="stat-value">{selectedNPC.current_hp}/{selectedNPC.max_hp}</span>
              </div>
              <div className="stat">
                <span className="stat-label">AC</span>
                <span className="stat-value">{selectedNPC.armor_class}</span>
              </div>
              <div className="stat">
                <span className="stat-label">CR</span>
                <span className="stat-value">{selectedNPC.challenge_rating || '0'}</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {showGoalModal && (
        <div className="modal-overlay" {...getClickableProps(() => setShowGoalModal(false))}>
          <div className="modal-content" {...getClickableProps(e => e.stopPropagation())}>
            <h3>Create NPC Goal</h3>
            
            <div className="form-group">
              <label>Goal Type</label>
              <select
                value={newGoal.goal_type}
                onChange={(e) => setNewGoal({...newGoal, goal_type: e.target.value})}
              >
                <option value="acquire_wealth">Acquire Wealth</option>
                <option value="gain_influence">Gain Influence</option>
                <option value="improve_skill">Improve Skill</option>
                <option value="complete_quest">Complete Quest</option>
                <option value="build_relationship">Build Relationship</option>
                <option value="seek_knowledge">Seek Knowledge</option>
                <option value="gain_power">Gain Power</option>
                <option value="find_artifact">Find Artifact</option>
              </select>
            </div>

            <div className="form-group">
              <label>Description</label>
              <textarea
                value={newGoal.description}
                onChange={(e) => setNewGoal({...newGoal, description: e.target.value})}
                placeholder="Describe the specific goal..."
                rows="3"
              />
            </div>

            <div className="form-group">
              <label>Priority (1-5)</label>
              <input
                type="number"
                min="1"
                max="5"
                value={newGoal.priority}
                onChange={(e) => setNewGoal({...newGoal, priority: parseInt(e.target.value)})}
              />
            </div>

            <div className="modal-actions">
              <button onClick={() => setShowGoalModal(false)}>Cancel</button>
              <button 
                onClick={createGoal}
                disabled={!newGoal.description}
                className="btn-primary"
              >
                Create Goal
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default NPCAutonomy;