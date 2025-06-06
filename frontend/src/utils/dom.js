/**
 * DOM utility functions for safe content manipulation
 */

/**
 * Safely sets text content of an element
 * @param {HTMLElement} element - The element to update
 * @param {string} text - The text content to set
 */
export function setTextContent(element, text) {
    if (element) {
        element.textContent = text;
    }
}

/**
 * Creates an element with safe text content
 * @param {string} tag - The HTML tag name
 * @param {Object} options - Options for the element
 * @returns {HTMLElement}
 */
export function createElement(tag, options = {}) {
    const element = document.createElement(tag);
    
    if (options.className) {
        element.className = options.className;
    }
    
    if (options.id) {
        element.id = options.id;
    }
    
    if (options.textContent) {
        element.textContent = options.textContent;
    }
    
    if (options.attributes) {
        Object.entries(options.attributes).forEach(([key, value]) => {
            element.setAttribute(key, value);
        });
    }
    
    if (options.children) {
        options.children.forEach(child => {
            if (typeof child === 'string') {
                element.appendChild(document.createTextNode(child));
            } else {
                element.appendChild(child);
            }
        });
    }
    
    if (options.events) {
        Object.entries(options.events).forEach(([event, handler]) => {
            element.addEventListener(event, handler);
        });
    }
    
    return element;
}

/**
 * Safely appends children to a parent element
 * @param {HTMLElement} parent - The parent element
 * @param {...(HTMLElement|string)} children - Children to append
 */
export function appendChildren(parent, ...children) {
    children.forEach(child => {
        if (typeof child === 'string') {
            parent.appendChild(document.createTextNode(child));
        } else if (child instanceof HTMLElement) {
            parent.appendChild(child);
        }
    });
}

/**
 * Clears all children from an element
 * @param {HTMLElement} element - The element to clear
 */
export function clearElement(element) {
    while (element.firstChild) {
        element.removeChild(element.firstChild);
    }
}

/**
 * Creates a form input element safely
 * @param {Object} options - Input options
 * @returns {HTMLElement}
 */
export function createInput(options = {}) {
    const input = document.createElement('input');
    
    if (options.type) input.type = options.type;
    if (options.id) input.id = options.id;
    if (options.name) input.name = options.name;
    if (options.value !== undefined) input.value = options.value;
    if (options.placeholder) input.placeholder = options.placeholder;
    if (options.required) input.required = true;
    if (options.min !== undefined) input.min = options.min;
    if (options.max !== undefined) input.max = options.max;
    if (options.className) input.className = options.className;
    
    return input;
}

/**
 * Creates a select element with options
 * @param {Object} options - Select options
 * @returns {HTMLElement}
 */
export function createSelect(options = {}) {
    const select = document.createElement('select');
    
    if (options.id) select.id = options.id;
    if (options.name) select.name = options.name;
    if (options.required) select.required = true;
    if (options.className) select.className = options.className;
    
    if (options.options) {
        options.options.forEach(opt => {
            const option = document.createElement('option');
            option.value = opt.value;
            option.textContent = opt.text;
            if (opt.selected) option.selected = true;
            select.appendChild(option);
        });
    }
    
    return select;
}

/**
 * Sanitizes user input to prevent XSS
 * Only use this if you absolutely need to insert HTML
 * @param {string} html - The HTML string to sanitize
 * @returns {string}
 */
export function sanitizeHTML(html) {
    const div = document.createElement('div');
    div.textContent = html;
    return div.innerHTML;
}