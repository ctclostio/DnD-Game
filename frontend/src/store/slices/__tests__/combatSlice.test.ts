import { configureStore } from '@reduxjs/toolkit';
import combatReducer, {
  startCombat,
  executeCombatAction,
  addParticipant,
  removeParticipant,
  updateParticipant,
  applyDamage,
  applyHealing,
  nextTurn,
  endCombat,
  restoreState,
} from '../combatSlice';
import dmToolsReducer from '../dmToolsSlice';
import { CombatParticipant, CombatAction } from '../../../types/game';

// Mock the addUndoableAction to avoid circular dependencies
jest.mock('../dmToolsSlice', () => ({
  ...jest.requireActual('../dmToolsSlice'),
  addUndoableAction: jest.fn(() => ({ type: 'dmTools/addUndoableAction', payload: {} })),
}));

describe('combatSlice', () => {
  let store: ReturnType<typeof configureStore>;

  const createMockParticipant = (id: string, name: string, initiative = 10): CombatParticipant => ({
    id,
    name,
    type: 'character',
    initiative,
    initiativeModifier: 2,
    hitPointsCurrent: 20,
    hitPointsMax: 20,
    temporaryHitPoints: 0,
    armorClass: 15,
    movementMax: 30,
    movementUsed: 0,
    hasActed: false,
    hasBonusActed: false,
    hasReacted: false,
    conditions: [],
    isNPC: false,
  });

  beforeEach(() => {
    jest.clearAllMocks();
    store = configureStore({
      reducer: {
        combat: combatReducer,
        dmTools: dmToolsReducer,
      },
    });
  });

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const state = store.getState().combat;
      
      expect(state).toEqual({
        active: false,
        sessionId: null,
        round: 0,
        turn: 0,
        participants: {
          ids: [],
          entities: {},
        },
        initiativeOrder: [],
        currentParticipantId: null,
        history: [],
        historyIndex: -1,
        pendingAction: null,
        isLoading: {},
        errors: {},
      });
    });
  });

  describe('startCombat', () => {
    it('should start combat with participants', async () => {
      const participants = [
        createMockParticipant('char-1', 'Aragorn'),
        createMockParticipant('char-2', 'Legolas'),
        createMockParticipant('char-3', 'Gimli'),
      ];

      await store.dispatch(startCombat({
        sessionId: 'session-123',
        participants,
      }));

      const state = store.getState().combat;
      
      expect(state.active).toBe(true);
      expect(state.sessionId).toBe('session-123');
      expect(state.round).toBe(1);
      expect(state.participants.ids).toHaveLength(3);
      expect(state.initiativeOrder).toHaveLength(3);
      expect(state.currentParticipantId).toBe(state.initiativeOrder[0]);
      
      // Initiatives should be rolled
      Object.values(state.participants.entities).forEach(p => {
        expect(p.initiative).toBeGreaterThanOrEqual(1 + p.initiativeModifier);
        expect(p.initiative).toBeLessThanOrEqual(20 + p.initiativeModifier);
      });
      
      // Should be sorted by initiative
      for (let i = 0; i < state.initiativeOrder.length - 1; i++) {
        const current = state.participants.entities[state.initiativeOrder[i]];
        const next = state.participants.entities[state.initiativeOrder[i + 1]];
        expect(current.initiative).toBeGreaterThanOrEqual(next.initiative);
      }
    });

    it('should handle pending state', () => {
      const participants = [createMockParticipant('char-1', 'Aragorn')];
      
      store.dispatch(startCombat({
        sessionId: 'session-123',
        participants,
      }));

      const state = store.getState().combat;
      expect(state.isLoading.startCombat).toBe(true);
    });
  });

  describe('addParticipant', () => {
    beforeEach(async () => {
      // Start combat first
      await store.dispatch(startCombat({
        sessionId: 'session-123',
        participants: [
          createMockParticipant('char-1', 'Aragorn', 15),
          createMockParticipant('char-2', 'Legolas', 18),
        ],
      }));
    });

    it('should add a participant and re-sort initiative', () => {
      const newParticipant = createMockParticipant('char-3', 'Gimli', 20);
      
      store.dispatch(addParticipant(newParticipant));

      const state = store.getState().combat;
      expect(state.participants.ids).toContain('char-3');
      expect(state.participants.entities['char-3']).toEqual(newParticipant);
      
      // Initiative order should be re-sorted
      expect(state.initiativeOrder[0]).toBe('char-3'); // Highest initiative
    });
  });

  describe('removeParticipant', () => {
    beforeEach(async () => {
      await store.dispatch(startCombat({
        sessionId: 'session-123',
        participants: [
          createMockParticipant('char-1', 'Aragorn', 15),
          createMockParticipant('char-2', 'Legolas', 18),
          createMockParticipant('char-3', 'Gimli', 12),
        ],
      }));
    });

    it('should remove a participant', () => {
      store.dispatch(removeParticipant('char-2'));

      const state = store.getState().combat;
      expect(state.participants.ids).not.toContain('char-2');
      expect(state.participants.entities['char-2']).toBeUndefined();
      expect(state.initiativeOrder).not.toContain('char-2');
    });
  });

  describe('updateParticipant', () => {
    beforeEach(async () => {
      await store.dispatch(startCombat({
        sessionId: 'session-123',
        participants: [createMockParticipant('char-1', 'Aragorn')],
      }));
    });

    it('should update participant properties', () => {
      store.dispatch(updateParticipant({
        id: 'char-1',
        changes: {
          hitPointsCurrent: 15,
          hasActed: true,
          conditions: ['poisoned'],
        },
      }));

      const state = store.getState().combat;
      const participant = state.participants.entities['char-1'];
      
      expect(participant.hitPointsCurrent).toBe(15);
      expect(participant.hasActed).toBe(true);
      expect(participant.conditions).toEqual(['poisoned']);
    });
  });

  describe('applyDamage', () => {
    beforeEach(async () => {
      await store.dispatch(startCombat({
        sessionId: 'session-123',
        participants: [createMockParticipant('char-1', 'Aragorn')],
      }));
    });

    it('should apply damage to hit points', () => {
      store.dispatch(applyDamage({
        targetId: 'char-1',
        damage: 8,
        type: 'slashing',
      }));

      const state = store.getState().combat;
      const participant = state.participants.entities['char-1'];
      
      expect(participant.hitPointsCurrent).toBe(12);
    });

    it('should apply damage to temporary hit points first', () => {
      // Add temp HP
      store.dispatch(updateParticipant({
        id: 'char-1',
        changes: { temporaryHitPoints: 5 },
      }));

      store.dispatch(applyDamage({
        targetId: 'char-1',
        damage: 8,
        type: 'slashing',
      }));

      const state = store.getState().combat;
      const participant = state.participants.entities['char-1'];
      
      expect(participant.temporaryHitPoints).toBe(0);
      expect(participant.hitPointsCurrent).toBe(17); // 20 - 3 overflow damage
    });

    it('should add unconscious condition when HP reaches 0', () => {
      store.dispatch(applyDamage({
        targetId: 'char-1',
        damage: 25,
        type: 'slashing',
      }));

      const state = store.getState().combat;
      const participant = state.participants.entities['char-1'];
      
      expect(participant.hitPointsCurrent).toBe(0);
      expect(participant.conditions).toContain('unconscious');
    });

    it('should not reduce HP below 0', () => {
      store.dispatch(applyDamage({
        targetId: 'char-1',
        damage: 50,
        type: 'slashing',
      }));

      const state = store.getState().combat;
      const participant = state.participants.entities['char-1'];
      
      expect(participant.hitPointsCurrent).toBe(0);
    });
  });

  describe('applyHealing', () => {
    beforeEach(async () => {
      await store.dispatch(startCombat({
        sessionId: 'session-123',
        participants: [createMockParticipant('char-1', 'Aragorn')],
      }));
      
      // Damage the participant first
      store.dispatch(applyDamage({
        targetId: 'char-1',
        damage: 15,
        type: 'slashing',
      }));
    });

    it('should heal hit points', () => {
      store.dispatch(applyHealing({
        targetId: 'char-1',
        healing: 10,
      }));

      const state = store.getState().combat;
      const participant = state.participants.entities['char-1'];
      
      expect(participant.hitPointsCurrent).toBe(15);
    });

    it('should not heal above max HP', () => {
      store.dispatch(applyHealing({
        targetId: 'char-1',
        healing: 30,
      }));

      const state = store.getState().combat;
      const participant = state.participants.entities['char-1'];
      
      expect(participant.hitPointsCurrent).toBe(20); // Max HP
    });

    it('should remove unconscious condition when healed', () => {
      // First make unconscious
      store.dispatch(applyDamage({
        targetId: 'char-1',
        damage: 20,
        type: 'slashing',
      }));

      let state = store.getState().combat;
      expect(state.participants.entities['char-1'].conditions).toContain('unconscious');

      // Then heal
      store.dispatch(applyHealing({
        targetId: 'char-1',
        healing: 5,
      }));

      state = store.getState().combat;
      const participant = state.participants.entities['char-1'];
      
      expect(participant.hitPointsCurrent).toBe(5);
      expect(participant.conditions).not.toContain('unconscious');
    });
  });

  describe('nextTurn', () => {
    beforeEach(async () => {
      await store.dispatch(startCombat({
        sessionId: 'session-123',
        participants: [
          createMockParticipant('char-1', 'Aragorn', 15),
          createMockParticipant('char-2', 'Legolas', 18),
          createMockParticipant('char-3', 'Gimli', 12),
        ],
      }));
    });

    it('should advance to next participant', () => {
      const initialState = store.getState().combat;
      const firstParticipant = initialState.currentParticipantId;
      
      store.dispatch(nextTurn());

      const state = store.getState().combat;
      expect(state.currentParticipantId).not.toBe(firstParticipant);
      expect(state.turn).toBe(1);
    });

    it('should reset turn flags for current participant', () => {
      // Set some flags
      const currentId = store.getState().combat.currentParticipantId!;
      store.dispatch(updateParticipant({
        id: currentId,
        changes: {
          hasActed: true,
          hasBonusActed: true,
          hasReacted: true,
          movementUsed: 20,
        },
      }));

      store.dispatch(nextTurn());

      const state = store.getState().combat;
      const participant = state.participants.entities[currentId];
      
      expect(participant.hasActed).toBe(false);
      expect(participant.hasBonusActed).toBe(false);
      expect(participant.hasReacted).toBe(false);
      expect(participant.movementUsed).toBe(0);
    });

    it('should increment round when wrapping', () => {
      // Advance through all participants
      store.dispatch(nextTurn()); // To second
      store.dispatch(nextTurn()); // To third
      
      const stateBeforeWrap = store.getState().combat;
      expect(stateBeforeWrap.round).toBe(1);
      
      store.dispatch(nextTurn()); // Back to first (new round)

      const state = store.getState().combat;
      expect(state.round).toBe(2);
      expect(state.currentParticipantId).toBe(state.initiativeOrder[0]);
    });
  });

  describe('executeCombatAction', () => {
    beforeEach(async () => {
      await store.dispatch(startCombat({
        sessionId: 'session-123',
        participants: [
          createMockParticipant('char-1', 'Aragorn', 15),
          createMockParticipant('char-2', 'Legolas', 18),
        ],
      }));
    });

    it('should execute attack action', async () => {
      const state = store.getState().combat;
      const currentId = state.currentParticipantId!;
      
      const action: CombatAction = {
        type: 'ATTACK',
        actorId: currentId,
        targetId: 'char-1',
        data: {
          attackRoll: 18,
          damageRoll: 8,
          damageType: 'slashing',
        },
      };

      await store.dispatch(executeCombatAction(action));

      const newState = store.getState().combat;
      const actor = newState.participants.entities[currentId];
      
      expect(actor.hasActed).toBe(true);
      expect(newState.history).toHaveLength(1);
    });

    it('should execute move action', async () => {
      const state = store.getState().combat;
      const currentId = state.currentParticipantId!;
      
      const action: CombatAction = {
        type: 'MOVE',
        actorId: currentId,
        data: {
          distance: 15,
          path: [],
        },
      };

      await store.dispatch(executeCombatAction(action));

      const newState = store.getState().combat;
      const actor = newState.participants.entities[currentId];
      
      expect(actor.movementUsed).toBe(15);
    });

    it('should reject action if not actor turn', async () => {
      const state = store.getState().combat;
      const notCurrentId = state.participants.ids.find(id => id !== state.currentParticipantId)!;
      
      const action: CombatAction = {
        type: 'ATTACK',
        actorId: notCurrentId,
        targetId: 'char-1',
        data: {},
      };

      await store.dispatch(executeCombatAction(action));

      const newState = store.getState().combat;
      expect(newState.errors.action).toContain('Not your turn');
    });

    it('should set pending action', () => {
      const state = store.getState().combat;
      const currentId = state.currentParticipantId!;
      
      const action: CombatAction = {
        type: 'ATTACK',
        actorId: currentId,
        targetId: 'char-1',
        data: {},
      };

      store.dispatch(executeCombatAction(action));

      const pendingState = store.getState().combat;
      expect(pendingState.pendingAction).toEqual(action);
    });
  });

  describe('endCombat', () => {
    it('should reset to initial state', async () => {
      await store.dispatch(startCombat({
        sessionId: 'session-123',
        participants: [createMockParticipant('char-1', 'Aragorn')],
      }));

      store.dispatch(endCombat());

      const state = store.getState().combat;
      expect(state).toEqual({
        active: false,
        sessionId: null,
        round: 0,
        turn: 0,
        participants: {
          ids: [],
          entities: {},
        },
        initiativeOrder: [],
        currentParticipantId: null,
        history: [],
        historyIndex: -1,
        pendingAction: null,
        isLoading: {},
        errors: {},
      });
    });
  });

  describe('restoreState', () => {
    it('should restore partial state', async () => {
      await store.dispatch(startCombat({
        sessionId: 'session-123',
        participants: [createMockParticipant('char-1', 'Aragorn')],
      }));

      const previousState = {
        round: 5,
        turn: 10,
      };

      store.dispatch(restoreState(previousState));

      const state = store.getState().combat;
      expect(state.round).toBe(5);
      expect(state.turn).toBe(10);
      expect(state.active).toBe(true); // Other state preserved
    });
  });
});