import React, { memo, useCallback, useState, useMemo } from 'react';
import { CharacterData, CharacterOptions } from './CharacterBuilder';

interface BackgroundSelectionStepProps {
  characterData: CharacterData;
  onUpdate: (field: keyof CharacterData, value: CharacterData[keyof CharacterData]) => void;
  options: CharacterOptions;
  errors: Record<string, string>;
  touched: Record<string, boolean>;
  setFieldTouched: (field: keyof CharacterData, touched: boolean) => void;
}

const formatBackgroundName = (background: string): string => {
  return background
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
};

const BackgroundCard = memo(({
  background,
  description,
  isSelected,
  onClick,
}: {
  background: string;
  description: string;
  isSelected: boolean;
  onClick: () => void;
}) => (
  <div
    className={`selection-card background-card ${isSelected ? 'selected' : ''}`}
    onClick={onClick}
    role="button"
    tabIndex={0}
    onKeyDown={(e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        onClick();
      }
    }}
  >
    <h3>{formatBackgroundName(background)}</h3>
    <p className="background-description">{description}</p>
  </div>
));

BackgroundCard.displayName = 'BackgroundCard';

// Background descriptions
const BACKGROUND_DESCRIPTIONS: Record<string, string> = {
  acolyte: 'You have spent your life in service to a temple',
  criminal: 'You have a criminal past and connections to the underworld',
  folk_hero: 'You come from humble beginnings but are destined for greatness',
  noble: 'You were born into wealth and privilege',
  sage: 'You spent years learning the lore of the multiverse',
  soldier: 'You fought in battles for your nation or lord',
  hermit: 'You lived in seclusion for a formative part of your life',
  entertainer: 'You thrive in front of an audience',
  guild_artisan: 'You are a member of an artisan\'s guild',
  outlander: 'You grew up in the wilds, far from civilization',
};

// Background features
const BACKGROUND_FEATURES: Record<string, { skills: string[]; languages: number; tools?: string[] }> = {
  acolyte: { skills: ['Insight', 'Religion'], languages: 2 },
  criminal: { skills: ['Deception', 'Stealth'], languages: 0, tools: ['Thieves\' tools', 'Gaming set'] },
  folk_hero: { skills: ['Animal Handling', 'Survival'], languages: 0, tools: ['Artisan\'s tools', 'Vehicles (land)'] },
  noble: { skills: ['History', 'Persuasion'], languages: 1, tools: ['Gaming set'] },
  sage: { skills: ['Arcana', 'History'], languages: 2 },
  soldier: { skills: ['Athletics', 'Intimidation'], languages: 0, tools: ['Gaming set', 'Vehicles (land)'] },
  hermit: { skills: ['Medicine', 'Religion'], languages: 1, tools: ['Herbalism kit'] },
  entertainer: { skills: ['Acrobatics', 'Performance'], languages: 0, tools: ['Disguise kit', 'Musical instrument'] },
  guild_artisan: { skills: ['Insight', 'Persuasion'], languages: 1, tools: ['Artisan\'s tools'] },
  outlander: { skills: ['Athletics', 'Survival'], languages: 1, tools: ['Musical instrument'] },
};

