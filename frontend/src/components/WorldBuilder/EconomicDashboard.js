import React, { useState, useEffect } from 'react';
import { getClickableProps, getSelectableProps } from '../../utils/accessibility';

const EconomicDashboard = ({ sessionId, settlements }) => {
    const [tradeRoutes, setTradeRoutes] = useState([]);
    const [selectedRoute, setSelectedRoute] = useState(null);
    const [showRouteCreator, setShowRouteCreator] = useState(false);
    const [routeForm, setRouteForm] = useState({
        startSettlementId: '',
        endSettlementId: '',
        goodsTraded: [''],
        hazards: ['']
    });
    const [economicStats, setEconomicStats] = useState(null);

    useEffect(() => {
        loadTradeRoutes();
        loadEconomicStats();
    }, [sessionId]);

    const loadTradeRoutes = async () => {
        try {
            const response = await fetch(`/api/v1/sessions/${sessionId}/trade-routes`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                setTradeRoutes(data);
            }
        } catch (err) {
            console.error('Failed to load trade routes:', err);
        }
    };

    const loadEconomicStats = async () => {
        try {
            const response = await fetch(`/api/v1/sessions/${sessionId}/economic-stats`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (response.ok) {
                const stats = await response.json();
                setEconomicStats(stats);
            }
        } catch (err) {
            console.error('Failed to load economic stats:', err);
        }
    };

    const createTradeRoute = async (e) => {
        e.preventDefault();
        
        try {
            const response = await fetch(`/api/v1/sessions/${sessionId}/trade-routes`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    ...routeForm,
                    goodsTraded: routeForm.goodsTraded.filter(g => g.trim()),
                    hazards: routeForm.hazards.filter(h => h.trim())
                })
            });

            if (response.ok) {
                const newRoute = await response.json();
                setTradeRoutes(prev => [...prev, newRoute]);
                setShowRouteCreator(false);
                setRouteForm({
                    startSettlementId: '',
                    endSettlementId: '',
                    goodsTraded: [''],
                    hazards: ['']
                });
            }
        } catch (err) {
            console.error('Failed to create trade route:', err);
        }
    };

    const simulateMarkets = async () => {
        try {
            const response = await fetch(`/api/v1/sessions/${sessionId}/markets/simulate`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (response.ok) {
                await loadEconomicStats();
                alert('Market conditions updated');
            }
        } catch (err) {
            console.error('Failed to simulate markets:', err);
        }
    };

    const handleArrayChange = (field, index, value) => {
        setRouteForm(prev => ({
            ...prev,
            [field]: prev[field].map((item, i) => i === index ? value : item)
        }));
    };

    const addArrayItem = (field) => {
        setRouteForm(prev => ({
            ...prev,
            [field]: [...prev[field], '']
        }));
    };

    const removeArrayItem = (field, index) => {
        setRouteForm(prev => ({
            ...prev,
            [field]: prev[field].filter((_, i) => i !== index)
        }));
    };

    const getRouteStatus = (route) => {
        if (route.banditActivity > 7) return 'dangerous';
        if (route.banditActivity > 4) return 'risky';
        if (route.profitability > 7) return 'prosperous';
        if (route.profitability > 4) return 'stable';
        return 'struggling';
    };

    const getSettlementById = (id) => {
        return settlements.find(s => s.id === id);
    };

    return (
        <div className="economic-dashboard">
            <div className="dashboard-header">
                <h3>Economic Dashboard</h3>
                <div className="header-actions">
                    <button onClick={() => setShowRouteCreator(true)} className="btn btn-primary">
                        + Create Trade Route
                    </button>
                    <button onClick={simulateMarkets} className="btn btn-secondary">
                        Simulate Markets
                    </button>
                </div>
            </div>

            {economicStats && (
                <div className="economic-overview">
                    <h4>Economic Overview</h4>
                    <div className="stats-grid">
                        <div className="stat-card">
                            <span className="stat-icon">💰</span>
                            <div className="stat-info">
                                <span className="stat-label">Total Trade Volume</span>
                                <span className="stat-value">{economicStats.totalTradeVolume?.toLocaleString() || 0} gp</span>
                            </div>
                        </div>
                        <div className="stat-card">
                            <span className="stat-icon">📈</span>
                            <div className="stat-info">
                                <span className="stat-label">Economic Growth</span>
                                <span className="stat-value">{economicStats.economicGrowth || 0}%</span>
                            </div>
                        </div>
                        <div className="stat-card">
                            <span className="stat-icon">🛣️</span>
                            <div className="stat-info">
                                <span className="stat-label">Active Trade Routes</span>
                                <span className="stat-value">{tradeRoutes.filter(r => r.profitability > 0).length}</span>
                            </div>
                        </div>
                        <div className="stat-card">
                            <span className="stat-icon">⚠️</span>
                            <div className="stat-info">
                                <span className="stat-label">Economic Threats</span>
                                <span className="stat-value">{economicStats.threats || 0}</span>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            <div className="trade-routes-section">
                <h4>Trade Routes</h4>
                <div className="routes-layout">
                    <div className="routes-list">
                        {tradeRoutes.length === 0 ? (
                            <div className="empty-state">
                                <p>No trade routes established</p>
                                <p className="hint">Create trade routes between settlements</p>
                            </div>
                        ) : (
                            tradeRoutes.map(route => {
                                const startSettlement = getSettlementById(route.startSettlementId);
                                const endSettlement = getSettlementById(route.endSettlementId);
                                const status = getRouteStatus(route);
                                
                                return (
                                    <div
                                        key={route.id}
                                        className={`route-card ${selectedRoute?.id === route.id ? 'selected' : ''} ${status}`}
                                        {...getSelectableProps(() => setSelectedRoute(route), selectedRoute?.id === route.id)}
                                    >
                                        <div className="route-header">
                                            <div className="route-path">
                                                <span className="settlement">{startSettlement?.name || 'Unknown'}</span>
                                                <span className="arrow">→</span>
                                                <span className="settlement">{endSettlement?.name || 'Unknown'}</span>
                                            </div>
                                            <span className={`route-status ${status}`}>{status}</span>
                                        </div>

                                        <div className="route-stats">
                                            <div className="stat">
                                                <span>Distance:</span>
                                                <span>{route.distance} days</span>
                                            </div>
                                            <div className="stat">
                                                <span>Profitability:</span>
                                                <span>{route.profitability}/10</span>
                                            </div>
                                            <div className="stat">
                                                <span>Bandit Activity:</span>
                                                <span className={route.banditActivity > 5 ? 'danger' : ''}>
                                                    {route.banditActivity}/10
                                                </span>
                                            </div>
                                        </div>

                                        <div className="goods-traded">
                                            {route.goodsTraded?.slice(0, 3).map((good, index) => (
                                                <span key={index} className="good-tag">{good}</span>
                                            ))}
                                            {route.goodsTraded?.length > 3 && (
                                                <span className="more-goods">+{route.goodsTraded.length - 3} more</span>
                                            )}
                                        </div>
                                    </div>
                                );
                            })
                        )}
                    </div>

                    {selectedRoute && (
                        <div className="route-detail-panel">
                            <h4>Route Details</h4>
                            
                            <div className="detail-section">
                                <h5>Route Information</h5>
                                <p><strong>Distance:</strong> {selectedRoute.distance} days of travel</p>
                                <p><strong>Profitability:</strong> {selectedRoute.profitability}/10</p>
                                <p><strong>Bandit Activity:</strong> {selectedRoute.banditActivity}/10</p>
                                <p><strong>Protection Cost:</strong> {selectedRoute.protectionCost} gp/trip</p>
                            </div>

                            <div className="detail-section">
                                <h5>Goods Traded</h5>
                                <div className="goods-list">
                                    {selectedRoute.goodsTraded?.map((good, index) => (
                                        <span key={index} className="good-item">{good}</span>
                                    ))}
                                </div>
                            </div>

                            <div className="detail-section">
                                <h5>Hazards</h5>
                                {selectedRoute.hazards?.length > 0 ? (
                                    <ul>
                                        {selectedRoute.hazards.map((hazard, index) => (
                                            <li key={index}>{hazard}</li>
                                        ))}
                                    </ul>
                                ) : (
                                    <p>No known hazards</p>
                                )}
                            </div>

                            {selectedRoute.controllingFaction && (
                                <div className="detail-section">
                                    <h5>Controlling Faction</h5>
                                    <p>{selectedRoute.controllingFaction}</p>
                                </div>
                            )}

                            {selectedRoute.ancientSiteNearby && (
                                <div className="ancient-warning">
                                    <span className="icon">🗿</span>
                                    <span>Route passes near ancient sites - increased supernatural hazards</span>
                                </div>
                            )}

                            {selectedRoute.leyLineIntersection && (
                                <div className="leyline-note">
                                    <span className="icon">✨</span>
                                    <span>Crosses ley lines - magical goods trade more profitable</span>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>

            {showRouteCreator && (
                <div className="modal-overlay" {...getClickableProps(() => setShowRouteCreator(false))}>
                    <div className="modal-content route-creator" {...getClickableProps((e) => e.stopPropagation())}>
                        <div className="modal-header">
                            <h3>Create Trade Route</h3>
                            <button className="close-button" onClick={() => setShowRouteCreator(false)}>×</button>
                        </div>

                        <form onSubmit={createTradeRoute}>
                            <div className="form-row">
                                <div className="form-group">
                                    <label>Start Settlement*</label>
                                    <select
                                        value={routeForm.startSettlementId}
                                        onChange={(e) => setRouteForm(prev => ({ 
                                            ...prev, 
                                            startSettlementId: e.target.value 
                                        }))}
                                        required
                                    >
                                        <option value="">Select Settlement</option>
                                        {settlements.map(settlement => (
                                            <option key={settlement.id} value={settlement.id}>
                                                {settlement.name} ({settlement.type})
                                            </option>
                                        ))}
                                    </select>
                                </div>

                                <div className="form-group">
                                    <label>End Settlement*</label>
                                    <select
                                        value={routeForm.endSettlementId}
                                        onChange={(e) => setRouteForm(prev => ({ 
                                            ...prev, 
                                            endSettlementId: e.target.value 
                                        }))}
                                        required
                                    >
                                        <option value="">Select Settlement</option>
                                        {settlements
                                            .filter(s => s.id !== routeForm.startSettlementId)
                                            .map(settlement => (
                                                <option key={settlement.id} value={settlement.id}>
                                                    {settlement.name} ({settlement.type})
                                                </option>
                                            ))}
                                    </select>
                                </div>
                            </div>

                            <div className="form-group">
                                <label>Goods Traded</label>
                                {routeForm.goodsTraded.map((good, index) => (
                                    <div key={index} className="array-input">
                                        <input
                                            type="text"
                                            value={good}
                                            onChange={(e) => handleArrayChange('goodsTraded', index, e.target.value)}
                                            placeholder="e.g., Grain, Iron, Magical Components"
                                        />
                                        {routeForm.goodsTraded.length > 1 && (
                                            <button
                                                type="button"
                                                onClick={() => removeArrayItem('goodsTraded', index)}
                                                className="remove-button"
                                            >
                                                ×
                                            </button>
                                        )}
                                    </div>
                                ))}
                                <button
                                    type="button"
                                    onClick={() => addArrayItem('goodsTraded')}
                                    className="add-button"
                                >
                                    + Add Good
                                </button>
                            </div>

                            <div className="form-group">
                                <label>Known Hazards</label>
                                {routeForm.hazards.map((hazard, index) => (
                                    <div key={index} className="array-input">
                                        <input
                                            type="text"
                                            value={hazard}
                                            onChange={(e) => handleArrayChange('hazards', index, e.target.value)}
                                            placeholder="e.g., Bandit camps, Monster territory"
                                        />
                                        {routeForm.hazards.length > 1 && (
                                            <button
                                                type="button"
                                                onClick={() => removeArrayItem('hazards', index)}
                                                className="remove-button"
                                            >
                                                ×
                                            </button>
                                        )}
                                    </div>
                                ))}
                                <button
                                    type="button"
                                    onClick={() => addArrayItem('hazards')}
                                    className="add-button"
                                >
                                    + Add Hazard
                                </button>
                            </div>

                            <div className="creator-actions">
                                <button type="button" onClick={() => setShowRouteCreator(false)}>
                                    Cancel
                                </button>
                                <button 
                                    type="submit" 
                                    className="btn btn-primary"
                                    disabled={!routeForm.startSettlementId || !routeForm.endSettlementId}
                                >
                                    Create Route
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};

export default EconomicDashboard;