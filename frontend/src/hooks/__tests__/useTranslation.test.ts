import { renderHook, act } from '@testing-library/react';
import { useTranslation } from '../useTranslation';

// Mock the i18n module
const mockListeners = new Set<() => void>();
const mockI18n: { [key: string]: jest.Mock } = {
  getLocale: jest.fn(() => 'en'),
  setLocale: jest.fn((locale: string) => {
    mockI18n.getLocale.mockReturnValue(locale);
    mockListeners.forEach(listener => listener());
  }),
  subscribe: jest.fn((listener: () => void) => {
    mockListeners.add(listener);
    return () => mockListeners.delete(listener);
  }),
  t: jest.fn((key: string, params?: Record<string, string | number>) => {
    if (params) {
      return `${key} with params: ${JSON.stringify(params)}`;
    }
    return key;
  }),
  getAvailableLocales: jest.fn(() => ['en', 'es', 'fr']),
  formatNumber: jest.fn((value: number, options?: Intl.NumberFormatOptions) => {
    return new Intl.NumberFormat(mockI18n.getLocale(), options).format(value);
  }),
  formatDate: jest.fn((date: Date | string, options?: Intl.DateTimeFormatOptions) => {
    const dateObj = typeof date === 'string' ? new Date(date) : date;
    return new Intl.DateTimeFormat(mockI18n.getLocale(), options).format(dateObj);
  }),
  formatRelativeTime: jest.fn((date: Date | string) => {
    return '2 hours ago';
  }),
};

jest.mock('../../i18n', () => ({
  default: mockI18n,
}));

describe('useTranslation', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockListeners.clear();
    mockI18n.getLocale.mockReturnValue('en');
  });

  it('should provide translation function', () => {
    const { result } = renderHook(() => useTranslation());

    expect(result.current.t).toBeDefined();
    expect(typeof result.current.t).toBe('function');
  });

  it('should translate keys', () => {
    const { result } = renderHook(() => useTranslation());

    const translated = result.current.t('common.save');
    expect(mockI18n.t).toHaveBeenCalledWith('common.save', undefined);
    expect(translated).toBe('common.save');
  });

  it('should translate with parameters', () => {
    const { result } = renderHook(() => useTranslation());

    const params = { name: 'John', count: 5 };
    const translated = result.current.t('greeting.hello', params);
    
    expect(mockI18n.t).toHaveBeenCalledWith('greeting.hello', params);
    expect(translated).toBe('greeting.hello with params: {"name":"John","count":5}');
  });

  it('should return current locale', () => {
    const { result } = renderHook(() => useTranslation());

    expect(result.current.locale).toBe('en');
    expect(mockI18n.getLocale).toHaveBeenCalled();
  });

  it('should change locale', () => {
    const { result } = renderHook(() => useTranslation());

    act(() => {
      result.current.setLocale('es');
    });

    expect(mockI18n.setLocale).toHaveBeenCalledWith('es');
    expect(result.current.locale).toBe('es');
  });

  it('should update when locale changes externally', () => {
    const { result } = renderHook(() => useTranslation());

    expect(result.current.locale).toBe('en');

    // Simulate external locale change
    act(() => {
      mockI18n.getLocale.mockReturnValue('fr');
      mockListeners.forEach(listener => listener());
    });

    expect(result.current.locale).toBe('fr');
  });

  it('should return available locales', () => {
    const { result } = renderHook(() => useTranslation());

    expect(result.current.availableLocales).toEqual(['en', 'es', 'fr']);
    expect(mockI18n.getAvailableLocales).toHaveBeenCalled();
  });

  it('should format numbers', () => {
    const { result } = renderHook(() => useTranslation());

    const formatted = result.current.formatNumber(1234.56);
    expect(mockI18n.formatNumber).toHaveBeenCalledWith(1234.56, undefined);
    expect(formatted).toBe('1,234.56');
  });

  it('should format numbers with options', () => {
    const { result } = renderHook(() => useTranslation());

    const options: Intl.NumberFormatOptions = {
      style: 'currency',
      currency: 'USD',
    };

    const formatted = result.current.formatNumber(99.99, options);
    expect(mockI18n.formatNumber).toHaveBeenCalledWith(99.99, options);
    expect(formatted).toMatch(/\$99\.99/);
  });

  it('should format dates', () => {
    const { result } = renderHook(() => useTranslation());

    const date = new Date('2023-12-25');
    const formatted = result.current.formatDate(date);
    
    expect(mockI18n.formatDate).toHaveBeenCalledWith(date, undefined);
    expect(formatted).toMatch(/12\/25\/2023/);
  });

  it('should format dates from strings', () => {
    const { result } = renderHook(() => useTranslation());

    const dateString = '2023-12-25T10:30:00Z';
    const formatted = result.current.formatDate(dateString);
    
    expect(mockI18n.formatDate).toHaveBeenCalledWith(dateString, undefined);
    expect(formatted).toBeTruthy();
  });

  it('should format dates with options', () => {
    const { result } = renderHook(() => useTranslation());

    const date = new Date('2023-12-25');
    const options: Intl.DateTimeFormatOptions = {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    };

    const formatted = result.current.formatDate(date, options);
    expect(mockI18n.formatDate).toHaveBeenCalledWith(date, options);
    expect(formatted).toMatch(/December 25, 2023/);
  });

  it('should format relative time', () => {
    const { result } = renderHook(() => useTranslation());

    const date = new Date();
    date.setHours(date.getHours() - 2);

    const formatted = result.current.formatRelativeTime(date);
    expect(mockI18n.formatRelativeTime).toHaveBeenCalledWith(date);
    expect(formatted).toBe('2 hours ago');
  });

  it('should format relative time from strings', () => {
    const { result } = renderHook(() => useTranslation());

    const dateString = new Date().toISOString();
    const formatted = result.current.formatRelativeTime(dateString);
    
    expect(mockI18n.formatRelativeTime).toHaveBeenCalledWith(dateString);
    expect(formatted).toBe('2 hours ago');
  });

  it('should unsubscribe on unmount', () => {
    const unsubscribeMock = jest.fn();
    mockI18n.subscribe.mockReturnValue(unsubscribeMock);

    const { unmount } = renderHook(() => useTranslation());

    expect(mockI18n.subscribe).toHaveBeenCalled();
    
    unmount();

    expect(unsubscribeMock).toHaveBeenCalled();
  });

  it('should handle locale changes with number formatting', () => {
    const { result } = renderHook(() => useTranslation());

    // Format in English
    let formatted = result.current.formatNumber(1234.56);
    expect(formatted).toBe('1,234.56');

    // Change to Spanish (which uses different separators)
    act(() => {
      result.current.setLocale('es');
    });

    // Mock Spanish formatting
    mockI18n.formatNumber.mockImplementationOnce((value: number) => {
      return new Intl.NumberFormat('es', undefined).format(value);
    });

    formatted = result.current.formatNumber(1234.56);
    expect(formatted).toMatch(/1\.234,56|1,234\.56/); // Depends on system locale support
  });

  it('should maintain stable function references', () => {
    const { result, rerender } = renderHook(() => useTranslation());

    const { t, setLocale, formatNumber, formatDate, formatRelativeTime } = result.current;

    rerender();

    expect(result.current.t).toBe(t);
    expect(result.current.setLocale).toBe(setLocale);
    expect(result.current.formatNumber).toBe(formatNumber);
    expect(result.current.formatDate).toBe(formatDate);
    expect(result.current.formatRelativeTime).toBe(formatRelativeTime);
  });
});