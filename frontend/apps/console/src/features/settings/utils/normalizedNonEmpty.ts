// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {normalizeOrigin} from './origin';

/**
 * Normalizes each value and drops the empty ones.
 *
 * @param values - The raw origin values to normalize
 * @returns The normalized origins with empty entries removed
 */
export default function normalizedNonEmpty(values: string[]): string[] {
  return values.map(normalizeOrigin).filter((value) => value !== '');
}
