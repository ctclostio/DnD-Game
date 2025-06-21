import { useCallback } from 'react';
import { useDispatch } from 'react-redux';
import { addNotification } from '../store/slices/uiSlice';

interface ErrorHandlerOptions {
  showNotification?: boolean;
  logToConsole?: boolean;
  fallbackMessage?: string;
}

export function useErrorHandler(defaultOptions?: ErrorHandlerOptions | string) {
  const dispatch = useDispatch();
  
  // Handle legacy string parameter for backward compatibility
  const normalizedDefaultOptions = typeof defaultOptions === 'string' 
    ? { fallbackMessage: defaultOptions } 
    : defaultOptions;

  const handleError = useCallback((error: Error | unknown, options?: ErrorHandlerOptions) => {
    const config = { ...normalizedDefaultOptions, ...options };
    
    let errorMessage: string;
    if (error instanceof Error) {
      errorMessage = error.message;
    } else if (typeof error === 'string') {
      errorMessage = error;
    } else {
      errorMessage = config.fallbackMessage ?? 'An unexpected error occurred';
    }
    
    // Log to console if enabled
    if (config.logToConsole !== false) {
      console.error('Error:', error);
    }
    
    // Show notification if enabled
    if (config.showNotification !== false) {
      // Dispatch error notification action
      dispatch(addNotification({
        type: 'error',
        message: errorMessage,
      }));
    }
    
    // Return the error for further handling if needed
    return error;
  }, [dispatch, normalizedDefaultOptions]);

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

  const handleFormError = useCallback((error: any) => {
    // Check for validation errors in response
    if (error?.response?.data?.errors) {
      // This is a validation error with field-specific errors
      handleError(error, {
        fallbackMessage: 'Please check the form errors',
      });
      return error.response.data.errors;
    }
    
    // Check for custom error message in response
    if (error?.response?.data?.message) {
      handleError(error, {
        fallbackMessage: error.response.data.message,
      });
      return {};
    }
    
    // Handle regular errors
    handleError(error, {
      fallbackMessage: error instanceof Error ? error.message : 'Failed to submit form. Please try again.',
    });
    
    return {};
  }, [handleError]);

  return { handleFormError };
}

// Hook for handling API errors
export function useApiErrorHandler() {
  const { handleError } = useErrorHandler();

  const handleApiError = useCallback((error: any) => {
    let message = 'An error occurred while communicating with the server';
    
    // Check for axios-style error structure
    if (error?.response) {
      const status = error.response.status;
      const customMessage = error.response.data?.message;
      
      if (customMessage) {
        message = customMessage;
      } else if (status === 401) {
        message = 'Your session has expired. Please log in again.';
      } else if (status === 403) {
        message = 'You do not have permission to perform this action.';
      } else if (status === 404) {
        message = 'The requested resource was not found.';
      } else if (status === 500) {
        message = 'A server error occurred. Please try again later.';
      }
    } else if (error?.request) {
      // Network error
      message = 'Network error. Please check your connection.';
    } else if (error?.code === 'ECONNABORTED') {
      // Timeout error
      message = 'Request timed out. Please try again.';
    }
    
    handleError(error, { fallbackMessage: message });
  }, [handleError]);

  const handleAsyncApiCall = useCallback(<T,>(
    apiCall: () => Promise<T>
  ) => {
    return async (): Promise<T | null> => {
      try {
        return await apiCall();
      } catch (error) {
        handleApiError(error);
        return null;
      }
    };
  }, [handleApiError]);

  return { handleApiError, handleAsyncApiCall };
}