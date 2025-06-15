// Core game type definitions

export type DamageType = 'acid' | 'bludgeoning' | 'cold' | 'fire' | 'force' | 
  'lightning' | 'necrotic' | 'piercing' | 'poison' | 'psychic' | 'radiant' | 
  'slashing' | 'thunder';

export type AbilityScore = 'strength' | 'dexterity' | 'constitution' | 
  'intelligence' | 'wisdom' | 'charisma';

export type Skill = 'acrobatics' | 'animalHandling' | 'arcana' | 'athletics' | 
  'deception' | 'history' | 'insight' | 'intimidation' | 'investigation' | 
  'medicine' | 'nature' | 'perception' | 'performance' | 'persuasion' | 
  'religion' | 'sleightOfHand' | 'stealth' | 'survival';

export type Condition = 'blinded' | 'charmed' | 'deafened' | 'frightened' | 
  'grappled' | 'incapacitated' | 'invisible' | 'paralyzed' | 'petrified' | 
  'poisoned' | 'prone' | 'restrained' | 'stunned' | 'unconscious' | 'exhaustion';

export type ActionType = 'action' | 'bonus_action' | 'reaction' | 'movement';

export interface AbilityScores {
  strength: number;
  dexterity: number;
  constitution: number;
  intelligence: number;
  wisdom: number;
  charisma: number;
}

export interface SkillProficiencies {
  [key: string]: boolean;
}

export interface SavingThrowProficiencies {
  strength: boolean;
  dexterity: boolean;
  constitution: boolean;
  intelligence: boolean;
  wisdom: boolean;
  charisma: boolean;
}

export interface DiceRoll {
  dice: string; // e.g., "2d6+3"
  result: number;
  rolls: number[];
  modifier: number;
  advantage?: boolean;
  disadvantage?: boolean;
  critical?: boolean;
  fumble?: boolean;
}

export interface Attack {
  id: string;
  name: string;
  attackBonus: number;
  damage: string;
  damageType: DamageType;
  range: string;
  properties: string[];
}

export interface SpellSlot {
  level: number;
  total: number;
  used: number;
}

export interface Spell {
  id: string;
  name: string;
  level: number;
  school: string;
  castingTime: string;
  range: string;
  components: string[];
  duration: string;
  description: string;
  damage?: string;
  savingThrow?: AbilityScore;
  attackRoll?: boolean;
}

export interface Equipment {
  id: string;
  name: string;
  type: string;
  quantity: number;
  weight: number;
  equipped: boolean;
  properties?: Record<string, string | number>;
}

export interface Character {
  id: string;
  userId: string;
  name: string;
  race: string;
  subrace?: string;
  class: string;
  subclass?: string;
  level: number;
  experiencePoints: number;
  background: string;
  alignment: string;
  
  // Ability Scores
  abilityScores: AbilityScores;
  
  // Combat Stats
  armorClass: number;
  initiative: number;
  speed: number;
  hitPointsMax: number;
  hitPointsCurrent: number;
  temporaryHitPoints: number;
  hitDice: string;
  hitDiceUsed: number;
  
  // Death Saves
  deathSaves: {
    successes: number;
    failures: number;
  };
  
  // Proficiencies
  proficiencyBonus: number;
  skillProficiencies: SkillProficiencies;
  savingThrowProficiencies: SavingThrowProficiencies;
  
  // Features & Traits
  features: string[];
  traits: string[];
  
  // Spellcasting
  spellcastingAbility?: AbilityScore;
  spellSaveDC?: number;
  spellAttackBonus?: number;
  spellSlots: SpellSlot[];
  spellsKnown: string[];
  spellsPrepared: string[];
  
  // Equipment
  equipment: Equipment[];
  attacks: Attack[];
  
  // Status
  conditions: Condition[];
  exhaustionLevel: number;
  inspiration: boolean;
  
  // Resources
  resources: {
    [key: string]: {
      current: number;
      max: number;
    };
  };
  
  createdAt: string;
  updatedAt: string;
}

export interface CombatParticipant {
  id: string;
  characterId?: string;
  name: string;
  initiative: number;
  initiativeModifier: number;
  armorClass: number;
  hitPointsMax: number;
  hitPointsCurrent: number;
  temporaryHitPoints: number;
  conditions: Condition[];
  isPlayer: boolean;
  isActive: boolean;
  
  // Turn tracking
  hasActed: boolean;
  hasBonusActed: boolean;
  hasReacted: boolean;
  movementUsed: number;
  movementMax: number;
  
  // Concentration
  concentrating: boolean;
  concentratingOn?: string;
}

export interface CombatRound {
  number: number;
  participants: CombatParticipant[];
  currentTurn: number;
  events: CombatEvent[];
}

export interface CombatEvent {
  id: string;
  timestamp: string;
  type: 'attack' | 'damage' | 'heal' | 'condition' | 'movement' | 'spell' | 'other';
  actorId: string;
  targetId?: string;
  description: string;
  details: Record<string, unknown>;
}

export interface GameSession {
  id: string;
  name: string;
  dmId: string;
  playerIds: string[];
  campaignId?: string;
  
  // Combat
  combatActive: boolean;
  currentRound?: CombatRound;
  combatHistory: CombatRound[];
  
  // Session State
  sessionNotes: string;
  sharedResources: Record<string, string>;
  mapData?: MapData;
  
  createdAt: string;
  updatedAt: string;
}

export interface Campaign {
  id: string;
  name: string;
  description: string;
  dmId: string;
  playerIds: string[];
  sessions: string[];
  worldData: WorldData;
  notes: string;
  createdAt: string;
  updatedAt: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  role: 'player' | 'dm' | 'admin';
  characterIds: string[];
  campaignIds: string[];
  preferences: {
    theme: 'light' | 'dark';
    autoRoll: boolean;
    notifications: boolean;
  };
}

export interface MapData {
  imageUrl: string;
  gridSize: number;
  tokens: MapToken[];
}

export interface MapToken {
  id: string;
  characterId?: string;
  x: number;
  y: number;
  size: number;
  color: string;
}

export interface WorldData {
  deities: string[];
  locations: Record<string, WorldLocation>;
  factions: Faction[];
}

export interface WorldLocation {
  id: string;
  name: string;
  description: string;
  population: number;
}

export interface Faction {
  id: string;
  name: string;
  description: string;
  reputation: number;
}
