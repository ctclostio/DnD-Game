import React, { useState, useEffect } from 'react';
import { FaGlobeAmericas, FaCloud, FaHeart, FaPlus, FaTrash, FaEdit, FaMagic } from 'react-icons/fa';
import api from '../../services/api';

const ConditionalRules = ({ ruleTemplate, onConditionsUpdate }) => {
  const [conditions, setConditions] = useState([]);
  const [showAddCondition, setShowAddCondition] = useState(false);
  const [editingCondition, setEditingCondition] = useState(null);
  const [newCondition, setNewCondition] = useState({
    name: '',
    context_type: 'plane',
    context_value: '',
    modifiers: {},
    priority: 1,
    description: ''
  });

  useEffect(() => {
    if (ruleTemplate?.conditional_modifiers) {
      setConditions(ruleTemplate.conditional_modifiers);
    }
  }, [ruleTemplate]);

  const contextTypes = {
    plane: {
      icon: <FaGlobeAmericas />,
      label: 'Plane of Existence',
      options: [
        'Material Plane',
        'Feywild',
        'Shadowfell',
        'Elemental Plane of Fire',
        'Elemental Plane of Water',
        'Elemental Plane of Air',
        'Elemental Plane of Earth',
        'Astral Plane',
        'Ethereal Plane',
        'Nine Hells',
        'Abyss',
        'Mechanus',
        'Limbo'
      ]
    },
    weather: {
      icon: <FaCloud />,
      label: 'Weather Condition',
      options: [
        'Clear',
        'Rain',
        'Storm',
        'Snow',
        'Fog',
        'Extreme Heat',
        'Extreme Cold',
        'Magical Storm',
        'Ash Storm',
        'Sandstorm'
      ]
    },
    emotion: {
      icon: <FaHeart />,
      label: 'Emotional State',
      options: [
        'Calm',
        'Angry',
        'Frightened',
        'Joyful',
        'Sad',
        'Confused',
        'Inspired',
        'Desperate',
        'Determined',
        'Berserk'
      ]
    },
    terrain: {
      icon: '⛰️',
      label: 'Terrain Type',
      options: [
        'Plains',
        'Forest',
        'Mountain',
        'Swamp',
        'Desert',
        'Arctic',
        'Coast',
        'Underground',
        'Underwater',
        'Urban'
      ]
    },
    time: {
      icon: '🕐',
      label: 'Time of Day',
      options: [
        'Dawn',
        'Morning',
        'Noon',
        'Afternoon',
        'Dusk',
        'Night',
        'Midnight',
        'Witching Hour'
      ]
    },
    moon_phase: {
      icon: '🌙',
      label: 'Moon Phase',
      options: [
        'New Moon',
        'Waxing Crescent',
        'First Quarter',
        'Waxing Gibbous',
        'Full Moon',
        'Waning Gibbous',
        'Last Quarter',
        'Waning Crescent'
      ]
    }
  };

  const modifierTypes = [
    { key: 'damage_multiplier', label: 'Damage Multiplier', type: 'number', min: 0.1, max: 5 },
    { key: 'accuracy_bonus', label: 'Accuracy Bonus', type: 'number', min: -10, max: 10 },
    { key: 'cost_modifier', label: 'Resource Cost', type: 'number', min: 0.1, max: 3 },
    { key: 'duration_modifier', label: 'Duration Modifier', type: 'number', min: 0.1, max: 5 },
    { key: 'range_modifier', label: 'Range Modifier', type: 'number', min: 0.5, max: 3 },
    { key: 'cooldown_modifier', label: 'Cooldown Modifier', type: 'number', min: 0.1, max: 3 },
    { key: 'area_modifier', label: 'Area of Effect', type: 'number', min: 0.5, max: 3 },
    { key: 'save_dc_modifier', label: 'Save DC Modifier', type: 'number', min: -5, max: 5 },
    { key: 'additional_effect', label: 'Additional Effect', type: 'text' },
    { key: 'restriction', label: 'Restriction', type: 'text' }
  ];

  const handleAddCondition = () => {
    if (newCondition.name && newCondition.context_value) {
      const condition = {
        ...newCondition,
        id: Date.now().toString()
      };
      
      const updatedConditions = [...conditions, condition];
      setConditions(updatedConditions);
      onConditionsUpdate(updatedConditions);
      
      // Reset form
      setNewCondition({
        name: '',
        context_type: 'plane',
        context_value: '',
        modifiers: {},
        priority: 1,
        description: ''
      });
      setShowAddCondition(false);
    }
  };

  const handleUpdateCondition = () => {
    if (editingCondition) {
      const updatedConditions = conditions.map(c => 
        c.id === editingCondition.id ? editingCondition : c
      );
      setConditions(updatedConditions);
      onConditionsUpdate(updatedConditions);
      setEditingCondition(null);
    }
  };

  const handleDeleteCondition = (conditionId) => {
    const updatedConditions = conditions.filter(c => c.id !== conditionId);
    setConditions(updatedConditions);
    onConditionsUpdate(updatedConditions);
  };

  const renderModifierInput = (modifierType, value, onChange) => {
    if (modifierType.type === 'number') {
      return (
        <input
          type="number"
          value={value || ''}
          onChange={(e) => onChange(parseFloat(e.target.value) || 0)}
          min={modifierType.min}
          max={modifierType.max}
          step="0.1"
        />
      );
    } else {
      return (
        <input
          type="text"
          value={value || ''}
          onChange={(e) => onChange(e.target.value)}
          placeholder={`e.g., Gains fire damage`}
        />
      );
    }
  };

  const getModifierDescription = (key, value) => {
    switch (key) {
      case 'damage_multiplier':
        return `${value > 1 ? '+' : ''}${Math.round((value - 1) * 100)}% damage`;
      case 'accuracy_bonus':
        return `${value > 0 ? '+' : ''}${value} to hit`;
      case 'cost_modifier':
        return `${Math.round(value * 100)}% resource cost`;
      case 'duration_modifier':
        return `${Math.round(value * 100)}% duration`;
      case 'range_modifier':
        return `${Math.round(value * 100)}% range`;
      case 'cooldown_modifier':
        return `${Math.round(value * 100)}% cooldown`;
      case 'area_modifier':
        return `${Math.round(value * 100)}% area`;
      case 'save_dc_modifier':
        return `${value > 0 ? '+' : ''}${value} to save DC`;
      default:
        return value;
    }
  };

  const renderConditionForm = (condition, isEditing = false) => {
    const currentCondition = isEditing ? editingCondition : newCondition;
    const setCurrentCondition = isEditing ? setEditingCondition : setNewCondition;

    return (
      <div className="condition-form">
        <div className="form-row">
          <div className="form-field">
            <label>Condition Name</label>
            <input
              type="text"
              value={currentCondition.name}
              onChange={(e) => setCurrentCondition({ ...currentCondition, name: e.target.value })}
              placeholder="e.g., Feywild Enhancement"
            />
          </div>
          <div className="form-field">
            <label>Priority</label>
            <input
              type="number"
              value={currentCondition.priority}
              onChange={(e) => setCurrentCondition({ ...currentCondition, priority: parseInt(e.target.value) })}
              min="1"
              max="10"
            />
          </div>
        </div>

        <div className="form-row">
          <div className="form-field">
            <label>Context Type</label>
            <select
              value={currentCondition.context_type}
              onChange={(e) => setCurrentCondition({ 
                ...currentCondition, 
                context_type: e.target.value,
                context_value: '' 
              })}
            >
              {Object.entries(contextTypes).map(([key, type]) => (
                <option key={key} value={key}>{type.label}</option>
              ))}
            </select>
          </div>
          <div className="form-field">
            <label>Context Value</label>
            <select
              value={currentCondition.context_value}
              onChange={(e) => setCurrentCondition({ ...currentCondition, context_value: e.target.value })}
            >
              <option value="">Select {contextTypes[currentCondition.context_type].label}</option>
              {contextTypes[currentCondition.context_type].options.map(option => (
                <option key={option} value={option}>{option}</option>
              ))}
            </select>
          </div>
        </div>

        <div className="form-field">
          <label>Description</label>
          <textarea
            value={currentCondition.description}
            onChange={(e) => setCurrentCondition({ ...currentCondition, description: e.target.value })}
            placeholder="Describe how this condition affects the rule..."
          />
        </div>

        <div className="modifiers-section">
          <h5>Modifiers</h5>
          {modifierTypes.map(modType => (
            <div key={modType.key} className="modifier-row">
              <label>{modType.label}</label>
              {renderModifierInput(
                modType,
                currentCondition.modifiers[modType.key],
                (value) => setCurrentCondition({
                  ...currentCondition,
                  modifiers: {
                    ...currentCondition.modifiers,
                    [modType.key]: value
                  }
                })
              )}
            </div>
          ))}
        </div>

        <div className="form-actions">
          <button 
            className="btn-primary" 
            onClick={isEditing ? handleUpdateCondition : handleAddCondition}
          >
            {isEditing ? 'Update Condition' : 'Add Condition'}
          </button>
          <button 
            className="btn-secondary" 
            onClick={() => isEditing ? setEditingCondition(null) : setShowAddCondition(false)}
          >
            Cancel
          </button>
        </div>
      </div>
    );
  };

  return (
    <div className="conditional-rules">
      <div className="conditional-header">
        <h3><FaMagic /> Conditional Reality System</h3>
        <p className="section-description">
          Define how your rule changes based on environmental, emotional, or planar contexts
        </p>
      </div>

      {/* Active Conditions */}
      <div className="conditions-list">
        {conditions.map(condition => (
          <div key={condition.id} className="condition-card">
            <div className="condition-header">
              <div className="condition-title">
                <span className="condition-icon">
                  {contextTypes[condition.context_type].icon}
                </span>
                <h4>{condition.name}</h4>
                <span className="condition-priority">Priority {condition.priority}</span>
              </div>
              <div className="condition-actions">
                <button onClick={() => setEditingCondition(condition)}>
                  <FaEdit />
                </button>
                <button onClick={() => handleDeleteCondition(condition.id)}>
                  <FaTrash />
                </button>
              </div>
            </div>

            <div className="condition-context">
              <span className="context-label">Active in:</span>
              <span className="context-value">{condition.context_value}</span>
            </div>

            {condition.description && (
              <p className="condition-description">{condition.description}</p>
            )}

            <div className="condition-modifiers">
              {Object.entries(condition.modifiers).filter(([_, value]) => value).map(([key, value]) => (
                <div key={key} className="modifier-tag">
                  {getModifierDescription(key, value)}
                </div>
              ))}
            </div>
          </div>
        ))}

        {conditions.length === 0 && (
          <div className="empty-conditions">
            <p>No conditional modifiers defined yet</p>
            <p className="hint">Add conditions to make your rule adapt to different contexts</p>
          </div>
        )}
      </div>

      {/* Add/Edit Form */}
      {(showAddCondition || editingCondition) && (
        <div className="condition-form-container">
          <h4>{editingCondition ? 'Edit Condition' : 'New Condition'}</h4>
          {renderConditionForm(null, !!editingCondition)}
        </div>
      )}

      {/* Add Button */}
      {!showAddCondition && !editingCondition && (
        <button 
          className="add-condition-btn"
          onClick={() => setShowAddCondition(true)}
        >
          <FaPlus /> Add Conditional Modifier
        </button>
      )}

      {/* Context Preview */}
      {conditions.length > 0 && (
        <div className="context-preview">
          <h4>Context Preview</h4>
          <p>See how your rule behaves in different situations:</p>
          
          <div className="preview-grid">
            {Object.entries(contextTypes).map(([type, config]) => {
              const relevantConditions = conditions.filter(c => c.context_type === type);
              if (relevantConditions.length === 0) return null;

              return (
                <div key={type} className="preview-category">
                  <h5>{config.icon} {config.label}</h5>
                  {relevantConditions.map(condition => (
                    <div key={condition.id} className="preview-item">
                      <span className="preview-context">{condition.context_value}:</span>
                      <span className="preview-effect">{condition.name}</span>
                    </div>
                  ))}
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};

export default ConditionalRules;