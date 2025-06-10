import { renderHook, act } from '@testing-library/react';
import { useOptimizedState, useOptimizedForm } from '../useOptimizedState';

describe('useOptimizedState', () => {
  it('should initialize with initial value', () => {
    const { result } = renderHook(() => useOptimizedState('initial'));
    
    expect(result.current.value).toBe('initial');
    expect(result.current.previousValue).toBeUndefined();
  });

  it('should update value and track changes', () => {
    const { result } = renderHook(() => useOptimizedState(10));
    
    act(() => {
      result.current.setValue(20);
    });

    expect(result.current.value).toBe(20);
    expect(result.current.previousValue).toBe(10);
  });

  it('should not update when value is the same', () => {
    const onChange = jest.fn();
    const { result } = renderHook(() => useOptimizedState('test', { onUpdate: onChange }));
    
    // Set same value
    act(() => {
      result.current.setValue('test');
    });

    expect(onChange).not.toHaveBeenCalled();
  });

  it('should use custom compare function', () => {
    const compareObjects = (a: any, b: any) => a.id === b.id;
    const { result } = renderHook(() =>
      useOptimizedState({ id: 1, name: 'test' }, { compareFunction: compareObjects })
    );
    
    // Update with same id but different name
    act(() => {
      result.current.setValue({ id: 1, name: 'updated' });
    });

    // Should not update due to custom compare
    expect(result.current.value.name).toBe('test');

    // Update with different id
    act(() => {
      result.current.setValue({ id: 2, name: 'new' });
    });

    expect(result.current.value.id).toBe(2);
  });

  it('should call onChange callback', () => {
    const onChange = jest.fn();
    const { result } = renderHook(() => useOptimizedState(0, { onUpdate: onChange }));
    
    act(() => {
      result.current.setValue(prev => prev + 5);
    });

    expect(onChange).toHaveBeenCalledWith(5, 0);
    expect(onChange).toHaveBeenCalledTimes(1);
  });

  it('should reset to initial value', () => {
    const { result } = renderHook(() => useOptimizedState('initial'));
    
    // Change value
    act(() => {
      result.current.setValue('changed');
    });
    
    expect(result.current.value).toBe('changed');

    // Reset
    act(() => {
      result.current.reset();
    });

    expect(result.current.value).toBe('initial');
    expect(result.current.previousValue).toBe('changed');
  });

  it('should handle functional updates', () => {
    const { result } = renderHook(() => useOptimizedState(10));
    
    act(() => {
      result.current.setValue(prev => prev + 5);
    });

    expect(result.current.value).toBe(15);
    expect(result.current.previousValue).toBe(10);
  });
});

