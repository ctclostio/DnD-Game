import React, { memo, useMemo } from 'react';
import { CharacterData } from './CharacterBuilder';

interface ReviewStepProps {
  characterData: CharacterData;
}

const formatValue = (value: string): string => {
  return value
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
};

const AbilityScoreDisplay = memo(({
  ability,
  score,
}: {
  ability: string;
  score: number;
}) => {
  const modifier = Math.floor((score - 10) / 2);
  
  return (
    <div className="ability-score-display">
      <span className="ability-name">{ability}</span>
      <span className="score-value">{score}</span>
      <span className="modifier">
        ({modifier >= 0 ? '+' : ''}{modifier})
      </span>
    </div>
  );
});

AbilityScoreDisplay.displayName = 'AbilityScoreDisplay';

export const ReviewStep = memo(({ characterData }: ReviewStepProps) => {
  // Calculate derived stats
  const derivedStats = useMemo(() => {
    const conModifier = Math.floor(((characterData.abilityScores.CON || 10) - 10) / 2);
    
    // Hit points calculation (simplified - normally depends on class)
    const baseHP = 10; // Default, would vary by class
    const hitPoints = baseHP + conModifier;
    
    // AC calculation (simplified - normally includes armor)
    const dexModifier = Math.floor(((characterData.abilityScores.DEX || 10) - 10) / 2);
    const armorClass = 10 + dexModifier;
    
    // Initiative
    const initiative = dexModifier;
    
    // Proficiency bonus (level 1)
    const proficiencyBonus = 2;
    
    return {
      hitPoints,
      armorClass,
      initiative,
      proficiencyBonus,
    };
  }, [characterData.abilityScores]);

  const isCustomCharacter = characterData.race === 'custom' || characterData.class === 'custom';

  return (
    <div className="step-content review-step">
      <h2>Review Your Character</h2>
      <p className="step-description">
        Review your character details before creating. You can go back to make changes if needed.
      </p>

      {isCustomCharacter && (
        <div className="custom-character-notice">
          <p>🎨 This character includes custom content that will be balanced for play.</p>
        </div>
      )}

      <div className="review-sections">
        <section className="review-section basic-info-section">
          <h3>Basic Information</h3>
          <div className="review-grid">
            <div className="review-item">
              <label>Name:</label>
              <span>{characterData.name || 'Not set'}</span>
            </div>
            <div className="review-item">
              <label>Alignment:</label>
              <span>{characterData.alignment}</span>
            </div>
          </div>
        </section>

        <section className="review-section race-class-section">
          <h3>Race & Class</h3>
          <div className="review-grid">
            <div className="review-item">
              <label>Race:</label>
              <span>
                {characterData.race === 'custom' 
                  ? characterData.customRaceData?.name || 'Custom Race'
                  : formatValue(characterData.race || 'Not selected')
                }
              </span>
            </div>
            {characterData.subrace && (
              <div className="review-item">
                <label>Subrace:</label>
                <span>{formatValue(characterData.subrace)}</span>
              </div>
            )}
            <div className="review-item">
              <label>Class:</label>
              <span>
                {characterData.class === 'custom'
                  ? characterData.customClassData?.name || 'Custom Class'
                  : formatValue(characterData.class || 'Not selected')
                }
              </span>
            </div>
            <div className="review-item">
              <label>Background:</label>
              <span>{formatValue(characterData.background || 'Not selected')}</span>
            </div>
          </div>
        </section>

        <section className="review-section ability-scores-section">
          <h3>Ability Scores</h3>
          <div className="ability-scores-review">
            {['STR', 'DEX', 'CON', 'INT', 'WIS', 'CHA'].map(ability => (
              <AbilityScoreDisplay
                key={ability}
                ability={ability}
                score={characterData.abilityScores[ability] || 10}
              />
            ))}
          </div>
        </section>

        <section className="review-section skills-section">
          <h3>Selected Skills</h3>
          {characterData.selectedSkills.length > 0 ? (
            <ul className="skills-list">
              {characterData.selectedSkills.map(skill => (
                <li key={skill}>{skill}</li>
              ))}
            </ul>
          ) : (
            <p className="no-skills">No skills selected</p>
          )}
        </section>

        <section className="review-section derived-stats-section">
          <h3>Starting Statistics</h3>
          <div className="stats-grid">
            <div className="stat-item">
              <label>Hit Points:</label>
              <span>{derivedStats.hitPoints}</span>
            </div>
            <div className="stat-item">
              <label>Armor Class:</label>
              <span>{derivedStats.armorClass}</span>
            </div>
            <div className="stat-item">
              <label>Initiative:</label>
              <span>{derivedStats.initiative >= 0 ? '+' : ''}{derivedStats.initiative}</span>
            </div>
            <div className="stat-item">
              <label>Proficiency Bonus:</label>
              <span>+{derivedStats.proficiencyBonus}</span>
            </div>
            <div className="stat-item">
              <label>Speed:</label>
              <span>30 ft</span>
            </div>
            <div className="stat-item">
              <label>Level:</label>
              <span>1</span>
            </div>
          </div>
        </section>

        {characterData.customRaceData && (
          <section className="review-section custom-race-review">
            <h3>Custom Race Details</h3>
            <div className="custom-content">
              <h4>{characterData.customRaceData.name}</h4>
              <p>{characterData.customRaceData.description}</p>
              {characterData.customRaceData.desiredTraits && (
                <div className="custom-traits">
                  <strong>Desired Traits:</strong>
                  <p>{characterData.customRaceData.desiredTraits}</p>
                </div>
              )}
            </div>
          </section>
        )}

        {characterData.customClassData && (
          <section className="review-section custom-class-review">
            <h3>Custom Class Details</h3>
            <div className="custom-content">
              <h4>{characterData.customClassData.name}</h4>
              <p>{characterData.customClassData.description}</p>
              {characterData.customClassData.role && (
                <p><strong>Role:</strong> {formatValue(characterData.customClassData.role)}</p>
              )}
              {characterData.customClassData.playstyle && (
                <p><strong>Playstyle:</strong> {formatValue(characterData.customClassData.playstyle)}</p>
              )}
            </div>
          </section>
        )}
      </div>

      <div className="character-ready-message">
        <p>✨ Your character is ready to be created!</p>
        <p className="hint">Click "Create Character" to finalize, or use "Previous" to make changes.</p>
      </div>
    </div>
  );
});

ReviewStep.displayName = 'ReviewStep';