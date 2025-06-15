import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import { CombatState, CombatAction, EntityState, RootState } from '../../types/state';
import { CombatParticipant } from '../../types/game';
import { addUndoableAction } from './dmToolsSlice';

const initialState: CombatState = {
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
};

// Helper function to create entity state
function createEntityState<T>(items: T[], getId: (item: T) => string): EntityState<T> {
  const ids = items.map(getId);
  const entities = items.reduce((acc, item) => {
    acc[getId(item)] = item;
    return acc;
  }, {} as { [id: string]: T });
  
  return { ids, entities };
}

// Async thunks for combat actions
export const startCombat = createAsyncThunk(
  'combat/start',
  async ({ sessionId, participants }: { sessionId: string; participants: CombatParticipant[] }) => {
    // Roll initiative for all participants
    const withInitiative = participants.map(p => ({
      ...p,
      initiative: Math.floor(Math.random() * 20) + 1 + p.initiativeModifier,
    }));
    
    // Sort by initiative (descending)
    withInitiative.sort((a, b) => b.initiative - a.initiative);
    
    return {
      sessionId,
      participants: withInitiative,
    };
  }
);

export const executeCombatAction = createAsyncThunk(
  'combat/executeAction',
  async (action: CombatAction, { getState, dispatch }) => {
    const state = getState() as RootState;
    const combat = state.combat as CombatState;
    
    // Validate action
    const actor = combat.participants.entities[action.actorId];
    if (!actor) {
      throw new Error('Actor not found');
    }
    
    // Check if it's the actor's turn
    if (combat.currentParticipantId !== action.actorId) {
      throw new Error('Not your turn');
    }
    
    // Validate action type availability
    switch (action.type) {
      case 'ATTACK':
        if (actor.hasActed) {
          throw new Error('Already used action this turn');
        }
        break;
      case 'MOVE':
        const remainingMovement = actor.movementMax - actor.movementUsed;
        if (remainingMovement <= 0) {
          throw new Error('No movement remaining');
        }
        break;
      // Add more validation as needed
    }
    
    // Create undoable action for DM
    const previousState = {
      participants: combat.participants,
      currentParticipantId: combat.currentParticipantId,
      round: combat.round,
      turn: combat.turn,
    };
    
    dispatch(addUndoableAction({
      id: `combat-${Date.now()}`,
      type: 'combat-action',
      timestamp: Date.now(),
      description: `${actor.name} performed ${action.type}`,
      undo: () => ({
        type: 'combat/restoreState',
        payload: previousState,
      }),
      redo: () => ({
        type: 'combat/executeCombatAction',
        payload: action,
      }),
    }));
    
    return action;
  }
);

