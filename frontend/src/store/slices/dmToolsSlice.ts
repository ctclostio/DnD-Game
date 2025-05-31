import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { DMToolsState, UndoableAction } from '../../types/state';

const MAX_UNDO_HISTORY = 50;

const initialState: DMToolsState = {
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
};

const dmToolsSlice = createSlice({
  name: 'dmTools',
  initialState,
  reducers: {
    // Add an undoable action
    addUndoableAction: (state, action: PayloadAction<UndoableAction>) => {
      state.undoStack.push(action.payload);
      // Limit stack size
      if (state.undoStack.length > MAX_UNDO_HISTORY) {
        state.undoStack.shift();
      }
      // Clear redo stack when new action is performed
      state.redoStack = [];
      state.canUndo = true;
      state.canRedo = false;
    },
    
    // Perform undo
    undo: (state) => {
      if (state.undoStack.length > 0) {
        const action = state.undoStack.pop()!;
        state.redoStack.push(action);
        state.canUndo = state.undoStack.length > 0;
        state.canRedo = true;
        // The actual undo operation will be handled by middleware
      }
    },
    
    // Perform redo
    redo: (state) => {
      if (state.redoStack.length > 0) {
        const action = state.redoStack.pop()!;
        state.undoStack.push(action);
        state.canUndo = true;
        state.canRedo = state.redoStack.length > 0;
        // The actual redo operation will be handled by middleware
      }
    },
    
    // Clear history
    clearHistory: (state) => {
      state.undoStack = [];
      state.redoStack = [];
      state.canUndo = false;
      state.canRedo = false;
    },
    
    // Notes management
    updateSessionNotes: (state, action: PayloadAction<string>) => {
      state.sessionNotes = action.payload;
    },
    
    updateCampaignNotes: (state, action: PayloadAction<string>) => {
      state.campaignNotes = action.payload;
    },
    
    // Quick references
    setQuickReferences: (state, action: PayloadAction<{ conditions: any[]; rules: any[] }>) => {
      state.quickReferences = action.payload;
    },
    
    // Loading states
    setLoading: (state, action: PayloadAction<{ key: string; value: boolean }>) => {
      state.isLoading[action.payload.key] = action.payload.value;
    },
    
    // Error handling
    setError: (state, action: PayloadAction<{ key: string; error: string | null }>) => {
      state.errors[action.payload.key] = action.payload.error;
    },
  },
});

export const {
  addUndoableAction,
  undo,
  redo,
  clearHistory,
  updateSessionNotes,
  updateCampaignNotes,
  setQuickReferences,
  setLoading,
  setError,
} = dmToolsSlice.actions;

export default dmToolsSlice.reducer;