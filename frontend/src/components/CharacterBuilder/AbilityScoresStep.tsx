import React, { memo, useCallback, useState, useMemo } from 'react';
import { CharacterData } from './CharacterBuilder';

interface AbilityScoresStepProps {
  characterData: CharacterData;
  onUpdate: (field: keyof CharacterData, value: any) => void;
  onMultipleUpdate: (updates: Partial<CharacterData>) => void;
  errors: Record<string, string>;
  touched: Record<string, boolean>;
  setFieldTouched: (field: keyof CharacterData, touched: boolean) => void;
}

const ABILITIES = ['STR', 'DEX', 'CON', 'INT', 'WIS', 'CHA'];
const ABILITY_NAMES = {
  STR: 'Strength',
  DEX: 'Dexterity',
  CON: 'Constitution',
  INT: 'Intelligence',
  WIS: 'Wisdom',
  CHA: 'Charisma',
};

const STANDARD_ARRAY = [15, 14, 13, 12, 10, 8];
const POINT_BUY_TOTAL = 27;
const POINT_BUY_COSTS: Record<number, number> = {
  8: 0, 9: 1, 10: 2, 11: 3, 12: 4, 13: 5, 14: 7, 15: 9,
};

const AbilityScoreInput = memo(({
  ability,
  value,
  onChange,
  min = 3,
  max = 18,
  disabled = false,
}: {
  ability: string;
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
  disabled?: boolean;
}) => {
  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = parseInt(e.target.value) || 0;
    if (newValue >= min && newValue <= max) {
      onChange(newValue);
    }
  }, [onChange, min, max]);

  const modifier = useMemo(() => {
    return Math.floor((value - 10) / 2);
  }, [value]);

  return (
    <div className="ability-score-input">
      <label>{ABILITY_NAMES[ability as keyof typeof ABILITY_NAMES]}</label>
      <div className="score-controls">
        <button
          type="button"
          onClick={() => onChange(Math.max(min, value - 1))}
          disabled={disabled || value <= min}
          className="score-btn decrease"
        >
          -
        </button>
        <input
          type="number"
          value={value || ''}
          onChange={handleChange}
          min={min}
          max={max}
          disabled={disabled}
          className="score-value"
        />
        <button
          type="button"
          onClick={() => onChange(Math.min(max, value + 1))}
          disabled={disabled || value >= max}
          className="score-btn increase"
        >
          +
        </button>
      </div>
      <div className="modifier">
        {modifier >= 0 ? '+' : ''}{modifier}
      </div>
    </div>
  );
});

AbilityScoreInput.displayName = 'AbilityScoreInput';

