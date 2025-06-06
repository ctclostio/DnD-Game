import { en } from './translations/en';
import { es } from './translations/es';

export type TranslationKey = typeof en;

export interface I18nConfig {
  locale: string;
  fallbackLocale: string;
  translations: Record<string, TranslationKey>;
}

class I18n {
  private locale: string;
  private fallbackLocale: string;
  private translations: Record<string, any>;
  private listeners: Set<() => void> = new Set();

  constructor(config: I18nConfig) {
    this.locale = config.locale;
    this.fallbackLocale = config.fallbackLocale;
    this.translations = config.translations;
  }

  // Get current locale
  getLocale(): string {
    return this.locale;
  }

  // Set locale
  setLocale(locale: string): void {
    if (this.translations[locale]) {
      this.locale = locale;
      localStorage.setItem('locale', locale);
      this.notifyListeners();
    } else {
      console.warn(`Locale '${locale}' not found. Using fallback.`);
      this.locale = this.fallbackLocale;
    }
  }

  // Get translation
  t(key: string, params?: Record<string, any>): string {
    const keys = key.split('.');
    let translation = this.translations[this.locale] || this.translations[this.fallbackLocale];
    
    for (const k of keys) {
      translation = translation?.[k];
      if (!translation) break;
    }

    if (typeof translation !== 'string') {
      console.warn(`Translation not found for key: ${key}`);
      return key;
    }

    // Replace parameters
    if (params) {
      return translation.replace(/\{\{(\w+)\}\}/g, (match, param) => {
        return params[param] !== undefined ? String(params[param]) : match;
      });
    }

    return translation;
  }

  // Subscribe to locale changes
  subscribe(listener: () => void): () => void {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  // Notify listeners
  private notifyListeners(): void {
    this.listeners.forEach(listener => listener());
  }

  // Get available locales
  getAvailableLocales(): string[] {
    return Object.keys(this.translations);
  }

  // Format number based on locale
  formatNumber(value: number, options?: Intl.NumberFormatOptions): string {
    return new Intl.NumberFormat(this.locale, options).format(value);
  }

  // Format date based on locale
  formatDate(date: Date | string, options?: Intl.DateTimeFormatOptions): string {
    const dateObj = typeof date === 'string' ? new Date(date) : date;
    return new Intl.DateTimeFormat(this.locale, options).format(dateObj);
  }

  // Format relative time
  formatRelativeTime(date: Date | string): string {
    const dateObj = typeof date === 'string' ? new Date(date) : date;
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - dateObj.getTime()) / 1000);
    
    const rtf = new Intl.RelativeTimeFormat(this.locale, { numeric: 'auto' });
    
    if (diffInSeconds < 60) {
      return rtf.format(-diffInSeconds, 'second');
    } else if (diffInSeconds < 3600) {
      return rtf.format(-Math.floor(diffInSeconds / 60), 'minute');
    } else if (diffInSeconds < 86400) {
      return rtf.format(-Math.floor(diffInSeconds / 3600), 'hour');
    } else if (diffInSeconds < 2592000) {
      return rtf.format(-Math.floor(diffInSeconds / 86400), 'day');
    } else if (diffInSeconds < 31536000) {
      return rtf.format(-Math.floor(diffInSeconds / 2592000), 'month');
    } else {
      return rtf.format(-Math.floor(diffInSeconds / 31536000), 'year');
    }
  }
}

// Initialize i18n
const savedLocale = localStorage.getItem('locale') || navigator.language.split('-')[0];
const i18n = new I18n({
  locale: savedLocale,
  fallbackLocale: 'en',
  translations: {
    en,
    es,
  },
});

export default i18n;