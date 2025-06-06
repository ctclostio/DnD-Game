import { useCallback } from 'react';
import { useDispatch } from 'react-redux';

interface ErrorHandlerOptions {
  showNotification?: boolean;
  logToConsole?: boolean;
  fallbackMessage?: string;
}

export function useErrorHandler(defaultOptions?: ErrorHandlerOptions) {
  const dispatch = useDispatch();

  const handleError = useCallback((error: Error | unknown, options?: ErrorHandlerOptions) => {
    const config = { ...defaultOptions, ...options };
    const errorMessage = error instanceof Error ? error.message : 'An unexpected error occurred';
    
    // Log to console if enabled
    if (config.logToConsole !== false) {
      console.error('Error handled:', error);
    }
    
    // Show notification if enabled
    if (config.showNotification !== false) {
      // Dispatch error notification action
      dispatch({
        type: 'ui/showNotification',
        payload: {
          type: 'error',
          message: config.fallbackMessage || errorMessage,
        },
      });
    }
    
    // Return the error for further handling if needed
    return error;
  }, [dispatch, defaultOptions]);

  const handleAsyncError = useCallback(async <T,>(
    asyncFn: () => Promise<T>,
    options?: ErrorHandlerOptions
  ): Promise<T | null> => {
    try {
      return await asyncFn();
    } catch (error) {
      handleError(error, options);
      return null;
    }
  }, [handleError]);

  return { handleError, handleAsyncError };
}

// Hook for handling form errors
export function useFormErrorHandler() {
  const { handleError } = useErrorHandler();

  const handleFormError = useCallback((error: unknown, fieldErrors?: Record<string, string>) => {
    if (error instanceof Error && error.message.includes('validation')) {
      // Handle validation errors differently
      return fieldErrors || {};
    }
    
    handleError(error, {
      fallbackMessage: 'Failed to submit form. Please try again.',
    });
    
    return {};
  }, [handleError]);

  return { handleFormError };
}

// Hook for handling API errors
export function useApiErrorHandler() {
  const { handleError } = useErrorHandler();

  const handleApiError = useCallback((error: unknown) => {
    let message = 'An error occurred while communicating with the server';
    
    if (error instanceof Error) {
      if (error.message.includes('401')) {
        message = 'Your session has expired. Please log in again.';
      } else if (error.message.includes('403')) {
        message = 'You do not have permission to perform this action.';
      } else if (error.message.includes('404')) {
        message = 'The requested resource was not found.';
      } else if (error.message.includes('500')) {
        message = 'Server error. Please try again later.';
      }
    }
    
    handleError(error, { fallbackMessage: message });
  }, [handleError]);

  return { handleApiError };
}