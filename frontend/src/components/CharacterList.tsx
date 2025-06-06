import React, { memo, useMemo, useCallback, useState } from 'react';
import { CharacterCard } from './CharacterCard';

interface Character {
  id: string;
  name: string;
  race: string;
  class: string;
  level: number;
  currentHP: number;
  maxHP: number;
  imageUrl?: string;
}

interface CharacterListProps {
  characters: Character[];
  onCharacterSelect?: (character: Character) => void;
  onCharacterDelete?: (id: string) => void;
  searchTerm?: string;
  filterBy?: {
    class?: string;
    race?: string;
    minLevel?: number;
    maxLevel?: number;
  };
  loading?: boolean;
}

export const CharacterList = memo(({
  characters,
  onCharacterSelect,
  onCharacterDelete,
  searchTerm = '',
  filterBy = {},
  loading = false,
}: CharacterListProps) => {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [sortBy, setSortBy] = useState<'name' | 'level' | 'class'>('name');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');
  
  // Memoized filtered and sorted characters
  const filteredCharacters = useMemo(() => {
    let filtered = characters;
    
    // Apply search filter
    if (searchTerm) {
      const term = searchTerm.toLowerCase();
      filtered = filtered.filter(char => 
        char.name.toLowerCase().includes(term) ||
        char.class.toLowerCase().includes(term) ||
        char.race.toLowerCase().includes(term)
      );
    }
    
    // Apply filters
    if (filterBy.class) {
      filtered = filtered.filter(char => char.class === filterBy.class);
    }
    if (filterBy.race) {
      filtered = filtered.filter(char => char.race === filterBy.race);
    }
    if (filterBy.minLevel !== undefined) {
      filtered = filtered.filter(char => char.level >= filterBy.minLevel!);
    }
    if (filterBy.maxLevel !== undefined) {
      filtered = filtered.filter(char => char.level <= filterBy.maxLevel!);
    }
    
    // Sort
    const sorted = [...filtered].sort((a, b) => {
      let compareValue = 0;
      
      switch (sortBy) {
        case 'name':
          compareValue = a.name.localeCompare(b.name);
          break;
        case 'level':
          compareValue = a.level - b.level;
          break;
        case 'class':
          compareValue = a.class.localeCompare(b.class);
          break;
      }
      
      return sortOrder === 'asc' ? compareValue : -compareValue;
    });
    
    return sorted;
  }, [characters, searchTerm, filterBy, sortBy, sortOrder]);
  
  // Memoized callbacks
  const handleCharacterSelect = useCallback((character: Character) => {
    setSelectedId(character.id);
    if (onCharacterSelect) {
      onCharacterSelect(character);
    }
  }, [onCharacterSelect]);
  
  const handleSort = useCallback((field: 'name' | 'level' | 'class') => {
    if (sortBy === field) {
      setSortOrder(prev => prev === 'asc' ? 'desc' : 'asc');
    } else {
      setSortBy(field);
      setSortOrder('asc');
    }
  }, [sortBy]);
  
  // Stats calculation
  const stats = useMemo(() => {
    const total = filteredCharacters.length;
    const avgLevel = total > 0 
      ? filteredCharacters.reduce((sum, char) => sum + char.level, 0) / total 
      : 0;
    
    const classCounts = filteredCharacters.reduce((acc, char) => {
      acc[char.class] = (acc[char.class] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);
    
    return { total, avgLevel, classCounts };
  }, [filteredCharacters]);
  
  if (loading) {
    return (
      <div className="character-list-loading">
        <div className="spinner" />
        <p>Loading characters...</p>
      </div>
    );
  }
  
  return (
    <div className="character-list-container">
      <div className="character-list-header">
        <div className="list-stats">
          <span>{stats.total} characters</span>
          {stats.total > 0 && (
            <span>Average level: {stats.avgLevel.toFixed(1)}</span>
          )}
        </div>
        
        <div className="sort-controls">
          <button 
            onClick={() => handleSort('name')}
            className={`sort-btn ${sortBy === 'name' ? 'active' : ''}`}
          >
            Name {sortBy === 'name' && (sortOrder === 'asc' ? '↑' : '↓')}
          </button>
          <button 
            onClick={() => handleSort('level')}
            className={`sort-btn ${sortBy === 'level' ? 'active' : ''}`}
          >
            Level {sortBy === 'level' && (sortOrder === 'asc' ? '↑' : '↓')}
          </button>
          <button 
            onClick={() => handleSort('class')}
            className={`sort-btn ${sortBy === 'class' ? 'active' : ''}`}
          >
            Class {sortBy === 'class' && (sortOrder === 'asc' ? '↑' : '↓')}
          </button>
        </div>
      </div>
      
      {filteredCharacters.length === 0 ? (
        <div className="empty-state">
          <p>No characters found</p>
          {searchTerm && (
            <p>Try adjusting your search term "{searchTerm}"</p>
          )}
        </div>
      ) : (
        <div className="character-grid">
          {filteredCharacters.map(character => (
            <CharacterCard
              key={character.id}
              character={character}
              onSelect={handleCharacterSelect}
              onDelete={onCharacterDelete}
              isSelected={selectedId === character.id}
            />
          ))}
        </div>
      )}
      
      {stats.total > 0 && (
        <div className="class-distribution">
          <h4>Class Distribution</h4>
          <div className="class-bars">
            {Object.entries(stats.classCounts).map(([className, count]) => (
              <div key={className} className="class-bar">
                <span className="class-name">{className}</span>
                <div className="bar">
                  <div 
                    className="bar-fill"
                    style={{ width: `${(count / stats.total) * 100}%` }}
                  />
                </div>
                <span className="class-count">{count}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
});

CharacterList.displayName = 'CharacterList';