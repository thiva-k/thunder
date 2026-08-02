// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {ApiError} from '@thunderid/types';

/**
 * A minimal subset of the i18next TFunction interface used by {@link getErrorMessage}.
 */
type TranslateFn = (key: string, options?: {defaultValue: string}) => string;

/**
 * Extracts a localized error message from an API error response.
 *
 * Attempts to resolve a specific i18n message for the error code returned
 * by the API (e.g. `errors.APP-1020`). If no specific translation exists,
 * falls back to the provided generic key.
 *
 * @param error - The error thrown by the mutation
 * @param t - The i18next translation function scoped to the relevant namespace
 * @param fallbackKey - i18n key to use when no specific message is found (e.g. `'create.error'`)
 * @returns Localized error message string
 *
 * @example
 * ```typescript
 * onError: (error) => {
 *   showToast(getErrorMessage(error, t, 'create.error'), 'error');
 * }
 * ```
 *
 * @public
 */
export default function getErrorMessage(error: Error, t: TranslateFn, fallbackKey: string): string {
  const apiError = (error as {response?: {data?: ApiError}}).response?.data;

  if (apiError?.code) {
    const specific = t(`errors.${apiError.code}`, {defaultValue: ''});

    if (specific) {
      return specific;
    }
  }

  return t(fallbackKey);
}
