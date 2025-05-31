import { Middleware } from '@reduxjs/toolkit';
import { undo, redo } from '../slices/dmToolsSlice';

export const undoMiddleware: Middleware = (store) => (next) => (action) => {
  // Handle undo action
  if (undo.match(action)) {
    const state = store.getState();
    const undoStack = state.dmTools.undoStack;
    
    if (undoStack.length > 0) {
      const lastAction = undoStack[undoStack.length - 1];
      
      // Execute the undo function
      try {
        const undoAction = lastAction.undo();
        if (undoAction) {
          store.dispatch(undoAction);
        }
      } catch (error) {
        console.error('Undo operation failed:', error);
      }
    }
  }
  
  // Handle redo action
  if (redo.match(action)) {
    const state = store.getState();
    const redoStack = state.dmTools.redoStack;
    
    if (redoStack.length > 0) {
      const lastAction = redoStack[redoStack.length - 1];
      
      // Execute the redo function
      try {
        const redoAction = lastAction.redo();
        if (redoAction) {
          store.dispatch(redoAction);
        }
      } catch (error) {
        console.error('Redo operation failed:', error);
      }
    }
  }
  
  return next(action);
};