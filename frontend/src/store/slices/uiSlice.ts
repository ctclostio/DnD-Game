import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { UIState, Notification } from '../../types/state';

const initialState: UIState = {
  theme: 'dark',
  sidebarOpen: true,
  modals: {},
  notifications: [],
  shortcuts: {
    enabled: true,
    customBindings: {},
  },
};

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    // Theme management
    setTheme: (state, action: PayloadAction<'light' | 'dark'>) => {
      state.theme = action.payload;
      // Persist to localStorage
      localStorage.setItem('theme', action.payload);
    },
    
    toggleTheme: (state) => {
      state.theme = state.theme === 'light' ? 'dark' : 'light';
      localStorage.setItem('theme', state.theme);
    },
    
    // Sidebar management
    toggleSidebar: (state) => {
      state.sidebarOpen = !state.sidebarOpen;
    },
    
    setSidebarOpen: (state, action: PayloadAction<boolean>) => {
      state.sidebarOpen = action.payload;
    },
    
    // Modal management
    openModal: (state, action: PayloadAction<{ key: string; data?: Record<string, unknown> }>) => {
      const { key, data } = action.payload;
      state.modals[key] = {
        isOpen: true,
        data,
      };
    },
    
    closeModal: (state, action: PayloadAction<string>) => {
      const key = action.payload;
      if (state.modals[key]) {
        state.modals[key].isOpen = false;
      }
    },
    
    closeAllModals: (state) => {
      Object.keys(state.modals).forEach(key => {
        state.modals[key].isOpen = false;
      });
    },
    
    // Notification management
    addNotification: (state, action: PayloadAction<Omit<Notification, 'id' | 'timestamp'>>) => {
      const notification: Notification = {
        ...action.payload,
        id: `notification-${Date.now()}-${Math.random()}`,
        timestamp: Date.now(),
      };
      state.notifications.push(notification);
      
      // Limit notifications to prevent memory issues
      if (state.notifications.length > 10) {
        state.notifications.shift();
      }
    },
    
    removeNotification: (state, action: PayloadAction<string>) => {
      state.notifications = state.notifications.filter(n => n.id !== action.payload);
    },
    
    clearNotifications: (state) => {
      state.notifications = [];
    },
    
    // Keyboard shortcuts
    setShortcutsEnabled: (state, action: PayloadAction<boolean>) => {
      state.shortcuts.enabled = action.payload;
    },
    
    setCustomBinding: (state, action: PayloadAction<{ action: string; binding: string }>) => {
      const { action: actionName, binding } = action.payload;
      state.shortcuts.customBindings[actionName] = binding;
    },
    
    resetCustomBindings: (state) => {
      state.shortcuts.customBindings = {};
    },
    
    // Bulk updates
    updateUISettings: (state, action: PayloadAction<Partial<UIState>>) => {
      return { ...state, ...action.payload };
    },
  },
});

export const {
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
} = uiSlice.actions;

export default uiSlice.reducer;
