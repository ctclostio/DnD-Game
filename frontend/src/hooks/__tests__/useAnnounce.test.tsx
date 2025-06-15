import React from 'react';
import { renderHook, act } from '@testing-library/react';
import { useAnnounce } from '../useAnnounce';

// Mock the AccessibilityProvider context
const mockAnnounceToScreenReader = jest.fn();

jest.mock('../../components/AccessibilityProvider', () => ({
  useAccessibility: () => ({
    announceToScreenReader: mockAnnounceToScreenReader,
    settings: {
      highContrast: false,
      largeText: false,
      reduceMotion: false,
      screenReaderMode: false,
      keyboardNavigation: true,
      focusIndicator: true,
    },
    updateSetting: jest.fn(),
  }),
}));

describe('useAnnounce', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should provide announce functions', () => {
    const { result } = renderHook(() => useAnnounce());

    expect(result.current.announce).toBeDefined();
    expect(result.current.announceError).toBeDefined();
    expect(result.current.announceSuccess).toBeDefined();
    expect(result.current.announceLoading).toBeDefined();
    expect(typeof result.current.announce).toBe('function');
    expect(typeof result.current.announceError).toBe('function');
    expect(typeof result.current.announceSuccess).toBe('function');
    expect(typeof result.current.announceLoading).toBe('function');
  });

  describe('announce', () => {
    it('should announce message with default priority', () => {
      const { result } = renderHook(() => useAnnounce());

      act(() => {
        result.current.announce('Test message');
      });

      expect(mockAnnounceToScreenReader).toHaveBeenCalledWith('Test message', undefined);
    });

    it('should announce message with polite priority', () => {
      const { result } = renderHook(() => useAnnounce());

      act(() => {
        result.current.announce('Polite message', 'polite');
      });

      expect(mockAnnounceToScreenReader).toHaveBeenCalledWith('Polite message', 'polite');
    });

    it('should announce message with assertive priority', () => {
      const { result } = renderHook(() => useAnnounce());

      act(() => {
        result.current.announce('Urgent message', 'assertive');
      });

      expect(mockAnnounceToScreenReader).toHaveBeenCalledWith('Urgent message', 'assertive');
    });
  });

  describe('announceError', () => {
    it('should announce error with assertive priority', () => {
      const { result } = renderHook(() => useAnnounce());

      act(() => {
        result.current.announceError('Something went wrong');
      });

      expect(mockAnnounceToScreenReader).toHaveBeenCalledWith(
        'Error: Something went wrong',
        'assertive'
      );
    });

    it('should prefix error messages', () => {
      const { result } = renderHook(() => useAnnounce());

      act(() => {
        result.current.announceError('Network connection failed');
      });

      expect(mockAnnounceToScreenReader).toHaveBeenCalledWith(
        'Error: Network connection failed',
        'assertive'
      );
    });
  });

  describe('announceSuccess', () => {
    it('should announce success with polite priority', () => {
      const { result } = renderHook(() => useAnnounce());

      act(() => {
        result.current.announceSuccess('Operation completed');
      });

      expect(mockAnnounceToScreenReader).toHaveBeenCalledWith(
        'Success: Operation completed',
        'polite'
      );
    });

    it('should prefix success messages', () => {
      const { result } = renderHook(() => useAnnounce());

      act(() => {
        result.current.announceSuccess('File uploaded');
      });

      expect(mockAnnounceToScreenReader).toHaveBeenCalledWith(
        'Success: File uploaded',
        'polite'
      );
    });
  });

  describe('announceLoading', () => {
    it('should announce loading state with default message', () => {
      const { result } = renderHook(() => useAnnounce());

      act(() => {
        result.current.announceLoading(true);
      });

      expect(mockAnnounceToScreenReader).toHaveBeenCalledWith(
        'Loading content',
        'polite'
      );
    });

    it('should announce loading complete with default message', () => {
      const { result } = renderHook(() => useAnnounce());

      act(() => {
        result.current.announceLoading(false);
      });

      expect(mockAnnounceToScreenReader).toHaveBeenCalledWith(
        'Content loaded',
        'polite'
      );
    });

    it('should use custom loading message', () => {
      const { result } = renderHook(() => useAnnounce());

      act(() => {
        result.current.announceLoading(true, 'Fetching user data');
      });

      expect(mockAnnounceToScreenReader).toHaveBeenCalledWith(
        'Fetching user data',
        'polite'
      );
    });

    it('should use custom complete message', () => {
      const { result } = renderHook(() => useAnnounce());

      act(() => {
        result.current.announceLoading(false, 'Loading...', 'User data loaded successfully');
      });

      expect(mockAnnounceToScreenReader).toHaveBeenCalledWith(
        'User data loaded successfully',
        'polite'
      );
    });

    it('should handle loading state transitions', () => {
      const { result } = renderHook(() => useAnnounce());

      // Start loading
      act(() => {
        result.current.announceLoading(true, 'Processing request');
      });

      expect(mockAnnounceToScreenReader).toHaveBeenLastCalledWith(
        'Processing request',
        'polite'
      );

      // Complete loading
      act(() => {
        result.current.announceLoading(false, 'Processing request', 'Request completed');
      });

      expect(mockAnnounceToScreenReader).toHaveBeenLastCalledWith(
        'Request completed',
        'polite'
      );

      expect(mockAnnounceToScreenReader).toHaveBeenCalledTimes(2);
    });
  });

  it('should maintain stable function references', () => {
    const { result, rerender } = renderHook(() => useAnnounce());

    const { announce, announceError, announceSuccess, announceLoading } = result.current;

    rerender();

    expect(result.current.announce).toBe(announce);
    expect(result.current.announceError).toBe(announceError);
    expect(result.current.announceSuccess).toBe(announceSuccess);
    expect(result.current.announceLoading).toBe(announceLoading);
  });

  it('should handle rapid announcements', () => {
    const { result } = renderHook(() => useAnnounce());

    act(() => {
      result.current.announce('Message 1');
      result.current.announceError('Error 1');
      result.current.announceSuccess('Success 1');
      result.current.announceLoading(true);
      result.current.announceLoading(false);
    });

    expect(mockAnnounceToScreenReader).toHaveBeenCalledTimes(5);
    expect(mockAnnounceToScreenReader).toHaveBeenNthCalledWith(1, 'Message 1', undefined);
    expect(mockAnnounceToScreenReader).toHaveBeenNthCalledWith(2, 'Error: Error 1', 'assertive');
    expect(mockAnnounceToScreenReader).toHaveBeenNthCalledWith(3, 'Success: Success 1', 'polite');
    expect(mockAnnounceToScreenReader).toHaveBeenNthCalledWith(4, 'Loading content', 'polite');
    expect(mockAnnounceToScreenReader).toHaveBeenNthCalledWith(5, 'Content loaded', 'polite');
  });

  it('should handle empty messages', () => {
    const { result } = renderHook(() => useAnnounce());

    act(() => {
      result.current.announce('');
      result.current.announceError('');
      result.current.announceSuccess('');
    });

    expect(mockAnnounceToScreenReader).toHaveBeenCalledWith('', undefined);
    expect(mockAnnounceToScreenReader).toHaveBeenCalledWith('Error: ', 'assertive');
    expect(mockAnnounceToScreenReader).toHaveBeenCalledWith('Success: ', 'polite');
  });

  it('should handle special characters in messages', () => {
    const { result } = renderHook(() => useAnnounce());

    const specialMessage = 'Test <script>alert("xss")</script> & special chars';

    act(() => {
      result.current.announce(specialMessage);
    });

    expect(mockAnnounceToScreenReader).toHaveBeenCalledWith(specialMessage, undefined);
  });
});

// Test error handling when context is not provided
describe('useAnnounce - Error Handling', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.resetModules();
  });

  it.skip('should throw error when used outside AccessibilityProvider', () => {
    // This test is skipped because testing the error from useContext is complicated
    // in a test environment. The AccessibilityProvider itself tests this behavior.
  });
});
