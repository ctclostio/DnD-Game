import { configureStore } from '@reduxjs/toolkit';
import uiReducer, {
  setTheme,
  toggleTheme,
  toggleSidebar,
  setSidebarOpen,
  openModal,
  closeModal,
  closeAllModals,
  addNotification,
  removeNotification,
  clearNotifications,
  setShortcutsEnabled,
  setCustomBinding,
  resetCustomBindings,
  updateUISettings,
} from '../uiSlice';

// Mock localStorage directly
const mockSetItem = jest.fn();
const mockGetItem = jest.fn();
const mockRemoveItem = jest.fn();
const mockClear = jest.fn();

Object.defineProperty(window, 'localStorage', {
  value: {
    setItem: mockSetItem,
    getItem: mockGetItem,
    removeItem: mockRemoveItem,
    clear: mockClear,
  },
  writable: true,
});

// Mock Date.now() for consistent notification IDs
const mockNow = 1640995200000; // 2022-01-01T00:00:00.000Z
const originalDateNow = Date.now;

// Helper functions to reduce nesting
const createStore = () => configureStore({
  reducer: {
    ui: uiReducer,
  },
});

const addMultipleNotifications = (store: any, count: number) => {
  for (let i = 0; i < count; i++) {
    store.dispatch(addNotification({
      type: 'info',
      message: `Notification ${i}`,
    }));
  }
};

const setupNotifications = (store: any) => {
  store.dispatch(addNotification({ type: 'info', message: 'Notification 1' }));
  store.dispatch(addNotification({ type: 'success', message: 'Notification 2' }));
  store.dispatch(addNotification({ type: 'error', message: 'Notification 3' }));
};

