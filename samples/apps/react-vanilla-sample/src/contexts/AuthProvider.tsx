// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import { useState, useEffect, useCallback, useRef } from 'react';
import type { ReactNode } from 'react';
import AuthContext from './AuthContext';
import { getCurrentUserProfile, type UserProfile } from '../services/userProfileService';

/**
 * AuthProvider component to manage authentication state.
 * 
 * @param children - The children components to be wrapped by the AuthProvider.
 * @returns 
 */
const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [token, setToken] = useState<string | null>(() => sessionStorage.getItem('authToken'));
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
  const tokenRef = useRef(token);
  tokenRef.current = token;

  useEffect(() => {
    if (token === null) {
      sessionStorage.removeItem('authToken');
    } else {
      sessionStorage.setItem('authToken', token);
    }
  }, [token]);

  const refreshUserProfile = useCallback(async () => {
    if (!token) {
      setUserProfile(null);
      return null;
    }

    const tokenSnapshot = token;
    const profile = await getCurrentUserProfile(tokenSnapshot);

    if (tokenRef.current === tokenSnapshot) {
      setUserProfile(profile);
    }
    return profile;
  }, [token]);

  useEffect(() => {
    if (!token) {
      setUserProfile(null);
      return;
    }

    void refreshUserProfile().catch(() => {
      setUserProfile(null);
    });
  }, [refreshUserProfile, token]);

  const clearToken = useCallback(() => {
    setToken(null);
    setUserProfile(null);
  }, []);

  return (
    <AuthContext.Provider value={{ token, setToken, clearToken, userProfile, refreshUserProfile }}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthProvider;
