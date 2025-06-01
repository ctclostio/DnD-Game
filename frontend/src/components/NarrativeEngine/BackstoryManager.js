import React, { useState } from 'react';
import { FaBook, FaPlus, FaEdit, FaTrash, FaCheck, FaTimes, FaStar, FaHistory } from 'react-icons/fa';
import api from '../../services/api';

const BackstoryManager = ({ characterId, backstoryElements, onBackstoryUpdate }) => {
  const [isAdding, setIsAdding] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [newElement, setNewElement] = useState({
    type: 'origin',
    content: '',
    tags: [],
    weight: 1.0
  });
  const [tagInput, setTagInput] = useState('');

  const elementTypes = {
    origin: { icon: '🏛️', label: 'Origin', color: '#3498db' },
    trauma: { icon: '💔', label: 'Trauma', color: '#e74c3c' },
    goal: { icon: '🎯', label: 'Goal', color: '#2ecc71' },
    relationship: { icon: '🤝', label: 'Relationship', color: '#f39c12' },
    secret: { icon: '🤫', label: 'Secret', color: '#9b59b6' }
  };

  const handleAddElement = async () => {
    try {
      const response = await api.post('/narrative/backstory', {
        ...newElement,
        character_id: characterId
      });
      
      onBackstoryUpdate([...backstoryElements, response.data]);
      setIsAdding(false);
      setNewElement({
        type: 'origin',
        content: '',
        tags: [],
        weight: 1.0
      });
      setTagInput('');
    } catch (error) {
      console.error('Failed to add backstory element:', error);
    }
  };

  const handleDeleteElement = async (elementId) => {
    if (!window.confirm('Are you sure you want to delete this backstory element?')) {
      return;
    }

    try {
      await api.delete(`/narrative/backstory/${elementId}`);
      onBackstoryUpdate(backstoryElements.filter(el => el.id !== elementId));
    } catch (error) {
      console.error('Failed to delete backstory element:', error);
    }
  };

  const handleAddTag = (e) => {
    if (e.key === 'Enter' && tagInput.trim()) {
      e.preventDefault();
      if (!newElement.tags.includes(tagInput.trim())) {
        setNewElement({
          ...newElement,
          tags: [...newElement.tags, tagInput.trim()]
        });
      }
      setTagInput('');
    }
  };

  const removeTag = (tagToRemove) => {
    setNewElement({
      ...newElement,
      tags: newElement.tags.filter(tag => tag !== tagToRemove)
    });
  };

  const getUsageIndicator = (element) => {
    if (!element.used) return { text: 'Unused', class: 'unused' };
    if (element.usage_count === 1) return { text: 'Used Once', class: 'used-once' };
    if (element.usage_count > 3) return { text: 'Frequently Used', class: 'used-frequently' };
    return { text: `Used ${element.usage_count} times`, class: 'used-multiple' };
  };

  return (
    <div className="backstory-manager">
      <div className="backstory-header">
        <h3><FaBook /> Character Backstory</h3>
        {!isAdding && (
          <button className="btn-primary" onClick={() => setIsAdding(true)}>
            <FaPlus /> Add Element
          </button>
        )}
      </div>

      {isAdding && (
        <div className="backstory-form">
          <h4>Add Backstory Element</h4>
          
          <div className="form-group">
            <label>Type:</label>
            <div className="type-selector">
              {Object.entries(elementTypes).map(([type, info]) => (
                <button
                  key={type}
                  className={`type-button ${newElement.type === type ? 'active' : ''}`}
                  onClick={() => setNewElement({ ...newElement, type })}
                  style={{ borderColor: newElement.type === type ? info.color : '#ddd' }}
                >
                  <span className="type-icon">{info.icon}</span>
                  <span>{info.label}</span>
                </button>
              ))}
            </div>
          </div>

          <div className="form-group">
            <label>Content:</label>
            <textarea
              value={newElement.content}
              onChange={(e) => setNewElement({ ...newElement, content: e.target.value })}
              placeholder="Describe this element of your backstory..."
              rows="4"
            />
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
                {newElement.tags.map(tag => (
                  <span key={tag} className="tag">
                    {tag}
                    <button onClick={() => removeTag(tag)}>&times;</button>
                  </span>
                ))}
              </div>
            </div>
          </div>

          <div className="form-group">
            <label>
              Importance (Weight):
              <input
                type="range"
                min="0.1"
                max="2.0"
                step="0.1"
                value={newElement.weight}
                onChange={(e) => setNewElement({ ...newElement, weight: parseFloat(e.target.value) })}
              />
              <span className="weight-value">{newElement.weight.toFixed(1)}</span>
            </label>
          </div>

          <div className="form-actions">
            <button className="btn-primary" onClick={handleAddElement}>
              <FaCheck /> Add Element
            </button>
            <button className="btn-secondary" onClick={() => {
              setIsAdding(false);
              setNewElement({ type: 'origin', content: '', tags: [], weight: 1.0 });
              setTagInput('');
            }}>
              <FaTimes /> Cancel
            </button>
          </div>
        </div>
      )}

      <div className="backstory-elements">
        {backstoryElements.length === 0 ? (
          <div className="empty-state">
            <FaHistory />
            <h4>No Backstory Elements</h4>
            <p>Add backstory elements to enrich your character's narrative</p>
          </div>
        ) : (
          backstoryElements.map(element => {
            const typeInfo = elementTypes[element.type];
            const usage = getUsageIndicator(element);
            
            return (
              <div 
                key={element.id} 
                className={`backstory-element ${element.used ? 'used' : ''}`}
                style={{ borderLeftColor: typeInfo.color }}
              >
                <div className="element-header">
                  <div className="element-type" style={{ color: typeInfo.color }}>
                    <span className="type-icon">{typeInfo.icon}</span>
                    <span>{typeInfo.label}</span>
                  </div>
                  <div className="element-meta">
                    <span className={`usage-indicator ${usage.class}`}>
                      {usage.text}
                    </span>
                    <div className="element-weight">
                      {[...Array(Math.ceil(element.weight))].map((_, i) => (
                        <FaStar 
                          key={i} 
                          className={i < element.weight ? 'filled' : 'empty'}
                        />
                      ))}
                    </div>
                  </div>
                </div>

                <div className="element-content">
                  <p>{element.content}</p>
                </div>

                {element.tags && element.tags.length > 0 && (
                  <div className="element-tags">
                    {element.tags.map(tag => (
                      <span key={tag} className="tag">{tag}</span>
                    ))}
                  </div>
                )}

                <div className="element-actions">
                  {editingId !== element.id && (
                    <>
                      <button 
                        className="btn-icon" 
                        onClick={() => setEditingId(element.id)}
                        title="Edit"
                      >
                        <FaEdit />
                      </button>
                      <button 
                        className="btn-icon danger" 
                        onClick={() => handleDeleteElement(element.id)}
                        title="Delete"
                      >
                        <FaTrash />
                      </button>
                    </>
                  )}
                </div>

                {element.used && element.usage_count > 0 && (
                  <div className="usage-history">
                    <small>
                      This element has been woven into {element.usage_count} narrative{element.usage_count > 1 ? 's' : ''}
                    </small>
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      <div className="backstory-tips">
        <h4>Tips for Creating Compelling Backstory</h4>
        <ul>
          <li><strong>Origins:</strong> Where did your character come from? What shaped their early life?</li>
          <li><strong>Traumas:</strong> What hardships have they faced? What scars do they carry?</li>
          <li><strong>Goals:</strong> What drives them forward? What do they hope to achieve?</li>
          <li><strong>Relationships:</strong> Who matters to them? Who have they lost or left behind?</li>
          <li><strong>Secrets:</strong> What are they hiding? What would they never want revealed?</li>
        </ul>
        <p>
          <em>The AI will weave these elements into your story, creating personalized narratives 
          that resonate with your character's history.</em>
        </p>
      </div>
    </div>
  );
};

export default BackstoryManager;