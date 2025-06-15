import React from 'react';
import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { RootState, AppDispatch } from '@store/index';
import { logout } from '@store/slices/authSlice';

const Layout: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const navigate = useNavigate();
  const user = useSelector((state: RootState) => state.auth.user);

  const handleLogout = async () => {
    await dispatch(logout());
    navigate('/login');
  };

  return (
    <div className="app-container">
      <nav className="navbar">
        <div className="nav-brand">
          <h1>D&D Online</h1>
        </div>
        
        <ul className="nav-links">
          <li>
            <NavLink to="/dashboard" className={({ isActive }) => isActive ? 'active' : ''}>
              Dashboard
            </NavLink>
          </li>
          <li>
            <NavLink to="/characters" className={({ isActive }) => isActive ? 'active' : ''}>
              Characters
            </NavLink>
          </li>
          {user?.role === 'dm' && (
            <>
              <li>
                <NavLink to="/world-builder" className={({ isActive }) => isActive ? 'active' : ''}>
                  World Builder
                </NavLink>
              </li>
              <li>
                <NavLink to="/dm-tools" className={({ isActive }) => isActive ? 'active' : ''}>
                  DM Tools
                </NavLink>
              </li>
            </>
          )}
        </ul>

        <div className="nav-user">
          <span className="username">{user?.username}</span>
          <button onClick={handleLogout} className="btn-logout">
            Logout
          </button>
        </div>
      </nav>

      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
};

export default Layout;
