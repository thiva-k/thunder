// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import React, { createContext } from 'react';
import type { UserProfile } from '../services/userProfileService';

/**
 * AuthContext provides authentication state management for the application.
 * It allows components to access the current authentication token and provides methods
 * to set and clear the token.
 */
type AuthContextType = {
  token: string | null;
  setToken: React.Dispatch<React.SetStateAction<string | null>>;
  clearToken: () => void;
  userProfile: UserProfile | null;
  refreshUserProfile: () => Promise<UserProfile | null>;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export default AuthContext;
