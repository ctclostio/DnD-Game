import React, { useState, useCallback, useMemo, memo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useOptimizedForm } from '../../hooks/useOptimizedState';
import { BasicInfoStep } from './BasicInfoStep';
import { RaceSelectionStep } from './RaceSelectionStep';
import { ClassSelectionStep } from './ClassSelectionStep';
import { AbilityScoresStep } from './AbilityScoresStep';
import { BackgroundSelectionStep } from './BackgroundSelectionStep';
import { SkillSelectionStep } from './SkillSelectionStep';
import { ReviewStep } from './ReviewStep';
import { ApiService } from '../../services/api';

const apiService = new ApiService();

export interface CharacterData {
  name: string;
  race: string;
  subrace?: string;
  class: string;
  background: string;
  alignment: string;
  abilityScoreMethod: 'standard_array' | 'point_buy' | 'manual' | 'roll';
  abilityScores: Record<string, number>;
  selectedSkills: string[];
  customRaceData?: {
    name: string;
    description: string;
    desiredTraits?: string;
    generationStyle?: 'balanced' | 'flavorful' | 'powerful';
  };
  customClassData?: {
    name: string;
    description: string;
    role?: string;
    playstyle?: string;
  };
}

interface CharacterOptions {
  races: string[];
  classes: string[];
  backgrounds: string[];
  skills: string[];
  aiEnabled: boolean;
}

const INITIAL_CHARACTER_DATA: CharacterData = {
  name: '',
  race: '',
  class: '',
  background: '',
  alignment: 'True Neutral',
  abilityScoreMethod: 'standard_array',
  abilityScores: {},
  selectedSkills: [],
};

const STEP_TITLES = [
  'Basic Information',
  'Choose Your Race',
  'Choose Your Class',
  'Ability Scores',
  'Background',
  'Skills',
  'Review & Create',
];

export const CharacterBuilder = memo(() => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(0);
  const [options, setOptions] = useState<CharacterOptions | null>(null);
  const [isCustomMode, setIsCustomMode] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const form = useOptimizedForm<CharacterData>(INITIAL_CHARACTER_DATA, {
    validateOnChange: true,
    validator: useCallback((values: CharacterData) => {
      const errors: Record<string, string> = {};
      
      if (currentStep === 0 && !values.name.trim()) {
        errors.name = 'Character name is required';
      }
      
      if (currentStep === 1 && !values.race) {
        errors.race = 'Please select a race';
      }
      
      if (currentStep === 2 && !values.class) {
        errors.class = 'Please select a class';
      }
      
      if (currentStep === 3 && Object.keys(values.abilityScores).length < 6) {
        errors.abilityScores = 'Please assign all ability scores';
      }
      
      return errors;
    }, [currentStep]),
  });

  // Load options on mount
  React.useEffect(() => {
    const loadOptions = async () => {
      try {
        setLoading(true);
        const data = await apiService.get('/characters/options');
        setOptions(data);
      } catch (err) {
        setError('Failed to load character options');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    loadOptions();
  }, []);

  // Navigation handlers
  const handleNext = useCallback(async () => {
    if (currentStep === STEP_TITLES.length - 1) {
      // Create character
      try {
        setLoading(true);
        const character = await apiService.post('/characters', form.values);
        navigate(`/characters/${character.id}`);
      } catch (err) {
        setError('Failed to create character');
        console.error(err);
      } finally {
        setLoading(false);
      }
    } else {
      setCurrentStep(prev => prev + 1);
    }
  }, [currentStep, form.values, navigate]);

  const handlePrevious = useCallback(() => {
    setCurrentStep(prev => Math.max(0, prev - 1));
  }, []);

  const toggleCustomMode = useCallback(() => {
    setIsCustomMode(prev => !prev);
  }, []);

  // Progress calculation
  const progress = useMemo(() => {
    return ((currentStep + 1) / STEP_TITLES.length) * 100;
  }, [currentStep]);

  // Can proceed check
  const canProceed = useMemo(() => {
    const stepErrors = Object.keys(form.errors);
    return stepErrors.length === 0;
  }, [form.errors]);

  // Render current step
  const renderStep = useMemo(() => {
    if (!options) return null;

    const stepProps = {
      characterData: form.values,
      onUpdate: form.setFieldValue,
      onMultipleUpdate: form.setMultipleValues,
      options,
      isCustomMode,
      errors: form.errors,
      touched: form.touched,
      setFieldTouched: form.setFieldTouched,
    };

    switch (currentStep) {
      case 0:
        return <BasicInfoStep {...stepProps} />;
      case 1:
        return <RaceSelectionStep {...stepProps} />;
      case 2:
        return <ClassSelectionStep {...stepProps} />;
      case 3:
        return <AbilityScoresStep {...stepProps} />;
      case 4:
        return <BackgroundSelectionStep {...stepProps} />;
      case 5:
        return <SkillSelectionStep {...stepProps} />;
      case 6:
        return <ReviewStep {...stepProps} />;
      default:
        return null;
    }
  }, [currentStep, form, options, isCustomMode]);

  if (loading && !options) {
    return (
      <div className="character-builder-loading">
        <div className="spinner" />
        <p>Loading character options...</p>
      </div>
    );
  }

  if (error && !options) {
    return (
      <div className="character-builder-error">
        <p>{error}</p>
        <button onClick={() => window.location.reload()}>Retry</button>
      </div>
    );
  }

  return (
    <div className="character-builder">
      <div className="builder-header">
        <h1>Create Your Character</h1>
        <div className="mode-toggle">
          <label className="toggle">
            <input 
              type="checkbox" 
              checked={isCustomMode}
              onChange={toggleCustomMode}
            />
            <span className="slider" />
          </label>
          <span>
            Custom/Homebrew Mode {options?.aiEnabled ? '(AI Powered)' : '(Basic)'}
          </span>
        </div>
      </div>
      
      <div className="progress-bar">
        <div className="progress-fill" style={{ width: `${progress}%` }} />
        <div className="progress-steps">
          {STEP_TITLES.map((title, index) => (
            <div 
              key={title}
              className={`progress-step ${index === currentStep ? 'active' : ''} ${index < currentStep ? 'completed' : ''}`}
            >
              <span className="step-number">{index + 1}</span>
              <span className="step-title">{title}</span>
            </div>
          ))}
        </div>
      </div>
      
      <div className="builder-content">
        {renderStep}
      </div>
      
      <div className="builder-navigation">
        <button 
          className="btn-secondary" 
          onClick={handlePrevious}
          disabled={currentStep === 0}
        >
          Previous
        </button>
        <button 
          className="btn-primary" 
          onClick={handleNext}
          disabled={!canProceed || loading}
        >
          {loading ? 'Loading...' : currentStep === STEP_TITLES.length - 1 ? 'Create Character' : 'Next'}
        </button>
      </div>
      
      {error && (
        <div className="builder-error-message">
          {error}
        </div>
      )}
    </div>
  );
});

CharacterBuilder.displayName = 'CharacterBuilder';