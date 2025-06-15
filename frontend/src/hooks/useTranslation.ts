import { useState, useEffect, useCallback } from 'react';
import i18n from '../i18n';

interface UseTranslationReturn {
  t: (key: string, params?: Record<string, any>) => string;
  locale: string;
  setLocale: (locale: string) => void;
  availableLocales: string[];
  formatNumber: (value: number, options?: Intl.NumberFormatOptions) => string;
  formatDate: (date: Date | string, options?: Intl.DateTimeFormatOptions) => string;
  formatRelativeTime: (date: Date | string) => string;
}

export function useTranslation(): UseTranslationReturn {
  const [locale, setLocaleState] = useState(i18n.getLocale());

  useEffect(() => {
    const unsubscribe = i18n.subscribe(() => {
      setLocaleState(i18n.getLocale());
    });

    return unsubscribe;
  }, []);

  const setLocale = useCallback((newLocale: string) => {
    i18n.setLocale(newLocale);
  }, []);

  const t = useCallback((key: string, params?: Record<string, any>) => {
    return i18n.t(key, params);
  }, []);

  const formatNumber = useCallback((value: number, options?: Intl.NumberFormatOptions) => {
    return i18n.formatNumber(value, options);
  }, []);

  const formatDate = useCallback((date: Date | string, options?: Intl.DateTimeFormatOptions) => {
    return i18n.formatDate(date, options);
  }, []);

  const formatRelativeTime = useCallback((date: Date | string) => {
    return i18n.formatRelativeTime(date);
  }, []);

  return {
    t,
    locale,
    setLocale,
    availableLocales: i18n.getAvailableLocales(),
    formatNumber,
    formatDate,
    formatRelativeTime,
  };
}
