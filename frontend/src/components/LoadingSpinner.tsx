import React from 'react';

interface LoadingSpinnerProps {
  fullScreen?: boolean;
  message?: string;
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ fullScreen = false, message = 'Loading...' }) => {
  const spinnerClass = fullScreen ? 'loading-spinner-fullscreen' : 'loading-spinner';
  
  return (
    <div className={spinnerClass}>
      <div className="spinner-container">
        <div className="d20-spinner">
          <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
            <polygon 
              points="50,5 90,25 90,75 50,95 10,75 10,25" 
              fill="none" 
              stroke="currentColor" 
              strokeWidth="2"
            />
            <text x="50" y="55" textAnchor="middle" className="d20-number">20</text>
          </svg>
        </div>
        <p className="loading-message">{message}</p>
      </div>
    </div>
  );
};

export default LoadingSpinner;
