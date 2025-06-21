import React, { useState } from 'react';
import { FaPlus, FaTrash, FaEdit, FaCog } from 'react-icons/fa';

const PropertyPanel = ({
  selectedNode,
  onPropertyChange,
  parameters,
  onParameterAdd,
  onParameterUpdate,
  onParameterDelete
}) => {
  const [showAddParameter, setShowAddParameter] = useState(false);
  const [newParameter, setNewParameter] = useState({
    name: '',
    display_name: '',
    type: 'number',
    default_value: null,
    constraints: {
      required: false
    },
    description: ''
  });

  const handlePropertyChange = (key, value) => {
    if (selectedNode) {
      onPropertyChange(selectedNode.id, {
        ...selectedNode.properties,
        [key]: value
      });
    }
  };
  
  // Helper functions to reduce nesting
  const updateArrayItem = (array, index, newValue) => {
    const newArray = [...(array || [])];
    newArray[index] = newValue;
    return newArray;
  };
  
  const removeArrayItem = (array, index) => {
    return (array || []).filter((_, i) => i !== index);
  };
  
  const addArrayItem = (array) => {
    return [...(array || []), ''];
  };

  const renderPropertyInput = (key, value, propertyDef = {}) => {
    const type = propertyDef.type || typeof value;

    switch (type) {
      case 'boolean':
        return (
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={value || false}
              onChange={(e) => handlePropertyChange(key, e.target.checked)}
            />
            <span>{propertyDef.label || key}</span>
          </label>
        );

      case 'number':
        return (
          <input
            type="number"
            value={value || 0}
            onChange={(e) => handlePropertyChange(key, parseFloat(e.target.value))}
            min={propertyDef.min}
            max={propertyDef.max}
            step={propertyDef.step || 1}
          />
        );

      case 'select':
        return (
          <select
            value={value || ''}
            onChange={(e) => handlePropertyChange(key, e.target.value)}
          >
            <option value="">Select...</option>
            {propertyDef.options?.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        );

      case 'array':
        return (
          <div className="array-input">
            {(value || []).map((item, index) => (
              <div key={index} className="array-item">
                <input
                  type="text"
                  value={item}
                  onChange={(e) => 
                    handlePropertyChange(key, updateArrayItem(value, index, e.target.value))
                  }
                />
                <button
                  onClick={() => 
                    handlePropertyChange(key, removeArrayItem(value, index))
                  }
                >
                  <FaTrash />
                </button>
              </div>
            ))}
            <button
              className="add-array-item"
              onClick={() => handlePropertyChange(key, addArrayItem(value))}
            >
              <FaPlus /> Add Item
            </button>
          </div>
        );

      default:
        return (
          <input
            type="text"
            value={value || ''}
            onChange={(e) => handlePropertyChange(key, e.target.value)}
          />
        );
    }
  };

  const getPropertyDefinitions = () => {
    // Define property schemas for different node types
    const schemas = {
      trigger_action: {
        action_types: {
          type: 'array',
          label: 'Action Types',
          description: 'Which actions trigger this rule'
        }
      },
      trigger_damage: {
        damage_types: {
          type: 'array',
          label: 'Damage Types',
          description: 'Which damage types trigger this rule'
        },
        threshold: {
          type: 'number',
          label: 'Minimum Damage',
          min: 0,
          description: 'Minimum damage to trigger'
        }
      },
      trigger_time: {
        whose_turn: {
          type: 'select',
          label: 'Whose Turn',
          options: [
            { value: 'self', label: 'Self' },
            { value: 'any', label: 'Any' },
            { value: 'enemy', label: 'Enemy' },
            { value: 'ally', label: 'Ally' }
          ]
        },
        phase: {
          type: 'select',
          label: 'Turn Phase',
          options: [
            { value: 'start', label: 'Start of Turn' },
            { value: 'end', label: 'End of Turn' }
          ]
        }
      },
      condition_compare: {
        operator: {
          type: 'select',
          label: 'Comparison',
          options: [
            { value: '>', label: 'Greater Than' },
            { value: '>=', label: 'Greater or Equal' },
            { value: '<', label: 'Less Than' },
            { value: '<=', label: 'Less or Equal' },
            { value: '==', label: 'Equals' },
            { value: '!=', label: 'Not Equals' }
          ]
        }
      },
      condition_roll: {
        ability: {
          type: 'select',
          label: 'Ability',
          options: [
            { value: 'strength', label: 'Strength' },
            { value: 'dexterity', label: 'Dexterity' },
            { value: 'constitution', label: 'Constitution' },
            { value: 'intelligence', label: 'Intelligence' },
            { value: 'wisdom', label: 'Wisdom' },
            { value: 'charisma', label: 'Charisma' }
          ]
        },
        dc: {
          type: 'number',
          label: 'Difficulty Class',
          min: 1,
          max: 30
        }
      },
      action_damage: {
        damage_dice: {
          type: 'string',
          label: 'Damage Dice',
          placeholder: 'e.g., 2d6+3'
        },
        damage_type: {
          type: 'select',
          label: 'Damage Type',
          options: [
            { value: 'slashing', label: 'Slashing' },
            { value: 'piercing', label: 'Piercing' },
            { value: 'bludgeoning', label: 'Bludgeoning' },
            { value: 'fire', label: 'Fire' },
            { value: 'cold', label: 'Cold' },
            { value: 'lightning', label: 'Lightning' },
            { value: 'thunder', label: 'Thunder' },
            { value: 'acid', label: 'Acid' },
            { value: 'poison', label: 'Poison' },
            { value: 'necrotic', label: 'Necrotic' },
            { value: 'radiant', label: 'Radiant' },
            { value: 'psychic', label: 'Psychic' },
            { value: 'force', label: 'Force' }
          ]
        }
      },
      action_effect: {
        effect_type: {
          type: 'select',
          label: 'Effect Type',
          options: [
            { value: 'condition', label: 'Apply Condition' },
            { value: 'buff', label: 'Buff' },
            { value: 'debuff', label: 'Debuff' },
            { value: 'movement', label: 'Movement' },
            { value: 'invisibility', label: 'Invisibility' },
            { value: 'advantage', label: 'Grant Advantage' },
            { value: 'disadvantage', label: 'Impose Disadvantage' }
          ]
        },
        duration: {
          type: 'select',
          label: 'Duration',
          options: [
            { value: 'instant', label: 'Instant' },
            { value: '1_turn', label: '1 Turn' },
            { value: '1_minute', label: '1 Minute' },
            { value: '10_minutes', label: '10 Minutes' },
            { value: '1_hour', label: '1 Hour' },
            { value: 'concentration', label: 'Concentration' }
          ]
        }
      },
      calc_math: {
        operation: {
          type: 'select',
          label: 'Operation',
          options: [
            { value: '+', label: 'Add' },
            { value: '-', label: 'Subtract' },
            { value: '*', label: 'Multiply' },
            { value: '/', label: 'Divide' },
            { value: '^', label: 'Power' },
            { value: 'min', label: 'Minimum' },
            { value: 'max', label: 'Maximum' }
          ]
        }
      },
      calc_random: {
        dice_notation: {
          type: 'string',
          label: 'Dice',
          placeholder: 'e.g., 1d20'
        }
      }
    };

    return schemas[selectedNode?.type] || {};
  };

  const handleAddParameter = () => {
    if (newParameter.name) {
      onParameterAdd(newParameter);
      setNewParameter({
        name: '',
        display_name: '',
        type: 'number',
        default_value: null,
        constraints: {
          required: false
        },
        description: ''
      });
      setShowAddParameter(false);
    }
  };

  return (
    <div className="property-panel">
      {selectedNode ? (
        <>
          <h3><FaCog /> Node Properties</h3>
          
          <div className="property-group">
            <h4>Node Info</h4>
            <div className="property-field">
              <label>ID</label>
              <input type="text" value={selectedNode.id} disabled />
            </div>
            <div className="property-field">
              <label>Type</label>
              <input type="text" value={selectedNode.type} disabled />
            </div>
          </div>

          <div className="property-group">
            <h4>Configuration</h4>
            {Object.entries(selectedNode.properties || {}).map(([key, value]) => {
              const propertyDef = getPropertyDefinitions()[key] || {};
              
              return (
                <div key={key} className="property-field">
                  <label>
                    {propertyDef.label || key}
                    {propertyDef.description && (
                      <span className="property-help" title={propertyDef.description}>?</span>
                    )}
                  </label>
                  {renderPropertyInput(key, value, propertyDef)}
                </div>
              );
            })}
          </div>
        </>
      ) : (
        <div className="no-selection">
          <p>Select a node to edit its properties</p>
        </div>
      )}

      {/* Rule Parameters Section */}
      <div className="parameters-section">
        <h3>Rule Parameters</h3>
        <p className="section-description">
          Define customizable parameters that users can adjust when using this rule
        </p>

        {parameters.map((param, index) => (
          <div key={index} className="parameter-item">
            <div className="parameter-header">
              <span className="parameter-name">{param.display_name || param.name}</span>
              <div className="parameter-actions">
                <button onClick={() => {
                  // Open edit modal
                  console.log('Edit parameter', param);
                }}>
                  <FaEdit />
                </button>
                <button onClick={() => onParameterDelete(index)}>
                  <FaTrash />
                </button>
              </div>
            </div>
            <div className="parameter-details">
              <span className="parameter-type">{param.type}</span>
              {param.default_value && (
                <span className="parameter-default">Default: {param.default_value}</span>
              )}
            </div>
            {param.description && (
              <p className="parameter-description">{param.description}</p>
            )}
          </div>
        ))}

        {showAddParameter ? (
          <div className="add-parameter-form">
            <div className="form-field">
              <label>Internal Name</label>
              <input
                type="text"
                value={newParameter.name}
                onChange={(e) => setNewParameter({ ...newParameter, name: e.target.value })}
                placeholder="e.g., damage_bonus"
              />
            </div>
            <div className="form-field">
              <label>Display Name</label>
              <input
                type="text"
                value={newParameter.display_name}
                onChange={(e) => setNewParameter({ ...newParameter, display_name: e.target.value })}
                placeholder="e.g., Damage Bonus"
              />
            </div>
            <div className="form-field">
              <label>Type</label>
              <select
                value={newParameter.type}
                onChange={(e) => setNewParameter({ ...newParameter, type: e.target.value })}
              >
                <option value="number">Number</option>
                <option value="string">Text</option>
                <option value="boolean">Yes/No</option>
                <option value="choice">Choice</option>
                <option value="entity_reference">Entity Reference</option>
              </select>
            </div>
            <div className="form-field">
              <label>Description</label>
              <textarea
                value={newParameter.description}
                onChange={(e) => setNewParameter({ ...newParameter, description: e.target.value })}
                placeholder="Describe what this parameter does..."
              />
            </div>
            <div className="form-actions">
              <button className="btn-primary" onClick={handleAddParameter}>
                Add Parameter
              </button>
              <button className="btn-secondary" onClick={() => setShowAddParameter(false)}>
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <button className="add-parameter-btn" onClick={() => setShowAddParameter(true)}>
            <FaPlus /> Add Parameter
          </button>
        )}
      </div>
    </div>
  );
};

export default PropertyPanel;