import storage from 'redux-persist/lib/storage';
import { PersistConfig } from 'redux-persist';
import { AuthState, CharacterState, UIState, GameDataState, RootState } from '../types/state';

// Persist configurations for different slices
export const authPersistConfig: PersistConfig<AuthState> = {
  key: 'auth',
  storage,
  whitelist: ['user', 'token'], // Only persist user and token
};

export const characterPersistConfig: PersistConfig<CharacterState> = {
  key: 'characters',
  storage,
  whitelist: ['characters', 'currentCharacterId'],
};

export const uiPersistConfig: PersistConfig<UIState> = {
  key: 'ui',
  storage,
  whitelist: ['theme', 'shortcuts', 'sidebarOpen'],
};

export const gameDataPersistConfig: PersistConfig<GameDataState> = {
  key: 'gameData',
  storage,
  whitelist: ['spells', 'equipment', 'classes', 'races'],
  // Game data can be large, so we might want to use a different storage strategy
};

// Combat state should NOT be persisted to avoid inconsistencies
// WebSocket state should NOT be persisted as connections are ephemeral

export const rootPersistConfig: PersistConfig<RootState> = {
  key: 'root',
  storage,
  blacklist: ['combat', 'websocket', 'dmTools'], // These will not be persisted
};