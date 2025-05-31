import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import AuthGuard from './components/AuthGuard';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import CharacterBuilder from './pages/CharacterBuilder';
import Characters from './pages/Characters';
import GameSession from './pages/GameSession';
import CombatView from './pages/Combat';
import WorldBuilder from './pages/WorldBuilder';
import DMTools from './pages/DMTools';

const App: React.FC = () => {
  return (
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
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
        </Route>
      </Route>
    </Routes>
  );
};

export default App;