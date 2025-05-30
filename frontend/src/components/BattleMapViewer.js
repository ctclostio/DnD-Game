import React, { useState, useEffect } from 'react';
import '../styles/battle-map.css';

const BattleMapViewer = ({ battleMap }) => {
    const [selectedCell, setSelectedCell] = useState(null);
    const [showTacticalNotes, setShowTacticalNotes] = useState(true);
    
    if (!battleMap) return null;

    const { 
        grid_size_x, 
        grid_size_y, 
        terrain_features = [],
        obstacle_positions = [],
        cover_positions = [],
        hazard_zones = [],
        spawn_points = {},
        tactical_notes = [],
        visual_theme = 'default'
    } = battleMap;

    // Create grid
    const grid = [];
    for (let y = 0; y < grid_size_y; y++) {
        const row = [];
        for (let x = 0; x < grid_size_x; x++) {
            row.push({ x, y });
        }
        grid.push(row);
    }

    const getCellContent = (x, y) => {
        // Check terrain features
        for (const feature of terrain_features) {
            if (feature.position.x === x && feature.position.y === y) {
                return { type: 'terrain', data: feature };
            }
        }

        // Check obstacles
        for (const obstacle of obstacle_positions) {
            if (obstacle.position.x === x && obstacle.position.y === y) {
                return { type: 'obstacle', data: obstacle };
            }
        }

        // Check cover
        for (const cover of cover_positions) {
            if (cover.position.x === x && cover.position.y === y) {
                return { type: 'cover', data: cover };
            }
        }

        // Check hazards
        for (const hazard of hazard_zones) {
            for (const cell of hazard.area || []) {
                if (cell.x === x && cell.y === y) {
                    return { type: 'hazard', data: hazard };
                }
            }
        }

        // Check spawn points
        if (spawn_points.party) {
            for (const spawn of spawn_points.party) {
                if (spawn.x === x && spawn.y === y) {
                    return { type: 'spawn-party', data: spawn };
                }
            }
        }

        if (spawn_points.enemies) {
            for (const spawn of spawn_points.enemies) {
                if (spawn.x === x && spawn.y === y) {
                    return { type: 'spawn-enemy', data: spawn };
                }
            }
        }

        return null;
    };

    const getCellClass = (content) => {
        if (!content) return 'empty';
        
        const classes = ['cell', content.type];
        
        if (content.type === 'terrain') {
            classes.push(`terrain-${content.data.type}`);
        }
        
        if (content.type === 'cover') {
            classes.push(`cover-${content.data.cover_type}`);
        }
        
        return classes.join(' ');
    };

    const getCellSymbol = (content) => {
        if (!content) return '';
        
        const symbols = {
            'terrain': {
                'wall': '█',
                'pillar': '◼',
                'tree': '🌳',
                'water': '💧',
                'elevation': '▲'
            },
            'obstacle': '▪',
            'cover': {
                'half': '◗',
                'three_quarters': '◧',
                'full': '■'
            },
            'hazard': '⚠',
            'spawn-party': 'P',
            'spawn-enemy': 'E'
        };

        if (content.type === 'terrain') {
            return symbols.terrain[content.data.type] || '?';
        }
        
        if (content.type === 'cover') {
            return symbols.cover[content.data.cover_type] || symbols.cover;
        }
        
        return symbols[content.type] || '';
    };

    const handleCellClick = (x, y) => {
        const content = getCellContent(x, y);
        setSelectedCell({ x, y, content });
    };

    const getTacticalNotesForCell = (x, y) => {
        return tactical_notes.filter(note => 
            note.position && note.position.x === x && note.position.y === y
        );
    };

    return (
        <div className={`battle-map-viewer theme-${visual_theme}`}>
            <div className="map-controls">
                <label>
                    <input
                        type="checkbox"
                        checked={showTacticalNotes}
                        onChange={(e) => setShowTacticalNotes(e.target.checked)}
                    />
                    Show Tactical Notes
                </label>
                <span className="grid-size">
                    Grid: {grid_size_x} × {grid_size_y}
                </span>
            </div>

            <div className="map-container">
                <div className="battle-grid" style={{
                    gridTemplateColumns: `repeat(${grid_size_x}, 1fr)`,
                    gridTemplateRows: `repeat(${grid_size_y}, 1fr)`
                }}>
                    {grid.map((row, y) => 
                        row.map((cell, x) => {
                            const content = getCellContent(x, y);
                            const notes = getTacticalNotesForCell(x, y);
                            const isSelected = selectedCell && selectedCell.x === x && selectedCell.y === y;
                            
                            return (
                                <div
                                    key={`${x}-${y}`}
                                    className={`grid-cell ${getCellClass(content)} ${isSelected ? 'selected' : ''}`}
                                    onClick={() => handleCellClick(x, y)}
                                    title={content?.data?.type || ''}
                                >
                                    <span className="cell-symbol">
                                        {getCellSymbol(content)}
                                    </span>
                                    {showTacticalNotes && notes.length > 0 && (
                                        <span className={`tactical-marker importance-${notes[0].importance}`}>
                                            ★
                                        </span>
                                    )}
                                </div>
                            );
                        })
                    )}
                </div>

                <div className="map-legend">
                    <h4>Legend</h4>
                    <div className="legend-items">
                        <div className="legend-item">
                            <span className="legend-symbol terrain-wall">█</span>
                            <span>Wall</span>
                        </div>
                        <div className="legend-item">
                            <span className="legend-symbol terrain-tree">🌳</span>
                            <span>Tree/Cover</span>
                        </div>
                        <div className="legend-item">
                            <span className="legend-symbol hazard">⚠</span>
                            <span>Hazard</span>
                        </div>
                        <div className="legend-item">
                            <span className="legend-symbol spawn-party">P</span>
                            <span>Party Spawn</span>
                        </div>
                        <div className="legend-item">
                            <span className="legend-symbol spawn-enemy">E</span>
                            <span>Enemy Spawn</span>
                        </div>
                    </div>
                </div>
            </div>

            {selectedCell && selectedCell.content && (
                <div className="cell-info">
                    <h4>Cell Details ({selectedCell.x}, {selectedCell.y})</h4>
                    <div className="info-content">
                        <p><strong>Type:</strong> {selectedCell.content.type}</p>
                        {selectedCell.content.type === 'terrain' && (
                            <>
                                <p><strong>Terrain:</strong> {selectedCell.content.data.type}</p>
                                {selectedCell.content.data.properties && (
                                    <p><strong>Properties:</strong> {selectedCell.content.data.properties.join(', ')}</p>
                                )}
                            </>
                        )}
                        {selectedCell.content.type === 'hazard' && (
                            <>
                                <p><strong>Hazard:</strong> {selectedCell.content.data.type}</p>
                                <p><strong>Damage:</strong> {selectedCell.content.data.damage_dice} {selectedCell.content.data.damage_type}</p>
                                <p><strong>Save:</strong> DC {selectedCell.content.data.save_dc} {selectedCell.content.data.save_type}</p>
                            </>
                        )}
                        {selectedCell.content.type === 'cover' && (
                            <p><strong>Cover:</strong> {selectedCell.content.data.cover_type}</p>
                        )}
                    </div>
                    
                    {getTacticalNotesForCell(selectedCell.x, selectedCell.y).map((note, idx) => (
                        <div key={idx} className={`tactical-note importance-${note.importance}`}>
                            <strong>Tactical Note:</strong> {note.note}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default BattleMapViewer;