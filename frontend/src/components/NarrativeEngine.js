import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Tab, Tabs, TabList, TabPanel } from 'react-tabs';
import 'react-tabs/style/react-tabs.css';
import { 
  FaBook, FaTheaterMasks, FaChessBoard, FaHistory, FaBrain,
  FaExclamationTriangle, FaEye, FaComments, FaCogs
} from 'react-icons/fa';
import api from '../services/api';
import BackstoryManager from './NarrativeEngine/BackstoryManager';
import ConsequenceTracker from './NarrativeEngine/ConsequenceTracker';
import PerspectiveViewer from './NarrativeEngine/PerspectiveViewer';
import StoryThreads from './NarrativeEngine/StoryThreads';
import CharacterMemories from './NarrativeEngine/CharacterMemories';
import NarrativeProfile from './NarrativeEngine/NarrativeProfile';
import '../styles/narrative-engine.css';

const NarrativeEngine = () => {
  const { sessionId } = useParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [activeTab, setActiveTab] = useState(0);
  const [narrativeData, setNarrativeData] = useState({
    profile: null,
    backstory: [],
    consequences: [],
    worldEvents: [],
    memories: [],
    threads: []
  });
  const [selectedCharacter, setSelectedCharacter] = useState(null);
  const [characters, setCharacters] = useState([]);
  const [isDM, setIsDM] = useState(false);

  useEffect(() => {
    fetchInitialData();
  }, [sessionId]);

  const fetchInitialData = async () => {
    try {
      setLoading(true);
      
      // Fetch session data to determine if user is DM
      const sessionResponse = await api.get(`/game/sessions/${sessionId}`);
      const session = sessionResponse.data;
      const currentUser = JSON.parse(localStorage.getItem('user'));
      setIsDM(session.dm_user_id === currentUser.id);

      // Fetch characters
      const charactersResponse = await api.get('/characters');
      setCharacters(charactersResponse.data);
      
      if (charactersResponse.data.length > 0) {
        setSelectedCharacter(charactersResponse.data[0]);
        await fetchCharacterNarrativeData(charactersResponse.data[0].id);
      }

      // Fetch session-wide narrative data
      await fetchSessionNarrativeData();
      
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const fetchCharacterNarrativeData = async (characterId) => {
    try {
      // Fetch narrative profile
      const profileResponse = await api.get(`/narrative/profile/${characterId}`);
      
      // Fetch backstory elements
      const backstoryResponse = await api.get(`/narrative/backstory/${characterId}`);
      
      // Fetch memories
      const memoriesResponse = await api.get(`/narrative/memory/${characterId}`);
      
      setNarrativeData(prev => ({
        ...prev,
        profile: profileResponse.data,
        backstory: backstoryResponse.data,
        memories: memoriesResponse.data
      }));
    } catch (err) {
      console.error('Failed to fetch character narrative data:', err);
    }
  };

  const fetchSessionNarrativeData = async () => {
    try {
      // Fetch pending consequences
      const consequencesResponse = await api.get(`/narrative/consequences/${sessionId}`);
      
      // Fetch active narrative threads
      const threadsResponse = await api.get('/narrative/threads');
      
      setNarrativeData(prev => ({
        ...prev,
        consequences: consequencesResponse.data,
        threads: threadsResponse.data.filter(thread => 
          thread.metadata?.session_id === sessionId || thread.status === 'active'
        )
      }));
    } catch (err) {
      console.error('Failed to fetch session narrative data:', err);
    }
  };

  const handleCharacterChange = async (e) => {
    const characterId = e.target.value;
    const character = characters.find(c => c.id === characterId);
    setSelectedCharacter(character);
    
    if (character) {
      await fetchCharacterNarrativeData(character.id);
    }
  };

  const handleRecordAction = async (action) => {
    try {
      await api.post('/narrative/action', {
        ...action,
        session_id: sessionId,
        character_id: selectedCharacter.id
      });
      
      // Refresh consequences after recording action
      setTimeout(() => {
        fetchSessionNarrativeData();
      }, 2000);
    } catch (err) {
      console.error('Failed to record action:', err);
    }
  };

  const handleCreateWorldEvent = async (event) => {
    try {
      const response = await api.post('/narrative/event', {
        ...event,
        metadata: { session_id: sessionId }
      });
      
      // Generate perspectives for the event
      if (event.generate_perspectives) {
        await api.post('/narrative/generate/perspectives', {
          event_id: response.data.id,
          sources: event.perspective_sources || [],
          session_id: sessionId
        });
      }
      
      return response.data;
    } catch (err) {
      console.error('Failed to create world event:', err);
      throw err;
    }
  };

  if (loading) {
    return (
      <div className="narrative-loading">
        <div className="loading-spinner"></div>
        <p>Loading narrative data...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="narrative-error">
        <FaExclamationTriangle />
        <p>Error loading narrative engine: {error}</p>
      </div>
    );
  }

  return (
    <div className="narrative-engine">
      <div className="narrative-header">
        <h2><FaBook /> Dynamic Storytelling Engine</h2>
        
        {characters.length > 0 && (
          <div className="character-selector">
            <label>Active Character:</label>
            <select value={selectedCharacter?.id || ''} onChange={handleCharacterChange}>
              {characters.map(char => (
                <option key={char.id} value={char.id}>
                  {char.name} - {char.class} {char.level}
                </option>
              ))}
            </select>
          </div>
        )}
      </div>

      <Tabs selectedIndex={activeTab} onSelect={setActiveTab}>
        <TabList>
          <Tab><FaBrain /> Profile</Tab>
          <Tab><FaBook /> Backstory</Tab>
          <Tab><FaChessBoard /> Consequences</Tab>
          <Tab><FaTheaterMasks /> Perspectives</Tab>
          <Tab><FaHistory /> Memories</Tab>
          <Tab><FaComments /> Story Threads</Tab>
          {isDM && <Tab><FaCogs /> DM Tools</Tab>}
        </TabList>

        <TabPanel>
          <NarrativeProfile 
            profile={narrativeData.profile}
            characterId={selectedCharacter?.id}
            onProfileUpdate={(updatedProfile) => {
              setNarrativeData(prev => ({ ...prev, profile: updatedProfile }));
            }}
          />
        </TabPanel>

        <TabPanel>
          <BackstoryManager
            characterId={selectedCharacter?.id}
            backstoryElements={narrativeData.backstory}
            onBackstoryUpdate={(updatedBackstory) => {
              setNarrativeData(prev => ({ ...prev, backstory: updatedBackstory }));
            }}
          />
        </TabPanel>

        <TabPanel>
          <ConsequenceTracker
            consequences={narrativeData.consequences}
            sessionId={sessionId}
            isDM={isDM}
            onRecordAction={handleRecordAction}
            onConsequenceUpdate={() => fetchSessionNarrativeData()}
          />
        </TabPanel>

        <TabPanel>
          <PerspectiveViewer
            sessionId={sessionId}
            characterId={selectedCharacter?.id}
            isDM={isDM}
            onCreateEvent={handleCreateWorldEvent}
          />
        </TabPanel>

        <TabPanel>
          <CharacterMemories
            memories={narrativeData.memories}
            characterId={selectedCharacter?.id}
            sessionId={sessionId}
            onMemoryCreate={(memory) => {
              setNarrativeData(prev => ({
                ...prev,
                memories: [...prev.memories, memory]
              }));
            }}
          />
        </TabPanel>

        <TabPanel>
          <StoryThreads
            threads={narrativeData.threads}
            sessionId={sessionId}
            isDM={isDM}
            onThreadUpdate={() => fetchSessionNarrativeData()}
          />
        </TabPanel>

        {isDM && (
          <TabPanel>
            <div className="dm-narrative-tools">
              <h3>Dungeon Master Narrative Tools</h3>
              
              <div className="dm-tools-grid">
                <div className="dm-tool-card">
                  <h4><FaEye /> World Event Creator</h4>
                  <p>Create significant events that affect the entire game world</p>
                  <button 
                    className="btn-primary"
                    onClick={() => {
                      // Open world event modal
                      const modal = document.getElementById('world-event-modal');
                      if (modal) modal.style.display = 'block';
                    }}
                  >
                    Create World Event
                  </button>
                </div>

                <div className="dm-tool-card">
                  <h4><FaChessBoard /> Consequence Manager</h4>
                  <p>View and trigger consequences from player actions</p>
                  <button 
                    className="btn-primary"
                    onClick={() => setActiveTab(2)}
                  >
                    Manage Consequences
                  </button>
                </div>

                <div className="dm-tool-card">
                  <h4><FaComments /> Story Thread Weaver</h4>
                  <p>Create and manage overarching narrative threads</p>
                  <button 
                    className="btn-primary"
                    onClick={() => setActiveTab(5)}
                  >
                    Manage Threads
                  </button>
                </div>

                <div className="dm-tool-card">
                  <h4><FaTheaterMasks /> Perspective Generator</h4>
                  <p>Generate multiple viewpoints for events</p>
                  <button 
                    className="btn-primary"
                    onClick={() => setActiveTab(3)}
                  >
                    Generate Perspectives
                  </button>
                </div>
              </div>

              {/* Quick Action Recorder */}
              <div className="quick-action-recorder">
                <h4>Quick Action Recorder</h4>
                <form onSubmit={(e) => {
                  e.preventDefault();
                  const formData = new FormData(e.target);
                  handleRecordAction({
                    action_type: formData.get('action_type'),
                    target_type: formData.get('target_type'),
                    action_description: formData.get('description'),
                    moral_weight: formData.get('moral_weight'),
                    immediate_result: formData.get('result')
                  });
                  e.target.reset();
                }}>
                  <div className="form-row">
                    <input
                      name="action_type"
                      placeholder="Action type (e.g., 'kill', 'save', 'steal')"
                      required
                    />
                    <input
                      name="target_type"
                      placeholder="Target type (e.g., 'npc', 'item', 'location')"
                      required
                    />
                  </div>
                  <textarea
                    name="description"
                    placeholder="Describe the action in detail..."
                    required
                  />
                  <div className="form-row">
                    <select name="moral_weight" required>
                      <option value="">Moral Weight</option>
                      <option value="good">Good</option>
                      <option value="evil">Evil</option>
                      <option value="neutral">Neutral</option>
                      <option value="chaotic">Chaotic</option>
                      <option value="lawful">Lawful</option>
                    </select>
                    <input
                      name="result"
                      placeholder="Immediate result"
                      required
                    />
                  </div>
                  <button type="submit" className="btn-primary">
                    Record Action
                  </button>
                </form>
              </div>
            </div>
          </TabPanel>
        )}
      </Tabs>

      {/* World Event Modal */}
      <div id="world-event-modal" className="modal" style={{ display: 'none' }}>
        <div className="modal-content">
          <span 
            className="close" 
            onClick={() => {
              const modal = document.getElementById('world-event-modal');
              if (modal) modal.style.display = 'none';
            }}
          >
            &times;
          </span>
          <h3>Create World Event</h3>
          <form onSubmit={async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            try {
              await handleCreateWorldEvent({
                type: formData.get('type'),
                name: formData.get('name'),
                description: formData.get('description'),
                location: formData.get('location'),
                participants: formData.get('participants').split(',').map(p => p.trim()),
                immediate_effects: formData.get('effects').split(',').map(e => e.trim()),
                generate_perspectives: formData.get('generate_perspectives') === 'on'
              });
              e.target.reset();
              const modal = document.getElementById('world-event-modal');
              if (modal) modal.style.display = 'none';
              fetchSessionNarrativeData();
            } catch (err) {
              alert('Failed to create world event');
            }
          }}>
            <input
              name="type"
              placeholder="Event type (e.g., 'battle', 'discovery', 'betrayal')"
              required
            />
            <input
              name="name"
              placeholder="Event name"
              required
            />
            <textarea
              name="description"
              placeholder="Describe the event..."
              rows="4"
              required
            />
            <input
              name="location"
              placeholder="Location"
              required
            />
            <input
              name="participants"
              placeholder="Participants (comma-separated)"
              required
            />
            <input
              name="effects"
              placeholder="Immediate effects (comma-separated)"
              required
            />
            <label>
              <input
                type="checkbox"
                name="generate_perspectives"
                defaultChecked
              />
              Generate multiple perspectives
            </label>
            <div className="modal-actions">
              <button type="submit" className="btn-primary">Create Event</button>
              <button 
                type="button" 
                className="btn-secondary"
                onClick={() => {
                  const modal = document.getElementById('world-event-modal');
                  if (modal) modal.style.display = 'none';
                }}
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default NarrativeEngine;