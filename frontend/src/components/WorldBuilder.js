import React, { useState, useEffect } from 'react';
import SettlementManager from './WorldBuilder/SettlementManager';
import FactionManager from './WorldBuilder/FactionManager';
import WorldEventViewer from './WorldBuilder/WorldEventViewer';
import EconomicDashboard from './WorldBuilder/EconomicDashboard';
import '../styles/world-builder.css';

const WorldBuilder = ({ sessionId }) => {
    const [activeTab, setActiveTab] = useState('settlements');
    const [worldData, setWorldData] = useState({
        settlements: [],
        factions: [],
        worldEvents: [],
        economicData: null
    });
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        if (sessionId) {
            loadWorldData();
        }
    }, [sessionId]);

    const loadWorldData = async () => {
        setLoading(true);
        setError(null);
        
        try {
            // Load all world data in parallel
            const [settlements, factions, worldEvents] = await Promise.all([
                fetchSettlements(),
                fetchFactions(),
                fetchWorldEvents()
            ]);

            setWorldData({
                settlements,
                factions,
                worldEvents,
                economicData: null // Will be loaded by economic dashboard
            });
        } catch (err) {
            setError('Failed to load world data');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const fetchSettlements = async () => {
        const response = await fetch(`/api/v1/sessions/${sessionId}/settlements`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });
        
        if (!response.ok) throw new Error('Failed to fetch settlements');
        return response.json();
    };

    const fetchFactions = async () => {
        const response = await fetch(`/api/v1/sessions/${sessionId}/factions`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });
        
        if (!response.ok) throw new Error('Failed to fetch factions');
        return response.json();
    };

    const fetchWorldEvents = async () => {
        const response = await fetch(`/api/v1/sessions/${sessionId}/world-events/active`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });
        
        if (!response.ok) throw new Error('Failed to fetch world events');
        return response.json();
    };

    const handleDataUpdate = (type, data) => {
        setWorldData(prev => ({
            ...prev,
            [type]: data
        }));
    };

    if (loading) {
        return (
            <div className="world-builder-loading">
                <div className="spinner"></div>
                <p>Loading world data...</p>
            </div>
        );
    }

    if (error) {
        return (
            <div className="world-builder-error">
                <p>{error}</p>
                <button onClick={loadWorldData}>Retry</button>
            </div>
        );
    }

    return (
        <div className="world-builder">
            <div className="world-builder-header">
                <h2>World Builder</h2>
                <p className="world-theme">Ancient World of Eternal Shadows</p>
            </div>

            <div className="world-builder-tabs">
                <button
                    className={`tab-button ${activeTab === 'settlements' ? 'active' : ''}`}
                    onClick={() => setActiveTab('settlements')}
                >
                    <span className="tab-icon">🏰</span>
                    Settlements
                </button>
                <button
                    className={`tab-button ${activeTab === 'factions' ? 'active' : ''}`}
                    onClick={() => setActiveTab('factions')}
                >
                    <span className="tab-icon">⚔️</span>
                    Factions
                </button>
                <button
                    className={`tab-button ${activeTab === 'events' ? 'active' : ''}`}
                    onClick={() => setActiveTab('events')}
                >
                    <span className="tab-icon">🌟</span>
                    World Events
                </button>
                <button
                    className={`tab-button ${activeTab === 'economy' ? 'active' : ''}`}
                    onClick={() => setActiveTab('economy')}
                >
                    <span className="tab-icon">💰</span>
                    Economy
                </button>
            </div>

            <div className="world-builder-content">
                {activeTab === 'settlements' && (
                    <SettlementManager
                        sessionId={sessionId}
                        settlements={worldData.settlements}
                        onUpdate={(data) => handleDataUpdate('settlements', data)}
                    />
                )}
                
                {activeTab === 'factions' && (
                    <FactionManager
                        sessionId={sessionId}
                        factions={worldData.factions}
                        settlements={worldData.settlements}
                        onUpdate={(data) => handleDataUpdate('factions', data)}
                    />
                )}
                
                {activeTab === 'events' && (
                    <WorldEventViewer
                        sessionId={sessionId}
                        worldEvents={worldData.worldEvents}
                        settlements={worldData.settlements}
                        factions={worldData.factions}
                        onUpdate={(data) => handleDataUpdate('worldEvents', data)}
                    />
                )}
                
                {activeTab === 'economy' && (
                    <EconomicDashboard
                        sessionId={sessionId}
                        settlements={worldData.settlements}
                        onUpdate={(data) => handleDataUpdate('economicData', data)}
                    />
                )}
            </div>

            <div className="world-summary">
                <h3>World Overview</h3>
                <div className="summary-stats">
                    <div className="stat">
                        <span className="stat-label">Settlements</span>
                        <span className="stat-value">{worldData.settlements.length}</span>
                    </div>
                    <div className="stat">
                        <span className="stat-label">Factions</span>
                        <span className="stat-value">{worldData.factions.length}</span>
                    </div>
                    <div className="stat">
                        <span className="stat-label">Active Events</span>
                        <span className="stat-value">{worldData.worldEvents.length}</span>
                    </div>
                    <div className="stat">
                        <span className="stat-label">World Corruption</span>
                        <span className="stat-value corruption">
                            {calculateWorldCorruption(worldData.settlements)}%
                        </span>
                    </div>
                </div>
            </div>
        </div>
    );
};

// Helper function to calculate overall world corruption
const calculateWorldCorruption = (settlements) => {
    if (!settlements || settlements.length === 0) return 0;
    
    const totalCorruption = settlements.reduce((sum, settlement) => 
        sum + (settlement.corruptionLevel || 0), 0
    );
    
    return Math.round((totalCorruption / settlements.length) * 10);
};

export default WorldBuilder;