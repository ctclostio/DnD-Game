import { renderHook, act } from '@testing-library/react';
import { useDebounce, useDebouncedCallback, useDebouncedSearch } from '../useDebounce';

// Mock timers
jest.useFakeTimers();

describe('useDebounce', () => {
  afterEach(() => {
    jest.clearAllTimers();
  });

  it('should return initial value immediately', () => {
    const { result } = renderHook(() => useDebounce('initial', 500));
    expect(result.current).toBe('initial');
  });

  it('should debounce value changes', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      {
        initialProps: { value: 'initial', delay: 500 },
      }
    );

    expect(result.current).toBe('initial');

    // Update value
    rerender({ value: 'updated', delay: 500 });
    
    // Value should not change immediately
    expect(result.current).toBe('initial');

    // Fast forward time
    act(() => {
      jest.advanceTimersByTime(499);
    });
    expect(result.current).toBe('initial');

    // After delay, value should update
    act(() => {
      jest.advanceTimersByTime(1);
    });
    expect(result.current).toBe('updated');
  });

  it('should cancel previous timeout on rapid changes', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      {
        initialProps: { value: 'initial', delay: 300 },
      }
    );

    // Make rapid changes
    rerender({ value: 'change1', delay: 300 });
    act(() => {
      jest.advanceTimersByTime(100);
    });
    
    rerender({ value: 'change2', delay: 300 });
    act(() => {
      jest.advanceTimersByTime(100);
    });
    
    rerender({ value: 'change3', delay: 300 });
    
    // Value should still be initial
    expect(result.current).toBe('initial');

    // Fast forward to complete the last debounce
    act(() => {
      jest.advanceTimersByTime(300);
    });
    
    // Should only have the last value
    expect(result.current).toBe('change3');
  });

  it('should handle different data types', () => {
    // Test with object
    const { result: objectResult } = renderHook(() => 
      useDebounce({ name: 'test', value: 42 }, 100)
    );
    
    expect(objectResult.current).toEqual({ name: 'test', value: 42 });

    // Test with array
    const { result: arrayResult } = renderHook(() => 
      useDebounce([1, 2, 3], 100)
    );
    
    expect(arrayResult.current).toEqual([1, 2, 3]);

    // Test with number
    const { result: numberResult } = renderHook(() => 
      useDebounce(123, 100)
    );
    
    expect(numberResult.current).toBe(123);
  });

  it('should handle delay changes', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      {
        initialProps: { value: 'initial', delay: 200 },
      }
    );

    rerender({ value: 'updated', delay: 500 });
    
    // Advance time based on original delay
    act(() => {
      jest.advanceTimersByTime(200);
    });
    
    // Value should not update yet due to new delay
    expect(result.current).toBe('initial');

    // Advance remaining time
    act(() => {
      jest.advanceTimersByTime(300);
    });
    
    expect(result.current).toBe('updated');
  });
});

