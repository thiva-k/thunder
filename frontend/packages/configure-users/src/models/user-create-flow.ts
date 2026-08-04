// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

/**
 * User creation step identifiers used in the creation wizard flow
 * to track the current step and navigate between steps.
 *
 * @public
 */
export const UserCreateFlowStep = {
  USER_TYPE: 'USER_TYPE',
  ORGANIZATION_UNIT: 'ORGANIZATION_UNIT',
  USER_DETAILS: 'USER_DETAILS',
} as const;

/**
 * User creation step type
 *
 * @public
 */
export type UserCreateFlowStep = keyof typeof UserCreateFlowStep;
