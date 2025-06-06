import React, { memo, useCallback, useMemo } from 'react';
import { CharacterData, CharacterOptions } from './CharacterBuilder';

interface SkillSelectionStepProps {
  characterData: CharacterData;
  onUpdate: (field: keyof CharacterData, value: any) => void;
  options: CharacterOptions;
  errors: Record<string, string>;
  touched: Record<string, boolean>;
  setFieldTouched: (field: keyof CharacterData, touched: boolean) => void;
}

// Skills grouped by ability
const SKILL_GROUPS: Record<string, string[]> = {
  Strength: ['Athletics'],
  Dexterity: ['Acrobatics', 'Sleight of Hand', 'Stealth'],
  Intelligence: ['Arcana', 'History', 'Investigation', 'Nature', 'Religion'],
  Wisdom: ['Animal Handling', 'Insight', 'Medicine', 'Perception', 'Survival'],
  Charisma: ['Deception', 'Intimidation', 'Performance', 'Persuasion'],
};

// Number of skills per class
const CLASS_SKILL_COUNTS: Record<string, number> = {
  barbarian: 2,
  bard: 3,
  cleric: 2,
  druid: 2,
  fighter: 2,
  monk: 2,
  paladin: 2,
  ranger: 3,
  rogue: 4,
  sorcerer: 2,
  warlock: 2,
  wizard: 2,
};

// Available skills per class
const CLASS_SKILLS: Record<string, string[]> = {
  barbarian: ['Animal Handling', 'Athletics', 'Intimidation', 'Nature', 'Perception', 'Survival'],
  bard: ['Acrobatics', 'Animal Handling', 'Arcana', 'Athletics', 'Deception', 'History', 
         'Insight', 'Intimidation', 'Investigation', 'Medicine', 'Nature', 'Perception', 
         'Performance', 'Persuasion', 'Religion', 'Sleight of Hand', 'Stealth', 'Survival'],
  cleric: ['History', 'Insight', 'Medicine', 'Persuasion', 'Religion'],
  druid: ['Arcana', 'Animal Handling', 'Insight', 'Medicine', 'Nature', 'Perception', 'Religion', 'Survival'],
  fighter: ['Acrobatics', 'Animal Handling', 'Athletics', 'History', 'Insight', 'Intimidation', 'Perception', 'Survival'],
  monk: ['Acrobatics', 'Athletics', 'History', 'Insight', 'Religion', 'Stealth'],
  paladin: ['Athletics', 'Insight', 'Intimidation', 'Medicine', 'Persuasion', 'Religion'],
  ranger: ['Animal Handling', 'Athletics', 'Insight', 'Investigation', 'Nature', 'Perception', 'Stealth', 'Survival'],
  rogue: ['Acrobatics', 'Athletics', 'Deception', 'Insight', 'Intimidation', 'Investigation', 
          'Perception', 'Performance', 'Persuasion', 'Sleight of Hand', 'Stealth'],
  sorcerer: ['Arcana', 'Deception', 'Insight', 'Intimidation', 'Persuasion', 'Religion'],
  warlock: ['Arcana', 'Deception', 'History', 'Intimidation', 'Investigation', 'Nature', 'Religion'],
  wizard: ['Arcana', 'History', 'Insight', 'Investigation', 'Medicine', 'Religion'],
};

const SkillCheckbox = memo(({
  skill,
  isSelected,
  isDisabled,
  onChange,
  modifier,
}: {
  skill: string;
  isSelected: boolean;
  isDisabled: boolean;
  onChange: (skill: string, checked: boolean) => void;
  modifier: number;
}) => (
  <label className={`skill-checkbox ${isDisabled ? 'disabled' : ''}`}>
    <input
      type="checkbox"
      checked={isSelected}
      disabled={isDisabled}
      onChange={(e) => onChange(skill, e.target.checked)}
    />
    <span className="skill-name">{skill}</span>
    <span className="skill-modifier">
      {modifier >= 0 ? '+' : ''}{modifier}
    </span>
  </label>
));

