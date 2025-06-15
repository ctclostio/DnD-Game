import React, { memo, useCallback } from 'react';
import { CharacterData } from './CharacterBuilder';

interface BasicInfoStepProps {
  characterData: CharacterData;
  onUpdate: (field: keyof CharacterData, value: CharacterData[keyof CharacterData]) => void;
  errors: Record<string, string>;
  touched: Record<string, boolean>;
  setFieldTouched: (field: keyof CharacterData, touched: boolean) => void;
}

const ALIGNMENTS = [
  'Lawful Good',
  'Neutral Good',
  'Chaotic Good',
  'Lawful Neutral',
  'True Neutral',
  'Chaotic Neutral',
  'Lawful Evil',
  'Neutral Evil',
  'Chaotic Evil',
];

export const BasicInfoStep = memo(({
  characterData,
  onUpdate,
  errors,
  touched,
  setFieldTouched,
}: BasicInfoStepProps) => {
  const handleNameChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    onUpdate('name', e.target.value);
  }, [onUpdate]);

  const handleNameBlur = useCallback(() => {
    setFieldTouched('name', true);
  }, [setFieldTouched]);

  const handleAlignmentChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    onUpdate('alignment', e.target.value);
  }, [onUpdate]);

  const showNameError = touched.name && errors.name;

  return (
    <div className="step-content basic-info-step">
      <h2>Basic Information</h2>
      <p className="step-description">
        Let's start with the basics. What shall we call your character?
      </p>
      
      <div className="form-group">
        <label htmlFor="characterName">
          Character Name <span className="required">*</span>
        </label>
        <input
          type="text"
          id="characterName"
          value={characterData.name}
          onChange={handleNameChange}
          onBlur={handleNameBlur}
          placeholder="Enter character name"
          className={showNameError ? 'error' : ''}
          aria-invalid={showNameError ? 'true' : 'false'}
          aria-describedby={showNameError ? 'characterName-error' : undefined}
          autoFocus
        />
        {showNameError && (
          <span id="characterName-error" className="error-message" aria-live="polite">{errors.name}</span>
        )}
        <span className="field-hint">
          Choose a name that fits your character's background and personality
        </span>
      </div>
      
      <div className="form-group">
        <label htmlFor="alignment">Alignment</label>
        <select
          id="alignment"
          value={characterData.alignment}
          onChange={handleAlignmentChange}
        >
          {ALIGNMENTS.map(alignment => (
            <option key={alignment} value={alignment}>
              {alignment}
            </option>
          ))}
        </select>
        <span className="field-hint">
          Your character's moral and ethical outlook
        </span>
      </div>
      
      <div className="alignment-guide">
        <h3>Alignment Guide</h3>
        <div className="alignment-grid">
          <div className="alignment-cell">
            <strong>Lawful Good:</strong> Honor and compassion
          </div>
          <div className="alignment-cell">
            <strong>Neutral Good:</strong> Doing the right thing
          </div>
          <div className="alignment-cell">
            <strong>Chaotic Good:</strong> Freedom and kindness
          </div>
          <div className="alignment-cell">
            <strong>Lawful Neutral:</strong> Order and tradition
          </div>
          <div className="alignment-cell">
            <strong>True Neutral:</strong> Balance in all things
          </div>
          <div className="alignment-cell">
            <strong>Chaotic Neutral:</strong> Personal freedom
          </div>
          <div className="alignment-cell">
            <strong>Lawful Evil:</strong> Tyranny and domination
          </div>
          <div className="alignment-cell">
            <strong>Neutral Evil:</strong> Pure selfishness
          </div>
          <div className="alignment-cell">
            <strong>Chaotic Evil:</strong> Destruction and chaos
          </div>
        </div>
      </div>
    </div>
  );
});

BasicInfoStep.displayName = 'BasicInfoStep';
