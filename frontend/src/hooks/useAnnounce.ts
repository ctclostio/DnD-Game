import { useCallback } from 'react';
import { useAccessibility } from '../components/AccessibilityProvider';

interface UseAnnounceReturn {
  announce: (message: string, priority?: 'polite' | 'assertive') => void;
  announceError: (message: string) => void;
  announceSuccess: (message: string) => void;
  announceLoading: (isLoading: boolean, loadingMessage?: string, completeMessage?: string) => void;
}

export function useAnnounce(): UseAnnounceReturn {
  const { announceToScreenReader } = useAccessibility();

  const announce = useCallback((message: string, priority?: 'polite' | 'assertive') => {
    announceToScreenReader(message, priority);
  }, [announceToScreenReader]);

  const announceError = useCallback((message: string) => {
    announceToScreenReader(`Error: ${message}`, 'assertive');
  }, [announceToScreenReader]);

  const announceSuccess = useCallback((message: string) => {
    announceToScreenReader(`Success: ${message}`, 'polite');
  }, [announceToScreenReader]);

  const announceLoading = useCallback((
    isLoading: boolean,
    loadingMessage: string = 'Loading content',
    completeMessage: string = 'Content loaded'
  ) => {
    if (isLoading) {
      announceToScreenReader(loadingMessage, 'polite');
    } else {
      announceToScreenReader(completeMessage, 'polite');
    }
  }, [announceToScreenReader]);

  return {
    announce,
    announceError,
    announceSuccess,
    announceLoading,
  };
}
