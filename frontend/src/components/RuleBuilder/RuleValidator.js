import React, { useState, useEffect } from 'react';
import { FaCheckCircle, FaExclamationTriangle, FaTimesCircle, FaInfoCircle, FaPlay, FaCog } from 'react-icons/fa';
import api from '../../services/api';

const RuleValidator = ({ logicGraph, ruleTemplate, onValidationComplete }) => {
  const [validationResults, setValidationResults] = useState(null);
  const [isValidating, setIsValidating] = useState(false);
  const [autoValidate, setAutoValidate] = useState(true);
  const [testScenario, setTestScenario] = useState({
    character_level: 5,
    target_ac: 15,
    current_hp: 30,
    max_hp: 40,
    ability_scores: {
      strength: 16,
      dexterity: 14,
      constitution: 15,
      intelligence: 12,
      wisdom: 13,
      charisma: 10
    }
  });
  const [showTestConfig, setShowTestConfig] = useState(false);

  useEffect(() => {
    if (autoValidate && logicGraph?.nodes?.length > 0) {
      const debounceTimer = setTimeout(() => {
        validateRule();
      }, 500);
      
      return () => clearTimeout(debounceTimer);
    }
  }, [logicGraph, autoValidate]);

  const validateRule = async () => {
    setIsValidating(true);
    
    try {
      // Local validation first
      const localValidation = performLocalValidation();
      
      // If local validation passes and we have a template ID, do server validation
      if (localValidation.is_valid && ruleTemplate?.id) {
        const response = await api.post(`/api/rules/templates/${ruleTemplate.id}/validate`, {
          test_scenario: testScenario
        });
        
        setValidationResults({
          ...response.data,
          local_validation: localValidation
        });
      } else {
        setValidationResults({
          local_validation: localValidation,
          server_validation: null
        });
      }
      
      if (onValidationComplete) {
        onValidationComplete(validationResults);
      }
    } catch (error) {
      console.error('Validation error:', error);
      setValidationResults({
        local_validation: performLocalValidation(),
        server_validation: {
          is_valid: false,
          errors: ['Failed to validate rule on server']
        }
      });
    } finally {
      setIsValidating(false);
    }
  };

  const performLocalValidation = () => {
    const errors = [];
    const warnings = [];
    const info = [];

    // Check if graph exists
    if (!logicGraph || !logicGraph.nodes) {
      errors.push({
        type: 'error',
        message: 'No logic graph defined',
        node_id: null
      });
      return { is_valid: false, errors, warnings, info };
    }

    // Check for start node
    if (!logicGraph.start_node_id) {
      errors.push({
        type: 'error',
        message: 'No start node defined. Set a trigger node as the starting point.',
        node_id: null
      });
    }

    // Check for at least one trigger
    const triggers = logicGraph.nodes.filter(n => n.type.startsWith('trigger_'));
    if (triggers.length === 0) {
      errors.push({
        type: 'error',
        message: 'Rule must have at least one trigger node',
        node_id: null
      });
    }

    // Check for at least one action
    const actions = logicGraph.nodes.filter(n => n.type.startsWith('action_'));
    if (actions.length === 0) {
      warnings.push({
        type: 'warning',
        message: 'Rule has no action nodes. It will not produce any effects.',
        node_id: null
      });
    }

    // Check for disconnected nodes
    const connectedNodes = new Set();
    logicGraph.connections.forEach(conn => {
      connectedNodes.add(conn.from_node_id);
      connectedNodes.add(conn.to_node_id);
    });

    logicGraph.nodes.forEach(node => {
      if (!connectedNodes.has(node.id) && node.id !== logicGraph.start_node_id) {
        warnings.push({
          type: 'warning',
          message: `Node "${node.subtype || node.type}" is disconnected`,
          node_id: node.id
        });
      }
    });

    // Check for cycles
    const hasCycle = detectCycles(logicGraph);
    if (hasCycle) {
      errors.push({
        type: 'error',
        message: 'Logic graph contains cycles. This will cause infinite loops.',
        node_id: null
      });
    }

    // Validate node properties
    logicGraph.nodes.forEach(node => {
      const validation = validateNodeProperties(node);
      errors.push(...validation.errors);
      warnings.push(...validation.warnings);
    });

    // Check complexity
    const complexity = calculateComplexity(logicGraph);
    if (complexity > 50) {
      warnings.push({
        type: 'warning',
        message: `Rule is very complex (score: ${complexity}). Consider simplifying.`,
        node_id: null
      });
    }

    // Provide helpful info
    info.push({
      type: 'info',
      message: `Rule contains ${logicGraph.nodes.length} nodes and ${logicGraph.connections.length} connections`
    });

    return {
      is_valid: errors.length === 0,
      errors,
      warnings,
      info,
      complexity_score: complexity
    };
  };

  const detectCycles = (graph) => {
    const visited = new Set();
    const recursionStack = new Set();

    const hasCycleDFS = (nodeId) => {
      visited.add(nodeId);
      recursionStack.add(nodeId);

      const outgoingConnections = graph.connections.filter(c => c.from_node_id === nodeId);
      
      for (const conn of outgoingConnections) {
        if (!visited.has(conn.to_node_id)) {
          if (hasCycleDFS(conn.to_node_id)) {
            return true;
          }
        } else if (recursionStack.has(conn.to_node_id)) {
          return true;
        }
      }

      recursionStack.delete(nodeId);
      return false;
    };

    for (const node of graph.nodes) {
      if (!visited.has(node.id)) {
        if (hasCycleDFS(node.id)) {
          return true;
        }
      }
    }

    return false;
  };

  const validateNodeProperties = (node) => {
    const errors = [];
    const warnings = [];

    // Type-specific validation
    switch (node.type) {
      case 'condition_compare':
        if (!node.properties?.operator) {
          errors.push({
            type: 'error',
            message: 'Comparison node missing operator',
            node_id: node.id
          });
        }
        break;

      case 'action_damage':
        if (!node.properties?.damage_dice) {
          errors.push({
            type: 'error',
            message: 'Damage action missing damage dice',
            node_id: node.id
          });
        } else if (!isValidDiceNotation(node.properties.damage_dice)) {
          errors.push({
            type: 'error',
            message: `Invalid dice notation: ${node.properties.damage_dice}`,
            node_id: node.id
          });
        }
        break;

      case 'condition_roll':
        if (!node.properties?.dc || node.properties.dc < 1 || node.properties.dc > 30) {
          warnings.push({
            type: 'warning',
            message: 'Roll condition has unusual DC value',
            node_id: node.id
          });
        }
        break;

      case 'calc_random':
        if (!node.properties?.dice_notation) {
          errors.push({
            type: 'error',
            message: 'Random calculation missing dice notation',
            node_id: node.id
          });
        }
        break;
    }

    return { errors, warnings };
  };

  const isValidDiceNotation = (notation) => {
    const diceRegex = /^\d+d\d+([+-]\d+)?$/;
    return diceRegex.test(notation);
  };

  const calculateComplexity = (graph) => {
    let complexity = 0;
    
    // Base complexity from node count
    complexity += graph.nodes.length * 2;
    
    // Additional complexity for connections
    complexity += graph.connections.length;
    
    // Extra complexity for certain node types
    graph.nodes.forEach(node => {
      if (node.type.startsWith('flow_')) complexity += 3;
      if (node.type.includes('roll')) complexity += 2;
      if (node.type.includes('calc')) complexity += 2;
    });
    
    return complexity;
  };

  const getSeverityIcon = (severity) => {
    switch (severity) {
      case 'error':
        return <FaTimesCircle style={{ color: '#e74c3c' }} />;
      case 'warning':
        return <FaExclamationTriangle style={{ color: '#f39c12' }} />;
      case 'info':
        return <FaInfoCircle style={{ color: '#3498db' }} />;
      case 'success':
        return <FaCheckCircle style={{ color: '#27ae60' }} />;
      default:
        return null;
    }
  };

  const getValidationSummary = () => {
    if (!validationResults) return null;

    const localVal = validationResults.local_validation;
    const hasErrors = localVal.errors.length > 0;
    const hasWarnings = localVal.warnings.length > 0;

    if (hasErrors) {
      return {
        icon: <FaTimesCircle style={{ color: '#e74c3c' }} />,
        text: `${localVal.errors.length} error${localVal.errors.length > 1 ? 's' : ''} found`,
        color: '#e74c3c'
      };
    } else if (hasWarnings) {
      return {
        icon: <FaExclamationTriangle style={{ color: '#f39c12' }} />,
        text: `${localVal.warnings.length} warning${localVal.warnings.length > 1 ? 's' : ''}`,
        color: '#f39c12'
      };
    } else {
      return {
        icon: <FaCheckCircle style={{ color: '#27ae60' }} />,
        text: 'Rule is valid',
        color: '#27ae60'
      };
    }
  };

  const summary = getValidationSummary();

  return (
    <div className="rule-validator">
      <div className="validator-header">
        <h3>Validation Results</h3>
        <div className="validator-controls">
          <label className="auto-validate-toggle">
            <input
              type="checkbox"
              checked={autoValidate}
              onChange={(e) => setAutoValidate(e.target.checked)}
            />
            <span>Auto-validate</span>
          </label>
          <button 
            className="btn-icon"
            onClick={() => setShowTestConfig(!showTestConfig)}
            title="Test Configuration"
          >
            <FaCog />
          </button>
          <button
            className="btn-primary"
            onClick={validateRule}
            disabled={isValidating}
          >
            <FaPlay /> Validate
          </button>
        </div>
      </div>

      {/* Test Configuration */}
      {showTestConfig && (
        <div className="test-config">
          <h4>Test Scenario Configuration</h4>
          <div className="config-grid">
            <div className="config-field">
              <label>Character Level</label>
              <input
                type="number"
                value={testScenario.character_level}
                onChange={(e) => setTestScenario({
                  ...testScenario,
                  character_level: parseInt(e.target.value)
                })}
                min="1"
                max="20"
              />
            </div>
            <div className="config-field">
              <label>Target AC</label>
              <input
                type="number"
                value={testScenario.target_ac}
                onChange={(e) => setTestScenario({
                  ...testScenario,
                  target_ac: parseInt(e.target.value)
                })}
                min="10"
                max="30"
              />
            </div>
            <div className="config-field">
              <label>Current HP</label>
              <input
                type="number"
                value={testScenario.current_hp}
                onChange={(e) => setTestScenario({
                  ...testScenario,
                  current_hp: parseInt(e.target.value)
                })}
                min="1"
                max="999"
              />
            </div>
            <div className="config-field">
              <label>Max HP</label>
              <input
                type="number"
                value={testScenario.max_hp}
                onChange={(e) => setTestScenario({
                  ...testScenario,
                  max_hp: parseInt(e.target.value)
                })}
                min="1"
                max="999"
              />
            </div>
          </div>
        </div>
      )}

      {/* Validation Summary */}
      {summary && (
        <div 
          className="validation-summary"
          style={{ borderColor: summary.color }}
        >
          {summary.icon}
          <span>{summary.text}</span>
          {validationResults?.local_validation?.complexity_score && (
            <span className="complexity-badge">
              Complexity: {validationResults.local_validation.complexity_score}
            </span>
          )}
        </div>
      )}

      {/* Validation Results */}
      {validationResults && (
        <div className="validation-results">
          {/* Errors */}
          {validationResults.local_validation.errors.length > 0 && (
            <div className="validation-section errors">
              <h4>Errors</h4>
              {validationResults.local_validation.errors.map((error, index) => (
                <div key={index} className="validation-item">
                  {getSeverityIcon('error')}
                  <span>{error.message}</span>
                  {error.node_id && (
                    <button 
                      className="btn-link"
                      onClick={() => {
                        // Focus on the problematic node
                        const nodeElement = document.getElementById(`node-${error.node_id}`);
                        if (nodeElement) {
                          nodeElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
                          nodeElement.classList.add('highlight-error');
                          setTimeout(() => nodeElement.classList.remove('highlight-error'), 2000);
                        }
                      }}
                    >
                      Go to node
                    </button>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Warnings */}
          {validationResults.local_validation.warnings.length > 0 && (
            <div className="validation-section warnings">
              <h4>Warnings</h4>
              {validationResults.local_validation.warnings.map((warning, index) => (
                <div key={index} className="validation-item">
                  {getSeverityIcon('warning')}
                  <span>{warning.message}</span>
                  {warning.node_id && (
                    <button 
                      className="btn-link"
                      onClick={() => {
                        const nodeElement = document.getElementById(`node-${warning.node_id}`);
                        if (nodeElement) {
                          nodeElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
                          nodeElement.classList.add('highlight-warning');
                          setTimeout(() => nodeElement.classList.remove('highlight-warning'), 2000);
                        }
                      }}
                    >
                      Go to node
                    </button>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Info */}
          {validationResults.local_validation.info.length > 0 && (
            <div className="validation-section info">
              <h4>Information</h4>
              {validationResults.local_validation.info.map((info, index) => (
                <div key={index} className="validation-item">
                  {getSeverityIcon('info')}
                  <span>{info.message}</span>
                </div>
              ))}
            </div>
          )}

          {/* Server Validation Results */}
          {validationResults.server_validation && (
            <div className="validation-section server">
              <h4>Server Validation</h4>
              {validationResults.server_validation.is_valid ? (
                <div className="validation-item">
                  {getSeverityIcon('success')}
                  <span>Rule compiled successfully and is ready to use</span>
                </div>
              ) : (
                validationResults.server_validation.errors.map((error, index) => (
                  <div key={index} className="validation-item">
                    {getSeverityIcon('error')}
                    <span>{error}</span>
                  </div>
                ))
              )}
              
              {validationResults.server_validation.execution_result && (
                <div className="execution-result">
                  <h5>Test Execution Result</h5>
                  <pre>{JSON.stringify(validationResults.server_validation.execution_result, null, 2)}</pre>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Loading State */}
      {isValidating && (
        <div className="validation-loading">
          <div className="spinner" />
          <p>Validating rule...</p>
        </div>
      )}

      {/* Empty State */}
      {!isValidating && !validationResults && (
        <div className="validation-empty">
          <p>Click "Validate" to check your rule for errors</p>
        </div>
      )}
    </div>
  );
};

export default RuleValidator;