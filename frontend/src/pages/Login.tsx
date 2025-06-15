import React, { useState, FormEvent } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { RootState, AppDispatch } from '@store/index';
import { login, clearError } from '@store/slices/authSlice';

const Login: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const navigate = useNavigate();
  const { isLoading, error } = useSelector((state: RootState) => state.auth);

  const [formData, setFormData] = useState({
    username: '',
    password: '',
  });

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();

    const result = await dispatch(login(formData));

    if (login.fulfilled.match(result)) {
      navigate('/dashboard');
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    if (error) {
      dispatch(clearError());
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h2>Login to D&D Online</h2>
        
        {error && (
          <div id="login-error" className="alert alert-error" role="alert">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="username">Username</label>
            <input
              type="text"
              id="username"
              name="username"
              value={formData.username}
              onChange={handleInputChange}
              required
              autoFocus
              autoComplete="username"
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
              name="password"
              value={formData.password}
              onChange={handleInputChange}
              required
              autoComplete="current-password"
              aria-describedby={error ? 'login-error' : undefined}
              aria-invalid={error ? true : undefined}
            />
          </div>

          <button
            type="submit"
            className="btn btn-primary btn-block"
            disabled={isLoading}
          >
            {isLoading ? 'Logging in...' : 'Login'}
          </button>
        </form>

        <div className="auth-footer">
          <p>
            Don't have an account? <Link to="/register">Register here</Link>
          </p>
        </div>
      </div>
    </div>
  );
};

export default Login;
