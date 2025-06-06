import React, { memo, useCallback, useState, useMemo } from 'react';
import { CharacterData, CharacterOptions } from './CharacterBuilder';
import { CustomClassForm } from './CustomClassForm';

interface ClassSelectionStepProps {
  characterData: CharacterData;
  onUpdate: (field: keyof CharacterData, value: any) => void;
  options: CharacterOptions;
  isCustomMode: boolean;
  errors: Record<string, string>;
  touched: Record<string, boolean>;
  setFieldTouched: (field: keyof CharacterData, touched: boolean) => void;
}

const formatClassName = (className: string): string => {
  return className
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
};

const ClassCard = memo(({
  className,
  isSelected,
  isCustom,
  onClick,
}: {
  className: string;
  isSelected: boolean;
  isCustom?: boolean;
  onClick: () => void;
}) => (
  <div
    className={`selection-card class-card ${isSelected ? 'selected' : ''} ${isCustom ? 'custom-class-card' : ''}`}
    onClick={onClick}
    role="button"
    tabIndex={0}
    onKeyDown={(e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        onClick();
      }
    }}
  >
    <h3>{isCustom ? '🎯 Custom Class' : formatClassName(className)}</h3>
    <p className="class-description">
      {isCustom ? 'Create your own unique class' : getClassTagline(className)}
    </p>
  </div>
));

ClassCard.displayName = 'ClassCard';

// Helper function to get class taglines
const getClassTagline = (className: string): string => {
  const taglines: Record<string, string> = {
    barbarian: 'Primal warrior of rage',
    bard: 'Master of song and magic',
    cleric: 'Divine agent of the gods',
    druid: 'Guardian of nature',
    fighter: 'Master of martial combat',
    monk: 'Warrior of ki and discipline',
    paladin: 'Holy warrior of justice',
    ranger: 'Hunter and tracker',
    rogue: 'Master of stealth and skill',
    sorcerer: 'Born of magic',
    warlock: 'Wielder of eldritch power',
    wizard: 'Scholar of the arcane',
  };
  
  return taglines[className.toLowerCase()] || 'Click to select';
};

export const ClassSelectionStep = memo(({
  characterData,
  onUpdate,
  options,
  isCustomMode,
  errors,
  touched,
  setFieldTouched,
}: ClassSelectionStepProps) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [showDetails, setShowDetails] = useState(false);

  const handleClassSelect = useCallback((className: string) => {
    onUpdate('class', className);
    setFieldTouched('class', true);
    setShowDetails(true);
  }, [onUpdate, setFieldTouched]);

  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  }, []);

  // Filter classes based on search
  const filteredClasses = useMemo(() => {
    if (!searchTerm) return options.classes;
    
    const term = searchTerm.toLowerCase();
    return options.classes.filter(cls => 
      formatClassName(cls).toLowerCase().includes(term) ||
      getClassTagline(cls).toLowerCase().includes(term)
    );
  }, [options.classes, searchTerm]);

  const showClassError = touched.class && errors.class;

  // Get class features for display
  const getClassFeatures = useCallback((className: string): string[] => {
    const features: Record<string, string[]> = {
      barbarian: ['Rage', 'Unarmored Defense', 'Reckless Attack', 'Danger Sense'],
      bard: ['Bardic Inspiration', 'Spellcasting', 'Jack of All Trades', 'Expertise'],
      cleric: ['Divine Spellcasting', 'Channel Divinity', 'Divine Domain', 'Healing'],
      druid: ['Druidic', 'Spellcasting', 'Wild Shape', 'Nature Magic'],
      fighter: ['Fighting Style', 'Second Wind', 'Action Surge', 'Extra Attack'],
      monk: ['Martial Arts', 'Ki', 'Unarmored Movement', 'Deflect Missiles'],
      paladin: ['Divine Sense', 'Lay on Hands', 'Fighting Style', 'Divine Smite'],
      ranger: ['Favored Enemy', 'Natural Explorer', 'Fighting Style', 'Spellcasting'],
      rogue: ['Sneak Attack', 'Thieves\' Cant', 'Cunning Action', 'Expertise'],
      sorcerer: ['Spellcasting', 'Sorcerous Origin', 'Font of Magic', 'Metamagic'],
      warlock: ['Otherworldly Patron', 'Pact Magic', 'Eldritch Invocations', 'Pact Boon'],
      wizard: ['Spellcasting', 'Arcane Recovery', 'Arcane Tradition', 'Ritual Casting'],
    };
    
    return features[className.toLowerCase()] || [];
  }, []);

  return (
    <div className="step-content class-selection-step">
      <h2>Choose Your Class</h2>
      <p className="step-description">
        Your class defines your character's abilities and role in the party.
      </p>

      {!isCustomMode && options.classes.length > 8 && (
        <div className="search-box">
          <input
            type="text"
            placeholder="Search classes..."
            value={searchTerm}
            onChange={handleSearchChange}
            className="search-input"
          />
        </div>
      )}

      {showClassError && (
        <div className="error-banner">{errors.class}</div>
      )}

      <div className="selection-grid class-grid">
        {filteredClasses.map(cls => (
          <ClassCard
            key={cls}
            className={cls}
            isSelected={characterData.class === cls}
            onClick={() => handleClassSelect(cls)}
          />
        ))}
        
        {(isCustomMode || options.aiEnabled) && (
          <ClassCard
            className="custom"
            isSelected={characterData.class === 'custom'}
            isCustom
            onClick={() => handleClassSelect('custom')}
          />
        )}
      </div>

      {characterData.class === 'custom' && (
        <CustomClassForm
          customClassData={characterData.customClassData}
          onUpdate={(data) => onUpdate('customClassData', data)}
          aiEnabled={options.aiEnabled}
        />
      )}

      {characterData.class && characterData.class !== 'custom' && showDetails && (
        <div className="class-details">
          <h3>{formatClassName(characterData.class)} Details</h3>
          <div className="class-info">
            <div className="class-features">
              <h4>Key Features:</h4>
              <ul>
                {getClassFeatures(characterData.class).map((feature, index) => (
                  <li key={index}>{feature}</li>
                ))}
              </ul>
            </div>
            <div className="class-description-full">
              <p>{getClassTagline(characterData.class)}</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
});

ClassSelectionStep.displayName = 'ClassSelectionStep';