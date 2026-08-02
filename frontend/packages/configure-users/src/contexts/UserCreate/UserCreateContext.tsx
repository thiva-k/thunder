// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {Context} from 'react';
import {createContext} from 'react';
import type {UserCreateFlowStep} from '../../models/user-create-flow';
import type {SchemaInterface} from '../../models/users';

/**
 * User creation context state interface.
 *
 * Provides centralized state management for the user creation wizard flow.
 *
 * @public
 */
export interface UserCreateContextType {
  currentStep: UserCreateFlowStep;
  setCurrentStep: (step: UserCreateFlowStep) => void;

  selectedSchema: SchemaInterface | null;
  setSelectedSchema: (schema: SchemaInterface | null) => void;

  selectedOuId: string | null;
  setSelectedOuId: (ouId: string | null) => void;

  formValues: Record<string, unknown>;
  setFormValues: (values: Record<string, unknown>) => void;

  error: string | null;
  setError: (error: string | null) => void;

  reset: () => void;
}

const UserCreateContext: Context<UserCreateContextType | undefined> = createContext<UserCreateContextType | undefined>(
  undefined,
);

export default UserCreateContext;
