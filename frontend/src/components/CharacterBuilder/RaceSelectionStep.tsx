import React, { memo, useCallback, useState, useMemo } from 'react';
import { CharacterData, CharacterOptions } from './CharacterBuilder';
import { CustomRaceForm } from './CustomRaceForm';

interface RaceSelectionStepProps {
  characterData: CharacterData;
  onUpdate: (field: keyof CharacterData, value: any) => void;
  onMultipleUpdate: (updates: Partial<CharacterData>) => void;
  options: CharacterOptions;
  isCustomMode: boolean;
  errors: Record<string, string>;
  touched: Record<string, boolean>;
  setFieldTouched: (field: keyof CharacterData, touched: boolean) => void;
}

const formatRaceName = (race: string): string => {
  return race
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
};

const RaceCard = memo(({
  race,
  isSelected,
  isCustom,
  onClick,
}: {
  race: string;
  isSelected: boolean;
  isCustom?: boolean;
  onClick: () => void;
}) => (
  <div
    className={`selection-card race-card ${isSelected ? 'selected' : ''} ${isCustom ? 'custom-race-card' : ''}`}
    onClick={onClick}
    role="button"
    tabIndex={0}
    onKeyDown={(e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        onClick();
      }
    }}
  >
    <h3>{isCustom ? '🎨 Custom Race' : formatRaceName(race)}</h3>
    <p className="race-description">
      {isCustom ? 'Create your own unique race' : 'Click to select'}
    </p>
  </div>
));

RaceCard.displayName = 'RaceCard';

export const RaceSelectionStep = memo(({
  characterData,
  onUpdate,
  onMultipleUpdate,
  options,
  isCustomMode,
  errors,
  touched,
  setFieldTouched,
}: RaceSelectionStepProps) => {
  const [searchTerm, setSearchTerm] = useState('');

  const handleRaceSelect = useCallback((race: string) => {
    onUpdate('race', race);
    setFieldTouched('race', true);
    
    // Clear subrace when changing race
    if (race !== characterData.race) {
      onUpdate('subrace', '');
    }
  }, [onUpdate, setFieldTouched, characterData.race]);

  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  }, []);

  // Filter races based on search
  const filteredRaces = useMemo(() => {
    if (!searchTerm) return options.races;
    
    const term = searchTerm.toLowerCase();
    return options.races.filter(race => 
      formatRaceName(race).toLowerCase().includes(term)
    );
  }, [options.races, searchTerm]);

  const showRaceError = touched.race && errors.race;

  return (
    <div className="step-content race-selection-step">
      <h2>Choose Your Race</h2>
      <p className="step-description">
        Your race determines your character's ancestry and grants unique abilities.
      </p>

      {!isCustomMode && options.races.length > 8 && (
        <div className="search-box">
          <input
            type="text"
            placeholder="Search races..."
            value={searchTerm}
            onChange={handleSearchChange}
            className="search-input"
          />
        </div>
      )}

      {showRaceError && (
        <div className="error-banner">{errors.race}</div>
      )}

      <div className="selection-grid race-grid">
        {filteredRaces.map(race => (
          <RaceCard
            key={race}
            race={race}
            isSelected={characterData.race === race}
            onClick={() => handleRaceSelect(race)}
          />
        ))}
        
        {(isCustomMode || options.aiEnabled) && (
          <RaceCard
            race="custom"
            isSelected={characterData.race === 'custom'}
            isCustom
            onClick={() => handleRaceSelect('custom')}
          />
        )}
      </div>

      {characterData.race === 'custom' && (
        <CustomRaceForm
          customRaceData={characterData.customRaceData}
          onUpdate={(data) => onUpdate('customRaceData', data)}
          aiEnabled={options.aiEnabled}
        />
      )}

      {characterData.race && characterData.race !== 'custom' && (
        <div className="race-details">
          <h3>Race Details: {formatRaceName(characterData.race)}</h3>
          <div className="race-info">
            <p>Loading race information...</p>
          </div>
        </div>
      )}
    </div>
  );
});

RaceSelectionStep.displayName = 'RaceSelectionStep';