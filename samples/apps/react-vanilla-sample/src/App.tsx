// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import { Navigate, Route, BrowserRouter as Router, Routes, useLocation } from 'react-router';
import HomePage from './pages/HomePage';
import InvitePage from './pages/InvitePage';
import LoginPage from './pages/LoginPage';
import ProfilePage from './pages/ProfilePage';
import AuthProvider from './contexts/AuthProvider';
import useAuth from './hooks/useAuth';
import './App.css';

const App = () => {
  const { token } = useAuth();
  const location = useLocation();

  return (
    <Routes>
      <Route path="/" element={token ? <HomePage /> : <LoginPage />} key={location.key} />
      <Route path="/profile" element={token ? <ProfilePage /> : <Navigate to="/" replace />} />
      <Route path="/invite" element={<InvitePage />} />
    </Routes>
  );
};

const AppWrapper = () => (
  <AuthProvider>
    <Router>
      <App />
    </Router>
  </AuthProvider>
);

export default AppWrapper;
