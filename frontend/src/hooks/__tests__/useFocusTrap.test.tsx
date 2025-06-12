import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { renderHook } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useFocusTrap } from '../useFocusTrap';

// Test component that uses the hook
const TestComponent: React.FC<{
  options?: Parameters<typeof useFocusTrap>[0];
  children?: React.ReactNode;
}> = ({ options = {}, children }) => {
  const containerRef = useFocusTrap<HTMLDivElement>(options);

  return (
    <div>
      <button data-testid="outside-before">Outside Before</button>
      <div ref={containerRef} data-testid="trap-container">
        {children || (
          <>
            <button data-testid="first-button">First</button>
            <input data-testid="input" type="text" />
            <textarea data-testid="textarea" />
            <select data-testid="select">
              <option>Option</option>
            </select>
            <a href="#" data-testid="link">Link</a>
            <button data-testid="last-button">Last</button>
          </>
        )}
      </div>
      <button data-testid="outside-after">Outside After</button>
    </div>
  );
};

describe('useFocusTrap', () => {
  let previousActiveElement: Element | null;

  beforeEach(() => {
    previousActiveElement = document.activeElement;
  });

  afterEach(() => {
    // Clean up focus
    if (previousActiveElement instanceof HTMLElement) {
      previousActiveElement.focus();
    }
  });

  it('should focus first focusable element on mount', async () => {
    render(<TestComponent />);

    await waitFor(() => {
      expect(document.activeElement).toBe(screen.getByTestId('first-button'));
    });
  });

  it('should focus initial focus element when specified', async () => {
    render(
      <TestComponent options={{ initialFocus: '[data-testid="input"]' }} />
    );

    await waitFor(() => {
      expect(document.activeElement).toBe(screen.getByTestId('input'));
    });
  });

  it('should trap Tab navigation', async () => {
    render(<TestComponent />);
    const user = userEvent.setup();

    const firstButton = screen.getByTestId('first-button');
    const lastButton = screen.getByTestId('last-button');

    // Wait for initial focus setup
    await waitFor(() => {
      expect(document.activeElement).toBe(firstButton);
    });

    // Focus last button
    lastButton.focus();
    expect(document.activeElement).toBe(lastButton);

    // Tab should wrap to first element
    await user.tab();
    expect(document.activeElement).toBe(firstButton);
  });

  it('should trap Shift+Tab navigation', async () => {
    render(<TestComponent />);
    const user = userEvent.setup();

    const firstButton = screen.getByTestId('first-button');
    const lastButton = screen.getByTestId('last-button');

    // Focus first button
    firstButton.focus();
    expect(document.activeElement).toBe(firstButton);

    // Shift+Tab should wrap to last element
    await user.tab({ shift: true });
    expect(document.activeElement).toBe(lastButton);
  });

  it('should navigate through all focusable elements', async () => {
    render(<TestComponent />);
    const user = userEvent.setup();

    const elements = [
      screen.getByTestId('first-button'),
      screen.getByTestId('input'),
      screen.getByTestId('textarea'),
      screen.getByTestId('select'),
      screen.getByTestId('link'),
      screen.getByTestId('last-button'),
    ];

    // Wait for initial focus
    await waitFor(() => {
      expect(document.activeElement).toBe(elements[0]);
    });

    // Tab through all elements
    for (let i = 1; i < elements.length; i++) {
      await user.tab();
      expect(document.activeElement).toBe(elements[i]);
    }

    // One more tab should wrap to first
    await user.tab();
    expect(document.activeElement).toBe(elements[0]);
  });

  it('should prevent clicks outside when allowOutsideClick is false', () => {
    render(<TestComponent options={{ allowOutsideClick: false }} />);

    const outsideButton = screen.getByTestId('outside-after');
    const firstButton = screen.getByTestId('first-button');

    // Focus should move to first element initially
    firstButton.focus();

    // Click outside should be prevented
    const clickEvent = new MouseEvent('mousedown', { bubbles: true });
    const preventDefaultSpy = jest.spyOn(clickEvent, 'preventDefault');
    
    outsideButton.dispatchEvent(clickEvent);

    expect(preventDefaultSpy).toHaveBeenCalled();
    expect(document.activeElement).toBe(firstButton);
  });

  it('should allow clicks outside when allowOutsideClick is true', () => {
    render(<TestComponent options={{ allowOutsideClick: true }} />);

    const outsideButton = screen.getByTestId('outside-after');

    const clickEvent = new MouseEvent('mousedown', { bubbles: true });
    const preventDefaultSpy = jest.spyOn(clickEvent, 'preventDefault');
    
    outsideButton.dispatchEvent(clickEvent);

    expect(preventDefaultSpy).not.toHaveBeenCalled();
  });

  it('should return focus on unmount when returnFocus is true', async () => {
    const outsideButton = document.createElement('button');
    document.body.appendChild(outsideButton);
    outsideButton.focus();

    const { unmount } = render(<TestComponent options={{ returnFocus: true }} />);

    // Wait for initial focus
    await waitFor(() => {
      expect(document.activeElement).not.toBe(outsideButton);
    });

    unmount();

    expect(document.activeElement).toBe(outsideButton);
    document.body.removeChild(outsideButton);
  });

  it('should not return focus on unmount when returnFocus is false', async () => {
    const outsideButton = document.createElement('button');
    document.body.appendChild(outsideButton);
    outsideButton.focus();

    const { unmount } = render(<TestComponent options={{ returnFocus: false }} />);

    // Wait for initial focus
    await waitFor(() => {
      expect(document.activeElement).not.toBe(outsideButton);
    });

    unmount();

    expect(document.activeElement).not.toBe(outsideButton);
    document.body.removeChild(outsideButton);
  });

  it('should skip disabled elements', async () => {
    render(
      <TestComponent>
        <button data-testid="enabled">Enabled</button>
        <button data-testid="disabled" disabled>Disabled</button>
        <input data-testid="enabled-input" />
        <input data-testid="disabled-input" disabled />
      </TestComponent>
    );

    const user = userEvent.setup();
    const enabledButton = screen.getByTestId('enabled');
    const enabledInput = screen.getByTestId('enabled-input');

    // Should focus first enabled element
    await waitFor(() => {
      expect(document.activeElement).toBe(enabledButton);
    });

    // Tab should skip disabled button
    await user.tab();
    expect(document.activeElement).toBe(enabledInput);

    // Tab again should wrap to first (skipping disabled input)
    await user.tab();
    expect(document.activeElement).toBe(enabledButton);
  });

  it('should skip hidden elements', async () => {
    render(
      <TestComponent>
        <button data-testid="visible">Visible</button>
        <button data-testid="display-none" style={{ display: 'none' }}>Hidden</button>
        <button data-testid="visibility-hidden" style={{ visibility: 'hidden' }}>Hidden</button>
        <button data-testid="visible-2">Visible 2</button>
      </TestComponent>
    );

    const user = userEvent.setup();
    const visibleButton = screen.getByTestId('visible');
    const visibleButton2 = screen.getByTestId('visible-2');

    // Should focus first visible element
    await waitFor(() => {
      expect(document.activeElement).toBe(visibleButton);
    });

    // Tab should skip hidden elements
    await user.tab();
    expect(document.activeElement).toBe(visibleButton2);
  });

  it('should handle container with no focusable elements', async () => {
    render(
      <TestComponent>
        <div>No focusable elements</div>
      </TestComponent>
    );

    // Should not throw error
    const container = screen.getByTestId('trap-container');
    expect(container).toBeInTheDocument();
  });

  it('should not trap focus when disabled', async () => {
    render(<TestComponent options={{ enabled: false }} />);
    const user = userEvent.setup();

    const outsideAfter = screen.getByTestId('outside-after');
    const lastButton = screen.getByTestId('last-button');

    lastButton.focus();
    
    // Tab should go to outside element
    await user.tab();
    expect(document.activeElement).toBe(outsideAfter);
  });

  it('should handle dynamic enable/disable', async () => {
    const { rerender } = render(<TestComponent options={{ enabled: true }} />);
    const user = userEvent.setup();

    const firstButton = screen.getByTestId('first-button');
    const lastButton = screen.getByTestId('last-button');
    const outsideAfter = screen.getByTestId('outside-after');

    // Wait for initial focus setup
    await waitFor(() => {
      expect(document.activeElement).toBe(firstButton);
    });

    // Initially enabled - should trap
    lastButton.focus();
    await user.tab();
    expect(document.activeElement).toBe(firstButton);

    // Disable trap
    rerender(<TestComponent options={{ enabled: false }} />);
    
    lastButton.focus();
    await user.tab();
    expect(document.activeElement).toBe(outsideAfter);
  });

  it('should handle elements with negative tabindex', async () => {
    render(
      <TestComponent>
        <button data-testid="first-tab">First</button>
        <button data-testid="negative-tab" tabIndex={-1}>Negative</button>
        <button data-testid="last-tab">Last</button>
      </TestComponent>
    );

    const user = userEvent.setup();
    const firstTab = screen.getByTestId('first-tab');
    const negativeTab = screen.getByTestId('negative-tab');
    const lastTab = screen.getByTestId('last-tab');

    // Wait for initial focus
    await waitFor(() => {
      expect(document.activeElement).toBe(firstTab);
    });

    // Tab should skip negative tabindex element
    await user.tab();
    expect(document.activeElement).toBe(lastTab);

    // Tab again should wrap back to first
    await user.tab();
    expect(document.activeElement).toBe(firstTab);

    // Negative tabindex element should never receive focus via tab
    expect(negativeTab).not.toBe(document.activeElement);
  });

  it('should return stable ref', () => {
    const { result, rerender } = renderHook(() => useFocusTrap());

    const ref1 = result.current;
    rerender();
    const ref2 = result.current;

    expect(ref1).toBe(ref2);
  });
});