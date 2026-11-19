// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {FlowExecutionFailure, getErrorMessage} from '@thunderid/utils';

/** The part of i18next's translate function this uses. */
type TranslateFn = (key: string, options?: Record<string, unknown>) => string;

/** Flow error codes the owner verification flow answers a send with, and what each means to an administrator. */
const OWNER_VERIFICATION_ERRORS: Record<string, {key: string; defaultValue: string}> = {
  'FET-1038': {
    key: 'agents:owner.verification.errors.noEmailSender',
    defaultValue:
      'No email sender is selected on the agent owner verification flow. Select one on its Send Email step, then try again.',
  },
  'FET-1042': {
    key: 'agents:owner.verification.errors.noSmsSender',
    defaultValue:
      'No SMS sender is selected on the agent owner verification flow. Select one on its Send SMS step, then try again.',
  },
  'FET-1112': {
    key: 'agents:owner.verification.errors.notVerifiable',
    defaultValue:
      'This owner cannot be verified. Only an external owner named by an email address or mobile number can.',
  },
};

/** The message for a failed send of an owner verification link. */
export default function getOwnerVerificationErrorMessage(error: Error, t: TranslateFn): string {
  const known = error instanceof FlowExecutionFailure && error.code ? OWNER_VERIFICATION_ERRORS[error.code] : undefined;
  if (known) {
    return t(known.key, {defaultValue: known.defaultValue});
  }
  return getErrorMessage(
    error,
    t,
    'agents:owner.verification.sendFailed',
    'The verification link could not be sent. Please try again.',
  );
}
