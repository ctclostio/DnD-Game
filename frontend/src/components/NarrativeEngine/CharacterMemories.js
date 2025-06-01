import React, { useState } from 'react';
import { 
  FaHistory, FaBrain, FaHeart, FaLink, FaPlus, 
  FaStar, FaTag, FaClock, FaFilter 
} from 'react-icons/fa';
import api from '../../services/api';

const CharacterMemories = ({ memories, characterId, sessionId, onMemoryCreate }) => {
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [filter, setFilter] = useState('all');
  const [sortBy, setSortBy] = useState('weight');
  const [selectedTags, setSelectedTags] = useState([]);
  const [newMemory, setNewMemory] = useState({
    memory_type: 'discovery',
    content: '',
    emotional_weight: 0.5,
    tags: []
  });
  const [tagInput, setTagInput] = useState('');

  const memoryTypes = {
    decision: { icon: '🤔', label: 'Decision', color: '#3498db' },
    consequence: { icon: '⚡', label: 'Consequence', color: '#e74c3c' },
    relationship: { icon: '🤝', label: 'Relationship', color: '#f39c12' },
    discovery: { icon: '🔍', label: 'Discovery', color: '#2ecc71' },
    loss: { icon: '💔', label: 'Loss', color: '#9b59b6' },
    achievement: { icon: '🏆', label: 'Achievement', color: '#f1c40f' },
    trauma: { icon: '😰', label: 'Trauma', color: '#c0392b' }
  };

  const allTags = [...new Set(memories.flatMap(m => m.tags || []))];

  const filteredMemories = memories
    .filter(memory => {
      if (filter === 'all') return true;
      if (filter === 'active') return memory.active;
      if (filter === 'high-impact') return memory.emotional_weight > 0.7;
      return memory.memory_type === filter;
    })
    .filter(memory => {
      if (selectedTags.length === 0) return true;
      return selectedTags.some(tag => memory.tags?.includes(tag));
    })
    .sort((a, b) => {
      if (sortBy === 'weight') return b.emotional_weight - a.emotional_weight;
      if (sortBy === 'recent') return new Date(b.created_at) - new Date(a.created_at);
      if (sortBy === 'referenced') return b.reference_count - a.reference_count;
      return 0;
    });

  const handleCreateMemory = async (e) => {
    e.preventDefault();
    
    try {
      const response = await api.post('/narrative/memory', {
        ...newMemory,
        character_id: characterId,
        session_id: sessionId
      });
      
      onMemoryCreate(response.data);
      setShowCreateForm(false);
      setNewMemory({
        memory_type: 'discovery',
        content: '',
        emotional_weight: 0.5,
        tags: []
      });
      setTagInput('');
    } catch (error) {
      console.error('Failed to create memory:', error);
    }
  };

  const handleAddTag = (e) => {
    if (e.key === 'Enter' && tagInput.trim()) {
      e.preventDefault();
      if (!newMemory.tags.includes(tagInput.trim())) {
        setNewMemory({
          ...newMemory,
          tags: [...newMemory.tags, tagInput.trim()]
        });
      }
      setTagInput('');
    }
  };

  const removeTag = (tagToRemove) => {
    setNewMemory({
      ...newMemory,
      tags: newMemory.tags.filter(tag => tag !== tagToRemove)
    });
  };

  const toggleTagFilter = (tag) => {
    setSelectedTags(prev =>
      prev.includes(tag)
        ? prev.filter(t => t !== tag)
        : [...prev, tag]
    );
  };

  const getEmotionalWeightDisplay = (weight) => {
    const stars = Math.ceil(weight * 5);
    return (
      <div className="emotional-weight">
        {[...Array(5)].map((_, i) => (
          <FaStar 
            key={i} 
            className={i < stars ? 'filled' : 'empty'}
            style={{ color: i < stars ? '#f1c40f' : '#ddd' }}
          />
        ))}
      </div>
    );
  };

  const getConnectionsDisplay = (connections) => {
    if (!connections || connections.length === 0) return null;
    
    return (
      <div className="memory-connections">
        <FaLink />
        <span>{connections.length} connected memories</span>
      </div>
    );
  };

  return (
    <div className="character-memories">
      <div className="memories-header">
        <h3><FaHistory /> Narrative Memories</h3>
        <button 
          className="btn-primary"
          onClick={() => setShowCreateForm(!showCreateForm)}
        >
          <FaPlus /> Create Memory
        </button>
      </div>

      {showCreateForm && (
        <form className="memory-form" onSubmit={handleCreateMemory}>
          <h4>Create New Memory</h4>
          
          <div className="form-group">
            <label>Type:</label>
            <div className="type-selector">
              {Object.entries(memoryTypes).map(([type, info]) => (
                <button
                  key={type}
                  type="button"
                  className={`type-button ${newMemory.memory_type === type ? 'active' : ''}`}
                  onClick={() => setNewMemory({ ...newMemory, memory_type: type })}
                  style={{ borderColor: newMemory.memory_type === type ? info.color : '#ddd' }}
                >
                  <span className="type-icon">{info.icon}</span>
                  <span>{info.label}</span>
                </button>
              ))}
            </div>
          </div>

          <div className="form-group">
            <label>Memory Content:</label>
            <textarea
              value={newMemory.content}
              onChange={(e) => setNewMemory({ ...newMemory, content: e.target.value })}
              placeholder="Describe this memory..."
              rows="3"
              required
            />
          </div>

          <div className="form-group">
            <label>
              Emotional Weight:
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                value={newMemory.emotional_weight}
                onChange={(e) => setNewMemory({ 
                  ...newMemory, 
                  emotional_weight: parseFloat(e.target.value) 
                })}
              />
              {getEmotionalWeightDisplay(newMemory.emotional_weight)}
            </label>
          </div>

          <div className="form-group">
            <label>Tags (press Enter to add):</label>
            <div className="tag-input-container">
              <input
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={handleAddTag}
                placeholder="Add tags..."
              />
              <div className="tag-list">
                {newMemory.tags.map(tag => (
                  <span key={tag} className="tag">
                    {tag}
                    <button type="button" onClick={() => removeTag(tag)}>&times;</button>
                  </span>
                ))}
              </div>
            </div>
          </div>

          <div className="form-actions">
            <button type="submit" className="btn-primary">Create Memory</button>
            <button 
              type="button" 
              className="btn-secondary"
              onClick={() => {
                setShowCreateForm(false);
                setNewMemory({
                  memory_type: 'discovery',
                  content: '',
                  emotional_weight: 0.5,
                  tags: []
                });
                setTagInput('');
              }}
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      <div className="memories-controls">
        <div className="filter-controls">
          <label><FaFilter /> Filter:</label>
          <select value={filter} onChange={(e) => setFilter(e.target.value)}>
            <option value="all">All Memories</option>
            <option value="active">Active Only</option>
            <option value="high-impact">High Impact</option>
            {Object.entries(memoryTypes).map(([type, info]) => (
              <option key={type} value={type}>{info.label}</option>
            ))}
          </select>
        </div>

        <div className="sort-controls">
          <label>Sort by:</label>
          <select value={sortBy} onChange={(e) => setSortBy(e.target.value)}>
            <option value="weight">Emotional Weight</option>
            <option value="recent">Most Recent</option>
            <option value="referenced">Most Referenced</option>
          </select>
        </div>
      </div>

      {allTags.length > 0 && (
        <div className="tag-filters">
          <label>Filter by tags:</label>
          <div className="tag-filter-list">
            {allTags.map(tag => (
              <button
                key={tag}
                className={`tag-filter ${selectedTags.includes(tag) ? 'active' : ''}`}
                onClick={() => toggleTagFilter(tag)}
              >
                <FaTag /> {tag}
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="memories-timeline">
        {filteredMemories.length === 0 ? (
          <div className="empty-state">
            <FaBrain />
            <h4>No Memories Found</h4>
            <p>Create memories to track important moments in your character's journey</p>
          </div>
        ) : (
          filteredMemories.map(memory => {
            const typeInfo = memoryTypes[memory.memory_type];
            
            return (
              <div 
                key={memory.id} 
                className={`memory-card ${memory.active ? 'active' : 'inactive'}`}
                style={{ borderLeftColor: typeInfo.color }}
              >
                <div className="memory-header">
                  <div className="memory-type" style={{ color: typeInfo.color }}>
                    <span className="type-icon">{typeInfo.icon}</span>
                    <span>{typeInfo.label}</span>
                  </div>
                  <div className="memory-meta">
                    {!memory.active && (
                      <span className="inactive-badge">Dormant</span>
                    )}
                    {getEmotionalWeightDisplay(memory.emotional_weight)}
                  </div>
                </div>

                <div className="memory-content">
                  <p>{memory.content}</p>
                </div>

                {memory.tags && memory.tags.length > 0 && (
                  <div className="memory-tags">
                    {memory.tags.map(tag => (
                      <span key={tag} className="tag">{tag}</span>
                    ))}
                  </div>
                )}

                <div className="memory-footer">
                  <div className="memory-stats">
                    {getConnectionsDisplay(memory.connections)}
                    {memory.reference_count > 0 && (
                      <div className="reference-count">
                        <FaHistory />
                        <span>Referenced {memory.reference_count} times</span>
                      </div>
                    )}
                  </div>
                  <div className="memory-timestamp">
                    <FaClock />
                    <small>{new Date(memory.created_at).toLocaleDateString()}</small>
                  </div>
                </div>

                {memory.last_referenced && (
                  <div className="last-referenced">
                    <small>
                      Last referenced: {new Date(memory.last_referenced).toLocaleDateString()}
                    </small>
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      <div className="memories-info">
        <FaBrain />
        <p>
          The AI uses these memories to create callbacks and emotional resonance in future narratives. 
          High-weight memories are more likely to influence story generation and character reactions.
        </p>
      </div>
    </div>
  );
};

export default CharacterMemories;