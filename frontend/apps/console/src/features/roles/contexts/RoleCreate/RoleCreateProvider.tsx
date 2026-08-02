// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {PropsWithChildren} from 'react';
import {useState, useMemo, useCallback} from 'react';
import RoleCreateContext, {type RoleCreateContextType} from './RoleCreateContext';
import type {ResourcePermissions} from '../../models/role';
import {RoleCreateFlowStep} from '../../models/role-create-flow';

const INITIAL_STATE = {
  currentStep: RoleCreateFlowStep.ORGANIZATION_UNIT as RoleCreateFlowStep,
  name: '',
  ouId: '',
  error: null as string | null,
  permissions: [] as ResourcePermissions[],
};

/**
 * React context provider component that provides role creation state
 * to all child components in the wizard flow.
 *
 * @public
 */
export default function RoleCreateProvider({children}: PropsWithChildren) {
  const [currentStep, setCurrentStep] = useState<RoleCreateFlowStep>(INITIAL_STATE.currentStep);
  const [name, setName] = useState<string>(INITIAL_STATE.name);
  const [ouId, setOuId] = useState<string>(INITIAL_STATE.ouId);
  const [error, setError] = useState<string | null>(INITIAL_STATE.error);
  const [permissions, setPermissions] = useState<ResourcePermissions[]>(INITIAL_STATE.permissions);

  const reset = useCallback((): void => {
    setCurrentStep(INITIAL_STATE.currentStep);
    setName(INITIAL_STATE.name);
    setOuId(INITIAL_STATE.ouId);
    setError(INITIAL_STATE.error);
    setPermissions(INITIAL_STATE.permissions);
  }, []);

  const contextValue: RoleCreateContextType = useMemo(
    () => ({
      currentStep,
      setCurrentStep,
      name,
      setName,
      ouId,
      setOuId,
      error,
      setError,
      permissions,
      setPermissions,
      reset,
    }),
    [currentStep, name, ouId, error, permissions, reset],
  );

  return <RoleCreateContext.Provider value={contextValue}>{children}</RoleCreateContext.Provider>;
}
