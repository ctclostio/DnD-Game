// Redux state type definitions
import { Character, Campaign, GameSession, CombatParticipant, Spell, Equipment } from './game';

// Normalized entity states
export interface EntityState<T> {
  ids: string[];
  entities: {
    [id: string]: T;
  };
}

// UI States
export interface LoadingState {
  [key: string]: boolean;
}

export interface ErrorState {
  [key: string]: string | null;
}

// Auth State
export interface AuthState {
  user: {
    id: string;
    username: string;
    email: string;
    role: 'player' | 'dm' | 'admin';
  } | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;
}

// Character State
export interface CharacterState {
  characters: EntityState<Character>;
  currentCharacterId: string | null;
  isLoading: LoadingState;
  errors: ErrorState;
}

// Campaign State
export interface CampaignState {
  campaigns: EntityState<Campaign>;
  currentCampaignId: string | null;
  isLoading: LoadingState;
  errors: ErrorState;
}

// Game Session State
export interface GameSessionState {
  sessions: EntityState<GameSession>;
  currentSessionId: string | null;
  isConnected: boolean;
  connectionError: string | null;
  isLoading: LoadingState;
  errors: ErrorState;
}

// Combat State
export interface CombatState {
  active: boolean;
  sessionId: string | null;
  round: number;
  turn: number;
  participants: EntityState<CombatParticipant>;
  initiativeOrder: string[];
  currentParticipantId: string | null;
  
  // Combat history for undo/redo
  history: CombatHistoryEntry[];
  historyIndex: number;
  
  // Temporary states during actions
  pendingAction: CombatAction | null;
  
  isLoading: LoadingState;
  errors: ErrorState;
}

export interface CombatHistoryEntry {
  timestamp: number;
  action: CombatAction;
  previousState: Partial<CombatState>;
  description: string;
}

// Game Data State (spells, items, rules, etc.)
export interface GameDataState {
  spells: EntityState<Spell>;
  equipment: EntityState<Equipment>;
  classes: EntityState<DndClass>;
  races: EntityState<DndRace>;
  isLoaded: boolean;
  isLoading: boolean;
  error: string | null;
}

// WebSocket State
export interface WebSocketState {
  connected: boolean;
  reconnecting: boolean;
  error: string | null;
  rooms: {
    [roomId: string]: {
      connected: boolean;
      participants: string[];
    };
  };
}

// UI State
export interface UIState {
  theme: 'light' | 'dark';
  sidebarOpen: boolean;
  modals: {
    [key: string]: {
      isOpen: boolean;
      data?: Record<string, unknown>;
    };
  };
  notifications: Notification[];
  shortcuts: {
    enabled: boolean;
    customBindings: { [key: string]: string };
  };
}

export interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  message: string;
  duration?: number;
  timestamp: number;
}

// DM Tools State
export interface DMToolsState {
  // Undo/redo functionality
  canUndo: boolean;
  canRedo: boolean;
  undoStack: UndoableAction[];
  redoStack: UndoableAction[];
  
  // Quick references
  quickReferences: {
    conditions: Condition[];
    rules: Rule[];
  };
  
  // Notes
  sessionNotes: string;
  campaignNotes: string;
  
  isLoading: LoadingState;
  errors: ErrorState;
}

export interface UndoableAction {
  id: string;
  type: string;
  timestamp: number;
  description: string;
  undo: () => void;
  redo: () => void;
}

// Root State
export interface RootState {
  auth: AuthState;
  character: CharacterState;
  gameSession: GameSessionState;
  combat: CombatState;
  ui: UIState;
  dmTools: DMToolsState;
  websocket: WebSocketState;
}

// Action payload types
export interface WebSocketMessage {
  type: string;
  roomId: string;
  data: Record<string, unknown>;
  timestamp: number;
}

export interface CombatAction {
  type: 'ATTACK' | 'DAMAGE' | 'HEAL' | 'CONDITION_ADD' | 'CONDITION_REMOVE' | 
        'MOVE' | 'END_TURN' | 'INITIATIVE_ROLL';
  actorId: string;
  targetId?: string;
  data: {
    distance?: number;
    damage?: number;
    heal?: number;
    condition?: Condition;
    spellId?: string;
  };
}

export interface DndClass {
  id: string;
  name: string;
  description: string;
  hitDie: number;
  proficiencies: string[];
  features: string[];
}

export interface DndRace {
  id: string;
  name: string;
  description: string;
  speed: number;
  abilityScoreBonuses: Record<string, number>;
  traits: string[];
}

export interface Condition {
  id: string;
  name: string;
  description: string;
}

export interface Rule {
  id: string;
  name: string;
  description: string;
}
