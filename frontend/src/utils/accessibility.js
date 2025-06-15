// Accessibility utilities for keyboard navigation

/**
 * Handles keyboard events for clickable elements
 * Triggers click handler on Enter or Space key press
 * @param {Function} onClick - The click handler function
 * @returns {Function} Keyboard event handler
 */
export const handleKeyboardClick = (onClick) => (e) => {
    if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        onClick(e);
    }
};

/**
 * Props to make a div behave like a button for accessibility
 * @param {Function} onClick - The click handler function
 * @param {Object} additionalProps - Any additional props to spread
 * @returns {Object} Props object with accessibility attributes
 */
export const getClickableProps = (onClick, additionalProps = {}) => ({
    onClick,
    onKeyDown: handleKeyboardClick(onClick),
    role: 'button',
    tabIndex: 0,
    'aria-pressed': false,
    ...additionalProps
});

/**
 * Props for selectable items (like cards that can be selected)
 * @param {Function} onClick - The click handler function
 * @param {boolean} isSelected - Whether the item is currently selected
 * @param {Object} additionalProps - Any additional props to spread
 * @returns {Object} Props object with accessibility attributes
 */
export const getSelectableProps = (onClick, isSelected = false, additionalProps = {}) => ({
    onClick,
    onKeyDown: handleKeyboardClick(onClick),
    role: 'button',
    tabIndex: 0,
    'aria-pressed': isSelected,
    'aria-selected': isSelected,
    ...additionalProps
});