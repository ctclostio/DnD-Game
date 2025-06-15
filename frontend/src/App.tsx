import React, { Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import AuthGuard from './components/AuthGuard';
import LoadingSpinner from './components/LoadingSpinner';
import { ErrorBoundary, RouteErrorBoundary } from './components/ErrorBoundary';
import { useAuthSync } from './hooks/useAuthSync';

// Lazy load all page components
const Login = lazy(() => import('./pages/Login'));
const Register = lazy(() => import('./pages/Register'));
const Dashboard = lazy(() => import('./pages/Dashboard'));
const CharacterBuilder = lazy(() => import('./pages/CharacterBuilder'));
const Characters = lazy(() => import('./pages/Characters'));
const GameSession = lazy(() => import('./pages/GameSession'));
const CombatView = lazy(() => import('./pages/Combat'));
const WorldBuilder = lazy(() => import('./pages/WorldBuilder'));
const DMTools = lazy(() => import('./pages/DMTools'));
const RuleBuilder = lazy(() => import('./components/RuleBuilder'));

const App: React.FC = () => {
  useAuthSync();
  return (
    <ErrorBoundary
      onError={(error, errorInfo) => {
        // Send to error reporting service in production
        console.error('App Error:', error, errorInfo);
      }}
    >
      <Suspense fallback={<LoadingSpinner fullScreen />}>
        <Routes>
          <Route path="/login" element={
            <RouteErrorBoundary>
              <Login />
            </RouteErrorBoundary>
          } />
          <Route path="/register" element={
            <RouteErrorBoundary>
              <Register />
            </RouteErrorBoundary>
          } />
          
          <Route element={<AuthGuard />}>
            <Route element={<Layout />}>
              <Route path="/dashboard" element={
                <RouteErrorBoundary>
                  <Dashboard />
                </RouteErrorBoundary>
              } />
              <Route path="/characters" element={
                <RouteErrorBoundary>
                  <Characters />
                </RouteErrorBoundary>
              } />
              <Route path="/characters/new" element={
                <RouteErrorBoundary>
                  <CharacterBuilder />
                </RouteErrorBoundary>
              } />
              <Route path="/characters/:id" element={
                <RouteErrorBoundary>
                  <CharacterBuilder />
                </RouteErrorBoundary>
              } />
              <Route path="/game-session/:id" element={
                <RouteErrorBoundary>
                  <GameSession />
                </RouteErrorBoundary>
              } />
              <Route path="/combat/:sessionId" element={
                <RouteErrorBoundary>
                  <CombatView />
                </RouteErrorBoundary>
              } />
              <Route path="/world-builder" element={
                <RouteErrorBoundary>
                  <WorldBuilder />
                </RouteErrorBoundary>
              } />
              <Route path="/dm-tools" element={
                <RouteErrorBoundary>
                  <DMTools />
                </RouteErrorBoundary>
              } />
              <Route path="/dm-tools/rule-builder" element={
                <RouteErrorBoundary>
                  <RuleBuilder />
                </RouteErrorBoundary>
              } />
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
            </Route>
          </Route>
        </Routes>
      </Suspense>
    </ErrorBoundary>
  );
};

export default App;
