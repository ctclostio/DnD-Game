import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): State {
    // Update state so the next render will show the fallback UI
    return {
      hasError: true,
      error,
      errorInfo: null,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Log the error to an error reporting service
    console.error('Error caught by boundary:', error, errorInfo);
    
    // Call the optional error handler
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }
    
    // Update state with error info
    this.setState({
      errorInfo,
    });
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  render() {
    if (this.state.hasError) {
      // Custom fallback UI
      if (this.props.fallback) {
        return <>{this.props.fallback}</>;
      }

      // Default error UI
      return (
        <div className="error-boundary-default">
          <h2>Oops! Something went wrong</h2>
          <details style={{ whiteSpace: 'pre-wrap' }}>
            <summary>Error details</summary>
            {this.state.error && this.state.error.toString()}
            <br />
            {this.state.errorInfo && this.state.errorInfo.componentStack}
          </details>
          <button onClick={this.handleReset}>Try again</button>
        </div>
      );
    }

    return this.props.children;
  }
}

// Specialized error boundary for async operations
interface AsyncErrorBoundaryState extends State {
  resetKey: number;
}

export class AsyncErrorBoundary extends Component<Props, AsyncErrorBoundaryState> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      resetKey: 0,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<AsyncErrorBoundaryState> {
    return {
      hasError: true,
      error,
      errorInfo: null,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Async error caught:', error, errorInfo);
    
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }
    
    this.setState({ errorInfo });
  }

  handleReset = () => {
    this.setState(prevState => ({
      hasError: false,
      error: null,
      errorInfo: null,
      resetKey: prevState.resetKey + 1,
    }));
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="error-boundary-async">
          <h3>Failed to load content</h3>
          <p>{this.state.error?.message || 'An unexpected error occurred'}</p>
          <button onClick={this.handleReset}>Retry</button>
        </div>
      );
    }

    return <div key={this.state.resetKey}>{this.props.children}</div>;
  }
}

// Route-level error boundary with navigation support
interface RouteErrorBoundaryProps extends Props {
  resetPath?: string;
}

export class RouteErrorBoundary extends Component<RouteErrorBoundaryProps, State> {
  constructor(props: RouteErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error,
      errorInfo: null,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Route error:', error, errorInfo);
    
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }
    
    this.setState({ errorInfo });
  }

  handleGoHome = () => {
    // Navigate to home or reset path
    window.location.href = this.props.resetPath || '/';
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="error-boundary-route">
          <h2>Page Error</h2>
          <p>We encountered an error loading this page.</p>
          <div className="error-actions">
            <button onClick={this.handleGoHome}>Go to Home</button>
            <button onClick={() => window.location.reload()}>Reload Page</button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
