import React, { useState, useEffect } from 'react';
import InitiativeTracker from './InitiativeTracker';
import CombatView from './CombatView';
import DiceRollerView from './DiceRollerView';
import CampaignManager from './CampaignManager';
import { getGameSession, getCharacters, getNPCsBySession } from '../services/api';
import { connectWebSocket, disconnectWebSocket, onMessage, sendMessage } from '../services/websocket';
import '../styles/game-session.css';

const GameSession = ({ sessionId, userId, isDM }) => {
    const [session, setSession] = useState(null);
    const [characters, setCharacters] = useState([]);
    const [npcs, setNpCs] = useState([]);
    const [activeCombat, setActiveCombat] = useState(null);
    const [activeTab, setActiveTab] = useState('initiative');
    const [chatMessages, setChatMessages] = useState([]);
    const [chatInput, setChatInput] = useState('');
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        if (sessionId) {
            loadSessionData();
            setupWebSocket();
        }

        return () => {
            disconnectWebSocket();
        };
    }, [sessionId]);

    const loadSessionData = async () => {
        setLoading(true);
        try {
            const [sessionRes, charsRes, npcsRes] = await Promise.all([
                getGameSession(sessionId),
                getCharacters(),
                getNPCsBySession(sessionId)
            ]);

            setSession(sessionRes.data);
            setCharacters(charsRes.data || []);
            setNPCs(npcsRes.data || []);
            setError('');
        } catch (err) {
            setError('Failed to load session data');
            console.error('Error loading session:', err);
        } finally {
            setLoading(false);
        }
    };

    const setupWebSocket = () => {
        connectWebSocket(sessionId);
        
        onMessage((message) => {
            switch (message.type) {
                case 'combat':
                    handleCombatUpdate(message.data);
                    break;
                case 'chat':
                    handleChatMessage(message.data);
                    break;
                case 'dice_roll':
                    handleDiceRoll(message.data);
                    break;
                default:
                    console.debug('Unknown message type:', message.type);
            }
        });
    };

    const handleCombatUpdate = (data) => {
        if (data.type === 'combat_started' || data.type === 'turn_changed') {
            setActiveCombat(data.combat);
        } else if (data.type === 'combat_ended') {
            setActiveCombat(null);
        }
    };

    const handleChatMessage = (data) => {
        setChatMessages(prev => [...prev, {
            id: Date.now(),
            user: data.user,
            message: data.message,
            timestamp: new Date().toLocaleTimeString()
        }]);
    };

    const handleDiceRoll = (data) => {
        setChatMessages(prev => [...prev, {
            id: Date.now(),
            user: data.user,
            message: `🎲 Rolled ${data.roll}: ${data.result} (${data.purpose || 'General'})`,
            timestamp: new Date().toLocaleTimeString(),
            isDiceRoll: true
        }]);
    };

    const sendChatMessage = () => {
        if (chatInput.trim()) {
            sendMessage({
                type: 'chat',
                data: {
                    message: chatInput,
                    sessionId: sessionId
                }
            });
            setChatInput('');
        }
    };

    if (loading) return <div className="loading">Loading session...</div>;
    if (error) return <div className="error">{error}</div>;
    if (!session) return <div className="error">Session not found</div>;

    return (
        <div className="game-session">
            <div className="session-header">
                <h2>{session.name}</h2>
                <div className="session-info">
                    <span className="session-code">Code: {session.code}</span>
                    <span className="participant-count">
                        {session.participants?.length || 0} Players
                    </span>
                    {isDM && <span className="dm-badge">DM</span>}
                </div>
            </div>

            <div className="session-content">
                <div className="main-panel">
                    <div className="tab-navigation">
                        <button 
                            className={activeTab === 'initiative' ? 'active' : ''}
                            onClick={() => setActiveTab('initiative')}
                        >
                            Initiative Tracker
                        </button>
                        <button 
                            className={activeTab === 'combat' ? 'active' : ''}
                            onClick={() => setActiveTab('combat')}
                            disabled={!activeCombat}
                        >
                            Combat View
                        </button>
                        <button 
                            className={activeTab === 'dice' ? 'active' : ''}
                            onClick={() => setActiveTab('dice')}
                        >
                            Dice Roller
                        </button>
                        <button 
                            className={activeTab === 'campaign' ? 'active' : ''}
                            onClick={() => setActiveTab('campaign')}
                        >
                            Campaign Manager
                        </button>
                    </div>

                    <div className="tab-content">
                        {activeTab === 'initiative' && (
                            <InitiativeTracker
                                gameSessionId={sessionId}
                                characters={characters.filter(c => 
                                    session.participants?.some(p => p.character_id === c.id)
                                )}
                                npcs={npcs}
                                onCombatUpdate={setActiveCombat}
                            />
                        )}

                        {activeTab === 'combat' && activeCombat && (
                            <CombatView
                                combat={activeCombat}
                                userId={userId}
                                isDM={isDM}
                            />
                        )}

                        {activeTab === 'dice' && (
                            <DiceRollerView sessionId={sessionId} />
                        )}
                        
                        {activeTab === 'campaign' && (
                            <CampaignManager 
                                gameSessionId={sessionId} 
                                isDM={isDM}
                            />
                        )}
                    </div>
                </div>

                <div className="side-panel">
                    <div className="chat-container">
                        <h3>Chat & Activity</h3>
                        <div className="chat-messages">
                            {chatMessages.map(msg => (
                                <div 
                                    key={msg.id} 
                                    className={`chat-message ${msg.isDiceRoll ? 'dice-roll' : ''}`}
                                >
                                    <span className="chat-user">{msg.user}:</span>
                                    <span className="chat-text">{msg.message}</span>
                                    <span className="chat-time">{msg.timestamp}</span>
                                </div>
                            ))}
                        </div>
                        <div className="chat-input-container">
                            <input
                                type="text"
                                value={chatInput}
                                onChange={(e) => setChatInput(e.target.value)}
                                onKeyPress={(e) => e.key === 'Enter' && sendChatMessage()}
                                placeholder="Type a message..."
                                className="chat-input"
                            />
                            <button onClick={sendChatMessage} className="chat-send">
                                Send
                            </button>
                        </div>
                    </div>

                    {isDM && (
                        <div className="dm-controls">
                            <h3>DM Controls</h3>
                            <button className="dm-button">Manage NPCs</button>
                            <button className="dm-button">Session Settings</button>
                            <button className="dm-button">End Session</button>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default GameSession;