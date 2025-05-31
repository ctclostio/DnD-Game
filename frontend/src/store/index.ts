import { configureStore, combineReducers } from '@reduxjs/toolkit';
import { persistStore, persistReducer, FLUSH, REHYDRATE, PAUSE, PERSIST, PURGE, REGISTER } from 'redux-persist';
import storage from 'redux-persist/lib/storage';

// Import slices
import authReducer from './slices/authSlice';
import characterReducer from './slices/characterSlice';
import gameSessionReducer from './slices/gameSessionSlice';
import combatReducer from './slices/combatSlice';
import uiReducer from './slices/uiSlice';
import dmToolsReducer from './slices/dmToolsSlice';
import websocketReducer from './slices/websocketSlice';

// Import middleware
import { undoMiddleware } from './middleware/undoMiddleware';
import { websocketMiddleware } from './middleware/websocketMiddleware';

// Import persist configs
import { authPersistConfig, characterPersistConfig, uiPersistConfig } from './config';

// Create persisted reducers
const persistedAuthReducer = persistReducer(authPersistConfig, authReducer);
const persistedCharacterReducer = persistReducer(characterPersistConfig, characterReducer);
const persistedUIReducer = persistReducer(uiPersistConfig, uiReducer);

// Root reducer
const rootReducer = combineReducers({
  auth: persistedAuthReducer,
  character: persistedCharacterReducer,
  gameSession: gameSessionReducer,
  combat: combatReducer,
  ui: persistedUIReducer,
  dmTools: dmToolsReducer,
  websocket: websocketReducer,
});

// Configure store
export const store = configureStore({
  reducer: rootReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types from redux-persist
        ignoredActions: [FLUSH, REHYDRATE, PAUSE, PERSIST, PURGE, REGISTER],
        // Ignore these paths in the state
        ignoredPaths: ['dmTools.undoStack', 'dmTools.redoStack'],
      },
    })
    .concat(undoMiddleware)
    .concat(websocketMiddleware),
  devTools: process.env.NODE_ENV !== 'production' && {
    name: 'D&D Online Game',
    trace: true,
    traceLimit: 25,
  },
});

// Create persistor
export const persistor = persistStore(store);

// Export types
export type RootState = ReturnType<typeof rootReducer>;
export type AppDispatch = typeof store.dispatch;

// Type-safe hooks
import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux';

export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;