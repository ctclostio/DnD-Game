import { useCallback } from 'react';
import { useAppDispatch, useAppSelector } from '@store/index';
import { undo, redo, addUndoableAction } from '@store/slices/dmToolsSlice';

export const useUndo = () => {
  const dispatch = useAppDispatch();
  const { canUndo, canRedo } = useAppSelector(state => state.dmTools);
  
  const performUndo = useCallback(() => {
    if (canUndo) {
      dispatch(undo());
    }
  }, [dispatch, canUndo]);
  
  const performRedo = useCallback(() => {
    if (canRedo) {
      dispatch(redo());
    }
  }, [dispatch, canRedo]);
  
  const addUndoable = useCallback((
    description: string,
    undoAction: () => any,
    redoAction: () => any
  ) => {
    dispatch(addUndoableAction({
      id: `action-${Date.now()}`,
      type: 'user-action',
      timestamp: Date.now(),
      description,
      undo: undoAction,
      redo: redoAction,
    }));
  }, [dispatch]);
  
  return {
    undo: performUndo,
    redo: performRedo,
    canUndo,
    canRedo,
    addUndoable,
  };
};