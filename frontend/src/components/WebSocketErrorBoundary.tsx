import React, { Component, ReactNode } from 'react';
import { disconnectWebSocket } from '../services/websocket';

interface Props {
  children: ReactNode;
  roomId?: string;
}

interface State {
  hasError: boolean;
  error: Error | null;
  reconnectAttempts: number;
}

export class WebSocketErrorBoundary extends Component<Props, State> {
  private reconnectTimer: NodeJS.Timeout | null = null;
  private maxReconnectAttempts = 3;

  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      reconnectAttempts: 0,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    // Check if it's a WebSocket-related error
    if (error.message.includes('WebSocket') || error.message.includes('Connection')) {
      return {
        hasError: true,
        error,
      };
    }
    // Let other errors bubble up
    throw error;
  }

  componentDidCatch(error: Error) {
    console.error('WebSocket error caught:', error);
    
    // Clean up WebSocket connection
    this.cleanupWebSocket();
    
    // Attempt to reconnect if under the limit
    if (this.state.reconnectAttempts < this.maxReconnectAttempts) {
      this.scheduleReconnect();
    }
  }

  componentWillUnmount() {
    this.cleanupReconnect();
    this.cleanupWebSocket();
  }

  cleanupWebSocket = () => {
    try {
      disconnectWebSocket();
    } catch (error) {
      console.error('Error during WebSocket cleanup:', error);
    }
  };

  cleanupReconnect = () => {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  };

  scheduleReconnect = () => {
    this.cleanupReconnect();
    
    const delay = Math.min(1000 * Math.pow(2, this.state.reconnectAttempts), 10000);
    
    this.reconnectTimer = setTimeout(() => {
      this.setState(prevState => ({
        hasError: false,
        error: null,
        reconnectAttempts: prevState.reconnectAttempts + 1,
      }));
    }, delay);
  };

  handleManualReconnect = () => {
    this.setState({
      hasError: false,
      error: null,
      reconnectAttempts: 0,
    });
  };

  handleGoBack = () => {
    // Clean up and navigate back
    this.cleanupWebSocket();
    window.history.back();
  };

  render() {
    if (this.state.hasError) {
      const canRetry = this.state.reconnectAttempts < this.maxReconnectAttempts;

      return (
        <div className="websocket-error-boundary">
          <h3>Connection Error</h3>
          <p>We're having trouble connecting to the game server.</p>
          
          {this.state.reconnectAttempts > 0 && (
            <p>Reconnection attempts: {this.state.reconnectAttempts}/{this.maxReconnectAttempts}</p>
          )}
          
          <div className="error-actions">
            {canRetry ? (
              <button onClick={this.handleManualReconnect}>Try Again</button>
            ) : (
              <p>Maximum reconnection attempts reached.</p>
            )}
            <button onClick={this.handleGoBack}>Leave Game</button>
          </div>
          
          {process.env.NODE_ENV === 'development' && (
            <details>
              <summary>Error Details</summary>
              <pre>{this.state.error?.toString()}</pre>
            </details>
          )}
        </div>
      );
    }

    return this.props.children;
  }
}