import React, { Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import AuthGuard from './components/AuthGuard';
import LoadingSpinner from './components/LoadingSpinner';

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
  return (
    <Suspense fallback={<LoadingSpinner fullScreen />}>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        
        <Route element={<AuthGuard />}>
          <Route element={<Layout />}>
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/characters" element={<Characters />} />
            <Route path="/characters/new" element={<CharacterBuilder />} />
            <Route path="/characters/:id" element={<CharacterBuilder />} />
            <Route path="/game-session/:id" element={<GameSession />} />
            <Route path="/combat/:sessionId" element={<CombatView />} />
            <Route path="/world-builder" element={<WorldBuilder />} />
            <Route path="/dm-tools" element={<DMTools />} />
            <Route path="/dm-tools/rule-builder" element={<RuleBuilder />} />
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
          </Route>
        </Route>
      </Routes>
    </Suspense>
  );
};

export default App;