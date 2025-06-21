import { Middleware } from '@reduxjs/toolkit';
import { undo, redo } from '../slices/dmToolsSlice';

// Helper function to execute an operation from the stack
const executeStackOperation = (
  store: any,
  stack: any[],
  operationName: string,
  operationFn: (action: any) => any
) => {
  if (stack.length === 0) return;
  
  const lastAction = stack[stack.length - 1];
  
  try {
    const resultAction = operationFn(lastAction);
    if (resultAction) {
      store.dispatch(resultAction);
    }
  } catch (error) {
    console.error(`${operationName} operation failed:`, error);
  }
};

// Helper to handle undo operation
const handleUndo = (store: any) => {
  const state = store.getState();
  const undoStack = state.dmTools.undoStack;
  executeStackOperation(store, undoStack, 'Undo', (action) => action.undo());
};

// Helper to handle redo operation
const handleRedo = (store: any) => {
  const state = store.getState();
  const redoStack = state.dmTools.redoStack;
  executeStackOperation(store, redoStack, 'Redo', (action) => action.redo());
};

export const undoMiddleware: Middleware = (store) => (next) => (action) => {
  // Handle undo action
  if (undo.match(action)) {
    handleUndo(store);
  }
  
  // Handle redo action
  if (redo.match(action)) {
    handleRedo(store);
  }
  
  return next(action);
};