import React, { memo } from 'react';
import { useTranslation } from '../hooks/useTranslation';

interface LanguageOption {
  code: string;
  name: string;
  flag: string;
}

const languages: LanguageOption[] = [
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'es', name: 'Español', flag: '🇪🇸' },
];

export const LanguageSelector = memo(() => {
  const { locale, setLocale, availableLocales } = useTranslation();

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setLocale(e.target.value);
  };

  return (
    <div className="language-selector">
      <label htmlFor="language-select" className="visually-hidden">
        Select Language
      </label>
      <select
        id="language-select"
        value={locale}
        onChange={handleChange}
        className="language-select"
        aria-label="Select language"
      >
        {languages
          .filter(lang => availableLocales.includes(lang.code))
          .map(lang => (
            <option key={lang.code} value={lang.code}>
              {lang.flag} {lang.name}
            </option>
          ))}
      </select>
    </div>
  );
});

LanguageSelector.displayName = 'LanguageSelector';

// Compact version for header/navbar
export const CompactLanguageSelector = memo(() => {
  const { locale, setLocale, availableLocales } = useTranslation();

  const currentLanguage = languages.find(lang => lang.code === locale);

  const handleClick = () => {
    const currentIndex = availableLocales.indexOf(locale);
    const nextIndex = (currentIndex + 1) % availableLocales.length;
    setLocale(availableLocales[nextIndex]);
  };

  return (
    <button
      className="language-toggle"
      onClick={handleClick}
      aria-label={`Current language: ${currentLanguage?.name}. Click to change.`}
      title="Change language"
    >
      <span className="language-flag">{currentLanguage?.flag}</span>
      <span className="language-code">{locale.toUpperCase()}</span>
    </button>
  );
});

CompactLanguageSelector.displayName = 'CompactLanguageSelector';