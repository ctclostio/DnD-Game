import { useEffect, useRef, useCallback } from 'react';

interface UseFocusTrapOptions {
  enabled?: boolean;
  returnFocus?: boolean;
  initialFocus?: string; // Selector for initial focus element
  allowOutsideClick?: boolean;
}

export function useFocusTrap<T extends HTMLElement = HTMLDivElement>(
  options: UseFocusTrapOptions = {}
) {
  const {
    enabled = true,
    returnFocus = true,
    initialFocus,
    allowOutsideClick = false,
  } = options;

  const containerRef = useRef<T>(null);
  const previouslyFocusedElement = useRef<HTMLElement | null>(null);

  // Get all focusable elements within the container
  const getFocusableElements = useCallback((): HTMLElement[] => {
    if (!containerRef.current) return [];

    const focusableSelectors = [
      'a[href]',
      'button:not([disabled])',
      'textarea:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      '[tabindex]:not([tabindex="-1"])',
    ].join(', ');

    return Array.from(
      containerRef.current.querySelectorAll<HTMLElement>(focusableSelectors)
    ).filter(el => {
      // Check if element is visible
      const style = window.getComputedStyle(el);
      return style.display !== 'none' && style.visibility !== 'hidden';
    });
  }, []);

  // Handle tab key navigation
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (!enabled || e.key !== 'Tab') return;

    const focusableElements = getFocusableElements();
    if (focusableElements.length === 0) return;

    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];
    const activeElement = document.activeElement as HTMLElement;

    // Trap focus within container
    if (e.shiftKey) {
      // Shift + Tab
      if (activeElement === firstElement) {
        e.preventDefault();
        lastElement.focus();
      }
    } else {
      // Tab
      if (activeElement === lastElement) {
        e.preventDefault();
        firstElement.focus();
      }
    }
  }, [enabled, getFocusableElements]);

  // Handle clicks outside the container
  const handleClickOutside = useCallback((e: MouseEvent) => {
    if (!enabled || allowOutsideClick) return;

    if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
      e.preventDefault();
      e.stopPropagation();
      
      // Return focus to the container
      const focusableElements = getFocusableElements();
      if (focusableElements.length > 0) {
        focusableElements[0].focus();
      }
    }
  }, [enabled, allowOutsideClick, getFocusableElements]);

  // Initialize focus trap
  useEffect(() => {
    if (!enabled) return;

    // Store currently focused element
    previouslyFocusedElement.current = document.activeElement as HTMLElement;

    // Set initial focus
    setTimeout(() => {
      if (initialFocus && containerRef.current) {
        const initialElement = containerRef.current.querySelector<HTMLElement>(initialFocus);
        if (initialElement) {
          initialElement.focus();
          return;
        }
      }

      // Otherwise focus first focusable element
      const focusableElements = getFocusableElements();
      if (focusableElements.length > 0) {
        focusableElements[0].focus();
      }
    }, 0);

    // Add event listeners
    document.addEventListener('keydown', handleKeyDown);
    document.addEventListener('mousedown', handleClickOutside);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.removeEventListener('mousedown', handleClickOutside);

      // Return focus to previously focused element
      if (returnFocus && previouslyFocusedElement.current) {
        previouslyFocusedElement.current.focus();
      }
    };
  }, [enabled, initialFocus, returnFocus, handleKeyDown, handleClickOutside, getFocusableElements]);

  return containerRef;
}