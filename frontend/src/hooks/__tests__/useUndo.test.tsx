import React from 'react';
import { renderHook, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { useUndo } from '../useUndo';
import dmToolsSlice, { undo, redo, addUndoableAction } from '../../store/slices/dmToolsSlice';

// Mock the store hooks
const mockDispatch = jest.fn();
const mockUseAppSelector = jest.fn();

jest.mock('../../store/index', () => ({
  useAppDispatch: () => mockDispatch,
  useAppSelector: (selector: any) => mockUseAppSelector(selector),
}));

// Create a test store
const createTestStore = (initialState = {}) => {
  return configureStore({
    reducer: {
      dmTools: dmToolsSlice,
    },
    preloadedState: {
      dmTools: {
        canUndo: false,
        canRedo: false,
        ...initialState,
      },
    },
  });
};

describe('useUndo', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAppSelector.mockImplementation((selector) => 
      selector({
        dmTools: {
          canUndo: false,
          canRedo: false,
        },
      })
    );
  });

  it('should provide undo/redo functions and state', () => {
    const { result } = renderHook(() => useUndo());

    expect(result.current.undo).toBeDefined();
    expect(result.current.redo).toBeDefined();
    expect(result.current.canUndo).toBe(false);
    expect(result.current.canRedo).toBe(false);
    expect(result.current.addUndoable).toBeDefined();
  });

  it('should dispatch undo action when canUndo is true', () => {
    mockUseAppSelector.mockImplementation((selector) => 
      selector({
        dmTools: {
          canUndo: true,
          canRedo: false,
        },
      })
    );

    const { result } = renderHook(() => useUndo());

    act(() => {
      result.current.undo();
    });

    expect(mockDispatch).toHaveBeenCalledWith(undo());
  });

  it('should not dispatch undo action when canUndo is false', () => {
    mockUseAppSelector.mockImplementation((selector) => 
      selector({
        dmTools: {
          canUndo: false,
          canRedo: false,
        },
      })
    );

    const { result } = renderHook(() => useUndo());

    act(() => {
      result.current.undo();
    });

    expect(mockDispatch).not.toHaveBeenCalled();
  });

  it('should dispatch redo action when canRedo is true', () => {
    mockUseAppSelector.mockImplementation((selector) => 
      selector({
        dmTools: {
          canUndo: false,
          canRedo: true,
        },
      })
    );

    const { result } = renderHook(() => useUndo());

    act(() => {
      result.current.redo();
    });

    expect(mockDispatch).toHaveBeenCalledWith(redo());
  });

  it('should not dispatch redo action when canRedo is false', () => {
    mockUseAppSelector.mockImplementation((selector) => 
      selector({
        dmTools: {
          canUndo: false,
          canRedo: false,
        },
      })
    );

    const { result } = renderHook(() => useUndo());

    act(() => {
      result.current.redo();
    });

    expect(mockDispatch).not.toHaveBeenCalled();
  });

  it('should add undoable action', () => {
    const { result } = renderHook(() => useUndo());

    const undoAction = jest.fn();
    const redoAction = jest.fn();
    const description = 'Test action';

    act(() => {
      result.current.addUndoable(description, undoAction, redoAction);
    });

    expect(mockDispatch).toHaveBeenCalledWith(
      addUndoableAction({
        id: expect.stringMatching(/^action-\d+$/),
        type: 'user-action',
        timestamp: expect.any(Number),
        description,
        undo: undoAction,
        redo: redoAction,
      })
    );
  });

  it('should generate unique action IDs', () => {
    const { result } = renderHook(() => useUndo());

    const actions: any[] = [];
    mockDispatch.mockImplementation((action) => {
      actions.push(action);
    });

    act(() => {
      result.current.addUndoable('Action 1', jest.fn(), jest.fn());
      result.current.addUndoable('Action 2', jest.fn(), jest.fn());
    });

    expect(actions).toHaveLength(2);
    expect(actions[0].payload.id).not.toBe(actions[1].payload.id);
  });

  it('should update when state changes', () => {
    const { result, rerender } = renderHook(() => useUndo());

    expect(result.current.canUndo).toBe(false);
    expect(result.current.canRedo).toBe(false);

    // Update mock to return new state
    mockUseAppSelector.mockImplementation((selector) => 
      selector({
        dmTools: {
          canUndo: true,
          canRedo: true,
        },
      })
    );

    rerender();

    expect(result.current.canUndo).toBe(true);
    expect(result.current.canRedo).toBe(true);
  });

  it('should maintain stable function references', () => {
    const { result, rerender } = renderHook(() => useUndo());

    const { undo, redo, addUndoable } = result.current;

    rerender();

    expect(result.current.undo).toBe(undo);
    expect(result.current.redo).toBe(redo);
    expect(result.current.addUndoable).toBe(addUndoable);
  });

  it('should handle rapid undo/redo calls', () => {
    mockUseAppSelector.mockImplementation((selector) => 
      selector({
        dmTools: {
          canUndo: true,
          canRedo: true,
        },
      })
    );

    const { result } = renderHook(() => useUndo());

    act(() => {
      result.current.undo();
      result.current.undo();
      result.current.redo();
      result.current.undo();
    });

    expect(mockDispatch).toHaveBeenCalledTimes(4);
    expect(mockDispatch).toHaveBeenNthCalledWith(1, undo());
    expect(mockDispatch).toHaveBeenNthCalledWith(2, undo());
    expect(mockDispatch).toHaveBeenNthCalledWith(3, redo());
    expect(mockDispatch).toHaveBeenNthCalledWith(4, undo());
  });
});

// Integration test with real Redux store
describe('useUndo - Integration', () => {
  it('should work with real Redux store', () => {
    const store = createTestStore({
      canUndo: true,
      canRedo: false,
      history: [],
      currentIndex: -1,
    });

    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <Provider store={store}>{children}</Provider>
    );

    // Reset mocks to use real implementation
    jest.unmock('../../store/index');
    jest.resetModules();
    
    // Re-import to get unmocked version
    const { useUndo: realUseUndo } = require('../useUndo');

    const { result } = renderHook(() => realUseUndo(), { wrapper });

    expect(result.current.canUndo).toBeDefined();
    expect(result.current.canRedo).toBeDefined();
  });
});