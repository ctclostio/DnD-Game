import React, { useEffect } from 'react';
import authService from '../services/auth';

export function AuthGuard({ children }) {
    useEffect(() => {
        if (!authService.isAuthenticated()) {
            window.location.href = '/login';
        }
    }, []);

    if (!authService.isAuthenticated()) {
        return <div>Redirecting to login...</div>;
    }

    return children;
}