SkillCheckbox.displayName = 'SkillCheckbox';

export const SkillSelectionStep = memo(({
  characterData,
  onUpdate,
  options,
  errors,
  touched,
  setFieldTouched,
}: SkillSelectionStepProps) => {
  const selectedClass = characterData.class === 'custom' ? 'fighter' : characterData.class; // Default to fighter for custom
  const maxSkills = CLASS_SKILL_COUNTS[selectedClass] || 2;
  const availableSkills = CLASS_SKILLS[selectedClass] || [];

  const handleSkillToggle = useCallback((skill: string, checked: boolean) => {
    const currentSkills = [...(characterData.selectedSkills || [])];
    
    if (checked) {
      if (currentSkills.length < maxSkills) {
        currentSkills.push(skill);
      }
    } else {
      const index = currentSkills.indexOf(skill);
      if (index > -1) {
        currentSkills.splice(index, 1);
      }
    }
    
    onUpdate('selectedSkills', currentSkills);
    setFieldTouched('selectedSkills', true);
  }, [characterData.selectedSkills, maxSkills, onUpdate, setFieldTouched]);

  // Calculate skill modifiers based on ability scores
  const getSkillModifier = useCallback((skill: string): number => {
    const ability = Object.entries(SKILL_GROUPS).find(([_, skills]) => 
      skills.includes(skill)
    )?.[0];
    
    if (!ability) return 0;
    
    const abilityScore = characterData.abilityScores[ability.slice(0, 3).toUpperCase()] || 10;
    return Math.floor((abilityScore - 10) / 2);
  }, [characterData.abilityScores]);

  const remainingSkills = maxSkills - (characterData.selectedSkills?.length || 0);
  
  const showSkillError = touched.selectedSkills && errors.selectedSkills;

  // Group available skills by ability
  const groupedSkills = useMemo(() => {
    const groups: Record<string, string[]> = {};
    
    Object.entries(SKILL_GROUPS).forEach(([ability, skills]) => {
      const availableInGroup = skills.filter(skill => availableSkills.includes(skill));
      if (availableInGroup.length > 0) {
        groups[ability] = availableInGroup;
      }
    });
    
    return groups;
  }, [availableSkills]);

  return (
    <div className="step-content skill-selection-step">
      <h2>Choose Your Skills</h2>
      <p className="step-description">
        As a {selectedClass}, you can choose <strong>{maxSkills}</strong> skills from the list below.
        Skills represent your character's training and expertise.
      </p>

      <div className="skill-counter">
        <span className={remainingSkills === 0 ? 'complete' : ''}>
          Skills Selected: {characterData.selectedSkills?.length || 0} / {maxSkills}
        </span>
        {remainingSkills > 0 && (
          <span className="remaining">({remainingSkills} remaining)</span>
        )}
      </div>

      {showSkillError && (
        <div className="error-banner">{errors.selectedSkills}</div>
      )}

      <div className="skills-container">
        {Object.entries(groupedSkills).map(([ability, skills]) => (
          <div key={ability} className="skill-group">
            <h3>{ability}</h3>
            <div className="skill-list">
              {skills.map(skill => (
                <SkillCheckbox
                  key={skill}
                  skill={skill}
                  isSelected={characterData.selectedSkills?.includes(skill) || false}
                  isDisabled={
                    !characterData.selectedSkills?.includes(skill) && 
                    characterData.selectedSkills?.length >= maxSkills
                  }
                  onChange={handleSkillToggle}
                  modifier={getSkillModifier(skill)}
                />
              ))}
            </div>
          </div>
        ))}
      </div>

      <div className="skill-tips">
        <h3>Skill Selection Tips</h3>
        <ul>
          <li>Choose skills that complement your class and playstyle</li>
          <li>Consider your party composition - avoid too much overlap</li>
          <li>Skills with higher ability modifiers will be more effective</li>
          <li>Some skills like Perception and Investigation are universally useful</li>
        </ul>
      </div>
    </div>
  );
});

SkillSelectionStep.displayName = 'SkillSelectionStep';