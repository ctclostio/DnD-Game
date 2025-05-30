import React, { useState, useEffect } from 'react';
import { getCombatAnalytics } from '../services/api';
import '../styles/combat-analytics.css';

const CombatAnalyticsView = ({ combatId }) => {
    const [analytics, setAnalytics] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        if (combatId) {
            loadAnalytics();
        }
    }, [combatId]);

    const loadAnalytics = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await getCombatAnalytics(combatId);
            setAnalytics(data);
        } catch (err) {
            setError('Failed to load combat analytics');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    if (loading) return <div className="loading">Loading analytics...</div>;
    if (error) return <div className="error-message">{error}</div>;
    if (!analytics) return null;

    const { analytics: combatStats, combatant_reports, tactical_analysis, recommendations } = analytics;

    return (
        <div className="combat-analytics-view">
            <h3>Combat Analysis Report</h3>
            
            {/* Combat Overview */}
            <div className="analytics-section">
                <h4>Combat Overview</h4>
                <div className="overview-grid">
                    <div className="stat-card">
                        <span className="stat-label">Duration</span>
                        <span className="stat-value">{combatStats.combat_duration} rounds</span>
                    </div>
                    <div className="stat-card">
                        <span className="stat-label">Total Damage</span>
                        <span className="stat-value">{combatStats.total_damage_dealt}</span>
                    </div>
                    <div className="stat-card">
                        <span className="stat-label">Total Healing</span>
                        <span className="stat-value">{combatStats.total_healing_done}</span>
                    </div>
                    <div className="stat-card">
                        <span className="stat-label">Tactical Rating</span>
                        <span className="stat-value">{combatStats.tactical_rating}/10</span>
                    </div>
                </div>
            </div>

            {/* MVP Section */}
            {combatStats.mvp_id && (
                <div className="analytics-section mvp-section">
                    <h4>🏆 Combat MVP</h4>
                    <div className="mvp-info">
                        {combatant_reports?.find(r => r.analytics.combatant_id === combatStats.mvp_id)?.analytics.combatant_name || 'Unknown'}
                        <span className="mvp-reason">Highest damage dealer</span>
                    </div>
                </div>
            )}

            {/* Combatant Performance */}
            <div className="analytics-section">
                <h4>Individual Performance</h4>
                <div className="combatants-grid">
                    {combatant_reports?.map(report => (
                        <div key={report.analytics.id} className={`combatant-card ${report.performance_rating}`}>
                            <div className="combatant-header">
                                <h5>{report.analytics.combatant_name}</h5>
                                <span className={`performance-badge ${report.performance_rating}`}>
                                    {report.performance_rating}
                                </span>
                            </div>
                            
                            <div className="combatant-stats">
                                <div className="stat-row">
                                    <span>Damage Dealt:</span>
                                    <span>{report.analytics.damage_dealt}</span>
                                </div>
                                <div className="stat-row">
                                    <span>Damage Taken:</span>
                                    <span>{report.analytics.damage_taken}</span>
                                </div>
                                {report.analytics.healing_done > 0 && (
                                    <div className="stat-row">
                                        <span>Healing Done:</span>
                                        <span>{report.analytics.healing_done}</span>
                                    </div>
                                )}
                                {report.analytics.attacks_made > 0 && (
                                    <div className="stat-row">
                                        <span>Hit Rate:</span>
                                        <span>
                                            {Math.round((report.analytics.attacks_hit / report.analytics.attacks_made) * 100)}%
                                        </span>
                                    </div>
                                )}
                                {report.analytics.critical_hits > 0 && (
                                    <div className="stat-row highlight">
                                        <span>Critical Hits:</span>
                                        <span>{report.analytics.critical_hits}</span>
                                    </div>
                                )}
                            </div>

                            {report.highlights && report.highlights.length > 0 && (
                                <div className="combatant-highlights">
                                    {report.highlights.map((highlight, idx) => (
                                        <div key={idx} className="highlight">
                                            ✨ {highlight}
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            </div>

            {/* Tactical Analysis */}
            {tactical_analysis && (
                <div className="analytics-section">
                    <h4>Tactical Analysis</h4>
                    <div className="tactical-grid">
                        <div className="tactical-score">
                            <span className="score-label">Positioning</span>
                            <div className="score-bar">
                                <div 
                                    className="score-fill" 
                                    style={{ width: `${tactical_analysis.positioning_score * 10}%` }}
                                />
                            </div>
                            <span className="score-value">{tactical_analysis.positioning_score}/10</span>
                        </div>
                        <div className="tactical-score">
                            <span className="score-label">Resource Management</span>
                            <div className="score-bar">
                                <div 
                                    className="score-fill" 
                                    style={{ width: `${tactical_analysis.resource_management * 10}%` }}
                                />
                            </div>
                            <span className="score-value">{tactical_analysis.resource_management}/10</span>
                        </div>
                        <div className="tactical-score">
                            <span className="score-label">Target Priority</span>
                            <div className="score-bar">
                                <div 
                                    className="score-fill" 
                                    style={{ width: `${tactical_analysis.target_prioritization * 10}%` }}
                                />
                            </div>
                            <span className="score-value">{tactical_analysis.target_prioritization}/10</span>
                        </div>
                        <div className="tactical-score">
                            <span className="score-label">Teamwork</span>
                            <div className="score-bar">
                                <div 
                                    className="score-fill" 
                                    style={{ width: `${tactical_analysis.teamwork_score * 10}%` }}
                                />
                            </div>
                            <span className="score-value">{tactical_analysis.teamwork_score}/10</span>
                        </div>
                    </div>

                    {tactical_analysis.missed_opportunities && tactical_analysis.missed_opportunities.length > 0 && (
                        <div className="missed-opportunities">
                            <h5>Missed Opportunities</h5>
                            <ul>
                                {tactical_analysis.missed_opportunities.map((opp, idx) => (
                                    <li key={idx}>{opp}</li>
                                ))}
                            </ul>
                        </div>
                    )}
                </div>
            )}

            {/* Recommendations */}
            {recommendations && recommendations.length > 0 && (
                <div className="analytics-section recommendations">
                    <h4>Recommendations for Next Combat</h4>
                    <ul>
                        {recommendations.map((rec, idx) => (
                            <li key={idx}>
                                <span className="rec-icon">💡</span>
                                {rec}
                            </li>
                        ))}
                    </ul>
                </div>
            )}
        </div>
    );
};

export default CombatAnalyticsView;