describe('useDebouncedCallback', () => {
  afterEach(() => {
    jest.clearAllTimers();
  });

  it('should debounce callback execution', () => {
    const mockCallback = jest.fn();
    const { result } = renderHook(() => useDebouncedCallback(mockCallback, 300));

    // Call the debounced function multiple times
    act(() => {
      result.current('arg1');
      result.current('arg2');
      result.current('arg3');
    });

    // Callback should not be called immediately
    expect(mockCallback).not.toHaveBeenCalled();

    // Fast forward time
    act(() => {
      jest.advanceTimersByTime(300);
    });

    // Callback should be called once with last arguments
    expect(mockCallback).toHaveBeenCalledTimes(1);
    expect(mockCallback).toHaveBeenCalledWith('arg3');
  });

  it('should handle callback updates', () => {
    const mockCallback1 = jest.fn();
    const mockCallback2 = jest.fn();
    
    const { result, rerender } = renderHook(
      ({ callback, delay }) => useDebouncedCallback(callback, delay),
      {
        initialProps: { callback: mockCallback1, delay: 200 },
      }
    );

    // Call with first callback
    act(() => {
      result.current('test');
    });

    // Update to second callback
    rerender({ callback: mockCallback2, delay: 200 });

    // Fast forward time
    act(() => {
      jest.advanceTimersByTime(200);
    });

    // Second callback should be called, not the first
    expect(mockCallback1).not.toHaveBeenCalled();
    expect(mockCallback2).toHaveBeenCalledWith('test');
  });

  it('should cleanup timeout on unmount', () => {
    const mockCallback = jest.fn();
    const { result, unmount } = renderHook(() => 
      useDebouncedCallback(mockCallback, 300)
    );

    // Call the debounced function
    act(() => {
      result.current('test');
    });

    // Unmount before timeout completes
    unmount();

    // Fast forward time
    act(() => {
      jest.advanceTimersByTime(300);
    });

    // Callback should not be called after unmount
    expect(mockCallback).not.toHaveBeenCalled();
  });

  it('should preserve this context', () => {
    const obj = {
      value: 'test',
      method: jest.fn(function(this: any, arg: string) {
        return `${this.value}-${arg}`;
      }),
    };

    const { result } = renderHook(() => 
      useDebouncedCallback(obj.method.bind(obj), 100)
    );

    act(() => {
      result.current('arg');
    });

    act(() => {
      jest.advanceTimersByTime(100);
    });

    expect(obj.method).toHaveBeenCalledWith('arg');
  });
});

describe('useDebouncedSearch', () => {
  afterEach(() => {
    jest.clearAllTimers();
  });

  it('should initialize with default values', () => {
    const { result } = renderHook(() => useDebouncedSearch());

    expect(result.current.searchTerm).toBe('');
    expect(result.current.debouncedSearchTerm).toBe('');
    expect(typeof result.current.setSearchTerm).toBe('function');
    expect(typeof result.current.clearSearch).toBe('function');
  });

  it('should initialize with custom values', () => {
    const { result } = renderHook(() => useDebouncedSearch('initial search', 500));

    expect(result.current.searchTerm).toBe('initial search');
    expect(result.current.debouncedSearchTerm).toBe('initial search');
  });

  it('should update search term immediately but debounce the result', () => {
    const { result } = renderHook(() => useDebouncedSearch('', 300));

    // Update search term
    act(() => {
      result.current.setSearchTerm('test search');
    });

    // Search term updates immediately
    expect(result.current.searchTerm).toBe('test search');
    // Debounced term hasn't updated yet
    expect(result.current.debouncedSearchTerm).toBe('');

    // Fast forward time
    act(() => {
      jest.advanceTimersByTime(300);
    });

    // Now debounced term should be updated
    expect(result.current.debouncedSearchTerm).toBe('test search');
  });

  it('should clear search', () => {
    const { result } = renderHook(() => useDebouncedSearch('existing search', 200));

    expect(result.current.searchTerm).toBe('existing search');
    expect(result.current.debouncedSearchTerm).toBe('existing search');

    // Clear search
    act(() => {
      result.current.clearSearch();
    });

    expect(result.current.searchTerm).toBe('');

    // Fast forward time
    act(() => {
      jest.advanceTimersByTime(200);
    });

    expect(result.current.debouncedSearchTerm).toBe('');
  });

  it('should handle rapid search updates', () => {
    const { result } = renderHook(() => useDebouncedSearch('', 250));

    // Simulate rapid typing
    act(() => {
      result.current.setSearchTerm('t');
    });
    act(() => {
      jest.advanceTimersByTime(50);
      result.current.setSearchTerm('te');
    });
    act(() => {
      jest.advanceTimersByTime(50);
      result.current.setSearchTerm('tes');
    });
    act(() => {
      jest.advanceTimersByTime(50);
      result.current.setSearchTerm('test');
    });

    // Debounced value should still be empty
    expect(result.current.debouncedSearchTerm).toBe('');

    // Complete the debounce
    act(() => {
      jest.advanceTimersByTime(250);
    });

    // Should only have the final value
    expect(result.current.debouncedSearchTerm).toBe('test');
  });
});
