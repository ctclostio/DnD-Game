import { useState, useCallback, useRef, useMemo } from 'react';
import { unstable_batchedUpdates } from 'react-dom';

/**
 * Hook that provides optimized state management with built-in performance features
 */
export function useOptimizedState<T>(
  initialValue: T | (() => T),
  options?: {
    compareFunction?: (prev: T, next: T) => boolean;
    onUpdate?: (newValue: T, prevValue: T) => void;
  }
) {
  const [state, setState] = useState(initialValue);
  const previousValueRef = useRef<T | undefined>(undefined);
  const memoizedOptions = useMemo(() => options, [options?.compareFunction, options?.onUpdate]);
  
  // Custom setState that only updates if value actually changed
  const setOptimizedState = useCallback((newValue: T | ((prev: T) => T)) => {
    setState(prev => {
      const nextValue = typeof newValue === 'function' 
        ? (newValue as (prev: T) => T)(prev)
        : newValue;
      
      // Use custom compare function if provided
      const hasChanged = memoizedOptions?.compareFunction
        ? !memoizedOptions.compareFunction(prev, nextValue)
        : prev !== nextValue;
      
      if (hasChanged) {
        previousValueRef.current = prev;
        
        // Call onUpdate callback if provided
        if (memoizedOptions?.onUpdate) {
          memoizedOptions.onUpdate(nextValue, prev);
        }
        
        return nextValue;
      }
      
      return prev;
    });
  }, [memoizedOptions]);
  
  // Reset to initial value
  const reset = useCallback(() => {
    const initial = typeof initialValue === 'function' 
      ? (initialValue as () => T)()
      : initialValue;
    setOptimizedState(initial);
  }, [initialValue, setOptimizedState]);
  
  return {
    value: state,
    setValue: setOptimizedState,
    previousValue: previousValueRef.current,
    reset,
  };
}

/**
 * Hook for managing complex form state with optimizations
 */
export function useOptimizedForm<T extends Record<string, unknown>>(
  initialValues: T,
  options?: {
    validateOnChange?: boolean;
    validator?: (values: T) => Record<string, string>;
    shallowCompare?: boolean;
    ariaLive?: boolean;
  }
) {
  const [values, setValues] = useState(initialValues);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  // Memoized validation
  const validate = useMemo(() => {
    return options?.validator || (() => ({}));
  }, [options?.validator]);
  
  // Update single field
  const setFieldValue = useCallback(<K extends keyof T>(field: K, value: T[K]) => {
    unstable_batchedUpdates(() => {
      setValues(prev => {
      if (prev[field] === value) return prev;
      
      const newValues = { ...prev, [field]: value };
      
      // Mark field as touched
      setTouched(prevTouched => ({ ...prevTouched, [field]: true }));
      
      // Validate on change if enabled
      if (options?.validateOnChange) {
        const newErrors = validate(newValues);
        // Filter out empty error strings
        const filteredErrors = Object.entries(newErrors).reduce((acc, [key, error]) => {
          if (error) acc[key] = error;
          return acc;
        }, {} as Record<string, string>);
        setErrors(filteredErrors);
      }
      
      return newValues;
    });
    });
  }, [validate, options?.validateOnChange]);
  
  // Mark field as touched
  const setFieldTouched = useCallback((field: keyof T, isTouched = true) => {
    setTouched(prev => {
      if (prev[field as string] === isTouched) return prev;
      return { ...prev, [field]: isTouched };
    });
  }, []);
  
  // Bulk update values
  const setMultipleValues = useCallback((updates: Partial<T>) => {
    unstable_batchedUpdates(() => {
      setValues(prev => {
      const hasChanges = Object.entries(updates).some(
        ([key, value]) => prev[key] !== value
      );
      
      if (!hasChanges) return prev;
      
      return { ...prev, ...updates };
    });
    });
  }, []);
  
  // Reset form
  const reset = useCallback(() => {
    setValues(initialValues);
    setErrors({});
    setTouched({});
    setIsSubmitting(false);
  }, [initialValues]);
  
  // Submit handler
  const handleSubmit = useCallback(
    (onSubmit: (values: T) => void | Promise<void>, options?: { ariaLive?: boolean }) => {
      return async (e?: React.FormEvent) => {
        if (e) e.preventDefault();
        
        setIsSubmitting(true);
        const validationErrors = validate(values);
        setErrors(prev => ({
          ...prev,
          ...validationErrors,
          ...(options?.ariaLive ? { __ariaLive: 'true' } : {})
        }));
        
        if (Object.keys(validationErrors).length === 0) {
          try {
            await onSubmit(values);
          } catch (error) {
            console.error('Form submission error:', error);
          }
        }
        
        setIsSubmitting(false);
      };
    },
    [values, validate]
  );
  
  // Check if form is valid
  const isValid = useMemo(() => {
    // Check both validation errors and manually set errors
    const validationErrors = validate(values);
    const hasValidationErrors = Object.values(validationErrors).some(error => !!error);
    const hasManualErrors = Object.values(errors).some(error => !!error);
    return !hasValidationErrors && !hasManualErrors;
  }, [values, validate, errors]);
  
  // Check if form is dirty
  const isDirty = useMemo(() => {
    if (options?.shallowCompare) {
      return Object.keys(values).some(key => values[key] !== initialValues[key]);
    }
    return !Object.is(values, initialValues);
  }, [values, initialValues, options?.shallowCompare]);
  
  return {
    values,
    errors,
    touched,
    isSubmitting,
    isValid,
    isDirty,
    setFieldValue,
    setFieldTouched,
    setFieldError: <K extends keyof T>(field: K, message: string) =>
      setErrors(prev => ({ ...prev, [field]: message })),
    setMultipleValues,
    handleSubmit,
    reset,
  };
}
