// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {PropsWithChildren} from 'react';
import {useState, useMemo, useCallback} from 'react';
import UserCreateContext, {type UserCreateContextType} from './UserCreateContext';
import {UserCreateFlowStep} from '../../models/user-create-flow';
import type {SchemaInterface} from '../../models/users';

/**
 * Initial state values for user creation.
 *
 * @internal
 */
const INITIAL_STATE = {
  currentStep: UserCreateFlowStep.USER_TYPE as UserCreateFlowStep,
  selectedSchema: null as SchemaInterface | null,
  selectedOuId: null as string | null,
  formValues: {} as Record<string, unknown>,
  error: null as string | null,
};

/**
 * React context provider component that provides user creation state
 * to all child components in the wizard flow.
 *
 * @public
 */
export default function UserCreateProvider({children}: PropsWithChildren) {
  const [currentStep, setCurrentStep] = useState<UserCreateFlowStep>(INITIAL_STATE.currentStep);
  const [selectedSchema, setSelectedSchema] = useState<SchemaInterface | null>(INITIAL_STATE.selectedSchema);
  const [selectedOuId, setSelectedOuId] = useState<string | null>(INITIAL_STATE.selectedOuId);
  const [formValues, setFormValues] = useState<Record<string, unknown>>(INITIAL_STATE.formValues);
  const [error, setError] = useState<string | null>(INITIAL_STATE.error);

  const reset = useCallback((): void => {
    setCurrentStep(INITIAL_STATE.currentStep);
    setSelectedSchema(INITIAL_STATE.selectedSchema);
    setSelectedOuId(INITIAL_STATE.selectedOuId);
    setFormValues(INITIAL_STATE.formValues);
    setError(INITIAL_STATE.error);
  }, []);

  const contextValue: UserCreateContextType = useMemo(
    () => ({
      currentStep,
      setCurrentStep,
      selectedSchema,
      setSelectedSchema,
      selectedOuId,
      setSelectedOuId,
      formValues,
      setFormValues,
      error,
      setError,
      reset,
    }),
    [currentStep, selectedSchema, selectedOuId, formValues, error, reset],
  );

  return <UserCreateContext.Provider value={contextValue}>{children}</UserCreateContext.Provider>;
}
