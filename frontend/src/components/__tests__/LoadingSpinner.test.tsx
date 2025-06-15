import React from 'react';
import { render, screen } from '@testing-library/react';
import LoadingSpinner from '../LoadingSpinner';

describe('LoadingSpinner', () => {
  it('renders with default message', () => {
    render(<LoadingSpinner />);
    
    expect(screen.getByText('Loading...')).toBeInTheDocument();
    expect(screen.getByText('20')).toBeInTheDocument(); // D20 number
  });

  it('renders with custom message', () => {
    const customMessage = 'Please wait while we load your data';
    render(<LoadingSpinner message={customMessage} />);
    
    expect(screen.getByText(customMessage)).toBeInTheDocument();
  });

  it('renders as normal spinner by default', () => {
    const { container } = render(<LoadingSpinner />);
    
    const spinnerDiv = container.querySelector('.loading-spinner');
    expect(spinnerDiv).toBeInTheDocument();
    expect(spinnerDiv).not.toHaveClass('loading-spinner-fullscreen');
  });

  it('renders fullscreen variant', () => {
    const { container } = render(<LoadingSpinner fullScreen />);
    
    const spinnerDiv = container.querySelector('.loading-spinner-fullscreen');
    expect(spinnerDiv).toBeInTheDocument();
  });

  it('renders D20 SVG', () => {
    const { container } = render(<LoadingSpinner />);
    
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
    
    const polygon = container.querySelector('polygon');
    expect(polygon).toBeInTheDocument();
    expect(polygon).toHaveAttribute('points', '50,5 90,25 90,75 50,95 10,75 10,25');
  });

  it('displays loading message with correct class', () => {
    render(<LoadingSpinner message="Rolling dice..." />);
    
    const message = screen.getByText('Rolling dice...');
    expect(message).toHaveClass('loading-message');
  });

  it('renders spinner container', () => {
    const { container } = render(<LoadingSpinner />);
    
    const spinnerContainer = container.querySelector('.spinner-container');
    expect(spinnerContainer).toBeInTheDocument();
    
    const d20Spinner = container.querySelector('.d20-spinner');
    expect(d20Spinner).toBeInTheDocument();
  });
});
