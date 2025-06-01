import React, { useState, useEffect } from 'react';
import { FaBalanceScale, FaChartLine, FaExclamationTriangle, FaCheckCircle, FaRedo, FaCog } from 'react-icons/fa';
import api from '../../services/api';

const BalanceAnalysis = ({ ruleTemplate, onAnalysisComplete }) => {
  const [analysis, setAnalysis] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [selectedScenario, setSelectedScenario] = useState('all');
  const [showSettings, setShowSettings] = useState(false);
  const [analysisSettings, setAnalysisSettings] = useState({
    simulation_count: 1000,
    level_range: { min: 1, max: 20 },
    scenarios: ['pvp', 'pve', 'exploration', 'roleplay']
  });

  useEffect(() => {
    if (ruleTemplate?.id && ruleTemplate?.logic_graph?.nodes?.length > 0) {
      runAnalysis();
    }
  }, [ruleTemplate?.id]);

  const runAnalysis = async () => {
    if (!ruleTemplate?.id) return;

    setLoading(true);
    setError(null);

    try {
      const response = await api.post(`/api/rules/templates/${ruleTemplate.id}/analyze`, analysisSettings);
      setAnalysis(response.data);
      if (onAnalysisComplete) {
        onAnalysisComplete(response.data);
      }
    } catch (err) {
      setError('Failed to analyze rule balance');
      console.error('Balance analysis error:', err);
    } finally {
      setLoading(false);
    }
  };

  const getBalanceColor = (score) => {
    if (score >= 80) return '#27ae60';
    if (score >= 60) return '#f39c12';
    if (score >= 40) return '#e67e22';
    return '#e74c3c';
  };

  const getBalanceLabel = (score) => {
    if (score >= 80) return 'Well Balanced';
    if (score >= 60) return 'Mostly Balanced';
    if (score >= 40) return 'Needs Adjustment';
    return 'Unbalanced';
  };

  const renderScenarioAnalysis = (scenario) => {
    const scenarioData = analysis.scenario_results[scenario];
    if (!scenarioData) return null;

    return (
      <div className="scenario-analysis">
        <h4>{scenario.toUpperCase()}</h4>
        
        <div className="metrics-grid">
          <div className="metric">
            <label>Win Rate</label>
            <div className="metric-value">{Math.round(scenarioData.win_rate * 100)}%</div>
            <div className="metric-bar">
              <div 
                className="metric-fill"
                style={{
                  width: `${scenarioData.win_rate * 100}%`,
                  backgroundColor: getBalanceColor(scenarioData.win_rate * 100)
                }}
              />
            </div>
          </div>

          <div className="metric">
            <label>Avg. Damage/Round</label>
            <div className="metric-value">{scenarioData.average_damage.toFixed(1)}</div>
          </div>

          <div className="metric">
            <label>Action Economy</label>
            <div className="metric-value">{scenarioData.action_economy.toFixed(2)}</div>
          </div>

          <div className="metric">
            <label>Resource Cost</label>
            <div className="metric-value">{scenarioData.resource_efficiency.toFixed(1)}</div>
          </div>
        </div>

        {scenarioData.issues && scenarioData.issues.length > 0 && (
          <div className="scenario-issues">
            <h5>Potential Issues</h5>
            {scenarioData.issues.map((issue, index) => (
              <div key={index} className="issue-item">
                <FaExclamationTriangle style={{ color: '#e74c3c' }} />
                <span>{issue}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    );
  };

  const renderBalanceChart = () => {
    if (!analysis) return null;

    const categories = ['Power', 'Complexity', 'Versatility', 'Fun Factor', 'Resource Efficiency'];
    const values = [
      analysis.power_level * 20,
      (1 - analysis.complexity_rating) * 100,
      analysis.versatility_score * 100,
      analysis.fun_factor * 100,
      analysis.resource_efficiency * 100
    ];

    return (
      <div className="balance-radar-chart">
        <svg viewBox="0 0 400 400" className="radar-svg">
          {/* Draw grid */}
          {[20, 40, 60, 80, 100].map(radius => (
            <polygon
              key={radius}
              points={categories.map((_, i) => {
                const angle = (i / categories.length) * 2 * Math.PI - Math.PI / 2;
                const x = 200 + (radius * 1.5) * Math.cos(angle);
                const y = 200 + (radius * 1.5) * Math.sin(angle);
                return `${x},${y}`;
              }).join(' ')}
              fill="none"
              stroke="#333"
              strokeOpacity="0.2"
            />
          ))}

          {/* Draw axes */}
          {categories.map((_, i) => {
            const angle = (i / categories.length) * 2 * Math.PI - Math.PI / 2;
            const x = 200 + 150 * Math.cos(angle);
            const y = 200 + 150 * Math.sin(angle);
            return (
              <line
                key={i}
                x1="200"
                y1="200"
                x2={x}
                y2={y}
                stroke="#333"
                strokeOpacity="0.2"
              />
            );
          })}

          {/* Draw data */}
          <polygon
            points={values.map((value, i) => {
              const angle = (i / categories.length) * 2 * Math.PI - Math.PI / 2;
              const x = 200 + (value * 1.5) * Math.cos(angle);
              const y = 200 + (value * 1.5) * Math.sin(angle);
              return `${x},${y}`;
            }).join(' ')}
            fill="#3498db"
            fillOpacity="0.3"
            stroke="#3498db"
            strokeWidth="2"
          />

          {/* Labels */}
          {categories.map((category, i) => {
            const angle = (i / categories.length) * 2 * Math.PI - Math.PI / 2;
            const x = 200 + 170 * Math.cos(angle);
            const y = 200 + 170 * Math.sin(angle);
            return (
              <text
                key={i}
                x={x}
                y={y}
                textAnchor="middle"
                dominantBaseline="middle"
                className="radar-label"
              >
                {category}
              </text>
            );
          })}
        </svg>
      </div>
    );
  };

  return (
    <div className="balance-analysis">
      <div className="analysis-header">
        <h3><FaBalanceScale /> Balance Analysis</h3>
        <div className="analysis-actions">
          <button
            className="btn-icon"
            onClick={() => setShowSettings(!showSettings)}
            title="Analysis Settings"
          >
            <FaCog />
          </button>
          <button
            className="btn-primary"
            onClick={runAnalysis}
            disabled={loading || !ruleTemplate?.id}
          >
            <FaRedo /> Re-analyze
          </button>
        </div>
      </div>

      {showSettings && (
        <div className="analysis-settings">
          <h4>Analysis Settings</h4>
          <div className="setting-group">
            <label>Simulation Count</label>
            <input
              type="number"
              value={analysisSettings.simulation_count}
              onChange={(e) => setAnalysisSettings({
                ...analysisSettings,
                simulation_count: parseInt(e.target.value)
              })}
              min="100"
              max="10000"
              step="100"
            />
          </div>
          <div className="setting-group">
            <label>Level Range</label>
            <div className="range-inputs">
              <input
                type="number"
                value={analysisSettings.level_range.min}
                onChange={(e) => setAnalysisSettings({
                  ...analysisSettings,
                  level_range: { ...analysisSettings.level_range, min: parseInt(e.target.value) }
                })}
                min="1"
                max="20"
              />
              <span>to</span>
              <input
                type="number"
                value={analysisSettings.level_range.max}
                onChange={(e) => setAnalysisSettings({
                  ...analysisSettings,
                  level_range: { ...analysisSettings.level_range, max: parseInt(e.target.value) }
                })}
                min="1"
                max="20"
              />
            </div>
          </div>
          <div className="setting-group">
            <label>Scenarios</label>
            <div className="scenario-toggles">
              {['pvp', 'pve', 'exploration', 'roleplay'].map(scenario => (
                <label key={scenario} className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={analysisSettings.scenarios.includes(scenario)}
                    onChange={(e) => {
                      if (e.target.checked) {
                        setAnalysisSettings({
                          ...analysisSettings,
                          scenarios: [...analysisSettings.scenarios, scenario]
                        });
                      } else {
                        setAnalysisSettings({
                          ...analysisSettings,
                          scenarios: analysisSettings.scenarios.filter(s => s !== scenario)
                        });
                      }
                    }}
                  />
                  <span>{scenario.toUpperCase()}</span>
                </label>
              ))}
            </div>
          </div>
        </div>
      )}

      {loading && (
        <div className="analysis-loading">
          <div className="spinner" />
          <p>Running {analysisSettings.simulation_count} simulations across {analysisSettings.scenarios.length} scenarios...</p>
        </div>
      )}

      {error && (
        <div className="analysis-error">
          <FaExclamationTriangle />
          <p>{error}</p>
        </div>
      )}

      {analysis && !loading && (
        <>
          {/* Overall Balance Score */}
          <div className="overall-balance">
            <div 
              className="balance-score-circle"
              style={{ borderColor: getBalanceColor(analysis.balance_score) }}
            >
              <div className="score-value">{Math.round(analysis.balance_score)}</div>
              <div className="score-label">{getBalanceLabel(analysis.balance_score)}</div>
            </div>
            
            <div className="balance-summary">
              <p>{analysis.overall_assessment}</p>
              
              {analysis.balance_suggestions && analysis.balance_suggestions.length > 0 && (
                <div className="suggestions">
                  <h4>Suggestions</h4>
                  {analysis.balance_suggestions.map((suggestion, index) => (
                    <div key={index} className="suggestion-item">
                      <FaCheckCircle style={{ color: '#27ae60' }} />
                      <span>{suggestion}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Radar Chart */}
          <div className="analysis-chart">
            <h4>Balance Metrics</h4>
            {renderBalanceChart()}
          </div>

          {/* Scenario Tabs */}
          <div className="scenario-analysis-section">
            <div className="scenario-tabs">
              <button
                className={selectedScenario === 'all' ? 'active' : ''}
                onClick={() => setSelectedScenario('all')}
              >
                All Scenarios
              </button>
              {Object.keys(analysis.scenario_results).map(scenario => (
                <button
                  key={scenario}
                  className={selectedScenario === scenario ? 'active' : ''}
                  onClick={() => setSelectedScenario(scenario)}
                >
                  {scenario.toUpperCase()}
                </button>
              ))}
            </div>

            <div className="scenario-content">
              {selectedScenario === 'all' ? (
                <div className="all-scenarios">
                  {Object.keys(analysis.scenario_results).map(scenario => (
                    <div key={scenario}>
                      {renderScenarioAnalysis(scenario)}
                    </div>
                  ))}
                </div>
              ) : (
                renderScenarioAnalysis(selectedScenario)
              )}
            </div>
          </div>

          {/* Comparison with Similar Rules */}
          {analysis.comparison_data && (
            <div className="rule-comparison">
              <h4>Comparison with Similar Rules</h4>
              <div className="comparison-chart">
                <FaChartLine />
                <p>Your rule is {analysis.comparison_data.percentile}th percentile in power level</p>
                <div className="comparison-bar">
                  <div 
                    className="percentile-marker"
                    style={{ left: `${analysis.comparison_data.percentile}%` }}
                  />
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default BalanceAnalysis;