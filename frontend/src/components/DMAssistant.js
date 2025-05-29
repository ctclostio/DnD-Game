import React, { useState, useEffect, useRef } from 'react';
import { api } from '../services/api';
import { getWebSocketService } from '../services/websocket';
import NPCDialogue from './DMAssistant/NPCDialogue';
import LocationGenerator from './DMAssistant/LocationGenerator';
import CombatNarrator from './DMAssistant/CombatNarrator';
import StoryElements from './DMAssistant/StoryElements';
import EnvironmentalHazards from './DMAssistant/EnvironmentalHazards';
import GeneratedContent from './DMAssistant/GeneratedContent';
import '../styles/dm-assistant.css';

const DMAssistant = ({ gameSessionId, currentCombat }) => {
    const [activeTab, setActiveTab] = useState('npc');
    const [isGenerating, setIsGenerating] = useState(false);
    const [generatedContent, setGeneratedContent] = useState([]);
    const [savedNPCs, setSavedNPCs] = useState([]);
    const [savedLocations, setSavedLocations] = useState([]);
    const [storyElements, setStoryElements] = useState([]);
    const wsRef = useRef(null);
    const requestIdRef = useRef(0);

    useEffect(() => {
        // Initialize WebSocket connection for real-time updates
        const ws = getWebSocketService();
        wsRef.current = ws;

        // Subscribe to DM Assistant messages
        const handleDMAssistantMessage = (message) => {
            if (message.type === 'dm_assistant_response' || 
                message.type.startsWith('dm_assistant_')) {
                handleWebSocketResponse(message);
            }
        };

        ws.subscribe('dm_assistant', handleDMAssistantMessage);

        // Load saved content
        loadSavedContent();

        return () => {
            ws.unsubscribe('dm_assistant', handleDMAssistantMessage);
        };
    }, [gameSessionId]);

    const loadSavedContent = async () => {
        try {
            // Load NPCs
            const npcsResponse = await api.get(`/dm-assistant/sessions/${gameSessionId}/npcs`);
            setSavedNPCs(npcsResponse.data);

            // Load locations
            const locationsResponse = await api.get(`/dm-assistant/sessions/${gameSessionId}/locations`);
            setSavedLocations(locationsResponse.data);

            // Load story elements
            const storyResponse = await api.get(`/dm-assistant/sessions/${gameSessionId}/story-elements`);
            setStoryElements(storyResponse.data);
        } catch (error) {
            console.error('Error loading saved content:', error);
        }
    };

    const handleWebSocketResponse = (message) => {
        if (message.streaming) {
            // Handle streaming updates
            updateGeneratingStatus(message);
        } else if (message.complete) {
            // Handle completed generation
            handleGenerationComplete(message);
        }
    };

    const updateGeneratingStatus = (message) => {
        // Update UI with streaming progress
        if (message.data.progress) {
            // Show progress bar or status
            console.log('Generation progress:', message.data.progress);
        }
    };

    const handleGenerationComplete = (message) => {
        setIsGenerating(false);
        
        // Add to generated content history
        const newContent = {
            id: Date.now(),
            type: message.type,
            data: message.data,
            timestamp: new Date()
        };
        
        setGeneratedContent(prev => [newContent, ...prev].slice(0, 10)); // Keep last 10

        // Refresh saved content if needed
        if (message.type === 'location_generation' || message.type === 'npc_generation') {
            loadSavedContent();
        }
    };

    const sendDMAssistantRequest = async (type, parameters, context = {}) => {
        const requestId = `req_${++requestIdRef.current}`;
        setIsGenerating(true);

        try {
            // Send via WebSocket for streaming support
            if (wsRef.current && wsRef.current.isConnected()) {
                wsRef.current.send({
                    type: 'dm_assistant_request',
                    requestId,
                    data: {
                        type,
                        gameSessionId,
                        parameters,
                        context
                    }
                });
            } else {
                // Fallback to HTTP API
                const response = await api.post('/dm-assistant', {
                    type,
                    gameSessionId,
                    parameters,
                    context,
                    streamResponse: false
                });
                
                handleGenerationComplete({
                    type: `${type}_response`,
                    data: response.data,
                    complete: true
                });
            }
        } catch (error) {
            console.error('DM Assistant error:', error);
            setIsGenerating(false);
        }
    };

    const tabs = [
        { id: 'npc', label: 'NPCs & Dialogue', icon: '🗣️' },
        { id: 'location', label: 'Locations', icon: '🏰' },
        { id: 'combat', label: 'Combat', icon: '⚔️' },
        { id: 'story', label: 'Story & Twists', icon: '📖' },
        { id: 'hazards', label: 'Hazards', icon: '⚠️' },
        { id: 'content', label: 'Generated', icon: '📜' }
    ];

    return (
        <div className="dm-assistant">
            <div className="dm-assistant-header">
                <h2>🧙‍♂️ AI Dungeon Master Assistant</h2>
                <div className="dm-assistant-status">
                    {isGenerating && (
                        <div className="generating-indicator">
                            <div className="spinner"></div>
                            <span>Generating...</span>
                        </div>
                    )}
                </div>
            </div>

            <div className="dm-assistant-tabs">
                {tabs.map(tab => (
                    <button
                        key={tab.id}
                        className={`dm-tab ${activeTab === tab.id ? 'active' : ''}`}
                        onClick={() => setActiveTab(tab.id)}
                        disabled={isGenerating}
                    >
                        <span className="tab-icon">{tab.icon}</span>
                        <span className="tab-label">{tab.label}</span>
                    </button>
                ))}
            </div>

            <div className="dm-assistant-content">
                {activeTab === 'npc' && (
                    <NPCDialogue
                        gameSessionId={gameSessionId}
                        savedNPCs={savedNPCs}
                        onGenerate={sendDMAssistantRequest}
                        isGenerating={isGenerating}
                    />
                )}

                {activeTab === 'location' && (
                    <LocationGenerator
                        gameSessionId={gameSessionId}
                        savedLocations={savedLocations}
                        onGenerate={sendDMAssistantRequest}
                        isGenerating={isGenerating}
                    />
                )}

                {activeTab === 'combat' && (
                    <CombatNarrator
                        gameSessionId={gameSessionId}
                        currentCombat={currentCombat}
                        onGenerate={sendDMAssistantRequest}
                        isGenerating={isGenerating}
                    />
                )}

                {activeTab === 'story' && (
                    <StoryElements
                        gameSessionId={gameSessionId}
                        storyElements={storyElements}
                        onGenerate={sendDMAssistantRequest}
                        onUseElement={(elementId) => {
                            api.post(`/dm-assistant/story-elements/${elementId}/use`);
                            loadSavedContent();
                        }}
                        isGenerating={isGenerating}
                    />
                )}

                {activeTab === 'hazards' && (
                    <EnvironmentalHazards
                        gameSessionId={gameSessionId}
                        currentLocation={savedLocations[0]} // Current location
                        onGenerate={sendDMAssistantRequest}
                        isGenerating={isGenerating}
                    />
                )}

                {activeTab === 'content' && (
                    <GeneratedContent
                        content={generatedContent}
                        onReuse={(item) => {
                            // Copy to clipboard or reuse content
                            navigator.clipboard.writeText(
                                typeof item.data === 'string' ? item.data : JSON.stringify(item.data, null, 2)
                            );
                        }}
                    />
                )}
            </div>

            {/* Quick Actions Bar */}
            <div className="dm-quick-actions">
                <h3>Quick Actions</h3>
                <div className="quick-action-buttons">
                    <button 
                        onClick={() => sendDMAssistantRequest('npc_dialogue', {
                            npcName: 'Quick NPC',
                            npcPersonality: ['mysterious', 'helpful'],
                            situation: 'Meeting the party for the first time'
                        })}
                        disabled={isGenerating}
                    >
                        🗣️ Quick Dialogue
                    </button>
                    <button 
                        onClick={() => sendDMAssistantRequest('plot_twist', {}, { 
                            currentPlot: 'Party investigating missing villagers' 
                        })}
                        disabled={isGenerating}
                    >
                        🎭 Generate Plot Twist
                    </button>
                    <button 
                        onClick={() => sendDMAssistantRequest('environmental_hazard', {
                            locationType: 'dungeon',
                            difficulty: 5
                        })}
                        disabled={isGenerating}
                    >
                        ⚡ Random Hazard
                    </button>
                </div>
            </div>
        </div>
    );
};

export default DMAssistant;