describe('uiSlice', () => {
  let store: ReturnType<typeof configureStore>;

  beforeEach(() => {
    jest.clearAllMocks();
    Date.now = jest.fn(() => mockNow);
    store = createStore();
  });

  afterEach(() => {
    Date.now = originalDateNow;
  });

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const state = store.getState().ui;
      
      expect(state).toEqual({
        theme: 'dark',
        sidebarOpen: true,
        modals: {},
        notifications: [],
        shortcuts: {
          enabled: true,
          customBindings: {},
        },
      });
    });
  });

  describe('theme management', () => {
    describe('setTheme', () => {
      it('should set theme to light', () => {
        store.dispatch(setTheme('light'));

        const state = store.getState().ui;
        expect(state.theme).toBe('light');
        expect(mockSetItem).toHaveBeenCalledWith('theme', 'light');
      });

      it('should set theme to dark', () => {
        store.dispatch(setTheme('light'));
        store.dispatch(setTheme('dark'));

        const state = store.getState().ui;
        expect(state.theme).toBe('dark');
        expect(mockSetItem).toHaveBeenLastCalledWith('theme', 'dark');
      });
    });

    describe('toggleTheme', () => {
      it('should toggle from dark to light', () => {
        store.dispatch(toggleTheme());

        const state = store.getState().ui;
        expect(state.theme).toBe('light');
        expect(mockSetItem).toHaveBeenCalledWith('theme', 'light');
      });

      it('should toggle from light to dark', () => {
        store.dispatch(setTheme('light'));
        mockSetItem.mockClear();
        
        store.dispatch(toggleTheme());

        const state = store.getState().ui;
        expect(state.theme).toBe('dark');
        expect(mockSetItem).toHaveBeenCalledWith('theme', 'dark');
      });

      it('should toggle multiple times', () => {
        store.dispatch(toggleTheme()); // dark -> light
        store.dispatch(toggleTheme()); // light -> dark
        store.dispatch(toggleTheme()); // dark -> light

        const state = store.getState().ui;
        expect(state.theme).toBe('light');
      });
    });
  });

  describe('sidebar management', () => {
    describe('toggleSidebar', () => {
      it('should toggle sidebar from open to closed', () => {
        store.dispatch(toggleSidebar());

        const state = store.getState().ui;
        expect(state.sidebarOpen).toBe(false);
      });

      it('should toggle sidebar from closed to open', () => {
        store.dispatch(toggleSidebar());
        store.dispatch(toggleSidebar());

        const state = store.getState().ui;
        expect(state.sidebarOpen).toBe(true);
      });
    });

    describe('setSidebarOpen', () => {
      it('should set sidebar open', () => {
        store.dispatch(setSidebarOpen(false));
        store.dispatch(setSidebarOpen(true));

        const state = store.getState().ui;
        expect(state.sidebarOpen).toBe(true);
      });

      it('should set sidebar closed', () => {
        store.dispatch(setSidebarOpen(false));

        const state = store.getState().ui;
        expect(state.sidebarOpen).toBe(false);
      });
    });
  });

  describe('modal management', () => {
    describe('openModal', () => {
      it('should open a modal without data', () => {
        store.dispatch(openModal({ key: 'characterCreate' }));

        const state = store.getState().ui;
        expect(state.modals.characterCreate).toEqual({
          isOpen: true,
          data: undefined,
        });
      });

      it('should open a modal with data', () => {
        const modalData = { characterId: 'char-123', mode: 'edit' };
        store.dispatch(openModal({ key: 'characterEdit', data: modalData }));

        const state = store.getState().ui;
        expect(state.modals.characterEdit).toEqual({
          isOpen: true,
          data: modalData,
        });
      });

      it('should open multiple modals', () => {
        store.dispatch(openModal({ key: 'modal1' }));
        store.dispatch(openModal({ key: 'modal2', data: { test: true } }));

        const state = store.getState().ui;
        expect(Object.keys(state.modals)).toHaveLength(2);
        expect(state.modals.modal1.isOpen).toBe(true);
        expect(state.modals.modal2.isOpen).toBe(true);
      });

      it('should update existing modal', () => {
        store.dispatch(openModal({ key: 'testModal', data: { version: 1 } }));
        store.dispatch(openModal({ key: 'testModal', data: { version: 2 } }));

        const state = store.getState().ui;
        expect(state.modals.testModal.data.version).toBe(2);
      });
    });

    describe('closeModal', () => {
      beforeEach(() => {
        store.dispatch(openModal({ key: 'modal1' }));
        store.dispatch(openModal({ key: 'modal2' }));
      });

      it('should close specific modal', () => {
        store.dispatch(closeModal('modal1'));

        const state = store.getState().ui;
        expect(state.modals.modal1.isOpen).toBe(false);
        expect(state.modals.modal2.isOpen).toBe(true);
      });

      it('should handle closing non-existent modal', () => {
        store.dispatch(closeModal('nonExistent'));

        const state = store.getState().ui;
        expect(state.modals.nonExistent).toBeUndefined();
      });

      it('should preserve modal data when closing', () => {
        const data = { important: 'data' };
        store.dispatch(openModal({ key: 'dataModal', data }));
        store.dispatch(closeModal('dataModal'));

        const state = store.getState().ui;
        expect(state.modals.dataModal.isOpen).toBe(false);
        expect(state.modals.dataModal.data).toEqual(data);
      });
    });

    describe('closeAllModals', () => {
      it('should close all open modals', () => {
        store.dispatch(openModal({ key: 'modal1' }));
        store.dispatch(openModal({ key: 'modal2' }));
        store.dispatch(openModal({ key: 'modal3' }));

        store.dispatch(closeAllModals());

        const state = store.getState().ui;
        expect(state.modals.modal1.isOpen).toBe(false);
        expect(state.modals.modal2.isOpen).toBe(false);
        expect(state.modals.modal3.isOpen).toBe(false);
      });

      it('should handle empty modals', () => {
        store.dispatch(closeAllModals());

        const state = store.getState().ui;
        expect(state.modals).toEqual({});
      });
    });
  });

  describe('notification management', () => {
    describe('addNotification', () => {
      it('should add info notification', () => {
        store.dispatch(addNotification({
          type: 'info',
          message: 'Character saved successfully',
        }));

        const state = store.getState().ui;
        expect(state.notifications).toHaveLength(1);
        expect(state.notifications[0]).toEqual({
          id: expect.stringContaining('notification-'),
          type: 'info',
          message: 'Character saved successfully',
          timestamp: mockNow,
        });
      });

      it('should add notification with duration', () => {
        store.dispatch(addNotification({
          type: 'success',
          message: 'Level up!',
          duration: 5000,
        }));

        const state = store.getState().ui;
        expect(state.notifications[0].duration).toBe(5000);
      });

      it('should add different notification types', () => {
        const notificationTypes: Array<'info' | 'success' | 'warning' | 'error'> = 
          ['info', 'success', 'warning', 'error'];
        
        for (const type of notificationTypes) {
          store.dispatch(addNotification({
            type,
            message: `${type} notification`,
          }));
        }

        const state = store.getState().ui;
        expect(state.notifications).toHaveLength(4);
        for (let i = 0; i < state.notifications.length; i++) {
          expect(state.notifications[i].type).toBe(notificationTypes[i]);
        }
      });

      it('should limit notifications to 10', () => {
        // Add 12 notifications
        addMultipleNotifications(store, 12);

        const state = store.getState().ui;
        expect(state.notifications).toHaveLength(10);
        // First two should be removed
        expect(state.notifications[0].message).toBe('Notification 2');
        expect(state.notifications[9].message).toBe('Notification 11');
      });
    });

    describe('removeNotification', () => {
      beforeEach(() => {
        setupNotifications(store);
      });

      it('should remove specific notification', () => {
        const state = store.getState().ui;
        const idToRemove = state.notifications[1].id;

        store.dispatch(removeNotification(idToRemove));

        const newState = store.getState().ui;
        expect(newState.notifications).toHaveLength(2);
        const found = newState.notifications.find(n => n.id === idToRemove);
        expect(found).toBeUndefined();
      });

      it('should handle removing non-existent notification', () => {
        store.dispatch(removeNotification('non-existent-id'));

        const state = store.getState().ui;
        expect(state.notifications).toHaveLength(3);
      });
    });

    describe('clearNotifications', () => {
      it('should clear all notifications', () => {
        // Add notifications
        store.dispatch(addNotification({ type: 'info', message: 'Test 1' }));
        store.dispatch(addNotification({ type: 'error', message: 'Test 2' }));

        store.dispatch(clearNotifications());

        const state = store.getState().ui;
        expect(state.notifications).toEqual([]);
      });

      it('should handle clearing empty notifications', () => {
        store.dispatch(clearNotifications());

        const state = store.getState().ui;
        expect(state.notifications).toEqual([]);
      });
    });
  });

  describe('keyboard shortcuts', () => {
    describe('setShortcutsEnabled', () => {
      it('should enable shortcuts', () => {
        store.dispatch(setShortcutsEnabled(false));
        store.dispatch(setShortcutsEnabled(true));

        const state = store.getState().ui;
        expect(state.shortcuts.enabled).toBe(true);
      });

      it('should disable shortcuts', () => {
        store.dispatch(setShortcutsEnabled(false));

        const state = store.getState().ui;
        expect(state.shortcuts.enabled).toBe(false);
      });
    });

    describe('setCustomBinding', () => {
      it('should set custom key binding', () => {
        store.dispatch(setCustomBinding({ 
          action: 'rollDice', 
          binding: 'ctrl+d' 
        }));

        const state = store.getState().ui;
        expect(state.shortcuts.customBindings.rollDice).toBe('ctrl+d');
      });

      it('should override existing binding', () => {
        store.dispatch(setCustomBinding({ action: 'save', binding: 'ctrl+s' }));
        store.dispatch(setCustomBinding({ action: 'save', binding: 'cmd+s' }));

        const state = store.getState().ui;
        expect(state.shortcuts.customBindings.save).toBe('cmd+s');
      });

      it('should set multiple custom bindings', () => {
        store.dispatch(setCustomBinding({ action: 'undo', binding: 'ctrl+z' }));
        store.dispatch(setCustomBinding({ action: 'redo', binding: 'ctrl+y' }));
        store.dispatch(setCustomBinding({ action: 'search', binding: 'ctrl+f' }));

        const state = store.getState().ui;
        expect(Object.keys(state.shortcuts.customBindings)).toHaveLength(3);
        expect(state.shortcuts.customBindings.undo).toBe('ctrl+z');
        expect(state.shortcuts.customBindings.redo).toBe('ctrl+y');
        expect(state.shortcuts.customBindings.search).toBe('ctrl+f');
      });
    });

    describe('resetCustomBindings', () => {
      it('should reset all custom bindings', () => {
        // Set some bindings
        store.dispatch(setCustomBinding({ action: 'action1', binding: 'key1' }));
        store.dispatch(setCustomBinding({ action: 'action2', binding: 'key2' }));

        store.dispatch(resetCustomBindings());

        const state = store.getState().ui;
        expect(state.shortcuts.customBindings).toEqual({});
      });
    });
  });

  describe('updateUISettings', () => {
    it('should update multiple settings at once', () => {
      store.dispatch(updateUISettings({
        theme: 'light',
        sidebarOpen: false,
      }));

      const state = store.getState().ui;
      expect(state.theme).toBe('light');
      expect(state.sidebarOpen).toBe(false);
    });

    it('should update nested settings', () => {
      store.dispatch(updateUISettings({
        shortcuts: {
          enabled: false,
          customBindings: { test: 'binding' },
        },
      }));

      const state = store.getState().ui;
      expect(state.shortcuts.enabled).toBe(false);
      expect(state.shortcuts.customBindings.test).toBe('binding');
    });

    it('should preserve unspecified settings', () => {
      // Set some initial state
      store.dispatch(setTheme('light'));
      store.dispatch(addNotification({ type: 'info', message: 'Test' }));

      store.dispatch(updateUISettings({
        sidebarOpen: false,
      }));

      const state = store.getState().ui;
      expect(state.theme).toBe('light');
      expect(state.notifications).toHaveLength(1);
      expect(state.sidebarOpen).toBe(false);
    });
  });

  describe('complex scenarios', () => {
    it('should handle modal and notification workflow', () => {
      // Open character creation modal
      store.dispatch(openModal({ key: 'characterCreate' }));
      
      // User creates character - show success notification
      store.dispatch(addNotification({
        type: 'success',
        message: 'Character created successfully!',
        duration: 3000,
      }));
      
      // Close modal
      store.dispatch(closeModal('characterCreate'));
      
      // Open character edit modal with data
      store.dispatch(openModal({ 
        key: 'characterEdit', 
        data: { characterId: 'char-123' } 
      }));

      const state = store.getState().ui;
      expect(state.modals.characterCreate.isOpen).toBe(false);
      expect(state.modals.characterEdit.isOpen).toBe(true);
      expect(state.modals.characterEdit.data.characterId).toBe('char-123');
      expect(state.notifications).toHaveLength(1);
    });

    it('should handle theme preference workflow', () => {
      // User prefers light theme
      store.dispatch(setTheme('light'));
      
      // User experiments with dark theme
      store.dispatch(toggleTheme());
      expect(store.getState().ui.theme).toBe('dark');
      
      // User goes back to light
      store.dispatch(toggleTheme());
      expect(store.getState().ui.theme).toBe('light');
      
      // Verify localStorage was updated each time
      expect(mockSetItem).toHaveBeenCalledTimes(3);
    });

    it('should handle keyboard shortcut customization', () => {
      // User customizes shortcuts
      store.dispatch(setCustomBinding({ action: 'quickSave', binding: 'ctrl+q' }));
      store.dispatch(setCustomBinding({ action: 'toggleCombat', binding: 'ctrl+c' }));
      
      // User disables shortcuts temporarily
      store.dispatch(setShortcutsEnabled(false));
      
      let state = store.getState().ui;
      expect(state.shortcuts.enabled).toBe(false);
      expect(state.shortcuts.customBindings.quickSave).toBe('ctrl+q');
      
      // User re-enables and resets bindings
      store.dispatch(setShortcutsEnabled(true));
      store.dispatch(resetCustomBindings());
      
      state = store.getState().ui;
      expect(state.shortcuts.enabled).toBe(true);
      expect(state.shortcuts.customBindings).toEqual({});
    });

    it('should handle notification overflow scenario', () => {
      // Simulate rapid notifications (e.g., during combat)
      const messages = [
        'Goblin attacks!',
        'Critical hit!',
        'You take 5 damage',
        'Healing potion used',
        'Level up!',
        'New spell learned',
        'Quest completed',
        'Item found',
        'Skill check passed',
        'Initiative rolled',
        'Combat started',
        'Enemy defeated',
      ];

      for (let i = 0; i < messages.length; i++) {
        const message = messages[i];
        store.dispatch(addNotification({
          type: i % 2 === 0 ? 'info' : 'success',
          message,
        }));
      }

      const state = store.getState().ui;
      // Should only keep last 10
      expect(state.notifications).toHaveLength(10);
      // First 2 notifications should have been removed
      expect(state.notifications[0].message).toBe('You take 5 damage');
      expect(state.notifications[9].message).toBe('Enemy defeated');
    });

    it('should maintain complete UI state', () => {
      // Set various UI states
      store.dispatch(setTheme('light'));
      store.dispatch(setSidebarOpen(false));
      store.dispatch(openModal({ key: 'settings', data: { tab: 'general' } }));
      store.dispatch(openModal({ key: 'help' }));
      store.dispatch(addNotification({ type: 'info', message: 'Welcome!' }));
      store.dispatch(setShortcutsEnabled(false));
      store.dispatch(setCustomBinding({ action: 'menu', binding: 'esc' }));

      const state = store.getState().ui;
      
      // Verify all states
      expect(state.theme).toBe('light');
      expect(state.sidebarOpen).toBe(false);
      expect(state.modals.settings.isOpen).toBe(true);
      expect(state.modals.settings.data.tab).toBe('general');
      expect(state.modals.help.isOpen).toBe(true);
      expect(state.notifications).toHaveLength(1);
      expect(state.shortcuts.enabled).toBe(false);
      expect(state.shortcuts.customBindings.menu).toBe('esc');
    });
  });
});