describe('useOptimizedForm', () => {
  const initialValues = {
    username: '',
    email: '',
    age: 0,
  };

  it('should initialize with initial values', () => {
    const { result } = renderHook(() => useOptimizedForm(initialValues));
    
    expect(result.current.values).toEqual(initialValues);
    expect(result.current.errors).toEqual({});
    expect(result.current.touched).toEqual({});
    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.isValid).toBe(true);
    expect(result.current.isDirty).toBe(false);
  });

  it('should update field value', () => {
    const { result } = renderHook(() => useOptimizedForm(initialValues));
    
    act(() => {
      result.current.setFieldValue('username', 'testuser');
    });

    expect(result.current.values.username).toBe('testuser');
    expect(result.current.touched.username).toBe(true);
    expect(result.current.isDirty).toBe(true);
  });

  it('should not update if value is the same', () => {
    const { result } = renderHook(() => useOptimizedForm({ name: 'test' }));
    
    const initialRenderCount = result.current.values;
    
    act(() => {
      result.current.setFieldValue('name', 'test');
    });

    // Values object should be the same reference
    expect(result.current.values).toBe(initialRenderCount);
  });

  it('should validate on change when validateOnChange is true', () => {
    const validators = {
      email: (value: string) => {
        if (!value.includes('@')) return 'Invalid email';
        return undefined;
      },
    };

    const { result } = renderHook(() =>
      useOptimizedForm(initialValues, {
        validator: (values) => ({
          email: validators.email(values.email) || ''
        }),
        validateOnChange: true
      })
    );
    
    act(() => {
      result.current.setFieldValue('email', 'invalid');
    });

    expect(result.current.errors.email).toBe('Invalid email');
    expect(result.current.isValid).toBe(false);

    act(() => {
      result.current.setFieldValue('email', 'valid@email.com');
    });

    expect(result.current.errors.email).toBeUndefined();
    expect(result.current.isValid).toBe(true);
  });

  it('should set multiple field values', () => {
    const { result } = renderHook(() => useOptimizedForm(initialValues));
    
    act(() => {
      result.current.setMultipleValues({
        username: 'newuser',
        email: 'new@email.com',
      });
    });

    expect(result.current.values.username).toBe('newuser');
    expect(result.current.values.email).toBe('new@email.com');
    expect(result.current.values.age).toBe(0); // Unchanged
  });

  it('should set field error', () => {
    const { result } = renderHook(() => useOptimizedForm(initialValues));
    
    act(() => {
      result.current.setFieldError('username', 'Username is required');
    });

    expect(result.current.errors.username).toBe('Username is required');
    expect(result.current.isValid).toBe(false);
  });

  it('should set multiple errors', () => {
    const { result } = renderHook(() => useOptimizedForm(initialValues));
    
    act(() => {
    act(() => {
      result.current.setFieldError('username', 'Required');
      result.current.setFieldError('email', 'Invalid');
    });
        username: 'Required',
        email: 'Invalid',
      });
    });

    expect(result.current.errors.username).toBe('Required');
    expect(result.current.errors.email).toBe('Invalid');
    expect(result.current.isValid).toBe(false);
  });

  it('should handle form submission', async () => {
    const onSubmit = jest.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() => useOptimizedForm(initialValues));
    
    const mockEvent = {
      preventDefault: jest.fn(),
    } as any;

    await act(async () => {
      await result.current.handleSubmit(onSubmit)(mockEvent);
    });

    expect(mockEvent.preventDefault).toHaveBeenCalled();
    expect(onSubmit).toHaveBeenCalledWith(initialValues);
    expect(result.current.isSubmitting).toBe(false);
  });

  it('should validate before submission', async () => {
    const validators = {
      username: (value: string) => {
        if (!value) return 'Username is required';
        return undefined;
      },
    };

    const onSubmit = jest.fn();
    const { result } = renderHook(() =>
      useOptimizedForm(initialValues, {
        validator: (values) => ({ username: validators.username(values.username) || '' })
      })
    );
    
    const mockEvent = {
      preventDefault: jest.fn(),
    } as any;

    await act(async () => {
      await result.current.handleSubmit(onSubmit)(mockEvent);
    });

    expect(onSubmit).not.toHaveBeenCalled();
    expect(result.current.errors.username).toBe('Username is required');
    expect(result.current.isSubmitting).toBe(false);
  });

  it('should reset form', () => {
    const { result } = renderHook(() => useOptimizedForm(initialValues));
    
    // Make changes
    act(() => {
      result.current.setFieldValue('username', 'changed');
      result.current.setFieldError('email', 'Error');
    });

    expect(result.current.isDirty).toBe(true);
    expect(result.current.errors.email).toBe('Error');

    // Reset
    act(() => {
      result.current.reset();
    });

    expect(result.current.values).toEqual(initialValues);
    expect(result.current.errors).toEqual({});
    expect(result.current.touched).toEqual({});
    expect(result.current.isDirty).toBe(false);
  });

  it('should validate entire form', () => {
    const validators = {
      username: (value: string) => !value ? 'Required' : undefined,
      email: (value: string) => {
        if (!value) return 'Required';
        if (!value.includes('@')) return 'Invalid email';
        return undefined;
      },
    };

    const { result } = renderHook(() =>
      useOptimizedForm(initialValues, {
        validator: (values) => ({
          username: validators.username(values.username),
          email: validators.email(values.email) || ''
        })
      })
    );
    
    act(() => {
      const errors = result.current.validate(values);
      expect(errors).toEqual({
        username: 'Required',
        email: 'Required',
      });
    });

    // Update values
    act(() => {
      result.current.setMultipleValues({
        username: 'testuser',
        email: 'invalid',
      });
    });

    act(() => {
      const errors = result.current.validate(result.current.values);
      expect(errors).toEqual({
        email: 'Invalid email',
      });
    });
  });

  it('should track individual field validation', () => {
    const validators = {
      age: (value: number) => value < 18 ? 'Must be 18 or older' : undefined,
    };

    const { result } = renderHook(() =>
      useOptimizedForm(initialValues, {
        validator: (values) => ({ age: validators.age(values.age) || '' })
      })
    );
    
    act(() => {
      const error = validators.age(result.current.values.age);
      expect(error).toBe('Must be 18 or older');
    });

    act(() => {
      result.current.setFieldValue('age', 21);
      const error = validators.age(result.current.values.age);
      expect(error).toBeUndefined();
    });
  });
});