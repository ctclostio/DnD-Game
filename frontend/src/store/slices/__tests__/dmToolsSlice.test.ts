import { configureStore } from '@reduxjs/toolkit';
import dmToolsReducer, {
  addUndoableAction,
  undo,
  redo,
  clearHistory,
  updateSessionNotes,
  updateCampaignNotes,
  setQuickReferences,
  setLoading,
  setError,
} from '../dmToolsSlice';
import { UndoableAction } from '../../../types/state';

// Helper functions defined outside describe blocks to reduce nesting
const createMockUndoableAction = (id: string, description: string): UndoableAction => ({
  id,
  type: 'test-action',
  timestamp: Date.now(),
  description,
  undo: () => ({ type: 'TEST_UNDO', payload: { id } }),
  redo: () => ({ type: 'TEST_REDO', payload: { id } }),
});

const createStore = () => configureStore({
  reducer: {
    dmTools: dmToolsReducer,
  },
});

const addMultipleActions = (store: ReturnType<typeof createStore>, count: number) => {
  for (let i = 1; i <= count; i++) {
    const action = createMockUndoableAction(`action-${i}`, `Test action ${i}`);
    store.dispatch(addUndoableAction(action));
  }
};

// Helper to extract action IDs to reduce nesting
const extractActionIds = (actions: UndoableAction[]): string[] => {
  return actions.map(a => a.id);
};