export const BackgroundSelectionStep = memo(({
  characterData,
  onUpdate,
  options,
  errors,
  touched,
  setFieldTouched,
}: BackgroundSelectionStepProps) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [showCustomForm, setShowCustomForm] = useState(false);
  const [customBackground, setCustomBackground] = useState({
    name: '',
    description: '',
    feature: '',
  });

  const handleBackgroundSelect = useCallback((background: string) => {
    onUpdate('background', background);
    setFieldTouched('background', true);
    setShowCustomForm(background === 'custom');
  }, [onUpdate, setFieldTouched]);

  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  }, []);

  // Filter backgrounds based on search
  const filteredBackgrounds = useMemo(() => {
    if (!searchTerm) return options.backgrounds;
    
    const term = searchTerm.toLowerCase();
    return options.backgrounds.filter(bg => 
      formatBackgroundName(bg).toLowerCase().includes(term) ||
      (BACKGROUND_DESCRIPTIONS[bg] || '').toLowerCase().includes(term)
    );
  }, [options.backgrounds, searchTerm]);

  const selectedBackgroundFeatures = useMemo(() => {
    if (!characterData.background || characterData.background === 'custom') return null;
    return BACKGROUND_FEATURES[characterData.background];
  }, [characterData.background]);

  const handleCustomBackgroundSubmit = useCallback(() => {
    if (customBackground.name && customBackground.description) {
      // In a real app, this would save the custom background
      console.debug('Custom background:', customBackground);
    }
  }, [customBackground]);

  const showBackgroundError = touched.background && errors.background;

  return (
    <div className="step-content background-selection-step">
      <h2>Choose Your Background</h2>
      <p className="step-description">
        Your background reveals where you came from and your place in the world.
      </p>

      {options.backgrounds.length > 6 && (
        <div className="search-box">
          <input
            type="text"
            placeholder="Search backgrounds..."
            value={searchTerm}
            onChange={handleSearchChange}
            className="search-input"
          />
        </div>
      )}

      {showBackgroundError && (
        <div className="error-banner">{errors.background}</div>
      )}

      <div className="selection-grid background-grid">
        {filteredBackgrounds.map(background => (
          <BackgroundCard
            key={background}
            background={background}
            description={BACKGROUND_DESCRIPTIONS[background] || 'A unique background'}
            isSelected={characterData.background === background}
            onClick={() => handleBackgroundSelect(background)}
          />
        ))}
        
        <BackgroundCard
          background="custom"
          description="Create your own unique background story"
          isSelected={characterData.background === 'custom'}
          onClick={() => handleBackgroundSelect('custom')}
        />
      </div>

      {characterData.background === 'custom' && showCustomForm && (
        <div className="custom-background-form">
          <h3>Create Custom Background</h3>
          
          <div className="form-group">
            <label htmlFor="customBackgroundName">Background Name</label>
            <input
              type="text"
              id="customBackgroundName"
              value={customBackground.name}
              onChange={(e) => setCustomBackground(prev => ({ ...prev, name: e.target.value }))}
              placeholder="e.g., Street Urchin, Wandering Scholar"
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="customBackgroundDesc">Description</label>
            <textarea
              id="customBackgroundDesc"
              rows={3}
              value={customBackground.description}
              onChange={(e) => setCustomBackground(prev => ({ ...prev, description: e.target.value }))}
              placeholder="Describe your character's background..."
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="customBackgroundFeature">Special Feature</label>
            <input
              type="text"
              id="customBackgroundFeature"
              value={customBackground.feature}
              onChange={(e) => setCustomBackground(prev => ({ ...prev, feature: e.target.value }))}
              placeholder="e.g., Guild Membership, Noble Privilege"
            />
          </div>
          
          <button 
            className="btn-primary"
            onClick={handleCustomBackgroundSubmit}
            disabled={!customBackground.name || !customBackground.description}
          >
            Save Custom Background
          </button>
        </div>
      )}

      {selectedBackgroundFeatures && (
        <div className="background-features">
          <h3>{formatBackgroundName(characterData.background)} Features</h3>
          <div className="features-grid">
            <div className="feature">
              <strong>Skill Proficiencies:</strong>
              <p>{selectedBackgroundFeatures.skills.join(', ')}</p>
            </div>
            
            {selectedBackgroundFeatures.languages > 0 && (
              <div className="feature">
                <strong>Languages:</strong>
                <p>Any {selectedBackgroundFeatures.languages}</p>
              </div>
            )}
            
            {selectedBackgroundFeatures.tools && (
              <div className="feature">
                <strong>Tool Proficiencies:</strong>
                <p>{selectedBackgroundFeatures.tools.join(', ')}</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
});

BackgroundSelectionStep.displayName = 'BackgroundSelectionStep';
