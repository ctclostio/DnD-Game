import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';

interface AccessibilitySettings {
  highContrast: boolean;
  largeText: boolean;
  reduceMotion: boolean;
  screenReaderMode: boolean;
  keyboardNavigation: boolean;
  focusIndicator: boolean;
}

interface AccessibilityContextType {
  settings: AccessibilitySettings;
  updateSetting: (key: keyof AccessibilitySettings, value: boolean) => void;
  announceToScreenReader: (message: string, priority?: 'polite' | 'assertive') => void;
}

const defaultSettings: AccessibilitySettings = {
  highContrast: false,
  largeText: false,
  reduceMotion: false,
  screenReaderMode: false,
  keyboardNavigation: true,
  focusIndicator: true,
};

const AccessibilityContext = createContext<AccessibilityContextType | undefined>(undefined);

export const useAccessibility = () => {
  const context = useContext(AccessibilityContext);
  if (!context) {
    throw new Error('useAccessibility must be used within AccessibilityProvider');
  }
  return context;
};

interface AccessibilityProviderProps {
  children: React.ReactNode;
}

export const AccessibilityProvider: React.FC<AccessibilityProviderProps> = ({ children }) => {
  const [settings, setSettings] = useState<AccessibilitySettings>(() => {
    const saved = localStorage.getItem('a11y-settings');
    return saved ? JSON.parse(saved) : defaultSettings;
  });

  // Check system preferences
  useEffect(() => {
    // Check for reduced motion preference
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)');
    if (prefersReducedMotion.matches) {
      setSettings(prev => ({ ...prev, reduceMotion: true }));
    }

    // Check for high contrast preference
    const prefersHighContrast = window.matchMedia('(prefers-contrast: high)');
    if (prefersHighContrast.matches) {
      setSettings(prev => ({ ...prev, highContrast: true }));
    }

    // Listen for changes
    const handleMotionChange = (e: MediaQueryListEvent) => {
      setSettings(prev => ({ ...prev, reduceMotion: e.matches }));
    };

    const handleContrastChange = (e: MediaQueryListEvent) => {
      setSettings(prev => ({ ...prev, highContrast: e.matches }));
    };

    prefersReducedMotion.addEventListener('change', handleMotionChange);
    prefersHighContrast.addEventListener('change', handleContrastChange);

    return () => {
      prefersReducedMotion.removeEventListener('change', handleMotionChange);
      prefersHighContrast.removeEventListener('change', handleContrastChange);
    };
  }, []);

  // Apply settings to document
  useEffect(() => {
    const root = document.documentElement;
    
    // High contrast
    root.classList.toggle('high-contrast', settings.highContrast);
    
    // Large text
    root.classList.toggle('large-text', settings.largeText);
    
    // Reduce motion
    root.classList.toggle('reduce-motion', settings.reduceMotion);
    
    // Focus indicator
    root.classList.toggle('enhanced-focus', settings.focusIndicator);
    
    // Save to localStorage
    localStorage.setItem('a11y-settings', JSON.stringify(settings));
  }, [settings]);

  const updateSetting = useCallback((key: keyof AccessibilitySettings, value: boolean) => {
    setSettings(prev => ({ ...prev, [key]: value }));
  }, []);

  // Screen reader announcements
  const announceToScreenReader = useCallback((message: string, priority: 'polite' | 'assertive' = 'polite') => {
    const announcement = document.createElement('div');
    announcement.setAttribute('role', 'status');
    announcement.setAttribute('aria-live', priority);
    announcement.setAttribute('aria-atomic', 'true');
    announcement.className = 'sr-only';
    announcement.textContent = message;
    
    document.body.appendChild(announcement);
    
    // Remove after announcement
    setTimeout(() => {
      document.body.removeChild(announcement);
    }, 1000);
  }, []);

  // Keyboard navigation helpers
  useEffect(() => {
    if (!settings.keyboardNavigation) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // Skip to main content with '1'
      if (e.key === '1' && !e.ctrlKey && !e.altKey && !e.metaKey) {
        const main = document.querySelector('main');
        if (main && !(e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement)) {
          e.preventDefault();
          (main as HTMLElement).focus();
          main.scrollIntoView({ behavior: 'smooth' });
        }
      }

      // Navigate landmarks with F6
      if (e.key === 'F6') {
        e.preventDefault();
        const landmarks = document.querySelectorAll('[role="navigation"], [role="main"], [role="complementary"], nav, main, aside');
        const currentIndex = Array.from(landmarks).findIndex(el => el.contains(document.activeElement));
        const nextIndex = (currentIndex + 1) % landmarks.length;
        const nextLandmark = landmarks[nextIndex] as HTMLElement;
        if (nextLandmark) {
          nextLandmark.focus();
          nextLandmark.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [settings.keyboardNavigation]);

  const value: AccessibilityContextType = {
    settings,
    updateSetting,
    announceToScreenReader,
  };

  return (
    <AccessibilityContext.Provider value={value}>
      {children}
    </AccessibilityContext.Provider>
  );
};
