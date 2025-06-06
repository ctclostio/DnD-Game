import { useState, useCallback, useRef, useMemo } from 'react';

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
  const previousValueRef = useRef(state);
  
  // Custom setState that only updates if value actually changed
  const setOptimizedState = useCallback((newValue: T | ((prev: T) => T)) => {
    setState(prev => {
      const nextValue = typeof newValue === 'function' 
        ? (newValue as (prev: T) => T)(prev)
        : newValue;
      
      // Use custom compare function if provided
      const hasChanged = options?.compareFunction
        ? !options.compareFunction(prev, nextValue)
        : prev !== nextValue;
      
      if (hasChanged) {
        previousValueRef.current = prev;
        
        // Call onUpdate callback if provided
        if (options?.onUpdate) {
          options.onUpdate(nextValue, prev);
        }
        
        return nextValue;
      }
      
      return prev;
    });
  }, [options?.compareFunction, options?.onUpdate]);
  
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
export function useOptimizedForm<T extends Record<string, any>>(
  initialValues: T,
  options?: {
    validateOnChange?: boolean;
    validator?: (values: T) => Record<string, string>;
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
  const setFieldValue = useCallback((field: keyof T, value: any) => {
    setValues(prev => {
      if (prev[field] === value) return prev;
      
      const newValues = { ...prev, [field]: value };
      
      // Validate on change if enabled
      if (options?.validateOnChange && touched[field as string]) {
        const newErrors = validate(newValues);
        setErrors(newErrors);
      }
      
      return newValues;
    });
  }, [touched, validate, options?.validateOnChange]);
  
  // Mark field as touched
  const setFieldTouched = useCallback((field: keyof T, isTouched = true) => {
    setTouched(prev => {
      if (prev[field as string] === isTouched) return prev;
      return { ...prev, [field]: isTouched };
    });
  }, []);
  
  // Bulk update values
  const setMultipleValues = useCallback((updates: Partial<T>) => {
    setValues(prev => {
      const hasChanges = Object.entries(updates).some(
        ([key, value]) => prev[key] !== value
      );
      
      if (!hasChanges) return prev;
      
      return { ...prev, ...updates };
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
    (onSubmit: (values: T) => void | Promise<void>) => {
      return async (e?: React.FormEvent) => {
        if (e) e.preventDefault();
        
        setIsSubmitting(true);
        const validationErrors = validate(values);
        setErrors(validationErrors);
        
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
    return Object.keys(validate(values)).length === 0;
  }, [values, validate]);
  
  // Check if form is dirty
  const isDirty = useMemo(() => {
    return Object.entries(values).some(
      ([key, value]) => value !== initialValues[key]
    );
  }, [values, initialValues]);
  
  return {
    values,
    errors,
    touched,
    isSubmitting,
    isValid,
    isDirty,
    setFieldValue,
    setFieldTouched,
    setMultipleValues,
    handleSubmit,
    reset,
  };
}