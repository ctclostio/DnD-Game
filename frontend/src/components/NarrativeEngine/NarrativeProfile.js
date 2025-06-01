import React, { useState } from 'react';
import { FaBrain, FaEdit, FaSave, FaTimes } from 'react-icons/fa';
import api from '../../services/api';

const NarrativeProfile = ({ profile, characterId, onProfileUpdate }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [editedProfile, setEditedProfile] = useState(profile || {
    preferences: {
      themes: [],
      tone: [],
      complexity: 3,
      moral_alignment: 'neutral',
      pacing: 'moderate',
      combat_narrative: 0.5
    },
    play_style: 'balanced'
  });

  const themeOptions = [
    'redemption', 'revenge', 'discovery', 'sacrifice', 'power',
    'friendship', 'betrayal', 'mystery', 'survival', 'destiny',
    'corruption', 'hope', 'tragedy', 'comedy', 'romance'
  ];

  const toneOptions = [
    'dark', 'heroic', 'comedic', 'tragic', 'epic',
    'gritty', 'whimsical', 'mysterious', 'hopeful', 'dramatic'
  ];

  const playStyleOptions = [
    'combat-focused', 'roleplay-heavy', 'explorer', 'strategist',
    'problem-solver', 'social', 'balanced', 'chaotic', 'methodical'
  ];

  const handleSave = async () => {
    try {
      const response = await api.put(`/narrative/profile/${characterId}`, editedProfile);
      onProfileUpdate(response.data);
      setIsEditing(false);
    } catch (error) {
      console.error('Failed to update profile:', error);
    }
  };

  const toggleTheme = (theme) => {
    const themes = editedProfile.preferences.themes || [];
    const newThemes = themes.includes(theme)
      ? themes.filter(t => t !== theme)
      : [...themes, theme];
    
    setEditedProfile({
      ...editedProfile,
      preferences: {
        ...editedProfile.preferences,
        themes: newThemes
      }
    });
  };

  const toggleTone = (tone) => {
    const tones = editedProfile.preferences.tone || [];
    const newTones = tones.includes(tone)
      ? tones.filter(t => t !== tone)
      : [...tones, tone];
    
    setEditedProfile({
      ...editedProfile,
      preferences: {
        ...editedProfile.preferences,
        tone: newTones
      }
    });
  };

  if (!profile && !isEditing) {
    return (
      <div className="narrative-profile empty-state">
        <FaBrain />
        <h4>No Narrative Profile</h4>
        <p>Create a narrative profile to personalize your storytelling experience</p>
        <button className="btn-primary" onClick={() => setIsEditing(true)}>
          Create Profile
        </button>
      </div>
    );
  }

  return (
    <div className="narrative-profile">
      <div className="profile-header">
        <h3><FaBrain /> Narrative Profile</h3>
        {!isEditing ? (
          <button className="btn-edit" onClick={() => setIsEditing(true)}>
            <FaEdit /> Edit
          </button>
        ) : (
          <div className="edit-actions">
            <button className="btn-save" onClick={handleSave}>
              <FaSave /> Save
            </button>
            <button className="btn-cancel" onClick={() => {
              setIsEditing(false);
              setEditedProfile(profile);
            }}>
              <FaTimes /> Cancel
            </button>
          </div>
        )}
      </div>

      <div className="profile-content">
        {/* Play Style */}
        <div className="profile-section">
          <h4>Play Style</h4>
          {!isEditing ? (
            <div className="play-style-display">
              <span className="style-badge">{profile?.play_style || 'Not Set'}</span>
            </div>
          ) : (
            <select
              value={editedProfile.play_style}
              onChange={(e) => setEditedProfile({ ...editedProfile, play_style: e.target.value })}
              className="style-selector"
            >
              {playStyleOptions.map(style => (
                <option key={style} value={style}>
                  {style.charAt(0).toUpperCase() + style.slice(1).replace('-', ' ')}
                </option>
              ))}
            </select>
          )}
        </div>

        {/* Story Themes */}
        <div className="profile-section">
          <h4>Preferred Story Themes</h4>
          {!isEditing ? (
            <div className="theme-tags">
              {profile?.preferences?.themes?.length > 0 ? (
                profile.preferences.themes.map(theme => (
                  <span key={theme} className="theme-tag">{theme}</span>
                ))
              ) : (
                <span className="no-data">No themes selected</span>
              )}
            </div>
          ) : (
            <div className="theme-selector">
              {themeOptions.map(theme => (
                <label key={theme} className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={editedProfile.preferences.themes?.includes(theme) || false}
                    onChange={() => toggleTheme(theme)}
                  />
                  {theme}
                </label>
              ))}
            </div>
          )}
        </div>

        {/* Narrative Tone */}
        <div className="profile-section">
          <h4>Preferred Narrative Tone</h4>
          {!isEditing ? (
            <div className="tone-tags">
              {profile?.preferences?.tone?.length > 0 ? (
                profile.preferences.tone.map(tone => (
                  <span key={tone} className="tone-tag">{tone}</span>
                ))
              ) : (
                <span className="no-data">No tones selected</span>
              )}
            </div>
          ) : (
            <div className="tone-selector">
              {toneOptions.map(tone => (
                <label key={tone} className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={editedProfile.preferences.tone?.includes(tone) || false}
                    onChange={() => toggleTone(tone)}
                  />
                  {tone}
                </label>
              ))}
            </div>
          )}
        </div>

        {/* Complexity and Pacing */}
        <div className="profile-section">
          <h4>Story Preferences</h4>
          {!isEditing ? (
            <div className="preferences-display">
              <div className="preference-item">
                <label>Complexity:</label>
                <div className="complexity-bar">
                  <div 
                    className="complexity-fill" 
                    style={{ width: `${(profile?.preferences?.complexity || 3) * 20}%` }}
                  />
                </div>
                <span>{profile?.preferences?.complexity || 3}/5</span>
              </div>
              <div className="preference-item">
                <label>Pacing:</label>
                <span className="pacing-value">{profile?.preferences?.pacing || 'moderate'}</span>
              </div>
              <div className="preference-item">
                <label>Combat vs Narrative:</label>
                <div className="balance-bar">
                  <div 
                    className="balance-fill" 
                    style={{ width: `${(profile?.preferences?.combat_narrative || 0.5) * 100}%` }}
                  />
                </div>
                <span>{Math.round((profile?.preferences?.combat_narrative || 0.5) * 100)}% Combat</span>
              </div>
            </div>
          ) : (
            <div className="preferences-editor">
              <div className="preference-control">
                <label>
                  Complexity (1-5):
                  <input
                    type="range"
                    min="1"
                    max="5"
                    value={editedProfile.preferences.complexity}
                    onChange={(e) => setEditedProfile({
                      ...editedProfile,
                      preferences: {
                        ...editedProfile.preferences,
                        complexity: parseInt(e.target.value)
                      }
                    })}
                  />
                  <span>{editedProfile.preferences.complexity}</span>
                </label>
              </div>
              <div className="preference-control">
                <label>
                  Pacing:
                  <select
                    value={editedProfile.preferences.pacing}
                    onChange={(e) => setEditedProfile({
                      ...editedProfile,
                      preferences: {
                        ...editedProfile.preferences,
                        pacing: e.target.value
                      }
                    })}
                  >
                    <option value="fast">Fast</option>
                    <option value="moderate">Moderate</option>
                    <option value="slow-burn">Slow Burn</option>
                  </select>
                </label>
              </div>
              <div className="preference-control">
                <label>
                  Combat Focus (0-100%):
                  <input
                    type="range"
                    min="0"
                    max="100"
                    value={editedProfile.preferences.combat_narrative * 100}
                    onChange={(e) => setEditedProfile({
                      ...editedProfile,
                      preferences: {
                        ...editedProfile.preferences,
                        combat_narrative: parseInt(e.target.value) / 100
                      }
                    })}
                  />
                  <span>{Math.round(editedProfile.preferences.combat_narrative * 100)}%</span>
                </label>
              </div>
            </div>
          )}
        </div>

        {/* Moral Alignment */}
        <div className="profile-section">
          <h4>Moral Decision Tendency</h4>
          {!isEditing ? (
            <div className="moral-display">
              <span className="moral-badge">{profile?.preferences?.moral_alignment || 'neutral'}</span>
            </div>
          ) : (
            <select
              value={editedProfile.preferences.moral_alignment}
              onChange={(e) => setEditedProfile({
                ...editedProfile,
                preferences: {
                  ...editedProfile.preferences,
                  moral_alignment: e.target.value
                }
              })}
              className="moral-selector"
            >
              <option value="lawful_good">Lawful Good</option>
              <option value="neutral_good">Neutral Good</option>
              <option value="chaotic_good">Chaotic Good</option>
              <option value="lawful_neutral">Lawful Neutral</option>
              <option value="neutral">True Neutral</option>
              <option value="chaotic_neutral">Chaotic Neutral</option>
              <option value="lawful_evil">Lawful Evil</option>
              <option value="neutral_evil">Neutral Evil</option>
              <option value="chaotic_evil">Chaotic Evil</option>
            </select>
          )}
        </div>

        {/* Decision History Summary */}
        {profile?.decision_history?.length > 0 && (
          <div className="profile-section">
            <h4>Decision History</h4>
            <div className="decision-summary">
              <p>{profile.decision_history.length} significant decisions recorded</p>
              <div className="recent-decisions">
                {profile.decision_history.slice(-3).map((decision, index) => (
                  <div key={index} className="decision-item">
                    <span className="decision-context">{decision.context}</span>
                    <span className="decision-choice">→ {decision.decision}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default NarrativeProfile;