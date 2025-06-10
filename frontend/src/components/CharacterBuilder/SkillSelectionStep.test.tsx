import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { SkillSelectionStep } from './SkillSelectionStep';

const baseCharacterData = {
  class: 'fighter',
  abilityScores: {
    STR: 16,
    DEX: 12,
    CON: 14,
    INT: 10,
    WIS: 10,
    CHA: 8,
  },
  selectedSkills: [],
};

const baseOptions = {};

describe('SkillSelectionStep', () => {
  it('renders skill groups and checkboxes', () => {
    render(
      <SkillSelectionStep
        characterData={baseCharacterData}
        onUpdate={jest.fn()}
        options={baseOptions}
        errors={{}}
        touched={{}}
        setFieldTouched={jest.fn()}
      />
    );
    expect(screen.getByText('Choose Your Skills')).toBeInTheDocument();
    expect(screen.getByText('Strength')).toBeInTheDocument();
    expect(screen.getByText('Athletics')).toBeInTheDocument();
    expect(screen.getByLabelText('Athletics')).toBeInTheDocument();
  });

  it('calls onUpdate and setFieldTouched when a skill is selected', () => {
    const onUpdate = jest.fn();
    const setFieldTouched = jest.fn();
    render(
      <SkillSelectionStep
        characterData={{ ...baseCharacterData, selectedSkills: [] }}
        onUpdate={onUpdate}
        options={baseOptions}
        errors={{}}
        touched={{}}
        setFieldTouched={setFieldTouched}
      />
    );
    const checkbox = screen.getByLabelText('Athletics') as HTMLInputElement;
    fireEvent.click(checkbox);
    expect(onUpdate).toHaveBeenCalledWith('selectedSkills', ['Athletics']);
    expect(setFieldTouched).toHaveBeenCalledWith('selectedSkills', true);
  });

  it('disables checkboxes when max skills are selected', () => {
    const maxSkills = 2;
    const selectedSkills = ['Athletics', 'Animal Handling'];
    render(
      <SkillSelectionStep
        characterData={{ ...baseCharacterData, selectedSkills }}
        onUpdate={jest.fn()}
        options={baseOptions}
        errors={{}}
        touched={{}}
        setFieldTouched={jest.fn()}
      />
    );
    // Should be checked and enabled
    expect((screen.getByLabelText('Athletics') as HTMLInputElement).checked).toBe(true);
    expect((screen.getByLabelText('Athletics') as HTMLInputElement).disabled).toBe(false);
    // Should be checked and enabled
    expect((screen.getByLabelText('Animal Handling') as HTMLInputElement).checked).toBe(true);
    expect((screen.getByLabelText('Animal Handling') as HTMLInputElement).disabled).toBe(false);
    // Should be disabled if not selected
    expect((screen.getByLabelText('Intimidation') as HTMLInputElement).disabled).toBe(true);
  });

  it('shows error banner if there is a skill error and touched', () => {
    render(
      <SkillSelectionStep
        characterData={baseCharacterData}
        onUpdate={jest.fn()}
        options={baseOptions}
        errors={{ selectedSkills: 'You must select 2 skills.' }}
        touched={{ selectedSkills: true }}
        setFieldTouched={jest.fn()}
      />
    );
    expect(screen.getByText('You must select 2 skills.')).toBeInTheDocument();
  });
});
