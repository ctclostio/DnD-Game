import React, { memo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

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

interface CharacterCardProps {
  character: Character;
  onSelect?: (character: Character) => void;
  onDelete?: (id: string) => void;
  isSelected?: boolean;
  showActions?: boolean;
}

// Memoized character card component
export const CharacterCard = memo(({ 
  character, 
  onSelect, 
  onDelete,
  isSelected = false,
  showActions = true 
}: CharacterCardProps) => {
  const navigate = useNavigate();
  
  // Memoized callbacks to prevent re-renders
  const handleClick = useCallback(() => {
    if (onSelect) {
      onSelect(character);
    } else {
      navigate(`/characters/${character.id}`);
    }
  }, [character, onSelect, navigate]);
  
  const handleEdit = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(`/characters/${character.id}`);
  }, [character.id, navigate]);
  
  const handleDelete = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    if (onDelete && window.confirm(`Are you sure you want to delete ${character.name}?`)) {
      onDelete(character.id);
    }
  }, [character.id, character.name, onDelete]);
  
  const hpPercentage = (character.currentHP / character.maxHP) * 100;
  const hpColor = hpPercentage > 50 ? 'green' : hpPercentage > 25 ? 'yellow' : 'red';
  
  return (
    <div 
      className={`character-card ${isSelected ? 'selected' : ''}`}
      onClick={handleClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          handleClick();
        }
      }}
    >
      {character.imageUrl && (
        <img 
          src={character.imageUrl} 
          alt={character.name}
          className="character-avatar"
          loading="lazy"
        />
      )}
      
      <div className="character-info">
        <h3>{character.name}</h3>
        <p className="character-details">
          Level {character.level} {character.race} {character.class}
        </p>
        
        <div className="character-hp">
          <div className="hp-bar">
            <div 
              className={`hp-fill hp-${hpColor}`}
              style={{ width: `${hpPercentage}%` }}
            />
          </div>
          <span className="hp-text">
            {character.currentHP}/{character.maxHP} HP
          </span>
        </div>
      </div>
      
      {showActions && (
        <div className="character-actions">
          <button 
            onClick={handleEdit}
            className="action-btn edit-btn"
            aria-label={`Edit ${character.name}`}
          >
            Edit
          </button>
          {onDelete && (
            <button 
              onClick={handleDelete}
              className="action-btn delete-btn"
              aria-label={`Delete ${character.name}`}
            >
              Delete
            </button>
          )}
        </div>
      )}
    </div>
  );
}, (prevProps, nextProps) => {
  // Custom comparison function for memo
  return (
    prevProps.character.id === nextProps.character.id &&
    prevProps.character.currentHP === nextProps.character.currentHP &&
    prevProps.character.level === nextProps.character.level &&
    prevProps.isSelected === nextProps.isSelected &&
    prevProps.showActions === nextProps.showActions
  );
});

CharacterCard.displayName = 'CharacterCard';
