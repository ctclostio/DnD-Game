import {
  setTextContent,
  createElement,
  appendChildren,
  clearElement,
  createInput,
  createSelect,
  sanitizeHTML
} from '../dom';

describe('DOM utilities', () => {
  describe('setTextContent', () => {
    it('should set text content of an element', () => {
      const element = document.createElement('div');
      setTextContent(element, 'Hello World');
      expect(element.textContent).toBe('Hello World');
    });

    it('should handle null element gracefully', () => {
      expect(() => setTextContent(null, 'test')).not.toThrow();
    });
  });

  describe('createElement', () => {
    it('should create element with basic options', () => {
      const element = createElement('div', {
        className: 'test-class',
        id: 'test-id',
        textContent: 'Test Content'
      });

      expect(element.tagName).toBe('DIV');
      expect(element.className).toBe('test-class');
      expect(element.id).toBe('test-id');
      expect(element.textContent).toBe('Test Content');
    });

    it('should add attributes', () => {
      const element = createElement('button', {
        attributes: {
          'data-test': 'value',
          'aria-label': 'Test Button'
        }
      });

      expect(element.getAttribute('data-test')).toBe('value');
      expect(element.getAttribute('aria-label')).toBe('Test Button');
    });

    it('should add event listeners', () => {
      const clickHandler = jest.fn();
      const element = createElement('button', {
        events: {
          click: clickHandler
        }
      });

      element.click();
      expect(clickHandler).toHaveBeenCalled();
    });

    it('should append children', () => {
      const child1 = document.createElement('span');
      const element = createElement('div', {
        children: [child1, 'Text node']
      });

      expect(element.children.length).toBe(1);
      expect(element.children[0]).toBe(child1);
      expect(element.textContent).toContain('Text node');
    });
  });

  describe('appendChildren', () => {
    it('should append multiple children', () => {
      const parent = document.createElement('div');
      const child1 = document.createElement('span');
      const child2 = document.createElement('button');

      appendChildren(parent, child1, 'text', child2);

      expect(parent.children.length).toBe(2);
      expect(parent.textContent).toContain('text');
    });
  });

  describe('clearElement', () => {
    it('should remove all children from element', () => {
      const parent = document.createElement('div');
      parent.innerHTML = '<span>1</span><span>2</span><span>3</span>';
      
      expect(parent.children.length).toBe(3);
      
      clearElement(parent);
      
      expect(parent.children.length).toBe(0);
      expect(parent.textContent).toBe('');
    });
  });

  describe('createInput', () => {
    it('should create input with options', () => {
      const input = createInput({
        type: 'text',
        id: 'test-input',
        name: 'testName',
        value: 'test value',
        placeholder: 'Enter text',
        required: true,
        className: 'form-input'
      });

      expect(input.type).toBe('text');
      expect(input.id).toBe('test-input');
      expect(input.name).toBe('testName');
      expect(input.value).toBe('test value');
      expect(input.placeholder).toBe('Enter text');
      expect(input.required).toBe(true);
      expect(input.className).toBe('form-input');
    });

    it('should create number input with min/max', () => {
      const input = createInput({
        type: 'number',
        min: 1,
        max: 100
      });

      expect(input.type).toBe('number');
      expect(input.min).toBe('1');
      expect(input.max).toBe('100');
    });
  });

  describe('createSelect', () => {
    it('should create select with options', () => {
      const select = createSelect({
        id: 'test-select',
        name: 'testSelect',
        required: true,
        options: [
          { value: '', text: 'Choose...' },
          { value: 'opt1', text: 'Option 1' },
          { value: 'opt2', text: 'Option 2', selected: true }
        ]
      });

      expect(select.id).toBe('test-select');
      expect(select.name).toBe('testSelect');
      expect(select.required).toBe(true);
      expect(select.options.length).toBe(3);
      expect(select.options[0].text).toBe('Choose...');
      expect(select.options[2].selected).toBe(true);
    });
  });

  describe('sanitizeHTML', () => {
    it('should escape HTML tags', () => {
      const dangerous = '<script>alert("XSS")</script>';
      const sanitized = sanitizeHTML(dangerous);
      
      expect(sanitized).not.toContain('<script>');
      expect(sanitized).toContain('&lt;script&gt;');
    });

    it('should escape quotes and special characters', () => {
      const input = '<div class="test">Hello & "World"</div>';
      const sanitized = sanitizeHTML(input);
      
      expect(sanitized).toContain('&lt;div');
      expect(sanitized).toContain('&quot;');
      expect(sanitized).toContain('&amp;');
    });
  });
});