const combatSlice = createSlice({
  name: 'combat',
  initialState,
  reducers: {
    // Add participant
    addParticipant: (state, action: PayloadAction<CombatParticipant>) => {
      const participant = action.payload;
      state.participants.ids.push(participant.id);
      state.participants.entities[participant.id] = participant;
      
      // Re-sort initiative order
      state.initiativeOrder = [...state.participants.ids].sort((a, b) => {
        const aInit = state.participants.entities[a].initiative;
        const bInit = state.participants.entities[b].initiative;
        return bInit - aInit;
      });
    },
    
    // Remove participant
    removeParticipant: (state, action: PayloadAction<string>) => {
      const id = action.payload;
      state.participants.ids = state.participants.ids.filter(pid => pid !== id);
      delete state.participants.entities[id];
      state.initiativeOrder = state.initiativeOrder.filter(pid => pid !== id);
    },
    
    // Update participant
    updateParticipant: (state, action: PayloadAction<{ id: string; changes: Partial<CombatParticipant> }>) => {
      const { id, changes } = action.payload;
      if (state.participants.entities[id]) {
        state.participants.entities[id] = {
          ...state.participants.entities[id],
          ...changes,
        };
      }
    },
    
    // Damage/Heal
    applyDamage: (state, action: PayloadAction<{ targetId: string; damage: number; type: string }>) => {
      const { targetId, damage } = action.payload;
      const target = state.participants.entities[targetId];
      if (target) {
        // Apply damage to temp HP first
        let remainingDamage = damage;
        if (target.temporaryHitPoints > 0) {
          const tempDamage = Math.min(remainingDamage, target.temporaryHitPoints);
          target.temporaryHitPoints -= tempDamage;
          remainingDamage -= tempDamage;
        }
        
        // Apply remaining damage to HP
        target.hitPointsCurrent = Math.max(0, target.hitPointsCurrent - remainingDamage);
        
        // Check for unconscious
        if (target.hitPointsCurrent === 0 && !target.conditions.includes('unconscious')) {
          target.conditions.push('unconscious');
        }
      }
    },
    
    applyHealing: (state, action: PayloadAction<{ targetId: string; healing: number }>) => {
      const { targetId, healing } = action.payload;
      const target = state.participants.entities[targetId];
      if (target) {
        target.hitPointsCurrent = Math.min(target.hitPointsMax, target.hitPointsCurrent + healing);
        
        // Remove unconscious if healed
        if (target.hitPointsCurrent > 0) {
          target.conditions = target.conditions.filter(c => c !== 'unconscious');
        }
      }
    },
    
    // Turn management
    nextTurn: (state) => {
      const currentIndex = state.initiativeOrder.indexOf(state.currentParticipantId || '');
      const nextIndex = (currentIndex + 1) % state.initiativeOrder.length;
      
      // Reset turn flags for current participant
      if (state.currentParticipantId) {
        const current = state.participants.entities[state.currentParticipantId];
        if (current) {
          current.hasActed = false;
          current.hasBonusActed = false;
          current.hasReacted = false;
          current.movementUsed = 0;
        }
      }
      
      // If we've wrapped around, increment round
      if (nextIndex === 0) {
        state.round++;
      }
      
      state.currentParticipantId = state.initiativeOrder[nextIndex];
      state.turn++;
    },
    
    // End combat
    endCombat: (state) => {
      return initialState;
    },
    
    // Restore state (for undo)
    restoreState: (state, action: PayloadAction<Partial<CombatState>>) => {
      return { ...state, ...action.payload };
    },
  },
  extraReducers: (builder) => {
    builder
      // Start combat
      .addCase(startCombat.pending, (state) => {
        state.isLoading.startCombat = true;
      })
      .addCase(startCombat.fulfilled, (state, action) => {
        state.isLoading.startCombat = false;
        state.active = true;
        state.sessionId = action.payload.sessionId;
        state.round = 1;
        state.turn = 0;
        
        const participantState = createEntityState(action.payload.participants, p => p.id);
        state.participants = participantState;
        state.initiativeOrder = action.payload.participants.map(p => p.id);
        state.currentParticipantId = state.initiativeOrder[0];
      })
      .addCase(startCombat.rejected, (state, action) => {
        state.isLoading.startCombat = false;
        state.errors.startCombat = action.error.message || 'Failed to start combat';
      })
      
      // Execute combat action
      .addCase(executeCombatAction.pending, (state, action) => {
        state.pendingAction = action.meta.arg;
      })
      .addCase(executeCombatAction.fulfilled, (state, action) => {
        state.pendingAction = null;
        const combatAction = action.payload;
        
        // Handle different action types
        switch (combatAction.type) {
          case 'ATTACK':
            const actor = state.participants.entities[combatAction.actorId];
            if (actor) {
              actor.hasActed = true;
            }
            break;
          case 'MOVE':
            const mover = state.participants.entities[combatAction.actorId];
            if (mover && typeof combatAction.data.distance === 'number') {
              mover.movementUsed += combatAction.data.distance;
            }
            break;
          // Handle other action types
        }
        
        // Add to history
        state.history.push({
          timestamp: Date.now(),
          action: combatAction,
          previousState: {},
          description: `${combatAction.type} action`,
        });
      })
      .addCase(executeCombatAction.rejected, (state, action) => {
        state.pendingAction = null;
        state.errors.action = action.error.message || 'Action failed';
      });
  },
});

export const {
  addParticipant,
  removeParticipant,
  updateParticipant,
  applyDamage,
  applyHealing,
  nextTurn,
  endCombat,
  restoreState,
} = combatSlice.actions;

export default combatSlice.reducer;
