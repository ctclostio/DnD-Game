import React, { useState, useEffect } from 'react';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { Tab, Tabs, TabList, TabPanel } from 'react-tabs';
import 'react-tabs/style/react-tabs.css';
import {
  FaCogs, FaPlay, FaBalanceScale, FaMagic, FaSave,
  FaShare, FaBook, FaExclamationTriangle, FaCheckCircle
} from 'react-icons/fa';
import api from '../services/api';
import VisualLogicEditor from './RuleBuilder/VisualLogicEditor';
import NodePalette from './RuleBuilder/NodePalette';
import PropertyPanel from './RuleBuilder/PropertyPanel';
import BalanceAnalysis from './RuleBuilder/BalanceAnalysis';
import ConditionalRules from './RuleBuilder/ConditionalRules';
import RuleLibrary from './RuleBuilder/RuleLibrary';
import RuleValidator from './RuleBuilder/RuleValidator';
import '../styles/rule-builder.css';

const RuleBuilder = () => {
  const [activeTab, setActiveTab] = useState(0);
  const [loading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  
  // Rule template state
  const [ruleTemplate, setRuleTemplate] = useState({
    name: '',
    description: '',
    category: 'ability',
    logic_graph: {
      nodes: [],
      connections: [],
      start_node_id: '',
      variables: {}
    },
    parameters: [],
    conditional_rules: [],
    tags: [],
    is_public: false
  });

  const [selectedNode, setSelectedNode] = useState(null);
  const [nodeTemplates, setNodeTemplates] = useState([]);
  const [validationResult, setValidationResult] = useState(null);
  const [balanceMetrics, setBalanceMetrics] = useState(null);
  const [isAnalyzing, setIsAnalyzing] = useState(false);

  // Fetch node templates on mount
  useEffect(() => {
    fetchNodeTemplates();
  }, []);

  const fetchNodeTemplates = async () => {
    try {
      const response = await api.get('/rules/nodes/templates');
      setNodeTemplates(response.data);
    } catch (err) {
      console.error('Failed to fetch node templates:', err);
    }
  };

  const handleNodeAdd = (nodeTemplate) => {
    const newNode = {
      id: `node_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      type: nodeTemplate.node_type,
      subtype: nodeTemplate.subtype,
      position: { x: 100, y: 100 }, // Default position
      properties: { ...nodeTemplate.default_properties },
      inputs: [...nodeTemplate.input_ports],
      outputs: [...nodeTemplate.output_ports]
    };

    setRuleTemplate(prev => ({
      ...prev,
      logic_graph: {
        ...prev.logic_graph,
        nodes: [...prev.logic_graph.nodes, newNode]
      }
    }));
  };

  const handleNodeUpdate = (nodeId, updates) => {
    setRuleTemplate(prev => ({
      ...prev,
      logic_graph: {
        ...prev.logic_graph,
        nodes: prev.logic_graph.nodes.map(node =>
          node.id === nodeId ? { ...node, ...updates } : node
        )
      }
    }));
  };

  const handleNodeDelete = (nodeId) => {
    setRuleTemplate(prev => ({
      ...prev,
      logic_graph: {
        ...prev.logic_graph,
        nodes: prev.logic_graph.nodes.filter(node => node.id !== nodeId),
        connections: prev.logic_graph.connections.filter(
          conn => conn.from_node_id !== nodeId && conn.to_node_id !== nodeId
        )
      }
    }));
    
    if (selectedNode?.id === nodeId) {
      setSelectedNode(null);
    }
  };

  const handleConnectionAdd = (connection) => {
    setRuleTemplate(prev => ({
      ...prev,
      logic_graph: {
        ...prev.logic_graph,
        connections: [...prev.logic_graph.connections, {
          id: `conn_${Date.now()}`,
          ...connection
        }]
      }
    }));
  };

  const handleConnectionDelete = (connectionId) => {
    setRuleTemplate(prev => ({
      ...prev,
      logic_graph: {
        ...prev.logic_graph,
        connections: prev.logic_graph.connections.filter(conn => conn.id !== connectionId)
      }
    }));
  };

  const validateRule = async () => {
    try {
      const response = await api.post('/rules/templates/validate', ruleTemplate);
      setValidationResult(response.data);
      return response.data.valid;
    } catch (err) {
      setError('Validation failed');
      return false;
    }
  };

  const analyzeBalance = async () => {
    if (!ruleTemplate.id) {
      setError('Save the rule before analyzing balance');
      return;
    }

    setIsAnalyzing(true);
    try {
      const response = await api.post(`/rules/templates/${ruleTemplate.id}/analyze`);
      setBalanceMetrics(response.data);
      setActiveTab(2); // Switch to balance tab
    } catch (err) {
      setError('Balance analysis failed');
    } finally {
      setIsAnalyzing(false);
    }
  };

  const saveRule = async () => {
    const isValid = await validateRule();
    if (!isValid) {
      setError('Rule validation failed. Check the validator tab for details.');
      return;
    }

    setSaving(true);
    try {
      let response;
      if (ruleTemplate.id) {
        response = await api.put(`/rules/templates/${ruleTemplate.id}`, ruleTemplate);
      } else {
        response = await api.post('/rules/templates?analyze=true', ruleTemplate);
      }
      
      setRuleTemplate(response.data);
      setBalanceMetrics(response.data.balance_metrics);
      setError(null);
      
      // Show success message
      alert('Rule saved successfully!');
    } catch (err) {
      setError('Failed to save rule');
    } finally {
      setSaving(false);
    }
  };

  const publishRule = async () => {
    if (!ruleTemplate.id) {
      setError('Save the rule before publishing');
      return;
    }

    if (ruleTemplate.approval_status !== 'approved') {
      setError('Rule must be approved before publishing');
      return;
    }

    try {
      await api.post(`/rules/templates/${ruleTemplate.id}/publish`);
      setRuleTemplate(prev => ({ ...prev, is_public: true }));
      alert('Rule published to library!');
    } catch (err) {
      setError('Failed to publish rule');
    }
  };

  const testRule = () => {
    // Open test modal or navigate to test page
    alert('Rule testing coming soon!');
  };

  return (
    <DndProvider backend={HTML5Backend}>
      <div className="rule-builder">
        <div className="rule-builder-header">
          <div className="rule-info">
            <input
              type="text"
              className="rule-name-input"
              placeholder="Rule Name"
              value={ruleTemplate.name}
              onChange={(e) => setRuleTemplate(prev => ({ ...prev, name: e.target.value }))}
            />
            <select
              className="rule-category-select"
              value={ruleTemplate.category}
              onChange={(e) => setRuleTemplate(prev => ({ ...prev, category: e.target.value }))}
            >
              <option value="spell">Spell</option>
              <option value="ability">Ability</option>
              <option value="item">Item</option>
              <option value="environmental">Environmental</option>
              <option value="condition">Condition</option>
            </select>
          </div>

          <div className="rule-actions">
            <button 
              className="btn-validate"
              onClick={validateRule}
              disabled={loading}
            >
              <FaCheckCircle /> Validate
            </button>
            <button 
              className="btn-test"
              onClick={testRule}
              disabled={!ruleTemplate.id || loading}
            >
              <FaPlay /> Test
            </button>
            <button 
              className="btn-analyze"
              onClick={analyzeBalance}
              disabled={!ruleTemplate.id || isAnalyzing}
            >
              <FaBalanceScale /> {isAnalyzing ? 'Analyzing...' : 'Analyze Balance'}
            </button>
            <button 
              className="btn-save"
              onClick={saveRule}
              disabled={saving}
            >
              <FaSave /> {saving ? 'Saving...' : 'Save'}
            </button>
            <button 
              className="btn-publish"
              onClick={publishRule}
              disabled={!ruleTemplate.id || ruleTemplate.approval_status !== 'approved'}
            >
              <FaShare /> Publish
            </button>
          </div>
        </div>

        {error && (
          <div className="error-banner">
            <FaExclamationTriangle /> {error}
            <button onClick={() => setError(null)}>&times;</button>
          </div>
        )}

        <div className="rule-builder-content">
          <Tabs selectedIndex={activeTab} onSelect={setActiveTab}>
            <TabList>
              <Tab><FaCogs /> Visual Logic</Tab>
              <Tab><FaMagic /> Conditional Rules</Tab>
              <Tab><FaBalanceScale /> Balance Analysis</Tab>
              <Tab><FaCheckCircle /> Validator</Tab>
              <Tab><FaBook /> Library</Tab>
            </TabList>

            <TabPanel>
              <div className="visual-logic-workspace">
                <div className="node-palette-container">
                  <NodePalette 
                    nodeTemplates={nodeTemplates}
                    onNodeAdd={handleNodeAdd}
                  />
                </div>
                
                <div className="editor-container">
                  <VisualLogicEditor
                    logicGraph={ruleTemplate.logic_graph}
                    selectedNode={selectedNode}
                    onNodeSelect={setSelectedNode}
                    onNodeUpdate={handleNodeUpdate}
                    onNodeDelete={handleNodeDelete}
                    onConnectionAdd={handleConnectionAdd}
                    onConnectionDelete={handleConnectionDelete}
                    onStartNodeSet={(nodeId) => {
                      setRuleTemplate(prev => ({
                        ...prev,
                        logic_graph: {
                          ...prev.logic_graph,
                          start_node_id: nodeId
                        }
                      }));
                    }}
                  />
                </div>

                <div className="property-panel-container">
                  <PropertyPanel
                    selectedNode={selectedNode}
                    onPropertyChange={(nodeId, properties) => {
                      handleNodeUpdate(nodeId, { properties });
                    }}
                    parameters={ruleTemplate.parameters}
                    onParameterAdd={(param) => {
                      setRuleTemplate(prev => ({
                        ...prev,
                        parameters: [...prev.parameters, param]
                      }));
                    }}
                    onParameterUpdate={(index, param) => {
                      setRuleTemplate(prev => ({
                        ...prev,
                        parameters: prev.parameters.map((p, i) => 
                          i === index ? param : p
                        )
                      }));
                    }}
                    onParameterDelete={(index) => {
                      setRuleTemplate(prev => ({
                        ...prev,
                        parameters: prev.parameters.filter((_, i) => i !== index)
                      }));
                    }}
                  />
                </div>
              </div>
            </TabPanel>

            <TabPanel>
              <ConditionalRules
                conditionalRules={ruleTemplate.conditional_rules}
                onRuleAdd={(rule) => {
                  setRuleTemplate(prev => ({
                    ...prev,
                    conditional_rules: [...prev.conditional_rules, rule]
                  }));
                }}
                onRuleUpdate={(index, rule) => {
                  setRuleTemplate(prev => ({
                    ...prev,
                    conditional_rules: prev.conditional_rules.map((r, i) =>
                      i === index ? rule : r
                    )
                  }));
                }}
                onRuleDelete={(index) => {
                  setRuleTemplate(prev => ({
                    ...prev,
                    conditional_rules: prev.conditional_rules.filter((_, i) => i !== index)
                  }));
                }}
              />
            </TabPanel>

            <TabPanel>
              <BalanceAnalysis
                balanceMetrics={balanceMetrics}
                ruleTemplate={ruleTemplate}
                onReanalyze={analyzeBalance}
                isAnalyzing={isAnalyzing}
              />
            </TabPanel>

            <TabPanel>
              <RuleValidator
                ruleTemplate={ruleTemplate}
                validationResult={validationResult}
                onValidate={validateRule}
              />
            </TabPanel>

            <TabPanel>
              <RuleLibrary
                onImportRule={(importedRule) => {
                  setRuleTemplate({
                    ...importedRule,
                    id: undefined, // Remove ID to create new rule
                    is_public: false,
                    approval_status: 'pending'
                  });
                  setActiveTab(0); // Go back to editor
                }}
              />
            </TabPanel>
          </Tabs>
        </div>

        <div className="rule-description">
          <textarea
            placeholder="Describe your rule..."
            value={ruleTemplate.description}
            onChange={(e) => setRuleTemplate(prev => ({ ...prev, description: e.target.value }))}
            rows="3"
          />
        </div>
      </div>
    </DndProvider>
  );
};

export default RuleBuilder;