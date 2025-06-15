import React, { memo, useCallback, useState } from 'react';
import { useDebouncedCallback } from '../../hooks/useDebounce';
import apiService from '../../services/api';

interface CustomRaceData {
  name: string;
  description: string;
  desiredTraits?: string;
  generationStyle?: 'balanced' | 'flavorful' | 'powerful';
}

interface CustomRaceFormProps {
  customRaceData?: CustomRaceData;
  onUpdate: (data: CustomRaceData) => void;
  aiEnabled: boolean;
}

interface GeneratedRacePreview {
  name: string;
  description: string;
  traits: {
    abilityScoreIncrease: string;
    size: string;
    speed: string;
    specialAbilities: string[];
  };
}

export const CustomRaceForm = memo(({ 
  customRaceData,
  onUpdate,
  aiEnabled,
}: CustomRaceFormProps) => {
  const [isGenerating, setIsGenerating] = useState(false);
  const [generatedPreview, setGeneratedPreview] = useState<GeneratedRacePreview | null>(null);

  const updateField = useCallback((field: keyof CustomRaceData, value: string) => {
    onUpdate({
      ...customRaceData,
      name: customRaceData?.name || '',
      description: customRaceData?.description || '',
      [field]: value,
    });
  }, [customRaceData, onUpdate]);

  const handleNameChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    updateField('name', e.target.value);
  }, [updateField]);

  const handleDescriptionChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    updateField('description', e.target.value);
  }, [updateField]);

  const handleTraitsChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    updateField('desiredTraits', e.target.value);
  }, [updateField]);

  const handleStyleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    updateField('generationStyle', e.target.value as any);
  }, [updateField]);

  const generateCustomRace = useDebouncedCallback(async () => {
    if (!aiEnabled || !customRaceData?.name || !customRaceData?.description) {
      return;
    }

    setIsGenerating(true);
    try {
      const preview = await apiService.generateCustomRace({
        name: customRaceData.name,
        description: customRaceData.description,
        desiredTraits: customRaceData.desiredTraits,
        style: customRaceData.generationStyle,
      });
      setGeneratedPreview(preview as GeneratedRacePreview);
    } catch (error) {
      console.error('Failed to generate custom race:', error);
    } finally {
      setIsGenerating(false);
    }
  }, 1000);

  return (
    <div className="custom-race-form">
      <h3>Create Your Custom Race</h3>
      {aiEnabled && (
        <p className="ai-notice">
          🤖 Our AI will help balance your custom race and generate appropriate traits!
        </p>
      )}
      
      <div className="form-group">
        <label htmlFor="customRaceName">
          Race Name <span className="required">*</span>
        </label>
        <input
          type="text"
          id="customRaceName"
          value={customRaceData?.name || ''}
          onChange={handleNameChange}
          placeholder="e.g., Celestial Tiefling, Frostborn Dwarf"
        />
      </div>
      
      <div className="form-group">
        <label htmlFor="customRaceDescription">
          Description <span className="required">*</span>
        </label>
        <textarea
          id="customRaceDescription"
          rows={4}
          value={customRaceData?.description || ''}
          onChange={handleDescriptionChange}
          placeholder="Describe your race's appearance, culture, origins, and any unique features..."
        />
      </div>
      
      <div className="form-group">
        <label htmlFor="customRaceTraits">Desired Traits (Optional)</label>
        <textarea
          id="customRaceTraits"
          rows={3}
          value={customRaceData?.desiredTraits || ''}
          onChange={handleTraitsChange}
          placeholder="List any specific abilities or traits you'd like (e.g., darkvision, natural armor, elemental resistance)"
        />
      </div>
      
      {aiEnabled && (
        <>
          <div className="ai-generation-options">
            <h4>Generation Style</h4>
            <div className="radio-group">
              <label>
                <input
                  type="radio"
                  name="generationStyle"
                  value="balanced"
                  checked={(customRaceData?.generationStyle || 'balanced') === 'balanced'}
                  onChange={handleStyleChange}
                />
                Balanced (Standard D&D power level)
              </label>
              <label>
                <input
                  type="radio"
                  name="generationStyle"
                  value="flavorful"
                  checked={customRaceData?.generationStyle === 'flavorful'}
                  onChange={handleStyleChange}
                />
                Flavorful (Focus on unique abilities)
              </label>
              <label>
                <input
                  type="radio"
                  name="generationStyle"
                  value="powerful"
                  checked={customRaceData?.generationStyle === 'powerful'}
                  onChange={handleStyleChange}
                />
                Powerful (Slightly stronger, for experienced players)
              </label>
            </div>
          </div>
          
          <button
            className="btn-primary generate-race-btn"
            onClick={generateCustomRace}
            disabled={!customRaceData?.name || !customRaceData?.description || isGenerating}
            aria-disabled={!customRaceData?.name || !customRaceData?.description || isGenerating}
          >
            {isGenerating ? '🎲 Generating...' : '🎲 Generate Custom Race'}
          </button>
        </>
      )}
      
      {generatedPreview && (
        <div className="race-preview generated-preview">
          <h4>Generated Race Preview</h4>
          <div className="preview-content">
            <h5>{generatedPreview.name}</h5>
            <p>{generatedPreview.description}</p>
            <div className="traits-list">
              <div className="trait">
                <strong>Ability Score Increase:</strong> {generatedPreview.traits.abilityScoreIncrease}
              </div>
              <div className="trait">
                <strong>Size:</strong> {generatedPreview.traits.size}
              </div>
              <div className="trait">
                <strong>Speed:</strong> {generatedPreview.traits.speed}
              </div>
              <div className="trait">
                <strong>Special Abilities:</strong>
                <ul>
                  {generatedPreview.traits.specialAbilities.map((ability: string, index: number) => (
                    <li key={index}>{ability}</li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
});

CustomRaceForm.displayName = 'CustomRaceForm';
