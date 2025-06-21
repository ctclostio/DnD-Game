import React, { memo, useCallback, useState } from 'react';
import { useDebouncedCallback } from '../../hooks/useDebounce';

interface CustomClassData {
  name: string;
  description: string;
  role?: string;
  playstyle?: string;
}

interface CustomClassFormProps {
  customClassData?: CustomClassData;
  onUpdate: (data: CustomClassData) => void;
  aiEnabled: boolean;
}

const CLASS_ROLES = [
  { value: 'tank', label: 'Tank (Defender)' },
  { value: 'damage', label: 'Damage Dealer' },
  { value: 'healer', label: 'Healer/Support' },
  { value: 'controller', label: 'Controller/Utility' },
  { value: 'hybrid', label: 'Hybrid/Versatile' },
];

const PLAYSTYLES = [
  { value: 'melee', label: 'Melee Combat' },
  { value: 'ranged', label: 'Ranged Combat' },
  { value: 'magic', label: 'Spellcasting' },
  { value: 'stealth', label: 'Stealth & Subterfuge' },
  { value: 'support', label: 'Support & Buffing' },
  { value: 'versatile', label: 'Versatile/Mixed' },
];

export const CustomClassForm = memo(({
  customClassData,
  onUpdate,
  aiEnabled,
}: CustomClassFormProps) => {
  const [isGenerating, setIsGenerating] = useState(false);
  const [generatedPreview, setGeneratedPreview] = useState<any>(null);

  const updateField = useCallback((field: keyof CustomClassData, value: string) => {
    onUpdate({
      ...customClassData,
      name: customClassData?.name || '',
      description: customClassData?.description || '',
      [field]: value,
    });
  }, [customClassData, onUpdate]);

  const handleNameChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    updateField('name', e.target.value);
  }, [updateField]);

  const handleDescriptionChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    updateField('description', e.target.value);
  }, [updateField]);

  const handleRoleChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    updateField('role', e.target.value);
  }, [updateField]);

  const handlePlaystyleChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    updateField('playstyle', e.target.value);
  }, [updateField]);

  const generateCustomClass = useDebouncedCallback(async () => {
    if (!aiEnabled || !customClassData?.name || !customClassData?.description) {
      return;
    }

    setIsGenerating(true);
    try {
      // FUTURE: Implement AI generation endpoint when backend API is ready
      // Expected endpoint: POST /api/ai/generate-class
      console.log('Generating custom class:', customClassData);
      
      // Mock response for demonstration purposes
      setTimeout(() => {
        setGeneratedPreview({
          name: customClassData.name,
          description: 'A unique class with special abilities...',
          hitDice: 'd10',
          primaryAbility: 'Strength or Dexterity',
          savingThrows: ['Strength', 'Constitution'],
          features: [
            { level: 1, name: 'Unique Feature', description: 'A special ability...' },
            { level: 2, name: 'Fighting Style', description: 'Choose a combat specialty...' },
            { level: 3, name: 'Subclass Choice', description: 'Select your path...' },
          ],
        });
        setIsGenerating(false);
      }, 1500);
    } catch (error) {
      console.error('Failed to generate custom class:', error);
      setIsGenerating(false);
    }
  }, 1000);

  return (
    <div className="custom-class-form">
      <h3>Create Your Custom Class</h3>
      {aiEnabled && (
        <p className="ai-notice">
          🤖 Our AI will design a balanced class with unique features and abilities!
        </p>
      )}
      
      <div className="form-group">
        <label htmlFor="customClassName">
          Class Name <span className="required">*</span>
        </label>
        <input
          type="text"
          id="customClassName"
          value={customClassData?.name || ''}
          onChange={handleNameChange}
          placeholder="e.g., Shadow Dancer, Spell Blade, Beast Master"
        />
      </div>
      
      <div className="form-group">
        <label htmlFor="customClassDescription">
          Description <span className="required">*</span>
        </label>
        <textarea
          id="customClassDescription"
          rows={4}
          value={customClassData?.description || ''}
          onChange={handleDescriptionChange}
          placeholder="Describe your class's role, combat style, source of power, and unique features..."
        />
      </div>
      
      <div className="form-row">
        <div className="form-group">
          <label htmlFor="customClassRole">Primary Role</label>
          <select
            id="customClassRole"
            value={customClassData?.role || ''}
            onChange={handleRoleChange}
          >
            <option value="">Select a role...</option>
            {CLASS_ROLES.map(role => (
              <option key={role.value} value={role.value}>
                {role.label}
              </option>
            ))}
          </select>
        </div>
        
        <div className="form-group">
          <label htmlFor="customClassPlaystyle">Preferred Playstyle</label>
          <select
            id="customClassPlaystyle"
            value={customClassData?.playstyle || ''}
            onChange={handlePlaystyleChange}
          >
            <option value="">Select a playstyle...</option>
            {PLAYSTYLES.map(style => (
              <option key={style.value} value={style.value}>
                {style.label}
              </option>
            ))}
          </select>
        </div>
      </div>
      
      {aiEnabled && (
        <button
          className="btn-primary generate-class-btn"
          onClick={generateCustomClass}
          disabled={!customClassData?.name || !customClassData?.description || isGenerating}
        >
          {isGenerating ? '🎲 Generating...' : '🎲 Generate Custom Class'}
        </button>
      )}
      
      {generatedPreview && (
        <div className="class-preview generated-preview">
          <h4>Generated Class Preview</h4>
          <div className="preview-content">
            <h5>{generatedPreview.name}</h5>
            <p>{generatedPreview.description}</p>
            <div className="class-stats">
              <div className="stat">
                <strong>Hit Dice:</strong> {generatedPreview.hitDice}
              </div>
              <div className="stat">
                <strong>Primary Ability:</strong> {generatedPreview.primaryAbility}
              </div>
              <div className="stat">
                <strong>Saving Throws:</strong> {generatedPreview.savingThrows.join(', ')}
              </div>
            </div>
            <div className="class-features">
              <h6>Key Features:</h6>
              {generatedPreview.features.map((feature: any, index: number) => (
                <div key={index} className="feature">
                  <strong>Level {feature.level} - {feature.name}:</strong>
                  <p>{feature.description}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
});

CustomClassForm.displayName = 'CustomClassForm';