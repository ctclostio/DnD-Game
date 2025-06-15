import React, { memo } from 'react';
import { useAccessibility } from './AccessibilityProvider';
import { useTranslation } from '../hooks/useTranslation';

export const AccessibilitySettings = memo(() => {
  const { settings, updateSetting } = useAccessibility();
  const { t } = useTranslation();

  const settingsConfig = [
    {
      key: 'highContrast' as const,
      label: 'High Contrast Mode',
      description: 'Increase color contrast for better visibility',
      icon: '🎨',
    },
    {
      key: 'largeText' as const,
      label: 'Large Text',
      description: 'Increase font size throughout the application',
      icon: '🔍',
    },
    {
      key: 'reduceMotion' as const,
      label: 'Reduce Motion',
      description: 'Minimize animations and transitions',
      icon: '🚫',
    },
    {
      key: 'screenReaderMode' as const,
      label: 'Screen Reader Mode',
      description: 'Optimize for screen reader usage',
      icon: '🔊',
    },
    {
      key: 'keyboardNavigation' as const,
      label: 'Enhanced Keyboard Navigation',
      description: 'Additional keyboard shortcuts and navigation aids',
      icon: '⌨️',
    },
    {
      key: 'focusIndicator' as const,
      label: 'Enhanced Focus Indicators',
      description: 'More visible focus outlines for keyboard navigation',
      icon: '🎯',
    },
  ];

  return (
    <div className="accessibility-settings">
      <h2>{t('nav.settings')} - Accessibility</h2>
      <p className="settings-description">
        Customize the application to meet your accessibility needs.
      </p>

      <div className="settings-grid">
        {settingsConfig.map(({ key, label, description, icon }) => (
          <div key={key} className="setting-item">
            <div className="setting-header">
              <span className="setting-icon" aria-hidden="true">{icon}</span>
              <label htmlFor={`a11y-${key}`} className="setting-label">
                {label}
              </label>
            </div>
            <p className="setting-description">{description}</p>
            <div className="setting-control">
              <input
                type="checkbox"
                id={`a11y-${key}`}
                checked={settings[key]}
                onChange={(e) => updateSetting(key, e.target.checked)}
                className="toggle-input"
                role="switch"
                aria-checked={settings[key]}
              />
              <label htmlFor={`a11y-${key}`} className="toggle-label">
                <span className="toggle-slider" />
                <span className="sr-only">
                  {settings[key] ? 'Enabled' : 'Disabled'}
                </span>
              </label>
            </div>
          </div>
        ))}
      </div>

      <div className="keyboard-shortcuts">
        <h3>Keyboard Shortcuts</h3>
        <dl className="shortcuts-list">
          <div className="shortcut-item">
            <dt><kbd>1</kbd></dt>
            <dd>Skip to main content</dd>
          </div>
          <div className="shortcut-item">
            <dt><kbd>F6</kbd></dt>
            <dd>Navigate between landmarks</dd>
          </div>
          <div className="shortcut-item">
            <dt><kbd>Tab</kbd></dt>
            <dd>Move to next interactive element</dd>
          </div>
          <div className="shortcut-item">
            <dt><kbd>Shift + Tab</kbd></dt>
            <dd>Move to previous interactive element</dd>
          </div>
          <div className="shortcut-item">
            <dt><kbd>Enter</kbd> / <kbd>Space</kbd></dt>
            <dd>Activate buttons and links</dd>
          </div>
          <div className="shortcut-item">
            <dt><kbd>Esc</kbd></dt>
            <dd>Close dialogs and menus</dd>
          </div>
        </dl>
      </div>
    </div>
  );
});

AccessibilitySettings.displayName = 'AccessibilitySettings';