describe('dmToolsSlice', () => {
  let store: ReturnType<typeof configureStore>;

  beforeEach(() => {
    store = createStore();
  });

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const state = store.getState().dmTools;
      
      expect(state).toEqual({
        canUndo: false,
        canRedo: false,
        undoStack: [],
        redoStack: [],
        quickReferences: {
          conditions: [],
          rules: [],
        },
        sessionNotes: '',
        campaignNotes: '',
        isLoading: {},
        errors: {},
      });
    });
  });

  describe('undo/redo functionality', () => {
    describe('addUndoableAction', () => {
      it('should add action to undo stack', () => {
        const action = createMockUndoableAction('action-1', 'Test action 1');
        
        store.dispatch(addUndoableAction(action));

        const state = store.getState().dmTools;
        expect(state.undoStack).toHaveLength(1);
        expect(state.undoStack[0]).toEqual(action);
        expect(state.canUndo).toBe(true);
        expect(state.canRedo).toBe(false);
      });

      it('should clear redo stack when adding new action', () => {
        // Add an action and undo it
        const action1 = createMockUndoableAction('action-1', 'Test action 1');
        store.dispatch(addUndoableAction(action1));
        store.dispatch(undo());
        
        let state = store.getState().dmTools;
        expect(state.redoStack).toHaveLength(1);
        
        // Add a new action
        const action2 = createMockUndoableAction('action-2', 'Test action 2');
        store.dispatch(addUndoableAction(action2));
        
        state = store.getState().dmTools;
        expect(state.redoStack).toHaveLength(0);
        expect(state.undoStack).toHaveLength(1);
        expect(state.undoStack[0]).toEqual(action2);
      });

      it('should limit undo stack size to MAX_UNDO_HISTORY (50)', () => {
        // Add 51 actions
        addMultipleActions(store, 51);

        const state = store.getState().dmTools;
        expect(state.undoStack).toHaveLength(50);
        // First action should be removed
        expect(state.undoStack[0].id).toBe('action-1');
        expect(state.undoStack[49].id).toBe('action-50');
      });

      it('should handle multiple actions in sequence', () => {
        const actions = [
          createMockUndoableAction('action-1', 'Test action 1'),
          createMockUndoableAction('action-2', 'Test action 2'),
          createMockUndoableAction('action-3', 'Test action 3'),
        ];

        for (const action of actions) {
          store.dispatch(addUndoableAction(action));
        }

        const state = store.getState().dmTools;
        expect(state.undoStack).toHaveLength(3);
        const ids = extractActionIds(state.undoStack);
        expect(ids).toEqual(['action-1', 'action-2', 'action-3']);
        expect(state.canUndo).toBe(true);
      });
    });

    describe('undo', () => {
      beforeEach(() => {
        // Add some actions to undo
        addMultipleActions(store, 3);
      });

      it('should move action from undo stack to redo stack', () => {
        store.dispatch(undo());

        const state = store.getState().dmTools;
        expect(state.undoStack).toHaveLength(2);
        expect(state.redoStack).toHaveLength(1);
        expect(state.redoStack[0].id).toBe('action-3');
        expect(state.canUndo).toBe(true);
        expect(state.canRedo).toBe(true);
      });

      it('should handle multiple undos', () => {
        store.dispatch(undo());
        store.dispatch(undo());

        const state = store.getState().dmTools;
        expect(state.undoStack).toHaveLength(1);
        expect(state.redoStack).toHaveLength(2);
        expect(extractActionIds(state.redoStack)).toEqual(['action-3', 'action-2']);
        expect(state.canUndo).toBe(true);
        expect(state.canRedo).toBe(true);
      });

      it('should set canUndo to false when stack is empty', () => {
        store.dispatch(undo());
        store.dispatch(undo());
        store.dispatch(undo());

        const state = store.getState().dmTools;
        expect(state.undoStack).toHaveLength(0);
        expect(state.canUndo).toBe(false);
        expect(state.canRedo).toBe(true);
      });

      it('should not crash when undo on empty stack', () => {
        // Clear all actions
        store.dispatch(clearHistory());
        
        // Try to undo
        store.dispatch(undo());

        const state = store.getState().dmTools;
        expect(state.undoStack).toHaveLength(0);
        expect(state.redoStack).toHaveLength(0);
        expect(state.canUndo).toBe(false);
        expect(state.canRedo).toBe(false);
      });
    });

    describe('redo', () => {
      beforeEach(() => {
        // Add actions and undo them
        addMultipleActions(store, 3);
        store.dispatch(undo());
        store.dispatch(undo());
      });

      it('should move action from redo stack to undo stack', () => {
        store.dispatch(redo());

        const state = store.getState().dmTools;
        expect(state.undoStack).toHaveLength(2);
        expect(state.redoStack).toHaveLength(1);
        expect(state.undoStack[1].id).toBe('action-2');
        expect(state.canUndo).toBe(true);
        expect(state.canRedo).toBe(true);
      });

      it('should handle multiple redos', () => {
        store.dispatch(redo());
        store.dispatch(redo());

        const state = store.getState().dmTools;
        expect(state.undoStack).toHaveLength(3);
        expect(state.redoStack).toHaveLength(0);
        expect(extractActionIds(state.undoStack)).toEqual(['action-1', 'action-2', 'action-3']);
        expect(state.canRedo).toBe(false);
      });

      it('should set canRedo to false when stack is empty', () => {
        store.dispatch(redo());
        store.dispatch(redo());

        const state = store.getState().dmTools;
        expect(state.redoStack).toHaveLength(0);
        expect(state.canRedo).toBe(false);
      });

      it('should not crash when redo on empty stack', () => {
        // Clear redo stack
        store.dispatch(clearHistory());
        
        // Try to redo
        store.dispatch(redo());

        const state = store.getState().dmTools;
        expect(state.undoStack).toHaveLength(0);
        expect(state.redoStack).toHaveLength(0);
        expect(state.canUndo).toBe(false);
        expect(state.canRedo).toBe(false);
      });
    });

    describe('clearHistory', () => {
      it('should clear both undo and redo stacks', () => {
        // Add some actions
        addMultipleActions(store, 3);
        store.dispatch(undo());

        // Clear history
        store.dispatch(clearHistory());

        const state = store.getState().dmTools;
        expect(state.undoStack).toHaveLength(0);
        expect(state.redoStack).toHaveLength(0);
        expect(state.canUndo).toBe(false);
        expect(state.canRedo).toBe(false);
      });
    });
  });

  describe('notes management', () => {
    describe('updateSessionNotes', () => {
      it('should update session notes', () => {
        const notes = 'Players encountered a dragon in the cave.';
        
        store.dispatch(updateSessionNotes(notes));

        const state = store.getState().dmTools;
        expect(state.sessionNotes).toBe(notes);
      });

      it('should overwrite existing session notes', () => {
        store.dispatch(updateSessionNotes('Initial notes'));
        store.dispatch(updateSessionNotes('Updated notes'));

        const state = store.getState().dmTools;
        expect(state.sessionNotes).toBe('Updated notes');
      });

      it('should handle empty notes', () => {
        store.dispatch(updateSessionNotes('Some notes'));
        store.dispatch(updateSessionNotes(''));

        const state = store.getState().dmTools;
        expect(state.sessionNotes).toBe('');
      });
    });

    describe('updateCampaignNotes', () => {
      it('should update campaign notes', () => {
        const notes = 'The ancient prophecy speaks of five heroes...';
        
        store.dispatch(updateCampaignNotes(notes));

        const state = store.getState().dmTools;
        expect(state.campaignNotes).toBe(notes);
      });

      it('should maintain separate session and campaign notes', () => {
        const sessionNotes = 'Session specific notes';
        const campaignNotes = 'Campaign overview notes';
        
        store.dispatch(updateSessionNotes(sessionNotes));
        store.dispatch(updateCampaignNotes(campaignNotes));

        const state = store.getState().dmTools;
        expect(state.sessionNotes).toBe(sessionNotes);
        expect(state.campaignNotes).toBe(campaignNotes);
      });
    });
  });

  describe('quick references', () => {
    it('should set quick references', () => {
      const references = {
        conditions: [
          { name: 'Blinded', description: 'Cannot see' },
          { name: 'Charmed', description: 'Cannot attack charmer' },
        ],
        rules: [
          { name: 'Advantage', description: 'Roll twice, take higher' },
          { name: 'Disadvantage', description: 'Roll twice, take lower' },
        ],
      };

      store.dispatch(setQuickReferences(references));

      const state = store.getState().dmTools;
      expect(state.quickReferences).toEqual(references);
    });

    it('should replace existing references', () => {
      const initial = {
        conditions: [{ name: 'Poisoned', description: 'Disadvantage on attacks' }],
        rules: [{ name: 'Cover', description: '+2 AC' }],
      };
      
      const updated = {
        conditions: [{ name: 'Stunned', description: 'Incapacitated' }],
        rules: [{ name: 'Flanking', description: 'Advantage on attacks' }],
      };

      store.dispatch(setQuickReferences(initial));
      store.dispatch(setQuickReferences(updated));

      const state = store.getState().dmTools;
      expect(state.quickReferences).toEqual(updated);
    });

    it('should handle empty references', () => {
      store.dispatch(setQuickReferences({
        conditions: [],
        rules: [],
      }));

      const state = store.getState().dmTools;
      expect(state.quickReferences.conditions).toEqual([]);
      expect(state.quickReferences.rules).toEqual([]);
    });
  });

  describe('loading and error states', () => {
    describe('setLoading', () => {
      it('should set loading state for a key', () => {
        store.dispatch(setLoading({ key: 'fetchRules', value: true }));

        const state = store.getState().dmTools;
        expect(state.isLoading.fetchRules).toBe(true);
      });

      it('should handle multiple loading states', () => {
        store.dispatch(setLoading({ key: 'fetchRules', value: true }));
        store.dispatch(setLoading({ key: 'saveNotes', value: true }));
        store.dispatch(setLoading({ key: 'fetchRules', value: false }));

        const state = store.getState().dmTools;
        expect(state.isLoading.fetchRules).toBe(false);
        expect(state.isLoading.saveNotes).toBe(true);
      });
    });

    describe('setError', () => {
      it('should set error for a key', () => {
        const error = 'Failed to fetch rules';
        store.dispatch(setError({ key: 'fetchRules', error }));

        const state = store.getState().dmTools;
        expect(state.errors.fetchRules).toBe(error);
      });

      it('should clear error when setting null', () => {
        store.dispatch(setError({ key: 'fetchRules', error: 'Some error' }));
        store.dispatch(setError({ key: 'fetchRules', error: null }));

        const state = store.getState().dmTools;
        expect(state.errors.fetchRules).toBeNull();
      });

      it('should handle multiple error states', () => {
        store.dispatch(setError({ key: 'fetchRules', error: 'Rules error' }));
        store.dispatch(setError({ key: 'saveNotes', error: 'Notes error' }));

        const state = store.getState().dmTools;
        expect(state.errors.fetchRules).toBe('Rules error');
        expect(state.errors.saveNotes).toBe('Notes error');
      });
    });
  });

  describe('complex scenarios', () => {
    it('should handle a complete undo/redo workflow', () => {
      const actions = [
        createMockUndoableAction('move-1', 'Move character to A5'),
        createMockUndoableAction('attack-1', 'Attack goblin'),
        createMockUndoableAction('spell-1', 'Cast fireball'),
      ];

      // Perform actions
      for (const action of actions) {
        store.dispatch(addUndoableAction(action));
      }

      // Undo last action
      store.dispatch(undo());
      let state = store.getState().dmTools;
      expect(state.undoStack).toHaveLength(2);
      expect(state.redoStack[0].description).toBe('Cast fireball');

      // Undo another
      store.dispatch(undo());
      state = store.getState().dmTools;
      expect(state.undoStack).toHaveLength(1);
      expect(state.redoStack).toHaveLength(2);

      // Redo one
      store.dispatch(redo());
      state = store.getState().dmTools;
      expect(state.undoStack).toHaveLength(2);
      expect(state.undoStack[1].description).toBe('Attack goblin');

      // Add new action (should clear redo stack)
      const newAction = createMockUndoableAction('heal-1', 'Heal fighter');
      store.dispatch(addUndoableAction(newAction));
      
      state = store.getState().dmTools;
      expect(state.undoStack).toHaveLength(3);
      expect(state.redoStack).toHaveLength(0);
      expect(state.canRedo).toBe(false);
    });

    it('should maintain all state independently', () => {
      // Set up various states
      store.dispatch(updateSessionNotes('Combat round 3'));
      store.dispatch(updateCampaignNotes('Chapter 2: The Dark Forest'));
      store.dispatch(setQuickReferences({
        conditions: [{ name: 'Frightened', desc: 'Disadvantage' }],
        rules: [{ name: 'Grapple', desc: 'Contested check' }],
      }));
      store.dispatch(setLoading({ key: 'autoSave', value: true }));
      store.dispatch(setError({ key: 'networkError', error: 'Connection lost' }));
      
      // Add some undoable actions
      store.dispatch(addUndoableAction(createMockUndoableAction('action-1', 'Action 1')));
      store.dispatch(addUndoableAction(createMockUndoableAction('action-2', 'Action 2')));
      store.dispatch(undo());

      // Verify all states are maintained
      const state = store.getState().dmTools;
      expect(state.sessionNotes).toBe('Combat round 3');
      expect(state.campaignNotes).toBe('Chapter 2: The Dark Forest');
      expect(state.quickReferences.conditions).toHaveLength(1);
      expect(state.quickReferences.rules).toHaveLength(1);
      expect(state.isLoading.autoSave).toBe(true);
      expect(state.errors.networkError).toBe('Connection lost');
      expect(state.undoStack).toHaveLength(1);
      expect(state.redoStack).toHaveLength(1);
    });
  });
});