import React, { useState } from 'react';
import { getClickableProps, getSelectableProps } from '../../utils/accessibility';

const NPCDialogue = ({ gameSessionId, savedNPCs, onGenerate, isGenerating }) => {
    const [selectedNPC, setSelectedNPC] = useState(null);
    const [newNPCForm, setNewNPCForm] = useState({
        name: '',
        race: '',
        occupation: '',
        personality: '',
        voiceStyle: '',
        motivations: ''
    });
    const [dialogueForm, setDialogueForm] = useState({
        situation: '',
        playerInput: '',
        previousContext: ''
    });
    const [showNewNPCForm, setShowNewNPCForm] = useState(false);

    const handleCreateNPC = () => {
        const personalityTraits = newNPCForm.personality.split(',').map(t => t.trim());
        
        onGenerate('npc_creation', {
            role: `${newNPCForm.occupation} ${newNPCForm.race}`,
            context: {
                name: newNPCForm.name,
                race: newNPCForm.race,
                occupation: newNPCForm.occupation,
                personality: personalityTraits,
                voiceStyle: newNPCForm.voiceStyle,
                motivations: newNPCForm.motivations
            }
        });

        setShowNewNPCForm(false);
        setNewNPCForm({
            name: '',
            race: '',
            occupation: '',
            personality: '',
            voiceStyle: '',
            motivations: ''
        });
    };

    const handleGenerateDialogue = () => {
        if (!selectedNPC) {
            alert('Please select an NPC first');
            return;
        }

        onGenerate('npc_dialogue', {
            npcName: selectedNPC.name,
            npcPersonality: selectedNPC.personalityTraits,
            dialogueStyle: selectedNPC.dialogueStyle,
            situation: dialogueForm.situation,
            playerInput: dialogueForm.playerInput,
            previousContext: dialogueForm.previousContext
        });
    };

    const quickPersonalities = [
        { label: 'Gruff Veteran', traits: 'battle-hardened, cynical, protective' },
        { label: 'Wise Sage', traits: 'knowledgeable, cryptic, patient' },
        { label: 'Cheerful Merchant', traits: 'friendly, opportunistic, gossipy' },
        { label: 'Mysterious Stranger', traits: 'secretive, observant, cautious' },
        { label: 'Noble Knight', traits: 'honorable, brave, dutiful' }
    ];

    const quickSituations = [
        'First meeting with the party',
        'Party asking for information',
        'During combat',
        'Making a deal or trade',
        'Revealing important information',
        'Casual conversation in tavern'
    ];

    return (
        <div className="npc-dialogue-panel">
            <div className="panel-header">
                <h3>NPC Dialogue Generator</h3>
                <button 
                    className="create-npc-btn"
                    onClick={() => setShowNewNPCForm(!showNewNPCForm)}
                >
                    + Create New NPC
                </button>
            </div>

            {showNewNPCForm && (
                <div className="new-npc-form">
                    <h4>Create New NPC</h4>
                    <div className="form-grid">
                        <input
                            type="text"
                            placeholder="NPC Name"
                            value={newNPCForm.name}
                            onChange={(e) => setNewNPCForm({...newNPCForm, name: e.target.value})}
                        />
                        <input
                            type="text"
                            placeholder="Race (e.g., Human, Elf)"
                            value={newNPCForm.race}
                            onChange={(e) => setNewNPCForm({...newNPCForm, race: e.target.value})}
                        />
                        <input
                            type="text"
                            placeholder="Occupation (e.g., Blacksmith, Innkeeper)"
                            value={newNPCForm.occupation}
                            onChange={(e) => setNewNPCForm({...newNPCForm, occupation: e.target.value})}
                        />
                        <input
                            type="text"
                            placeholder="Voice/Speech Style"
                            value={newNPCForm.voiceStyle}
                            onChange={(e) => setNewNPCForm({...newNPCForm, voiceStyle: e.target.value})}
                        />
                    </div>
                    
                    <div className="personality-input">
                        <label>Personality Traits (comma-separated)</label>
                        <input
                            type="text"
                            placeholder="e.g., friendly, mysterious, cautious"
                            value={newNPCForm.personality}
                            onChange={(e) => setNewNPCForm({...newNPCForm, personality: e.target.value})}
                        />
                        <div className="quick-personality-chips">
                            {quickPersonalities.map((qp, idx) => (
                                <button
                                    key={idx}
                                    className="chip"
                                    onClick={() => setNewNPCForm({...newNPCForm, personality: qp.traits})}
                                >
                                    {qp.label}
                                </button>
                            ))}
                        </div>
                    </div>

                    <textarea
                        placeholder="Motivations and goals..."
                        value={newNPCForm.motivations}
                        onChange={(e) => setNewNPCForm({...newNPCForm, motivations: e.target.value})}
                        rows={3}
                    />

                    <div className="form-actions">
                        <button onClick={handleCreateNPC} disabled={isGenerating}>
                            Generate NPC
                        </button>
                        <button onClick={() => setShowNewNPCForm(false)} className="cancel-btn">
                            Cancel
                        </button>
                    </div>
                </div>
            )}

            <div className="npc-selection">
                <h4>Select NPC</h4>
                <div className="npc-grid">
                    {savedNPCs.map(npc => (
                        <div
                            key={npc.id}
                            className={`npc-card ${selectedNPC?.id === npc.id ? 'selected' : ''}`}
                            {...getSelectableProps(() => setSelectedNPC(npc), selectedNPC?.id === npc.id)}
                        >
                            <div className="npc-name">{npc.name}</div>
                            <div className="npc-details">
                                {npc.race} {npc.occupation}
                            </div>
                            <div className="npc-traits">
                                {npc.personalityTraits?.slice(0, 3).join(', ')}
                            </div>
                        </div>
                    ))}
                    {savedNPCs.length === 0 && (
                        <div className="empty-state">
                            No NPCs created yet. Create your first NPC above!
                        </div>
                    )}
                </div>
            </div>

            {selectedNPC && (
                <div className="dialogue-generator">
                    <h4>Generate Dialogue for {selectedNPC.name}</h4>
                    
                    <div className="npc-info">
                        <p><strong>Personality:</strong> {selectedNPC.personalityTraits?.join(', ')}</p>
                        <p><strong>Voice:</strong> {selectedNPC.voiceDescription}</p>
                        <p><strong>Motivations:</strong> {selectedNPC.motivations}</p>
                    </div>

                    <div className="dialogue-form">
                        <div className="situation-select">
                            <label>Situation</label>
                            <select
                                value={dialogueForm.situation}
                                onChange={(e) => setDialogueForm({...dialogueForm, situation: e.target.value})}
                            >
                                <option value="">Custom situation...</option>
                                {quickSituations.map((sit, idx) => (
                                    <option key={idx} value={sit}>{sit}</option>
                                ))}
                            </select>
                            {!dialogueForm.situation && (
                                <input
                                    type="text"
                                    placeholder="Describe the situation..."
                                    onChange={(e) => setDialogueForm({...dialogueForm, situation: e.target.value})}
                                />
                            )}
                        </div>

                        <div className="player-input">
                            <label>What the player says/does:</label>
                            <textarea
                                placeholder="e.g., 'Have you seen any strange activity lately?'"
                                value={dialogueForm.playerInput}
                                onChange={(e) => setDialogueForm({...dialogueForm, playerInput: e.target.value})}
                                rows={2}
                            />
                        </div>

                        <div className="context-input">
                            <label>Previous Context (optional):</label>
                            <textarea
                                placeholder="Any relevant previous interactions or context..."
                                value={dialogueForm.previousContext}
                                onChange={(e) => setDialogueForm({...dialogueForm, previousContext: e.target.value})}
                                rows={2}
                            />
                        </div>

                        <button
                            className="generate-dialogue-btn"
                            onClick={handleGenerateDialogue}
                            disabled={isGenerating || !dialogueForm.playerInput}
                        >
                            Generate Dialogue Response
                        </button>
                    </div>

                    {/* Dialogue History */}
                    {selectedNPC.generatedDialogue && selectedNPC.generatedDialogue.length > 0 && (
                        <div className="dialogue-history">
                            <h5>Recent Dialogue</h5>
                            {selectedNPC.generatedDialogue.slice(-5).map((entry, idx) => (
                                <div key={idx} className="dialogue-entry">
                                    <div className="dialogue-context">{entry.context}</div>
                                    <div className="dialogue-text">"{entry.dialogue}"</div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};

export default NPCDialogue;