// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useContext} from 'react';
import UserCreateContext, {type UserCreateContextType} from './UserCreateContext';

/**
 * Hook to access the user creation context.
 *
 * @throws Error if used outside of UserCreateProvider
 * @returns The user creation context value
 *
 * @public
 */
export default function useUserCreate(): UserCreateContextType {
  const context = useContext(UserCreateContext);

  if (!context) {
    throw new Error('useUserCreate must be used within a UserCreateProvider');
  }

  return context;
}
