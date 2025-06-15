import { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { logout } from '../store/slices/authSlice';

/**
 * Sync authentication state across multiple tabs by listening
 * for localStorage changes and dispatching logout when the
 * access token is removed in another tab.
 */
export function useAuthSync() {
  const dispatch = useDispatch();

  useEffect(() => {
    const handleStorage = (e: StorageEvent) => {
      if (e.key === 'access_token' && e.oldValue && !e.newValue) {
        // Token was removed in another tab -> log out here as well
        dispatch(logout());
      }
    };
    window.addEventListener('storage', handleStorage);
    return () => window.removeEventListener('storage', handleStorage);
  }, [dispatch]);
}
