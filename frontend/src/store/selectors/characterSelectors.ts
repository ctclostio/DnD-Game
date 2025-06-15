import { createSelector } from 'reselect';
import { RootState } from '../index';
import { Character, AbilityScore, SpellSlot } from '../../types/game';

// Base selectors
const selectCharacterState = (state: RootState) => state.character;
const selectCharacters = (state: RootState) => state.character.characters;
const selectCurrentCharacterId = (state: RootState) => state.character.currentCharacterId;

// Get all characters as array
export const selectCharactersArray = createSelector(
  [selectCharacters],
  (characters) => characters.ids.map((id: string) => characters.entities[id])
);

// Get current character
export const selectCurrentCharacter = createSelector(
  [selectCharacters, selectCurrentCharacterId],
  (characters, currentId) => currentId ? characters.entities[currentId] : null
);

// Get character by ID
export const selectCharacterById = (characterId: string) =>
  createSelector(
    [selectCharacters],
    (characters) => characters.entities[characterId]
  );

// Calculate ability modifiers
export const selectCharacterModifiers = createSelector(
  [selectCurrentCharacter],
  (character) => {
    if (!character) return null;
    
    const modifiers = {} as Record<AbilityScore, number>;
    
    Object.entries(character.abilityScores).forEach(([ability, score]: [string, number]) => {
      modifiers[ability as AbilityScore] = Math.floor((score - 10) / 2);
    });
    
    return modifiers;
  }
);

// Get spell slots summary
export const selectSpellSlotsSummary = createSelector(
  [selectCurrentCharacter],
  (character) => {
    if (!character || !character.spellSlots) return null;
    
    return character.spellSlots.reduce((summary, slot: SpellSlot) => {
      summary[`level${slot.level}`] = {
        total: slot.total,
        used: slot.used,
        remaining: slot.total - slot.used,
      };
      return summary;
    }, {} as Record<string, { total: number; used: number; remaining: number }>);
  }
);

// Get character combat stats
export const selectCharacterCombatStats = createSelector(
  [selectCurrentCharacter, selectCharacterModifiers],
  (character, modifiers) => {
    if (!character || !modifiers) return null;
    
    return {
      initiative: modifiers.dexterity,
      hitPoints: {
        current: character.hitPointsCurrent,
        max: character.hitPointsMax,
        temp: character.temporaryHitPoints,
        percentage: (character.hitPointsCurrent / character.hitPointsMax) * 100,
      },
      armorClass: character.armorClass,
      speed: character.speed,
      proficiencyBonus: character.proficiencyBonus,
      spellSaveDC: character.spellSaveDC,
      spellAttackBonus: character.spellAttackBonus,
    };
  }
);

// Get character skill bonuses
export const selectCharacterSkillBonuses = createSelector(
  [selectCurrentCharacter, selectCharacterModifiers],
  (character, modifiers) => {
    if (!character || !modifiers) return null;
    
    const skillAbilityMap: Record<string, AbilityScore> = {
      acrobatics: 'dexterity',
      animalHandling: 'wisdom',
      arcana: 'intelligence',
      athletics: 'strength',
      deception: 'charisma',
      history: 'intelligence',
      insight: 'wisdom',
      intimidation: 'charisma',
      investigation: 'intelligence',
      medicine: 'wisdom',
      nature: 'intelligence',
      perception: 'wisdom',
      performance: 'charisma',
      persuasion: 'charisma',
      religion: 'intelligence',
      sleightOfHand: 'dexterity',
      stealth: 'dexterity',
      survival: 'wisdom',
    };
    
    const skillBonuses: Record<string, number> = {};
    
    Object.entries(skillAbilityMap).forEach(([skill, ability]) => {
      const abilityModifier = modifiers[ability];
      const isProficient = character.skillProficiencies[skill];
      skillBonuses[skill] = abilityModifier + (isProficient ? character.proficiencyBonus : 0);
    });
    
    return skillBonuses;
  }
);

// Get saving throw bonuses
export const selectCharacterSavingThrows = createSelector(
  [selectCurrentCharacter, selectCharacterModifiers],
  (character, modifiers) => {
    if (!character || !modifiers) return null;
    
    const savingThrows = {} as Record<AbilityScore, number>;
    
    Object.entries(modifiers).forEach(([ability, modifier]) => {
      const isProficient = character.savingThrowProficiencies[ability as AbilityScore];
      savingThrows[ability as AbilityScore] = modifier + (isProficient ? character.proficiencyBonus : 0);
    });
    
    return savingThrows;
  }
);

// Get prepared spells count
export const selectPreparedSpellsInfo = createSelector(
  [selectCurrentCharacter, selectCharacterModifiers],
  (character, modifiers) => {
    if (!character || !modifiers || !character.spellcastingAbility) return null;
    
    const spellcastingModifier = modifiers[character.spellcastingAbility];
    const maxPrepared = Math.max(1, character.level + spellcastingModifier);
    
    return {
      prepared: character.spellsPrepared.length,
      max: maxPrepared,
      remaining: maxPrepared - character.spellsPrepared.length,
    };
  }
);

// Check if character is alive
export const selectIsCharacterAlive = createSelector(
  [selectCurrentCharacter],
  (character) => {
    if (!character) return false;
    return character.hitPointsCurrent > 0 || character.deathSaves.successes >= 3;
  }
);

// Get characters by level range
export const selectCharactersByLevelRange = (minLevel: number, maxLevel: number) =>
  createSelector(
    [selectCharactersArray],
    (characters) => characters.filter((char: Character) => 
      char.level >= minLevel && char.level <= maxLevel
    )
  );