export const AbilityScoresStep = memo(({
  characterData,
  onUpdate,
  onMultipleUpdate,
  errors,
  touched,
  setFieldTouched,
}: AbilityScoresStepProps) => {
  const [availableScores, setAvailableScores] = useState<number[]>(STANDARD_ARRAY);
  const [rolledScores, setRolledScores] = useState<number[]>([]);

  const handleMethodChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    const method = e.target.value as CharacterData['abilityScoreMethod'];
    onUpdate('abilityScoreMethod', method);
    
    // Reset scores when changing method
    const resetScores: Record<string, number> = {};
    ABILITIES.forEach(ability => {
      resetScores[ability] = method === 'point_buy' ? 8 : 10;
    });
    onUpdate('abilityScores', resetScores);
    
    if (method === 'standard_array') {
      setAvailableScores([...STANDARD_ARRAY]);
    }
  }, [onUpdate]);

  const handleScoreChange = useCallback((ability: string, value: number) => {
    const newScores = {
      ...characterData.abilityScores,
      [ability]: value,
    };
    onUpdate('abilityScores', newScores);
    setFieldTouched('abilityScores', true);
  }, [characterData.abilityScores, onUpdate, setFieldTouched]);

  const handleStandardArrayAssign = useCallback((ability: string, score: number) => {
    const currentScore = characterData.abilityScores[ability];
    
    // If ability already has a score, return it to available
    if (currentScore && availableScores.indexOf(currentScore) === -1) {
      setAvailableScores(prev => [...prev, currentScore].sort((a, b) => b - a));
    }
    
    // Assign new score
    handleScoreChange(ability, score);
    
    // Remove from available
    setAvailableScores(prev => prev.filter(s => s !== score));
  }, [characterData.abilityScores, availableScores, handleScoreChange]);

  const rollAbilityScores = useCallback(() => {
    const rolls = ABILITIES.map(() => {
      // Roll 4d6, drop lowest
      const dice = Array(4).fill(0).map(() => Math.floor(Math.random() * 6) + 1);
      dice.sort((a, b) => b - a);
      return dice.slice(0, 3).reduce((sum, d) => sum + d, 0);
    });
    
    setRolledScores(rolls);
    
    // Auto-assign rolled scores
    const newScores: Record<string, number> = {};
    ABILITIES.forEach((ability, index) => {
      newScores[ability] = rolls[index];
    });
    onUpdate('abilityScores', newScores);
  }, [onUpdate]);

  // Calculate point buy total
  const pointBuyTotal = useMemo(() => {
    if (characterData.abilityScoreMethod !== 'point_buy') return 0;
    
    return Object.values(characterData.abilityScores).reduce((total, score) => {
      return total + (POINT_BUY_COSTS[score] || 0);
    }, 0);
  }, [characterData.abilityScoreMethod, characterData.abilityScores]);

  const pointsRemaining = POINT_BUY_TOTAL - pointBuyTotal;
  
  const showScoreError = touched.abilityScores && errors.abilityScores;

  return (
    <div className="step-content ability-scores-step">
      <h2>Determine Ability Scores</h2>
      <p className="step-description">
        Your ability scores determine your character's natural talents and weaknesses.
      </p>

      <div className="form-group">
        <label htmlFor="scoreMethod">Score Generation Method</label>
        <select
          id="scoreMethod"
          value={characterData.abilityScoreMethod}
          onChange={handleMethodChange}
        >
          <option value="standard_array">Standard Array (Recommended)</option>
          <option value="point_buy">Point Buy</option>
          <option value="roll">Roll (4d6 drop lowest)</option>
          <option value="manual">Manual Entry</option>
        </select>
      </div>

      {showScoreError && (
        <div className="error-banner">{errors.abilityScores}</div>
      )}

      {characterData.abilityScoreMethod === 'standard_array' && (
        <div className="standard-array-info">
          <p>Assign these scores to your abilities: {availableScores.join(', ')}</p>
        </div>
      )}

      {characterData.abilityScoreMethod === 'point_buy' && (
        <div className="point-buy-info">
          <p>Points Remaining: <strong>{pointsRemaining}</strong> / {POINT_BUY_TOTAL}</p>
          <p className="hint">Each ability starts at 8. Higher scores cost more points.</p>
        </div>
      )}

      {characterData.abilityScoreMethod === 'roll' && (
        <div className="roll-controls">
          <button 
            className="btn-primary roll-btn"
            onClick={rollAbilityScores}
          >
            🎲 Roll Ability Scores
          </button>
          {rolledScores.length > 0 && (
            <p className="rolled-info">Rolled: {rolledScores.join(', ')}</p>
          )}
        </div>
      )}

      <div className="ability-scores-grid">
        {ABILITIES.map(ability => (
          <div key={ability} className="ability-score-container">
            {characterData.abilityScoreMethod === 'standard_array' ? (
              <div className="standard-array-assign">
                <label>{ABILITY_NAMES[ability as keyof typeof ABILITY_NAMES]}</label>
                <select
                  value={characterData.abilityScores[ability] || ''}
                  onChange={(e) => {
                    const score = parseInt(e.target.value);
                    if (score) handleStandardArrayAssign(ability, score);
                  }}
                >
                  <option value="">Select...</option>
                  {availableScores.map(score => (
                    <option key={score} value={score}>{score}</option>
                  ))}
                  {Boolean(characterData.abilityScores[ability]) && (
                    <option value={characterData.abilityScores[ability]}>
                      {characterData.abilityScores[ability]} (Current)
                    </option>
                  )}
                </select>
              </div>
            ) : (
              <AbilityScoreInput
                ability={ability}
                value={characterData.abilityScores[ability] || 10}
                onChange={(value) => handleScoreChange(ability, value)}
                min={characterData.abilityScoreMethod === 'point_buy' ? 8 : 3}
                max={characterData.abilityScoreMethod === 'point_buy' ? 15 : 18}
                disabled={characterData.abilityScoreMethod === 'roll' && rolledScores.length === 0}
              />
            )}
          </div>
        ))}
      </div>

      <div className="ability-descriptions">
        <h3>Ability Score Guide</h3>
        <dl>
          <dt>Strength</dt>
          <dd>Physical power: melee attacks, jumping, carrying capacity</dd>
          
          <dt>Dexterity</dt>
          <dd>Agility: AC, initiative, ranged attacks, stealth</dd>
          
          <dt>Constitution</dt>
          <dd>Endurance: hit points, stamina, resisting poison</dd>
          
          <dt>Intelligence</dt>
          <dd>Reasoning: knowledge skills, wizard spells, investigation</dd>
          
          <dt>Wisdom</dt>
          <dd>Awareness: perception, insight, cleric/druid spells</dd>
          
          <dt>Charisma</dt>
          <dd>Force of personality: social skills, bard/sorcerer spells</dd>
        </dl>
      </div>
    </div>
  );
});

AbilityScoresStep.displayName = 'AbilityScoresStep';